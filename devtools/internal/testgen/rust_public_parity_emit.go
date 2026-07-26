package testgen

import (
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
	"github.com/sky1core/bid754/devtools/tools/go2rs/apiemit"
)

// rustTmpl renders a Rust code template using @NAME@ placeholders instead of
// fmt's positional %s/%d verbs. Several templates below repeat the same
// substitution (e.g. go_symbol) across many diagnostic strings. Named
// placeholders keep substitution sites self-describing and avoid positional
// argument-count drift.
func rustTmpl(tmpl string, kv ...string) string {
	if len(kv)%2 != 0 {
		panic(fmt.Sprintf("rustTmpl: odd kv count (%d)", len(kv)))
	}
	return strings.NewReplacer(kv...).Replace(tmpl)
}

// parityWidth is the per-Decimal-width record the shape emitters below use to
// stay a single mechanical template reused across widths, mirroring apiemit's
// own widthSpec. Every field is an explicit
// literal on decimal64/decimal32/decimal128ParityWidth (never derived by
// string-splicing at a call site), matching the project's no-implicit-fallback
// convention. All three widths have a parityWidth record;
// parityWidthForBidgoFunc resolves the right one from each emitted row's
// Bid32/Bid64/Bid128 census prefix.
type parityWidth struct {
	selfType   string // "Decimal64" | "Decimal32" | "Decimal128" -- Rust value-type name
	corpus     string // "CORPUS_64" | "CORPUS_32" | "CORPUS_128" -- the value-bit-pattern corpus const
	pairs      string // "PAIRS_64" | "PAIRS_32" | "PAIRS_128"
	triples    string // "TRIPLES_64" | "TRIPLES_32" | "TRIPLES_128"
	minQuantum int64  // inclusive cohort-exponent lower bound
	maxQuantum int64  // inclusive cohort-exponent upper bound
	precision  int    // maximum decimal coefficient digit count

	// Auxiliary census bidgo_function names this width's composed shapes
	// (sign, signaling_eq_compose/signaling_not_eq_compose, parse's NaN
	// fast-path skip, the trait-parity test) need to resolve a port beyond
	// their own row's BidgoFunction.
	bidgoIsNaN       string
	bidgoIsZero      string
	bidgoIsSigned    string
	bidgoSignalingGE string
	bidgoSignalingLE string
	bidgoQuietEqual  string
	bidgoQuietLess   string
	bidgoQuietGreat  string

	// bidgoFromString is the bidgo function this width's raw from-string port
	// entrypoint used by the parse family.
	bidgoFromString string
}

var decimal64ParityWidth = parityWidth{
	selfType: "Decimal64", corpus: "CORPUS_64", pairs: "PAIRS_64", triples: "TRIPLES_64",
	minQuantum: -398, maxQuantum: 369, precision: 16,
	bidgoIsNaN: "Bid64IsNaN", bidgoIsZero: "Bid64IsZero", bidgoIsSigned: "Bid64IsSigned",
	bidgoSignalingGE: "Bid64SignalingGreaterEqual", bidgoSignalingLE: "Bid64SignalingLessEqual",
	bidgoQuietEqual: "Bid64QuietEqual", bidgoQuietLess: "Bid64QuietLess", bidgoQuietGreat: "Bid64QuietGreater",
	bidgoFromString: "Bid64FromString",
}

var decimal32ParityWidth = parityWidth{
	selfType: "Decimal32", corpus: "CORPUS_32", pairs: "PAIRS_32", triples: "TRIPLES_32",
	minQuantum: -101, maxQuantum: 90, precision: 7,
	bidgoIsNaN: "Bid32IsNaN", bidgoIsZero: "Bid32IsZero", bidgoIsSigned: "Bid32IsSigned",
	bidgoSignalingGE: "Bid32SignalingGreaterEqual", bidgoSignalingLE: "Bid32SignalingLessEqual",
	bidgoQuietEqual: "Bid32QuietEqual", bidgoQuietLess: "Bid32QuietLess", bidgoQuietGreat: "Bid32QuietGreater",
	bidgoFromString: "Bid32FromStringRaw",
}

// decimal128ParityWidth is the Decimal128 width record. bidgoFromString names
// Bid128FromString directly (not a special-cased raw wrapper): unlike width
// 64 (bid754-rs/src/generated/bid64_from_string.rs is pub(crate), so the
// parity leg's from-string rows route through the doc(hidden)
// bid754::bid64_from_string_raw compat function instead -- see
// resolveFromStringRawPort's doc comment), width 128's
// generated::bid128_string::bid128_from_string is fully pub (verified
// against the actual generated source, mirroring width 32's own
// already-pub bid32_from_string_raw), so it is called directly like every
// other shape in this file.
var decimal128ParityWidth = parityWidth{
	selfType: "Decimal128", corpus: "CORPUS_128", pairs: "PAIRS_128", triples: "TRIPLES_128",
	minQuantum: -6176, maxQuantum: 6111, precision: 34,
	bidgoIsNaN: "Bid128IsNaN", bidgoIsZero: "Bid128IsZero", bidgoIsSigned: "Bid128IsSigned",
	bidgoSignalingGE: "Bid128SignalingGreaterEqual", bidgoSignalingLE: "Bid128SignalingLessEqual",
	bidgoQuietEqual: "Bid128QuietEqual", bidgoQuietLess: "Bid128QuietLess", bidgoQuietGreat: "Bid128QuietGreater",
	bidgoFromString: "Bid128FromString",
}

// parityWidthForBidgoFunc resolves the width record from a census
// bidgo_function name's own "Bid64"/"Bid32"/"Bid128" prefix. This works
// uniformly for every row including Context's (Add64BIDWithContext's
// BidgoFunction is Bid64AddWithFlags; Add128BIDWithContext's is Bid128Add),
// where row.RustOwner is "Context" rather than a Decimal<w> type name and so
// cannot itself be used to resolve the width. Bid128 is checked before Bid64/
// Bid32 would ever be ambiguous with it (none of their names share the
// "Bid128" prefix), so a fixed check order is safe here.
func parityWidthForBidgoFunc(bidgoFn, goSymbol string) (parityWidth, error) {
	switch {
	case strings.HasPrefix(bidgoFn, "Bid128"):
		return decimal128ParityWidth, nil
	case strings.HasPrefix(bidgoFn, "Bid64"):
		return decimal64ParityWidth, nil
	case strings.HasPrefix(bidgoFn, "Bid32"):
		return decimal32ParityWidth, nil
	default:
		return parityWidth{}, fmt.Errorf("rust public parity: cannot resolve a decimal width from bidgo_function %q (go_symbol %q); only Bid32*/Bid64*/Bid128* are supported today", bidgoFn, goSymbol)
	}
}

// Width-generic helpers.
//
// pubFrom/portArg/pubBitsExpr/portBitsExpr/resultFmtSpec are this file's
// counterparts of apiemit's widthSpec.selfArg/wrapResult (devtools/tools/
// go2rs/apiemit/rust_templates.go): the single generation point for the
// corpus-bit-pattern <-> public-Decimal128 <-> port-BID_UINT128 conversion
// on this side of the independent-mapping boundary. They reuse the
// to_port128/from_port128 free functions
// (emitRustParityStaticHelpers) -- this file's OWN [u8;16]<->BID_UINT128
// conversion, independent of apiemit's private bid_uint128_from_le_bytes/
// to_le_bytes -- rather than calling into apiemit's helpers, preserving the
// independent-mapping guarantee (a shared-conversion-helper bug
// would otherwise be invisible to this gate).
//
// bitsLen/pairsLen/triplesLen/nonNaNLen centralize per-width corpus-size
// selection in one place.
func (w parityWidth) pubFrom(varName string) string {
	if w.selfType == "Decimal128" {
		return "Decimal128::from_le_bytes(" + varName + ")"
	}
	return w.selfType + "::from_bits(" + varName + ")"
}

func (w parityWidth) portArg(varName string) string {
	if w.selfType == "Decimal128" {
		return "to_port128(" + varName + ")"
	}
	return varName
}

func (w parityWidth) pubBitsExpr(pubVar string) string {
	if w.selfType == "Decimal128" {
		return pubVar + ".to_le_bytes()"
	}
	return pubVar + ".to_bits()"
}

func (w parityWidth) portBitsExpr(portVar string) string {
	if w.selfType == "Decimal128" {
		return "from_port128(" + portVar + ")"
	}
	return portVar
}

func (w parityWidth) resultFmtSpec() string {
	if w.selfType == "Decimal128" {
		return "{:?}"
	}
	return "{:#x}"
}

func rustCanonicalQNaNBits(w parityWidth) string {
	switch w.selfType {
	case "Decimal32":
		return "0x7c00_0000u32"
	case "Decimal64":
		return "0x7c00_0000_0000_0000u64"
	case "Decimal128":
		return "[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x7c]"
	default:
		panic("rust public parity: canonical qNaN requested for unsupported width " + w.selfType)
	}
}

// opFmtSpec is the diagnostic format specifier for a corpus operand of width
// w in a failure message: [u8; 16] has no LowerHex impl, so a width-128
// operand uses `{:x?}` (Debug-hex) instead of `{:#x}`.
func (w parityWidth) opFmtSpec() string {
	if w.selfType == "Decimal128" {
		return "{:x?}"
	}
	return "{:#x}"
}

func (w parityWidth) bitsLen(corpus publicParityCorpus) int {
	switch w.selfType {
	case "Decimal32":
		return len(corpus.Bits32)
	case "Decimal128":
		return len(corpus.Bits128)
	default:
		return len(corpus.Bits64)
	}
}

func (w parityWidth) pairsLen(corpus publicParityCorpus) int {
	switch w.selfType {
	case "Decimal32":
		return len(corpus.Pairs32)
	case "Decimal128":
		return len(corpus.Pairs128)
	default:
		return len(corpus.Pairs64)
	}
}

func (w parityWidth) triplesLen(corpus publicParityCorpus) int {
	switch w.selfType {
	case "Decimal32":
		return len(corpus.Triples32)
	case "Decimal128":
		return len(corpus.Triples128)
	default:
		return len(corpus.Triples64)
	}
}

func (w parityWidth) nonNaNLen(corpus publicParityCorpus) int {
	switch w.selfType {
	case "Decimal32":
		return corpus.NonNaN32
	case "Decimal128":
		return corpus.NonNaN128
	default:
		return corpus.NonNaN64
	}
}

// portCallStmt renders the Rust statement(s) binding pr (the port call's raw
// result) and praw (the raised-flags u32) from a direct
// bid754::generated::* port call, dispatching on whether bidgoFn uses the
// pfpsf-output-parameter convention (apiemit.PortPfpsf) or the tuple-return
// convention every other operation uses -- the parity-runner-side sibling of
// apiemit's flagsCallStmt (devtools/tools/go2rs/apiemit/decimal_emit.go),
// same independence argument as PortPathFor/PortBoolResult (this is a
// structural fact about the generated function's signature, not a
// flag/rounding semantic mapping).
func portCallStmt(bidgoFn, module, fn string, args []string) string {
	call := fmt.Sprintf("bid754::generated::%s::%s", module, fn)
	if apiemit.PortPfpsf(bidgoFn) {
		allArgs := append(append([]string{}, args...), "&mut praw")
		return fmt.Sprintf("let mut praw = 0u32;\n        let pr = %s(%s);", call, strings.Join(allArgs, ", "))
	}
	return fmt.Sprintf("let (pr, praw) = %s(%s);", call, strings.Join(args, ", "))
}

// -- NaN-literal round-trip corpus (payload/sign format coverage) --
//
// The vm_string/display parity shape deliberately skips NaN operands (a NaN
// display is hand-plumbing over the port, so it cannot be compared against a
// raw port to_string call), and the parse-family shapes only check
// is_nan()/is_signaling() on NaN inputs, not the payload/sign BITS. That
// leaves the width-specific NaN payload+sign encode/decode
// (apiemit's parse_decimal<w>_nan / format_decimal<w>_nan, and especially
// Decimal128's native-u128 110-bit path spanning both BID_UINT128 words)
// with no gate. This corpus closes that gap with a width-parameterized
// round-trip property check (see emitRustNaNRoundtripTest), replacing the
// prior reliance on bid754-rs/tests/public_api_smoke.rs (which round-trips
// Decimal64 NaNs only).

// parityNaNCase is one NaN-literal corpus entry with its INDEPENDENTLY
// computed expected canonical display and sign/signaling classification
// (parityNaNExpectation, a re-derivation of the documented grammar+canonical
// rules, never a call into the Rust format under test).
type parityNaNCase struct {
	Input     string
	Display   string
	Negative  bool
	Signaling bool
}

// canonicalNaNLimit returns width w's canonical NaN-payload display limit
// (10^6 / 10^15 / 10^33), as a big.Int so the 128-bit 10^33 fits. Mirrors
// bid754-go/types_bid_nan_payload.go's per-width canonical limits (and the
// apiemit-side decimal<w> constants). The parseable-payload max is limit-1
// for every width (parse_uint_payload/parse_u128_payload reject >= limit via
// the fast path), so any payload < limit both parses and, if nonzero,
// renders; only 0 is suppressed.
func canonicalNaNLimit(w parityWidth) *big.Int {
	switch w.selfType {
	case "Decimal32":
		return big.NewInt(1_000_000)
	case "Decimal64":
		return big.NewInt(1_000_000_000_000_000)
	default: // Decimal128
		return new(big.Int).Exp(big.NewInt(10), big.NewInt(33), nil)
	}
}

// parityNaNExpectation is the independent Go oracle for a NaN literal's
// parse+display contract, re-deriving bid754-go/types_bid_nan_payload.go's
// parseBIDNaNLiteral grammar and payload canonicalization rules directly
// (case-insensitive [+|-](nan|qnan|snan)[digits], leading ASCII space/tab
// ignored, all other whitespace rejected, leading zeros stripped; payload
// rendered on display only when nonzero and below the width's canonical
// limit). It returns ok=false if the
// input is not a valid NaN literal that parses at this width (payload out of
// the parseable [0, limit) range), so the corpus builder can reject a
// mis-authored entry at generation time.
func parityNaNExpectation(input string, w parityWidth) (parityNaNCase, bool) {
	limit := canonicalNaNLimit(w)
	trimmed := strings.TrimLeft(input, " \t")
	negative := false
	rest := trimmed
	switch {
	case strings.HasPrefix(rest, "+"):
		rest = rest[1:]
	case strings.HasPrefix(rest, "-"):
		negative = true
		rest = rest[1:]
	}
	lower := strings.ToLower(rest)
	signaling := false
	var payload string
	switch {
	case strings.HasPrefix(lower, "snan"):
		signaling = true
		payload = rest[4:]
	case strings.HasPrefix(lower, "qnan"):
		payload = rest[4:]
	case strings.HasPrefix(lower, "nan"):
		payload = rest[3:]
	default:
		return parityNaNCase{}, false
	}
	for _, r := range payload {
		if r < '0' || r > '9' {
			return parityNaNCase{}, false
		}
	}
	payload = strings.TrimLeft(payload, "0")
	payloadVal := new(big.Int)
	if payload != "" {
		v, ok := payloadVal.SetString(payload, 10)
		if !ok {
			return parityNaNCase{}, false
		}
		payloadVal = v
	}
	if payloadVal.Cmp(limit) >= 0 {
		// Not parseable at this width via the NaN fast path.
		return parityNaNCase{}, false
	}
	sign := "+"
	if negative {
		sign = "-"
	}
	token := "NaN"
	if signaling {
		token = "SNaN"
	}
	rendered := ""
	if payloadVal.Sign() != 0 {
		rendered = payloadVal.String()
	}
	return parityNaNCase{Input: input, Display: sign + token + rendered, Negative: negative, Signaling: signaling}, true
}

// publicParityNaNCases builds the NaN round-trip corpus for width w: a base
// set shared by every width (sign/signaling/spelling/whitespace/leading-zero/
// zero-payload variety, small payloads that fit BID32's 6 digits) plus
// width-specific large payloads. For Decimal128 the width-specific set MUST
// include payloads >= 2^64 so the native-u128 path that spans BOTH
// BID_UINT128 words (lo = payload as u64, hi = (payload >> 64) as u64 | combo)
// is actually exercised, not just the low-word-only range 32/64 also cover.
func publicParityNaNCases(w parityWidth) []parityNaNCase {
	inputs := []string{
		"NaN", "-NaN", "+NaN",
		"SNaN", "-SNaN", "+qnan", "snan", "QNAN",
		"NaN42", "-SNaN7", "qnan123", "NaN999",
		"SNaN000123",    // leading zeros -> canonical "123"
		"NaN0", "-NaN0", // zero payload -> suppressed on display
		" \tNaN55", // leading ASCII space/tab -> trimmed
	}
	switch w.selfType {
	case "Decimal64":
		inputs = append(inputs,
			"NaN999999999999999",  // 10^15 - 1 (canonical max for BID64)
			"SNaN123456789012345", // 15-digit signaling payload
		)
	case "Decimal128":
		inputs = append(inputs,
			"NaN18446744073709551616",               // 2^64 exactly: lo word 0, first hi-word payload bit set
			"SNaN18446744073709551617",              // 2^64 + 1: spans both words
			"-NaN123456789012345678901234567890123", // 33-digit negative payload, < 10^33
			"NaN999999999999999999999999999999999",  // 10^33 - 1 (canonical max for BID128)
		)
	default: // Decimal32
		inputs = append(inputs,
			"NaN999999", // 10^6 - 1 (canonical max for BID32)
			"SNaN314159",
		)
	}
	cases := make([]parityNaNCase, 0, len(inputs))
	for _, in := range inputs {
		c, ok := parityNaNExpectation(in, w)
		if !ok {
			panic(fmt.Sprintf("rust public parity: NaN corpus input %q is not a valid parseable NaN literal for %s", in, w.selfType))
		}
		cases = append(cases, c)
	}
	return cases
}

// rustParityCaseSummary is the resolved (function name, apiemit shape tag,
// pinned case count) for one emitted census row, used to build the driver
// table and the expected-count constants.
type rustParityCaseSummary struct {
	GoSymbol string
	FuncName string
	Shape    string
	Cases    int
}

// emitRustParityFile renders the complete generated Rust parity runner: the
// independent flag/rounding/class mappings, the shared bit-literal corpus
// (reused verbatim from the Go leg's buildPublicParityCorpus), one function
// per emitted census row, the driver test, and the PartialEq/PartialOrd
// trait-parity test that extends beyond the per-symbol census.
func emitRustParityFile(rows []rustParityInventoryRow, constRows []rustConstInventoryRow, corpus publicParityCorpus, masks publicParityExceptionMasks) ([]byte, error) {
	var b strings.Builder
	b.WriteString(genmarker.Line("testgen") + "\n")
	b.WriteString(rustParityFileDoc)
	b.WriteString("\n")
	b.WriteString("use bid754::{Context, Decimal128, Decimal32, Decimal64, RoundingMode};\n\n")

	emitRustParityStaticHelpers(&b, masks)
	emitRustMixedFMAFusednessSentinelRows(&b)
	emitRustParityCorpusTables(&b, corpus)

	seenNames := map[string]string{}
	var summaries []rustParityCaseSummary
	for _, row := range rows {
		funcName, cases, err := emitRustParityUnit(&b, row, corpus)
		if err != nil {
			return nil, err
		}
		if prior, dup := seenNames[funcName]; dup {
			return nil, fmt.Errorf("rust public parity: generated function name %q collides between go_symbol %q and %q", funcName, prior, row.GoSymbol)
		}
		seenNames[funcName] = row.GoSymbol
		summaries = append(summaries, rustParityCaseSummary{GoSymbol: row.GoSymbol, FuncName: funcName, Shape: row.Shape, Cases: cases})
	}

	emitRustParityDriver(&b, summaries)
	if err := emitRustTraitParityTest(&b, corpus, decimal64ParityWidth); err != nil {
		return nil, err
	}
	if err := emitRustTraitParityTest(&b, corpus, decimal32ParityWidth); err != nil {
		return nil, err
	}
	if err := emitRustTraitParityTest(&b, corpus, decimal128ParityWidth); err != nil {
		return nil, err
	}

	// NaN payload/sign round-trip coverage for all three widths. The
	// shared NanCase struct is emitted once; each width then gets its own
	// corpus const + test.
	b.WriteString(`/// One NaN-literal round-trip corpus entry: the input literal plus its
/// independently computed expected canonical display and sign/signaling
/// classification (see the Go-side parityNaNExpectation oracle).
struct NanCase {
    input: &'static str,
    display: &'static str,
    negative: bool,
    signaling: bool,
}

`)
	emitRustNaNRoundtripTest(&b, decimal64ParityWidth)
	emitRustNaNRoundtripTest(&b, decimal32ParityWidth)
	emitRustNaNRoundtripTest(&b, decimal128ParityWidth)

	// ZERO/ONE/PI/E constants parity is accounted separately from
	// EXPECTED_PARITY_CASES/EXPECTED_NAN_ROUNDTRIP_CASES_* (its own
	// EXPECTED_CONST_PARITY_CASES_<digits> self-check per width), mirroring
	// the NaN round-trip section's accounting style.
	if err := emitRustConstParityTests(&b, constRows); err != nil {
		return nil, err
	}

	// Flagless-sibling equivalence leg (rust_public_parity_flagless_emit.go):
	// the Rust mirror of the Go leg, accounted separately from
	// EXPECTED_PARITY_CASES like the trait/NaN/constants legs above.
	if err := emitRustParityFlaglessSibling(&b); err != nil {
		return nil, err
	}
	return []byte(b.String()), nil
}

// emitRustMixedFMAFusednessSentinelRows carries the generator-owned, ordered
// mixed-FMA sentinel census into the generated Rust runner. The verification
// anchor compares this literal byte-for-byte with the independently maintained
// pin; the Rust generator never reads that pin.
func emitRustMixedFMAFusednessSentinelRows(b *strings.Builder) {
	rows := ffiMixedFMAFusednessRows()
	fmt.Fprintf(b, "const MIXED_FMA_FUSEDNESS_SENTINEL_ROWS: [&str; %d] = [\n", len(rows))
	for _, row := range rows {
		fmt.Fprintf(b, "    %q,\n", row)
	}
	b.WriteString("];\n\n")
}

