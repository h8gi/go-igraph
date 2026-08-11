# Package architecture

This document records how go-igraph should grow while keeping a Go-native API
over C/igraph. It is an architecture decision, not a directory-reorganization
plan.

## Decision

`github.com/h8gi/go-igraph` remains one public Go package. New feature domains
should normally be added as cohesive files in that package, with adjacent C
shims where a direct cgo call is not sufficient. Package-independent developer
tools may live outside the package, but graph algorithms and their cgo resource
management should not be split into public domain packages or `internal/`
packages merely to reduce the number of files in the repository root.

This decision can be revisited when a proposed boundary has its own useful
public vocabulary, does not need access to the private graph representation,
and avoids duplicating ownership, locking, validation, and conversion logic.
File count alone is not sufficient reason to introduce a package boundary.

## Current shape

At the time of this decision, the root contains 138 Go files, including 94 test
files, plus 19 C files and 20 headers. The package documentation exposes about
349 declarations, 181 of which claim user-facing C/igraph bindings. Another 51
upstream functions are explicit internal dependencies. There are about 180
methods on `*Graph`.

The files already fall into three practical groups:

- public graph types, selectors, value conversions, ownership, locking, error
  translation, and RNG coordination shared by all domains;
- feature slices such as centrality, community detection, flow, layout, and
  clique operations, with focused tests and optional `_cgo.c`/`_cgo.h` pairs;
- development tools and generated documentation under `tools/` and `docs/`.

Tests intentionally use both package forms. External `igraph_test` tests verify
the caller-visible contract. Same-package tests exercise cleanup, failure
injection, locking, and other behavior that cannot be tested through the public
API alone.

These counts are a point-in-time description rather than architecture limits.
They can be reproduced with `find`, `go doc`, and the binding annotations used
by `tools/api_coverage.py`.

## Why one public package remains appropriate

In Go, moving files to a subdirectory creates another package; it is not a
cosmetic source-tree grouping mechanism. A public domain package such as
`igraph/community` could not add methods to the non-local `igraph.Graph` type.
It would need free functions, a new wrapper type, or exported access to graph
internals. Each choice would fragment the API or weaken the existing ownership
boundary.

The current package also has one private representation of `igraph_t`, one
graph lock and `Close` contract, and shared conversion and cleanup helpers. C
types and temporary storage remain inside this boundary. Keeping algorithms
with that boundary makes it possible to:

- borrow graph and slice inputs only for a synchronous call;
- serialize operations safely with `Close` and lock multiple graphs in a
  stable order;
- translate every upstream error without exposing C types;
- destroy partially initialized C values on every failure path; and
- copy ordinary results into Go-owned values before releasing graph locks.

Splitting domains before those concerns have a package-independent interface
would create wrappers or duplicated cgo plumbing and could introduce import
cycles between core graph types and algorithms. It would also make API
discovery depend on knowing which upstream domain owns an operation. A single
package lets callers start from `Graph` and Go documentation instead.

The large root directory is primarily a contributor-navigation concern. It
does not require callers to import more packages and is better addressed with
consistent names, documentation, and feature-sized reviews.

## Alternatives considered

### Public domain subpackages

Packages for areas such as `layout`, `community`, or `flow` would shorten the
root file listing, but most operations still require `*Graph` and shared
selectors, matrices, ownership rules, and locking. Free functions would be
less discoverable than methods, wrapper graph types would obscure ownership,
and exporting an implementation handle would expose the cgo boundary. This is
not recommended for the current API.

A future package is reasonable for a genuinely independent facility whose
types are useful without private graph access. It must have a clear import
direction and must not create an alternative graph lifetime model.

### Internal implementation packages

An `internal/` package is useful only for package-independent Go code. Moving
algorithms there would still require an exported bridge to private graph state,
and moving low-level graph state there would make the public package a broad
forwarding layer. Both add indirection without a demonstrated safety or reuse
benefit.

Developer tooling, parsers, inventory models, and generators may be extracted
when they become shared Go programs. The current Python coverage and reference
tools should remain in `tools/`; placing them under Go's `internal/` convention
would not improve their boundary.

### Generated or consolidated cgo glue

Generation is appropriate for repetitive, declarative adapters when the input,
generated output, ownership template, and regeneration check are all reviewed
and deterministic. It must not infer public API design from C signatures or
hide cleanup and error paths that need individual tests.

