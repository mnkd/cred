package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/mnkd/cred/internal/config"
	"github.com/mnkd/cred/internal/keyring"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check cred's environment and detect issues",
	Long: `Run diagnostic checks on cred's environment.

Checks:
  - OS keyring accessibility (read/write)
  - Config file existence and validity
  - Consistency between config metadata and keyring entries

Examples:
  cred doctor`,
	RunE: func(cmd *cobra.Command, args []string) error {
		hasError := false

		// 1. キーリングの可用性チェック
		fmt.Fprintln(os.Stderr, "Checking keyring...")
		testService := keyring.ServiceName("__doctor_test__")
		testKey := "__doctor_probe__"
		testValue := "ok"

		if err := keyring.Set(testService, testKey, testValue); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ Keyring: write failed: %v\n", err)
			hasError = true
		} else {
			val, err := keyring.Get(testService, testKey)
			if err != nil || val != testValue {
				fmt.Fprintf(os.Stderr, "  ✗ Keyring: read failed\n")
				hasError = true
			} else {
				fmt.Fprintln(os.Stderr, "  ✓ Keyring: accessible")
			}
			// テストエントリを削除
			_ = keyring.Delete(testService, testKey)
		}

		// 2. 設定ファイルのチェック
		fmt.Fprintln(os.Stderr, "Checking config...")
		configPath, err := config.ConfigPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ Config path: %v\n", err)
			hasError = true
		} else {
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "  - Config: %s (not yet created)\n", configPath)
			} else {
				fmt.Fprintf(os.Stderr, "  ✓ Config: %s\n", configPath)
			}
		}

		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ Config load: %v\n", err)
			hasError = true
		} else {
			fmt.Fprintf(os.Stderr, "  ✓ Config: version %d, %d namespace(s)\n", cfg.Version, len(cfg.Namespaces))

			// 3. config と keyring の整合性チェック
			fmt.Fprintln(os.Stderr, "Checking consistency...")
			for nsName, nsConfig := range cfg.Namespaces {
				service := keyring.ServiceName(nsName)
				for key := range nsConfig.Keys {
					_, err := keyring.Get(service, key)
					if err != nil {
						if errors.Is(err, keyring.ErrNotFound) {
							fmt.Fprintf(os.Stderr, "  ⚠ %s/%s: in config but not in keyring\n", nsName, key)
							hasError = true
						} else {
							fmt.Fprintf(os.Stderr, "  ✗ %s/%s: keyring error: %v\n", nsName, key, err)
							hasError = true
						}
					} else {
						fmt.Fprintf(os.Stderr, "  ✓ %s/%s\n", nsName, key)
					}
				}
			}
		}

		if hasError {
			fmt.Fprintln(os.Stderr, "\nIssues found. See above for details.")
		} else {
			fmt.Fprintln(os.Stderr, "\nAll checks passed.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
