FROM golang:1.18

RUN apt-get update && \
    apt-get install -y libigraph-dev libxml2-dev && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY . .

RUN go test ./...

CMD ["go", "test", "./..."]
