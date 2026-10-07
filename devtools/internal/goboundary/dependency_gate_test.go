package goboundary

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type dependencyGateFixture struct {
	root     string
	makefile string
	env      []string
}

func newDependencyGateFixture(t *testing.T) dependencyGateFixture {
	t.Helper()
	makefile, err := filepath.Abs("../../../Makefile")
	if err != nil {
		t.Fatal(err)
	}
	f := dependencyGateFixture{
		root:     t.TempDir(),
		makefile: makefile,
		env: append(os.Environ(),
			"PATH="+filepath.Join(runtime.GOROOT(), "bin")+string(os.PathListSeparator)+os.Getenv("PATH"),
			"GOTOOLCHAIN=local", "GOWORK=off", "GOFLAGS=", "GOPROXY=off", "GOSUMDB=off", "CGO_ENABLED=1"),
	}
	f.write(t, "bid754-go/go.mod", "module "+modulePath+"\n\ngo 1.23\n")
	f.write(t, "bid754-go/probe.go", "package probe\nimport (\"net\"; \"os/user\"; \""+modulePath+"/helper\")\nvar _ = net.IPv4len\nvar _ = user.Current\nfunc Value() int { return helper.Value() }\n")
	f.write(t, "bid754-go/helper/helper.go", "package helper\nfunc Value() int { return 7 }\n")
	return f
}

func (f dependencyGateFixture) write(t *testing.T, name, source string) {
	t.Helper()
	path := filepath.Join(f.root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
}

func (f dependencyGateFixture) require(t *testing.T, path string) {
	t.Helper()
	f.write(t, "bid754-go/go.mod", "module "+modulePath+"\n\ngo 1.23\nrequire "+path+" v0.0.0\nreplace "+path+" => ../external\n")
	f.write(t, "external/go.mod", "module "+path+"\n\ngo 1.23\n")
	f.write(t, "external/value.go", "package external\nfunc Value() int { return 7 }\n")
}

func (f dependencyGateFixture) check(t *testing.T, target, rejectedPath string) {
	t.Helper()
	cmd := exec.Command("make", "-f", f.makefile, target, "GO_MODULES=bid754-go")
	cmd.Dir = f.root
	cmd.Env = f.env
	out, err := cmd.CombinedOutput()
	if rejectedPath != "" {
		if err == nil || !strings.Contains(string(out), "ERROR:") || !strings.Contains(string(out), rejectedPath) {
			t.Fatalf("%s must reject %s through its dependency check: %v\n%s", target, rejectedPath, err, out)
		}
		return
	}
	want := "bid754-go: stdlib-only"
	if target == "verify-portable-purity" {
		want = "bid754-go: default build is cgo-free"
	}
	if err != nil || !strings.Contains(string(out), want) {
		t.Fatalf("%s baseline: %v\n%s", target, err, out)
	}
}

func (f dependencyGateFixture) build(t *testing.T, cgo, wantFailure string) {
	t.Helper()
	cmd := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "./...")
	cmd.Dir = filepath.Join(f.root, "bid754-go")
	cmd.Env = append(append([]string{}, f.env...), "CGO_ENABLED="+cgo)
	out, err := cmd.CombinedOutput()
	if wantFailure != "" {
		if err == nil || !strings.Contains(string(out), wantFailure) {
			t.Fatalf("CGO_ENABLED=%s must fail with %q: %v\n%s", cgo, wantFailure, err, out)
		}
	} else if err != nil {
		t.Fatalf("CGO_ENABLED=%s build: %v\n%s", cgo, err, out)
	}
}

func TestZeroDependencyGateRejectsExternalModules(t *testing.T) {
	for _, tc := range []struct {
		name string
		path string
		use  string
	}{
		{"shared_prefix", modulePath + "-external", "runtime"},
		{"nested_module", modulePath + "/external", "runtime"},
		{"unrelated_module", "example.com/external", "runtime"},
		{"unused_require", "example.com/external", "unused"},
		{"test_only", modulePath + "-external", "test"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newDependencyGateFixture(t)
			f.check(t, "verify-zero-deps", "")
			f.require(t, tc.path)
			switch tc.use {
			case "runtime":
				f.write(t, "bid754-go/helper/helper.go", "package helper\nimport external \""+tc.path+"\"\nfunc Value() int { return external.Value() }\n")
			case "test":
				f.write(t, "bid754-go/probe_test.go", "package probe\nimport (\"testing\"; external \""+tc.path+"\")\nfunc TestValue(t *testing.T) { if external.Value() != 7 { t.Fatal(\"value\") } }\n")
			}
			f.build(t, "1", "")
			f.check(t, "verify-zero-deps", tc.path)
		})
	}
}

