#!/usr/bin/env python3
"""Build and validate the explicit go-igraph upstream binding inventory."""

from __future__ import annotations

import argparse
import re
import sys
import tempfile
from collections import Counter
from dataclasses import dataclass
from pathlib import Path

from upstream_api import (
    Annotation,
    discover_annotations,
    discover_exported_go_declarations,
    discover_go_calls,
    discover_upstream_api,
    download_or_get_source,
    load_config,
    locate_include,
)


@dataclass(frozen=True)
class ConfiguredDisposition:
    status: str
    domain: str
    rationale: str
    go_api: str | None = None


CONFIGURED_DISPOSITION_STATUSES = {
    "composed": "Composed",
    "deferred": "Deferred",
    "intentionally_unsupported": "Intentionally unsupported",
}
GO_API_RE = re.compile(r"[A-Z][A-Za-z0-9_]*")


def configured_dispositions(config: dict) -> dict[str, ConfiguredDisposition]:
    raw = config.get("dispositions", {})
    if not isinstance(raw, dict):
        raise ValueError("invalid binding inventory:\n- dispositions must be an object")

    errors: list[str] = []
    unknown_statuses = sorted(set(raw) - set(CONFIGURED_DISPOSITION_STATUSES))
    errors.extend(f"unknown configured disposition {status}" for status in unknown_statuses)
    dispositions: dict[str, ConfiguredDisposition] = {}

    for status in CONFIGURED_DISPOSITION_STATUSES:
        entries = raw.get(status, {})
        if not isinstance(entries, dict):
            errors.append(f"configured disposition {status} must be an object")
            continue
        for name, metadata in entries.items():
            if name in dispositions:
                errors.append(f"overlapping configured disposition for {name}")
                continue
            if not isinstance(metadata, dict):
                errors.append(f"configured disposition metadata for {name} must be an object")
                continue
            domain = metadata.get("domain")
            rationale = metadata.get("rationale")
            go_api = metadata.get("go_api")
            if not isinstance(domain, str) or not domain.strip():
                errors.append(f"configured disposition for {name} requires a non-empty domain")
            if not isinstance(rationale, str) or not rationale.strip():
                errors.append(f"configured disposition for {name} requires a non-empty rationale")
            if status == "composed":
                if not isinstance(go_api, str) or not GO_API_RE.fullmatch(go_api):
                    errors.append(f"composed disposition for {name} requires an exported Go API")
            elif go_api is not None:
                errors.append(f"configured disposition {status} for {name} must not define a Go API")
            if (
                isinstance(domain, str)
                and domain.strip()
                and isinstance(rationale, str)
                and rationale.strip()
                and (status != "composed" or isinstance(go_api, str) and GO_API_RE.fullmatch(go_api))
                and (status == "composed" or go_api is None)
            ):
                dispositions[name] = ConfiguredDisposition(status, domain, rationale, go_api)

    if errors:
        raise ValueError("invalid binding inventory:\n- " + "\n- ".join(errors))
    return dispositions


def validate_inventory(
    declarations: dict[str, str],
    calls: set[str],
    annotations: list[Annotation],
    dispositions: dict[str, ConfiguredDisposition],
    exported_go: set[str],
) -> None:
    errors: list[str] = []
    known = set(declarations)
    for annotation in annotations:
        if annotation.upstream not in known:
            errors.append(f"{annotation.path}:{annotation.line}: unknown upstream symbol {annotation.upstream}")
        if annotation.kind == "bind" and not annotation.go_symbol[:1].isupper():
            errors.append(f"{annotation.path}:{annotation.line}: binding target {annotation.go_symbol} is not exported")
    by_upstream = Counter(item.upstream for item in annotations)
    errors.extend(f"duplicate annotation for {name}" for name, count in sorted(by_upstream.items()) if count > 1)
    errors.extend(f"unknown configured disposition symbol {name}" for name in sorted(set(dispositions) - known))
    annotated = {item.upstream for item in annotations}
    errors.extend(f"overlapping annotation and configured disposition for {name}" for name in sorted(annotated & set(dispositions)))
    unsupported = {name for name, item in dispositions.items() if item.status == "intentionally_unsupported"}
    errors.extend(
        f"composed disposition for {name} references missing Go API {item.go_api}"
        for name, item in sorted(dispositions.items())
        if item.status == "composed" and item.go_api not in exported_go
    )
    classified = annotated | unsupported
    errors.extend(f"unclassified production C call {name}" for name in sorted(calls - classified))
    if errors:
        raise ValueError("invalid binding inventory:\n- " + "\n- ".join(errors))


