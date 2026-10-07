package scriptcheck

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type policyFixture struct {
	root, makefile string
	environment    []string
}

func newPolicyFixture(t *testing.T) policyFixture {
	t.Helper()
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	f := policyFixture{root: t.TempDir(), makefile: filepath.Join(repo, "Makefile")}
	for _, name := range []string{"verify_rust_package.sh", "verify_rust_overflow_policy.sh", "check_generated_marker_coverage.sh", "generated_artifacts.py"} {
		data, err := os.ReadFile(filepath.Join(repo, "devtools/scripts", name))
		if err != nil {
			t.Fatal(err)
		}
		f.write(t, "devtools/scripts/"+name, string(data))
	}
	return f
}

func (f policyFixture) write(t *testing.T, name, data string) {
	t.Helper()
	path := filepath.Join(f.root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
}

func (f policyFixture) run(t *testing.T, directory, reject string, args ...string) string {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = filepath.Join(f.root, directory)
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "CARGO_ENCODED_RUSTFLAGS=") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	cmd.Env = append(cmd.Env, "RUSTFLAGS=", "CARGO_TARGET_DIR="+filepath.Join(f.root, "target"))
	cmd.Env = append(cmd.Env, f.environment...)
	out, err := cmd.CombinedOutput()
	if reject == "" {
		if err != nil {
			t.Fatalf("%v baseline: %v\n%s", args, err, out)
		}
	} else if err == nil || !strings.Contains(string(out), reject) {
		t.Fatalf("%v must fail with %q: %v\n%s", args, reject, err, out)
	}
	return string(out)
}

func (f policyFixture) gate(t *testing.T, target, reject string) {
	t.Helper()
	f.run(t, "", reject, "make", "-f", f.makefile, target)
}

const policyCargoManifest = `[package]
name = "bid754"
version = "0.0.0"
edition = "2021"
publish = false
include = ["src/**", "Cargo.toml", "LICENSE", "README.md"]
[features]
verification = []
[dependencies]
num-bigint = "=0.4.6"
num-traits = "=0.2.19"
`

func (f policyFixture) rust(t *testing.T) {
	t.Helper()
	f.write(t, "bid754-rs/Cargo.toml", policyCargoManifest)
	f.write(t, "bid754-rs/LICENSE", "test fixture\n")
	f.write(t, "bid754-rs/README.md", "test fixture\n")
	f.write(t, "bid754-rs/src/lib.rs", "pub fn value() -> i32 { 7 }\n")
	f.write(t, "bid754-rs/src/generated/api/mod.rs", "#![deny(arithmetic_overflow, overflowing_literals)]\n")
	f.run(t, "bid754-rs", "", "cargo", "generate-lockfile")
}

