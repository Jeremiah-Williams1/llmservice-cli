package llmservice

import "testing"

// This is the highest-value test file in the repo: convert.go is the one
// place raw map[string]interface{} manipulation happens, so it's the one
// place a typo'd field name or wrong type silently breaks at runtime
// instead of compile time. No cluster needed for any of these - pure
// data transformation.

func TestToUnstructured_SetsCorrectApiVersionAndKind(t *testing.T) {
	t.Skip("TODO: ToUnstructured(...).GetAPIVersion() == apps.jerremiah.dev/v1alpha1, GetKind() == LLMService")
}

func TestToUnstructured_MapsAllSpecFields(t *testing.T) {
	t.Skip("TODO: build a Spec with every field set to a distinct value, " +
		"convert, then assert each field landed at the right nested path " +
		"via unstructured.NestedString/NestedInt64")
}

func TestToUnstructured_GPUCountIsInt64NotInt32(t *testing.T) {
	t.Skip("TODO: unstructured content requires int64 - assert the stored " +
		"type is int64, not int32, or NestedInt64 will fail at read time")
}

func TestFromUnstructured_RoundTrip(t *testing.T) {
	t.Skip("TODO: Spec -> ToUnstructured -> FromUnstructured -> assert equal to original Spec")
}

func TestFromUnstructured_MissingFieldsDontPanic(t *testing.T) {
	t.Skip("TODO: pass a minimal/partial unstructured object (as if an old " +
		"CR predates a newer field) and assert FromUnstructured returns an " +
		"error or zero value, not a panic")
}
