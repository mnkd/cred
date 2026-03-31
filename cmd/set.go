package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mitsuru/cred/internal/config"
	"github.com/mitsuru/cred/internal/keyring"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var setDescription string

var setCmd = &cobra.Command{
	Use:   "set KEY [VALUE]",
	Short: "Store a credential in the OS keyring",
	Long: `Store a secret value in the OS keyring.

If VALUE is omitted, you will be prompted to enter it.
If VALUE is "-", it will be read from stdin (useful for piping).

Examples:
  cred set OPENAI_API_KEY                           # Prompted input
  cred set OPENAI_API_KEY sk-xxx                    # Direct value
  cred set OPENAI_API_KEY sk-xxx -d "OpenAI prod"  # With description
  echo "sk-xxx" | cred set OPENAI_API_KEY -         # From stdin
  cat token.txt | cred set GITHUB_TOKEN -           # From file via pipe
  cred set ANTHROPIC_API_KEY -n work                # Store in namespace`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		var value string
		if len(args) == 2 {
			if args[1] == "-" {
				// Read from stdin
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					value = strings.TrimSpace(scanner.Text())
				}
				if err := scanner.Err(); err != nil {
					return fmt.Errorf("failed to read from stdin: %w", err)
				}
			} else {
				value = args[1]
			}
		} else {
			// セキュアプロンプト: 入力値を画面に表示しない
			fmt.Fprintf(os.Stderr, "Enter value for %s: ", key)
			bytesValue, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Fprintln(os.Stderr)
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			value = strings.TrimSpace(string(bytesValue))
		}

		if value == "" {
			return fmt.Errorf("value cannot be empty")
		}

		// Load config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		ns := cfg.ActiveNamespace(namespace)
		service := keyring.ServiceName(ns)

		// Store in keyring
		if err := keyring.Set(service, key, value); err != nil {
			return fmt.Errorf("failed to store in keyring: %w", err)
		}

		// Update config metadata
		now := time.Now().UTC().Format(time.RFC3339)
		meta := config.KeyMeta{
			Description: setDescription,
			EnvVar:      key, // default env var name = key name
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		// Preserve created_at if updating existing key
		if existing, ok := cfg.Namespaces[ns]; ok {
			if existingMeta, ok := existing.Keys[key]; ok {
				meta.CreatedAt = existingMeta.CreatedAt
				if setDescription == "" {
					meta.Description = existingMeta.Description
				}
			}
		}

		cfg.SetKeyMeta(ns, key, meta)
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Stored %s in namespace %q\n", key, ns)
		return nil
	},
}

func init() {
	setCmd.Flags().StringVarP(&setDescription, "description", "d", "", "description of this credential")
	rootCmd.AddCommand(setCmd)
}
