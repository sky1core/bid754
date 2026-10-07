import json
import os
from pathlib import Path
import shutil
import shlex
import subprocess
import tarfile
import tempfile
import unittest
import zipfile


REPO = Path(__file__).resolve().parents[2]
ARCHIVE = REPO / "devtools/third_party/intel_dfp/IntelRDFPMathLib20U4.tar.gz"
SCRIPT = REPO / "devtools/scripts/setup_generation_inputs.sh"


class GenerationInputProvenanceTest(unittest.TestCase):
    def test_dectest_rejects_failed_input_inspection(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            scripts = root / "devtools/scripts"
            inputs = root / "devtools/tests"
            scripts.mkdir(parents=True)
            inputs.mkdir(parents=True)
            shutil.copy2(SCRIPT, scripts / SCRIPT.name)
            archive = inputs / "dectest.zip"
            shutil.copy2(REPO / "devtools/tests/dectest.zip", archive)
            with zipfile.ZipFile(archive) as source:
                source.extractall(inputs)
            good = subprocess.run(["bash", str(scripts / SCRIPT.name), "dectest"],
                                  capture_output=True, text=True)
            self.assertEqual(good.returncode, 0, good.stdout + good.stderr)
            pinned_add = (inputs / "add.decTest").read_bytes()
            (inputs / "add.decTest").write_text("foreign source\n")
            tool_dir = root / "failed-find"
            tool_dir.mkdir()
            wrapper = tool_dir / "find"
            wrapper.write_text("#!/bin/sh\nexit 23\n")
            wrapper.chmod(0o700)
            failed = subprocess.run(["bash", str(scripts / SCRIPT.name), "dectest"],
                                    env={**os.environ, "PATH": str(tool_dir) + os.pathsep + os.environ["PATH"]},
                                    capture_output=True, text=True)
            self.assertNotEqual(failed.returncode, 0, failed.stdout + failed.stderr)
            self.assertIn("failed to enumerate pinned IBM", failed.stdout + failed.stderr)
            (inputs / "add.decTest").write_bytes(pinned_add)
            verified = subprocess.run(["bash", str(scripts / SCRIPT.name), "verify-dectest"],
                                      capture_output=True, text=True)
            self.assertEqual(verified.returncode, 0, verified.stdout + verified.stderr)
            partial_tools = root / "partial-basename"
            partial_tools.mkdir()
            basename = shutil.which("basename")
            self.assertIsNotNone(basename)
            wrapper = partial_tools / "basename"
            wrapper.write_text("#!/bin/sh\ncase \"$1\" in */add.decTest) exit 23;; esac\n"
                               "exec " + shlex.quote(basename) + " \"$@\"\n")
            wrapper.chmod(0o700)
            (inputs / "add.decTest").write_text("foreign source\n")
            partial = subprocess.run(["bash", str(scripts / SCRIPT.name), "verify-dectest"],
                                     env={**os.environ, "PATH": str(partial_tools) + os.pathsep + os.environ["PATH"]},
                                     capture_output=True, text=True)
            self.assertNotEqual(partial.returncode, 0, partial.stdout + partial.stderr)
            self.assertIn("differs from pinned 2.62 archive", partial.stdout + partial.stderr)
            (inputs / "add.decTest").write_bytes(pinned_add)
            for name, tool, code, message in (
                ("unpack", "unzip", "exit 23", "failed to unpack pinned IBM"),
                ("list", "find", "exit 23", "failed to enumerate pinned IBM"),
                ("empty list", "find", "exit 0", "failed to enumerate pinned IBM"),
                ("sort", "sort", "exit 23", "failed to enumerate pinned IBM"),
            ):
                with self.subTest(name=name):
                    tools = root / name
                    tools.mkdir()
                    wrapper = tools / tool
                    wrapper.write_text("#!/bin/sh\n" + code + "\n")
                    wrapper.chmod(0o700)
                    rejected = subprocess.run(["bash", str(scripts / SCRIPT.name), "verify-dectest"],
                                              env={**os.environ, "PATH": str(tools) + os.pathsep + os.environ["PATH"]},
                                              capture_output=True, text=True)
                    self.assertNotEqual(rejected.returncode, 0, rejected.stdout + rejected.stderr)
                    self.assertIn(message, rejected.stdout + rejected.stderr)
            (inputs / "add.decTest").write_text("foreign source\n")
            shutil.copy2(REPO / "Makefile", root / "Makefile")
            wrong_source = subprocess.run(["make", "test-portable-dectest"], cwd=root,
                                          capture_output=True, text=True)
            self.assertNotEqual(wrong_source.returncode, 0, wrong_source.stdout + wrong_source.stderr)
            self.assertIn("differs from pinned 2.62 archive", wrong_source.stdout + wrong_source.stderr)
            with archive.open("ab") as output:
                output.write(b"foreign")
            wrong_archive = subprocess.run(["bash", str(scripts / SCRIPT.name), "verify-dectest"],
                                           capture_output=True, text=True)
            self.assertNotEqual(wrong_archive.returncode, 0, wrong_archive.stdout + wrong_archive.stderr)
            self.assertIn("checksum mismatch", wrong_archive.stdout + wrong_archive.stderr)

    def test_existing_source_alias_must_reach_pinned_tree(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            scripts = root / "devtools/scripts"
            intel = root / "devtools/third_party/intel_dfp"
            scripts.mkdir(parents=True)
            intel.mkdir(parents=True)
            shutil.copy2(SCRIPT, scripts / SCRIPT.name)
            shutil.copy2(ARCHIVE, intel / ARCHIVE.name)
            with tarfile.open(intel / ARCHIVE.name) as source:
                source.extractall(intel, filter="data")

            result = subprocess.run(["bash", str(scripts / SCRIPT.name), "intel"],
                                    capture_output=True, text=True)
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assertTrue((intel / "src").samefile(intel / "LIBRARY/src"))

            (intel / "src").unlink()
            foreign = root / "foreign-src"
            foreign.mkdir()
            shutil.copy2(intel / "LIBRARY/src/bid_conf.h", foreign / "bid_conf.h")
            (foreign / "bid_conf.h").write_bytes(b"foreign source\n")
            (intel / "src").symlink_to(foreign, target_is_directory=True)
            result = subprocess.run(["bash", str(scripts / SCRIPT.name), "intel"],
                                    capture_output=True, text=True)
            self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assertIn("src", result.stdout + result.stderr)

    def test_read_only_native_provenance_checks_pin_source_and_build_stamp(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            scripts = root / "devtools/scripts"
            intel = root / "devtools/third_party/intel_dfp"
            scripts.mkdir(parents=True)
            intel.mkdir(parents=True)
            shutil.copy2(SCRIPT, scripts / SCRIPT.name)
            shutil.copy2(ARCHIVE, intel / ARCHIVE.name)
            with tarfile.open(intel / ARCHIVE.name) as source:
                source.extractall(intel, filter="data")
            (intel / "src").symlink_to("LIBRARY/src", target_is_directory=True)
            lib = intel / "lib"
            lib.mkdir()
            (lib / "libbid.a").write_bytes(b"fixture archive")
            stamp = lib / ".libbid.build-flags"
            aux = "-DBID_SIZE_LONG=8" if os.uname().machine in ("arm64", "aarch64") else ""
            stamp.write_text("CALL_BY_REF=0\nGLOBAL_RND=0\nGLOBAL_FLAGS=0\n"
                             "UNCHANGED_BINARY_FLAGS=0\nCFLAGS_AUX=" + aux + "\n"
                             "CFLAGS_OPT=-O3 -ffp-contract=off\n")

            def verify():
                return subprocess.run(["bash", str(scripts / SCRIPT.name), "verify-intel"],
                                      capture_output=True, text=True)

            good = verify()
            self.assertEqual(good.returncode, 0, good.stdout + good.stderr)
            for tool, message in (("tar", "unpack pinned Intel"), ("find", "enumerate pinned Intel")):
                with self.subTest(failing_tool=tool):
                    tool_dir = root / ("failed-" + tool)
                    tool_dir.mkdir()
                    wrapper = tool_dir / tool
                    wrapper.write_text("#!/bin/sh\nexit 23\n")
                    wrapper.chmod(0o700)
                    failed_tool = subprocess.run(["bash", str(scripts / SCRIPT.name), "verify-intel"],
                                                 env={**os.environ, "PATH": str(tool_dir) + os.pathsep + os.environ["PATH"]},
                                                 capture_output=True, text=True)
                    self.assertNotEqual(failed_tool.returncode, 0, failed_tool.stdout + failed_tool.stderr)
                    self.assertIn(message, failed_tool.stdout + failed_tool.stderr)
            pinned_stamp = stamp.read_text()
            stamp.write_text(pinned_stamp.replace("-O3 -ffp-contract=off", "-O3 -ffast-math"))
            overridden = subprocess.run(["bash", str(scripts / SCRIPT.name), "verify-intel"],
                                        env={**os.environ, "INTEL_DFP_OPT_CFLAGS": "-O3 -ffast-math"},
                                        capture_output=True, text=True)
            self.assertNotEqual(overridden.returncode, 0, overridden.stdout + overridden.stderr)
            self.assertIn("pinned", overridden.stdout + overridden.stderr)
            stamp.write_text(pinned_stamp)
            shutil.copy2(REPO / "Makefile", root / "Makefile")
            direct_good = subprocess.run(["make", "verify-native-inputs"], cwd=root,
                                         capture_output=True, text=True)
            self.assertEqual(direct_good.returncode, 0, direct_good.stdout + direct_good.stderr)
            stamp.unlink()
            plan = json.loads((REPO / "devtools/verification_plan.json").read_text())
            targets = {gate["target"] for gate in plan["gates"]
                       if "native" in gate.get("prerequisites", [])}
            targets.update({"bench-native", "bench-quick", "explore-fresh-seed"})
            for target in sorted(targets):
                with self.subTest(target=target):
                    missing = subprocess.run(["make", target], cwd=root,
                                             capture_output=True, text=True)
                    self.assertNotEqual(missing.returncode, 0, missing.stdout + missing.stderr)
                    self.assertIn("native library or build stamp missing", missing.stdout + missing.stderr)
            stamp.write_text(pinned_stamp)
            stamp.write_text("CALL_BY_REF=1\n")
            bad_stamp = verify()
            self.assertNotEqual(bad_stamp.returncode, 0, bad_stamp.stdout + bad_stamp.stderr)
            self.assertIn("build", bad_stamp.stdout + bad_stamp.stderr)
            stamp.write_text("CALL_BY_REF=0\nGLOBAL_RND=0\nGLOBAL_FLAGS=0\n"
                             "UNCHANGED_BINARY_FLAGS=0\nCFLAGS_AUX=" + aux + "\n"
                             "CFLAGS_OPT=-O3 -ffp-contract=off\n")
            with (intel / ARCHIVE.name).open("ab") as output:
                output.write(b"foreign")
            bad_archive = verify()
            self.assertNotEqual(bad_archive.returncode, 0, bad_archive.stdout + bad_archive.stderr)
            self.assertIn("checksum", bad_archive.stdout + bad_archive.stderr)


if __name__ == "__main__":
    unittest.main()
