package cmd

import (
	"fmt"
	"os"

	"github.com/mitsuru/cred/internal/config"
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

func init() {
	nsCreateCmd.Flags().StringP("description", "d", "", "namespace description")
	nsCmd.AddCommand(nsListCmd)
	nsCmd.AddCommand(nsCreateCmd)
	nsCmd.AddCommand(nsDefaultCmd)
	rootCmd.AddCommand(nsCmd)
}
