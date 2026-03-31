//go:build darwin

package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/mnkd/cred/internal/config"
	"github.com/mnkd/cred/internal/keyring"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run -- COMMAND [ARGS...]",
	Short: "Run a command with credentials injected as environment variables",
	Long: `Execute a command with all credentials from the current namespace
injected as environment variables.

The env var name for each credential is determined by the "env_var"
field in config (defaults to the key name).

Examples:
  cred run -- python app.py                              # Python
  cred run -- node server.js                             # Node.js
  cred run -n production -- ./deploy.sh                  # From namespace
  cred run -- env | grep OPENAI                          # Verify injection
  cred run -- docker run --env OPENAI_API_KEY app:latest # Docker
  cred run -- sh -c 'echo $API_KEY && python app.py'     # Multiple commands`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: false,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		ns := cfg.ActiveNamespace(namespace)
		nsConfig, ok := cfg.Namespaces[ns]
		if !ok || len(nsConfig.Keys) == 0 {
			fmt.Fprintf(os.Stderr, "⚠ No credentials in namespace %q, running without injection\n", ns)
		}

		// Build environment: current env + credentials
		environ := os.Environ()
		injected := 0

		if ok {
			service := keyring.ServiceName(ns)
			for key, meta := range nsConfig.Keys {
				value, err := keyring.Get(service, key)
				if err != nil {
					if errors.Is(err, keyring.ErrNotFound) {
						fmt.Fprintf(os.Stderr, "⚠ Skipping %s: not found in keyring\n", key)
						continue
					}
					return fmt.Errorf("failed to retrieve %s: %w", key, err)
				}

				envVar := meta.EnvVar
				if envVar == "" {
					envVar = key
				}
				environ = append(environ, envVar+"="+value)
				injected++
			}
		}

		fmt.Fprintf(os.Stderr, "→ Injected %d credential(s) from namespace %q\n", injected, ns)

		// Find the command binary
		binary, err := exec.LookPath(args[0])
		if err != nil {
			return fmt.Errorf("command not found: %s", args[0])
		}

		// Replace current process with the command (exec)
		return syscall.Exec(binary, args, environ)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
