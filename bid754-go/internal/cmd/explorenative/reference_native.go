//go:build cgo && bid754_native

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/sky1core/bid754/bid754-go/internal/decimalprobe"
	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

// This file is the cgo differential/oracle leg of the reference exploration
// campaigns (relations, uniform-finite) and of exact fixed-input JSONL replay.
// It runs the decimalref model, pinned Intel C, and the Go mechanical port on
// every campaign input and records their pairwise disagreements as findings.
// It is a discovery/audit tool, not a regular verification domain; it never
// claims verification closure (docs/TEST_GENERATION_SPEC.md Auxiliary Gates).

const pinnedCArchiveName = "IntelRDFPMathLib20U4.tar.gz"

// ---------- emitted config/summary/error records ----------

type configRecordV2 struct {
	Type                   string   `json:"type"`
	Tool                   string   `json:"tool"`
	Campaign               string   `json:"campaign"`
	FormatVersion          int      `json:"format_version"`
	GeneratorFormatVersion int      `json:"generator_format_version"`
	OracleFormatVersion    int      `json:"oracle_format_version"`
	Seed                   string   `json:"seed"`
	CasesPerTarget         int      `json:"cases_per_target"`
	Widths                 []int    `json:"widths"`
	Modes                  []string `json:"modes"`
	Ops                    []string `json:"ops,omitempty"`
	Families               []string `json:"families,omitempty"`
	ShrinkAttempts         int      `json:"shrink_attempts"`
	ShrinkEnabled          bool     `json:"shrink_enabled"`
	Commit                 string   `json:"commit,omitempty"`
	GoToolchain            string   `json:"go_toolchain"`
	CArchiveSHA256         string   `json:"c_archive_sha256"`
	GoOS                   string   `json:"go_os"`
	GoArch                 string   `json:"go_arch"`
}

type summaryRecordV2 struct {
	Type            string `json:"type"`
	Campaign        string `json:"campaign"`
	Targets         int    `json:"targets"`
	Generated       int    `json:"generated"`
	Reached         int    `json:"reached"`
	OracleCompleted int    `json:"oracle_completed"`
	GenerateError   int    `json:"generate_error"`
	OracleError     int    `json:"oracle_error"`
	TargetCompleted int    `json:"target_completed"`
	ExecutionError  int    `json:"execution_error"`
	Findings        int    `json:"findings"`
	ShrinkError     int    `json:"shrink_error"`
	Comparisons     int    `json:"comparisons"`
	ElapsedMS       int64  `json:"elapsed_ms"`
}

type generateErrorRecord struct {
	Type     string `json:"type"`
	Campaign string `json:"campaign"`
	Family   string `json:"family,omitempty"`
	Op       string `json:"op"`
	Width    int    `json:"width"`
	Mode     string `json:"mode"`
	Error    string `json:"error"`
}

type executionErrorRecord struct {
	Type     string     `json:"type"`
	Campaign string     `json:"campaign"`
	Family   string     `json:"family"`
	Op       string     `json:"op"`
	Width    int        `json:"width"`
	Mode     string     `json:"mode"`
	Stage    string     `json:"stage"`
	Error    string     `json:"error"`
	Original *sampleRec `json:"original"`
}

// ---------- per-case evaluation shared by sweep, replay, and shrink ----------

type caseOutcome struct {
	ref             decimalref.Result
	oracleErr       error
	cBits           string
	cFlags          uint32
	gBits           string
	gFlags          uint32
	disc            discrepancy
	targetCompleted bool
	compared        bool
	executionStage  string
}

