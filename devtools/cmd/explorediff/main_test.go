package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveSeed(t *testing.T) {
	if _, _, err := resolveSeed("not-a-number", 1); err == nil {
		t.Fatal("junk seed text must be rejected")
	}
	seed, source, err := resolveSeed("754", 1)
	if err != nil || seed != 754 || source != "flag" {
		t.Fatalf("decimal seed: got %d %q %v", seed, source, err)
	}
	seed, source, err = resolveSeed("0x2f2", 1)
	if err != nil || seed != 754 || source != "flag" {
		t.Fatalf("hex seed: got %d %q %v", seed, source, err)
	}
	a, source, err := resolveSeed("", 1)
	if err != nil || source != "time" {
		t.Fatalf("time seed: got %q %v", source, err)
	}
	b, _, _ := resolveSeed("", 2)
	if a == b {
		t.Fatal("adjacent clock values must scramble to different seeds")
	}
}

func TestReproCommandCarriesEveryStreamParameter(t *testing.T) {
	cfg := config{cases: 123, bias: 0.5, ops: "fma", widths: "128", modes: "toward_zero"}
	cmd := reproCommand(cfg, "/repo", 42)
	for _, want := range []string{"-repo /repo", "-seed 42", "-cases 123", "-bias 0.5",
		"-ops fma", "-widths 128", "-modes toward_zero", "go build -o"} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("repro command %q missing %q", cmd, want)
		}
	}
	// `go run` folds the exit-3 counterexample signal into exit 1; the repro
	// command must run a built binary directly.
	if strings.Contains(cmd, "go run") {
		t.Fatalf("repro command must not use go run: %q", cmd)
	}
}

// TestRelayRecords drives the record relay with a canned subprocess stream
// and checks the findings file carries every line while the console carries
// the digest.
func TestRelayRecords(t *testing.T) {
	stream := strings.Join([]string{
		`{"type":"config","tool":"explorediff","seed":"42"}`,
		`{"type":"mismatch","target":"d128/fma","mode":"toward_zero","case_index":7,"x":"aa","y":"bb","z":"cc","c_bits":"11","c_flags":"00000028","go_bits":"22","go_flags":"00000028"}`,
		`{"type":"target_summary","target":"d128/fma","cases":10,"comparisons":50,"mismatches":1,"elapsed_ms":3}`,
		`{"type":"summary","targets":1,"cases":10,"comparisons":50,"mismatches":1,"elapsed_ms":3}`,
	}, "\n") + "\n"

	file, err := os.Create(filepath.Join(t.TempDir(), "findings.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var console bytes.Buffer
	mismatches, sawSummary, err := relayRecords(strings.NewReader(stream), file, &console)
	if err != nil || mismatches != 1 || !sawSummary {
		t.Fatalf("relayRecords = %d %v %v", mismatches, sawSummary, err)
	}
	written, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	if string(written) != stream {
		t.Fatalf("findings file does not carry the full stream:\n%s", written)
	}
	out := console.String()
	for _, want := range []string{"MISMATCH d128/fma", "x=aa y=bb z=cc", "C=11/00000028", "go=22/00000028",
		"target d128/fma", "TOTAL targets=1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("console output missing %q:\n%s", want, out)
		}
	}
}

func TestRelayRecordsRejectsTruncatedStream(t *testing.T) {
	stream := `{"type":"config","tool":"explorediff"}` + "\n"
	file, err := os.Create(filepath.Join(t.TempDir(), "findings.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var console bytes.Buffer
	_, sawSummary, err := relayRecords(strings.NewReader(stream), file, &console)
	if err != nil || sawSummary {
		t.Fatalf("truncated stream: sawSummary=%v err=%v", sawSummary, err)
	}
}
