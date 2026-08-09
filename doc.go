// Package igraph provides idiomatic Go bindings for the igraph C library.
//
// Graph and Vector values own C resources and should be closed when they are
// no longer needed. Close is idempotent, and methods on a closed resource
// return ErrClosed. Matrix, Path, DiameterResult, AveragePathLengthResult,
// VertexSelector, and EdgeSelector are ordinary Go values and never require
// Close. Slices returned by the package are Go-owned copies and remain valid
// after the source graph or temporary C resource is closed. Algorithm option
// slices are borrowed only for the call and copied before C/igraph uses them.
package igraph
