# go-igraph

Go bindings for [igraph](https://igraph.org/).

## Prerequisites

`go-igraph` requires C-igraph version 1.0.0 or later.

Install C-igraph using package managers (e.g., `brew install igraph` on macOS) or by building it from source.

## Ownership and shared data

Public APIs use Go values and never require callers to handle C types. `Graph`
and `Vector` own C resources and should be closed; matrices, selectors, and
returned slices are Go-owned and need no cleanup. The
[shared data ownership contract](docs/shared-data-ownership.md) documents
lifetime, copying, nil, empty, and failure-path behavior in detail.

## Verification

Run the full local equivalent of the required CI job with:

```sh
make verify
```

This checks formatting, `go vet`, tests, the statement-coverage floor, the
coverage-tool tests, and the generated upstream API inventory.

## Upstream API coverage

The [upstream API roadmap](docs/upstream-api-roadmap.md) defines the binding
strategy, milestones, and completion criteria.

Milestone 3's fundamental algorithms include degree and neighborhoods,
components, BFS/DFS, shortest paths and distance matrices, density, diameter,
average path length, and transitivity. Milestone 4 adds selector-aware
closeness, harmonic and betweenness centrality, eigenvector centrality, HITS,
standard and personalized PageRank, and graph centralization. Milestone 5 adds
atomic selector-based deletion, independently owned induced and edge subgraphs,
connected-component graph decomposition, atomic simplification and direction
conversion, common graph operators, articulation points, bridges, and biconnected
components. Milestone 6 is complete: it adds community structure algorithms,
including modularity metrics and matrix calculations, flat community detection
(Multilevel, Leiden, Label Propagation, Infomap, Fluid), hierarchical community
detection and dendrogram cuts (Walktrap, FastGreedy, Edge Betweenness), spectral and
optimization community detection (Leading Eigenvector, Spinglass, Optimal Modularity),
and community partition comparison metrics (VI, NMI, Split-Join, Rand, Adjusted Rand).
All returned community partitions and dendrograms are Go-owned and survive source
graph closure. Milestone 10 adds general and color-aware graph isomorphism,
explicit first mappings and exact counts, bounded VF2 and LAD mapping
enumeration, induced LAD matching with domains, canonical graphs, automorphism
generators, and exact arbitrary-precision automorphism group sizes. Its graph
operands are borrowed under stable locking; mapping, permutation, generator,
and count results are Go-owned, while canonical graphs are independently
caller-closed. Milestone 11 begins with complete-graph, clique, and independent-
set decisions, clique and independence numbers, bounded ordinary and largest-
clique enumeration, clique-size histograms, and shared inclusive-range and
bounded-enumeration contracts. Executable examples demonstrate selector order, weighted distances,
restricted traversal, distance centrality, personalized ranking, deletion mappings,
component graphs, flat community detection, dendrogram cuts, bounded matching,
canonicalization, and automorphism analysis.

The generated [API coverage report](docs/api-coverage.md) compares the functions
exported by a pinned upstream igraph release with explicit `//igraph:bind` and
`//igraph:internal` annotations in production Go files. Unknown, duplicate, and
unclassified bindings fail the coverage check. This is an inventory for
planning binding work, not a claim of behavioral compatibility.

Regenerate the report (downloads the configured upstream source archive):

```sh
make coverage
```

Check that the committed report is current:

```sh
make coverage-check
```

GitHub Actions runs `make verify` for every pull request and for pushes to
`main`.

For an offline run, point the tool at an already extracted igraph source tree:

```sh
python3 tools/api_coverage.py --source-dir /path/to/igraph-1.0.1
```

## Docker

The Docker build pins both Go and the upstream C/igraph release, builds igraph
from source, and runs `go vet` and the Go tests. CI uses this same image instead
of the older igraph package from the Ubuntu repositories:

```sh
make docker-test
```

Run the tests against the working tree and write `coverage.out` locally with:

```sh
make docker-coverage
```

CI also enforces a statement coverage floor, currently 90%:

```sh
make docker-coverage-check
```

Override `COVERAGE_MIN` to test a different threshold.

Override `IGRAPH_VERSION` when validating a future upstream release.

The pinned version and upstream URLs live in
`tools/api_coverage_config.json`. Update the config and regenerate the report
when changing the comparison baseline.
