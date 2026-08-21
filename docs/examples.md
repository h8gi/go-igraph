# Guidelines for Go Examples in `go-igraph`

`go-igraph` follows standard Go library conventions to provide both package documentation on `pkg.go.dev` and standalone, runnable demonstration applications.

---

## 1. Documentation Examples (`pkg.go.dev`)

Documentation examples are placed in test files located in the root repository directory (`package igraph_test`).

### File Naming Convention
Examples are grouped by functional domain:
- `example_basic_test.go`: Basic graph creation, deletion, decomposition, and general graph operations.
- `example_paths_test.go`: Shortest/widest paths, distance matrices,
  reachability, derived graphs, and Eulerian traversals.
- `example_centrality_test.go`: Centrality metrics (closeness, betweenness, PageRank).
- `example_community_test.go`: Community detection algorithms (multilevel, walktrap, etc.).
- `example_flow_test.go`: Network flows, cuts, and connectivity.
- `example_isomorphism_test.go`: Isomorphism decisions, explicit mappings, bounded enumeration, canonical graphs, and automorphisms.
- `example_clique_test.go`: Clique extrema, bounded maximum-weight results, and independent sets.
- `example_cycle_test.go`: Bounded cycle enumeration, cycle bases, feedback sets, and topological verification.
- `example_bipartite_test.go`: Bipartite matrix construction, reproducible random generation, explicit partitions, source-ID-preserving projection, and weighted matching.
- `example_spatial_test.go`: Spatial graph construction, edge-ID-aligned distances, weighted routing, convex hulls, and proximity-graph comparison.
- `example_attributes_test.go`: Typed graph, vertex, and edge metadata with ID-aligned inspection and mutation.

### Function Naming & Structure
- Each example function must start with `Example` (e.g. `ExampleGraph_MaxFlow`, `ExampleNewGraphFromEdges`).
- Must use black-box test package (`package igraph_test`).
- Must end with `// Output:` comment asserting exact standard output so that `go test` validates it automatically.

---

## 2. Standalone Runnable Demos (`examples/` directory)

For end-to-end tutorial scripts or realistic CLI demos, `go-igraph` maintains standalone `package main` applications in `examples/`.
The `examples/paths/` program demonstrates bounded route alternatives and
directed reachability without exposing C types.

### Directory Structure
Each example is placed in its own subdirectory under `examples/`:
- `examples/maxflow/main.go`
- `examples/community/main.go`
- `examples/shortest_path/main.go`
- `examples/random/main.go`
- `examples/layout/main.go`
- `examples/isomorphism/main.go`
- `examples/cliques/main.go`
- `examples/cycles/main.go`
- `examples/bipartite/main.go`
- `examples/spatial/main.go`
- `examples/attributes/main.go`

### Usage & Verification
Users can run any standalone example using:
```bash
go run ./examples/maxflow
```

The attributes program demonstrates a complete typed GraphML export and
reimport while keeping the file and both graphs under explicit caller
ownership:

```bash
go run ./examples/attributes
```

During build & integration checks, all standalone examples under `examples/` are compiled to ensure API compatibility.
