// Hand-written unreachability certificates for four mutation sites of the
// mechanical port. It lives outside every generation path and must stay
// hand-written.
//
// Why this gate exists. The 2026-07 mutation audit left four mutants that no
// verification corpus can distinguish, because the mutated token sits in a
// region no input reaches:
//
//	add128_inline.go:117  bid_get_add128  bit:<<->>>
//	bid128_add.go:1248    Bid128Add       aor:dec->inc
//	bid128_add.go:1388    Bid128Add       aor:dec->inc
//	div64.go:267          Bid64Div        cmp:==->!=
//
// For a provably dead region no behavioural test can ever kill the mutant, so
// the only verification available is a certificate: a machine-checked proof of
// unreachability plus an anchor that fails the moment the region, or any
// premise the proof rests on, changes. That is what this file is. Every
// certificate below has two halves:
//
//   - a *source anchor*: the dominating guards, the closed world of writes to
//     the variables the proof reasons about, and the dead region itself are
//     pinned against the port's own AST. A hand edit, a mutation, or a
//     refactor inside the region fails the pin instead of silently passing.
//     This is deliberate: pinning is the fail-closed half, and without it a
//     numeric bound recomputed from the port's tables would keep passing after
//     the call site it describes had moved on.
//   - a *premise check*: the numeric facts the proof needs, verified over a
//     closed enumeration of the port's own tables and helpers rather than a
//     restatement of the algorithm's intent.
//
// One premise (the double-precision quotient estimate in Bid64Div, see
// bid64DivEstimateSlack) is an IEEE-754 property of the host, not a table
// fact; it is checked over a boundary plus seeded corpus and is called out as
// the single sampled premise in this file.
//
// Maintenance note: the anchors compare gofmt renderings with whitespace
// collapsed, so line rewrapping is invisible to them but a token change is not.
// Any intended edit inside a certified region - including one that follows from
// a toolchain formatting change - has to update the pinned text here and, with
// it, re-justify the unreachability argument. That coupling is the point.
package bidgo

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"math"
	"math/big"
	"math/bits"
	"math/rand"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// source anchor layer
// ---------------------------------------------------------------------------

// portFile is one parsed file of the port, read from the package directory the
// test runs in, so the anchors always describe the tree under test.
type portFile struct {
	name string
	fset *token.FileSet
	file *ast.File
}

func loadPortFile(t *testing.T, name string) *portFile {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return &portFile{name: name, fset: fset, file: f}
}

// flat renders a node with gofmt and collapses its whitespace to single
// spaces, so a pin survives line rewrapping but not a token change.
func (p *portFile) flat(t *testing.T, n ast.Node) string {
	t.Helper()
	var buf bytes.Buffer
	if err := format.Node(&buf, p.fset, n); err != nil {
		t.Fatalf("%s: render node at %s: %v", p.name, p.fset.Position(n.Pos()), err)
	}
	return strings.Join(strings.Fields(buf.String()), " ")
}

func (p *portFile) line(n ast.Node) int { return p.fset.Position(n.Pos()).Line }

func (p *portFile) funcDecl(t *testing.T, name string) *ast.FuncDecl {
	t.Helper()
	for _, d := range p.file.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == name && fd.Body != nil {
			return fd
		}
	}
	t.Fatalf("%s: function %s not found", p.name, name)
	return nil
}

// blockWithRun returns the single block inside root whose statement list holds
// want as a contiguous run of flattened statements. Requiring exactly one match
// is what makes the run both the locator and the pin: if any statement in the
// run changes, or the run stops being contiguous, or a second copy appears, the
// anchor fails.
func (p *portFile) blockWithRun(t *testing.T, root ast.Node, want []string) *ast.BlockStmt {
	t.Helper()
	if len(want) == 0 {
		t.Fatal("blockWithRun needs a non-empty run")
	}
	var hits []*ast.BlockStmt
	ast.Inspect(root, func(n ast.Node) bool {
		b, ok := n.(*ast.BlockStmt)
		if !ok || len(b.List) < len(want) {
			return true
		}
		for start := 0; start+len(want) <= len(b.List); start++ {
			match := true
			for i, w := range want {
				if p.flat(t, b.List[start+i]) != w {
					match = false
					break
				}
			}
			if match {
				hits = append(hits, b)
				break
			}
		}
		return true
	})
	if len(hits) != 1 {
		t.Fatalf("%s: expected exactly one block carrying the pinned %d-statement run, found %d\nfirst pinned statement: %s",
			p.name, len(want), len(hits), want[0])
	}
	return hits[0]
}

// pinRun asserts that blk's statements, starting at index start, are exactly
// the flattened statements in want.
func (p *portFile) pinRun(t *testing.T, label string, blk *ast.BlockStmt, start int, want []string) {
	t.Helper()
	if start < 0 || start+len(want) > len(blk.List) {
		t.Fatalf("%s: %s: block has %d statements, pinned run needs [%d,%d)",
			p.name, label, len(blk.List), start, start+len(want))
	}
	for i, w := range want {
		if got := p.flat(t, blk.List[start+i]); got != w {
			t.Fatalf("%s: %s: statement %d (line %d) drifted\n got: %s\nwant: %s",
				p.name, label, start+i, p.line(blk.List[start+i]), got, w)
		}
	}
}

func (p *portFile) pinNode(t *testing.T, label string, n ast.Node, want string) {
	t.Helper()
	if got := p.flat(t, n); got != want {
		t.Fatalf("%s: %s (line %d) drifted\n got: %s\nwant: %s", p.name, label, p.line(n), got, want)
	}
}

func (p *portFile) pinStmtCount(t *testing.T, label string, blk *ast.BlockStmt, want int) {
	t.Helper()
	if len(blk.List) != want {
		t.Fatalf("%s: %s: block (line %d) has %d statements, pinned %d",
			p.name, label, p.line(blk), len(blk.List), want)
	}
}

func (p *portFile) stmtAt(t *testing.T, label string, blk *ast.BlockStmt, i int) ast.Stmt {
	t.Helper()
	if i < 0 || i >= len(blk.List) {
		t.Fatalf("%s: %s: no statement %d in block of %d (line %d)", p.name, label, i, len(blk.List), p.line(blk))
	}
	return blk.List[i]
}

func (p *portFile) asIf(t *testing.T, label string, s ast.Stmt) *ast.IfStmt {
	t.Helper()
	is, ok := s.(*ast.IfStmt)
	if !ok {
		t.Fatalf("%s: %s: statement at line %d is %T, want an if statement", p.name, label, p.line(s), s)
	}
	return is
}

func (p *portFile) elseBlock(t *testing.T, label string, is *ast.IfStmt) *ast.BlockStmt {
	t.Helper()
	b, ok := is.Else.(*ast.BlockStmt)
	if !ok {
		t.Fatalf("%s: %s: if at line %d has else of type %T, want a block", p.name, label, p.line(is), is.Else)
	}
	return b
}

func (p *portFile) elseIf(t *testing.T, label string, is *ast.IfStmt) *ast.IfStmt {
	t.Helper()
	e, ok := is.Else.(*ast.IfStmt)
	if !ok {
		t.Fatalf("%s: %s: if at line %d has else of type %T, want an else-if", p.name, label, p.line(is), is.Else)
	}
	return e
}

// portWrite is one statement that can change target's value.
type portWrite struct {
	line int
	text string
}

// writes returns, in source order, every statement inside root that assigns to
// target, increments or decrements it, or takes its address. Recording address
// operators matters: without them a later `f(&v)` could change the value the
// certificate reasons about while the pinned assignment list stayed intact.
func (p *portFile) writes(t *testing.T, root ast.Node, target string) []portWrite {
	t.Helper()
	var out []portWrite
	add := func(n ast.Node, text string) {
		out = append(out, portWrite{line: p.line(n), text: text})
	}
	ast.Inspect(root, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range s.Lhs {
				if rootIdentName(lhs) == target {
					add(s, p.flat(t, s))
					break
				}
			}
		case *ast.IncDecStmt:
			if rootIdentName(s.X) == target {
				add(s, p.flat(t, s))
			}
		case *ast.ValueSpec:
			for _, name := range s.Names {
				if name.Name == target && len(s.Values) > 0 {
					add(s, p.flat(t, s))
					break
				}
			}
		case *ast.UnaryExpr:
			if s.Op == token.AND && rootIdentName(s.X) == target {
				add(s, "ADDRESS-TAKEN "+p.flat(t, s))
			}
		}
		return true
	})
	return out
}

