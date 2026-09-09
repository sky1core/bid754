package testgen

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type oracleDependencyPackage struct {
	ImportPath string
	Dir        string
	Module     *struct{ Path string }
}

func oracleProductionDependencies(dir, target string, forbiddenDirs []string) error {
	cmd := exec.Command("go", "list", "-deps", "-json", target)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("production dependency graph: %w", err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(out)))
	packages := map[string]oracleDependencyPackage{}
	for {
		var pkg oracleDependencyPackage
		err := decoder.Decode(&pkg)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		packages[pkg.ImportPath] = pkg
	}
	for _, pkg := range packages {
		for _, prefix := range []string{"github.com/sky1core/bid754/bid754-codec-go", "github.com/sky1core/bid754/bid754-go"} {
			if pkg.ImportPath == prefix || strings.HasPrefix(pkg.ImportPath, prefix+"/") || (pkg.Module != nil && pkg.Module.Path == prefix) {
				return fmt.Errorf("production oracle dependency reaches %s", pkg.ImportPath)
			}
		}
		for _, root := range forbiddenDirs {
			actual, err := filepath.EvalSymlinks(pkg.Dir)
			if err != nil {
				return err
			}
			canonical, err := filepath.EvalSymlinks(root)
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(canonical, actual)
			if err != nil {
				return err
			}
			if rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
				return fmt.Errorf("production oracle dependency reaches directory %s through %s", canonical, pkg.ImportPath)
			}
		}
	}
	if len(packages) == 0 {
		return fmt.Errorf("empty production dependency graph")
	}
	return nil
}

func TestBidCodecVectorGeneratorDoesNotImportBidCodecUnderTest(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if err := oracleProductionDependencies(root, "./internal/testgen", []string{filepath.Join(root, "../bid754-codec-go"), filepath.Join(root, "../bid754-go")}); err != nil {
		t.Fatal(err)
	}
}

func TestOracleProductionDependencyGraph(t *testing.T) {
	for _, tc := range []struct {
		name, production, testSource, helper, modulePath string
		reject                                           bool
	}{
		{name: "baseline", production: "package generator\nimport _ \"fmt\"\n"},
		{name: "renamed_direct", production: "package generator\nimport alias \"github.com/sky1core/bid754/bid754-codec-go\"\nvar _ = alias.Value\n", modulePath: "github.com/sky1core/bid754/bid754-codec-go", reject: true},
		{name: "indirect", production: "package generator\nimport _ \"fixture/helper\"\n", helper: "package helper\nimport _ \"github.com/sky1core/bid754/bid754-codec-go\"\n", modulePath: "github.com/sky1core/bid754/bid754-codec-go", reject: true},
		{name: "public_runtime", production: "package generator\nimport _ \"github.com/sky1core/bid754/bid754-go\"\n", modulePath: "github.com/sky1core/bid754/bid754-go", reject: true},
		{name: "module_alias", production: "package generator\nimport _ \"alternate.invalid/decimal\"\n", modulePath: "alternate.invalid/decimal", reject: true},
		{name: "test_only", production: "package generator\n", testSource: "package generator\nimport _ \"github.com/sky1core/bid754/bid754-codec-go\"\n", modulePath: "github.com/sky1core/bid754/bid754-codec-go"},
		{name: "explicit_process_tooling", production: "package generator\nimport \"os/exec\"\nvar _ = exec.Command\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			codec := filepath.Join(root, "production-codec")
			if err := os.MkdirAll(codec, 0755); err != nil {
				t.Fatal(err)
			}
			module := "module fixture\n\ngo 1.23\n"
			if tc.modulePath != "" {
				module += fmt.Sprintf("require %s v0.0.0\nreplace %s => ./production-codec\n", tc.modulePath, tc.modulePath)
			}
			files := map[string]string{"go.mod": module, "renamed.go": tc.production, "production-codec/value.go": "package codec\nconst Value=1\n", "production-codec/go.mod": "module " + tc.modulePath + "\n\ngo 1.23\n"}
			if tc.testSource != "" {
				files["dependency_test.go"] = tc.testSource
			}
			if tc.helper != "" {
				files["helper/renamed.go"] = tc.helper
			}
			for name, body := range files {
				path := filepath.Join(root, name)
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
			}
			err := oracleProductionDependencies(root, ".", []string{codec})
			if tc.reject {
				if err == nil || !strings.Contains(err.Error(), "production oracle dependency reaches") {
					t.Fatalf("want forbidden dependency, got %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}