// emitRustNaNRoundtripTest emits width w's NAN_CASES_<digits> corpus const
// and the generated_public_api_nan_roundtrip_<digits> test. The test asserts,
// per corpus entry: (1) parse(input) succeeds and is a NaN whose sign and
// signaling bit match the literal; (2) display(parse(input)) equals the
// independently computed canonical string (this is the direct 110-bit
// payload-extraction check for Decimal128); (3) parse(display(parse(input)))
// reproduces the exact same bits -- an oracle-independent fixed-point that
// catches any payload/sign alteration in parse_decimal<w>_nan /
// format_decimal<w>_nan even if the display oracle shared a bug. Accounted
// separately from EXPECTED_PARITY_CASES (like the trait-parity test), with
// its own self-checking EXPECTED_NAN_ROUNDTRIP_CASES_<digits> constant.
func emitRustNaNRoundtripTest(b *strings.Builder, w parityWidth) {
	digits := strings.TrimPrefix(w.selfType, "Decimal")
	cases := publicParityNaNCases(w)

	fmt.Fprintf(b, "const NAN_CASES_%s: &[NanCase] = &[\n", digits)
	for _, c := range cases {
		fmt.Fprintf(b, "    NanCase { input: %q, display: %q, negative: %v, signaling: %v },\n", c.Input, c.Display, c.Negative, c.Signaling)
	}
	b.WriteString("];\n\n")

	fmt.Fprintf(b, "const EXPECTED_NAN_ROUNDTRIP_CASES_%s: usize = %d;\n\n", digits, len(cases))

	pubBitsD := w.pubBitsExpr("d")
	pubBitsD2 := w.pubBitsExpr("d2")
	fmtSpec := w.resultFmtSpec()
	b.WriteString(rustTmpl(`#[test]
fn generated_public_api_nan_roundtrip_@DIGITS@() {
    let mut failures: Vec<String> = Vec::new();
    let mut count = 0usize;
    for nc in NAN_CASES_@DIGITS@ {
        let d = match @SELF@::parse(nc.input) {
            Ok(d) => d,
            Err(e) => {
                failures.push(format!("nan roundtrip @DIGITS@: input {:?}: unexpected parse error {}", nc.input, e));
                continue;
            }
        };
        if !d.is_nan() {
            failures.push(format!("nan roundtrip @DIGITS@: input {:?}: parse result is not a NaN", nc.input));
        }
        if d.is_signaling() != nc.signaling {
            failures.push(format!("nan roundtrip @DIGITS@: input {:?}: signaling mismatch public={} expected={}", nc.input, d.is_signaling(), nc.signaling));
        }
        if d.is_sign_negative() != nc.negative {
            failures.push(format!("nan roundtrip @DIGITS@: input {:?}: sign mismatch public={} expected={}", nc.input, d.is_sign_negative(), nc.negative));
        }
        let shown = d.to_string();
        if shown != nc.display {
            failures.push(format!("nan roundtrip @DIGITS@: input {:?}: display mismatch public={:?} expected={:?}", nc.input, shown, nc.display));
        }
        // Oracle-independent fixed point: parse -> display -> parse must
        // reproduce the exact same bits. A payload/sign bit dropped or
        // altered by parse_decimal@DIGITS@_nan or format_decimal@DIGITS@_nan
        // makes the re-parse diverge here even if the display-string oracle
        // above happened to share the bug.
        match @SELF@::parse(&shown) {
            Ok(d2) => {
                if @PUBBITSD@ != @PUBBITSD2@ {
                    failures.push(format!("nan roundtrip @DIGITS@: input {:?}: parse->display->parse bits diverged public=@FMTSPEC@ reparsed=@FMTSPEC@", nc.input, @PUBBITSD@, @PUBBITSD2@));
                }
            }
            Err(e) => failures.push(format!("nan roundtrip @DIGITS@: input {:?}: re-parse of display {:?} failed: {}", nc.input, shown, e)),
        }
        count += 1;
    }
    assert_eq!(count, EXPECTED_NAN_ROUNDTRIP_CASES_@DIGITS@, "nan roundtrip case count drifted");
    assert!(
        failures.is_empty(),
        "NaN payload/sign round-trip failures ({} total):\n{}",
        failures.len(),
        failures.join("\n")
    );
}
`, "@DIGITS@", digits, "@SELF@", w.selfType, "@PUBBITSD@", pubBitsD, "@PUBBITSD2@", pubBitsD2, "@FMTSPEC@", fmtSpec))
}

// ZERO/ONE/PI/E constants parity.
//
// go2rs apiemit computes each constant's compile-time bits with its own
// independent BID interchange-format encoder (const_bits.go), never the
// generated from_string parser -- so this gate's "const bits == parse
// (documented literal) bits" check is a real cross-check between two
// separately-derived computations (an apiemit typo, a wrong bias/shift in
// const_bits.go, or a manifest literal that drifted from the port's own
// parse result all move this gate), not a tautology against the same code
// path the constant itself was computed from.

// constParityCase is one resolved rust_api_surface_inventory.json "constants" row
// for a single width: the compile-time associated const name (RustConst)
// versus the documented Literal a fresh runtime parse must reproduce.
type constParityCase struct {
	RustConst string
	Literal   string
}

// emitRustConstParityTests groups the emitted constants-inventory rows by width
// and renders one generated_public_api_const_parity_<digits> test per width
// (decimal32ParityWidth/decimal64ParityWidth/decimal128ParityWidth --
// deliberately reusing the SAME three parityWidth records and their
// pubBitsExpr/resultFmtSpec helpers the main parity/NaN-roundtrip sections
// use above, so the to_bits()-vs-to_le_bytes() and {:#x}-vs-{:?} width forks
// stay a single source of truth). Reports an error if any emitted constants row's
// rust_owner does not resolve to one of those three widths (a silently
// dropped row would otherwise just be a smaller generated test with no
// generation-time signal).
func emitRustConstParityTests(b *strings.Builder, constRows []rustConstInventoryRow) error {
	byWidth := map[string][]constParityCase{}
	total := 0
	for _, r := range constRows {
		byWidth[r.RustOwner] = append(byWidth[r.RustOwner], constParityCase{RustConst: r.RustConst, Literal: r.Literal})
		total++
	}
	widths := []parityWidth{decimal32ParityWidth, decimal64ParityWidth, decimal128ParityWidth}
	accounted := 0
	for _, w := range widths {
		cases := byWidth[w.selfType]
		sort.Slice(cases, func(i, j int) bool { return cases[i].RustConst < cases[j].RustConst })
		emitRustConstParityTest(b, w, cases)
		accounted += len(cases)
	}
	if accounted != total {
		return fmt.Errorf("rust public parity: constants verification has %d emitted row(s) but only %d resolved to a known Decimal32/64/128 owner; check rust_api_surface_inventory.json constants[].rust_owner", total, accounted)
	}
	return nil
}

// emitRustConstParityTest renders width w's EXPECTED_CONST_PARITY_CASES_
// <digits> self-check constant and its generated_public_api_const_parity_
// <digits> test: for every case, parse the documented literal fresh and
// assert its bits equal the compile-time associated const's bits (mirrors
// emitRustNaNRoundtripTest's per-width corpus-loop-plus-self-check shape).
func emitRustConstParityTest(b *strings.Builder, w parityWidth, cases []constParityCase) {
	digits := strings.TrimPrefix(w.selfType, "Decimal")
	fmt.Fprintf(b, "const EXPECTED_CONST_PARITY_CASES_%s: usize = %d;\n\n", digits, len(cases))

	b.WriteString(rustTmpl(`#[test]
fn generated_public_api_const_parity_@DIGITS@() {
    let mut failures: Vec<String> = Vec::new();
    let mut count = 0usize;
`, "@DIGITS@", digits))

	fmtSpec := w.resultFmtSpec()
	for _, c := range cases {
		constBits := w.pubBitsExpr(w.selfType + "::" + c.RustConst)
		parsedBits := w.pubBitsExpr("parsed")
		// The documented literal (@LIT@, a quoted Rust &str expression such as
		// "3.141593") is passed as a runtime format! ARGUMENT ({:?} in the
		// message text), never spliced as raw text into the message string's
		// own quotes: splicing @LIT@ directly into "...parse(@LIT@)..." would
		// nest an unescaped '"' inside the outer string literal.
		b.WriteString(rustTmpl(`    match @SELF@::parse(@LIT@) {
        Ok(parsed) => {
            let const_bits = @CONSTBITS@;
            let parsed_bits = @PARSEDBITS@;
            if const_bits != parsed_bits {
                failures.push(format!(
                    "const parity @SELF@::@NAME@: compile-time bits @FMT@ != parse({:?}) bits @FMT@",
                    const_bits, @LIT@, parsed_bits
                ));
            }
        }
        Err(e) => failures.push(format!("const parity @SELF@::@NAME@: parse({:?}) failed: {}", @LIT@, e)),
    }
    count += 1;
`, "@SELF@", w.selfType, "@NAME@", c.RustConst, "@LIT@", fmt.Sprintf("%q", c.Literal), "@CONSTBITS@", constBits, "@PARSEDBITS@", parsedBits, "@FMT@", fmtSpec))
	}

	b.WriteString(rustTmpl(`    assert_eq!(count, EXPECTED_CONST_PARITY_CASES_@DIGITS@, "const parity case count drifted");
    assert!(
        failures.is_empty(),
        "const == parse(literal) parity failures ({} total):\n{}",
        failures.len(),
        failures.join("\n")
    );
}
`, "@DIGITS@", digits))
}

const rustParityFileDoc = `//! Public-API parity gate: exercises every emitted public-API wrapper
//! symbol recorded in devtools/generated/testspec/rust_api_surface_inventory.json
//! ("status": "emitted") against an independent direct call to the same
//! crate::generated::* port function the wrapper itself calls, comparing
//! result bits and mapped exception flags. Mirrors
//! bid754-go/generated_public_parity_dispatch_test.go: the flag-bit and
//! rounding-mode mappings below (map_port_flags, PARITY_MODES) are
//! independent numeric literals, not this crate's own
//! ExceptionFlags::from_bidgo / to_bidgo_rounding converters, so a
//! shared-converter bug between the wrapper and this gate stays visible. This
//! is an architecture-contract gate proving the wrapper routes through the
//! generated port and preserves semantics, not a fifth regular verification
//! domain (same framing as the Go leg).
//!
//! Scope: emitted rows cover Decimal32, Decimal64, and Decimal128 value methods
//! and constructors, plus the
//! per-width Context::add methods. Each shape emitter is parameterized by a
//! parityWidth record resolved from the row's own census bidgo_function prefix
//! (Bid64.../Bid32.../Bid128...), not a hardcoded width, so all widths share
//! one template per shape. The width-128 records carry the pfpsf-pointer flag
//! convention and the to_le_bytes/from_le_bytes accessor names. The generator
//! still reports an error on any emitted row whose bidgo_function does not resolve
//! to one of the three widths rather than guessing at its convention.
//!
//! Invalid-RoundingMode accounting: Go's RoundingMode is an open int, so every
//! public Go surface taking one has a defined behavior for a value outside the
//! five IEEE directions (types_bidgo_invalid_mode.go: FlagInvalidOperation
//! plus either the canonical quiet NaN of the result width or the mirrored
//! NaN-input result of the conversion families). Rust's RoundingMode
//! (bid754-rs/src/generated/api/types.rs) is a closed five-variant enum, so an
//! invalid mode value is not constructible at all: the type system represents
//! the constraint (docs/IEEE754_SPEC.md's constructibility principle), not a
//! runtime branch this gate could exercise.
//!
//! This is reported here rather than silently omitted. The Go generated
//! public-API parity corpus does carry that contract for the arithmetic mode
//! shapes: the generated body of every same-width explicit-mode arithmetic
//! wrapper ({Add,Sub,Mul,Div,Quantize,FMA,ScaleB,Sqrt,RoundIntegralExact}
//! WithMode) and every mixed-width explicit-mode free function ends in a
//! rejection case that calls the wrapper with RoundingMode(99) and compares
//! the result against a pinned canonical quiet-NaN bit literal plus exactly
//! FlagInvalidOperation, preceded by a valid-mode control asserting the case
//! is not vacuous. Those cases are counted inside their own shapes'
//! expectedPublicParityCasesByShape entries in
//! bid754-go/generated_public_parity_cases_test.go, one per wrapper; no total
//! is restated in this comment because a hand-carried count here would drift
//! from the Go leg. The conversion, integer-constructor, string-constructor,
//! and context mode shapes do not carry a generated rejection case yet: their
//! expected rejection value is derived per target type rather than from the
//! result width. The accounting for this Rust leg is unchanged either way:
//! 0 cases affected, because the input does not exist in this API and so
//! there is nothing to skip.
`

func emitRustParityStaticHelpers(b *strings.Builder, masks publicParityExceptionMasks) {
	b.WriteString(`/// bidgo-domain rounding value for round-ties-to-even (the IEEE default),
/// used directly wherever a port call needs a rounding argument but the
/// public wrapper does not expose a mode choice.
const BIDGO_ROUND_NEAREST_EVEN: i64 = 0;

`)
	fmt.Fprintf(b, `/// Maps the raw Intel status word to the public ExceptionFlags bit layout
/// using numeric literals extracted from the pinned Intel bid_functions.h
/// (BID_*_EXCEPTION), returned as a raw u32 comparable against
/// ExceptionFlags::bits(). This gate must not reuse the wrapper's own
/// ExceptionFlags::from_bidgo converter: a shared converter bug would be
/// invisible. bid754-rs/tests/public_api_smoke.rs's hand-written
/// map_bidgo_flags pins the same literals outside this generated file.
fn map_port_flags(raw: u32) -> u32 {
    let mut out: u32 = 0;
    if raw & 0x%02x != 0 {
        out |= bid754::ExceptionFlags::INEXACT.bits();
    }
    if raw & 0x%02x != 0 {
        out |= bid754::ExceptionFlags::UNDERFLOW.bits();
    }
    if raw & 0x%02x != 0 {
        out |= bid754::ExceptionFlags::OVERFLOW.bits();
    }
    if raw & 0x%02x != 0 {
        out |= bid754::ExceptionFlags::DIVISION_BY_ZERO.bits();
    }
    if raw & 0x%02x != 0 {
        out |= bid754::ExceptionFlags::INVALID_OPERATION.bits();
    }
    out
}

`, masks.Inexact, masks.Underflow, masks.Overflow, masks.DivByZero, masks.Invalid)

	b.WriteString(`/// Pairs each public RoundingMode with its port-domain integer as
/// independent numeric literals (mirrors the Go leg's publicParityModes); a
/// wrapper whose to_bidgo_rounding mapping drifts diverges in value here.
const PARITY_MODES: &[(RoundingMode, i64)] = &[
    (RoundingMode::NearestEven, 0),
    (RoundingMode::NearestAway, 4),
    (RoundingMode::TowardZero, 3),
    (RoundingMode::TowardPositive, 2),
    (RoundingMode::TowardNegative, 1),
];

/// Maps the Intel class_t integer (0-9) to the public DecimalClass GDA
/// spelling independently of the crate's own decimal_class_from_bid_class
/// (mirrors the Go leg's publicParityClassName).
fn class_name(c: i64) -> &'static str {
    match c {
        0 => "sNaN",
        1 => "NaN",
        2 => "-Infinity",
        3 => "-Normal",
        4 => "-Subnormal",
        5 => "-Zero",
        6 => "+Zero",
        7 => "+Subnormal",
        8 => "+Normal",
        9 => "+Infinity",
        _ => "NaN",
    }
}

/// Reinterprets a little-endian 128-bit byte pattern as the port operand type
/// (bid754::gen_types::BID_UINT128), matching the value types' 1:1 byte
/// correspondence. A pure conversion via safe from_le_bytes composition, not
/// a flag/rounding mapping, so sharing the field layout with the crate's own
/// (equivalent, but pub(crate) and therefore inaccessible from here) helper
/// does not relax the independent-mapping guarantee.
fn to_port128(bytes: [u8; 16]) -> bid754::gen_types::BID_UINT128 {
    let mut lo = [0u8; 8];
    let mut hi = [0u8; 8];
    lo.copy_from_slice(&bytes[0..8]);
    hi.copy_from_slice(&bytes[8..16]);
    bid754::gen_types::BID_UINT128 {
        lo: u64::from_le_bytes(lo),
        hi: u64::from_le_bytes(hi),
    }
}

fn from_port128(v: bid754::gen_types::BID_UINT128) -> [u8; 16] {
    let mut bytes = [0u8; 16];
    bytes[0..8].copy_from_slice(&v.lo.to_le_bytes());
    bytes[8..16].copy_from_slice(&v.hi.to_le_bytes());
    bytes
}

`)
}

func emitRustParityCorpusTables(b *strings.Builder, corpus publicParityCorpus) {
	b.WriteString("const CORPUS_32: &[u32] = &[\n")
	for _, v := range corpus.Bits32 {
		fmt.Fprintf(b, "    0x%08x,\n", v)
	}
	b.WriteString("];\n\n")

	b.WriteString("const CORPUS_64: &[u64] = &[\n")
	for _, v := range corpus.Bits64 {
		fmt.Fprintf(b, "    0x%016x,\n", v)
	}
	b.WriteString("];\n\n")

	b.WriteString("const CORPUS_128: &[[u8; 16]] = &[\n")
	for _, raw := range corpus.Bits128 {
		b.WriteString("    [")
		for i, by := range raw {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(b, "0x%02x", by)
		}
		b.WriteString("],\n")
	}
	b.WriteString("];\n\n")

	b.WriteString("const PAIRS_32: &[(usize, usize)] = &[\n")
	for _, p := range corpus.Pairs32 {
		fmt.Fprintf(b, "    (%d, %d),\n", p[0], p[1])
	}
	b.WriteString("];\n\n")

	b.WriteString("const PAIRS_64: &[(usize, usize)] = &[\n")
	for _, p := range corpus.Pairs64 {
		fmt.Fprintf(b, "    (%d, %d),\n", p[0], p[1])
	}
	b.WriteString("];\n\n")

	b.WriteString("const PAIRS_128: &[(usize, usize)] = &[\n")
	for _, p := range corpus.Pairs128 {
		fmt.Fprintf(b, "    (%d, %d),\n", p[0], p[1])
	}
	b.WriteString("];\n\n")

	b.WriteString("const TRIPLES_32: &[(usize, usize, usize)] = &[\n")
	for _, t := range corpus.Triples32 {
		fmt.Fprintf(b, "    (%d, %d, %d),\n", t[0], t[1], t[2])
	}
	b.WriteString("];\n\n")

	b.WriteString("const TRIPLES_64: &[(usize, usize, usize)] = &[\n")
	for _, t := range corpus.Triples64 {
		fmt.Fprintf(b, "    (%d, %d, %d),\n", t[0], t[1], t[2])
	}
	b.WriteString("];\n\n")

	b.WriteString("const TRIPLES_128: &[(usize, usize, usize)] = &[\n")
	for _, t := range corpus.Triples128 {
		fmt.Fprintf(b, "    (%d, %d, %d),\n", t[0], t[1], t[2])
	}
	b.WriteString("];\n\n")

	b.WriteString("const SCALEB_EXPS: &[i64] = &[")
	for i, e := range publicParityScaleBExps {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(b, "%d", e)
	}
	b.WriteString("];\n\n")

	writeIntSlice := func(name, rustType string, n int, at func(i int) string) {
		fmt.Fprintf(b, "const %s: &[%s] = &[", name, rustType)
		for i := 0; i < n; i++ {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(at(i))
		}
		b.WriteString("];\n\n")
	}
	writeIntSlice("INT_CORPUS_32", "i32", len(publicParityIntCorpus32), func(i int) string {
		return strconv.FormatInt(int64(publicParityIntCorpus32[i]), 10)
	})
	writeIntSlice("UINT_CORPUS_32", "u32", len(publicParityUintCorpus32), func(i int) string {
		return strconv.FormatUint(uint64(publicParityUintCorpus32[i]), 10)
	})
	writeIntSlice("INT_CORPUS_64", "i64", len(publicParityIntCorpus64), func(i int) string {
		return strconv.FormatInt(publicParityIntCorpus64[i], 10) + "i64"
	})
	writeIntSlice("UINT_CORPUS_64", "u64", len(publicParityUintCorpus64), func(i int) string {
		return strconv.FormatUint(publicParityUintCorpus64[i], 10) + "u64"
	})

	b.WriteString(`/// One string-shape corpus entry with its generation-time classification
/// against the documented public parsing contract (mirrors the Go leg's
/// parityStringCase / publicParityStringCorpusCases verbatim -- same inputs,
/// same kind/signaling/nan_min_width/cohort_min_width classification, since
/// the Rust wrapper contract mirrors Go's exactly). Each minimum-width field
/// is zero when no supported width can preserve that property.
struct StringCase {
    input: &'static str,
    kind: &'static str,
    signaling: bool,
    nan_min_width: u32,
    cohort_min_width: u32,
}

const STRING_CASES: &[StringCase] = &[
`)
	for _, sc := range publicParityStringCorpusCases {
		fmt.Fprintf(b, "    StringCase { input: %q, kind: %q, signaling: %v, nan_min_width: %d, cohort_min_width: %d },\n", sc.Input, sc.Kind, sc.Signaling, sc.NaNMinWidth, sc.CohortMinWidth)
	}
	b.WriteString("];\n\n")
}

// ---- shared diagnostic helpers ----

func ctxLabelAndPlaceholders(arity int, w parityWidth) (label, placeholders string) {
	label = "operand"
	if arity > 1 {
		label = "operands"
	}
	parts := make([]string, arity)
	for i := range parts {
		parts[i] = w.opFmtSpec()
	}
	return label, strings.Join(parts, ",")
}

func operandVarsForArity(arity int) []string {
	vars := make([]string, arity)
	for i := range vars {
		vars[i] = fmt.Sprintf("v%d", i)
	}
	return vars
}

// resolvePort looks up the crate::generated module/function apiemit resolved
// a bidgo function name to, failing closed with a generation-time error that
// names the missing table entry instead of silently skipping the row.
func resolvePort(bidgoFn, goSymbol string) (module, fn string, err error) {
	module, fn, ok := apiemit.PortPathFor(bidgoFn)
	if !ok {
		return "", "", fmt.Errorf("rust public parity: no apiemit port path for bidgo_function %q (go_symbol %q); extend devtools/tools/go2rs/apiemit's portPath table first (apiemit itself would fail to emit a wrapper calling an unresolved port function, so this indicates the verification and apiemit's table have drifted)", bidgoFn, goSymbol)
	}
	return module, fn, nil
}

