package platformdigest

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func worktreeCommand(t *testing.T, dir string, name string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
	return out
}

func worktreeFixture(t *testing.T) string {
	t.Helper()
	t.Setenv("BID754_SNAPSHOT_ARCHIVE", "")
	t.Setenv("BID754_SNAPSHOT_ID", "")
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "repo")
	worktreeCommand(t, root, "git", "init", "--quiet", dir)
	for _, rel := range []string{"devtools/scripts/print_tree_id.sh", "devtools/scripts/lib/worktree_files.py", "devtools/scripts/lib/source_snapshot.py"} {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, rel)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, rel), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeWorktreeFile(t, dir, "README.md", "tracked content")
	worktreeCommand(t, dir, "git", "add", "--", "README.md")
	return dir
}

func writeWorktreeFile(t *testing.T, dir, rel, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, rel), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func archiveWorktree(t *testing.T, dir string) map[string]tar.Header {
	t.Helper()
	paths := worktreeCommand(t, dir, "python3", "-B", "devtools/scripts/lib/worktree_files.py")
	cmd := exec.Command("tar", "--no-recursion", "--exclude=.git", "--null", "-T", "-", "-cf", "-")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "COPYFILE_DISABLE=1")
	cmd.Stdin = bytes.NewReader(paths)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	data, err := cmd.Output()
	if err != nil {
		t.Fatalf("archive worktree: %v\n%s", err, stderr.Bytes())
	}
	entries := map[string]tar.Header{}
	reader := tar.NewReader(bytes.NewReader(data))
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		name := strings.TrimSuffix(strings.TrimPrefix(header.Name, "./"), "/")
		if _, exists := entries[name]; exists {
			t.Fatalf("duplicate archive entry %q", name)
		}
		entries[name] = *header
	}
	return entries
}

func TestWorktreeArchiveDeletionAndLinks(t *testing.T) {
	dir := worktreeFixture(t)
	index := worktreeCommand(t, dir, "git", "ls-files", "--stage", "-z")
	if err := os.Rename(filepath.Join(dir, "README.md"), filepath.Join(t.TempDir(), "saved")); err != nil {
		t.Fatal(err)
	}
	writeWorktreeFile(t, dir, "-file with\nnewline", "content")
	if err := os.Symlink("absent", filepath.Join(dir, "dangling")); err != nil {
		t.Fatal(err)
	}
	worktreeCommand(t, dir, "git", "init", "--quiet", "nested")
	writeWorktreeFile(t, dir, "nested/source", "nested content")
	if err := os.Symlink("nested", filepath.Join(dir, "directory-link")); err != nil {
		t.Fatal(err)
	}
	entries := archiveWorktree(t, dir)
	if _, exists := entries["README.md"]; exists {
		t.Fatal("deleted tracked file was archived")
	}
	for _, name := range []string{"-file with\nnewline", "nested/source"} {
		if entries[name].Typeflag != tar.TypeReg {
			t.Fatalf("missing regular file %q", name)
		}
	}
	for name, target := range map[string]string{"dangling": "absent", "directory-link": "nested"} {
		if entries[name].Typeflag != tar.TypeSymlink || entries[name].Linkname != target {
			t.Fatalf("symlink %q was lost or dereferenced: %+v", name, entries[name])
		}
	}
	for name := range entries {
		if strings.Contains("/"+name+"/", "/.git/") || strings.HasPrefix(name, "directory-link/") {
			t.Fatalf("excluded metadata or symlink child archived: %q", name)
		}
	}
	if !bytes.Equal(index, worktreeCommand(t, dir, "git", "ls-files", "--stage", "-z")) {
		t.Fatal("enumerating the worktree changed the index")
	}
	if err := os.Mkdir(filepath.Join(dir, "README.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeWorktreeFile(t, dir, "README.md/child", "replacement directory")
	entries = archiveWorktree(t, dir)
	if entries["README.md"].Typeflag != tar.TypeDir || entries["README.md/child"].Typeflag != tar.TypeReg {
		t.Fatal("tracked file replaced by directory was not preserved")
	}
}

func TestWorktreeDigestNestedRepository(t *testing.T) {
	dir := worktreeFixture(t)
	worktreeCommand(t, dir, "git", "init", "--quiet", "nested")
	writeWorktreeFile(t, dir, "nested/.gitignore", "ignored.log\n")
	writeWorktreeFile(t, dir, "nested/source", "first")
	writeWorktreeFile(t, dir, "nested/ignored.log", "also transferred")
	if err := os.Symlink("absent", filepath.Join(dir, "nested/link")); err != nil {
		t.Fatal(err)
	}
	fingerprint := func() string {
		return strings.TrimSpace(string(worktreeCommand(t, dir, "python3", "-B", "devtools/scripts/lib/worktree_files.py", "--fingerprint")))
	}
	first := fingerprint()
	if len(first) != 64 || first != fingerprint() {
		t.Fatalf("unstable or missing fingerprint: %s", first)
	}
	if got := strings.TrimSpace(string(worktreeCommand(t, dir, "bash", "devtools/scripts/print_tree_id.sh"))); got != "unknown" {
		t.Fatalf("tree without history should be unknown, got %q", got)
	}
	writeWorktreeFile(t, dir, "nested/.git/description", "metadata changed")
	if first != fingerprint() {
		t.Fatal("excluded nested Git metadata changed the fingerprint")
	}
	for _, rel := range []string{"nested/source", "nested/ignored.log"} {
		before := fingerprint()
		writeWorktreeFile(t, dir, rel, "changed")
		if before == fingerprint() {
			t.Fatalf("transferred content %q was absent from fingerprint", rel)
		}
	}
	before := fingerprint()
	if err := os.Rename(filepath.Join(dir, "nested/link"), filepath.Join(t.TempDir(), "saved-link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("different-absent", filepath.Join(dir, "nested/link")); err != nil {
		t.Fatal(err)
	}
	if before == fingerprint() {
		t.Fatal("symlink target was absent from fingerprint")
	}
}

func TestWorktreeArchiveDirectoryReplacement(t *testing.T) {
	for _, replacement := range []string{"file", "symlink"} {
		t.Run(replacement, func(t *testing.T) {
			dir := worktreeFixture(t)
			if err := os.Mkdir(filepath.Join(dir, "tracked-directory"), 0o755); err != nil {
				t.Fatal(err)
			}
			writeWorktreeFile(t, dir, "tracked-directory/source", "tracked")
			worktreeCommand(t, dir, "git", "add", "--", "tracked-directory/source")
			saved := filepath.Join(t.TempDir(), "saved-directory")
			if err := os.Rename(filepath.Join(dir, "tracked-directory"), saved); err != nil {
				t.Fatal(err)
			}
			if replacement == "file" {
				writeWorktreeFile(t, dir, "tracked-directory", "replacement file")
			} else if err := os.Symlink(saved, filepath.Join(dir, "tracked-directory")); err != nil {
				t.Fatal(err)
			}
			entries := archiveWorktree(t, dir)
			if _, exists := entries["tracked-directory/source"]; exists {
				t.Fatal("deleted indexed child or symlink target was archived")
			}
			want := byte(tar.TypeReg)
			if replacement == "symlink" {
				want = tar.TypeSymlink
			}
			if entries["tracked-directory"].Typeflag != want {
				t.Fatal("replacement file type was not preserved")
			}
		})
	}
}
