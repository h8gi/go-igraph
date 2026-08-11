# Upstream API roadmap

This roadmap describes how go-igraph will grow from a small binding into a
coherent Go interface for C/igraph. It prioritizes useful, safe API slices over
maximizing a raw function count.

The generated [API coverage report](api-coverage.md) remains the inventory of
the pinned upstream release. GitHub milestones and issues should track the
execution of the work described here.

The [package architecture decision](package-architecture.md) defines why the
binding remains one public Go package and how to interpret completeness. Raw C
declaration coverage is an inventory dimension, not a product-completion
percentage.

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
configuration records composed, deferred, and deliberately unsupported
functions with a domain and rationale. The report distinguishes:

- user-facing bindings;
- internal implementation dependencies;
- composed APIs;
- intentionally unsupported functions;
- deferred declarations; and
- missing functions.

`Missing` now means that a declaration has not received a reviewed disposition.
A domain is complete when each relevant upstream declaration has a reviewed
disposition and its coherent Go-facing contract satisfies the definition of
done; its one-to-one binding percentage does not need to reach 100%. The
roadmap remains authoritative for the longer API and product rationale.

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

Status: Milestones 5 through 12 are complete. The later domain milestones
below remain candidates. Later milestone numbers express dependency order, not
a commitment to bind every function in the named upstream headers; each milestone must still
be split into reviewable issues with explicit API contracts.

The next stage should deepen the general-purpose graph API before expanding
into increasingly specialized domains. In particular, graph-returning
operations and stochastic algorithms need shared ownership, ID-mapping, and
reproducibility rules rather than one-off conventions in each binding.

### Milestone 5: Graph transformation and decomposition

Goal: make derived graphs and destructive edits safe building blocks for later
algorithms.

Status: complete. The transformation layer was delivered as a dependency-ordered
sequence of focused issues:

