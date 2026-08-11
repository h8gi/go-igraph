.PHONY: coverage coverage-check coverage-tool-test docker-test docker-race docker-coverage docker-coverage-check format-check verify c-ref

IGRAPH_VERSION ?= 1.0.1
DOCKER_IMAGE ?= go-igraph-test
COVERAGE_MIN ?= 90.0

coverage:
	python3 tools/api_coverage.py

coverage-check:
	python3 tools/api_coverage.py --check

coverage-tool-test:
	python3 -m unittest discover -s tools -p 'test_*.py' -v

c-ref:
	python3 tools/igraph_c_ref.py lookup $(SYMBOL)


format-check:
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "The following Go files need formatting:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

docker-test:
	docker build --build-arg IGRAPH_VERSION=$(IGRAPH_VERSION) --tag $(DOCKER_IMAGE) .

docker-coverage: docker-test
	docker run --rm --volume "$(CURDIR):/workspace" $(DOCKER_IMAGE)

docker-coverage-check: docker-test
	docker run --rm $(DOCKER_IMAGE) sh -c 'go test -coverpkg=github.com/h8gi/go-igraph ./... -coverprofile=/tmp/coverage.out && coverage=$$(go tool cover -func=/tmp/coverage.out | awk "/^total:/ { sub(/%/, \"\", \$$3); print \$$3 }") && echo "statement coverage: $$coverage% (minimum $(COVERAGE_MIN)%)" && awk -v coverage="$$coverage" -v minimum="$(COVERAGE_MIN)" "BEGIN { exit coverage >= minimum ? 0 : 1 }"'

docker-race: docker-test
	docker run --rm $(DOCKER_IMAGE) go test -race ./...

verify: format-check docker-coverage-check docker-race coverage-tool-test coverage-check
