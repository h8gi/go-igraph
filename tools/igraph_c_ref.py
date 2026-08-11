#!/usr/bin/env python3
"""CLI and helper tool for referencing upstream C igraph documentation and source code."""

from __future__ import annotations

import argparse
import json
import os
import re
import shutil
import sys
import tarfile
import urllib.request
from dataclasses import dataclass
from pathlib import Path

DEFAULT_IGRAPH_VERSION = "1.0.1"
DOCS_BASE_URL = "https://igraph.org/c/html/latest"

HEADER_TO_DOC_MODULE: dict[str, str] = {
    "igraph_layout.h": "igraph-Layout.html",
    "igraph_community.h": "igraph-Community.html",
    "igraph_centrality.h": "igraph-Centrality.html",
    "igraph_structural.h": "igraph-Structural.html",
    "igraph_paths.h": "igraph-Shortest-Paths.html",
    "igraph_flow.h": "igraph-Flow.html",
    "igraph_operators.h": "igraph-Operators.html",
    "igraph_constructors.h": "igraph-Generators.html",
    "igraph_foreign.h": "igraph-Foreign.html",
    "igraph_cliques.h": "igraph-Cliques.html",
    "igraph_bipartite.h": "igraph-Bipartite.html",
    "igraph_components.h": "igraph-Components.html",
    "igraph_eigen.h": "igraph-Spectral.html",
    "igraph_error.h": "igraph-Error-handling.html",
    "igraph_datatype.h": "igraph-Data-structures.html",
    "igraph_vector.h": "igraph-Data-structures.html",
    "igraph_matrix.h": "igraph-Data-structures.html",
    "igraph_arpack.h": "igraph-Spectral.html",
    "igraph_matching.h": "igraph-Matching.html",
    "igraph_graphicality.h": "igraph-Graphicality.html",
    "igraph_separators.h": "igraph-Separators.html",
    "igraph_isomorphism.h": "igraph-Isomorphism.html",
    "igraph_dsep.h": "igraph-D-Separation.html",
    "igraph_motifs.h": "igraph-Motifs.html",
    "igraph_random.h": "igraph-Random.html",
    "igraph_visitors.h": "igraph-Visitors.html",
}

EXPORT_RE = re.compile(
    r"^\s*(?:(?:IGRAPH_EXPERIMENTAL|IGRAPH_DEPRECATED)\s+)?IGRAPH_EXPORT"
    r"(?:(?!;).)*?\b(igraph_[A-Za-z0-9_]+)\s*\(([^)]*)\)",
    re.MULTILINE | re.DOTALL,
)
CALL_RE = re.compile(r"\bC\.(igraph_[A-Za-z0-9_]+)\b")
ANNOTATION_RE = re.compile(r"^\s*//igraph:(bind|internal)\s+(igraph_[A-Za-z0-9_]+)\s*$")


@dataclass
class CFunctionInfo:
    name: str
    header_relative: str
    params_text: str
    doc_url: str
    source_file: str | None = None
    source_line: int | None = None


def get_repo_root() -> Path:
    return Path(__file__).resolve().parent.parent


def get_cache_dir() -> Path:
    cache = os.environ.get("GO_IGRAPH_CACHE_DIR")
    if cache:
        return Path(cache)
    return Path.home() / ".cache" / "go-igraph"


def ensure_c_igraph_source(version: str = DEFAULT_IGRAPH_VERSION) -> Path:
    cache_dir = get_cache_dir()
    version_dir = cache_dir / f"igraph-{version}"
    if version_dir.exists() and (version_dir / "include").exists():
        return version_dir

    cache_dir.mkdir(parents=True, exist_ok=True)
    tar_path = cache_dir / f"igraph-{version}.tar.gz"
    url = f"https://github.com/igraph/igraph/releases/download/{version}/igraph-{version}.tar.gz"

    if not tar_path.exists():
        print(f"Downloading C igraph {version} source from {url}...", file=sys.stderr)
        urllib.request.urlretrieve(url, tar_path)

    temp_extract = cache_dir / f"extract-{version}"
    if temp_extract.exists():
        shutil.rmtree(temp_extract)
    temp_extract.mkdir(parents=True, exist_ok=True)

    print(f"Extracting C igraph {version} source...", file=sys.stderr)
    with tarfile.open(tar_path, "r:gz") as tar:
        if hasattr(tarfile, "data_filter"):
            tar.extractall(path=temp_extract, filter="data")
        else:
            tar.extractall(path=temp_extract)

    children = list(temp_extract.iterdir())
    root_src = children[0] if len(children) == 1 and children[0].is_dir() else temp_extract

    if version_dir.exists():
        shutil.rmtree(version_dir)
    shutil.move(str(root_src), str(version_dir))
    if temp_extract.exists():
        shutil.rmtree(temp_extract, ignore_errors=True)

    return version_dir


