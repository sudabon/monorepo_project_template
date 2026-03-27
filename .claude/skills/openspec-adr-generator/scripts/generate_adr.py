#!/usr/bin/env python3
from __future__ import annotations

import argparse
import dataclasses
import os
import re
import sys
from pathlib import Path
from typing import Iterable

SECTION_RE = re.compile(r'^(#{1,6})\s+(.+?)\s*$')


@dataclasses.dataclass
class Decision:
    number: int
    title: str
    selected_option: str
    rationale: list[str]
    alternatives: list[str]


@dataclasses.dataclass
class Sections:
    proposal_why: str = ""
    proposal_impact: str = ""
    design_context: str = ""
    design_goals: str = ""
    design_risks: str = ""
    design_migration: str = ""
    design_open_questions: str = ""
    decisions: list[Decision] = dataclasses.field(default_factory=list)


def read_text(path: Path) -> str:
    return path.read_text(encoding='utf-8')


def normalize_heading(text: str) -> str:
    return re.sub(r'\s+', ' ', text.strip().lower())


def extract_section(markdown: str, heading_names: Iterable[str]) -> str:
    wanted = {normalize_heading(h) for h in heading_names}
    lines = markdown.splitlines()
    start = None
    level = None
    for i, line in enumerate(lines):
        m = SECTION_RE.match(line)
        if not m:
            continue
        hdr_level = len(m.group(1))
        title = normalize_heading(m.group(2))
        if title in wanted:
            start = i + 1
            level = hdr_level
            break
    if start is None:
        return ""
    end = len(lines)
    for i in range(start, len(lines)):
        m = SECTION_RE.match(lines[i])
        if m and len(m.group(1)) <= level:
            end = i
            break
    return "\n".join(lines[start:end]).strip()


DECISION_HEADER_RE = re.compile(r'^#{3,6}\s*decision\s+(\d+)\s*:\s*(.+?)\s*$', re.IGNORECASE | re.MULTILINE)


def extract_field_values(block: str, label: str) -> list[str]:
    pattern = re.compile(rf'^\*\*{re.escape(label)}\*\*\s*:\s*(.+?)\s*$', re.IGNORECASE | re.MULTILINE)
    return [m.group(1).strip() for m in pattern.finditer(block)]


def split_bullets(text: str) -> list[str]:
    items: list[str] = []
    for raw in text.splitlines():
        line = raw.strip()
        if not line:
            continue
        line = re.sub(r'^[-*]\s+', '', line)
        items.append(line)
    if not items and text.strip():
        items.append(text.strip())
    return items


RISK_ITEM_RE = re.compile(r'^-\s+\*\*\[(.+?)\]\*\*\s*(.+)$', re.MULTILINE)


def risk_keywords_for(decision: Decision) -> set[str]:
    blob = f"{decision.title} {decision.selected_option} {' '.join(decision.alternatives)} {' '.join(decision.rationale)}".lower()
    pairs = {
        'latency': {'latency', 'low-latency', '遅延', 'レイテンシ', 'streaming', 'stt', 'tts'},
        'webrtc': {'webrtc', 'aiortc', 'janus', 'mediasoup', 'media server', 'media'},
        'llm': {'gpt', 'openai', 'claude', 'llm', 'api'},
        'rag': {'rag', 'faiss', 'langchain', 'retriever', 'vector'},
        'database': {'postgres', 'postgresql', 'mysql', 'sqlalchemy', 'asyncpg', 'database', 'db'},
        'frontend': {'frontend', 'client', 'dashboard', 'spa', 'react'},
        'websocket': {'websocket', 'sse', 'signaling', 'push'},
        'infra': {'cloud run', 'container', 'single instance', 'single-inst', 'availability'},
    }
    matched: set[str] = set()
    for key, words in pairs.items():
        if any(word in blob for word in words):
            matched.add(key)
    return matched


def select_related_risks(risks_text: str, decision: Decision) -> list[str]:
    wanted = risk_keywords_for(decision)
    related: list[str] = []
    for kind, content in RISK_ITEM_RE.findall(risks_text):
        kind_norm = kind.strip().lower()
        if kind_norm in wanted or any(token in content.lower() for token in wanted):
            related.append(f"**[{kind}]** {content.strip()}")
    return related


def select_related_open_questions(open_text: str, decision: Decision) -> list[str]:
    lines = [ln.strip() for ln in open_text.splitlines() if ln.strip().startswith('- ')]
    blob = f"{decision.title} {decision.selected_option} {' '.join(decision.alternatives)}".lower()
    keywords = set(re.findall(r'[a-z][a-z0-9\-+/.]{2,}', blob))
    selected: list[str] = []
    for line in lines:
        if any(k in line.lower() for k in keywords):
            selected.append(re.sub(r'^-\s+', '', line))
    return selected


