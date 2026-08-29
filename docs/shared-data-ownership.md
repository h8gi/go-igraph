# Shared data ownership

The shared data layer is the boundary between ordinary Go values and the
temporary C/igraph values used to call upstream APIs. Public APIs never expose
C types, C-backed slices, or cleanup functions for internal values.

## Public lifetime rules

| Value | Storage | Caller action | Lifetime rule |
| --- | --- | --- | --- |
| `*Graph` | owns an `igraph_t` | call `Close` | `Close` is idempotent; methods that require a live graph return `ErrClosed` afterwards |
| graph-reader input and result | caller-owned `*os.File`; independently owned `*Graph` | keep the file open during the call; close the returned graph | edge-list, GraphML, and GML readers synchronously snapshot from the current offset without changing it or closing the file; the result retains no file or snapshot storage |
| graph-writer input | borrowed live `*Graph` and caller-owned `*os.File` | keep both open during the call; retain responsibility for closing both | edge-list, GraphML, and GML writers retain neither value, never close the file, propagate conversion/write/flush failures, and serialize locale-sensitive C I/O internally |
| `*Vector` | owns an `igraph_vector_t` | call `Close` | construction copies slice input; `Close` is idempotent; methods that require a live vector return `ErrClosed` afterwards |
| `Matrix` | Go-owned immutable value | none | constructors and `Rows` copy their input or result |
| `AttributeMetadata` | Go-owned name, scope, and type | none | names never alias C storage; metadata remains valid after graph closure |
| spatial point matrices and `NearestNeighborOptions` | borrowed Go values | none | point row `i` maps to vertex ID `i`; matrices and optional bound pointers are copied or read only during a synchronous call and are never retained |
| `ConvexHullResult` and spatial edge lengths | Go-owned values | none | hull indices align with coordinate rows; edge lengths align with edge IDs; all returned storage survives input or graph closure |
| nearest-neighbor graph result | independently owned `*Graph` | close `Graph` | point row `i` becomes vertex ID `i`; the graph retains no point matrix or option pointer and survives independently of all inputs |
| Delaunay, Gabriel, and relative-neighborhood graph results | independently owned `*Graph` | close `Graph` | point row `i` becomes vertex ID `i`; each undirected graph retains no point matrix and survives independently of its input |
| beta-skeleton graph results | independently owned `*Graph` | close `Graph` | lune and circle constructors borrow their point matrix and return an undirected graph whose vertex IDs preserve point-row identity |
| `BetaWeightedGabrielResult` | independently owned `*Graph` plus Go-owned thresholds | close `Graph` | `ThresholdBetas` aligns with graph edge IDs and survives graph closure; positive infinity is a valid persistence marker |
| `Path` | Go-owned result value | none | vertex and edge slices remain valid after the graph is closed; an unreachable target has `Found == false` and non-nil empty slices |
| `*SpectralEmbeddingResult` | Go-owned result value | none | `X`, `Y`, and `SingularValues` are Go-owned copies and remain valid after the graph is closed; `Y` is an empty matrix for undirected graphs and `SingularValues` is empty for edgeless graphs |
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
| isomorphism decision result | Go-owned Boolean | none | both graph operands are borrowed only for the synchronous call; the result remains valid after either graph is closed |
| `IsomorphismResult`, `SubgraphIsomorphismResult` | Go-owned mappings and Boolean | none | color inputs are borrowed and copied for the call; mapping slices are non-nil, explicitly directed, and remain valid after both graphs close |
| `MappingEnumerationResult` | Go-owned nested mapping slices and Boolean | none | enumeration requires a positive bound; each mapping is independently copied and remains valid after both graphs close |
| `CanonicalGraphResult` | independently owned `*Graph` plus a Go-owned permutation | close `Graph` | graph and source-to-canonical permutation survive source closure |
| automorphism generators and group size | Go-owned nested slices or `*big.Int` | none | generators and exact size survive graph closure and retain no Bliss or C storage |
| clique-family decisions and scalar extrema | Go-owned Boolean or `int` | none | selectors are borrowed for the synchronous call; results retain no graph or C storage |
| `VertexSetRange`, `VertexSetEnumerationOptions` | Go-owned option values | none | optional bound pointers are read only during the synchronous call; no pointer is retained |
| `VertexSetEnumeration` | Go-owned nested slices and Boolean | none | bounded results and every nested vertex set are non-nil independent copies that survive graph closure |
| cycle predicates, topological order, `Cycle`, and `GirthResult` | Go-owned scalars or slices | none | witnesses and orders are non-nil copies that remain valid and mutable after graph closure |
| `SimpleCyclesResult` and cycle-basis results | Go-owned nested slices and Boolean | none | bounded paired cycles and every basis element are non-nil independent copies that survive graph closure |
| feedback edge/vertex sets | Go-owned slices | none | weight inputs are borrowed and copied for the synchronous call; result IDs survive graph closure |
| dyad/triad census and triangle results | Go-owned structs, slices, or nested fixed arrays | none | selectors are borrowed and materialized for the synchronous call; exact counts and triangle IDs survive graph closure |
| RANDESU motif results | Go-owned slices or scalars | none | cut probabilities, seeds, and sample selectors are borrowed synchronously; histograms retain upstream NaN class markers and survive graph closure |
| graphlet basis and projection results | Go-owned nested clique slices and aligned real vectors | none | weights, clique bases, and initial coefficients are borrowed and copied synchronously; all returned storage survives graph closure |
| `DifferenceResult` | independently owned `*Graph` plus left-operand `GraphIDMapping` | close `Graph` | vertex mapping is exact; edge mapping follows the documented structural convention |
| `CompositionResult` | independently owned `*Graph`, Go-owned vertex mappings and edge provenance | close `Graph` | `Edges` is indexed by result edge ID and preserves one-to-many source participation |
| `CommunityPartition` | Go-owned result value and slices | none | membership, sizes, community count, and modularity score remain valid and mutable after graph closure |
| `SpinglassSingleResult` | Go-owned result value and slices | none | community member vertex IDs, cohesion, adhesion, inner links, and outer links remain valid after graph closure |
| `MaxFlowResult`, `MinCutResult`, `STMinCutResult` | Go-owned result value and slices | none | scalar values, flow vectors, cut edge sets, and partition vertex sets remain valid and mutable after graph closure |
| `STCut` | Go-owned result value and slices | none | cut edge set and source partition vertex set remain valid and mutable after graph closure |
| `ResidualGraphResult`, `GomoryHuTreeResult`, `DominatorTreeResult`, `TarjanReductionResult` | independently owned `*Graph` plus Go-owned result vectors | close `Graph` | returned tree and residual graph survive source graph closure; result vectors remain valid after graph closure |

