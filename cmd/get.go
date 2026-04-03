package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/mnkd/cred/internal/config"
	"github.com/mnkd/cred/internal/keyring"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get KEY",
	Short: "Retrieve a credential from the OS keyring",
	Long: `Retrieve a secret value from the OS keyring and print it to stdout.

The value is printed without a trailing newline, making it safe for
command substitution and piping.

Examples:
  cred get OPENAI_API_KEY                                    # Print to stdout
  cred get DB_PASSWORD -n production                         # From namespace
  export OPENAI_API_KEY=$(cred get OPENAI_API_KEY)           # Shell export
  cred get OPENAI_API_KEY | pbcopy                           # Copy to clipboard (macOS)
  cred get OPENAI_API_KEY | xclip                            # Copy to clipboard (Linux)
  curl -H "Authorization: Bearer $(cred get TOKEN)" URL      # Use in curl`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		ns := cfg.ActiveNamespace(namespace)
		service := keyring.ServiceName(ns)

		value, err := keyring.Get(service, key)
		if err != nil {
			if errors.Is(err, keyring.ErrNotFound) {
				return fmt.Errorf("credential %q not found in namespace %q", key, ns)
			}
			return fmt.Errorf("failed to retrieve from keyring: %w", err)
		}

		if isJSON() {
			out := map[string]string{"key": key, "value": value}
			return json.NewEncoder(os.Stdout).Encode(out)
		}

		// Text output: no trailing newline (safe for $() substitution)
		fmt.Fprint(os.Stdout, value)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
