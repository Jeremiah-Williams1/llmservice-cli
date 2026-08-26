package cmd

import (
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Show the current spec and status of an LLMService",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO:
		// 1. name := args[0]
		// 2. obj, err := client.Resource(k8s.LLMServiceGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
		// 3. spec, status, err := llmservice.FromUnstructured(obj)
		// 4. print both - status should surface what deploy/spec alone can't
		//    (e.g. .status.conditions, ready replicas, ScaledObject state)
		panic("TODO: not implemented")
	},
}

func init() {
	// nothing to register yet - status only needs the shared --namespace flag
}
