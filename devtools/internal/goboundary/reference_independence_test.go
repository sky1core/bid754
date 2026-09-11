package goboundary

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func checkReferenceDependencies(packages []listedPackage, name string) error {
	model := modulePath + "/internal/" + name
	found := false
	for _, p := range packages {
		if p.ImportPath == model {
			found = true
			continue
		}
		if p.ImportPath == modulePath+"/internal/decimalref" {
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
	for _, name := range []string{"decimalref", "tier1ref"} {
		t.Run(name, func(t *testing.T) { checkReferenceDependencyBoundary(t, name) })
	}
}

func checkReferenceDependencyBoundary(t *testing.T, name string) {
	source, err := filepath.Abs("../../../bid754-go")
	if err != nil {
		t.Fatal(err)
	}
	module := t.TempDir()
	copyGoModuleSources(t, source, module)
	check := func(phase string) error {
		t.Helper()
		if out, err := runGo(module, "0", "build", "./internal/"+name); err != nil {
			t.Fatalf("%s model build: %v\n%s", phase, err, out)
		}
		out, err := runGo(module, "0", "list", "-deps", "-json", "./internal/"+name)
		if err != nil {
			t.Fatalf("%s dependency listing: %v\n%s", phase, err, out)
		}
		packages, err := decodePackages(out)
		if err != nil {
			t.Fatal(err)
		}
		return checkReferenceDependencies(packages, name)
	}
	if err := check("baseline"); err != nil {
		t.Fatal(err)
	}
	t.Log("REFERENCE-DEPENDENCIES baseline built: standard library only")
	path := filepath.Join(module, "internal/"+name+"/production_dependency.go")
	if err := os.WriteFile(path, []byte("package "+name+"\nimport _ \""+modulePath+"\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := check("mutant"); err == nil {
		t.Fatal("compiled reference with a production dependency was accepted")
	} else {
		t.Logf("REFERENCE-DEPENDENCIES compiled mutant rejected: %v", err)
	}
}