// pinWrites asserts the closed world of writes to target inside root.
func (p *portFile) pinWrites(t *testing.T, root ast.Node, target string, want []string) []portWrite {
	t.Helper()
	got := p.writes(t, root, target)
	if len(got) != len(want) {
		t.Fatalf("%s: writes to %s: found %d, pinned %d\ngot: %v", p.name, target, len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].text != w {
			t.Fatalf("%s: write %d to %s (line %d) drifted\n got: %s\nwant: %s",
				p.name, i, target, got[i].line, got[i].text, w)
		}
	}
	return got
}

// noWriteBetween asserts nothing inside root changes target strictly between
// the end of from and the start of to.
func (p *portFile) noWriteBetween(t *testing.T, root ast.Node, target string, from, to ast.Node) {
	t.Helper()
	lo, hi := from.End(), to.Pos()
	ast.Inspect(root, func(n ast.Node) bool {
		var pos token.Pos
		var text string
		switch s := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range s.Lhs {
				if rootIdentName(lhs) == target {
					pos, text = s.Pos(), p.flat(t, s)
				}
			}
		case *ast.IncDecStmt:
			if rootIdentName(s.X) == target {
				pos, text = s.Pos(), p.flat(t, s)
			}
		case *ast.UnaryExpr:
			if s.Op == token.AND && rootIdentName(s.X) == target {
				pos, text = s.Pos(), "ADDRESS-TAKEN "+p.flat(t, s)
			}
		}
		if pos != token.NoPos && pos >= lo && pos < hi {
			t.Fatalf("%s: %s is written at line %d, between the pinned anchors (lines %d..%d): %s",
				p.name, target, p.fset.Position(pos).Line,
				p.fset.Position(lo).Line, p.fset.Position(hi).Line, text)
		}
		return true
	})
}

func rootIdentName(e ast.Expr) string {
	for {
		switch x := e.(type) {
		case *ast.Ident:
			return x.Name
		case *ast.SelectorExpr:
			e = x.X
		case *ast.IndexExpr:
			e = x.X
		case *ast.StarExpr:
			e = x.X
		case *ast.ParenExpr:
			e = x.X
		default:
			return ""
		}
	}
}

// ---------------------------------------------------------------------------
// small exact helpers
// ---------------------------------------------------------------------------

func pow10big(n int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}

func pow2big(n int) *big.Int {
	return new(big.Int).Lsh(big.NewInt(1), uint(n))
}

func decDigits(v *big.Int) int {
	if v.Sign() == 0 {
		return 1
	}
	return len(v.String())
}

// ---------------------------------------------------------------------------
// shared table premises
// ---------------------------------------------------------------------------

// TestPortDigitEstimateTablesAreExact pins the two tables every digit-count
// normalisation in the port derives its scale from:
//
//	bid_estimate_decimal_digits[b] is the exact decimal digit count of 2^b, and
//	bid_power10_index_binexp[b]    is 10^bid_estimate_decimal_digits[b].
//
// Together these give the exact digit count of any v with binary exponent b:
// because 2^(b+1) < 10*2^b, the interval [2^b, 2^(b+1)) spans at most one
// decade boundary, and that boundary is exactly 10^est[b]. Both the
// add128_inline and the div64 certificate rest on this, so it is checked once.
func TestPortDigitEstimateTablesAreExact(t *testing.T) {
	if len(bid_power10_table_128) < 39 {
		t.Fatalf("bid_power10_table_128 has %d entries, expected at least 39", len(bid_power10_table_128))
	}
	for k := 0; k <= 19; k++ {
		want := pow10big(k)
		got := new(big.Int).SetUint64(bid_power10_table_128[k].lo)
		if bid_power10_table_128[k].hi != 0 || got.Cmp(want) != 0 {
			t.Fatalf("bid_power10_table_128[%d] = %016x%016x, want 10^%d",
				k, bid_power10_table_128[k].hi, bid_power10_table_128[k].lo, k)
		}
	}
	for b := 0; b < len(bid_estimate_decimal_digits); b++ {
		est := bid_estimate_decimal_digits[b]
		p := pow2big(b)
		if got := decDigits(p); got != est {
			t.Fatalf("bid_estimate_decimal_digits[%d] = %d, but 2^%d has %d decimal digits", b, est, b, got)
		}
		// The decade boundary inside [2^b, 2^(b+1)) is 10^est[b].
		if pow10big(est-1).Cmp(p) > 0 || pow10big(est).Cmp(p) <= 0 {
			t.Fatalf("bid_estimate_decimal_digits[%d] = %d does not bracket 2^%d in [10^%d, 10^%d)", b, est, b, est-1, est)
		}
	}
	for b := 0; b < len(bid_power10_index_binexp); b++ {
		if b >= len(bid_estimate_decimal_digits) {
			t.Fatalf("bid_power10_index_binexp is longer (%d) than bid_estimate_decimal_digits (%d)",
				len(bid_power10_index_binexp), len(bid_estimate_decimal_digits))
		}
		want := pow10big(bid_estimate_decimal_digits[b])
		if new(big.Int).SetUint64(bid_power10_index_binexp[b]).Cmp(want) != 0 {
			t.Fatalf("bid_power10_index_binexp[%d] = %d, want 10^%d = %s",
				b, bid_power10_index_binexp[b], bid_estimate_decimal_digits[b], want)
		}
	}
}

// ---------------------------------------------------------------------------
// certificate 1: add128_inline.go:115-119 (bid_get_add128)
// ---------------------------------------------------------------------------

