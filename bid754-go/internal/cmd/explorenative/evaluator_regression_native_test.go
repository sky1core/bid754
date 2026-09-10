//go:build cgo && bid754_native

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalprobe"
	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

func TestEvaluatorRegressionOverlays(t *testing.T) {
	source, err := filepath.Abs("reference_native.go")
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []string{"baseline", "go-panic", "c-panic", "model-panic", "model-error", "comparison-panic", "go-flags-change", "go-flags-fixed", "c-flags-change", "c-flags-fixed"} {
		t.Run(fixture, func(t *testing.T) {
			content := string(original)
			switch fixture {
			case "go-panic":
				content = strings.Replace(content, "gB, gF := goEval(w, c.Op, tup, m.native)", "gB, gF := goEval(w, c.Op, tup, m.native); panic(\"target fixture\")", 1)
			case "c-panic":
				content = strings.Replace(content, "cB, cF := cEval(w, c.Op, tup, m.native)", "cB, cF := cEval(w, c.Op, tup, m.native); panic(\"target fixture\")", 1)
			case "model-panic":
				content = strings.Replace(content, "return decimalref.Evaluate(c)", "panic(\"model fixture\")", 1)
			case "model-error":
				content = strings.Replace(content, "return decimalref.Evaluate(c)", "return decimalref.Result{}, fmt.Errorf(\"model fixture\")", 1)
			case "comparison-panic":
				content = strings.Replace(content, "oc.disc.cgoValue =", "panic(\"comparison fixture\"); oc.disc.cgoValue =", 1)
			case "go-flags-change", "go-flags-fixed", "c-flags-change", "c-flags-fixed":
				leg := "g"
				if strings.HasPrefix(fixture, "c-") {
					leg = "c"
				}
				needle := leg + "B, " + leg + "F := "
				start := strings.Index(content, needle)
				if start < 0 {
					t.Fatal("missing evaluator call")
				}
				end := start + strings.Index(content[start:], "\n")
				mutation := "; " + leg + "F |= flagInvalid"
				if strings.HasSuffix(fixture, "change") {
					mutation = "; if tup[0].lo == 0x3280000a && tup[1].lo == 0x32800014 { " + leg + "F |= flagInvalid } else { " + leg + "F |= flagUnderflow }"
				}
				content = content[:end] + mutation + content[end:]
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, "reference_native.go")
			if err := os.WriteFile(replacement, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{source: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlayPath := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("go", "test", "-overlay", overlayPath, "-tags", "bid754_native", "-run", "^TestEvaluatorRegressionFixture$", "-count=1", "-v", ".")
			cmd.Env = append(os.Environ(), "EXPLORENATIVE_REGRESSION_FIXTURE="+fixture)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("fixture %s: %v\n%s", fixture, err, out)
			}
			if !strings.Contains(string(out), "--- PASS: TestEvaluatorRegressionFixture") {
				t.Fatalf("fixture did not execute: %s", out)
			}
		})
	}
}

func TestEvaluatorRegressionFixture(t *testing.T) {
	fixture := os.Getenv("EXPLORENATIVE_REGRESSION_FIXTURE")
	if fixture == "" {
		t.Skip("overlay child only")
	}
	sample := decimalprobe.Sample{Version: 1, Family: "uniform-finite", Case: decimalref.Case{Width: 32, Op: "add", Mode: "nearest_even", Operands: []string{"3280000a", "32800014"}}}
	oc, err := evaluateCase(sample.Case)
	if strings.Contains(fixture, "flags-") {
		if err != nil || oc.oracleErr != nil {
			t.Fatalf("evaluation: %v %v", err, oc.oracleErr)
		}
		rc := &refCampaign{shrinkAttempts: 64}
		f := &findingRec{}
		rc.attachShrink(f, sample, oc)
		if f.Shrink.Failed || f.Shrink.Attempts <= 1 {
			t.Fatalf("shrink did not test candidates: %+v", f.Shrink)
		}
		if strings.HasSuffix(fixture, "change") {
			if f.Shrunk != nil || f.Shrink.Accepted != 0 {
				t.Fatalf("invalid changed to underflow: %+v", f)
			}
		} else {
			if f.Shrunk == nil || f.Shrink.Accepted == 0 {
				t.Fatalf("stable flag defect did not shrink: %+v", f)
			}
			shrunk, err := evaluateCase(f.Shrunk.Case)
			if err != nil || shrunk.flagDifferences() != oc.flagDifferences() {
				t.Fatalf("shrunk flags changed: %+v %v", shrunk, err)
			}
		}
		return
	}
	oracleError, executionError, targetCompleted, comparisons := 0, 0, 1, 1
	switch fixture {
	case "model-panic", "model-error":
		oracleError, comparisons = 1, 0
	case "go-panic", "c-panic":
		executionError, targetCompleted, comparisons = 1, 0, 0
	case "comparison-panic":
		executionError, comparisons = 1, 0
	}
	if (err != nil) != (executionError == 1) || (oc.oracleErr != nil) != (oracleError == 1) || oc.targetCompleted != (targetCompleted == 1) || oc.compared != (comparisons == 1) {
		t.Fatalf("wrong evaluator stages: %+v err=%v", oc, err)
	}
	for _, campaign := range []string{campaignUniform, campaignRel, "replay"} {
		t.Run(campaign, func(t *testing.T) {
			dir := t.TempDir()
			output, err := os.Create(filepath.Join(dir, "output.jsonl"))
			if err != nil {
				t.Fatal(err)
			}
			saved := os.Stdout
			os.Stdout = output
			var runErr error
			func() {
				defer func() { os.Stdout = saved }()
				if campaign == "replay" {
					line, err := json.Marshal(findingRec{Type: "sample", Sample: toSampleRec(sample)})
					if err != nil {
						t.Fatal(err)
					}
					path := filepath.Join(dir, "replay.jsonl")
					if err := os.WriteFile(path, line, 0600); err != nil {
						t.Fatal(err)
					}
					runErr = runReplay(path, "")
				} else {
					runErr = runReference(campaign, "1", 1, "add", "32", "nearest_even", 0, "")
				}
			}()
			if err := output.Close(); err != nil {
				t.Fatal(err)
			}
			if (runErr != nil) != (oracleError+executionError > 0) {
				t.Fatalf("wrong run status: %v", runErr)
			}
			data, err := os.ReadFile(output.Name())
			if err != nil {
				t.Fatal(err)
			}
			var counters, errors, summaries, findings int
			for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
				var kind struct{ Type string }
				if err := json.Unmarshal([]byte(line), &kind); err != nil {
					t.Fatal(err)
				}
				switch kind.Type {
				case "execution_error":
					errors++
					var r executionErrorRecord
					if err := json.Unmarshal([]byte(line), &r); err != nil {
						t.Fatal(err)
					}
					if r.Stage != strings.TrimSuffix(fixture, "-panic") || r.Error == "" || r.Original == nil || r.Op != "add" || r.Width != 32 {
						t.Fatalf("incomplete execution diagnostic: %s", line)
					}
				case "finding":
					findings++
				case "counters", "summary":
					var r summaryRecordV2
					if err := json.Unmarshal([]byte(line), &r); err != nil {
						t.Fatal(err)
					}
					n := r.Generated
					if n == 0 || r.Reached != n || r.GenerateError != 0 || r.OracleCompleted != n*(1-oracleError) || r.OracleError != n*oracleError || r.TargetCompleted != n*targetCompleted || r.ExecutionError != n*executionError || r.Comparisons != n*comparisons {
						t.Fatalf("wrong counts: %s", line)
					}
					if kind.Type == "counters" {
						counters++
					} else {
						summaries++
						if n != counters {
							t.Fatalf("summary lost cases: %s", line)
						}
					}
				}
			}
			if counters == 0 || summaries != 1 || errors != counters*executionError || findings != counters*oracleError {
				t.Fatalf("lost records: %s", data)
			}
		})
	}
}

