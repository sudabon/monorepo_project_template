# Extraction rules

Use these rules when generating ADRs from OpenSpec documents.

## Inputs
- `openspec/changes/{name}/proposal.md`
- `openspec/changes/{name}/design.md`

## Section mapping
- `proposal.md` / `Why` -> ADR `Context`
- `design.md` / `Context` -> ADR `Context`
- `design.md` / `Decisions` -> ADR `Decision`, `Rationale`, `Alternatives Considered`
- `design.md` / `Risks / Trade-offs` -> ADR `Risks / Trade-offs`
- `design.md` / `Migration Plan` -> ADR `Migration / Follow-up`
- `design.md` / `Open Questions` -> ADR `Open Questions`

## Decision granularity
- Generate one ADR per decision inside `design.md`.
- Prefer a separate summary ADR only when the user asks for it.
- Do not force feature-spec details into ADRs when they are not architectural decisions.

## Output rules
- Default output root is `docs/adr/{name}`.
- If the repository already uses `decs/adr`, it is acceptable to write there instead.
- Preserve Markdown and UTF-8.
- Keep files deterministic and easy to review in git.
