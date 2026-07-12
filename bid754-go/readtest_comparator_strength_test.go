package bid754

// Hand-written comparator-strength anchor for the generated readtest gate.
//
// devtools/verification_anchors.json pins the readtest *case counts* outside
// every generation path so a generator regression that shrinks the case set
// cannot re-pin its own smaller counts. This file is the same principle
// applied to comparator *semantics*: it pins how strict the generated shared
// comparison helpers (bid754-go/generated_readtest_shared.go, emitted by
// devtools testgen) must be. A relaxation of the comparator meaning — for
// example replacing the Intel readtest.c exact-bit round-trip with a value
// comparison — keeps every case count identical, so no counting gate can see
// it; only a semantic anchor that calls the helpers directly can.
//
// The assertions further below pin the comparator *bodies*, but a generator
// regression could leave those bodies untouched and instead stop calling them
// — emitting a less strict inline comparison inside the generated runner. That
// keeps every case count identical (counting gate blind), regenerates cleanly
// (verify-generated only checks reproducibility, not which helper is called),
// and fails fewer native rows (a looser compare passes more), so no existing
// gate sees it. The TestReadtestComparatorBinding* tests close that gap: they
// statically parse the checked-in generated runner and dispatch sources and
// assert each anchored comparator is still invoked in a result-consuming
// position on a live path reachable from the file's case-decision entry
// function (same-file call graph — a bare identifier mention, an alternate call in
// unreachable code, or a call whose result is discarded does not satisfy it).
// That binding is part of this anchor for the same reason the
// semantics are — it must live outside the generation path so the template
// change that drops the call cannot regenerate the check that would catch it.
//
// This file is deliberately hand-written and kept OUTSIDE the generation
// path (like devtools/verification_anchors.json, and like the hand-written
// public tests in parse_nan_validation_public_test.go): if it were generated,
// the same template change that relaxes a comparator could regenerate the
// anchor to match. Do not move these assertions into a generated file.
//
// The pinned upstream meaning is Intel readtest.c check_results
// (devtools/third_party/intel_dfp/TESTS/readtest.c):
//   - check32/check64/check128 (readtest.c:562-599) are exact bit compares,
//     so cohort members with equal value but different quantum stay distinct;
//   - decimal-literal expected fields are parsed with the library's own
//     bid*_from_string at the row rounding mode (get_test, readtest.c:1018+),
//     and to_string rows round-trip the produced string the same way
//     (readtest.c:1453-1489);
//   - CMP_EQUALSTATUS rows (readtest.c:1497-1514) pass on exact bits OR when
//     the library's own bid*_quiet_not_equal reports the values equal, so +0
//     and -0 compare equal and equal-value cohort members pass, while any NaN
//     result still requires an exact bit match (quiet_not_equal is true for
//     NaN operands);
//   - frexp/modf secondary outputs are compared exactly
//     (i1 != i2 / R64_1 != B64 / check128(R_1, B), readtest.c:1480,1489).
//
// Every rejection case below carries the same contract: if it starts passing
// (or an acceptance case starts failing), the generated gate has drifted from
// Intel readtest.c semantics — fix the generator, never this anchor.
//
// All calls run through the goport backend (goportReadtestStringBackend /
// goportReadtestGeneratedSigned); the native backend shares the same
// generated helper bodies, so anchoring one backend anchors the comparator
// logic for both.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// comparatorStrengthBits converts a readtest operand literal to the bits
// string an actual generated runner would hold as the operation result. Bit
// literals pass through untouched; decimal literals go through the goport
// bid*_from_string exactly like a dispatch result. This intentionally does
// NOT use readtestValueBits (a helper under test) for the "got" side, so a
// relaxed expected-side parse cannot mask itself.
func comparatorStrengthBits(t *testing.T, format, literal string, rounding int) string {
	t.Helper()
	if len(literal) > 0 && literal[0] == '[' {
		return literal
	}
	switch format {
	case "decimal32":
		raw, _ := goportReadtestBID32FromString(literal, rounding)
		return fmt.Sprintf("[%08x]", raw)
	case "decimal64":
		raw, _ := goportReadtestBID64FromString(literal, rounding)
		return fmt.Sprintf("[%016x]", raw)
	case "decimal128":
		raw, _ := goportReadtestBID128FromString(literal, rounding)
		return formatReadtestBits128(raw)
	default:
		t.Fatalf("unsupported format %q", format)
		return ""
	}
}

// TestReadtestOperandParsingUsesNearestEvenRawPort pins the get_ops conversion
// boundary itself. Intel readtest.c forces nearest-even while parsing decimal
// operands, and its corpus deliberately includes values that round or change
// range. The NaN payload case additionally proves the harness calls Intel's
// mechanical from-string path rather than the stronger public raw wrapper.
// Routing through public constructors would therefore reject valid official
// rows or change the operand bits before the operation is exercised.
func TestReadtestOperandParsingUsesNearestEvenRawPort(t *testing.T) {
	for _, input := range []string{"1.2345678", "1e97", "1e-102", "NaN123"} {
		got, err := parseReadtestBits32(input)
		if err != nil {
			t.Fatalf("parseReadtestBits32(%q): %v", input, err)
		}
		want, _ := goportReadtestBID32FromString(input, 0)
		if got != want {
			t.Errorf("parseReadtestBits32(%q) = %#x, want nearest-even raw port %#x", input, got, want)
		}
	}

	for _, input := range []string{"1.2345678901234567", "1e385", "1e-399", "NaN123"} {
		got, err := parseReadtestBits64(input)
		if err != nil {
			t.Fatalf("parseReadtestBits64(%q): %v", input, err)
		}
		want, _ := goportReadtestBID64FromString(input, 0)
		if got != want {
			t.Errorf("parseReadtestBits64(%q) = %#x, want nearest-even raw port %#x", input, got, want)
		}
	}

	for _, input := range []string{"1.2345678901234567890123456789012345", "1e6145", "1e-6177", "NaN123"} {
		got, err := parseReadtestBits128(input)
		if err != nil {
			t.Fatalf("parseReadtestBits128(%q): %v", input, err)
		}
		want, _ := goportReadtestBID128FromString(input, 0)
		if got != want {
			t.Errorf("parseReadtestBits128(%q) = %x, want nearest-even raw port %x", input, got, want)
		}
	}
}

