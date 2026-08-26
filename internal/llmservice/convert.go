package llmservice

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	apiVersion = "apps.jerremiah.dev/v1alpha1"
	kind       = "LLMService"
)

// ToUnstructured builds the object the dynamic client can Create()/Update().
// This is where Spec's clean Go fields get flattened into the
// map[string]interface{} shape the K8s API actually expects on the wire.
func ToUnstructured(name, namespace string, spec Spec) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion(apiVersion)
	obj.SetKind(kind)
	obj.SetName(name)
	obj.SetNamespace(namespace)

	// unstructured.SetNestedField writes into obj.Object at the given
	// path, creating intermediate maps as needed. The path arguments
	// ("spec", "model") mirror exactly where the field lives in the
	// resulting JSON - "spec.model", "spec.gpuCount", etc.
	//
	// Model is required - always set, no nil/empty check.
	_ = unstructured.SetNestedField(obj.Object, spec.Model, "spec", "model")

	// Image/CPURequest/MemoryRequest/Port: "" or 0 means "not provided by
	// the user" - skip the field entirely so the CRD's kubebuilder default
	// applies server-side, rather than writing an explicit empty value
	// that would override the default.
	if spec.Image != "" {
		_ = unstructured.SetNestedField(obj.Object, spec.Image, "spec", "image")
	}
	if spec.CPURequest != "" {
		_ = unstructured.SetNestedField(obj.Object, spec.CPURequest, "spec", "cpuRequest")
	}
	if spec.MemoryRequest != "" {
		_ = unstructured.SetNestedField(obj.Object, spec.MemoryRequest, "spec", "memoryRequest")
	}
	if spec.Port != 0 {
		_ = unstructured.SetNestedField(obj.Object, int64(spec.Port), "spec", "port")
	}

	// Replicas/GPUCount: *int32 in the CRD, no default - nil means "not
	// provided," so skip rather than writing a misleading 0.
	if spec.Replicas != nil {
		_ = unstructured.SetNestedField(obj.Object, int64(*spec.Replicas), "spec", "replicas")
	}
	if spec.GPUCount != nil {
		_ = unstructured.SetNestedField(obj.Object, int64(*spec.GPUCount), "spec", "gpuCount")
	}

	// AutoscalingMin/Max: same nil-means-unset reasoning, nested one level
	// deeper under spec.autoscaling to match AutoscalingConfig's JSON tags.
	if spec.AutoscalingMin != nil {
		_ = unstructured.SetNestedField(obj.Object, int64(*spec.AutoscalingMin), "spec", "autoscaling", "minReplicaCount")
	}
	if spec.AutoscalingMax != nil {
		_ = unstructured.SetNestedField(obj.Object, int64(*spec.AutoscalingMax), "spec", "autoscaling", "maxReplicaCount")
	}

	return obj
}

// FromUnstructured does the reverse - pulls spec/status fields back out of
// what Get() returns, for `status` to print. Uses the NestedX helpers
// rather than raw map indexing, since those handle missing-field cases
// (an old CR predating a newer field, or a field left unset because the
// CRD default hasn't been applied yet) without panicking.
func FromUnstructured(obj *unstructured.Unstructured) (Spec, Status, error) {
	var spec Spec
	var status Status

	// NestedString/NestedInt64/etc return (value, found, err).
	// found=false means the key wasn't present at all - not the same as
	// "present but zero/empty." err is non-nil only if a field along the
	// path exists but isn't the expected type (e.g. spec.port is a string
	// instead of a number) - a real schema-mismatch bug worth surfacing.

	model, _, err := unstructured.NestedString(obj.Object, "spec", "model")
	if err != nil {
		return spec, status, fmt.Errorf("reading spec.model: %w", err)
	}
	spec.Model = model

	image, _, err := unstructured.NestedString(obj.Object, "spec", "image")
	if err != nil {
		return spec, status, fmt.Errorf("reading spec.image: %w", err)
	}
	spec.Image = image

	cpuRequest, _, err := unstructured.NestedString(obj.Object, "spec", "cpuRequest")
	if err != nil {
		return spec, status, fmt.Errorf("reading spec.cpuRequest: %w", err)
	}
	spec.CPURequest = cpuRequest

	memoryRequest, _, err := unstructured.NestedString(obj.Object, "spec", "memoryRequest")
	if err != nil {
		return spec, status, fmt.Errorf("reading spec.memoryRequest: %w", err)
	}
	spec.MemoryRequest = memoryRequest

	port, _, err := unstructured.NestedInt64(obj.Object, "spec", "port")
	if err != nil {
		return spec, status, fmt.Errorf("reading spec.port: %w", err)
	}
	spec.Port = int32(port)

	// Replicas/GPUCount are *int32 in Spec - only allocate a pointer if
	// the field was actually found, otherwise leave nil.
	if replicas, found, err := unstructured.NestedInt64(obj.Object, "spec", "replicas"); err != nil {
		return spec, status, fmt.Errorf("reading spec.replicas: %w", err)
	} else if found {
		r := int32(replicas)
		spec.Replicas = &r
	}

	if gpuCount, found, err := unstructured.NestedInt64(obj.Object, "spec", "gpuCount"); err != nil {
		return spec, status, fmt.Errorf("reading spec.gpuCount: %w", err)
	} else if found {
		g := int32(gpuCount)
		spec.GPUCount = &g
	}

	if min, found, err := unstructured.NestedInt64(obj.Object, "spec", "autoscaling", "minReplicaCount"); err != nil {
		return spec, status, fmt.Errorf("reading spec.autoscaling.minReplicaCount: %w", err)
	} else if found {
		m := int32(min)
		spec.AutoscalingMin = &m
	}

	if max, found, err := unstructured.NestedInt64(obj.Object, "spec", "autoscaling", "maxReplicaCount"); err != nil {
		return spec, status, fmt.Errorf("reading spec.autoscaling.maxReplicaCount: %w", err)
	} else if found {
		m := int32(max)
		spec.AutoscalingMax = &m
	}

	// status.phase - fine for this to just come back "" if the reconciler
	// hasn't written it yet (e.g. right after Create, before the operator
	// has reconciled once). Not an error case.
	phase, _, err := unstructured.NestedString(obj.Object, "status", "phase")
	if err != nil {
		return spec, status, fmt.Errorf("reading status.phase: %w", err)
	}
	status.Phase = phase

	return spec, status, nil
}
