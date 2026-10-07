from pathlib import Path
import re
import subprocess
import tempfile
import unittest


REPO = Path(__file__).resolve().parents[2]
SCRIPT = REPO / "devtools/scripts/verify_bidcodec_packages.sh"
GOOD_PACKAGE = """from enum import Enum
class Kind(Enum):
    NORMAL = 1
class Codec:
    kind = Kind.NORMAL
    coefficient = 1
def decode32(raw):
    return Codec()
def to_string(codec):
    return '+1E+0'
"""


class InstalledPythonSmokeTest(unittest.TestCase):
    def test_smoke_uses_installed_package_even_from_source_directory(self):
        script = SCRIPT.read_text()
        matches = re.findall(r'(?m)^  ("\$py_venv/bin/python"[^\n]*<<\'PY\'\n.*?\nPY)$', script, re.S)
        self.assertEqual(len(matches), 1, "expected one installed-wheel smoke in package gate")
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            source = root / "bid754-codec-py"
            package = source / "bid_codec"
            package.mkdir(parents=True)
            (package / "__init__.py").write_text(GOOD_PACKAGE)
            venv = root / "venv"
            subprocess.run(["python3", "-m", "venv", str(venv)], check=True)
            purelib = subprocess.check_output(
                [str(venv / "bin/python"), "-c", "import sysconfig; print(sysconfig.get_paths()['purelib'])"],
                text=True).strip()
            installed = Path(purelib) / "bid_codec"
            installed.mkdir()
            (installed / "__init__.py").write_text("raise ImportError('broken installed wheel')\n")

            def smoke():
                return subprocess.run(["bash", "-c", "py_venv=$1\n" + matches[0], "bash", str(venv)],
                                      cwd=source, capture_output=True, text=True)

            broken = smoke()
            self.assertNotEqual(broken.returncode, 0, broken.stdout + broken.stderr)
            self.assertIn("broken installed wheel", broken.stderr)
            (installed / "__init__.py").write_text(GOOD_PACKAGE)
            healthy = smoke()
            self.assertEqual(healthy.returncode, 0, healthy.stdout + healthy.stderr)


if __name__ == "__main__":
    unittest.main()
