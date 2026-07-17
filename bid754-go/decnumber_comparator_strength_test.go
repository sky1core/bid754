package bid754

// Hand-written comparator-strength anchor for the decNumber differential
// gate.
//
// devtools/verification_anchors.json pins the gate's *case counts and
// stream hashes* outside every generation path; this file is the same
// principle applied to comparator *semantics*: it pins how strict the
// generated exact comparator and the fixed decNumber status projection
// (bid754-go/generated_decnumber_differential_shared_test.go, emitted by
// devtools testgen) must be. A relaxation — value-only comparison that
// ignores the cohort, a dropped flag bit, a projection that tolerates an
// unknown status bit — keeps every case count and stream hash identical,
// so no counting gate can see it; only a semantic anchor that calls the
// helpers directly can. It lives outside the generation path on purpose
// (GUARDRAILS principle): a generator regression cannot re-pin it.
//
// The binding test closes the remaining gap: a generator could leave the
// comparator body intact but stop calling it, emitting a looser inline
// comparison in the runner instead. TestDecnumberDiffComparatorBinding
// statically parses the checked-in generated runner and requires
// decnumberDiffCompare to be invoked in a result-consuming decision
// position inside decnumberDiffCheckCase, and decnumberDiffCheckCase to be
// reachable from both differential driver tests.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"math/big"
	"testing"
)

func decnumberDiffStrengthFinite(sign bool, coeff int64, exp int32) decnumberDiffTriple {
	return decnumberDiffTriple{kind: decnumberDiffKindFinite, sign: sign, coeff: big.NewInt(coeff), exp: exp}
}

// TestDecnumberDiffComparatorStrength pins the exact comparison surface:
// finite values compare class, sign, coefficient, and quantum exponent
// (cohort-exact); infinities compare sign; NaN sign and payload stay out of
// the surface but the quiet/signaling class distinction stays in; the
// 5-flag words compare bit-for-bit.
func TestDecnumberDiffComparatorStrength(t *testing.T) {
	base := decnumberDiffStrengthFinite(false, 1230000, -3)
	const baseFlags = decnumberDiffFlagInexact

	if !decnumberDiffCompare(base, baseFlags, decnumberDiffStrengthFinite(false, 1230000, -3), baseFlags) {
		t.Fatalf("comparator rejects an identical (triple, flags) pair")
	}

	mismatches := []struct {
		name   string
		other  decnumberDiffTriple
		flags  uint32
		reason string
	}{
		{"cohort-only", decnumberDiffStrengthFinite(false, 123000, -2), baseFlags,
			"same value in a different cohort must mismatch (preferred-exponent verification is a core target)"},
		{"one-ulp", decnumberDiffStrengthFinite(false, 1230001, -3), baseFlags, "one-ulp coefficient difference"},
		{"sign", decnumberDiffStrengthFinite(true, 1230000, -3), baseFlags, "finite sign difference"},
		{"class-inf", decnumberDiffTriple{kind: decnumberDiffKindInf}, baseFlags, "finite vs infinity"},
		{"class-qnan", decnumberDiffTriple{kind: decnumberDiffKindQNaN}, baseFlags, "finite vs quiet NaN"},
	}
	for _, tc := range mismatches {
		if decnumberDiffCompare(base, baseFlags, tc.other, tc.flags) {
			t.Errorf("comparator accepts %s: %s", tc.name, tc.reason)
		}
	}

	for _, flagBit := range []uint32{
		decnumberDiffFlagInvalid,
		decnumberDiffFlagDivZero,
		decnumberDiffFlagOverflow,
		decnumberDiffFlagUnderflow,
		decnumberDiffFlagInexact,
	} {
		if decnumberDiffCompare(base, baseFlags, decnumberDiffStrengthFinite(false, 1230000, -3), baseFlags^flagBit) {
			t.Errorf("comparator tolerates a single-bit 5-flag difference (bit %#02x)", flagBit)
		}
	}

	// Zero results: the quantum exponent and the zero sign are inside the
	// comparison surface (IEEE zero-sign rules and zero quantum preservation
	// are gate targets).
	zero := decnumberDiffStrengthFinite(false, 0, 2)
	if decnumberDiffCompare(zero, 0, decnumberDiffStrengthFinite(false, 0, 1), 0) {
		t.Errorf("comparator accepts zero results with different quantum exponents")
	}
	if decnumberDiffCompare(zero, 0, decnumberDiffStrengthFinite(true, 0, 2), 0) {
		t.Errorf("comparator accepts zero results with different signs")
	}

	// Infinity sign is compared; NaN sign and payload are projected out of
	// the v1 surface, but the quiet/signaling class distinction is not.
	if decnumberDiffCompare(decnumberDiffTriple{kind: decnumberDiffKindInf}, 0,
		decnumberDiffTriple{kind: decnumberDiffKindInf, sign: true}, 0) {
		t.Errorf("comparator accepts +Inf vs -Inf")
	}
	if !decnumberDiffCompare(decnumberDiffTriple{kind: decnumberDiffKindQNaN}, decnumberDiffFlagInvalid,
		decnumberDiffTriple{kind: decnumberDiffKindQNaN, sign: true}, decnumberDiffFlagInvalid) {
		t.Errorf("comparator rejects qNaN results differing only in the projected-out NaN sign")
	}
	if decnumberDiffCompare(decnumberDiffTriple{kind: decnumberDiffKindQNaN}, 0,
		decnumberDiffTriple{kind: decnumberDiffKindSNaN}, 0) {
		t.Errorf("comparator accepts a quiet vs signaling NaN class difference")
	}
}