`Graph` and `Vector` install finalizers as a leak fallback, but deterministic
code should still use `Close`, normally with `defer` or `t.Cleanup`.

## Attribute runtime and value boundary

The package installs C-igraph's C attribute table during package
initialization, before exported graph constructors can run. The table is
process-global in pinned C-igraph 1.0.1, so installation is guarded by
`sync.Once`, repeated setup is read-only, and callers cannot detach or replace
it through the Go API. If another C consumer has already installed a different
table, package initialization fails instead of creating graphs whose attribute
storage would be managed by inconsistent handlers. Pinned C-igraph 1.0.1 marks
the C handler experimental; no handler-specific storage or replacement
mechanism is therefore exposed as a stable Go contract.

Public attribute vocabulary is restricted to graph, vertex, and edge scopes
and Boolean, numeric, and string scalar types. C tables, unions, attribute
records, object values, generated record lists, and their destructors remain
internal. Attribute names must be non-empty valid UTF-8 without embedded NUL
bytes. String values may be empty but follow the same UTF-8 and NUL rules.
Metadata name/type collections must be length-aligned and contain no duplicate
name within one scope. Conversion eagerly copies names and returns a non-nil
Go-owned slice, including for an empty collection.

Graph-level attribute metadata is returned in lexical name order. Typed
getters distinguish `ErrAttributeNotFound` from `ErrAttributeTypeMismatch` and
copy string results into Go storage. Setters borrow names and values only for
the synchronous call; C-igraph copies them before return. Same-type setters
overwrite, cross-type setters fail without modification, empty string values
are valid, and numeric setters reject NaN and infinities. Removing a missing
name is an explicit not-found error, while removing all graph attributes is
idempotent and does not touch vertex or edge attributes.

