package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
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
		`{"type":"config","tool":"explorediff","format_version":1,"cases_per_target":10,"ops":["fma"],"widths":[128],"modes":["nearest_even","toward_negative","toward_positive","toward_zero","nearest_away"],"seed":"42"}`,
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

// TestRelayCountsFindings checks the new-campaign records: a finding is a
// counterexample and the campaign summary is recognized.
func TestRelayCountsFindings(t *testing.T) {
	stream := strings.Join([]string{
		`{"type":"config","tool":"explorediff","campaign":"relations","format_version":2,"cases_per_target":5,"ops":["mul"],"widths":[64],"modes":["nearest_even"],"families":["mul_tie"]}`,
		`{"type":"finding","campaign":"relations","family":"mul_tie","op":"mul","width":64,"mode":"nearest_even","classes":["reference-go:flags","c-go:value"]}`,
		`{"type":"counters","campaign":"relations","family":"mul_tie","op":"mul","width":64,"mode":"nearest_even","generated":5,"reached":5,"oracle_completed":5,"target_completed":5,"comparisons":5}`,
		`{"type":"summary","campaign":"relations","targets":1,"generated":5,"reached":5,"oracle_completed":5,"target_completed":5,"comparisons":5,"findings":1}`,
	}, "\n") + "\n"

	file, err := os.Create(filepath.Join(t.TempDir(), "findings.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var console bytes.Buffer
	counterexamples, sawSummary, err := relayRecords(strings.NewReader(stream), file, &console)
	if err != nil || counterexamples != 1 || !sawSummary {
		t.Fatalf("relayRecords = %d %v %v", counterexamples, sawSummary, err)
	}
	out := console.String()
	for _, want := range []string{"FINDING relations mul_tie/mul d64/nearest_even", "reference-go:flags,c-go:value",
		"counters mul_tie", "TOTAL campaign=relations", "findings=1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("console missing %q:\n%s", want, out)
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
	if err == nil || sawSummary {
		t.Fatalf("truncated stream: sawSummary=%v err=%v", sawSummary, err)
	}
}

func TestDriverPreflightRejectsWithoutOutput(t *testing.T) {
	base := config{repo: filepath.Join(t.TempDir(), "missing-repo"), campaign: "relations", cases: 1, bias: 0.25, ops: "add", widths: "32", modes: "nearest_even"}
	for name, change := range map[string]func(*config){
		"sqrt":             func(c *config) { c.ops = "sqrt" },
		"zero cases":       func(c *config) { c.cases = 0 },
		"negative shrink":  func(c *config) { c.shrinkAttempts = -1 },
		"duplicate ops":    func(c *config) { c.ops = "add,add" },
		"duplicate widths": func(c *config) { c.widths = "32,32" },
		"unknown mode":     func(c *config) { c.modes = "odd" },
	} {
		t.Run(name, func(t *testing.T) {
			cfg := base
			cfg.out = filepath.Join(t.TempDir(), "new", "out.jsonl")
			change(&cfg)
			if code, err := run(cfg); code != 1 || err == nil || strings.Contains(err.Error(), "checkout") {
				t.Fatalf("validation occurred too late: %d %v", code, err)
			}
			if _, err := os.Stat(filepath.Dir(cfg.out)); !os.IsNotExist(err) {
				t.Fatalf("output side effect: %v", err)
			}
		})
	}
	for _, campaign := range []string{"legacy", "relations", "uniform-finite"} {
		cfg := base
		cfg.campaign = campaign
		cfg.ops = ""
		if err := resolveConfig(&cfg); err != nil {
			t.Fatal(err)
		}
		want := 6
		if campaign == "legacy" {
			want = 7
		}
		if len(strings.Split(cfg.ops, ",")) != want {
			t.Fatal(cfg.ops)
		}
	}
}

func TestOutputCannotOverwriteReplayOrExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.jsonl")
	original := []byte("preserve me\n")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	for _, replay := range []string{"", path} {
		cfg := config{campaign: "relations", cases: 1, ops: "add", widths: "32", modes: "nearest_even", replay: replay, out: path}
		if code, err := run(cfg); code != 1 || err == nil || !strings.Contains(err.Error(), "already exists") {
			t.Fatalf("%d %v", code, err)
		}
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(got, original) {
			t.Fatal("existing file changed")
		}
	}
}

func TestFlagDuplicatesAndUnknowns(t *testing.T) {
	for _, args := range [][]string{{"-ops", "add", "--ops=sub"}, {"-cases=1", "-cases=2"}, {"-typo", "1"}, {"argument"}, {"--", "-ops", "add"}} {
		if err := validateFlagArgs(args); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}

func TestShellQuotePaths(t *testing.T) {
	path := "/tmp/a 'quote' $(echo forbidden) `echo forbidden`\nfile"
	out, err := exec.Command("sh", "-c", "printf '%s' "+shellQuote(path)).Output()
	if err != nil || string(out) != path {
		t.Fatalf("shell path changed: %q %v", out, err)
	}
	cfg := config{replay: path}
	if cmd := reproCommand(cfg, "/tmp/repo space", 1); !strings.Contains(cmd, "-replay "+shellQuote(path)) || !strings.Contains(cmd, "cd '/tmp/repo space/devtools'") {
		t.Fatal(cmd)
	}
}

func TestRelayRejectsInvalidTotals(t *testing.T) {
	configLine := `{"type":"config","tool":"explorediff","campaign":"replay","format_version":2}` + "\n"
	counters := `{"type":"counters","campaign":"replay","generated":1,"reached":1,"oracle_completed":1,"target_completed":1,"comparisons":1}` + "\n"
	summary := `{"type":"summary","campaign":"replay","targets":1,"generated":1,"reached":1,"oracle_completed":1,"target_completed":1,"comparisons":1}` + "\n"
	valid := configLine + counters + summary
	if _, _, err := relayRecords(strings.NewReader(valid), io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	for name, stream := range map[string]string{
		"arbitrary summary":         `{"type":"summary"}`,
		"missing":                   configLine + counters,
		"duplicate":                 valid + summary,
		"count mismatch":            strings.Replace(valid, `"comparisons":1`, `"comparisons":2`, 1),
		"impossible":                strings.Replace(valid, `"oracle_completed":1`, `"oracle_completed":2`, 1),
		"unknown type":              configLine + `{"type":"anything"}`,
		"truncated":                 valid[:len(valid)-3],
		"duplicate JSON":            strings.Replace(valid, `"targets":1`, `"targets":2,"targets":1`, 1),
		"unreported finding":        configLine + counters + strings.Replace(summary, `"comparisons":1`, `"comparisons":1,"findings":1`, 1),
		"case alias JSON":           strings.Replace(valid, `"targets":1`, `"targets":2,"Targets":1`, 1),
		"missing target completion": strings.Replace(valid, `"target_completed":1`, `"target_completed":0`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := relayRecords(strings.NewReader(stream), io.Discard, io.Discard); err == nil {
				t.Fatal("accepted invalid stream")
			}
		})
	}
	for _, stream := range []string{
		configLine + `{"type":"generate_error","campaign":"replay","error":"failed"}` + "\n" + `{"type":"counters","campaign":"replay","generated":1,"generate_error":1}` + "\n" + `{"type":"summary","campaign":"replay","targets":1,"generated":1,"generate_error":1}` + "\n",
		configLine + `{"type":"finding","campaign":"replay","classes":["oracle-error"]}` + "\n" + `{"type":"counters","campaign":"replay","generated":1,"reached":1,"oracle_error":1,"target_completed":1}` + "\n" + `{"type":"summary","campaign":"replay","targets":1,"generated":1,"reached":1,"oracle_error":1,"target_completed":1,"comparisons":0,"findings":1}` + "\n",
		configLine + `{"type":"execution_error","campaign":"replay","error":"go panic"}` + "\n" + `{"type":"counters","campaign":"replay","generated":1,"reached":1,"oracle_completed":1,"execution_error":1}` + "\n" + `{"type":"summary","campaign":"replay","targets":1,"generated":1,"reached":1,"oracle_completed":1,"execution_error":1}` + "\n",
	} {
		if _, saw, err := relayRecords(strings.NewReader(stream), io.Discard, io.Discard); err == nil || !saw {
			t.Fatalf("failed campaign must have auditable summary and error: %v %v", saw, err)
		}
	}
}

func TestPublishRecordsExclusiveAndFlushErrors(t *testing.T) {
	for _, test := range []string{"new", "existing", "closed"} {
		t.Run(test, func(t *testing.T) {
			dir := t.TempDir()
			spool, err := os.CreateTemp(dir, "spool-")
			if err != nil {
				t.Fatal(err)
			}
			buf := bufio.NewWriter(spool)
			if _, err := buf.WriteString("auditable records\n"); err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(dir, "output.jsonl")
			if test == "existing" {
				if err := os.WriteFile(out, []byte("preserved"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if test == "closed" {
				if err := spool.Close(); err != nil {
					t.Fatal(err)
				}
			}
			err = publishRecords(buf, spool, out)
			switch test {
			case "new":
				if err != nil {
					t.Fatal(err)
				}
				got, err := os.ReadFile(out)
				if err != nil || string(got) != "auditable records\n" {
					t.Fatalf("lost records: %q %v", got, err)
				}
			case "existing":
				if err == nil {
					t.Fatal("overwrote existing output")
				}
				got, err := os.ReadFile(out)
				if err != nil || string(got) != "preserved" {
					t.Fatalf("changed existing output: %q %v", got, err)
				}
			case "closed":
				if err == nil {
					t.Fatal("lost flush/seek/close errors")
				}
				if _, err := os.Stat(out); !os.IsNotExist(err) {
					t.Fatal("published failed spool")
				}
			}
		})
	}
}
