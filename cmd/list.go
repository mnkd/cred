package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/mitsuru/cred/internal/config"
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

		if isJSON() {
			return listJSON(cfg)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)

		if listAll {
			fmt.Fprintln(w, "NAMESPACE\tKEY\tENV_VAR\tDESCRIPTION")
			namespaces := sortedKeys(cfg.Namespaces)
			for _, nsName := range namespaces {
				ns := cfg.Namespaces[nsName]
				keys := sortedKeyMetaKeys(ns.Keys)
				for _, key := range keys {
					meta := ns.Keys[key]
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", nsName, key, meta.EnvVar, meta.Description)
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
			fmt.Fprintln(w, "KEY\tENV_VAR\tDESCRIPTION\tUPDATED")
			keys := sortedKeyMetaKeys(ns.Keys)
			for _, key := range keys {
				meta := ns.Keys[key]
				warning := staleWarning(meta.UpdatedAt)
				fmt.Fprintf(w, "%s\t%s\t%s\t%s%s\n", key, meta.EnvVar, meta.Description, meta.UpdatedAt, warning)
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

type listEntry struct {
	Namespace   string `json:"namespace,omitempty"`
	Key         string `json:"key"`
	EnvVar      string `json:"env_var"`
	Description string `json:"description,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

func listJSON(cfg *config.Config) error {
	var entries []listEntry

	if listAll {
		for nsName, ns := range cfg.Namespaces {
			for key, meta := range ns.Keys {
				entries = append(entries, listEntry{
					Namespace:   nsName,
					Key:         key,
					EnvVar:      meta.EnvVar,
					Description: meta.Description,
					UpdatedAt:   meta.UpdatedAt,
				})
			}
		}
	} else {
		nsName := cfg.ActiveNamespace(namespace)
		ns, ok := cfg.Namespaces[nsName]
		if !ok || len(ns.Keys) == 0 {
			return json.NewEncoder(os.Stdout).Encode([]listEntry{})
		}
		for key, meta := range ns.Keys {
			entries = append(entries, listEntry{
				Key:         key,
				EnvVar:      meta.EnvVar,
				Description: meta.Description,
				UpdatedAt:   meta.UpdatedAt,
			})
		}
	}

	return json.NewEncoder(os.Stdout).Encode(entries)
}

// staleWarning は更新日時が閾値を超えている場合に警告文字列を返す。
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
		return fmt.Sprintf(" ⚠ %d日未更新", days)
	}
	return ""
}

func init() {
	listCmd.Flags().BoolVarP(&listAll, "all", "a", false, "list keys across all namespaces")
	rootCmd.AddCommand(listCmd)
}
