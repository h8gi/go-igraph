# Contributing to go-igraph

Thank you for helping improve `go-igraph`.

## Development requirements

The local toolchain needs Go, Python 3, Docker, and `make`. The Docker build
pins both Go and the upstream C/igraph release, builds igraph from source, and
provides the same environment used by CI.

## Verification

Run the full local equivalent of the required CI job before submitting a pull
request:

```sh
make verify
```

This checks formatting, `go vet`, tests, race detection, the statement coverage
floor, the coverage-tool tests, and the generated upstream API inventory.

Useful focused targets are:

```sh
make docker-test           # build the test image and run go vet and tests
make docker-race           # run the test suite with the race detector
make docker-coverage       # write coverage.out into the working tree
make docker-coverage-check # enforce the coverage floor (90% by default)
```

Set `COVERAGE_MIN` to test a different statement coverage threshold. Set
`IGRAPH_VERSION` when validating a future upstream release.

## Upstream API coverage

The generated [API coverage report](docs/api-coverage.md) compares the functions
exported by the pinned upstream igraph release with explicit `//igraph:bind` and
`//igraph:internal` annotations in production Go files. It is an inventory for
planning binding work, not a claim of behavioral compatibility. The
[upstream API roadmap](docs/upstream-api-roadmap.md) defines the binding
strategy, milestones, and completion criteria.

Regenerate and check the report with:

```sh
make coverage
make coverage-check
```

Regeneration downloads the configured upstream source archive. For an offline
run, point the tool at an extracted source tree:

```sh
python3 tools/api_coverage.py --source-dir /path/to/igraph-1.0.1
```

The pinned version and upstream URLs live in
`tools/api_coverage_config.json`. Update that configuration and regenerate the
report when changing the comparison baseline.

## Examples

Follow [the example guidelines](docs/examples.md) when adding package examples
or standalone programs. Documentation examples must include expected output so
that `go test` validates them.

## Pull requests

Keep each pull request focused on one issue and include `Closes #<issue>` in
the description. Document whether public inputs are borrowed or copied and
whether returned values are Go-owned. New bindings should cover initialization
failure, upstream errors, early returns, empty values, and use after `Close`
where those paths apply.

## Contribution licensing

By submitting a contribution, you agree to license it under the project's
GNU General Public License, version 2 or later (`GPL-2.0-or-later`), and you
represent that you have the right to do so. Preserve applicable third-party
copyright and license notices, and identify any code or documentation adapted
from another source in the pull request.

The same responsibility applies to AI-assisted contributions: contributors
must review the resulting material and ensure that they have the rights needed
to submit it under the project license.
