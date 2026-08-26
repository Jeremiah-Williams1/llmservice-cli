package k8s

import (
	"fmt"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
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
// If kubeconfigPath is empty, falls back to the default location
// (~/.kube/config), same behavior as kubectl.
func NewDynamicClient(kubeconfigPath string) (dynamic.Interface, error) {
	resolvedPath := kubeconfigPath
	if resolvedPath == "" {
		home := homedir.HomeDir()
		if home == "" {
			return nil, fmt.Errorf("no --kubeconfig provided and could not determine home directory for default")
		}
		resolvedPath = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("loading kubeconfig from %s: %w", resolvedPath, err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("building dynamic client: %w", err)
	}

	return client, nil
}
