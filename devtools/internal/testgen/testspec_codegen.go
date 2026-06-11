package testgen

import (
	"fmt"
	"os"
	"path/filepath"
)

type testspecPackageTemplate struct {
	OutputPath   string
	TemplatePath string
}

// testspecPackageTemplates lists the verbatim templates that make up the
// generated bid754-go/internal/testspec package: the shared spec data model
// plus the LoadGenerated index+shard loader consumed by the bid754-go
// verification harness and platformdigest. The schema source of truth stays
// in this package; the harness coverage assertions catch template drift.
var testspecPackageTemplates = []testspecPackageTemplate{
	{OutputPath: "../bid754-go/internal/testspec/spec.go", TemplatePath: "internal/testgen/testspec_templates/spec.go.tmpl"},
	{OutputPath: "../bid754-go/internal/testspec/spec_io.go", TemplatePath: "internal/testgen/testspec_templates/spec_io.go.tmpl"},
}

func WriteTestspecPackageOutputs(repoRoot string) error {
	files, err := GenerateTestspecPackageOutputs(repoRoot)
	if err != nil {
		return err
	}
	for path, data := range files {
		fullPath := filepath.Join(repoRoot, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write generated testspec package %q: %w", fullPath, err)
		}
	}
	return nil
}

func GenerateTestspecPackageOutputs(repoRoot string) (map[string][]byte, error) {
	outputs := make(map[string][]byte, len(testspecPackageTemplates))
	for _, item := range testspecPackageTemplates {
		data, err := os.ReadFile(filepath.Join(repoRoot, item.TemplatePath))
		if err != nil {
			return nil, fmt.Errorf("read generated testspec template %q: %w", item.TemplatePath, err)
		}
		outputs[item.OutputPath] = []byte(dectestGeneratedSourceFromTemplate(data))
	}
	return formatGeneratedGoOutputs(outputs)
}
