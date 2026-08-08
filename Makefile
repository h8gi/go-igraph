.PHONY: coverage coverage-check

coverage:
	python3 tools/api_coverage.py

coverage-check:
	python3 tools/api_coverage.py --check
