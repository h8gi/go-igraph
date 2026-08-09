// Package igraph provides idiomatic Go bindings for the igraph C library.
//
// Graph and Vector values own C resources and should be closed when they are
// no longer needed. Close is idempotent, and methods on a closed resource
// return ErrClosed. Selectors, Matrix, Path, ConnectedComponents, traversal
// results, and structural result values are Go-owned and never require Close.
// Every returned slice, nested slice, matrix, and path remains valid after the
// source graph or temporary C resource is closed.
//
// Algorithm option and selector inputs are borrowed only for a call; any slice
// passed to C is first copied into temporary C storage. DirectionOut is the
// zero/default direction. Direction is ignored by undirected graphs. A nil
// weight slice requests an unweighted calculation, while a non-nil slice must
// contain one finite value per edge. Duplicate explicit selectors preserve
// caller order in returned results.
//
// Unreachable distances are positive infinity and an unreachable Path has
// Found false with non-nil empty slices. APIs with mathematically undefined
// scalar results document whether they return NaN or accept a mode that
// substitutes zero.
package igraph
