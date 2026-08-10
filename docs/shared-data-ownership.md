# Shared data ownership

The shared data layer is the boundary between ordinary Go values and the
temporary C/igraph values used to call upstream APIs. Public APIs never expose
C types, C-backed slices, or cleanup functions for internal values.

## Public lifetime rules

| Value | Storage | Caller action | Lifetime rule |
| --- | --- | --- | --- |
| `*Graph` | owns an `igraph_t` | call `Close` | `Close` is idempotent; methods that require a live graph return `ErrClosed` afterwards |
| `*Vector` | owns an `igraph_vector_t` | call `Close` | construction copies slice input; `Close` is idempotent; methods that require a live vector return `ErrClosed` afterwards |
| `Matrix` | Go-owned immutable value | none | constructors and `Rows` copy their input or result |
| `Path` | Go-owned result value | none | vertex and edge slices remain valid after the graph is closed; an unreachable target has `Found == false` and non-nil empty slices |
| `DiameterResult` | Go-owned result value | none | scalar fields and its `Path` remain valid after the graph is closed; when no diameter path exists, endpoint IDs are -1 and path slices are non-nil empty slices |
| `AveragePathLengthResult` | Go-owned result value | none | contains scalar length and unreachable ordered-pair count; retains no graph or C resource |
| transitivity and density results | Go-owned scalar or slice | none | scalar results retain no resource; local-transitivity slices remain valid and mutable after graph closure |
| `VertexSelector` | Go-owned value | none | constructors copy explicit IDs; no graph or C resource is retained |
| `EdgeSelector` | Go-owned value | none | constructors copy explicit IDs or pairs; no graph or C resource is retained |
| selection result | Go-owned slice | none | remains valid and mutable after the graph is closed |
| `ConnectedComponents` | Go-owned value and slices | none | membership and sizes remain valid and mutable after the graph is closed |
| articulation-point and bridge results | Go-owned slices | none | zero-based IDs remain valid and mutable after the graph is closed |
| `BiconnectedComponents` | Go-owned value and nested slices | none | component edge/vertex IDs and articulation points remain valid and mutable after the graph is closed |
| `BFSResult`, `DFSResult` | Go-owned slices | none | traversal options are borrowed only during the call; results remain valid and mutable after the graph is closed |
| centrality score/result values | Go-owned slices and scalars | none | selector, cutoff, reset, weight, and solver inputs are borrowed only for the call; returned scores and metadata remain valid after graph closure |
| `CentralizationResult` | Go-owned slice and scalars | none | generic score input is copied; specialized results retain no graph, solver, or C resource |
| `IDMapping` | Go-owned slices | none | `OldToNew` is indexed by source ID; `NewToOld` is indexed by derived ID; both survive graph closure |
| `GraphIDMapping` | two Go-owned `IDMapping` values | none | vertex and edge mappings follow the same direction, sentinel, and lifetime rules |
| `GraphTransformationResult` | Go-owned mappings and availability flag | none | vertex mapping is identity; an unavailable one-to-many edge mapping has non-nil empty slices and must not be read as empty provenance |
| `VertexMappedGraphResult` (`InducedSubgraphResult`, `EdgeSubgraphResult`) | independently owned `*Graph` plus Go-owned vertex mapping | close `Graph` | the aliases share one ownership shape; operation docs define exact mapping provenance |
| decomposition result | non-nil slice of independently owned `*Graph` values | close every graph | source and sibling closure do not invalidate any returned component |
| `BinaryGraphOperatorResult` | independently owned `*Graph` plus two Go-owned `GraphIDMapping` values | close `Graph` | operand-to-result mappings and graph survive closure of either operand |
| `DifferenceResult` | independently owned `*Graph` plus left-operand `GraphIDMapping` | close `Graph` | vertex mapping is exact; edge mapping follows the documented structural convention |
| `CompositionResult` | independently owned `*Graph`, Go-owned vertex mappings and edge provenance | close `Graph` | `Edges` is indexed by result edge ID and preserves one-to-many source participation |
| `CommunityPartition` | Go-owned result value and slices | none | membership, sizes, community count, and modularity score remain valid and mutable after graph closure |
| `SpinglassSingleResult` | Go-owned result value and slices | none | community member vertex IDs, cohesion, adhesion, inner links, and outer links remain valid after graph closure |
| `MaxFlowResult`, `MinCutResult`, `STMinCutResult` | Go-owned result value and slices | none | scalar values, flow vectors, cut edge sets, and partition vertex sets remain valid and mutable after graph closure |
| `STCut` | Go-owned result value and slices | none | cut edge set and source partition vertex set remain valid and mutable after graph closure |
| `ResidualGraphResult`, `GomoryHuTreeResult`, `DominatorTreeResult`, `TarjanReductionResult` | independently owned `*Graph` plus Go-owned result vectors | close `Graph` | returned tree and residual graph survive source graph closure; result vectors remain valid after graph closure |