// emitRustParityUnit dispatches one emitted census row to its shape template
// and returns the generated Rust function name plus its pinned case count.
// Every shape's width is resolved from the row's own BidgoFunction prefix
// (parityWidthForBidgoFunc), so a single dispatch table covers every emitted
// width without a per-width duplicate switch.
func emitRustParityUnit(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus) (string, int, error) {
	w, err := parityWidthForBidgoFunc(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	switch row.Shape {
	case "unary":
		return emitSimpleOp(b, row, corpus, w, 1, false, false, false, resultDec64)
	case "unary_with_flags_no_round":
		return emitSimpleOp(b, row, corpus, w, 1, false, true, false, resultDec64)
	case "unary_int_with_flags_no_round":
		return emitSimpleOp(b, row, corpus, w, 1, false, true, false, resultInt)
	case "unary_with_flags_default_round":
		return emitSimpleOp(b, row, corpus, w, 1, true, true, false, resultDec64)
	case "unary_mode_drop_flags":
		return emitSimpleOp(b, row, corpus, w, 1, true, false, true, resultDec64)
	case "predicate":
		return emitSimpleOp(b, row, corpus, w, 1, false, false, false, resultBool)
	case "binary":
		return emitSimpleOp(b, row, corpus, w, 2, true, false, false, resultDec64)
	case "binary_with_flags":
		return emitSimpleOp(b, row, corpus, w, 2, true, true, false, resultDec64)
	case "binary_mode_flags":
		return emitModeBinary(b, row, corpus, w)
	case "mixed_binary_mode_flags_dd", "mixed_binary_mode_flags_dq", "mixed_binary_mode_flags_qd", "mixed_binary_mode_flags_qq":
		return emitMixedModeBinary(b, row, corpus, w)
	case "mixed_ternary_mode_flags_ddd", "mixed_ternary_mode_flags_ddq", "mixed_ternary_mode_flags_dqd", "mixed_ternary_mode_flags_dqq",
		"mixed_ternary_mode_flags_qdd", "mixed_ternary_mode_flags_qdq", "mixed_ternary_mode_flags_qqd", "mixed_ternary_mode_flags_qqq":
		return emitMixedModeTernary(b, row, corpus, w)
	case "mixed_unary_mode_flags_d", "mixed_unary_mode_flags_q":
		return emitMixedModeUnary(b, row, corpus, w)
	case "unary_mode_flags":
		return emitModeUnaryArith(b, row, corpus, w)
	case "ternary_mode_flags":
		return emitModeTernary(b, row, corpus, w)
	case "scaleb_mode":
		return emitModeScaleB(b, row, corpus, w)
	case "binary_flags_no_round":
		return emitSimpleOp(b, row, corpus, w, 2, false, true, false, resultDec64)
	case "binary_drop_flags":
		return emitSimpleOp(b, row, corpus, w, 2, true, false, true, resultDec64)
	case "copysign":
		return emitSimpleOp(b, row, corpus, w, 2, false, false, false, resultDec64)
	case "same_quantum":
		return emitSimpleOp(b, row, corpus, w, 2, false, false, false, resultBool)
	case "compare_bool_flags":
		return emitSimpleOp(b, row, corpus, w, 2, false, true, false, resultBool)
	case "fma":
		return emitSimpleOp(b, row, corpus, w, 3, true, true, false, resultDec64)

	case "to_binary32":
		return emitModeUnary(b, row, corpus, w, modeResultF32, true)
	case "to_binary64":
		return emitModeUnary(b, row, corpus, w, modeResultF64, true)
	case "to_binary128":
		return emitModeUnary(b, row, corpus, w, modeResultBin128, true)
	case "to_decimal32":
		return emitModeUnary(b, row, corpus, w, modeResultDec32, true)
	case "to_decimal64":
		return emitModeUnary(b, row, corpus, w, modeResultDec64, false)
	case "to_decimal64_mode":
		// Decimal128->Decimal64 narrowing takes an explicit rounding mode
		// (inexact in general), unlike Decimal32's parameterless exact
		// to_decimal64 (modeResultDec64, hasMode=false): hasMode=true here.
		return emitModeUnary(b, row, corpus, w, modeResultDec64, true)
	case "to_decimal128":
		return emitModeUnary(b, row, corpus, w, modeResultDec128, false)

	case "scaleb":
		return emitScaleB(b, row, corpus, w)
	case "next_toward":
		return emitNextToward(b, row, corpus, w)
	case "class":
		return emitClass(b, row, corpus, w)
	case "sign":
		return emitSign(b, row, corpus, w)
	case "total_cmp", "total_cmp_mag":
		return emitTotalCmp(b, row, corpus, w)
	case "signaling_eq_compose":
		return emitSignalingEqCompose(b, row, corpus, w)
	case "signaling_not_eq_compose":
		return emitSignalingNotEqCompose(b, row, corpus, w)
	case "radix_const":
		return emitRadixConst(b, row, w)
	case "copy_fold":
		return emitCopyFold(b, row, corpus, w)
	case "display":
		return emitDisplay(b, row, corpus, w)

	case "parse", "parse_fold":
		return emitParse(b, row, w)
	case "parse_raw":
		return emitParseRaw(b, row, w)
	case "parse_with_flags":
		return emitParseWithFlags(b, row, w)
	case "parse_mode":
		return emitParseMode(b, row, w)
	case "from_i64_exact_or_error":
		return emitFromIntExactOrError(b, row, w, "INT_CORPUS_64")
	case "from_i32_exact_or_error":
		return emitFromIntExactOrError(b, row, w, "INT_CORPUS_32")
	case "from_i32_exact", "from_u32_exact", "from_i64_exact", "from_u64_exact":
		return emitFromExact(b, row, w)
	case "from_i64_mode", "from_u64_mode", "from_i32_mode", "from_u32_mode":
		return emitFromMode(b, row, w)

	case "context_binary_with_flags":
		return emitContextBinaryWithFlags(b, row, corpus, w)

	default:
		if strings.HasPrefix(row.Shape, "to_i") || strings.HasPrefix(row.Shape, "to_u") {
			return emitConvertInt(b, row, corpus, w)
		}
		return "", 0, fmt.Errorf("rust public parity: no emitter for shape %q (go_symbol %q)", row.Shape, row.GoSymbol)
	}
}

var rustMixedShapeWidths = map[string][2]parityWidth{
	"mixed_binary_mode_flags_dd": {decimal64ParityWidth, decimal64ParityWidth},
	"mixed_binary_mode_flags_dq": {decimal64ParityWidth, decimal128ParityWidth},
	"mixed_binary_mode_flags_qd": {decimal128ParityWidth, decimal64ParityWidth},
	"mixed_binary_mode_flags_qq": {decimal128ParityWidth, decimal128ParityWidth},
}

var rustMixedTernaryShapeWidths = map[string][3]parityWidth{
	"mixed_ternary_mode_flags_ddd": {decimal64ParityWidth, decimal64ParityWidth, decimal64ParityWidth},
	"mixed_ternary_mode_flags_ddq": {decimal64ParityWidth, decimal64ParityWidth, decimal128ParityWidth},
	"mixed_ternary_mode_flags_dqd": {decimal64ParityWidth, decimal128ParityWidth, decimal64ParityWidth},
	"mixed_ternary_mode_flags_dqq": {decimal64ParityWidth, decimal128ParityWidth, decimal128ParityWidth},
	"mixed_ternary_mode_flags_qdd": {decimal128ParityWidth, decimal64ParityWidth, decimal64ParityWidth},
	"mixed_ternary_mode_flags_qdq": {decimal128ParityWidth, decimal64ParityWidth, decimal128ParityWidth},
	"mixed_ternary_mode_flags_qqd": {decimal128ParityWidth, decimal128ParityWidth, decimal64ParityWidth},
	"mixed_ternary_mode_flags_qqq": {decimal128ParityWidth, decimal128ParityWidth, decimal128ParityWidth},
}

var rustMixedUnaryShapeWidths = map[string]parityWidth{
	"mixed_unary_mode_flags_d": decimal64ParityWidth,
	"mixed_unary_mode_flags_q": decimal128ParityWidth,
}

func rustMixedOperation(goSymbol string) (string, error) {
	for _, op := range []string{"Add", "Sub", "Mul", "Div"} {
		if strings.HasPrefix(goSymbol, op) {
			return op, nil
		}
	}
	return "", fmt.Errorf("rust public parity: mixed-width go_symbol %q has no supported Add/Sub/Mul/Div prefix", goSymbol)
}

// emitMixedModeBinary independently exercises one generated Rust associated
// function for Intel's D/Q mixed-width arithmetic. It uses the source-width
// pair table and value conversion for each operand, while result comparison is
// normalized by the destination width resolved from the census port name.
func emitMixedModeBinary(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, result parityWidth) (string, int, error) {
	operands, ok := rustMixedShapeWidths[row.Shape]
	if !ok {
		return "", 0, fmt.Errorf("rust public parity: mixed-width shape %q has no independent operand-width mapping", row.Shape)
	}
	left, right := operands[0], operands[1]
	op, err := rustMixedOperation(row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	disc, err := mixedModeBinaryDiscriminantOperands(op, parityWidthDigits(result), [2]int{parityWidthDigits(left), parityWidthDigits(right)})
	if err != nil {
		return "", 0, err
	}
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := (len(parityLabelPairs) + len(disc)) * len(publicParityModeOrderNames)
	pubBits, portBits := result.pubBitsExpr("pv"), result.portBitsExpr("pr")

	fmt.Fprintf(b, "fn %s(failures: &mut Vec<String>) -> usize {\n", funcName)
	b.WriteString("    let mut count = 0usize;\n")
	fmt.Fprintf(b, "    for pair_index in 0..%s.len() {\n", left.pairs)
	fmt.Fprintf(b, "        let left_pair = %s[pair_index];\n", left.pairs)
	fmt.Fprintf(b, "        let right_pair = %s[pair_index];\n", right.pairs)
	fmt.Fprintf(b, "        let left_bits = %s[left_pair.0];\n", left.corpus)
	fmt.Fprintf(b, "        let right_bits = %s[right_pair.1];\n", right.corpus)
	b.WriteString("        for &(mode, port_mode) in PARITY_MODES {\n")
	indent := "            "
	fmt.Fprintf(b, "%slet (pv, pf) = %s::%s(%s, %s, mode);\n", indent, result.selfType, row.RustSurface, left.pubFrom("left_bits"), right.pubFrom("right_bits"))
	b.WriteString(indent + portCallStmt(row.BidgoFunction, module, fn, []string{left.portArg("left_bits"), right.portArg("right_bits"), "port_mode"}) + "\n")
	ctxFmt := "operands " + left.opFmtSpec() + "," + right.opFmtSpec() + " mode {:?}"
	ctxArgs := "left_bits, right_bits, mode"
	fmt.Fprintf(b, "%sif %s != %s {\n", indent, pubBits, portBits)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: result mismatch public=%s port=%s\", %s, %s, %s));\n", indent, row.GoSymbol, ctxFmt, result.resultFmtSpec(), result.resultFmtSpec(), ctxArgs, pubBits, portBits)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%sif pf.bits() != map_port_flags(praw) {\n", indent)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: flag mismatch public={:#x} port={:#x}\", %s, pf.bits(), map_port_flags(praw)));\n", indent, row.GoSymbol, ctxFmt, ctxArgs)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%scount += 1;\n", indent)
	b.WriteString("        }\n    }\n")

	if len(disc) > 0 {
		leftType, _ := rustModeDiscTypes(left)
		rightType, _ := rustModeDiscTypes(right)
		_, seenInit := rustModeDiscTypes(result)
		fmt.Fprintf(b, "    let disc_pairs: &[(%s, %s)] = &[\n", leftType, rightType)
		for _, pair := range disc {
			leftLit, err := modeDiscRustLiteral(parityWidthDigits(left), pair[0])
			if err != nil {
				return "", 0, fmt.Errorf("%s left discriminant: %w", row.GoSymbol, err)
			}
			rightLit, err := modeDiscRustLiteral(parityWidthDigits(right), pair[1])
			if err != nil {
				return "", 0, fmt.Errorf("%s right discriminant: %w", row.GoSymbol, err)
			}
			fmt.Fprintf(b, "        (%s, %s),\n", leftLit, rightLit)
		}
		b.WriteString("    ];\n")
		b.WriteString("    for &(left_bits, right_bits) in disc_pairs {\n")
		fmt.Fprintf(b, "        let mut mode_seen = %s;\n", seenInit)
		b.WriteString("        for (mi, &(mode, port_mode)) in PARITY_MODES.iter().enumerate() {\n")
		fmt.Fprintf(b, "%slet (pv, pf) = %s::%s(%s, %s, mode);\n", indent, result.selfType, row.RustSurface, left.pubFrom("left_bits"), right.pubFrom("right_bits"))
		b.WriteString(indent + portCallStmt(row.BidgoFunction, module, fn, []string{left.portArg("left_bits"), right.portArg("right_bits"), "port_mode"}) + "\n")
		discCtxFmt := "discriminant operands " + left.opFmtSpec() + "," + right.opFmtSpec() + " mode {:?}"
		fmt.Fprintf(b, "%sif %s != %s {\n", indent, pubBits, portBits)
		fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: result mismatch public=%s port=%s\", left_bits, right_bits, mode, %s, %s));\n", indent, row.GoSymbol, discCtxFmt, result.resultFmtSpec(), result.resultFmtSpec(), pubBits, portBits)
		fmt.Fprintf(b, "%s}\n", indent)
		fmt.Fprintf(b, "%sif pf.bits() != map_port_flags(praw) {\n", indent)
		fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: flag mismatch public={:#x} port={:#x}\", left_bits, right_bits, mode, pf.bits(), map_port_flags(praw)));\n", indent, row.GoSymbol, discCtxFmt)
		fmt.Fprintf(b, "%s}\n", indent)
		fmt.Fprintf(b, "%smode_seen[mi] = %s;\n", indent, pubBits)
		fmt.Fprintf(b, "%scount += 1;\n", indent)
		b.WriteString("        }\n")
		fmt.Fprintf(b, "        if mode_seen.iter().all(|s| *s == mode_seen[0]) {\n")
		fmt.Fprintf(b, "            failures.push(format!(\"public parity %s: discriminant operands %s,%s: every rounding mode produced the same result\", left_bits, right_bits));\n", row.GoSymbol, left.opFmtSpec(), right.opFmtSpec())
		b.WriteString("        }\n    }\n")
	}
	b.WriteString("    count\n}\n\n")
	return funcName, cases, nil
}

// validateRustMixedTernaryRow keeps the Rust parity generator's D/Q mapping
// independent of apiemit's mapping. A manifest row must name the exact Intel
// mixed FMA symbol, associated function, owner, and port implied by its shape;
// otherwise generation fails instead of comparing two identically miswired
// paths.
func validateRustMixedTernaryRow(row rustParityInventoryRow, result parityWidth) (string, [3]parityWidth, error) {
	operands, ok := rustMixedTernaryShapeWidths[row.Shape]
	if !ok {
		return "", [3]parityWidth{}, fmt.Errorf("rust public parity: mixed ternary shape %q has no independent operand-width mapping", row.Shape)
	}
	code := strings.TrimPrefix(row.Shape, "mixed_ternary_mode_flags_")
	resultWidth := parityWidthDigits(result)
	if (resultWidth == 64 && code == "ddd") || (resultWidth == 128 && code == "qqq") {
		return "", [3]parityWidth{}, fmt.Errorf("rust public parity: %s is a same-width FMA shape, not an Intel mixed-width extension for Decimal%d", code, resultWidth)
	}
	wantSymbol := fmt.Sprintf("FMA%d%sBIDWithMode", resultWidth, strings.ToUpper(code))
	wantSurface := "fma_" + code + "_with_mode"
	wantPort := fmt.Sprintf("Bid%d%sFma", resultWidth, code)
	if row.GoSymbol != wantSymbol || row.RustOwner != result.selfType || row.RustSurface != wantSurface || row.BidgoFunction != wantPort {
		return "", [3]parityWidth{}, fmt.Errorf("rust public parity: mixed FMA row %q does not match shape %q; want symbol=%q owner=%q surface=%q port=%q, got owner=%q surface=%q port=%q", row.GoSymbol, row.Shape, wantSymbol, result.selfType, wantSurface, wantPort, row.RustOwner, row.RustSurface, row.BidgoFunction)
	}
	return code, operands, nil
}

// emitMixedModeTernary independently exercises one generated Rust associated
// function for Intel's mixed-width FMA family. Each operand is selected and
// converted through its declared D/Q width, preserving operand order (which
// is observable for NaN selection), while the result and flags are compared
// against the direct generated Go-port entrypoint.
func emitMixedModeTernary(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, result parityWidth) (string, int, error) {
	code, operands, err := validateRustMixedTernaryRow(row, result)
	if err != nil {
		return "", 0, err
	}
	probeFunction := fmt.Sprintf("bid%d%s_fma", parityWidthDigits(result), code)
	shape, ok := ffiMixedDecimalShapeFor(probeFunction)
	if !ok {
		return "", 0, fmt.Errorf("rust public parity: mixed FMA %q has no FFI shape for its fusedness sentinel", row.GoSymbol)
	}
	probe, ok := ffiMixedFMAFusednessProbeFor(probeFunction)
	if !ok {
		return "", 0, fmt.Errorf("rust public parity: mixed FMA %q has no fusedness sentinel for %q", row.GoSymbol, probeFunction)
	}
	if err := validateFFIFusednessProbe(probe, shape); err != nil {
		return "", 0, fmt.Errorf("rust public parity: mixed FMA %q fusedness sentinel: %w", row.GoSymbol, err)
	}
	if probe.rounding != 0 {
		return "", 0, fmt.Errorf("rust public parity: mixed FMA %q fusedness sentinel rounding = %d, want nearest-even (0)", row.GoSymbol, probe.rounding)
	}
	xWidth, yWidth, zWidth := operands[0], operands[1], operands[2]
	operandDigits := [3]int{parityWidthDigits(xWidth), parityWidthDigits(yWidth), parityWidthDigits(zWidth)}
	disc, err := mixedModeTernaryDiscriminantOperands("FMA", parityWidthDigits(result), operandDigits)
	if err != nil {
		return "", 0, err
	}
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := (len(parityLabelTriples)+len(disc))*len(publicParityModeOrderNames) + 1
	pubBits, portBits := result.pubBitsExpr("pv"), result.portBitsExpr("pr")
	resultFmt := result.resultFmtSpec()

	fmt.Fprintf(b, "fn %s(failures: &mut Vec<String>) -> usize {\n", funcName)
	b.WriteString("    let mut count = 0usize;\n")
	fmt.Fprintf(b, "    for triple_index in 0..%s.len() {\n", xWidth.triples)
	fmt.Fprintf(b, "        let x_triple = %s[triple_index];\n", xWidth.triples)
	fmt.Fprintf(b, "        let y_triple = %s[triple_index];\n", yWidth.triples)
	fmt.Fprintf(b, "        let z_triple = %s[triple_index];\n", zWidth.triples)
	fmt.Fprintf(b, "        let x_bits = %s[x_triple.0];\n", xWidth.corpus)
	fmt.Fprintf(b, "        let y_bits = %s[y_triple.1];\n", yWidth.corpus)
	fmt.Fprintf(b, "        let z_bits = %s[z_triple.2];\n", zWidth.corpus)
	b.WriteString("        for &(mode, port_mode) in PARITY_MODES {\n")
	indent := "            "
	fmt.Fprintf(b, "%slet (pv, pf) = %s::%s(%s, %s, %s, mode);\n", indent, result.selfType, row.RustSurface, xWidth.pubFrom("x_bits"), yWidth.pubFrom("y_bits"), zWidth.pubFrom("z_bits"))
	b.WriteString(indent + portCallStmt(row.BidgoFunction, module, fn, []string{xWidth.portArg("x_bits"), yWidth.portArg("y_bits"), zWidth.portArg("z_bits"), "port_mode"}) + "\n")
	ctxFmt := "operands " + xWidth.opFmtSpec() + "," + yWidth.opFmtSpec() + "," + zWidth.opFmtSpec() + " mode {:?}"
	fmt.Fprintf(b, "%sif %s != %s {\n", indent, pubBits, portBits)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: result mismatch public=%s port=%s\", x_bits, y_bits, z_bits, mode, %s, %s));\n", indent, row.GoSymbol, ctxFmt, resultFmt, resultFmt, pubBits, portBits)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%sif pf.bits() != map_port_flags(praw) {\n", indent)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: flag mismatch public={:#x} port={:#x}\", x_bits, y_bits, z_bits, mode, pf.bits(), map_port_flags(praw)));\n", indent, row.GoSymbol, ctxFmt)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%scount += 1;\n", indent)
	b.WriteString("        }\n    }\n")

	if len(disc) > 0 {
		xType, _ := rustModeDiscTypes(xWidth)
		yType, _ := rustModeDiscTypes(yWidth)
		zType, _ := rustModeDiscTypes(zWidth)
		_, seenInit := rustModeDiscTypes(result)
		fmt.Fprintf(b, "    let disc_triples: &[(%s, %s, %s)] = &[\n", xType, yType, zType)
		for _, triple := range disc {
			xLit, err := modeDiscRustLiteral(parityWidthDigits(xWidth), triple[0])
			if err != nil {
				return "", 0, fmt.Errorf("%s x discriminant: %w", row.GoSymbol, err)
			}
			yLit, err := modeDiscRustLiteral(parityWidthDigits(yWidth), triple[1])
			if err != nil {
				return "", 0, fmt.Errorf("%s y discriminant: %w", row.GoSymbol, err)
			}
			zLit, err := modeDiscRustLiteral(parityWidthDigits(zWidth), triple[2])
			if err != nil {
				return "", 0, fmt.Errorf("%s z discriminant: %w", row.GoSymbol, err)
			}
			fmt.Fprintf(b, "        (%s, %s, %s),\n", xLit, yLit, zLit)
		}
		b.WriteString("    ];\n")
		b.WriteString("    for &(x_bits, y_bits, z_bits) in disc_triples {\n")
		fmt.Fprintf(b, "        let mut mode_seen = %s;\n", seenInit)
		b.WriteString("        for (mi, &(mode, port_mode)) in PARITY_MODES.iter().enumerate() {\n")
		fmt.Fprintf(b, "%slet (pv, pf) = %s::%s(%s, %s, %s, mode);\n", indent, result.selfType, row.RustSurface, xWidth.pubFrom("x_bits"), yWidth.pubFrom("y_bits"), zWidth.pubFrom("z_bits"))
		b.WriteString(indent + portCallStmt(row.BidgoFunction, module, fn, []string{xWidth.portArg("x_bits"), yWidth.portArg("y_bits"), zWidth.portArg("z_bits"), "port_mode"}) + "\n")
		discCtxFmt := "discriminant operands " + xWidth.opFmtSpec() + "," + yWidth.opFmtSpec() + "," + zWidth.opFmtSpec() + " mode {:?}"
		fmt.Fprintf(b, "%sif %s != %s {\n", indent, pubBits, portBits)
		fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: result mismatch public=%s port=%s\", x_bits, y_bits, z_bits, mode, %s, %s));\n", indent, row.GoSymbol, discCtxFmt, resultFmt, resultFmt, pubBits, portBits)
		fmt.Fprintf(b, "%s}\n", indent)
		fmt.Fprintf(b, "%sif pf.bits() != map_port_flags(praw) {\n", indent)
		fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: flag mismatch public={:#x} port={:#x}\", x_bits, y_bits, z_bits, mode, pf.bits(), map_port_flags(praw)));\n", indent, row.GoSymbol, discCtxFmt)
		fmt.Fprintf(b, "%s}\n", indent)
		fmt.Fprintf(b, "%smode_seen[mi] = %s;\n", indent, pubBits)
		fmt.Fprintf(b, "%scount += 1;\n", indent)
		b.WriteString("        }\n")
		fmt.Fprintf(b, "        if mode_seen.iter().all(|s| *s == mode_seen[0]) {\n")
		fmt.Fprintf(b, "            failures.push(format!(\"public parity %s: discriminant operands %s,%s,%s: every rounding mode produced the same result\", x_bits, y_bits, z_bits));\n", row.GoSymbol, xWidth.opFmtSpec(), yWidth.opFmtSpec(), zWidth.opFmtSpec())
		b.WriteString("        }\n    }\n")
	}
	if err := emitRustMixedFMAFusednessCase(b, row, result, operands, probe, module, fn); err != nil {
		return "", 0, err
	}
	b.WriteString("    count\n}\n\n")
	return funcName, cases, nil
}

