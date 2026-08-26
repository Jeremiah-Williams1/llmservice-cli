package llmservice

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func int32Ptr(v int32) *int32 { return &v }

func TestToUnstructured_SetsCorrectApiVersionAndKind(t *testing.T) {
	obj := ToUnstructured("my-model", "default", Spec{Model: "meta-llama/Llama-3-8b"})

	if got := obj.GetAPIVersion(); got != "apps.jerremiah.dev/v1alpha1" {
		t.Errorf("APIVersion = %q, want %q", got, "apps.jerremiah.dev/v1alpha1")
	}
	if got := obj.GetKind(); got != "LLMService" {
		t.Errorf("Kind = %q, want %q", got, "LLMService")
	}
	if got := obj.GetName(); got != "my-model" {
		t.Errorf("Name = %q, want %q", got, "my-model")
	}
	if got := obj.GetNamespace(); got != "default" {
		t.Errorf("Namespace = %q, want %q", got, "default")
	}
}

func TestToUnstructured_MapsAllSpecFields(t *testing.T) {
	spec := Spec{
		Model:          "meta-llama/Llama-3-8b",
		Image:          "vllm/vllm-openai:v0.5.0",
		CPURequest:     "500m",
		MemoryRequest:  "4Gi",
		Port:           8080,
		Replicas:       int32Ptr(3),
		GPUCount:       int32Ptr(2),
		AutoscalingMin: int32Ptr(1),
		AutoscalingMax: int32Ptr(5),
	}

	obj := ToUnstructured("my-model", "default", spec)

	assertString(t, obj, "meta-llama/Llama-3-8b", "spec", "model")
	assertString(t, obj, "vllm/vllm-openai:v0.5.0", "spec", "image")
	assertString(t, obj, "500m", "spec", "cpuRequest")
	assertString(t, obj, "4Gi", "spec", "memoryRequest")
	assertInt64(t, obj, 8080, "spec", "port")
	assertInt64(t, obj, 3, "spec", "replicas")
	assertInt64(t, obj, 2, "spec", "gpuCount")
	assertInt64(t, obj, 1, "spec", "autoscaling", "minReplicaCount")
	assertInt64(t, obj, 5, "spec", "autoscaling", "maxReplicaCount")
}

func TestToUnstructured_UnsetOptionalFieldsAreOmitted(t *testing.T) {
	// Only Model set - everything else left at zero value / nil, mirroring
	// a user who only passed --model on the CLI.
	spec := Spec{Model: "meta-llama/Llama-3-8b"}

	obj := ToUnstructured("my-model", "default", spec)

	assertFieldAbsent(t, obj, "spec", "image")
	assertFieldAbsent(t, obj, "spec", "cpuRequest")
	assertFieldAbsent(t, obj, "spec", "memoryRequest")
	assertFieldAbsent(t, obj, "spec", "port")
	assertFieldAbsent(t, obj, "spec", "replicas")
	assertFieldAbsent(t, obj, "spec", "gpuCount")
	assertFieldAbsent(t, obj, "spec", "autoscaling")
}

