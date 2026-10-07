package bidgo

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func auxIntelReadtestPath() string {
	if path, ok := os.LookupEnv("BID754_AUX_READTEST_FILE"); ok {
		return path
	}
	return "../../../devtools/third_party/intel_dfp/TESTS/readtest.in"
}

func openAuxIntelReadtest(t *testing.T) *os.File {
	t.Helper()
	file, err := os.Open(auxIntelReadtestPath())
	if errors.Is(err, os.ErrNotExist) {
		t.Skipf("Intel test file missing: %v", err)
	}
	if err != nil {
		t.Fatalf("open Intel test file: %v", err)
	}
	t.Cleanup(func() { file.Close() })
	return file
}

func finishAuxIntelReadtest(t *testing.T, scanner *bufio.Scanner, prefix string, passed, failed int) {
	t.Helper()
	if err := scanner.Err(); err != nil {
		t.Fatalf("%s: read Intel test file: %v", prefix, err)
	}
	if passed+failed == 0 {
		t.Fatalf("%s: no matching runnable rows", prefix)
	}
}

func TestAuxIntelReadtestRunnerRejectsBadFixtures(t *testing.T) {
	dir := t.TempDir()
	empty := filepath.Join(dir, "empty.in")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	unrelated := filepath.Join(dir, "unrelated.in")
	if err := os.WriteFile(unrelated, []byte("bid32_add 0 [0] [0] 00\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	denied := filepath.Join(dir, "denied.in")
	if err := os.WriteFile(denied, []byte("unused\n"), 0o000); err != nil {
		t.Fatal(err)
	}
	selector := `^TestBid64(RoundIntegral.*|NearbyInt|Modf|Sqrt|ToBid128|ToBid32)IntelReadtest$`
	for _, tc := range []struct {
		name, path, wantError string
		wantSkip              bool
	}{
		{"missing", filepath.Join(dir, "missing.in"), "", true},
		{"empty", empty, "no matching runnable rows", false},
		{"no matching rows", unrelated, "no matching runnable rows", false},
		{"read failure", dir, "read Intel test file", false},
		{"permission denied", denied, "open Intel test file", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "permission denied" && os.Geteuid() == 0 {
				t.Skip("root can read mode-000 files")
			}
			cmd := exec.Command("go", "test", "-run", selector, "-count=1", "-v", ".")
			cmd.Env = append(os.Environ(), "BID754_AUX_READTEST_FILE="+tc.path)
			out, err := cmd.CombinedOutput()
			if tc.wantSkip {
				if err != nil || strings.Count(string(out), "--- SKIP:") != 11 {
					t.Fatalf("missing file should skip all 11 runners: %v\n%s", err, out)
				}
				return
			}
			if err == nil || strings.Count(string(out), "--- FAIL:") != 11 || !strings.Contains(string(out), tc.wantError) {
				t.Fatalf("bad fixture should fail all 11 runners: %v\n%s", err, out)
			}
		})
	}
}
