package testgen

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// This file is the single source of the `go_full` BID codec vector consumer's
// expectation tables: the additional (non-required) consumer generated for the
// bid754-go full decimal library's PUBLIC parse surface (NewDecimal{32,64,128},
// NewDecimal*WithFlags, NewDecimal*WithMode). Unlike `rust_full`, which
// re-verifies an embedded Components codec against the standalone codec
// contract, bid754-go has no embedded Components codec; the surface under test
// is the public fromString contract itself, whose grammar and rejection
// boundaries intentionally differ from the codec fromString schema. The maps
// below pin, per record, the expected PUBLIC observation class:
//
//   - "exact":    Direct succeeds; WithFlags succeeds with zero flags and the
//                 same bits; WithMode(NearestEven) matches WithFlags; and the
//                 public render/parse closure holds (Direct(v.String()) == v).
//   - "rounded":  Direct errors (exact-only contract); WithFlags succeeds with
//                 nonzero flags (the IEEE flag channel reports the rounding /
//                 range excursion instead of silence); WithMode matches.
//   - "rejected": every family errors with a zero value and zero flags (public
//                 grammar violation, silent-cohort trap, or a NaN payload
//                 outside the width's range).
//
// The classes are generator-owned pinned data with the same standing as the
// string_vectors `expected` literals: they were cross-measured against the
// public API (all 79 rows x 3 widths x 3 families, 2026-07 probe; the extreme
// rows are additionally bit+flag pinned by the hand-written
// bid754-go/parse_literal_boundary_public_test.go), and the generated runner
// re-executes them on every `make test-bidcodec` run. Both maps are
// closed-world in both directions against the generated record lists: a new
// record without an expectation, a stale expectation without a record, or an
// unknown class name fails generation. Like every generation-path file, this
// file must not read or write devtools/verification_anchors.json; the go_full
// consumed/skipped totals are pinned there by hand.

// bidCodecGoFullClasses is the closed set of go_full observation classes.
var bidCodecGoFullClasses = map[string]bool{
	"exact":    true,
	"rounded":  true,
	"rejected": true,
}

// bidCodecGoFullFromStringClasses maps every reject_vectors from_string input
// to its go_full class. The codec rejects each of these inputs at its grammar
// or schema boundary; the public Go parse observation differs by class:
// grammar-invalid inputs and exact-value silent-cohort / NaN-payload traps are
// "rejected" in all three families, while inputs whose only defect is a value
// outside every width's range are "rounded" (Direct errors, WithFlags reports
// overflow/underflow+inexact). Classes are width-independent on this channel.
var bidCodecGoFullFromStringClasses = map[string]string{
	// Public-grammar rejects (malformed syntax, non-ASCII, whitespace, signs).
	"":        "rejected",
	"NaNabc":  "rejected",
	"SNaN-1":  "rejected",
	"1.2.3":   "rejected",
	"1E":      "rejected",
	"1Eabc":   "rejected",
	"NaN+5":   "rejected",
	"1E１":     "rejected",
	"1E1_0":   "rejected",
	"１２３":     "rejected",
	"NaN١٢":   "rejected",
	"1E 5":    "rejected",
	"\u00a01": "rejected", // non-breaking space leading the token
	"½":       "rejected",
	"++1":     "rejected",
	"--1":     "rejected",
	"+-1":     "rejected",
	"1E++5":   "rejected",
	"1..2":    "rejected",
	".":       "rejected",
	// Value-exact written cohorts no width can encode (the port parses these
	// with exact status, so the public silent-cohort trap rejects them).
	"10000000000000000000000000000000000": "rejected", // 10^34, 35 written digits
	"1" + strings.Repeat("0", 39):         "rejected", // 10^39, 40 written digits
	// NaN payloads above every width's payload range (invalid operation).
	"NaN1000000000000000000000000000000000": "rejected", // payload 10^33
	"NaN1" + strings.Repeat("0", 39):        "rejected", // payload 10^39
	// Values outside every width's range or precision: the flag channel
	// reports overflow/underflow/inexact, the exact-only channel errors.
	"1E2147483648":                        "rounded", // overflow+inexact
	"99999999999999999999999999999999999": "rounded", // 35 nines, inexact
	"1.0E-2147483648":                     "rounded", // underflow+inexact
	"0.1E-2147483648":                     "rounded", // underflow+inexact
	"1.0E+2147483649":                     "rounded", // overflow+inexact
	"1E9007199254740992":                  "rounded", // overflow+inexact
	"1E-9007199254740992":                 "rounded", // underflow+inexact
	"1.0E-9223372036854775808":            "rounded", // underflow+inexact
	"1E" + strings.Repeat("9", 25):        "rounded", // overflow+inexact
	// Multi-byte UTF-8 carrying a valid ASCII prefix. The public grammar is
	// ASCII, so every family rejects; what these rows add over the other
	// grammar rejects is that the parser walks INTO the multi-byte character
	// before rejecting, which is where a &str cut in the generated Rust
	// panics instead of returning the error.
	"1é":         "rejected",
	"12é":        "rejected",
	"123é":       "rejected",
	"1234é":      "rejected",
	"12345678é":  "rejected",
	"1.2345678é": "rejected",
	"-123é":      "rejected",
	"+123é":      "rejected",
	"1.23é":      "rejected",
	"1中":         "rejected",
	"123中":       "rejected",
	"1234中":      "rejected",
	"😀":          "rejected",
	"1😀":         "rejected",
	"12😀":        "rejected",
	"123😀":       "rejected",
	"1234😀":      "rejected",
	"snané":      "rejected",
	"snaé":       "rejected",
	"é":          "rejected",
	"中文字":        "rejected",
	"aé":         "rejected",
	"aaé":        "rejected",
	"aaaé":       "rejected",
	"nané":       "rejected",
	"İnf":        "rejected", // Unicode lowers U+0130 to ASCII 'i'; the ASCII fold must not
}