// TestReadtestDecimalRowComparatorIsBitExact anchors the CMP_FUZZYSTATUS
// decimal-row comparator (readtestDecimalRowEqual) to the Intel readtest.c
// check32/check64/check128 exact-bit meaning. Every "reject" row is a pair
// that a value-level comparison would accept; if any of them starts passing,
// the gate has been relaxed below Intel readtest.c strictness.
func TestReadtestDecimalRowComparatorIsBitExact(t *testing.T) {
	cases := []struct {
		name     string
		format   string
		expected string
		got      string // decimal literal or [bits]; converted like a dispatch result
		want     bool
		detail   string
	}{
		// --- rejections: same value, different encoding. A pass here means
		// the exact-bit compare degraded into a value compare.
		{"cohort64", "decimal64", "+16E+0", "+160E-1", false,
			"equal-value cohort members (16E0 vs 160E-1) must stay distinct under check64"},
		{"cohort32", "decimal32", "1E+2", "+10E+1", false,
			"equal-value cohort members (1E2 vs 10E1) must stay distinct under check32"},
		{"zero quantum high", "decimal128", "+0E+6111", "+0E+6100", false,
			"zeros with different quantum must stay distinct under check128"},
		{"zero quantum low", "decimal128", "+0E-6176", "+0E+0", false,
			"zeros with different quantum must stay distinct under check128"},
		{"signed zero", "decimal64", "+0E+0", "-0E+0", false,
			"+0 and -0 must stay distinct in the exact-bit branch (only CMP_EQUALSTATUS may equate them)"},
		{"nan payload", "decimal64", "[7c00000000000001]", "[7c00000000000002]", false,
			"NaN payloads must be compared exactly"},
		{"nan signaling bit", "decimal64", "[7c00000000000001]", "[7e00000000000001]", false,
			"quiet vs signaling NaN with the same payload must stay distinct"},
		{"plain value", "decimal64", "1E+0", "2E+0", false,
			"different values must never compare equal"},
		// --- acceptances: if any of these fails, the comparator became
		// stricter than Intel readtest.c (over-rigid gates hide real rows
		// behind spurious failures and invite blanket skips).
		{"identical bits", "decimal64", "[0000000000000001]", "[0000000000000001]", true,
			"identical bits must compare equal"},
		{"same cohort spelling 64", "decimal64", "+15E-1", "1.5", true,
			"different spellings of the same cohort member must compare equal after the from_string round-trip"},
		{"same cohort spelling 32", "decimal32", "1E+2", "0.01E+4", true,
			"leading-zero/exponent respellings of the same cohort member must compare equal"},
		{"comma literal vs canonical", "decimal128", "[0,5]", "[00000000000000000000000000000005]", true,
			"the upstream getop128 comma form must compare equal to the canonical 32-nybble form"},
		{"canonical vs comma literal", "decimal128", "[00000000000000000000000000000005]", "[0,5]", true,
			"the canonical 32-nybble form must compare equal to the upstream getop128 comma form"},
		{"realistic comma literal", "decimal128", "[2080000000000000,0]", "[20800000000000000000000000000000]", true,
			"high-word comma literals must compare equal to their canonical form"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotBits := comparatorStrengthBits(t, tc.format, tc.got, 0)
			equal, err := readtestDecimalRowEqual(tc.format, tc.expected, gotBits, 0, goportReadtestStringBackend)
			if err != nil {
				t.Fatalf("readtestDecimalRowEqual(%q, %q, %q): %v", tc.format, tc.expected, gotBits, err)
			}
			if equal != tc.want {
				t.Errorf("readtestDecimalRowEqual(%q, %q, %q) = %v, want %v — gate less strict/stricter than Intel readtest.c: %s",
					tc.format, tc.expected, gotBits, equal, tc.want, tc.detail)
			}
		})
	}
}

// TestReadtestQuietEqualComparatorMatchesIntelSemantics anchors the
// CMP_EQUALSTATUS comparator (readtestQuietEqual) to the Intel readtest.c
// meaning at lines 1497-1514: exact bits OR the library's own
// bid*_quiet_not_equal reporting equality. The acceptance rows pin the
// intentionally looser side (±0, equal-value cohorts); the rejection rows pin
// the still-strict side (NaN needs exact bits).
func TestReadtestQuietEqualComparatorMatchesIntelSemantics(t *testing.T) {
	cases := []struct {
		name     string
		format   string
		expected string
		got      string
		want     bool
		detail   string
	}{
		// --- Intel-mandated acceptances. A failure here means the gate got
		// stricter than readtest.c and CMP_EQUALSTATUS rows would fail
		// spuriously.
		{"signed zero equal 64", "decimal64", "+0E+0", "-0E+0", true,
			"quiet_not_equal(+0,-0)=0 upstream, so CMP_EQUALSTATUS must accept opposite-sign zeros"},
		{"signed zero equal 32", "decimal32", "+0E+0", "-0E+0", true,
			"quiet_not_equal(+0,-0)=0 upstream, so CMP_EQUALSTATUS must accept opposite-sign zeros"},
		{"cohort equal 64", "decimal64", "+16E+0", "+160E-1", true,
			"CMP_EQUALSTATUS is a value comparison for non-NaN results; equal-value cohort members must pass"},
		{"cohort equal 128", "decimal128", "+16E+0", "+160E-1", true,
			"CMP_EQUALSTATUS is a value comparison for non-NaN results; equal-value cohort members must pass"},
		{"zero quantum equal", "decimal128", "0E+6100", "0E+6111", true,
			"CMP_EQUALSTATUS compares zeros by value regardless of quantum"},
		{"identical nan bits", "decimal64", "[7c00000000000001]", "[7c00000000000001]", true,
			"bit-identical NaNs must pass through the exact-bit branch"},
		// --- Intel-mandated rejections. A pass here means the quiet-equal
		// fallback was widened beyond readtest.c (e.g. treating any two NaNs
		// as equal), which silently drops NaN payload coverage.
		{"nan payload mismatch", "decimal64", "[7c00000000000001]", "[7c00000000000002]", false,
			"quiet_not_equal is true for NaN operands, so NaN results require an exact bit match"},
		{"nan signaling mismatch", "decimal64", "[7c00000000000001]", "[7e00000000000001]", false,
			"quiet vs signaling NaN must not compare equal under CMP_EQUALSTATUS"},
		{"nan payload mismatch 128", "decimal128", "[7c000000000000000000000000000001]", "[7c000000000000000000000000000002]", false,
			"quiet_not_equal is true for NaN operands, so NaN results require an exact bit match"},
		{"different values", "decimal64", "1E+0", "2E+0", false,
			"different values must never compare equal"},
		{"different zero vs nonzero", "decimal32", "0E+0", "1E-10", false,
			"zero and nonzero must never compare equal"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotBits := comparatorStrengthBits(t, tc.format, tc.got, 0)
			equal, err := readtestQuietEqual(tc.format, tc.expected, gotBits, 0, goportReadtestStringBackend, goportReadtestGeneratedSigned)
			if err != nil {
				t.Fatalf("readtestQuietEqual(%q, %q, %q): %v", tc.format, tc.expected, gotBits, err)
			}
			if equal != tc.want {
				t.Errorf("readtestQuietEqual(%q, %q, %q) = %v, want %v — gate less strict/stricter than Intel readtest.c: %s",
					tc.format, tc.expected, gotBits, equal, tc.want, tc.detail)
			}
		})
	}
}

