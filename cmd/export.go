package cmd

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/mnkd/cred/internal/config"
	"github.com/mnkd/cred/internal/keyring"
	"github.com/spf13/cobra"
)

var exportKeys string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export credentials as .env format to stdout",
	Long: `Export all credentials from the current namespace in .env format.
This is the reverse of "cred import".

Examples:
  cred export                                        # Export all keys
  cred export > .env.local                           # Save to file
  cred export -n production                          # From namespace
  cred export --keys OPENAI_API_KEY,DB_PASSWORD      # Specific keys only`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		ns := cfg.ActiveNamespace(namespace)
		nsConfig, ok := cfg.Namespaces[ns]
		if !ok || len(nsConfig.Keys) == 0 {
			return fmt.Errorf("no credentials in namespace %q", ns)
		}

		// Filter by --keys flag
		var filter map[string]bool
		if exportKeys != "" {
			filter = make(map[string]bool)
			for _, k := range strings.Split(exportKeys, ",") {
				filter[strings.TrimSpace(k)] = true
			}
		}

		service := keyring.ServiceName(ns)
		keys := make([]string, 0, len(nsConfig.Keys))
		for k := range nsConfig.Keys {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		exported := 0
		for _, key := range keys {
			if filter != nil && !filter[key] {
				continue
			}

			value, err := keyring.Get(service, key)
			if err != nil {
				if errors.Is(err, keyring.ErrNotFound) {
					fmt.Fprintf(os.Stderr, "⚠ Skipping %s: not found in keyring\n", key)
					continue
				}
				return fmt.Errorf("failed to retrieve %s: %w", key, err)
			}

			meta := nsConfig.Keys[key]
			envVar := meta.EnvVar
			if envVar == "" {
				envVar = key
			}
			fmt.Fprintf(os.Stdout, "%s=%s\n", envVar, value)
			exported++
		}

		fmt.Fprintf(os.Stderr, "→ Exported %d credential(s) from namespace %q\n", exported, ns)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVar(&exportKeys, "keys", "", "comma-separated list of keys to export")
	rootCmd.AddCommand(exportCmd)
}
