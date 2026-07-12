package testgen

// This file is the single source of the BID codec `string_vectors` domain: the
// deterministic SUCCESS-channel string records written into
// bid754-codec-vectors/vectors.json. Each record pins the string surface on
// its own: `fromString(input)` must SUCCEED, `toString` of the parsed
// Components must render the exact `expected` string, and — the round-trip
// closure leg every generated consumer also executes — `fromString(expected)`
// must itself succeed and re-render as the same `expected` string (a toString
// fixed point), in every language consumer.
//
// Why a third channel exists: the `vectors` channel starts from decodable BID
// bits, so it can only reach Components values that some BID width encodes;
// the `reject_vectors` channel pins inputs that must FAIL. Neither pins the
// fromString→toString result for contract-valid Components that no BID width
// can encode — above all the int32-extreme exponents, where a to_string
// implementation computing the adjusted exponent (exponent + digits - 1) in
// 32-bit arithmetic silently wraps the sign (the Java consumer rendered
// "12E+2147483647" as "+1.2E-2147483648" while the other six consumers agreed
// on "+1.2E+2147483648", and no generated domain caught the divergence). The
// grammar-edge rows (whitespace trim, empty integer/fraction sides, leading
// zeros, payload normalization, case-insensitive specials) pin the successful
// parse normalizations the reject channel only bounds from the failure side.
//
// The closure leg exists because the fuzzer found the contract was not closed:
// FromString("10E2147483647") succeeded, ToString rendered "+1.0E+2147483648"
// (adjusted exponent int32 max + 1), and the parser then REJECTED its own
// rendering under the old "the exponent literal itself must fit int32" rule —
// identically in all seven consumers. The shared exponent rule is now
// final-value-based: the literal must be below the shared exact-integer bound
// 2^53 in magnitude (the grammar constant every consumer checks exactly) and
// only the fraction-adjusted FINAL exponent must fit int32, which makes
// parse(render(x)) total (a valid Components with n coefficient digits and
// exponent e renders the adjusted-exponent literal e+n-1 ∈
// [int32 min, int32 max + 33], far below 2^53, and reparsing subtracts the
// same n-1).
//
// Expected values are pinned literals. The devtools reference codec
// (bid_codec_reference.go) carries a to_string renderer (bidCodecDecimalString)
// but no from_string parser, so the generator cannot derive or re-assert these
// expectations at generation time; the rows were cross-measured on all seven
// language consumers (Go, standalone Rust, bid754-rs bid_codec, Java, Python,
// JS, Swift produced byte-identical output for every row, 2026-07 probe, after
// the Java 64-bit adjusted-exponent fix) and every generated consumer harness
// re-executes them on each `make test-bidcodec` run.
//
// Like every generation-path file, this file must not read or write
// devtools/verification_anchors.json (no generator or emitter may read or
// write that anchor); the domain total is pinned there by hand.

// bidCodecStringVector is one success-channel string record: fromString(Input)
// must succeed and toString of the result must equal Expected exactly.
type bidCodecStringVector struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

// bidCodecStringVectorRows is the deterministic literal table (order is the
// emission order). No row carries a capability tag: every input is a pure
// ASCII string, constructible in all language consumers, so each consumer
// consumes every row.
var bidCodecStringVectorRows = []bidCodecStringVector{
	// int32-extreme exponents: adjusted exponent exceeds int32 (the divergence
	// class the Java consumer silently wrapped) and its negative-edge mirrors.
	{"12E+2147483647", "+1.2E+2147483648"},
	{"9999999999999999999999999999999999E+2147483647", "+9.999999999999999999999999999999999E+2147483680"},
	{"1.5E2147483647", "+1.5E+2147483647"},
	{"999E+2147483645", "+9.99E+2147483647"},
	{"1E-2147483648", "+1E-2147483648"},
	{"0.1E-2147483647", "+1E-2147483648"},
	{"0E+2147483647", "+0E+2147483647"},
	{"0E-2147483648", "+0E-2147483648"},
	// Round-trip closure at the rendered top edge (the fuzzer-found case): the
	// input parses to coefficient 10 / exponent int32 max, whose rendering
	// carries the adjusted-exponent literal int32 max + 1 — the rendering
	// itself is the next row's input shape, so these rows pin both the parse
	// and the closure leg. The literal-past-int32 rows moved here from
	// reject_vectors when the exponent rule became final-value-based.
	{"10E2147483647", "+1.0E+2147483648"},
	{"1.0E2147483648", "+1.0E+2147483648"},
	{"0.001E2147483649", "+1E+2147483646"},
	// zero forms: fraction-adjusted zero exponent and leading-zero collapse.
	{"0.00", "+0E-2"},
	{"000", "+0"},
	// grammar edges: empty integer/fraction side, leading zeros with a kept
	// fractional tail.
	{".5", "+5E-1"},
	{"5.", "+5E+0"},
	{"001.100", "+1.100E+0"},
	// NaN payload normalization (leading zeros dropped), schema-max payload
	// kept verbatim, and case-insensitive special tokens.
	{"NaN000123", "+NaN123"},
	{"-InFiNiTy", "-Inf"},
	{"NaN999999999999999999999999999999999", "+NaN999999999999999999999999999999999"},
	// surrounding ASCII whitespace trim (TAB before, CR after).
	{"\t1\r", "+1E+0"},
}

// bidCodecStringVectors returns the string_vectors records in emission order.
func bidCodecStringVectors() []bidCodecStringVector {
	return bidCodecStringVectorRows
}