func TestFlagDifferenceKeepsDirection(t *testing.T) {
	ref, err := decimalref.Evaluate(decimalref.Case{Width: 32, Op: "add", Mode: "nearest_even", Operands: []string{"32800001", "32800002"}})
	if err != nil {
		t.Fatal(err)
	}
	bits, err := decimalref.Encode(32, ref.Value)
	if err != nil {
		t.Fatal(err)
	}
	missing, spurious := caseOutcome{ref: ref}, caseOutcome{ref: ref}
	missing.ref.Flags = flagInvalid
	spurious.gFlags = flagInvalid
	missing.disc.refGoValue, missing.disc.refGoFlags = refLegDiff(32, missing.ref, bits, missing.gFlags)
	spurious.disc.refGoValue, spurious.disc.refGoFlags = refLegDiff(32, spurious.ref, bits, spurious.gFlags)
	if !sameClasses(missing.disc.classes(), spurious.disc.classes()) {
		t.Fatal("fixture classes differ")
	}
	if missing.ref.Flags^missing.gFlags != spurious.ref.Flags^spurious.gFlags {
		t.Fatal("fixture XOR differs")
	}
	if missing.flagDifferences()[0] == spurious.flagDifferences()[0] {
		t.Fatal("missing invalid became spurious invalid")
	}
}