// TestReadtestToStringComparatorRequiresExactCohort anchors the to_string row
// comparator (readtestToStringRowEqual) to readtest.c:1453-1489: the produced
// string is round-tripped through the library's own from_string and compared
// as exact bits, and the round-trip status is surfaced so the caller can fold
// it into the operation status like the upstream *pfpsf accumulation.
func TestReadtestToStringComparatorRequiresExactCohort(t *testing.T) {
	// Rejection: a produced string denoting the same value in the wrong
	// cohort must fail. If this passes, to_string output checking degraded
	// into a value comparison.
	equal, _, err := readtestToStringRowEqual("decimal64", "+15E+0", "+150E-1", 0, goportReadtestStringBackend)
	if err != nil {
		t.Fatalf("readtestToStringRowEqual: %v", err)
	}
	if equal {
		t.Error("readtestToStringRowEqual accepted a wrong-cohort string (+150E-1 for +15E+0) — to_string gate less strict than Intel readtest.c")
	}

	// Acceptance: spelling variants of the same cohort member must pass.
	equal, status, err := readtestToStringRowEqual("decimal64", "15", "+15E+0", 0, goportReadtestStringBackend)
	if err != nil {
		t.Fatalf("readtestToStringRowEqual: %v", err)
	}
	if !equal {
		t.Error("readtestToStringRowEqual rejected an exact-cohort spelling variant — gate stricter than Intel readtest.c")
	}
	if normalizeReadtestStatus(status) != "00" {
		t.Errorf("exact round-trip status = %q, want 00", status)
	}

	// Status surfacing: a produced string whose round-trip is inexact must
	// report the inexact flag so the caller's status comparison sees it.
	// Dropping this replays the upstream *pfpsf accumulation bug readtest.c
	// avoids (readtest.c:1453-1457).
	equal, status, err = readtestToStringRowEqual("decimal64", "5000000000000000E-15", "5.0000000000000000001", 0, goportReadtestStringBackend)
	if err != nil {
		t.Fatalf("readtestToStringRowEqual: %v", err)
	}
	if !equal {
		t.Error("readtestToStringRowEqual rejected a value that rounds to the expected bits — round-trip parse broken")
	}
	if normalizeReadtestStatus(status) != "20" {
		t.Errorf("inexact round-trip status = %q, want 20 — dropping the round-trip status relaxes the status gate", status)
	}
	combined, err := readtestCombineStatus("00", status)
	if err != nil {
		t.Fatalf("readtestCombineStatus: %v", err)
	}
	if normalizeReadtestStatus(combined) != "20" {
		t.Errorf("combined status = %q, want 20", combined)
	}
}

// TestReadtestSecondaryOutputComparatorIsExact anchors the frexp/modf
// secondary-output comparator (readtestSecondaryOutputEqual) to the upstream
// i1 != i2 / R64_1 != B64 / check128(R_1, B) checks (readtest.c:1480,1489).
// Dropping or loosening the secondary compare silently un-tests the second
// output of every frexp/modf row while all case counts stay identical.
func TestReadtestSecondaryOutputComparatorIsExact(t *testing.T) {
	frexpOperands := []string{"+1E+9", "10"}
	sec := readtestSecondaryOutput{Kind: readtestSecondaryInt, OperandIndex: 1, Int: 10}
	if ok, err := readtestSecondaryOutputEqual(sec, frexpOperands); err != nil || !ok {
		t.Errorf("frexp secondary int match = (%v, %v), want (true, nil)", ok, err)
	}
	sec.Int = 9
	if ok, err := readtestSecondaryOutputEqual(sec, frexpOperands); err != nil || ok {
		t.Errorf("frexp secondary int mismatch accepted (ok=%v, err=%v) — gate less strict than Intel readtest.c i1 != i2", ok, err)
	}

	exactBits, _ := goportReadtestBID64FromString("+3E+0", 0)
	cohortBits, _ := goportReadtestBID64FromString("+30E-1", 0)
	if exactBits == cohortBits {
		t.Fatal("test precondition broken: +3E+0 and +30E-1 must encode differently")
	}
	modfOperands := []string{"+3.5E+0", "+3E+0"}
	secDec := readtestSecondaryOutput{Kind: readtestSecondaryDec64, OperandIndex: 1, Bits64: exactBits}
	if ok, err := readtestSecondaryOutputEqual(secDec, modfOperands); err != nil || !ok {
		t.Errorf("modf secondary exact-bit match = (%v, %v), want (true, nil)", ok, err)
	}
	secDec.Bits64 = cohortBits
	if ok, err := readtestSecondaryOutputEqual(secDec, modfOperands); err != nil || ok {
		t.Errorf("modf secondary accepted an equal-value wrong-quantum integral part (ok=%v, err=%v) — gate less strict than Intel readtest.c R64_1 != B64", ok, err)
	}

	secDec.OperandIndex = 5
	if _, err := readtestSecondaryOutputEqual(secDec, modfOperands); err == nil {
		t.Error("out-of-range secondary operand index must error, not silently pass")
	}
}

// TestReadtestStatusComparatorDistinguishesFlags anchors the status
// normalization used by every generated runner's status comparison. Statuses
// are distinct flag sets; any normalization that collapses different flags
// (or fails to equate spelling variants of the same flags) breaks the
// expected_status != *pfpsf mirror of readtest.c.
func TestReadtestStatusComparatorDistinguishesFlags(t *testing.T) {
	if normalizeReadtestStatus("00") == normalizeReadtestStatus("20") {
		t.Error("status 00 and 20 normalized equal — status gate no longer distinguishes flags")
	}
	if normalizeReadtestStatus("01") == normalizeReadtestStatus("05") {
		t.Error("status 01 and 05 normalized equal — status gate no longer distinguishes flags")
	}
	if normalizeReadtestStatus("120") == normalizeReadtestStatus("20") {
		t.Error("status 120 and 20 normalized equal — normalization must not truncate high flag bits")
	}
	for _, spelling := range []string{"0x20", "0X20", " 20 ", "20"} {
		if normalizeReadtestStatus(spelling) != "20" {
			t.Errorf("normalizeReadtestStatus(%q) = %q, want 20 — spelling variants of the same flags must normalize equal", spelling, normalizeReadtestStatus(spelling))
		}
	}
	if normalizeReadtestStatus("5") != "05" {
		t.Errorf("normalizeReadtestStatus(\"5\") = %q, want 05", normalizeReadtestStatus("5"))
	}
	combined, err := readtestCombineStatus("01", "20")
	if err != nil {
		t.Fatalf("readtestCombineStatus: %v", err)
	}
	if normalizeReadtestStatus(combined) != "21" {
		t.Errorf("readtestCombineStatus(01, 20) = %q, want 21 — status accumulation must OR flags like the upstream *pfpsf", combined)
	}
}