func TestRustPackageGateRejectsForeignCode(t *testing.T) {
	f := newPolicyFixture(t)
	f.rust(t)
	baseline := "pub const TEXT: &str = \"extern \\\"C\\\" { }\";\nextern crate num_traits;\n"
	f.write(t, "bid754-rs/src/lib.rs", baseline)
	f.write(t, "bid754-rs/examples/native_probe.rs", "unsafe extern \"C\" { fn abs(value: i32) -> i32; }\nfn main() {}\n")
	f.gate(t, "verify-rust-package", "")
	for _, tc := range []struct{ name, source string }{
		{"single_line", "unsafe extern \"C\" { fn abs(value: i32) -> i32; }"},
		{"multiline", "unsafe extern\n\"C\"\n{ fn abs(value: i32) -> i32; }"},
		{"comments", "unsafe /* comment */ extern /* comment */ \"C\" { fn abs(value: i32) -> i32; }"},
		{"macro", "macro_rules! foreign { () => { unsafe extern \"C\" { fn abs(value: i32) -> i32; } }; } foreign!();"},
		{"include", "include!(\"foreign.inc\");"},
		{"allow_override", "#[allow(unsafe_code)] unsafe extern \"C\" { fn abs(value: i32) -> i32; }"},
		{"verification_feature", "#[cfg(feature = \"verification\")] unsafe extern \"C\" { fn abs(value: i32) -> i32; }"},
		{"lint_cap_override", "unsafe extern \"C\" { fn abs(value: i32) -> i32; }"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := f
			if tc.name == "lint_cap_override" {
				f.environment = []string{"CARGO_ENCODED_RUSTFLAGS=--cap-lints=allow"}
			}
			f.write(t, "bid754-rs/src/lib.rs", baseline+tc.source+"\n")
			f.write(t, "bid754-rs/src/foreign.inc", "unsafe extern \"C\" { fn abs(value: i32) -> i32; }\n")
			f.run(t, "bid754-rs", "", "cargo", "check", "--locked", "--lib", "--features", "verification")
			f.gate(t, "verify-rust-package", "unsafe-code")
		})
	}
	for _, tc := range []struct{ name, path, kind, target, manifest string }{
		{"auto_main", "src/main.rs", "bin", "bid754", ""},
		{"auto_bin", "src/bin/probe.rs", "bin", "probe", ""},
		{"nested_bin", "src/bin/probe/main.rs", "bin", "probe", ""},
		{"feature_gated_bin", "src/bin/probe.rs", "bin", "probe", ""},
		{"explicit_bin", "src/probe.rs", "bin", "probe", "[[bin]]\nname = \"probe\"\npath = \"src/probe.rs\"\n"},
		{"required_feature_bin", "src/probe.rs", "bin", "probe", "[[bin]]\nname = \"probe\"\npath = \"src/probe.rs\"\nrequired-features = [\"verification\"]\n"},
		{"packaged_example", "src/probe.rs", "example", "probe", "[[example]]\nname = \"probe\"\npath = \"src/probe.rs\"\n"},
		{"packaged_test", "src/probe.rs", "test", "probe", "[[test]]\nname = \"probe\"\npath = \"src/probe.rs\"\nharness = false\n"},
		{"packaged_bench", "src/probe.rs", "bench", "probe", "[[bench]]\nname = \"probe\"\npath = \"src/probe.rs\"\nharness = false\n"},
		{"bin_lint_cap_override", "src/bin/probe.rs", "bin", "probe", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newPolicyFixture(t)
			f.rust(t)
			f.write(t, "bid754-rs/Cargo.toml", policyCargoManifest+"\n"+tc.manifest)
			f.write(t, "bid754-rs/src/lib.rs", "#![forbid(unsafe_code)]\npub fn value() -> i32 { 7 }\n")
			f.write(t, "bid754-rs/"+tc.path, "fn main() { assert_eq!(bid754::value(), 7); }\n")
			if tc.name == "bin_lint_cap_override" {
				f.environment = []string{"CARGO_ENCODED_RUSTFLAGS=--cap-lints=allow"}
			}
			f.gate(t, "verify-rust-package", "")
			packaged := f.run(t, "bid754-rs", "", "cargo", "package", "--list", "--locked", "--allow-dirty")
			if !strings.Contains("\n"+packaged, "\n"+tc.path+"\n") {
				t.Fatalf("fixture target %s is not packaged:\n%s", tc.path, packaged)
			}
			source := "unsafe extern \"C\" { fn abs(value: i32) -> i32; }\nfn main() { assert_eq!(unsafe { abs(-7) }, 7); }\n"
			if tc.name == "feature_gated_bin" {
				source = "#[cfg(feature = \"verification\")]\nunsafe extern \"C\" { fn abs(value: i32) -> i32; }\nfn main() { #[cfg(feature = \"verification\")] assert_eq!(unsafe { abs(-7) }, 7); }\n"
			}
			f.write(t, "bid754-rs/"+tc.path, source)
			f.run(t, "bid754-rs", "", "cargo", "check", "--locked", "--all-features", "--"+tc.kind, tc.target)
			f.gate(t, "verify-rust-package", "unsafe-code")
		})
	}
}

func TestRustOverflowGateParsesProfiles(t *testing.T) {
	f := newPolicyFixture(t)
	f.rust(t)
	f.write(t, "bid754-rs/src/lib.rs", "#![allow(\n arithmetic_overflow,\n overflowing_literals\n)]\npub fn value() -> i32 { 7 }\n")
	for i := 0; i < 104; i++ {
		name := string(rune('a'+i/26)) + string(rune('a'+i%26)) + ".rs"
		if i == 0 {
			name = "bid64_from_string.rs"
		}
		f.write(t, filepath.Join("bid754-rs/src/generated", name), "")
	}
	f.gate(t, "verify-rust-overflow", "")
	f.write(t, "bid754-rs/Cargo.toml", policyCargoManifest+"\n# overflow-checks = false\n[profile.dev]\noverflow-checks=true\n")
	f.gate(t, "verify-rust-overflow", "")
	for _, profile := range []string{
		"[profile.dev]\noverflow-checks=false\n",
		"[profile.dev]\n\"overflow-checks\"\t=\tfalse\n",
		"[profile]\ndev = { overflow-checks = false }\n",
		"[profile.test.build-override]\noverflow-checks = false\n",
		"[profile.release.package.num-bigint]\noverflow-checks = false\n",
	} {
		f.write(t, "bid754-rs/Cargo.toml", policyCargoManifest+"\n"+profile)
		f.gate(t, "verify-rust-overflow", "must not disable Rust overflow checks")
	}
	f.write(t, "bid754-rs/Cargo.toml", policyCargoManifest+"\n[profile.dev\n")
	f.gate(t, "verify-rust-overflow", "TOMLDecodeError")
}

