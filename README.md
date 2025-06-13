# go-igraph Docker

This repository includes a Dockerfile for running the tests with the
required `libigraph` and `libxml2` libraries.

## Build the image

```bash
docker build -t go-igraph .
```

The build process installs the required packages and executes `go test`
to verify that the environment compiles the project.

## Run tests inside the container

After building the image, tests can be executed by running:

```bash
docker run --rm go-igraph
```

This runs `go test ./...` inside the container.