func TestPortablePurityGateRejectsReachableCgo(t *testing.T) {
	for _, location := range []string{"root", "own_dependency", "external_dependency", "external_with_pure_go"} {
		t.Run(location, func(t *testing.T) {
			f := newDependencyGateFixture(t)
			f.check(t, "verify-portable-purity", "")
			f.build(t, "0", "")
			path := modulePath
			switch location {
			case "root":
				f.write(t, "bid754-go/cgo.go", "package probe\nimport \"C\"\nvar cgoValue C.int\n")
			case "own_dependency":
				path += "/helper"
				f.write(t, "bid754-go/helper/cgo.go", "package helper\nimport \"C\"\nvar cgoValue C.int\n")
			default:
				path += "-external"
				f.require(t, path)
				f.write(t, "bid754-go/helper/helper.go", "package helper\nimport external \""+path+"\"\nfunc Value() int { return external.Value() }\n")
				f.write(t, "external/value.go", "package external\nimport \"C\"\nfunc Value() int { return int(C.int(7)) }\n")
				if location == "external_with_pure_go" {
					f.write(t, "external/pure.go", "//go:build !cgo\n\npackage external\nfunc Value() int { return 7 }\n")
				}
			}
			f.build(t, "1", "")
			if location == "external_dependency" {
				f.build(t, "0", "build constraints exclude all Go files")
			} else {
				f.build(t, "0", "")
			}
			f.check(t, "verify-portable-purity", path)
		})
	}
}

func TestPortablePurityGateRejectsReachableSwig(t *testing.T) {
	for _, extension := range []string{"swig", "swigcxx"} {
		for _, location := range []string{"root", "own_dependency"} {
			t.Run(extension+"_"+location, func(t *testing.T) {
				f := newDependencyGateFixture(t)
				f.check(t, "verify-portable-purity", "")
				path := modulePath
				dir := "bid754-go"
				if location == "own_dependency" {
					path += "/helper"
					dir += "/helper"
				}
				f.write(t, dir+"/probe."+extension, "%module probe\n")
				f.check(t, "verify-portable-purity", path)
			})
		}
	}
}

func TestDependencyGatesRejectInspectionFailure(t *testing.T) {
	for _, target := range []string{"verify-zero-deps", "verify-portable-purity"} {
		for _, fault := range []string{"go_list", "output_filter"} {
			t.Run(target+"/"+fault, func(t *testing.T) {
				f := newDependencyGateFixture(t)
				f.check(t, target, "")
				if fault == "go_list" {
					f.write(t, "bid754-go/go.mod", "invalid go.mod\n")
				} else {
					for _, filter := range []string{"grep", "awk", "sort"} {
						name := "bin/" + filter
						f.write(t, name, "#!/bin/sh\necho 'injected output filter failure' >&2\nexit 42\n")
						if err := os.Chmod(filepath.Join(f.root, name), 0700); err != nil {
							t.Fatal(err)
						}
					}
					f.env = append(f.env, "PATH="+filepath.Join(f.root, "bin")+string(os.PathListSeparator)+filepath.Join(runtime.GOROOT(), "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
				}
				cmd := exec.Command("make", "-f", f.makefile, target, "GO_MODULES=bid754-go")
				cmd.Dir = f.root
				cmd.Env = f.env
				out, err := cmd.CombinedOutput()
				if err == nil || strings.Contains(string(out), "✅ bid754-go:") {
					t.Fatalf("%s must fail on %s errors: %v\n%s", target, fault, err, out)
				}
				if fault == "output_filter" && !strings.Contains(string(out), "injected output filter failure") {
					t.Fatalf("%s did not reach the injected filter fault:\n%s", target, out)
				}
			})
		}
	}
}

func TestDependencyGatesAllowNativeTaggedCgo(t *testing.T) {
	f := newDependencyGateFixture(t)
	f.write(t, "bid754-go/native.go", "//go:build bid754_native\n\npackage probe\nimport \"C\"\nvar nativeValue C.int\n")
	f.env = append(f.env, "GOFLAGS=-tags=bid754_native", "GOWORK="+filepath.Join(f.root, "missing.work"))
	for _, target := range []string{"verify-zero-deps", "verify-portable-purity"} {
		f.check(t, target, "")
	}
}