// TestGetAdd128SmallCoefficientArmIsUnreachable certifies that the arm
//
//	if coefficient_x < 1000000000000000 {
//	        coefficient_x -= uint64(D)
//	        coefficient_x = uint64(D) + (coefficient_x << 1) + (coefficient_x << 3)
//	        exponent_x--
//	}
//
// in bid_get_add128 is dead, which is why the mutation of either shift on the
// middle line (audit site add128_inline.go:3342) cannot be observed.
//
// Proof. On entry to the guard, coefficient_x has been normalised to exactly 16
// decimal digits and, when the operand signs differ, bumped off the low decade
// edge:
//
//	extra_dx = 16 - digits_x
//	coefficient_x *= 10^extra_dx                     -> [10^15, 10^16)
//	if signs differ && coefficient_x == 10^15 { coefficient_x = 10^16 }
//
// Between that normalisation and the guard the only writes to coefficient_x are
// the three `coefficient_x += uint64(D)` statements of the pinned rounding
// switch, and in every one of them D == -1 requires the two signs to differ:
//
//	BID_ROUNDING_UP     guarded by sign_y == 0,     so D == -1 needs sign_x != 0
//	BID_ROUNDING_DOWN   guarded by sign_y != 0,     so D == -1 needs sign_x == 0
//	BID_ROUNDING_TO_ZERO guarded by sign_y != sign_x, where D is the literal -1
//
// So the value reaching the guard is >= 10^15 in every case: equal signs keep
// coefficient_x >= 10^15 and can only add +1, differing signs guarantee
// coefficient_x > 10^15 before a -1 is applied, and the empty default arm
// leaves coefficient_x untouched. The guard is therefore never taken.
func TestGetAdd128SmallCoefficientArmIsUnreachable(t *testing.T) {
	p := loadPortFile(t, "add128_inline.go")
	fn := p.funcDecl(t, "bid_get_add128")

	// --- anchor: closed world of writes to the two variables the proof reads.
	p.pinWrites(t, fn, "coefficient_x", []string{
		"coefficient_x *= bid_power10_table_128[extra_dx].lo",
		"coefficient_x = 10000000000000000",
		"coefficient_x += uint64(D)",
		"coefficient_x += uint64(D)",
		"coefficient_x += uint64(D)",
		"coefficient_x -= uint64(D)",
		"coefficient_x = uint64(D) + (coefficient_x << 1) + (coefficient_x << 3)",
	})
	p.pinWrites(t, fn, "D", []string{
		"D = int64(sign_x ^ sign_y)",
		"D >>= 63",
		"D = D + D + 1",
		"D = int64(sign_x ^ sign_y)",
		"D >>= 63",
		"D = D + D + 1",
		"D = 0 - 1",
	})

	// --- anchor: the 16-digit normalisation, as one contiguous run.
	normBlock := p.blockWithRun(t, fn, []string{
		"tempx := math.Float64bits(float64(coefficient_x))",
		"bin_expon_cx = int((tempx&MASK_BINARY_EXPONENT)>>52) - 0x3ff",
		"digits_x = bid_estimate_decimal_digits[bin_expon_cx]",
		"if coefficient_x >= bid_power10_table_128[digits_x].lo { digits_x++ }",
		"extra_dx = 16 - digits_x",
		"coefficient_x *= bid_power10_table_128[extra_dx].lo",
		"if (sign_x^sign_y) != 0 && (coefficient_x == 1000000000000000) { extra_dx++ coefficient_x = 10000000000000000 }",
		"exponent_x -= extra_dx",
	})

	// --- anchor: the rounding switch and the dead guard, adjacent.
	const pinnedSwitch = "switch rounding_mode { " +
		"case BID_ROUNDING_UP: if sign_y == 0 { D = int64(sign_x ^ sign_y) D >>= 63 D = D + D + 1 coefficient_x += uint64(D) } " +
		"case BID_ROUNDING_DOWN: if sign_y != 0 { D = int64(sign_x ^ sign_y) D >>= 63 D = D + D + 1 coefficient_x += uint64(D) } " +
		"case BID_ROUNDING_TO_ZERO: if sign_y != sign_x { D = 0 - 1 coefficient_x += uint64(D) } " +
		"default: }"
	const pinnedDeadArm = "if coefficient_x < 1000000000000000 { " +
		"coefficient_x -= uint64(D) " +
		"coefficient_x = uint64(D) + (coefficient_x << 1) + (coefficient_x << 3) " +
		"exponent_x-- }"
	switchBlock := p.blockWithRun(t, fn, []string{pinnedSwitch, pinnedDeadArm})
	p.pinStmtCount(t, "rounding-correction block", switchBlock, 2)
	// The block is the body of the directed-rounding guard, so the empty
	// default arm is only reachable for rounding modes outside the three cases
	// - and those leave coefficient_x untouched either way.
	var roundingGuard *ast.IfStmt
	ast.Inspect(fn, func(n ast.Node) bool {
		if is, ok := n.(*ast.IfStmt); ok && is.Body == switchBlock {
			roundingGuard = is
		}
		return true
	})
	if roundingGuard == nil {
		t.Fatal("add128_inline.go: the rounding-correction block is not an if body")
	}
	p.pinNode(t, "directed-rounding guard", roundingGuard.Cond, "(rounding_mode & 3) != 0")
	if p.line(switchBlock) < p.line(normBlock) {
		t.Fatalf("add128_inline.go: the rounding switch (line %d) precedes the normalisation block (line %d); the certificate's ordering premise is broken",
			p.line(switchBlock), p.line(normBlock))
	}

	// --- premise: the D idiom yields -1 exactly when the two signs differ.
	for _, sx := range []uint64{0, 0x8000000000000000} {
		for _, sy := range []uint64{0, 0x8000000000000000} {
			d := int64(sx ^ sy)
			d >>= 63
			d = d + d + 1
			differ := sx != sy
			if differ && d != -1 {
				t.Fatalf("D idiom: sign_x=%016x sign_y=%016x differ but D = %d, want -1", sx, sy, d)
			}
			if !differ && d != 1 {
				t.Fatalf("D idiom: sign_x=%016x sign_y=%016x equal but D = %d, want +1", sx, sy, d)
			}
		}
	}

	// --- premise: the normalisation lands coefficient_x in [10^15, 10^16).
	// digits_x is the exact digit count (TestPortDigitEstimateTablesAreExact),
	// so it is enough to check the extremes of every digit class.
	lo15, hi16 := pow10big(15), pow10big(16)
	for d := 1; d <= 16; d++ {
		scale := pow10big(16 - d)
		if got := new(big.Int).Mul(pow10big(d-1), scale); got.Cmp(lo15) != 0 {
			t.Fatalf("normalisation: smallest %d-digit coefficient scales to %s, want 10^15", d, got)
		}
		maxCoeff := new(big.Int).Sub(pow10big(d), big.NewInt(1))
		if got := new(big.Int).Mul(maxCoeff, scale); got.Cmp(hi16) >= 0 || got.Cmp(lo15) < 0 {
			t.Fatalf("normalisation: largest %d-digit coefficient scales to %s, outside [10^15, 10^16)", d, got)
		}
	}

	// --- premise: out-of-contract coefficients trap in the normalisation
	// rather than reaching the guard, so they cannot revive the arm.
	zeroBinExpon := int((math.Float64bits(float64(uint64(0)))&MASK_BINARY_EXPONENT)>>52) - 0x3ff
	if zeroBinExpon >= 0 {
		t.Fatalf("coefficient_x == 0 yields bin_expon_cx = %d; the certificate needs it to index bid_estimate_decimal_digits out of range", zeroBinExpon)
	}
	for _, cx := range []uint64{10000000000000000, 1 << 54} {
		digits := portDigitCount64(cx)
		if 16-digits >= 0 {
			t.Fatalf("coefficient_x = %d yields digits_x = %d and extra_dx = %d; the certificate needs a negative index there",
				cx, digits, 16-digits)
		}
	}
}

// portDigitCount64 reproduces the port's digit-count idiom (the four pinned
// statements of the normalisation run) so the out-of-contract premise is
// evaluated with the port's own tables.
func portDigitCount64(coefficient uint64) int {
	tempx := math.Float64bits(float64(coefficient))
	binExpon := int((tempx&MASK_BINARY_EXPONENT)>>52) - 0x3ff
	digits := bid_estimate_decimal_digits[binExpon]
	if coefficient >= bid_power10_table_128[digits].lo {
		digits++
	}
	return digits
}

// ---------------------------------------------------------------------------
// certificates 2 and 3: bid128_add.go delta == P34, C1 == 10^(q1-1) branch
// ---------------------------------------------------------------------------

// bid128AddSpecialBranch anchors the `delta == P34 && C1 == 10^(q1-1) &&
// x_sign != y_sign` block of Bid128Add, which carries both bid128_add
// certificates, and returns the pinned pieces they descend into.
type bid128AddSpecialBranch struct {
	p            *portFile
	fn           *ast.FuncDecl
	blk          *ast.BlockStmt // the branch body
	roundC2      *ast.IfStmt    // if ind >= 0 { ... }
	highf2star   *ast.IfStmt    // the highf2star ind-arm chain
	inexactChain *ast.IfStmt    // the inexactness ind-arm chain
	correction   *ast.IfStmt    // if rnd_mode != BID_ROUNDING_TO_NEAREST { ... }
}

const pinnedC1ScaleDispatch = "if scale >= 20 { C1 = __mul_128x64_to_128(C1_lo, bid_ten2k128[scale-20]) } " +
	"else { if q1 <= 19 { C1 = __mul_64x64_to_128(C1_lo, bid_ten2k64[scale]) } " +
	"else { C1.hi = C1_hi C1.lo = C1_lo C1 = __mul_128x64_to_128(bid_ten2k64[scale], C1) } }"

const pinnedHighf2starChain = "if ind <= 2 { highf2star.hi = 0x0 highf2star.lo = 0x0 } " +
	"else if ind <= 21 { highf2star.hi = 0x0 highf2star.lo = R256.w2 & bid_maskhigh128[ind] } " +
	"else { highf2star.hi = R256.w3 & bid_maskhigh128[ind] highf2star.lo = R256.w2 }"

