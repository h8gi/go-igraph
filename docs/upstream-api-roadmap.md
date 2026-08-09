# Upstream API roadmap

This roadmap describes how go-igraph will grow from a small binding into a
coherent Go interface for C/igraph. It prioritizes useful, safe API slices over
maximizing a raw function count.

The generated [API coverage report](api-coverage.md) remains the inventory of
the pinned upstream release. GitHub milestones and issues should track the
execution of the work described here.

## Principles

- Design an idiomatic Go API instead of reproducing C signatures mechanically.
- Prefer correctness, explicit ownership, and maintainability over backward
  compatibility.
- Add APIs in complete, testable feature slices.
- Build shared conversion and selector types before algorithms that depend on
  them.
- Every binding must propagate C/igraph errors and define ownership of returned
  C resources.
- Behavioral tests are required; increasing the inventory count alone is not
  sufficient.

## Coverage semantics

The report uses explicit annotations instead of treating every direct
`C.igraph_*` call as user-facing coverage. This avoids the earlier problems
where:

- calls inside cgo C helpers are not discovered;
- lifecycle plumbing can increase the count without adding user-visible
  functionality;
- one Go API may intentionally compose several upstream functions;
- an upstream function may be intentionally unsupported or internal-only.

An exported Go declaration uses `//igraph:bind` to claim a user-facing upstream
function. `//igraph:internal` marks implementation dependencies, while the tool
configuration records deliberately unsupported functions. The report
distinguishes:

- user-facing bindings;
- internal implementation dependencies;
- intentionally unsupported functions;
- missing functions.

## Milestone 1: Core graph API

Goal: create, inspect, and modify ordinary graphs without reaching into cgo.

Status: complete. The graph API covers the planned inspection, mutation,
lookup, copying, and deterministic-construction areas, and reached its initial
target with 26 user-facing bindings. The binding inventory and behavioral
checks are enforced in CI.

Implemented areas:

- graph counts and directionality;
- add vertices and edges;
- adjacency and edge lookup;
- copy and basic graph construction;
- full, ring, star, tree, and other common deterministic constructors.

Initial target: 20–30 well-tested user-facing bindings.

Completion criteria:

- APIs follow Go naming and error conventions;
- resources have explicit ownership;
- directed and undirected behavior is tested;
- the explicit binding inventory is enforced in CI.

## Milestone 2: Shared data and selection layer

Goal: provide reusable infrastructure needed by most algorithms.

Status: complete. The public lifetime, copying, nil, and empty contracts are
recorded in [Shared data ownership](shared-data-ownership.md). The generated
inventory classifies the lifecycle functions used by the common wrappers, and
`make verify` enforces formatting, vet, behavioral tests, statement coverage,
inventory-tool tests, and report freshness in both local development and CI.

Implemented areas:

- integer, real, boolean, and string conversions;
- matrices;
- vertex and edge selectors;
- iterators;
- safe conversion between Go slices and C/igraph data;
- result types that own and release C resources correctly.

Completion criteria:

- no API requires callers to handle C types;
- conversion and ownership rules are documented and tested;
- common helpers remove duplicated cgo code from later bindings.

Implementation decisions:

- temporary integer, real, boolean, and string vectors copy at the Go/C
  boundary and return Go-owned slices;
- `Matrix` is an immutable Go value and C matrices are temporary adapters;
- selectors are reusable Go values that copy explicit input and retain no
  graph or C resource;
- selection is eagerly materialized into Go-owned slices so C iterators never
  escape or outlive their graph lock;
- only public `Graph` and `Vector` values require `Close`.

## Milestone 3: Fundamental algorithms

Goal: cover the graph analysis operations expected from a general-purpose
igraph binding.

Status: complete. The fundamental algorithm layer provides 15 user-facing
bindings over the Milestone 2 selector, conversion, and ownership primitives.
The algorithms share direction and weight validation, preserve public selector
order and duplicates, convert upstream failures to Go errors, and copy every
result out of temporary C storage while the graph lock is held.

Implemented areas:

- degree and neighborhood queries;
- connected components;
- BFS and DFS;
- shortest paths and distances;
- density, diameter, transitivity, and related structural properties.

Completion criteria:

- scalar, slice, and matrix results have stable Go representations;
- invalid selectors and disconnected graphs are tested;
- behavior is checked against small graphs with known answers.

Completion evidence:

- scalar undefined values use documented `NaN` or positive-infinity
  conventions, collection results are non-nil Go-owned slices, paths use an
  explicit `Found` flag, and dense distances use the immutable Go `Matrix`;
- focused and cross-feature tests cover invalid and duplicate selectors, empty
  selections, directed and undirected graphs, loops, weighted and unweighted
  calls, disconnected graphs, upstream failures, and calls after `Close`;
- known-answer tests cover every implemented area, including weak versus
  strong components, traversal forests, weighted shortest paths, and
  transitivity's undefined-value modes;
- all fallible algorithm and temporary-resource calls use centralized C
  wrappers that install and restore non-aborting thread-local igraph handlers.

## Milestone 4: Advanced algorithms

Goal: expand into specialized analysis after the common abstractions are
stable.

Planned areas:

- centrality;
- community detection;
- layouts;
- flows and cuts;
- isomorphism;
- random graph generators;
- domain-specific modules such as spatial, bipartite, and motif APIs.

These areas should be split into independent GitHub milestones when work begins.

## Definition of done for a binding

A binding is complete only when all of the following are true:

- the public API is idiomatic Go and does not expose C types;
- all C/igraph error codes are handled;
- ownership and `Close` behavior are explicit where applicable;
- nil, empty, invalid, directed, and undirected cases are considered;
- behavioral tests validate results, not only successful execution;
- statement coverage remains at or above the CI threshold;
- the upstream binding inventory is regenerated and checked;
- public API documentation explains semantics that differ from C/igraph.

## GitHub project management

Use repository documentation and GitHub features for different purposes:

- this document owns long-lived direction, ordering, and completion criteria;
- GitHub milestones group work for one roadmap milestone;
- issues describe reviewable feature slices and link to the milestone;
- pull requests implement one coherent slice and close its issue;
- the generated API report measures inventory progress.

This separation keeps project direction versioned with the code while GitHub
tracks the current execution state.