def parse_decisions(decisions_markdown: str) -> list[Decision]:
    matches = list(DECISION_HEADER_RE.finditer(decisions_markdown))
    decisions: list[Decision] = []
    for idx, match in enumerate(matches):
        start = match.end()
        end = matches[idx + 1].start() if idx + 1 < len(matches) else len(decisions_markdown)
        block = decisions_markdown[start:end].strip()
        selected = extract_field_values(block, '選択')
        rationale = extract_field_values(block, '理由')
        alternatives = extract_field_values(block, '代替案')
        decisions.append(
            Decision(
                number=int(match.group(1)),
                title=match.group(2).strip(),
                selected_option=selected[0] if selected else '',
                rationale=split_bullets(rationale[0] if rationale else ''),
                alternatives=split_bullets(alternatives[0] if alternatives else ''),
            )
        )
    return decisions


def slugify(text: str) -> str:
    manual = {
        normalize_heading('音声処理パイプライン'): 'sequential-voice-pipeline',
        normalize_heading('WebRTC実装方式'): 'backend-webrtc-with-aiortc',
        normalize_heading('LLMの選択'): 'gpt-4o-mini-as-llm',
        normalize_heading('ナレッジ検索（RAG）の構成'): 'langchain-faiss-for-rag',
        normalize_heading('フロントエンド・クライアントの分離'): 'separate-frontend-and-client',
        normalize_heading('データベース設計方針'): 'postgresql-async-stack',
        normalize_heading('リアルタイム通知方式'): 'websocket-for-realtime-events',
    }
    key = normalize_heading(text)
    for source, slug in manual.items():
        if source in key:
            return slug
    text = text.lower()
    text = text.replace('gpt-4o-mini', 'gpt-4o-mini')
    text = re.sub(r'[^a-z0-9]+', '-', text)
    text = re.sub(r'-+', '-', text).strip('-')
    return text or 'adr'


def summarize_context(sections: Sections) -> str:
    parts = [p.strip() for p in [sections.proposal_why, sections.design_context] if p.strip()]
    return '\n\n'.join(parts)


def render_bullets(items: list[str]) -> str:
    if not items:
        return '- None noted from source documents.'
    return '\n'.join(f'- {item}' for item in items)


def render_adr(change_name: str, date_text: str, sections: Sections, decision: Decision) -> str:
    related_risks = select_related_risks(sections.design_risks, decision)
    related_questions = select_related_open_questions(sections.design_open_questions, decision)

    consequences: list[str] = []
    for bullet in decision.rationale:
        consequences.append(bullet)
    if decision.selected_option:
        consequences.insert(0, f"This ADR standardizes the project on: {decision.selected_option}")

    return f"""# ADR-{decision.number:03d}: {decision.title}\n\n- Status: Accepted\n- Date: {date_text}\n- Source Change: {change_name}\n\n## Context\n\n{summarize_context(sections) or 'Context could not be extracted automatically. Review proposal.md and design.md.'}\n\n## Decision\n\n{decision.selected_option or 'Decision could not be extracted automatically.'}\n\n## Rationale\n\n{render_bullets(decision.rationale)}\n\n## Alternatives Considered\n\n{render_bullets(decision.alternatives)}\n\n## Consequences\n\n{render_bullets(consequences)}\n\n## Risks / Trade-offs\n\n{render_bullets(related_risks)}\n\n## Migration / Follow-up\n\n{sections.design_migration.strip() or 'No migration or follow-up details were extracted.'}\n\n## Open Questions\n\n{render_bullets(related_questions)}\n\n## References\n\n- /openspec/changes/{change_name}/proposal.md\n- /openspec/changes/{change_name}/design.md\n"""


def choose_output_root(project_root: Path, requested: str | None) -> Path:
    if requested:
        return (project_root / requested).resolve()
    decs = project_root / 'decs' / 'adr'
    docs = project_root / 'docs' / 'adr'
    if decs.exists() and not docs.exists():
        return decs.resolve()
    return docs.resolve()


def discover_change_root(project_root: Path, change_name: str | None) -> tuple[str, Path]:
    base = project_root / 'openspec' / 'changes'
    if not base.exists():
        raise FileNotFoundError(f'Missing openspec changes directory: {base}')
    if change_name:
        target = base / change_name
        if not target.exists():
            raise FileNotFoundError(f'Change directory not found: {target}')
        return change_name, target

    candidates: list[Path] = []
    for child in sorted(base.iterdir()):
        if child.is_dir() and (child / 'proposal.md').exists() and (child / 'design.md').exists():
            candidates.append(child)
    if len(candidates) == 1:
        return candidates[0].name, candidates[0]
    names = ', '.join(c.name for c in candidates) if candidates else '(none found)'
    raise RuntimeError(f'Unable to infer change name automatically. Pass --change. Candidates: {names}')