func TestRustPackageGateRejectsDependencyOverrides(t *testing.T) {
	for _, location := range []string{"manifest_patch", "config_patch", "replace", "transitive_patch", "directory_source", "directory_source_in_cache", "archive_checksum"} {
		t.Run(location, func(t *testing.T) {
			f := newPolicyFixture(t)
			if location == "directory_source_in_cache" || location == "archive_checksum" {
				f.environment = []string{"CARGO_HOME=" + filepath.Join(f.root, "cargo-home")}
			}
			f.rust(t)
			f.gate(t, "verify-rust-package", "")
			f.write(t, "native-bigint/Cargo.toml", "[package]\nname = \"num-bigint\"\nversion = \"0.4.6\"\nedition = \"2021\"\n")
			f.write(t, "native-bigint/src/lib.rs", "unsafe extern \"C\" { fn abs(value: i32) -> i32; }\npub fn native_value() -> i32 { unsafe { abs(-7) } }\n")
			patch := "[patch.crates-io]\nnum-bigint = { path = \"../native-bigint\" }\n"
			switch location {
			case "manifest_patch":
				f.write(t, "bid754-rs/Cargo.toml", policyCargoManifest+"\n"+patch)
			case "config_patch":
				f.write(t, "bid754-rs/.cargo/config.toml", patch)
			case "replace":
				f.write(t, "bid754-rs/Cargo.toml", policyCargoManifest+"\n[replace]\n\"num-bigint:0.4.6\" = { path = \"../native-bigint\" }\n")
			case "archive_checksum":
				archives, err := filepath.Glob(filepath.Join(f.root, "cargo-home/registry/cache/*/num-bigint-0.4.6.crate"))
				if err != nil || len(archives) != 1 {
					t.Fatalf("expected one cached archive: %v, %v", archives, err)
				}
				data, err := os.ReadFile(archives[0])
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(archives[0], append(data, 0), 0600); err != nil {
					t.Fatal(err)
				}
			case "directory_source", "directory_source_in_cache":
				directory := "local-sources"
				if location == "directory_source_in_cache" {
					directory = "cargo-home/registry/src/directory-override"
				}
				var metadata struct {
					Packages []struct {
						Name         string `json:"name"`
						Version      string `json:"version"`
						Source       string `json:"source"`
						ManifestPath string `json:"manifest_path"`
					} `json:"packages"`
				}
				data := f.run(t, "bid754-rs", "", "cargo", "metadata", "--locked", "--format-version", "1")
				if err := json.Unmarshal([]byte(data), &metadata); err != nil {
					t.Fatal(err)
				}
				f.run(t, "bid754-rs", "", "cargo", "vendor", "--locked", "--versioned-dirs", filepath.Join(f.root, "vendor"))
				for _, dependency := range metadata.Packages {
					if dependency.Source == "" {
						continue
					}
					stem := dependency.Name + "-" + dependency.Version
					if err := os.CopyFS(filepath.Join(f.root, directory, stem), os.DirFS(filepath.Dir(dependency.ManifestPath))); err != nil {
						t.Fatal(err)
					}
					checksum, err := os.ReadFile(filepath.Join(f.root, "vendor", stem, ".cargo-checksum.json"))
					if err != nil {
						t.Fatal(err)
					}
					f.write(t, filepath.Join(directory, stem, ".cargo-checksum.json"), string(checksum))
				}
				f.write(t, "bid754-rs/.cargo/config.toml", fmt.Sprintf("[source.crates-io]\nreplace-with = \"local\"\n[source.local]\ndirectory = %q\n", filepath.Join(f.root, directory)))
				f.gate(t, "verify-rust-package", "")
				name := directory + "/num-bigint-0.4.6/src/lib.rs"
				source, err := os.ReadFile(filepath.Join(f.root, name))
				if err != nil {
					t.Fatal(err)
				}
				source = append(source, []byte("\nunsafe extern \"C\" { fn abs(value: i32) -> i32; }\npub fn native_value() -> i32 { unsafe { abs(-7) } }\n")...)
				f.write(t, name, string(source))
				checksumPath := directory + "/num-bigint-0.4.6/.cargo-checksum.json"
				encoded, err := os.ReadFile(filepath.Join(f.root, checksumPath))
				if err != nil {
					t.Fatal(err)
				}
				var checksum map[string]any
				if err := json.Unmarshal(encoded, &checksum); err != nil {
					t.Fatal(err)
				}
				checksum["files"].(map[string]any)["src/lib.rs"] = fmt.Sprintf("%x", sha256.Sum256(source))
				encoded, err = json.Marshal(checksum)
				if err != nil {
					t.Fatal(err)
				}
				f.write(t, checksumPath, string(encoded))
				f.environment = append(f.environment, "CARGO_TARGET_DIR="+filepath.Join(f.root, "mutated-target"))
			case "transitive_patch":
				var metadata struct {
					Packages []struct {
						Name         string `json:"name"`
						ManifestPath string `json:"manifest_path"`
					} `json:"packages"`
				}
				data := f.run(t, "bid754-rs", "", "cargo", "metadata", "--locked", "--format-version", "1")
				if err := json.Unmarshal([]byte(data), &metadata); err != nil {
					t.Fatal(err)
				}
				found := false
				for _, dependency := range metadata.Packages {
					if dependency.Name == "num-integer" {
						if err := os.CopyFS(filepath.Join(f.root, "local-integer"), os.DirFS(filepath.Dir(dependency.ManifestPath))); err != nil {
							t.Fatal(err)
						}
						found = true
					}
				}
				if !found {
					t.Fatal("fixture must resolve num-integer transitively")
				}
				f.write(t, "bid754-rs/Cargo.toml", policyCargoManifest+"\n[patch.crates-io]\nnum-integer = { path = \"../local-integer\" }\n")
			}
			if location != "transitive_patch" && location != "archive_checksum" {
				f.write(t, "bid754-rs/src/lib.rs", "#![forbid(unsafe_code)]\npub fn value() -> i32 { num_bigint::native_value() }\n#[test] fn invokes_native() { assert_eq!(value(), 7); }\n")
			}
			f.run(t, "bid754-rs", "", "cargo", "generate-lockfile")
			f.run(t, "bid754-rs", "", "cargo", "test", "--locked", "--lib")
			f.gate(t, "verify-rust-package", "resolved dependency")
		})
	}
}