func rustFusednessBitsType(bits ffiFusednessBits) (string, error) {
	switch bits.width {
	case 64:
		return "u64", nil
	case 128:
		return "[u8; 16]", nil
	default:
		return "", fmt.Errorf("rust public parity: fusedness literal has unsupported width %d", bits.width)
	}
}

// rustFusednessBitsLiteral renders the table's semantic (hi, lo) words as the
// raw Rust value representation. Decimal128 is deliberately emitted lo-word
// first in little-endian byte order, exactly matching Decimal128::from_le_bytes
// and the FFI shard representation.
func rustFusednessBitsLiteral(bits ffiFusednessBits) (string, error) {
	switch bits.width {
	case 64:
		return fmt.Sprintf("0x%016xu64", bits.lo), nil
	case 128:
		parts := make([]string, 16)
		for i := 0; i < 8; i++ {
			parts[i] = fmt.Sprintf("0x%02x", byte(bits.lo>>uint(8*i)))
			parts[8+i] = fmt.Sprintf("0x%02x", byte(bits.hi>>uint(8*i)))
		}
		return "[" + strings.Join(parts, ", ") + "]", nil
	default:
		return "", fmt.Errorf("rust public parity: fusedness literal has unsupported width %d", bits.width)
	}
}

func emitRustFusednessOperand(b *strings.Builder, name string, bits ffiFusednessBits) error {
	typeName, err := rustFusednessBitsType(bits)
	if err != nil {
		return err
	}
	literal, err := rustFusednessBitsLiteral(bits)
	if err != nil {
		return err
	}
	fmt.Fprintf(b, "        let %s_bits: %s = %s;\n", name, typeName, literal)
	return nil
}

func emitRustFusednessWiden(b *strings.Builder, name string, bits ffiFusednessBits) error {
	switch bits.width {
	case 64:
		fmt.Fprintf(b, "        let (%s_q, %s_widen_raw) = bid754::generated::to_bid12864::bid64_to_bid128(%s_bits);\n", name, name, name)
	case 128:
		fmt.Fprintf(b, "        let %s_q = to_port128(%s_bits);\n", name, name)
		fmt.Fprintf(b, "        let %s_widen_raw = 0u32;\n", name)
	default:
		return fmt.Errorf("rust public parity: fusedness operand %s has unsupported width %d", name, bits.width)
	}
	return nil
}

// emitRustMixedFMAFusednessCase adds one nearest-even known-answer case to a
// mixed FMA wrapper. It closes both sides of the contract: the public wrapper
// and its direct generated port must equal the fused Intel result, while an
// explicitly sequential generated-port composition must equal the pinned
// forbidden result. Every conversion/operation status is ORed so flags remain
// sticky across the full composition.
func emitRustMixedFMAFusednessCase(b *strings.Builder, row rustParityInventoryRow, result parityWidth, operands [3]parityWidth, probe ffiFusednessProbe, module, fn string) error {
	if apiemit.PortPfpsf(row.BidgoFunction) {
		return fmt.Errorf("rust public parity: mixed FMA %q unexpectedly uses a pfpsf output parameter", row.GoSymbol)
	}
	for i, width := range operands {
		if got, want := probe.operands[i].width, parityWidthDigits(width); got != want {
			return fmt.Errorf("rust public parity: mixed FMA %q fusedness operand %d width = %d, want %d", row.GoSymbol, i, got, want)
		}
	}
	if got, want := probe.expected.bits.width, parityWidthDigits(result); got != want {
		return fmt.Errorf("rust public parity: mixed FMA %q fusedness result width = %d, want %d", row.GoSymbol, got, want)
	}

	expectedType, err := rustFusednessBitsType(probe.expected.bits)
	if err != nil {
		return err
	}
	expectedLiteral, err := rustFusednessBitsLiteral(probe.expected.bits)
	if err != nil {
		return err
	}
	forbiddenLiteral, err := rustFusednessBitsLiteral(probe.forbidden.bits)
	if err != nil {
		return err
	}

	b.WriteString("    {\n")
	for i, name := range []string{"fused_x", "fused_y", "fused_z"} {
		if err := emitRustFusednessOperand(b, name, probe.operands[i]); err != nil {
			return fmt.Errorf("rust public parity: mixed FMA %q: %w", row.GoSymbol, err)
		}
	}
	fmt.Fprintf(b, "        let expected_bits: %s = %s;\n", expectedType, expectedLiteral)
	fmt.Fprintf(b, "        let forbidden_bits: %s = %s;\n", expectedType, forbiddenLiteral)
	fmt.Fprintf(b, "        let expected_raw = 0x%08xu32;\n", probe.expected.flags)
	fmt.Fprintf(b, "        let forbidden_raw = 0x%08xu32;\n", probe.forbidden.flags)
	b.WriteString("        let fused_mode = RoundingMode::NearestEven;\n")
	b.WriteString("        let fused_port_mode = BIDGO_ROUND_NEAREST_EVEN;\n")
	fmt.Fprintf(b, "        let (fused_pv, fused_pf) = %s::%s(%s, %s, %s, fused_mode);\n",
		result.selfType, row.RustSurface,
		operands[0].pubFrom("fused_x_bits"), operands[1].pubFrom("fused_y_bits"), operands[2].pubFrom("fused_z_bits"))
	fmt.Fprintf(b, "        let (fused_pr, fused_praw) = bid754::generated::%s::%s(%s, %s, %s, fused_port_mode);\n",
		module, fn,
		operands[0].portArg("fused_x_bits"), operands[1].portArg("fused_y_bits"), operands[2].portArg("fused_z_bits"))

	pubBits := result.pubBitsExpr("fused_pv")
	portBits := result.portBitsExpr("fused_pr")
	resultFmt := result.resultFmtSpec()
	fmt.Fprintf(b, "        if %s != expected_bits {\n", pubBits)
	fmt.Fprintf(b, "            failures.push(format!(\"public parity %s fusedness: public result mismatch public=%s expected=%s\", %s, expected_bits));\n", row.GoSymbol, resultFmt, resultFmt, pubBits)
	b.WriteString("        }\n")
	b.WriteString("        if fused_pf.bits() != map_port_flags(expected_raw) {\n")
	fmt.Fprintf(b, "            failures.push(format!(\"public parity %s fusedness: public flags mismatch public={:#x} expected={:#x}\", fused_pf.bits(), map_port_flags(expected_raw)));\n", row.GoSymbol)
	b.WriteString("        }\n")
	fmt.Fprintf(b, "        if %s != expected_bits {\n", portBits)
	fmt.Fprintf(b, "            failures.push(format!(\"public parity %s fusedness: direct-port result mismatch port=%s expected=%s\", %s, expected_bits));\n", row.GoSymbol, resultFmt, resultFmt, portBits)
	b.WriteString("        }\n")
	b.WriteString("        if fused_praw != expected_raw {\n")
	fmt.Fprintf(b, "            failures.push(format!(\"public parity %s fusedness: direct-port raw flags mismatch port={:#x} expected={:#x}\", fused_praw, expected_raw));\n", row.GoSymbol)
	b.WriteString("        }\n")

	for i, name := range []string{"fused_x", "fused_y", "fused_z"} {
		if err := emitRustFusednessWiden(b, name, probe.operands[i]); err != nil {
			return fmt.Errorf("rust public parity: mixed FMA %q: %w", row.GoSymbol, err)
		}
	}
	b.WriteString("        let (fused_product, fused_mul_raw) = bid754::generated::bid128_mul::bid128_mul(fused_x_q, fused_y_q, fused_port_mode);\n")
	b.WriteString("        let mut composed_raw = fused_x_widen_raw | fused_y_widen_raw | fused_z_widen_raw | fused_mul_raw;\n")
	b.WriteString("        let fused_sum = bid754::generated::bid128_add::bid128_add(fused_product, fused_z_q, fused_port_mode, &mut composed_raw);\n")
	if result.selfType == "Decimal64" {
		b.WriteString("        let (composed_result, fused_narrow_raw) = bid754::generated::bid128_conversions::bid128_to_bid64(fused_sum, fused_port_mode);\n")
		b.WriteString("        composed_raw |= fused_narrow_raw;\n")
	} else {
		b.WriteString("        let composed_result = fused_sum;\n")
	}
	composedBits := result.portBitsExpr("composed_result")
	fmt.Fprintf(b, "        if %s != forbidden_bits {\n", composedBits)
	fmt.Fprintf(b, "            failures.push(format!(\"public parity %s fusedness: sequential result mismatch composed=%s forbidden=%s\", %s, forbidden_bits));\n", row.GoSymbol, resultFmt, resultFmt, composedBits)
	b.WriteString("        }\n")
	b.WriteString("        if composed_raw != forbidden_raw {\n")
	fmt.Fprintf(b, "            failures.push(format!(\"public parity %s fusedness: sequential raw flags mismatch composed={:#x} forbidden={:#x}\", composed_raw, forbidden_raw));\n", row.GoSymbol)
	b.WriteString("        }\n")
	fmt.Fprintf(b, "        if %s == %s && composed_raw == fused_praw {\n", composedBits, portBits)
	fmt.Fprintf(b, "            failures.push(\"public parity %s fusedness: sequential composition did not differ from direct FMA in bits+raw-flags\".to_string());\n", row.GoSymbol)
	b.WriteString("        }\n")
	b.WriteString("        count += 1;\n")
	b.WriteString("    }\n")
	return nil
}

func validateRustMixedUnaryRow(row rustParityInventoryRow, result parityWidth) (parityWidth, error) {
	operand, ok := rustMixedUnaryShapeWidths[row.Shape]
	if !ok {
		return parityWidth{}, fmt.Errorf("rust public parity: mixed unary shape %q has no independent operand-width mapping", row.Shape)
	}
	code := strings.TrimPrefix(row.Shape, "mixed_unary_mode_flags_")
	resultWidth := parityWidthDigits(result)
	if (resultWidth == 64 && code != "q") || (resultWidth == 128 && code != "d") {
		return parityWidth{}, fmt.Errorf("rust public parity: mixed sqrt shape %q is unsupported for Decimal%d", row.Shape, resultWidth)
	}
	wantSymbol := fmt.Sprintf("Sqrt%d%sBIDWithMode", resultWidth, strings.ToUpper(code))
	wantSurface := "sqrt_" + code + "_with_mode"
	wantPort := fmt.Sprintf("Bid%d%sSqrt", resultWidth, code)
	if row.GoSymbol != wantSymbol || row.RustOwner != result.selfType || row.RustSurface != wantSurface || row.BidgoFunction != wantPort {
		return parityWidth{}, fmt.Errorf("rust public parity: mixed sqrt row %q does not match shape %q; want symbol=%q owner=%q surface=%q port=%q, got owner=%q surface=%q port=%q", row.GoSymbol, row.Shape, wantSymbol, result.selfType, wantSurface, wantPort, row.RustOwner, row.RustSurface, row.BidgoFunction)
	}
	return operand, nil
}

// emitMixedModeUnary independently exercises Intel's two unlike-width square
// roots. The source corpus is decoded at the operand width and the result is
// compared at the destination width; no widen/sqrt/narrow composition enters
// this verification path.
func emitMixedModeUnary(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, result parityWidth) (string, int, error) {
	operand, err := validateRustMixedUnaryRow(row, result)
	if err != nil {
		return "", 0, err
	}
	disc, err := mixedModeUnaryDiscriminantOperands("Sqrt", parityWidthDigits(result), parityWidthDigits(operand))
	if err != nil {
		return "", 0, err
	}
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := (operand.bitsLen(corpus) + len(disc)) * len(publicParityModeOrderNames)
	pubBits, portBits := result.pubBitsExpr("pv"), result.portBitsExpr("pr")
	resultFmt := result.resultFmtSpec()
	indent := "            "

	fmt.Fprintf(b, "fn %s(failures: &mut Vec<String>) -> usize {\n", funcName)
	b.WriteString("    let mut count = 0usize;\n")
	fmt.Fprintf(b, "    for &value_bits in %s {\n", operand.corpus)
	b.WriteString("        for &(mode, port_mode) in PARITY_MODES {\n")
	fmt.Fprintf(b, "%slet (pv, pf) = %s::%s(%s, mode);\n", indent, result.selfType, row.RustSurface, operand.pubFrom("value_bits"))
	b.WriteString(indent + portCallStmt(row.BidgoFunction, module, fn, []string{operand.portArg("value_bits"), "port_mode"}) + "\n")
	ctxFmt := "operand " + operand.opFmtSpec() + " mode {:?}"
	fmt.Fprintf(b, "%sif %s != %s {\n", indent, pubBits, portBits)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: result mismatch public=%s port=%s\", value_bits, mode, %s, %s));\n", indent, row.GoSymbol, ctxFmt, resultFmt, resultFmt, pubBits, portBits)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%sif pf.bits() != map_port_flags(praw) {\n", indent)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: flag mismatch public={:#x} port={:#x}\", value_bits, mode, pf.bits(), map_port_flags(praw)));\n", indent, row.GoSymbol, ctxFmt)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%scount += 1;\n", indent)
	b.WriteString("        }\n    }\n")

	if len(disc) > 0 {
		elemType, _ := rustModeDiscTypes(operand)
		_, seenInit := rustModeDiscTypes(result)
		fmt.Fprintf(b, "    let disc_values: &[%s] = &[\n", elemType)
		for _, value := range disc {
			lit, err := modeDiscRustLiteral(parityWidthDigits(operand), value)
			if err != nil {
				return "", 0, fmt.Errorf("%s discriminant: %w", row.GoSymbol, err)
			}
			fmt.Fprintf(b, "        %s,\n", lit)
		}
		b.WriteString("    ];\n")
		b.WriteString("    for &value_bits in disc_values {\n")
		fmt.Fprintf(b, "        let mut mode_seen = %s;\n", seenInit)
		b.WriteString("        for (mi, &(mode, port_mode)) in PARITY_MODES.iter().enumerate() {\n")
		fmt.Fprintf(b, "%slet (pv, pf) = %s::%s(%s, mode);\n", indent, result.selfType, row.RustSurface, operand.pubFrom("value_bits"))
		b.WriteString(indent + portCallStmt(row.BidgoFunction, module, fn, []string{operand.portArg("value_bits"), "port_mode"}) + "\n")
		discCtxFmt := "discriminant operand " + operand.opFmtSpec() + " mode {:?}"
		fmt.Fprintf(b, "%sif %s != %s {\n", indent, pubBits, portBits)
		fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: result mismatch public=%s port=%s\", value_bits, mode, %s, %s));\n", indent, row.GoSymbol, discCtxFmt, resultFmt, resultFmt, pubBits, portBits)
		fmt.Fprintf(b, "%s}\n", indent)
		fmt.Fprintf(b, "%sif pf.bits() != map_port_flags(praw) {\n", indent)
		fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: flag mismatch public={:#x} port={:#x}\", value_bits, mode, pf.bits(), map_port_flags(praw)));\n", indent, row.GoSymbol, discCtxFmt)
		fmt.Fprintf(b, "%s}\n", indent)
		fmt.Fprintf(b, "%smode_seen[mi] = %s;\n", indent, pubBits)
		fmt.Fprintf(b, "%scount += 1;\n", indent)
		b.WriteString("        }\n")
		fmt.Fprintf(b, "        if mode_seen.iter().all(|s| *s == mode_seen[0]) {\n")
		fmt.Fprintf(b, "            failures.push(format!(\"public parity %s: discriminant operand %s: every rounding mode produced the same result\", value_bits));\n", row.GoSymbol, operand.opFmtSpec())
		b.WriteString("        }\n    }\n")
	}
	b.WriteString("    count\n}\n\n")
	return funcName, cases, nil
}

type simpleOpResultKind int

const (
	resultDec64 simpleOpResultKind = iota
	resultBool
	// resultInt is the integer logBFormat primary result (ILogB): both the
	// public wrapper and the port return i64, so the two sides are compared
	// directly with no bit-normalization and no "!= 0" truth conversion.
	resultInt
)

// emitSimpleOp handles every arity-1/2/3 same-width-operand shape with an
// optional BIDGO_ROUND_NEAREST_EVEN rounding argument on the port side and a
// none/returned/dropped flags shape: unary, unary_with_flags_no_round,
// unary_int_with_flags_no_round, unary_with_flags_default_round,
// unary_mode_drop_flags, predicate (arity
// 1); binary, binary_with_flags, binary_flags_no_round, binary_drop_flags,
// copysign, same_quantum, compare_bool_flags (arity 2); fma (arity 3). This
// single generic template, parameterized by width w, covers the large
// majority of emitted rows across every width because apiemit's shape
// taxonomy already pins the exact calling convention per tag (verified
// against the concrete generated bid754-rs/src/generated/api/decimal64.rs
// and decimal32.rs for every case). resultBool additionally needs
// apiemit.PortBoolResult: Decimal32's IsInf/IsNaN/IsZero/SameQuantum port
// functions already return Rust bool (unlike every other width's
// Intel-mirroring numeric "!= 0" convention), so the comparison must not
// blindly assume "!= 0" compiles.
func emitSimpleOp(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, w parityWidth, arity int, hasRounding, hasFlags, dropFlags bool, resultKind simpleOpResultKind) (string, int, error) {
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	vars := operandVarsForArity(arity)

	var loopOpen, loopClose string
	var cases int
	switch arity {
	case 1:
		loopOpen = rustTmpl("    for &v0 in @CORPUS@ {\n", "@CORPUS@", w.corpus)
		cases = w.bitsLen(corpus)
	case 2:
		loopOpen = rustTmpl("    for &(i0, i1) in @PAIRS@ {\n        let v0 = @CORPUS@[i0];\n        let v1 = @CORPUS@[i1];\n", "@PAIRS@", w.pairs, "@CORPUS@", w.corpus)
		cases = w.pairsLen(corpus)
	case 3:
		loopOpen = rustTmpl("    for &(i0, i1, i2) in @TRIPLES@ {\n        let v0 = @CORPUS@[i0];\n        let v1 = @CORPUS@[i1];\n        let v2 = @CORPUS@[i2];\n", "@TRIPLES@", w.triples, "@CORPUS@", w.corpus)
		cases = w.triplesLen(corpus)
	default:
		return "", 0, fmt.Errorf("rust public parity: unsupported arity %d for go_symbol %q", arity, row.GoSymbol)
	}
	loopClose = "    }\n"
	indent := "        "

	fmt.Fprintf(b, "fn %s(failures: &mut Vec<String>) -> usize {\n", funcName)
	fmt.Fprintf(b, "    let mut count = 0usize;\n")
	b.WriteString(loopOpen)

	pubArgs := make([]string, 0, arity-1)
	for _, v := range vars[1:] {
		pubArgs = append(pubArgs, w.pubFrom(v))
	}
	pubCall := fmt.Sprintf("%s.%s(%s)", w.pubFrom(vars[0]), row.RustSurface, strings.Join(pubArgs, ", "))
	if hasFlags {
		fmt.Fprintf(b, "%slet (pv, pf) = %s;\n", indent, pubCall)
	} else {
		fmt.Fprintf(b, "%slet pv = %s;\n", indent, pubCall)
	}

	portArgs := make([]string, len(vars))
	for i, v := range vars {
		portArgs[i] = w.portArg(v)
	}
	if hasRounding {
		portArgs = append(portArgs, "BIDGO_ROUND_NEAREST_EVEN")
	}
	// "binary" is the one shape whose port call genuinely forks by width
	// (mirrors apiemit's identical bins-loop fork, decimal_emit.go): every
	// 32/64 bidgo function this shape uses is a truly separate, flags-less
	// function, but Decimal128's census maps e.g. Add and AddWithFlags to
	// the SAME bidgo_function, whose generated:: signature always carries a
	// flags word (pfpsf or tuple) even though this shape's own wrapper
	// method signature drops it -- the independent port call here must
	// still satisfy that signature and then discard the flags.
	alwaysHasPortFlags := dropFlags || (w.selfType == "Decimal128" && row.Shape == "binary")
	if hasFlags || alwaysHasPortFlags {
		b.WriteString(indent + portCallStmt(row.BidgoFunction, module, fn, portArgs) + "\n")
		if !hasFlags {
			fmt.Fprintf(b, "%slet _ = praw;\n", indent)
		}
	} else {
		// Genuinely flags-less port function for every width (unary,
		// predicate, copysign, same_quantum): a bare non-tuple call, exactly
		// as before this widening.
		portCall := fmt.Sprintf("bid754::generated::%s::%s(%s)", module, fn, strings.Join(portArgs, ", "))
		fmt.Fprintf(b, "%slet pr = %s;\n", indent, portCall)
	}

	label, placeholders := ctxLabelAndPlaceholders(arity, w)
	ctxArgs := strings.Join(vars, ", ")

	switch resultKind {
	case resultBool:
		prBool := "(pr != 0)"
		if apiemit.PortBoolResult(row.BidgoFunction) {
			prBool = "pr"
		}
		fmt.Fprintf(b, "%sif pv != %s {\n", indent, prBool)
		fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s %s: result mismatch public={} port={}\", %s, pv, %s));\n", indent, row.GoSymbol, label, placeholders, ctxArgs, prBool)
		fmt.Fprintf(b, "%s}\n", indent)
	case resultInt:
		fmt.Fprintf(b, "%sif pv != pr {\n", indent)
		fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s %s: result mismatch public={} port={}\", %s, pv, pr));\n", indent, row.GoSymbol, label, placeholders, ctxArgs)
		fmt.Fprintf(b, "%s}\n", indent)
	default:
		pubBits, portBits, fmtSpec := w.pubBitsExpr("pv"), w.portBitsExpr("pr"), w.resultFmtSpec()
		fmt.Fprintf(b, "%sif %s != %s {\n", indent, pubBits, portBits)
		fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s %s: result mismatch public=%s port=%s\", %s, %s, %s));\n", indent, row.GoSymbol, label, placeholders, fmtSpec, fmtSpec, ctxArgs, pubBits, portBits)
		fmt.Fprintf(b, "%s}\n", indent)
	}
	if hasFlags {
		fmt.Fprintf(b, "%sif pf.bits() != map_port_flags(praw) {\n", indent)
		fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s %s: flag mismatch public={:#x} port={:#x}\", %s, pf.bits(), map_port_flags(praw)));\n", indent, row.GoSymbol, label, placeholders, ctxArgs)
		fmt.Fprintf(b, "%s}\n", indent)
	}
	fmt.Fprintf(b, "%scount += 1;\n", indent)
	b.WriteString(loopClose)
	b.WriteString("    count\n}\n\n")
	return funcName, cases, nil
}

type modeResultKind int

const (
	modeResultF32 modeResultKind = iota
	modeResultF64
	modeResultBin128
	modeResultDec32
	modeResultDec64
	modeResultDec128
)