func rawToWords(width int, raw string) (words, error) {
	switch width {
	case 32:
		v, err := strconv.ParseUint(raw, 16, 32)
		if err != nil {
			return words{}, err
		}
		return words{lo: v}, nil
	case 64:
		v, err := strconv.ParseUint(raw, 16, 64)
		if err != nil {
			return words{}, err
		}
		return words{lo: v}, nil
	case 128:
		parts := strings.Split(raw, ":")
		if len(parts) != 2 {
			return words{}, fmt.Errorf("128-bit raw must be hi:lo")
		}
		hi, err := strconv.ParseUint(parts[0], 16, 64)
		if err != nil {
			return words{}, err
		}
		lo, err := strconv.ParseUint(parts[1], 16, 64)
		if err != nil {
			return words{}, err
		}
		return words{lo: lo, hi: hi}, nil
	}
	return words{}, fmt.Errorf("unsupported width %d", width)
}

func evaluateCase(c decimalref.Case) (oc caseOutcome, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s panic: %v", oc.executionStage, r)
		}
	}()
	oc.ref, oc.oracleErr = evaluateModel(c)
	oc.executionStage = "input"
	m, ok := modeByName(c.Mode)
	if !ok {
		return oc, fmt.Errorf("unknown mode %q", c.Mode)
	}
	w, ok := widthByBits(c.Width)
	if !ok {
		return oc, fmt.Errorf("unknown width %d", c.Width)
	}
	if len(c.Operands) != opArity(c.Op) {
		return oc, fmt.Errorf("%s needs %d operands, got %d", c.Op, opArity(c.Op), len(c.Operands))
	}
	var tup [3]words
	for i, raw := range c.Operands {
		wd, err := rawToWords(c.Width, raw)
		if err != nil {
			return oc, fmt.Errorf("operand %d: %w", i, err)
		}
		tup[i] = wd
	}

	oc.executionStage = "c"
	cB, cF := cEval(w, c.Op, tup, m.native)
	oc.cBits, oc.cFlags = formatValue(w, cB), cF
	oc.executionStage = "go"
	gB, gF := goEval(w, c.Op, tup, m.native)
	oc.gBits, oc.gFlags = formatValue(w, gB), gF
	oc.targetCompleted = true
	oc.executionStage = "comparison"
	oc.disc.cgoValue = oc.cBits != oc.gBits
	oc.disc.cgoFlags = cF != gF
	if oc.oracleErr == nil {
		oc.disc.refGoValue, oc.disc.refGoFlags = refLegDiff(c.Width, oc.ref, oc.gBits, gF)
		oc.disc.refCValue, oc.disc.refCFlags = refLegDiff(c.Width, oc.ref, oc.cBits, cF)
		oc.compared = true
	}
	return oc, nil
}

func evaluateModel(c decimalref.Case) (result decimalref.Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("model panic: %v", r)
		}
	}()
	return decimalref.Evaluate(c)
}

type flagDifference struct{ missing, spurious uint32 }

func flagDiff(want, got uint32) flagDifference {
	return flagDifference{missing: want &^ got, spurious: got &^ want}
}

func (oc caseOutcome) flagDifferences() [3]flagDifference {
	return [3]flagDifference{flagDiff(oc.ref.Flags, oc.gFlags), flagDiff(oc.ref.Flags, oc.cFlags), flagDiff(oc.cFlags, oc.gFlags)}
}

func decodeOperands(width int, raws []string) ([]decimalref.Decimal, error) {
	out := make([]decimalref.Decimal, len(raws))
	for i, raw := range raws {
		d, err := decimalref.Decode(width, raw)
		if err != nil {
			return nil, fmt.Errorf("operand %d: %w", i, err)
		}
		out[i] = d
	}
	return out, nil
}

func hexFlags(f uint32) string { return fmt.Sprintf("%08x", f) }

func toSampleRec(s decimalprobe.Sample) *sampleRec {
	return &sampleRec{Version: s.Version, Family: s.Family, Case: s.Case}
}

func sameOperands(a, b decimalprobe.Sample) bool {
	if len(a.Case.Operands) != len(b.Case.Operands) {
		return false
	}
	for i := range a.Case.Operands {
		if a.Case.Operands[i] != b.Case.Operands[i] {
			return false
		}
	}
	return true
}

// ---------- reference campaign sweep ----------

