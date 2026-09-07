package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"jerremiah.dev/llmservice-cli/internal/k8s"
	"jerremiah.dev/llmservice-cli/internal/llmservice"
)

var statusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Show the current spec and status of an LLMService",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		ctx := context.Background()

		obj, err := dynClient.Resource(k8s.LLMServiceGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("%s not found in namespace %s", name, namespace)
		}
		if err != nil {
			return fmt.Errorf("getting %s: %w", name, err)
		}

		spec, status, err := llmservice.FromUnstructured(obj)
		if err != nil {
			return fmt.Errorf("reading %s: %w", name, err)
		}

		printStatus(name, spec, status)
		return nil
	},
}

// crdDefaults mirrors the +kubebuilder:default markers on LLMServiceSpec.
// Single source of truth for "what does the CRD fall back to" - update
// this map if the CRD's defaults ever change, rather than hunting through
// per-field print logic.
var crdDefaults = map[string]string{
	"image":         "vllm/vllm-openai:latest",
	"cpuRequest":    "100m",
	"memoryRequest": "2Gi",
	"port":          "8000",
}

// orDefault returns value if non-empty, otherwise looks up fieldKey in
// crdDefaults and reports what the CRD will fall back to. Fields with no
// CRD default (fieldKey not in the map) just get "(not set)".
func orDefault(value, fieldKey string) string {
	if value != "" {
		return value
	}
	if def, ok := crdDefaults[fieldKey]; ok {
		return fmt.Sprintf("(not set - CRD default: %s)", def)
	}
	return "(not set)"
}

// printStatus is separated from RunE so it can be unit tested (or reused,
// e.g. by `deploy` printing a summary after a successful apply) without
// needing a Cobra command or a cluster.
func printStatus(name string, spec llmservice.Spec, status llmservice.Status) {
	portStr := ""
	if spec.Port != 0 {
		portStr = fmt.Sprintf("%d", spec.Port)
	}

	fmt.Printf("%s\n", name)
	fmt.Printf("  phase:          %s\n", orNotSet(status.Phase))
	fmt.Printf("  model:          %s\n", orNotSet(spec.Model))
	fmt.Printf("  image:          %s\n", orDefault(spec.Image, "image"))
	fmt.Printf("  replicas:       %s\n", int32PtrOrNotSet(spec.Replicas))
	fmt.Printf("  gpu count:      %s\n", int32PtrOrNotSet(spec.GPUCount))
	fmt.Printf("  cpu request:    %s\n", orDefault(spec.CPURequest, "cpuRequest"))
	fmt.Printf("  memory request: %s\n", orDefault(spec.MemoryRequest, "memoryRequest"))
	fmt.Printf("  port:           %s\n", orDefault(portStr, "port"))
	fmt.Printf("  autoscaling:    min=%s max=%s\n",
		int32PtrOrNotSet(spec.AutoscalingMin), int32PtrOrNotSet(spec.AutoscalingMax))
}

// orNotSet is for fields with no CRD default at all (Model is required so
// this shouldn't normally trigger; status.Phase is empty simply because
// the reconciler hasn't reported yet, not because of a CRD default).
func orNotSet(s string) string {
	if s == "" {
		return "(not set)"
	}
	return s
}

func int32PtrOrNotSet(p *int32) string {
	if p == nil {
		return "(not set)"
	}
	return fmt.Sprintf("%d", *p)
}

func init() {
	// nothing to register yet - status only needs the shared --namespace flag
}