Vertex and edge metadata follows the same lexical order. Full value getters
return non-nil Go-owned slices aligned with vertex IDs or edge IDs and remain
valid after graph closure. Full-vector setters borrow and copy inputs during
the call, require exact graph-size alignment, and are the only element setters
that create an attribute; nil and empty both mean a zero-length vector and are
valid only for an empty scope. Scalar setters require an existing attribute and
a valid current ID. Topology growth appends C-igraph's explicit missing-value
defaults: numeric NaN, empty string, and Boolean false. Numeric setter inputs
must still be finite, while getters preserve a NaN missing marker until callers
replace it. Removal operates on complete named vectors, and remove-all methods
affect only their selected vertex or edge scope.

Internal attribute records own their type-specific C vector and are destroyed
exactly once after successful initialization. A failed record or generated-list
initializer transfers no ownership to Go; partially initialized upstream
storage is cleaned by the upstream initializer's error path. Resize, type
checking, and list access propagate upstream error codes as Go errors. These
helpers establish the cleanup boundary used by the graph-, vertex-, edge-, and
interchange slices without exposing C lifecycle operations publicly.

Graph interchange duplicates the caller's file descriptor into a temporary
`FILE *`; all success and failure paths close only that duplicate. Readers use
a private byte snapshot so the caller's current offset is unchanged. Writers
flush the duplicate before returning so delayed I/O errors are observable.
Locale changes needed by C-igraph are serialized and restored internally.
GraphML is the supported attributed round trip for Boolean, numeric, and
string values. Imported and transformed graphs own separate attribute storage,
while returned metadata, value slices, and ID mappings are Go-owned and remain
valid after every participating graph is closed.

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

Graph-isomorphism decisions follow the same multi-graph borrowing and locking
contract. `Isomorphic` treats receiver and argument symmetrically.
`ContainsSubgraphIsomorphicTo` names the receiver as the target and the
argument as the pattern, matching the package-level `PatternToTarget` mapping
direction used by specialized matchers. General isomorphism accepts the graph
shapes supported by igraph's dispatcher, including multigraphs; general
subgraph isomorphism is restricted upstream to simple graphs. Directedness
mismatches and unsupported shapes are returned as errors rather than retained
as package state.

VF2 options pair colors by operand and color dimension. A nil pair disables
that dimension; providing only one side is rejected. Non-nil color slices,
including empty slices for empty graphs, are length-checked and copied into
temporary integer vectors. VF2 inputs are checked for loops and parallel edges
before matching because upstream requires simple graphs but does not perform
that validation itself. Equal-size mappings use `SourceToTarget` and
`TargetToSource`. Subgraph mappings use `PatternToTarget`; the reverse
`TargetToPattern` mapping contains `RemovedID` for target vertices not used by
the match. A non-match returns non-nil empty slices in every direction.

VF2 count operations return a checked Go `int`; conversion rejects values that
cannot be represented. Mapping enumeration never uses the upstream unbounded
collection functions. A C-only callback copies at most `MaxMappings` mappings
into an owned temporary vector list and requests normal early termination only
after observing one additional match. `Truncated` therefore means that more
mappings actually exist, including when the bound is reached exactly without
another match. Equal-size enumeration returns source-to-target mappings;
subgraph enumeration returns pattern-to-target mappings. The outer slice and
every nested slice are non-nil Go-owned values and share no C or Go backing
storage with sibling mappings.

LAD domain input is a borrowed outer slice indexed by pattern vertex. Each
inner slice is copied and must contain unique, in-range target vertex IDs; an
empty inner domain is an immediate, successful non-match without entering the
solver. Nil domains permit every target vertex. LAD rejects directedness
mismatches and parallel edges before the upstream call, while loops remain
supported. First-match results use the shared `PatternToTarget` and
`TargetToPattern` directions.

Bounded LAD enumeration does not request upstream's unbounded `maps` output.
It repeatedly requests one mapping and partitions the remaining domain space
into disjoint prefix constraints that exclude the returned mapping. The search
stops after observing one result beyond `MaxMappings`, so `Truncated` has the
same exact meaning as for bounded VF2 enumeration. Temporary domain vector
lists own copied vectors and are destroyed after every solver call; returned
mappings are non-nil independent Go slices.

