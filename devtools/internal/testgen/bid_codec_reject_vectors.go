package testgen

import (
	"fmt"
	"sort"
	"strings"
)

// This file is the single source of the BID codec `reject_vectors` domain: the
// programmatically-derived reject cases written into bid754-codec-vectors/vectors.json
// and the per-language capability / type-domain tables the generated consumer
// harnesses hardcode. Cases are derived from the per-width limit tables below,
// so adding a width or a boundary auto-expands the generated reject set; no case
// is hand-enumerated per file. It must not read or write
// devtools/verification_anchors.json (that anchor is pinned outside the
// generation path on purpose, so no generator or emitter may read or write
// it).

// bidCodecRejectVector is one reject record. The `encode` channel carries a
// Components value that must be rejected by Encode{Type}; the `to_string`
// channel carries a schema-invalid Components value that must be rejected by
// ToString; the `from_string` channel carries a malformed input string that
// must be rejected by FromString.
// coefficient/payload are decimal strings (same encoding as the `vectors`
// records) so magnitudes at or above 2^64/2^128 and negative values are
// expressible. Component fields use omitempty so from_string records stay
// `{channel, input, reason}`, encode records carry `type`, and to_string records
// carry the same Components shape without a width.
type bidCodecRejectVector struct {
	Channel string `json:"channel"`
	Type    string `json:"type,omitempty"`
	// Input is a pointer so a from_string record always emits an `input` field —
	// including the empty-string reject case (`"input": ""`) — while encode
	// records (Input == nil) omit it entirely.
	Input       *string `json:"input,omitempty"`
	Sign        bool    `json:"sign,omitempty"`
	Kind        string  `json:"kind,omitempty"`
	Coefficient string  `json:"coefficient,omitempty"`
	Exponent    int32   `json:"exponent,omitempty"`
	Payload     string  `json:"payload,omitempty"`
	Reason      string  `json:"reason"`
	Requires    string  `json:"requires,omitempty"`
}

// bidCodecRejectWidthLimit pins, per BID width, the first out-of-range value on
// each field boundary. CoeffOver is width max + 1 (10^k), ExpMin/ExpMax are the
// valid unbiased exponent bounds (the reject cases use ExpMin-1 and ExpMax+1),
// and PayloadOver is the first non-canonical NaN payload (10^m). Every width now
// carries a payload limit: the Components payload field represents the full
// BID128 110-bit payload, so bid128's 10^33 limit is expressible in all six
// languages (10^33 < 2^128) and needs no capability tag.
type bidCodecRejectWidthLimit struct {
	Type        string
	CoeffOver   string
	ExpMin      int32
	ExpMax      int32
	PayloadOver string
}

var bidCodecRejectWidthLimits = []bidCodecRejectWidthLimit{
	{Type: "bid32", CoeffOver: "10000000", ExpMin: -101, ExpMax: 90, PayloadOver: "1000000"},
	{Type: "bid64", CoeffOver: "10000000000000000", ExpMin: -398, ExpMax: 369, PayloadOver: "1000000000000000"},
	{Type: "bid128", CoeffOver: "10000000000000000000000000000000000", ExpMin: -6176, ExpMax: 6111, PayloadOver: "1000000000000000000000000000000000"},
}

// bidCodecTwoPow128Decimal is 2^128, a coefficient that overflows even a 128-bit
// integer, so only big-integer Components languages can construct it.
const bidCodecTwoPow128Decimal = "340282366920938463463374607431768211456"

// bidCodecFromStringRejectCase is one malformed FromString input plus a tag
// naming the grammar boundary it violates. These absorb and replace the
// hardcoded malformed lists the per-language runners used to carry.
type bidCodecFromStringRejectCase struct {
	Input  string
	Reason string
}

