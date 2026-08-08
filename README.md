# go-igraph

Go bindings for [igraph](https://igraph.org/).

## Upstream API coverage

The generated [API coverage report](docs/api-coverage.md) compares the functions
exported by a pinned upstream igraph release with direct `C.igraph_*` calls in
production Go files. This is an inventory for planning binding work, not a claim
of behavioral compatibility.

Regenerate the report (downloads the configured upstream source archive):

```sh
make coverage
```

Check that the committed report is current:

```sh
make coverage-check
```

GitHub Actions checks Go formatting, runs `go vet` and the Go and coverage-tool
tests, and runs `make coverage-check` for every pull request and for pushes to
`main`.

For an offline run, point the tool at an already extracted igraph source tree:

```sh
python3 tools/api_coverage.py --source-dir /path/to/igraph-1.0.1
```

## Docker

The Docker build pins both Go and the upstream C/igraph release, builds igraph
from source, and runs the Go tests:

```sh
make docker-test
```

Run the tests against the working tree and write `coverage.out` locally with:

```sh
make docker-coverage
```

Override `IGRAPH_VERSION` when validating a future upstream release.

The pinned version and upstream URLs live in
`tools/api_coverage_config.json`. Update the config and regenerate the report
when changing the comparison baseline.
