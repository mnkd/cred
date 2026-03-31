package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/mnkd/cred/internal/config"
	"github.com/mnkd/cred/internal/keyring"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete KEY",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a credential from the OS keyring",
	Long: `Delete a secret from the OS keyring and remove its metadata from config.

Examples:
  cred delete OPENAI_API_KEY              # Delete from default namespace
  cred rm OPENAI_API_KEY                  # Alias: rm
  cred remove OPENAI_API_KEY              # Alias: remove
  cred delete AWS_SECRET_KEY -n work      # Delete from specific namespace`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		ns := cfg.ActiveNamespace(namespace)
		service := keyring.ServiceName(ns)

		// Delete from keyring
		if err := keyring.Delete(service, key); err != nil {
			if errors.Is(err, keyring.ErrNotFound) {
				return fmt.Errorf("credential %q not found in namespace %q", key, ns)
			}
			return fmt.Errorf("failed to delete from keyring: %w", err)
		}

		// Remove metadata from config
		cfg.DeleteKeyMeta(ns, key)
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Deleted %s from namespace %q\n", key, ns)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
