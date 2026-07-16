//go:build arm64

package bidgo

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var bidLongSqrt128AssemblySink BID_UINT128

func TestBidLongSqrt128SeedDoesNotContractFMA(t *testing.T) {
	bidLongSqrt128AssemblySink, _ = Bid128Sqrt(
		BID_UINT128{lo: 2, hi: 0x3040000000000000},
		BID_ROUNDING_TO_NEAREST,
	)
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source path")
	}
	// Build a stable binary with its Go symbol table; the transient executable
	// used by `go test` is not guaranteed to remain readable by child tools.
	exe := filepath.Join(t.TempDir(), "bidgo.test")
	compile := exec.Command("go", "test", "-c", "-o", exe, ".")
	compile.Dir = filepath.Dir(sourceFile)
	if out, err := compile.CombinedOutput(); err != nil {
		t.Fatalf("compile bidgo test executable: %v\n%s", err, out)
	}

	const symbolPattern = "bid_long_sqrt128|noFmaMulAddF64"
	out, err := exec.Command("go", "tool", "objdump", "-s", symbolPattern, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("objdump %s: %v\n%s", symbolPattern, err, out)
	}
	assembly := string(out)
	if !strings.Contains(assembly, "bid_long_sqrt128") {
		t.Fatalf("objdump output does not contain bid_long_sqrt128:\n%s", assembly)
	}
	// The helper is normally inlined into bid_long_sqrt128 and therefore has no
	// standalone symbol. If a future compiler leaves it out of line, the same
	// objdump pattern includes that body in this scan as well.
	for _, fused := range []string{"FMADD", "FMSUB", "FNMADD", "FNMSUB"} {
		if strings.Contains(assembly, fused) {
			t.Errorf("sqrt128 seed path contains contracted %s instruction:\n%s", fused, assembly)
		}
	}
}