// emitModeUnary handles the width-w-to-{f32,f64,Binary128,Decimal32,
// Decimal64,Decimal128} conversion family: to_binary32/64/128 (mode-taking,
// every width) and to_decimal32 (mode-taking narrowing, Decimal64 only) loop
// PARITY_MODES; to_decimal64/to_decimal128 (Decimal32/Decimal64's exact
// widenings) have no mode parameter at all (the port entrypoints
// bid32_to_bid64/bid<w>_to_bid128 take none), so hasMode=false skips the
// mode loop and calls the port with no rounding arg.
func emitModeUnary(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, w parityWidth, kind modeResultKind, hasMode bool) (string, int, error) {
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := w.bitsLen(corpus)
	if hasMode {
		cases *= len(publicParityModeOrderNames)
	}

	fmt.Fprintf(b, "fn %s(failures: &mut Vec<String>) -> usize {\n", funcName)
	fmt.Fprintf(b, "    let mut count = 0usize;\n")
	fmt.Fprintf(b, "    for &v0 in %s {\n", w.corpus)
	indent := "        "
	modeLoopIndent := indent
	if hasMode {
		fmt.Fprintf(b, "%sfor &(mode, port_mode) in PARITY_MODES {\n", indent)
		modeLoopIndent = indent + "    "
	}

	pubCall := fmt.Sprintf("%s.%s()", w.pubFrom("v0"), row.RustSurface)
	if hasMode {
		pubCall = fmt.Sprintf("%s.%s(mode)", w.pubFrom("v0"), row.RustSurface)
	}
	fmt.Fprintf(b, "%slet (pv, pf) = %s;\n", modeLoopIndent, pubCall)

	portArgs := []string{w.portArg("v0")}
	if hasMode {
		portArgs = append(portArgs, "port_mode")
	}
	b.WriteString(modeLoopIndent + portCallStmt(row.BidgoFunction, module, fn, portArgs) + "\n")

	ctxFmt := "operand " + w.opFmtSpec()
	ctxArgs := "v0"
	if hasMode {
		ctxFmt = "operand " + w.opFmtSpec() + " mode {:?}"
		ctxArgs = "v0, mode"
	}

	var pubBits, portBits string
	switch kind {
	case modeResultF32, modeResultF64:
		// The public wrapper returns f32/f64 for every width, so pv.to_bits()
		// is the bit pattern. The PORT result forks by width: Bid64/Bid32's
		// to_binary32/64 return that same bit pattern as a plain u32/u64
		// (compare directly), but Bid128's return a native f32/f64 (a pfpsf
		// function -- see emitToBinary32Op's doc comment), so its bits need
		// pr.to_bits() to compare (and to stay NaN-safe: a bit comparison,
		// not a float ==).
		if w.selfType == "Decimal128" {
			pubBits, portBits = "pv.to_bits()", "pr.to_bits()"
		} else {
			pubBits, portBits = "pv.to_bits()", "pr"
		}
	case modeResultDec32, modeResultDec64:
		pubBits, portBits = "pv.to_bits()", "pr"
	case modeResultBin128, modeResultDec128:
		pubBits, portBits = "pv.to_le_bytes()", "from_port128(pr)"
	}
	fmtSpec := "{:#x}"
	if kind == modeResultBin128 || kind == modeResultDec128 {
		fmtSpec = "{:?}"
	}
	fmt.Fprintf(b, "%sif %s != %s {\n", modeLoopIndent, pubBits, portBits)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: result mismatch public=%s port=%s\", %s, %s, %s));\n", modeLoopIndent, row.GoSymbol, ctxFmt, fmtSpec, fmtSpec, ctxArgs, pubBits, portBits)
	fmt.Fprintf(b, "%s}\n", modeLoopIndent)
	fmt.Fprintf(b, "%sif pf.bits() != map_port_flags(praw) {\n", modeLoopIndent)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: flag mismatch public={:#x} port={:#x}\", %s, pf.bits(), map_port_flags(praw)));\n", modeLoopIndent, row.GoSymbol, ctxFmt, ctxArgs)
	fmt.Fprintf(b, "%s}\n", modeLoopIndent)
	fmt.Fprintf(b, "%scount += 1;\n", modeLoopIndent)
	if hasMode {
		fmt.Fprintf(b, "%s}\n", indent)
	}
	b.WriteString("    }\n")
	b.WriteString("    count\n}\n\n")
	return funcName, cases, nil
}

// publicParityModeOrderNames gives emitModeUnary/emitConvertInt/emitFromMode/
// emitContextBinaryWithFlags the mode count without importing the Go-side
// RoundingMode identifiers (this file only needs the count and, for
// emitConvertInt, the ordered bidgo variant-name suffixes below).
var publicParityModeOrderNames = []string{"NearestEven", "NearestAway", "TowardZero", "TowardPositive", "TowardNegative"}

// emitModeBinary renders the binary_mode_flags shape ({Add,Sub,Mul,Div}WithMode):
// two operand legs under the PARITY_MODES cycle, always comparing result bits
// and flags against the port called at port_mode (not the hardcoded
// BIDGO_ROUND_NEAREST_EVEN emitSimpleOp passes). The shared PAIRS_<w> corpus
// probes NaN/infinity/sign/noncanonical routing but is mostly exact under
// every mode (its Mul/Div combinations at widths 32/64 never round), so a
// second leg runs the per-operation mode-discriminant table
// (modeBinaryDiscriminants, the same generation-time source the Go leg uses,
// so both runners exercise identical bit patterns): operand pairs whose exact
// result does not fit the width's precision, with a generated per-pair
// assertion that the five wrapper results are not all identical. A wrapper
// that drops its RoundingMode fails both the per-mode bit compare and that
// discrimination assertion, and a corpus edit that stops discriminating fails
// the gate instead of silently making it vacuous. The port call's per-width
// flags convention (pfpsf vs tuple) comes from portCallStmt, identical to
// every other flags-carrying shape.
func emitModeBinary(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, w parityWidth) (string, int, error) {
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	op := modeBinaryOpFromGoSymbol(row.GoSymbol)
	disc, err := modeBinaryDiscriminantOperands(op, parityWidthDigits(w))
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := (w.pairsLen(corpus) + len(disc)) * len(publicParityModeOrderNames)

	pubBits, portBits, fmtSpec := w.pubBitsExpr("pv"), w.portBitsExpr("pr"), w.resultFmtSpec()

	fmt.Fprintf(b, "fn %s(failures: &mut Vec<String>) -> usize {\n", funcName)
	fmt.Fprintf(b, "    let mut count = 0usize;\n")

	// Shared routing-corpus leg.
	fmt.Fprintf(b, "    for &(i0, i1) in %s {\n", w.pairs)
	fmt.Fprintf(b, "        let v0 = %s[i0];\n", w.corpus)
	fmt.Fprintf(b, "        let v1 = %s[i1];\n", w.corpus)
	fmt.Fprintf(b, "        for &(mode, port_mode) in PARITY_MODES {\n")
	indent := "            "

	pubCall := fmt.Sprintf("%s.%s(%s, mode)", w.pubFrom("v0"), row.RustSurface, w.pubFrom("v1"))
	fmt.Fprintf(b, "%slet (pv, pf) = %s;\n", indent, pubCall)

	portArgs := []string{w.portArg("v0"), w.portArg("v1"), "port_mode"}
	b.WriteString(indent + portCallStmt(row.BidgoFunction, module, fn, portArgs) + "\n")

	ctxFmt := "operands " + w.opFmtSpec() + "," + w.opFmtSpec() + " mode {:?}"
	ctxArgs := "v0, v1, mode"
	fmt.Fprintf(b, "%sif %s != %s {\n", indent, pubBits, portBits)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: result mismatch public=%s port=%s\", %s, %s, %s));\n", indent, row.GoSymbol, ctxFmt, fmtSpec, fmtSpec, ctxArgs, pubBits, portBits)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%sif pf.bits() != map_port_flags(praw) {\n", indent)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: flag mismatch public={:#x} port={:#x}\", %s, pf.bits(), map_port_flags(praw)));\n", indent, row.GoSymbol, ctxFmt, ctxArgs)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%scount += 1;\n", indent)
	b.WriteString("        }\n")
	b.WriteString("    }\n")

	// Mode-discriminant leg: inexact operand pairs plus the per-pair
	// discrimination assertion.
	var discTupleType, seenInit string
	switch w.selfType {
	case "Decimal32":
		discTupleType, seenInit = "(u32, u32)", "[0u32; 5]"
	case "Decimal128":
		discTupleType, seenInit = "([u8; 16], [u8; 16])", "[[0u8; 16]; 5]"
	default:
		discTupleType, seenInit = "(u64, u64)", "[0u64; 5]"
	}
	fmt.Fprintf(b, "    let disc_pairs: &[%s] = &[\n", discTupleType)
	for _, dp := range disc {
		aLit, err := modeDiscRustLiteral(parityWidthDigits(w), dp[0])
		if err != nil {
			return "", 0, fmt.Errorf("%s: %w", row.GoSymbol, err)
		}
		bLit, err := modeDiscRustLiteral(parityWidthDigits(w), dp[1])
		if err != nil {
			return "", 0, fmt.Errorf("%s: %w", row.GoSymbol, err)
		}
		fmt.Fprintf(b, "        (%s, %s),\n", aLit, bLit)
	}
	fmt.Fprintf(b, "    ];\n")
	fmt.Fprintf(b, "    for &(da, db) in disc_pairs {\n")
	fmt.Fprintf(b, "        let mut mode_seen = %s;\n", seenInit)
	fmt.Fprintf(b, "        for (mi, &(mode, port_mode)) in PARITY_MODES.iter().enumerate() {\n")
	discCall := fmt.Sprintf("%s.%s(%s, mode)", w.pubFrom("da"), row.RustSurface, w.pubFrom("db"))
	fmt.Fprintf(b, "%slet (pv, pf) = %s;\n", indent, discCall)
	discPortArgs := []string{w.portArg("da"), w.portArg("db"), "port_mode"}
	b.WriteString(indent + portCallStmt(row.BidgoFunction, module, fn, discPortArgs) + "\n")
	discCtxFmt := "discriminant operands " + w.opFmtSpec() + "," + w.opFmtSpec() + " mode {:?}"
	discCtxArgs := "da, db, mode"
	fmt.Fprintf(b, "%sif %s != %s {\n", indent, pubBits, portBits)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: result mismatch public=%s port=%s\", %s, %s, %s));\n", indent, row.GoSymbol, discCtxFmt, fmtSpec, fmtSpec, discCtxArgs, pubBits, portBits)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%sif pf.bits() != map_port_flags(praw) {\n", indent)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: flag mismatch public={:#x} port={:#x}\", %s, pf.bits(), map_port_flags(praw)));\n", indent, row.GoSymbol, discCtxFmt, discCtxArgs)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%smode_seen[mi] = %s;\n", indent, w.pubBitsExpr("pv"))
	fmt.Fprintf(b, "%scount += 1;\n", indent)
	b.WriteString("        }\n")
	fmt.Fprintf(b, "        if mode_seen.iter().all(|s| *s == mode_seen[0]) {\n")
	fmt.Fprintf(b, "            failures.push(format!(\"public parity %s: discriminant operands %s,%s: every rounding mode produced the same result; the mode-discriminant corpus entry no longer discriminates\", da, db));\n", row.GoSymbol, w.opFmtSpec(), w.opFmtSpec())
	fmt.Fprintf(b, "        }\n")
	b.WriteString("    }\n")
	b.WriteString("    count\n}\n\n")
	return funcName, cases, nil
}

// modeBinaryOpFromGoSymbol extracts the operation name a binary_mode_flags
// census row exercises: the method name after the receiver dot, with the
// WithMode suffix stripped ("Decimal64BID.MulWithMode" -> "Mul").
func modeBinaryOpFromGoSymbol(goSymbol string) string {
	method := goSymbol
	if idx := strings.LastIndex(goSymbol, "."); idx >= 0 {
		method = goSymbol[idx+1:]
	}
	return strings.TrimSuffix(method, "WithMode")
}

// parityWidthDigits maps a parityWidth record to its numeric width for
// lookups in the width-keyed generation-time tables.
func parityWidthDigits(w parityWidth) int {
	switch w.selfType {
	case "Decimal32":
		return 32
	case "Decimal128":
		return 128
	default:
		return 64
	}
}

// emitParseMode renders the parse_mode shape (NewDecimal<w>WithMode /
// parse_with_mode): the full string corpus under every PARITY_MODES entry
// with the same documented parsing contract as parse_with_flags (blank and
// digitless-finite inputs error, NaN-literal branch pins
// NaN-ness/signaling/zero flags, port inputs bit/flag-compare against the
// from-string port called at port_mode and pin the
// error-iff-non-NaN-grammar-NaN rule), then the parse mode-discriminant
// literals (decimal strings one digit past the width's precision, the same
// generation-time source the Go leg uses) with the per-case
// not-all-identical assertion.
func emitParseMode(b *strings.Builder, row rustParityInventoryRow, w parityWidth) (string, int, error) {
	fromStringCall, nanExpr, err := resolveFromStringModePort(row.BidgoFunction, row.GoSymbol, w, "sc.input", "port_mode")
	if err != nil {
		return "", 0, err
	}
	disc, err := modeParseDiscriminantStrings(parityWidthDigits(w))
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := (len(publicParityStringCorpusCases) + len(disc)) * len(publicParityModeOrderNames)

	b.WriteString(rustTmpl(`fn @FUNC@(failures: &mut Vec<String>) -> usize {
    let mut count = 0usize;
    for sc in STRING_CASES {
        for &(mode, port_mode) in PARITY_MODES {
            let got = @SELF@::parse_with_mode(sc.input, mode);
            match sc.kind {
                "nan_literal" => match (sc.nan_min_width != 0 && sc.nan_min_width <= @WIDTH@, got) {
                    (true, Ok((pv, pf))) => {
                        if !pv.is_nan() {
                            failures.push(format!("public parity @SYM@: input {:?} mode {:?}: expected a NaN result from the NaN literal branch", sc.input, mode));
                        }
                        if pv.is_signaling() != sc.signaling {
                            failures.push(format!("public parity @SYM@: input {:?} mode {:?}: signaling bit mismatch public={} literal={}", sc.input, mode, pv.is_signaling(), sc.signaling));
                        }
                        if !pf.is_empty() {
                            failures.push(format!("public parity @SYM@: input {:?} mode {:?}: NaN literal branch must raise no flags, got {:?}", sc.input, mode, pf));
                        }
                    }
                    (true, Err(e)) => failures.push(format!("public parity @SYM@: input {:?} mode {:?}: unexpected error {}", sc.input, mode, e)),
                    (false, Ok((pv, pf))) => failures.push(format!("public parity @SYM@: input {:?} mode {:?}: unrepresentable NaN payload returned Ok((@FMTSPEC@, {:#x}))", sc.input, mode, @PVBITS@, pf.bits())),
                    (false, Err(e)) => {
                        if e.input() != sc.input {
                            failures.push(format!("public parity @SYM@: input {:?} mode {:?}: ParseDecimalError retained {:?}", sc.input, mode, e.input()));
                        }
                    }
                },
                "blank" | "invalid_syntax" => match got {
                    Ok((pv, pf)) => failures.push(format!("public parity @SYM@: input {:?} mode {:?}: rejected input kind {:?} returned Ok((@FMTSPEC@, {:#x}))", sc.input, mode, sc.kind, @PVBITS@, pf.bits())),
                    Err(e) => {
                        if e.input() != sc.input {
                            failures.push(format!("public parity @SYM@: input {:?} mode {:?}: ParseDecimalError retained {:?}", sc.input, mode, e.input()));
                        }
                    }
                },
                _ => {
                    let (pr, praw) = @FROMSTR@;
                    let pr_is_nan = @NANEXPR@;
                    let port_flags = map_port_flags(praw);
                    let silent_cohort_coercion = sc.kind == "cohort_coercion"
                        && (sc.cohort_min_width == 0 || sc.cohort_min_width > @WIDTH@)
                        && praw == 0;
                    let should_error = pr_is_nan
                        || port_flags & bid754::ExceptionFlags::INVALID_OPERATION.bits() != 0
                        || silent_cohort_coercion;
                    match got {
                        Ok((pv, pf)) if should_error => {
                            failures.push(format!("public parity @SYM@: input {:?} mode {:?}: expected an error, got Ok((@FMTSPEC@, {:#x}))", sc.input, mode, @PVBITS@, pf.bits()));
                        }
                        Ok((pv, pf)) => {
                            if @PVBITS@ != @PORTBITS@ {
                                failures.push(format!("public parity @SYM@: input {:?} mode {:?}: result mismatch public=@FMTSPEC@ port=@FMTSPEC@", sc.input, mode, @PVBITS@, @PORTBITS@));
                            }
                            if pf.bits() != port_flags {
                                failures.push(format!("public parity @SYM@: input {:?} mode {:?}: flag mismatch public={:#x} port={:#x}", sc.input, mode, pf.bits(), port_flags));
                            }
                        }
                        Err(e) if !should_error => {
                            failures.push(format!("public parity @SYM@: input {:?} mode {:?}: unexpected error {} for port result @FMTSPEC@", sc.input, mode, e, @PORTBITS@));
                        }
                        Err(e) => {
                            if e.input() != sc.input {
                                failures.push(format!("public parity @SYM@: input {:?} mode {:?}: ParseDecimalError retained {:?}", sc.input, mode, e.input()));
                            }
                        }
                    }
                }
            }
            count += 1;
        }
    }
`, "@FUNC@", funcName, "@SELF@", w.selfType, "@WIDTH@", strconv.Itoa(parityWidthDigits(w)), "@SYM@", row.GoSymbol,
		"@FROMSTR@", fromStringCall, "@NANEXPR@", nanExpr,
		"@PVBITS@", w.pubBitsExpr("pv"), "@PORTBITS@", w.portBitsExpr("pr"), "@FMTSPEC@", w.resultFmtSpec()))

	// Mode-discriminant leg.
	_, seenInit := rustModeDiscTypes(w)
	discCall, discNanExpr, err := resolveFromStringModePort(row.BidgoFunction, row.GoSymbol, w, "ds", "port_mode")
	if err != nil {
		return "", 0, err
	}
	_ = discNanExpr
	fmt.Fprintf(b, "    let disc_inputs: &[&str] = &[\n")
	for _, ds := range disc {
		fmt.Fprintf(b, "        %q,\n", ds)
	}
	fmt.Fprintf(b, "    ];\n")
	fmt.Fprintf(b, "    for &ds in disc_inputs {\n")
	fmt.Fprintf(b, "        let mut mode_seen = %s;\n", seenInit)
	fmt.Fprintf(b, "        for (mi, &(mode, port_mode)) in PARITY_MODES.iter().enumerate() {\n")
	indent := "            "
	fmt.Fprintf(b, "%smatch %s::parse_with_mode(ds, mode) {\n", indent, w.selfType)
	fmt.Fprintf(b, "%s    Ok((pv, pf)) => {\n", indent)
	fmt.Fprintf(b, "%s        let (pr, praw) = %s;\n", indent, discCall)
	fmt.Fprintf(b, "%s        if %s != %s {\n", indent, w.pubBitsExpr("pv"), w.portBitsExpr("pr"))
	fmt.Fprintf(b, "%s            failures.push(format!(\"public parity %s: discriminant input {:?} mode {:?}: result mismatch public=%s port=%s\", ds, mode, %s, %s));\n", indent, row.GoSymbol, w.resultFmtSpec(), w.resultFmtSpec(), w.pubBitsExpr("pv"), w.portBitsExpr("pr"))
	fmt.Fprintf(b, "%s        }\n", indent)
	fmt.Fprintf(b, "%s        if pf.bits() != map_port_flags(praw) {\n", indent)
	fmt.Fprintf(b, "%s            failures.push(format!(\"public parity %s: discriminant input {:?} mode {:?}: flag mismatch public={:#x} port={:#x}\", ds, mode, pf.bits(), map_port_flags(praw)));\n", indent, row.GoSymbol)
	fmt.Fprintf(b, "%s        }\n", indent)
	fmt.Fprintf(b, "%s        mode_seen[mi] = %s;\n", indent, w.pubBitsExpr("pv"))
	fmt.Fprintf(b, "%s    }\n", indent)
	fmt.Fprintf(b, "%s    Err(e) => failures.push(format!(\"public parity %s: discriminant input {:?} mode {:?}: unexpected error {}\", ds, mode, e)),\n", indent, row.GoSymbol)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%scount += 1;\n", indent)
	b.WriteString("        }\n")
	fmt.Fprintf(b, "        if mode_seen.iter().all(|s| *s == mode_seen[0]) {\n")
	fmt.Fprintf(b, "            failures.push(format!(\"public parity %s: discriminant input {:?}: every rounding mode produced the same result; the mode-discriminant corpus entry no longer discriminates\", ds));\n", row.GoSymbol)
	fmt.Fprintf(b, "        }\n")
	b.WriteString("    }\n")
	b.WriteString("    count\n}\n\n")
	return funcName, cases, nil
}

// rustModeDiscTypes gives the Rust corpus element type, seen-array
// initializer, and tuple element type string for a width's discriminant leg.
func rustModeDiscTypes(w parityWidth) (elemType, seenInit string) {
	switch w.selfType {
	case "Decimal32":
		return "u32", "[0u32; 5]"
	case "Decimal128":
		return "[u8; 16]", "[[0u8; 16]; 5]"
	default:
		return "u64", "[0u64; 5]"
	}
}

