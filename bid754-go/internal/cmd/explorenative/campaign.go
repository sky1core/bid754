package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sky1core/bid754/bid754-go/internal/decimalprobe"
	"io"
	"os"
	"strings"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

// This file is the portable (no-cgo) core of the reference exploration
// campaigns: campaign selection, the JSONL replay-input schema and its strict
// decoder, the numeric-situation counters, and the finding record shape. The
// cgo differential/oracle orchestration is in reference_native.go. Legacy
// C-vs-Go exploration (explore.go / main_native.go legacy path) is untouched.

const (
	campaignLegacy  = "legacy"
	campaignRel     = "relations"
	campaignUniform = "uniform-finite"
)

// referenceConfigFormatVersion tags new-campaign output; legacy stays 1.
const referenceConfigFormatVersion = 2

// oracleFormatVersion is the decimalref reference-model contract version this
// tool targets. The generator format version travels per sample (Sample.Version).
const oracleFormatVersion = 1

// IEEE 5-flag word bits (Intel BID / decimalref layout).
const (
	flagInvalid        uint32 = 0x01
	flagDivisionByZero uint32 = 0x04
	flagOverflow       uint32 = 0x08
	flagUnderflow      uint32 = 0x10
	flagInexact        uint32 = 0x20
)

var referenceOps = map[string]bool{
	"add": true, "sub": true, "mul": true, "div": true, "fma": true, "quantize": true,
}

func supportedReferenceOp(op string) bool { return referenceOps[op] }

func validateCampaign(name string) error {
	switch name {
	case campaignLegacy, campaignRel, campaignUniform:
		return nil
	}
	return fmt.Errorf("unknown campaign %q (known: legacy, relations, uniform-finite)", name)
}

func knownMode(name string) bool {
	_, ok := modeByName(name)
	return ok
}

// ---------- record schema (emitted and, for sample/finding, replay input) ----------

type sampleRec struct {
	Version int             `json:"version"`
	Family  string          `json:"family"`
	Case    decimalref.Case `json:"case"`
}

type legBlock struct {
	Bits  string `json:"bits"`
	Flags string `json:"flags"`
}

type refBlock struct {
	Raw          string `json:"raw"`
	Kind         string `json:"kind"`
	Negative     bool   `json:"negative"`
	Coeff        string `json:"coeff"`
	Exp          int    `json:"exp"`
	Flags        string `json:"flags"`
	Rounding     string `json:"rounding"`
	Cancellation int    `json:"cancellation"`
	Carry        bool   `json:"carry"`
}

type shrinkBlock struct {
	Attempts  int    `json:"attempts"`
	Accepted  int    `json:"accepted"`
	Exhausted bool   `json:"exhausted"`
	Enabled   bool   `json:"enabled"`
	Failed    bool   `json:"failed"`
	Error     string `json:"error,omitempty"`
}

// findingRec is one finding line. It doubles as the strict replay-input record:
// a hand-authored sample line sets only Type+Sample, a finding line sets
// Original/Shrunk plus descriptive fields. Every legitimate field is declared,
// so DisallowUnknownFields rejects only genuinely unknown fields.
type findingRec struct {
	Type                   string       `json:"type"`
	Campaign               string       `json:"campaign,omitempty"`
	Seed                   string       `json:"seed,omitempty"`
	Family                 string       `json:"family,omitempty"`
	Op                     string       `json:"op,omitempty"`
	Width                  int          `json:"width,omitempty"`
	Mode                   string       `json:"mode,omitempty"`
	Classes                []string     `json:"classes,omitempty"`
	Sample                 *sampleRec   `json:"sample,omitempty"`
	Original               *sampleRec   `json:"original,omitempty"`
	Shrunk                 *sampleRec   `json:"shrunk,omitempty"`
	Shrink                 *shrinkBlock `json:"shrink,omitempty"`
	Reference              *refBlock    `json:"reference,omitempty"`
	OracleError            string       `json:"oracle_error,omitempty"`
	C                      *legBlock    `json:"c,omitempty"`
	Go                     *legBlock    `json:"go,omitempty"`
	Commit                 string       `json:"commit,omitempty"`
	GoToolchain            string       `json:"go_toolchain,omitempty"`
	CArchiveSHA256         string       `json:"c_archive_sha256,omitempty"`
	GeneratorFormatVersion int          `json:"generator_format_version,omitempty"`
	OracleFormatVersion    int          `json:"oracle_format_version,omitempty"`
	ConfiguredModes        []string     `json:"configured_modes,omitempty"`
	ConfiguredCampaign     string       `json:"configured_campaign,omitempty"`
}

// ---------- strict replay decoding ----------

