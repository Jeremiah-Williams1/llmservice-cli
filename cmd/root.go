package cmd

import (
	"github.com/spf13/cobra"
)

// TODO: persistent flags shared by all subcommands.
// At minimum: --kubeconfig, --namespace. Maybe --context later.
var (
	kubeconfigPath string
	namespace      string
)

var rootCmd = &cobra.Command{
	Use:   "llmservice-cli",
	Short: "Deploy, inspect, and roll back LLMService resources",
	// TODO: decide whether the dynamic client gets built once here
	// (e.g. in PersistentPreRunE) and passed down, or built lazily
	// per-subcommand. Once here avoids reconnecting per command.
}

func init() {
	// TODO: register persistent flags on rootCmd here.
	// rootCmd.PersistentFlags().StringVar(&kubeconfigPath, "kubeconfig", "", "...")
	// rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "...")

	// TODO: rootCmd.AddCommand(deployCmd, statusCmd, rollbackCmd)
}

// Execute is called by main.go. Kept separate from main() so it's testable.
func Execute() error {
	return rootCmd.Execute()
}