// emitModeUnaryArith renders the unary_mode_flags shape: the shared corpus leg
// plus the unary mode-discriminant leg
// (modeUnaryDiscriminants, the same generation-time source the Go leg uses)
// with the per-case not-all-identical assertion, both cycled through
// PARITY_MODES against the port called at port_mode.
func emitModeUnaryArith(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, w parityWidth) (string, int, error) {
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	op := modeBinaryOpFromGoSymbol(row.GoSymbol)
	disc, err := modeUnaryDiscriminantOperands(op, parityWidthDigits(w))
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := (w.bitsLen(corpus) + len(disc)) * len(publicParityModeOrderNames)

	pubBits, portBits, fmtSpec := w.pubBitsExpr("pv"), w.portBitsExpr("pr"), w.resultFmtSpec()
	indent := "            "

	fmt.Fprintf(b, "fn %s(failures: &mut Vec<String>) -> usize {\n", funcName)
	fmt.Fprintf(b, "    let mut count = 0usize;\n")

	// Shared routing-corpus leg.
	fmt.Fprintf(b, "    for &v0 in %s {\n", w.corpus)
	fmt.Fprintf(b, "        for &(mode, port_mode) in PARITY_MODES {\n")
	fmt.Fprintf(b, "%slet (pv, pf) = %s.%s(mode);\n", indent, w.pubFrom("v0"), row.RustSurface)
	b.WriteString(indent + portCallStmt(row.BidgoFunction, module, fn, []string{w.portArg("v0"), "port_mode"}) + "\n")
	ctxFmt := "operand " + w.opFmtSpec() + " mode {:?}"
	fmt.Fprintf(b, "%sif %s != %s {\n", indent, pubBits, portBits)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: result mismatch public=%s port=%s\", v0, mode, %s, %s));\n", indent, row.GoSymbol, ctxFmt, fmtSpec, fmtSpec, pubBits, portBits)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%sif pf.bits() != map_port_flags(praw) {\n", indent)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: flag mismatch public={:#x} port={:#x}\", v0, mode, pf.bits(), map_port_flags(praw)));\n", indent, row.GoSymbol, ctxFmt)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%scount += 1;\n", indent)
	b.WriteString("        }\n    }\n")

	// Mode-discriminant leg.
	elemType, seenInit := rustModeDiscTypes(w)
	fmt.Fprintf(b, "    let disc_vals: &[%s] = &[\n", elemType)
	for _, dv := range disc {
		lit, err := modeDiscRustLiteral(parityWidthDigits(w), dv)
		if err != nil {
			return "", 0, fmt.Errorf("%s: %w", row.GoSymbol, err)
		}
		fmt.Fprintf(b, "        %s,\n", lit)
	}
	fmt.Fprintf(b, "    ];\n")
	fmt.Fprintf(b, "    for &dv in disc_vals {\n")
	fmt.Fprintf(b, "        let mut mode_seen = %s;\n", seenInit)
	fmt.Fprintf(b, "        for (mi, &(mode, port_mode)) in PARITY_MODES.iter().enumerate() {\n")
	fmt.Fprintf(b, "%slet (pv, pf) = %s.%s(mode);\n", indent, w.pubFrom("dv"), row.RustSurface)
	b.WriteString(indent + portCallStmt(row.BidgoFunction, module, fn, []string{w.portArg("dv"), "port_mode"}) + "\n")
	discCtxFmt := "discriminant operand " + w.opFmtSpec() + " mode {:?}"
	fmt.Fprintf(b, "%sif %s != %s {\n", indent, pubBits, portBits)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: result mismatch public=%s port=%s\", dv, mode, %s, %s));\n", indent, row.GoSymbol, discCtxFmt, fmtSpec, fmtSpec, pubBits, portBits)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%sif pf.bits() != map_port_flags(praw) {\n", indent)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: flag mismatch public={:#x} port={:#x}\", dv, mode, pf.bits(), map_port_flags(praw)));\n", indent, row.GoSymbol, discCtxFmt)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%smode_seen[mi] = %s;\n", indent, w.pubBitsExpr("pv"))
	fmt.Fprintf(b, "%scount += 1;\n", indent)
	b.WriteString("        }\n")
	fmt.Fprintf(b, "        if mode_seen.iter().all(|s| *s == mode_seen[0]) {\n")
	fmt.Fprintf(b, "            failures.push(format!(\"public parity %s: discriminant operand %s: every rounding mode produced the same result; the mode-discriminant corpus entry no longer discriminates\", dv));\n", row.GoSymbol, w.opFmtSpec())
	fmt.Fprintf(b, "        }\n")
	b.WriteString("    }\n")
	b.WriteString("    count\n}\n\n")
	return funcName, cases, nil
}

// emitModeTernary renders the ternary_mode_flags shape (FMAWithMode): the
// shared triples leg plus the ternary mode-discriminant leg with the
// per-case assertion (emitModeBinary's structure at arity 3).
func emitModeTernary(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, w parityWidth) (string, int, error) {
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	op := modeBinaryOpFromGoSymbol(row.GoSymbol)
	disc, err := modeTernaryDiscriminantOperands(op, parityWidthDigits(w))
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := (w.triplesLen(corpus) + len(disc)) * len(publicParityModeOrderNames)

	pubBits, portBits, fmtSpec := w.pubBitsExpr("pv"), w.portBitsExpr("pr"), w.resultFmtSpec()
	indent := "            "

	fmt.Fprintf(b, "fn %s(failures: &mut Vec<String>) -> usize {\n", funcName)
	fmt.Fprintf(b, "    let mut count = 0usize;\n")

	// Shared routing-corpus leg.
	fmt.Fprintf(b, "    for &(i0, i1, i2) in %s {\n", w.triples)
	fmt.Fprintf(b, "        let v0 = %s[i0];\n", w.corpus)
	fmt.Fprintf(b, "        let v1 = %s[i1];\n", w.corpus)
	fmt.Fprintf(b, "        let v2 = %s[i2];\n", w.corpus)
	fmt.Fprintf(b, "        for &(mode, port_mode) in PARITY_MODES {\n")
	fmt.Fprintf(b, "%slet (pv, pf) = %s.%s(%s, %s, mode);\n", indent, w.pubFrom("v0"), row.RustSurface, w.pubFrom("v1"), w.pubFrom("v2"))
	b.WriteString(indent + portCallStmt(row.BidgoFunction, module, fn, []string{w.portArg("v0"), w.portArg("v1"), w.portArg("v2"), "port_mode"}) + "\n")
	ctxFmt := "operands " + w.opFmtSpec() + "," + w.opFmtSpec() + "," + w.opFmtSpec() + " mode {:?}"
	fmt.Fprintf(b, "%sif %s != %s {\n", indent, pubBits, portBits)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: result mismatch public=%s port=%s\", v0, v1, v2, mode, %s, %s));\n", indent, row.GoSymbol, ctxFmt, fmtSpec, fmtSpec, pubBits, portBits)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%sif pf.bits() != map_port_flags(praw) {\n", indent)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: flag mismatch public={:#x} port={:#x}\", v0, v1, v2, mode, pf.bits(), map_port_flags(praw)));\n", indent, row.GoSymbol, ctxFmt)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%scount += 1;\n", indent)
	b.WriteString("        }\n    }\n")

	// Mode-discriminant leg.
	elemType, seenInit := rustModeDiscTypes(w)
	fmt.Fprintf(b, "    let disc_triples: &[(%s, %s, %s)] = &[\n", elemType, elemType, elemType)
	for _, dt := range disc {
		aLit, err := modeDiscRustLiteral(parityWidthDigits(w), dt[0])
		if err != nil {
			return "", 0, fmt.Errorf("%s: %w", row.GoSymbol, err)
		}
		bLit, err := modeDiscRustLiteral(parityWidthDigits(w), dt[1])
		if err != nil {
			return "", 0, fmt.Errorf("%s: %w", row.GoSymbol, err)
		}
		cLit, err := modeDiscRustLiteral(parityWidthDigits(w), dt[2])
		if err != nil {
			return "", 0, fmt.Errorf("%s: %w", row.GoSymbol, err)
		}
		fmt.Fprintf(b, "        (%s, %s, %s),\n", aLit, bLit, cLit)
	}
	fmt.Fprintf(b, "    ];\n")
	fmt.Fprintf(b, "    for &(da, db, dc) in disc_triples {\n")
	fmt.Fprintf(b, "        let mut mode_seen = %s;\n", seenInit)
	fmt.Fprintf(b, "        for (mi, &(mode, port_mode)) in PARITY_MODES.iter().enumerate() {\n")
	fmt.Fprintf(b, "%slet (pv, pf) = %s.%s(%s, %s, mode);\n", indent, w.pubFrom("da"), row.RustSurface, w.pubFrom("db"), w.pubFrom("dc"))
	b.WriteString(indent + portCallStmt(row.BidgoFunction, module, fn, []string{w.portArg("da"), w.portArg("db"), w.portArg("dc"), "port_mode"}) + "\n")
	discCtxFmt := "discriminant operands " + w.opFmtSpec() + "," + w.opFmtSpec() + "," + w.opFmtSpec() + " mode {:?}"
	fmt.Fprintf(b, "%sif %s != %s {\n", indent, pubBits, portBits)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: result mismatch public=%s port=%s\", da, db, dc, mode, %s, %s));\n", indent, row.GoSymbol, discCtxFmt, fmtSpec, fmtSpec, pubBits, portBits)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%sif pf.bits() != map_port_flags(praw) {\n", indent)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: flag mismatch public={:#x} port={:#x}\", da, db, dc, mode, pf.bits(), map_port_flags(praw)));\n", indent, row.GoSymbol, discCtxFmt)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%smode_seen[mi] = %s;\n", indent, w.pubBitsExpr("pv"))
	fmt.Fprintf(b, "%scount += 1;\n", indent)
	b.WriteString("        }\n")
	fmt.Fprintf(b, "        if mode_seen.iter().all(|s| *s == mode_seen[0]) {\n")
	fmt.Fprintf(b, "            failures.push(format!(\"public parity %s: discriminant operands %s,%s,%s: every rounding mode produced the same result; the mode-discriminant corpus entry no longer discriminates\", da, db, dc));\n", row.GoSymbol, w.opFmtSpec(), w.opFmtSpec(), w.opFmtSpec())
	fmt.Fprintf(b, "        }\n")
	b.WriteString("    }\n")
	b.WriteString("    count\n}\n\n")
	return funcName, cases, nil
}

// emitModeScaleB renders the scaleb_mode shape (ScaleBWithMode): the shared
// corpus x SCALEB_EXPS leg plus the scaleB mode-discriminant leg (operand +
// boundary exponent tuples) with the per-case assertion. scaleB is exact
// inside the format range, so the shared leg alone cannot prove mode
// forwarding; the discriminant exponents sit at the overflow/underflow
// boundary where the five directions split.
func emitModeScaleB(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, w parityWidth) (string, int, error) {
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	op := modeBinaryOpFromGoSymbol(row.GoSymbol)
	disc, err := modeScaleBDiscriminantCases(op, parityWidthDigits(w))
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := (w.bitsLen(corpus)*len(publicParityScaleBExps) + len(disc)) * len(publicParityModeOrderNames)

	pubBits, portBits, fmtSpec := w.pubBitsExpr("pv"), w.portBitsExpr("pr"), w.resultFmtSpec()
	indent := "                "

	fmt.Fprintf(b, "fn %s(failures: &mut Vec<String>) -> usize {\n", funcName)
	fmt.Fprintf(b, "    let mut count = 0usize;\n")

	// Shared routing-corpus leg.
	fmt.Fprintf(b, "    for &v0 in %s {\n", w.corpus)
	fmt.Fprintf(b, "        for &exp in SCALEB_EXPS {\n")
	fmt.Fprintf(b, "            for &(mode, port_mode) in PARITY_MODES {\n")
	fmt.Fprintf(b, "%slet (pv, pf) = %s.%s(exp, mode);\n", indent, w.pubFrom("v0"), row.RustSurface)
	b.WriteString(indent + portCallStmt(row.BidgoFunction, module, fn, []string{w.portArg("v0"), "exp", "port_mode"}) + "\n")
	ctxFmt := "operand " + w.opFmtSpec() + " exp {} mode {:?}"
	fmt.Fprintf(b, "%sif %s != %s {\n", indent, pubBits, portBits)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: result mismatch public=%s port=%s\", v0, exp, mode, %s, %s));\n", indent, row.GoSymbol, ctxFmt, fmtSpec, fmtSpec, pubBits, portBits)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%sif pf.bits() != map_port_flags(praw) {\n", indent)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: flag mismatch public={:#x} port={:#x}\", v0, exp, mode, pf.bits(), map_port_flags(praw)));\n", indent, row.GoSymbol, ctxFmt)
	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%scount += 1;\n", indent)
	b.WriteString("            }\n        }\n    }\n")

	// Mode-discriminant leg.
	elemType, seenInit := rustModeDiscTypes(w)
	discIndent := "            "
	fmt.Fprintf(b, "    let disc_cases: &[(%s, i64)] = &[\n", elemType)
	for _, dc := range disc {
		lit, err := modeDiscRustLiteral(parityWidthDigits(w), dc.Val)
		if err != nil {
			return "", 0, fmt.Errorf("%s: %w", row.GoSymbol, err)
		}
		fmt.Fprintf(b, "        (%s, %d),\n", lit, dc.Exp)
	}
	fmt.Fprintf(b, "    ];\n")
	fmt.Fprintf(b, "    for &(dv, dexp) in disc_cases {\n")
	fmt.Fprintf(b, "        let mut mode_seen = %s;\n", seenInit)
	fmt.Fprintf(b, "        for (mi, &(mode, port_mode)) in PARITY_MODES.iter().enumerate() {\n")
	fmt.Fprintf(b, "%slet (pv, pf) = %s.%s(dexp, mode);\n", discIndent, w.pubFrom("dv"), row.RustSurface)
	b.WriteString(discIndent + portCallStmt(row.BidgoFunction, module, fn, []string{w.portArg("dv"), "dexp", "port_mode"}) + "\n")
	discCtxFmt := "discriminant operand " + w.opFmtSpec() + " exp {} mode {:?}"
	fmt.Fprintf(b, "%sif %s != %s {\n", discIndent, pubBits, portBits)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: result mismatch public=%s port=%s\", dv, dexp, mode, %s, %s));\n", discIndent, row.GoSymbol, discCtxFmt, fmtSpec, fmtSpec, pubBits, portBits)
	fmt.Fprintf(b, "%s}\n", discIndent)
	fmt.Fprintf(b, "%sif pf.bits() != map_port_flags(praw) {\n", discIndent)
	fmt.Fprintf(b, "%s    failures.push(format!(\"public parity %s: %s: flag mismatch public={:#x} port={:#x}\", dv, dexp, mode, pf.bits(), map_port_flags(praw)));\n", discIndent, row.GoSymbol, discCtxFmt)
	fmt.Fprintf(b, "%s}\n", discIndent)
	fmt.Fprintf(b, "%smode_seen[mi] = %s;\n", discIndent, w.pubBitsExpr("pv"))
	fmt.Fprintf(b, "%scount += 1;\n", discIndent)
	b.WriteString("        }\n")
	fmt.Fprintf(b, "        if mode_seen.iter().all(|s| *s == mode_seen[0]) {\n")
	fmt.Fprintf(b, "            failures.push(format!(\"public parity %s: discriminant operand %s exp {}: every rounding mode produced the same result; the mode-discriminant corpus entry no longer discriminates\", dv, dexp));\n", row.GoSymbol, w.opFmtSpec())
	fmt.Fprintf(b, "        }\n")
	b.WriteString("    }\n")
	b.WriteString("    count\n}\n\n")
	return funcName, cases, nil
}

// modeDiscRustLiteral renders one encoded mode-discriminant operand as a Rust
// literal of the width's corpus element type.
func modeDiscRustLiteral(width int, o modeDiscOperand) (string, error) {
	switch width {
	case 32:
		bits, err := encodeModeDiscOperand32(o)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("0x%08x", bits), nil
	case 64:
		bits, err := encodeModeDiscOperand64(o)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("0x%016x", bits), nil
	default:
		raw, err := encodeModeDiscOperand128(o)
		if err != nil {
			return "", err
		}
		parts := make([]string, len(raw))
		for i, by := range raw {
			parts[i] = fmt.Sprintf("0x%02x", by)
		}
		return "[" + strings.Join(parts, ", ") + "]", nil
	}
}

func emitScaleB(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, w parityWidth) (string, int, error) {
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := w.bitsLen(corpus) * len(publicParityScaleBExps)
	portStmt := portCallStmt(row.BidgoFunction, module, fn, []string{w.portArg("v0"), "exp", "BIDGO_ROUND_NEAREST_EVEN"})

	b.WriteString(rustTmpl(`fn @FUNC@(failures: &mut Vec<String>) -> usize {
    let mut count = 0usize;
    for &v0 in @CORPUS@ {
        for &exp in SCALEB_EXPS {
            let (pv, pf) = @PUBFROM@.@SURF@(exp);
            @PORTSTMT@
            if @PUBBITS@ != @PORTBITS@ {
                failures.push(format!("public parity @SYM@: operand @OPFMT@ exp {}: result mismatch public=@FMTSPEC@ port=@FMTSPEC@", v0, exp, @PUBBITS@, @PORTBITS@));
            }
            if pf.bits() != map_port_flags(praw) {
                failures.push(format!("public parity @SYM@: operand @OPFMT@ exp {}: flag mismatch public={:#x} port={:#x}", v0, exp, pf.bits(), map_port_flags(praw)));
            }
            count += 1;
        }
    }
    count
}

`, "@FUNC@", funcName, "@CORPUS@", w.corpus, "@PUBFROM@", w.pubFrom("v0"), "@SURF@", row.RustSurface, "@PORTSTMT@", portStmt, "@PUBBITS@", w.pubBitsExpr("pv"), "@PORTBITS@", w.portBitsExpr("pr"), "@OPFMT@", w.opFmtSpec(), "@FMTSPEC@", w.resultFmtSpec(), "@SYM@", row.GoSymbol))
	return funcName, cases, nil
}

const nextTowardTargets = 5

func emitNextToward(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, w parityWidth) (string, int, error) {
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := w.bitsLen(corpus) * nextTowardTargets
	// NextToward's target parameter is always Decimal128-width for every
	// receiver width (matches every Go width's NextToward), so the target
	// port arg is always to_port128(tb) regardless of w -- only the receiver
	// operand v0 uses w's own portArg. Bid<w>NextToward is tuple-return for
	// every width (verified), so no pfpsf branch is needed here.
	b.WriteString(rustTmpl(`fn @FUNC@(failures: &mut Vec<String>) -> usize {
    let mut count = 0usize;
    for &v0 in @CORPUS@ {
        for &tb in &CORPUS_128[..@N@] {
            let target = Decimal128::from_le_bytes(tb);
            let (pv, pf) = @PUBFROM@.@SURF@(target);
            let (pr, praw) = bid754::generated::@MOD@::@FN@(@PORTV0@, to_port128(tb));
            if @PUBBITS@ != @PORTBITS@ {
                failures.push(format!("public parity @SYM@: operand @OPFMT@ target {:x?}: result mismatch public=@FMTSPEC@ port=@FMTSPEC@", v0, tb, @PUBBITS@, @PORTBITS@));
            }
            if pf.bits() != map_port_flags(praw) {
                failures.push(format!("public parity @SYM@: operand @OPFMT@ target {:x?}: flag mismatch public={:#x} port={:#x}", v0, tb, pf.bits(), map_port_flags(praw)));
            }
            count += 1;
        }
    }
    count
}

`, "@FUNC@", funcName, "@CORPUS@", w.corpus, "@N@", strconv.Itoa(nextTowardTargets), "@PUBFROM@", w.pubFrom("v0"), "@SURF@", row.RustSurface, "@MOD@", module, "@FN@", fn, "@PORTV0@", w.portArg("v0"), "@PUBBITS@", w.pubBitsExpr("pv"), "@PORTBITS@", w.portBitsExpr("pr"), "@OPFMT@", w.opFmtSpec(), "@FMTSPEC@", w.resultFmtSpec(), "@SYM@", row.GoSymbol))
	return funcName, cases, nil
}

func emitClass(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, w parityWidth) (string, int, error) {
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := w.bitsLen(corpus)
	b.WriteString(rustTmpl(`fn @FUNC@(failures: &mut Vec<String>) -> usize {
    let mut count = 0usize;
    for &v0 in @CORPUS@ {
        let pv = @PUBFROM@.@SURF@();
        let pr = bid754::generated::@MOD@::@FN@(@PORTV0@);
        if pv.to_string() != class_name(pr) {
            failures.push(format!("public parity @SYM@: operand @OPFMT@: result mismatch public={} port={}", v0, pv, class_name(pr)));
        }
        count += 1;
    }
    count
}

`, "@FUNC@", funcName, "@CORPUS@", w.corpus, "@PUBFROM@", w.pubFrom("v0"), "@SURF@", row.RustSurface, "@MOD@", module, "@FN@", fn, "@PORTV0@", w.portArg("v0"), "@OPFMT@", w.opFmtSpec(), "@SYM@", row.GoSymbol))
	return funcName, cases, nil
}

