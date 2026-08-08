#!/usr/bin/env python3
"""Build and validate the explicit go-igraph upstream binding inventory."""

from __future__ import annotations

import argparse
import json
import re
import os
import subprocess
import sys
import tarfile
import tempfile
import urllib.request
from collections import Counter
from dataclasses import dataclass
from pathlib import Path

EXPORT_RE = re.compile(
    r"^\s*(?:(?:IGRAPH_EXPERIMENTAL|IGRAPH_DEPRECATED)\s+)?IGRAPH_EXPORT"
    r"(?:(?!;).)*?\b(igraph_[A-Za-z0-9_]+)\s*\(",
    re.MULTILINE | re.DOTALL,
)
CALL_RE = re.compile(r"\bC\.(igraph_[A-Za-z0-9_]+)\s*\(")
ANNOTATION_RE = re.compile(r"^\s*//igraph:(bind|internal)\s+(igraph_[A-Za-z0-9_]+)\s*$")
DECL_RE = re.compile(r"^\s*(?:func|type|var|const)\s+(?:\([^)]*\)\s*)?([A-Za-z_][A-Za-z0-9_]*)")


@dataclass(frozen=True)
class Annotation:
    kind: str
    upstream: str
    go_symbol: str
    path: str
    line: int


def strip_comments(text: str) -> str:
    return re.sub(r"/\*.*?\*/|//[^\n]*", "", text, flags=re.DOTALL)


def discover_upstream_api(include_dir: Path) -> dict[str, str]:
    declarations: dict[str, str] = {}
    for header in sorted(include_dir.rglob("*.h")):
        text = strip_comments(header.read_text(encoding="utf-8", errors="replace"))
        for match in EXPORT_RE.finditer(text):
            declarations.setdefault(match.group(1), header.relative_to(include_dir).as_posix())
    # Several public API families (notably vectors) are instantiated from PMT
    # headers. Parse the preprocessed umbrella header so those declarations are
    # part of the inventory too; direct declarations above retain their more
    # precise source header.
    umbrella = include_dir / "igraph.h"
    if umbrella.exists():
        command = [os.environ.get("CC", "cc"), "-E", "-P", "-I", str(include_dir), str(umbrella)]
        try:
            process = subprocess.run(command, check=False, capture_output=True, text=True)
            expanded = process.stdout
            if not expanded:
                raise ValueError(process.stderr.strip() or "preprocessor produced no output")
        except (FileNotFoundError, ValueError) as error:
            raise ValueError(f"could not preprocess {umbrella}: {error}") from error
        for match in EXPORT_RE.finditer(expanded):
            declarations.setdefault(match.group(1), "igraph.h (generated)")
    return declarations


def production_go_files(repo: Path) -> list[Path]:
    return sorted(
        path for path in repo.rglob("*.go")
        if not path.name.endswith("_test.go") and ".git" not in path.parts
    )


def discover_go_calls(repo: Path) -> set[str]:
    calls: set[str] = set()
    for path in production_go_files(repo):
        calls.update(name for name in CALL_RE.findall(path.read_text(encoding="utf-8")) if not name.endswith("_t"))
    return calls


def discover_annotations(repo: Path) -> list[Annotation]:
    annotations: list[Annotation] = []
    for path in production_go_files(repo):
        pending: list[tuple[str, str, int]] = []
        for number, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
            match = ANNOTATION_RE.match(line)
            if match:
                pending.append((match.group(1), match.group(2), number))
                continue
            declaration = DECL_RE.match(line)
            if declaration and pending:
                symbol = declaration.group(1)
                annotations.extend(
                    Annotation(kind, upstream, symbol, path.relative_to(repo).as_posix(), source_line)
                    for kind, upstream, source_line in pending
                )
                pending = []
                continue
            if pending and line.strip() and not line.lstrip().startswith("//"):
                raise ValueError(f"{path}:{pending[0][2]}: annotation is not attached to a Go declaration")
        if pending:
            raise ValueError(f"{path}:{pending[0][2]}: annotation is not attached to a Go declaration")
    return annotations


def validate_inventory(
    declarations: dict[str, str], calls: set[str], annotations: list[Annotation], unsupported: set[str]
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
    errors.extend(f"unknown intentionally unsupported symbol {name}" for name in sorted(unsupported - known))
    classified = {item.upstream for item in annotations} | unsupported
    errors.extend(f"unclassified production C call {name}" for name in sorted(calls - classified))
    if errors:
        raise ValueError("invalid binding inventory:\n- " + "\n- ".join(errors))


def locate_include(source: Path) -> Path:
    direct = source / "include"
    if direct.is_dir():
        return direct
    candidates = list(source.glob("*/include"))
    if len(candidates) == 1:
        return candidates[0]
    raise ValueError(f"could not locate a unique include directory below {source}")


def download_source(url: str, destination: Path) -> Path:
    archive = destination / "upstream.tar.gz"
    urllib.request.urlretrieve(url, archive)
    with tarfile.open(archive, "r:gz") as bundle:
        root = destination.resolve()
        for member in bundle.getmembers():
            target = (destination / member.name).resolve()
            if root not in target.parents and target != root:
                raise ValueError("unsafe path in upstream archive")
        bundle.extractall(destination, filter="data")
    return destination


def render(config: dict, declarations: dict[str, str], annotations: list[Annotation]) -> str:
    bindings = {item.upstream: item for item in annotations if item.kind == "bind"}
    internal = {item.upstream: item for item in annotations if item.kind == "internal"}
    unsupported = set(config.get("intentionally_unsupported", []))
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
        f"- Intentionally unsupported: **{len(unsupported)}**", "",
        "Coverage is based on explicit `//igraph:bind` annotations on exported Go declarations.", "",
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
        elif name in unsupported:
            status, owner = "Intentionally unsupported", "—"
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
    config = json.loads(args.config.read_text(encoding="utf-8"))
    output = args.output if args.output.is_absolute() else repo / args.output
    calls, annotations = discover_go_calls(repo), discover_annotations(repo)
    with tempfile.TemporaryDirectory(prefix="go-igraph-coverage-") as temp:
        source = args.source_dir or download_source(config["source_archive_url"], Path(temp))
        declarations = discover_upstream_api(locate_include(source))
        try:
            validate_inventory(declarations, calls, annotations, set(config.get("intentionally_unsupported", [])))
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