func loadBid128AddSpecialBranch(t *testing.T) *bid128AddSpecialBranch {
	t.Helper()
	p := loadPortFile(t, "bid128_add.go")
	fn := p.funcDecl(t, "Bid128Add")

	// The branch is located, and simultaneously pinned, by the run that sets
	// up x1, the P34+1-q1 scale, the C1 scaling dispatch and tmp64.
	blk := p.blockWithRun(t, fn, []string{
		"x1 = q2 - 1",
		"scale = 34 - q1 + 1",
		pinnedC1ScaleDispatch,
		"tmp64 = C1.lo",
		"ind = x1 - 1",
	})

	// The enclosing if selects this block as its else, and its condition is the
	// negation of "x_sign == y_sign or C1 != 10^(q1-1)".
	var parent *ast.IfStmt
	ast.Inspect(fn, func(n ast.Node) bool {
		is, ok := n.(*ast.IfStmt)
		if ok && is.Else == ast.Stmt(blk) {
			parent = is
		}
		return true
	})
	if parent == nil {
		t.Fatal("bid128_add.go: the pinned branch is not the else of an if statement")
	}
	p.pinNode(t, "delta==P34 special-branch guard", parent.Cond,
		"x_sign == y_sign || (q1 <= 20 && (C1_hi != 0 || C1_lo != bid_ten2k64[q1-1])) "+
			"|| (q1 >= 21 && (C1_hi != bid_ten2k128[q1-21].hi || C1_lo != bid_ten2k128[q1-21].lo))")

	b := &bid128AddSpecialBranch{p: p, fn: fn, blk: blk}
	b.roundC2 = p.asIf(t, "round-C2 arm", p.stmtAt(t, "branch body", blk, 5))
	p.pinNode(t, "round-C2 arm guard", b.roundC2.Cond, "ind >= 0")

	body := b.roundC2.Body
	p.pinRun(t, "round-C2 prologue", body, 0, []string{
		"C2.lo = C2_lo",
		"C2.hi = C2_hi",
		"if ind <= 18 { C2.lo = C2.lo + bid_midpoint64[ind] if C2.lo < C2_lo { C2.hi++ } } " +
			"else { C2.lo = C2.lo + bid_midpoint128[ind-19].lo C2.hi = C2.hi + bid_midpoint128[ind-19].hi if C2.lo < C2_lo { C2.hi++ } }",
		"R256 = __mul_128x128_to_256(C2, bid_ten2mk128[ind])",
		pinnedHighf2starChain,
		"if ind >= 3 { shift = bid_shiftright128[ind] " +
			"if shift < 64 { R256.w2 = (R256.w2 >> uint(shift)) | (R256.w3 << uint(64-shift)) R256.w3 = (R256.w3 >> uint(shift)) } " +
			"else { R256.w2 = (R256.w3 >> uint(shift-64)) R256.w3 = 0x0 } }",
		"is_inexact_lt_midpoint = 0",
		"is_inexact_gt_midpoint = 0",
		"is_midpoint_lt_even = 0",
		"is_midpoint_gt_even = 0",
	})
	b.highf2star = p.asIf(t, "highf2star chain", p.stmtAt(t, "round-C2 body", body, 4))
	b.inexactChain = p.asIf(t, "inexactness chain", p.stmtAt(t, "round-C2 body", body, 10))
	p.pinNode(t, "inexactness chain first arm", b.inexactChain.Cond, "ind <= 2")

	b.correction = p.asIf(t, "RN-to-directed correction", p.stmtAt(t, "branch body", blk, 14))
	p.pinNode(t, "RN-to-directed correction guard", b.correction.Cond, "rnd_mode != BID_ROUNDING_TO_NEAREST")
	return b
}

// TestBid128AddHighF2StarBorrowIsUnreachable certifies that the borrow arm
//
//	if tmp64A > highf2star.lo {
//	        tmp64B--
//	}
//
// of the `3 <= ind <= 21` inexactness arm (audit site bid128_add.go:48307) is
// dead, so the mutation of that decrement cannot be observed.
//
// Proof, entirely structural. In the `3 <= ind <= 21` arm of the highf2star
// chain the port assigns the literal `highf2star.hi = 0x0`, and nothing writes
// highf2star again before the inexactness chain reads it (both chains branch on
// the same unmodified ind, so they select the same arm). With highf2star.hi
// == 0 the arm's guard reduces to its second and third disjuncts, each of which
// requires highf2star.lo >= bid_onehalf128[ind]. The arm then computes
// tmp64A = highf2star.lo - bid_onehalf128[ind] with unsigned arithmetic, which
// for lo >= onehalf does not borrow, so tmp64A <= highf2star.lo and the guard
// `tmp64A > highf2star.lo` is false.
func TestBid128AddHighF2StarBorrowIsUnreachable(t *testing.T) {
	b := loadBid128AddSpecialBranch(t)
	p := b.p

	// --- anchor: the ind <= 21 arm of the highf2star chain zeroes .hi.
	p.pinNode(t, "highf2star ind-arm chain", b.highf2star, pinnedHighf2starChain)
	highArm := p.elseIf(t, "highf2star ind<=21 arm", b.highf2star)
	p.pinNode(t, "highf2star ind<=21 guard", highArm.Cond, "ind <= 21")
	p.pinRun(t, "highf2star ind<=21 body", highArm.Body, 0, []string{
		"highf2star.hi = 0x0",
		"highf2star.lo = R256.w2 & bid_maskhigh128[ind]",
	})
	p.pinStmtCount(t, "highf2star ind<=21 body", highArm.Body, 2)

	// --- anchor: the matching inexactness arm and the dead borrow.
	inexactArm := p.elseIf(t, "inexactness ind<=21 arm", b.inexactChain)
	p.pinNode(t, "inexactness ind<=21 guard", inexactArm.Cond, "ind <= 21")
	gt := p.asIf(t, "f2* > 1/2 test", p.stmtAt(t, "inexactness ind<=21 body", inexactArm.Body, 0))
	p.pinNode(t, "f2* > 1/2 condition", gt.Cond,
		"highf2star.hi > 0x0 || (highf2star.hi == 0x0 && highf2star.lo > bid_onehalf128[ind]) "+
			"|| (highf2star.hi == 0x0 && highf2star.lo == bid_onehalf128[ind] && (R256.w1 != 0 || R256.w0 != 0))")
	p.pinRun(t, "f2* - 1/2 body", gt.Body, 0, []string{
		"tmp64A = highf2star.lo - bid_onehalf128[ind]",
		"tmp64B = highf2star.hi",
		"if tmp64A > highf2star.lo { tmp64B-- }",
	})

	// --- anchor: highf2star and ind are untouched between the two chains.
	p.noWriteBetween(t, b.fn, "highf2star", b.highf2star, gt)
	p.noWriteBetween(t, b.fn, "ind", b.highf2star, gt)

	// --- premise: unsigned subtraction with lo >= onehalf cannot borrow, so
	// the guard's `tmp64A > highf2star.lo` shape is unsatisfiable. Checked over
	// the port's own bid_onehalf128 rows for the arm's whole ind range,
	// including the extremal lo values the arm can present.
	for ind := 3; ind <= 21; ind++ {
		half := bid_onehalf128[ind]
		for _, lo := range []uint64{half, half + 1, ^uint64(0), half | (half - 1)} {
			if lo < half {
				continue
			}
			if tmp64A := lo - half; tmp64A > lo {
				t.Fatalf("ind=%d lo=%016x onehalf=%016x: lo-onehalf = %016x borrowed; the borrow arm stops being dead",
					ind, lo, half, tmp64A)
			}
		}
		if half == 0 {
			t.Fatalf("ind=%d: bid_onehalf128[ind] is zero, so highf2star.lo >= onehalf no longer excludes a borrow", ind)
		}
	}
}