// bidCodecGoFullStringVectorClasses maps every string_vectors input to its
// go_full classes for widths 32, 64, and 128 in that order. The codec parses
// every one of these successfully; on the fixed-width public surface the
// int32-extreme exponent rows are range excursions ("rounded"), the extreme
// zero-cohort row and the trailing-whitespace grammar row are "rejected"
// (surrounding-whitespace trim is codec grammar, not public Go grammar; the
// zero asymmetry — 0E+2147483647 rejected, 0E-2147483648 rounded — follows the
// pinned port behavior of clamping high zero exponents with exact status while
// raising underflow+inexact at the low extreme, exactly like the hand-pinned
// "0e91"/"0e-398" boundary rows), and the representable rows are "exact" with
// the public render/parse closure asserted. NaN payload 10^33-1 fits only
// Decimal128, so that row is width-dependent.
var bidCodecGoFullStringVectorClasses = map[string][3]string{
	"12E+2147483647": {"rounded", "rounded", "rounded"},
	"9999999999999999999999999999999999E+2147483647": {"rounded", "rounded", "rounded"},
	"1.5E2147483647":                       {"rounded", "rounded", "rounded"},
	"999E+2147483645":                      {"rounded", "rounded", "rounded"},
	"1E-2147483648":                        {"rounded", "rounded", "rounded"},
	"0.1E-2147483647":                      {"rounded", "rounded", "rounded"},
	"0E+2147483647":                        {"rejected", "rejected", "rejected"},
	"0E-2147483648":                        {"rounded", "rounded", "rounded"},
	"10E2147483647":                        {"rounded", "rounded", "rounded"},
	"1.0E2147483648":                       {"rounded", "rounded", "rounded"},
	"0.001E2147483649":                     {"rounded", "rounded", "rounded"},
	"0.00":                                 {"exact", "exact", "exact"},
	"000":                                  {"exact", "exact", "exact"},
	".5":                                   {"exact", "exact", "exact"},
	"5.":                                   {"exact", "exact", "exact"},
	"001.100":                              {"exact", "exact", "exact"},
	"NaN000123":                            {"exact", "exact", "exact"},
	"-InFiNiTy":                            {"exact", "exact", "exact"},
	"NaN999999999999999999999999999999999": {"rejected", "rejected", "exact"},
	"\t1\r":                                {"rejected", "rejected", "rejected"},
}

// bidCodecGoFullFromStringRecords returns the from_string reject records in
// emission order after verifying the expectation map is closed-world against
// them in both directions.
func bidCodecGoFullFromStringRecords() []bidCodecRejectVector {
	var records []bidCodecRejectVector
	seen := map[string]bool{}
	for _, r := range bidCodecRejectVectors() {
		if r.Channel != "from_string" {
			continue
		}
		if r.Input == nil {
			panic("BID codec go_full: from_string reject record without an input field")
		}
		if r.Requires != "" {
			panic(fmt.Sprintf("BID codec go_full: from_string reject record %q carries a requires tag; the go_full channel split assumes the from_string channel is capability-ungated", *r.Input))
		}
		class, ok := bidCodecGoFullFromStringClasses[*r.Input]
		if !ok {
			panic(fmt.Sprintf("BID codec go_full: from_string reject input %q has no expectation class", *r.Input))
		}
		if !bidCodecGoFullClasses[class] {
			panic(fmt.Sprintf("BID codec go_full: from_string input %q maps to unknown class %q", *r.Input, class))
		}
		if seen[*r.Input] {
			panic(fmt.Sprintf("BID codec go_full: duplicate from_string reject input %q cannot key the expectation map", *r.Input))
		}
		seen[*r.Input] = true
		records = append(records, r)
	}
	if len(seen) != len(bidCodecGoFullFromStringClasses) {
		for input := range bidCodecGoFullFromStringClasses {
			if !seen[input] {
				panic(fmt.Sprintf("BID codec go_full: stale from_string expectation for input %q (no such reject record)", input))
			}
		}
	}
	return records
}

