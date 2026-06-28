# Development Workflow Rules

Use this rule for normal implementation work.

## Branch And Worktree

- Stay on the current branch.
- Do not create, switch, or reset branches unless explicitly requested.
- Do not create a worktree unless explicitly requested.

## Editing

- Inspect the existing code path before editing.
- Keep edits narrow and compatible with existing patterns.
- Do not perform unrelated refactors.
- Do not delete or rewrite user changes.
- Use deterministic output for generated files.

## Commits

- Commit after small completed changes when asked to develop a larger feature.
- Use imperative commit messages with a capital first letter and no `feat:`, `docs:`, or AI prefix.
- Do not commit broken intermediate states if tests for that step fail.

## Dependencies

- Prefer the standard library.
- If adding a dependency, update `go.mod`, `go.sum`, relevant docs, and dependency policy context.
- Run `go mod tidy` only when dependency changes require it.