// TestReadtest128BitLiteralParserRejectsInvalidForms anchors the 128-bit
// literal parser (parseReadtestBits128) and normalizer used across the
// dispatch and comparison paths. Loosening them lets malformed or truncated
// 128-bit rows pass (or silently equal unrelated values) instead of failing
// the row like the upstream sscanf-based getop128.
func TestReadtest128BitLiteralParserRejectsInvalidForms(t *testing.T) {
	invalid := []string{
		"[5]",                                 // no high/low split
		"[0000000000000005]",                  // 16 nybbles, no comma: not a 128-bit literal
		"[000000000000000000000000000000005]", // 33 nybbles
		"[,5]",                                // empty high word
		"[5,]",                                // empty low word
		"[zz,5]",                              // non-hex high word
		"[0,zz]",                              // non-hex low word
	}
	for _, literal := range invalid {
		if _, err := parseReadtestBits128(literal); err == nil {
			t.Errorf("parseReadtestBits128(%q) accepted an invalid 128-bit literal — malformed rows must fail, not pass", literal)
		}
	}

	if normalizeReadtestBits("[0,5]", 128) != normalizeReadtestBits("[00000000000000000000000000000005]", 128) {
		t.Error("comma form [0,5] must normalize equal to its canonical 32-nybble form")
	}
	if normalizeReadtestBits("[0,5]", 128) == normalizeReadtestBits("[0,6]", 128) {
		t.Error("[0,5] and [0,6] normalized equal — 128-bit literal comparison lost the low word")
	}
	if normalizeReadtestBits("[1,0]", 128) == normalizeReadtestBits("[0,1]", 128) {
		t.Error("[1,0] and [0,1] normalized equal — 128-bit literal comparison lost the high/low word split")
	}
	// An invalid 16-nybble no-comma literal must not alias a valid comma
	// literal through the generic zero-trim fallback.
	if normalizeReadtestBits("[0000000000000005]", 128) == normalizeReadtestBits("[0,5]", 128) {
		t.Error("invalid 16-nybble literal aliased a valid 128-bit literal — normalization widened beyond getop128")
	}
}

// TestReadtestExpectedLiteralParseHonorsRowRounding anchors that the
// expected-literal parse (readtestValueBits) forwards the row rounding mode
// to the backend from_string, mirroring readtest.c get_test which parses the
// expected field after get_ops restores the row rounding. Forcing a fixed
// rounding here would silently alter every non-RN row whose expected
// literal needs rounding.
func TestReadtestExpectedLiteralParseHonorsRowRounding(t *testing.T) {
	const literal = "0.7777777777" // needs coefficient rounding in decimal32
	bitsRN, _, err := readtestValueBits("decimal32", literal, 0, goportReadtestStringBackend)
	if err != nil {
		t.Fatalf("readtestValueBits(RN): %v", err)
	}
	bitsRZ, _, err := readtestValueBits("decimal32", literal, 3, goportReadtestStringBackend)
	if err != nil {
		t.Fatalf("readtestValueBits(RZ): %v", err)
	}
	if bitsRN == bitsRZ {
		t.Errorf("readtestValueBits parsed %q identically under RN and RZ (%s) — row rounding no longer reaches the expected-literal parse", literal, bitsRN)
	}
	directRN, _ := goportReadtestBID32FromString(literal, 0)
	if normalizeReadtestBits(bitsRN, 32) != normalizeReadtestBits(fmt.Sprintf("[%08x]", directRN), 32) {
		t.Errorf("readtestValueBits RN parse %s diverges from the backend from_string %08x", bitsRN, directRN)
	}
	directRZ, _ := goportReadtestBID32FromString(literal, 3)
	if normalizeReadtestBits(bitsRZ, 32) != normalizeReadtestBits(fmt.Sprintf("[%08x]", directRZ), 32) {
		t.Errorf("readtestValueBits RZ parse %s diverges from the backend from_string %08x", bitsRZ, directRZ)
	}
}

// readtestComparatorBindingReport holds the static-analysis outcome of scanning
// one generated readtest source: whether its declared case-decision entry
// function still exists and which anchored comparators it fails to reach
// through a call expression.
type readtestComparatorBindingReport struct {
	runnerName    string
	runnerPresent bool
	missingCalls  []string
}

// problems renders the report as human-readable binding violations; an empty
// slice means the generated source still routes through every anchored
// comparator.
func (r readtestComparatorBindingReport) problems() []string {
	var problems []string
	if r.runnerName != "" && !r.runnerPresent {
		problems = append(problems, fmt.Sprintf("case-decision entry function %s no longer exists", r.runnerName))
	}
	for _, name := range r.missingCalls {
		problems = append(problems, fmt.Sprintf("anchored comparator %s is not invoked in any result-consuming position on a live path reachable from %s", name, r.runnerName))
	}
	return problems
}

