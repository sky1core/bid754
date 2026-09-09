package platformdigest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyDigestEvidence(t *testing.T) {
	const tree = "0123456789012345678901234567890123456789"
	const sum = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	record := func(platform string) string {
		parts := strings.Split(platform, "/")
		return fmt.Sprintf("PLATFORM-DIGEST-TREE %s\nPLATFORM-DIGEST goos=%s goarch=%s cases=3668 sha256=%s\n", tree, parts[0], parts[1], sum)
	}
	tests := []struct {
		name   string
		mutate func(map[string]string)
		want   string
	}{
		{name: "three platforms agree"},
		{name: "missing platform", mutate: func(files map[string]string) { delete(files, "digest_linux_arm64.txt") }, want: "missing required platforms"},
		{name: "stale matching evidence", mutate: func(files map[string]string) {
			for name, text := range files {
				files[name] = strings.ReplaceAll(text, tree, strings.Repeat("a", 40))
			}
		}, want: "tree mismatch"},
		{name: "dirty evidence", mutate: func(files map[string]string) {
			files["digest_linux_arm64.txt"] = strings.ReplaceAll(files["digest_linux_arm64.txt"], tree, tree+"-dirty")
		}, want: "tree mismatch"},
		{name: "value mismatch", mutate: func(files map[string]string) {
			files["digest_linux_arm64.txt"] = strings.ReplaceAll(files["digest_linux_arm64.txt"], sum, strings.Repeat("b", 64))
		}, want: "digest mismatch"},
		{name: "count mismatch", mutate: func(files map[string]string) {
			files["digest_linux_arm64.txt"] = strings.ReplaceAll(files["digest_linux_arm64.txt"], "cases=3668", "cases=3667")
		}, want: "case-count"},
		{name: "malformed hash", mutate: func(files map[string]string) {
			files["digest_linux_arm64.txt"] = strings.ReplaceAll(files["digest_linux_arm64.txt"], sum, "arbitrary-text")
		}, want: "invalid digest record"},
		{name: "empty corpus", mutate: func(files map[string]string) {
			for name, text := range files {
				files[name] = strings.ReplaceAll(text, "cases=3668", "cases=0")
			}
		}, want: "invalid digest record"},
		{name: "duplicate records", mutate: func(files map[string]string) {
			files["digest_linux_arm64.txt"] += files["digest_linux_arm64.txt"]
		}, want: "exactly one"},
		{name: "truncated record", mutate: func(files map[string]string) {
			files["digest_linux_arm64.txt"] = "PLATFORM-DIGEST-TREE " + tree + "\n"
		}, want: "exactly one"},
		{name: "renamed platform", mutate: func(files map[string]string) {
			files["digest_linux_arm64.txt"] = record("linux/amd64")
		}, want: "filename disagrees"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			files := map[string]string{}
			for _, platform := range []string{"darwin/arm64", "linux/amd64", "linux/arm64"} {
				files["digest_"+strings.ReplaceAll(platform, "/", "_")+".txt"] = record(platform)
			}
			if tc.mutate != nil {
				tc.mutate(files)
			}
			for name, text := range files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command("bash", "../../scripts/verify_digest.sh", "--results-dir", dir, "--expected-tree", tree,
				"--require-platform", "darwin/arm64", "--require-platform", "linux/amd64", "--require-platform", "linux/arm64")
			output, err := cmd.CombinedOutput()
			if tc.want == "" {
				if err != nil || !strings.Contains(string(output), "3 platforms agree") {
					t.Fatalf("valid evidence rejected: %v\n%s", err, output)
				}
			} else if err == nil || !strings.Contains(string(output), tc.want) {
				t.Fatalf("want failure containing %q, got %v\n%s", tc.want, err, output)
			}
		})
	}
}
