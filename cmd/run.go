//go:build darwin

package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/mnkd/cred/internal/config"
	"github.com/mnkd/cred/internal/env"
	"github.com/mnkd/cred/internal/keyring"
	"github.com/mnkd/cred/internal/mask"
	"github.com/mnkd/cred/internal/ref"
	"github.com/spf13/cobra"
)

var (
	envFiles  []string
	noMasking bool
)

// getSecret is replaced in tests.
var getSecret = keyring.Get

var runCmd = &cobra.Command{
	Use:   "run -- COMMAND [ARGS...]",
	Short: "Run a command with credentials injected as environment variables",
	Long: `Execute a command with credentials injected as environment variables.

References: any environment variable whose value is cred://<namespace>/<KEY>
is replaced with the secret from the keyring. --env-file loads KEY=VALUE
lines from a dotenv file; values may also be references. A value that starts
with cred:/ or cred// but is not a valid reference (e.g. cred:/work/KEY) is
an error, so a typo is not passed on as a plain value.

Namespace injection: every key in the namespace is injected under its own
name. It runs when -n is given, or when no references and no --env-file are
used (then the default namespace is injected).

Precedence (later wins): inherited environment, namespace injection,
--env-file in the order given. References are resolved after merging.
If any reference cannot be resolved, the command is not started.

Masking: secrets read from the keyring are replaced with
<concealed by cred> in the command's stdout and stderr. Secrets shorter than
8 bytes are not masked. The command's stdout and stderr are then pipes, not
a terminal, so some commands drop colors or interactive output; output that
does not end with a newline may be held back while it could be the start of
a secret. --no-masking turns masking off and runs the command as before.
Masking prevents accidental disclosure; it is not a security boundary.

All cred messages go to stderr; stdout belongs to the command.

Examples:
  cred run -- python app.py                              # Inject default namespace
  cred run -n production -- ./deploy.sh                  # Inject a namespace
  API_TOKEN=cred://work/JIRA_TOKEN cred run -- ./sync.sh # Resolve a reference
  cred run --env-file .env.cred -- python app.py         # References in a file
  cred run -- env | grep OPENAI                          # Verify injection (value concealed)
  cred run --no-masking -- ./interactive-tool            # Keep the terminal
  cred run -- sh -c 'echo $API_KEY && python app.py'     # Multiple commands`,
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		ns := cfg.ActiveNamespace(namespace)
		var keys []string
		if nsConfig, ok := cfg.Namespaces[ns]; ok {
			// Sort keys for deterministic injection order (fixes #7)
			keys = sortedKeys(nsConfig.Keys)
		}

		runEnv, err := buildRunEnv(runEnvInput{
			Environ:           os.Environ(),
			Namespace:         ns,
			NamespaceKeys:     keys,
			NamespaceExplicit: namespace != "",
			EnvFiles:          envFiles,
		}, os.Stderr)
		if err != nil {
			return err
		}

		// Find the command binary
		binary, err := exec.LookPath(args[0])
		if err != nil {
			return fmt.Errorf("command not found: %s", args[0])
		}

		if !noMasking && len(mask.Values(runEnv.Secrets)) > 0 {
			code, err := runMasked(binary, args, runEnv.Environ, runEnv.Secrets,
				childIO{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr})
			if err != nil {
				return err
			}
			if code != 0 {
				return &ExitCodeError{Code: code}
			}
			return nil
		}

		// Nothing to mask: replace current process with the command (exec),
		// so it keeps the terminal, and the exit code and signals belong to
		// it without a supervising parent.
		return syscall.Exec(binary, args, runEnv.Environ)
	},
}

// runEnvInput holds everything buildRunEnv needs to assemble the environment.
type runEnvInput struct {
	Environ           []string // inherited environment
	Namespace         string   // namespace for bulk injection
	NamespaceKeys     []string // sorted keys registered in Namespace
	NamespaceExplicit bool     // -n was given
	EnvFiles          []string // --env-file paths in order
}

// runEnv is the environment for the command and the secrets it contains.
type runEnv struct {
	Environ []string // KEY=VALUE entries for the command
	Secrets []string // every value read from the keyring, unfiltered
}