// evaluateReadtestComparatorBinding statically checks that a generated readtest
// source keeps invoking the anchored comparator helpers in result-consuming
// positions on live paths reachable from its case-decision entry function. It
// is deliberately a pure function of the source text — no repo file access and
// no devtools import — so the negative self-test below can feed it a synthetic
// relaxed source and prove the check has real detection power.
//
// The analysis is reachability-based, not file-scoped and not identifier-appearance
// based: starting from the body of the entry function named by runnerName (a
// top-level receiverless function that must exist), it follows same-file
// unqualified function calls transitively (cycle-safe). An anchored helper only
// counts when it is the callee of an actual call expression inside a function
// reachable from that entry AND the call sits in a consumed position: a
// condition (if/for condition, switch tag), the RHS of an assignment or short
// declaration with at least one non-blank left-hand side, a return value, an
// argument to another call, or an operand of a unary/binary expression. A bare
// call statement and an all-blank assignment (`_ = cmp(...)`, `_, _ = cmp(...)`)
// discard the result and do not count. A same-file helper extraction (runner →
// intermediate helper → comparator) still passes — reachability follows every
// call, consumed or not; only the anchored comparator's own call site must
// consume its result. Known invalid arrangements still fail: a bare identifier
// reference is not a call, a comparator call parked in a function the entry never
// reaches is not reachable, and a discarded-result call is not consumed.
//
// What this binding pins — and what it does not: it pins that the anchored
// comparators are invoked in consumed positions on live reachable paths.
// Whether the consumed result actually feeds the final pass/fail verdict (an
// assigned-but-never-used variable, a condition whose outcome changes nothing)
// is dataflow, not syntax, and stays an accepted residual limit of this
// anchor; that axis is covered by the standing requirement to preserve the
// readtest comparator semantics and by testgen diff review.
//
// Function literals count as reachable only when their execution is
// evident from the call site itself: immediately invoked (the literal is the
// callee), or passed as an argument to a *.Run(...) selector call — the
// runners' t.Run(id, func(...){...}) shape, where the testing package
// guarantees the callback runs. An arbitrary callee taking a closure argument
// gives no static guarantee it ever invokes it, so every other argument
// position is conservatively excluded — a parked closure (assigned, declared,
// blank-discarded) or a closure handed to a never-invoking callee cannot
// add unrelated comparator calls to the reachable set. If a future runner
// shape trips this exclusion, the binding fails loudly toward over-alarm, not
// relaxation.
//
// Two documented boundaries of the analysis:
//   - Dead statements: calls inside *obvious* dead statements are not counted
//     — statements after a return in the same block (BlockStmt or CaseClause
//     body) and the then-block of a literal `if false`. This catches the
//     realistic accidental relaxation where a template edit early-returns a
//     relaxed compare and strands the comparator block as dead code. No further
//     control-flow analysis is attempted (panic/os.Exit termination, constant
//     folding, the else of `if true`, unreachable-branch solving); a dead
//     branch beyond these two shapes is an accepted residual limit.
//   - The *.Run selector match is purely syntactic: with go/ast alone there is
//     no type information, so the receiver is not verified to be a
//     *testing.T/B. A generator that actively emits a same-file no-op type
//     with a Run method to route an unrelated closure would not be caught; that
//     class sits outside this verification scope (accidental loss of strictness
//     plus simple substitutes) and is an accepted limitation.
//
// Reachability deliberately stops at the file boundary — each
// generated file is checked against its own (entry, required-comparators)
// contract, so a cross-file hop cannot smuggle the verdict away from the
// anchored helpers of the file that owns it.
func evaluateReadtestComparatorBinding(src, runnerName string, requiredCalls []string) (readtestComparatorBindingReport, error) {
	if runnerName == "" {
		return readtestComparatorBindingReport{}, fmt.Errorf("binding entry function name must not be empty")
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "generated.go", src, 0)
	if err != nil {
		return readtestComparatorBindingReport{}, err
	}
	// Index the file's top-level receiverless functions so the walk below can
	// follow same-file calls; the anchored comparators are all same-package
	// helpers invoked as unqualified idents.
	decls := map[string]*ast.FuncDecl{}
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil {
			decls[fn.Name.Name] = fn
		}
	}
	entry, runnerPresent := decls[runnerName]
	called := map[string]bool{}
	visited := map[string]bool{}
	var walk func(fn *ast.FuncDecl)
	walk = func(fn *ast.FuncDecl) {
		if fn == nil || fn.Body == nil || visited[fn.Name.Name] {
			return
		}
		visited[fn.Name.Name] = true
		// Function literals are only descended into when their execution is
		// evident from the call site: immediately invoked (CallExpr.Fun is
		// the literal), or passed as an argument to a *.Run(...) selector
		// call — the runners' t.Run(id, func(...){...}) shape, where the
		// testing package guarantees the callback runs. An arbitrary callee
		// (ignore(func(){...})) gives no static guarantee it invokes its
		// argument, so all other argument positions stay unwired; likewise a
		// parked literal — assigned, declared, or blank-discarded — is dead
		// code. Descending into either would let an unrelated closure inside the
		// runner satisfy the binding. The Run match is name-only (no type
		// info in go/ast), an accepted limitation documented above.
		//
		// Obvious dead statements are excluded too: statements after a return
		// in the same statement list and the then-block of a literal
		// `if false` never execute, so calls inside them must not satisfy the
		// binding — that is exactly where an accidental template relaxation
		// (early-return of a relaxed compare) strands the old comparator block.
		// A return inside a nested block (e.g. `if cond { return }`) kills
		// only its own list, never the statements following the if in the
		// outer block. No deeper control-flow analysis is attempted.
		wired := map[*ast.FuncLit]bool{}
		dead := map[ast.Node]bool{}
		// consumed marks expressions whose value is used by their parent —
		// the consumed positions listed in the doc comment. Inspect is
		// pre-order, so a parent marks its consuming child expressions before
		// the child CallExpr is visited.
		consumed := map[ast.Node]bool{}
		markDeadAfterReturn := func(stmts []ast.Stmt) {
			returned := false
			for _, stmt := range stmts {
				if returned {
					dead[stmt] = true
					continue
				}
				if _, ok := stmt.(*ast.ReturnStmt); ok {
					returned = true
				}
			}
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if dead[n] {
				return false
			}
			switch node := n.(type) {
			case *ast.BlockStmt:
				markDeadAfterReturn(node.List)
			case *ast.CaseClause:
				markDeadAfterReturn(node.Body)
			case *ast.IfStmt:
				if ident, ok := node.Cond.(*ast.Ident); ok && ident.Name == "false" {
					dead[node.Body] = true
				}
				consumed[node.Cond] = true
			case *ast.ForStmt:
				if node.Cond != nil {
					consumed[node.Cond] = true
				}
			case *ast.SwitchStmt:
				if node.Tag != nil {
					consumed[node.Tag] = true
				}
			case *ast.AssignStmt:
				// Only an assignment that binds at least one non-blank
				// left-hand side consumes its RHS; `_ = cmp(...)` and
				// `_, _ = cmp(...)` discard the result.
				nonBlank := false
				for _, lhs := range node.Lhs {
					if ident, ok := lhs.(*ast.Ident); !ok || ident.Name != "_" {
						nonBlank = true
					}
				}
				if nonBlank {
					for _, rhs := range node.Rhs {
						consumed[rhs] = true
					}
				}
			case *ast.ReturnStmt:
				for _, result := range node.Results {
					consumed[result] = true
				}
			case *ast.BinaryExpr:
				consumed[node.X] = true
				consumed[node.Y] = true
			case *ast.UnaryExpr:
				consumed[node.X] = true
			case *ast.ParenExpr:
				if consumed[node] {
					consumed[node.X] = true
				}
			case *ast.FuncLit:
				// Inspect visits the enclosing CallExpr before its children,
				// so a literal wired into a call is already marked here.
				return wired[node]
			case *ast.CallExpr:
				for _, arg := range node.Args {
					consumed[arg] = true
				}
				switch fun := node.Fun.(type) {
				case *ast.FuncLit:
					wired[fun] = true
				case *ast.SelectorExpr:
					if fun.Sel.Name == "Run" {
						for _, arg := range node.Args {
							if lit, ok := arg.(*ast.FuncLit); ok {
								wired[lit] = true
							}
						}
					}
				case *ast.Ident:
					// Reachability follows every same-file call — a bare
					// helper call still executes its body — but only a
					// consumed call site satisfies a required comparator.
					if consumed[node] {
						called[fun.Name] = true
					}
					if next, ok := decls[fun.Name]; ok {
						walk(next)
					}
				}
			}
			return true
		})
	}
	if runnerPresent {
		walk(entry)
	}
	report := readtestComparatorBindingReport{runnerName: runnerName, runnerPresent: runnerPresent}
	for _, name := range requiredCalls {
		if !called[name] {
			report.missingCalls = append(report.missingCalls, name)
		}
	}
	return report, nil
}