def render(config: dict, declarations: dict[str, str], annotations: list[Annotation]) -> str:
    bindings = {item.upstream: item for item in annotations if item.kind == "bind"}
    internal = {item.upstream: item for item in annotations if item.kind == "internal"}
    dispositions = configured_dispositions(config)
    composed = {name: item for name, item in dispositions.items() if item.status == "composed"}
    deferred = {name for name, item in dispositions.items() if item.status == "deferred"}
    unsupported = {name for name, item in dispositions.items() if item.status == "intentionally_unsupported"}
    total = len(declarations)
    percent = 100 * len(bindings) / total if total else 0
    by_header = Counter(declarations.values())
    bound_by_header = Counter(declarations[name] for name in bindings)
    lines = [
        "# Upstream igraph API coverage", "",
        "> Generated by `tools/api_coverage.py`; do not edit by hand.", "",
        f"- Upstream: [{config['upstream']} {config['version']}]({config['release_url']})",
        f"- User-facing bindings: **{len(bindings)} / {total} ({percent:.2f}%)**",
        f"- Internal dependencies: **{len(internal)}**",
        f"- Composed APIs: **{len(composed)}**",
        f"- Intentionally unsupported: **{len(unsupported)}**",
        f"- Deferred declarations: **{len(deferred)}**", "",
        "User-facing and internal coverage is based on explicit source annotations. Composed,",
        "deferred, and intentionally unsupported declarations are configured with a domain and",
        "rationale in `tools/api_coverage_config.json`. `Missing` therefore means unreviewed, not",
        "merely unbound.", "",
        "Headers marked `(generated)` are declaration-discovery locations. PMT-generated",
        "APIs can appear through several public headers, so the report records the first",
        "preprocessed header where each declaration is found rather than a canonical owner.", "",
        "## Summary by header", "", "| Header | Bound | Total | Coverage |",
        "| --- | ---: | ---: | ---: |",
    ]
    for header in sorted(by_header):
        count, header_total = bound_by_header[header], by_header[header]
        lines.append(f"| `{header}` | {count} | {header_total} | {100 * count / header_total:.2f}% |")
    lines += ["", "## Function inventory", "", "| Function | Header | Status | Go API |", "| --- | --- | --- | --- |"]
    for name in sorted(declarations):
        if name in bindings:
            status, owner = "User-facing", f"`{bindings[name].go_symbol}`"
        elif name in internal:
            status, owner = "Internal", f"`{internal[name].go_symbol}`"
        elif name in composed:
            status, owner = CONFIGURED_DISPOSITION_STATUSES[composed[name].status], f"`{composed[name].go_api}`"
        elif name in unsupported:
            status, owner = "Intentionally unsupported", "—"
        elif name in deferred:
            status, owner = "Deferred", "—"
        else:
            status, owner = "Missing", "—"
        lines.append(f"| `{name}` | `{declarations[name]}` | {status} | {owner} |")
    lines.append("")
    return "\n".join(lines)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--config", type=Path, default=Path("tools/api_coverage_config.json"))
    parser.add_argument("--source-dir", type=Path, help="local igraph source tree (offline mode)")
    parser.add_argument("--output", type=Path, default=Path("docs/api-coverage.md"))
    parser.add_argument("--check", action="store_true", help="fail if the generated report differs")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    repo = Path(__file__).resolve().parents[1]
    config = load_config(args.config)
    output = args.output if args.output.is_absolute() else repo / args.output
    calls = discover_go_calls(repo)
    annotations = discover_annotations(repo)
    exported_go = discover_exported_go_declarations(repo)
    if args.source_dir:
        source = args.source_dir
    else:
        source = download_or_get_source(config)
    declarations = discover_upstream_api(locate_include(source))
    try:
        dispositions = configured_dispositions(config)
        validate_inventory(declarations, calls, annotations, dispositions, exported_go)
    except ValueError as error:
        print(error, file=sys.stderr)
        return 1
    report = render(config, declarations, annotations)
    if args.check:
        if not output.exists() or output.read_text(encoding="utf-8") != report:
            print(f"{output} is out of date; run make coverage", file=sys.stderr)
            return 1
        return 0
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(report, encoding="utf-8")
    print(f"wrote {output}: {sum(a.kind == 'bind' for a in annotations)}/{len(declarations)} user-facing bindings")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