// decodeReplayLine strictly decodes one replay-input JSONL line. Unknown
// fields, trailing data, and non-sample/non-finding record types are rejected;
// there is no silent field drop.
func decodeReplayLine(line []byte) (*findingRec, error) {
	if err := rejectDuplicateKeys(json.NewDecoder(bytes.NewReader(line))); err != nil {
		return nil, err
	}
	var peek struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(line, &peek); err != nil {
		return nil, fmt.Errorf("invalid JSON record: %w", err)
	}
	switch peek.Type {
	case "sample", "finding":
	case "":
		return nil, fmt.Errorf("record is missing a \"type\"")
	default:
		return nil, fmt.Errorf("replay input accepts only sample/finding records, got type %q", peek.Type)
	}

	dec := json.NewDecoder(bytes.NewReader(line))
	dec.DisallowUnknownFields()
	var rec findingRec
	if err := dec.Decode(&rec); err != nil {
		return nil, err
	}
	if dec.More() {
		return nil, fmt.Errorf("trailing data after JSON record")
	}
	return &rec, nil
}

// replaySamples resolves the record into the recorded sample cases to replay.
// It rejects ambiguous records that mix sample and finding slots.
func (r *findingRec) replaySamples() ([]*sampleRec, error) {
	switch r.Type {
	case "sample":
		if r.Sample == nil {
			return nil, fmt.Errorf("sample record is missing \"sample\"")
		}
		if r.Original != nil || r.Shrunk != nil {
			return nil, fmt.Errorf("sample record must not carry original/shrunk")
		}
		return []*sampleRec{r.Sample}, nil
	case "finding":
		if r.Original == nil {
			return nil, fmt.Errorf("finding record is missing \"original\"")
		}
		if r.Sample != nil {
			return nil, fmt.Errorf("finding record must not carry a top-level sample")
		}
		out := []*sampleRec{r.Original}
		if r.Shrunk != nil {
			out = append(out, r.Shrunk)
		}
		return out, nil
	}
	return nil, fmt.Errorf("unexpected record type %q", r.Type)
}

// validate checks a recorded sample is a well-formed reference-campaign case
// with supported op/mode/width and finite, decodable operands.
func (s *sampleRec) validate() error {
	if !supportedReferenceOp(s.Case.Op) {
		return fmt.Errorf("unsupported reference op %q", s.Case.Op)
	}
	_, err := decimalprobe.Validate(decimalprobe.Sample{Version: s.Version, Family: s.Family, Case: s.Case})
	return err
}

// ---------- numeric-situation counters ----------

type counterSet struct {
	op    string
	arity int

	Generated             int
	Reached               int
	OracleCompleted       int
	GenerateError         int
	OracleError           int
	TargetCompleted       int
	ExecutionError        int
	Comparisons           int
	executionBeforeTarget int
	oracleExecutionError  int

	ResultClass map[string]int
	InputClass  map[string]int
	Rounding    map[string]int
	Flags       map[string]int
	Cancel      map[string]int
	Carry       map[string]int
}

func newCounterSet(op string) *counterSet {
	return &counterSet{
		op:          op,
		arity:       opArity(op),
		ResultClass: map[string]int{},
		InputClass:  map[string]int{},
		Rounding:    map[string]int{},
		Flags:       map[string]int{},
		Cancel:      map[string]int{},
		Carry:       map[string]int{},
	}
}

func cancellationBin(n int) string {
	switch {
	case n <= 0:
		return "0"
	case n <= 3:
		return "1-3"
	default:
		return "4+"
	}
}

// addOracle records one case whose reference model result completed. Input and
// result classes come from the independent decimalref.Classify; rounding,
// flags, cancellation, and carry come from the model result.
func (c *counterSet) addOracle(width int, res decimalref.Result, operands []decimalref.Decimal) error {
	c.OracleCompleted++
	rc, err := decimalref.Classify(width, res.Value)
	if err != nil {
		return fmt.Errorf("classify result: %w", err)
	}
	c.ResultClass[rc]++
	for i, o := range operands {
		ic, err := decimalref.Classify(width, o)
		if err != nil {
			return fmt.Errorf("classify operand %d: %w", i, err)
		}
		c.InputClass[ic]++
	}
	rounding := res.Rounding
	if rounding == "" {
		rounding = "exact"
	}
	c.Rounding[rounding]++
	for _, f := range []struct {
		name string
		bit  uint32
	}{
		{"invalid", flagInvalid}, {"division_by_zero", flagDivisionByZero},
		{"overflow", flagOverflow}, {"underflow", flagUnderflow}, {"inexact", flagInexact},
	} {
		if res.Flags&f.bit != 0 {
			c.Flags[f.name]++
		}
	}
	c.Cancel[cancellationBin(res.Cancellation)]++
	if res.Carry {
		c.Carry["true"]++
	} else {
		c.Carry["false"]++
	}
	return nil
}

