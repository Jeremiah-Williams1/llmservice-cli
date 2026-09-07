package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"jerremiah.dev/llmservice-cli/internal/k8s"
	"jerremiah.dev/llmservice-cli/internal/llmservice"
)

// Simple rollback mechanism: each `deploy` stores the PREVIOUS spec as a
// JSON blob in the previousSpecAnnotation (defined in deploy.go) before
// applying a new one. `rollback` reads that single annotation and
// re-applies it as the current spec. This gives one level of undo, not a
// revision log - a second rollback in a row will error, since the first
// rollback clears the annotation rather than swapping it (no redo).

var rollbackCmd = &cobra.Command{
	Use:   "rollback [name]",
	Short: "Revert an LLMService to its previous spec (one level of undo)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		ctx := context.Background()
		resourceClient := dynClient.Resource(k8s.LLMServiceGVR).Namespace(namespace)

		existing, err := resourceClient.Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("%s not found in namespace %s", name, namespace)
		}
		if err != nil {
			return fmt.Errorf("getting %s: %w", name, err)
		}

		annotations := existing.GetAnnotations()
		previousSpecJSON, ok := annotations[previousSpecAnnotation]
		if !ok || previousSpecJSON == "" {
			return fmt.Errorf(
				"%s has no previous spec recorded - nothing to roll back to "+
					"(rollback only works after at least one `deploy` update, "+
					"not on the first deploy)", name)
		}

		var previousSpec llmservice.Spec
		if err := json.Unmarshal([]byte(previousSpecJSON), &previousSpec); err != nil {
			return fmt.Errorf("decoding stored previous spec for %s: %w", name, err)
		}

		// Same conflict-retry reasoning as deploy.go: re-fetch and rebuild
		// on each attempt so the resourceVersion sent is always current.
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			latest, err := resourceClient.Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			obj := llmservice.ToUnstructured(name, namespace, previousSpec)
			obj.SetResourceVersion(latest.GetResourceVersion())
			// Clear the annotation on rollback rather than carrying it
			// forward - this is one level of undo, not a redo stack.
			obj.SetAnnotations(map[string]string{
				previousSpecAnnotation: "",
			})

			_, err = resourceClient.Update(ctx, obj, metav1.UpdateOptions{})
			return err
		})
		if err != nil {
			return fmt.Errorf("rolling back %s: %w", name, err)
		}
		fmt.Printf("rolled back %s\n", name)
		return nil
	},
}

func init() {
	// nothing to register yet - rollback only needs the shared --namespace flag
}
