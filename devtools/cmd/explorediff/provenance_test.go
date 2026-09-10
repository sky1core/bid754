package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceIdentityBindsUntrackedContentAndSnapshot(t *testing.T) {
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	gitOutput := func(root string, args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid", "GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	fixture := t.TempDir()
	gitOutput(fixture, "init", "--quiet", "--template=")
	tree := gitOutput(fixture, "write-tree")
	head := gitOutput(fixture, "commit-tree", tree, "-m", "fixture")
	write := func(path, content string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(fixture, ".git/HEAD"), head+"\n")
	write(filepath.Join(fixture, ".gitignore"), "__pycache__/\n")
	lib := filepath.Join(fixture, "devtools/scripts/lib")
	if err := os.MkdirAll(lib, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"source_snapshot.py", "worktree_files.py"} {
		data, err := os.ReadFile(filepath.Join(repo, "devtools/scripts/lib", name))
		if err != nil {
			t.Fatal(err)
		}
		write(filepath.Join(lib, name), string(data))
	}
	path := filepath.Join(fixture, "untracked.go")
	write(path, "package first\n")
	first, err := sourceIdentity(fixture)
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := sourceIdentity(fixture)
	if err != nil || first != repeated {
		t.Fatalf("unstable matching source: %s %s %v", first, repeated, err)
	}
	snapshot := filepath.Join(t.TempDir(), "source.tar")
	script := filepath.Join(lib, "source_snapshot.py")
	out, err := exec.Command("python3", "-B", script, "create", snapshot, "--root", fixture).CombinedOutput()
	if err != nil {
		t.Fatalf("snapshot create: %v %s", err, out)
	}
	id := strings.TrimSpace(string(out))
	verify := func() ([]byte, error) {
		return exec.Command("python3", "-B", script, "verify-source", snapshot, "--expected-id", id, "--root", fixture).CombinedOutput()
	}
	if out, err := verify(); err != nil {
		t.Fatalf("matching snapshot rejected: %v %s", err, out)
	}
	write(path, "package second\n")
	second, err := sourceIdentity(fixture)
	if err != nil || second == first {
		t.Fatalf("untracked contents not bound: %s %s %v", first, second, err)
	}
	if out, err := verify(); err == nil || !strings.Contains(string(out), "mismatch") {
		t.Fatalf("changed snapshot accepted: %v %s", err, out)
	}
}

func TestBuildEnvironmentRejectsForeignNativePaths(t *testing.T) {
	for _, key := range []string{"CGO_CFLAGS", "CGO_LDFLAGS", "GOFLAGS", "GOWORK", "CC", "CXX", "CPATH", "LIBRARY_PATH"} {
		t.Setenv(key, "foreign-tree")
	}
	env := strings.Join(buildEnvironment(), "\n")
	if strings.Contains(env, "foreign-tree") || !strings.Contains(env, "CGO_ENABLED=1") || !strings.Contains(env, "GOENV=off") || !strings.Contains(env, "GOWORK=off") {
		t.Fatal(env)
	}
}
