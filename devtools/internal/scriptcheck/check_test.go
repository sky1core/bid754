package scriptcheck

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestShellSyntaxChecksEverySelectedScript(t *testing.T) {
	checker, err := filepath.Abs("../../scripts/check_scripts.py")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	init := exec.Command("git", "init", "--quiet", dir)
	if out, err := init.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	for _, name := range []string{"first.sh", "second script.sh"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/usr/bin/env bash\ntrue\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	check := func(wantFailure bool) {
		t.Helper()
		cmd := exec.Command("python3", "-B", checker)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if wantFailure {
			if err == nil || !strings.Contains(string(out), "shell syntax check failed: second script.sh") || !strings.Contains(string(out), "syntax error") {
				t.Fatalf("missing intended syntax failure: %v: %s", err, out)
			}
		} else if err != nil || !strings.Contains(string(out), "SCRIPT-SYNTAX-CHECK checked=2") {
			t.Fatalf("baseline failed: %v: %s", err, out)
		}
	}
	check(false)
	if err := os.WriteFile(filepath.Join(dir, "second script.sh"), []byte("#!/usr/bin/env bash\nif\n"), 0600); err != nil {
		t.Fatal(err)
	}
	check(true)
}