var bidCodecFromStringRejectCases = []bidCodecFromStringRejectCase{
	// The original hardcoded list carried by every runner.
	{"", "empty"},
	{"NaNabc", "malformed_nan_payload"},
	{"SNaN-1", "malformed_nan_payload"},
	{"1.2.3", "multiple_decimal_points"},
	{"1E", "malformed_exponent"},
	{"1Eabc", "malformed_exponent"},
	{"1E2147483648", "exponent_out_of_int32"}, // no fraction: the FINAL exponent is out of int32
	// Divergence inputs the old lists did not exercise (stdlib parser leniency,
	// Unicode digits/whitespace, schema-limit and payload-range escapes).
	{"NaN+5", "malformed_nan_payload"},
	{"1E１", "non_ascii"},            // fullwidth digit ONE in the exponent
	{"1E1_0", "underscore"},         // digit-group underscore in the exponent
	{"１２３", "non_ascii"},            // fullwidth digits in the coefficient
	{"NaN١٢", "non_ascii"},          // arabic-indic digits in the payload
	{"1E 5", "internal_whitespace"}, // ASCII space inside the token
	{" 1", "non_ascii"},             // non-breaking space leading the token
	{"½", "non_ascii"},              // vulgar fraction one half
	{"99999999999999999999999999999999999", "coefficient_exceeds_schema_max"}, // 35 nines
	{"10000000000000000000000000000000000", "coefficient_exceeds_schema_max"}, // 10^34 = schema max + 1
	// Fraction-adjusted FINAL exponent one past either int32 edge. The literal
	// itself may exceed int32 (it only needs to stay below the shared
	// exact-integer bound 2^53): "0.001E2147483649" and "1.0E2147483648" moved
	// to the string_vectors success channel when the exponent rule became
	// final-value-based (round-trip closure), and these one-past rows pin the
	// new boundary from the reject side.
	{"1.0E-2147483648", "exponent_out_of_int32"}, // final -2147483649, below int32 min
	{"0.1E-2147483648", "exponent_out_of_int32"}, // final -2147483649, below int32 min
	{"1.0E+2147483649", "exponent_out_of_int32"}, // final +2147483648, above int32 max
	// The shared exact-integer exponent-literal bound 2^53, one-at-the-edge in
	// both signs: |literal| < 2^53 is a GRAMMAR rule (the widest bound every
	// consumer's number type can check exactly — JavaScript's safe-integer
	// range pins it; the other six consumers enforce the same constant
	// explicitly), so these two rows fix the exact literal boundary in every
	// language at the literal step through the same error channel. Both final
	// values are also outside int32, so the reason tag stays true.
	{"1E9007199254740992", "exponent_out_of_int32"},  // literal +2^53, at the shared bound
	{"1E-9007199254740992", "exponent_out_of_int32"}, // literal -2^53, at the shared bound
	// Exponent literal exactly i64 min with a one-digit fraction adjustment.
	// Under the shared 2^53 literal bound every language now rejects this at
	// the literal step (i64 parse succeeds in the 64-bit languages, then the
	// 2^53 bound fails; JS fails safe-integer; Python fails the same explicit
	// bound); the row also pins that no consumer reaches an unchecked fold —
	// Swift's earlier unchecked Int arithmetic would have trapped/crashed on
	// exactly this shape, which the no-panic contract forbids.
	{"1.0E-9223372036854775808", "exponent_out_of_int32"},
	{"NaN1000000000000000000000000000000000", "nan_payload_exceeds_schema_max"}, // payload 10^33 = schema max + 1 (2^64 is now a valid payload)
	{"++1", "malformed_sign"},
	{"--1", "malformed_sign"},
	{"+-1", "malformed_sign"},
	{"1E++5", "malformed_exponent"},
	{"1..2", "multiple_decimal_points"},
	{".", "no_digits"},
	// Machine-integer accumulation-overflow probes. The rows above never exceed
	// 35 coefficient/payload digits or a 10-digit exponent literal, so a
	// fixed-width consumer's checked digit accumulation (u128 coefficient and
	// payload in Rust/Swift, i64/long exponent accumulators) was never pushed
	// past its machine range. These literals overflow those accumulators
	// mid-parse (10^39 > u128 max ~3.4*10^38; a 25-digit exponent literal is
	// beyond both the i64 accumulator range and the shared 2^53 literal bound)
	// while big-integer languages fail the same schema caps and the same
	// explicit 2^53 exponent-literal bound — every consumer surfaces the same
	// error channel, never wrap, saturate, or panic/trap.
	{"1" + strings.Repeat("0", 39), "coefficient_exceeds_schema_max"},    // coefficient 10^39 overflows u128 accumulation
	{"1E" + strings.Repeat("9", 25), "exponent_out_of_int32"},            // exponent literal beyond i64 accumulation and the 2^53 bound
	{"NaN1" + strings.Repeat("0", 39), "nan_payload_exceeds_schema_max"}, // payload 10^39 overflows u128 accumulation
	// Multi-byte UTF-8 at a range of byte offsets. A Go parser slices a string
	// at any byte offset and never faults, so a probe written as a slice passes
	// every Go-side gate; the generated Rust cuts a &str at the same offset and
	// panics when the cut is inside a character. The public parse surface must
	// reject all of these through its error type in both languages, never by
	// unwinding, which docs/SPEC.md requires of every public API path.
	//
	// Some of the non_ascii rows above already trip a cut at some offsets, so
	// these rows are not what first made the class detectable -- what was
	// missing was a consumer that ran ANY of them through the generated Rust
	// public parse path (bid754-rs/tests/bid_codec_parse_vectors.rs, added with
	// this block). What the rows add is a full offset grid, so detection does
	// not depend on which pre-existing input happens to collide with the probe
	// width a future defect uses.
	//
	// Measured over the whole non_ascii class, the first-non-ASCII byte offset
	// by character width is:
	//
	//   2-byte: 0 1 2 3 4 8 9
	//   3-byte: 0 1 2 3 4
	//   4-byte: 0 1 2 3 4
	//
	// 0..4 spans every fixed-position read on the parse path (the sign/radix/
	// digit probe at 1 and the four-byte "snan"/"infi" prefix probes), and the
	// 8/9 rows land inside the digit loop, past every fixed probe. Both the
	// numeric and the letter entry path are represented.
	{"1é", "non_ascii"},    // cut at offset 1, 2-byte character
	{"12é", "non_ascii"},   // cut at offset 2
	{"123é", "non_ascii"},  // cut at offset 3
	{"1234é", "non_ascii"}, // cut at offset 4, the four-byte prefix probe width
	{"-123é", "non_ascii"}, // minus consumed first, so the probe cuts past it
	{"+123é", "non_ascii"}, // plus takes the other sign branch
	{"1.23é", "non_ascii"}, // radix point on the path to the same cut
	{"1中", "non_ascii"},    // 3-byte character, cut at offset 1
	{"123中", "non_ascii"},  // 3-byte, offset 3
	{"1234中", "non_ascii"}, // 3-byte, offset 4
	{"😀", "non_ascii"},     // 4-byte character, offset 0
	{"1😀", "non_ascii"},    // 4-byte, offset 1
	{"12😀", "non_ascii"},   // 4-byte, offset 2
	{"123😀", "non_ascii"},  // 4-byte, offset 3
	{"1234😀", "non_ascii"}, // 4-byte, offset 4
	{"snané", "non_ascii"}, // sNaN prefix probe with a multi-byte tail
	{"snaé", "non_ascii"},  // one byte SHORT of the probe width, so the cut is inside the character
	{"é", "non_ascii"},     // leading 2-byte character, no ASCII prefix
	{"中文字", "non_ascii"},   // 3-byte characters, no ASCII prefix
	// Past every probe width, inside the digit loop. The rows above all place
	// the character within the first four bytes, where the fixed-width Inf/NaN
	// probes read; a coefficient long enough to clear them reaches the loop
	// that walks digits one at a time, which is a separate read site.
	{"12345678é", "non_ascii"},
	{"1.2345678é", "non_ascii"},
	// The rows above all start with a digit or sign, so they enter the numeric
	// path. A leading letter takes the special-case branch instead, where the
	// Inf/NaN probes run against the whole token -- a different set of probes
	// reading the same offsets. The short forms put the character at offsets 1
	// and 2 on that path, which the digit rows cover only on the numeric path.
	{"aé", "non_ascii"},
	{"aaé", "non_ascii"},
	{"aaaé", "non_ascii"},
	{"nané", "non_ascii"},
	// U+0130 LATIN CAPITAL LETTER I WITH DOT ABOVE. Unicode simple case
	// mapping lowers it to ASCII 'i', so a parser that folds with Unicode
	// semantics reads this as "inf" and returns Infinity, while the pinned
	// Intel C and its ASCII byte fold reject it. The row pins the ASCII fold.
	{"İnf", "non_ascii"},
}