// TestBid128AddSubtractOneLowWordBorrowIsUnreachable certifies that the borrow
// arm
//
//	if C1_lo == 0xffffffffffffffff {
//	        C1_hi--
//	}
//
// of the `C1 = C1 - 1` directed-rounding correction (audit site
// bid128_add.go:53301) is dead, so the mutation of that decrement cannot be
// observed.
//
// Proof. The borrow needs C1_lo == 0 before the decrement, i.e. the 128-bit
// difference assembled a few lines earlier must be an exact multiple of 2^64.
// In this branch it cannot be:
//
//	C1 == 10^(q1-1) is the branch condition, and the scale the port applies is
//	scale = P34 + 1 - q1, so the scaled left operand is exactly
//	10^(q1-1) * 10^(35-q1) = 10^34 for every q1, giving
//	tmp64 = low64(10^34) = 0x378d8e6400000000.
//
//	The right operand is C2 rounded to q2 - x1 = 1 decimal digit, so R256.w3 is
//	zero and R256.w2 is a single-digit value (at most 11 including the midpoint
//	carry, and at least 1 because C2 >= 10^(q2-1) = 10^x1).
//
// Hence C1.lo = 0x378d8e6400000000 - R256.w2 lies in
// [0x378d8e63fffffff5, 0x378d8e63ffffffff] and is never zero; the two's
// complement negation of the following sign fixup maps zero to zero and so
// cannot introduce one either. C1_lo - 1 therefore never wraps.
func TestBid128AddSubtractOneLowWordBorrowIsUnreachable(t *testing.T) {
	b := loadBid128AddSpecialBranch(t)
	p := b.p

	// --- anchor: the difference, the sign fixup and the C1_lo/C1_hi handoff.
	p.pinRun(t, "P34 difference", b.blk, 6, []string{
		"C1.lo = C1.lo - R256.w2",
		"C1.hi = C1.hi - R256.w3",
		"if C1.lo > tmp64 { C1.hi-- }",
		"if C1.hi >= 0x8000000000000000 { C1.lo = ^C1.lo C1.lo++ C1.hi = ^C1.hi if C1.lo == 0x0 { C1.hi++ } tmp_sign = y_sign } else { tmp_sign = x_sign }",
		"x_sign = tmp_sign",
		"if x1 >= 1 { y_exp = y_exp + (uint64(x1) << 49) }",
		"C1_hi = C1.hi",
		"C1_lo = C1.lo",
	})

	// --- anchor: the dead borrow inside the "C1 = C1 - 1" arm.
	bigIf := p.asIf(t, "directed-rounding split", p.stmtAt(t, "correction body", b.correction.Body, 0))
	minusOne := p.elseIf(t, "C1 = C1 - 1 arm", bigIf)
	p.pinNode(t, "C1 = C1 - 1 guard", minusOne.Cond,
		"(is_midpoint_lt_even != 0 || is_inexact_gt_midpoint != 0) && "+
			"((x_sign != 0 && (rnd_mode == BID_ROUNDING_UP || rnd_mode == BID_ROUNDING_TO_ZERO)) || "+
			"(x_sign == 0 && (rnd_mode == BID_ROUNDING_DOWN || rnd_mode == BID_ROUNDING_TO_ZERO)))")
	p.pinRun(t, "C1 = C1 - 1 body", minusOne.Body, 0, []string{
		"C1_lo = C1_lo - 1",
		"if C1_lo == 0xffffffffffffffff { C1_hi-- }",
		"if C1_hi == 0x0000314dc6448d93 && C1_lo == 0x38c15b09ffffffff { C1_hi = 0x0001ed09bead87c0 C1_lo = 0x378d8e63ffffffff y_exp = y_exp - EXP_P1 }",
	})
	p.pinStmtCount(t, "C1 = C1 - 1 body", minusOne.Body, 3)

	// --- anchor: the midpoint fixup is the only other writer of R256.w2 after
	// the rounding, and it can only decrement it by one.
	p.pinNode(t, "midpoint fixup", p.stmtAt(t, "round-C2 body", b.roundC2.Body, 11),
		"if (R256.w1 != 0 || R256.w0 != 0) && (highf2star.hi == 0) && (highf2star.lo == 0) && "+
			"(R256.w1 < bid_ten2mk128trunc[ind].hi || (R256.w1 == bid_ten2mk128trunc[ind].hi && R256.w0 <= bid_ten2mk128trunc[ind].lo)) { "+
			"if (tmp64+R256.w2)&0x01 != 0 { R256.w2-- if R256.w2 == 0xffffffffffffffff { R256.w3-- } "+
			"is_midpoint_lt_even = 1 is_inexact_lt_midpoint = 0 is_inexact_gt_midpoint = 0 } "+
			"else { is_midpoint_gt_even = 1 is_inexact_lt_midpoint = 0 is_inexact_gt_midpoint = 0 } }")
	// --- anchor: the ind == -1 (x1 == 0) arm feeds R256 straight from C2.
	p.pinNode(t, "ind == -1 arm", p.elseBlock(t, "round-C2 else", b.roundC2),
		"{ R256.w2 = C2_lo R256.w3 = C2_hi is_midpoint_lt_even = 0 is_midpoint_gt_even = 0 "+
			"is_inexact_lt_midpoint = 0 is_inexact_gt_midpoint = 0 }")
	// --- anchor: C1_lo and C1_hi are not touched between the handoff and the
	// correction, and the correction's other arm is a separate branch.
	p.noWriteBetween(t, b.blk, "C1_lo", p.stmtAt(t, "branch body", b.blk, 13), b.correction)
	p.noWriteBetween(t, b.blk, "C1_hi", p.stmtAt(t, "branch body", b.blk, 13), b.correction)
	// --- anchor: the scaled C1 survives the C2 rounding untouched, so the low
	// word entering the subtraction is still the low word of 10^34.
	p.noWriteBetween(t, b.blk, "C1", p.stmtAt(t, "branch body", b.blk, 3), p.stmtAt(t, "branch body", b.blk, 6))
	// --- anchor: closed world of R256 writes in the branch, so R256.w2 at the
	// subtraction is the rounded single-digit C2, at most decremented once.
	p.pinWrites(t, b.blk, "R256", []string{
		"R256 = __mul_128x128_to_256(C2, bid_ten2mk128[ind])",
		"R256.w2 = (R256.w2 >> uint(shift)) | (R256.w3 << uint(64-shift))",
		"R256.w3 = (R256.w3 >> uint(shift))",
		"R256.w2 = (R256.w3 >> uint(shift-64))",
		"R256.w3 = 0x0",
		"R256.w2--",
		"R256.w3--",
		"R256.w2 = C2_lo",
		"R256.w3 = C2_hi",
	})

	// --- premise: the scaled left operand is exactly 10^34 for every q1, so
	// tmp64 is the low word of 10^34. The enumeration runs to 35 rather than
	// P34: q1 == 35 is already outside the canonical Decimal128 coefficient
	// range, and any larger q1 makes the pinned dispatch index a table out of
	// range, so 35 is the widest domain that can reach the subtraction at all.
	want34 := pow10big(34)
	const lowWordOf10Pow34 = uint64(0x378d8e6400000000)
	for q1 := 1; q1 <= 35; q1++ {
		scale := 34 - q1 + 1
		c1 := pow10big(q1 - 1)
		scaled := bid128AddScaleC1(t, c1, q1, scale)
		if scaled.Cmp(want34) != 0 {
			t.Fatalf("q1=%d scale=%d: the port's C1 scaling gives %s, want 10^34", q1, scale, scaled)
		}
		if lo := new(big.Int).And(scaled, new(big.Int).SetUint64(^uint64(0))).Uint64(); lo != lowWordOf10Pow34 {
			t.Fatalf("q1=%d: low word of the scaled C1 is %016x, want %016x", q1, lo, lowWordOf10Pow34)
		}
	}

	// --- premise: the rounded C2 fits a single decimal digit, so R256 stays
	// far below the low word of 10^34. The port's rounding is monotone in C2
	// (a 128x128 product by a constant followed by a right shift), so the
	// extremes of each q2 digit class bound the whole class; interior draws are
	// checked against those extremes so the monotonicity the bound relies on is
	// not merely asserted.
	const maxRoundedC2 = 11
	r := rand.New(rand.NewSource(53301754))
	for q2 := 1; q2 <= 35; q2++ {
		lo := pow10big(q2 - 1)
		hi := new(big.Int).Sub(pow10big(q2), big.NewInt(1))
		span := new(big.Int).Sub(hi, lo)
		loW2, _ := bid128AddRoundC2ToOneDigit(t, lo, q2)
		hiW2, _ := bid128AddRoundC2ToOneDigit(t, hi, q2)
		probes := []*big.Int{lo, hi}
		for i := 0; i < 64 && span.Sign() > 0; i++ {
			probes = append(probes, new(big.Int).Add(lo, new(big.Int).Rand(r, span)))
		}
		for _, c2 := range probes {
			w2, w3 := bid128AddRoundC2ToOneDigit(t, c2, q2)
			if w3 != 0 {
				t.Fatalf("q2=%d C2=%s: rounded C2 high word is %016x, want 0", q2, c2, w3)
			}
			if w2 < 1 || w2 > maxRoundedC2 {
				t.Fatalf("q2=%d C2=%s: rounded C2 is %d, outside [1, %d]; the certificate's single-digit bound is broken",
					q2, c2, w2, maxRoundedC2)
			}
			if w2 < loW2 || w2 > hiW2 {
				t.Fatalf("q2=%d C2=%s: rounded C2 is %d, outside the digit class extremes [%d, %d]; the monotonicity the bound relies on is broken",
					q2, c2, w2, loW2, hiW2)
			}
		}
	}
	// The midpoint fixup can subtract one from R256.w2, so the reachable window
	// is [0, maxRoundedC2]; it must stay clear of the low word of 10^34.
	if maxRoundedC2 >= lowWordOf10Pow34 {
		t.Fatalf("the single-digit bound %d is not below the low word of 10^34 (%016x)", maxRoundedC2, lowWordOf10Pow34)
	}
}