// TestReadtestComparatorBindingRoutesThroughAnchoredHelpers is the binding half
// of this anchor. The semantic tests above pin how strict the shared comparator
// bodies are; this test pins that the checked-in generated runner and dispatch
// sources still invoke those bodies in result-consuming positions on live
// reachable paths, so a generator regression cannot replace the pinned
// comparators with a less strict inline comparison while keeping every count,
// reproducibility, and native gate green. Whether the consumed result feeds
// the final verdict is dataflow beyond this syntax check (see the
// evaluateReadtestComparatorBinding doc). Each required comparator here is
// pinned for semantics by a sibling test named in the comment beside it.
func TestReadtestComparatorBindingRoutesThroughAnchoredHelpers(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve anchor source path")
	}
	dir := filepath.Dir(thisFile)

	// The comparators the two generated case runners must invoke in consumed
	// positions on live reachable paths. Each is semantically anchored by the
	// sibling test cited beside it, so the semantic anchor plus this call-site
	// binding together pin: the comparator body is strict AND the runner
	// invokes it where its result is consumed.
	coreComparators := []string{
		"readtestDecimalRowEqual",      // TestReadtestDecimalRowComparatorIsBitExact
		"readtestQuietEqual",           // TestReadtestQuietEqualComparatorMatchesIntelSemantics
		"readtestToStringRowEqual",     // TestReadtestToStringComparatorRequiresExactCohort
		"readtestSecondaryOutputEqual", // TestReadtestSecondaryOutputComparatorIsExact
		"readtestCombineStatus",        // TestReadtestStatusComparatorDistinguishesFlags
		"normalizeReadtestStatus",      // TestReadtestStatusComparatorDistinguishesFlags
		"normalizeReadtestBits",        // TestReadtest128BitLiteralParserRejectsInvalidForms
	}

	targets := []struct {
		file     string
		runner   string
		required []string
	}{
		// Case runners: the case-decision entry function must exist and must
		// route every decimal/quiet-equal/to-string/secondary/status verdict
		// through the anchored comparators.
		{"generated_readtest_goport_cases_test.go", "TestGeneratedReadCasesGoPort", coreComparators},
		{"generated_readtest_cases_native_test.go", "TestGeneratedReadCases", coreComparators},
		// Dispatch tables: the 128-bit dispatch entry must still parse operands
		// through the anchored parseReadtestBits128 (pinned by
		// TestReadtest128BitLiteralParserRejectsInvalidForms), so a loosened
		// inline 128-bit parse cannot let malformed operands through.
		{"generated_readtest_goport_dispatch_test.go", "goportReadtestGeneratedBID128", []string{"parseReadtestBits128"}},
		{"generated_readtest_dispatch_native.go", "nativeReadtestGeneratedBID128", []string{"parseReadtestBits128"}},
	}

	for _, tgt := range targets {
		tgt := tgt
		t.Run(tgt.file, func(t *testing.T) {
			// These sources are checked in, so a missing file is a hard
			// failure, never a skip: a generator that stops emitting a runner
			// must not be able to silently disable its own binding. The native
			// sources are gated behind //go:build cgo && bid754_native and are
			// not compiled into the default test build, which is exactly why
			// the binding reads them as text and parses them here — a compiled
			// reference could not reach them from this untagged anchor.
			src, err := os.ReadFile(filepath.Join(dir, tgt.file))
			if err != nil {
				t.Fatalf("read checked-in generated source %s: %v", tgt.file, err)
			}
			report, err := evaluateReadtestComparatorBinding(string(src), tgt.runner, tgt.required)
			if err != nil {
				t.Fatalf("parse generated source %s: %v", tgt.file, err)
			}
			for _, p := range report.problems() {
				t.Errorf("%s: %s — the generated readtest path no longer invokes the anchored comparator in a consumed position on a live path; a relaxed inline comparison would pass every count/reproducibility/native gate, so fix the generator, never this anchor", tgt.file, p)
			}
		})
	}
}