// bidCodecEncodeRejectVectors derives the encode-channel reject records from the
// per-width limit table plus the three field-domain cases that only big-integer
// / signed-field Components languages can construct.
func bidCodecEncodeRejectVectors() []bidCodecRejectVector {
	var out []bidCodecRejectVector
	for _, w := range bidCodecRejectWidthLimits {
		// coefficient above the width maximum (max + 1)
		out = append(out, bidCodecRejectVector{
			Channel: "encode", Type: w.Type, Kind: "normal",
			Coefficient: w.CoeffOver, Reason: "coefficient_overflow",
		})
		// unbiased exponent below the width minimum (applies to zero and normal;
		// zero isolates the exponent check)
		out = append(out, bidCodecRejectVector{
			Channel: "encode", Type: w.Type, Kind: "zero",
			Exponent: w.ExpMin - 1, Reason: "exponent_out_of_range",
		})
		// unbiased exponent above the width maximum
		out = append(out, bidCodecRejectVector{
			Channel: "encode", Type: w.Type, Kind: "zero",
			Exponent: w.ExpMax + 1, Reason: "exponent_out_of_range",
		})
		// NaN payload at or above the width canonical payload limit
		if w.PayloadOver != "" {
			out = append(out, bidCodecRejectVector{
				Channel: "encode", Type: w.Type, Kind: "qnan",
				Payload: w.PayloadOver, Reason: "payload_overflow",
			})
		}

		// Fields outside the selected kind's representation domain must not be
		// silently discarded. Keep one field varied per record so accepting any
		// individual extraneous field cannot hide behind rejection of another.
		out = append(out,
			bidCodecRejectVector{
				Channel: "encode", Type: w.Type, Kind: "normal",
				Coefficient: "0", Reason: "zero_coefficient_not_representable_for_normal",
			},
			bidCodecRejectVector{
				Channel: "encode", Type: w.Type, Kind: "normal",
				Coefficient: "1", Payload: "1", Reason: "payload_not_representable_for_normal",
			},
			bidCodecRejectVector{
				Channel: "encode", Type: w.Type, Kind: "zero",
				Coefficient: "1", Reason: "coefficient_not_representable_for_zero",
			},
			bidCodecRejectVector{
				Channel: "encode", Type: w.Type, Kind: "zero",
				Payload: "1", Reason: "payload_not_representable_for_zero",
			},
			bidCodecRejectVector{
				Channel: "encode", Type: w.Type, Kind: "inf",
				Coefficient: "1", Reason: "coefficient_not_representable_for_infinity",
			},
			bidCodecRejectVector{
				Channel: "encode", Type: w.Type, Kind: "inf",
				Exponent: 1, Reason: "exponent_not_representable_for_infinity",
			},
			bidCodecRejectVector{
				Channel: "encode", Type: w.Type, Kind: "inf",
				Payload: "1", Reason: "payload_not_representable_for_infinity",
			},
			bidCodecRejectVector{
				Channel: "encode", Type: w.Type, Kind: "qnan",
				Coefficient: "1", Reason: "coefficient_not_representable_for_qnan",
			},
			bidCodecRejectVector{
				Channel: "encode", Type: w.Type, Kind: "qnan",
				Exponent: 1, Reason: "exponent_not_representable_for_qnan",
			},
			bidCodecRejectVector{
				Channel: "encode", Type: w.Type, Kind: "snan",
				Coefficient: "1", Reason: "coefficient_not_representable_for_snan",
			},
			bidCodecRejectVector{
				Channel: "encode", Type: w.Type, Kind: "snan",
				Exponent: 1, Reason: "exponent_not_representable_for_snan",
			},
		)
	}

	// Field-domain cases gated by capability: a consumer whose Components field
	// types cannot express the value skips the record and reports the skip.
	out = append(out, bidCodecRejectVector{
		Channel: "encode", Type: "bid128", Kind: "normal",
		Coefficient: bidCodecTwoPow128Decimal, Reason: "coefficient_overflow",
		Requires: "bignum_coefficient",
	})
	out = append(out, bidCodecRejectVector{
		Channel: "encode", Type: "bid32", Kind: "normal",
		Coefficient: "-1", Reason: "negative_coefficient",
		Requires: "negative_coefficient",
	})
	out = append(out, bidCodecRejectVector{
		Channel: "encode", Type: "bid32", Kind: "qnan",
		Payload: "-1", Reason: "negative_payload",
		Requires: "negative_payload",
	})
	return out
}

