package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"

	"jerremiah.dev/llmservice-cli/internal/k8s"
)

// Persistent flags shared by all subcommands.
var (
	kubeconfigPath string
	namespace      string
)

// dynClient is built once in PersistentPreRunE and read by every
// subcommand's RunE. Package-level var is the simplest way to share it
// across cmd/*.go without threading it through every function signature -
// acceptable here since this is a small CLI with one client, not a
// library other code will import.
var dynClient dynamic.Interface

var rootCmd = &cobra.Command{
	Use:   "llmservice-cli",
	Short: "Deploy, inspect, and roll back LLMService resources",
	// PersistentPreRunE runs once, after flag parsing, before ANY
	// subcommand's RunE - so by the time deploy/status/rollback execute,
	// dynClient is guaranteed to be ready. Using PersistentPreRunE (not
	// PreRunE) means this also runs for subcommands, not just rootCmd
	// itself.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		client, err := k8s.NewDynamicClient(kubeconfigPath)
		if err != nil {
			return fmt.Errorf("connecting to cluster: %w", err)
		}
		dynClient = client
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&kubeconfigPath, "kubeconfig", "",
		"path to kubeconfig file (defaults to ~/.kube/config)")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default",
		"namespace to operate in")

	rootCmd.AddCommand(deployCmd, statusCmd, rollbackCmd)
}

// Execute is called by main.go. Kept separate from main() so it's testable.
func Execute() error {
	return rootCmd.Execute()
}
