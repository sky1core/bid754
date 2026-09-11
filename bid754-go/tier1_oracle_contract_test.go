package bid754

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestTier1MissingOracleReplay(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, self, "-test.run=^FuzzTier1BigDecimal/seed#0$", "-test.v")
	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "BID754_BIGDECIMAL_") || strings.HasPrefix(v, "BID754_TIER1_") {
			continue
		}
		cmd.Env = append(cmd.Env, v)
	}
	out, err := cmd.CombinedOutput()
	if err == nil || ctx.Err() != nil || !strings.Contains(string(out), "explicit Tier1 fuzzing/replay requires Java and generated Rust") {
		t.Fatalf("missing-oracle replay: %v\n%s", err, out)
	}
}

func TestTier1OracleProtocols(t *testing.T) {
	configured, err := tier1Configured()
	if err != nil {
		t.Fatal(err)
	}
	if !configured {
		t.Skip("requires Tier1 comparison processes")
	}
	for _, leg := range []struct {
		name, path     string
		args, requests []string
	}{
		{"rust", os.Getenv("BID754_TIER1_RUST"), nil, []string{
			`{"version":1,"case":{"width":32,"op":"from_int","mode":"nearest_even","operands":[],"param":"2147483648","target":32}}`,
			`{"version":1,"case":{"width":32,"op":"scaleb","mode":"nearest_even","operands":["32800001"],"param":"9223372036854775808","target":0}}`,
			`{"version":1,"case":{"width":32,"op":"quiet_less","mode":"nearest_even","operands":["32800001"],"param":"","target":0}}`,
			`{"version":1,"case":{"width":32,"op":"convert","mode":"nearest_even","operands":["32800001"],"param":"","target":32}}`,
			`{"version":1,"case":{"width":32,"op":"rem","mode":"nearest_even","operands":["032800001","32800002"],"param":"","target":0}}`,
			strings.Repeat(" ", 4097),
		}},
		{"java", os.Getenv("BID754_BIGDECIMAL_JAVA"), []string{"-cp", os.Getenv("BID754_BIGDECIMAL_CLASSES"), "Tier1BigDecimalProbe"}, []string{
			"1\t32\tfrom_int\tnearest_even\t2147483648\t32",
			"1\t32\tscaleb\tnearest_even\t9223372036854775808\t0\t1\t0",
			"1\t32\trem\tnearest_even\t\t0\t01\t0\t2\t0",
			"1\t32\tconvert\tnearest_even\t\t32\t1\t0",
			"1\t32\tquiet_less\tnearest_even\t\t0\t1\t91\t2\t0",
			strings.Repeat(" ", 1025),
		}},
	} {
		for i, request := range leg.requests {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			cmd := exec.CommandContext(ctx, leg.path, leg.args...)
			cmd.Stdin = strings.NewReader(request + "\n")
			out, err := cmd.CombinedOutput()
			timedOut := ctx.Err() != nil
			cancel()
			if err == nil || timedOut {
				t.Fatalf("%s input %d accepted or timed out: %v\n%s", leg.name, i, err, out)
			}
		}
	}
}