The default remains an adjacent `<domain>_cgo.c` and `<domain>_cgo.h` pair.
Combining all shims into a single file would reduce file count but enlarge the
rebuild and review surface and weaken feature ownership. Shims can be shared
when they implement a real cross-domain primitive rather than coincidentally
similar calls.

## Dependency and placement rules

Dependencies should point inward in this order:

1. exported feature APIs use shared private validation, selection, locking,
   conversion, RNG, and ownership helpers;
2. those helpers own all interaction with temporary C values;
3. focused C shims call pinned C/igraph APIs and return error codes or plain
   output values to Go;
4. tools inspect source and annotations but are never runtime dependencies of
   the package.

The following placement rules keep that direction visible:

- `<domain>.go` contains the public vocabulary and implementation for one
  coherent feature slice;
- `<domain>_cgo.c` and `<domain>_cgo.h` contain callbacks or error-handler
  wrappers needed by that slice; direct cgo calls may stay in the Go file;
- `<domain>_test.go` prefers `package igraph_test` for public behavior;
- `<domain>_internal_test.go` uses `package igraph` for failure seams, resource
  cleanup, and private invariants;
- `milestone<N>_integration_test.go` is reserved for cross-feature contracts
  that justify a milestone-wide test;
- `example_<domain>_test.go` contains output-checked package examples, while
  `examples/<domain>/` contains a runnable program when it teaches a broader
  workflow;
- shared helpers belong in a domain-neutral file only after at least two real
  consumers establish the common contract; and
- generated files must identify their generator and are checked for freshness
  by `make verify`.

New files should use domain vocabulary rather than upstream header names when
the Go API combines several headers. A feature slice should remain reviewable
as one issue and must follow the ownership and failure-path requirements in the
upstream API roadmap.

## What “complete” means

Raw C declaration coverage is an inventory, not a product-completion score.
The pinned report currently includes generated containers and low-level
facilities alongside graph algorithms. Exposing every declaration would leak
C lifecycle operations, callbacks, file handles, solver details, or APIs that
are more safely composed behind one bounded Go operation.

Every upstream function relevant to an audited domain should instead receive
one of these dispositions:

- **user-facing**: an idiomatic exported Go API binds the function;
- **composed**: a Go API intentionally uses or replaces one or more upstream
  functions without exposing each function one-for-one;
- **internal**: the function implements a public contract but is not itself a
  useful public Go API;
- **intentionally unsupported**: exposing the function would violate the
  package's safety, ownership, concurrency, or API-design rules;
- **deferred**: the function belongs to a domain whose Go contract and use case
  have not yet been designed.

A domain is complete when every relevant declaration has a reviewed
disposition, the user-facing feature slice is coherent, and the roadmap's
binding definition of done is satisfied. Its raw bound percentage may remain
below 100%. The overall `181 / 2015` figure describes only one disposition and
must not be used alone to prioritize work.

Future coverage reporting should separate algorithmic public APIs from
generated containers and low-level support declarations, show composed
coverage explicitly, and summarize dispositions by audited domain. Until the
tool supports those distinctions, roadmap prose and configuration remain the
source of truth for composed, unsupported, and deferred decisions.

## Consequences and next steps

No package move, import-path change, or compatibility layer follows from this
decision. New domains can continue using the established feature-slice pattern.
The next architecture work should be limited to focused follow-up issues:

1. extend the coverage model and generated report with explicit composed and
   deferred/domain-audit dispositions; and
2. select the next domain milestone using user value and the availability of a
   coherent Go ownership contract, not the largest raw declaration gap.

With those definitions in place, the recommended next feature domain is cycle
analysis. The pinned `igraph_cycles.h` surface is small and cohesive, and its
scalar, basis, ordering, and explicitly bounded enumeration operations can
reuse the graph, direction, nested-result, and bounded-enumeration contracts
already established by earlier milestones. The experimental status of
`igraph_simple_cycles` and the disposition of its callback variant must be
explicit in that milestone plan.

Motifs and graphlets are a reasonable second candidate, but should follow
cycle analysis rather than be grouped into the same milestone. Their sampling,
cut-probability, callback, weighted, and RNG semantics require a separate
contract review. This ordering is a design recommendation, not authorization
to create the milestones before their focused planning issues are accepted.

Reconsider a larger restructure only with measurements that identify a
specific compilation, navigation, duplication, or safety problem and a
proposed boundary that improves it.
