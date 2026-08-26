package k8s

import "testing"

// TODO: NewDynamicClient talks to a real cluster via rest.Config, so it's
// not directly unit-testable without a running apiserver (or envtest).
// What IS worth a plain unit test here: LLMServiceGVR has the exact
// group/version/resource we expect. Cheap, catches typos, no cluster needed.
func TestLLMServiceGVR(t *testing.T) {
	t.Skip("TODO: assert LLMServiceGVR.Group/Version/Resource match expected values")
}