def build_doc_url(header_name: str, function_name: str) -> str:
    header_basename = Path(header_name).name
    doc_page = HEADER_TO_DOC_MODULE.get(header_basename)
    if doc_page:
        return f"{DOCS_BASE_URL}/{doc_page}#{function_name}"
    return f"{DOCS_BASE_URL}/cigraph-index.html#{function_name}"


def parse_c_declarations(src_dir: Path) -> dict[str, CFunctionInfo]:
    include_dir = src_dir / "include"
    if not include_dir.exists():
        include_dir = src_dir
    results: dict[str, CFunctionInfo] = {}

    headers = sorted(include_dir.rglob("*.h"))
    for header in headers:
        header_rel = header.relative_to(include_dir).as_posix()
        text = header.read_text(encoding="utf-8", errors="replace")
        text_no_comments = re.sub(r"/\*.*?\*/|//[^\n]*", "", text, flags=re.DOTALL)
        for match in EXPORT_RE.finditer(text_no_comments):
            fn_name = match.group(1)
            params = match.group(2).strip()
            doc_url = build_doc_url(header_rel, fn_name)
            results[fn_name] = CFunctionInfo(
                name=fn_name,
                header_relative=header_rel,
                params_text=params,
                doc_url=doc_url,
            )

    c_sources = sorted(src_dir.rglob("*.c"))
    for c_file in c_sources:
        try:
            lines = c_file.read_text(encoding="utf-8", errors="replace").splitlines()
        except Exception:
            continue
        rel_c = c_file.relative_to(src_dir).as_posix()
        for idx, line in enumerate(lines, 1):
            for fn_name in results:
                if (
                    results[fn_name].source_file is None
                    and fn_name in line
                    and ("(" in line or (idx < len(lines) and "(" in lines[idx]))
                    and not line.strip().startswith("//")
                    and not line.strip().startswith("/*")
                ):
                    results[fn_name].source_file = rel_c
                    results[fn_name].source_line = idx

    return results


def find_go_references(repo_dir: Path, symbol: str) -> dict[str, list[str]]:
    references: dict[str, list[str]] = {
        "annotations": [],
        "cgo_calls": [],
    }

    go_files = sorted(
        p for p in repo_dir.rglob("*.go") if ".git" not in p.parts and not p.name.startswith(".worktrees")
    )
    for gfile in go_files:
        rel_path = gfile.relative_to(repo_dir).as_posix()
        lines = gfile.read_text(encoding="utf-8", errors="replace").splitlines()
        for idx, line in enumerate(lines, 1):
            if ANNOTATION_RE.match(line) and symbol in line:
                references["annotations"].append(f"{rel_path}:{idx}: {line.strip()}")
            elif f"C.{symbol}" in line:
                references["cgo_calls"].append(f"{rel_path}:{idx}: {line.strip()}")

    return references


