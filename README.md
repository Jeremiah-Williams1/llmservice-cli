# llmservice-cli

Cobra CLI for `LLMService` resources (`apps.jerremiah.dev/v1alpha1`),
served by the `llmservice-operator` reconciler in a separate repo.

Deliberately **not** importing the operator's Go types — uses
`k8s.io/client-go/dynamic` against the CRD's GVR instead, so this repo
has zero build-time dependency on the operator repo. See
`internal/llmservice/types.go` for the tradeoff that comes with that
choice (hand-maintained struct, no compile-time drift detection).

## Status

Bare scaffold. Every file with a `panic("TODO: not implemented")` or
`t.Skip(...)` is intentionally unfinished — being filled in one file at
a time.

## Layout

- `main.go` — entry point, just calls `cmd.Execute()`
- `cmd/` — Cobra commands: `root.go` (shared flags/client wiring),
  `deploy.go`, `status.go`, `rollback.go`
- `internal/k8s/` — dynamic client construction + the `LLMServiceGVR`
  constant (the one place CRD coordinates are hardcoded)
- `internal/llmservice/` — `types.go` (hand-mirrored Spec/Status
  structs) and `convert.go` (the bridge to/from
  `unstructured.Unstructured`)

## Rollback scope

Simple, one-level undo: `deploy` stores the previous spec as a JSON
annotation before applying a new one; `rollback` reads it back. Not a
revision history.
