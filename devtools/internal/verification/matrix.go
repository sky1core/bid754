package verification

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

func ValidateMatrix(root string, plan Plan, dir, snapshot, invocation string) error {
	if len(plan.Matrix) == 0 {
		return fmt.Errorf("execution plan has no required matrix")
	}
	required := map[string]bool{}
	for _, entry := range plan.Matrix {
		required[entry.Profile+"@"+entry.Platform] = true
	}
	seen := map[string]bool{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "result.json" {
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("result must not be a symlink: %s", path)
		}
		var result Result
		if err := ReadJSON(path, &result); err != nil {
			return err
		}
		key := result.Profile + "@" + result.Platform
		if !required[key] || seen[key] {
			return fmt.Errorf("unexpected or duplicate matrix result %s", key)
		}
		if err := Validate(root, plan, filepath.Dir(path), result, snapshot, invocation); err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
		seen[key] = true
		return nil
	})
	if err != nil {
		return err
	}
	for key := range required {
		if !seen[key] {
			return fmt.Errorf("missing required matrix result %s", key)
		}
	}
	return nil
}