type refCampaign struct {
	campaign       string
	seedStr        string
	commit         string
	goToolchain    string
	cArchiveSHA    string
	modes          []string
	shrinkAttempts int
	emit           func(any) error
	shrinkErrors   int
}

func (rc *refCampaign) makeFinding(sample decimalprobe.Sample, oc caseOutcome) *findingRec {
	classes := oc.disc.classes()
	if oc.oracleErr != nil {
		classes = append(classes, "oracle-error")
	}
	f := &findingRec{
		Type: "finding", Campaign: rc.campaign, Seed: rc.seedStr,
		Family: sample.Family, Op: sample.Case.Op, Width: sample.Case.Width, Mode: sample.Case.Mode,
		Classes:  classes,
		Original: toSampleRec(sample),
		Commit:   rc.commit, GoToolchain: rc.goToolchain, CArchiveSHA256: rc.cArchiveSHA,
		GeneratorFormatVersion: sample.Version, OracleFormatVersion: oracleFormatVersion,
		ConfiguredModes: rc.modes, ConfiguredCampaign: rc.campaign,
	}
	if oc.cBits != "" {
		f.C = &legBlock{Bits: oc.cBits, Flags: hexFlags(oc.cFlags)}
	}
	if oc.gBits != "" {
		f.Go = &legBlock{Bits: oc.gBits, Flags: hexFlags(oc.gFlags)}
	}
	if oc.oracleErr != nil {
		f.OracleError = oc.oracleErr.Error()
	} else {
		f.Reference = encodeReference(sample.Case.Width, oc.ref)
	}
	rc.attachShrink(f, sample, oc)
	if f.Shrink.Failed {
		rc.shrinkErrors++
	}
	return f
}

// attachShrink runs decimalprobe.Shrink for findings that carry an arithmetic
// discrepancy, requiring every candidate to preserve the original discrepancy
// classes and exact missing/spurious flag masks, refusing execution changes.
func (rc *refCampaign) attachShrink(f *findingRec, sample decimalprobe.Sample, oc caseOutcome) {
	defer func() {
		if r := recover(); r != nil {
			f.Shrunk = nil
			f.Shrink = &shrinkBlock{Enabled: true, Failed: true, Error: fmt.Sprintf("shrink panic: %v", r)}
		}
	}()
	if rc.shrinkAttempts <= 0 || oc.oracleErr != nil {
		f.Shrink = &shrinkBlock{Enabled: false}
		return
	}
	original := oc.disc.classes()
	if len(original) == 0 {
		f.Shrink = &shrinkBlock{Enabled: false}
		return
	}
	fails := func(cand decimalprobe.Sample) (ok bool, err error) {
		defer func() {
			if r := recover(); r != nil {
				ok, err = false, fmt.Errorf("shrink evaluation panic: %v", r)
			}
		}()
		co, e := evaluateCase(cand.Case)
		if e != nil {
			return false, e
		}
		if co.oracleErr != nil {
			return false, co.oracleErr
		}
		return sameClasses(co.disc.classes(), original) && co.flagDifferences() == oc.flagDifferences(), nil
	}
	shrunk, stats, err := decimalprobe.Shrink(sample, rc.shrinkAttempts, fails)
	f.Shrink = &shrinkBlock{Attempts: stats.Attempts, Accepted: stats.Accepted, Exhausted: stats.Exhausted, Enabled: true}
	if err != nil {
		f.Shrink.Failed = true
		f.Shrink.Error = err.Error()
		return
	}
	if !sameOperands(shrunk, sample) {
		f.Shrunk = toSampleRec(shrunk)
	}
}

type targetTotals struct {
	generated, reached, oracleCompleted, generateError, oracleError, targetCompleted, executionError, findings, comparisons int
}

