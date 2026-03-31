package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/mitsuru/cred/internal/config"
	"github.com/mitsuru/cred/internal/env"
	"github.com/mitsuru/cred/internal/keyring"
	"github.com/spf13/cobra"
)

var (
	importOverwrite bool
	importDryRun    bool
)

var importCmd = &cobra.Command{
	Use:   "import FILE",
	Short: "Import credentials from a .env file into the OS keyring",
	Long: `Parse a .env file and store each key-value pair in the OS keyring.

Supports standard .env format:
  KEY=VALUE
  KEY="quoted value"
  export KEY=VALUE
  # comments

Examples:
  cred import .env                         # Import to default namespace
  cred import .env.production -n production # Import to specific namespace
  cred import .env --dry-run               # Preview without storing
  cred import .env --overwrite             # Overwrite existing keys
  cred import .env --overwrite --dry-run   # Preview overwrite`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		entries, err := env.ParseFile(filePath)
		if err != nil {
			return err
		}

		if len(entries) == 0 {
			fmt.Fprintln(os.Stderr, "No entries found in file")
			return nil
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		ns := cfg.ActiveNamespace(namespace)
		service := keyring.ServiceName(ns)
		now := time.Now().UTC().Format(time.RFC3339)

		imported := 0
		skipped := 0

		for _, entry := range entries {
			// Check if key already exists
			if !importOverwrite {
				if nsConfig, ok := cfg.Namespaces[ns]; ok {
					if _, exists := nsConfig.Keys[entry.Key]; exists {
						fmt.Fprintf(os.Stderr, "  skip: %s (already exists, use --overwrite)\n", entry.Key)
						skipped++
						continue
					}
				}
			}

			if importDryRun {
				fmt.Fprintf(os.Stderr, "  would import: %s\n", entry.Key)
				imported++
				continue
			}

			// Store in keyring
			if err := keyring.Set(service, entry.Key, entry.Value); err != nil {
				fmt.Fprintf(os.Stderr, "  error: %s: %v\n", entry.Key, err)
				continue
			}

			// Update config metadata
			meta := config.KeyMeta{
				EnvVar:    entry.Key,
				CreatedAt: now,
				UpdatedAt: now,
			}
			cfg.SetKeyMeta(ns, entry.Key, meta)
			imported++
			fmt.Fprintf(os.Stderr, "  ✓ %s\n", entry.Key)
		}

		if !importDryRun {
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
		}

		action := "Imported"
		if importDryRun {
			action = "Would import"
		}
		fmt.Fprintf(os.Stderr, "\n%s %d credential(s) to namespace %q", action, imported, ns)
		if skipped > 0 {
			fmt.Fprintf(os.Stderr, " (%d skipped)", skipped)
		}
		fmt.Fprintln(os.Stderr)

		return nil
	},
}

func init() {
	importCmd.Flags().BoolVar(&importOverwrite, "overwrite", false, "overwrite existing credentials")
	importCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "preview import without storing")
	rootCmd.AddCommand(importCmd)
}
