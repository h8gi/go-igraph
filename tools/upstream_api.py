#!/usr/bin/env python3
"""Shared upstream C igraph API discovery, configuration, and annotation tools."""

from __future__ import annotations

import json
import os
import re
import shutil
import subprocess
import sys
import tarfile
import tempfile
import urllib.request
from dataclasses import dataclass
from pathlib import Path

EXPORT_RE = re.compile(
    r"^\s*(?:(?:IGRAPH_EXPERIMENTAL|IGRAPH_DEPRECATED)\s+)?IGRAPH_EXPORT"
    r"(?:(?!;).)*?\b(igraph_[A-Za-z0-9_]+)\s*\(([^)]*)\)",
    re.MULTILINE | re.DOTALL,
)
CALL_RE = re.compile(r"\bC\.(igraph_[A-Za-z0-9_]+)\b")
ANNOTATION_RE = re.compile(r"^\s*//igraph:(bind|internal)\s+(igraph_[A-Za-z0-9_]+)\s*$")
DECL_RE = re.compile(r"^\s*(?:func|type|var|const)\s+(?:\([^)]*\)\s*)?([A-Za-z_][A-Za-z0-9_]*)")

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

DOCS_BASE_URL = "https://igraph.org/c/html/latest"


@dataclass(frozen=True)
class Annotation:
    kind: str
    upstream: str
    go_symbol: str
    path: str
    line: int


@dataclass(frozen=True)
class CDeclaration:
    name: str
    header: str
    params: str
    doc_url: str


def load_config(config_path: Path) -> dict:
    return json.loads(config_path.read_text(encoding="utf-8"))


def get_cache_dir() -> Path:
    cache = os.environ.get("GO_IGRAPH_CACHE_DIR")
    if cache:
        return Path(cache)
    return Path.home() / ".cache" / "go-igraph"


def download_or_get_source(config: dict) -> Path:
    version = config.get("version", "1.0.1")
    cache_dir = get_cache_dir()
    version_dir = cache_dir / f"igraph-{version}"
    if version_dir.exists() and (version_dir / "include").exists():
        return version_dir

    cache_dir.mkdir(parents=True, exist_ok=True)
    url = config["source_archive_url"]
    archive = cache_dir / f"igraph-{version}.tar.gz"
    if not archive.exists():
        urllib.request.urlretrieve(url, archive)

    temp_extract = cache_dir / f"extract-{version}"
    if temp_extract.exists():
        shutil.rmtree(temp_extract)
    temp_extract.mkdir(parents=True, exist_ok=True)

    with tarfile.open(archive, "r:gz") as bundle:
        if hasattr(tarfile, "data_filter"):
            bundle.extractall(temp_extract, filter="data")
        else:
            bundle.extractall(temp_extract)

    children = list(temp_extract.iterdir())
    root_src = children[0] if len(children) == 1 and children[0].is_dir() else temp_extract

    if version_dir.exists():
        shutil.rmtree(version_dir)
    shutil.move(str(root_src), str(version_dir))
    if temp_extract.exists():
        shutil.rmtree(temp_extract, ignore_errors=True)

    return version_dir


def locate_include(source: Path) -> Path:
    direct = source / "include"
    if direct.is_dir():
        return direct
    candidates = list(source.glob("*/include"))
    if len(candidates) == 1:
        return candidates[0]
    raise ValueError(f"could not locate a unique include directory below {source}")


def build_doc_url(header_name: str, function_name: str) -> str:
    header_basename = Path(header_name).name.replace(" (generated)", "")
    doc_page = HEADER_TO_DOC_MODULE.get(header_basename)
    if doc_page:
        return f"{DOCS_BASE_URL}/{doc_page}#{function_name}"
    return f"{DOCS_BASE_URL}/cigraph-index.html#{function_name}"


def strip_comments(text: str) -> str:
    return re.sub(r"/\*.*?\*/|//[^\n]*", "", text, flags=re.DOTALL)


