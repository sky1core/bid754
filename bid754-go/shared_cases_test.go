package bid754

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/testspec"
)

func TestGeneratedDectestSuiteSelection(t *testing.T) {
	spec := loadSharedSpecForTest(t)

	testCases := []struct {
		pattern string
		want    int
	}{
		{pattern: "ds*.decTest", want: 1},
		{pattern: "dd*.decTest", want: 33},
		{pattern: "dq*.decTest", want: 33},
		{pattern: "*.decTest", want: 10},
	}

	for _, tc := range testCases {
		files, err := getTestFiles("../devtools/tests", tc.pattern)
		if err != nil {
			t.Fatalf("getTestFiles(%q) error: %v", tc.pattern, err)
		}
		if len(files) != tc.want {
			t.Fatalf("getTestFiles(%q) returned %d files, want %d", tc.pattern, len(files), tc.want)
		}
	}

	if len(spec.DectestSuites) != 4 {
		t.Fatalf("generated suite count = %d, want 4", len(spec.DectestSuites))
	}

	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddQuantize.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddToIntegral.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddCopySign.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddClass.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddSameQuantum.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddMin.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddMax.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddMinMag.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddMaxMag.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddCompareTotal.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddCompareTotalMag.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddAbs.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddPlus.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddMinus.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddFMA.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddLogB.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddScaleB.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddRemainder.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddRemainderNear.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddNextToward.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddNextPlus.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddNextMinus.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddReduce.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqQuantize.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqToIntegral.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqCopySign.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqClass.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqSameQuantum.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqMin.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqMax.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqMinMag.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqMaxMag.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqCompareTotal.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqCompareTotalMag.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqAbs.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqPlus.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqMinus.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqFMA.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqLogB.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqScaleB.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqRemainder.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqRemainderNear.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqNextToward.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqNextPlus.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqNextMinus.decTest")
	assertGeneratedSuiteContains(t, spec, "*.decTest", "tests/quantize.decTest")
	assertGeneratedSuiteContains(t, spec, "*.decTest", "tests/tointegral.decTest")
	assertGeneratedSuiteContains(t, spec, "*.decTest", "tests/tointegralx.decTest")
	assertGeneratedSuiteContains(t, spec, "*.decTest", "tests/comparesig.decTest")
	assertGeneratedSuiteMissing(t, spec, "*.decTest", "tests/randoms.decTest")
	assertGeneratedSuiteMissing(t, spec, "*.decTest", "tests/testall.decTest")
}

func assertGeneratedSuiteContains(t *testing.T, spec testspec.SharedSpec, pattern, wantFile string) {
	t.Helper()

	for _, suite := range spec.DectestSuites {
		if suite.Pattern != pattern {
			continue
		}
		for _, file := range suite.Files {
			if file == wantFile {
				return
			}
		}
		t.Fatalf("generated suite %q does not include %q", pattern, wantFile)
	}

	t.Fatalf("generated suite %q not found", pattern)
}

func assertGeneratedSuiteMissing(t *testing.T, spec testspec.SharedSpec, pattern, unwantedFile string) {
	t.Helper()

	for _, suite := range spec.DectestSuites {
		if suite.Pattern != pattern {
			continue
		}
		for _, file := range suite.Files {
			if file == unwantedFile {
				t.Fatalf("generated suite %q unexpectedly includes %q", pattern, unwantedFile)
			}
		}
		return
	}

	t.Fatalf("generated suite %q not found", pattern)
}

