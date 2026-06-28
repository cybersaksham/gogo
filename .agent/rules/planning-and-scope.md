# Planning And Scope Rules

Use this rule when work maps to `.plans` or spans multiple packages.

## Plans Are Source Material

- Read the relevant `.plans/*.md` file before changing a planned feature area.
- Keep implementation aligned with package names, commands, tests, and acceptance criteria in the plan.
- If the plan and current code disagree, inspect current code and preserve compatibility unless explicitly asked to rework architecture.

## Scope Control

- Implement the requested task, not adjacent nice-to-have work.
- If a change exposes a concrete bug in the same path, fix it and verify it.
- If a larger redesign is needed, document the reason before making broad edits.

## Incremental Delivery

- Prefer small vertical increments: API, implementation, tests, docs.
- For cross-cutting changes, update the smallest shared contract first, then adapt callers.