Canonical labeling and automorphism operations borrow optional vertex colors,
length-check and copy them, and reject parallel edges before calling Bliss
because upstream otherwise returns unreliable results. Loops remain supported.
Canonical permutations are exposed as source-to-canonical mappings even though
the permutation consumed by `igraph_permute_vertices` is canonical-to-source;
the binding validates and inverts the upstream permutation explicitly. A
canonical graph owns a separate `igraph_t`, is caller-closed, and survives its
source graph.

Automorphism generators are copied from the temporary vector list into non-nil
Go slices. Each generator is a zero-based source-to-source permutation. Exact
group size uses the internal Bliss default heuristic, parses the decimal
result into a new `*big.Int`, and frees the Bliss-owned decimal string on every
successful path. No Bliss options or info structures cross the public API.

Clique membership borrows a `VertexSelector`, validates it against the current
graph, rejects duplicate explicit IDs, and copies explicit IDs into temporary
C-owned selector storage. Scalar clique and independence numbers normalize a
temporary graph copy before querying so loops and parallel edges consistently
have adjacency-only semantics; the source graph is never mutated. Direction is
ignored by scalar extrema and independent-set membership. `IsClique` exposes
the upstream directed choice explicitly.

Clique-family enumeration uses inclusive optional size bounds and a required
positive result limit. Bound pointers are read only while validating a call and
are never retained. Enumeration implementations observe one additional match,
so `Truncated` means another matching result actually existed. Returned outer
and inner slices are Go-owned and non-nil. `Cliques` requests exactly one result
beyond the public limit, sorts vertex IDs within each set, and does not promise
an outer enumeration order. `LargestCliques` composes the scalar clique number
with the same bounded enumerator instead of collecting through the upstream
unbounded largest-clique API. Histogram counts are checked for finiteness,
integrality, non-negativity, and Go `int` range before being returned.

Maximal-clique enumeration shares the same bounded result and temporary graph
normalization. `MaximalCliquesFromVertices` copies its explicit vertex input;
the input partitions upstream's internal initial search roots. It is neither an
induced-subgraph filter nor a requirement that returned cliques contain one of
the supplied IDs. Empty input returns an empty non-nil result, while duplicates
and invalid IDs are rejected before entering C. Count and histogram results are
checked before conversion to Go `int` values.

Weighted-clique inputs contain exactly one positive Go integer per vertex. The
slice is borrowed, validated for length and exact C `double` representation,
and copied into a temporary real vector. The positive total is also bounded to
the exact integer range of C `double`, preventing silent truncation by pinned
igraph. Optional weight bounds are positive inclusive Go integers and never
expose upstream zero/negative sentinel conventions. `MaximumWeightCliques`
computes the scalar optimum and bounded ties under one graph lock and one
normalized temporary graph, while all returned sets remain Go-owned.

Independent-vertex-set enumeration uses the same borrowed graph, inclusive
optional size range, positive result limit, exact truncation, canonical inner
vertex ordering, and Go-owned nested-slice contracts as clique enumeration.
Ordinary, maximal, and composed largest-set queries all ignore direction,
loops, and parallel edges through temporary graph normalization. Empty graphs
return a non-nil empty result, including where pinned igraph would otherwise
represent the empty set as one maximal result. No outer result order is
promised.

Cycle-analysis methods borrow the graph and option values only for the
synchronous call. `Cycle` keeps vertex and edge IDs aligned in traversal order;
bounded enumeration additionally keeps every paired cycle independently owned.
Topological orders, girth witnesses, cycle bases, and feedback sets are copied
before temporary C vectors or vector lists are destroyed. All collection-valued
results are non-nil, including acyclic, empty, and no-result cases.

`SimpleCycles`, `FundamentalCycleBasis`, and `MinimumCycleBasis` bind APIs marked
experimental in pinned igraph 1.0.1, which is stated on their public methods.
The callback-only simple-cycle declaration remains intentionally unsupported:
the bounded collector observes one additional result to provide exact
truncation without retaining a C callback. Cycle-basis options omit the pinned
unused weight parameters. Feedback weight slices are length-checked, restricted
to finite non-negative values, and copied into temporary C vectors; no solver
or upstream result storage escapes the call.