func (f policyFixture) artifacts(t *testing.T) {
	t.Helper()
	manifest, err := json.Marshal(map[string][]string{"files": {"generated.go"}, "directories": {"generated"}})
	if err != nil {
		t.Fatal(err)
	}
	f.write(t, "devtools/generated_artifacts.json", string(manifest))
	for _, prefix := range []string{"", "backup/"} {
		f.write(t, prefix+"generated.go", "// Code generated by fixture; DO NOT EDIT.\n")
		f.write(t, prefix+"generated/value.json", "[1]\n")
	}
}

func TestGeneratedArtifactComparatorRejectsDrift(t *testing.T) {
	f := newPolicyFixture(t)
	f.artifacts(t)
	compare := func(reject string) {
		t.Helper()
		f.run(t, "", reject, "python3", "-B", "devtools/scripts/generated_artifacts.py", "compare", "backup")
	}
	compare("")
	for _, name := range []string{"generated.go", "generated/value.json", "generated/extra.json"} {
		f.artifacts(t)
		f.write(t, name, "changed\n")
		compare("generated artifacts differ:")
	}
	if err := os.Remove(filepath.Join(f.root, "generated/extra.json")); err != nil {
		t.Fatal(err)
	}
	f.artifacts(t)
	if err := os.Remove(filepath.Join(f.root, "backup/generated.go")); err != nil {
		t.Fatal(err)
	}
	compare("missing artifact file:")
	f.artifacts(t)
	f.run(t, "", "backup must differ", "python3", "-B", "devtools/scripts/generated_artifacts.py", "compare", ".")
}

func TestGeneratedMarkerCoverageRejectsUnregisteredComparisons(t *testing.T) {
	f := newPolicyFixture(t)
	f.artifacts(t)
	f.write(t, ".gitignore", "backup/\n")
	f.write(t, "devtools/scripts/generated_marker_exceptions.txt", "")
	f.run(t, "", "", "git", "init", "--quiet")
	f.gate(t, "check-generated-markers", "")
	f.write(t, "unregistered.go", "// Code generated by fixture; DO NOT EDIT.\n")
	f.write(t, "Makefile", "verify-generated:\n\t@cmp -s unregistered.go $$tmpdir/backup/unregistered.go || true\n")
	f.gate(t, "check-generated-markers", "unregistered.go")
}

