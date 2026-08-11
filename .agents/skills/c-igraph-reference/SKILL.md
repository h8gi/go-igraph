---
name: c-igraph-reference
description: Guide and tool integration for referencing C igraph official documentation (https://igraph.org/c/html/latest/), header declarations, and memory/error contracts when porting or inspecting functions in go-igraph.
---

# C igraph Reference Guide

When working on `go-igraph`, developers and agents need quick, accurate access to upstream C igraph declarations, official documentation URLs, and Go binding contracts.

## Upstream Sources & Configuration

- **Official C igraph HTML Documentation**: `https://igraph.org/c/html/latest/` (moving `latest`; the CLI labels it separately from the pinned source version)
- **Upstream C igraph Version**: Defined in `tools/api_coverage_config.json` (currently `1.0.1`).

---

## Quick Reference CLI Tool

Use `tools/igraph_c_ref.py` (or `make c-ref SYMBOL=...`) to look up exact C igraph API definitions, fetch official doc URLs, and inspect Go binding locations:

```bash
# Look up exact C declaration, header, official doc URL, and Go bindings/cgo calls
python3 tools/igraph_c_ref.py lookup igraph_layout_sugiyama

# Print exact official documentation URL for a C symbol
python3 tools/igraph_c_ref.py url igraph_modularity

# Shortcut via Makefile
make c-ref SYMBOL=igraph_betweenness
```

---

## Core Binding Rules & Verification Checklist

As specified in `AGENTS.md`, ensure the following contracts are maintained when wrapping C igraph functions:

1. **Resource Ownership & Destructors**:
   - Every initialized C structure (`igraph_vector_t`, `igraph_matrix_t`, `igraph_t`, etc.) must be freed with its corresponding destructor (e.g. `igraph_vector_destroy(&vec)`).
2. **Error Propagation**:
   - Non-zero C `igraph_error_t` return codes must be captured and returned as Go `error` values.
3. **Array & Matrix Layout**:
   - Indexing is 0-based in C igraph.
   - `igraph_matrix_t` is column-major in memory.
4. **Clean Testing**:
   - Test cases must cover initialization failure, upstream errors, early returns, empty values, and resource cleanup.
