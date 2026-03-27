# Examples

## Example user intent

> Run /opsx-codex-apply for the active OpenSpec change.

## Example orchestration summary

1. Read the active change artifacts.
2. Build a Codex prompt from the spec and tasks.
3. Run the repository wrapper for Codex apply.
4. Run format, lint, and test.
5. Send failures back to Codex if needed.
6. Run Claude review after checks pass.

## Example Codex brief

```text
Implement the active OpenSpec change.

Goal:
Add the approved repository changes for the active spec without expanding scope.

Context:
- change id: improve-auth-retries
- affected areas: auth service, retry policy, unit tests
- constraints: keep public API stable, preserve existing logging pattern

Tasks:
1. Add retry policy support in the auth service.
2. Update integration points to use the new policy.
3. Add tests for success, failure, and max retry behavior.

Done when:
- implementation matches the active spec
- formatting passes
- lint passes
- tests pass
- changes remain scoped to this work
```

## Example review repair brief

```text
Repair the implementation based on Claude review findings.

Problems to fix:
- auth service mixes transport errors with domain errors; preserve the existing domain error boundary
- missing test coverage for zero retry configuration
- one helper name is inconsistent with adjacent modules

Keep the change scoped to the active OpenSpec work and do not redesign unrelated code.
```
