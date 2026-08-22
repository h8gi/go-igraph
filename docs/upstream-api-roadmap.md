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

Status: Milestones 5 through 15 are complete. The later domain milestones
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

Status: complete. All planned bindings are implemented, and integration, examples, ownership/race coverage, documentation, and inventory audits were completed in [#204](https://github.com/h8gi/go-igraph/issues/204).

Implemented areas:

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
- RANDESU motif counting validates subgraph sizes (3 or 4), finite cut probabilities in `[0, 1]`, and cut probability length matching size. Histograms are Go-owned `[]float64` values that preserve upstream NaN markers for impossible isomorphism classes. Estimation accepts either a positive random sample size or a non-empty unique explicit vertex selector, returns a possibly fractional estimate, and executes stochastic work safely within `withRNG` with optional `Seed *uint64`.
- `igraph_motifs_randesu_callback` is marked as `intentionally_unsupported` to prevent unsafe C callbacks across the cgo boundary.
- Graphlet APIs treat nil weights as unit weights and otherwise require one finite non-negative value per edge. Candidate bases, thresholds, and projected coefficients are copied into aligned Go-owned slices. Projection validates and copies complete, unique clique inputs and sets `startMu` only for a non-empty, length-matched initial coefficient vector.

Final reviewed disposition: All 12 declarations across `igraph_motifs.h` and `igraph_graphlets.h` are accounted for (11 user-facing bindings and 1 intentionally unsupported callback), with no deferred declarations. Package and standalone examples exercise both motif and graphlet workflows, and the integration pipeline verifies aligned Go-owned results through graph closure.

## Milestone 14: Bipartite networks and matching

Goal: construct, validate, convert, project, generate, and match bipartite
networks through one explicit partition contract, with source provenance and
Go-owned graph and value results.

Status: complete. The reference workflows, shared contracts, and
dependency-ordered issue plan established in
[#211](https://github.com/h8gi/go-igraph/issues/211) are implemented and
audited through [#217](https://github.com/h8gi/go-igraph/issues/217).

Reference workflows:

- affiliation and recommendation networks construct a user-item or
  author-paper graph from a biadjacency matrix, recover the row and column
  source IDs, project either mode, and retain projection multiplicities; and
- assignment networks validate a supplied partition and compute unweighted or
  weighted maximum matchings without exposing upstream unmatched sentinels.

Delivered areas:

- explicit partition discovery, validation, and deterministic construction
  ([#212](https://github.com/h8gi/go-igraph/issues/212));
- unweighted and weighted biadjacency construction and matrix conversion with
  explicit row and column vertex-ID mappings
  ([#213](https://github.com/h8gi/go-igraph/issues/213));
- one-mode and two-mode projection with independently owned graphs, source
  vertex mappings, and edge multiplicities
  ([#214](https://github.com/h8gi/go-igraph/issues/214));
- matching validation and maximum cardinality or weighted matching through
  Go-owned result types
  ([#215](https://github.com/h8gi/go-igraph/issues/215));
- reproducible random bipartite graph generators under the existing seeded RNG
  isolation contract
  ([#216](https://github.com/h8gi/go-igraph/issues/216)); and
- the final contract audit, integration pipeline, executable examples,
  ownership and race coverage, documentation, and inventory update
  ([#217](https://github.com/h8gi/go-igraph/issues/217)).

Execution order: #211 establishes the domain dispositions, reference
workflows, shared contracts, and issue plan. #212 defines the explicit
partition vocabulary used by every later slice. #213, #214, #215, and #216 may
then proceed independently where their implementation does not require another
slice's result type. #217 follows all implementation issues and removes every
stale deferred disposition in the final audit.

Shared bipartite contract: a partition is explicit Go-owned data aligned with
vertex IDs; it is not hidden graph attribute state and does not introduce a
second graph lifetime model. Public partition, matrix, weight, and options
inputs are borrowed only for the synchronous call and copied before C retains
or mutates them. Returned partitions, mappings, multiplicities, matchings,
matrices, and weights are non-nil Go-owned values that remain valid after graph
closure. Every returned graph is independently owned, survives source and
sibling closure, and must be closed by the caller.

Biadjacency results include row and column source vertex IDs so matrix position
is never an implicit provenance contract. Weighted construction returns weights
aligned with the resulting edge IDs. Projection results identify the source
vertex corresponding to every projected vertex and align multiplicities with
projected edge IDs. Matching results use Go-native pairs or equivalent result
values instead of exposing negative upstream mate sentinels; output order is
not a compatibility promise unless pinned igraph documents it.

Partition length and validity, matrix dimensions and entries, weight length and
finiteness, integer conversion, directed orientation, loops, parallel edges,
empty modes, isolated vertices, and disconnected graphs must be validated or
documented for each operation. Stochastic constructors use optional
`Seed *uint64` values under `withRNG`, require exact replay for equal seeds, and
must not interfere across concurrent calls. No C callback, attribute table,
solver object, graph storage, or temporary vector escapes the synchronous call.

Completion criteria:

- both reference workflows are expressible without C types, hidden attributes,
  or caller-inferred vertex ordering;
- initialization failure, upstream error, early return, checked conversion,
  and partial graph/vector construction clean up all owned resources;
- empty, invalid, directed, undirected, loop, parallel-edge, disconnected,
  weighted, seeded, and use-after-`Close` behaviors are documented and tested;
- graph results and all aligned Go-owned values survive source and sibling
  closure, while repeated `Close` remains safe;
- seeded isolation, concurrent reads, and graph closure are covered under the
  race detector; and
- every declaration in `igraph_bipartite.h` and `igraph_matching.h` has a final
  reviewed disposition, examples cover both workflows, and `make verify`
  passes while preserving the statement coverage floor.

Final reviewed disposition: all 14 declarations across `igraph_bipartite.h`
and `igraph_matching.h` are user-facing bindings, with no deferred or
unsupported declarations. Package examples and `examples/bipartite/` cover
affiliation projection and assignment matching, while the integration pipeline
also exercises matrix round trips, weighted and unweighted matching, seeded
generation, bipartite layout, ownership across closure, and race-safe access.

## Milestone 15: Paths, reachability, and routing

Goal: deepen the general-purpose graph API with finite and explicitly bounded
route analysis, distance-derived metrics, reachability and derived graphs,
widest/Voronoi/spanner workflows, and Eulerian traversals.

Status: complete. The reference workflows, shared contracts, initial deferred
inventory, and dependency-ordered issue plan were established in
[#225](https://github.com/h8gi/go-igraph/issues/225).

Reference workflows:

- route analysis computes selector-ordered shortest routes, bounded alternative
  routes, cutoff distances and summaries, widest routes, and a sparse spanner
  while retaining aligned vertex and edge identities; and
- directed reachability computes reachable sets and counts, constructs
  independently owned neighborhood and transitive-closure graphs with explicit
  source provenance, partitions vertices by graph distance, and produces an
  Eulerian traversal when one exists.

Planned areas:

- finite batch shortest paths and explicitly bounded k-shortest and simple-path
  enumeration with aligned Go-owned vertex and edge sequences
  ([#226](https://github.com/h8gi/go-igraph/issues/226));
- cutoff distances, algorithm selection, eccentricity, radius, center,
  pseudo-diameter, efficiency, and path-length histograms
  ([#227](https://github.com/h8gi/go-igraph/issues/227));
- Go-owned reachable sets and counts plus independently owned neighborhood and
  transitive-closure graphs
  ([#228](https://github.com/h8gi/go-igraph/issues/228));
- widest paths and widths, graph Voronoi partitioning, and spanners with
  explicit provenance ([#229](https://github.com/h8gi/go-igraph/issues/229));
- Eulerian existence checks and aligned vertex/edge traversals
  ([#230](https://github.com/h8gi/go-igraph/issues/230)); and
- final integration, examples, documentation, race/ownership coverage, and
  inventory audit ([#231](https://github.com/h8gi/go-igraph/issues/231)).

Execution order: #225 establishes the initial dispositions and shared
contracts. #226, #227, #228, and #230 may then proceed independently. #229
reuses the distance and reachability decisions from #227 and #228. #231 follows
all implementation slices and removes every stale deferred disposition.

Shared path and reachability contract: `Path`, `PathOptions`,
`VertexSelector`, immutable `Matrix`, edge-weight validation, checked integer
conversion, stable graph locking, and independently owned graph results remain
the common vocabulary. Public slice, selector, matrix, weight, capacity,
generator, and option inputs are borrowed only for the synchronous call and
copied before C retains or mutates them. Returned paths, matrices, partitions,
histograms, mappings, and nested collections are non-nil Go-owned values that
survive graph closure. Returned neighborhood, closure, or spanner graphs are
independently owned, survive source and sibling closure, and must be closed by
the caller.

Finite batch operations preserve materialized selector order and duplicates.
Every path result aligns its vertex and edge sequences and represents
unreachability without leaking upstream negative sentinels. Potentially
exponential enumeration requires explicit non-negative limits before entering
C. The unbounded all-shortest-path declarations remain deferred until #226
proves that allocation can be bounded before materialization; otherwise the
final audit records them as intentionally unsupported. Algorithm-specific
shortest-path, distance, and widest-path declarations should normally be
composed behind automatic Go selection rather than exposed as parallel public
methods.

The #226 path-enumeration slice resolves that allocation review:
`ShortestPaths` preserves finite target selector order and duplicates,
`KShortestPaths` requires an explicit positive count, and `SimplePaths`
requires an explicit positive result limit and probes one additional result to
report truncation. The two unbounded all-shortest-path declarations are
intentionally unsupported, A* is intentionally unsupported because its
heuristic callback would cross the cgo boundary, algorithm-specific
shortest-path variants are composed behind `ShortestPath` or `ShortestPaths`,
and low-level path conversion helpers are composed by the aligned Go-owned
`Path` result.

The #227 distance-metrics slice adds selector-ordered cutoff matrices,
eccentricities, radius, the experimental upstream graph-center operation,
pseudo-diameter estimates with explicit start and seed controls, global and
local efficiency, and Go-owned path-length histograms. Weighted operations
require finite non-negative edge lengths. Algorithm-specific distance
declarations are composed behind `Distances` and `CutoffDistances`; callers do
not select Bellman-Ford, Dijkstra, Johnson, or Floyd-Warshall directly.

The #228 reachability slice converts upstream component-indexed bitsets into
vertex-indexed Go-owned reachable sets, adds per-vertex reachable counts, and
returns independently owned neighborhood and transitive-closure graphs.
Neighborhood results preserve root selector order and duplicates and include a
Go-owned mapping from each result vertex back to its source vertex ID.

The #229 routing slice adds aligned widest paths and width matrices, Voronoi
partitions with deterministic or seeded-random tie breaking, and independently
owned spanner graphs with source-edge provenance. Width inputs are required and
finite; Voronoi and spanner lengths are finite and non-negative.

The #230 Eulerian slice adds existence queries and aligned Go-owned path and
cycle traversals. A missing traversal is represented by `Found == false` with
non-nil empty vertex and edge slices rather than as an error.

Reachability results do not expose C bitsets. Derived graph results include
source vertex or edge provenance wherever IDs can change and such provenance
is meaningful. Nil weights select unweighted calculation; non-nil weights or
capacities are copied, edge-aligned, finite, and follow operation-specific sign
constraints. Cutoffs, limits, stretch factors, unreachable values, Voronoi
ties, directed modes, empty values, loops, parallel edges, and disconnected
graphs have explicit Go-native semantics. Random Voronoi tie-breaking, if
exposed, uses optional `Seed *uint64` under `withRNG`; deterministic modes do
not touch package RNG state. No C callback, bitset, iterator, graph storage, or
temporary vector/list escapes the synchronous call.

Initial disposition: the seven declarations already covered in these headers
remain user-facing. The other 39 declarations are deferred to #226–#230. This
includes algorithm variants and low-level path conversion helpers whose final
status may be composed or internal, A* whose callback boundary requires a
safety decision, and unbounded all-shortest-path declarations whose allocation
contract requires explicit review. The final audit must resolve every deferred
entry.

Final reviewed disposition: every declaration in `igraph_paths.h`,
`igraph_reachability.h`, `igraph_neighborhood.h`, and `igraph_eulerian.h` has a
non-deferred disposition. The bounded Go APIs cover the two reference
workflows; algorithm variants and low-level conversions are composed, while
unbounded all-shortest-path materialization and the A* callback boundary are
intentionally unsupported. Package and standalone examples cover routing and
reachability, and the integration pipeline verifies ownership, race safety,
the generated inventory, and the statement coverage floor.

Completion criteria:

- both reference workflows are expressible without C types, caller-inferred ID
  ordering, negative sentinels, or unbounded result allocation;
- initialization failure, upstream error, early return, checked conversion,
  and partial nested-vector or multi-graph construction clean up all resources;
- empty, invalid, directed, undirected, loop, parallel-edge, disconnected,
  unreachable, weighted, capacity, tied, bounded, and use-after-`Close`
  behavior is documented and tested;
- graph results and all aligned Go-owned values survive source and sibling
  closure, while repeated `Close` remains safe;
- seeded isolation where applicable, concurrent reads, and graph closure are
  covered under the race detector; and
- every declaration in `igraph_paths.h`, `igraph_reachability.h`,
  `igraph_neighborhood.h`, and `igraph_eulerian.h` has a final reviewed
  disposition, examples cover both workflows, and `make verify` passes while
  preserving the statement coverage floor.

## Milestone 16: Spatial graphs and geometry

Goal: construct spatial graphs from point-coordinate matrices, compute
edge-aligned spatial lengths, and extract two-dimensional convex hulls through
Go-owned graph and value results.

Status: complete. The reference workflows, shared contracts, initial deferred
inventory, and dependency-ordered issue plan were established in
[#239](https://github.com/h8gi/go-igraph/issues/239), and implementation and
final audit were completed through #240–#245.

Reference workflows:

- spatial routing constructs a nearest-neighbor or proximity graph from a
  point matrix, computes edge lengths in edge-ID order, and passes those
  lengths directly to weighted path, centrality, or routing APIs; and
- planar geometry extracts an index-preserving convex hull and compares
  Delaunay, Gabriel, relative-neighborhood, and beta-skeleton graphs over the
  same point rows.

Planned areas:

- shared point-matrix validation, distance-metric vocabulary, optional bound
  semantics, graph adoption, and aligned-result contracts
  ([#240](https://github.com/h8gi/go-igraph/issues/240));
- aligned two-dimensional convex hulls and spatial edge lengths
  ([#241](https://github.com/h8gi/go-igraph/issues/241));
- directed or undirected nearest-neighbor graph construction with optional
  neighbor-count and distance bounds
  ([#242](https://github.com/h8gi/go-igraph/issues/242));
- Delaunay, Gabriel, and relative-neighborhood graph construction
  ([#243](https://github.com/h8gi/go-igraph/issues/243));
- lune- and circle-based beta skeletons plus beta-weighted Gabriel graphs
  ([#244](https://github.com/h8gi/go-igraph/issues/244)); and
- final integration, examples, documentation, race/ownership coverage, and
  inventory audit ([#245](https://github.com/h8gi/go-igraph/issues/245)).

Execution order: #239 establishes the initial dispositions and milestone-wide
contract. #240 defines the common metric, validation, graph-adoption, and
aligned-result vocabulary. #241, #242, and #243 may then proceed independently.
#244 follows #240 and reuses the graph-plus-aligned-values result contract from
#241 where practical. #245 follows all implementation slices and removes every
stale deferred disposition.

Shared spatial contract: an immutable `Matrix` represents a point set, with row
`i` corresponding to vertex ID `i` and columns representing spatial dimensions.
Matrix inputs are borrowed only for the synchronous call and copied into
temporary C storage; no C pointer or matrix storage escapes. Public metrics are
Go-native Euclidean and Manhattan choices rather than an exposed C enum.
Operation-specific dimension, duplicate-point, finite-coordinate, neighbor,
cutoff, and beta constraints are validated explicitly.

Spatial graph constructors return independently owned `Graph` values that must
be closed by the caller and remain valid independently of the input matrix and
other results. Vertex IDs retain point-row identity. Graph edge order is not a
compatibility promise, but every accompanying edge-value slice is aligned with
the returned graph's edge IDs. `SpatialEdgeLengths` similarly returns one
Go-owned value per receiver edge in edge-ID order. Convex-hull point indices and
coordinate rows are aligned, non-nil, Go-owned values.

Optional nearest-neighbor limits use Go option values rather than upstream
negative sentinels. A missing maximum-neighbor count or cutoff means unbounded
for that dimension of the search; explicit counts and cutoffs are
non-negative. Beta-skeleton parameters are positive and finite. A missing
weighted-Gabriel maximum beta requests the upstream unlimited calculation,
while positive infinity in a returned threshold is a valid result indicating
that the edge persists beyond the requested search range.

Eight declarations in `igraph_spatial.h` are experimental in pinned igraph
1.0.1; only `igraph_convex_hull_2d` is not. This status is stated on every
corresponding public API. Experimental status does not relax ownership, error,
cleanup, validation, or behavioral-test requirements.

Initial disposition: all nine currently missing declarations in
`igraph_spatial.h` are deferred to #241–#244. The final audit must resolve each
as user-facing, composed, internal, or intentionally unsupported and leave no
stale deferred or accidentally missing declaration in the audited domain.

Final reviewed disposition: all nine declarations in `igraph_spatial.h` are
user-facing bindings, with no deferred, composed, internal, or intentionally
unsupported declarations in the domain. Package examples and
`examples/spatial/` cover weighted spatial routing and planar geometry, while
the milestone integration pipeline verifies hull/proximity invariants,
edge-ID-aligned values, ownership across closure, concurrent construction, and
race-safe spatial edge-length reads. Generated coverage reports 9/9 spatial
declarations with a final disposition and zero deferred declarations overall.

Completion criteria:

- both reference workflows are expressible without C types, negative
  sentinels, hidden attributes, or caller-inferred point/vertex/edge alignment;
- point matrices, options, and metric inputs are borrowed only for synchronous
  calls, while graphs, hulls, lengths, weights, and aligned collections follow
  explicit independent ownership rules;
- initialization failure, upstream error, checked conversion, partial-result
  construction, and early-return paths release all temporary C resources;
- empty, invalid, non-finite, duplicate, degenerate, multidimensional, tied,
  directed, undirected, loop, parallel-edge, bounded, and use-after-`Close`
  behavior is documented and tested where relevant;
- experimental upstream status and any absence of stable edge-order guarantees
  are visible in public documentation;
- race tests cover concurrent read-only edge-length calculations and closure,
  and graph constructors require no package-global mutable state; and
- every declaration in `igraph_spatial.h` has a final reviewed disposition,
  examples cover both workflows, and `make verify` passes while preserving the
  statement coverage floor.

## Milestone 17: Attributes and graph interchange

Goal: attach typed metadata to graphs, vertices, and edges, preserve or
explicitly combine that metadata across graph transformations, and exchange
attributed graphs through practical file formats without exposing C attribute
tables or file handles.

Status: complete. The reference workflows, shared contracts, initial
deferred inventory, and dependency-ordered issue plan were established in
[#253](https://github.com/h8gi/go-igraph/issues/253). The shared typed
vocabulary, process-global attribute runtime, checked metadata conversion, and
record/list cleanup boundary are established in
[#254](https://github.com/h8gi/go-igraph/issues/254). Graph-, vertex-, and
edge-level typed attributes, transformation combination policies, readers, and
writers are implemented through [#259](https://github.com/h8gi/go-igraph/issues/259).
[#260](https://github.com/h8gi/go-igraph/issues/260) completes the composed
attributed-interchange workflow, examples, ownership and concurrency review,
and final inventory audit.

Reference workflows:

- attributed analysis attaches boolean, numeric, and string metadata to a
  graph, its vertices, and its edges; inspects and updates those values; then
  copies, simplifies, decomposes, or combines graphs with explicit attribute
  preservation and combination semantics; and
- attributed interchange reads a GraphML or GML graph, inspects and modifies
  its topology and metadata through the same typed APIs, writes it, and reads
  it back without undocumented loss of supported information.

Planned areas:

- shared attribute scope/type vocabulary, checked string and metadata
  conversion, one-time C attribute-table installation, and cleanup seams
  ([#254](https://github.com/h8gi/go-igraph/issues/254));
- graph-level boolean, numeric, and string metadata
  ([#255](https://github.com/h8gi/go-igraph/issues/255));
- vertex- and edge-ID-aligned boolean, numeric, and string metadata
  ([#256](https://github.com/h8gi/go-igraph/issues/256));
- explicit preservation and Go-native combination policies across graph
  transformations and operators
  ([#257](https://github.com/h8gi/go-igraph/issues/257));
- independently owned edge-list, GraphML, and GML graph readers plus reviewed
  dispositions for the remaining upstream readers
  ([#258](https://github.com/h8gi/go-igraph/issues/258));
- richer graph writers, format capability contracts, and attributed
  round-trip coverage
  ([#259](https://github.com/h8gi/go-igraph/issues/259)); and
- final integration, examples, documentation, race/ownership coverage, and
  inventory audit ([#260](https://github.com/h8gi/go-igraph/issues/260)).

Execution order: #253 establishes the initial dispositions and milestone-wide
contract. #254 defines the shared runtime, type, conversion, and cleanup
vocabulary. #255 and #256 may then proceed independently. #257 follows both
typed attribute slices. #258 follows #254 through #256; #259 follows the typed
attribute APIs and readers so that it can verify round trips. #260 follows all
implementation slices and removes every stale deferred disposition.

Shared attribute contract: public APIs expose Go-native attribute scopes and
boolean, numeric, and string types rather than C tables, records, unions,
vectors, or generated containers. Names, scalar values, slices, options, and
files are borrowed only for the synchronous call. Returned attribute metadata
and collections are non-nil Go-owned values, ordered by vertex or edge ID where
applicable, and remain valid after graph closure. Missing names, type mismatch,
embedded NUL bytes, empty values, vector-length mismatch, non-finite numeric
input, invalid IDs, and overwrite behavior are explicit.

The package installs and owns the upstream C attribute table exactly once
before attributed graphs are created. Callers cannot replace it or observe raw
attribute storage. Graph locks cover complete reads and mutations; writes are
serialized, racing `Close` either waits or returns `ErrClosed`, and imported,
copied, and derived graphs own their attribute storage independently. Attribute
combination uses Go-native policies for supported operations and keeps raw
`igraph_attribute_combination_t` values and custom C callbacks internal.
Operations that cannot preserve or combine a requested attribute without loss
must reject the request or document deterministic loss rather than silently
invent provenance.

Reader contract: `ReadEdgeList`, `ReadGraphML`, and `ReadGML` borrow an open,
seekable regular `*os.File` only for the synchronous call. They snapshot bytes
from its current offset without changing that offset or closing the file, and
return an independently owned graph. Edge-list options state minimum vertex
count and directedness explicitly. GraphML selects a zero-based top-level graph
and imports supported Boolean, numeric, and string attributes; unsupported
attribute types return an error. GML preserves its directed flag and node-ID
mapping while importing simple numeric and string attributes. Parse and index
errors return no partially owned graph.

Final reader disposition for #258: `igraph_read_graph_edgelist`,
`igraph_read_graph_graphml`, and `igraph_read_graph_gml` are user-facing;
`igraph_enter_safelocale` and `igraph_exit_safelocale` are internal reader
plumbing. NCOL, LGL, Pajek, DL, GraphDB, and DIMACS-flow readers are
intentionally unsupported: their symbolic identity/weight choices, lossy or
specialized format semantics, or auxiliary flow outputs require dedicated
policy/result APIs. No reader declaration reviewed by #258 remains deferred.

The graph-level slice from
[#255](https://github.com/h8gi/go-igraph/issues/255) lists Boolean, numeric, and
string metadata in stable name order and provides typed scalar getters,
setters, and removal operations. Missing names and wrong types have distinct
errors, same-type setters overwrite, empty strings are retained, and numeric
setters reject non-finite values before mutation. Names and strings are copied
by C-igraph during setters; getter results and metadata are Go-owned. Reads use
the graph read lock, writes are serialized, and a racing `Close` waits or makes
the operation return `ErrClosed`.

The vertex/edge slice from
[#256](https://github.com/h8gi/go-igraph/issues/256) provides the same typed
metadata, scalar, vector, and removal contracts in vertex-ID and edge-ID order.
Full-vector setters are the only operations that create an element attribute;
their length must exactly match the current vertex or edge count, with nil and
empty both denoting length zero. Scalar setters update an existing typed
attribute at a checked ID. After topology growth, C-igraph's missing-value
defaults are exposed explicitly as NaN for numeric, the empty string for
string, and false for Boolean attributes; callers can fill them with scalar or
aligned-vector setters. Input slices are borrowed synchronously and returned
slices are non-nil Go-owned copies, including on empty graphs.

The transformation slice from
[#257](https://github.com/h8gi/go-igraph/issues/257) adds Go-native, typed
attribute-combination policies while keeping upstream combination records,
callbacks, and sentinels internal. Lossless copies, deletion, subgraphs,
decomposition, directedness conversion, difference, and complement preserve
their documented graph, vertex, and edge values. Simplification, undirected
edge collapse, union, intersection, disjoint-union graph conflicts, and
composition require explicit scope-specific rules exactly where multiple
values reach one result element; same-name type conflicts require an explicit
drop rule. Mutations remain clone-and-swap atomic and
derived operator graphs reconstruct Go-visible values from exact provenance
under stable operand locks.

Shared interchange contract: public graph readers return independently owned
`Graph` values and destroy every partially initialized graph or attribute
record on parse, conversion, or adoption failure. Reader and writer file
arguments remain caller-owned, stay open after the synchronous call, and never
leak a `FILE *`. Format options replace upstream integer flags, indices, and
sentinels with checked Go values. Each format documents directedness, graph
indexing, supported attribute types and reserved names, identity/order
guarantees, and any deterministic loss. Locale changes, parser callbacks, and
flush handling remain internal and cannot weaken concurrent calls or error
propagation.

The core interchange slice covers edge lists for topology and GraphML/GML for
attributed round trips. NCOL, LGL, Pajek, DL, GraphDB, DOT, LEDA, and DIMACS
functions receive individual reviewed dispositions based on whether their
format-specific semantics support a coherent Go API; the milestone does not
expose them merely to increase declaration coverage.

Historical initial disposition: the 146 declarations discovered through
`igraph_attributes.h`, including its generated attribute-record list
operations, and the 18 currently missing declarations in `igraph_foreign.h`
are deferred to #254–#260. Five generated declarations already used by the
integer-vector-list infrastructure remain internal, and the two existing
foreign writer bindings remain user-facing. The final audit must resolve every
deferred declaration as user-facing, composed, internal, or intentionally
unsupported and leave no stale deferred or accidentally missing declaration in
either audited domain. The final audit is now complete: no declaration in the
attributes or graph-interchange domains remains deferred. Public typed APIs,
composed transformations, internal runtime/container/`FILE *`/locale plumbing,
and intentionally unsupported lossy or specialized formats each have a final
reviewed disposition in the generated coverage report. These domains are
marked completed in the inventory configuration, so a future deferred entry
fails validation.

Runtime-foundation disposition: #254 classifies attribute-table installation
and presence checks plus the record and generated-list initialization,
destruction, sizing, type-checking, and resize operations used by the shared
cleanup boundary as ten internal dependencies. The remaining 154 scoped
declarations stay deferred to the focused value, transformation, interchange,
and final-audit issues.

Completion criteria:

- both reference workflows are expressible without C types, hidden global
  mutation, caller-inferred ownership or ID alignment, or undocumented format
  loss;
- all public names, values, slices, options, and files have explicit borrowed
  or copied lifetimes, while graphs and returned metadata follow explicit
  independent ownership rules;
- initialization failure, upstream or parse error, early return, checked
  conversion, partial graph/record construction, and writer flush failure
  release every temporary C resource;
- missing, invalid, empty, non-finite, duplicate-name, wrong-type,
  length-mismatch, malformed-input, unsupported-format/type, directed,
  undirected, loop, parallel-edge, and use-after-`Close` behavior is
  documented and tested where relevant;
- copies, subgraphs, decomposition, simplification, direction conversion, and
  multi-graph operators preserve or combine attributes according to explicit
  contracts and independently owned results;
- race tests cover concurrent attribute reads, serialized mutations,
  concurrent imports and exports, stable multi-graph locking, one-time table
  installation, and graph closure; and
- every scoped declaration in `igraph_attributes.h` and
  `igraph_foreign.h` has a final reviewed disposition, examples cover both
  workflows, and `make verify` passes while preserving the statement
  coverage floor.

Completion evidence: `milestone17_integration_test.go` exercises attributed
GraphML import, typed inspection and mutation, union with scope-specific
combination rules, exact vertex/edge provenance, source closure, export, and
reimport. Focused tests cover malformed and unsupported inputs, empty graphs,
flush and initialization failures, use after `Close`, concurrent attribute
access, imports, exports, stable multi-graph locking, and close races.
`example_attributes_test.go` is output checked and
`examples/attributes/main.go` is a standalone attributed round-trip program.

## Milestone 18: Mixing, similarity, and local statistics

Goal: turn attributed graphs and aligned graph snapshots into coherent mixing,
neighborhood-similarity, citation-coupling, and weighted local-statistics
workflows without exposing C selectors, vectors, matrices, or ownership rules.

Status: planned. The reference workflows, shared contracts, initial deferred
inventory, and dependency-ordered issue plan are established in
[#269](https://github.com/h8gi/go-igraph/issues/269).

Reference workflows:

- attributed mixing analysis imports a GraphML graph, reads typed vertex
  categories or numeric values and optional edge weights, then computes scalar
  assortativity and category- or degree-aligned joint distributions; and
- local structural comparison computes neighborhood similarity for selected
  vertices, vertex pairs, or graph edges, then evaluates weighted local scans
  within one graph or across two snapshots sharing a vertex-ID space.

Planned areas:

- selector-aligned neighborhood similarity plus cocitation and bibliographic
  coupling matrices ([#270](https://github.com/h8gi/go-igraph/issues/270));
- allocation-conscious pair- and edge-selected Jaccard and Dice similarity
  ([#271](https://github.com/h8gi/go-igraph/issues/271));
- categorical, numeric, and degree assortativity
  ([#272](https://github.com/h8gi/go-igraph/issues/272));
- category- and degree-aligned joint distributions
  ([#273](https://github.com/h8gi/go-igraph/issues/273));
- weighted same-graph radius and caller-supplied subset scans
  ([#274](https://github.com/h8gi/go-igraph/issues/274));
- weighted cross-graph scans over aligned snapshots with stable multi-graph
  locking ([#275](https://github.com/h8gi/go-igraph/issues/275)); and
- final integration, examples, documentation, race/ownership coverage, and
  inventory audit ([#276](https://github.com/h8gi/go-igraph/issues/276)).

Execution order: #269 establishes the milestone-wide contract and initial
dispositions. #270 and #272 may then proceed independently. #271 follows #270
and reuses its metric, direction, loop, and alignment vocabulary. #273 follows
#272 and reuses its category, weighting, normalization, and undefined-result
contracts. #274 depends only on #269 and establishes radius and nested-subset
conversion; #275 follows #274 and adds aligned graph identity and stable
multi-graph locking. #276 follows every implementation slice and removes every
stale deferred disposition.

Shared ownership and concurrency contract: graphs, selectors, pair and subset
collections, category and numeric values, weights, and option values are
borrowed only for a synchronous call. Go slices are copied into temporary
C-owned storage when C must retain them during the call. Every returned scalar,
slice, matrix, and axis value is Go-owned, remains valid after graph closure,
and shares no mutable result storage with another call. A receiver lock covers
validation, C execution, and result copying. Operations on two graphs use the
existing stable-address lock order, acquire each distinct graph once, preserve
argument order inside the operation, and reject nil or closed operands with
`ErrClosed`.

Similarity alignment contract: Jaccard and Dice matrix rows follow the exact
materialized order of the row `VertexSelector`, and columns follow the exact
materialized order of the column selector. Explicit selector duplicates
produce duplicate rows or columns rather than being deduplicated. Square
inverse-log-weighted, cocitation, and bibliographic-coupling results use one
selector on both axes in its materialized order. Pair results follow pair input
order, including repetitions. Edge-selected results follow the materialized
edge-ID order, including repeated edge IDs, with endpoint orientation taken
from the graph. Empty selections return dimensionally aligned non-nil Go
results. Matrix dimensions and products are checked before C allocation.

Similarity direction and loop contract: public APIs reuse the package's
Go-native incoming, outgoing, and all-neighbor direction modes. Jaccard and
Dice expose loop inclusion explicitly; enabling it adds each compared vertex
to its own neighborhood as defined upstream. Inverse-log-weighted similarity
has no loop option because the upstream measure does not define one. Cocitation
on a directed graph compares incoming citation neighborhoods, while
bibliographic coupling compares outgoing reference neighborhoods; their public
documentation states this interpretation instead of accepting an inapplicable
direction option. Pair orientation does not change a symmetric similarity
value, though direction mode changes the neighborhoods being compared.

Mixing input and axis contract: categorical APIs use an immutable Go category
value constructed from either integer or string labels. Construction copies
the labels. Each source and target category set is compacted independently to
dense zero-based indices in stable first-occurrence order, preventing sparse
or hostile labels from controlling matrix allocation. Joint type results
return Go-owned row and column label axes alongside the matrix; row `i`
describes an edge source category and column `j` an edge target category when
direction is honored. An undirected result uses the same endpoint category
semantics symmetrically. Numeric assortativity uses one vertex-aligned value
slice for both endpoints by default and permits an explicit second
vertex-aligned target slice only when directed mixing is requested. Degree
distribution row and column indices denote exact source and target degrees for
the selected direction modes, starting at degree zero; explicit checked maxima
bound both axes and prevent implicit unbounded matrix allocation.

Mixing weighting and scalar contract: every supplied value slice has exactly
one entry per vertex and every supplied weight slice exactly one finite entry
per edge. Raw weighted sums may use finite signed weights, while probability
normalization rejects negative weights or a non-positive total. Pinned igraph
1.0.1 does not implement weighted nominal assortativity, so that API does not
silently accept and ignore weights. Directedness and normalization are explicit
Go options and are ignored only where graph structure makes them inapplicable,
as documented. Mathematically undefined scalar results, including empty or
constant-value cases, remain `NaN`; the binding does not invent a zero result.

Local scan contract: a non-negative integer radius selects one value per
vertex in vertex-ID order. Radius zero is degree for unit weights and strength
for supplied weights. Larger radii count edges, or sum edge weights, in the
direction-selected neighborhood induced for each vertex. Caller-supplied
subsets are borrowed nested Go slices and return one value per subset in input
order. Repeated whole subsets are valid and produce repeated result entries;
duplicate vertex IDs within one subset are rejected so accidental repetition
cannot multiply counts. Invalid IDs are rejected before C execution. Nested
sizes, integer conversions, vector-list initialization, and partial cleanup
are checked explicitly. `igraph_local_scan_neighborhood_ecount` is deprecated
in pinned igraph 1.0.1 and will be composed through the subset contract or
marked intentionally unsupported rather than exposed as a duplicate public
API.

Cross-graph scan contract: the first graph supplies neighborhoods and the
second supplies counted edges and optional edge-ID-aligned weights. Both graphs
must have the same vertex count, and vertex ID `i` must identify the same entity
in both snapshots; topology and edge IDs need not match. Results align with
that shared vertex-ID space. Equal operands are valid, reversed concurrent
calls cannot deadlock, and closing either graph races safely by waiting for the
operation or causing it to return `ErrClosed`. Comparison weights are aligned
only to the second graph and are borrowed for the call.

Validation and failure contract: invalid selector kinds or IDs, malformed
pairs, invalid direction modes, negative radii or bounds, value-length and
weight-length mismatches, non-finite inputs, incompatible graph snapshots,
checked-conversion overflow, and impossible matrix or nested-result dimensions
fail before unsafe allocation or C execution where possible. Initialization
failure, upstream error, early return, and partial nested-result construction
destroy every initialized C resource exactly once. Focused tests cover empty
and degenerate graphs, isolates, loops, parallel edges, duplicates, directed
and undirected modes, weighted and unweighted calls, use after `Close`, result
independence, concurrent reads, reversed operands, repeated operands, and close
races.

Initial disposition: all 23 declarations in pinned igraph 1.0.1 from
`igraph_cocitation.h`, `igraph_mixing.h`, and `igraph_scan.h` are deferred to
#270–#275. The final audit in #276 must resolve each declaration as
user-facing, composed, internal, or intentionally unsupported, with no stale
deferred or accidentally missing declaration in the three audited domains.

Completion criteria:

- both reference workflows are expressible without C types, caller-inferred
  ownership, ambiguous alignment, sparse-label allocation, or hidden graph
  identity assumptions;
- direction, loops, weighting, normalization, undefined values, category and
  degree axes, selector ordering, radius behavior, and snapshot compatibility
  are documented and tested;
- initialization failure, upstream error, early return, checked conversion,
  and partial matrix, vector, or nested-list construction release every
  temporary C resource;
- all returned values are independently Go-owned, and race tests cover
  concurrent reads, repeated and reversed graph operands, use after `Close`,
  and graph-close races;
- matrix/pair similarity, scalar/distribution mixing, and same-/cross-graph
  scan results satisfy independently computed known-answer invariants; and
- every scoped declaration has a final reviewed disposition, examples cover
  both workflows, statement coverage remains at least 90.0%, and `make verify`
  passes.

## Later domain milestones

Other specialized upstream domains remain candidates after the milestones
above. They should advance when a concrete use case can define a coherent Go
API and its resource model, not merely to increase the inventory percentage.

Select the next domain by user value, shared-infrastructure readiness, and the
ability to define a complete Go ownership and concurrency contract.

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
