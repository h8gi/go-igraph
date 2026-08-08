.PHONY: coverage coverage-check docker-test docker-coverage

IGRAPH_VERSION ?= 1.0.1
DOCKER_IMAGE ?= go-igraph-test

coverage:
	python3 tools/api_coverage.py

coverage-check:
	python3 tools/api_coverage.py --check

docker-test:
	docker build --build-arg IGRAPH_VERSION=$(IGRAPH_VERSION) --tag $(DOCKER_IMAGE) .

docker-coverage: docker-test
	docker run --rm --volume "$(CURDIR):/workspace" $(DOCKER_IMAGE)