RANDESU motif methods borrow cut-probability slices and copy them into
temporary C vectors. Histogram results are copied into non-nil Go slices;
finite entries are checked exact non-negative counts, while NaN preserves
upstream markers for impossible isomorphism classes. Explicit sample selectors
are materialized and copied, reject duplicates, and are mutually exclusive
with random sample sizes. Random sampling and stochastic cutting execute under
the package RNG lock; a non-nil seed is applied only within that serialized
call. Estimates are finite non-negative real values and may be fractional;
total motif counts are checked exact Go-owned integers.

Graphlet methods borrow optional edge weights; nil means unit weights, while a
non-nil slice is length-checked, restricted to finite non-negative values, and
copied into a temporary C vector. Candidate and complete decomposition results
copy every clique and aligned threshold or coefficient into Go-owned storage.
Projection additionally copies and canonicalizes caller-supplied cliques,
rejects invalid IDs, repeated vertices, incomplete or duplicate cliques, and
uses a non-empty initial coefficient slice as the iterative starting point.
All temporary vector lists own their nested C vectors and are destroyed on
success, conversion failure, and upstream error paths.

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
separately closable `Graph`. Copies, deletion results, induced and edge
subgraphs, and decomposed components preserve graph attributes and the vertex
and edge attributes selected by their upstream ID permutation. Each result has
independent attribute storage: changing or closing a source, component, or
sibling does not affect another graph.

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

Public `AttributeCombinationPolicy` values select drop, first, last, numeric
sum/product/minimum/maximum/mean, or string concatenation without
exposing `igraph_attribute_combination_t`, callbacks, variadic sentinels, or C
storage. Policies and override maps are borrowed synchronously. Unknown names,
invalid rules, and rules incompatible with an attribute type are rejected
before mutation. Internal combination lists are destroyed after every success
or partial construction failure.

Simplification preserves graph and vertex attributes. Loop-only simplification
uses first-value semantics to preserve every surviving edge attribute. If
parallel removal may merge attributed edges, `SimplifyOptions.EdgeAttributes`
is required, even when the desired rule is to drop them. Directed conversion
preserves graph and vertex attributes and copies each source edge attribute to
its derived edge. Per-edge undirected conversion preserves edge attributes;
collapse and mutual conversion require an explicit edge policy when attributes
exist. All three mutations remain clone-transform-swap atomic after policy
validation and combination-list allocation.

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

Operator results reconstruct attributes from these exact mappings while the
operand locks remain held. Disjoint union concatenates same-typed vertex and
edge values in operand order; same-name graph attributes and element type
conflicts require a policy. Union and
intersection preserve a single contributing value and require the matching
`GraphOperatorAttributePolicy` scope when two same-name values reach one result
element. Composition applies the same rule to graph and vertex mappings and to
each exact contributing edge pair. A nil policy is therefore valid only when
no combination is needed. Difference retains the left graph, vertex, and
surviving edge attributes. Complement retains graph and vertex attributes and
has no source edge attributes to copy. Type conflicts require an explicit drop
rule or return `ErrAttributeTypeMismatch`. Any validation, allocation, or
upstream failure closes a partial result and leaves both operands unchanged;
successful results own independent attribute storage.

`DisjointUnionMany`, `UnionMany`, and `IntersectionMany` accept ordinary
borrowed Go slices and return `ManyGraphOperatorResult`. `Inputs` is aligned
with operand order, including repeated graph pointers, and contains exact
non-nil Go-owned vertex and edge mappings. Distinct operands are locked once in
stable order. The temporary C graph-pointer vector and nested edge-map list are
internal, destroyed before return, and never become public lifecycle objects.
An empty operand slice returns an independently owned directed null graph and a
non-nil empty mapping slice, matching the documented upstream convention.
Disjoint-union mappings preserve each input's order with cumulative offsets;
union and intersection mappings preserve upstream's exact source-to-result
edge correspondence. Attribute restoration uses all exact mappings and the
same scope-specific conflict policy as binary operators.

`GraphPower` borrows its source under the graph lock and returns an
independently owned simple graph with an exact identity vertex mapping. Order
zero retains vertices and removes all edges; positive orders connect vertices
reachable within that many steps. Respecting direction preserves directed
reachability, while ignoring it produces an undirected result. Graph and vertex
attributes are independently preserved and edge attributes are intentionally
dropped because power edges have no unique source edge.

