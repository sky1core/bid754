package testgen

// Decimal32 unary exhaustive differential gate codegen.
//
// Generates the native cgo shim, the long runner, and the stub for the
// `d32_exhaustive` gate: every declared unary Decimal32 lane is compared
// bit+flag exact between pinned Intel BID C and the Go mechanical port over
// the full 32-bit input space (bit patterns 0..2^32-1, non-canonical
// included). The gate is exhaustive, so there is no case corpus to emit:
// the generated artifacts carry the lane table, the loop-bound and count
// constants, and the routing-sentinel rows. Result digests cannot be
// computed at generation time (they require executing pinned Intel C), so
// they are hand-pinned per lane in devtools/verification_anchors.json and
// bound to the runner's per-lane digest log lines by cmd/verifylog
// (domain d32-exhaustive), never by this generator.
//
// Lane selection (where a finite input space is enumerable, enumeration
// replaces sampling; unary Decimal32 only):
//   - bid32_sqrt across all five rounding modes,
//   - bid32_round_integral_exact across all five rounding modes,
//   - the five fixed-attribute bid32_round_integral_* variants,
//   - the exact width promotions bid32_to_bid64 and bid32_to_bid128.
// bid32_negate, bid32_abs, and bid32_copy are deliberately excluded: their
// ports are single sign-bit/mask/identity expressions with no
// data-dependent control flow and no status flags, and the readtest and
// FFI bit-compare domains already exercise them; an exhaustive sweep adds
// no discriminating power for that shape of operation.
//
// GUARDRAILS: this generator never reads or writes
// devtools/verification_anchors.json or devtools/verification_sentinels.json.
// Sentinel rows land only in the generated runner; the human pin flows
// through `cmd/testgen -print-d32-exhaustive-sentinel-anchors` stdout and a
// manual paste audited by TestVerificationAnchorsMatchGeneratedArtifacts.

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

const (
	d32ExhaustiveNativeShimGeneratedPath = "../bid754-go/generated_d32_exhaustive_native.go"
	d32ExhaustiveRunnerGeneratedPath     = "../bid754-go/generated_d32_exhaustive_long_test.go"
	d32ExhaustiveStubGeneratedPath       = "../bid754-go/generated_d32_exhaustive_stub_test.go"

	d32ExhaustiveCasesPerLane = uint64(1) << 32
)

//go:embed d32_exhaustive_templates/*
var d32ExhaustiveTemplates embed.FS

// d32ExhaustiveLaneSpec declares one exhaustive lane: a single Intel
// bid32 unary entry point at one fixed rounding attribute (or no rounding
// attribute), swept over the entire 32-bit input space.
type d32ExhaustiveLaneSpec struct {
	// name is the lane identity: the anchors digest-map key, the runner log
	// token, and the generated lane-name literal entry.
	name string
	// opConst is the template-static runner op constant suffix
	// (d32ExhaustiveOp<opConst>).
	opConst string
	// opToken is the sentinel-row operation token. Mode-taking lanes share
	// one token across their five mode lanes; fixed lanes use their name.
	opToken string
	// nativeMode is the Intel raw rounding mode for mode-taking lanes and
	// -1 for fixed-attribute and conversion lanes.
	nativeMode int
}

// d32ExhaustiveModeOrder is the tier1ArithmeticModes native mode order.
var d32ExhaustiveModeOrder = []struct {
	name   string
	native int
}{
	{name: "nearest_even", native: 0},
	{name: "nearest_away", native: 4},
	{name: "toward_zero", native: 3},
	{name: "toward_positive", native: 2},
	{name: "toward_negative", native: 1},
}

// d32ExhaustiveLaneSpecs returns the fixed lane table in execution order.
func d32ExhaustiveLaneSpecs() []d32ExhaustiveLaneSpec {
	lanes := []d32ExhaustiveLaneSpec{}
	for _, mode := range d32ExhaustiveModeOrder {
		lanes = append(lanes, d32ExhaustiveLaneSpec{
			name:       "sqrt_" + mode.name,
			opConst:    "Sqrt",
			opToken:    "sqrt",
			nativeMode: mode.native,
		})
	}
	for _, mode := range d32ExhaustiveModeOrder {
		lanes = append(lanes, d32ExhaustiveLaneSpec{
			name:       "round_integral_exact_" + mode.name,
			opConst:    "RoundIntegralExact",
			opToken:    "round_integral_exact",
			nativeMode: mode.native,
		})
	}
	fixed := []struct {
		name    string
		opConst string
	}{
		{name: "round_integral_nearest_even", opConst: "RoundIntegralNearestEven"},
		{name: "round_integral_nearest_away", opConst: "RoundIntegralNearestAway"},
		{name: "round_integral_zero", opConst: "RoundIntegralZero"},
		{name: "round_integral_positive", opConst: "RoundIntegralPositive"},
		{name: "round_integral_negative", opConst: "RoundIntegralNegative"},
		{name: "to_bid64", opConst: "ToBid64"},
		{name: "to_bid128", opConst: "ToBid128"},
	}
	for _, lane := range fixed {
		lanes = append(lanes, d32ExhaustiveLaneSpec{
			name:       lane.name,
			opConst:    lane.opConst,
			opToken:    lane.name,
			nativeMode: -1,
		})
	}
	return lanes
}

// d32ExhaustiveOperationCount counts distinct op constants (an op with five
// mode lanes is one operation).
func d32ExhaustiveOperationCount(lanes []d32ExhaustiveLaneSpec) int {
	ops := map[string]bool{}
	for _, lane := range lanes {
		ops[lane.opConst] = true
	}
	return len(ops)
}

