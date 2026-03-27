# Runbook

## Purpose

This skill lets Claude Code behave like an orchestrator for an OpenSpec implementation flow where **Codex writes code** and **Claude reviews it**.

## Recommended control loop

1. Resolve the active OpenSpec change.
2. Read `proposal.md`, `spec.md`, and `tasks.md`.
3. Produce a compact Codex brief.
4. Start Codex through the repository wrapper or `codex exec`.
5. Run the project's formatter.
6. Run the project's linter.
7. Run the project's tests.
8. If checks fail, summarize failures and send them back to Codex.
9. If checks pass, perform Claude review.
10. If review fails, summarize findings and send them back to Codex.
11. End only when all gates are green.

## Wrapper-first strategy

Prefer a repository wrapper because it gives one stable entrypoint for:

- `codex exec`
- check commands
- retry budgeting
- state files
- structured logs

Suggested wrappers:

- `./ai/opsx-codex-apply.sh`
- `./scripts/opsx-codex-apply.sh`
- `make opsx-codex-apply`

## Minimal Codex brief template

```text
Use the repository's active OpenSpec change to implement the approved work.

Goal:
[one paragraph]

Context:
- change id: [id]
- files/modules: [items]
- constraints: [items]

Tasks:
1. ...
2. ...
3. ...

Done when:
- spec intent is implemented
- formatting passes
- lint passes
- tests pass
- scope stays minimal
```

## Failure brief template

```text
Repair the current implementation. Keep scope minimal.

Problems to fix:
- formatter failed in [file]: [reason]
- lint failed in [file]: [reason]
- tests failed in [suite]: [reason]

Do not redesign unrelated areas. Make the smallest set of changes needed to pass all checks.
```

## Claude review checklist

Review after checks pass:

- Does the diff satisfy the OpenSpec intent?
- Are any tasks skipped or only partially implemented?
- Are there edge cases or regressions?
- Is the architecture still aligned with existing patterns?
- Did the change widen public API or scope unnecessarily?
- Are tests meaningful instead of superficial?

## Stop conditions

Stop the loop when:

- checks pass and review passes
- the retry budget is exhausted
- the user explicitly asks to stop

When stopping due to retries, report:
- what passed
- what failed
- the next best Codex repair brief