func TestGeneratedArtifactManifestRejectsUnsafePaths(t *testing.T) {
	f := newPolicyFixture(t)
	f.artifacts(t)
	for _, manifest := range []string{
		`{"files": [], "directories": []}`,
		`{"files": ["../outside"], "directories": []}`,
		`{"files": ["generated*"], "directories": []}`,
		`{"files": ["-option"], "directories": []}`,
		`{"files": [], "directories": ["."]}`,
		`{"files": ["generated.go", "generated.go"], "directories": []}`,
		`{"files": ["generated/value.json"], "directories": ["generated"]}`,
		`{"files": ["generated.go"], "files": [], "directories": []}`,
	} {
		f.write(t, "devtools/generated_artifacts.json", manifest)
		f.run(t, "", "ERROR:", "python3", "-B", "devtools/scripts/generated_artifacts.py", "list", "all")
	}
	f.artifacts(t)
	if err := os.Symlink(filepath.Join(f.root, "generated.go"), filepath.Join(f.root, "generated/link.go")); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"files", "directories", "all"} {
		f.run(t, "", "artifact must not be a symlink:", "python3", "-B", "devtools/scripts/generated_artifacts.py", "list", kind)
	}
}

func TestGeneratedMarkerCoverageRejectsInspectionFailures(t *testing.T) {
	for _, fault := range []string{"git_error_after_output", "git_diagnostic_with_success", "unreadable_exceptions", "temporary_file_creation_denied", "unreadable_unregistered"} {
		t.Run(fault, func(t *testing.T) {
			if (fault == "temporary_file_creation_denied" || fault == "unreadable_unregistered") && runtime.GOOS != "darwin" {
				t.Skip("macOS Bash temporary-file failure requires sandbox-exec")
			}
			f := newPolicyFixture(t)
			f.artifacts(t)
			f.write(t, ".gitignore", "backup/\nbin/\n")
			f.write(t, "devtools/scripts/generated_marker_exceptions.txt", "")
			f.run(t, "", "", "git", "init", "--quiet")
			f.gate(t, "check-generated-markers", "")
			switch fault {
			case "git_error_after_output", "git_diagnostic_with_success":
				realGit, err := exec.LookPath("git")
				if err != nil {
					t.Fatal(err)
				}
				status := 42
				if fault == "git_diagnostic_with_success" {
					status = 0
				}
				f.write(t, "bin/git", fmt.Sprintf("#!/usr/bin/env python3\nimport os, subprocess, sys\nresult = subprocess.run([os.environ['POLICY_REAL_GIT'], *sys.argv[1:]])\nif sys.argv[1] == 'grep':\n print('injected read error', file=sys.stderr)\n sys.exit(%d)\nsys.exit(result.returncode)\n", status))
				if err := os.Chmod(filepath.Join(f.root, "bin/git"), 0700); err != nil {
					t.Fatal(err)
				}
				f.environment = []string{"POLICY_REAL_GIT=" + realGit, "PATH=" + filepath.Join(f.root, "bin") + string(os.PathListSeparator) + os.Getenv("PATH")}
				f.gate(t, "check-generated-markers", "ERROR:")
			case "unreadable_exceptions":
				name := filepath.Join(f.root, "devtools/scripts/generated_marker_exceptions.txt")
				if err := os.Rename(name, name+".saved"); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(name, 0700); err != nil {
					t.Fatal(err)
				}
				f.gate(t, "check-generated-markers", "ERROR:")
			case "unreadable_unregistered":
				f.write(t, "unregistered.go", "// Code generated by fixture; DO NOT EDIT.\n")
				blocked, err := filepath.EvalSymlinks(filepath.Join(f.root, "unregistered.go"))
				if err != nil {
					t.Fatal(err)
				}
				profile := fmt.Sprintf("(version 1) (allow default) (deny file-read-data (literal %q))", blocked)
				f.run(t, "", "unregistered.go", "/usr/bin/sandbox-exec", "-p", profile, "make", "-f", f.makefile, "check-generated-markers")
			case "temporary_file_creation_denied":
				command := []string{"/usr/bin/sandbox-exec", "-p", "(version 1) (allow default) (deny file-write-create)", "make", "-f", f.makefile, "check-generated-markers"}
				f.run(t, "", "", command...)
				f.write(t, "unregistered.go", "// Code generated by fixture; DO NOT EDIT.\n")
				f.run(t, "", "unregistered.go", command...)
			}
		})
	}
}
