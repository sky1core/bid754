package platformdigest

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func snapshotCommand(t *testing.T, dir string, args ...string) []byte {
	t.Helper()
	return worktreeCommand(t, dir, "python3", append([]string{"-B", "devtools/scripts/lib/source_snapshot.py"}, args...)...)
}

func createSnapshot(t *testing.T, dir string) (string, string) {
	t.Helper()
	archive := filepath.Join(t.TempDir(), "source.tar")
	identity := strings.TrimSpace(string(snapshotCommand(t, dir, "create", archive)))
	data, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	checksum := sha256.Sum256(data)
	if identity != hex.EncodeToString(checksum[:]) {
		t.Fatalf("snapshot ID does not identify transferred bytes: %s", identity)
	}
	info, err := os.Stat(archive)
	if err != nil || info.Mode().Perm() != 0o444 {
		t.Fatalf("snapshot must be read-only: %v, %v", info, err)
	}
	return archive, identity
}

func snapshotFailure(t *testing.T, dir, want string, args ...string) {
	t.Helper()
	cmd := exec.Command("python3", append([]string{"-B", "devtools/scripts/lib/source_snapshot.py"}, args...)...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), want) {
		t.Fatalf("%v: want failure %q, got %v\n%s", args, want, err, out)
	}
}

func snapshotFixture(t *testing.T) string {
	t.Helper()
	dir := worktreeFixture(t)
	writeWorktreeFile(t, dir, ".gitignore", "ignored-root\n")
	writeWorktreeFile(t, dir, "ignored-root", "not transferred")
	writeWorktreeFile(t, dir, "-file with\nnewline", "literal path")
	writeWorktreeFile(t, dir, "unicode-한글", "byte path")
	writeWorktreeFile(t, dir, "executable", "executable content")
	if err := os.Chmod(filepath.Join(dir, "executable"), 0o751); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "mode-dir"), 0o751); err != nil {
		t.Fatal(err)
	}
	writeWorktreeFile(t, dir, "mode-dir/child", "directory mode")
	for name, target := range map[string]string{"dangling": "absent", "external": t.TempDir(), "directory-link": "mode-dir"} {
		if err := os.Symlink(target, filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
	}
	worktreeCommand(t, dir, "git", "init", "--quiet", "nested")
	writeWorktreeFile(t, dir, "nested/.gitignore", "ignored\n")
	writeWorktreeFile(t, dir, "nested/ignored", "nested ignored content is transferred")
	writeWorktreeFile(t, dir, "nested/tracked", "nested tracked")
	worktreeCommand(t, filepath.Join(dir, "nested"), "git", "add", "--", "tracked")
	if err := os.Rename(filepath.Join(dir, "README.md"), filepath.Join(t.TempDir(), "deleted")); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestSourceSnapshotRoundTrip(t *testing.T) {
	dir := snapshotFixture(t)
	archive, identity := createSnapshot(t, dir)
	_, secondID := createSnapshot(t, dir)
	if identity != secondID {
		t.Fatalf("unchanged source produced different snapshots: %s %s", identity, secondID)
	}
	destination := t.TempDir()
	snapshotCommand(t, dir, "restore", archive, "--root", destination, "--expected-id", identity)
	snapshotCommand(t, dir, "verify", archive, "--root", destination, "--expected-id", identity)
	snapshotCommand(t, destination, "verify-source", archive, "--root", destination, "--expected-id", identity)
	for _, repo := range []string{"", "nested"} {
		before := worktreeCommand(t, filepath.Join(dir, repo), "git", "ls-files", "--stage", "-z")
		after := worktreeCommand(t, filepath.Join(destination, repo), "git", "ls-files", "--stage", "-z")
		if !bytes.Equal(before, after) {
			t.Fatalf("tracked index changed for repository %q", repo)
		}
	}
	for _, rel := range []string{"README.md", "ignored-root", "nested/.git/description"} {
		if _, err := os.Lstat(filepath.Join(destination, rel)); !os.IsNotExist(err) {
			t.Fatalf("excluded/deleted path restored: %s: %v", rel, err)
		}
	}
	for _, rel := range []string{"-file with\nnewline", "unicode-한글", "nested/ignored"} {
		before, _ := os.ReadFile(filepath.Join(dir, rel))
		after, err := os.ReadFile(filepath.Join(destination, rel))
		if err != nil || !bytes.Equal(before, after) {
			t.Fatalf("source bytes lost for %q: %v", rel, err)
		}
	}
}

func TestSourceSnapshotPortableSymlinkMode(t *testing.T) {
	dir := worktreeFixture(t)
	link := filepath.Join(dir, "dangling")
	if err := os.Symlink("absent", link); err != nil {
		t.Fatal(err)
	}
	archive, identity := createSnapshot(t, dir)
	f, err := os.Open(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	reader := tar.NewReader(f)
	canonicalLink := false
	for {
		h, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if h.Name != "manifest.json" {
			continue
		}
		var manifest struct {
			Entries []struct {
				Kind string
				Mode int
			}
		}
		if err := json.NewDecoder(reader).Decode(&manifest); err != nil {
			t.Fatal(err)
		}
		for _, entry := range manifest.Entries {
			if entry.Kind == "symlink" {
				canonicalLink = entry.Mode == 0o777
			}
		}
	}
	if !canonicalLink {
		t.Fatal("symlink mode is not portable")
	}
	if runtime.GOOS == "darwin" {
		worktreeCommand(t, dir, "python3", "-c", "import os,sys; os.chmod(sys.argv[1], 0o700, follow_symlinks=False)", link)
		_, after := createSnapshot(t, dir)
		if after != identity {
			t.Fatal("platform-specific symlink permissions changed portable snapshot identity")
		}
	}
	destination := t.TempDir()
	snapshotCommand(t, dir, "restore", archive, "--root", destination, "--expected-id", identity)
	snapshotCommand(t, destination, "verify-source", archive, "--root", destination, "--expected-id", identity)
}

func TestSourceSnapshotRestoredCorruptions(t *testing.T) {
	dir := snapshotFixture(t)
	archive, identity := createSnapshot(t, dir)
	for _, tc := range []struct {
		name, want string
		change     func(*testing.T, string)
	}{
		{"bytes", "restored file content mismatch: executable", func(t *testing.T, r string) { writeWorktreeFile(t, r, "executable", "changed content") }},
		{"mode", "restored mode mismatch: executable", func(t *testing.T, r string) {
			if err := os.Chmod(filepath.Join(r, "executable"), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
		{"directory-mode", "restored mode mismatch: mode-dir", func(t *testing.T, r string) {
			if err := os.Chmod(filepath.Join(r, "mode-dir"), 0o755); err != nil {
				t.Fatal(err)
			}
		}},
		{"symlink", "restored symlink target mismatch: dangling", func(t *testing.T, r string) {
			if err := os.Rename(filepath.Join(r, "dangling"), filepath.Join(t.TempDir(), "saved")); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink("different-absent", filepath.Join(r, "dangling")); err != nil {
				t.Fatal(err)
			}
		}},
		{"index", "restored tracked index mismatch:", func(t *testing.T, r string) {
			worktreeCommand(t, r, "git", "update-index", "--force-remove", "--", "README.md")
		}},
		{"nested-index", "restored tracked index mismatch: nested", func(t *testing.T, r string) {
			worktreeCommand(t, filepath.Join(r, "nested"), "git", "update-index", "--force-remove", "--", "tracked")
		}},
		{"deletion", "unexpected restored path: README.md", func(t *testing.T, r string) { writeWorktreeFile(t, r, "README.md", "resurrected") }},
		{"missing-file", "restored path missing: executable", func(t *testing.T, r string) {
			if err := os.Rename(filepath.Join(r, "executable"), filepath.Join(t.TempDir(), "saved")); err != nil {
				t.Fatal(err)
			}
		}},
		{"symlink-parent", "restored file type mismatch: mode-dir", func(t *testing.T, r string) {
			saved := filepath.Join(t.TempDir(), "saved")
			if err := os.Rename(filepath.Join(r, "mode-dir"), saved); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(saved, filepath.Join(r, "mode-dir")); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			destination := t.TempDir()
			snapshotCommand(t, dir, "restore", archive, "--root", destination, "--expected-id", identity)
			tc.change(t, destination)
			snapshotFailure(t, dir, tc.want, "verify", archive, "--root", destination, "--expected-id", identity)
		})
	}
}

func mutateSnapshot(t *testing.T, archive string, change func(*tar.Header, []byte) (*tar.Header, []byte)) string {
	t.Helper()
	data, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	writer := tar.NewWriter(&output)
	reader := tar.NewReader(bytes.NewReader(data))
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		header, body = change(header, body)
		if header == nil {
			continue
		}
		header.Size = int64(len(body))
		if err := writer.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write(body); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "corrupt.tar")
	if err := os.WriteFile(path, output.Bytes(), 0o444); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSourceSnapshotArchiveCorruptions(t *testing.T) {
	dir := snapshotFixture(t)
	archive, identity := createSnapshot(t, dir)
	for _, tc := range []struct {
		name, want string
		change     func(*tar.Header, []byte) (*tar.Header, []byte)
	}{
		{"manifest-missing", "snapshot manifest missing", func(h *tar.Header, b []byte) (*tar.Header, []byte) {
			if h.Name == "manifest.json" {
				return nil, nil
			}
			return h, b
		}},
		{"bytes", "snapshot file content mismatch:", func(h *tar.Header, b []byte) (*tar.Header, []byte) {
			if strings.HasPrefix(h.Name, "blobs/") && len(b) > 0 {
				b[0] ^= 1
			}
			return h, b
		}},
		{"archive-mode", "invalid snapshot archive member:", func(h *tar.Header, b []byte) (*tar.Header, []byte) { h.Mode = 0o777; return h, b }},
		{"member-link", "invalid snapshot archive member:", func(h *tar.Header, b []byte) (*tar.Header, []byte) {
			if h.Name == "blobs/0" {
				h.Typeflag = tar.TypeSymlink
				h.Linkname = "/tmp/outside"
				b = nil
			}
			return h, b
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			corrupt := mutateSnapshot(t, archive, tc.change)
			destination := t.TempDir()
			snapshotFailure(t, dir, tc.want, "verify-source", corrupt, "--root", destination, "--expected-id", identity)
			snapshotFailure(t, dir, tc.want, "restore", corrupt, "--root", destination, "--expected-id", identity)
			files, err := os.ReadDir(destination)
			if err != nil || len(files) != 0 {
				t.Fatalf("invalid archive wrote destination: %v %v", files, err)
			}
		})
	}
	snapshotFailure(t, dir, "snapshot ID mismatch:", "restore", archive, "--root", t.TempDir(), "--expected-id", strings.Repeat("0", 64))
	snapshotFailure(t, dir, "snapshot ID mismatch:", "verify-source", archive, "--root", dir, "--expected-id", strings.Repeat("0", 64))
}

func TestSourceSnapshotFrozenReceiver(t *testing.T) {
	dir := snapshotFixture(t)
	archive := filepath.Join(t.TempDir(), "source.tar")
	receiver := filepath.Join(t.TempDir(), "receiver.py")
	identity := strings.TrimSpace(string(snapshotCommand(t, dir, "create", archive, "--receiver", receiver)))
	writeWorktreeFile(t, dir, "executable", "changed after capture")
	writeWorktreeFile(t, dir, "devtools/scripts/lib/source_snapshot.py", "raise RuntimeError('live source must not run')\n")
	destination := t.TempDir()
	cmd := exec.Command("python3", "-B", receiver, "restore", "-", "--root", destination, "--expected-id", identity)
	input, err := os.Open(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	cmd.Stdin = input
	out, err := cmd.CombinedOutput()
	if err != nil || strings.TrimSpace(string(out)) != identity {
		t.Fatalf("frozen stream restore: %v\n%s", err, out)
	}
	got, err := os.ReadFile(filepath.Join(destination, "executable"))
	if err != nil || string(got) != "executable content" {
		t.Fatalf("snapshot reread live source: %q %v", got, err)
	}
	worktreeCommand(t, destination, "python3", "-B", receiver, "verify", archive, "--root", destination, "--expected-id", identity)
}

func commitWorktreeFixture(t *testing.T, dir string) {
	t.Helper()
	worktreeCommand(t, dir, "git", "add", "--", "README.md", "devtools/scripts/print_tree_id.sh", "devtools/scripts/lib/worktree_files.py", "devtools/scripts/lib/source_snapshot.py")
	tree := strings.TrimSpace(string(worktreeCommand(t, dir, "git", "write-tree")))
	commit := strings.TrimSpace(string(worktreeCommand(t, dir, "git", "-c", "user.name=Snapshot Fixture", "-c", "user.email=fixture@example.invalid", "commit-tree", tree, "-m", "fixture")))
	worktreeCommand(t, dir, "git", "update-ref", "HEAD", commit)
}

func TestSourceSnapshotShellWrapperIdentity(t *testing.T) {
	dir := worktreeFixture(t)
	commitWorktreeFixture(t, dir)
	worktreeCommand(t, dir, "git", "config", "core.filemode", "false")
	wrapper := func() string {
		return strings.TrimSpace(string(worktreeCommand(t, dir, "bash", "devtools/scripts/print_tree_id.sh")))
	}
	first := wrapper()
	if len(first) != 64 || first != wrapper() {
		t.Fatalf("clean shell wrapper ID unstable: %s", first)
	}
	archive, identity := createSnapshot(t, dir)
	if first != identity {
		t.Fatalf("shell wrapper not bound to snapshot bytes: %s != %s", first, identity)
	}
	if got := strings.TrimSpace(string(worktreeCommand(t, dir, "bash", "devtools/scripts/print_tree_id.sh", "--snapshot", archive))); got != first {
		t.Fatalf("snapshot shell wrapper differs: %s != %s", got, first)
	}
	if err := os.Chmod(filepath.Join(dir, "README.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	if status := worktreeCommand(t, dir, "git", "status", "--porcelain"); len(status) != 0 {
		t.Fatalf("fixture must reproduce Git-clean mode change: %s", status)
	}
	if first == wrapper() {
		t.Fatal("shell wrapper ignored mode change hidden by Git status")
	}
	writeWorktreeFile(t, dir, "README.md", "dirty content")
	before := wrapper()
	if !strings.HasSuffix(before, "-dirty") {
		t.Fatalf("dirty guard lost: %s", before)
	}
	writeWorktreeFile(t, dir, "README.md", "different dirty content")
	if before == wrapper() {
		t.Fatal("shell wrapper used constant dirty fingerprint")
	}
	worktreeCommand(t, dir, "git", "init", "--quiet", "nested")
	writeWorktreeFile(t, dir, "nested/source", "first nested content")
	before = wrapper()
	writeWorktreeFile(t, dir, "nested/source", "second nested content")
	if before == wrapper() {
		t.Fatal("shell wrapper ignored nested repository contents")
	}
}

func TestSourceSnapshotUnsafeManifest(t *testing.T) {
	dir := snapshotFixture(t)
	archive, identity := createSnapshot(t, dir)
	for _, path := range []string{"../escape", "/absolute", ".git/config", "dangling/child"} {
		t.Run(path, func(t *testing.T) {
			corrupt := mutateSnapshot(t, archive, func(h *tar.Header, b []byte) (*tar.Header, []byte) {
				if h.Name == "manifest.json" {
					var manifest map[string]any
					if err := json.Unmarshal(b, &manifest); err != nil {
						t.Fatal(err)
					}
					manifest["entries"].([]any)[0].(map[string]any)["path"] = hex.EncodeToString([]byte(path))
					var err error
					b, err = json.Marshal(manifest)
					if err != nil {
						t.Fatal(err)
					}
					b = append(b, '\n')
				}
				return h, b
			})
			snapshotFailure(t, dir, "unsafe snapshot", "restore", corrupt, "--root", t.TempDir(), "--expected-id", identity)
		})
	}
}

func TestSourceSnapshotIndexStagesAndReplacements(t *testing.T) {
	dir := worktreeFixture(t)
	if err := os.Mkdir(filepath.Join(dir, "replaced-dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeWorktreeFile(t, dir, "replaced-dir/child", "tracked child")
	worktreeCommand(t, dir, "git", "add", "--", "replaced-dir/child")
	if err := os.Rename(filepath.Join(dir, "replaced-dir"), filepath.Join(t.TempDir(), "saved")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("absent", filepath.Join(dir, "replaced-dir")); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(dir, "README.md"), filepath.Join(t.TempDir(), "saved")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "README.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeWorktreeFile(t, dir, "README.md/child", "replacement directory")
	worktreeCommand(t, dir, "git", "init", "--quiet", "nested")
	writeWorktreeFile(t, dir, "nested/source", "nested commit content")
	nested := filepath.Join(dir, "nested")
	worktreeCommand(t, nested, "git", "add", "--", "source")
	tree := strings.TrimSpace(string(worktreeCommand(t, nested, "git", "write-tree")))
	commit := strings.TrimSpace(string(worktreeCommand(t, nested, "git", "-c", "user.name=Snapshot Fixture", "-c", "user.email=fixture@example.invalid", "commit-tree", tree, "-m", "fixture")))
	worktreeCommand(t, nested, "git", "update-ref", "HEAD", commit)
	worktreeCommand(t, dir, "git", "add", "--", "nested")
	oid := strings.TrimSpace(string(worktreeCommand(t, nested, "git", "rev-parse", "HEAD:source")))
	cmd := exec.Command("git", "update-index", "-z", "--index-info")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("100644 " + oid + " 1\tconflict\x00" + "100755 " + oid + " 2\tconflict\x00" + "120000 " + oid + " 3\tconflict\x00")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("stage fixture: %v\n%s", err, out)
	}
	archive, identity := createSnapshot(t, dir)
	destination := t.TempDir()
	snapshotCommand(t, dir, "restore", archive, "--root", destination, "--expected-id", identity)
	if before, after := worktreeCommand(t, dir, "git", "ls-files", "--stage", "-z"), worktreeCommand(t, destination, "git", "ls-files", "--stage", "-z"); !bytes.Equal(before, after) {
		t.Fatal("conflict stages, gitlink, or deleted tracked entries changed")
	}
	snapshotCommand(t, destination, "verify-source", archive, "--root", destination, "--expected-id", identity)
	if target, err := os.Readlink(filepath.Join(destination, "replaced-dir")); err != nil || target != "absent" {
		t.Fatalf("directory replacement lost: %s %v", target, err)
	}
	if data, err := os.ReadFile(filepath.Join(destination, "README.md/child")); err != nil || string(data) != "replacement directory" {
		t.Fatalf("file replacement lost: %s %v", data, err)
	}
}

func TestSourceSnapshotUnsafeIndex(t *testing.T) {
	dir := snapshotFixture(t)
	archive, identity := createSnapshot(t, dir)
	for _, path := range []string{"../escape", ".GiT/config"} {
		t.Run(path, func(t *testing.T) {
			corrupt := mutateSnapshot(t, archive, func(h *tar.Header, b []byte) (*tar.Header, []byte) {
				if h.Name == "manifest.json" {
					var manifest map[string]any
					if err := json.Unmarshal(b, &manifest); err != nil {
						t.Fatal(err)
					}
					record := manifest["repositories"].([]any)[0].(map[string]any)
					record["index"] = hex.EncodeToString([]byte("100644 " + strings.Repeat("a", 40) + " 0\t" + path + "\x00"))
					var err error
					b, err = json.Marshal(manifest)
					if err != nil {
						t.Fatal(err)
					}
					b = append(b, '\n')
				}
				return h, b
			})
			snapshotFailure(t, dir, "unsafe snapshot path:", "restore", corrupt, "--root", t.TempDir(), "--expected-id", identity)
		})
	}
}

func TestSourceSnapshotDestinationSafety(t *testing.T) {
	dir := worktreeFixture(t)
	archive, identity := createSnapshot(t, dir)
	destination := t.TempDir()
	writeWorktreeFile(t, destination, "user-file", "preserve")
	snapshotFailure(t, dir, "restore destination must be empty", "restore", archive, "--root", destination, "--expected-id", identity)
	if got, err := os.ReadFile(filepath.Join(destination, "user-file")); err != nil || string(got) != "preserve" {
		t.Fatalf("destination changed: %s %v", got, err)
	}
	snapshotFailure(t, dir, "snapshot output must be outside the source tree", "create", filepath.Join(dir, "source.tar"))
	snapshotFailure(t, dir, "File exists", "create", archive)
}

func TestSourceSnapshotSHA256Index(t *testing.T) {
	source := worktreeFixture(t)
	dir := filepath.Join(t.TempDir(), "repo")
	worktreeCommand(t, source, "git", "init", "--quiet", "--object-format=sha256", dir)
	for _, rel := range []string{"README.md", "devtools/scripts/print_tree_id.sh", "devtools/scripts/lib/worktree_files.py", "devtools/scripts/lib/source_snapshot.py"} {
		data, err := os.ReadFile(filepath.Join(source, rel))
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
	worktreeCommand(t, dir, "git", "add", "--", "README.md")
	archive, identity := createSnapshot(t, dir)
	destination := t.TempDir()
	snapshotCommand(t, dir, "restore", archive, "--root", destination, "--expected-id", identity)
	if got := strings.TrimSpace(string(worktreeCommand(t, destination, "git", "rev-parse", "--show-object-format"))); got != "sha256" {
		t.Fatalf("object format lost: %s", got)
	}
	snapshotCommand(t, destination, "verify-source", archive, "--root", destination, "--expected-id", identity)
}

func TestSourceSnapshotDigestCompatibility(t *testing.T) {
	dir := worktreeFixture(t)
	commitWorktreeFixture(t, dir)
	archive, identity := createSnapshot(t, dir)
	tree := strings.TrimSpace(string(snapshotCommand(t, dir, "tree-id", archive, "--expected-id", identity)))
	if tree != identity {
		t.Fatalf("clean tree contract lost: %s != %s", tree, identity)
	}
	results := t.TempDir()
	for _, platform := range []string{"darwin/arm64", "linux/amd64"} {
		parts := strings.Split(platform, "/")
		record := "PLATFORM-DIGEST-TREE " + tree + "\nPLATFORM-DIGEST goos=" + parts[0] + " goarch=" + parts[1] + " cases=1 sha256=" + strings.Repeat("a", 64) + "\n"
		writeWorktreeFile(t, results, "digest_"+parts[0]+"_"+parts[1]+".txt", record)
	}
	out := worktreeCommand(t, ".", "bash", "../../scripts/verify_digest.sh", "--results-dir", results, "--expected-tree", tree, "--require-platform", "darwin/arm64", "--require-platform", "linux/amd64")
	if !strings.Contains(string(out), "2 platforms agree") {
		t.Fatalf("snapshot ID rejected by verify_digest: %s", out)
	}
}

func addIgnoredSnapshotBuildOutputs(t *testing.T, dir string) {
	t.Helper()
	worktreeCommand(t, dir, "python3", "-m", "py_compile", "devtools/scripts/lib/worktree_files.py")
	for _, rel := range []string{".build/object.o", "test_results/verification-linux/snapshot/leg/result.json", "mode-dir/build.log"} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, rel)), 0o755); err != nil {
			t.Fatal(err)
		}
		writeWorktreeFile(t, dir, rel, "ignored build output")
	}
}

func TestSourceSnapshotVerifySourceAfterBuild(t *testing.T) {
	dir := snapshotFixture(t)
	writeWorktreeFile(t, dir, ".gitignore", "ignored-root\n/.build/\n/test_results/\n__pycache__/\n*.log\n")
	worktreeCommand(t, dir, "git", "add", "--", ".gitignore")
	archive, identity := createSnapshot(t, dir)
	for _, tc := range []struct {
		name, want string
		change     func(*testing.T, string)
	}{
		{"ignored-build-output", "", func(*testing.T, string) {}},
		{"bytes", "source entry mismatch: executable", func(t *testing.T, r string) { writeWorktreeFile(t, r, "executable", "corrupt") }},
		{"missing", "source path missing: executable", func(t *testing.T, r string) {
			if err := os.Rename(filepath.Join(r, "executable"), filepath.Join(t.TempDir(), "saved")); err != nil {
				t.Fatal(err)
			}
		}},
		{"extra", "unexpected source path: z-extra-source", func(t *testing.T, r string) { writeWorktreeFile(t, r, "z-extra-source", "extra") }},
		{"mode", "source entry mismatch: executable", func(t *testing.T, r string) {
			worktreeCommand(t, r, "git", "config", "core.filemode", "false")
			if err := os.Chmod(filepath.Join(r, "executable"), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
		{"directory-mode", "source entry mismatch: mode-dir", func(t *testing.T, r string) {
			if err := os.Chmod(filepath.Join(r, "mode-dir"), 0o755); err != nil {
				t.Fatal(err)
			}
		}},
		{"link", "source entry mismatch: dangling", func(t *testing.T, r string) {
			if err := os.Rename(filepath.Join(r, "dangling"), filepath.Join(t.TempDir(), "saved")); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink("new-target", filepath.Join(r, "dangling")); err != nil {
				t.Fatal(err)
			}
		}},
		{"index-remove", "source tracked index mismatch", func(t *testing.T, r string) {
			worktreeCommand(t, r, "git", "update-index", "--force-remove", "--", "README.md")
		}},
		{"index-add", "source tracked index mismatch", func(t *testing.T, r string) {
			worktreeCommand(t, r, "git", "add", "--", "executable")
		}},
		{"nested-index", "source tracked index mismatch", func(t *testing.T, r string) {
			worktreeCommand(t, filepath.Join(r, "nested"), "git", "update-index", "--force-remove", "--", "tracked")
		}},
		{"ignore-changed", "source entry mismatch: .gitignore", func(t *testing.T, r string) {
			writeWorktreeFile(t, r, ".gitignore", "ignored-root\n/.build/\n/test_results/\n__pycache__/\n*.log\nexecutable\n")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			destination := t.TempDir()
			snapshotCommand(t, dir, "restore", archive, "--root", destination, "--expected-id", identity)
			addIgnoredSnapshotBuildOutputs(t, destination)
			snapshotFailure(t, destination, "unexpected restored path:", "verify", archive, "--root", destination, "--expected-id", identity)
			if got := strings.TrimSpace(string(snapshotCommand(t, destination, "verify-source", archive, "--root", destination, "--expected-id", identity))); got != identity {
				t.Fatalf("source verification lost archived identity: %s", got)
			}
			tc.change(t, destination)
			if tc.want != "" {
				snapshotFailure(t, destination, tc.want, "verify-source", archive, "--root", destination, "--expected-id", identity)
			}
		})
	}
}

func TestSourceSnapshotContainerTreeID(t *testing.T) {
	for _, dirty := range []bool{false, true} {
		name := "clean"
		if dirty {
			name = "dirty"
		}
		t.Run(name, func(t *testing.T) {
			dir := worktreeFixture(t)
			writeWorktreeFile(t, dir, ".gitignore", "/.build/\n/test_results/\n__pycache__/\n*.log\n")
			worktreeCommand(t, dir, "git", "add", "--", ".gitignore")
			commitWorktreeFixture(t, dir)
			if dirty {
				writeWorktreeFile(t, dir, "README.md", "dirty source")
			}
			archive, identity := createSnapshot(t, dir)
			tree := strings.TrimSpace(string(snapshotCommand(t, dir, "tree-id", archive, "--expected-id", identity)))
			destination := t.TempDir()
			snapshotCommand(t, dir, "restore", archive, "--root", destination, "--expected-id", identity)
			if got := strings.TrimSpace(string(snapshotCommand(t, destination, "current-tree-id"))); got != "unknown" {
				t.Fatalf("restored fixture unexpectedly has HEAD: %s", got)
			}
			if err := os.Mkdir(filepath.Join(destination, "mode-dir"), 0o755); err != nil {
				t.Fatal(err)
			}
			addIgnoredSnapshotBuildOutputs(t, destination)
			run := func(archiveEnv, idEnv string) ([]byte, error) {
				cmd := exec.Command("bash", "devtools/scripts/print_tree_id.sh")
				cmd.Dir = destination
				cmd.Env = append(os.Environ(), "BID754_SNAPSHOT_ARCHIVE="+archiveEnv, "BID754_SNAPSHOT_ID="+idEnv)
				return cmd.CombinedOutput()
			}
			if out, err := run(archive, identity); err != nil || strings.TrimSpace(string(out)) != tree {
				t.Fatalf("container wrapper did not preserve archived tree ID: want %s, got %s, %v", tree, out, err)
			}
			for _, env := range [][2]string{{archive, ""}, {"", identity}, {archive, strings.Repeat("0", 64)}} {
				if out, err := run(env[0], env[1]); err == nil {
					t.Fatalf("container wrapper accepted missing/mismatched archive identity: %s", out)
				}
			}
			writeWorktreeFile(t, destination, "README.md", "corruption after build")
			if out, err := run(archive, identity); err == nil || !strings.Contains(string(out), "source entry mismatch: README.md") {
				t.Fatalf("container wrapper accepted source corruption: %s, %v", out, err)
			}
		})
	}
}
