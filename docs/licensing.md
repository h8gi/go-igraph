# Licensing and provenance

## Project license

`go-igraph` is free software licensed under the GNU General Public License as
published by the Free Software Foundation, either version 2 of the License, or
(at your option) any later version (`GPL-2.0-or-later`). The canonical GPLv2
text is in [`LICENSE`](../LICENSE), and the copyright and application notice is
in [`COPYRIGHT`](../COPYRIGHT).

Except where another notice applies, this grant covers the Go, C, header,
test, tool, example, and documentation material for which the project
copyright holder has licensing authority. It also covers that material in
prior repository revisions. Third-party material is not relicensed and
remains under its original terms.

The GPL regulates copying, modification, and distribution. It does not
restrict running the software privately. Anyone distributing source or a
linked program must evaluate and satisfy the GPL requirements applicable to
that distribution.

## C/igraph dependency

This repository does not vendor the C/igraph implementation. The cgo binding
uses `pkg-config` and C/igraph's public C interface to link with a separately
installed C/igraph library. C/igraph states that its code is licensed under
GPL version 2 or later:

- <https://igraph.org/c/html/latest/>
- <https://github.com/igraph/igraph/blob/main/COPYING>

The exact obligations for a distributed binary can also depend on how
C/igraph was built and which optional dependencies it contains. In
particular, an enabled GPLv3-only dependency can require distributing the
combined work under GPLv3. The `or-later` choice for `go-igraph` preserves
that compatibility.

## Repository history

Early revisions beginning with commit
[`61c5ec1`](https://github.com/h8gi/go-igraph/commit/61c5ec190001a3b1591864af2734618239ed4f4e)
contained a Go adaptation of C/igraph's `examples/tutorial/tutorial1.c`. The
adapted sample function was removed in commit
[`b9f4029`](https://github.com/h8gi/go-igraph/commit/b9f4029b56f51c23b9fae1233c157d5ab8031254),
but remains visible in Git history. C/igraph distributes that tutorial with
its GPL-2.0-or-later code. To the extent that the historical adaptation is
copyrightable or derived from that tutorial, it is distributed under the
same GPL-2.0-or-later terms and its upstream copyright remains intact.

The current binding contains interoperability declarations using C/igraph
API names, types, and signatures, but does not contain a vendored copy of
C/igraph's algorithm implementations.

## Contributions

Contributions are accepted under GPL-2.0-or-later. Contributors must have the
right to submit their work on those terms and must preserve or disclose any
applicable third-party attribution. See [`CONTRIBUTING.md`](../CONTRIBUTING.md).

This document records the project's licensing intent and provenance; it is
not legal advice.
