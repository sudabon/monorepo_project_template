---
name: openspec-adr-generator
description: generate adr markdown files from openspec proposal.md and design.md in a project repository. use when the user asks to create, extract, convert, or refresh architecture decision records from openspec changes, especially when the repo has openspec/changes and needs docs/adr or decs/adr output. this skill assumes it is invoked from the project root and should read local files, split design decisions into one adr per decision, and write the generated markdown files back into the repository.
---

# OpenSpec ADR Generator

This skill generates ADR markdown files from OpenSpec change documents stored under `openspec/changes/{name}`. It is designed for repository-local use from the project root and writes output to `docs/adr/{name}` by default, or `decs/adr/{name}` when that convention already exists or the user explicitly asks for it.

## Workflow

1. Confirm the repository root contains `openspec/changes`.
2. Determine the target change name.
   - If the user provided a change name, use it.
   - Otherwise, if exactly one change directory contains both `proposal.md` and `design.md`, use it automatically.
   - If multiple valid change directories exist, ask the user which change to use.
3. Run `scripts/generate_adr.py` from the project root.
4. Review the generated ADR files for obvious extraction errors.
5. Tell the user where the ADR files were written and summarize what was created.

## Commands

From the project root, use:

```bash
python /home/oai/skills/openspec-adr-generator/scripts/generate_adr.py --project-root . --change <change-name>
```

Optional flags:

```bash
python /home/oai/skills/openspec-adr-generator/scripts/generate_adr.py --project-root . --change <change-name> --with-summary-adr
python /home/oai/skills/openspec-adr-generator/scripts/generate_adr.py --project-root . --change <change-name> --overwrite
python /home/oai/skills/openspec-adr-generator/scripts/generate_adr.py --project-root . --change <change-name> --output-subdir decs/adr
```

If exactly one valid change exists, `--change` may be omitted.

## Output behavior

- Generate one ADR file per decision found in `design.md`.
- Name files with a zero-padded numeric prefix and a stable kebab-case slug.
- Use this default structure: `Context`, `Decision`, `Rationale`, `Alternatives Considered`, `Consequences`, `Risks / Trade-offs`, `Migration / Follow-up`, `Open Questions`, `References`.
- Keep the generated markdown editable by humans; do not over-compress or over-summarize.

## Review guidance

After running the script, quickly inspect the generated ADRs when any of these are true:

- `design.md` uses unusual headings or mixed languages
- a decision lacks `選択`, `理由`, or `代替案`
- multiple decisions produce similar filenames
- the user asked for higher-quality editorial cleanup

When review is needed, use `references/adr-template.md` and `references/extraction-rules.md` as the source of truth for structure and mapping.

## Notes

- Default to `docs/adr/{name}` for new repositories.
- If the repository already uses `decs/adr`, keep that convention instead of forcing `docs/adr`.
- Treat the generated ADRs as a first draft that is suitable for git review.
- Do not invent decisions that are not grounded in the source files.
