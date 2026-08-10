ARG GO_VERSION=1.26.0
FROM golang:${GO_VERSION}-bookworm AS test

ARG IGRAPH_VERSION=1.0.1

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        cmake \
        curl \
        g++ \
        libxml2-dev \
        make \
        pkg-config \
    && curl --fail --location --silent --show-error \
        "https://github.com/igraph/igraph/releases/download/${IGRAPH_VERSION}/igraph-${IGRAPH_VERSION}.tar.gz" \
        --output /tmp/igraph.tar.gz \
    && mkdir /tmp/igraph-source \
    && tar -xzf /tmp/igraph.tar.gz --strip-components=1 -C /tmp/igraph-source \
    && cmake -S /tmp/igraph-source -B /tmp/igraph-build \
        -DBUILD_SHARED_LIBS=ON \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_INSTALL_PREFIX=/usr/local \
    && cmake --build /tmp/igraph-build --parallel \
    && cmake --install /tmp/igraph-build \
    && rm -rf /var/lib/apt/lists/* /tmp/igraph-build /tmp/igraph-source /tmp/igraph.tar.gz

ENV LD_LIBRARY_PATH=/usr/local/lib

WORKDIR /workspace

COPY . .

RUN go vet ./... \
    && go test ./...

CMD ["go", "test", "-coverpkg=github.com/h8gi/go-igraph", "./...", "-coverprofile=coverage.out"]
