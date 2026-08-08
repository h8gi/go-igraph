import tempfile
import unittest
from pathlib import Path

from api_coverage import (
    Annotation,
    discover_annotations,
    discover_go_calls,
    discover_upstream_api,
    production_go_files,
    render,
    validate_inventory,
)


class ApiCoverageTest(unittest.TestCase):
    def test_discovers_multiline_exported_declarations(self):
        with tempfile.TemporaryDirectory() as temp:
            include = Path(temp)
            (include / "igraph_example.h").write_text(
                "IGRAPH_EXPORT igraph_error_t\nigraph_example(const igraph_t *graph);\n"
                "IGRAPH_EXPERIMENTAL IGRAPH_EXPORT igraph_error_t igraph_new_api(void);\n"
                "igraph_error_t igraph_private(void);\n",
                encoding="utf-8",
            )
            self.assertEqual(
                discover_upstream_api(include),
                {
                    "igraph_example": "igraph_example.h",
                    "igraph_new_api": "igraph_example.h",
                },
            )

    def test_finds_production_calls_and_excludes_tests(self):
        with tempfile.TemporaryDirectory() as temp:
            repo = Path(temp)
            (repo / "graph.go").write_text("package p\n// C.igraph_real()\n", encoding="utf-8")
            (repo / "graph_test.go").write_text("package p\n// C.igraph_test()\n", encoding="utf-8")
            self.assertEqual([p.name for p in production_go_files(repo)], ["graph.go"])
            self.assertEqual(discover_go_calls(repo), {"igraph_real"})

    def test_discovers_annotations_attached_to_declarations(self):
        with tempfile.TemporaryDirectory() as temp:
            repo = Path(temp)
            (repo / "graph.go").write_text(
                "package p\n"
                "//igraph:bind igraph_empty\n"
                "func NewGraph() {}\n"
                "//igraph:internal igraph_destroy\n"
                "func (g *Graph) close() {}\n",
                encoding="utf-8",
            )
            self.assertEqual(
                discover_annotations(repo),
                [
                    Annotation("bind", "igraph_empty", "NewGraph", "graph.go", 2),
                    Annotation("internal", "igraph_destroy", "close", "graph.go", 4),
                ],
            )

    def test_rejects_unknown_duplicate_and_unclassified_symbols(self):
        declarations = {"igraph_empty": "igraph.h", "igraph_destroy": "igraph.h"}
        annotations = [
            Annotation("bind", "igraph_empty", "NewGraph", "graph.go", 1),
            Annotation("internal", "igraph_empty", "close", "graph.go", 2),
            Annotation("bind", "igraph_unknown", "Unknown", "graph.go", 3),
        ]
        with self.assertRaisesRegex(ValueError, "duplicate annotation for igraph_empty"):
            validate_inventory(declarations, {"igraph_destroy"}, annotations, set())

    def test_rejects_binding_on_unexported_go_symbol(self):
        annotation = Annotation("bind", "igraph_empty", "newGraph", "graph.go", 1)
        with self.assertRaisesRegex(ValueError, "binding target newGraph is not exported"):
            validate_inventory({"igraph_empty": "igraph.h"}, set(), [annotation], set())

    def test_renders_all_inventory_classifications_deterministically(self):
        declarations = {
            "igraph_bound": "b.h",
            "igraph_internal": "a.h",
            "igraph_missing": "a.h",
            "igraph_unsupported": "b.h",
        }
        annotations = [
            Annotation("bind", "igraph_bound", "Bound", "graph.go", 1),
            Annotation("internal", "igraph_internal", "helper", "graph.go", 2),
        ]
        config = {
            "upstream": "igraph/igraph",
            "version": "1",
            "release_url": "https://example.test",
            "intentionally_unsupported": ["igraph_unsupported"],
        }
        report = render(config, declarations, annotations)
        self.assertIn("| `igraph_bound` | `b.h` | User-facing | `Bound` |", report)
        self.assertIn("| `igraph_internal` | `a.h` | Internal | `helper` |", report)
        self.assertIn("| `igraph_unsupported` | `b.h` | Intentionally unsupported | — |", report)
        self.assertIn("| `igraph_missing` | `a.h` | Missing | — |", report)


if __name__ == "__main__":
    unittest.main()