func (c *counterSet) reconcileEvaluation() error {
	if err := c.reconcile(); err != nil {
		return err
	}
	if c.ExecutionError < 0 || c.TargetCompleted < 0 || c.Comparisons < 0 ||
		c.TargetCompleted > c.Reached || c.ExecutionError > c.Reached ||
		c.executionBeforeTarget < 0 || c.executionBeforeTarget > c.ExecutionError ||
		c.oracleExecutionError < 0 || c.oracleExecutionError > c.OracleError || c.oracleExecutionError > c.ExecutionError ||
		c.TargetCompleted+c.executionBeforeTarget != c.Reached ||
		c.Comparisons > c.TargetCompleted || c.Comparisons > c.OracleCompleted ||
		c.Comparisons+c.ExecutionError > c.Reached ||
		c.Comparisons+c.ExecutionError+c.OracleError-c.oracleExecutionError != c.Reached {
		return fmt.Errorf("evaluation counts do not reconcile: reached=%d oracle_completed=%d oracle_error=%d target_completed=%d execution_error=%d comparisons=%d", c.Reached, c.OracleCompleted, c.OracleError, c.TargetCompleted, c.ExecutionError, c.Comparisons)
	}
	return nil
}

// reconcile enforces that no case is dropped: generated splits into
// reached+generate_error, reached into oracle_completed+oracle_error, and the
// per-completion histograms sum to oracle_completed (input classes times arity).
func (c *counterSet) reconcile() error {
	if c.Generated != c.Reached+c.GenerateError {
		return fmt.Errorf("generated %d != reached %d + generate_error %d", c.Generated, c.Reached, c.GenerateError)
	}
	if c.Reached != c.OracleCompleted+c.OracleError {
		return fmt.Errorf("reached %d != oracle_completed %d + oracle_error %d", c.Reached, c.OracleCompleted, c.OracleError)
	}
	if sum(c.ResultClass) != c.OracleCompleted {
		return fmt.Errorf("result_class total %d != oracle_completed %d", sum(c.ResultClass), c.OracleCompleted)
	}
	if sum(c.Rounding) != c.OracleCompleted {
		return fmt.Errorf("rounding total %d != oracle_completed %d", sum(c.Rounding), c.OracleCompleted)
	}
	if sum(c.Carry) != c.OracleCompleted {
		return fmt.Errorf("carry total %d != oracle_completed %d", sum(c.Carry), c.OracleCompleted)
	}
	if sum(c.Cancel) != c.OracleCompleted {
		return fmt.Errorf("cancellation total %d != oracle_completed %d", sum(c.Cancel), c.OracleCompleted)
	}
	if sum(c.InputClass) != c.OracleCompleted*c.arity {
		return fmt.Errorf("input_class total %d != oracle_completed %d * arity %d", sum(c.InputClass), c.OracleCompleted, c.arity)
	}
	return nil
}

func sum(m map[string]int) int {
	t := 0
	for _, v := range m {
		t += v
	}
	return t
}

// countersRecord is one emitted per-target counters line (intent vs actual).
type countersRecord struct {
	Type             string         `json:"type"`
	Campaign         string         `json:"campaign"`
	Family           string         `json:"family,omitempty"`
	Op               string         `json:"op"`
	Width            int            `json:"width"`
	Mode             string         `json:"mode"`
	Generated        int            `json:"generated"`
	Reached          int            `json:"reached"`
	OracleCompleted  int            `json:"oracle_completed"`
	GenerateError    int            `json:"generate_error"`
	OracleError      int            `json:"oracle_error"`
	TargetCompleted  int            `json:"target_completed"`
	ExecutionError   int            `json:"execution_error"`
	Comparisons      int            `json:"comparisons"`
	ResultClass      map[string]int `json:"result_class"`
	InputClass       map[string]int `json:"input_class"`
	Rounding         map[string]int `json:"rounding"`
	Flags            map[string]int `json:"flags"`
	CancellationBins map[string]int `json:"cancellation_bins"`
	Carry            map[string]int `json:"carry"`
	ElapsedMS        int64          `json:"elapsed_ms"`
}

func (c *counterSet) record(campaign, family string, width int, mode string, elapsedMS int64) *countersRecord {
	return &countersRecord{
		Type: "counters", Campaign: campaign, Family: family, Op: c.op, Width: width, Mode: mode,
		Generated: c.Generated, Reached: c.Reached, OracleCompleted: c.OracleCompleted,
		GenerateError: c.GenerateError, OracleError: c.OracleError,
		TargetCompleted: c.TargetCompleted, ExecutionError: c.ExecutionError, Comparisons: c.Comparisons,
		ResultClass: c.ResultClass, InputClass: c.InputClass, Rounding: c.Rounding,
		Flags: c.Flags, CancellationBins: c.Cancel, Carry: c.Carry, ElapsedMS: elapsedMS,
	}
}

