#!/usr/bin/env python3
"""Unit tests for tools/igraph_c_ref.py."""

import unittest
from pathlib import Path
from unittest.mock import patch

from tools.igraph_c_ref import (
    build_doc_url,
    generate_audit_tips,
    HEADER_TO_DOC_MODULE,
    DOCS_BASE_URL,
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

    def test_generate_audit_tips(self):
        tips = generate_audit_tips(
            "igraph_test",
            "const igraph_t *graph, igraph_matrix_t *res, igraph_vector_t *weights",
        )
        tips_text = " ".join(tips)
        self.assertIn("igraph_vector_t", tips_text)
        self.assertIn("igraph_matrix_t", tips_text)
        self.assertIn("igraph_t", tips_text)
        self.assertIn("igraph_error_t", tips_text)


if __name__ == "__main__":
    unittest.main()
