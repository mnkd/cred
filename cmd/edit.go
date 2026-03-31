package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mnkd/cred/internal/config"
	"github.com/mnkd/cred/internal/keyring"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit KEY",
	Short: "Edit a credential value using $EDITOR",
	Long: `Open the current value of a credential in your editor.
After saving and closing, the new value is stored in the keyring.

Uses $EDITOR (falls back to vi).

Examples:
  cred edit OPENAI_API_KEY
  EDITOR=code cred edit CONFIG_JSON
  cred edit DB_PASSWORD -n production`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		ns := cfg.ActiveNamespace(namespace)
		service := keyring.ServiceName(ns)

		// 現在の値を取得
		current, err := keyring.Get(service, key)
		if err != nil {
			if errors.Is(err, keyring.ErrNotFound) {
				return fmt.Errorf("credential %q not found in namespace %q", key, ns)
			}
			return fmt.Errorf("failed to retrieve from keyring: %w", err)
		}

		// 一時ファイルに書き出し
		tmpFile, err := os.CreateTemp("", "cred-edit-*")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		if _, err := tmpFile.WriteString(current); err != nil {
			tmpFile.Close()
			return fmt.Errorf("failed to write temp file: %w", err)
		}
		tmpFile.Close()

		// エディタで開く
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
		editorCmd := exec.Command(editor, tmpPath)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr
		if err := editorCmd.Run(); err != nil {
			return fmt.Errorf("editor exited with error: %w", err)
		}

		// 編集後の値を読み込み
		data, err := os.ReadFile(tmpPath)
		if err != nil {
			return fmt.Errorf("failed to read edited file: %w", err)
		}
		newValue := strings.TrimRight(string(data), "\n")

		if newValue == current {
			fmt.Fprintln(os.Stderr, "No changes made")
			return nil
		}

		if newValue == "" {
			return fmt.Errorf("value cannot be empty")
		}

		// キーリングに保存
		if err := keyring.Set(service, key, newValue); err != nil {
			return fmt.Errorf("failed to store in keyring: %w", err)
		}

		// メタデータの updated_at を更新
		if nsConfig, ok := cfg.Namespaces[ns]; ok {
			if meta, ok := nsConfig.Keys[key]; ok {
				meta.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
				cfg.SetKeyMeta(ns, key, meta)
				if err := cfg.Save(); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}
			}
		}

		fmt.Fprintf(os.Stderr, "✓ Updated %s in namespace %q\n", key, ns)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
