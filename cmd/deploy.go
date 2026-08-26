package cmd

import (
	"github.com/spf13/cobra"
)

// TODO: flag vars for --model, --image, --replicas, --gpu-count,
// --cpu-request, --memory-request, --port, --autoscaling-min, --autoscaling-max

var deployCmd = &cobra.Command{
	Use:   "deploy [name]",
	Short: "Deploy a new LLMService",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO:
		// 1. name := args[0]
		// 2. build llmservice.Spec from flags
		// 3. before creating: snapshot current spec (if any) for rollback -
		//    see rollback.go for what "simple rollback" needs here
		// 4. obj := llmservice.ToUnstructured(name, namespace, spec)
		// 5. client.Resource(k8s.LLMServiceGVR).Namespace(namespace).Create(ctx, obj, metav1.CreateOptions{})
		panic("TODO: not implemented")
	},
}

func init() {
	// TODO: register flags on deployCmd
}