// bid128AddScaleC1 replays the pinned three-way C1 scaling dispatch with the
// port's own helpers and power-of-ten tables.
func bid128AddScaleC1(t *testing.T, c1 *big.Int, q1, scale int) *big.Int {
	t.Helper()
	if !c1.IsUint64() && q1 <= 19 {
		t.Fatalf("q1=%d: 10^(q1-1) does not fit in 64 bits", q1)
	}
	var out BID_UINT128
	switch {
	case scale >= 20:
		out = __mul_128x64_to_128(c1.Uint64(), bid_ten2k128[scale-20])
	case q1 <= 19:
		out = __mul_64x64_to_128(c1.Uint64(), bid_ten2k64[scale])
	default:
		wide := big128(c1)
		out = __mul_128x64_to_128(bid_ten2k64[scale], wide)
	}
	return big.NewInt(0).Add(new(big.Int).Lsh(new(big.Int).SetUint64(out.hi), 64), new(big.Int).SetUint64(out.lo))
}

// bid128AddRoundC2ToOneDigit replays the pinned "round C2 to one decimal digit"
// sequence (midpoint add, 128x128 product against bid_ten2mk128, and the
// bid_shiftright128 extraction) and returns R256.w2, R256.w3 as the port leaves
// them. For q2 == 1 the port takes the ind == -1 arm and copies C2 directly.
func bid128AddRoundC2ToOneDigit(t *testing.T, c2 *big.Int, q2 int) (uint64, uint64) {
	t.Helper()
	x1 := q2 - 1
	ind := x1 - 1
	c2w := big128(c2)
	if ind < 0 {
		return c2w.lo, c2w.hi
	}
	before := c2w.lo
	if ind <= 18 {
		c2w.lo += bid_midpoint64[ind]
		if c2w.lo < before {
			c2w.hi++
		}
	} else {
		c2w.lo += bid_midpoint128[ind-19].lo
		c2w.hi += bid_midpoint128[ind-19].hi
		if c2w.lo < before {
			c2w.hi++
		}
	}
	r := __mul_128x128_to_256(c2w, bid_ten2mk128[ind])
	w2, w3 := r.w2, r.w3
	if ind >= 3 {
		shift := bid_shiftright128[ind]
		if shift < 64 {
			w2 = (w2 >> uint(shift)) | (w3 << uint(64-shift))
			w3 = w3 >> uint(shift)
		} else {
			w2 = w3 >> uint(shift-64)
			w3 = 0
		}
	}
	// Cross-check against the exact quotient the arm is meant to produce.
	if q2 >= 2 {
		half := new(big.Int).Mul(big.NewInt(5), pow10big(x1-1))
		want := new(big.Int).Div(new(big.Int).Add(c2, half), pow10big(x1))
		if !want.IsUint64() || want.Uint64() != w2 {
			t.Fatalf("q2=%d C2=%s: port rounds to %d, exact floor((C2 + 5*10^(x1-1))/10^x1) = %s", q2, c2, w2, want)
		}
	}
	return w2, w3
}

func big128(v *big.Int) BID_UINT128 {
	mask := new(big.Int).SetUint64(^uint64(0))
	lo := new(big.Int).And(v, mask).Uint64()
	hi := new(big.Int).Rsh(v, 64)
	if !hi.IsUint64() {
		panic("value wider than 128 bits")
	}
	return BID_UINT128{lo: lo, hi: hi.Uint64()}
}

// ---------------------------------------------------------------------------
// certificate 4: div64.go:267 (Bid64Div trailing-zero elimination)
// ---------------------------------------------------------------------------

// bid64DivEstimateSlack is how far the port's double-precision quotient
// estimate may sit below the exact quotient floor(cx/cy) after the single
// `Q += uint64(D)` correction. This is the one premise in this file that is a
// property of IEEE-754 double arithmetic rather than of a port table: cx and cy
// are both below 10^16 < 2^54, so fl(cx|1)/fl(cy) carries at most three
// roundings of relative size 2^-53, and the absolute error over a quotient
// bounded by 10^16 stays well under 8. It is checked over the boundary plus
// seeded corpus in TestBid64DivQuotientIsNormalisedTo16Digits.
const bid64DivEstimateSlack = 8

