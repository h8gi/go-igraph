#!/usr/bin/env python3
"""Unit tests for tools/upstream_api.py and tools/igraph_c_ref.py."""

import tempfile
import tarfile
import unittest
from pathlib import Path

from upstream_api import (
    build_doc_url,
    discover_upstream_declarations,
    DOCS_BASE_URL,
    find_cgo_call_locations,
    load_or_build_declaration_index,
    download_or_get_source,
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
                "// C.igraph_betweenness(nil) is documentation, not a call.\n"
                "func Foo() {\n"
                "    C.igraph_betweenness(nil)\n"
                "}\n",
                encoding="utf-8",
            )
            (repo / "algorithm_cgo.c").write_text(
                "// igraph_betweenness(graph, weights, result, vids, directed, normalized);\n"
                "void wrapper(void) {\n"
                "    igraph_betweenness(graph, weights, result, vids, directed, normalized);\n"
                "    igraph_betweenness_cutoff(graph, weights, result, vids, directed, cutoff);\n"
                "}\n",
                encoding="utf-8",
            )
            (repo / "igraph_error_cgo.h").write_text(
                "#define CALL() igraph_betweenness(graph, weights, result, vids, directed, normalized)\n",
                encoding="utf-8",
            )
            calls = find_cgo_call_locations(repo, "igraph_betweenness")
            self.assertEqual(len(calls), 3)
            self.assertEqual(calls[0][0], "betweenness.go")
            self.assertEqual(calls[0][1], 5)
            self.assertIn("C.igraph_betweenness(nil)", calls[0][2])
            self.assertEqual(calls[1][0], "algorithm_cgo.c")
            self.assertEqual(calls[1][1], 3)
            self.assertEqual(calls[2][0], "igraph_error_cgo.h")

    def test_find_cgo_call_locations_in_embedded_preambles(self):
        with tempfile.TemporaryDirectory() as temp:
            repo = Path(temp)
            (repo / "line_preamble.go").write_text(
                "package main\n"
                "// #include <igraph.h>\n"
                "// // igraph_vector_int_list_size(list) is only a C comment.\n"
                "// static int wrapper(const igraph_vector_int_list_t *list) {\n"
                "//   igraph_vector_int_list_size_extra(list);\n"
                "//   return igraph_vector_int_list_size(list);\n"
                "// }\n"
                'import "C"\n',
                encoding="utf-8",
            )
            (repo / "block_preamble.go").write_text(
                "package main\n"
                "/*\n"
                "#include <igraph.h>\n"
                "static int wrapper(const igraph_vector_int_list_t *list) {\n"
                "  return igraph_vector_int_list_size(list);\n"
                "}\n"
                "*/\n"
                'import "C"\n',
                encoding="utf-8",
            )
            (repo / "detached_comment.go").write_text(
                "package main\n"
                "// igraph_vector_int_list_size(list) is documentation.\n"
                "\n"
                'import "C"\n',
                encoding="utf-8",
            )
            calls = find_cgo_call_locations(repo, "igraph_vector_int_list_size")
            self.assertEqual(
                calls,
                [
                    (
                        "block_preamble.go",
                        5,
                        "return igraph_vector_int_list_size(list);",
                    ),
                    (
                        "line_preamble.go",
                        6,
                        "//   return igraph_vector_int_list_size(list);",
                    ),
                ],
            )

    def test_preserves_complete_declarations_and_nested_parameters(self):
        with tempfile.TemporaryDirectory() as temp:
            include = Path(temp)
            (include / "igraph_example.h").write_text(
                "IGRAPH_EXPORT void igraph_destroy(igraph_t *graph);\n"
                "IGRAPH_EXPORT igraph_int_t igraph_vcount(const igraph_t *graph);\n"
                "IGRAPH_EXPORT handler_t *igraph_set_handler(void (*handler)(int));\n",
                encoding="utf-8",
            )
            declarations = discover_upstream_declarations(include)
            self.assertEqual(
                declarations["igraph_destroy"].declaration,
                "IGRAPH_EXPORT void igraph_destroy(igraph_t *graph);",
            )
            self.assertEqual(
                declarations["igraph_vcount"].declaration,
                "IGRAPH_EXPORT igraph_int_t igraph_vcount(const igraph_t *graph);",
            )
            self.assertEqual(
                declarations["igraph_set_handler"].params,
                "void (*handler)(int)",
            )

    def test_declaration_index_reuses_matching_cache(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            include = root / "include"
            include.mkdir()
            (include / "igraph_example.h").write_text(
                "IGRAPH_EXPORT void igraph_destroy(void);\n",
                encoding="utf-8",
            )
            config = {
                "version": "test",
                "source_archive_url": "https://example.test/igraph.tar.gz",
                "documentation_base_url": DOCS_BASE_URL,
            }
            from unittest.mock import patch

            with patch("upstream_api.get_cache_dir", return_value=root):
                first = load_or_build_declaration_index(config, include)
                (include / "igraph_example.h").write_text("broken", encoding="utf-8")
                second = load_or_build_declaration_index(config, include)
            self.assertEqual(first, second)

    def test_declaration_index_tolerates_unwritable_cache(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            include = root / "include"
            include.mkdir()
            (include / "igraph_example.h").write_text(
                "IGRAPH_EXPORT void igraph_destroy(void);\n",
                encoding="utf-8",
            )
            config = {
                "version": "test",
                "source_archive_url": "https://example.test/igraph.tar.gz",
                "documentation_base_url": DOCS_BASE_URL,
            }
            from unittest.mock import patch

            with (
                patch("upstream_api.get_cache_dir", return_value=root),
                patch("upstream_api.tempfile.NamedTemporaryFile", side_effect=PermissionError),
            ):
                declarations = load_or_build_declaration_index(config, include)
            self.assertIn("igraph_destroy", declarations)

    def test_download_rejects_archive_path_traversal(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            archive = root / "igraph-test.tar.gz"
            payload = root / "payload"
            payload.write_text("unsafe", encoding="utf-8")
            with tarfile.open(archive, "w:gz") as bundle:
                bundle.add(payload, arcname="../escape")
            config = {
                "version": "test",
                "source_archive_url": "https://example.test/igraph.tar.gz",
            }
            from unittest.mock import patch

            with patch("upstream_api.get_cache_dir", return_value=root):
                with self.assertRaisesRegex(ValueError, "unsafe path"):
                    download_or_get_source(config)


if __name__ == "__main__":
    unittest.main()
