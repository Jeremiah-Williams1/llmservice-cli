package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"jerremiah.dev/llmservice-cli/internal/k8s"
	"jerremiah.dev/llmservice-cli/internal/llmservice"
)

// previousSpecAnnotation stores the spec that was live immediately before
// a deploy overwrote it, as JSON. rollback.go reads this back. Shared
// between deploy.go and rollback.go since both live in package cmd.
const previousSpecAnnotation = "llmservice-cli/previous-spec"

var (
	flagModel          string
	flagImage          string
	flagCPURequest     string
	flagMemoryRequest  string
	flagPort           int32
	flagReplicas       int32
	flagGPUCount       int32
	flagAutoscalingMin int32
	flagAutoscalingMax int32
)

var deployCmd = &cobra.Command{
	Use:   "deploy [name]",
	Short: "Deploy a new LLMService, or update an existing one",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		spec := buildSpecFromFlags(cmd)

		ctx := context.Background()
		resourceClient := dynClient.Resource(k8s.LLMServiceGVR).Namespace(namespace)

		existing, err := resourceClient.Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			// First deploy for this name - nothing to snapshot for rollback.
			obj := llmservice.ToUnstructured(name, namespace, spec)
			if _, err := resourceClient.Create(ctx, obj, metav1.CreateOptions{}); err != nil {
				return fmt.Errorf("creating %s: %w", name, err)
			}
			fmt.Printf("created %s\n", name)
			return nil
		}
		if err != nil {
			return fmt.Errorf("checking for existing %s: %w", name, err)
		}

		// Already exists - snapshot its current spec into the new object's
		// annotation before overwriting, so rollback has something to
		// revert to. Update (not Create) requires the current
		// resourceVersion, or the API server rejects it as a conflict.
		previousSpec, _, err := llmservice.FromUnstructured(existing)
		if err != nil {
			return fmt.Errorf("reading current spec of %s before update: %w", name, err)
		}
		previousSpecJSON, err := json.Marshal(previousSpec)
		if err != nil {
			return fmt.Errorf("encoding previous spec for rollback annotation: %w", err)
		}

		obj := llmservice.ToUnstructured(name, namespace, spec)
		obj.SetResourceVersion(existing.GetResourceVersion())
		obj.SetAnnotations(map[string]string{
			previousSpecAnnotation: string(previousSpecJSON),
		})

		if _, err := resourceClient.Update(ctx, obj, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("updating %s: %w", name, err)
		}
		fmt.Printf("updated %s (previous spec saved for rollback)\n", name)
		return nil
	},
}

// buildSpecFromFlags reads parsed flag values into a Spec. For the *int32
// fields, cmd.Flags().Changed(...) distinguishes "user passed 0
// explicitly" from "user didn't pass this flag at all" - the zero value
// alone can't tell those apart, since 0 is a meaningful value (e.g.
// --replicas 0).
func buildSpecFromFlags(cmd *cobra.Command) llmservice.Spec {
	spec := llmservice.Spec{
		Model:         flagModel,
		Image:         flagImage,
		CPURequest:    flagCPURequest,
		MemoryRequest: flagMemoryRequest,
		Port:          flagPort,
	}

	if cmd.Flags().Changed("replicas") {
		v := flagReplicas
		spec.Replicas = &v
	}
	if cmd.Flags().Changed("gpu-count") {
		v := flagGPUCount
		spec.GPUCount = &v
	}
	if cmd.Flags().Changed("autoscaling-min") {
		v := flagAutoscalingMin
		spec.AutoscalingMin = &v
	}
	if cmd.Flags().Changed("autoscaling-max") {
		v := flagAutoscalingMax
		spec.AutoscalingMax = &v
	}

	return spec
}

func init() {
	deployCmd.Flags().StringVar(&flagModel, "model", "", "model identifier (required)")
	deployCmd.Flags().StringVar(&flagImage, "image", "", "container image (default: CRD default)")
	deployCmd.Flags().StringVar(&flagCPURequest, "cpu-request", "", "CPU request, e.g. 500m (default: CRD default)")
	deployCmd.Flags().StringVar(&flagMemoryRequest, "memory-request", "", "memory request, e.g. 4Gi (default: CRD default)")
	deployCmd.Flags().Int32Var(&flagPort, "port", 0, "container port (default: CRD default)")
	deployCmd.Flags().Int32Var(&flagReplicas, "replicas", 0, "replica count")
	deployCmd.Flags().Int32Var(&flagGPUCount, "gpu-count", 0, "GPUs per replica")
	deployCmd.Flags().Int32Var(&flagAutoscalingMin, "autoscaling-min", 0, "KEDA min replica count")
	deployCmd.Flags().Int32Var(&flagAutoscalingMax, "autoscaling-max", 0, "KEDA max replica count")

	_ = deployCmd.MarkFlagRequired("model")
}