// TestBid64DivTrailingZeroBlockHighDigitIsNeverZero certifies that the
// condition
//
//	if digit == 0 && tdigit[1] == 0 {
//
// of Bid64Div's long trailing-zero elimination (audit site div64.go:6396)
// cannot change value when its first comparison is inverted, because
// tdigit[1] != 0 whenever the condition is evaluated. With tdigit[1] != 0 both
// `digit == 0 && tdigit[1] == 0` and `digit != 0 && tdigit[1] == 0` are false,
// so the mutant is equivalent on every reachable input.
//
// Proof. tdigit[1] accumulates only non-negative contributions: after the two
// initialisers the block's whole closed world of writes is `tdigit[1] +=
// bid_convert_table[j][k][1]` and `tdigit[1]++`. The loop walks the 7-bit
// chunks of QX32 = uint32(Q >> 26) from the bottom up and stops when the
// remaining chunks are zero, so it visits the chunk holding QX32's most
// significant set bit; that chunk's k is therefore non-zero. For every chunk
// index j >= 3 and every k >= 1 the table's high limb bid_convert_table[j][k][1]
// is at least one, so tdigit[1] >= 1 as soon as QX32 >= 2^21.
//
// QX32 >= 2^21 follows from the quotient normalisation: Bid64Div reaches the
// block with 2^47 <= Q < 2^58, so Q >> 26 lands in [2^21, 2^32) and the
// uint32 conversion is lossless. That bound is certified separately by
// TestBid64DivQuotientIsNormalisedTo16Digits.
func TestBid64DivTrailingZeroBlockHighDigitIsNeverZero(t *testing.T) {
	p := loadPortFile(t, "div64.go")
	fn := p.funcDecl(t, "Bid64Div")

	// --- anchor: the long trailing-zero block, located and pinned by its own
	// statement run. The pinned run contains the mutated condition, so any edit
	// to it fails here.
	const pinnedChunkLoop = "for j := 0; QX32 != 0; j, QX32 = j+1, QX32>>7 { k := int(QX32 & 127) " +
		"tdigit[0] += bid_convert_table[j][k][0] tdigit[1] += bid_convert_table[j][k][1] " +
		"if tdigit[0] >= 100000000 { tdigit[0] -= 100000000 tdigit[1]++ } }"
	const pinnedNzerosSwitch = "if digit == 0 && tdigit[1] == 0 { nzeros += 16 } " +
		"else { if digit == 0 { nzeros += 8 digit = tdigit[1] } " +
		"PD := uint64(digit) * 0x068DB8BB digit_h := uint32(PD >> 40) digit_low := digit - digit_h*10000 " +
		"if digit_low == 0 { nzeros += 4 } else { digit_h = digit_low } " +
		"if (digit_h & 1) == 0 { nzeros += int(3 & (uint32(bid_packed_10000_zeros[digit_h>>3]) >> (digit_h & 7))) } }"
	blk := p.blockWithRun(t, fn, []string{
		"var tdigit [3]uint32",
		"tdigit[0] = uint32(Q & 0x3ffffff)",
		"tdigit[1] = 0",
		"QX := Q >> 26",
		"QX32 := uint32(QX)",
		"nzeros = 0",
		pinnedChunkLoop,
		"digit := tdigit[0]",
		pinnedNzerosSwitch,
		"if nzeros > 0 { CT := __mul_64x64_to_128(Q, bid_reciprocals10_64[nzeros]) amount := bid_short_recip_scale[nzeros] Q = CT.hi >> uint(amount) }",
		"diff_expon += nzeros",
	})
	p.pinStmtCount(t, "long trailing-zero block", blk, 11)

	// --- anchor: the block is the else of the short-coefficient test, itself
	// inside the exact-division test.
	var shortTest *ast.IfStmt
	ast.Inspect(fn, func(n ast.Node) bool {
		if is, ok := n.(*ast.IfStmt); ok && is.Else == ast.Stmt(blk) {
			shortTest = is
		}
		return true
	})
	if shortTest == nil {
		t.Fatal("div64.go: the long trailing-zero block is not the else of an if statement")
	}
	p.pinNode(t, "short-coefficient test", shortTest.Cond, "coefficient_x <= 1024 && coefficient_y <= 1024")
	var exactTest *ast.IfStmt
	ast.Inspect(fn, func(n ast.Node) bool {
		is, ok := n.(*ast.IfStmt)
		if !ok || is.Body == nil {
			return true
		}
		for _, st := range is.Body.List {
			if st == ast.Stmt(shortTest) {
				exactTest = is
			}
		}
		return true
	})
	if exactTest == nil {
		t.Fatal("div64.go: the short-coefficient test is not directly inside an if statement")
	}
	p.pinNode(t, "exact-division test", exactTest.Cond, "R == 0")

	// --- anchor: closed world of writes to tdigit, so nothing can lower
	// tdigit[1] after the loop has raised it.
	p.pinWrites(t, fn, "tdigit", []string{
		"tdigit[0] = uint32(Q & 0x3ffffff)",
		"tdigit[1] = 0",
		"tdigit[0] += bid_convert_table[j][k][0]",
		"tdigit[1] += bid_convert_table[j][k][1]",
		"tdigit[0] -= 100000000",
		"tdigit[1]++",
	})
	// --- anchor: closed world of writes to Q, so the normalisation premise
	// covers every value that can reach the block.
	p.pinWrites(t, fn, "Q", []string{
		"Q = 0",
		"Q = uint64(dq)",
		"Q += uint64(D)",
		"Q *= bid_power10_table_128[ed2].lo",
		"Q += Q2",
		"Q += Q2",
		"Q = CT.hi >> uint(amount)",
		"Q = CT.hi >> uint(amount)",
		"Q += R >> 63",
		"Q++",
	})
	p.pinWrites(t, fn, "ed2", []string{
		"ed2 = bid_estimate_decimal_digits[bin_index] + ed1",
		"ed2 = 16 - bid_estimate_decimal_digits[bin_expon_cx] - int(DU)",
	})

	// --- anchor: both quotient-normalisation paths, because
	// TestBid64DivQuotientIsNormalisedTo16Digits reasons about exactly these
	// statements. The coefficient comparison that selects between them is
	// pinned with them.
	scaleUpBlock := p.blockWithRun(t, fn, []string{
		"tempx := math.Float32bits(float32(coefficient_x))",
		"tempy := math.Float32bits(float32(coefficient_y))",
		"bin_index = int((tempy - tempx) >> 23)",
		"A = coefficient_x * bid_power10_index_binexp[bin_index]",
		"B = coefficient_y",
		"temp_b := float64(B)",
		"DU = (A - B) >> 63",
		"ed1 = 15 + int(DU)",
		"ed2 = bid_estimate_decimal_digits[bin_index] + ed1",
		"T = bid_power10_table_128[ed1].lo",
		"CA = __mul_64x64_to_128(A, T)",
		"Q = 0",
		"diff_expon = diff_expon - ed2",
	})
	p.blockWithRun(t, fn, []string{
		"A2 := coefficient_x | 1",
		"da := float64(A2)",
		"db = float64(coefficient_y)",
		"dq := da / db",
		"Q = uint64(dq)",
		"R = coefficient_x - coefficient_y*Q",
		"tempq := math.Float64bits(dq)",
		"bin_expon_cx = int((tempq >> 52)) - 0x3ff",
		"D = int64(R) >> 63",
		"Q += uint64(D)",
		"R += coefficient_y & uint64(D)",
		"if int64(R) <= 0 { res = get_BID64(sign_x^sign_y, diff_expon, Q+R, rndMode) return res }",
		"DU = bid_power10_index_binexp[bin_expon_cx] - Q - 1",
		"DU >>= 63",
		"ed2 = 16 - bid_estimate_decimal_digits[bin_expon_cx] - int(DU)",
		"T = bid_power10_table_128[ed2].lo",
		"CA = __mul_64x64_to_128(R, T)",
		"B = coefficient_y",
		"Q *= bid_power10_table_128[ed2].lo",
		"diff_expon -= ed2",
	})
	var pathSelect *ast.IfStmt
	ast.Inspect(fn, func(n ast.Node) bool {
		if is, ok := n.(*ast.IfStmt); ok && is.Body == scaleUpBlock {
			pathSelect = is
		}
		return true
	})
	if pathSelect == nil {
		t.Fatal("div64.go: the scale-up path is not an if body")
	}
	p.pinNode(t, "quotient path selection", pathSelect.Cond, "coefficient_x < coefficient_y")
	// --- anchor: the quick divide that turns CA/B into the final quotient.
	p.blockWithRun(t, fn, []string{
		"Q2 = CA.lo / B",
		"B2 = B + B",
		"B4 = B2 + B2",
		"R = CA.lo - Q2*B",
		"Q += Q2",
	})

	// --- premise: every chunk index the loop can reach at or above 3 has a
	// non-zero high limb for every non-zero k, so a set bit at position >= 21
	// of QX32 forces tdigit[1] >= 1.
	if len(bid_convert_table) != 5 {
		t.Fatalf("bid_convert_table has %d chunk rows, expected 5", len(bid_convert_table))
	}
	for j := 3; j < len(bid_convert_table); j++ {
		for k := 1; k < len(bid_convert_table[j]); k++ {
			if bid_convert_table[j][k][1] == 0 {
				t.Fatalf("bid_convert_table[%d][%d][1] == 0; a quotient whose top chunk is %d would leave tdigit[1] at zero", j, k, k)
			}
		}
	}
	// The table's two-limb base-10^8 decomposition is what makes the high limb
	// meaningful. Intel keeps only the low 16 decimal digits of each chunk
	// weight (the 10^16 place and above are dropped, because the routine never
	// strips more than 16 zeros), so the identity is modulo 10^16. Checking it
	// over the whole table also bounds the limbs: with every low limb below
	// 10^8 and every high limb below 10^8, the loop's at most five chunk
	// contributions plus five carries keep tdigit[1] under 2^32, so it cannot
	// wrap back to zero after the top chunk has raised it.
	mod16 := pow10big(16)
	for j := 0; j < len(bid_convert_table); j++ {
		for k := 0; k < len(bid_convert_table[j]); k++ {
			low, high := bid_convert_table[j][k][0], bid_convert_table[j][k][1]
			if low >= 100000000 || high >= 100000000 {
				t.Fatalf("bid_convert_table[%d][%d] = {%d, %d}: a limb is not below 10^8, so the loop's accumulation bound is broken", j, k, low, high)
			}
			want := new(big.Int).Mod(new(big.Int).Lsh(big.NewInt(int64(k)), uint(26+7*j)), mod16)
			got := new(big.Int).Add(new(big.Int).SetUint64(uint64(low)),
				new(big.Int).Mul(big.NewInt(100000000), new(big.Int).SetUint64(uint64(high))))
			if got.Cmp(want) != 0 {
				t.Fatalf("bid_convert_table[%d][%d] decodes to %s, want (k*2^(26+7j)) mod 10^16 = %s", j, k, got, want)
			}
		}
	}
	if maxAccum := int64(len(bid_convert_table))*100000000 + int64(len(bid_convert_table)); maxAccum >= 1<<32 {
		t.Fatalf("the loop can accumulate up to %d into the uint32 tdigit[1]; the no-wrap premise is broken", maxAccum)
	}

	// --- premise: with 2^47 <= Q < 2^58 the chunk that carries QX32's top set
	// bit has index in [3, 4] and non-zero k.
	for _, q := range bid64DivQuotientProbeValues() {
		qx32 := uint32(q >> 26)
		if uint64(qx32) != q>>26 {
			t.Fatalf("Q=%d: uint32(Q>>26) truncates; the certificate's Q < 2^58 premise is broken", q)
		}
		top := 31 - bits.LeadingZeros32(qx32)
		j := top / 7
		if j < 3 || j >= len(bid_convert_table) {
			t.Fatalf("Q=%d: QX32=%d puts its top set bit in chunk %d, outside the certified [3,%d] range",
				q, qx32, j, len(bid_convert_table)-1)
		}
		if k := (qx32 >> uint(7*j)) & 127; k == 0 {
			t.Fatalf("Q=%d: chunk %d of QX32=%d is zero although it holds the top set bit", q, j, qx32)
		}
	}
}

