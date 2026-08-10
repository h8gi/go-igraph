---
name: c-igraph-reference
description: Guide and tool integration for referencing C igraph official documentation (https://igraph.org/c/html/latest/), C source code, header declarations, and memory/error contracts when debugging or porting functions in go-igraph.
---

# C igraph Reference & Audit Guide

When working on `go-igraph`, bugs frequently arise from subtle mismatches between Go implementations and upstream C igraph semantics. Use this skill to consult official C igraph documentation, inspect upstream C source code, and systematically verify contracts.

## Upstream Documentation & Source Sources

- **Official C igraph HTML Documentation**:
  - Latest C API: `https://igraph.org/c/html/latest/`
  - Function Anchor Pattern: `https://igraph.org/c/html/latest/<module>.html#<function_name>`
    - Example: `https://igraph.org/c/html/latest/igraph-Layout.html#igraph_layout_sugiyama`
    - Example: `https://igraph.org/c/html/latest/igraph-Community.html#igraph_modularity`

- **Upstream C igraph Source Code**:
  - GitHub Repository: `https://github.com/igraph/igraph`
  - Target Version: `1.0.1` (see `Dockerfile` and `Makefile`)

---

## Quick Reference Tools

Use `tools/igraph_c_ref.py` to inspect C igraph API definitions, fetch official doc URLs, view upstream C code, and check Go binding bindings locally:

```bash
# Lookup C igraph function details, doc URL, C declaration, and Go binding mapping
python3 tools/igraph_c_ref.py lookup igraph_layout_sugiyama

# Get exact official documentation URL for a C symbol
python3 tools/igraph_c_ref.py url igraph_modularity

# Audit Go binding against common C igraph bug patterns (memory leakage, 0-indexing, error handling)
python3 tools/igraph_c_ref.py audit igraph_betweenness
```

---

## Common Bug Audit Checklist for C igraph Porting

When debugging or implementing a `go-igraph` wrapper, verify the following 5 critical areas against the C igraph source and documentation:

### 1. Resource Ownership & Destructors
- **Vectors / Matrices / Graphs**: Did the C call initialize or populate an `igraph_vector_t`, `igraph_matrix_t`, `igraph_strvector_t`, `igraph_arpack_options_t`, etc.?
- **Freeing Memory**: Ensure every initialized C structure is cleaned up using its corresponding destructor (e.g., `igraph_vector_destroy(&vec)`).
- **Go Ownership**: If data is copied to Go slices, free the temporary C vector before returning. If a wrapper holds C memory, ensure `runtime.SetFinalizer` or an explicit `.Close()` / `Free()` method is provided.

### 2. Error Propagation (`igraph_error_t`)
- Upstream C igraph functions return an `igraph_error_t` (0 / `IGRAPH_SUCCESS` on success, non-zero error code on failure).
- Ensure C calls are wrapped with error checking (e.g. `IGRAPH_CHECK` or `callC(...)`) and propagate errors cleanly to Go as idiomatic `error` values.
- Verify early returns in Go do not bypass C cleanup (`defer` or cleanup handlers must be used).

### 3. Array Indexing & Matrices
- **0-based vs 1-based indexing**: C igraph uses 0-based vertex and edge IDs. Go wrappers must preserve or map 0-based indices cleanly.
- **Matrix Layout**: C igraph `igraph_matrix_t` stores data in column-major order (`IGRAPH_MATRIX(m, row, col)`). Ensure Go matrix conversions respect column-major storage without accidentally transposing rows and columns.

### 4. Optional Parameters & NULL Pointers
- Upstream functions often allow `NULL` for optional outputs (e.g. `weights`, `res`, `parents`, `in_degree`).
- Verify whether passing `NULL` or a valid C pointer is expected when optional Go arguments (e.g., `nil` slices or `nil` options) are supplied.

### 5. Type Conversions & Precision
- C `igraph_integer_t` is typically 64-bit (`int64_t` or `long int`) or 32-bit depending on build options. Go code should convert via explicit C types (`C.igraph_integer_t`, `C.double`).
- Boolean flags must use `C.igraph_bool_t` (`true`/`false`).
