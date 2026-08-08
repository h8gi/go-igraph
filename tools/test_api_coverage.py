import pathlib
import tempfile
import unittest

import api_coverage


class CoverageTest(unittest.TestCase):
    def test_extracts_public_api_and_cgo_calls(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            include = root / "include"
            include.mkdir()
            (include / "igraph_demo.h").write_text(
                "IGRAPH_EXPORT igraph_error_t\nigraph_alpha(void);\n"
                "IGRAPH_EXPORT void igraph_beta(int value);\n"
            )
            (root / "demo.go").write_text("package demo\nfunc f() { C.igraph_alpha() }\n")
            (root / "demo_test.go").write_text("package demo\nfunc f() { C.igraph_beta() }\n")

            self.assertEqual(
                api_coverage.public_api(include),
                {"igraph_alpha": "igraph_demo.h", "igraph_beta": "igraph_demo.h"},
            )
            self.assertEqual(dict(api_coverage.cgo_calls(root)), {"igraph_alpha": ["demo.go:2"]})


if __name__ == "__main__":
    unittest.main()
