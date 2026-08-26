package llmservice

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ToUnstructured builds the object the dynamic client can Create()/Update().
// This is where Spec's clean Go fields get flattened into the
// map[string]interface{} shape the K8s API actually expects on the wire.
//
// TODO: build and return something like:
//
//	&unstructured.Unstructured{
//	    Object: map[string]interface{}{
//	        "apiVersion": "apps.jerremiah.dev/v1alpha1",
//	        "kind":       "LLMService",
//	        "metadata": map[string]interface{}{
//	            "name":      name,
//	            "namespace": namespace,
//	        },
//	        "spec": map[string]interface{}{
//	            "model":    spec.Model,
//	            "gpuCount": int64(spec.GPUCount), // note: int64, not int32 - unstructured wants int64
//	            // ... remaining fields
//	        },
//	    },
//	}
func ToUnstructured(name, namespace string, spec Spec) *unstructured.Unstructured {
	panic("TODO: not implemented")
}

// FromUnstructured does the reverse - pulls spec/status fields back out of
// what Get() returns, for `status` to print. Use unstructured.NestedString /
// NestedInt64 / etc helpers rather than raw map indexing - they handle
// missing-field cases without panicking.
func FromUnstructured(obj *unstructured.Unstructured) (Spec, Status, error) {
	panic("TODO: not implemented")
}
