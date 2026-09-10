//go:build cgo && bid754_native

package main

import (
	"encoding/json"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalprobe"
)

func TestNativeCLIEntry(t *testing.T) {
	if os.Getenv("EXPLORENATIVE_CLI_CHILD") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"explorenative"}, os.Args[i+1:]...)
			break
		}
	}
	flag.CommandLine = flag.NewFlagSet("explorenative", flag.ExitOnError)
	main()
	os.Exit(0)
}

func nativeCLI(t *testing.T, args ...string) ([]byte, error) {
	t.Helper()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(binary, append([]string{"-test.run=^TestNativeCLIEntry$", "--"}, args...)...)
	cmd.Dir = filepath.Join("..", "..", "..")
	cmd.Env = append(os.Environ(), "EXPLORENATIVE_CLI_CHILD=1")
	return cmd.CombinedOutput()
}

func nativeProvenanceArgs(t *testing.T) []string {
	t.Helper()
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("python3", "-B", filepath.Join(root, "devtools/scripts/lib/source_snapshot.py"), "current-tree-id", "--root", root).CombinedOutput()
	if err != nil {
		t.Fatalf("source identity: %s %v", out, err)
	}
	source := strings.TrimSpace(string(out))
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"-source-id", source, "-commit", source}
	for name, path := range map[string]string{"binary-sha256": binary, "c-library-sha256": filepath.Join(root, "devtools/third_party/intel_dfp/lib/libbid.a"), "c-archive-sha256": filepath.Join(root, "devtools/third_party/intel_dfp/IntelRDFPMathLib20U4.tar.gz")} {
		hash, err := hashFile(path)
		if err != nil {
			t.Fatal(err)
		}
		args = append(args, "-"+name, hash)
	}
	return args
}

func TestNativeCLINegatives(t *testing.T) {
	base := []string{"-campaign", "relations", "-seed", "1", "-cases", "1", "-widths", "32", "-modes", "nearest_even"}
	for _, extra := range [][]string{{"-ops", "sqrt"}, {"-ops", "add,sqrt"}, {"-cases", "0"}, {"-ops", "add,add"}, {"-ops", "add", "-ops", "sub"}, {"-unknown", "1"}, {"-shrink-attempts", "-1"}, {"-ops", ""}} {
		out, err := nativeCLI(t, append(append([]string{}, base...), extra...)...)
		if err == nil || strings.Contains(string(out), `"type":"summary"`) {
			t.Fatalf("accepted %v: %v %s", extra, err, out)
		}
	}
	for name, line := range map[string]string{"version": strings.Replace(sampleLine(""), `"version":1`, `"version":999`, 1), "false midpoint": `{"type":"sample","sample":{"version":1,"family":"add_below","case":{"width":32,"op":"add","mode":"nearest_even","operands":["328f4240","2f4c4b40"]}}}`, "empty": "", "truncated": `{"type":"sample"`, "duplicate": strings.Replace(sampleLine(""), `"version":1`, `"version":2,"version":1`, 1)} {
		path := filepath.Join(t.TempDir(), name+".jsonl")
		if err := os.WriteFile(path, []byte(line), 0600); err != nil {
			t.Fatal(err)
		}
		out, err := nativeCLI(t, "-replay", path)
		if err == nil || strings.Contains(string(out), `"type":"summary"`) {
			t.Fatalf("accepted %s: %v %s", name, err, out)
		}
	}
}