// bidCodecFromStringRejectVectors derives the from_string-channel reject records.
func bidCodecFromStringRejectVectors() []bidCodecRejectVector {
	out := make([]bidCodecRejectVector, 0, len(bidCodecFromStringRejectCases))
	for _, c := range bidCodecFromStringRejectCases {
		input := c.Input
		out = append(out, bidCodecRejectVector{
			Channel: "from_string", Input: &input, Reason: c.Reason,
		})
	}
	return out
}

// bidCodecToStringRejectVectors pins the width-independent Components schema
// accepted by the shared string renderer. Every accepted value must render to
// a string FromString can parse back without losing Kind or any nonzero field.
// These records isolate every lossy Kind/field combination plus the schema-wide
// coefficient and NaN-payload limits. Negative values reuse the existing
// capability tags because unsigned Components languages cannot construct them.
func bidCodecToStringRejectVectors() []bidCodecRejectVector {
	return []bidCodecRejectVector{
		{Channel: "to_string", Kind: "normal", Reason: "missing_coefficient_not_representable_for_normal"},
		{Channel: "to_string", Kind: "normal", Coefficient: "0", Reason: "zero_coefficient_not_representable_for_normal"},
		{Channel: "to_string", Kind: "normal", Coefficient: "10000000000000000000000000000000000", Reason: "coefficient_exceeds_schema_max"},
		{Channel: "to_string", Kind: "normal", Coefficient: "1", Payload: "1", Reason: "payload_not_representable_for_normal"},
		{Channel: "to_string", Kind: "zero", Coefficient: "1", Reason: "coefficient_not_representable_for_zero"},
		{Channel: "to_string", Kind: "zero", Payload: "1", Reason: "payload_not_representable_for_zero"},
		{Channel: "to_string", Kind: "inf", Coefficient: "1", Reason: "coefficient_not_representable_for_infinity"},
		{Channel: "to_string", Kind: "inf", Exponent: 1, Reason: "exponent_not_representable_for_infinity"},
		{Channel: "to_string", Kind: "inf", Payload: "1", Reason: "payload_not_representable_for_infinity"},
		{Channel: "to_string", Kind: "qnan", Coefficient: "1", Reason: "coefficient_not_representable_for_qnan"},
		{Channel: "to_string", Kind: "qnan", Exponent: 1, Reason: "exponent_not_representable_for_qnan"},
		{Channel: "to_string", Kind: "qnan", Payload: "1000000000000000000000000000000000", Reason: "nan_payload_exceeds_schema_max"},
		{Channel: "to_string", Kind: "snan", Coefficient: "1", Reason: "coefficient_not_representable_for_snan"},
		{Channel: "to_string", Kind: "snan", Exponent: 1, Reason: "exponent_not_representable_for_snan"},
		{Channel: "to_string", Kind: "snan", Payload: "1000000000000000000000000000000000", Reason: "nan_payload_exceeds_schema_max"},
		{Channel: "to_string", Kind: "normal", Coefficient: "-1", Reason: "negative_coefficient", Requires: "negative_coefficient"},
		{Channel: "to_string", Kind: "qnan", Payload: "-1", Reason: "negative_payload", Requires: "negative_payload"},
	}
}

