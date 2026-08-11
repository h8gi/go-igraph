#!/usr/bin/env python3
"""CLI tool for looking up upstream C igraph documentation, declarations, and Go bindings."""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

from upstream_api import (
    discover_annotations,
    download_or_get_source,
    find_cgo_call_locations,
    load_or_build_declaration_index,
    load_config,
    locate_include,
)


def get_repo_root() -> Path:
    return Path(__file__).resolve().parent.parent


def command_lookup(symbol: str, config_path: Path) -> int:
    repo = get_repo_root()
    config = load_config(config_path)
    source_dir = download_or_get_source(config)
    include_dir = locate_include(source_dir)
    declarations = load_or_build_declaration_index(config, include_dir)
    annotations = discover_annotations(repo)

    decl = declarations.get(symbol)
    print(f"=== Upstream C igraph {config['version']} Symbol: {symbol} ===")
    if decl:
        print(f"Header:           include/{decl.header}")
        docs_version = config.get("documentation_version", "latest")
        print(f"Official Doc URL ({docs_version}): {decl.doc_url}")
        print(f"C Declaration:    {decl.declaration}")
    else:
        print(f"Symbol '{symbol}' not found in upstream C igraph {config['version']} declarations.")
        docs_base_url = config["documentation_base_url"]
        print(f"Default Doc Index: {docs_base_url}/cigraph-index.html#{symbol}")

    print("\n=== Go Project Bindings & References ===")
    matching_annos = [a for a in annotations if a.upstream == symbol]
    if matching_annos:
        print("Annotations:")
        for ann in matching_annos:
            print(f"  - [{ann.kind}] {ann.path}:{ann.line}: Go symbol `{ann.go_symbol}`")
    else:
        print("Annotations: None found")

    cgo_calls = find_cgo_call_locations(repo, symbol)
    if cgo_calls:
        print(f"CGO Calls ({len(cgo_calls)} location(s)):")
        for path, line_num, line_str in cgo_calls[:10]:
            print(f"  - {path}:{line_num}: {line_str}")
        if len(cgo_calls) > 10:
            print(f"  ... and {len(cgo_calls) - 10} more location(s)")
    else:
        print("CGO Calls: None found")

    return 0 if decl else 1


def command_url(symbol: str, config_path: Path) -> int:
    config = load_config(config_path)
    source_dir = download_or_get_source(config)
    include_dir = locate_include(source_dir)
    declarations = load_or_build_declaration_index(config, include_dir)

    decl = declarations.get(symbol)
    if decl:
        print(decl.doc_url)
    else:
        print(f"{config['documentation_base_url']}/cigraph-index.html#{symbol}")
    return 0 if decl else 1


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Inspect upstream C igraph declarations, official documentation URLs, and Go bindings."
    )
    parser.add_argument(
        "--config",
        type=Path,
        default=Path("tools/api_coverage_config.json"),
        help="Path to api_coverage_config.json",
    )

    subparsers = parser.add_subparsers(dest="command", required=True)

    lookup_parser = subparsers.add_parser("lookup", help="Look up exact C igraph symbol details")
    lookup_parser.add_argument("symbol", help="Exact C symbol name, e.g. igraph_layout_sugiyama")

    url_parser = subparsers.add_parser("url", help="Print official documentation URL for a symbol")
    url_parser.add_argument("symbol", help="Exact C symbol name, e.g. igraph_modularity")

    args = parser.parse_args()

    config_path = args.config if args.config.is_absolute() else get_repo_root() / args.config

    if args.command == "lookup":
        return command_lookup(args.symbol, config_path)
    elif args.command == "url":
        return command_url(args.symbol, config_path)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
