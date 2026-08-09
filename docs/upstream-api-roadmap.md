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

## Milestone 4: Centrality algorithms

Goal: deliver the first advanced-algorithm slice as a coherent centrality API
over the selector, direction, weight, matrix, error, and ownership contracts
established by Milestones 2 and 3.

Status: complete. The centrality layer adds 22 user-facing bindings and reuses
the selector, direction, optional-weight, error, locking, and ownership
contracts established by Milestones 2 and 3.

Implemented areas:

- closeness and harmonic centrality, including explicit cutoff and
  reachability semantics ([#52](https://github.com/h8gi/go-igraph/issues/52));
- vertex and edge betweenness, including normalization and cutoff behavior
  ([#53](https://github.com/h8gi/go-igraph/issues/53));
- eigenvector centrality, HITS, PageRank, and personalized PageRank behind a
  Go-native solver boundary ([#54](https://github.com/h8gi/go-igraph/issues/54));
- graph-level degree, betweenness, closeness, and eigenvector centralization
  after their node-level contracts stabilize
  ([#55](https://github.com/h8gi/go-igraph/issues/55));
- a final integration, ownership, failure-path, and documentation audit
  ([#56](https://github.com/h8gi/go-igraph/issues/56)).

Completion criteria:

- vertex and edge results preserve materialized selector order, including
  duplicates, and every returned collection is a non-nil Go-owned value;
- direction, directed-path, weight, cutoff, normalization, reachability,
  reset-distribution, and undefined-value semantics are explicit and
  consistent across the API;
- ARPACK, PRPACK, and other upstream solver types do not appear in the public
  API, and solver defaults, convergence, warnings, and errors have documented
  Go behavior;
- initialization failure, upstream errors, warning paths, early returns,
  empty and degenerate graphs, disconnected graphs, invalid inputs, and use
  after `Close` release all temporary C resources and have focused coverage;
- known-answer or invariant-based tests cover directed and undirected,
  weighted and unweighted, normalized and raw, cutoff and unlimited, and
  personalized and non-personalized calculations;
- public documentation states whether inputs are borrowed or copied and that
  returned values are Go-owned; and
- the binding inventory is regenerated and reviewed, statement coverage stays
  above the CI threshold, and `make verify` passes.

Completion evidence:

- closeness, harmonic, vertex-betweenness, edge-betweenness, and PageRank
  results preserve selector order and duplicates while returning non-nil,
  Go-owned slices;
- cutoffs use explicit optional non-negative values, centrality families share
  positive or non-negative weight validators as required, and PageRank reset
  representations are mutually exclusive;
- eigenvector, HITS, and PageRank expose only Go solver settings, default to the
  recommended PRPACK PageRank backend, and never expose a C solver object;
- graph centralization returns node scores, raw or normalized values, and
  theoretical maxima without exposing null-graph or raw-node-count conventions;
- focused and cross-feature tests cover known answers, invariants, directed and
  undirected graphs, weighted and unweighted calls, cutoff boundaries,
  personalized resets, disconnected and degenerate graphs, initialization and
  upstream failures where testable, warning paths, and use after `Close`;
- package examples demonstrate selector-ordered distance centrality and
  personalized PageRank; and
- `make verify` checks the regenerated inventory, statement-coverage floor,
  vet, formatting, and the full behavioral suite against pinned igraph 1.0.1.

## Roadmap after Milestone 4

Status: proposed. The milestone numbers below express the intended dependency
order, not a commitment to bind every function in the named upstream headers.
Before implementation starts, each milestone should be split into reviewable
issues with an explicit API contract and binding target.

The next stage should deepen the general-purpose graph API before expanding
into increasingly specialized domains. In particular, graph-returning
operations and stochastic algorithms need shared ownership, ID-mapping, and
reproducibility rules rather than one-off conventions in each binding.

### Milestone 5: Graph transformation and decomposition

Goal: make derived graphs and destructive edits safe building blocks for later
algorithms.

Planned areas:

- vertex and edge deletion using the existing selectors;
- induced and edge subgraphs, decomposition into components, and independently
  owned graph results;
- simplification, direction conversion, and commonly used graph operators;
- articulation points, bridges, and biconnected structure; and
- explicit old-to-new vertex and edge mappings whenever an operation can
  renumber IDs.

The milestone must define atomic failure behavior for mutations, materialize
selectors before modifying a graph, and clean up every partially initialized
graph in multi-result operations. Returned graphs must be independently owned,
and mappings must remain valid after their source graphs are closed.

### Milestone 6: Community structure

Goal: provide a coherent partition API rather than unrelated wrappers for
individual community-detection functions.

Planned areas:

- partition quality and structure metrics such as modularity and coreness;
- an initial representative set of flat community algorithms, including
  multilevel or Leiden and label propagation;
- hierarchical algorithms such as fast-greedy, walktrap, or edge
  betweenness, after a common merge representation is defined; and
- Go-owned partition and hierarchy results with membership, group sizes,
  quality values, and algorithm-specific diagnostics where meaningful.

Weighted, directed, resolution, initial-membership, and fixed-membership
semantics must be explicit. Before exposing a randomized algorithm, the
package must also define a shared stochastic-execution contract covering how
seeding, concurrent calls, and C/igraph's RNG state interact; reproducibility
must not depend on an undocumented global side effect.

### Milestone 7: Connectivity, flows, and cuts

Goal: cover network robustness and capacity analysis on top of the graph and
result ownership rules established by Milestone 5.

Planned areas:

- edge and vertex connectivity;
- maximum flow and minimum cuts;
- edge- and vertex-disjoint paths;
- source-target cut enumeration where bounded result handling is practical;
  and
- Gomory-Hu trees and residual graphs as independently owned graph results.

Capacity inputs should follow the existing borrowed-input and validation
model. Result types must distinguish values, partitions, cut edges, and flow
vectors without exposing C storage, and tests must cover directedness,
parallel edges, loops, zero capacities, disconnected graphs, and partial
initialization failures.

### Milestone 8: Reproducible random graphs

Goal: apply and harden the stochastic-execution contract established for
Milestone 6 through a useful first slice of graph generators and random
transformations.

Planned areas:

- the shared seed and reproducibility policy under concurrent generator calls;
- Erdős-Rényi, degree-sequence, and preferential-attachment families;
- rewiring and sampling operations; and
- validation for graphical sequences, size overflow, and model-specific
  parameter domains.

The public API should expose model concepts rather than C RNG objects. Equal
inputs and seeds must have a documented reproducibility scope, and failure or
an interrupted call must not leave package-wide random state in a surprising
state.

### Milestone 9: Layouts and embeddings

Goal: return visualization-ready coordinates through the existing Go-owned
matrix boundary.

Planned areas:

- deterministic circle, grid, star, tree, and bipartite layouts;
- representative force-directed and distance-based 2D layouts;
- selected 3D layouts and spectral or manifold embeddings when their solver
  contracts can remain Go-native; and
- copied seed coordinates, bounds, weights, and fixed-vertex inputs.

Coordinate matrices must use one documented vertex-to-row convention and have
explicit dimensionality. Iteration limits, convergence, warnings, and empty or
degenerate graphs need the same error and ownership treatment as Milestone 4,
while stochastic layouts must reuse the shared reproducibility contract from
Milestone 6.

### Later domain milestones

Isomorphism and subgraph matching; cliques, cycles, motifs, and graphlets;
bipartite and spatial analysis; attributes and richer import/export; and other
specialized upstream domains remain candidates after the milestones above.
They should advance when a concrete use case can define a coherent Go API and
its resource model, not merely to increase the inventory percentage.

Across all future milestones, interruption and progress callbacks should stay
internal until they can cross the Go/C boundary without weakening concurrency,
cleanup, or error guarantees. Each milestone should finish with a cross-feature
audit, examples, regenerated inventory, and `make verify`, following the same
pattern as Milestones 3 and 4.

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