// bidCodecRejectVectors assembles the full reject_vectors array (from_string,
// then encode, then to_string records). Order is deterministic.
func bidCodecRejectVectors() []bidCodecRejectVector {
	out := bidCodecFromStringRejectVectors()
	out = append(out, bidCodecEncodeRejectVectors()...)
	out = append(out, bidCodecToStringRejectVectors()...)
	return out
}

// bidCodecRejectCapabilities is the master capability -> language map. A reject
// record tagged `requires` is only consumable by languages whose Components
// field types can construct the value; every other language skips it and reports
// the skip. Each generated consumer hardcodes its own set (emitted from here).
var bidCodecRejectCapabilities = map[string][]string{
	// Go's Components.Payload is now a signed *big.Int, so Go can construct a
	// negative payload (negative_payload) as well as a big coefficient and a
	// negative coefficient. Rust (u128) and Swift (UInt64 pair) remain unsigned
	// fixed-width, so they still skip every requires-tagged record.
	// "rust_full" is the additional (non-required) consumer generated for the
	// bid754-rs full library's embedded bid_codec module; it mirrors the
	// standalone Rust codec's u128 field types, so its capability set is the
	// same empty set as "rust".
	"go":        {"bignum_coefficient", "negative_coefficient", "negative_payload"},
	"rust":      {},
	"rust_full": {},
	"java":      {"bignum_coefficient", "negative_coefficient", "negative_payload"},
	"python":    {"bignum_coefficient", "negative_coefficient", "negative_payload"},
	"js":        {"bignum_coefficient", "negative_coefficient", "negative_payload"},
	"swift":     {},
}