func TestEvaluationCountsRejectUnfinishedStages(t *testing.T) {
	sample := decimalprobe.Sample{Version: 1, Family: "uniform-finite", Case: decimalref.Case{Width: 32, Op: "add", Mode: "nearest_even", Operands: []string{"32800001", "32800002"}}}
	for _, mutate := range []func(*counterSet){
		func(cs *counterSet) { cs.Comparisons = 0 },
		func(cs *counterSet) { cs.TargetCompleted = 0 },
		func(cs *counterSet) { cs.ExecutionError = 1 },
		func(cs *counterSet) { cs.Generated++ },
	} {
		cs := newCounterSet("add")
		cs.Generated = 1
		rc := &refCampaign{emit: func(any) error { return nil }}
		if _, err := rc.evaluateSample(cs, sample); err != nil {
			t.Fatal(err)
		}
		if err := cs.reconcileEvaluation(); err != nil {
			t.Fatal(err)
		}
		mutate(cs)
		if err := cs.reconcileEvaluation(); err == nil {
			t.Fatal("accepted inconsistent evaluation counts")
		}
	}
}

func TestEvaluationInputErrorsRemainDistinctFromModelErrors(t *testing.T) {
	for _, mode := range []string{"nearest_even", "invalid-mode"} {
		sample := decimalprobe.Sample{Version: 1, Family: "uniform-finite", Case: decimalref.Case{Width: 32, Op: "add", Mode: mode, Operands: []string{"bad-raw", "32800002"}}}
		cs := newCounterSet("add")
		cs.Generated = 1
		errors := 0
		rc := &refCampaign{emit: func(v any) error {
			if r, ok := v.(executionErrorRecord); ok {
				errors++
				if r.Stage != "input" || r.Error == "" {
					t.Fatalf("wrong input error: %+v", r)
				}
			}
			return nil
		}}
		if _, err := rc.evaluateSample(cs, sample); err != nil {
			t.Fatal(err)
		}
		if err := cs.reconcileEvaluation(); err != nil {
			t.Fatal(err)
		}
		if errors != 1 || cs.Reached != 1 || cs.OracleCompleted != 0 || cs.OracleError != 1 || cs.ExecutionError != 1 || cs.TargetCompleted != 0 || cs.Comparisons != 0 {
			t.Fatalf("input/model failures conflated: %+v", cs)
		}
	}
}