// d32ExhaustiveValidateLanes fails generation on a malformed lane table so
// no partial output can carry a duplicate or inconsistently moded lane.
func d32ExhaustiveValidateLanes(lanes []d32ExhaustiveLaneSpec) error {
	names := map[string]bool{}
	modesByToken := map[string]map[int]bool{}
	for _, lane := range lanes {
		if names[lane.name] {
			return fmt.Errorf("d32 exhaustive lane table: duplicate lane name %q", lane.name)
		}
		names[lane.name] = true
		if modesByToken[lane.opToken] == nil {
			modesByToken[lane.opToken] = map[int]bool{}
		}
		if modesByToken[lane.opToken][lane.nativeMode] {
			return fmt.Errorf("d32 exhaustive lane table: duplicate (op %q, mode %d) lane", lane.opToken, lane.nativeMode)
		}
		modesByToken[lane.opToken][lane.nativeMode] = true
		if lane.nativeMode < -1 || lane.nativeMode > 4 {
			return fmt.Errorf("d32 exhaustive lane table: lane %q carries out-of-range native mode %d", lane.name, lane.nativeMode)
		}
	}
	for _, token := range []string{"sqrt", "round_integral_exact"} {
		modes := modesByToken[token]
		if len(modes) != 5 {
			return fmt.Errorf("d32 exhaustive lane table: op %q covers %d modes, want all 5", token, len(modes))
		}
		for native := 0; native <= 4; native++ {
			if !modes[native] {
				return fmt.Errorf("d32 exhaustive lane table: op %q is missing native mode %d", token, native)
			}
		}
	}
	return nil
}

func d32ExhaustiveLaneRowLiterals(lanes []d32ExhaustiveLaneSpec) string {
	var out strings.Builder
	for _, lane := range lanes {
		fmt.Fprintf(&out, "\t{name: %q, opToken: %q, op: d32ExhaustiveOp%s, nativeMode: %d},\n",
			lane.name, lane.opToken, lane.opConst, lane.nativeMode)
	}
	return out.String()
}

func d32ExhaustiveLaneNameLiterals(lanes []d32ExhaustiveLaneSpec) string {
	var out strings.Builder
	for _, lane := range lanes {
		fmt.Fprintf(&out, "\t%q,\n", lane.name)
	}
	return out.String()
}

func WriteD32ExhaustiveOutputs(repoRoot string) error {
	files, err := GenerateD32ExhaustiveOutputs()
	if err != nil {
		return err
	}
	for path, data := range files {
		fullPath := filepath.Join(repoRoot, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write generated d32 exhaustive output %q: %w", fullPath, err)
		}
	}
	return nil
}

func GenerateD32ExhaustiveOutputs() (map[string][]byte, error) {
	lanes := d32ExhaustiveLaneSpecs()
	if err := d32ExhaustiveValidateLanes(lanes); err != nil {
		return nil, err
	}
	laneCount := uint64(len(lanes))
	totalComparisons := laneCount * d32ExhaustiveCasesPerLane

	// Routing sentinels: deterministic known-answer rows pinned through the
	// public-API oracle and self-asserted for lane coverage and mode/op
	// sensitivity; a selection failure aborts the whole generation run with
	// no partial output.
	sentinelRows, err := GenerateD32ExhaustiveSentinelRows()
	if err != nil {
		return nil, err
	}

	shimTemplate, err := d32ExhaustiveTemplates.ReadFile("d32_exhaustive_templates/go_d32_exhaustive_native.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("read d32 exhaustive native shim template: %w", err)
	}
	runnerTemplate, err := d32ExhaustiveTemplates.ReadFile("d32_exhaustive_templates/go_d32_exhaustive_long_test.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("read d32 exhaustive runner template: %w", err)
	}
	stubTemplate, err := d32ExhaustiveTemplates.ReadFile("d32_exhaustive_templates/go_d32_exhaustive_stub_test.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("read d32 exhaustive stub template: %w", err)
	}

	replacer := strings.NewReplacer(
		"@@D32X_LANE_COUNT@@", fmt.Sprint(laneCount),
		"@@D32X_OPERATION_COUNT@@", fmt.Sprint(d32ExhaustiveOperationCount(lanes)),
		"@@D32X_TOTAL_COMPARISONS@@", fmt.Sprint(totalComparisons),
		"@@D32X_LANE_ROWS@@", d32ExhaustiveLaneRowLiterals(lanes),
		"@@D32X_LANE_NAME_ROWS@@", d32ExhaustiveLaneNameLiterals(lanes),
		"@@D32X_SENTINEL_COUNT@@", fmt.Sprint(len(sentinelRows)),
		"@@D32X_SENTINEL_ROWS@@", d32ExhaustiveSentinelGoRowLiterals(sentinelRows),
	)

	outputs := map[string][]byte{
		d32ExhaustiveNativeShimGeneratedPath: []byte(genmarker.Line("testgen") + "\n" + replacer.Replace(string(shimTemplate))),
		d32ExhaustiveRunnerGeneratedPath:     []byte(genmarker.Line("testgen") + "\n" + replacer.Replace(string(runnerTemplate))),
		d32ExhaustiveStubGeneratedPath:       []byte(genmarker.Line("testgen") + "\n" + replacer.Replace(string(stubTemplate))),
	}
	return formatGeneratedGoOutputs(outputs)
}
