package apiemit

import (
	"fmt"
	"regexp"
	"sort"
)

// staticCheckAPIOutput validates two properties of the generated api/*.rs
// content before files are written. Generated public wrappers may reference
// the generated implementation, registry-derived types, sibling API modules,
// and the standard library. They also return failures through
// Result/ExceptionFlags rather than process-terminating constructs.
//
// The generated wrapper shape is deliberately narrow, so a text check is both
// sufficient and easier to keep reproducible than a syn/HIR analysis:
//
//  1. Every `crate::…` reference starts with `generated` or `gen_types`. A
//     `super::` chain may have depth one or two, which stays within
//     generated::api / generated; depth three reaches the crate root and is
//     outside the API wrapper subtree.
//  2. The output contains none of `unsafe`, `panic!`, `.unwrap(`, `.expect(`,
//     `unreachable!`, `todo!`, or `unimplemented!`. Patterns accept optional
//     whitespace before `!` or `(`. The `unsafe` expression uses word
//     boundaries so it does not match the `unsafe_code` identifier in the
//     crate attribute.
func staticCheckAPIOutput(files map[string]string) error {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if err := checkOneAPIFile(name, files[name]); err != nil {
			return err
		}
	}
	return nil
}

// expectedCrateFirstSegments is the exhaustive set of crate-root modules an
// `crate::…` reference in api/*.rs may name: the generated implementation
// port and its registry-derived value types. Everything else — the codec
// vector-verification module (crate::bid_codec), the constant/table support
// modules, and any future hand-written module — is outside what a
// routing/plumbing wrapper legitimately reaches through the `crate::` root.
var expectedCrateFirstSegments = map[string]bool{
	"generated": true,
	"gen_types": true,
}

// crateRefPattern captures the first path segment after `crate::`, tolerating
// whitespace around the `::` so formatting does not affect the result.
var crateRefPattern = regexp.MustCompile(`crate\s*::\s*([A-Za-z_][A-Za-z0-9_]*)`)

// deepSuperPathPattern matches a `super::` chain of depth 3 or more. From an
// api/*.rs file (crate::generated::api::*), a depth-3 chain reaches the crate
// root and past it, leaving the generated implementation subtree; depth ≤ 2
// stays inside generated::api / generated and does not match.
var deepSuperPathPattern = regexp.MustCompile(`(?:super\s*::\s*){3,}`)

// unsupportedConstructs are process-terminating constructs outside the public
// API error contract. Each pattern tolerates whitespace before its trailing
// `!`/`(` so formatting variants produce the same result.
// `unsafe` uses word boundaries so it does not match the `unsafe_code`
// identifier in `#![forbid(unsafe_code)]`.
var unsupportedConstructs = []struct {
	pattern *regexp.Regexp
	label   string
}{
	{regexp.MustCompile(`\bunsafe\b`), "unsafe"},
	{regexp.MustCompile(`panic\s*!`), "panic!"},
	{regexp.MustCompile(`\.\s*unwrap\s*\(`), ".unwrap("},
	{regexp.MustCompile(`\.\s*expect\s*\(`), ".expect("},
	{regexp.MustCompile(`unreachable\s*!`), "unreachable!"},
	{regexp.MustCompile(`todo\s*!`), "todo!"},
	{regexp.MustCompile(`unimplemented\s*!`), "unimplemented!"},
}

func checkOneAPIFile(name, code string) error {
	for _, m := range crateRefPattern.FindAllStringSubmatch(code, -1) {
		segment := m[1]
		if !expectedCrateFirstSegments[segment] {
			return fmt.Errorf("apiemit: static check failed for %s: crate::%s is outside the expected implementation modules (crate::generated, crate::gen_types)", name, segment)
		}
	}
	if loc := deepSuperPathPattern.FindString(code); loc != "" {
		return fmt.Errorf("apiemit: static check failed for %s: a super:: chain of depth 3+ (%q) leaves the generated::api subtree; expected depth <= 2", name, loc)
	}
	for _, tok := range unsupportedConstructs {
		if tok.pattern.MatchString(code) {
			return fmt.Errorf("apiemit: static check failed for %s: construct %q does not use the Result/ExceptionFlags error contract", name, tok.label)
		}
	}
	return nil
}