// TestDecnumberDiffProjectionStrength pins the fixed decNumber status
// projection table: mapped bits land on their IEEE flags, the three
// GDA-only condition bits are dropped, and every other bit is reported as
// unknown (the runner hard-fails on unknown bits; there is no runtime
// tolerance to widen).
func TestDecnumberDiffProjectionStrength(t *testing.T) {
	mapped := []struct {
		status uint32
		want   uint32
	}{
		{decnumberDiffDnInvalidOperation, decnumberDiffFlagInvalid},
		{decnumberDiffDnDivisionUndefined, decnumberDiffFlagInvalid},
		{decnumberDiffDnDivisionByZero, decnumberDiffFlagDivZero},
		{decnumberDiffDnOverflow, decnumberDiffFlagOverflow},
		{decnumberDiffDnUnderflow, decnumberDiffFlagUnderflow},
		{decnumberDiffDnInexact, decnumberDiffFlagInexact},
	}
	for _, tc := range mapped {
		flags5, unknown := decnumberDiffProjectStatus(tc.status)
		if flags5 != tc.want || unknown != 0 {
			t.Errorf("projection of %#08x = (%#02x, unknown %#08x), want (%#02x, 0)", tc.status, flags5, unknown, tc.want)
		}
	}
	for _, dropped := range []uint32{decnumberDiffDnRounded, decnumberDiffDnClamped, decnumberDiffDnSubnormal} {
		flags5, unknown := decnumberDiffProjectStatus(dropped)
		if flags5 != 0 || unknown != 0 {
			t.Errorf("GDA-only bit %#08x must be dropped, got (%#02x, unknown %#08x)", dropped, flags5, unknown)
		}
	}
	unknownBits := []uint32{
		decnumberDiffDnConversionSyntax,
		decnumberDiffDnDivisionImpossible,
		decnumberDiffDnInsufficientStore,
		decnumberDiffDnInvalidContext,
		0x00000100, // DEC_Lost_digits (DECSUBSET builds)
		0x00004000, // any future bit above the pinned table
	}
	for _, bit := range unknownBits {
		if _, unknown := decnumberDiffProjectStatus(bit); unknown == 0 {
			t.Errorf("status bit %#08x outside the projection table must be reported unknown (hard fail), got unknown=0", bit)
		}
	}
	// Composite: mapped + dropped + unknown decompose exactly.
	flags5, unknown := decnumberDiffProjectStatus(decnumberDiffDnInexact | decnumberDiffDnRounded | decnumberDiffDnDivisionImpossible)
	if flags5 != decnumberDiffFlagInexact || unknown != decnumberDiffDnDivisionImpossible {
		t.Errorf("composite projection = (%#02x, %#08x), want (%#02x, %#08x)",
			flags5, unknown, decnumberDiffFlagInexact, decnumberDiffDnDivisionImpossible)
	}
}

// TestDecnumberDiffComparatorBinding statically requires the checked-in
// generated runner to consume decnumberDiffCompare in a decision position
// inside decnumberDiffCheckCase, and both differential drivers to reach
// decnumberDiffCheckCase. A generated runner that stops calling the
// anchored comparator (emitting a looser inline comparison instead) fails
// here even though every count and hash stays identical.
func TestDecnumberDiffComparatorBinding(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "generated_decnumber_differential_native_test.go", nil, 0)
	if err != nil {
		t.Fatalf("parse generated decNumber differential runner: %v", err)
	}
	functions := map[string]*ast.FuncDecl{}
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil {
			functions[fn.Name.Name] = fn
		}
	}
	checkCase, ok := functions["decnumberDiffCheckCase"]
	if !ok {
		t.Fatalf("generated runner has no decnumberDiffCheckCase function")
	}
	comparatorDecides := false
	ast.Inspect(checkCase, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		unary, ok := ifStmt.Cond.(*ast.UnaryExpr)
		if !ok || unary.Op != token.NOT {
			return true
		}
		call, ok := unary.X.(*ast.CallExpr)
		if !ok {
			return true
		}
		if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "decnumberDiffCompare" {
			comparatorDecides = true
		}
		return true
	})
	if !comparatorDecides {
		t.Errorf("decnumberDiffCheckCase does not consume decnumberDiffCompare in a negated decision position; the anchored comparator is no longer the case verdict")
	}
	for _, driver := range []string{
		"TestGeneratedDecnumberDifferentialStructured",
		"TestGeneratedDecnumberDifferentialDeterministicRandom",
	} {
		fn, ok := functions[driver]
		if !ok {
			t.Errorf("generated runner has no %s driver", driver)
			continue
		}
		reached := false
		ast.Inspect(fn, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "decnumberDiffCheckCase" {
				reached = true
			}
			return true
		})
		if !reached {
			t.Errorf("%s does not reach decnumberDiffCheckCase; the driver no longer routes through the anchored comparison", driver)
		}
	}
}