// bidCodecGoFullStringVectorRecords returns the string_vectors records in
// emission order after verifying the per-width expectation map is closed-world
// against them in both directions.
func bidCodecGoFullStringVectorRecords() []bidCodecStringVector {
	records := bidCodecStringVectors()
	seen := map[string]bool{}
	for _, sv := range records {
		classes, ok := bidCodecGoFullStringVectorClasses[sv.Input]
		if !ok {
			panic(fmt.Sprintf("BID codec go_full: string_vectors input %q has no expectation classes", sv.Input))
		}
		for _, class := range classes {
			if !bidCodecGoFullClasses[class] {
				panic(fmt.Sprintf("BID codec go_full: string_vectors input %q maps to unknown class %q", sv.Input, class))
			}
		}
		if seen[sv.Input] {
			panic(fmt.Sprintf("BID codec go_full: duplicate string_vectors input %q cannot key the expectation map", sv.Input))
		}
		seen[sv.Input] = true
	}
	if len(seen) != len(bidCodecGoFullStringVectorClasses) {
		for input := range bidCodecGoFullStringVectorClasses {
			if !seen[input] {
				panic(fmt.Sprintf("BID codec go_full: stale string_vectors expectation for input %q (no such record)", input))
			}
		}
	}
	return records
}

// bidCodecGoFullRejectCounts derives the go_full consumed/skipped split from
// the generated reject records: the from_string channel is consumed against
// the public parse surface, every other channel is skipped because bid754-go
// has no public Components construction surface.
func bidCodecGoFullRejectCounts() (consumed, skipped int) {
	consumed = len(bidCodecGoFullFromStringRecords())
	return consumed, len(bidCodecRejectVectors()) - consumed
}

// bidCodecGoFullFromStringClassElems renders the go_full from_string
// expectation map entries `input: class,` in record emission order.
func bidCodecGoFullFromStringClassElems() string {
	var b strings.Builder
	for _, r := range bidCodecGoFullFromStringRecords() {
		fmt.Fprintf(&b, "\n\t%q: %q,", *r.Input, bidCodecGoFullFromStringClasses[*r.Input])
	}
	b.WriteString("\n")
	return b.String()
}

// rustStringLiteral renders s as a Rust string literal.
//
// Go's %q is not usable for Rust source: it escapes a non-printable rune in
// Go's \uXXXX form, which Rust rejects (Rust spells it \u{XXXX}), and the
// from_string reject corpus carries exactly such a rune (U+00A0, the
// non-breaking space row). Printable non-ASCII stays literal because Rust
// source is UTF-8, which keeps the multi-byte rows readable as the characters
// they are testing.
func rustStringLiteral(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r == utf8.RuneError || !unicode.IsPrint(r) {
				fmt.Fprintf(&b, `\u{%x}`, r)
				continue
			}
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// bidCodecRustFullParseFromStringClassElems renders the Rust public-parse
// consumer's expectation rows `(input, class),` in record emission order.
//
// The rows come from the same bidCodecGoFullFromStringClasses table the Go
// consumer uses. Sharing the table is deliberate: the Rust public API is
// generated from the same mechanical port the Go public API routes through, so
// a class that differs between the two languages is a defect, and generating
// both runners from one table makes that divergence a test failure rather than
// a pair of independently maintained expectations that can drift apart.
func bidCodecRustFullParseFromStringClassElems() string {
	var b strings.Builder
	for _, r := range bidCodecGoFullFromStringRecords() {
		fmt.Fprintf(&b, "\n    (%s, %s),", rustStringLiteral(*r.Input), rustStringLiteral(bidCodecGoFullFromStringClasses[*r.Input]))
	}
	b.WriteString("\n")
	return b.String()
}

// bidCodecGoFullStringClassElems renders the go_full string_vectors per-width
// expectation map entries `input: {c32, c64, c128},` in record emission order.
func bidCodecGoFullStringClassElems() string {
	var b strings.Builder
	for _, sv := range bidCodecGoFullStringVectorRecords() {
		classes := bidCodecGoFullStringVectorClasses[sv.Input]
		fmt.Fprintf(&b, "\n\t%q: {%q, %q, %q},", sv.Input, classes[0], classes[1], classes[2])
	}
	b.WriteString("\n")
	return b.String()
}
