package apiemit

import (
	"strings"
	"testing"
)

// TestStaticCheckAcceptsExpectedCode proves the check accepts
// on the kind of code the shape templates actually emit: crate::generated::*
// port calls, crate::gen_types conversions, a depth-2 super:: reference
// (generated::api -> generated), and the mod.rs #![forbid(unsafe_code)]
// attribute (whose "unsafe_code" substring must not trip the \bunsafe\b
// word-boundary check).
func TestStaticCheckAcceptsExpectedCode(t *testing.T) {
	files := map[string]string{
		"decimal64.rs": `
pub fn add(self, rhs: Decimal64) -> Decimal64 {
    Decimal64(crate::generated::add64::bid64_add(self.0, rhs.0, 0))
}
pub(crate) fn bid_uint128_from_le_bytes(bytes: [u8; 16]) -> crate::gen_types::BID_UINT128 {
    crate::gen_types::BID_UINT128 { w: [0, 0] }
}
pub fn c(self) -> DecimalClass {
    super::types::decimal_class_from_bid_class(0)
}
// depth-2 super stays inside crate::generated and must be allowed.
pub fn d() -> u32 { super::super::add64::bid64_add(0, 0, 0) as u32 }
`,
		"mod.rs": `#![forbid(unsafe_code)]
#![deny(arithmetic_overflow, overflowing_literals)]
mod types;
`,
	}
	if err := staticCheckAPIOutput(files); err != nil {
		t.Fatalf("expected generated API code to pass the static check, got: %v", err)
	}
}

// TestStaticCheckReportsUnexpectedCrateModule checks a reference to a
// crate-root module outside the expected implementation-domain set.
func TestStaticCheckReportsUnexpectedCrateModule(t *testing.T) {
	for _, src := range []string{
		`pub fn f() { crate::bid_codec::decode(); }`,
		// gen_constants and tables are registry support modules but are NOT in
		// the narrowed permitted set; a wrapper never needs them through crate::.
		`pub fn f() -> u32 { crate::gen_constants::BID_INVALID }`,
		`pub fn f() -> u32 { crate::tables::LUT[0] }`,
		// whitespace around :: must not hide the segment name.
		`pub fn f() { crate :: bid_codec :: decode(); }`,
	} {
		if err := staticCheckAPIOutput(map[string]string{"decimal64.rs": src}); err == nil {
			t.Errorf("expected a static check rejection for %q, got nil", src)
		}
	}
}

// TestStaticCheckReportsDeepSuperPath checks a super:: chain of depth 3+,
// which leaves generated::api and reaches the crate root.
func TestStaticCheckReportsDeepSuperPath(t *testing.T) {
	for _, src := range []string{
		`pub fn f() { super::super::super::bid_codec::decode(); }`,
		`pub fn f() { super :: super :: super :: bid_codec::decode(); }`,
	} {
		err := staticCheckAPIOutput(map[string]string{"decimal64.rs": src})
		if err == nil {
			t.Errorf("expected a static check failure for deep super path %q, got nil", src)
			continue
		}
		if !strings.Contains(err.Error(), "super:: chain of depth 3") {
			t.Errorf("expected a depth-3 super message for %q, got: %v", src, err)
		}
	}
}

// TestStaticCheckReportsUnsafe proves an actual `unsafe` keyword usage (not
// just the substring inside an identifier) fails generation.
func TestStaticCheckReportsUnsafe(t *testing.T) {
	files := map[string]string{
		"decimal64.rs": `pub fn f() -> u32 { unsafe { core::mem::transmute(0u32) } }`,
	}
	if err := staticCheckAPIOutput(files); err == nil {
		t.Fatal("expected a static check rejection for unsafe, got nil")
	}
}

// TestStaticCheckReportsPanicFamily checks every panic-family macro/method,
// including whitespace variants such as `panic ! (` and `x . expect (`.
func TestStaticCheckReportsPanicFamily(t *testing.T) {
	cases := map[string]string{
		"panic!":         `pub fn f() { panic!("no"); }`,
		"panic! spaced":  `pub fn f() { panic ! ("no"); }`,
		"unwrap(":        `pub fn f(x: Option<u32>) -> u32 { x.unwrap() }`,
		"unwrap( spaced": `pub fn f(x: Option<u32>) -> u32 { x . unwrap () }`,
		"expect(":        `pub fn f(x: Option<u32>) -> u32 { x.expect("no") }`,
		"expect( spaced": `pub fn f(x: Option<u32>) -> u32 { x . expect ("no") }`,
		"unreachable!":   `pub fn f() -> u32 { unreachable!() }`,
		"todo!":          `pub fn f() -> u32 { todo!() }`,
		"unimplemented!": `pub fn f() -> u32 { unimplemented!() }`,
	}
	for label, src := range cases {
		if err := staticCheckAPIOutput(map[string]string{"decimal64.rs": src}); err == nil {
			t.Errorf("case %s: expected a static check rejection, got nil", label)
		}
	}
}