def discover_upstream_declarations(include_dir: Path) -> dict[str, CDeclaration]:
    declarations: dict[str, CDeclaration] = {}
    headers = sorted(include_dir.rglob("*.h"))
    for header in headers:
        header_rel = header.relative_to(include_dir).as_posix()
        text = strip_comments(header.read_text(encoding="utf-8", errors="replace"))
        for match in EXPORT_RE.finditer(text):
            symbol = match.group(1)
            params = match.group(2).strip()
            if symbol not in declarations:
                declarations[symbol] = CDeclaration(
                    name=symbol,
                    header=header_rel,
                    params=params,
                    doc_url=build_doc_url(header_rel, symbol),
                )

    with tempfile.TemporaryDirectory(prefix="go-igraph-headers-") as temp:
        prepared = Path(temp) / "include"
        shutil.copytree(include_dir, prepared)
        (prepared / "igraph_export.h").write_text(
            "#define IGRAPH_EXPORT IGRAPH_EXPORT\n"
            "#define IGRAPH_NO_EXPORT IGRAPH_NO_EXPORT\n"
            "#define IGRAPH_DEPRECATED IGRAPH_DEPRECATED\n",
            encoding="utf-8",
        )
        for generated in ("igraph_config.h", "igraph_version.h"):
            target = prepared / generated
            if not target.exists():
                target.write_text("/* neutral coverage-tool stub */\n", encoding="utf-8")
        decls = prepared / "igraph_decls.h"
        if decls.exists():
            decls.write_text(
                decls.read_text(encoding="utf-8").replace(
                    "#define IGRAPH_PRIVATE_EXPORT IGRAPH_EXPORT",
                    "#define IGRAPH_PRIVATE_EXPORT IGRAPH_PRIVATE_EXPORT",
                ),
                encoding="utf-8",
            )
        for header in headers:
            relative = header.relative_to(include_dir)
            if relative.as_posix() == "igraph.h":
                continue
            command = [os.environ.get("CC", "cc"), "-E", "-P", "-I", str(prepared), str(prepared / relative)]
            try:
                process = subprocess.run(command, check=False, capture_output=True, text=True)
                expanded = process.stdout
            except FileNotFoundError as error:
                raise ValueError(f"could not run C preprocessor {command[0]}: {error}") from error
            if not expanded:
                continue
            for match in EXPORT_RE.finditer(expanded):
                symbol = match.group(1)
                params = match.group(2).strip()
                if symbol not in declarations:
                    header_str = f"{relative.as_posix()} (generated)"
                    declarations[symbol] = CDeclaration(
                        name=symbol,
                        header=header_str,
                        params=params,
                        doc_url=build_doc_url(relative.as_posix(), symbol),
                    )
    return declarations


def discover_upstream_api(include_dir: Path) -> dict[str, str]:
    decls = discover_upstream_declarations(include_dir)
    return {symbol: decl.header for symbol, decl in decls.items()}


def production_go_files(repo: Path) -> list[Path]:
    return sorted(
        path for path in repo.rglob("*.go")
        if not path.name.endswith("_test.go") and ".git" not in path.parts and not path.name.startswith(".worktrees")
    )


def discover_go_calls(repo: Path) -> set[str]:
    calls: set[str] = set()
    for path in production_go_files(repo):
        calls.update(name for name in CALL_RE.findall(path.read_text(encoding="utf-8")) if not name.endswith("_t"))
    return calls


def find_cgo_call_locations(repo: Path, symbol: str) -> list[tuple[str, int, str]]:
    results: list[tuple[str, int, str]] = []
    symbol_pattern = re.compile(rf"\bC\.{re.escape(symbol)}\b")
    for path in production_go_files(repo):
        rel_path = path.relative_to(repo).as_posix()
        for idx, line in enumerate(path.read_text(encoding="utf-8", errors="replace").splitlines(), 1):
            if symbol_pattern.search(line):
                results.append((rel_path, idx, line.strip()))
    return results


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