func (rc *refCampaign) runTarget(seed uint64, w widthParams, mode, family, op string,
	generate func(hi, lo uint64, exp int32, negative bool) (decimalprobe.Sample, error),
	cases int) (targetTotals, error) {

	cs := newCounterSet(op)
	rng := &splitMix64{state: targetSeed(seed, w, family+"/"+op+"/"+mode)}
	start := time.Now()
	var findings int
	for i := 0; i < cases; i++ {
		cs.Generated++
		hi, lo := rng.next(), rng.next()
		exp := int32(rng.next())
		negative := rng.next()&1 == 1
		sample, gerr := generate(hi, lo, exp, negative)
		if gerr == nil {
			gerr = toSampleRec(sample).validate()
		}
		if gerr != nil {
			cs.GenerateError++
			if err := rc.emit(generateErrorRecord{
				Type: "generate_error", Campaign: rc.campaign, Family: family, Op: op,
				Width: w.width, Mode: mode, Error: gerr.Error(),
			}); err != nil {
				return targetTotals{}, err
			}
			continue
		}
		found, err := rc.evaluateSample(cs, sample)
		if err != nil {
			return targetTotals{}, err
		}
		findings += found
	}
	if err := cs.reconcileEvaluation(); err != nil {
		return targetTotals{}, fmt.Errorf("target %s/%s/d%d/%s counters do not reconcile: %w", family, op, w.width, mode, err)
	}
	if err := rc.emit(cs.record(rc.campaign, family, w.width, mode, time.Since(start).Milliseconds())); err != nil {
		return targetTotals{}, err
	}
	return targetTotals{
		generated: cs.Generated, reached: cs.Reached, oracleCompleted: cs.OracleCompleted,
		generateError: cs.GenerateError, oracleError: cs.OracleError,
		findings: findings, comparisons: cs.Comparisons,
		targetCompleted: cs.TargetCompleted, executionError: cs.ExecutionError,
	}, nil
}

func (rc *refCampaign) evaluateSample(cs *counterSet, sample decimalprobe.Sample) (int, error) {
	cs.Reached++
	oc, executionErr := evaluateCase(sample.Case)
	if oc.oracleErr != nil {
		cs.OracleError++
	} else {
		operands, err := decodeOperands(sample.Case.Width, sample.Case.Operands)
		if err != nil {
			return 0, err
		}
		if err := cs.addOracle(sample.Case.Width, oc.ref, operands); err != nil {
			return 0, err
		}
	}
	if oc.targetCompleted {
		cs.TargetCompleted++
	}
	if oc.compared {
		cs.Comparisons++
	}
	if executionErr != nil {
		cs.ExecutionError++
		if !oc.targetCompleted {
			cs.executionBeforeTarget++
		}
		if oc.oracleErr != nil {
			cs.oracleExecutionError++
		}
		if err := rc.emit(executionErrorRecord{Type: "execution_error", Campaign: rc.campaign,
			Family: sample.Family, Op: sample.Case.Op, Width: sample.Case.Width, Mode: sample.Case.Mode,
			Stage: oc.executionStage, Error: executionErr.Error(), Original: toSampleRec(sample)}); err != nil {
			return 0, err
		}
	}
	if (executionErr == nil && oc.disc.any()) || oc.oracleErr != nil {
		if err := rc.emit(rc.makeFinding(sample, oc)); err != nil {
			return 0, err
		}
		return 1, nil
	}
	return 0, nil
}

