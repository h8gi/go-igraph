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
| `VertexSelector` | Go-owned value | none | constructors copy explicit IDs; no graph or C resource is retained |
| `EdgeSelector` | Go-owned value | none | constructors copy explicit IDs or pairs; no graph or C resource is retained |
| selection result | Go-owned slice | none | remains valid and mutable after the graph is closed |
| `ConnectedComponents` | Go-owned value and slices | none | membership and sizes remain valid and mutable after the graph is closed |
| `BFSResult`, `DFSResult` | Go-owned slices | none | traversal options are borrowed only during the call; results remain valid and mutable after the graph is closed |

`Graph` and `Vector` install finalizers as a leak fallback, but deterministic
code should still use `Close`, normally with `defer` or `t.Cleanup`.

Selectors are reusable. Their graph-independent shape is validated when they
are constructed, and bounds, missing endpoint pairs, and closed graphs are
validated each time they are materialized. `SelectedVertexIDs` and
`SelectedEdgeIDs` eagerly materialize a result while holding the graph lock;
they do not return an iterator that borrows the graph.

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

Density and whole-graph distance-summary options use the same optional weight
contract. Density accepts any finite weights. Diameter and average path length
delegate negative-weight rejection to their upstream shortest-path algorithm.
`IgnoreUnreachable` restricts summaries to reachable pairs; otherwise a
disconnected summary length is positive infinity. Diameter vertex and edge
paths, local transitivity slices, and all other structural metric results are
copied into Go-owned values before temporary C resources are destroyed. Local
transitivity follows materialized selector order, including duplicates; the
binding only deduplicates internally where required by the upstream call and
expands the result before returning it.

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
- `ConnectedComponents` on a graph with no vertices returns non-nil, empty
  membership and size slices with a component count of zero;
- breadth-first search requires at least one root and depth-first search
  requires one valid root, so traversal of an empty graph returns an error;
- `NewLattice` is the exception: it rejects nil and empty dimensions because
  a lattice needs at least one dimension.

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

No C object retains a pointer into Go memory. All C/igraph error codes are
converted to Go errors, and each constructor cleans up any successfully
initialized dependency before returning an error.

## Verification

Run the same complete verification entry point used by CI:

```sh
make verify
```

It checks formatting, runs `go vet`, tests and the statement-coverage floor
against the pinned igraph release in Docker, tests the inventory tool, and
checks that the generated API coverage report is current.