// bid64DivQuotientProbeValues returns quotient values on the certified window's
// edges plus every power-of-two and chunk boundary inside it, so the top-chunk
// argument is exercised at each place it could break.
func bid64DivQuotientProbeValues() []uint64 {
	out := []uint64{1 << 47, (1 << 58) - 1, 1000000000000000, 9999999999999999, 800000000000000}
	for b := 47; b <= 57; b++ {
		out = append(out, uint64(1)<<uint(b), (uint64(1)<<uint(b))+1, (uint64(1)<<uint(b+1))-1)
	}
	r := rand.New(rand.NewSource(75407540))
	for i := 0; i < 4096; i++ {
		v := (r.Uint64() % ((1 << 58) - (1 << 47))) + (1 << 47)
		out = append(out, v)
	}
	return out
}

// TestBid64DivQuotientIsNormalisedTo16Digits certifies the numeric premise the
// trailing-zero certificate depends on: when Bid64Div reaches its exact-division
// trailing-zero elimination, the quotient satisfies 2^47 <= Q < 2^58.
//
// Both entry paths compute the same value, Q = floor(cx * 10^ed2 / cy). The
// quick divide that produces it estimates its quotient in double precision and
// then corrects with at most five conditional subtractions, but the block is
// dominated by `if R == 0`, and a zero final remainder means CA == Q2*B
// exactly - so on every path that reaches the block the quotient is the exact
// floor, not an estimate:
//
//   - cx >= cy: Q_init = uint64(fl(cx|1)/fl(cy)) after one correction, and the
//     port only continues past `if int64(R) <= 0` when cx > cy*Q_init, so
//     Q_init < cx/cy exactly. The scale is ed2 = 16 - est[b] - DU with
//     b = binexp(fl quotient) and DU = 1 exactly when Q_init >= 10^est[b], so
//     Q_init >= 2^b - 1 and Q_init < 2^(b+1) hold by construction.
//   - cx < cy: A = cx * 10^est[bin_index] and Q = floor(A * 10^(15+DU) / cy)
//     with DU = 1 exactly when A < cy. Because 10^est[i] > 2^i and the float32
//     exponent difference bin_index is within one of binexp(cy) - binexp(cx),
//     A stays inside (cy/4, 80*cy).
//
// The enumerations below evaluate both windows over every binary exponent the
// port's tables admit and assert the result stays inside [2^47, 2^58).
func TestBid64DivQuotientIsNormalisedTo16Digits(t *testing.T) {
	lower, upper := pow2big(47), pow2big(58)
	one := big.NewInt(1)

	// --- cx >= cy path.
	for b := 0; b <= 53; b++ {
		est := bid_estimate_decimal_digits[b]
		for du := 0; du <= 1; du++ {
			ed2 := 16 - est - du
			if ed2 < 0 {
				// bid_power10_table_128[ed2] would index out of range, so the
				// port traps before the trailing-zero block instead of
				// reaching it with an unnormalised quotient.
				continue
			}
			scale := pow10big(ed2)
			// Smallest admissible Q_init for this (b, DU) pair.
			qLo := new(big.Int).Sub(pow2big(b), one)
			if du == 1 {
				qLo = pow10big(est)
			}
			if qLo.Cmp(one) < 0 {
				qLo = new(big.Int).Set(one)
			}
			if got := new(big.Int).Mul(qLo, scale); got.Cmp(lower) < 0 {
				t.Fatalf("cx>=cy path b=%d DU=%d: minimal quotient %s * 10^%d = %s is below 2^47",
					b, du, qLo, ed2, got)
			}
			// Largest admissible Q_init, plus the float slack and the one unit
			// that separates floor(cx/cy) from cx/cy itself.
			qHi := pow2big(b + 1)
			if du == 0 && pow10big(est).Cmp(qHi) < 0 {
				qHi = pow10big(est)
			}
			qHi = new(big.Int).Add(qHi, big.NewInt(bid64DivEstimateSlack+1))
			if got := new(big.Int).Mul(qHi, scale); got.Cmp(upper) >= 0 {
				t.Fatalf("cx>=cy path b=%d DU=%d: maximal quotient %s * 10^%d = %s reaches 2^58",
					b, du, qHi, ed2, got)
			}
		}
	}

	// --- cx < cy path. The float32 exponent difference is within one of the
	// exact binary exponent difference; check that lemma at the extremes of
	// every binade (float32 conversion is monotone, so the extremes bound the
	// whole binade) and then bound A/cy.
	for bx := 0; bx <= 53; bx++ {
		lo := uint64(1) << uint(bx)
		hi := (uint64(1) << uint(bx+1)) - 1
		for _, v := range []uint64{lo, hi} {
			e := int(math.Float32bits(float32(v))>>23) - 127
			if e != bx && e != bx+1 {
				t.Fatalf("float32 exponent lemma: v=%d has binary exponent %d but float32 exponent %d", v, bx, e)
			}
		}
	}
	// The largest bin_index the path can form is binexp(cy) - binexp(cx) + 1,
	// and both coefficients stay under 2^54, so index 55 must exist.
	if len(bid_power10_index_binexp) <= 55 || len(bid_estimate_decimal_digits) <= 55 {
		t.Fatalf("the digit-estimate tables (%d and %d entries) do not cover every bin_index the cx<cy path can form",
			len(bid_power10_index_binexp), len(bid_estimate_decimal_digits))
	}
	for i := 0; i < len(bid_estimate_decimal_digits); i++ {
		scale := pow10big(bid_estimate_decimal_digits[i])
		// 2^i < 10^est[i] gives A > cy/4; 10^est[i] <= 10*2^i gives A < 80*cy.
		if scale.Cmp(pow2big(i)) <= 0 {
			t.Fatalf("10^est[%d] = %s is not above 2^%d; the cx<cy path's A > cy/4 bound is broken", i, scale, i)
		}
		if scale.Cmp(new(big.Int).Mul(big.NewInt(10), pow2big(i))) > 0 {
			t.Fatalf("10^est[%d] = %s exceeds 10*2^%d; the cx<cy path's A < 80*cy bound is broken", i, scale, i)
		}
	}
	// DU == 1 (A < cy): Q = floor(A*10^16/cy) with A > cy/4.
	if got := new(big.Int).Div(pow10big(16), big.NewInt(4)); got.Cmp(lower) < 0 {
		t.Fatalf("cx<cy path DU=1: minimal quotient 10^16/4 = %s is below 2^47", got)
	}
	if got := pow10big(16); got.Cmp(upper) >= 0 {
		t.Fatalf("cx<cy path DU=1: maximal quotient 10^16 = %s reaches 2^58", got)
	}
	// DU == 0 (A >= cy): Q = floor(A*10^15/cy) with cy <= A < 80*cy.
	if got := pow10big(15); got.Cmp(lower) < 0 {
		t.Fatalf("cx<cy path DU=0: minimal quotient 10^15 = %s is below 2^47", got)
	}
	if got := new(big.Int).Mul(big.NewInt(80), pow10big(15)); got.Cmp(upper) >= 0 {
		t.Fatalf("cx<cy path DU=0: maximal quotient 80*10^15 = %s reaches 2^58", got)
	}

	// --- the sampled premise: for every operand pair that gets past the port's
	// own `if int64(R) <= 0` exact-result return, the corrected estimate sits
	// at or below the exact quotient and no more than bid64DivEstimateSlack
	// under it. Pairs that trip the early return never reach the trailing-zero
	// block, so they are outside the certificate's domain.
	checkEstimate := func(cx, cy uint64) {
		if cy == 0 || cx < cy {
			return
		}
		dq := float64(cx|1) / float64(cy)
		q := uint64(dq)
		r := cx - cy*q
		d := int64(r) >> 63
		q += uint64(d)
		r += cy & uint64(d)
		if int64(r) <= 0 {
			return
		}
		exact := cx / cy
		if q > exact || exact-q > bid64DivEstimateSlack {
			t.Fatalf("quotient estimate: cx=%d cy=%d gives Q_init=%d, exact %d (slack pinned at %d)",
				cx, cy, q, exact, bid64DivEstimateSlack)
		}
	}
	edges := []uint64{1, 2, 3, 9, 10, 1024, 1025, 999999999999999, 1000000000000000,
		1 << 53, (1 << 53) + 1, 9999999999999998, 9999999999999999}
	for _, cx := range edges {
		for _, cy := range edges {
			checkEstimate(cx, cy)
		}
	}
	r := rand.New(rand.NewSource(6396754))
	for i := 0; i < 1<<18; i++ {
		cx := r.Uint64()%9999999999999999 + 1
		cy := r.Uint64()%9999999999999999 + 1
		if cx < cy {
			cx, cy = cy, cx
		}
		checkEstimate(cx, cy)
	}
}