// bidCodecRejectCapsElems renders a language's capability tags as a
// comma-separated list of quoted string literals (empty string when the
// language has no capabilities), for splicing into that language's set literal.
func bidCodecRejectCapsElems(lang string) string {
	caps, ok := bidCodecRejectCapabilities[lang]
	if !ok {
		panic(fmt.Sprintf("BID codec reject capabilities: unknown language %q", lang))
	}
	quoted := make([]string, len(caps))
	for i, c := range caps {
		quoted[i] = fmt.Sprintf("%q", c)
	}
	return strings.Join(quoted, ", ")
}

// bidCodecRejectRequiresUniverse returns the sorted set of `requires` tags
// actually carried by the generated reject records. It is derived from the
// records (not the capability table) so a new tag introduced on the record
// side always enters the universe the consumers pin against.
func bidCodecRejectRequiresUniverse() []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range bidCodecRejectVectors() {
		if r.Requires != "" && !seen[r.Requires] {
			seen[r.Requires] = true
			out = append(out, r.Requires)
		}
	}
	sort.Strings(out)
	return out
}

// bidCodecRejectExpectedCounts derives, from the capability table and the
// generated records, how many reject records a language's consumer must
// consume versus skip. Every generated consumer hardcodes these two counts so
// a generator regression that flips records into the skipped bucket fails the
// runner instead of passing as "all skipped".
func bidCodecRejectExpectedCounts(lang string) (consumed, skipped int) {
	capsList, ok := bidCodecRejectCapabilities[lang]
	if !ok {
		panic(fmt.Sprintf("BID codec reject capabilities: unknown language %q", lang))
	}
	caps := map[string]bool{}
	for _, c := range capsList {
		caps[c] = true
	}
	for _, r := range bidCodecRejectVectors() {
		if r.Requires != "" && !caps[r.Requires] {
			skipped++
			continue
		}
		consumed++
	}
	return consumed, skipped
}

// bidCodecRejectUnsupportedElems renders, for a language, the requires tags it
// must skip (the record-derived universe minus its capability set) as quoted
// literals. Consumers assert every skipped record's tag is in this set, so an
// unknown new tag fails the harness instead of silently widening the skips.
func bidCodecRejectUnsupportedElems(lang string) string {
	capsList, ok := bidCodecRejectCapabilities[lang]
	if !ok {
		panic(fmt.Sprintf("BID codec reject capabilities: unknown language %q", lang))
	}
	caps := map[string]bool{}
	for _, c := range capsList {
		caps[c] = true
	}
	var quoted []string
	for _, tag := range bidCodecRejectRequiresUniverse() {
		if !caps[tag] {
			quoted = append(quoted, fmt.Sprintf("%q", tag))
		}
	}
	return strings.Join(quoted, ", ")
}

// bidCodecTypeDomainRejectCase is a type-domain reject that the shared JSON
// schema cannot express (a non-boolean sign, a wrong numeric field type, an
// out-of-int32/non-finite exponent, or a kind outside the defined set). It is
// emitted directly into the generated harnesses of the
// languages whose Components field types can construct it; each field is that
// language's Components construction expression, "" when unconstructible.
type bidCodecTypeDomainRejectCase struct {
	ID string
	Go string
	Py string
	JS string
}

