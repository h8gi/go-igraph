import tempfile
import unittest
from pathlib import Path

from api_coverage import discover_go_calls, discover_upstream_api, production_go_files


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


if __name__ == "__main__":
    unittest.main()
