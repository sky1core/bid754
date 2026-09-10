package goboundary

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func checkReferenceDependencies(packages []listedPackage) error {
	const model = modulePath + "/internal/decimalref"
	found := false
	for _, p := range packages {
		if p.ImportPath == model {
			found = true
			continue
		}
		if !p.Standard {
			return fmt.Errorf("exact reference imports non-standard dependency: %s", p.ImportPath)
		}
	}
	if !found {
		return fmt.Errorf("exact reference absent from dependency graph")
	}
	return nil
}

func TestExactReferenceDependencyBoundary(t *testing.T) {
	source, err := filepath.Abs("../../../bid754-go")
	if err != nil {
		t.Fatal(err)
	}
	module := t.TempDir()
	copyGoModuleSources(t, source, module)
	check := func(phase string) error {
		t.Helper()
		if out, err := runGo(module, "0", "build", "./internal/decimalref"); err != nil {
			t.Fatalf("%s model build: %v\n%s", phase, err, out)
		}
		out, err := runGo(module, "0", "list", "-deps", "-json", "./internal/decimalref")
		if err != nil {
			t.Fatalf("%s dependency listing: %v\n%s", phase, err, out)
		}
		packages, err := decodePackages(out)
		if err != nil {
			t.Fatal(err)
		}
		return checkReferenceDependencies(packages)
	}
	if err := check("baseline"); err != nil {
		t.Fatal(err)
	}
	t.Log("REFERENCE-DEPENDENCIES baseline built: standard library only")
	path := filepath.Join(module, "internal/decimalref/production_dependency.go")
	if err := os.WriteFile(path, []byte("package decimalref\nimport _ \""+modulePath+"\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := check("mutant"); err == nil {
		t.Fatal("compiled reference with a production dependency was accepted")
	} else {
		t.Logf("REFERENCE-DEPENDENCIES compiled mutant rejected: %v", err)
	}
}
