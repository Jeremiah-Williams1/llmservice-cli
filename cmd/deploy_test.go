package cmd

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"jerremiah.dev/llmservice-cli/internal/k8s"
	"jerremiah.dev/llmservice-cli/internal/llmservice"
)

// newTestFlagCmd builds a fresh *cobra.Command with the same flags
// deployCmd registers, bound to the same shared package-level vars
// (flagModel, flagReplicas, etc). A fresh FlagSet is needed per test
// because pflag's "Changed" bookkeeping lives on the FlagSet and never
// resets once true - reusing deployCmd's real FlagSet across tests would
// let one test's --replicas leak "Changed=true" into the next test.
func newTestFlagCmd() *cobra.Command {
	c := &cobra.Command{}
	c.Flags().StringVar(&flagModel, "model", "", "")
	c.Flags().StringVar(&flagImage, "image", "", "")
	c.Flags().StringVar(&flagCPURequest, "cpu-request", "", "")
	c.Flags().StringVar(&flagMemoryRequest, "memory-request", "", "")
	c.Flags().Int32Var(&flagPort, "port", 0, "")
	c.Flags().Int32Var(&flagReplicas, "replicas", 0, "")
	c.Flags().Int32Var(&flagGPUCount, "gpu-count", 0, "")
	c.Flags().Int32Var(&flagAutoscalingMin, "autoscaling-min", 0, "")
	c.Flags().Int32Var(&flagAutoscalingMax, "autoscaling-max", 0, "")
	return c
}

// resetFlagVars zeroes the shared package-level flag vars between tests,
// since they're global state that would otherwise leak between test cases.
func resetFlagVars() {
	flagModel = ""
	flagImage = ""
	flagCPURequest = ""
	flagMemoryRequest = ""
	flagPort = 0
	flagReplicas = 0
	flagGPUCount = 0
	flagAutoscalingMin = 0
	flagAutoscalingMax = 0
}

func TestBuildSpecFromFlags_ExplicitZeroBecomesNonNilPointer(t *testing.T) {
	resetFlagVars()
	c := newTestFlagCmd()
	if err := c.Flags().Parse([]string{"--model", "meta-llama/Llama-3-8b", "--replicas", "0"}); err != nil {
		t.Fatalf("parsing flags: %v", err)
	}

	spec := buildSpecFromFlags(c)

	if spec.Replicas == nil {
		t.Fatal("Replicas is nil, want a pointer to 0 - the flag was explicitly passed")
	}
	if *spec.Replicas != 0 {
		t.Errorf("Replicas = %d, want 0", *spec.Replicas)
	}
}

func TestBuildSpecFromFlags_UnpassedFlagStaysNil(t *testing.T) {
	resetFlagVars()
	c := newTestFlagCmd()
	// --gpu-count deliberately not passed.
	if err := c.Flags().Parse([]string{"--model", "meta-llama/Llama-3-8b"}); err != nil {
		t.Fatalf("parsing flags: %v", err)
	}

	spec := buildSpecFromFlags(c)

	if spec.GPUCount != nil {
		t.Errorf("GPUCount = %v, want nil since --gpu-count was never passed", *spec.GPUCount)
	}
}

func TestBuildSpecFromFlags_StringFieldsUseEmptyAsUnset(t *testing.T) {
	resetFlagVars()
	c := newTestFlagCmd()
	if err := c.Flags().Parse([]string{"--model", "meta-llama/Llama-3-8b"}); err != nil {
		t.Fatalf("parsing flags: %v", err)
	}

	spec := buildSpecFromFlags(c)

	if spec.Image != "" {
		t.Errorf("Image = %q, want empty since --image was never passed", spec.Image)
	}
}

func TestDeploy_CreatesNewResourceWhenNotFound(t *testing.T) {
	withFakeClient(t, newFakeDynClient(t)) // empty - nothing exists yet
	namespace = "default"
	resetFlagVars()

	c := newTestFlagCmd()
	if err := c.Flags().Parse([]string{"--model", "meta-llama/Llama-3-8b"}); err != nil {
		t.Fatalf("parsing flags: %v", err)
	}

	if err := deployCmd.RunE(c, []string{"my-model"}); err != nil {
		t.Fatalf("deploy returned error: %v", err)
	}

	got, err := dynClient.Resource(k8s.LLMServiceGVR).Namespace("default").
		Get(context.Background(), "my-model", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("expected resource to exist after deploy, got error: %v", err)
	}
	spec, _, err := llmservice.FromUnstructured(got)
	if err != nil {
		t.Fatalf("reading spec: %v", err)
	}
	if spec.Model != "meta-llama/Llama-3-8b" {
		t.Errorf("Model = %q, want %q", spec.Model, "meta-llama/Llama-3-8b")
	}
	if v, ok := got.GetAnnotations()[previousSpecAnnotation]; ok && v != "" {
		t.Errorf("previousSpecAnnotation = %q, want unset on first create", v)
	}
}

func TestDeploy_UpdatingExistingResourceStoresPreviousSpec(t *testing.T) {
	existing := llmservice.ToUnstructured("my-model", "default",
		llmservice.Spec{Model: "meta-llama/Llama-3-8b", GPUCount: int32Ptr(1)})
	existing.SetResourceVersion("1")

	withFakeClient(t, newFakeDynClient(t, existing))
	namespace = "default"
	resetFlagVars()

	c := newTestFlagCmd()
	if err := c.Flags().Parse([]string{"--model", "meta-llama/Llama-3-8b", "--gpu-count", "2"}); err != nil {
		t.Fatalf("parsing flags: %v", err)
	}

	if err := deployCmd.RunE(c, []string{"my-model"}); err != nil {
		t.Fatalf("deploy returned error: %v", err)
	}

	got, err := dynClient.Resource(k8s.LLMServiceGVR).Namespace("default").
		Get(context.Background(), "my-model", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("re-fetching after deploy: %v", err)
	}

	spec, _, err := llmservice.FromUnstructured(got)
	if err != nil {
		t.Fatalf("reading spec: %v", err)
	}
	if spec.GPUCount == nil || *spec.GPUCount != 2 {
		t.Errorf("GPUCount = %v, want 2 (the new value)", derefOrNil(spec.GPUCount))
	}

	previousJSON, ok := got.GetAnnotations()[previousSpecAnnotation]
	if !ok || previousJSON == "" {
		t.Fatal("expected previousSpecAnnotation to be set after updating an existing resource")
	}
	var previous llmservice.Spec
	if err := json.Unmarshal([]byte(previousJSON), &previous); err != nil {
		t.Fatalf("decoding stored previous spec: %v", err)
	}
	if previous.GPUCount == nil || *previous.GPUCount != 1 {
		t.Errorf("stored previous GPUCount = %v, want 1 (the value before this update)", derefOrNil(previous.GPUCount))
	}
}