// ---------- discrepancy classification (portable core) ----------

// discrepancy captures the three independent pairwise checks. reference-* is
// the decimalref model against a leg (numeric value/class/sign(+canonicality)
// and the IEEE 5-flag word, independently — not cohort equality). c-go is the
// legacy exact raw-bits/raw-flags check, retained separately.
type discrepancy struct {
	refGoValue, refGoFlags bool
	refCValue, refCFlags   bool
	cgoValue, cgoFlags     bool
}

func (d discrepancy) classes() []string {
	var out []string
	add := func(pair string, value, flags bool) {
		if value {
			out = append(out, pair+":value")
		}
		if flags {
			out = append(out, pair+":flags")
		}
	}
	add("reference-go", d.refGoValue, d.refGoFlags)
	add("reference-c", d.refCValue, d.refCFlags)
	add("c-go", d.cgoValue, d.cgoFlags)
	return out
}

func (d discrepancy) any() bool { return len(d.classes()) > 0 }

// refLegDiff isolates value from flag divergence between the model result and a
// leg, comparing values with independent zero flag words.
func refLegDiff(width int, ref decimalref.Result, bits string, flags uint32) (value, flag bool) {
	flag = ref.Flags != flags
	want := ref
	want.Flags = 0
	if err := decimalref.Compare(width, want, bits, 0); err != nil {
		value = true
	}
	return value, flag
}

func sameClasses(super, sub []string) bool {
	if len(super) != len(sub) {
		return false
	}
	set := make(map[string]bool, len(super))
	for _, s := range super {
		set[s] = true
	}
	for _, s := range sub {
		if !set[s] {
			return false
		}
		delete(set, s)
	}
	return len(set) == 0
}

// encodeReference renders the model result for a finding record.
func encodeReference(width int, r decimalref.Result) *refBlock {
	b := &refBlock{
		Kind:         r.Value.Kind,
		Negative:     r.Value.Negative,
		Exp:          r.Value.Exp,
		Flags:        fmt.Sprintf("%08x", r.Flags),
		Rounding:     r.Rounding,
		Cancellation: r.Cancellation,
		Carry:        r.Carry,
	}
	if r.Value.Coeff != nil {
		b.Coeff = r.Value.Coeff.String()
	}
	if raw, err := decimalref.Encode(width, r.Value); err == nil {
		b.Raw = raw
	}
	return b
}

func rejectDuplicateKeys(dec *json.Decoder) error {
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
			key, err := dec.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			name = strings.ToLower(strings.ToUpper(name))
			if !ok || seen[name] {
				return fmt.Errorf("duplicate or invalid JSON key %q", key)
			}
			seen[name] = true
			if err := rejectDuplicateKeys(dec); err != nil {
				return err
			}
		}
	case '[':
		for dec.More() {
			if err := rejectDuplicateKeys(dec); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
	_, err = dec.Token()
	return err
}

func loadReplay(path string) ([]*sampleRec, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return readReplay(f)
}

func readReplay(in io.Reader) ([]*sampleRec, error) {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	var samples []*sampleRec
	for line := 1; scanner.Scan(); line++ {
		if len(bytes.TrimSpace(scanner.Bytes())) == 0 {
			continue
		}
		rec, err := decodeReplayLine(scanner.Bytes())
		if err != nil {
			return nil, fmt.Errorf("replay line %d: %w", line, err)
		}
		batch, err := rec.replaySamples()
		if err != nil {
			return nil, fmt.Errorf("replay line %d: %w", line, err)
		}
		for _, s := range batch {
			if err := s.validate(); err != nil {
				return nil, fmt.Errorf("replay line %d: %w", line, err)
			}
			if (rec.Family != "" && rec.Family != s.Family) || (rec.Op != "" && rec.Op != s.Case.Op) || (rec.Width != 0 && rec.Width != s.Case.Width) || (rec.Mode != "" && rec.Mode != s.Case.Mode) {
				return nil, fmt.Errorf("replay line %d: conflicting sample metadata", line)
			}
		}
		if len(batch) == 2 && (batch[0].Family != batch[1].Family || batch[0].Case.Op != batch[1].Case.Op || batch[0].Case.Width != batch[1].Case.Width || batch[0].Case.Mode != batch[1].Case.Mode) {
			return nil, fmt.Errorf("replay line %d: conflicting original/shrunk", line)
		}
		samples = append(samples, batch...)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(samples) == 0 {
		return nil, fmt.Errorf("replay contains zero samples")
	}
	return samples, nil
}
