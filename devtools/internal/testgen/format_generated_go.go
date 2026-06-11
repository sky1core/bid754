package testgen

import (
	"fmt"
	"go/format"
	"strings"
)

func formatGeneratedGoOutputs(files map[string][]byte) (map[string][]byte, error) {
	for path, data := range files {
		if !strings.HasSuffix(path, ".go") {
			continue
		}
		formatted, err := format.Source(data)
		if err != nil {
			return nil, fmt.Errorf("format generated Go file %q: %w", path, err)
		}
		files[path] = formatted
	}
	return files, nil
}