`Graph` and `Vector` install finalizers as a leak fallback, but deterministic
code should still use `Close`, normally with `defer` or `t.Cleanup`.

Selectors are reusable. Their graph-independent shape is validated when they
are constructed, and bounds, missing endpoint pairs, and closed graphs are
validated each time they are materialized. `SelectedVertexIDs` and
`SelectedEdgeIDs` eagerly materialize a result while holding the graph lock;
they do not return an iterator that borrows the graph.

Derived-graph operations borrow every source graph only for the synchronous
call. They materialize selectors while the corresponding source lock is held,
before any source mutation can begin. Operations with multiple graph inputs
deduplicate repeated graph pointers, acquire all distinct locks in stable
address order, verify closure only after all locks are held, and unlock in
reverse order. Supplying the same graph more than once therefore cannot
self-deadlock, while a nil or closed input returns `ErrClosed`.

Every successfully returned derived `Graph` owns exactly one independently
initialized `igraph_t`. It remains usable after all source graphs are closed;
closing it repeatedly or in any order relative to sibling results is safe.
No returned graph shares a graph-list container or requires manual C cleanup.

`IDMapping.OldToNew` is indexed by source ID and contains the derived ID or
`RemovedID` (`-1`). `NewToOld` is indexed by derived ID and contains the lowest
source ID that maps to it, or `RemovedID` for an element with no source. The
lowest-ID rule makes many-to-one mappings deterministic. Identity mappings
contain the index in both directions. For a many-to-one transformation,
`NewToOld` is a deterministic representative mapping rather than a strict
inverse of `OldToNew`. Empty mappings contain non-nil empty slices. Both
directions are copied into Go storage and remain valid and mutable after source
and result graphs are closed.

`DeleteVertices` and `DeleteEdges` borrow their selector for the call and
materialize it into Go-owned IDs while holding the graph lock, before creating
or mutating any replacement graph. Duplicate IDs coalesce and selector order
does not affect deletion. Empty selectors are no-ops that return identity
vertex and edge mappings. Vertex deletion marks both removed vertices and all
incident edges with `RemovedID`; edge deletion returns an identity vertex
mapping and marks selected edges with `RemovedID`. Retained elements preserve
their relative ID order, including directed edges, loops, and parallel edges.

Deletion is fully atomic. The binding clones the graph, initializes selectors
and mapping outputs, invokes upstream igraph, converts and validates all
Go-owned mappings, and only then moves the completed clone into the public
`Graph`. Validation, initialization, upstream, or post-mutation conversion
failure destroys the clone and leaves the original graph unchanged. Successful
mappings retain no graph or C storage and survive closing the mutated graph.

Induced-subgraph selectors are materialized while the source graph is locked.
Duplicate vertex IDs are considered once, selector order is ignored, and the
result follows increasing source vertex ID order. Its vertex mapping is the
exact mapping returned by upstream igraph. The binding does not infer an edge
mapping because upstream does not return result edge-ID correspondence or a
result edge-order mapping.