func TestNativeCLIReplayCountsAndProvenance(t *testing.T) {
	path := filepath.Join(t.TempDir(), "replay 'quoted'.jsonl")
	if err := os.WriteFile(path, []byte(sampleLine("")+"\n"+sampleLine("")), 0600); err != nil {
		t.Fatal(err)
	}
	provenance := nativeProvenanceArgs(t)
	out, err := nativeCLI(t, append([]string{"-replay", path}, provenance...)...)
	if err != nil {
		t.Fatalf("replay: %v %s", err, out)
	}
	counts, summaries := 0, 0
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		var rec struct {
			Type                                     string
			Generated, Reached, Comparisons, Targets int
			OracleCompleted                          int `json:"oracle_completed"`
		}
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatal(err)
		}
		if rec.Type == "counters" {
			counts++
			if rec.Generated != 1 || rec.Reached != 1 || rec.OracleCompleted != 1 {
				t.Fatalf("bad replay counters: %s", line)
			}
		}
		if rec.Type == "summary" {
			summaries++
			if rec.Targets != 2 || rec.Generated != 2 || rec.Reached != 2 || rec.OracleCompleted != 2 || rec.Comparisons != 2 {
				t.Fatalf("bad replay summary: %s", line)
			}
		}
	}
	if counts != 2 || summaries != 1 {
		t.Fatalf("records: counters=%d summaries=%d", counts, summaries)
	}
	for _, field := range []string{"-source-id", "-c-library-sha256", "-binary-sha256", "-c-archive-sha256"} {
		bad := append([]string{}, provenance...)
		for i, v := range bad {
			if v == field {
				bad[i+1] = strings.Repeat("0", 64)
				break
			}
		}
		out, err := nativeCLI(t, append([]string{"-replay", path}, bad...)...)
		if err == nil || !strings.Contains(string(out), "mismatch") {
			t.Fatalf("accepted wrong %s: %v %s", field, err, out)
		}
	}
	out, err = nativeCLI(t, "-replay", path)
	if err == nil || !strings.Contains(string(out), "provenance") {
		t.Fatalf("missing provenance: %v %s", err, out)
	}
}

func TestNativeCLIRelationsHonorOps(t *testing.T) {
	args := []string{"-campaign", "relations", "-seed", "754", "-cases", "1", "-ops", "sub", "-widths", "32", "-modes", "nearest_even", "-shrink-attempts", "0"}
	out, err := nativeCLI(t, append(args, nativeProvenanceArgs(t)...)...)
	if err != nil {
		t.Fatalf("relations: %v %s", err, out)
	}
	count := 0
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		var rec struct{ Type, Op string }
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatal(err)
		}
		if rec.Type == "counters" {
			count++
			if rec.Op != "sub" {
				t.Fatal(line)
			}
		}
	}
	if count == 0 {
		t.Fatal("no selected targets")
	}
}

func TestReplayFlushErrorPropagates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.jsonl")
	if err := os.WriteFile(path, []byte(sampleLine("")), 0600); err != nil {
		t.Fatal(err)
	}
	closed, err := os.Create(filepath.Join(t.TempDir(), "closed"))
	if err != nil {
		t.Fatal(err)
	}
	if err := closed.Close(); err != nil {
		t.Fatal(err)
	}
	saved := os.Stdout
	os.Stdout = closed
	defer func() { os.Stdout = saved }()
	if err := runReplay(path, ""); err == nil {
		t.Fatal("lost buffered flush failure")
	}
}

func TestGenerateErrorsAreAuditable(t *testing.T) {
	var counters *countersRecord
	errors := 0
	rc := &refCampaign{campaign: campaignUniform, emit: func(v any) error {
		switch rec := v.(type) {
		case *countersRecord:
			counters = rec
		case generateErrorRecord:
			errors++
		}
		return nil
	}}
	w, _ := widthByBits(32)
	tot, err := rc.runTarget(1, w, "nearest_even", "uniform-finite", "sqrt", func(hi, lo uint64, exp int32, neg bool) (decimalprobe.Sample, error) {
		return decimalprobe.Uniform("sqrt", 32, "nearest_even", hi, lo, exp, neg)
	}, 3)
	if err != nil || errors != 3 || counters == nil || tot.generated != 3 || tot.generateError != 3 || tot.reached != 0 || tot.oracleCompleted != 0 {
		t.Fatalf("lost generate errors: %+v counters=%+v errors=%d err=%v", tot, counters, errors, err)
	}
}