func runReference(campaign, seedText string, cases int, opsText, widthsText, modesText string, shrinkAttempts int, commit string) (runErr error) {
	if err := validateOptions(campaign, cases, 0, opsText, widthsText, modesText, shrinkAttempts); err != nil {
		return err
	}
	if err := validateCampaign(campaign); err != nil {
		return err
	}
	if campaign == campaignLegacy {
		return fmt.Errorf("runReference called with legacy campaign")
	}
	if seedText == "" {
		return fmt.Errorf("-seed is required")
	}
	seed, err := strconv.ParseUint(seedText, 0, 64)
	if err != nil {
		return fmt.Errorf("invalid -seed %q: %v", seedText, err)
	}
	if cases <= 0 {
		return fmt.Errorf("-cases must be positive, got %d", cases)
	}
	if shrinkAttempts < 0 {
		return fmt.Errorf("-shrink-attempts must be >= 0, got %d", shrinkAttempts)
	}
	widths, widthBits, err := parseWidths(widthsText)
	if err != nil {
		return err
	}
	modeNames, err := parseModeNames(modesText)
	if err != nil {
		return err
	}

	ops, families, err := referenceTargets(campaign, opsText)
	if err != nil {
		return err
	}

	out := bufio.NewWriterSize(os.Stdout, 1<<16)
	defer func() { runErr = errors.Join(runErr, out.Flush()) }()
	emit := func(record any) error {
		b, err := json.Marshal(record)
		if err != nil {
			return err
		}
		_, err = out.Write(append(b, '\n'))
		return err
	}

	rc := &refCampaign{
		campaign: campaign, seedStr: strconv.FormatUint(seed, 10), commit: commit,
		goToolchain: runtime.Version(), cArchiveSHA: activeProvenance.CArchiveSHA256,
		modes: modeNames, shrinkAttempts: shrinkAttempts, emit: emit,
	}

	if err := emit(configRecordV2{
		Type: "config", Tool: "explorediff", Campaign: campaign, FormatVersion: referenceConfigFormatVersion,
		GeneratorFormatVersion: 1, OracleFormatVersion: oracleFormatVersion,
		Seed: rc.seedStr, CasesPerTarget: cases, Widths: widthBits, Modes: modeNames,
		Ops: ops, Families: families, ShrinkAttempts: shrinkAttempts, ShrinkEnabled: shrinkAttempts > 0,
		Commit: commit, GoToolchain: rc.goToolchain, CArchiveSHA256: rc.cArchiveSHA,
		GoOS: runtime.GOOS, GoArch: runtime.GOARCH,
	}); err != nil {
		return err
	}

	start := time.Now()
	var tot targetTotals
	targets := 0
	for _, w := range widths {
		for _, mode := range modeNames {
			if campaign == campaignUniform {
				for _, op := range ops {
					op, mode, w := op, mode, w
					tt, err := rc.runTarget(seed, w, mode, "uniform-finite", op,
						func(hi, lo uint64, exp int32, negative bool) (decimalprobe.Sample, error) {
							return decimalprobe.Uniform(op, w.width, mode, hi, lo, exp, negative)
						}, cases)
					if err != nil {
						return err
					}
					accumulate(&tot, tt)
					targets++
				}
			} else {
				for _, fam := range families {
					fam, mode, w := fam, mode, w
					op, _ := decimalprobe.FamilyOperation(fam)
					tt, err := rc.runTarget(seed, w, mode, fam, op,
						func(hi, lo uint64, exp int32, negative bool) (decimalprobe.Sample, error) {
							return decimalprobe.Generate(fam, w.width, mode, hi, lo, exp, negative)
						}, cases)
					if err != nil {
						return err
					}
					accumulate(&tot, tt)
					targets++
				}
			}
		}
	}

	if err := emit(summaryRecordV2{
		Type: "summary", Campaign: campaign, Targets: targets,
		Generated: tot.generated, Reached: tot.reached, OracleCompleted: tot.oracleCompleted,
		GenerateError: tot.generateError, OracleError: tot.oracleError,
		TargetCompleted: tot.targetCompleted, ExecutionError: tot.executionError,
		Findings: tot.findings, Comparisons: tot.comparisons, ShrinkError: rc.shrinkErrors,
		ElapsedMS: time.Since(start).Milliseconds(),
	}); err != nil {
		return err
	}
	if tot.generateError+tot.oracleError+tot.executionError+rc.shrinkErrors > 0 {
		return fmt.Errorf("campaign failed: generate_error=%d oracle_error=%d execution_error=%d shrink_error=%d", tot.generateError, tot.oracleError, tot.executionError, rc.shrinkErrors)
	}
	return nil
}