func emitSign(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, w parityWidth) (string, int, error) {
	zeroModule, zeroFn, err := resolvePort(w.bidgoIsZero, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	signedModule, signedFn, err := resolvePort(w.bidgoIsSigned, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	// Decimal32's Bid32IsZero already returns bool (apiemit.PortBoolResult);
	// every other width/predicate here returns the Intel-mirroring numeric
	// type compared via "!= 0". See emitSimpleOp's resultBool case for the
	// same distinction.
	zeroExpr := fmt.Sprintf("bid754::generated::%s::%s(%s)", zeroModule, zeroFn, w.portArg("v0"))
	if !apiemit.PortBoolResult(w.bidgoIsZero) {
		zeroExpr += " != 0"
	}
	signedExpr := fmt.Sprintf("bid754::generated::%s::%s(%s)", signedModule, signedFn, w.portArg("v0"))
	if !apiemit.PortBoolResult(w.bidgoIsSigned) {
		signedExpr += " != 0"
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := w.bitsLen(corpus)
	b.WriteString(rustTmpl(`fn @FUNC@(failures: &mut Vec<String>) -> usize {
    let mut count = 0usize;
    for &v0 in @CORPUS@ {
        let pv = @PUBFROM@.@SURF@();
        let is_zero = @ZEROEXPR@;
        let is_signed = @SIGNEDEXPR@;
        let want = if is_zero { 0 } else if is_signed { -1 } else { 1 };
        if pv != want {
            failures.push(format!("public parity @SYM@: operand @OPFMT@: result mismatch public={} port={}", v0, pv, want));
        }
        count += 1;
    }
    count
}

`, "@FUNC@", funcName, "@CORPUS@", w.corpus, "@PUBFROM@", w.pubFrom("v0"), "@SURF@", row.RustSurface, "@ZEROEXPR@", zeroExpr, "@SIGNEDEXPR@", signedExpr, "@OPFMT@", w.opFmtSpec(), "@SYM@", row.GoSymbol))
	return funcName, cases, nil
}

func emitTotalCmp(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, w parityWidth) (string, int, error) {
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := w.pairsLen(corpus)
	b.WriteString(rustTmpl(`fn @FUNC@(failures: &mut Vec<String>) -> usize {
    let mut count = 0usize;
    for &(i0, i1) in @PAIRS@ {
        let v0 = @CORPUS@[i0];
        let v1 = @CORPUS@[i1];
        let pv = @PUBFROM0@.@SURF@(@PUBFROM1@);
        let fwd = bid754::generated::@MOD@::@FN@(@PORTV0@, @PORTV1@);
        let rev = bid754::generated::@MOD@::@FN@(@PORTV1@, @PORTV0@);
        let want = if fwd != 0 && rev != 0 {
            core::cmp::Ordering::Equal
        } else if fwd != 0 {
            core::cmp::Ordering::Less
        } else {
            core::cmp::Ordering::Greater
        };
        if pv != want {
            failures.push(format!("public parity @SYM@: operands @OPFMT@,@OPFMT@: result mismatch public={:?} port={:?}", v0, v1, pv, want));
        }
        count += 1;
    }
    count
}

`, "@FUNC@", funcName, "@PAIRS@", w.pairs, "@CORPUS@", w.corpus, "@PUBFROM0@", w.pubFrom("v0"), "@PUBFROM1@", w.pubFrom("v1"), "@SURF@", row.RustSurface, "@MOD@", module, "@FN@", fn, "@PORTV0@", w.portArg("v0"), "@PORTV1@", w.portArg("v1"), "@OPFMT@", w.opFmtSpec(), "@SYM@", row.GoSymbol))
	return funcName, cases, nil
}

func emitSignalingEqCompose(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, w parityWidth) (string, int, error) {
	geModule, geFn, err := resolvePort(w.bidgoSignalingGE, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	leModule, leFn, err := resolvePort(w.bidgoSignalingLE, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := w.pairsLen(corpus)
	b.WriteString(rustTmpl(`fn @FUNC@(failures: &mut Vec<String>) -> usize {
    let mut count = 0usize;
    for &(i0, i1) in @PAIRS@ {
        let v0 = @CORPUS@[i0];
        let v1 = @CORPUS@[i1];
        let (pv, pf) = @PUBFROM0@.@SURF@(@PUBFROM1@);
        let (ge, ge_raw) = bid754::generated::@GEMOD@::@GEFN@(@PORTV0@, @PORTV1@);
        let (le, le_raw) = bid754::generated::@LEMOD@::@LEFN@(@PORTV0@, @PORTV1@);
        let want = ge != 0 && le != 0;
        if pv != want {
            failures.push(format!("public parity @SYM@: operands @OPFMT@,@OPFMT@: result mismatch public={} port={}", v0, v1, pv, want));
        }
        if pf.bits() != map_port_flags(ge_raw | le_raw) {
            failures.push(format!("public parity @SYM@: operands @OPFMT@,@OPFMT@: flag mismatch public={:#x} port={:#x}", v0, v1, pf.bits(), map_port_flags(ge_raw | le_raw)));
        }
        count += 1;
    }
    count
}

`, "@FUNC@", funcName, "@PAIRS@", w.pairs, "@CORPUS@", w.corpus, "@PUBFROM0@", w.pubFrom("v0"), "@PUBFROM1@", w.pubFrom("v1"), "@SURF@", row.RustSurface, "@GEMOD@", geModule, "@GEFN@", geFn, "@LEMOD@", leModule, "@LEFN@", leFn, "@PORTV0@", w.portArg("v0"), "@PORTV1@", w.portArg("v1"), "@OPFMT@", w.opFmtSpec(), "@SYM@", row.GoSymbol))
	return funcName, cases, nil
}

func emitSignalingNotEqCompose(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, w parityWidth) (string, int, error) {
	geModule, geFn, err := resolvePort(w.bidgoSignalingGE, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	leModule, leFn, err := resolvePort(w.bidgoSignalingLE, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := w.pairsLen(corpus)
	b.WriteString(rustTmpl(`fn @FUNC@(failures: &mut Vec<String>) -> usize {
    let mut count = 0usize;
    for &(i0, i1) in @PAIRS@ {
        let v0 = @CORPUS@[i0];
        let v1 = @CORPUS@[i1];
        let (pv, pf) = @PUBFROM0@.@SURF@(@PUBFROM1@);
        let (ge, ge_raw) = bid754::generated::@GEMOD@::@GEFN@(@PORTV0@, @PORTV1@);
        let (le, le_raw) = bid754::generated::@LEMOD@::@LEFN@(@PORTV0@, @PORTV1@);
        let want = !(ge != 0 && le != 0);
        if pv != want {
            failures.push(format!("public parity @SYM@: operands @OPFMT@,@OPFMT@: result mismatch public={} port={}", v0, v1, pv, want));
        }
        if pf.bits() != map_port_flags(ge_raw | le_raw) {
            failures.push(format!("public parity @SYM@: operands @OPFMT@,@OPFMT@: flag mismatch public={:#x} port={:#x}", v0, v1, pf.bits(), map_port_flags(ge_raw | le_raw)));
        }
        count += 1;
    }
    count
}

`, "@FUNC@", funcName, "@PAIRS@", w.pairs, "@CORPUS@", w.corpus, "@PUBFROM0@", w.pubFrom("v0"), "@PUBFROM1@", w.pubFrom("v1"), "@SURF@", row.RustSurface, "@GEMOD@", geModule, "@GEFN@", geFn, "@LEMOD@", leModule, "@LEFN@", leFn, "@PORTV0@", w.portArg("v0"), "@PORTV1@", w.portArg("v1"), "@OPFMT@", w.opFmtSpec(), "@SYM@", row.GoSymbol))
	return funcName, cases, nil
}

func emitRadixConst(b *strings.Builder, row rustParityInventoryRow, w parityWidth) (string, int, error) {
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	b.WriteString(rustTmpl(`fn @FUNC@(failures: &mut Vec<String>) -> usize {
    let pv = @SELF@::@SURF@;
    let pr = bid754::generated::@MOD@::@FN@();
    if i64::from(pv) != pr {
        failures.push(format!("public parity @SYM@: result mismatch public={} port={}", pv, pr));
    }
    1
}

`, "@FUNC@", funcName, "@SELF@", w.selfType, "@SURF@", row.RustSurface, "@MOD@", module, "@FN@", fn, "@SYM@", row.GoSymbol))
	return funcName, 1, nil
}

func emitCopyFold(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, w parityWidth) (string, int, error) {
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := w.bitsLen(corpus)
	b.WriteString(rustTmpl(`fn @FUNC@(failures: &mut Vec<String>) -> usize {
    let mut count = 0usize;
    for &v0 in @CORPUS@ {
        let a = @PUBFROM@;
        // Rust's Copy (implicit memberwise bit copy on a #[repr(transparent)]
        // single-field struct) is the language-level identity the Go .Copy
        // port entrypoint implements at runtime; comparing the two proves the
        // port's Copy is bit-identity, matching Rust's Copy trait semantics.
        let b_copy = a;
        let pr = bid754::generated::@MOD@::@FN@(@PORTV0@);
        if @COPYBITS@ != @PORTBITS@ {
            failures.push(format!("public parity @SYM@: operand @OPFMT@: result mismatch public=@FMTSPEC@ port=@FMTSPEC@", v0, @COPYBITS@, @PORTBITS@));
        }
        count += 1;
    }
    count
}

`, "@FUNC@", funcName, "@CORPUS@", w.corpus, "@PUBFROM@", w.pubFrom("v0"), "@MOD@", module, "@FN@", fn, "@PORTV0@", w.portArg("v0"), "@COPYBITS@", w.pubBitsExpr("b_copy"), "@PORTBITS@", w.portBitsExpr("pr"), "@OPFMT@", w.opFmtSpec(), "@FMTSPEC@", w.resultFmtSpec(), "@SYM@", row.GoSymbol))
	return funcName, cases, nil
}

func emitDisplay(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, w parityWidth) (string, int, error) {
	nanModule, nanFn, err := resolvePort(w.bidgoIsNaN, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	// Decimal32's Bid32IsNaN already returns bool; see emitSimpleOp's
	// resultBool case for the same distinction.
	nanExpr := fmt.Sprintf("bid754::generated::%s::%s(%s)", nanModule, nanFn, w.portArg("v0"))
	if !apiemit.PortBoolResult(w.bidgoIsNaN) {
		nanExpr += " != 0"
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := w.nonNaNLen(corpus)
	b.WriteString(rustTmpl(`fn @FUNC@(failures: &mut Vec<String>) -> usize {
    let mut count = 0usize;
    for &v0 in @CORPUS@ {
        // NaN entries take the hand-built display branch (which cannot be
        // compared against a raw port to_string call), so they are skipped
        // without counting, matching the Go leg's vm_string shape. Their
        // payload/sign display + round-trip is covered for every width by
        // generated_public_api_nan_roundtrip_<w> below (NOT by the
        // Decimal64-only smoke test).
        if @NANEXPR@ {
            continue;
        }
        let pv = @PUBFROM@.to_string();
        let pr = bid754::generated::@MOD@::@FN@(@PORTV0@);
        if pv != pr {
            failures.push(format!("public parity @SYM@: operand @OPFMT@: result mismatch public={:?} port={:?}", v0, pv, pr));
        }
        count += 1;
    }
    count
}

`, "@FUNC@", funcName, "@CORPUS@", w.corpus, "@PUBFROM@", w.pubFrom("v0"), "@NANEXPR@", nanExpr, "@MOD@", module, "@FN@", fn, "@PORTV0@", w.portArg("v0"), "@OPFMT@", w.opFmtSpec(), "@SYM@", row.GoSymbol))
	return funcName, cases, nil
}

// ---- string-shape (parse family) ----

// resolveFromStringRawPort resolves the width's own independent raw
// from-string comparison call (as a full Rust call expression over inputExpr)
// plus the independent NaN-check expression the parse family needs.
// fromStringBidgoFn identifies which port the caller wants.
//
// Decimal64's generated::bid64_from_string::bid64_from_string is pub(crate),
// so this external integration test uses the existing #[doc(hidden)] public
// bid64_from_string_raw entrypoint. The generated Decimal32 and Decimal128 raw
// functions are public and are called directly. This branch follows those
// declared visibility boundaries explicitly.
func resolveFromStringRawPort(fromStringBidgoFn, goSymbol string, w parityWidth, inputExpr string) (fromStringCall, nanExpr string, err error) {
	nanModule, nanFn, err := resolvePort(w.bidgoIsNaN, goSymbol)
	if err != nil {
		return "", "", err
	}
	nanExpr = fmt.Sprintf("bid754::generated::%s::%s(pr)", nanModule, nanFn)
	if !apiemit.PortBoolResult(w.bidgoIsNaN) {
		nanExpr += " != 0"
	}

	if w.selfType == "Decimal64" {
		return fmt.Sprintf("bid754::bid64_from_string_raw(%s, 0)", inputExpr), nanExpr, nil
	}
	module, fn, err := resolvePort(fromStringBidgoFn, goSymbol)
	if err != nil {
		return "", "", err
	}
	fromStringCall = fmt.Sprintf("bid754::generated::%s::%s(%s, BIDGO_ROUND_NEAREST_EVEN)", module, fn, inputExpr)
	return fromStringCall, nanExpr, nil
}

// resolveFromStringModePort is resolveFromStringRawPort with the rounding
// argument taken from the caller's expression (the parse_mode shape's
// port_mode loop variable) instead of the fixed nearest-even literal, keeping
// the same width-specific entrypoint selection (Decimal64's re-exported raw
// from-string helper vs the generated from-string port).
func resolveFromStringModePort(fromStringBidgoFn, goSymbol string, w parityWidth, inputExpr, modeExpr string) (fromStringCall, nanExpr string, err error) {
	nanModule, nanFn, err := resolvePort(w.bidgoIsNaN, goSymbol)
	if err != nil {
		return "", "", err
	}
	nanExpr = fmt.Sprintf("bid754::generated::%s::%s(pr)", nanModule, nanFn)
	if !apiemit.PortBoolResult(w.bidgoIsNaN) {
		nanExpr += " != 0"
	}

	if w.selfType == "Decimal64" {
		// bid64_from_string_raw's rounding parameter is i32 (the raw-port
		// re-export), unlike the generated from-string ports' i64.
		return fmt.Sprintf("bid754::bid64_from_string_raw(%s, %s as i32)", inputExpr, modeExpr), nanExpr, nil
	}
	module, fn, err := resolvePort(fromStringBidgoFn, goSymbol)
	if err != nil {
		return "", "", err
	}
	fromStringCall = fmt.Sprintf("bid754::generated::%s::%s(%s, %s)", module, fn, inputExpr, modeExpr)
	return fromStringCall, nanExpr, nil
}

func emitParse(b *strings.Builder, row rustParityInventoryRow, w parityWidth) (string, int, error) {
	fromStringCall, nanExpr, err := resolveFromStringRawPort(row.BidgoFunction, row.GoSymbol, w, "sc.input")
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := len(publicParityStringCorpusCases)
	b.WriteString(rustTmpl(`fn @FUNC@(failures: &mut Vec<String>) -> usize {
    let mut count = 0usize;
    for sc in STRING_CASES {
        let got = @SELF@::parse(sc.input);
        match sc.kind {
            "nan_literal" => match (sc.nan_min_width != 0 && sc.nan_min_width <= @WIDTH@, got) {
                (true, Ok(pv)) => {
                    if !pv.is_nan() {
                        failures.push(format!("public parity @SYM@: input {:?}: expected a NaN result from the NaN literal branch", sc.input));
                    }
                    if pv.is_signaling() != sc.signaling {
                        failures.push(format!("public parity @SYM@: input {:?}: signaling bit mismatch public={} literal={}", sc.input, pv.is_signaling(), sc.signaling));
                    }
                }
                (true, Err(e)) => failures.push(format!("public parity @SYM@: input {:?}: unexpected error {}", sc.input, e)),
                (false, Ok(pv)) => failures.push(format!("public parity @SYM@: input {:?}: unrepresentable NaN payload returned Ok(@FMTSPEC@)", sc.input, @PVBITS@)),
                (false, Err(e)) => {
                    if e.input() != sc.input {
                        failures.push(format!("public parity @SYM@: input {:?}: ParseDecimalError retained {:?}", sc.input, e.input()));
                    }
                }
            },
            "blank" | "invalid_syntax" => match got {
                Ok(pv) => failures.push(format!("public parity @SYM@: input {:?}: rejected input kind {:?} returned Ok(@FMTSPEC@)", sc.input, sc.kind, @PVBITS@)),
                Err(e) => {
                    if e.input() != sc.input {
                        failures.push(format!("public parity @SYM@: input {:?}: ParseDecimalError retained {:?}", sc.input, e.input()));
                    }
                }
            },
            _ => {
                let (pr, praw) = @FROMSTR@;
                let pr_is_nan = @NANEXPR@;
                let port_flags = map_port_flags(praw);
                let silent_cohort_coercion = sc.kind == "cohort_coercion"
                    && (sc.cohort_min_width == 0 || sc.cohort_min_width > @WIDTH@)
                    && praw == 0;
                let loss_mask = bid754::ExceptionFlags::INVALID_OPERATION.bits()
                    | bid754::ExceptionFlags::OVERFLOW.bits()
                    | bid754::ExceptionFlags::UNDERFLOW.bits()
                    | bid754::ExceptionFlags::INEXACT.bits();
                let should_error = pr_is_nan || port_flags & loss_mask != 0 || silent_cohort_coercion;
                match got {
                    Ok(pv) if should_error => {
                        failures.push(format!("public parity @SYM@: input {:?}: expected an exact-representation error, got Ok(@FMTSPEC@) with port flags {:#x}", sc.input, @PVBITS@, port_flags));
                    }
                    Ok(pv) => {
                        if @PVBITS@ != @PORTBITS@ {
                            failures.push(format!("public parity @SYM@: input {:?}: result mismatch public=@FMTSPEC@ port=@FMTSPEC@", sc.input, @PVBITS@, @PORTBITS@));
                        }
                    }
                    Err(e) if !should_error => {
                        failures.push(format!("public parity @SYM@: input {:?}: unexpected error {} for port result @FMTSPEC@", sc.input, e, @PORTBITS@));
                    }
                    Err(e) => {
                        if e.input() != sc.input {
                            failures.push(format!("public parity @SYM@: input {:?}: ParseDecimalError retained {:?}", sc.input, e.input()));
                        }
                    }
                }
            }
        }
        count += 1;
    }
    count
}

`, "@FUNC@", funcName, "@SELF@", w.selfType, "@WIDTH@", strconv.Itoa(parityWidthDigits(w)), "@FROMSTR@", fromStringCall, "@NANEXPR@", nanExpr, "@PVBITS@", w.pubBitsExpr("pv"), "@PORTBITS@", w.portBitsExpr("pr"), "@FMTSPEC@", w.resultFmtSpec(), "@SYM@", row.GoSymbol))
	return funcName, cases, nil
}

func emitParseRaw(b *strings.Builder, row rustParityInventoryRow, w parityWidth) (string, int, error) {
	fromStringCall, _, err := resolveFromStringRawPort(row.BidgoFunction, row.GoSymbol, w, "sc.input")
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := len(publicParityStringCorpusCases)
	b.WriteString(rustTmpl(`fn @FUNC@(failures: &mut Vec<String>) -> usize {
    let mut count = 0usize;
    for sc in STRING_CASES {
        let (pv, pf) = @SELF@::parse_raw(sc.input);
        if sc.kind == "nan_literal" {
            if sc.nan_min_width != 0 && sc.nan_min_width <= @WIDTH@ {
                if !pv.is_nan() {
                    failures.push(format!("public parity @SYM@: input {:?}: expected a NaN result from the NaN literal branch", sc.input));
                }
                if pv.is_signaling() != sc.signaling {
                    failures.push(format!("public parity @SYM@: input {:?}: signaling bit mismatch public={} literal={}", sc.input, pv.is_signaling(), sc.signaling));
                }
                if !pf.is_empty() {
                    failures.push(format!("public parity @SYM@: input {:?}: representable NaN literal must raise no flags, got {:?}", sc.input, pf));
                }
            } else {
                if @PVBITS@ != @QNANBITS@ {
                    failures.push(format!("public parity @SYM@: input {:?}: invalid-payload result bits @FMTSPEC@, want canonical qNaN @FMTSPEC@", sc.input, @PVBITS@, @QNANBITS@));
                }
                if pf.bits() != bid754::ExceptionFlags::INVALID_OPERATION.bits() {
                    failures.push(format!("public parity @SYM@: input {:?}: invalid-payload flags={:#x}, want InvalidOperation", sc.input, pf.bits()));
                }
            }
        } else {
            let (pr, praw) = @FROMSTR@;
            let silent_cohort_coercion = sc.kind == "cohort_coercion"
                && (sc.cohort_min_width == 0 || sc.cohort_min_width > @WIDTH@)
                && praw == 0;
            let malformed_input = sc.kind == "blank" || sc.kind == "invalid_syntax";
            if malformed_input || silent_cohort_coercion {
                if @PVBITS@ != @QNANBITS@ {
                    failures.push(format!("public parity @SYM@: input {:?}: rejected-input result bits @FMTSPEC@, want canonical qNaN @FMTSPEC@", sc.input, @PVBITS@, @QNANBITS@));
                }
                if pf.bits() != bid754::ExceptionFlags::INVALID_OPERATION.bits() {
                    failures.push(format!("public parity @SYM@: input {:?}: rejected-input flags={:#x}, want InvalidOperation", sc.input, pf.bits()));
                }
            } else {
                if @PVBITS@ != @PORTBITS@ {
                    failures.push(format!("public parity @SYM@: input {:?}: result mismatch public=@FMTSPEC@ port=@FMTSPEC@", sc.input, @PVBITS@, @PORTBITS@));
                }
                if pf.bits() != map_port_flags(praw) {
                    failures.push(format!("public parity @SYM@: input {:?}: flag mismatch public={:#x} port={:#x}", sc.input, pf.bits(), map_port_flags(praw)));
                }
            }
        }
        count += 1;
    }
    count
}

`, "@FUNC@", funcName, "@SELF@", w.selfType, "@WIDTH@", strconv.Itoa(parityWidthDigits(w)), "@FROMSTR@", fromStringCall, "@PVBITS@", w.pubBitsExpr("pv"), "@PORTBITS@", w.portBitsExpr("pr"), "@QNANBITS@", rustCanonicalQNaNBits(w), "@FMTSPEC@", w.resultFmtSpec(), "@SYM@", row.GoSymbol))
	return funcName, cases, nil
}

func emitParseWithFlags(b *strings.Builder, row rustParityInventoryRow, w parityWidth) (string, int, error) {
	fromStringCall, nanExpr, err := resolveFromStringRawPort(row.BidgoFunction, row.GoSymbol, w, "sc.input")
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := len(publicParityStringCorpusCases)
	b.WriteString(rustTmpl(`fn @FUNC@(failures: &mut Vec<String>) -> usize {
    let mut count = 0usize;
    for sc in STRING_CASES {
        let got = @SELF@::parse_with_flags(sc.input);
        match sc.kind {
            "nan_literal" => match (sc.nan_min_width != 0 && sc.nan_min_width <= @WIDTH@, got) {
                (true, Ok((pv, pf))) => {
                    if !pv.is_nan() {
                        failures.push(format!("public parity @SYM@: input {:?}: expected a NaN result from the NaN literal branch", sc.input));
                    }
                    if pv.is_signaling() != sc.signaling {
                        failures.push(format!("public parity @SYM@: input {:?}: signaling bit mismatch public={} literal={}", sc.input, pv.is_signaling(), sc.signaling));
                    }
                    if !pf.is_empty() {
                        failures.push(format!("public parity @SYM@: input {:?}: NaN literal branch must raise no flags, got {:?}", sc.input, pf));
                    }
                }
                (true, Err(e)) => failures.push(format!("public parity @SYM@: input {:?}: unexpected error {}", sc.input, e)),
                (false, Ok((pv, pf))) => failures.push(format!("public parity @SYM@: input {:?}: unrepresentable NaN payload returned Ok((@FMTSPEC@, {:#x}))", sc.input, @PVBITS@, pf.bits())),
                (false, Err(e)) => {
                    if e.input() != sc.input {
                        failures.push(format!("public parity @SYM@: input {:?}: ParseDecimalError retained {:?}", sc.input, e.input()));
                    }
                }
            },
            "blank" | "invalid_syntax" => match got {
                Ok((pv, pf)) => failures.push(format!("public parity @SYM@: input {:?}: rejected input kind {:?} returned Ok((@FMTSPEC@, {:#x}))", sc.input, sc.kind, @PVBITS@, pf.bits())),
                Err(e) => {
                    if e.input() != sc.input {
                        failures.push(format!("public parity @SYM@: input {:?}: ParseDecimalError retained {:?}", sc.input, e.input()));
                    }
                }
            },
            _ => {
                let (pr, praw) = @FROMSTR@;
                let pr_is_nan = @NANEXPR@;
                let port_flags = map_port_flags(praw);
                let silent_cohort_coercion = sc.kind == "cohort_coercion"
                    && (sc.cohort_min_width == 0 || sc.cohort_min_width > @WIDTH@)
                    && praw == 0;
                let should_error = pr_is_nan
                    || port_flags & bid754::ExceptionFlags::INVALID_OPERATION.bits() != 0
                    || silent_cohort_coercion;
                match got {
                    Ok((pv, pf)) if should_error => {
                        failures.push(format!("public parity @SYM@: input {:?}: expected an error, got Ok((@FMTSPEC@, {:#x}))", sc.input, @PVBITS@, pf.bits()));
                    }
                    Ok((pv, pf)) => {
                        if @PVBITS@ != @PORTBITS@ {
                            failures.push(format!("public parity @SYM@: input {:?}: result mismatch public=@FMTSPEC@ port=@FMTSPEC@", sc.input, @PVBITS@, @PORTBITS@));
                        }
                        if pf.bits() != port_flags {
                            failures.push(format!("public parity @SYM@: input {:?}: flag mismatch public={:#x} port={:#x}", sc.input, pf.bits(), port_flags));
                        }
                    }
                    Err(e) if !should_error => {
                        failures.push(format!("public parity @SYM@: input {:?}: unexpected error {} for port result @FMTSPEC@", sc.input, e, @PORTBITS@));
                    }
                    Err(e) => {
                        if e.input() != sc.input {
                            failures.push(format!("public parity @SYM@: input {:?}: ParseDecimalError retained {:?}", sc.input, e.input()));
                        }
                    }
                }
            }
        }
        count += 1;
    }
    count
}

`, "@FUNC@", funcName, "@SELF@", w.selfType, "@WIDTH@", strconv.Itoa(parityWidthDigits(w)), "@FROMSTR@", fromStringCall, "@NANEXPR@", nanExpr, "@PVBITS@", w.pubBitsExpr("pv"), "@PORTBITS@", w.portBitsExpr("pr"), "@FMTSPEC@", w.resultFmtSpec(), "@SYM@", row.GoSymbol))
	return funcName, cases, nil
}

// emitFromIntExactOrError handles NewDecimal<w>FromInt's direct mechanical-
// port route. Decimal32/64 reject a port Inexact result; Decimal128's int64
// domain is exact by construction. intCorpus follows the Go parameter width.
func emitFromIntExactOrError(b *strings.Builder, row rustParityInventoryRow, w parityWidth, intCorpus string) (string, int, error) {
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	var portSetup string
	switch row.BidgoFunction {
	case "Bid32FromInt32", "Bid64FromInt64":
		portSetup = fmt.Sprintf("let (pr, praw) = bid754::generated::%s::%s(x, BIDGO_ROUND_NEAREST_EVEN);\n        let should_error = map_port_flags(praw) & bid754::ExceptionFlags::INEXACT.bits() != 0;", module, fn)
	case "Bid128FromInt64":
		portSetup = fmt.Sprintf("let pr = bid754::generated::%s::%s(x);\n        let should_error = false;", module, fn)
	default:
		return "", 0, fmt.Errorf("rust public parity: exact-or-error integer shape has unsupported port %q for %s", row.BidgoFunction, row.GoSymbol)
	}
	cases := len(publicParityIntCorpus64)
	if intCorpus == "INT_CORPUS_32" {
		cases = len(publicParityIntCorpus32)
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	b.WriteString(rustTmpl(`fn @FUNC@(failures: &mut Vec<String>) -> usize {
    let mut count = 0usize;
    for &x in @INTCORPUS@ {
        let got = @SELF@::@SURF@(x);
        let text = x.to_string();
        @PORTSETUP@
        match got {
            Ok(pv) if should_error => {
                failures.push(format!("public parity @SYM@: operand {}: expected an exact-representation error, got @FMTSPEC@", x, @PVBITS@));
            }
            Ok(pv) => {
                if @PVBITS@ != @PORTBITS@ {
                    failures.push(format!("public parity @SYM@: operand {}: result mismatch public=@FMTSPEC@ port=@FMTSPEC@", x, @PVBITS@, @PORTBITS@));
                }
            }
            Err(e) if should_error => {
                let want_message = format!("integer {} is not exactly representable as @SELF@", x);
                if e.input() != text || e.target() != "@SELF@" || e.to_string() != want_message {
                    failures.push(format!("public parity @SYM@: operand {}: error mismatch input={:?} target={:?} message={:?}", x, e.input(), e.target(), e.to_string()));
                }
            }
            Err(e) => failures.push(format!("public parity @SYM@: operand {}: unexpected error {}", x, e)),
        }
        count += 1;
    }
    count
}

`, "@FUNC@", funcName, "@INTCORPUS@", intCorpus, "@SELF@", w.selfType, "@SURF@", row.RustSurface, "@PORTSETUP@", portSetup, "@PVBITS@", w.pubBitsExpr("pv"), "@PORTBITS@", w.portBitsExpr("pr"), "@FMTSPEC@", w.resultFmtSpec(), "@SYM@", row.GoSymbol))
	return funcName, cases, nil
}

// emitFromExact handles the exact-widening integer constructors that fold
// into an impl From<T>: Decimal64's from_i32_exact/from_u32_exact and
// Decimal128's from_i32_exact/from_u32_exact/from_i64_exact/from_u64_exact
// (BID128's 34 digits represent every int64/uint64 exactly, so it gains the
// two 64-bit From impls Decimal64 cannot). The port function is always
// tuple-less (a bare Bid<w>FromInt<N> value return, no flags for an exact
// conversion), so no pfpsf branch is needed; only the result wrap forks by
// width.
func emitFromExact(b *strings.Builder, row rustParityInventoryRow, w parityWidth) (string, int, error) {
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	var corpusName string
	switch row.Shape {
	case "from_i32_exact":
		corpusName = "INT_CORPUS_32"
	case "from_u32_exact":
		corpusName = "UINT_CORPUS_32"
	case "from_i64_exact":
		corpusName = "INT_CORPUS_64"
	case "from_u64_exact":
		corpusName = "UINT_CORPUS_64"
	default:
		return "", 0, fmt.Errorf("rust public parity: emitFromExact: unsupported shape %q for go_symbol %q", row.Shape, row.GoSymbol)
	}
	cases := 0
	switch corpusName {
	case "INT_CORPUS_32":
		cases = len(publicParityIntCorpus32)
	case "UINT_CORPUS_32":
		cases = len(publicParityUintCorpus32)
	case "INT_CORPUS_64":
		cases = len(publicParityIntCorpus64)
	case "UINT_CORPUS_64":
		cases = len(publicParityUintCorpus64)
	}
	b.WriteString(rustTmpl(`fn @FUNC@(failures: &mut Vec<String>) -> usize {
    let mut count = 0usize;
    for &x in @CORPUS@ {
        let pv: @SELF@ = @SELF@::from(x);
        let pr = bid754::generated::@MOD@::@FN@(x);
        if @PVBITS@ != @PORTBITS@ {
            failures.push(format!("public parity @SYM@: operand {}: result mismatch public=@FMTSPEC@ port=@FMTSPEC@", x, @PVBITS@, @PORTBITS@));
        }
        count += 1;
    }
    count
}

`, "@FUNC@", funcName, "@CORPUS@", corpusName, "@SELF@", w.selfType, "@MOD@", module, "@FN@", fn, "@PVBITS@", w.pubBitsExpr("pv"), "@PORTBITS@", w.portBitsExpr("pr"), "@FMTSPEC@", w.resultFmtSpec(), "@SYM@", row.GoSymbol))
	return funcName, cases, nil
}

func emitFromMode(b *strings.Builder, row rustParityInventoryRow, w parityWidth) (string, int, error) {
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	var corpusName string
	switch row.Shape {
	case "from_i64_mode":
		corpusName = "INT_CORPUS_64"
	case "from_u64_mode":
		corpusName = "UINT_CORPUS_64"
	case "from_i32_mode":
		corpusName = "INT_CORPUS_32"
	case "from_u32_mode":
		corpusName = "UINT_CORPUS_32"
	default:
		return "", 0, fmt.Errorf("rust public parity: emitFromMode: unsupported shape %q for go_symbol %q", row.Shape, row.GoSymbol)
	}
	var baseLen int
	switch corpusName {
	case "INT_CORPUS_64":
		baseLen = len(publicParityIntCorpus64)
	case "UINT_CORPUS_64":
		baseLen = len(publicParityUintCorpus64)
	case "INT_CORPUS_32":
		baseLen = len(publicParityIntCorpus32)
	case "UINT_CORPUS_32":
		baseLen = len(publicParityUintCorpus32)
	}
	cases := baseLen * len(publicParityModeOrderNames)
	// from_i64_mode/from_u64_mode are Decimal64-only and from_i32_mode/
	// from_u32_mode Decimal32-only (Decimal128 uses the exact From impls
	// instead), so this shape's port is always tuple-return; the width
	// helpers still parameterize the result wrap so a future 128 use would
	// be correct rather than silently wrong.
	b.WriteString(rustTmpl(`fn @FUNC@(failures: &mut Vec<String>) -> usize {
    let mut count = 0usize;
    for &x in @CORPUS@ {
        for &(mode, port_mode) in PARITY_MODES {
            let (pv, pf) = @SELF@::@SURF@(x, mode);
            let (pr, praw) = bid754::generated::@MOD@::@FN@(x, port_mode);
            if @PVBITS@ != @PORTBITS@ {
                failures.push(format!("public parity @SYM@: operand {} mode {:?}: result mismatch public=@FMTSPEC@ port=@FMTSPEC@", x, mode, @PVBITS@, @PORTBITS@));
            }
            if pf.bits() != map_port_flags(praw) {
                failures.push(format!("public parity @SYM@: operand {} mode {:?}: flag mismatch public={:#x} port={:#x}", x, mode, pf.bits(), map_port_flags(praw)));
            }
            count += 1;
        }
    }
    count
}

`, "@FUNC@", funcName, "@CORPUS@", corpusName, "@SELF@", w.selfType, "@SURF@", row.RustSurface, "@MOD@", module, "@FN@", fn, "@PVBITS@", w.pubBitsExpr("pv"), "@PORTBITS@", w.portBitsExpr("pr"), "@FMTSPEC@", w.resultFmtSpec(), "@SYM@", row.GoSymbol))
	return funcName, cases, nil
}

func emitContextBinaryWithFlags(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, w parityWidth) (string, int, error) {
	module, fn, err := resolvePort(row.BidgoFunction, row.GoSymbol)
	if err != nil {
		return "", 0, err
	}
	funcName := normalizeRustFnName(row.GoSymbol)
	cases := w.pairsLen(corpus) * len(publicParityModeOrderNames)
	// Add128BIDWithContext's port (Bid128Add) uses the pfpsf convention;
	// portCallStmt handles that the same way every other
	// pfpsf op does.
	portStmt := portCallStmt(row.BidgoFunction, module, fn, []string{w.portArg("v0"), w.portArg("v1"), "port_mode"})
	b.WriteString(rustTmpl(`fn @FUNC@(failures: &mut Vec<String>) -> usize {
    let mut count = 0usize;
    for &(i0, i1) in @PAIRS@ {
        let v0 = @CORPUS@[i0];
        let v1 = @CORPUS@[i1];
        for &(mode, port_mode) in PARITY_MODES {
            let mut ctx = Context::with_rounding(mode);
            let pv = ctx.@SURF@(@PUBFROM0@, @PUBFROM1@);
            @PORTSTMT@
            if @PVBITS@ != @PORTBITS@ {
                failures.push(format!("public parity @SYM@: operands @OPFMT@,@OPFMT@ mode {:?}: result mismatch public=@FMTSPEC@ port=@FMTSPEC@", v0, v1, mode, @PVBITS@, @PORTBITS@));
            }
            if ctx.flags.bits() != map_port_flags(praw) {
                failures.push(format!("public parity @SYM@: operands @OPFMT@,@OPFMT@ mode {:?}: flag mismatch public={:#x} port={:#x}", v0, v1, mode, ctx.flags.bits(), map_port_flags(praw)));
            }
            count += 1;
        }
    }
    count
}

`, "@FUNC@", funcName, "@PAIRS@", w.pairs, "@CORPUS@", w.corpus, "@PUBFROM0@", w.pubFrom("v0"), "@PUBFROM1@", w.pubFrom("v1"), "@SURF@", row.RustSurface, "@PORTSTMT@", portStmt, "@PVBITS@", w.pubBitsExpr("pv"), "@PORTBITS@", w.portBitsExpr("pr"), "@OPFMT@", w.opFmtSpec(), "@FMTSPEC@", w.resultFmtSpec(), "@SYM@", row.GoSymbol))
	return funcName, cases, nil
}

// ---- to_iN / to_uN (+ Exact) mode-dispatch family ----

var convertRoundingSuffixes = []string{"Rnint", "Rninta", "Int", "Ceil", "Floor"}
var convertExactRoundingSuffixes = []string{"Xrnint", "Xrninta", "Xint", "Xceil", "Xfloor"}

func emitConvertInt(b *strings.Builder, row rustParityInventoryRow, corpus publicParityCorpus, w parityWidth) (string, int, error) {
	exact := strings.HasSuffix(row.Shape, "_exact")
	stem := strings.TrimSuffix(strings.TrimPrefix(row.Shape, "to_"), "_exact")
	rustType := stem
	var base string
	var suffixes []string
	if exact {
		base = strings.TrimSuffix(row.BidgoFunction, "Xrnint")
		suffixes = convertExactRoundingSuffixes
	} else {
		base = strings.TrimSuffix(row.BidgoFunction, "Rnint")
		suffixes = convertRoundingSuffixes
	}
	type variant struct{ module, fn string }
	variants := make([]variant, len(suffixes))
	for i, suf := range suffixes {
		m, f, err := resolvePort(base+suf, row.GoSymbol)
		if err != nil {
			return "", 0, err
		}
		variants[i] = variant{m, f}
	}

	funcName := normalizeRustFnName(row.GoSymbol)
	cases := w.bitsLen(corpus) * len(publicParityModeOrderNames)
	// Every ConvertToInt<N>/ConvertToUint<N> dispatch leaf is tuple-return
	// for every width including 128 (verified against the generated::*
	// signatures -- these converters never use the pfpsf convention), so the
	// (pr, praw) tuple destructure is valid for all widths; only the operand
	// port arg forks by width (via w.portArg). The int result pr is a plain
	// iN/uN comparable directly for every width.
	portV0 := w.portArg("v0")

	fmt.Fprintf(b, "fn %s(failures: &mut Vec<String>) -> usize {\n", funcName)
	fmt.Fprintf(b, "    let mut count = 0usize;\n")
	fmt.Fprintf(b, "    for &v0 in %s {\n", w.corpus)
	fmt.Fprintf(b, "        for &(mode, _port_mode) in PARITY_MODES {\n")
	fmt.Fprintf(b, "            let (pv, pf) = %s.%s(mode);\n", w.pubFrom("v0"), row.RustSurface)
	fmt.Fprintf(b, "            let (pr, praw): (%s, u32) = match mode {\n", rustType)
	for i, m := range publicParityModeOrderNames {
		fmt.Fprintf(b, "                RoundingMode::%s => bid754::generated::%s::%s(%s),\n", m, variants[i].module, variants[i].fn, portV0)
	}
	fmt.Fprintf(b, "            };\n")
	fmt.Fprintf(b, "            if pv != pr {\n")
	fmt.Fprintf(b, "                failures.push(format!(\"public parity %s: operand %s mode {:?}: result mismatch public={} port={}\", v0, mode, pv, pr));\n", row.GoSymbol, w.opFmtSpec())
	fmt.Fprintf(b, "            }\n")
	fmt.Fprintf(b, "            if pf.bits() != map_port_flags(praw) {\n")
	fmt.Fprintf(b, "                failures.push(format!(\"public parity %s: operand %s mode {:?}: flag mismatch public={:#x} port={:#x}\", v0, mode, pf.bits(), map_port_flags(praw)));\n", row.GoSymbol, w.opFmtSpec())
	fmt.Fprintf(b, "            }\n")
	fmt.Fprintf(b, "            count += 1;\n")
	fmt.Fprintf(b, "        }\n")
	fmt.Fprintf(b, "    }\n")
	fmt.Fprintf(b, "    count\n}\n\n")
	return funcName, cases, nil
}

// ---- driver ----

func emitRustParityDriver(b *strings.Builder, summaries []rustParityCaseSummary) {
	sort.Slice(summaries, func(i, j int) bool { return summaries[i].GoSymbol < summaries[j].GoSymbol })

	total := 0
	byShape := map[string]int{}
	for _, s := range summaries {
		total += s.Cases
		byShape[s.Shape] += s.Cases
	}
	shapeNames := make([]string, 0, len(byShape))
	for name := range byShape {
		shapeNames = append(shapeNames, name)
	}
	sort.Strings(shapeNames)

	b.WriteString(`/// One driver-table row: the verified go_symbol (for diagnostics), the
/// apiemit shape tag (for the by-shape case-count accounting), and the
/// generated per-row check function.
struct ParityUnit {
    go_symbol: &'static str,
    shape: &'static str,
    run: fn(&mut Vec<String>) -> usize,
}

`)
	b.WriteString("const PARITY_UNITS: &[ParityUnit] = &[\n")
	for _, s := range summaries {
		fmt.Fprintf(b, "    ParityUnit { go_symbol: %q, shape: %q, run: %s },\n", s.GoSymbol, s.Shape, s.FuncName)
	}
	b.WriteString("];\n\n")

	b.WriteString(`/// The public-API parity gate exercises every emitted public wrapper against
/// an independent invocation of its pinned crate::generated::* port function
/// on the shared deterministic bit-literal corpus, comparing result bits and
/// mapped exception flags. It is an architecture-contract gate (does the
/// wrapper route through the port and preserve semantics?), not a fifth
/// regular verification domain (same framing as the Go leg). Case counts are
/// pinned here at generation time so a generator regression that shrinks the
/// corpus cannot silently re-pin a smaller surface.
`)
	fmt.Fprintf(b, "pub(crate) const EXPECTED_PARITY_WRAPPERS: usize = %d;\n", len(summaries))
	fmt.Fprintf(b, "pub(crate) const EXPECTED_PARITY_CASES: usize = %d;\n\n", total)
	fmt.Fprintf(b, "const EXPECTED_MIXED_FMA_FUSEDNESS_SENTINELS: usize = %d;\n\n", len(ffiMixedFMAFusednessProbes))

	b.WriteString("const EXPECTED_PARITY_CASES_BY_SHAPE: &[(&str, usize)] = &[\n")
	for _, name := range shapeNames {
		fmt.Fprintf(b, "    (%q, %d),\n", name, byShape[name])
	}
	b.WriteString("];\n\n")

	b.WriteString(`#[test]
fn generated_public_api_parity() {
    assert_eq!(
        MIXED_FMA_FUSEDNESS_SENTINEL_ROWS.len(),
        EXPECTED_MIXED_FMA_FUSEDNESS_SENTINELS,
        "mixed FMA fusedness sentinel census drifted"
    );
    assert_eq!(
        PARITY_UNITS.len(),
        EXPECTED_PARITY_WRAPPERS,
        "parity wrapper count drifted"
    );
    let mut failures: Vec<String> = Vec::new();
    let mut total = 0usize;
    let mut by_shape: std::collections::HashMap<&str, usize> = std::collections::HashMap::new();
    for unit in PARITY_UNITS {
        let n = (unit.run)(&mut failures);
        if n == 0 {
            failures.push(format!("public parity wrapper {:?} ran zero cases", unit.go_symbol));
        }
        total += n;
        *by_shape.entry(unit.shape).or_insert(0) += n;
    }
    if total != EXPECTED_PARITY_CASES {
        failures.push(format!(
            "expected {} parity cases, ran {}",
            EXPECTED_PARITY_CASES, total
        ));
    }
    for &(shape, want) in EXPECTED_PARITY_CASES_BY_SHAPE {
        let got = by_shape.get(shape).copied().unwrap_or(0);
        if got != want {
            failures.push(format!(
                "shape {:?}: expected {} cases, ran {}",
                shape, want, got
            ));
        }
    }
    if by_shape.len() != EXPECTED_PARITY_CASES_BY_SHAPE.len() {
        failures.push(format!(
            "ran {} shapes, pinned {}",
            by_shape.len(),
            EXPECTED_PARITY_CASES_BY_SHAPE.len()
        ));
    }
    assert!(
        failures.is_empty(),
        "public API parity failures ({} total):\n{}",
        failures.len(),
        failures.join("\n")
    );
}

`)
}

// emitRustTraitParityTest verifies the idiomatic PartialEq/PartialOrd trait
// impls (bid754-rs/src/generated/api/decimal<w>.rs) against a direct
// composition of the already-verified quiet_eq/quiet_lt/quiet_gt port calls,
// for width w. These traits are not their own rust_api_surface_inventory.json
// row (they are emergent Rust idiom built from three already-emitted,
// already-parity-tested wrapper methods), so this extends beyond the
// per-width census population rather than folding into
// EXPECTED_PARITY_WRAPPERS/EXPECTED_PARITY_CASES, keeping that anchor
// exactly traceable to the inventory's emitted row count. Called once per width,
// each with its own width-suffixed constant/test name so
// a width-32 trait-parity failure cannot be confused with width-64's.
func emitRustTraitParityTest(b *strings.Builder, corpus publicParityCorpus, w parityWidth) error {
	eqModule, eqFn, err := resolvePort(w.bidgoQuietEqual, "PartialEq/PartialOrd trait parity ("+w.selfType+")")
	if err != nil {
		return err
	}
	ltModule, ltFn, err := resolvePort(w.bidgoQuietLess, "PartialEq/PartialOrd trait parity ("+w.selfType+")")
	if err != nil {
		return err
	}
	gtModule, gtFn, err := resolvePort(w.bidgoQuietGreat, "PartialEq/PartialOrd trait parity ("+w.selfType+")")
	if err != nil {
		return err
	}
	pairs := w.pairsLen(corpus)
	b.WriteString(rustTmpl(`/// PartialEq/PartialOrd trait parity verifies the idiomatic trait
/// impls against a direct composition of the quiet_equal/quiet_less/
/// quiet_greater port calls, independently of the trait impls' own delegation
/// to @SELF@::quiet_eq/quiet_lt/quiet_gt. Flags are dropped by design (the
/// traits carry no flag channel), matching the Go leg's flag-less quiet-*
/// usage in this position.
const EXPECTED_TRAIT_PARITY_CASES_@DIGITS@: usize = @CASES@;

#[test]
fn generated_public_api_trait_parity_@DIGITS@() {
    let mut failures: Vec<String> = Vec::new();
    let mut count = 0usize;
    for &(i0, i1) in @PAIRS@ {
        let v0 = @CORPUS@[i0];
        let v1 = @CORPUS@[i1];
        let a = @PUBFROM0@;
        let b = @PUBFROM1@;
        let (eq_port, _) = bid754::generated::@EQMOD@::@EQFN@(@PORTV0@, @PORTV1@);
        let (lt_port, _) = bid754::generated::@LTMOD@::@LTFN@(@PORTV0@, @PORTV1@);
        let (gt_port, _) = bid754::generated::@GTMOD@::@GTFN@(@PORTV0@, @PORTV1@);
        if (a == b) != (eq_port != 0) {
            failures.push(format!("public parity PartialEq: operands @OPFMT@,@OPFMT@: public={} port={}", v0, v1, a == b, eq_port != 0));
        }
        let want_ord = if lt_port != 0 {
            Some(core::cmp::Ordering::Less)
        } else if gt_port != 0 {
            Some(core::cmp::Ordering::Greater)
        } else if eq_port != 0 {
            Some(core::cmp::Ordering::Equal)
        } else {
            None
        };
        if a.partial_cmp(&b) != want_ord {
            failures.push(format!("public parity PartialOrd: operands @OPFMT@,@OPFMT@: public={:?} port={:?}", v0, v1, a.partial_cmp(&b), want_ord));
        }
        count += 1;
    }
    assert_eq!(count, EXPECTED_TRAIT_PARITY_CASES_@DIGITS@, "trait parity case count drifted");
    assert!(
        failures.is_empty(),
        "PartialEq/PartialOrd trait parity failures ({} total):\n{}",
        failures.len(),
        failures.join("\n")
    );
}
`, "@DIGITS@", strings.TrimPrefix(w.selfType, "Decimal"), "@CASES@", strconv.Itoa(pairs),
		"@PAIRS@", w.pairs, "@CORPUS@", w.corpus, "@SELF@", w.selfType,
		"@PUBFROM0@", w.pubFrom("v0"), "@PUBFROM1@", w.pubFrom("v1"), "@PORTV0@", w.portArg("v0"), "@PORTV1@", w.portArg("v1"),
		"@OPFMT@", w.opFmtSpec(),
		"@EQMOD@", eqModule, "@EQFN@", eqFn, "@LTMOD@", ltModule, "@LTFN@", ltFn, "@GTMOD@", gtModule, "@GTFN@", gtFn))
	return nil
}
