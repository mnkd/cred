package cmd

import (
	"github.com/spf13/cobra"
)

var namespace string

var rootCmd = &cobra.Command{
	Use:   "cred",
	Short: "A credential manager for local development tools",
	Long: `cred stores and retrieves API keys and secrets using your OS keyring
(macOS Keychain, Linux Secret Service, Windows Credential Manager).

No more scattered .env files. No more expired sessions.

Quick start:
  cred set OPENAI_API_KEY              # Store a secret (prompted)
  cred get OPENAI_API_KEY              # Retrieve a secret
  cred list                            # List stored keys
  cred run -- python app.py            # Run with secrets as env vars
  cred import .env                     # Import from .env file
  cred ns create work -d "Work keys"    # Create a namespace
  cred set API_KEY -n work             # Store in specific namespace`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace to use (default: from config)")
}

func Execute() error {
	return rootCmd.Execute()
}