`ConnectNeighborhoodInPlace` adds the same bounded-reachability connections to
a clone and swaps only after upstream execution and mapping conversion succeed.
Its edge mapping covers every original edge and marks added edges with
`RemovedID` in `NewToOld`. Existing edge attributes survive; added edges receive
typed missing defaults. DirectionOut follows outgoing paths, DirectionIn
follows incoming paths while retaining their path orientation, and DirectionAll
ignores direction for reachability but adds one pinned-upstream-oriented edge
per missing unordered connection. Order zero is a validated identity operation.

`Join` borrows two same-directedness operands and returns an independently
owned graph. Left vertex IDs are preserved and right vertex IDs are offset by
the left vertex count. Exact structural edge mappings pair endpoint-equivalent
parallel edges by ascending IDs; newly created cross-operand edges have no
source. Graph, vertex, and original edge attributes are reconstructed through
`GraphOperatorAttributePolicy`, while new edges receive typed missing values.

`Product` and `RootedProduct` wrap experimental upstream functions. Their Go
ownership and validation contracts are stable for pinned igraph 1.0.1, but the
structural semantics must be re-audited on dependency upgrades. Both borrow
same-directedness operands and return an independently owned graph. Product
mode and rooted-product root are validated before the product call; the root is
a vertex of the second operand. Modular products additionally reject operands
with loops or parallel edges. Result vertex `(u,v)` has ID `u*|V2|+v`.
`GraphProductResult` exposes both result-to-pair provenance and non-nil
source-to-result lists as Go-owned values that outlive all graphs. Product
attributes are intentionally not propagated because product vertices and
edges do not have a unique source element.

`ContractVerticesInPlace` borrows a Go target-label slice and optional vertex
attribute policy while holding the receiver lock. Labels must be non-negative;
gaps are removed by ranking distinct labels in ascending order. The returned
vertex mapping records this exact many-to-one normalization and the edge
mapping is identity because contraction preserves every edge ID and its
attributes, even when endpoints become loops or parallel edges. A policy is
required when an actual merge combines vertex attributes. Graph and edge
attributes remain unchanged. Identity and empty mappings are validated no-ops.

`ReverseEdgesInPlace` materializes a borrowed `EdgeSelector` before mutation,
deduplicates selected IDs, and reverses each selected directed edge exactly
once. Empty selections are no-ops and undirected graphs are rejected. Vertex
and edge IDs, ordering, multiplicity, loops, and all attributes are preserved,
so both returned mappings are exact identities. Both mutation APIs allocate
and transform a clone and swap it into the receiver only after upstream work
and provenance checks succeed; all earlier failures destroy temporary C
vectors, selectors, combination records, and graph copies without mutation.

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

Cutoff-distance matrices, eccentricity and local-efficiency slices, graph
centers, pseudo-diameter endpoint records, and path-length histograms follow
the same synchronous borrowing rule and are returned as Go-owned values.
Distance-derived weighted operations require finite non-negative edge lengths.
Pseudo-diameter's optional start and seed values are borrowed only for the
call; a nil start delegates the starting-vertex choice to the package-locked
C/igraph RNG.

Reachability sets and counts are copied out of upstream component and bitset
storage before return. Neighborhood graphs and transitive closures are
independently owned and must each be closed by the caller. Every neighborhood
result includes a Go-owned source-vertex mapping that remains valid after the
source graph or any sibling graph is closed.

Widest-path sequences and matrices and Voronoi memberships/distances are
Go-owned. Spanner graphs are independently owned, must be closed by the caller,
and carry a Go-owned result-edge-to-source-edge provenance slice.