// FuzzArithmeticPortVsNativeResultOnlyNative is a native-only DIFFERENTIAL
// fuzz complement: it runs the same decTest-shaped case through two
// independent implementations — the Go mechanical port (the goport decTest
// adapters) and the decNumber native reference bridge — and fails on any
// result-value divergence or one-sided acceptance divergence. It is an
// auxiliary exploration tool, not a regular generated verification domain: it
// compares result values only (via the shared decTest value comparator), not
// decTest status or IEEE exception flags.
//
// Oracle principle: the comparison baseline is an independent implementation
// the fuzzer's mutated tuple cannot touch. The generated spec's IBM decTest
// fuzz cases are only the SEED source; their Expected field is deliberately
// not part of the fuzz tuple, because a mutated expected value is not an
// oracle (the previous design compared against it and produced false
// failures within seconds of mutation).
func FuzzArithmeticPortVsNativeResultOnlyNative(f *testing.F) {
	spec, err := loadSharedSpecFromDisk()
	if err != nil {
		f.Fatalf("load shared spec: %v", err)
	}

	for _, tc := range spec.FuzzCases {
		f.Add(tc.TestType, tc.Operation, tc.Operands[0], tc.Operands[1], tc.Precision, tc.RoundingMode, tc.MaxExponent, tc.MinExponent, tc.Clamp)
	}

	f.Fuzz(func(t *testing.T, testType, op, operand1, operand2 string, precision int, roundingMode string, maxExponent, minExponent, clamp int) {
		requireNative(t)

		// Shared supported-surface gate. Both engines must support the
		// combination, or the case is skipped:
		// - only the fixed-width BID suites map onto the Go port (the goport
		//   decTest leg excludes General's arbitrary-precision context), and
		//   only their canonical decTest contexts are meaningful — the native
		//   general bridge would happily compute at any mutated precision the
		//   fixed-width port cannot represent;
		// - dectestGoportSkipReason is the goport decTest leg's own
		//   supported-surface classifier (oracle ops, rounding modes, and the
		//   documented decNumber-vs-BID divergence classes such as NaN-payload
		//   precedence), reused so this fuzzer and the generated leg cannot
		//   drift apart. Without this gate the native bridge's leniency (it
		//   parses malformed operands to NaN and returns success) floods the
		//   run with noise instead of divergences.
		if !isGoportDectestRunnerSuite(testType) {
			t.Skipf("test type %q outside the shared port/native fixed-width surface", testType)
		}
		if !fuzzCanonicalDecTestContext(testType, precision, maxExponent, minExponent, clamp) {
			t.Skipf("non-canonical %s context (precision=%d emax=%d emin=%d clamp=%d)", testType, precision, maxExponent, minExponent, clamp)
		}
		normalizedOp := normalizeDecTestOperation(op)
		operands := []string{operand1, operand2}
		if fuzzDecTestUnaryOperation(normalizedOp) {
			// Unary read-path ops consume one operand; the second tuple slot
			// is meaningless for them and is dropped, matching both engines'
			// one-operand executors.
			operands = operands[:1]
		}
		tc := decTestCase{
			Operation:    op,
			Operands:     operands,
			RoundingMode: roundingMode,
			Precision:    precision,
			MaxExponent:  maxExponent,
			MinExponent:  minExponent,
			Clamp:        clamp,
		}
		if reason, ok := dectestGoportSkipReason(nil, tc); ok {
			t.Skipf("outside the shared goport oracle surface: %s", reason)
		}
		for _, operand := range operands {
			if !fuzzDecTestParseableOperand(operand) {
				// Harness artifact, not a divergence: on a conversion-syntax
				// operand the native C bridge discards the decNumber result and
				// returns the hardcoded string "NaN" (see the
				// strcpy(out, "NaN") arms in dectest_native.go), while the Go
				// port renders its own parse (e.g. Intel keeps the consumed
				// sign: "-" and "-0 " both parse to -NaN). The goport decTest
				// leg skips the same class as conversion_syntax_divergence,
				// but its check reads the decTest file's flag annotations,
				// which a fuzz tuple does not carry — so the operand grammar
				// is checked directly here. Malformed-input behavior is pinned
				// by the string/codec domains, not by this arithmetic
				// differential.
				t.Skipf("operand %q is not decTest-parseable (conversion-syntax divergence class)", operand)
			}
		}
		if testType == "decimal32" && !fuzzDecTestUnaryOperation(normalizedOp) && strings.ToLower(roundingMode) != "half_even" {
			// Harness artifact, not a divergence: the native decimal32 binary
			// bridge (bid754_decimal32_op) takes no rounding argument and
			// always computes at decNumber's default half-even, while the Go
			// port applies the case rounding mode.
			t.Skip("native decimal32 binary bridge is fixed at half_even rounding")
		}

		// The port case flags are deliberately dropped: this fuzzer is
		// result-only (the goport decTest leg owns flag parity).
		portResult, _, portErr := runDectestGoportCase(tc, testType)
		var nativeExec decTestExecResult
		var nativeErr error
		if fuzzDecTestUnaryOperation(normalizedOp) {
			nativeExec, nativeErr = executeDecTestReadOperation(tc, testType)
		} else {
			nativeExec, nativeErr = executeDecTestOperation(tc, testType)
		}

		switch {
		case portErr == nil && nativeErr == nil:
			if !compareDecimalResults(nativeExec.Result, portResult) {
				t.Fatalf("port/native result divergence: %s %q %s %q (test_type=%s rounding=%s): native %q, port %q",
					op, operand1, normalizedOp, operand2, testType, roundingMode, nativeExec.Result, portResult)
			}
		case portErr != nil && nativeErr != nil:
			t.Skipf("unsupported on both engines: port=%v native=%v", portErr, nativeErr)
		case portErr != nil:
			t.Fatalf("acceptance divergence: %s %q %q (test_type=%s rounding=%s): native succeeded with %q, Go port failed: %v",
				op, operand1, operand2, testType, roundingMode, nativeExec.Result, portErr)
		default:
			t.Fatalf("acceptance divergence: %s %q %q (test_type=%s rounding=%s): Go port succeeded with %q, native failed: %v",
				op, operand1, operand2, testType, roundingMode, portResult, nativeErr)
		}
	})
}

