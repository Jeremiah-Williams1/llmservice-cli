package cmd

import (
	"github.com/spf13/cobra"
)

// Simple rollback mechanism (agreed scope - not real multi-revision history):
// each `deploy` stores the PREVIOUS spec as a JSON blob in an annotation
// (e.g. "llmservice-cli/previous-spec") before applying the new one.
// `rollback` reads that single annotation and re-applies it as the current
// spec. This gives one level of undo, not a revision log.
//
// TODO (talk through before implementing):
//   - annotation key constant, e.g. const previousSpecAnnotation = "llmservice-cli/previous-spec"
//   - what happens on rollback of a resource that has never been deployed
//     via this CLI (annotation missing) - error out clearly, don't panic
//   - after rollback, should the annotation be cleared, or swapped so a
//     second rollback "redoes"? (simple version: probably just clear it -
//     no redo)

var rollbackCmd = &cobra.Command{
	Use:   "rollback [name]",
	Short: "Revert an LLMService to its previous spec (one level of undo)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO:
		// 1. name := args[0]
		// 2. obj, err := client.Resource(...).Namespace(namespace).Get(ctx, name, ...)
		// 3. read previousSpecAnnotation from obj.GetAnnotations()
		// 4. if missing -> return clear error, nothing to roll back to
		// 5. unmarshal annotation JSON into llmservice.Spec
		// 6. build new unstructured with that spec, Update() the resource
		panic("TODO: not implemented")
	},
}

func init() {
	// nothing to register yet - rollback only needs the shared --namespace flag
}
