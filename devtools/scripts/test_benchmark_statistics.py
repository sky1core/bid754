import ast
import contextlib
import io
import json
from pathlib import Path
import subprocess
import unittest

ROOT = Path(__file__).resolve().parents[2]


class BenchmarkStatisticsTest(unittest.TestCase):
    def test_python_report(self):
        source = ROOT / "bid754-codec-py/benchmarks/bench_runner.py"
        tree = ast.parse(source.read_text())
        tree.body = [node for node in tree.body
                     if isinstance(node, (ast.Import, ast.ImportFrom)) and
                     any(alias.name == "statistics" for alias in node.names)
                     or isinstance(node, ast.FunctionDef) and node.name == "_bench_row"]
        for samples, expected in [([1, 3], 2), ([9, 1, 5], 5), ([8], 8), ([8, 2, 4, 10], 6)]:
            timings = iter([1_000_000] + [value * 1000 for value in samples])
            namespace = {"_run_batch": lambda *args: next(timings),
                         "_CALIBRATION_ITERS": 1, "_TARGET_SAMPLE_NS": 1000}
            exec(compile(tree, str(source), "exec"), namespace)
            output = io.StringIO()
            with contextlib.redirect_stdout(output):
                namespace["_bench_row"]("fixture", len(samples), lambda _: None)
            self.assertIn(f"ns_op_median={expected:.1f} ", output.getvalue())

    def test_javascript_report(self):
        source = (ROOT / "bid754-codec-js/bench_runner.mjs").read_text()
        start = source.index("function benchRow(")
        end = source.index("\nconst n32 =", start)
        reporter = source[start:end]
        for samples, expected in [([1, 3], 2), ([9, 1, 5], 5), ([8], 8), ([8, 2, 4, 10], 6)]:
            program = "const SAMPLES=" + str(len(samples)) + ";" + \
                "const TARGET_SAMPLE_NS=1000n, CALIBRATION_ITERS=1;" + \
                "const timings=" + json.dumps([1000000] + [x * 1000 for x in samples]) + ";" + \
                "function runBatch(){return BigInt(timings.shift());}\n" + reporter + \
                "\nbenchRow('fixture', () => {});"
            result = subprocess.run(["node", "--input-type=module", "-e", program],
                                    capture_output=True, text=True, check=True)
            self.assertIn(f"ns_op_median={expected:.1f} ", result.stdout)


if __name__ == "__main__":
    unittest.main()
