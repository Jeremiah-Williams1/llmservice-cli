package cmd

import (
	"context"
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	"jerremiah.dev/llmservice-cli/internal/k8s"
	"jerremiah.dev/llmservice-cli/internal/llmservice"
)

// int32Ptr and derefOrNil are small test-only helpers for working with
// *int32 fields. Duplicated here (not imported from internal/llmservice's
// test file) because Go doesn't allow one package's tests to use another
// package's unexported test helpers - convert_test.go in internal/llmservice
// has its own copies of these for the same reason.
func int32Ptr(v int32) *int32 { return &v }

func derefOrNil(p *int32) interface{} {
	if p == nil {
		return nil
	}
	return *p
}

// newFakeDynClient wires up an in-memory dynamic.Interface pre-populated
// with the given objects. The fake client needs to know the "list kind"
// for any custom GVR it'll serve (LLMServiceList here) - the real API
// server infers this itself, the fake needs it told explicitly since it
// has no CRD schema to consult.
func newFakeDynClient(t *testing.T, objects ...*unstructured.Unstructured) *dynamicfake.FakeDynamicClient {
	t.Helper()
	scheme := runtime.NewScheme()
	gvrToListKind := map[schema.GroupVersionResource]string{
		k8s.LLMServiceGVR: "LLMServiceList",
	}
	objs := make([]runtime.Object, len(objects))
	for i, o := range objects {
		objs[i] = o
	}
	return dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, gvrToListKind, objs...)
}

// withFakeClient swaps the package-level dynClient for the duration of a
// test, restoring the original afterwards. dynClient being a mutable
// package var (flagged as a tradeoff back in root.go) is exactly what
// makes this necessary - a cleaner design would inject the client instead
// of reaching into a global, but this works for a small, non-parallel
// test suite.
func withFakeClient(t *testing.T, fake *dynamicfake.FakeDynamicClient) {
	t.Helper()
	original := dynClient
	dynClient = fake
	t.Cleanup(func() { dynClient = original })
}

func TestRollback_RevertsToStoredPreviousSpec(t *testing.T) {
	previous := llmservice.Spec{Model: "meta-llama/Llama-3-8b", GPUCount: int32Ptr(1)}
	current := llmservice.Spec{Model: "meta-llama/Llama-3-8b", GPUCount: int32Ptr(2)}

	existing := llmservice.ToUnstructured("my-model", "default", current)
	existing.SetResourceVersion("1")
	previousJSON, err := json.Marshal(previous)
	if err != nil {
		t.Fatalf("marshaling previous spec: %v", err)
	}
	existing.SetAnnotations(map[string]string{previousSpecAnnotation: string(previousJSON)})

	withFakeClient(t, newFakeDynClient(t, existing))
	namespace = "default"

	if err := rollbackCmd.RunE(rollbackCmd, []string{"my-model"}); err != nil {
		t.Fatalf("rollback returned error: %v", err)
	}

	got, err := dynClient.Resource(k8s.LLMServiceGVR).Namespace("default").
		Get(context.Background(), "my-model", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("re-fetching after rollback: %v", err)
	}
	spec, _, err := llmservice.FromUnstructured(got)
	if err != nil {
		t.Fatalf("reading spec after rollback: %v", err)
	}
	if spec.GPUCount == nil || *spec.GPUCount != 1 {
		t.Errorf("GPUCount after rollback = %v, want 1 (the previous spec)", derefOrNil(spec.GPUCount))
	}

	// Annotation should be cleared - this is one level of undo, not a redo
	// stack (see the comment in rollback.go for why).
	if v := got.GetAnnotations()[previousSpecAnnotation]; v != "" {
		t.Errorf("previousSpecAnnotation after rollback = %q, want empty", v)
	}
}

func TestRollback_ErrorsWhenNoPreviousSpecAnnotation(t *testing.T) {
	// No previousSpecAnnotation set at all - as if this object was created
	// by a single `deploy` and never updated since.
	existing := llmservice.ToUnstructured("my-model", "default", llmservice.Spec{Model: "meta-llama/Llama-3-8b"})
	existing.SetResourceVersion("1")

	withFakeClient(t, newFakeDynClient(t, existing))
	namespace = "default"

	err := rollbackCmd.RunE(rollbackCmd, []string{"my-model"})
	if err == nil {
		t.Fatal("expected an error when no previous spec is recorded, got nil")
	}
}

func TestRollback_ErrorsWhenResourceDoesNotExist(t *testing.T) {
	withFakeClient(t, newFakeDynClient(t)) // no objects at all
	namespace = "default"

	err := rollbackCmd.RunE(rollbackCmd, []string{"does-not-exist"})
	if err == nil {
		t.Fatal("expected an error for a nonexistent resource, got nil")
	}
}