func accumulate(tot *targetTotals, tt targetTotals) {
	tot.generated += tt.generated
	tot.reached += tt.reached
	tot.oracleCompleted += tt.oracleCompleted
	tot.generateError += tt.generateError
	tot.oracleError += tt.oracleError
	tot.targetCompleted += tt.targetCompleted
	tot.executionError += tt.executionError
	tot.findings += tt.findings
	tot.comparisons += tt.comparisons
}

func parseWidths(text string) ([]widthParams, []int, error) {
	var widths []widthParams
	var bits []int
	for _, part := range splitCSV(text) {
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid width %q", part)
		}
		w, ok := widthByBits(n)
		if !ok {
			return nil, nil, fmt.Errorf("unknown width %d (known: 32,64,128)", n)
		}
		widths = append(widths, w)
		bits = append(bits, n)
	}
	if len(widths) == 0 {
		return nil, nil, fmt.Errorf("need at least one width")
	}
	return widths, bits, nil
}

func parseModeNames(text string) ([]string, error) {
	var names []string
	for _, name := range splitCSV(text) {
		if !knownMode(name) {
			return nil, fmt.Errorf("unknown mode %q (known: %s)", name, modeNamesCSV())
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("need at least one mode")
	}
	return names, nil
}

// ---------- exact fixed-input JSONL replay ----------

func runReplay(path, commit string) (runErr error) {
	samples, err := loadReplay(path)
	if err != nil {
		return err
	}
	out := bufio.NewWriterSize(os.Stdout, 1<<16)
	defer func() { runErr = errors.Join(runErr, out.Flush()) }()
	emit := func(v any) error { return json.NewEncoder(out).Encode(v) }
	rc := &refCampaign{campaign: "replay", commit: commit, goToolchain: runtime.Version(), cArchiveSHA: activeProvenance.CArchiveSHA256, emit: emit}
	if err := emit(configRecordV2{Type: "config", Tool: "explorediff", Campaign: "replay", FormatVersion: referenceConfigFormatVersion, GeneratorFormatVersion: 1, OracleFormatVersion: oracleFormatVersion, Commit: commit, GoToolchain: rc.goToolchain, CArchiveSHA256: rc.cArchiveSHA, GoOS: runtime.GOOS, GoArch: runtime.GOARCH}); err != nil {
		return err
	}
	start := time.Now()
	var tot targetTotals
	for _, s := range samples {
		targetStart := time.Now()
		cs := newCounterSet(s.Case.Op)
		cs.Generated = 1
		found, err := rc.evaluateSample(cs, decimalprobe.Sample{Version: s.Version, Family: s.Family, Case: s.Case})
		if err != nil {
			return err
		}
		tot.findings += found
		if err := cs.reconcileEvaluation(); err != nil {
			return err
		}
		if err := emit(cs.record("replay", s.Family, s.Case.Width, s.Case.Mode, time.Since(targetStart).Milliseconds())); err != nil {
			return err
		}
		tot.generated++
		tot.reached++
		tot.comparisons += cs.Comparisons
		tot.targetCompleted += cs.TargetCompleted
		tot.executionError += cs.ExecutionError
		tot.oracleCompleted += cs.OracleCompleted
		tot.oracleError += cs.OracleError
	}
	if err := emit(summaryRecordV2{Type: "summary", Campaign: "replay", Targets: len(samples), Generated: tot.generated, Reached: tot.reached, OracleCompleted: tot.oracleCompleted, OracleError: tot.oracleError, TargetCompleted: tot.targetCompleted, ExecutionError: tot.executionError, Findings: tot.findings, Comparisons: tot.comparisons, ElapsedMS: time.Since(start).Milliseconds()}); err != nil {
		return err
	}
	if tot.oracleError+tot.executionError > 0 {
		return fmt.Errorf("replay failed: oracle_error=%d execution_error=%d", tot.oracleError, tot.executionError)
	}
	return nil
}

func appendOracle(oc caseOutcome) []string {
	classes := oc.disc.classes()
	if oc.oracleErr != nil {
		classes = append(classes, "oracle-error")
	}
	return classes
}
