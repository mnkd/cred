package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mitsuru/cred/internal/config"
	"github.com/mitsuru/cred/internal/keyring"
	"github.com/spf13/cobra"
)

var nsCmd = &cobra.Command{
	Use:     "namespace",
	Aliases: []string{"ns"},
	Short:   "Manage namespaces",
}

var nsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all namespaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		active := cfg.ActiveNamespace("")

		if isJSON() {
			type nsEntry struct {
				Name        string `json:"name"`
				Default     bool   `json:"default"`
				KeyCount    int    `json:"key_count"`
				Description string `json:"description,omitempty"`
			}
			var entries []nsEntry
			for name, ns := range cfg.Namespaces {
				entries = append(entries, nsEntry{
					Name:        name,
					Default:     name == active,
					KeyCount:    len(ns.Keys),
					Description: ns.Description,
				})
			}
			return json.NewEncoder(os.Stdout).Encode(entries)
		}

		for name, ns := range cfg.Namespaces {
			marker := "  "
			if name == active {
				marker = "* "
			}
			desc := ""
			if ns.Description != "" {
				desc = " - " + ns.Description
			}
			fmt.Fprintf(os.Stdout, "%s%s (%d keys)%s\n", marker, name, len(ns.Keys), desc)
		}
		return nil
	},
}

var nsCreateCmd = &cobra.Command{
	Use:   "create NAME",
	Short: "Create a new namespace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if _, exists := cfg.Namespaces[name]; exists {
			return fmt.Errorf("namespace %q already exists", name)
		}
		cfg.EnsureNamespace(name)
		desc, _ := cmd.Flags().GetString("description")
		if desc != "" {
			ns := cfg.Namespaces[name]
			ns.Description = desc
			cfg.Namespaces[name] = ns
		}
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "✓ Created namespace %q\n", name)
		return nil
	},
}

var nsDefaultCmd = &cobra.Command{
	Use:   "default [NAME]",
	Short: "Get or set the default namespace",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if len(args) == 0 {
			fmt.Println(cfg.ActiveNamespace(""))
			return nil
		}
		name := args[0]
		if _, exists := cfg.Namespaces[name]; !exists {
			return fmt.Errorf("namespace %q does not exist", name)
		}
		cfg.Default = name
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "✓ Default namespace set to %q\n", name)
		return nil
	},
}

var nsForce bool

var nsDeleteCmd = &cobra.Command{
	Use:   "delete NAME",
	Short: "Delete a namespace and all its credentials",
	Long: `Delete a namespace, removing all its credentials from the keyring
and metadata from config.

Examples:
  cred ns delete work
  cred ns delete staging --force    # Skip confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// デフォルト namespace の削除を禁止
		if name == cfg.ActiveNamespace("") {
			return fmt.Errorf("cannot delete the default namespace %q", name)
		}

		nsConfig, exists := cfg.Namespaces[name]
		if !exists {
			return fmt.Errorf("namespace %q does not exist", name)
		}

		// 確認プロンプト
		if !nsForce {
			fmt.Fprintf(os.Stderr, "Delete namespace %q and its %d credential(s)? [y/N] ", name, len(nsConfig.Keys))
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
				if answer != "y" && answer != "yes" {
					fmt.Fprintln(os.Stderr, "Cancelled")
					return nil
				}
			}
		}

		// キーリングからエントリを削除
		service := keyring.ServiceName(name)
		for key := range nsConfig.Keys {
			if err := keyring.Delete(service, key); err != nil {
				fmt.Fprintf(os.Stderr, "⚠ Failed to delete %s from keyring: %v\n", key, err)
			}
		}

		// 設定ファイルから namespace を削除
		delete(cfg.Namespaces, name)
		if err := cfg.Save(); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "✓ Deleted namespace %q\n", name)
		return nil
	},
}

func init() {
	nsCreateCmd.Flags().StringP("description", "d", "", "namespace description")
	nsDeleteCmd.Flags().BoolVar(&nsForce, "force", false, "skip confirmation prompt")
	nsCmd.AddCommand(nsListCmd)
	nsCmd.AddCommand(nsCreateCmd)
	nsCmd.AddCommand(nsDefaultCmd)
	nsCmd.AddCommand(nsDeleteCmd)
	rootCmd.AddCommand(nsCmd)
}