- shared derived-graph ownership, ID mappings, graph-list adoption, and stable
  multi-graph locking ([#66](https://github.com/h8gi/go-igraph/issues/66));
- selector-based atomic vertex and edge deletion
  ([#67](https://github.com/h8gi/go-igraph/issues/67));
- induced and edge subgraphs, decomposition into components, and independently
  owned graph results ([#68](https://github.com/h8gi/go-igraph/issues/68));
- atomic simplification and direction conversion
  ([#69](https://github.com/h8gi/go-igraph/issues/69));
- common graph operators with explicit provenance availability
  ([#70](https://github.com/h8gi/go-igraph/issues/70));
- articulation points, bridges, and biconnected structure
  ([#71](https://github.com/h8gi/go-igraph/issues/71)); and
- this integration, executable-example, ownership, failure-path, and
  documentation audit ([#72](https://github.com/h8gi/go-igraph/issues/72)).

Completion criteria:

- mutating names and graph-returning names distinguish receiver mutation from
  independently owned results, while selectors are materialized before every
  mutation;
- `RemovedID`, identity, many-to-one representatives, unavailable one-to-many
  mappings, exact upstream provenance, and deterministic structural mappings
  have one documented source-to-result contract;
- clone-and-swap mutations are atomic on validation, initialization, upstream,
  and conversion failure, and graph-list/nested-list partial results are cleaned
  at every ownership-transfer boundary;
- graph results survive source and sibling closure, mappings and nested
  collections are non-nil Go-owned values, and repeated `Close` is safe;
- multi-operand APIs deduplicate repeated graphs and lock distinct operands in
  deterministic order; and
- integration tests, race tests, generated inventory checks, and `make verify`
  pass against pinned igraph 1.0.1.

Completion evidence:

- deletion, subgraph, transformation, operator, and structural-decomposition
  suites cover directed and undirected graphs, loops, parallel edges,
  disconnected and empty graphs, empty and all-element selectors, invalid and
  closed inputs, upstream and initialization failures, and early returns;
- focused failure seams cover clone cleanup, mapping-vector cleanup, partial
  graph-list extraction, and partial nested-list conversion;
- integration pipelines combine deletion, independently owned subgraphs and
  components, simplification, direction conversion, graph operators,
  articulation points, bridges, and biconnected components, including source
  closure before result use;
- concurrent reversed and repeated operand tests run under the race detector;
- executable examples demonstrate deletion mappings and independently closed
  component graphs; and
- `make verify` enforces formatting, vet, behavioral tests, the statement
  coverage floor, inventory-tool tests, and generated inventory freshness.

### Milestone 6: Community structure

Goal: provide a coherent partition API rather than unrelated wrappers for
individual community-detection functions.

Status: complete. Delivered as a sequence of focused issues:

- partition types (`CommunityPartition`, `HierarchicalCommunity`), merge matrix
  conventions, continuous 0-indexed membership, and thread-safe C RNG seed locking ([#80](https://github.com/h8gi/go-igraph/issues/80));
- coreness, trussness, modularity calculation, and modularity matrix APIs ([#81](https://github.com/h8gi/go-igraph/issues/81));
- flat community detection algorithms (Multilevel/Louvain, Leiden, Label Propagation, Infomap, Fluid) ([#82](https://github.com/h8gi/go-igraph/issues/82));
- hierarchical community detection and dendrogram cuts (Walktrap, FastGreedy, Edge Betweenness) ([#83](https://github.com/h8gi/go-igraph/issues/83));
- spectral, simulated annealing, and exact optimization algorithms (Leading Eigenvector, Spinglass, SpinglassSingle, Optimal Modularity) ([#84](https://github.com/h8gi/go-igraph/issues/84));
- community partition comparison metrics (VI, NMI, Split-Join, Rand, Adjusted Rand) ([#85](https://github.com/h8gi/go-igraph/issues/85)); and
- this final contract audit, integration pipeline, executable examples, and documentation update ([#86](https://github.com/h8gi/go-igraph/issues/86)).

Completion criteria:

- all public APIs consistently communicate partition, dendrogram, and comparison behavior and expose no C graph, vector, matrix, enum, pointer, or RNG types;
- borrowed inputs, Go-owned result structs, continuous 0-indexed membership vectors, and reproducible RNG options are documented;
- failure-path tests cover initialization failure, upstream errors, invalid parameters, out-of-bounds steps, empty/degenerate graphs, repeated closure, and use after Close;
- integration pipelines combine partitioning, quality evaluation, dendrogram cut extraction, and partition comparison; and
- statement coverage remains above the CI threshold (>= 90.0%) and `make verify` passes.

### Milestone 7: Connectivity, flows, and cuts

Goal: cover network robustness and capacity analysis on top of the graph and
result ownership rules established by Milestone 5.

Status: complete. Delivered as a sequence of focused issues:

- core flow and cut result types (`MaxFlowResult`, `MinCutResult`, `STMinCutResult`), capacity validation, and max-flow / min-cut algorithms ([#94](https://github.com/h8gi/go-igraph/issues/94));
- edge and vertex connectivity, adhesion, cohesion, and edge-/vertex-disjoint paths ([#95](https://github.com/h8gi/go-igraph/issues/95));
- cut enumeration (`AllSTCuts` and `AllSTMincuts`) with safe multi-vector extraction ([#96](https://github.com/h8gi/go-igraph/issues/96));
- residual graphs, Gomory-Hu tree, dominator tree, and Tarjan reduction as independently owned graph results ([#97](https://github.com/h8gi/go-igraph/issues/97)); and
- this final contract audit, integration pipeline, executable examples, and documentation update ([#98](https://github.com/h8gi/go-igraph/issues/98)).

Completion criteria:

- capacity inputs follow the existing borrowed-input model (optional `[]float64` slice, where `nil` implies unit capacity `1.0` per edge, and non-nil slices are checked for length matching `g.NumEdges()` and non-negative values);
- result types distinguish scalar flow/cut values, edge flow vectors, cut edge sets, partition vertex sets, and tree graphs without exposing C storage or raw pointers;
- graph-returning flow operations (`ResidualGraph`, `ReverseResidualGraph`, `GomoryHuTree`, `DominatorTree`, `EvenTarjanReduction`) return independently owned `*Graph` instances that survive parent graph closure;
- error propagation covers invalid vertex IDs (source/target out of bounds), negative capacity values, invalid selector/vertex parameters, disconnected graphs, initialization failures, and use after `Close`;
- integration pipelines combine max flow, min cut, connectivity, residual graphs, Gomory-Hu trees, and dominator trees; and
- statement coverage remains above the CI threshold (>= 90.0%), generated inventory is updated, and `make verify` passes.

Completion evidence:

- core flow result structs expose flow values, edge flow vectors, cut vertex/edge partitions, and tree graph outputs;
- capacity slice validation strictly checks length matching `g.NumEdges()` and non-negative values, treating `nil` as unit capacity `1.0`;
- residual graphs, Gomory-Hu trees, dominator trees, and Even-Tarjan reductions return independently owned `*Graph` instances that survive parent closure;
- failure-path tests cover out-of-bounds sources/targets, negative capacities, invalid parameters, disconnected graphs, initialization failures, and use after `Close`;
- integration suite verifies max flow, min cut, connectivity, Gomory-Hu trees, and dominator trees;
- executable max-flow example demonstrates capacity networks and residual graphs; and
- `make verify` enforces formatting, vet, behavioral tests, statement coverage (>= 90.0%), and inventory report freshness.

### Milestone 8: Reproducible random graphs

Goal: apply and harden the stochastic-execution contract established for
Milestone 6 through a useful first slice of graph generators and random
transformations.

Status: complete. Delivered as a sequence of focused issues:

- Erdős-Rényi (G(n,m), G(n,p)), k-regular, and random tree generator APIs ([#111](https://github.com/h8gi/go-igraph/issues/111));
- degree-sequence graph generator and degree sequence validation APIs ([#112](https://github.com/h8gi/go-igraph/issues/112));
- Barabási-Albert preferential attachment, Watts-Strogatz small-world, and Stochastic Block Model (SBM) generator APIs ([#113](https://github.com/h8gi/go-igraph/issues/113));
- graph rewiring, random walk, and random spanning tree APIs ([#114](https://github.com/h8gi/go-igraph/issues/114)); and
- this final contract audit, integration pipeline, executable examples, and documentation update ([#115](https://github.com/h8gi/go-igraph/issues/115)).

Completion criteria:

- all random graph generators and stochastic operations accept an optional `Seed *uint64` option and use the package-wide thread-safe `withRNG` contract;
- generator parameters ($n \ge 0$, $m \ge 0$, $0 \le p \le 1$, degree sequences) are validated before calling C/igraph functions;
- graph generators return independently owned `*Graph` instances that survive source and caller scope without memory leaks;
- mutating operations (`Rewire`) are atomic on failure and leave receiver state unchanged if parameter validation fails;
- vector/path sampling operations (`RandomWalk`, `RandomSpanningTree`) return Go-owned slices and handle optional borrowed weight slices;
- failure-path tests cover invalid parameter domains, negative/overflow counts, non-graphical degree sequences, out-of-bounds start vertices, and use after `Close`;
- integration pipelines combine random graph generation, rewiring, random walks, and downstream graph analysis; and
- statement coverage remains above the CI threshold (>= 90.0%), generated inventory is updated, and `make verify` passes.

Completion evidence:

- all generator and sampling functions (`ErdosRenyiGNM`, `ErdosRenyiGNP`, `KRegularGame`, `RandomTreeGame`, `DegreeSequenceGame`, `BarabasiGame`, `WattsStrogatzGame`, `SBMGame`, `Rewire`, `RewireEdges`, `RandomWalk`, `RandomSpanningTree`) accept `Seed *uint64` and execute under the package-wide thread-safe `withRNG` contract;
- parameter validation (degree sequences, block sizes, integer bounds, overflow checks, float/probability bounds) rejects invalid inputs prior to calling C/igraph functions;
- in-place mutation (`Rewire`) is atomic on error, returning Go errors and preserving receiver state;
- graph generators and `RewireEdges` return Go-owned `*Graph` instances that survive source and caller scope;
- integration pipeline test suite and concurrent seed isolation tests under `-race` verify reproducibility and thread safety;
- executable example in `examples/random/` demonstrates reproducible random network generation and sampling; and
- `make verify` enforces formatting, vet, behavioral tests, statement coverage (>= 90.0%), and inventory report freshness.

### Milestone 9: Layouts and embeddings

Goal: return visualization-ready coordinates through the existing Go-owned
matrix boundary.

Status: complete. Delivered as a sequence of focused issues:

- layout coordinate contract and deterministic circle, star, grid, and random layout APIs ([#121](https://github.com/h8gi/go-igraph/issues/121));
- Reingold-Tilford tree, bipartite, and Sugiyama layered layout APIs ([#122](https://github.com/h8gi/go-igraph/issues/122));
- Fruchterman-Reingold, Kamada-Kawai, and MDS layout APIs ([#123](https://github.com/h8gi/go-igraph/issues/123));
- selected 3D layouts and adjacency/Laplacian spectral embedding APIs ([#124](https://github.com/h8gi/go-igraph/issues/124)); and
- final contract audit, integration pipeline, executable examples, and documentation update ([#125](https://github.com/h8gi/go-igraph/issues/125)).

Completion criteria:

- all layout and embedding APIs return Go-owned `Matrix` or `SpectralEmbeddingResult` values using the vertex-to-row convention (row `i` = vertex `i`) and explicit dimensionality;
- borrowed input slices (`order`, `roots`, `types`, `layers`, `Weights`) and options structs are strictly validated before calling C/igraph functions;
- stochastic layouts (`LayoutRandom`, `LayoutFruchtermanReingold`, `LayoutKamadaKawai`, `LayoutRandom3D`, `LayoutFruchtermanReingold3D`, `LayoutKamadaKawai3D`, spectral embeddings) accept an optional `Seed *uint64` and execute under the package-wide thread-safe `withRNG` contract;
- initial coordinates, distance matrices, and option bounds are copied or validated at the boundary, without exposing C option structs, solver objects, or raw C pointers;
- UMAP layouts are documented as intentionally unsupported until their solver contract can remain Go-native;
- failure-path tests cover invalid parameter domains, out-of-bounds vertex/order/layer selections, mismatched matrix dimensions, solver convergence failures, upstream errors, and use after `Close`;
- integration pipelines combine graph generation, layout computation, and spectral-embedding-based downstream analysis;
- race detector tests verify thread safety and seed isolation under concurrent layout calls; and
- statement coverage remains above the CI threshold (>= 90.0%), generated inventory is updated, and `make verify` passes.

Completion evidence:

- all layout functions (`LayoutCircle`, `LayoutStar`, `LayoutGrid`, `LayoutRandom`, `LayoutReingoldTilford`, `LayoutReingoldTilfordCircular`, `LayoutBipartite`, `LayoutSugiyama`, `LayoutFruchtermanReingold`, `LayoutKamadaKawai`, `LayoutMDS`, `LayoutRandom3D`, `LayoutGrid3D`, `LayoutSphere`, `LayoutFruchtermanReingold3D`, `LayoutKamadaKawai3D`) return Go-owned `Matrix` values with row `i` holding the coordinates of vertex `i` and explicit 2- or 3-column dimensionality, and the embeddings return `*SpectralEmbeddingResult` whose matrices and slices are Go-owned and survive graph closure;
- borrowed inputs are validated at the boundary: index slices (`order`, `roots`, `types`, `layers`) are length- and range-checked, `Weights` and `DegreeCorrection` are length- and finiteness-checked, per-axis bounds are length-checked and reject NaN (±Inf deliberately disables the bound on that side), initial coordinates and distance matrices are dimension-checked and copied, and no upstream enum, ARPACK object, or raw C pointer appears in any public signature;
- every stochastic entry point (`LayoutRandom`, `LayoutRandom3D`, `LayoutFruchtermanReingold`, `LayoutFruchtermanReingold3D`, `LayoutKamadaKawai`, `LayoutKamadaKawai3D`, `LayoutMDS`, `AdjacencySpectralEmbedding`, `LaplacianSpectralEmbedding`) accepts `Seed *uint64` and runs under the package-wide `withRNG` contract, including the ARPACK start vector of the spectral embeddings;
- UMAP layouts (`igraph_layout_umap`, `igraph_layout_umap_3d`, `igraph_layout_umap_compute_weights`) are recorded as intentionally unsupported in the generated inventory until their solver contract can remain Go-native;
- `TestMilestone9IntegrationPipeline` chains reproducible graph generation, deterministic circle seeding, bounded force-directed refinement, 3D layout, adjacency spectral embedding, `DimSelect`, and Laplacian embedding; `TestMilestone9ConcurrentSeedIsolation` runs seeded layout and embedding calls concurrently on independent graphs and requires exact agreement with serial references under `-race`;
- `example_layout_test.go` provides output-asserted package examples for the layout and embedding domain, and `examples/layout/main.go` demonstrates deterministic layouts and seed-reproducible force-directed layouts and embeddings; and
- the regenerated inventory reports the new bindings, and `make verify` passes with statement coverage at or above 90.0%.

### Milestone 10: Graph isomorphism and subgraph matching

Goal: provide coherent graph-isomorphism, subgraph-matching, canonical-labeling,
and automorphism APIs without exposing C callbacks, algorithm-specific storage,
or unbounded mapping collection.

Status: complete as a dependency-ordered sequence of focused issues:

- shared naming, operand, locking, ownership, and general isomorphism decision
  contracts ([#151](https://github.com/h8gi/go-igraph/issues/151));
- color-aware VF2 decisions and first mappings with explicit mapping directions
  ([#152](https://github.com/h8gi/go-igraph/issues/152));
- VF2 counts and explicitly bounded mapping enumeration with internal early
  termination ([#153](https://github.com/h8gi/go-igraph/issues/153));
- induced and non-induced LAD subgraph matching with validated optional domains
  ([#154](https://github.com/h8gi/go-igraph/issues/154));
- canonical labeling, independently owned canonical graphs, and automorphism
  generators and exact group sizes
  ([#155](https://github.com/h8gi/go-igraph/issues/155)); and
- the final contract audit, integration pipeline, executable examples, and
  documentation update ([#156](https://github.com/h8gi/go-igraph/issues/156)).

Completion criteria:

- multi-graph calls borrow their operands only for the synchronous call,
  deduplicate repeated graph pointers, and use the existing stable lock order so
  reversed and repeated operands cannot deadlock;
- operand roles and every mapping direction are explicit: names identify each
  mapping's index domain and value codomain, non-matches return non-nil empty
  mappings, and unmatched values use `RemovedID` only where the result contract
  defines it;
- optional vertex colors, edge colors, and LAD domains are validated and copied
  into temporary C storage, while returned mappings, nested mapping lists,
  canonical permutations, and automorphism generators are non-nil Go-owned
  values that survive input graph closure;
- exponential mapping enumeration requires an explicit positive bound and
  reports truncation; callback execution and normal early termination remain
  internal and no Go callback, C function pointer, Bliss structure, or raw
  solver option appears in the public API;
- canonical and automorphism operations reject multigraphs before invoking
  upstream functions that document unreliable multigraph results, and every
  independently returned canonical graph is caller-closed and survives source
  graph closure;
- initialization failure, upstream errors, integer conversion and count
  overflow, normal callback early stops, partial nested-list construction,
  empty and degenerate inputs, directedness mismatches, unsupported graph
  shapes, and use after `Close` are covered by focused tests; and
- integration and race tests, package and standalone examples, generated
  inventory checks, ownership documentation, and `make verify` pass against the
  pinned igraph release while statement coverage remains at or above 90.0%.

Execution order: #151 establishes the shared contract. #152 builds on it; #153
then reuses its colors and mapping types. #154 can proceed after #151 while
reusing the finalized mapping convention. #155 can proceed after #151 and the
shared nested-list ownership helpers are stable. #156 is the final audit.

Shared contract: methods name equal-size operands as `source` and `target`, and
subgraph operations name the smaller query graph `pattern` and the containing
graph `target`. A mapping name always states its direction; for example,
`PatternToTarget[i]` is the target vertex matched to pattern vertex `i`.
Non-matches return non-nil empty mapping slices. Reverse subgraph mappings are
indexed by target vertex and use `RemovedID` for unmatched target vertices.
Optional color inputs are accepted only when both corresponding operands
provide them, are borrowed and copied for the synchronous call, and are
validated against vertex or edge counts before entering C.

All multi-graph operations use `withLockedGraphs`: repeated graph pointers are
locked once, distinct graphs are locked in stable address order, and C graph
pointers are borrowed only until the callback returns. Scalar decisions and
all mapping, permutation, generator, and nested mapping results are Go-owned.
Enumeration APIs require an explicit positive bound and report whether further
matches existed. Public APIs do not expose algorithm-specific options, C
callbacks, function pointers, or C-owned storage.

Completion evidence:

- general decisions, color-aware VF2 first mappings and counts, bounded VF2
  callback enumeration, induced/non-induced LAD with copied domains, canonical
  labeling and graphs, automorphism generators, and exact `*big.Int` group
  sizes are covered by focused tests;
- `TestMilestone10IntegrationPipeline` combines canonicalization, color-aware
  matching, bounded subgraph enumeration, automorphism generators, exact group
  size, source closure, and continued canonical-graph use;
- `TestMilestone10RepeatedAndReversedConcurrency` covers repeated operands and
  reversed concurrent calls under the race detector;
- `example_isomorphism_test.go` provides output-asserted package examples and
  `examples/isomorphism/main.go` provides a standalone executable pipeline;
- failure tests cover vector-list initialization and partial conversion,
  invalid bounds, colors, domains and permutations, upstream directedness
  failures, normal callback early stop, empty and non-match results,
  unsupported graph shapes, and use after `Close`;
- all returned mappings, nested mappings, canonical permutations, generators,
  and exact sizes are Go-owned; canonical graphs are independently owned and
  caller-closed; and
- the regenerated inventory reports 166 user-facing bindings, including 10 of
  24 functions from `igraph_isomorphism.h`, and `make verify` passes with the
  statement coverage floor at 90.0%.

### Milestone 11: Cliques and independent vertex sets

Goal: provide coherent scalar, histogram, and explicitly bounded enumeration
APIs for cliques and independent vertex sets without exposing C callbacks, file
handles, sentinel bounds, or unbounded exponential result collection.

Status: complete as a dependency-ordered sequence of focused issues. The
shared contracts and scalar operations from #167 and bounded ordinary/largest
clique enumeration and histograms from #168, plus maximal-clique queries from
#169, positive-integer weighted queries from #170, and bounded independent-set
enumeration from #171, plus the cross-feature audit from #172, are implemented:

- this roadmap, shared contract, and issue-order plan
  ([#166](https://github.com/h8gi/go-igraph/issues/166));
- clique-family membership decisions, scalar extrema, and shared size-range and
  bounded-result types ([#167](https://github.com/h8gi/go-igraph/issues/167));
- bounded clique enumeration, largest-clique enumeration, and clique-size
  histograms ([#168](https://github.com/h8gi/go-igraph/issues/168));
- bounded maximal-clique enumeration, counts, histograms, and subset-seeded
  search ([#169](https://github.com/h8gi/go-igraph/issues/169));
- positive-integer weighted clique queries and bounded maximum-weight results
  ([#170](https://github.com/h8gi/go-igraph/issues/170));
- bounded ordinary, maximal, and largest independent vertex set enumeration
  ([#171](https://github.com/h8gi/go-igraph/issues/171)); and
- the final contract audit, integration pipeline, executable examples,
  inventory review, and documentation update
  ([#172](https://github.com/h8gi/go-igraph/issues/172)).

Completion criteria:

- all exponential collection APIs require an explicit positive result limit
  and report exact truncation by determining whether at least one additional
  matching result exists; upstream unlimited sentinels are never public;
- size and weight ranges use Go-native optional bounds with consistent inclusive
  semantics, reject inconsistent values before entering C, and do not overload
  zero with upstream's "unlimited" meaning;
- ordinary, largest, maximal, subset-seeded, weighted, and independent-set
  results share one documented bounded-enumeration shape; all returned nested
  slices and histograms are non-nil Go-owned values that survive graph closure;
- graph and slice/selector inputs are borrowed only for the synchronous call,
  explicit vertex IDs and positive integer vertex weights are validated and
  copied into temporary C storage, and no C pointer or storage escapes;
- direction, loop, parallel-edge, empty-graph, result-order, maximal-versus-
  maximum, and subset-seed semantics are explicit and tested against pinned
  igraph 1.0.1;
- callbacks used for exact early termination remain internal,
  `igraph_maximal_cliques_file` stays unsupported, and unbounded largest-result
  functions are composed behind the bounded public contract rather than exposed
  directly;
- initialization and upstream failures, integer/count conversion overflow,
  partial nested-list conversion, normal internal early stop, invalid ranges,
  limits, subsets, and weights, and use after `Close` have focused coverage; and
- integration and race tests, package and standalone examples, generated
  inventory checks, ownership documentation, and `make verify` pass against
  pinned igraph 1.0.1 while statement coverage remains at or above 90.0%.

Execution order: #166 establishes this plan. #167 defines the common public
vocabulary and scalar operations. #168 then establishes the reusable bounded
enumeration implementation. #169, #170, and #171 build independently on those
contracts, with #169 reusing subset materialization and #170 adding the stricter
positive-integer weight boundary. #172 is the final cross-feature audit.

Shared contract: collection-returning enumeration accepts an explicit positive
maximum result count and returns both a non-nil Go-owned `[][]int`-style value
and whether another result existed. Implementations may request one extra result
or stop an internal callback after the extra match, but must check the
`limit + 1` conversion for overflow and must not expose callback execution.

Final `igraph_cliques.h` disposition: the twelve bounded/scalar operations in
the generated inventory are user-facing. The callback and file-output APIs
(`igraph_cliques_callback`, `igraph_maximal_cliques_callback`, and
`igraph_maximal_cliques_file`) are intentionally not public because they would
expose callback or C/stdio ownership. The three unbounded collectors
(`igraph_largest_cliques`, `igraph_largest_weighted_cliques`, and
`igraph_largest_independent_vertex_sets`) are intentionally replaced by the
bounded composed APIs `LargestCliques`, `MaximumWeightCliques`, and
`LargestIndependentVertexSets`. No function in that header remains deferred.
Each returned vertex set is canonicalized when upstream ordering is not
guaranteed; outer enumeration order is not a compatibility promise unless the
pinned upstream API documents it.

Optional lower and upper size bounds have inclusive semantics and are distinct
from the required result limit. Weighted APIs expose positive integer weights
because pinned igraph 1.0.1 supports only positive integer weights and otherwise
truncates real inputs. A weight slice must contain exactly one value per vertex.
Subset maximal-clique search documents the upstream initial-vertex/search
semantics and does not present the subset as an induced-subgraph filter.

Graph inputs and input slices/selectors are borrowed only until the synchronous
call returns; explicit input values are copied into temporary C storage. Scalar
results, histograms, and all nested collections are Go-owned. Read-only methods
share the graph's existing lock and serialize safely with `Close`. Public APIs
expose no C vectors, callback/function pointers, file handles, or upstream
sentinel constants.

### Milestone 12: Cycle analysis

Goal: provide coherent acyclicity, cycle-witness, bounded simple-cycle,
cycle-basis, and feedback-set APIs without exposing C callbacks, negative
sentinels, unused parameters, or solver-specific enums.

Status: complete through a dependency-ordered sequence of focused issues:

- extend the generated inventory with composed and deferred domain dispositions
  ([#186](https://github.com/h8gi/go-igraph/issues/186));
- this roadmap, shared contract, upstream disposition, and issue-order plan
  ([#187](https://github.com/h8gi/go-igraph/issues/187));
- acyclicity predicates, topological ordering, one cycle witness, girth, and the
  shared cycle result vocabulary
  ([#188](https://github.com/h8gi/go-igraph/issues/188));
- explicitly bounded simple-cycle enumeration with exact truncation
  ([#189](https://github.com/h8gi/go-igraph/issues/189));
- fundamental and minimum cycle bases with explicit completeness, cutoff, and
  edge-ordering contracts
  ([#190](https://github.com/h8gi/go-igraph/issues/190));
- exact and approximate feedback edge sets plus exact feedback vertex sets
  ([#191](https://github.com/h8gi/go-igraph/issues/191)); and
- the final contract audit, integration pipeline, executable examples,
  inventory review, and documentation update
  ([#192](https://github.com/h8gi/go-igraph/issues/192)).

Completion criteria:

- `IsAcyclic` distinguishes the general directed/undirected predicate from
  `IsDAG`, which reports false for undirected graphs, while `TopologicalSort`
  accepts only incoming or outgoing orientation, rejects undirected graphs and
  non-loop directed cycles, and explicitly preserves pinned igraph's unusual
  behavior of ignoring self-loops even though the predicates report them as
  cycles;
- `FindCycle` and bounded simple-cycle enumeration return a shared `Cycle`
  shape whose non-nil Go-owned vertex and edge slices have equal lengths and
  matching traversal order, while acyclic/no-result cases use empty values
  rather than C sentinels;
- `Girth` documents that pinned igraph ignores direction, self-loops, and
  parallel-edge 2-cycles, and represents acyclic infinity and its empty witness
  without exposing C storage;
- every exponential simple-cycle collection requires an explicit positive
  result limit, checks `limit + 1` conversion, and reports exact truncation by
  observing one additional matching cycle;
- optional cycle-length and BFS cutoffs are Go-native values with inclusive or
  depth semantics stated explicitly; negative upstream unlimited sentinels are
  never public;
- fundamental-basis roots are range-checked, complete bases cover all weak
  components by default, cutoff-limited incomplete bases require an affirmative
  opt-in, and minimum-basis edge ordering is promised only when natural cycle
  order is explicitly requested;
- cycle-basis APIs omit the upstream `weights` parameters because pinned
  igraph 1.0.1 documents and implements them as unused, rather than creating a
  misleading public contract;
- feedback edge/vertex weights are borrowed and copied for the synchronous
  call, length-checked, and restricted to finite non-negative values; zero is
  valid, while negatives are rejected before entering C because pinned
  undirected feedback-arc behavior does not provide a coherent signed minimum
  objective; exact versus approximate feedback behavior is explicit;
- upstream implementation-specific IP enum values remain private, approximate
  feedback results are never described as minimum, and feedback-set validity
  is verified by deleting returned IDs from a copy and checking acyclicity;
- direction, self-loop, parallel-edge, empty, disconnected, experimental API,
  output-order, and use-after-`Close` behaviors are documented and tested against
  pinned igraph 1.0.1;
- initialization failures, upstream errors, early returns, checked integer
  conversions, partial paired/nested-list construction, and all cleanup paths
  have focused coverage; and
- integration and race tests, package and standalone examples, final domain
  dispositions, ownership documentation, generated inventory checks, and
  `make verify` pass while statement coverage remains at or above 90.0%.

Execution order: #186 establishes the disposition model and initial deferred
cycle inventory. #187 records this plan. #188 defines the shared public
vocabulary and non-enumerating foundation. #189 builds bounded paired-list
enumeration on that vocabulary. #190 may proceed after #187 while reusing the
nested edge-list conventions from #188/#189 where practical. #191 depends on
#188 so its output can be validated through the shared acyclicity contract.
#192 follows #189, #190, and #191 as the final cross-feature audit.

Shared cycle contract: `Cycle` contains corresponding `Vertices` and `Edges`
slices in traversal order; both slices and every nested collection are non-nil
and Go-owned. A bounded simple-cycle result contains `Cycles` plus `Truncated`.
Cycle start, orientation, and outer enumeration order are not compatibility
promises unless pinned igraph documents them. Methods use the existing
`DirectionMode`; undirected graphs ignore direction only where upstream does,
and modes that are meaningless for a specific operation return Go errors.

Simple-cycle length ranges use optional positive inclusive bounds. Enumeration
requires a positive maximum result count and internally requests one additional
cycle, never exposing `IGRAPH_UNLIMITED` or a callback. Graph/options inputs are
borrowed only until the synchronous call returns. Returned orders, witnesses,
cycles, bases, and feedback IDs remain valid after graph closure.

Cycle-basis results are nested edge-ID slices. Nil root/cutoff options request
all weak components and full computation. A cutoff cannot silently weaken the
zero-value contract: incomplete output must be explicitly allowed. Minimum
bases can request natural edge order around each cycle at its documented
performance cost; otherwise an element is an edge set, not an ordered path.

Feedback arc APIs expose Go-native automatic-exact and Eades-approximate
strategies, not the current IP backend constants. Feedback vertex sets expose
the sole exact operation without a meaningless strategy selector. Nil weights
mean unit weights; non-nil values must match the edge or vertex count and be
finite and non-negative. Zero weights are accepted, but negatives are rejected
before entering C: although pinned igraph accepts finite negatives, its
undirected maximum-spanning-forest-complement implementation does not define a
consistent global signed minimum. Values are copied into temporary C storage
and no C solver object or result storage escapes.

Final reviewed disposition: ten declarations are user-facing:
`igraph_is_acyclic`, `igraph_is_dag`, `igraph_topological_sorting`,
`igraph_find_cycle`, `igraph_girth`, `igraph_simple_cycles`,
`igraph_fundamental_cycles`, `igraph_minimum_cycle_basis`,
`igraph_feedback_arc_set`, and `igraph_feedback_vertex_set`.
`igraph_simple_cycles_callback`, the remaining ninth declaration from
`igraph_cycles.h`, is intentionally unsupported because the bounded Go-owned
collector provides early result limiting and exact truncation without exposing
or retaining a C callback. No declaration in the audited domain remains
deferred or accidentally missing. The experimental status of
`igraph_simple_cycles`, its callback, `igraph_fundamental_cycles`, and
`igraph_minimum_cycle_basis` in pinned igraph 1.0.1 is visible on the relevant
public methods and in this final audit.

## Milestone 13: Motifs and graphlets

Goal: Provide complete Go bindings for dyad census, triad census, triangle counting/listing, RANDESU motif counting/sampling, and graphlet decomposition/projection over `igraph_motifs.h` and `igraph_graphlets.h`.

Status: in progress. Initial inventory dispositions and API contract plan are established in [#200](https://github.com/h8gi/go-igraph/issues/200).

Implemented & Planned areas:

- Dyad census (`DyadCensus`) returning Go struct `DyadCensusResult{Mutual, Asymmetric, Null int64}` ([#201](https://github.com/h8gi/go-igraph/issues/201));
- Triad census (`TriadCensus`) returning a 16-element Go slice corresponding to Davis-Leinhardt 16 isomorphism classes ([#201](https://github.com/h8gi/go-igraph/issues/201));
- Triangle counting (`TrianglesCount`, `AdjacentTrianglesCount`) and listing (`TrianglesList` returning `[][3]int`) accepting `VertexSelector` ([#201](https://github.com/h8gi/go-igraph/issues/201));
- RANDESU motif counting (`MotifsRandesu`), stochastic estimation (`MotifsRandesuEstimate`), and total motif count (`MotifsRandesuNo`) for size 3 and 4 subgraphs with seed management under `withRNG` ([#202](https://github.com/h8gi/go-igraph/issues/202));
- Graphlet decomposition (`Graphlets`), candidate basis (`GraphletsCandidateBasis`), and projection (`GraphletsProject`) returning Go-owned clique and weight slices, supporting initial `Mu` handling (`startMu`), and validating weight vector bounds ([#203](https://github.com/h8gi/go-igraph/issues/203));
- Cross-feature integration testing, package and standalone examples, and final API audit ([#204](https://github.com/h8gi/go-igraph/issues/204)).

Execution order: #200 establishes the disposition model and initial deferred motif/graphlet inventory. #201 implements dyad, triad, and triangle APIs. #202 adds RANDESU motif counting and sampling. #203 provides graphlet basis, decomposition, and projection. #204 performs final audit, integration pipeline tests, and standalone examples.

Shared motif and graphlet contracts:
- `DyadCensus` returns Go struct `DyadCensusResult` with Go-owned `int64` fields.
- `TriadCensus` returns a Go-owned 16-element slice of `int64` representing Davis-Leinhardt classes.
- `TrianglesList` returns Go-owned `[][3]int` matrix layout without exposing raw C vectors.
- RANDESU motif counting validates subgraph sizes (3 or 4) and cut probability length matches size. Stochastic sampling options take optional `Seed *uint64` and execute safely within `withRNG`.
- `igraph_motifs_randesu_callback` is marked as `intentionally_unsupported` to prevent unsafe C callbacks across the cgo boundary.
- Graphlet APIs accept optional weights validated against `NumEdges()`, copy clique indices into Go-owned slices, and handle optional initial `Mu` vectors correctly (`startMu`).

Final reviewed disposition: All 12 declarations across `igraph_motifs.h` and `igraph_graphlets.h` are accounted for (11 user-facing bindings and 1 intentionally unsupported callback).

### Later domain milestones

Motifs and graphlets; bipartite and spatial analysis;
attributes and richer import/export; and other specialized upstream domains
remain candidates after the milestones above. They should advance when a
concrete use case can define a coherent Go API and its resource model, not
merely to increase the inventory percentage.

Select the next domain by user value, shared-infrastructure readiness, and the
ability to define a complete Go ownership and concurrency contract. Motifs and
graphlets should follow cycle analysis separately, after their sampling,
callback, weighted, and RNG contracts are designed.

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