Edge-subgraph selectors are likewise materialized before the upstream call.
Duplicate edge IDs are retained once and IDs are supplied to upstream in
increasing source-ID order. When isolated vertices are retained, vertex IDs do
not change and the vertex mapping is identity. When isolated vertices are
deleted, the binding first creates an independent edge subgraph containing all
source vertices, so its initial vertex IDs remain identical to source IDs. It
then deletes that result's isolated vertices and returns the exact mapping from
upstream `igraph_delete_vertices_map` in Go-owned storage. No edge mapping is
exposed because upstream does not return result edge-ID correspondence or a
result edge-order mapping. Component decomposition also returns no inferred
provenance: component ordering and vertex renumbering are upstream-defined.
Every component is removed from its temporary graph list and adopted into a
separately closable `Graph`.

`SimplifyInPlace`, `ConvertToDirectedInPlace`, and
`ConvertToUndirectedInPlace` are explicit mutating transformations. Each one
clones the receiver under its lock, completes the upstream transformation on
the clone, and swaps it into the receiver only after success. Initialization or
upstream failure destroys the clone and leaves the receiver unchanged. These
APIs construct structural source-to-result edge mappings from the source and
completed replacement edge lists. Simplification and undirected collapse use
the `IDMapping` many-to-one representative convention; undirected mutual maps
both members of each reciprocal pair to one result and marks unmatched edges
`RemovedID`. Per-edge undirected and non-mutual directed modes are one-to-one.
Mutual directed conversion is one-to-many and cannot be represented by
`IDMapping`, so `EdgeMappingAvailable` is false and both edge mapping slices
are non-nil and empty. This explicit state must not be interpreted as an empty
source graph. An identity or empty-to-empty operation has an available mapping.
For endpoint-equivalent parallel edges, one-to-one mappings pair edges in
ascending source and result edge-ID order. Undirected mutual conversion pairs
ascending source IDs from each direction and assigns ascending result IDs for
that endpoint group. These are deterministic structural conventions, not
attribute provenance.

The package does not expose graph attributes or raw
`igraph_attribute_combination_t` policies. Simplification passes no edge
attribute combination and therefore discards edge attributes when it runs.
Undirected collapse and mutual conversion likewise pass no combination and
discard edge attributes; per-edge undirected conversion preserves them under
upstream semantics. Directed conversion uses upstream attribute behavior with
no configurable policy.
Binary graph operators borrow operands while `withLockedGraphs` holds every
distinct graph lock in stable pointer order. Reversed operand order and repeated
operands therefore do not deadlock. Directedness must match for every binary
operator. Union and intersection use the larger operand vertex count; union
takes the larger and intersection the smaller parallel-edge multiplicity.
Difference preserves the left operand's vertices and subtracts right-edge
multiplicity. Disjoint union preserves operand order and offsets all right IDs.

Disjoint union mappings follow upstream's documented ordering. Union edge
mappings and intersection inverse edge mappings are copied from upstream and
converted to exact operand-to-result `GraphIDMapping` values. Composition keeps
upstream's per-result-edge pair of contributing operand edge IDs because a
single source edge may contribute to several result edges and cannot be
represented by `IDMapping`. Difference exposes its exact left identity vertex
mapping. Upstream provides no difference edge map, so simple edges are matched
exactly and endpoint-equivalent parallel edges are paired by ascending source
and result edge IDs, with excess left edges marked `RemovedID`. This is a
deterministic structural convention, not attribute provenance.
Complement exposes its exact identity vertex mapping; complement edges have no
source correspondence. Complement supports input loops but is intentionally
limited to simple-graph inputs: parallel edges are rejected because upstream's
complement traversal does not provide reliable multigraph presence semantics.
No unavailable edge provenance is guessed. All mapping and provenance slices
are non-nil Go-owned values and remain valid after every graph is closed.

The current public combination API is deliberately binary. The upstream
`disjoint_union_many`, `union_many`, and `intersection_many` entry points remain
unbound: their empty-list directedness convention and nested per-operand
provenance results do not share the binary result contract without adding a
separate graph-pointer-list and vector-list ownership surface.

Connected-component results are eagerly copied while holding the graph lock.
`Membership` is indexed by vertex ID and contains component IDs; `Sizes` is
indexed by component ID; and `Count` equals `len(Sizes)`. Component numbering
and ordering come from upstream igraph. No result retains the source graph or
any C storage.

