package llmservice

// Spec mirrors the LLMService CRD's spec fields. This is a hand-maintained
// copy, NOT imported from the operator repo — see convert.go for why.
//
// IMPORTANT: if apps.jerremiah.dev/v1alpha1.LLMServiceSpec changes in the
// operator repo, this struct must be updated to match by hand. Nothing
// here will catch drift automatically — that's the tradeoff of the
// dynamic-client approach.
type Spec struct {
	Model    string
	Image    string
	Replicas int32
	GPUCount int32

	CPURequest    string
	MemoryRequest string

	Port int32

	AutoscalingMin int32
	AutoscalingMax int32
}

// Status mirrors the fields we care about from .status, populated by
// FromUnstructured in convert.go. Only fields the CLI actually displays
// need to live here — this doesn't need to mirror the full CRD status.
type Status struct {
	// TODO: fill in once we know what the reconciler actually writes
	// back (e.g. Conditions, ReadyReplicas, ScaledObjectState).
}
