package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type relayRecord struct {
	Type            string   `json:"type"`
	Tool            string   `json:"tool"`
	FormatVersion   int      `json:"format_version"`
	Campaign        string   `json:"campaign"`
	Target          string   `json:"target"`
	Family          string   `json:"family"`
	Op              string   `json:"op"`
	Width           int      `json:"width"`
	Mode            string   `json:"mode"`
	CasesPerTarget  int      `json:"cases_per_target"`
	Ops             []string `json:"ops"`
	Widths          []int    `json:"widths"`
	Modes           []string `json:"modes"`
	Families        []string `json:"families"`
	Targets         int      `json:"targets"`
	Cases           int      `json:"cases"`
	Comparisons     int      `json:"comparisons"`
	Mismatches      int      `json:"mismatches"`
	Generated       int      `json:"generated"`
	Reached         int      `json:"reached"`
	OracleCompleted int      `json:"oracle_completed"`
	TargetCompleted int      `json:"target_completed"`
	ExecutionError  int      `json:"execution_error"`
	GenerateError   int      `json:"generate_error"`
	OracleError     int      `json:"oracle_error"`
	ShrinkError     int      `json:"shrink_error"`
	Findings        int      `json:"findings"`
	ElapsedMS       int64    `json:"elapsed_ms"`
	Classes         []string `json:"classes"`
	Error           string   `json:"error"`
	Shrink          *struct {
		Failed bool `json:"failed"`
	} `json:"shrink"`
	X       string `json:"x"`
	Y       string `json:"y"`
	Z       string `json:"z"`
	CBits   string `json:"c_bits"`
	CFlags  string `json:"c_flags"`
	GoBits  string `json:"go_bits"`
	GoFlags string `json:"go_flags"`
}