Eulerian path and cycle traversals reuse the Go-owned `Path` contract. Their
vertex and edge slices remain valid after the source graph is closed.

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
| integer vector list | owns its container and every copied integer vector | destroy the whole list after eager nested-slice conversion; partial conversion returns no result and still destroys every C vector |
| string vector | each validated Go string is copied by igraph; temporary `CString` values are freed immediately | destroy the vector on a partial set failure and after successful use |
| matrix | Go matrix data is copied into C and copied back into a new `Matrix` | destroy after the upstream call or result copy |
| explicit selector | backing vector is copied into a regular selector | destroy backing vector immediately; destroy the regular selector after its iterator |
| immediate selector | contains no allocation | never call the regular-selector destroy function |
| vertex or edge iterator | borrows its graph and selector only during eager materialization | destroy iterator first, then its owned selector, before releasing the graph lock |
| initialized graph result | ownership is moved, never copied from borrowed list storage, into exactly one public `Graph` | clear the moved-from value; the public `Graph.Close` performs the one destroy |
| graph list | owns its container and every graph still stored in it | remove transfers one element; destroy the container on success and all remaining elements plus already adopted graphs on any failure or early return |
| mutating replacement graph | source graph and inputs are borrowed under one lock; mapping vectors, when used, and the clone are temporary owners | destroy the clone and temporaries on every failure; on success destroy the prior graph exactly once after moving in the clone |
| Bliss exact-size string | Bliss allocates a decimal C string | copy into a Go `big.Int`, then release with `igraph_free`; no Bliss info object or string escapes the C wrapper |

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

Bipartite partitions are explicit Go-owned boolean slices aligned with vertex
IDs; they are not stored as hidden graph attributes. Constructor inputs are
borrowed only for the synchronous call and copied into temporary C vectors.
`Bipartite` copies a discovered partition into Go storage before releasing the
graph read lock. `NewBipartite` and `NewFullBipartite` return independently
owned graphs that the caller closes plus non-nil partition copies that remain
valid after graph closure. Failed construction adopts no graph, and temporary
partition and edge vectors are destroyed on every return path.

Biadjacency construction borrows immutable `Matrix` values and copies them into
temporary C matrices. Weighted construction copies its output weights into a
Go-owned slice aligned with returned graph edge IDs. Graph-to-matrix conversion
borrows the explicit partition and optional edge weights while holding the
graph read lock, then eagerly copies the matrix and row/column source vertex IDs
into Go-owned values. Partial initialization or conversion destroys every
temporary matrix and vector and, after successful graph initialization, either
adopts that graph exactly once or destroys it before returning an error.

Bipartite projection borrows the source graph and explicit partition while
holding the graph read lock. Each requested projection is initialized into a
separate `igraph_t`, adopted into an independently owned public `Graph`, and
paired with Go-owned source vertex IDs and edge-ID-aligned multiplicities.
Two-mode projection cleans both initialized graphs if either multiplicity
conversion fails. Closing the source or either sibling does not affect the
other graph or any returned slice.

Bipartite matching borrows the graph, explicit partition, candidate pairs,
optional edge weights, and epsilon only for the synchronous call. Candidate
pairs are converted into a temporary symmetric C mate vector whose `-1`
unmatched values never enter the public API. Maximum-matching mate vectors are
validated and copied into sorted Go-owned `MatchedPair` slices before temporary
C vectors are destroyed. No solver state or unmatched sentinel escapes, and
the result remains valid after graph closure.

Random bipartite constructors borrow `BipartiteRandomOptions` and its optional
seed pointer only for the synchronous call. Stochastic execution is serialized
under the package RNG lock, so equal seeds and options replay exactly without
interference from concurrent seeded calls. Each result owns an independent
graph and a non-nil Go-owned partition that remains valid after graph closure.
The GNP and GNM constructors produce simple graphs; IEA independently assigns
edge draws and may produce parallel edges. Bipartite self-loops are never
generated.

Spatial constructors copy immutable point matrices into temporary C storage;
point row `i` remains vertex ID `i`, and no constructor retains the matrix.
Every returned graph is independently owned. Hull coordinates, hull indices,
edge lengths, and weighted-Gabriel thresholds are copied into Go-owned storage
before temporary C resources are destroyed. Edge-value slices align with the
returned or receiver graph's edge IDs even though edge enumeration order itself
is unspecified. `SpatialEdgeLengths` holds the graph read lock for its complete
synchronous call, so concurrent reads are safe and a racing `Close` either
waits or causes the operation to return `ErrClosed`; spatial constructors use no
package-global mutable state.

## Verification

Run the same complete verification entry point used by CI:

```sh
make verify
```

It checks formatting, runs `go vet`, tests and the statement-coverage floor
against the pinned igraph release in Docker, tests the inventory tool, and
checks that the generated API coverage report is current.
