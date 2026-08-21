# go-igraph

[![Go Reference](https://pkg.go.dev/badge/github.com/h8gi/go-igraph.svg)](https://pkg.go.dev/github.com/h8gi/go-igraph)
[![CI](https://github.com/h8gi/go-igraph/actions/workflows/api-coverage.yml/badge.svg)](https://github.com/h8gi/go-igraph/actions/workflows/api-coverage.yml)

Go bindings for the [igraph](https://igraph.org/) network analysis library.

## Prerequisites

`go-igraph` requires C/igraph 1.0.0 or later and a working `pkg-config`
installation. Install C/igraph with your system package manager—for example:

```sh
brew install igraph # macOS with Homebrew
```

See the [igraph installation guide](https://igraph.org/c/html/latest/igraph-Installation.html)
for other platforms and source builds.

## Installation

Add `go-igraph` to your module:

```sh
go get github.com/h8gi/go-igraph
```

## Quick start

The following program creates an undirected graph with three vertices and two
edges:

```go
package main

import (
	"fmt"
	"log"

	igraph "github.com/h8gi/go-igraph"
)

func main() {
	graph, err := igraph.NewGraph()
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	if err := graph.AddVertices(3); err != nil {
		log.Fatal(err)
	}
	if err := graph.AddEdges([]igraph.Edge{
		{From: 0, To: 1},
		{From: 1, To: 2},
	}); err != nil {
		log.Fatal(err)
	}

	vertices, err := graph.VertexCount()
	if err != nil {
		log.Fatal(err)
	}
	edges, err := graph.EdgeCount()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("vertices: %d, edges: %d\n", vertices, edges)
}
```

`Graph` and `Vector` values own C resources and should be closed. Public input
slices are borrowed only for the duration of a call and copied when needed;
returned slices, matrices, and selectors are Go-owned. See the
[shared data ownership contract](docs/shared-data-ownership.md) for complete
lifetime, copying, nil, empty, and failure-path behavior.

## Typed attributes and interchange

Graphs expose typed Boolean, numeric, and string metadata at graph, vertex,
and edge scope. Vertex and edge value slices are aligned with their current
IDs, transformations use explicit attribute-combination policies where values
merge, and GraphML supports attributed export and reimport. See the
[attributes example](examples/attributes/main.go) for a complete workflow.

## Documentation

- [Package reference](https://pkg.go.dev/github.com/h8gi/go-igraph)
- [Runnable and package examples](docs/examples.md)
- [Upstream API roadmap](docs/upstream-api-roadmap.md)
- [Generated API coverage report](docs/api-coverage.md)
- [Package architecture and coverage strategy](docs/package-architecture.md)

## License

`go-igraph` is licensed under the GNU General Public License, version 2 or
later (`GPL-2.0-or-later`). See [LICENSE](LICENSE) for the license text and
[Licensing and provenance](docs/licensing.md) for the scope, upstream
dependency, and repository-history details.

`go-igraph` links to the separately distributed
[C/igraph](https://igraph.org/) library, which is also licensed under GPL
version 2 or later. Distributors of linked programs are responsible for
complying with the licenses of the C/igraph build and its enabled
dependencies.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for local verification, Docker workflows,
and API coverage maintenance.
