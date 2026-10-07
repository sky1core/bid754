import contextlib
import importlib.util
import io
import json
from pathlib import Path
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
SPEC = importlib.util.spec_from_file_location(
    "numeric_regressions", ROOT / "devtools/scripts/generate_numeric_regressions.py")
GENERATOR = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(GENERATOR)


class NumericRegressionCheckTest(unittest.TestCase):
    def test_parser_carry_oracle_preserves_tininess(self):
        rows = GENERATOR.parser_carry_rows()
        self.assertEqual(len(rows), 2400)
        indexed = {(fn, mode, args[0]): (bits, flags)
                   for fn, mode, args, bits, flags in rows}
        self.assertEqual(len(indexed), 2400)
        for width, literal, bits in [(32, "999999901e-104", 1000000),
                                     (64, "999999999999999901e-401", 1000000000000000)]:
            for sign, mode in [("", 2), ("-", 1)]:
                want = bits | (int(bool(sign)) << (width - 1))
                key = (f"bid{width}_from_string", mode, sign + literal)
                self.assertEqual(indexed[key], (f"{want:0{width // 4}x}", 0x30))

    def test_quantum_corpus_covers_exponents_layouts_signs_and_neighbors(self):
        rows = [row for row in GENERATOR.make_rows() if row[0] == "bid32_quantum"]
        self.assertEqual(len(rows), 896)
        finite = [row for row in rows if int(row[2][0], 16) & 0x78000000 != 0x78000000]
        observed = {(biased, sign, layout) for _, _, args, _, _ in finite
                    for bits in [int(args[0], 16)]
                    for sign in [(bits >> 31) & 1]
                    for layout in [bool(bits & 0x60000000 == 0x60000000)]
                    for biased in [((bits >> 21) if layout else (bits >> 23)) & 0xff]}
        self.assertEqual(observed, {(exp, sign, layout)
                                    for exp in range(192) for sign in (0, 1)
                                    for layout in (False, True)})
        self.assertEqual(len(rows) - len(finite), 8)
        for row in rows:
            GENERATOR.validate_quantum_row(row)

    def test_quantum_generation_rejects_malformed_shape_and_observation(self):
        row = next(row for row in GENERATOR.make_rows() if row[0] == "bid32_quantum")
        for bad in ((row[0], 1, row[2], row[3], row[4]),
                    (row[0], row[1], ("6000000",), row[3], row[4]),
                    (row[0], row[1], row[2], "ffffffff", row[4]),
                    (row[0], row[1], row[2], row[3], 0x20)):
            with self.assertRaisesRegex(ValueError, "malformed bid32_quantum"):
                GENERATOR.validate_quantum_row(bad)
            with self.assertRaisesRegex(ValueError, "malformed bid32_quantum"):
                GENERATOR.emit([bad], {}, check=True)
        with self.assertRaisesRegex(ValueError, "malformed pinned-C observation"):
            GENERATOR.emit([row], {GENERATOR.key(row): ["bogus", "00"]}, check=True)

    def test_check_rejects_changed_result_and_skip_without_repair(self):
        rows = GENERATOR.make_rows()
        observed = json.loads(GENERATOR.OBS.read_text())["rows"]
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            data = root / "devtools/testdata"
            data.mkdir(parents=True)
            manifest = root / "devtools/testgen_manifest.json"
            manifest.write_bytes((ROOT / "devtools/testgen_manifest.json").read_bytes())
            old_root, old_data = GENERATOR.ROOT, GENERATOR.DATA
            GENERATOR.ROOT, GENERATOR.DATA = root, data
            try:
                with contextlib.redirect_stdout(io.StringIO()):
                    GENERATOR.emit(rows, observed)
                    GENERATOR.emit(rows, observed, check=True)
                fixed = data / "readtest_numeric_boundary_bid32_from_string_fixed_point_underflow_cdiverge.in"
                for literal, mode, bits in [("99999980e-104", 1, "0001869f"),
                                            ("14999998e-108", 4, "00000001"),
                                            ("99999995e-103", 4, "000f4240")]:
                    spelling = format(GENERATOR.decimal.Decimal(literal), "f")
                    self.assertIn(f"bid32_from_string {mode} {spelling} [{bits}] 30\n",
                                  fixed.read_text())
                source = next(data.glob("*cdiverge.in"))
                original = source.read_text()
                lines = original.splitlines()
                fields = lines[1].split()
                fields[-1] = "7f"
                lines[1] = " ".join(fields)
                damaged = "\n".join(lines) + "\n"
                source.write_text(damaged)
                with self.assertRaisesRegex(ValueError, "artifact differs"):
                    GENERATOR.emit(rows, observed, check=True)
                self.assertEqual(source.read_text(), damaged)
                source.write_text(original)
                content = json.loads(manifest.read_text())
                block = next(b for b in content["readtests"]
                             if b["name"].startswith("numeric_boundary_")
                             and "native_compare_skip_reason" in b)
                del block["native_compare_skip_reason"]
                damaged = json.dumps(content, indent=2) + "\n"
                manifest.write_text(damaged)
                with self.assertRaisesRegex(ValueError, "artifact differs"):
                    GENERATOR.emit(rows, observed, check=True)
                self.assertEqual(manifest.read_text(), damaged)
            finally:
                GENERATOR.ROOT, GENERATOR.DATA = old_root, old_data


if __name__ == "__main__":
    unittest.main()
