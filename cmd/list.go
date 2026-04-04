package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/mnkd/cred/internal/config"
	"github.com/spf13/cobra"
)

const staleThreshold = 90 * 24 * time.Hour

var listAll bool

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List stored credential keys",
	Long: `List all credential keys registered in the current namespace.
Only key names and metadata are shown — secret values are never displayed.

Examples:
  cred list                  # List keys in default namespace
  cred ls                    # Alias: ls
  cred list -n production    # List keys in specific namespace
  cred list --all            # List keys across all namespaces
  cred ls -a                 # Short form`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)

		if listAll {
			fmt.Fprintln(w, "NAMESPACE\tKEY\tDESCRIPTION")
			namespaces := sortedKeys(cfg.Namespaces)
			for _, nsName := range namespaces {
				ns := cfg.Namespaces[nsName]
				keys := sortedKeyMetaKeys(ns.Keys)
				for _, key := range keys {
					meta := ns.Keys[key]
					fmt.Fprintf(w, "%s\t%s\t%s\n", nsName, key, meta.Description)
				}
			}
		} else {
			nsName := cfg.ActiveNamespace(namespace)
			ns, ok := cfg.Namespaces[nsName]
			if !ok || len(ns.Keys) == 0 {
				fmt.Fprintf(os.Stderr, "No credentials in namespace %q\n", nsName)
				return nil
			}

			fmt.Fprintf(os.Stderr, "Namespace: %s\n\n", nsName)
			fmt.Fprintln(w, "KEY\tDESCRIPTION\tUPDATED")
			keys := sortedKeyMetaKeys(ns.Keys)
			for _, key := range keys {
				meta := ns.Keys[key]
				warning := staleWarning(meta.UpdatedAt)
				fmt.Fprintf(w, "%s\t%s\t%s%s\n", key, meta.Description, meta.UpdatedAt, warning)
			}
		}

		return w.Flush()
	},
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedKeyMetaKeys(m map[string]config.KeyMeta) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// staleWarning returns a warning string if the updated_at timestamp exceeds the stale threshold.
func staleWarning(updatedAt string) string {
	if updatedAt == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return ""
	}
	age := time.Since(t)
	if age > staleThreshold {
		days := int(age.Hours() / 24)
		return fmt.Sprintf(" ⚠ %d days since last update", days)
	}
	return ""
}

func init() {
	listCmd.Flags().BoolVarP(&listAll, "all", "a", false, "list keys across all namespaces")
	rootCmd.AddCommand(listCmd)
}
