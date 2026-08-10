# C igraph Reference & Upstream Audit Guide

This directory documents how `go-igraph` references the upstream C igraph project, how to navigate the official C documentation, and how to verify Go binding contracts against the C source.

---

## Upstream Resources

- **Official C igraph Documentation**: [https://igraph.org/c/html/latest/](https://igraph.org/c/html/latest/)
- **Upstream C igraph Repository**: [https://github.com/igraph/igraph](https://github.com/igraph/igraph)
- **Target Upstream Version**: `1.0.1`

---

## Quick Reference CLI Tool

Use `tools/igraph_c_ref.py` to inspect C declarations, get exact official documentation URLs, download and search upstream C source code, and perform bug audit checks.

```bash
# Look up details, doc URL, C declaration, and Go usage for a symbol
python3 tools/igraph_c_ref.py lookup igraph_layout_sugiyama

# Print the official HTML documentation URL
python3 tools/igraph_c_ref.py url igraph_modularity

# Run bug audit checklist for a symbol
python3 tools/igraph_c_ref.py audit igraph_betweenness
```

---

## Antigravity AI Customization Skill

An Antigravity agent skill is configured at `.agents/skills/c-igraph-reference/SKILL.md`.
When debugging or implementing features with AI assistance, this skill automatically provides the agent with guidelines and tools to cross-reference C igraph contracts.

---

## Key Bug Prevention Principles

1. **Resource Ownership**: C structures (`igraph_vector_t`, `igraph_matrix_t`, `igraph_t`) allocated during function execution must be destroyed via `igraph_..._destroy()`.
2. **Matrix Indexing**: C igraph stores matrices in column-major order.
3. **Error Code Checking**: C functions return `igraph_error_t`. Non-zero return values must be converted to Go `error` return values.
4. **0-based Indexing**: C igraph vertex and edge IDs are 0-indexed.