func relayRecords(in io.Reader, file, console io.Writer) (int, bool, error) {
	checkedConsole := consoleWriter(console)
	console = checkedConsole
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	var cfg, tot relayRecord
	configured, summary := false, false
	findings, shrinkErrors, genErrors, pendingFindings, pendingOracle := 0, 0, 0, 0, 0
	pendingExecution := 0
	seen := map[string]bool{}
	fail := func(err error) (int, bool, error) { return findings, summary, err }
	for scanner.Scan() {
		line := scanner.Bytes()
		if _, err := file.Write(append(append([]byte{}, line...), '\n')); err != nil {
			return fail(err)
		}
		if summary {
			return fail(fmt.Errorf("record after summary"))
		}
		if err := uniqueJSON(json.NewDecoder(bytes.NewReader(line))); err != nil {
			return fail(err)
		}
		var rec relayRecord
		if err := json.Unmarshal(line, &rec); err != nil {
			return fail(err)
		}
		if rec.Type == "provenance" {
			if configured || seen["provenance"] {
				return fail(fmt.Errorf("duplicate/misplaced provenance"))
			}
			seen["provenance"] = true
			continue
		}
		if rec.Type == "config" {
			if configured || rec.Tool != "explorediff" || (rec.Campaign == "" && rec.FormatVersion != 1) || (rec.Campaign != "" && rec.FormatVersion != 2) {
				return fail(fmt.Errorf("invalid/duplicate config"))
			}
			if rec.Campaign != "" && rec.Campaign != "relations" && rec.Campaign != "uniform-finite" && rec.Campaign != "replay" {
				return fail(fmt.Errorf("unknown campaign"))
			}
			if rec.Campaign != "replay" && (rec.CasesPerTarget <= 0 || len(rec.Widths) == 0 || len(rec.Modes) == 0 || len(rec.Ops) == 0) {
				return fail(fmt.Errorf("missing campaign dimensions"))
			}
			cfg = rec
			configured = true
			continue
		}
		if !configured {
			return fail(fmt.Errorf("record without config"))
		}
		if rec.ElapsedMS < 0 || rec.Targets < 0 || rec.Cases < 0 || rec.Comparisons < 0 || rec.Mismatches < 0 || rec.Generated < 0 || rec.Reached < 0 || rec.OracleCompleted < 0 || rec.GenerateError < 0 || rec.OracleError < 0 || rec.ShrinkError < 0 || rec.Findings < 0 {
			return fail(fmt.Errorf("negative time or counts"))
		}
		switch rec.Type {
		case "execution_error":
			if cfg.Campaign == "" || rec.Campaign != cfg.Campaign || rec.Error == "" {
				return fail(fmt.Errorf("invalid execution_error"))
			}
			pendingExecution++
			fmt.Fprintf(console, "execution_error: %s\n", rec.Error)
		case "finding":
			if cfg.Campaign == "" || rec.Campaign != cfg.Campaign || len(rec.Classes) == 0 {
				return fail(fmt.Errorf("invalid finding"))
			}
			findings++
			pendingFindings++
			for _, c := range rec.Classes {
				if c == "oracle-error" {
					pendingOracle++
				}
			}
			if rec.Shrink != nil && rec.Shrink.Failed {
				shrinkErrors++
			}
			if findings <= 20 {
				fmt.Fprintf(console, "FINDING %s %s/%s d%d/%s classes=%s\n", rec.Campaign, rec.Family, rec.Op, rec.Width, rec.Mode, strings.Join(rec.Classes, ","))
			}
		case "mismatch":
			if cfg.Campaign != "" {
				return fail(fmt.Errorf("mismatch in reference campaign"))
			}
			findings++
			pendingFindings++
			if findings <= 20 {
				fmt.Fprintf(console, "MISMATCH %s %s x=%s y=%s z=%s C=%s/%s go=%s/%s\n", rec.Target, rec.Mode, rec.X, rec.Y, rec.Z, rec.CBits, rec.CFlags, rec.GoBits, rec.GoFlags)
			}
		case "generate_error":
			if cfg.Campaign == "" || rec.Campaign != cfg.Campaign || rec.Error == "" {
				return fail(fmt.Errorf("invalid generate_error"))
			}
			genErrors++
			fmt.Fprintf(console, "generate_error: %s\n", rec.Error)
		case "counters":
			if rec.ExecutionError != pendingExecution || rec.TargetCompleted < 0 || rec.TargetCompleted > rec.Reached || rec.ExecutionError < 0 || rec.ExecutionError > rec.Reached || rec.TargetCompleted+rec.ExecutionError < rec.Reached || rec.Comparisons > rec.TargetCompleted || rec.Comparisons > rec.OracleCompleted || rec.Comparisons+rec.ExecutionError > rec.Reached || rec.Comparisons+rec.ExecutionError+rec.OracleError < rec.Reached {
				return fail(fmt.Errorf("impossible execution/comparison counts"))
			}
			if cfg.Campaign == "" || rec.Campaign != cfg.Campaign || rec.Generated <= 0 || rec.Reached < 0 || rec.OracleCompleted < 0 || rec.GenerateError < 0 || rec.OracleError < 0 || rec.Generated != rec.Reached+rec.GenerateError || rec.Reached != rec.OracleCompleted+rec.OracleError || pendingFindings > rec.Reached || pendingOracle != rec.OracleError || genErrors != rec.GenerateError {
				return fail(fmt.Errorf("impossible target counters"))
			}
			if cfg.Campaign == "replay" {
				if rec.Generated != 1 {
					return fail(fmt.Errorf("replay counters must describe one sample"))
				}
			} else {
				key := fmt.Sprintf("%s/%s/%d/%s", rec.Family, rec.Op, rec.Width, rec.Mode)
				if seen[key] || rec.Generated != cfg.CasesPerTarget || !has(cfg.Ops, rec.Op) || !has(cfg.Modes, rec.Mode) || !hasWidth(cfg.Widths, rec.Width) || (cfg.Campaign == "relations" && !has(cfg.Families, rec.Family)) || (cfg.Campaign == "uniform-finite" && rec.Family != "uniform-finite") {
					return fail(fmt.Errorf("unexpected/duplicate target counters"))
				}
				seen[key] = true
			}
			tot.Targets++
			tot.Generated += rec.Generated
			tot.Reached += rec.Reached
			tot.OracleCompleted += rec.OracleCompleted
			tot.GenerateError += rec.GenerateError
			tot.OracleError += rec.OracleError
			tot.TargetCompleted += rec.TargetCompleted
			tot.ExecutionError += rec.ExecutionError
			tot.Comparisons += rec.Comparisons
			pendingExecution = 0
			pendingFindings = 0
			pendingOracle = 0
			genErrors = 0
			fmt.Fprintf(console, "counters %-16s %-8s d%d/%s generated=%d reached=%d oracle_completed=%d gen_err=%d oracle_err=%d\n", rec.Family, rec.Op, rec.Width, rec.Mode, rec.Generated, rec.Reached, rec.OracleCompleted, rec.GenerateError, rec.OracleError)
		case "target_summary":
			if cfg.Campaign != "" || seen[rec.Target] || rec.Cases != cfg.CasesPerTarget || rec.Comparisons != rec.Cases*len(cfg.Modes) || rec.Mismatches != pendingFindings || rec.Mismatches > rec.Comparisons {
				return fail(fmt.Errorf("invalid/duplicate legacy target"))
			}
			valid := false
			for _, w := range cfg.Widths {
				for _, op := range cfg.Ops {
					if rec.Target == fmt.Sprintf("d%d/%s", w, op) {
						valid = true
					}
				}
			}
			if !valid {
				return fail(fmt.Errorf("unknown target"))
			}
			seen[rec.Target] = true
			tot.Targets++
			tot.Cases += rec.Cases
			tot.Comparisons += rec.Comparisons
			tot.Mismatches += rec.Mismatches
			pendingFindings = 0
			fmt.Fprintf(console, "target %-14s cases=%d comparisons=%d mismatches=%d elapsed=%dms\n", rec.Target, rec.Cases, rec.Comparisons, rec.Mismatches, rec.ElapsedMS)
		case "summary":
			if rec.Campaign != cfg.Campaign || rec.Targets <= 0 || rec.Targets != tot.Targets || rec.Comparisons != tot.Comparisons || pendingFindings != 0 || genErrors != 0 || pendingExecution != 0 {
				return fail(fmt.Errorf("summary totals do not reconcile"))
			}
			if cfg.Campaign != "replay" {
				n := len(cfg.Ops) * len(cfg.Widths)
				if cfg.Campaign != "" {
					if cfg.Campaign == "relations" {
						n = len(cfg.Families) * len(cfg.Widths)
					}
					n *= len(cfg.Modes)
				}
				if n != rec.Targets {
					return fail(fmt.Errorf("summary missing configured targets"))
				}
			}
			if cfg.Campaign == "" {
				if rec.Cases != tot.Cases || rec.Mismatches != tot.Mismatches || rec.Mismatches != findings {
					return fail(fmt.Errorf("legacy summary mismatch"))
				}
				fmt.Fprintf(console, "TOTAL targets=%d cases=%d comparisons=%d mismatches=%d\n", rec.Targets, rec.Cases, rec.Comparisons, rec.Mismatches)
			} else {
				if rec.Generated != tot.Generated || rec.Reached != tot.Reached || rec.OracleCompleted != tot.OracleCompleted || rec.GenerateError != tot.GenerateError || rec.OracleError != tot.OracleError || rec.Findings != findings || rec.ShrinkError != shrinkErrors || rec.TargetCompleted != tot.TargetCompleted || rec.ExecutionError != tot.ExecutionError {
					return fail(fmt.Errorf("reference summary mismatch"))
				}
				fmt.Fprintf(console, "TOTAL campaign=%s targets=%d generated=%d reached=%d oracle_completed=%d gen_err=%d oracle_err=%d findings=%d\n", rec.Campaign, rec.Targets, rec.Generated, rec.Reached, rec.OracleCompleted, rec.GenerateError, rec.OracleError, rec.Findings)
			}
			summary = true
		default:
			return fail(fmt.Errorf("unknown subprocess record %q", rec.Type))
		}
	}
	if err := scanner.Err(); err != nil {
		return fail(err)
	}
	if !summary {
		return fail(fmt.Errorf("missing summary"))
	}
	if tot.GenerateError+tot.OracleError+tot.ExecutionError+shrinkErrors > 0 {
		return fail(fmt.Errorf("campaign execution failed: generate_error=%d oracle_error=%d execution_error=%d shrink_error=%d", tot.GenerateError, tot.OracleError, tot.ExecutionError, shrinkErrors))
	}
	return findings, true, checkedConsole.err
}

func has(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}
func hasWidth(values []int, want int) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

func uniqueJSON(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := tok.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for dec.More() {
			k, err := dec.Token()
			if err != nil {
				return err
			}
			key, ok := k.(string)
			key = strings.ToLower(strings.ToUpper(key))
			if !ok || seen[key] {
				return fmt.Errorf("duplicate JSON key")
			}
			seen[key] = true
			if err := uniqueJSON(dec); err != nil {
				return err
			}
		}
	case '[':
		for dec.More() {
			if err := uniqueJSON(dec); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("invalid JSON delimiter")
	}
	_, err = dec.Token()
	return err
}
