#!/usr/bin/env python3
"""Unit tests for tools/upstream_api.py and tools/igraph_c_ref.py."""

import tempfile
import unittest
from pathlib import Path

from upstream_api import (
    build_doc_url,
    DOCS_BASE_URL,
    find_cgo_call_locations,
)


class TestIgraphCRef(unittest.TestCase):
    def test_build_doc_url_known_header(self):
        url = build_doc_url("igraph_layout.h", "igraph_layout_sugiyama")
        expected = f"{DOCS_BASE_URL}/igraph-Layout.html#igraph_layout_sugiyama"
        self.assertEqual(url, expected)

    def test_build_doc_url_unknown_header(self):
        url = build_doc_url("igraph_unknown.h", "igraph_unknown_fn")
        expected = f"{DOCS_BASE_URL}/cigraph-index.html#igraph_unknown_fn"
        self.assertEqual(url, expected)

    def test_find_cgo_call_locations_exact_matching(self):
        with tempfile.TemporaryDirectory() as temp:
            repo = Path(temp)
            (repo / "betweenness.go").write_text(
                "package main\n"
                "// C.igraph_betweenness_cutoff()\n"
                "func Foo() {\n"
                "    C.igraph_betweenness(nil)\n"
                "}\n",
                encoding="utf-8",
            )
            calls = find_cgo_call_locations(repo, "igraph_betweenness")
            self.assertEqual(len(calls), 1)
            self.assertEqual(calls[0][0], "betweenness.go")
            self.assertEqual(calls[0][1], 4)
            self.assertIn("C.igraph_betweenness(nil)", calls[0][2])


if __name__ == "__main__":
    unittest.main()
