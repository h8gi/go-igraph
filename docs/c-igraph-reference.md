# C igraph Reference & Binding Guide

This directory documents how `go-igraph` references the upstream C igraph project, how to navigate the official C documentation, and how to query Go binding status.

---

## Upstream Resources

- **Official C igraph Documentation**: [https://igraph.org/c/html/latest/](https://igraph.org/c/html/latest/)
- **Upstream Repository**: [https://github.com/igraph/igraph](https://github.com/igraph/igraph)
- **Target Upstream Version**: `1.0.1` (configured in `tools/api_coverage_config.json`)

---

## Quick Reference CLI Tool

Use `tools/igraph_c_ref.py` to inspect C declarations, get exact official documentation URLs, and check Go annotations and `CGO` calls.

```bash
# Look up details, doc URL, C declaration, Go annotations, and cgo calls
python3 tools/igraph_c_ref.py lookup igraph_layout_sugiyama

# Print the official HTML documentation URL
python3 tools/igraph_c_ref.py url igraph_modularity

# Shortcut via Makefile
make c-ref SYMBOL=igraph_betweenness
```

---

## Shared Upstream Tools Architecture

- `tools/api_coverage_config.json`: The single source of truth for upstream repository, version, and source archive URLs.
- `tools/upstream_api.py`: Shared library that parses C declarations, header mappings, Go annotations (`//igraph:bind`), and exact `cgo` call locations.
- `tools/api_coverage.py`: Validates complete upstream coverage and generates `docs/api-coverage.md`.
- `tools/igraph_c_ref.py`: Fast interactive CLI for developers and AI agents.