var bidCodecTypeDomainRejectCases = []bidCodecTypeDomainRejectCase{
	{
		ID: "non_boolean_sign",
		Py: `Components(sign=1, kind=Kind.NORMAL, coefficient=1, exponent=0)`,
		JS: `{ sign: 1, kind: Kind.Normal, coefficient: 1n, exponent: 0, payload: 0n }`,
	},
	{
		ID: "non_integral_exponent",
		Py: `Components(kind=Kind.NORMAL, coefficient=1, exponent=1.5)`,
		JS: `{ sign: false, kind: Kind.Normal, coefficient: 1n, exponent: 1.5, payload: 0n }`,
	},
	{
		ID: "nan_exponent",
		Py: `Components(kind=Kind.NORMAL, coefficient=1, exponent=float("nan"))`,
		JS: `{ sign: false, kind: Kind.Normal, coefficient: 1n, exponent: Number.NaN, payload: 0n }`,
	},
	{
		ID: "exponent_above_int32",
		Py: `Components(kind=Kind.NORMAL, coefficient=1, exponent=2147483648)`,
		JS: `{ sign: false, kind: Kind.Normal, coefficient: 1n, exponent: 2147483648, payload: 0n }`,
	},
	{
		ID: "exponent_below_int32",
		Py: `Components(kind=Kind.NORMAL, coefficient=1, exponent=-2147483649)`,
		JS: `{ sign: false, kind: Kind.Normal, coefficient: 1n, exponent: -2147483649, payload: 0n }`,
	},
	{
		ID: "non_integer_coefficient",
		Py: `Components(kind=Kind.NORMAL, coefficient=1.5, exponent=0)`,
		JS: `{ sign: false, kind: Kind.Normal, coefficient: 1, exponent: 0, payload: 0n }`,
	},
	{
		ID: "non_integer_payload",
		Py: `Components(kind=Kind.QNAN, payload=1.5)`,
		JS: `{ sign: false, kind: Kind.QNaN, coefficient: 0n, exponent: 0, payload: 1 }`,
	},
	{
		ID: "boolean_coefficient",
		Py: `Components(kind=Kind.NORMAL, coefficient=True, exponent=0)`,
		JS: `{ sign: false, kind: Kind.Normal, coefficient: true, exponent: 0, payload: 0n }`,
	},
	{
		ID: "boolean_exponent",
		Py: `Components(kind=Kind.ZERO, exponent=True)`,
		JS: `{ sign: false, kind: Kind.Zero, coefficient: 0n, exponent: true, payload: 0n }`,
	},
	{
		ID: "boolean_payload",
		Py: `Components(kind=Kind.QNAN, payload=True)`,
		JS: `{ sign: false, kind: Kind.QNaN, coefficient: 0n, exponent: 0, payload: true }`,
	},
	{
		ID: "unrecognized_kind",
		Go: `Components{Kind: Kind(99)}`,
		Py: `Components(kind=99, coefficient=1, exponent=0)`,
		JS: `{ sign: false, kind: 99, coefficient: 1n, exponent: 0, payload: 0n }`,
	},
}

// bidCodecGoTypeDomainElems renders the Go type-domain reject slice elements
// `{"id", <ComponentsExpr>},` for the cases Go can construct.
func bidCodecGoTypeDomainElems() string {
	var b strings.Builder
	for _, tc := range bidCodecTypeDomainRejectCases {
		if tc.Go == "" {
			continue
		}
		fmt.Fprintf(&b, "\n\t\t{%q, %s},", tc.ID, tc.Go)
	}
	b.WriteString("\n\t")
	return b.String()
}

// bidCodecPyTypeDomainElems renders the Python type-domain reject list elements
// `("id", <ComponentsExpr>),` for the cases Python can construct.
func bidCodecPyTypeDomainElems() string {
	var b strings.Builder
	for _, tc := range bidCodecTypeDomainRejectCases {
		if tc.Py == "" {
			continue
		}
		fmt.Fprintf(&b, "\n    (%q, %s),", tc.ID, tc.Py)
	}
	b.WriteString("\n")
	return b.String()
}

// bidCodecJsTypeDomainElems renders the JS type-domain reject array elements
// `["id", <ComponentsExpr>],` for the cases JavaScript can construct.
func bidCodecJsTypeDomainElems() string {
	var b strings.Builder
	for _, tc := range bidCodecTypeDomainRejectCases {
		if tc.JS == "" {
			continue
		}
		fmt.Fprintf(&b, "\n  [%q, %s],", tc.ID, tc.JS)
	}
	b.WriteString("\n")
	return b.String()
}

// bidCodecRawDecodeRejectCase is a public raw-word decode input that a dynamic
// language can construct but that is not an exact word in the declared API
// domain. The generated Python and JavaScript harnesses invoke the production
// Decode* functions directly and require their language-specific error channel.
type bidCodecRawDecodeRejectCase struct {
	ID string
	Py string
	JS string
}