func TestFromUnstructured_RoundTrip(t *testing.T) {
	original := Spec{
		Model:          "meta-llama/Llama-3-8b",
		Image:          "vllm/vllm-openai:v0.5.0",
		CPURequest:     "500m",
		MemoryRequest:  "4Gi",
		Port:           8080,
		Replicas:       int32Ptr(3),
		GPUCount:       int32Ptr(2),
		AutoscalingMin: int32Ptr(1),
		AutoscalingMax: int32Ptr(5),
	}

	obj := ToUnstructured("my-model", "default", original)
	got, _, err := FromUnstructured(obj)
	if err != nil {
		t.Fatalf("FromUnstructured returned error: %v", err)
	}

	if got.Model != original.Model ||
		got.Image != original.Image ||
		got.CPURequest != original.CPURequest ||
		got.MemoryRequest != original.MemoryRequest ||
		got.Port != original.Port {
		t.Errorf("string/int32 fields did not round-trip: got %+v, want %+v", got, original)
	}
	if !ptrEqual(got.Replicas, original.Replicas) {
		t.Errorf("Replicas = %v, want %v", derefOrNil(got.Replicas), derefOrNil(original.Replicas))
	}
	if !ptrEqual(got.GPUCount, original.GPUCount) {
		t.Errorf("GPUCount = %v, want %v", derefOrNil(got.GPUCount), derefOrNil(original.GPUCount))
	}
	if !ptrEqual(got.AutoscalingMin, original.AutoscalingMin) {
		t.Errorf("AutoscalingMin = %v, want %v", derefOrNil(got.AutoscalingMin), derefOrNil(original.AutoscalingMin))
	}
	if !ptrEqual(got.AutoscalingMax, original.AutoscalingMax) {
		t.Errorf("AutoscalingMax = %v, want %v", derefOrNil(got.AutoscalingMax), derefOrNil(original.AutoscalingMax))
	}
}

func TestFromUnstructured_MissingOptionalFieldsStayNil(t *testing.T) {
	// Minimal object, as if an old CR predates the autoscaling field, or
	// the user only ever set --model.
	obj := ToUnstructured("my-model", "default", Spec{Model: "meta-llama/Llama-3-8b"})

	got, status, err := FromUnstructured(obj)
	if err != nil {
		t.Fatalf("FromUnstructured returned error: %v", err)
	}
	if got.Replicas != nil {
		t.Errorf("Replicas = %v, want nil", *got.Replicas)
	}
	if got.GPUCount != nil {
		t.Errorf("GPUCount = %v, want nil", *got.GPUCount)
	}
	if got.AutoscalingMin != nil {
		t.Errorf("AutoscalingMin = %v, want nil", *got.AutoscalingMin)
	}
	if status.Phase != "" {
		t.Errorf("Phase = %q, want empty (reconciler hasn't run yet)", status.Phase)
	}
}

func TestFromUnstructured_ReadsStatusPhase(t *testing.T) {
	obj := ToUnstructured("my-model", "default", Spec{Model: "meta-llama/Llama-3-8b"})
	_ = unstructured.SetNestedField(obj.Object, "Running", "status", "phase")

	_, status, err := FromUnstructured(obj)
	if err != nil {
		t.Fatalf("FromUnstructured returned error: %v", err)
	}
	if status.Phase != "Running" {
		t.Errorf("Phase = %q, want %q", status.Phase, "Running")
	}
}

// --- helpers ---

func assertString(t *testing.T, obj *unstructured.Unstructured, want string, path ...string) {
	t.Helper()
	got, found, err := unstructured.NestedString(obj.Object, path...)
	if err != nil {
		t.Fatalf("reading %v: %v", path, err)
	}
	if !found {
		t.Fatalf("%v not found in object", path)
	}
	if got != want {
		t.Errorf("%v = %q, want %q", path, got, want)
	}
}

func assertInt64(t *testing.T, obj *unstructured.Unstructured, want int64, path ...string) {
	t.Helper()
	got, found, err := unstructured.NestedInt64(obj.Object, path...)
	if err != nil {
		t.Fatalf("reading %v: %v", path, err)
	}
	if !found {
		t.Fatalf("%v not found in object", path)
	}
	if got != want {
		t.Errorf("%v = %d, want %d", path, got, want)
	}
}

func assertFieldAbsent(t *testing.T, obj *unstructured.Unstructured, path ...string) {
	t.Helper()
	_, found, err := unstructured.NestedFieldNoCopy(obj.Object, path...)
	if err != nil {
		t.Fatalf("reading %v: %v", path, err)
	}
	if found {
		t.Errorf("%v = present, want absent (unset field should be omitted, not zero-valued)", path)
	}
}

func ptrEqual(a, b *int32) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func derefOrNil(p *int32) interface{} {
	if p == nil {
		return nil
	}
	return *p
}
