package k8s

import "testing"

// NewDynamicClient talks to a real cluster via rest.Config, so it's not
// directly unit-testable without a running apiserver (or envtest). What
// IS worth a plain unit test here: LLMServiceGVR has the exact
// group/version/resource we expect. Cheap, catches typos, no cluster needed
// — and this is the one line in the whole repo where a typo silently sends
// every request to the wrong URL.
func TestLLMServiceGVR(t *testing.T) {
	want := struct {
		Group    string
		Version  string
		Resource string
	}{
		Group:    "apps.jerremiah.dev",
		Version:  "v1alpha1",
		Resource: "llmservices",
	}

	if LLMServiceGVR.Group != want.Group {
		t.Errorf("Group = %q, want %q", LLMServiceGVR.Group, want.Group)
	}
	if LLMServiceGVR.Version != want.Version {
		t.Errorf("Version = %q, want %q", LLMServiceGVR.Version, want.Version)
	}
	if LLMServiceGVR.Resource != want.Resource {
		t.Errorf("Resource = %q, want %q", LLMServiceGVR.Resource, want.Resource)
	}
}

// NewDynamicClient's error paths ARE testable without a cluster, since
// BuildConfigFromFlags/NewForConfig fail on bad local input before any
// network call happens (see the "never contacts the cluster" note in
// client.go). This catches a bad/missing kubeconfig path failing loudly
// with a useful error, rather than a nil client or a confusing panic.
func TestNewDynamicClient_BadPathReturnsError(t *testing.T) {
	_, err := NewDynamicClient("/definitely/does/not/exist/kubeconfig.yaml")
	if err == nil {
		t.Fatal("expected an error for a nonexistent kubeconfig path, got nil")
	}
}