var bidCodecRawDecodeRejectCases = []bidCodecRawDecodeRejectCase{
	{ID: "decode32_boolean", Py: `lambda: decode32(True)`, JS: `() => Reflect.apply(decode32, undefined, [true])`},
	{ID: "decode32_string", Py: `lambda: decode32("1")`, JS: `() => Reflect.apply(decode32, undefined, ["1"])`},
	{ID: "decode32_fraction", Py: `lambda: decode32(1.5)`, JS: `() => Reflect.apply(decode32, undefined, [1.5])`},
	{ID: "decode32_nan", Py: `lambda: decode32(float("nan"))`, JS: `() => Reflect.apply(decode32, undefined, [Number.NaN])`},
	{ID: "decode32_infinity", Py: `lambda: decode32(float("inf"))`, JS: `() => Reflect.apply(decode32, undefined, [Number.POSITIVE_INFINITY])`},
	{ID: "decode32_negative", Py: `lambda: decode32(-1)`, JS: `() => Reflect.apply(decode32, undefined, [-1])`},
	{ID: "decode32_overflow", Py: `lambda: decode32(1 << 32)`, JS: `() => Reflect.apply(decode32, undefined, [2 ** 32])`},
	{ID: "decode32_bigint", JS: `() => Reflect.apply(decode32, undefined, [1n])`},
	{ID: "decode64_boolean", Py: `lambda: decode64(True)`, JS: `() => Reflect.apply(decode64, undefined, [true])`},
	{ID: "decode64_string", Py: `lambda: decode64("1")`, JS: `() => Reflect.apply(decode64, undefined, ["1"])`},
	{ID: "decode64_wrong_numeric_type", Py: `lambda: decode64(1.5)`, JS: `() => Reflect.apply(decode64, undefined, [1])`},
	{ID: "decode64_negative", Py: `lambda: decode64(-1)`, JS: `() => Reflect.apply(decode64, undefined, [-1n])`},
	{ID: "decode64_overflow", Py: `lambda: decode64(1 << 64)`, JS: `() => Reflect.apply(decode64, undefined, [1n << 64n])`},
	{ID: "decode128_lo_boolean", Py: `lambda: decode128(True, 0)`, JS: `() => Reflect.apply(decode128, undefined, [true, 0n])`},
	{ID: "decode128_lo_string", Py: `lambda: decode128("1", 0)`, JS: `() => Reflect.apply(decode128, undefined, ["1", 0n])`},
	{ID: "decode128_lo_wrong_numeric_type", Py: `lambda: decode128(1.5, 0)`, JS: `() => Reflect.apply(decode128, undefined, [1, 0n])`},
	{ID: "decode128_lo_negative", Py: `lambda: decode128(-1, 0)`, JS: `() => Reflect.apply(decode128, undefined, [-1n, 0n])`},
	{ID: "decode128_lo_overflow", Py: `lambda: decode128(1 << 64, 0)`, JS: `() => Reflect.apply(decode128, undefined, [1n << 64n, 0n])`},
	{ID: "decode128_hi_boolean", Py: `lambda: decode128(0, True)`, JS: `() => Reflect.apply(decode128, undefined, [0n, true])`},
	{ID: "decode128_hi_string", Py: `lambda: decode128(0, "1")`, JS: `() => Reflect.apply(decode128, undefined, [0n, "1"])`},
	{ID: "decode128_hi_wrong_numeric_type", Py: `lambda: decode128(0, 1.5)`, JS: `() => Reflect.apply(decode128, undefined, [0n, 1])`},
	{ID: "decode128_hi_negative", Py: `lambda: decode128(0, -1)`, JS: `() => Reflect.apply(decode128, undefined, [0n, -1n])`},
	{ID: "decode128_hi_overflow", Py: `lambda: decode128(0, 1 << 64)`, JS: `() => Reflect.apply(decode128, undefined, [0n, 1n << 64n])`},
}

func bidCodecPyRawDecodeRejectElems() string {
	var b strings.Builder
	for _, tc := range bidCodecRawDecodeRejectCases {
		if tc.Py == "" {
			continue
		}
		fmt.Fprintf(&b, "\n    (%q, %s),", tc.ID, tc.Py)
	}
	b.WriteString("\n")
	return b.String()
}

func bidCodecJsRawDecodeRejectElems() string {
	var b strings.Builder
	for _, tc := range bidCodecRawDecodeRejectCases {
		if tc.JS == "" {
			continue
		}
		fmt.Fprintf(&b, "\n  [%q, %s],", tc.ID, tc.JS)
	}
	b.WriteString("\n")
	return b.String()
}
