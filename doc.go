// Package igraph provides idiomatic Go bindings for the igraph C library.
//
// Graph and Vector values own C resources and should be closed when they are
// no longer needed. Close is idempotent, and methods on a closed resource
// return ErrClosed. Selectors, Matrix, Path, ConnectedComponents,
// articulation-point and bridge slices, BiconnectedComponents, IDMapping,
// GraphIDMapping, subgraph and graph-operator result provenance, traversal
// results, structural results, centrality results, and centralization results
// are Go-owned and never require Close. Graphs returned by subgraph,
// decomposition, and graph-operator APIs are independently owned and must each
// be closed.
// Every returned slice, nested slice, matrix, and path remains valid after the
// source graph or temporary C resource is closed.
//
// Algorithm option and selector inputs are borrowed only for a call; any slice
// passed to C is first copied into temporary C storage. DirectionOut is the
// zero/default direction. Direction is ignored by undirected graphs. A nil
// weight slice requests an unweighted calculation, while a non-nil slice must
// contain one finite value per edge and may have stricter method-specific sign
// constraints. Distance-based and betweenness centralities require positive
// weights; spectral and ranking centralities require non-negative weights.
// Duplicate explicit selectors preserve caller order in returned results.
// Destructive deletion APIs materialize selectors before mutation and use a
// temporary graph so validation, initialization, upstream, and conversion
// failures leave the original graph unchanged.
// In-place simplification and direction conversion use the same clone-and-swap
// atomicity: an error leaves the receiver unchanged. These transformations do
// return exact structural edge mappings when IDMapping can represent the
// provenance. One-to-many mutual directed conversion explicitly marks its edge
// mapping unavailable instead of inventing a correspondence. Equivalent
// parallel edges use source/result edge-ID order as a deterministic structural
// convention, not as attribute provenance.
//
// Unreachable distances are positive infinity and an unreachable Path has
// Found false with non-nil empty slices. APIs with mathematically undefined
// scalar results document whether they return NaN or accept a mode that
// substitutes zero. Centrality cutoffs use nil for unlimited paths and explicit
// finite non-negative values otherwise; callers never pass upstream negative
// sentinels. Public solver options expose only convergence settings and never
// retain ARPACK, PRPACK, or other C solver objects.
package igraph