def load_sections(change_root: Path) -> Sections:
    proposal = read_text(change_root / 'proposal.md')
    design = read_text(change_root / 'design.md')
    decisions_text = extract_section(design, ['Decisions'])
    return Sections(
        proposal_why=extract_section(proposal, ['Why']),
        proposal_impact=extract_section(proposal, ['Impact']),
        design_context=extract_section(design, ['Context']),
        design_goals=extract_section(design, ['Goals / Non-Goals', 'Goals/Non-Goals']),
        design_risks=extract_section(design, ['Risks / Trade-offs', 'Risks/Trade-offs']),
        design_migration=extract_section(design, ['Migration Plan']),
        design_open_questions=extract_section(design, ['Open Questions']),
        decisions=parse_decisions(decisions_text),
    )


def build_summary_adr(change_name: str, date_text: str, sections: Sections) -> str:
    listing = '\n'.join(f'- ADR-{d.number:03d}: {d.title}' for d in sections.decisions) or '- No child ADRs extracted.'
    impact = sections.proposal_impact.strip() or 'Impact section could not be extracted automatically.'
    goals = sections.design_goals.strip() or 'Goals / Non-Goals section could not be extracted automatically.'
    return f"""# ADR-000: overall architecture principles for {change_name}\n\n- Status: Accepted\n- Date: {date_text}\n- Source Change: {change_name}\n\n## Context\n\n{summarize_context(sections) or 'Context could not be extracted automatically.'}\n\n## Decision\n\nThis change is implemented as a set of focused ADRs derived from the OpenSpec proposal and design documents.\n\n## Decision Drivers\n\n{goals}\n\n## Scope / Impact\n\n{impact}\n\n## Child ADRs\n\n{listing}\n\n## References\n\n- /openspec/changes/{change_name}/proposal.md\n- /openspec/changes/{change_name}/design.md\n"""


def write_file(path: Path, text: str, overwrite: bool) -> str:
    if path.exists() and not overwrite:
        return 'skipped'
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text, encoding='utf-8', newline='\n')
    return 'written'


def main() -> int:
    parser = argparse.ArgumentParser(description='Generate ADR markdown files from OpenSpec proposal.md and design.md.')
    parser.add_argument('--project-root', default='.', help='Project root that contains openspec/changes.')
    parser.add_argument('--change', help='Change name under openspec/changes. If omitted, autodetect when exactly one exists.')
    parser.add_argument('--output-subdir', help='Output directory relative to project root, e.g. docs/adr or decs/adr.')
    parser.add_argument('--overwrite', action='store_true', help='Overwrite existing ADR files.')
    parser.add_argument('--with-summary-adr', action='store_true', help='Also create ADR-000 summary file.')
    parser.add_argument('--dry-run', action='store_true', help='Print target files without writing.')
    args = parser.parse_args()

    project_root = Path(args.project_root).resolve()
    try:
        change_name, change_root = discover_change_root(project_root, args.change)
        sections = load_sections(change_root)
    except (FileNotFoundError, RuntimeError) as exc:
        print(f'ERROR: {exc}', file=sys.stderr)
        return 1

    if not sections.decisions:
        print('WARNING: no decisions were extracted from design.md', file=sys.stderr)
        return 0

    output_root = choose_output_root(project_root, args.output_subdir) / change_name
    date_text = os.environ.get('ADR_DATE', '1970-01-01')
    if date_text == '1970-01-01':
        from datetime import date
        date_text = date.today().isoformat()

    planned: list[tuple[Path, str]] = []
    if args.with_summary_adr:
        planned.append((output_root / '000-overall-architecture-principles.md', build_summary_adr(change_name, date_text, sections)))
    used_names: set[str] = set()
    for decision in sections.decisions:
        base = f"{decision.number:03d}-{slugify(decision.title)}.md"
        if base in used_names:
            stem = base[:-3]
            n = 2
            while f'{stem}-{n}.md' in used_names:
                n += 1
            base = f'{stem}-{n}.md'
        used_names.add(base)
        planned.append((output_root / base, render_adr(change_name, date_text, sections, decision)))

    if args.dry_run:
        for path, _ in planned:
            print(path)
        return 0

    written = skipped = 0
    for path, body in planned:
        result = write_file(path, body, overwrite=args.overwrite)
        if result == 'written':
            written += 1
            print(f'WROTE {path}')
        else:
            skipped += 1
            print(f'SKIPPED {path}')

    print(f'SUMMARY change={change_name} decisions={len(sections.decisions)} written={written} skipped={skipped} output_root={output_root}')
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
