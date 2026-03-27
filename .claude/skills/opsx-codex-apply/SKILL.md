---
name: opsx-codex-apply
description: orchestrate openspec apply-style implementation from claude code by delegating coding work to codex, then run repository checks and continue to claude review. use when a repository uses openspec and the user wants a claude code command such as /opsx-codex-apply to hand implementation to codex instead of coding directly in claude.
---

# Opsx Codex Apply

Use this skill as a **Claude Code orchestration command**, not as the primary implementer.

## Workflow

Follow this sequence:

1. Identify the active OpenSpec change and the relevant artifacts.
2. Read the current proposal, spec, and tasks before doing anything else.
3. Build a concise Codex handoff prompt from those artifacts.
4. Ask the runtime to execute the repository's Codex wrapper command if one exists.
5. After Codex finishes, run the repository checks in the expected order.
6. If checks fail, hand the failures back to Codex for repair.
7. When checks pass, perform a Claude review of the resulting diff.
8. If review fails, send the review findings back to Codex for another repair pass.
9. Stop only when checks pass and review is acceptable, or when the user explicitly asks to stop.

## Required behavior

- **Do not implement the feature directly in Claude Code** unless the user explicitly overrides this workflow.
- Prefer the repository's existing wrapper script if one exists, such as:
  - `./ai/opsx-codex-apply.sh`
  - `./scripts/opsx-codex-apply.sh`
  - `make opsx-codex-apply`
- If no wrapper exists, construct a Codex prompt and instruct the runtime to run `codex exec` with that prompt.
- Keep the Codex handoff short, concrete, and diff-oriented.
- Preserve the OpenSpec intent. Do not expand scope.
- Run formatting, linting, and tests after each Codex implementation pass.
- Treat failing checks as input for the next Codex repair pass.
- Perform Claude review only after the checks pass.
- Review for architecture fit, regression risk, spec compliance, and missing edge cases.
- If review is not acceptable, summarize the findings into a compact repair brief for Codex.
- Keep looping until one of these is true:
  - checks pass and review passes
  - the configured retry budget is exhausted
  - the user explicitly stops the workflow

## Inputs to gather

Read whichever of these exist in the repository for the active change:

- `openspec/changes/<change-id>/proposal.md`
- `openspec/changes/<change-id>/spec.md`
- `openspec/changes/<change-id>/tasks.md`

If the repository uses a different OpenSpec layout, adapt to that layout instead of forcing these exact paths.

## Codex handoff format

Use this structure when preparing the Codex prompt:

### Goal
One short paragraph stating what to implement.

### Context
- active change id
- relevant files or modules
- architecture constraints
- public API constraints

### Tasks
A short numbered list derived from OpenSpec tasks.

### Done when
- implementation matches the active spec
- formatting passes
- lint passes
- tests pass
- changes remain scoped to this work

## Review output format

When Claude review finds issues, summarize them like this before sending them back to Codex:

- **issue**
  - file or area
  - why it matters
  - concrete fix direction

Keep the repair brief short enough that Codex can act on it directly.

## Repository command preferences

When the repository defines canonical commands, prefer those over guessing. Typical preference order:

1. project wrapper command for codex apply
2. project formatter command
3. project linter command
4. project test command

If commands are ambiguous, inspect repository files such as `package.json`, `Makefile`, `justfile`, `pyproject.toml`, or CI workflows to infer the standard commands.

## References

- Use [runbook.md](references/runbook.md) for the detailed orchestration pattern.
- Use [examples.md](references/examples.md) for prompt and review examples.
