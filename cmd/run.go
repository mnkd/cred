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
	"github.com/mnkd/cred/internal/ref"
	"github.com/spf13/cobra"
)

var envFiles []string

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

All cred messages go to stderr; stdout belongs to the command.

Examples:
  cred run -- python app.py                              # Inject default namespace
  cred run -n production -- ./deploy.sh                  # Inject a namespace
  API_TOKEN=cred://work/JIRA_TOKEN cred run -- ./sync.sh # Resolve a reference
  cred run --env-file .env.cred -- python app.py         # References in a file
  cred run -- env | grep OPENAI                          # Verify injection
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

		environ, err := buildRunEnv(runEnvInput{
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

		// Replace current process with the command (exec), so the exit code
		// and signals belong to the command without a supervising parent.
		return syscall.Exec(binary, args, environ)
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

// buildRunEnv merges the inherited environment, namespace injection and
// env files (later wins), then resolves cred:// references once.
// Messages go to stderr; secret values are never written.
func buildRunEnv(in runEnvInput, stderr io.Writer) ([]string, error) {
	var fileEntries []env.Entry
	for _, path := range in.EnvFiles {
		entries, err := env.ParseFile(path)
		if err != nil {
			return nil, fmt.Errorf("--env-file %s: %w", path, err)
		}
		fileEntries = append(fileEntries, entries...)
	}

	environ := append([]string(nil), in.Environ...)

	// Without -n, references or env files mean the caller chose exactly what
	// to pass, so the default namespace is not injected wholesale.
	injectNS := in.NamespaceExplicit || (!hasReference(environ) && len(in.EnvFiles) == 0)
	injected := 0
	if injectNS {
		var err error
		environ, injected, err = injectNamespace(environ, in.Namespace, in.NamespaceKeys, stderr)
		if err != nil {
			return nil, err
		}
	}

	for _, e := range fileEntries {
		environ = setEnv(environ, e.Key, e.Value)
	}

	environ, resolved, err := resolveReferences(environ, stderr)
	if err != nil {
		return nil, err
	}

	// Report only after everything succeeded, so a run that does not start
	// the command never claims that credentials were injected.
	if injectNS {
		fmt.Fprintf(stderr, "→ Injected %d credential(s) from namespace %q\n", injected, in.Namespace)
	}
	if resolved > 0 {
		fmt.Fprintf(stderr, "→ Resolved %d reference(s)\n", resolved)
	}
	return environ, nil
}

func hasReference(environ []string) bool {
	for _, entry := range environ {
		if _, value, ok := strings.Cut(entry, "="); ok && ref.IsRef(value) {
			return true
		}
	}
	return false
}

func injectNamespace(environ []string, ns string, keys []string, stderr io.Writer) ([]string, int, error) {
	if len(keys) == 0 {
		fmt.Fprintf(stderr, "⚠ No credentials in namespace %q, running without injection\n", ns)
	}

	service := keyring.ServiceName(ns)
	injected := 0
	for _, key := range keys {
		value, err := getSecret(service, key)
		if err != nil {
			if errors.Is(err, keyring.ErrNotFound) {
				fmt.Fprintf(stderr, "⚠ Skipping %s: not found in keyring\n", key)
				continue
			}
			return nil, 0, fmt.Errorf("failed to retrieve %s: %w", key, err)
		}
		environ = setEnv(environ, key, value)
		injected++
	}
	return environ, injected, nil
}

// resolveReferences replaces every cred:// value with its secret and returns
// the number of variables resolved. Resolved values are not resolved again.
// Each distinct reference is looked up once, including failed ones, and all
// failures are reported before returning.
func resolveReferences(environ []string, stderr io.Writer) ([]string, int, error) {
	type lookup struct {
		secret string
		err    error
	}
	cache := map[string]lookup{}
	resolved, failed := 0, 0

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
		resolved++
	}

	if failed > 0 {
		return nil, 0, fmt.Errorf("failed to resolve %d reference(s); command not started", failed)
	}
	return environ, resolved, nil
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
	runCmd.Flags().StringArrayVar(&envFiles, "env-file", nil, "load KEY=VALUE lines from a dotenv file (repeatable; later files win)")
	rootCmd.AddCommand(runCmd)
}
