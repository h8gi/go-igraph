// Package igraph provides idiomatic Go bindings for the igraph C library.
//
// Graph and Vector values own C resources and should be closed when they are
// no longer needed. Close is idempotent, and methods on a closed resource
// return ErrClosed. Selectors, Matrix, Path, Cycle, GirthResult,
// SimpleCyclesResult, PathsResult, path collections, reachability, Voronoi,
// path histograms, Eulerian traversals, cycle-basis, feedback-set,
// census, triangle, motif, and
// graphlet collections,
// ConnectedComponents, articulation-point and bridge slices,
// BiconnectedComponents, IDMapping, GraphIDMapping, subgraph and graph-operator
// result provenance, traversal
// results, structural results, centrality results, centralization results,
// CommunityPartition, HierarchicalCommunity, and SpinglassSingleResult are
// Go-owned and never require Close. Graphs returned by subgraph,
// decomposition, graph-operator, neighborhood, transitive-closure, and spanner
// APIs are independently owned and must each
// be closed.
// Every returned slice, nested slice, matrix, path, community partition, and
// hierarchical community dendrogram remains valid after the source graph or
// temporary C resource is closed.
//
// AttributeScope, AttributeType, and AttributeMetadata provide the shared
// Go-native vocabulary for graph, vertex, and edge boolean, numeric, and
// string metadata. Attribute names and string values are UTF-8 and may not
// contain embedded NUL bytes; names are non-empty while values may be empty.
// Metadata returned by package APIs is Go-owned. The package installs and owns
// the process-global C attribute table during package initialization, before
// callers can construct graphs, and does not expose a replacement hook. The C
// handler is experimental in pinned igraph 1.0.1; its lifecycle remains an
// internal implementation detail rather than part of the Go API contract.
// GraphAttributes returns graph-level metadata in lexical name order. Typed
// graph attribute getters return ErrAttributeNotFound for missing names and
// ErrAttributeTypeMismatch for wrong types. Setters copy names and string
// values before returning, allow empty strings, reject non-finite numeric
// values, and only overwrite attributes of the same type.
// Vertex and edge attribute vectors are aligned with current IDs. Full-vector
// setters require exact scope length and are the only element operations that
// create attributes; scalar setters update an existing attribute at a checked
// ID. Nil and empty vectors both denote length zero. Topology growth appends
// missing-value defaults of numeric NaN, empty string, and Boolean false.
// Returned vectors are non-nil Go-owned copies that survive graph closure.
// Attribute names, scalar string values, and input vectors are borrowed only
// for a synchronous call and copied before return; no Go backing storage is
// retained by C-igraph.
//
// Graph readers and writers borrow an open, seekable regular *os.File only for
// the synchronous call and never close it. Readers snapshot from the current
// offset without changing it and return an independently owned graph. Writers
// retain neither the file nor graph data after return. GraphML preserves the
// supported Boolean, numeric, and string attributes; format-specific losses
// and options are documented by each method. Temporary FILE handles and
// process-locale coordination remain internal. Concurrent interchange calls
// are safe; a writer racing Close completes under the graph lock or returns
// ErrClosed.
//
// Algorithm option and selector inputs are borrowed only for a call; any slice
// passed to C is first copied into temporary C storage. DirectionOut is the
// zero/default direction. Direction is ignored by undirected graphs. A nil
// weight slice requests an unweighted calculation, while a non-nil slice must
// contain one finite value per edge and may have stricter method-specific sign
// constraints. Distance-based and betweenness centralities require positive
// weights; spectral and ranking centralities require non-negative weights.
// Graphlet methods treat nil weights as unit weights and otherwise require
// finite non-negative values on graphs that are simple when directions are
// ignored.
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
// Graph-returning transformations borrow their source graphs only for the
// synchronous call and return independently closable graphs. Multi-graph
// operators deduplicate repeated operands and acquire distinct graph locks in
// stable order. Many-graph union, intersection, and disjoint union return one
// exact Go-owned mapping per input and expose no C graph-pointer or mapping-list
// lifecycle. Component graphs own separate C resources; ID mappings,
// composition provenance, articulation points, bridges, and biconnected nested
// collections are non-nil Go-owned values that survive source closure.
// Robustness analysis includes separator and minimal-separator predicates,
// biconnectivity and biconnected decomposition, deterministic experimental
// bond/site/edge-list percolation curves, and cohesive block hierarchy. A
// CohesiveBlocksResult owns its nested Go slices while its BlockTree is an
// independently closable graph. Full separator enumeration is intentionally
// not exposed because pinned igraph materializes unbounded complete result
// lists before a Go-side limit could be applied.
//
// Community structure APIs provide flat algorithms (Multilevel, Leiden,
// Label Propagation, Infomap, Fluid), hierarchical algorithms (Walktrap,
// FastGreedy, Edge Betweenness), spectral algorithms (Leading Eigenvector),
// simulated annealing (Spinglass, SpinglassSingle), and exact optimization
// (Optimal Modularity). All returned community membership vectors use contiguous
// 0-indexed integer cluster IDs. Hierarchical communities expose dendrogram cuts
// via MembershipAt and OptimalMembership. Community comparison metrics
// (CompareCommunities, SplitJoinDistance) evaluate partition distance (VI, Split-Join)
// and similarity (NMI, Rand, Adjusted Rand). Stochastic algorithms accept an optional
// Seed parameter for reproducible random execution protected by a global package RNG lock.
// RANDESU motif sampling follows the same RNG contract. Its histogram keeps
// upstream NaN markers for impossible isomorphism classes; every finite count
// and all graphlet clique, threshold, and coefficient slices are Go-owned.
// Advanced random graph models use the same package RNG contract. Expected
// degree, fitness, preference, block, growth, citation, correlated, geometric,
// and latent-position inputs are borrowed synchronously. Returned graphs own
// independent C resources and must be closed; auxiliary type assignments,
// coordinates, and sample matrices are Go-owned. Public matrices always use
// one vertex or sample per row. Multi-graph results own each graph separately,
// and directed endpoint rewiring commits through clone-and-swap so failures
// leave the receiver and its attributes unchanged. APIs that mirror upstream
// experimental generators state that status explicitly.
//
// Unreachable distances are positive infinity and an unreachable Path has
// Found false with non-nil empty slices. APIs with mathematically undefined
// scalar results document whether they return NaN or accept a mode that
// substitutes zero. Centrality cutoffs use nil for unlimited paths and explicit
// finite non-negative values otherwise; callers never pass upstream negative
// sentinels. Feedback edge and vertex weights likewise require finite
// non-negative values and accept zero. Public solver options expose only
// convergence settings and never retain ARPACK, PRPACK, or other C solver
// objects.
package igraph