Path and distance options borrow a non-nil weight slice only for the duration
of the call. The slice must contain one finite value per edge and is copied to
temporary C storage; `nil` selects the unweighted operation. Distance matrices
follow the materialized source and target selector order, including duplicates,
and represent unreachable pairs as positive infinity. Path slices and matrices
remain valid after their source graph is closed.

A non-nil empty weight slice therefore represents a weighted call only for a
graph with no edges; it is not interchangeable with `nil` on a graph that has
edges. Path algorithms support finite negative weights and report reachable
negative cycles as upstream errors. Whole-graph diameter and average-path
summaries reject negative weights upstream, while density accepts them.

Density and whole-graph distance-summary options use the same optional weight
contract. Diameter and average path length delegate negative-weight rejection
to their upstream shortest-path algorithm.
`IgnoreUnreachable` restricts summaries to reachable pairs; otherwise a
disconnected summary length is positive infinity. Diameter vertex and edge
paths, local transitivity slices, and all other structural metric results are
copied into Go-owned values before temporary C resources are destroyed. Local
transitivity follows materialized selector order, including duplicates; the
binding only deduplicates internally where required by the upstream call and
expands the result before returning it.

Centrality options use method-specific edge-weight validation. Closeness,
harmonic centrality, and betweenness require one finite, strictly positive
weight per edge. Eigenvector centrality, HITS, and PageRank allow finite
non-negative weights. Every non-nil weight slice is borrowed only for the call
and copied into a temporary C vector. A nil centrality cutoff means unlimited;
a non-nil cutoff is borrowed and must be finite and non-negative.

PageRank reset distributions are borrowed and copied into temporary C storage;
reset selectors are borrowed and materialized while the graph lock is held.
The two reset forms are mutually exclusive. Returned PageRank scores follow the
materialized result-selector order, including duplicates. Solver option values
are copied into stack-local upstream options for one call; no public value owns
or exposes an upstream solver object.

Graph centralization inputs and outputs are ordinary Go values. Generic
centralization copies its score input into the returned result. Specialized
centralization routines return Go-owned node scores and scalars. Raw empty-graph
centralization uses non-nil empty scores, `NaN` value, and theoretical maximum
zero; normalized empty or single-vertex calculations return an error rather
than dividing by zero.

## Nil and empty values

Nil and empty slices have the same meaning at the shared data boundary unless
an API explicitly requires at least one value:

- numeric, boolean, and string conversion helpers create valid zero-length C
  vectors for either form;
- `NewVectorFromSlice(nil)` creates a zero-length vector;
- `NewMatrixFromRows(nil)` and `NewMatrixFromRows([][]float64{})` create a
  0-by-0 matrix;
- `VertexIDs()`, `EdgeIDs()`, and `EdgePairs(nil, directed)` select nothing;
- `NoVertices` and `NoEdges` select nothing, including on an empty graph;
- `NewGraphFromEdges(vertexCount, nil)` creates the requested isolated
  vertices, and `AddEdges(nil)` is a no-op;
- empty deletion selectors are no-ops with non-nil identity mappings; deleting
  all edges retains all vertices, while deleting all vertices yields a valid
  zero-vertex, zero-edge graph and non-nil empty reverse mappings;
- zero-option simplification and conversion to the graph's existing direction
  are no-ops; empty and edgeless graphs remain valid while directedness changes
  according to a requested conversion;
- `ConnectedComponents` on a graph with no vertices returns non-nil, empty
  membership and size slices with a component count of zero;
- cut-structure queries on empty or edgeless graphs return non-nil empty
  slices, including both outer component collections;
- an empty induced-subgraph selector returns an independently owned empty graph
  with a non-nil mapping; an empty edge selector either retains all isolated
  vertices or returns a null graph according to its deletion option;
- decomposition of a graph with no vertices returns a non-nil empty graph
  slice;
- breadth-first search requires at least one root and depth-first search
  requires one valid root, so traversal of an empty graph returns an error;
- `NewLattice` is the exception: it rejects nil and empty dimensions because
  a lattice needs at least one dimension;
- empty centrality selectors return non-nil empty score slices; an empty graph
  returns non-nil empty eigenvector, HITS, and PageRank score slices;
