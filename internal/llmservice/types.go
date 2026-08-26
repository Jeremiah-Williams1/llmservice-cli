package llmservice

// Spec mirrors the LLMService CRD's spec fields. This is a hand-maintained
// copy, NOT imported from the operator repo — see convert.go for why.
//
// IMPORTANT: if apps.jerremiah.dev/v1alpha1.LLMServiceSpec changes in the
// operator repo, this struct must be updated to match by hand. Nothing
// here will catch drift automatically — that's the tradeoff of the
// dynamic-client approach.
//
// Optionality mirrors the CRD exactly:
//   - Model: required, always set.
//   - Image, CPURequest, MemoryRequest: plain string + omitempty in the
//     CRD -> "" means unset here too, no pointer needed.
//   - Port: int32 + omitempty in the CRD -> 0 means unset.
//   - Replicas, GPUCount: *int32 in the CRD -> nil means unset here too,
//     since these have no default and 0 is a meaningfully different value
//     from "not provided."
//   - AutoscalingMin/Max: same reasoning, *int32 in the CRD's nested
//     AutoscalingConfig.
type Spec struct {
	Model string // required, no omitempty in the CRD

	Image         string // "" = unset, CRD defaults to vllm/vllm-openai:latest
	CPURequest    string // "" = unset, CRD defaults to 100m
	MemoryRequest string // "" = unset, CRD defaults to 2Gi
	Port          int32  // 0 = unset, CRD defaults to 8000

	Replicas *int32 // nil = unset, no CRD default
	GPUCount *int32 // nil = unset, no CRD default

	AutoscalingMin *int32 // nil = unset (spec.autoscaling.minReplicaCount)
	AutoscalingMax *int32 // nil = unset (spec.autoscaling.maxReplicaCount)
}

// Status mirrors the fields we care about from .status, populated by
// FromUnstructured in convert.go. The real CRD status has Conditions
// ([]metav1.Condition) and Phase (string) - starting with just Phase
// since that's the simplest thing `status` can print; Conditions can be
// added once we decide how much detail to surface.
type Status struct {
	Phase string
}
