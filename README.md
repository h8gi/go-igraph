# go-igraph

Go bindings for [igraph](https://igraph.org/).

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
standard and personalized PageRank, and graph centralization. Executable
examples demonstrate selector order, weighted distances, restricted traversal,
distance centrality, and personalized ranking.

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
