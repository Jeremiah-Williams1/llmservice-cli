package cmd

import "testing"

// These don't need a real cluster - use a k8s.io/client-go/dynamic/fake
// client (fake.NewSimpleDynamicClient) so the whole deploy->rollback flow
// can be tested without minikube. Worth setting up once deploy.go and
// rollback.go have real logic.

func TestRollback_RevertsToStoredPreviousSpec(t *testing.T) {
	t.Skip("TODO: deploy A, deploy B (stores A as previous-spec annotation), " +
		"rollback, assert live spec == A")
}

func TestRollback_ErrorsWhenNoPreviousSpecAnnotation(t *testing.T) {
	t.Skip("TODO: object with no previous-spec annotation -> rollback returns " +
		"a clear error, not a panic or silent no-op")
}