def generate_audit_tips(fn_name: str, params: str) -> list[str]:
    tips: list[str] = []
    if "vector" in params or "igraph_vector" in params:
        tips.append(
            "[Memory Leak / Ownership] Uses igraph_vector_t. Check if memory is freed with igraph_vector_destroy() and if Go slice index matches 0-based C array indexing."
        )
    if "matrix" in params or "igraph_matrix" in params:
        tips.append(
            "[Matrix Layout] Uses igraph_matrix_t. C igraph matrices store elements in column-major order. Verify row/column ordering in Go matrix conversion."
        )
    if "igraph_t" in params:
        tips.append(
            "[Graph Safety] Takes igraph_t graph parameter. Ensure C graph pointer is non-null and not modified unsafely if shared."
        )
    if "igraph_integer_t" in params:
        tips.append(
            "[Type Bounds] Uses igraph_integer_t. Cast Go int/int64 explicitly via C.igraph_integer_t to prevent overflow."
        )
    if "igraph_real_t" in params:
        tips.append(
            "[Precision] Uses igraph_real_t (C double). Map to Go float64."
        )
    if "NULL" in params or "weights" in params or "res" in params:
        tips.append(
            "[Optional Args] Check if NULL pointers are supported when optional Go arguments are omitted/nil."
        )
    tips.append(
        "[Error Check] Upstream returns igraph_error_t. Ensure non-zero C error codes trigger idiomatic Go errors."
    )
    return tips


def command_lookup(symbol: str, version: str = DEFAULT_IGRAPH_VERSION) -> None:
    repo = get_repo_root()
    src_dir = ensure_c_igraph_source(version)
    decls = parse_c_declarations(src_dir)

    info = decls.get(symbol)
    print(f"=== Upstream C igraph Symbol: {symbol} ===")
    if info:
        print(f"Header:           include/{info.header_relative}")
        print(f"Official Doc URL: {info.doc_url}")
        print(f"C Parameters:     ({info.params_text})")
        if info.source_file:
            print(f"C Implementation: {info.source_file}:{info.source_line}")
        else:
            print("C Implementation: Source location not cached")
    else:
        print(f"Symbol '{symbol}' not found in C igraph {version} header exports.")
        print(f"Default Doc Index: {DOCS_BASE_URL}/cigraph-index.html#{symbol}")

    print("\n=== Go Project Bindings & References ===")
    refs = find_go_references(repo, symbol)
    if refs["annotations"]:
        print("Annotations:")
        for ann in refs["annotations"]:
            print(f"  - {ann}")
    else:
        print("Annotations: None found")

    if refs["cgo_calls"]:
        print("CGO Calls:")
        for call in refs["cgo_calls"][:10]:
            print(f"  - {call}")
        if len(refs["cgo_calls"]) > 10:
            print(f"  ... and {len(refs['cgo_calls']) - 10} more calls")
    else:
        print("CGO Calls: None found")

    print("\n=== Bug Audit Checklist ===")
    params_str = info.params_text if info else symbol
    tips = generate_audit_tips(symbol, params_str)
    for tip in tips:
        print(f"- {tip}")


def command_url(symbol: str, version: str = DEFAULT_IGRAPH_VERSION) -> None:
    src_dir = ensure_c_igraph_source(version)
    decls = parse_c_declarations(src_dir)
    info = decls.get(symbol)
    if info:
        print(info.doc_url)
    else:
        print(f"{DOCS_BASE_URL}/cigraph-index.html#{symbol}")


def command_audit(symbol: str, version: str = DEFAULT_IGRAPH_VERSION) -> None:
    command_lookup(symbol, version)


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Inspect upstream C igraph documentation, declarations, and Go bindings."
    )
    parser.add_argument(
        "--version",
        default=DEFAULT_IGRAPH_VERSION,
        help="C igraph upstream version (default: 1.0.1)",
    )

    subparsers = parser.add_subparsers(dest="command", required=True)

    lookup_parser = subparsers.add_parser("lookup", help="Look up a C igraph symbol")
    lookup_parser.add_argument("symbol", help="C symbol name, e.g. igraph_layout_sugiyama")

    url_parser = subparsers.add_parser("url", help="Get official documentation URL for a symbol")
    url_parser.add_argument("symbol", help="C symbol name, e.g. igraph_modularity")

    audit_parser = subparsers.add_parser("audit", help="Run audit checks for a C symbol in Go")
    audit_parser.add_argument("symbol", help="C symbol name, e.g. igraph_betweenness")

    args = parser.parse_args()

    if args.command == "lookup":
        command_lookup(args.symbol, version=args.version)
    elif args.command == "url":
        command_url(args.symbol, version=args.version)
    elif args.command == "audit":
        command_audit(args.symbol, version=args.version)


if __name__ == "__main__":
    main()