- personalized PageRank reset distributions and reset selectors are exceptions
  that must contain positive mass or select at least one vertex.

Every successful API that returns a collection returns a non-nil empty Go
slice when the result has no elements. A zero selector value is intentionally
different from an empty explicit selector: `VertexSelector{}` selects all
vertices and `EdgeSelector{}` selects all edges.

## Internal C ownership

Temporary wrappers follow one rule: a successful initializer transfers one C
resource to the wrapper, and the same stack frame arranges its destruction.
Initialization failure does not create a resource that needs destruction.

| Wrapper | Input and result boundary | Cleanup order |
| --- | --- | --- |
| integer, real, boolean vectors | Go input is copied into C; results are copied back into non-nil Go slices | destroy after the upstream call or result copy |
| string vector | each validated Go string is copied by igraph; temporary `CString` values are freed immediately | destroy the vector on a partial set failure and after successful use |
| matrix | Go matrix data is copied into C and copied back into a new `Matrix` | destroy after the upstream call or result copy |
| explicit selector | backing vector is copied into a regular selector | destroy backing vector immediately; destroy the regular selector after its iterator |
| immediate selector | contains no allocation | never call the regular-selector destroy function |
| vertex or edge iterator | borrows its graph and selector only during eager materialization | destroy iterator first, then its owned selector, before releasing the graph lock |
| initialized graph result | ownership is moved, never copied from borrowed list storage, into exactly one public `Graph` | clear the moved-from value; the public `Graph.Close` performs the one destroy |
| graph list | owns its container and every graph still stored in it | remove transfers one element; destroy the container on success and all remaining elements plus already adopted graphs on any failure or early return |
| mutating replacement graph | source graph and inputs are borrowed under one lock; mapping vectors, when used, and the clone are temporary owners | destroy the clone and temporaries on every failure; on success destroy the prior graph exactly once after moving in the clone |

No C object retains a pointer into Go memory. All C/igraph error codes are
converted to Go errors, and each constructor cleans up any successfully
initialized dependency before returning an error.

All fallible Milestone 3 and 4 calls, including temporary integer/real vector,
matrix, vector-list, selector, and iterator initialization, pass through
central C wrappers. A wrapper installs igraph's non-aborting error and warning
handlers, performs exactly one upstream operation, and restores the prior
handlers before the cgo call returns. The pinned thread-safe igraph build keeps
handler state thread-local, so installation and restoration occur on the same
OS thread without a process-wide Go mutex. Initialization failure transfers no
resource; partial Go constructors destroy every dependency whose initializer
already succeeded. Centrality output vectors, selectors, reset vectors, weight
vectors, and stack-local solver options remain scoped to the graph lock and are
destroyed on upstream error, warning paths, early return, and success.
Graph-list initialization, insertion, and removal use the same non-aborting C
wrapper contract. Initialization failure transfers no list ownership. During
extraction, removed elements are either immediately adopted or destroyed;
conversion failure closes all previously adopted graphs and list destruction
cleans every element that has not yet been removed.

Flow, cut, and connectivity operations borrow optional capacity and flow
vectors only for the call. Non-nil capacity slices must match `g.NumEdges()` in
length and contain finite non-negative values (`>= 0`); `nil` specifies unit edge
capacity `1.0`. Vector outputs (flows, cut edge lists, partition vertex lists)
are copied eagerly into Go-owned slices while the graph lock is held, and temporary
C storage is destroyed before returning. Cut enumeration (`AllSTCuts` and `AllSTMincuts`)
extracts list of vector results into Go `[]STCut` values; if nested allocation
or conversion fails partway, all initialized C vectors and lists are cleanly freed.
Graph-returning flow APIs (`ResidualGraph`, `ReverseResidualGraph`, `GomoryHuTree`,
`DominatorTree`, `EvenTarjanReduction`) instantiate an independent `igraph_t` inside a
new Go `*Graph`. The returned graph survives closing the source graph and can be
closed independently.

## Verification

Run the same complete verification entry point used by CI:

```sh
make verify
```

It checks formatting, runs `go vet`, tests and the statement-coverage floor
against the pinned igraph release in Docker, tests the inventory tool, and
checks that the generated API coverage report is current.
