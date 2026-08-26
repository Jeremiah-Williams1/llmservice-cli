package k8s

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// LLMServiceGVR is the single place in this repo that knows the CRD's
// exact API coordinates. If the operator's group/version ever changes,
// this is the only line that needs to change.
var LLMServiceGVR = schema.GroupVersionResource{
	Group:    "apps.jerremiah.dev",
	Version:  "v1alpha1",
	Resource: "llmservices",
}

// NewDynamicClient builds a dynamic.Interface from a kubeconfig path.
// TODO:
//  1. clientcmd.BuildConfigFromFlags("", kubeconfigPath) -> *rest.Config
//  2. dynamic.NewForConfig(config) -> dynamic.Interface
func NewDynamicClient(kubeconfigPath string) (dynamic.Interface, error) {
	panic("TODO: not implemented")
}