// TestReadtestComparatorBindingCheckDetectsRelaxation is the detection-power
// self-test for the binding above. It exercises evaluateReadtestComparatorBinding
// against synthetic in-memory sources only — it never mutates a repo file — and
// proves the check catches the exact relaxation it exists to stop.
func TestReadtestComparatorBindingCheckDetectsRelaxation(t *testing.T) {
	const runner = "TestGeneratedReadCasesGoPort"
	required := []string{"readtestDecimalRowEqual", "readtestQuietEqual"}

	// Positive control: a runner that invokes the anchored comparators in
	// consumed positions (non-blank assignment, then a condition) must yield
	// no problems, so the check is not vacuously failing.
	strong := `package bid754
func TestGeneratedReadCasesGoPort() {
	equal, _ := readtestDecimalRowEqual(f, e, g, r, b)
	if !equal {
		panic(1)
	}
	ok, _ := readtestQuietEqual(f, e, g, r, b, d)
	if !ok {
		panic(1)
	}
}`
	report, err := evaluateReadtestComparatorBinding(strong, runner, required)
	if err != nil {
		t.Fatalf("parse strong synthetic source: %v", err)
	}
	if problems := report.problems(); len(problems) != 0 {
		t.Fatalf("binding check flagged a compliant runner: %v", problems)
	}

	// Refactor-tolerance control: a same-file helper extraction (runner →
	// intermediate helper → comparator), including a call cycle on the way,
	// must still pass — the binding follows the same-file call graph
	// transitively and must not loop forever on recursion.
	extracted := `package bid754
func TestGeneratedReadCasesGoPort() {
	verifyRow()
}
func verifyRow() {
	deeper()
}
func deeper() {
	verifyRow() // cycle back — walk must stay cycle-safe
	equal, _ := readtestDecimalRowEqual(f, e, g, r, b)
	if !equal {
		panic(1)
	}
	ok, _ := readtestQuietEqual(f, e, g, r, b, d)
	if !ok {
		panic(1)
	}
}`
	report, err = evaluateReadtestComparatorBinding(extracted, runner, required)
	if err != nil {
		t.Fatalf("parse extracted synthetic source: %v", err)
	}
	if problems := report.problems(); len(problems) != 0 {
		t.Fatalf("binding check flagged a helper-extraction refactor that still routes through the comparators: %v", problems)
	}

	// Wired-closure control: comparator calls inside function literals whose
	// execution is evident from the call site — passed to a *.Run(...) call
	// (the real runners' t.Run(id, func(...){...}) shape) or immediately
	// invoked — must count as reachable, or the binding would reject every
	// checked-in runner.
	closureRouted := `package bid754
func TestGeneratedReadCasesGoPort(t *testing.T) {
	t.Run("case", func(t *testing.T) {
		equal, _ := readtestDecimalRowEqual(f, e, g, r, b)
		if !equal {
			panic(1)
		}
	})
	func() {
		ok, _ := readtestQuietEqual(f, e, g, r, b, d)
		if !ok {
			panic(1)
		}
	}()
}`
	report, err = evaluateReadtestComparatorBinding(closureRouted, runner, required)
	if err != nil {
		t.Fatalf("parse closure-routed synthetic source: %v", err)
	}
	if problems := report.problems(); len(problems) != 0 {
		t.Fatalf("binding check flagged a runner that routes through call-wired closures like the real t.Run runners: %v", problems)
	}

	// Negative control: the exact failure mode this anchor exists to catch — the
	// runner drops the anchored comparator calls and inlines a less strict value
	// compare. The dangling `_ = readtestDecimalRowEqual` reference proves the
	// check is CallExpr-based: an identifier mention is not a call and must not
	// satisfy the binding.
	relaxed := `package bid754
func TestGeneratedReadCasesGoPort() {
	if got == expected { // inline value compare, no anchored comparator
		_ = got
	}
	_ = readtestDecimalRowEqual // referenced, never called
}`
	report, err = evaluateReadtestComparatorBinding(relaxed, runner, required)
	if err != nil {
		t.Fatalf("parse relaxed synthetic source: %v", err)
	}
	problems := report.problems()
	if len(problems) == 0 {
		t.Fatal("binding check passed a relaxed runner that dropped every anchored comparator call — the check has no detection power and cannot preserve the anchor")
	}
	joined := strings.Join(problems, "; ")
	for _, name := range required {
		if !strings.Contains(joined, name) {
			t.Errorf("relaxed-runner report %q does not name the dropped comparator %s", joined, name)
		}
	}

	// Unreachable-function control: the runner inlines a relaxed compare while an unused
	// same-file function still calls the anchored comparators. A file-scoped
	// scan would be satisfied by those parked calls; the reachability walk
	// must not be, because the runner never reaches them.
	unreachableAlternate := `package bid754
func TestGeneratedReadCasesGoPort() {
	if got == expected { // inline value compare, no anchored comparator
		_ = got
	}
}
func unreachableAlternate() { // never called by the runner
	equal, _ := readtestDecimalRowEqual(f, e, g, r, b)
	_ = equal
	ok, _ := readtestQuietEqual(f, e, g, r, b, d)
	_ = ok
}`
	report, err = evaluateReadtestComparatorBinding(unreachableAlternate, runner, required)
	if err != nil {
		t.Fatalf("parse unreachable-function synthetic source: %v", err)
	}
	problems = report.problems()
	if len(problems) == 0 {
		t.Fatal("binding check passed a runner whose comparator calls live only in an unreachable alternate function — unreachable calls must not satisfy the binding")
	}
	joined = strings.Join(problems, "; ")
	for _, name := range required {
		if !strings.Contains(joined, name) {
			t.Errorf("unreachable-function report %q does not name the unreachable comparator %s", joined, name)
		}
	}

	// Parked-closure control: the runner inlines a relaxed compare and parks the
	// comparator calls inside a function literal that is never invoked (blank
	// discard). The literal's body is dead code inside a reachable function, so
	// descending into it unconditionally would let this unrelated closure pass; the
	// binding must reject it.
	parked := `package bid754
func TestGeneratedReadCasesGoPort() {
	if got == expected { // inline value compare, no anchored comparator
		_ = got
	}
	_ = func() { // parked closure, never invoked and never passed to a call
		equal, _ := readtestDecimalRowEqual(f, e, g, r, b)
		_ = equal
		ok, _ := readtestQuietEqual(f, e, g, r, b, d)
		_ = ok
	}
}`
	report, err = evaluateReadtestComparatorBinding(parked, runner, required)
	if err != nil {
		t.Fatalf("parse parked-closure synthetic source: %v", err)
	}
	problems = report.problems()
	if len(problems) == 0 {
		t.Fatal("binding check passed a runner whose comparator calls live only in a parked, never-invoked closure — dead closure bodies must not satisfy the binding")
	}
	joined = strings.Join(problems, "; ")
	for _, name := range required {
		if !strings.Contains(joined, name) {
			t.Errorf("parked-closure report %q does not name the unreachable comparator %s", joined, name)
		}
	}

	// Unused-callback control: the runner inlines a relaxed compare and hands the
	// comparator calls to an arbitrary same-file callee inside a closure the
	// callee never executes. Nothing at the call site guarantees the closure
	// runs — only *.Run(...) arguments carry that guarantee (from the testing
	// package) — so the binding must reject this even though the closure is
	// syntactically a call argument.
	unusedCallback := `package bid754
func TestGeneratedReadCasesGoPort() {
	if got == expected { // inline value compare, no anchored comparator
		_ = got
	}
	ignore(func() { // handed to a no-op callee that never invokes it
		equal, _ := readtestDecimalRowEqual(f, e, g, r, b)
		_ = equal
		ok, _ := readtestQuietEqual(f, e, g, r, b, d)
		_ = ok
	})
}
func ignore(f func()) {}`
	report, err = evaluateReadtestComparatorBinding(unusedCallback, runner, required)
	if err != nil {
		t.Fatalf("parse unused-callback synthetic source: %v", err)
	}
	problems = report.problems()
	if len(problems) == 0 {
		t.Fatal("binding check passed a runner whose comparator calls live only in a closure handed to a never-invoking callee — non-Run callback arguments must not satisfy the binding")
	}
	joined = strings.Join(problems, "; ")
	for _, name := range required {
		if !strings.Contains(joined, name) {
			t.Errorf("unused-callback report %q does not name the unreachable comparator %s", joined, name)
		}
	}

	// Dead-after-return control: the realistic accidental relaxation — a
	// template edit early-returns a relaxed inline compare and the old comparator
	// block is left stranded after the return in the same statement list. The
	// stranded calls never execute, so the binding must reject this.
	deadAfterReturn := `package bid754
func TestGeneratedReadCasesGoPort() bool {
	if got == expected { // inline value compare, early-returned
		return true
	}
	return false
	equal, _ := readtestDecimalRowEqual(f, e, g, r, b) // stranded dead code
	_ = equal
	ok, _ := readtestQuietEqual(f, e, g, r, b, d)
	_ = ok
}`
	report, err = evaluateReadtestComparatorBinding(deadAfterReturn, runner, required)
	if err != nil {
		t.Fatalf("parse dead-after-return synthetic source: %v", err)
	}
	problems = report.problems()
	if len(problems) == 0 {
		t.Fatal("binding check passed a runner whose comparator calls sit after a return in the same block — stranded dead code must not satisfy the binding")
	}
	joined = strings.Join(problems, "; ")
	for _, name := range required {
		if !strings.Contains(joined, name) {
			t.Errorf("dead-after-return report %q does not name the stranded comparator %s", joined, name)
		}
	}

	// If-false control: comparator calls kept only inside a literal
	// `if false` block never execute and must not satisfy the binding.
	deadIfFalse := `package bid754
func TestGeneratedReadCasesGoPort() {
	if got == expected { // inline value compare, no anchored comparator
		_ = got
	}
	if false { // never executes
		equal, _ := readtestDecimalRowEqual(f, e, g, r, b)
		_ = equal
		ok, _ := readtestQuietEqual(f, e, g, r, b, d)
		_ = ok
	}
}`
	report, err = evaluateReadtestComparatorBinding(deadIfFalse, runner, required)
	if err != nil {
		t.Fatalf("parse if-false synthetic source: %v", err)
	}
	problems = report.problems()
	if len(problems) == 0 {
		t.Fatal("binding check passed a runner whose comparator calls live only inside `if false` — a never-executing branch must not satisfy the binding")
	}
	joined = strings.Join(problems, "; ")
	for _, name := range required {
		if !strings.Contains(joined, name) {
			t.Errorf("if-false report %q does not name the dead-branch comparator %s", joined, name)
		}
	}

	// Bare-call control: the runner decides via an inline compare and calls
	// the comparators only as bare expression statements — executed, but the
	// verdict is discarded on the floor. A discarded result must not satisfy
	// the binding.
	bareCall := `package bid754
func TestGeneratedReadCasesGoPort() {
	if got == expected { // inline value compare decides the case
		_ = got
	}
	readtestDecimalRowEqual(f, e, g, r, b) // bare call, result discarded
	readtestQuietEqual(f, e, g, r, b, d)
}`
	report, err = evaluateReadtestComparatorBinding(bareCall, runner, required)
	if err != nil {
		t.Fatalf("parse bare-call synthetic source: %v", err)
	}
	problems = report.problems()
	if len(problems) == 0 {
		t.Fatal("binding check passed a runner that only bare-calls the comparators and discards their results — a discarded verdict must not satisfy the binding")
	}
	joined = strings.Join(problems, "; ")
	for _, name := range required {
		if !strings.Contains(joined, name) {
			t.Errorf("bare-call report %q does not name the discarded comparator %s", joined, name)
		}
	}

	// All-blank-assignment control: assigning every comparator result to the
	// blank identifier is the same discard dressed as an assignment and must
	// not satisfy the binding either.
	blankOnly := `package bid754
func TestGeneratedReadCasesGoPort() {
	if got == expected { // inline value compare decides the case
		_ = got
	}
	_, _ = readtestDecimalRowEqual(f, e, g, r, b)
	_, _ = readtestQuietEqual(f, e, g, r, b, d)
}`
	report, err = evaluateReadtestComparatorBinding(blankOnly, runner, required)
	if err != nil {
		t.Fatalf("parse all-blank synthetic source: %v", err)
	}
	problems = report.problems()
	if len(problems) == 0 {
		t.Fatal("binding check passed a runner that blank-assigns every comparator result — an all-blank assignment must not satisfy the binding")
	}
	joined = strings.Join(problems, "; ")
	for _, name := range required {
		if !strings.Contains(joined, name) {
			t.Errorf("all-blank report %q does not name the discarded comparator %s", joined, name)
		}
	}

	// Acknowledged residual limit (documented in the
	// evaluateReadtestComparatorBinding doc): binding a non-blank variable
	// satisfies the consumed-position rule even if that variable is then only
	// blank-discarded — assigned-but-unused is dataflow beyond this syntax
	// check, covered by the standing requirement to preserve the readtest
	// comparator semantics and testgen diff review. This case pins where the
	// check's power ends so a future tightening is a deliberate decision, not
	// an accident.
	assignedUnused := `package bid754
func TestGeneratedReadCasesGoPort() {
	equal, _ := readtestDecimalRowEqual(f, e, g, r, b) // non-blank LHS: consumed position
	_ = equal                                          // ... but never actually used
	ok, _ := readtestQuietEqual(f, e, g, r, b, d)
	_ = ok
}`
	report, err = evaluateReadtestComparatorBinding(assignedUnused, runner, required)
	if err != nil {
		t.Fatalf("parse assigned-unused synthetic source: %v", err)
	}
	if problems := report.problems(); len(problems) != 0 {
		t.Fatalf("assigned-but-unused now fails the binding (%v) — the consumed-position boundary changed; if intentional, update the doc comment and this acknowledged case together", problems)
	}

	// Missing-runner control: if the case-decision entry function is removed or
	// renamed, the binding must fail loudly even when the helpers are still
	// called elsewhere, so the runner cannot be quietly detached from the
	// anchored comparators.
	noRunner := `package bid754
func somethingElse() {
	readtestDecimalRowEqual()
	readtestQuietEqual()
}`
	report, err = evaluateReadtestComparatorBinding(noRunner, runner, required)
	if err != nil {
		t.Fatalf("parse no-runner synthetic source: %v", err)
	}
	problems = report.problems()
	if len(problems) == 0 {
		t.Fatal("binding check passed a source with no case-decision entry function — a removed or renamed runner must fail the binding")
	}
	if !strings.Contains(strings.Join(problems, "; "), runner) {
		t.Errorf("no-runner report %v does not name the missing runner %s", problems, runner)
	}
}
