.PHONY: coverage coverage-check

coverage:
	python3 tools/api_coverage.py

coverage-check:
	@trap 'rm -f COVERAGE.tmp.md' EXIT; \
	python3 tools/api_coverage.py --output COVERAGE.tmp.md; \
	diff -u COVERAGE.md COVERAGE.tmp.md