// fuzzDecTestParseableOperand reports whether a fuzz operand is inside the
// strict decNumber (GDA) numeric grammar both engines parse as a value: after
// stripping decTest quoting, an optional single sign followed by either
// inf/infinity, NaN/sNaN with an optional ASCII-digit payload, or ASCII
// digits with at most one point and an optional exponent that has at least
// one ASCII digit — no whitespace and no other characters anywhere. Anything
// outside this grammar is decNumber conversion syntax, where the native
// bridge substitutes a hardcoded "NaN" string instead of an implementation
// result, so no engine comparison is possible (parseComparableDecimal is NOT
// reused here: it is the result comparator, whose deliberate leniency —
// whitespace trimming, qNaN spelling, non-digit payloads — is exactly what
// must be rejected on the operand side).
func fuzzDecTestParseableOperand(raw string) bool {
	s := strings.Trim(raw, "'\"") // the executors strip decTest quoting the same way
	if s == "" {
		return false
	}
	if s[0] == '+' || s[0] == '-' {
		s = s[1:]
	}
	lower := strings.ToLower(s)
	if lower == "inf" || lower == "infinity" {
		return true
	}
	for _, prefix := range []string{"snan", "nan"} {
		if strings.HasPrefix(lower, prefix) {
			return fuzzASCIIDigits(lower[len(prefix):], true)
		}
	}
	mantissa := lower
	if idx := strings.IndexByte(lower, 'e'); idx >= 0 {
		mantissa = lower[:idx]
		exponent := lower[idx+1:]
		if len(exponent) > 0 && (exponent[0] == '+' || exponent[0] == '-') {
			exponent = exponent[1:]
		}
		if !fuzzASCIIDigits(exponent, false) {
			return false
		}
	}
	if strings.Count(mantissa, ".") > 1 {
		return false
	}
	return fuzzASCIIDigits(strings.Replace(mantissa, ".", "", 1), false)
}

// fuzzASCIIDigits reports whether s is ASCII digits only; emptyOK allows the
// empty string (a bare NaN payload).
func fuzzASCIIDigits(s string, emptyOK bool) bool {
	if s == "" {
		return emptyOK
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// fuzzDecTestUnaryOperation mirrors the generated dispatch arity split: the
// read-path ops take one operand (native executeDecTestReadOperation, goport
// string/to-integral executors); everything else on the shared oracle surface
// is a two-operand operation.
func fuzzDecTestUnaryOperation(normalizedOp string) bool {
	switch normalizedOp {
	case "tosci", "toeng", "tointegral", "tointegralx":
		return true
	default:
		return false
	}
}

// fuzzCanonicalDecTestContext reports whether the context parameters are the
// canonical decTest directives of the fixed-width suite (the values the ds/dd/
// dq suites pin). The Go port computes in the fixed-width format regardless,
// so any other context would make the native general bridge compute something
// the port cannot express — a harness mismatch, not a divergence.
func fuzzCanonicalDecTestContext(testType string, precision, maxExponent, minExponent, clamp int) bool {
	switch testType {
	case "decimal32":
		return precision == 7 && maxExponent == 96 && minExponent == -95 && clamp == 1
	case "decimal64":
		return precision == 16 && maxExponent == 384 && minExponent == -383 && clamp == 1
	case "decimal128":
		return precision == 34 && maxExponent == 6144 && minExponent == -6143 && clamp == 1
	default:
		return false
	}
}

func loadSharedSpecForTest(t *testing.T) testspec.SharedSpec {
	t.Helper()
	spec, err := loadSharedSpecFromDisk()
	if err != nil {
		t.Fatalf("load shared spec: %v", err)
	}
	return spec
}

func loadSharedSpecFromDisk() (testspec.SharedSpec, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return testspec.SharedSpec{}, fmt.Errorf("resolve shared_cases_test.go path")
	}
	return testspec.LoadGenerated(filepath.Join(filepath.Dir(currentFile), "..", "devtools", "generated", "testspec", "spec_index.json"))
}