// buildRunEnv merges the inherited environment, namespace injection and
// env files (later wins), then resolves cred:// references once.
// Messages go to stderr; secret values are never written.
func buildRunEnv(in runEnvInput, stderr io.Writer) (runEnv, error) {
	var fileEntries []env.Entry
	for _, path := range in.EnvFiles {
		entries, err := env.ParseFile(path)
		if err != nil {
			return runEnv{}, fmt.Errorf("--env-file %s: %w", path, err)
		}
		fileEntries = append(fileEntries, entries...)
	}

	environ := append([]string(nil), in.Environ...)

	// Without -n, references or env files mean the caller chose exactly what
	// to pass, so the default namespace is not injected wholesale.
	injectNS := in.NamespaceExplicit || (!hasReference(environ) && len(in.EnvFiles) == 0)
	var nsSecrets []string
	if injectNS {
		var err error
		environ, nsSecrets, err = injectNamespace(environ, in.Namespace, in.NamespaceKeys, stderr)
		if err != nil {
			return runEnv{}, err
		}
	}

	for _, e := range fileEntries {
		environ = setEnv(environ, e.Key, e.Value)
	}

	environ, refSecrets, err := resolveReferences(environ, stderr)
	if err != nil {
		return runEnv{}, err
	}

	// Report only after everything succeeded, so a run that does not start
	// the command never claims that credentials were injected.
	if injectNS {
		fmt.Fprintf(stderr, "→ Injected %d credential(s) from namespace %q\n", len(nsSecrets), in.Namespace)
	}
	if len(refSecrets) > 0 {
		fmt.Fprintf(stderr, "→ Resolved %d reference(s)\n", len(refSecrets))
	}
	return runEnv{Environ: environ, Secrets: append(nsSecrets, refSecrets...)}, nil
}

func hasReference(environ []string) bool {
	for _, entry := range environ {
		if _, value, ok := strings.Cut(entry, "="); ok && ref.IsRef(value) {
			return true
		}
	}
	return false
}

// injectNamespace sets every key of ns and returns the values it read.
func injectNamespace(environ []string, ns string, keys []string, stderr io.Writer) ([]string, []string, error) {
	if len(keys) == 0 {
		fmt.Fprintf(stderr, "⚠ No credentials in namespace %q, running without injection\n", ns)
	}

	service := keyring.ServiceName(ns)
	var secrets []string
	for _, key := range keys {
		value, err := getSecret(service, key)
		if err != nil {
			if errors.Is(err, keyring.ErrNotFound) {
				fmt.Fprintf(stderr, "⚠ Skipping %s: not found in keyring\n", key)
				continue
			}
			return nil, nil, fmt.Errorf("failed to retrieve %s: %w", key, err)
		}
		environ = setEnv(environ, key, value)
		secrets = append(secrets, value)
	}
	return environ, secrets, nil
}

// resolveReferences replaces every cred:// value with its secret and returns
// one secret per variable resolved. Resolved values are not resolved again.
// Each distinct reference is looked up once, including failed ones, and all
// failures are reported before returning.
func resolveReferences(environ []string, stderr io.Writer) ([]string, []string, error) {
	type lookup struct {
		secret string
		err    error
	}
	cache := map[string]lookup{}
	var secrets []string
	failed := 0

	for i, entry := range environ {
		name, value, ok := strings.Cut(entry, "=")
		if !ok || !ref.IsRef(value) {
			continue
		}

		res, cached := cache[value]
		if !cached {
			r, err := ref.Parse(value)
			if err == nil {
				res.secret, err = getSecret(keyring.ServiceName(r.Namespace), r.Key)
			}
			res.err = err
			cache[value] = res
		}
		if res.err != nil {
			fmt.Fprintf(stderr, "✗ Cannot resolve %s (%s): %s\n", name, value, describeResolveError(res.err))
			failed++
			continue
		}

		environ[i] = name + "=" + res.secret
		secrets = append(secrets, res.secret)
	}

	if failed > 0 {
		return nil, nil, fmt.Errorf("failed to resolve %d reference(s); command not started", failed)
	}
	return environ, secrets, nil
}

func describeResolveError(err error) string {
	switch {
	case errors.Is(err, ref.ErrInvalid):
		return err.Error()
	case errors.Is(err, keyring.ErrNotFound):
		return "not found in keyring"
	case errors.Is(err, keyring.ErrTimeout):
		return "keyring timed out (access may be waiting for approval or denied)"
	default:
		return fmt.Sprintf("keyring error: %v", err)
	}
}

func init() {
	runCmd.Flags().BoolVar(&noMasking, "no-masking", false, "do not conceal secrets in the command's stdout and stderr")
	runCmd.Flags().StringArrayVar(&envFiles, "env-file", nil, "load KEY=VALUE lines from a dotenv file (repeatable; later files win)")
	rootCmd.AddCommand(runCmd)
}
