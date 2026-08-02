// Hand-written unreachability certificates for twelve mutation sites of the
// mechanical port. It lives outside every generation path and must stay
// hand-written.
//
// Why this gate exists. The 2026-07 mutation audit left twelve mutants that no
// verification corpus can distinguish, because the mutated token sits in a
// region no input reaches:
//
//	add128_inline.go:117  bid_get_add128  bit:<<->>>
//	bid128_add.go:1248    Bid128Add       aor:dec->inc
//	bid128_add.go:1388    Bid128Add       aor:dec->inc
//	bid128_fma_body.go:1813 bid_fma_cases_11_12 negcond:negate
//	bid128_fma_body.go:1832 bid_fma_cases_11_12 bit:|->&
//	div64.go:267          Bid64Div        cmp:==->!=
//	bid128_round.go:317   bid_round192_39_57 aor: ind-39
//	bid128_round.go:347   bid_round256_58_76 cmp: C.w2 == 0x0
//	bid128_round.go:430   bid_round256_58_76 shift: P512.w5 >> shift
//	bid128_round.go:613   bid_round256_58_76 aor: ind-39
//	bid128_sqrt.go:423    Bid128Sqrt      aor:inc->dec  (audit site 16)
//	bid128_sqrt.go:661    Bid128dSqrt     aor:inc->dec  (audit site 19)
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
// Two premises in this file are sampled rather than exhaustively proved, and both are called out as such at their use
// sites: the double-precision quotient estimate in Bid64Div (see bid64DivEstimateSlack) is an IEEE-754 property of the
// host, not a table fact; and the rounding-helper postcondition Cstar <= 10^(q-x) (see
// TestBid128RoundBoundaryArmsAreUnreachable) is a property of the helpers' own reciprocal tables and shift schedule that
// the current structure does not let this file derive from an AST pin. Both are checked over deterministic boundaries
// plus a fixed-seed corpus driven through the production helpers, and neither is presented as a proof.
//
// The coefficient hypotheses that postcondition is applied under are not sampled and must not become so: both
// 10^(q3-1) <= C3 < 10^q3 and 10^(q4-1) <= C4 < 10^q4 are derived from pinned source plus exact table, multiply and
// digit-bound reasoning. A sampled stand-in for either would leave the postcondition attached to nothing.
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
	"os"
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

// stmtWithText returns the single statement below root whose flattened source
// matches want. Requiring a single hit makes the statement text both a locator
// and a fail-closed source pin.
func (p *portFile) stmtWithText(t *testing.T, root ast.Node, want string) ast.Stmt {
	t.Helper()
	var hits []ast.Stmt
	ast.Inspect(root, func(n ast.Node) bool {
		s, ok := n.(ast.Stmt)
		if ok && p.flat(t, s) == want {
			hits = append(hits, s)
		}
		return true
	})
	if len(hits) != 1 {
		t.Fatalf("%s: expected exactly one statement %q below line %d, found %d",
			p.name, want, p.line(root), len(hits))
	}
	return hits[0]
}

// ifWithCond returns the single if statement below root whose condition has
// the pinned flattened form.
func (p *portFile) ifWithCond(t *testing.T, root ast.Node, want string) *ast.IfStmt {
	t.Helper()
	var hits []*ast.IfStmt
	ast.Inspect(root, func(n ast.Node) bool {
		is, ok := n.(*ast.IfStmt)
		if ok && p.flat(t, is.Cond) == want {
			hits = append(hits, is)
		}
		return true
	})
	if len(hits) != 1 {
		t.Fatalf("%s: expected exactly one if condition %q below line %d, found %d",
			p.name, want, p.line(root), len(hits))
	}
	return hits[0]
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
	pos  token.Pos
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
		out = append(out, portWrite{line: p.line(n), pos: n.Pos(), text: text})
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

// pinNoNestedDecl fails on any declaration of target below fd that is not a top-level var of fd's own body. writes()
// only records a ValueSpec that carries an initialiser, so a nested `var target T` would shadow the binding the pinned
// writes describe while staying invisible to pinWrites. A top-level declaration cannot shadow: it either is the pinned
// binding itself or collides with a parameter or another top-level var, which does not compile - and this test lives in
// the package, so a non-compiling port never reaches a passing run.
func (p *portFile) pinNoNestedDecl(t *testing.T, fd *ast.FuncDecl, target string) {
	t.Helper()
	top := map[*ast.ValueSpec]bool{}
	for _, s := range fd.Body.List {
		ds, ok := s.(*ast.DeclStmt)
		if !ok {
			continue
		}
		if gd, isGen := ds.Decl.(*ast.GenDecl); isGen {
			for _, spec := range gd.Specs {
				if vs, isVal := spec.(*ast.ValueSpec); isVal {
					top[vs] = true
				}
			}
		}
	}
	ast.Inspect(fd, func(n ast.Node) bool {
		vs, ok := n.(*ast.ValueSpec)
		if !ok || top[vs] {
			return true
		}
		for _, name := range vs.Names {
			if name.Name == target {
				t.Fatalf("%s: %s: nested declaration of %s at line %d shadows the pinned binding: %s",
					p.name, fd.Name.Name, target, p.line(vs), p.flat(t, vs))
			}
		}
		return true
	})
}

// pinNoPointerRebind fails on any statement that rebinds the identifier target itself - `target = ...`,
// `target := ...`, `target++` - as opposed to writing through it (`*target = ...`, `target.hi = ...`), which writes()
// and pinWrites already see. For a pointer parameter that distinction is the whole aliasing question: the caller-side
// slot pins fix which storage each parameter names on entry, and only this check keeps two of them from coming to name
// the same storage afterwards.
func (p *portFile) pinNoPointerRebind(t *testing.T, fd *ast.FuncDecl, target string) {
	t.Helper()
	bare := func(e ast.Expr) bool {
		id, ok := e.(*ast.Ident)
		return ok && id.Name == target
	}
	ast.Inspect(fd, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range s.Lhs {
				if bare(lhs) {
					t.Fatalf("%s: %s: the pointer %s is rebound at line %d: %s",
						p.name, fd.Name.Name, target, p.line(s), p.flat(t, s))
				}
			}
		case *ast.IncDecStmt:
			if bare(s.X) {
				t.Fatalf("%s: %s: the pointer %s is rebound at line %d: %s",
					p.name, fd.Name.Name, target, p.line(s), p.flat(t, s))
			}
		}
		return true
	})
}

// pinPointerUseContexts closes the world of uses of a pointer variable by AST context, so a second name for the same
// storage cannot appear. writes(), pinNoPointerRebind and the bare-argument censuses all key on a statement or an
// argument that spells target directly, and a copy escapes every one of them: `alias := target` names alias on the left,
// and `alias.hi = ...` afterwards writes through the same pointer with no statement in the function mentioning target at
// all. Here every identifier spelled target inside fd's body has to fall into one of three contexts:
//
//	target.f    - the operand of a field selector
//	*target     - the operand of a dereference
//	f(..., target, ...) - a bare argument at an argument index of a call this certificate has already pinned
//
// The first two read or write through the pointer without producing another name for it, and what they write is closed
// by pinWrites; the third is the hand-off the caller-side slot pins and the callee census already cover, and `allowed`
// maps such a call's position to the argument indices that may carry target. Every other context - an assignment or
// declaration right-hand side, a composite literal element, a return, an address-of, a range expression, an argument of
// any other call - can create a second name and fails here. A `p.target` selector where target is the field rather than
// the operand names some other value entirely and is skipped. `target.m(...)` is not a field select: the selector is the
// callee of a call, target is the receiver, and the callee holds the same pointer under another name, so it fails.
//
// Admitting the field select is only sound while target's type declares no methods. The method *value* `target.m` -
// spelled as a value rather than called, so it is not the callee of any call - is the same AST shape as a field read,
// and untyped AST cannot separate the two; it binds target as a receiver and so hands the pointer a second name that
// this classification would wave through as "field select". This helper cannot close that itself, so every caller has
// to pin it separately with pinTypeIsMethodFree: TestBid128RoundBoundaryArmsAreUnreachable runs that census over
// BID_UINT128 - the type of the C3 pointer it censuses - and the field-select admission here rests on it.
func (p *portFile) pinPointerUseContexts(t *testing.T, fd *ast.FuncDecl, target string, allowed map[token.Pos][]int) {
	t.Helper()
	parent := map[ast.Node]ast.Node{}
	var stack []ast.Node
	var uses []*ast.Ident
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		if len(stack) > 0 {
			parent[n] = stack[len(stack)-1]
		}
		if id, ok := n.(*ast.Ident); ok && id.Name == target {
			uses = append(uses, id)
		}
		stack = append(stack, n)
		return true
	})
	classify := func(id *ast.Ident) (string, bool) {
		switch par := parent[id].(type) {
		case *ast.SelectorExpr:
			if par.Sel == id {
				return "field name of another value", true
			}
			if par.X == id {
				// `target.f` and `target.m(...)` are the same AST shape, so a selector that is the callee of a
				// call is not a field read: it hands target to a method as its receiver, which names the same
				// storage inside the callee and escapes every census here. Untyped AST cannot tell a method
				// from a call through a func-valued field, so both fail.
				if ce, ok := parent[ast.Node(par)].(*ast.CallExpr); ok && ce.Fun == ast.Expr(par) {
					return "the callee of a call, so it hands target over as a method receiver", false
				}
				return "field select", true
			}
		case *ast.StarExpr:
			if par.X == id {
				return "dereference", true
			}
		case *ast.CallExpr:
			for i, a := range par.Args {
				if a != ast.Expr(id) {
					continue
				}
				for _, slot := range allowed[par.Pos()] {
					if i == slot {
						return "pinned call argument", true
					}
				}
				return "an argument of a call with no pinned slot for it", false
			}
		}
		return "a copy or another unclassified context", false
	}
	for _, id := range uses {
		kind, ok := classify(id)
		if ok {
			continue
		}
		enclosing := ast.Node(id)
		if par := parent[id]; par != nil {
			enclosing = par
		}
		t.Fatalf("%s: %s: the use of %s at line %d is %s, not a field select, a dereference or a pinned call "+
			"argument, so it can hand the pointer a second name: %s",
			p.name, fd.Name.Name, target, p.line(id), kind, p.flat(t, enclosing))
	}
}

// uniqueUnboundedLoop returns fd's single `for {}` - no init, no condition, no post - and fails on none or several.
func (p *portFile) uniqueUnboundedLoop(t *testing.T, fd *ast.FuncDecl) *ast.ForStmt {
	t.Helper()
	var out []*ast.ForStmt
	ast.Inspect(fd, func(n ast.Node) bool {
		if fs, ok := n.(*ast.ForStmt); ok && fs.Init == nil && fs.Cond == nil && fs.Post == nil {
			out = append(out, fs)
		}
		return true
	})
	if len(out) != 1 {
		t.Fatalf("%s: %s has %d unbounded for loops, pinned 1", p.name, fd.Name.Name, len(out))
	}
	return out[0]
}

// within reports whether n starts strictly inside blk.
func within(n ast.Node, blk *ast.BlockStmt) bool { return n.Pos() > blk.Pos() && n.Pos() < blk.End() }

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

// TestBid128FmaCases1112UnderflowTailIsUnreachable certifies both mutation
// sites in the x0 == ind tail of bid_fma_cases_11_12:
//
//	else if R128.hi == bid_midpoint128[ind-20].hi &&
//	        R128.lo == bid_midpoint128[ind-20].lo {
//		...
//	}
//	...
//	res.hi = p_sign | res.hi
//
// The whole underflow block that contains them is dead for Cases 11 and 12.
// At the caller, delta is initially q3+e3-q4-e4 and is negated on entry to the
// delta-lt-zero dispatcher. Both Case 11 and Case 12 require
// q4 < delta+q3, so at the callee boundary
//
//	e4-e3 = delta+q3-q4 >= 1.
//
// The finite addend exponent is e3=(z_exp>>49)-6176 >= -6176, hence the
// callee starts with e4 >= -6175. Before the underflow guard, e4 can change
// only in the ind > P34 rounding arm: it first gains
// (ind-P34)+incr_exp, where the three rounding helpers only produce
// incr_exp in {0,1}; the sole following decrement is nested in that same arm.
// The net change is therefore non-negative. The guard e4 < -6176 cannot be
// true, so both audited tokens below it are unreachable.
func TestBid128FmaCases1112UnderflowTailIsUnreachable(t *testing.T) {
	callerFile := loadPortFile(t, "bid128_fma.go")
	caller := callerFile.funcDecl(t, "bid128_ext_fma")
	bodyFile := loadPortFile(t, "bid128_fma_body.go")
	dispatch := bodyFile.funcDecl(t, "bid_fma_delta_lt_zero")
	cases := bodyFile.funcDecl(t, "bid_fma_cases_11_12")

	// --- caller anchor: e3 comes from an unsigned exponent field, and that
	// assignment plus the address handed to the main body are the closed world
	// of writes to e3 in the caller, so e3 >= -6176 holds at the call.
	callerFile.pinWrites(t, caller, "e3", []string{
		"e3 = int(z_exp>>49) - 6176",
		"ADDRESS-TAKEN &e3",
	})
	// delta has the pinned algebraic definition and no other write, and that
	// definition is contiguous with the only call that consumes it: nothing can
	// be inserted between them to rewrite e3, e4 or delta.
	callerFile.pinWrites(t, caller, "delta", []string{"delta := q3 + e3 - q4 - e4"})
	callerFile.blockWithRun(t, caller, []string{
		"delta := q3 + e3 - q4 - e4",
		"bid_fma_main_body(p34, &res, &is_midpoint_lt_even, &is_midpoint_gt_even, " +
			"&is_inexact_lt_midpoint, &is_inexact_gt_midpoint, p_sign, z_sign, &z_exp, &p_exp, " +
			"q3, q4, &e3, &e4, delta, &C3, C4, rnd_mode, pfpsf)",
	})

	// --- wrapper anchor: bid_fma_main_body copies e3/e4 by value out of the
	// caller's pointers and, on the delta < 0 side, hands them straight to the
	// dispatcher. The copies are a contiguous run, the delta split is the only
	// if in that block, its else body is exactly the dispatcher call, and
	// nothing writes e3, e4 or delta between the copies and the split.
	mainBody := bodyFile.funcDecl(t, "bid_fma_main_body")
	mainBlock := bodyFile.blockWithRun(t, mainBody, []string{"e3 := *e3_ptr", "e4 := *e4_ptr"})
	var e4Copy ast.Stmt
	var deltaSplits []*ast.IfStmt
	for _, stmt := range mainBlock.List {
		if bodyFile.flat(t, stmt) == "e4 := *e4_ptr" {
			e4Copy = stmt
		}
		if is, ok := stmt.(*ast.IfStmt); ok {
			deltaSplits = append(deltaSplits, is)
		}
	}
	if e4Copy == nil || len(deltaSplits) != 1 {
		t.Fatalf("%s: bid_fma_main_body body drifted: e4 copy found=%v, top-level ifs=%d",
			bodyFile.name, e4Copy != nil, len(deltaSplits))
	}
	deltaSplit := deltaSplits[0]
	bodyFile.pinNode(t, "main-body delta split", deltaSplit.Cond, "delta >= 0")
	for _, v := range []string{"e3", "e4", "delta"} {
		bodyFile.noWriteBetween(t, mainBody, v, e4Copy, deltaSplit)
	}
	bodyFile.pinWrites(t, mainBody, "delta", nil)
	deltaLtZero := bodyFile.elseBlock(t, "main-body delta < 0 arm", deltaSplit)
	bodyFile.pinStmtCount(t, "main-body delta < 0 arm", deltaLtZero, 1)
	bodyFile.pinNode(t, "main-body dispatcher call", deltaLtZero.List[0],
		"bid_fma_delta_lt_zero(p34, res, &is_midpoint_lt_even, &is_midpoint_gt_even, "+
			"&is_inexact_lt_midpoint, &is_inexact_gt_midpoint, p_sign, z_sign, &z_exp, &p_exp, "+
			"q3, q4, &e3, &e4, delta, C3, C4, rnd_mode, pfpsf)")

	// --- dispatcher anchor: negation immediately precedes the closed if/else
	// case chain. The Cases 11/12 arm has the exact common strict inequality
	// used by the proof and contains only the pinned call.
	// Closed world of delta writes in the whole dispatcher: only the negation
	// that dominates the case chain and the swap-arm recomputation in the
	// mutually exclusive Cases 8/9/10/13/14/18 arm. A delta write added
	// anywhere else in the function - including in an earlier arm's if Init -
	// breaks this pin.
	bodyFile.pinWrites(t, dispatch, "delta", []string{
		"delta = -delta",
		"delta = q3 + e3 - q4 - e4",
	})
	dispatchBlock := bodyFile.blockWithRun(t, dispatch, []string{"delta = -delta"})
	negateIndex := -1
	for i, stmt := range dispatchBlock.List {
		if bodyFile.flat(t, stmt) == "delta = -delta" {
			negateIndex = i
			break
		}
	}
	if negateIndex < 0 || negateIndex+1 >= len(dispatchBlock.List) {
		t.Fatalf("%s: delta negation no longer immediately dominates the case dispatch", bodyFile.name)
	}
	chain := bodyFile.asIf(t, "delta-lt case dispatch", dispatchBlock.List[negateIndex+1])
	// Walk that one if/else-if chain object; the Cases 11/12 arm has to be an
	// arm of it, not merely some if somewhere in the function.
	wantCaseCond := "(p34 <= delta && delta < q4 && q4 < delta+q3) || " +
		"(delta < p34 && p34 < q4 && q4 < delta+q3)"
	var caseArm *ast.IfStmt
	arms := 0
	for cur := chain; cur != nil; {
		arms++
		if bodyFile.flat(t, cur.Cond) == wantCaseCond {
			if caseArm != nil {
				t.Fatalf("%s: Cases 11/12 condition appears in more than one arm of the dispatch chain",
					bodyFile.name)
			}
			caseArm = cur
		}
		next, ok := cur.Else.(*ast.IfStmt)
		if !ok {
			break
		}
		cur = next
	}
	if caseArm == nil {
		t.Fatalf("%s: Cases 11/12 arm is not part of the if/else-if chain dominated by the delta negation (walked %d arms)",
			bodyFile.name, arms)
	}
	bodyFile.pinStmtCount(t, "Cases 11/12 arm", caseArm.Body, 1)
	bodyFile.pinNode(t, "Cases 11/12 call", caseArm.Body.List[0],
		"bid_fma_cases_11_12(p34, res, &is_midpoint_lt_even, &is_midpoint_gt_even, "+
			"&is_inexact_lt_midpoint, &is_inexact_gt_midpoint, p_sign, z_sign, q3, q4, "+
			"&e3, &e4, delta, C3, C4, rnd_mode, pfpsf)")

	// --- callee anchor: pin the exponent floor and every write to e4 before
	// the dead guard. This closes the proof against a new hidden mutation of
	// e4 or an address-taking call being inserted before the guard.
	// Closed world of expmin128 writes in the whole callee: the declaration is
	// the only one, and its address is never taken. Pinning the declaration's
	// existence alone would let a later `expmin128 = ...` or `f(&expmin128)`
	// move the floor the guard compares against without failing the test.
	bodyFile.pinWrites(t, cases, "expmin128", []string{"expmin128 := -6176"})
	var underflowHits []*ast.IfStmt
	ast.Inspect(cases, func(n ast.Node) bool {
		is, ok := n.(*ast.IfStmt)
		if ok && bodyFile.flat(t, is.Cond) == "e4 < expmin128" &&
			len(is.Body.List) > 0 && bodyFile.flat(t, is.Body.List[0]) == "x0 = expmin128 - e4" {
			underflowHits = append(underflowHits, is)
		}
		return true
	})
	if len(underflowHits) != 1 {
		t.Fatalf("%s: expected one e4 underflow guard starting with x0 assignment, found %d",
			bodyFile.name, len(underflowHits))
	}
	underflow := underflowHits[0]
	var beforeGuard []string
	// Position boundary, not a line comparison: everything textually before the
	// guard's condition counts, so a write packed onto the guard's own line -
	// including one in the guard's if Init - is caught instead of skipped.
	for _, w := range bodyFile.writes(t, cases, "e4") {
		if w.pos < underflow.Cond.Pos() {
			beforeGuard = append(beforeGuard, w.text)
		}
	}
	wantBeforeGuard := []string{
		"e4 := *e4_ptr",
		"e4 = e4 + x0 + incr_exp",
		"e4--",
	}
	if len(beforeGuard) != len(wantBeforeGuard) {
		t.Fatalf("%s: writes to e4 before the underflow guard = %v, want %v",
			bodyFile.name, beforeGuard, wantBeforeGuard)
	}
	for i := range wantBeforeGuard {
		if beforeGuard[i] != wantBeforeGuard[i] {
			t.Fatalf("%s: write %d to e4 before the underflow guard = %q, want %q",
				bodyFile.name, i, beforeGuard[i], wantBeforeGuard[i])
		}
	}

	// The increase and the only decrement are both nested in the ind>P34
	// block, with the decrement after the increase.
	digitSplit := bodyFile.ifWithCond(t, cases, "ind < p34")
	equalDigits := bodyFile.elseIf(t, "ind == p34 arm", digitSplit)
	bodyFile.pinNode(t, "ind == p34 guard", equalDigits.Cond, "ind == p34")
	wideDigits := bodyFile.elseBlock(t, "ind > p34 arm", equalDigits)
	increase := bodyFile.stmtWithText(t, wideDigits, "e4 = e4 + x0 + incr_exp")
	decrement := bodyFile.stmtWithText(t, wideDigits, "e4--")
	if increase.Pos() >= decrement.Pos() {
		t.Fatalf("%s: e4 decrement no longer follows the pinned ind>P34 increase", bodyFile.name)
	}
	bodyFile.stmtWithText(t, wideDigits, "x0 = ind - p34")

	// Closed world of incr_exp writes inside that arm: exactly the three
	// rounding-helper tuple assignments, so the value read by the increase can
	// only come from a helper and nothing can pre-set it.
	bodyFile.pinWrites(t, wideDigits, "incr_exp", []string{
		"R128, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round128_19_38(ind, x0, P128)",
		"R192, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round192_39_57(ind, x0, P192)",
		"R256, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round256_58_76(ind, x0, R256)",
	})

	// Every helper that can set incr_exp on this path is closed to literal
	// zero/one assignments. Named result parameters start at zero, so this
	// proves incr_exp >= 0 without replaying the rounding algorithms.
	roundFile := loadPortFile(t, "bid128_round.go")
	for _, name := range []string{"bid_round128_19_38", "bid_round192_39_57", "bid_round256_58_76"} {
		fn := roundFile.funcDecl(t, name)
		writes := roundFile.writes(t, fn, "incr_exp")
		if len(writes) == 0 {
			t.Fatalf("%s: %s no longer assigns incr_exp", roundFile.name, name)
		}
		for _, w := range writes {
			if w.text != "incr_exp = 0" && w.text != "incr_exp = 1" {
				t.Fatalf("%s: %s writes incr_exp outside {0,1} at line %d: %s",
					roundFile.name, name, w.line, w.text)
			}
		}
	}

	// --- mutation-sensitive dead-region anchor. The first pin fails on the
	// negated equality mutant; the second fails on the |->& mutant.
	x0Greater := bodyFile.asIf(t, "x0 underflow split", bodyFile.stmtAt(t, "underflow body", underflow.Body, 9))
	x0Equal := bodyFile.elseIf(t, "x0 == ind arm", x0Greater)
	bodyFile.pinNode(t, "x0 == ind guard", x0Equal.Cond, "x0 == ind")
	wideMidpoint := bodyFile.elseBlock(t, "ind > 19 midpoint arm",
		bodyFile.ifWithCond(t, x0Equal.Body, "ind <= 19"))
	midpointEqual := bodyFile.ifWithCond(t, wideMidpoint,
		"R128.hi == bid_midpoint128[ind-20].hi && R128.lo == bid_midpoint128[ind-20].lo")
	bodyFile.pinNode(t, "128-bit midpoint equality", midpointEqual.Cond,
		"R128.hi == bid_midpoint128[ind-20].hi && R128.lo == bid_midpoint128[ind-20].lo")
	bodyFile.stmtWithText(t, x0Equal.Body, "res.hi = p_sign | res.hi")

	// --- exact finite-domain premise. Decimal128 addends have at most 34
	// digits and products at most 68. The Case 11/12 predicates themselves
	// bound delta below q4, so this enumeration is the complete integer domain
	// at the dispatcher boundary. It computes the smallest possible initial
	// e4 directly from the pinned algebra.
	if P34 != 34 || MASK_EXP_128>>49 < 1 {
		t.Fatalf("Decimal128 precision/exponent anchors drifted: P34=%d MASK_EXP_128>>49=%d",
			P34, MASK_EXP_128>>49)
	}
	minE4 := int(^uint(0) >> 1)
	reached := 0
	for q3 := 1; q3 <= P34; q3++ {
		for q4 := 1; q4 <= 2*P34; q4++ {
			for delta := 0; delta < q4; delta++ {
				case11 := P34 <= delta && delta < q4 && q4 < delta+q3
				case12 := delta < P34 && P34 < q4 && q4 < delta+q3
				if !case11 && !case12 {
					continue
				}
				reached++
				gap := delta + q3 - q4
				if gap < 1 {
					t.Fatalf("Cases 11/12 q3=%d q4=%d delta=%d gives e4-e3=%d, want >=1",
						q3, q4, delta, gap)
				}
				e4 := -6176 + gap
				if e4 < minE4 {
					minE4 = e4
				}
			}
		}
	}
	if reached == 0 || minE4 != -6175 {
		t.Fatalf("Cases 11/12 domain: reached=%d minimum initial e4=%d, want nonzero and -6175",
			reached, minE4)
	}
	// In the only arm that can change e4 before the guard, ind>P34 gives
	// x0>=1, incr_exp is 0/1, and the optional decrement is at most one.
	for ind := P34 + 1; ind <= 2*P34; ind++ {
		for incrExp := 0; incrExp <= 1; incrExp++ {
			for decrement := 0; decrement <= 1; decrement++ {
				net := ind - P34 + incrExp - decrement
				if net < 0 {
					t.Fatalf("ind=%d incr_exp=%d decrement=%d gives negative e4 correction %d",
						ind, incrExp, decrement, net)
				}
			}
		}
	}
	if minE4 < -6176 {
		t.Fatalf("minimum e4 %d reaches the pinned underflow guard", minE4)
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

// ---------------------------------------------------------------------------
// certificate 5: bid128_round.go round-boundary arms (four mutation sites)
// ---------------------------------------------------------------------------

// portPkg parses every non-test file of the package directory, so the closed world of call sites below covers the
// package, not just the files named here.
func portPkg(t *testing.T) []*portFile {
	t.Helper()
	ents, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package directory: %v", err)
	}
	var out []*portFile
	for _, e := range ents {
		if n := e.Name(); !e.IsDir() && strings.HasSuffix(n, ".go") && !strings.HasSuffix(n, "_test.go") {
			out = append(out, loadPortFile(t, n))
		}
	}
	if len(out) == 0 {
		t.Fatal("package directory holds no non-test Go files")
	}
	return out
}

func (p *portFile) decls() []*ast.FuncDecl {
	var out []*ast.FuncDecl
	for _, d := range p.file.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Body != nil {
			out = append(out, fd)
		}
	}
	return out
}

// guardPath returns, outermost first, the conditions dominating target: each enclosing if's condition (negated when
// entered through its else) and "for"
func (p *portFile) guardPath(t *testing.T, fd *ast.FuncDecl, target ast.Node) string {
	t.Helper()
	par, stack := map[ast.Node]ast.Node{}, []ast.Node{}
	ast.Inspect(fd, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		if len(stack) > 0 {
			par[n] = stack[len(stack)-1]
		}
		stack = append(stack, n)
		return true
	})
	var out []string
	for cur := target; ; {
		up, ok := par[cur]
		if !ok {
			break
		}
		switch s := up.(type) {
		case *ast.IfStmt:
			if cur == ast.Node(s.Body) {
				out = append(out, p.flat(t, s.Cond))
			} else if cur == ast.Node(s.Else) {
				out = append(out, "!("+p.flat(t, s.Cond)+")")
			}
		case *ast.ForStmt:
			if cur == ast.Node(s.Body) {
				out = append(out, "for")
			}
		// Fail closed on every construct whose entry condition this walk cannot render, including the
		// deferred/concurrent/closure bodies whose dominating conditions are not the lexical enclosing ones at all.
		case *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.SelectStmt, *ast.RangeStmt,
			*ast.LabeledStmt, *ast.CaseClause, *ast.CommClause,
			*ast.FuncLit, *ast.GoStmt, *ast.DeferStmt:
			t.Fatalf("%s: %s: unhandled %T dominating line %d", p.name, fd.Name.Name, up, p.line(target))
		}
		cur = up
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return strings.Join(out, " && ")
}

type portCallSite struct {
	key, guards string
	args        []string
	node        *ast.CallExpr
	pf          *portFile
	fd          *ast.FuncDecl
}

// portCallsTo returns every call to name in the package and fails if the name is also used as a value, which would let
// a call site escape the closed world. The identifier census walks each whole ast.File - package-level GenDecl
// initialisers included - and exempts only the declaration's own name node, so a package-level function value or a
// package-level direct call is counted as a mention but not as one of the call sites below and breaks the census.
func portCallsTo(t *testing.T, files []*portFile, name string) []portCallSite {
	t.Helper()
	var out []portCallSite
	idents := 0
	for _, p := range files {
		declName := map[ast.Node]bool{}
		for _, d := range p.file.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == name {
				declName[fd.Name] = true
			}
		}
		ast.Inspect(p.file, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && id.Name == name && !declName[id] {
				idents++
			}
			return true
		})
		// Calls stay associated with their enclosing FuncDecl; a call reached from anywhere else has no enclosing
		// function to pin and must fail the census above rather than enter this world silently.
		for _, fd := range p.decls() {
			ast.Inspect(fd, func(n ast.Node) bool {
				ce, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				if id, isID := ce.Fun.(*ast.Ident); !isID || id.Name != name {
					return true
				}
				var args []string
				for _, a := range ce.Args {
					args = append(args, p.flat(t, a))
				}
				out = append(out, portCallSite{key: p.name + "/" + fd.Name.Name + "/" + p.flat(t, ce),
					args: args, guards: p.guardPath(t, fd, ce), node: ce, pf: p, fd: fd})
				return true
			})
		}
	}
	if idents != len(out) {
		t.Fatalf("%s: mentioned %d times, called %d times; a non-call reference bypasses the pinned call world", name, idents, len(out))
	}
	return out
}

func paramNames(fd *ast.FuncDecl) []string {
	var out []string
	for _, f := range fd.Type.Params.List {
		for _, n := range f.Names {
			out = append(out, n.Name)
		}
	}
	return out
}

// pinSlots fails unless, for every (index, name) in want, callee's parameter at that index is named name and the call
// passes the caller's variable of the same name, directly or by address. Same name on both sides is what carries a value
// into a premise that talks about it; a positional shuffle on either side fails.
func pinSlots(t *testing.T, label string, callee *ast.FuncDecl, args []string, want map[int]string) {
	t.Helper()
	names := paramNames(callee)
	for i, n := range want {
		if i >= len(names) || i >= len(args) || names[i] != n || (args[i] != n && args[i] != "&"+n) {
			t.Fatalf("%s: slot %d is (arg %v, param %v), want %s on both sides", label, i, args, names, n)
		}
	}
}

// pinForwardOnly fails if fd holds a loop or a goto that does not jump forward. Where it passes, source order bounds
// execution order inside fd: a statement can never run before one that precedes it in the file.
func (p *portFile) pinForwardOnly(t *testing.T, fd *ast.FuncDecl) {
	t.Helper()
	labels := map[string]token.Pos{}
	ast.Inspect(fd, func(n ast.Node) bool {
		if ls, ok := n.(*ast.LabeledStmt); ok {
			labels[ls.Label.Name] = ls.Pos()
		}
		return true
	})
	ast.Inspect(fd, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.ForStmt, *ast.RangeStmt:
			t.Fatalf("%s: %s has a loop at line %d; source order no longer bounds execution order",
				p.name, fd.Name.Name, p.line(n))
		case *ast.BranchStmt:
			if s.Tok == token.GOTO {
				if at, ok := labels[s.Label.Name]; !ok || at <= s.Pos() {
					t.Fatalf("%s: %s: goto %s at line %d is not a forward jump",
						p.name, fd.Name.Name, s.Label.Name, p.line(s))
				}
			}
		}
		return true
	})
}

// pinNoGoto fails on any goto in fd. pinForwardOnly still admits a forward goto, and a forward goto can jump over a
// statement while landing before the one that reads it: the skipped assignment would still precede the reader in source
// and still carry a conjunct prefix of the reader's guards, so both halves of a dominance argument built from
// stmtWithText + conjunctPrefix would keep holding for an assignment that never ran. Where a certificate reads a value
// out of one dominating assignment, that jump has to be excluded outright rather than merely pointed forward.
func (p *portFile) pinNoGoto(t *testing.T, fd *ast.FuncDecl) {
	t.Helper()
	ast.Inspect(fd, func(n ast.Node) bool {
		if s, ok := n.(*ast.BranchStmt); ok && s.Tok == token.GOTO {
			t.Fatalf("%s: %s: goto %s at line %d can bypass an assignment that still dominates its reader by source "+
				"order and guard prefix", p.name, fd.Name.Name, s.Label.Name, p.line(s))
		}
		return true
	})
}

// pinNoLoops fails on any for or range statement anywhere in fd. pinNoGoto excludes the one backward jump a goto can
// make, but a loop reintroduces re-entry without any goto at all: a closed world of writes plus a dominating chain then
// says which statements may write a variable and under which conditions, yet no longer says that the write the
// certificate reads is the last one to have run before the read. Where a certificate needs a value to still be the one a
// single dominating assignment produced, the enclosing body has to be single-pass.
func (p *portFile) pinNoLoops(t *testing.T, fd *ast.FuncDecl) {
	t.Helper()
	ast.Inspect(fd, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.ForStmt, *ast.RangeStmt:
			t.Fatalf("%s: %s: the loop at line %d lets a certified statement be re-entered on a state its dominating "+
				"chain was only ever established for on the first pass", p.name, fd.Name.Name, p.line(n))
		}
		return true
	})
}

// pinBlockAlwaysReturns fails unless every path out of blk is a return, so nothing after blk runs on the state blk
// catches. A certificate that reasons about a value only some earlier test admits needs exactly that: an early-exit
// guard is only a premise while it is impossible to fall out of it into the code below. blk's last statement being a
// return rules out reaching the end of the list, and banning every branch statement inside it - break, continue, goto,
// fallthrough - together with labels leaves a return, or a panic that reaches nothing below either, as the only way out.
func (p *portFile) pinBlockAlwaysReturns(t *testing.T, label string, blk *ast.BlockStmt) {
	t.Helper()
	if len(blk.List) == 0 {
		t.Fatalf("%s: %s: the block at line %d is empty, so control falls straight through it", p.name, label, p.line(blk))
	}
	last := blk.List[len(blk.List)-1]
	if _, ok := last.(*ast.ReturnStmt); !ok {
		t.Fatalf("%s: %s: the block at line %d ends in a %T rather than a return, so control can fall out of it into "+
			"the code this certificate reads", p.name, label, p.line(blk), last)
	}
	ast.Inspect(blk, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.BranchStmt:
			t.Fatalf("%s: %s: the %v at line %d leaves the block at line %d without returning",
				p.name, label, s.Tok, p.line(s), p.line(blk))
		case *ast.LabeledStmt:
			t.Fatalf("%s: %s: the label %s at line %d gives the block at line %d an exit this scan cannot follow",
				p.name, label, s.Label.Name, p.line(s), p.line(blk))
		}
		return true
	})
}

// pinReturnWorld asserts the closed world of return statements in fd, in source order. A function's exits need a census
// rather than a text pin per exit for two reasons, both live in the unpackers this certificate reads: two of
// unpack_BID128_value's three returns render identically, so stmtWithText cannot name either, and a return added later
// would simply be an exit no pin mentions. Pinning the whole list fails on both.
func (p *portFile) pinReturnWorld(t *testing.T, fd *ast.FuncDecl, want []string) {
	t.Helper()
	var got []ast.Stmt
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		if r, ok := n.(*ast.ReturnStmt); ok {
			got = append(got, r)
		}
		return true
	})
	if len(got) != len(want) {
		var texts []string
		for _, r := range got {
			texts = append(texts, p.flat(t, r))
		}
		t.Fatalf("%s: %s returns from %d places, pinned %d: %v", p.name, fd.Name.Name, len(got), len(want), texts)
	}
	for i, w := range want {
		if g := p.flat(t, got[i]); g != w {
			t.Fatalf("%s: %s: return %d (line %d) drifted\n got: %s\nwant: %s",
				p.name, fd.Name.Name, i, p.line(got[i]), g, w)
		}
	}
}

// The struct renderings the no-method censuses pin. Each admits only explicitly named scalar fields, so the type has no
// embedded field to promote a method from and no func-typed field a selector could call.
const (
	pinnedUint128Shape = "struct { lo uint64 hi uint64 }"
	pinnedUint256Shape = "struct { w0 uint64 w1 uint64 w2 uint64 w3 uint64 }"
)

// pinTypeIsMethodFree is the package-wide census that keeps a port type method-free, so that `v.name` on a variable of
// that type can only ever be a field selector. Two AST shapes need it, and writes()/noWriteBetween see neither. A call
// `v.m(...)` to a pointer-receiver method takes &v implicitly and can rewrite v with no assignment, inc/dec or
// address-of anywhere in the caller, so a pinned write world over v would be fail-open. The method *value* `v.m`,
// spelled rather than called, renders as the same AST shape as a field read and binds v as a receiver, so
// pinPointerUseContexts' field-select admission would wave it through. Neither exists while the type carries no method
// at all, which is what this census establishes.
//
// Two ways a method reaches the type without any receiver spelling typeName, so a match on the written receiver name
// alone would be fail-open. First, promotion: an embedded field's methods are the outer type's methods, with no
// receiver naming the outer type anywhere. shape pins the struct's exact rendering, so it holds only explicitly named
// fields and there is nothing to promote from. Second, a package-level type alias: `type A = T` makes `func (a A) m()`
// a method on T itself, so the name match runs on the alias-resolved receiver root rather than the written one.
//
// Where this scan stops. It reads every package-level type and method declaration in the non-test files portPkg parses,
// which is the whole world of methods on typeName: Go requires a method's receiver base type to be defined in the same
// package, so no other package can add one, and a type this package does not declare fails the lookup below rather than
// passing silently. It does not reason about func values or interface method sets - a func-typed struct field or an
// interface value called through a selector names a func, not the receiver storage, and the shape pin admits neither
// into the type. Anything the alias walk cannot decide - a cycle, a generic alias, a right-hand side with no root
// identifier - fails rather than resolving to a name that would pass the typeName check. consequence names, in the
// calling certificate's own terms, what stops holding if a method does appear.
func pinTypeIsMethodFree(t *testing.T, files []*portFile, typeName, shape, consequence string) {
	t.Helper()
	type portTypeSpec struct {
		file *portFile
		spec *ast.TypeSpec
	}
	typeSpecs := map[string]portTypeSpec{}
	for _, p := range files {
		for _, d := range p.file.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, s := range gd.Specs {
				ts, ok := s.(*ast.TypeSpec)
				if !ok {
					t.Fatalf("%s: a package-level type declaration holds a %T, so the alias graph cannot be read", p.name, s)
				}
				if prev, dup := typeSpecs[ts.Name.Name]; dup {
					t.Fatalf("%s: type %s is declared again after %s, so the alias graph is ambiguous",
						p.name, ts.Name.Name, prev.file.name)
				}
				typeSpecs[ts.Name.Name] = portTypeSpec{file: p, spec: ts}
			}
		}
	}
	target, ok := typeSpecs[typeName]
	if !ok {
		t.Fatalf("the package declares no package-level %s, so the no-method pin has no type to close", typeName)
	}
	if target.spec.Assign.IsValid() || target.spec.TypeParams != nil {
		t.Fatalf("%s: %s is not a plain non-generic defined type, so its method set cannot be read off its own "+
			"declaration: %s", target.file.name, typeName, target.file.flat(t, target.spec))
	}
	// The exact rendering is the pin: adding, embedding or retyping a field fails here, and with it any promoted method
	// reachable as `v.m`.
	if got := target.file.flat(t, target.spec.Type); got != shape {
		t.Fatalf("%s: %s is no longer the pinned embedded-field-free shape, so promoted methods on it are no "+
			"longer ruled out\n got: %s\nwant: %s", target.file.name, typeName, got, shape)
	}
	// resolveTypeName walks the alias graph from a written type name to the type it finally names. `type A = B` is
	// followed; `type A B` is a distinct defined type whose methods are not typeName's, so it ends the walk, as does a
	// name the package does not declare.
	resolveTypeName := func(what, name string) string {
		seen := map[string]bool{}
		for {
			ts, ok := typeSpecs[name]
			if !ok || !ts.spec.Assign.IsValid() {
				return name
			}
			if seen[name] {
				t.Fatalf("%s: the alias chain from %s revisits %s, so it cannot be resolved and %s cannot be "+
					"ruled out", ts.file.name, what, name, typeName)
			}
			seen[name] = true
			if ts.spec.TypeParams != nil {
				t.Fatalf("%s: %s, reached from %s, is a generic alias, so the chain resolves to no single type and "+
					"%s cannot be ruled out", ts.file.name, name, what, typeName)
			}
			next := rootIdentName(ts.spec.Type)
			if next == "" {
				t.Fatalf("%s: alias %s, reached from %s, has the unrecognised right-hand side %s, so the chain cannot "+
					"rule out %s", ts.file.name, name, what, ts.file.flat(t, ts.spec.Type), typeName)
			}
			name = next
		}
	}
	for _, p := range files {
		for _, d := range p.file.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Recv == nil {
				continue
			}
			// Fail-closed on shape: a receiver list that is not exactly one field, or a receiver type whose root
			// identifier this file cannot name - parenthesised, generic-instantiated with several type arguments, or
			// anything else unexpected - fails here instead of being skipped as "not typeName".
			if len(fd.Recv.List) != 1 {
				t.Fatalf("%s: %s has a %d-field receiver list, so the no-method pin cannot read its receiver type",
					p.name, fd.Name.Name, len(fd.Recv.List))
			}
			recv := fd.Recv.List[0].Type
			// rootIdentName peels *, (), [] and selectors, so both `T` and `*T` land on the same root name, and an
			// unpeelable shape returns "" and fails just above the check rather than below it.
			root := rootIdentName(recv)
			if root == "" {
				t.Fatalf("%s: %s has the unrecognised receiver type %s, so the no-method pin cannot rule out %s",
					p.name, fd.Name.Name, p.flat(t, recv), typeName)
			}
			// Resolve through the alias graph before matching: `type A = T` puts a method on T under a receiver that
			// spells A. Value, pointer and parenthesised receivers all peel to the same root above.
			if root = resolveTypeName("the receiver of "+fd.Name.Name, root); root == typeName {
				t.Fatalf("%s: %s is a method with receiver %s, so %s",
					p.name, fd.Name.Name, p.flat(t, recv), consequence)
			}
		}
	}
}

// conjunctPrefix reports whether the guard path a is b or one of its outermost conjunct prefixes, i.e. whether a's
// position is dominated by everything that dominates b - the form guardPath emits, joined with " && ".
func conjunctPrefix(a, b string) bool {
	return a == "" || a == b || strings.HasPrefix(b, a+" && ")
}

func (p *portFile) stmtIndex(t *testing.T, blk *ast.BlockStmt, want string) int {
	t.Helper()
	hit := -1
	for i, s := range blk.List {
		if p.flat(t, s) == want {
			if hit >= 0 {
				t.Fatalf("%s: %q occurs more than once in the pinned block", p.name, want)
			}
			hit = i
		}
	}
	if hit < 0 {
		t.Fatalf("%s: %q not found in the block at line %d", p.name, want, p.line(blk))
	}
	return hit
}

const (
	formQMinusP34 = iota
	formQMinus1
	formCase2Loop
)

const pinnedFmaCase1DoubleRound = "p34 <= delta-1 || (p34 == delta && e3+6176 < p34-q3) && " +
	"(p_sign != z_sign) && (delta == (q3 + scale + 1)) && " +
	"!((q3 <= 19 && C3.lo != bid_ten2k64[q3-1]) || (q3 == 20 && (C3.hi != 0 || C3.lo != bid_ten2k64[19])) || " +
	"(q3 >= 21 && (C3.hi != bid_ten2k128[q3-21].hi || C3.lo != bid_ten2k128[q3-21].lo)))"

const pinnedFmaCase1ppB = "!(res.hi != 0x0000314dc6448d93 || res.lo != 0x38c15b0a00000000) && e3 > expmin128"

// pinnedRoundCalls is the closed world of production calls to the two wide rounding helpers: file/function/call plus
// the full chain of dominating branch
var pinnedRoundCalls = []struct {
	key, guards string
	form        int
}{
	{"bid128_fma.go/bid128_ext_fma/bid_round192_39_57(q4, x0, P192)",
		"C3.hi == 0x0 && C3.lo == 0x0 && q4 > p34 && !(q4 <= 38) && q4 <= 57", formQMinusP34},
	{"bid128_fma.go/bid128_ext_fma/bid_round256_58_76(q4, x0, C4)",
		"C3.hi == 0x0 && C3.lo == 0x0 && q4 > p34 && !(q4 <= 38) && !(q4 <= 57)", formQMinusP34},
	{"bid128_fma_body.go/bid_fma_delta_ge_zero/bid_round192_39_57(q4, q4-1, P192)",
		pinnedFmaCase1DoubleRound + " && !(q4 == 1) && !(q4 <= 18) && !(q4 <= 38) && q4 <= 57", formQMinus1},
	{"bid128_fma_body.go/bid_fma_delta_ge_zero/bid_round256_58_76(q4, q4-1, C4)",
		pinnedFmaCase1DoubleRound + " && !(q4 == 1) && !(q4 <= 18) && !(q4 <= 38) && !(q4 <= 57)", formQMinus1},
	{"bid128_fma_body.go/bid_fma_case1ppB_psign_ne_zsign/bid_round192_39_57(q4, q4-1, P192)",
		pinnedFmaCase1ppB + " && !(q4 == 1) && !(q4 <= 18) && !(q4 <= 38) && q4 <= 57", formQMinus1},
	{"bid128_fma_body.go/bid_fma_case1ppB_psign_ne_zsign/bid_round256_58_76(q4, q4-1, C4)",
		pinnedFmaCase1ppB + " && !(q4 == 1) && !(q4 <= 18) && !(q4 <= 38) && !(q4 <= 57)", formQMinus1},
	{"bid128_fma_body.go/bid_fma_cases_2_to_6/bid_round192_39_57(q4, x0, P192)",
		"for && !(x0 == 0) && !(q4 <= 18) && !(q4 <= 38) && q4 <= 57", formCase2Loop},
	{"bid128_fma_body.go/bid_fma_cases_2_to_6/bid_round256_58_76(q4, x0, C4)",
		"for && !(x0 == 0) && !(q4 <= 18) && !(q4 <= 38) && !(q4 <= 57)", formCase2Loop},
	{"bid128_fma_body.go/bid_fma_case7/bid_round192_39_57(q4, x0, P192)",
		"!(q4 <= 38) && q4 <= 57", formQMinusP34},
	{"bid128_fma_body.go/bid_fma_case7/bid_round256_58_76(q4, x0, C4)",
		"!(q4 <= 38) && !(q4 <= 57)", formQMinusP34},
	{"bid128_fma_body.go/bid_fma_cases_11_12/bid_round192_39_57(ind, x0, P192)",
		"!(ind < p34) && !(ind == p34) && !(ind <= 38) && ind <= 57", formQMinusP34},
	{"bid128_fma_body.go/bid_fma_cases_11_12/bid_round256_58_76(ind, x0, R256)",
		"!(ind < p34) && !(ind == p34) && !(ind <= 38) && !(ind <= 57)", formQMinusP34},
	{"bid128_fma_body.go/bid_add_and_round/bid_round192_39_57(ind, x0, P192)",
		"!(ind <= p34) && !(ind <= 38) && ind <= 57", formQMinusP34},
	{"bid128_fma_body.go/bid_add_and_round/bid_round256_58_76(ind, x0, R256)",
		"!(ind <= p34) && !(ind <= 38) && !(ind <= 57)", formQMinusP34},
}

// entrySlots are the argument positions the certificate's (res,C3,q3,q4,C4,delta) premise travels through on every hop
// from bid128_ext_fma down to bid_fma_delta_ge_zero. res is in the list because the k bound reads `res - R128` while
// re-reading C3: if a hop passed C3 into the res slot the two pointers would name one variable, the scaling would
// overwrite C3 before the next iteration re-reads it, and the bound would be about a value the port never holds. Both
// are *BID_UINT128 on every hop, so that swap type-checks and no compiler error would catch it.
var entrySlots = map[int]string{1: "res", 10: "q3", 11: "q4", 14: "delta", 15: "C3", 16: "C4"}

// pinnedMainBodyArgs is the single bid_fma_main_body call of bid128_ext_fma, argument by argument. It also opens the
// aliasing chain: the loop's res and C3 come from `&res` and `&C3`, two distinct locals of bid128_ext_fma, so the
// subtraction the k bound reads cannot be writing through the same storage as C3.
var pinnedMainBodyArgs = []string{"p34", "&res", "&is_midpoint_lt_even", "&is_midpoint_gt_even",
	"&is_inexact_lt_midpoint", "&is_inexact_gt_midpoint", "p_sign", "z_sign", "&z_exp", "&p_exp",
	"q3", "q4", "&e3", "&e4", "delta", "&C3", "C4", "rnd_mode", "pfpsf"}

// pinnedP34Params: every port function receiving p34, with its parameter
var pinnedP34Params = map[string]int{
	"bid_fma_main_body": 0, "bid_fma_delta_ge_zero": 0, "bid_fma_delta_lt_zero": 0,
	"bid_fma_cases_2_to_6": 0, "bid_fma_case7": 0, "bid_fma_cases_11_12": 0,
	"bid_fma_case1ppB_psign_ne_zsign": 0, "bid_add_and_round": 4,
}

// TestBid128RoundBoundaryArmsAreUnreachable certifies four dead regions of
// bid128_round.go, one per audited mutant:
//
//	bid_round192_39_57  ind = q-x >= 40 overflow arm       (bid_ten2k256[ind-39].w0)
//	bid_round256_58_76  ind = x-1 <= 18 midpoint carry     (C.w2 == 0x0)
//	bid_round256_58_76  ind = x-1 <= 18 extraction arm     (P512.w5 >> shift)
//	bid_round256_58_76  40 <= ind = q-x <= 57 overflow arm (bid_ten2k256[ind-39].w0)
//
// All four live in q-x >= 39: the two x-1 arms need x <= 19 while the 256-bit helper's guards force q >= 58, and the two
// overflow arms need q-x >= 40. The certificate reduces to one bound over the closed world of the fourteen production
// calls - q-x <= 34, and x >= 24 whenever bid_round256_58_76 runs - established per x form:
//
//   - x = q-p34 (eight calls): q-x = p34 = 34, and q >= 58 at every 256-bit
//     call, so x = q-34 >= 24.
//   - x = q-1 (four calls): q-x = 1 and x = q-1 >= 57.
//   - the case2_repeat loop of bid_fma_cases_2_to_6 (two calls): x0 starts at delta+q4-p34 and is decremented only by
//     the loop's single back edge - the sole continue of the one unbounded loop, whose body ends in an unlabeled break,
//     so no other path re-enters it - so after k continues q4-x0 = 34-delta+k and x0 = delta+q4-34-k. The x0 = 0
//     initialisations of Cases (3), (5) and (6) reach no helper call: the back edge needs x0 >= 1, so x0 stays 0 and
//     the pinned `!(x0 == 0)` guard excludes the call. With delta >= 0 at k = 0 and delta >= 2 at k = 1, and k <= 1,
//     that leaves q4-x0 <= max(34-0, 34-2+1) = 34 and x0 >= q4-34 >= 24.
//
// The two delta floors come from the dispatch. bid_fma_delta_ge_zero enters Cases (2)-(6) only under
// !(delta <= 1 && p_sign != z_sign) and the back edge sits in the z_sign != p_sign arm, so a continue implies
// delta >= 2; and the loop is reached only through bid_fma_main_body's delta >= 0 arm or through the swap arm of
// bid_fma_delta_lt_zero, which is excluded below, so delta >= 0 throughout.
//
// k <= 1, using only the k = 0 -> 1 step. A second continue would have to fire at iteration 1, where the loop
// recomputes res = C3 * 10^(p34-q3+1) with 10^(q3-1) <= C3 < 10^q3 and 1 <= q3 <= p34, so res is in [10^34, 10^35) and
// its multiply stays inside 128 bits - exact, because the pinned scaling arms use two pinned helpers that return the
// low 128 bits of the true product. The subtrahend is C4 rounded to q4-x0 = 35-delta digits, hence at most
// 10^(35-delta) <= 10^33 for delta >= 2 - the sampled helper postcondition, which applies because 10^(q4-1) <= C4 <
// 10^q4 is established below, normalised by the pinned incr_exp arms, and small enough that the two low words the loop
// copies into R128 hold it whole. So res - R128 >= 9*10^33 > 10^33: the back edge's `res < 10^33` test and its
// `res == 10^33` midpoint alternative are both false. If x0 has instead reached 0, the back edge's own `x0 >= 1`
// conjunct fails. Either way there is no second continue. `res` is a different variable from C3 on every hop - the
// aliasing closure below - so the subtraction the back edge tests is not writing through the C3 the next iteration
// re-reads.
//
// Those two digit-count premises are only needed on the delta >= 0 route, because the swap arm of
// bid_fma_delta_lt_zero hands bid_fma_delta_ge_zero q4 = its own incoming q3 - bid128_ext_fma's digit count of a
// canonical C3, at most p34 = 34 - and both loop calls are guarded by `!(q4 <= 38)`, which is then false. On the
// delta >= 0 route (C3,q3) and (C4,q4) are bid128_ext_fma's own counted pairs, forwarded unchanged:
// bid_fma_main_body never writes them, and bid_fma_delta_ge_zero writes them only after the Cases (2)-(6) call, in a
// function with no loop and only forward gotos. (C4,q4) is counted from the pinned C4 = C1*C2 chain, whose seven arms
// each compute the exact product and answer q1+q2-1 or q1+q2 by comparing it against 10^(q1+q2-1); with
// 1 <= q1,q2 <= 34 counted off two canonical non-zero coefficients that bounds C4 to exactly q4 digits.
func TestBid128RoundBoundaryArmsAreUnreachable(t *testing.T) {
	files := portPkg(t)
	r := loadPortFile(t, "bid128_round.go")
	f192, f256 := r.funcDecl(t, "bid_round192_39_57"), r.funcDecl(t, "bid_round256_58_76")

	// --- anchor: each helper takes (q, x, C) in that order and never touches q or x again, so the q and x the pinned
	// caller-side bounds are about are the ones every `ind = x - 1` / `ind = q - x` below reads. Without the closed write
	// world a reassignment anywhere in the helper - or inside a block the pinned arms do not cover - could move ind into
	// a dead arm's range while every other anchor in this test still matched.
	for _, fd := range []*ast.FuncDecl{f192, f256} {
		if got := paramNames(fd); len(got) != 3 || got[0] != "q" || got[1] != "x" || got[2] != "C" {
			t.Fatalf("bid128_round.go: %s takes %v, pinned (q, x, C)", fd.Name.Name, got)
		}
		for _, v := range []string{"q", "x"} {
			r.pinWrites(t, fd, v, nil)
			r.pinNoNestedDecl(t, fd, v)
		}
		// Source position is used as execution order for the ind chains below, so no loop and no backward goto.
		r.pinForwardOnly(t, fd)
	}

	// --- anchor: ind is exactly these two functions of the arguments, so "ind <= 18" means x <= 19 and "ind >= 40"
	// means q-x >= 40.
	r.pinWrites(t, f192, "ind", []string{"ind = x - 1", "ind = q - x"})
	ind256 := r.pinWrites(t, f256, "ind", []string{"ind = x - 1", "ind = q - x"})
	r.pinNoNestedDecl(t, f192, "ind")
	r.pinNoNestedDecl(t, f256, "ind")

	// --- anchor: dead region 1, the bid_round192_39_57 overflow arm for ind>=40.
	ovf192 := r.asIf(t, "round192 overflow chain", f192.Body.List[r.stmtIndex(t, f192.Body, "ind = q - x")+1])
	r.pinNode(t, "round192 overflow chain head", ovf192.Cond, "ind <= 19")
	for _, want := range []string{"ind == 20", "ind <= 38", "ind == 39"} {
		ovf192 = r.elseIf(t, "round192 overflow arm "+want, ovf192)
		r.pinNode(t, "round192 overflow arm", ovf192.Cond, want)
	}
	r.pinNode(t, "round192 ind>=40 overflow arm", r.elseBlock(t, "round192 overflow tail", ovf192),
		"{ if Cstar.w2 == bid_ten2k256[ind-39].w2 && Cstar.w1 == bid_ten2k256[ind-39].w1 && "+
			"Cstar.w0 == bid_ten2k256[ind-39].w0 { Cstar.w0 = bid_ten2k256[ind-40].w0 "+
			"Cstar.w1 = bid_ten2k256[ind-40].w1 Cstar.w2 = bid_ten2k256[ind-40].w2 incr_exp = 1 } "+
			"else { incr_exp = 0 } }")

	// --- anchor: dead regions 2 and 3, the two ind<=18 arms of bid_round256_58_76.
	mid := r.asIf(t, "round256 midpoint chain", f256.Body.List[r.stmtIndex(t, f256.Body, "ind = x - 1")+1])
	r.pinNode(t, "round256 midpoint chain head", mid.Cond, "ind <= 18")
	r.pinRun(t, "round256 ind<=18 midpoint arm", mid.Body, 0, []string{
		"tmp64 = C.w0", "C.w0 = C.w0 + bid_midpoint64[ind]",
		"if C.w0 < tmp64 { C.w1++ if C.w1 == 0x0 { C.w2++ if C.w2 == 0x0 { C.w3++ } } }"})
	r.pinStmtCount(t, "round256 ind<=18 midpoint arm", mid.Body, 3)
	ext := r.asIf(t, "round256 extraction chain",
		f256.Body.List[r.stmtIndex(t, f256.Body, "shift = int(bid_Ex256m256[ind])")+1])
	r.pinNode(t, "round256 extraction chain head", ext.Cond, "ind <= 18")
	// The extraction arm's "ind" is the x-1 value: the if starts after the only write that can define it and ends
	// before the ind = q-x rewrite.
	if ext.Pos() <= ind256[0].pos || ext.End() >= ind256[1].pos {
		t.Fatalf("bid128_round.go: round256 extraction if (lines %d..%d) not strictly between %q (%d) and %q (%d)",
			r.line(ext), r.fset.Position(ext.End()).Line, ind256[0].text, ind256[0].line, ind256[1].text, ind256[1].line)
	}
	r.pinRun(t, "round256 ind<=18 extraction arm", ext.Body, 0, []string{
		"Cstar.w3 = (P512.w7 >> uint(shift))",
		"Cstar.w2 = (P512.w7 << uint(64-shift)) | (P512.w6 >> uint(shift))",
		"Cstar.w1 = (P512.w6 << uint(64-shift)) | (P512.w5 >> uint(shift))",
		"Cstar.w0 = (P512.w5 << uint(64-shift)) | (P512.w4 >> uint(shift))"})

	// --- anchor: dead region 4, the 40 <= ind <= 57 overflow arm.
	ovf256 := r.asIf(t, "round256 overflow chain", f256.Body.List[r.stmtIndex(t, f256.Body, "ind = q - x")+1])
	r.pinNode(t, "round256 overflow chain head", ovf256.Cond, "ind <= 19")
	for _, want := range []string{"ind == 20", "ind <= 38", "ind == 39", "ind <= 57"} {
		ovf256 = r.elseIf(t, "round256 overflow arm "+want, ovf256)
		r.pinNode(t, "round256 overflow arm", ovf256.Cond, want)
	}
	r.pinStmtCount(t, "round256 40<=ind<=57 overflow arm", ovf256.Body, 1)
	r.pinNode(t, "round256 40<=ind<=57 overflow arm", ovf256.Body.List[0],
		"if Cstar.w3 == 0x0 && Cstar.w2 == bid_ten2k256[ind-39].w2 && "+
			"Cstar.w1 == bid_ten2k256[ind-39].w1 && Cstar.w0 == bid_ten2k256[ind-39].w0 { "+
			"Cstar.w0 = bid_ten2k256[ind-40].w0 Cstar.w1 = bid_ten2k256[ind-40].w1 "+
			"Cstar.w2 = bid_ten2k256[ind-40].w2 incr_exp = 1 } else { incr_exp = 0 }")

	// --- anchor + premise: the closed world of production calls with their guards, and the bound q-x <= 34 (x >= 24 for
	// the 256-bit helper) each one yields under its pinned x form.
	seen := map[string]portCallSite{}
	for _, name := range []string{"bid_round192_39_57", "bid_round256_58_76"} {
		for _, s := range portCallsTo(t, files, name) {
			// Residual: two call sites rendering to the same key would collapse into one map entry and the count below
			// would still match the pinned world, hiding the second site's guards.
			if prev, dup := seen[s.key]; dup {
				t.Fatalf("two production calls share the key %s (guards %q and %q); the closed world would collapse",
					s.key, prev.guards, s.guards)
			}
			seen[s.key] = s
		}
	}
	if len(seen) != len(pinnedRoundCalls) {
		t.Fatalf("calls to the wide rounding helpers: found %d, pinned %d\ngot: %v", len(seen), len(pinnedRoundCalls), seen)
	}
	if P34 != 34 {
		t.Fatalf("P34 = %d; the certificate's p34 == 34 premise is broken", P34)
	}
	const maxContinues, minDeltaOpposite = 1, 2 // both proved by the case2_repeat block below
	for _, c := range pinnedRoundCalls {
		site, ok := seen[c.key]
		if !ok {
			t.Fatalf("pinned call site is gone: %s", c.key)
		}
		got := site.guards
		if got != c.guards {
			t.Fatalf("%s: dominating guards drifted\n got: %s\nwant: %s", c.key, got, c.guards)
		}
		v := strings.TrimSpace(strings.SplitN(c.key[strings.IndexByte(c.key, '(')+1:], ",", 2)[0])
		has := func(s string) bool { return strings.Contains(" "+got+" ", " "+s+" ") }
		if !has("!(" + v + " <= 38)") {
			t.Fatalf("%s: no pinned guard establishes %s >= 39", c.key, v)
		}
		wide := has("!(" + v + " <= 57)")
		if !wide && !has(v+" <= 57") {
			t.Fatalf("%s: no pinned guard bounds %s against 57", c.key, v)
		}
		minQ := 39
		if wide {
			minQ = 58
		}
		qMinusX, minX := P34, minQ-P34
		switch c.form {
		case formQMinusP34:
		case formQMinus1:
			qMinusX, minX = 1, minQ-1
		case formCase2Loop:
			if !has("!(x0 == 0)") {
				t.Fatalf("%s: the loop call is no longer guarded by x0 != 0", c.key)
			}
			// Residual: the guard path only says some enclosing loop exists. The k bound below is about case2_repeat,
			// so require the call to sit inside the single unbounded for{} of bid_fma_cases_2_to_6 itself, located in
			// the same AST the guards were read from.
			if site.fd.Name.Name != "bid_fma_cases_2_to_6" {
				t.Fatalf("%s: the case2_repeat form is pinned to bid_fma_cases_2_to_6, not %s", c.key, site.fd.Name.Name)
			}
			if rep := site.pf.uniqueUnboundedLoop(t, site.fd); !within(site.node, rep.Body) {
				t.Fatalf("%s: the call at line %d is outside the unbounded loop body at line %d",
					c.key, site.pf.line(site.node), site.pf.line(rep))
			}
			if v := P34 - minDeltaOpposite + maxContinues; v > qMinusX {
				qMinusX = v
			}
			if v := minQ - P34 + minDeltaOpposite - maxContinues; v < minX {
				minX = v
			}
		default:
			t.Fatalf("%s: unpinned x form", c.key)
		}
		if qMinusX > 38 {
			t.Fatalf("%s: q-x can reach %d; the ind>=39 dead regions stop being dead", c.key, qMinusX)
		}
		if wide && minX < 20 {
			t.Fatalf("%s: x can fall to %d; the ind<=18 dead regions stop being dead", c.key, minX)
		}
	}

	// --- anchor: p34 == 34 at those call sites. It is written once, from the constant; every consumer takes it in the
	// pinned position, is passed the identifier, and never assigns it.
	fma := loadPortFile(t, "bid128_fma.go")
	extFma := fma.funcDecl(t, "bid128_ext_fma")
	fma.pinWrites(t, extFma, "p34", []string{"p34 := P34", "p34 = P34"})
	fma.pinNoNestedDecl(t, extFma, "p34")
	for name, idx := range pinnedP34Params {
		var decl *ast.FuncDecl
		var dp *portFile
		for _, p := range files {
			for _, fd := range p.decls() {
				if fd.Name.Name == name {
					decl, dp = fd, p
				}
			}
		}
		if decl == nil {
			t.Fatalf("p34 consumer %s not found", name)
		}
		params := paramNames(decl)
		if idx >= len(params) || params[idx] != "p34" {
			t.Fatalf("%s: parameter %d is %v, want p34", name, idx, params)
		}
		dp.pinWrites(t, decl, "p34", nil)
		// Residual: writes() skips a value-less ValueSpec, so `var p34 int` in a nested block would shadow the pinned
		// parameter with a zero while leaving the pinned write world empty and intact.
		dp.pinNoNestedDecl(t, decl, "p34")
		for _, s := range portCallsTo(t, files, name) {
			if idx >= len(s.args) || s.args[idx] != "p34" {
				t.Fatalf("%s: a call passes %v as p34", name, s.args)
			}
			// Residual: the identifier named p34 in the caller must itself be a pinned p34. Only bid128_ext_fma (whose
			// p34 is pinned to P34 above) and the consumers themselves (whose p34 is a never-written parameter) may
			// call a consumer, so a new function holding a local p34 of its own cannot feed the chain.
			if _, ok := pinnedP34Params[s.fd.Name.Name]; !ok && s.fd.Name.Name != extFma.Name.Name {
				t.Fatalf("%s is called from %s/%s, neither %s nor a pinned p34 consumer; its p34 is outside the chain",
					name, s.pf.name, s.fd.Name.Name, extFma.Name.Name)
			}
		}
	}

	// --- anchor: in the four straight-line callers x0 is assigned once from the same q the call passes, neither is
	// touched in between, and that assignment dominates the call - it precedes it in source and every condition guarding
	// the assignment also guards the call - so the call's x0 is that value.
	body := loadPortFile(t, "bid128_fma_body.go")
	for _, c := range []struct {
		p        *portFile
		fn, w, q string
		writes   []string
	}{
		{fma, "bid128_ext_fma", "x0 = q4 - p34", "q4", []string{"x0 = q4 - p34", "x0 = expmin128 - e4"}},
		{body, "bid_fma_case7", "x0 = q4 - p34", "q4", []string{"x0 = q4 - p34"}},
		{body, "bid_fma_cases_11_12", "x0 = ind - p34", "ind", []string{"x0 = e4 - e3", "x0 = ind - p34", "x0 = expmin128 - e4"}},
		{body, "bid_add_and_round", "x0 = ind - p34", "ind", []string{"x0 = ind - p34", "x0 = expmin128 - e4"}},
	} {
		decl := c.p.funcDecl(t, c.fn)
		c.p.pinWrites(t, decl, "x0", c.writes)
		// Residual: writes() skips a value-less ValueSpec, so a nested `var x0 int` between the pinned assignment and the
		// call - or a nested `var q4 int` / `var ind int` around either - would rebind the name to a zero the pinned
		// assignment never wrote, while pinWrites and noWriteBetween both still passed. The x0 and the q the call reads
		// have to be the function's own top-level binding.
		c.p.pinNoNestedDecl(t, decl, "x0")
		c.p.pinNoNestedDecl(t, decl, c.q)
		// Residual: the dominance argument below is exactly "precedes in source" plus "guarded by a conjunct prefix",
		// and pinForwardOnly - which all four of these functions are held to - still admits a forward goto. A jump from
		// ahead of the pinned x0 assignment to a label between it and the call satisfies both halves while leaving x0
		// unset by that assignment, so these four are pinned goto-free instead.
		c.p.pinNoGoto(t, decl)
		checked := 0
		for _, name := range []string{"bid_round192_39_57", "bid_round256_58_76"} {
			for _, s := range portCallsTo(t, files, name) {
				if s.pf.name != c.p.name || s.fd.Name.Name != c.fn {
					continue
				}
				from := s.pf.stmtWithText(t, s.fd, c.w)
				if from.End() > s.node.Pos() {
					t.Fatalf("%s/%s: %q (line %d) no longer precedes the call at line %d", s.pf.name, c.fn, c.w, s.pf.line(from), s.pf.line(s.node))
				}
				if g := s.pf.guardPath(t, s.fd, from); !conjunctPrefix(g, s.guards) {
					t.Fatalf("%s/%s: %q guarded by %q, not a conjunct prefix of %q", s.pf.name, c.fn, c.w, g, s.guards)
				}
				s.pf.noWriteBetween(t, s.fd, c.q, from, s.node)
				s.pf.noWriteBetween(t, s.fd, "x0", from, s.node)
				checked++
			}
		}
		if checked != 2 {
			t.Fatalf("%s/%s: checked %d wide-helper calls, pinned 2", c.p.name, c.fn, checked)
		}
	}

	// --- the case2_repeat loop: closed inputs, three loops (case2_repeat plus two finite loops), one back edge.
	loop := body.funcDecl(t, "bid_fma_cases_2_to_6")
	for _, v := range []string{"q3", "q4", "delta", "p34", "C3", "p_sign"} {
		body.pinWrites(t, loop, v, nil)
		// Residual: writes() skips a value-less ValueSpec, so a nested `var q3 int` would rebind the name the loop's
		// arithmetic reads - to a zero, and to storage the caller never wrote - while leaving this write world empty
		// and intact. Every name the k bound and the arm selection read has to be the incoming parameter.
		body.pinNoNestedDecl(t, loop, v)
	}
	x0w := body.pinWrites(t, loop, "x0", []string{"x0 = delta + q4 - p34", "x0 = 0", "x0 = 0", "x0 = x0 - 1", "x0 = expmin128 - e3"})
	scw := body.pinWrites(t, loop, "scale", []string{"scale = p34 - q3", "scale = q3 - delta - q4", "scale = 0",
		"scale = delta + q4 - q3", "scale = scale + 1"})
	body.pinNoNestedDecl(t, loop, "x0")
	body.pinNoNestedDecl(t, loop, "scale")
	// C4 is a by-value parameter here, so the loop's own rewrite of it cannot reach the caller - but it does reach the
	// loop's later iterations, and the k bound reads C4 as the p34-digit-bounded product bid128_ext_fma counted. The two
	// writes are pinned into the Case (6) arm below, which also sets x0 = 0 and so cannot reach the back edge.
	c4w := body.pinWrites(t, loop, "C4", []string{"C4.w0 = P128.lo", "C4.w1 = P128.hi"})
	body.pinNoNestedDecl(t, loop, "C4")
	repeat := body.uniqueUnboundedLoop(t, loop) // case2_repeat
	loops, continues, gotos := 0, 0, 0
	var cont ast.Stmt
	ast.Inspect(loop, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.ForStmt, *ast.RangeStmt:
			loops++
		case *ast.BranchStmt:
			switch s.Tok {
			case token.CONTINUE:
				continues++
				cont = s
			case token.GOTO:
				// Residual: a goto is a back edge this accounting cannot see - it could re-enter the loop body
				// without executing the pinned continue, breaking the k bound. Pinned to none.
				gotos++
			}
		}
		return true
	})
	if loops != 3 || continues != 1 || gotos != 0 {
		t.Fatalf("bid_fma_cases_2_to_6: %d loops, %d continues, %d gotos; pinned 3, 1, 0", loops, continues, gotos)
	}
	// The sole continue belongs to that unbounded loop, and the loop body's last top-level statement is an unlabeled
	// break, so control cannot fall off the end of the body into another iteration: every extra iteration is one
	// execution of the single pinned back edge, which is what bounds k below.
	if cont.Pos() < repeat.Body.Pos() || cont.Pos() >= repeat.Body.End() {
		t.Fatalf("bid_fma_cases_2_to_6: sole continue (line %d) outside the unbounded loop at line %d", body.line(cont), body.line(repeat))
	}
	tail := repeat.Body.List[len(repeat.Body.List)-1]
	if br, ok := tail.(*ast.BranchStmt); !ok || br.Tok != token.BREAK || br.Label != nil {
		t.Fatalf("bid_fma_cases_2_to_6: case2_repeat body no longer ends in an unlabeled break; last (line %d) is %s",
			body.line(tail), body.flat(t, tail))
	}
	const backEdge = "e3 > expmin128 && ((res.hi < 0x0000314dc6448d93 || " +
		"(res.hi == 0x0000314dc6448d93 && res.lo < 0x38c15b0a00000000)) || " +
		"((is_inexact_lt_midpoint|is_midpoint_gt_even) != 0 && res.hi == 0x0000314dc6448d93 && " +
		"res.lo == 0x38c15b0a00000000)) && x0 >= 1"
	if got := body.guardPath(t, loop, cont); got != "for && !(z_sign == p_sign) && "+backEdge {
		t.Fatalf("case2_repeat back edge guards drifted: %s", got)
	}
	pinnedBackEdgeIf := "if " + backEdge + " { x0 = x0 - 1 e3 = e3 + scale scale = scale + 1 " +
		"is_inexact_lt_midpoint = 0 is_inexact_gt_midpoint = 0 is_midpoint_lt_even = 0 " +
		"is_midpoint_gt_even = 0 incr_exp = 0 continue }"
	body.pinNode(t, "case2_repeat back edge", body.ifWithCond(t, loop, backEdge), pinnedBackEdgeIf)
	// --- anchor: the loop body's own shape, so `res` at the back-edge test really is the scaled C3 minus the rounded C4,
	// and so `x0` at a wide-helper call is either the pre-loop initialisation or the back edge's decrement. The four
	// pinned top-level statements run in order; statements 1 and 2 touch no res; the opposite-sign arm subtracts R128
	// from res and then tests it; and the sole later write to x0 sits after the back edge, in an iteration whose only
	// remaining exit is the pinned tail break.
	if len(repeat.Body.List) < 4 {
		t.Fatalf("case2_repeat has %d top-level statements, too few to pin its shape", len(repeat.Body.List))
	}
	body.pinNode(t, "case2_repeat exponent adjust", repeat.Body.List[1], "e3 = e3 - scale")
	roundChain := body.asIf(t, "case2_repeat C4 rounding chain", repeat.Body.List[2])
	body.pinNode(t, "case2_repeat C4 rounding chain head", roundChain.Cond, "x0 == 0")
	// --- anchor: that chain arm by arm. The k <= 1 step reads R128 as C4 rounded to q4-x0 digits and bounded by
	// 10^(q4-x0), so both halves of the reading have to be pinned: which q4 class selects which arm, and how each arm's
	// result reaches R128. The x0 == 0 arm copies C4 itself, which is the same bound at q4-x0 = q4; the two narrow arms
	// leave the rounded value in R128 directly; and the two wide arms copy only the low two words of R192 and R256,
	// which is the whole rounded value because the subtrahend bound checked at the end of this test fits 128 bits. Text
	// pins on the incr_exp normalisations alone would leave an arm free to move its q4 bound, round a different operand,
	// or extract the wrong words, with every other anchor in this certificate still matching.
	incrExpNorm := func(lo, hi string) string {
		return "if incr_exp != 0 { if q4-x0 <= 19 { " + lo + " = bid_ten2k64[q4-x0] } else { " +
			lo + " = bid_ten2k128[q4-x0-20].lo " + hi + " = bid_ten2k128[q4-x0-20].hi } }"
	}
	body.pinRun(t, "case2_repeat x0 == 0 arm", roundChain.Body, 0, []string{"R128.hi = C4.w1", "R128.lo = C4.w0"})
	body.pinStmtCount(t, "case2_repeat x0 == 0 arm", roundChain.Body, 2)
	roundArm := roundChain
	for _, a := range []struct {
		cond  string
		stmts []string
	}{
		{"q4 <= 18", []string{
			"R64 = bid_round64_2_18(q4, x0, C4.w0, &incr_exp, &is_midpoint_lt_even, &is_midpoint_gt_even, " +
				"&is_inexact_lt_midpoint, &is_inexact_gt_midpoint)",
			"if incr_exp != 0 { R64 = bid_ten2k64[q4-x0] }",
			"R128.hi = 0", "R128.lo = R64"}},
		{"q4 <= 38", []string{"P128.hi = C4.w1", "P128.lo = C4.w0",
			"R128, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, " +
				"is_inexact_gt_midpoint = bid_round128_19_38(q4, x0, P128)",
			incrExpNorm("R128.lo", "R128.hi")}},
		{"q4 <= 57", []string{"P192.w2 = C4.w2", "P192.w1 = C4.w1", "P192.w0 = C4.w0",
			"R192, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, " +
				"is_inexact_gt_midpoint = bid_round192_39_57(q4, x0, P192)",
			incrExpNorm("R192.w0", "R192.w1"), "R128.hi = R192.w1", "R128.lo = R192.w0"}},
	} {
		roundArm = body.elseIf(t, "case2_repeat rounding arm "+a.cond, roundArm)
		body.pinNode(t, "case2_repeat rounding arm guard", roundArm.Cond, a.cond)
		body.pinRun(t, "case2_repeat rounding arm "+a.cond, roundArm.Body, 0, a.stmts)
		body.pinStmtCount(t, "case2_repeat rounding arm "+a.cond, roundArm.Body, len(a.stmts))
	}
	roundTail := body.elseBlock(t, "case2_repeat rounding chain tail", roundArm)
	body.pinRun(t, "case2_repeat rounding chain tail", roundTail, 0, []string{
		"R256, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, " +
			"is_inexact_gt_midpoint = bid_round256_58_76(q4, x0, C4)",
		incrExpNorm("R256.w0", "R256.w1"), "R128.hi = R256.w1", "R128.lo = R256.w0"})
	body.pinStmtCount(t, "case2_repeat rounding chain tail", roundTail, 4)
	// --- anchor: the closed world of writes to the subtrahend, every one of them inside that chain. Without it a
	// redefinition of R128 between the chain and the subtraction would leave the arm pins above describing a value the
	// back edge no longer tests.
	for _, w := range body.pinWrites(t, loop, "R128", []string{
		"R128.hi = C4.w1", "R128.lo = C4.w0", "R128.hi = 0", "R128.lo = R64",
		"R128, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, " +
			"is_inexact_gt_midpoint = bid_round128_19_38(q4, x0, P128)",
		"R128.lo = bid_ten2k64[q4-x0]", "R128.lo = bid_ten2k128[q4-x0-20].lo",
		"R128.hi = bid_ten2k128[q4-x0-20].hi",
		"R128.hi = R192.w1", "R128.lo = R192.w0", "R128.hi = R256.w1", "R128.lo = R256.w0"}) {
		if w.pos < roundChain.Pos() || w.pos >= roundChain.End() {
			t.Fatalf("bid_fma_cases_2_to_6: R128 is written at line %d, outside the C4 rounding chain: %s", w.line, w.text)
		}
	}
	body.pinNoNestedDecl(t, loop, "R128")
	signIf := body.asIf(t, "case2_repeat sign split", repeat.Body.List[3])
	body.pinNode(t, "case2_repeat sign split head", signIf.Cond, "z_sign == p_sign")
	for _, n := range []ast.Node{repeat.Body.List[1], repeat.Body.List[2]} {
		if w := body.writes(t, n, "res"); len(w) != 0 {
			t.Fatalf("case2_repeat: res is written at line %d between the scaling and the sign split: %s", w[0].line, w[0].text)
		}
	}
	oppose := body.elseBlock(t, "case2_repeat opposite-sign arm", signIf)
	body.pinRun(t, "case2_repeat opposite-sign subtraction", oppose, 0,
		[]string{"lsb = res.lo & 0x01", "tmp64 = res.lo", "res.lo = res.lo - R128.lo", "res.hi = res.hi - R128.hi",
			"if res.lo > tmp64 { res.hi-- }", pinnedBackEdgeIf})
	backEdgeIf := oppose.List[5]
	for i, w := range x0w {
		switch {
		case i == 3 && (w.pos < backEdgeIf.Pos() || w.pos >= backEdgeIf.End()):
			t.Fatalf("case2_repeat: the x0 decrement at line %d left the back edge: %s", w.line, w.text)
		case i == 4 && w.pos < backEdgeIf.End():
			t.Fatalf("case2_repeat: the tiny-path x0 write at line %d is no longer after the back edge: %s", w.line, w.text)
		}
	}
	// --- anchor + premise: the x0 = delta+q4-p34 initialisation belongs to the Cases (2)/(4) arm, the statement just
	// ahead of case2_repeat. Both disjuncts of its pinned guard bound delta below p34 - the first says so outright, the
	// second through delta < q3 <= p34 - so the back edge's route carries 2 <= delta <= p34-1, which is what keeps the
	// subtrahend bound below finite.
	repIdx := -1
	for i, s := range loop.Body.List {
		if ast.Stmt(repeat) == s {
			repIdx = i
		}
	}
	if repIdx < 1 {
		t.Fatalf("case2_repeat sits at top-level index %d, so the initialisation ahead of it is gone", repIdx)
	}
	initIf := body.asIf(t, "Cases (2)/(4) initialisation", loop.Body.List[repIdx-1])
	body.pinNode(t, "Cases (2)/(4) guard", initIf.Cond,
		"(q3 <= delta && delta < p34 && p34 < delta+q4) || (delta < q3 && p34 < delta+q4)")
	body.pinStmtCount(t, "Cases (2)/(4) initialisation", initIf.Body, 2)
	body.pinRun(t, "Cases (2)/(4) initialisation", initIf.Body, 0, []string{"scale = p34 - q3", "x0 = delta + q4 - p34"})

	// --- anchor: where each (scale, x0) write sits, not just what it says. The three pre-loop writes belong to the three
	// arms of the one if/else chain immediately ahead of case2_repeat, so exactly one of them runs and the loop's first
	// iteration reads it; the fourth pair is the back edge's own; and the fifth x0 write is the tiny path after it. Text
	// pins alone leave a scale assignment free to move to another arm, out of the chain, or into the loop body, which
	// would change the value the pinned `if scale == 0 { ... }` scaling reads on the Cases (2)/(4) route while every
	// text pin still matched.
	case6 := body.elseIf(t, "Case (6) arm", initIf)
	body.pinNode(t, "Case (6) guard", case6.Cond, "delta+q4 < q3")
	body.pinRun(t, "Case (6) head", case6.Body, 0, []string{"scale = q3 - delta - q4"})
	body.pinRun(t, "Case (6) tail", case6.Body, len(case6.Body.List)-2, []string{"scale = 0", "x0 = 0"})
	case35 := body.elseBlock(t, "Cases (3)/(5) arm", case6)
	body.pinStmtCount(t, "Cases (3)/(5) arm", case35, 2)
	body.pinRun(t, "Cases (3)/(5) arm", case35, 0, []string{"scale = delta + q4 - q3", "x0 = 0"})
	inside := func(w portWrite, n ast.Node) bool { return w.pos >= n.Pos() && w.pos < n.End() }
	for _, c := range []struct {
		label string
		w     portWrite
		in    ast.Node
	}{
		{"Cases (2)/(4) scale", scw[0], initIf.Body}, {"Case (6) scale", scw[1], case6.Body},
		{"Case (6) reset scale", scw[2], case6.Body}, {"Cases (3)/(5) scale", scw[3], case35},
		{"Cases (2)/(4) x0", x0w[0], initIf.Body}, {"Case (6) x0", x0w[1], case6.Body},
		{"Cases (3)/(5) x0", x0w[2], case35},
		{"Case (6) C4 low word", c4w[0], case6.Body}, {"Case (6) C4 high word", c4w[1], case6.Body},
	} {
		if !inside(c.w, c.in) {
			t.Fatalf("bid_fma_cases_2_to_6: the %s write at line %d left the arm at line %d: %s",
				c.label, c.w.line, body.line(c.in), c.w.text)
		}
	}
	if !inside(scw[4], backEdgeIf) {
		t.Fatalf("bid_fma_cases_2_to_6: the scale increment at line %d left the back edge: %s", scw[4].line, scw[4].text)
	}
	// The chain is the statement immediately ahead of the loop and the scaling is the loop's first statement, so nothing
	// runs between the selected arm and the read of scale. Its three arms are disjoint, so the Cases (2)/(4) route never
	// executes the Case (6) rewrite of C4 or either `x0 = 0`.
	if initIf.End() > repeat.Pos() {
		t.Fatalf("bid_fma_cases_2_to_6: the initialisation chain no longer ends before case2_repeat at line %d",
			body.line(repeat))
	}
	body.pinWrites(t, loop, "z_sign", []string{"z_sign = 0x0000000000000000", "z_sign = 0x8000000000000000",
		"z_sign = 0x0000000000000000", "z_sign = 0x8000000000000000"})
	body.pinNoNestedDecl(t, loop, "z_sign")
	var zeroBlocks []*ast.IfStmt
	ast.Inspect(loop, func(n ast.Node) bool {
		if is, ok := n.(*ast.IfStmt); ok && body.flat(t, is.Cond) == "res.hi == 0x0 && res.lo == 0x0" {
			zeroBlocks = append(zeroBlocks, is)
		}
		return true
	})
	if len(zeroBlocks) != 2 {
		t.Fatalf("bid_fma_cases_2_to_6 has %d pure-zero blocks, pinned 2", len(zeroBlocks))
	}
	for _, zb := range zeroBlocks {
		if _, ok := zb.Body.List[len(zb.Body.List)-1].(*ast.ReturnStmt); !ok {
			t.Fatalf("bid_fma_cases_2_to_6: the pure-zero block at line %d no longer returns", body.line(zb))
		}
	}
	for _, w := range body.writes(t, loop, "z_sign") {
		if (w.pos < zeroBlocks[0].Pos() || w.pos > zeroBlocks[0].End()) &&
			(w.pos < zeroBlocks[1].Pos() || w.pos > zeroBlocks[1].End()) {
			t.Fatalf("bid_fma_cases_2_to_6: z_sign is written at line %d outside the returning pure-zero blocks", w.line)
		}
	}
	// --- anchor: the dispatcher guard that puts delta >= 2 on that route.
	const dispatchInnerCond = "((q3 <= delta && delta < p34 && p34 < delta+q4) || (q3 <= delta && delta+q4 <= p34) || " +
		"(delta < q3 && p34 < delta+q4) || (delta < q3 && q3 <= delta+q4 && delta+q4 <= p34) || " +
		"(delta+q4 < q3)) && !(delta <= 1 && p_sign != z_sign)"
	const dispatchGuards = "!(p34 <= delta-1 || (p34 == delta && e3+6176 < p34-q3)) && !(p34 == delta) && " +
		dispatchInnerCond
	dispatchArgs := []string{"p34", "res", "&is_midpoint_lt_even", "&is_midpoint_gt_even",
		"&is_inexact_lt_midpoint", "&is_inexact_gt_midpoint", "p_sign", "z_sign", "&z_exp",
		"q3", "q4", "&e3", "&e4", "delta", "C3", "C4", "rnd_mode", "pfpsf"}
	d := portCallsTo(t, files, "bid_fma_cases_2_to_6")
	if len(d) != 1 {
		t.Fatalf("bid_fma_cases_2_to_6 has %d call sites, pinned 1: %v", len(d), d)
	}
	if want := "bid128_fma_body.go/bid_fma_delta_ge_zero/bid_fma_cases_2_to_6(" +
		strings.Join(dispatchArgs, ", ") + ")"; d[0].key != want {
		t.Fatalf("Cases (2)-(6) dispatch call drifted\n got: %s\nwant: %s", d[0].key, want)
	}
	if got := strings.Join(d[0].args, "|"); got != strings.Join(dispatchArgs, "|") {
		t.Fatalf("Cases (2)-(6) dispatch arguments drifted\n got: %s\nwant: %s", got, strings.Join(dispatchArgs, "|"))
	}
	if d[0].guards != dispatchGuards {
		t.Fatalf("Cases (2)-(6) dispatch guards drifted\n got: %s\nwant: %s", d[0].guards, dispatchGuards)
	}
	// The condition's last conjunct is what turns the back edge's z_sign != p_sign into delta >= 2.
	if !strings.HasSuffix(dispatchInnerCond, " && !(delta <= 1 && p_sign != z_sign)") {
		t.Fatalf("the dispatch condition no longer ends in the opposite-sign delta floor: %s", dispatchInnerCond)
	}
	// Residual: the guard chain above is evaluated where the conditions are, not where the call is. Locate the innermost
	// dominating if in the same AST - guardPath's own rendering makes it the last conjunct of the pinned chain - and
	// require the call to sit in its body with none of the tested values rewritten in between, so the delta, signs, q3,
	// q4 and C3 the loop receives are the ones the dispatch condition classified.
	if !strings.HasSuffix(dispatchGuards, " && "+dispatchInnerCond) {
		t.Fatalf("the pinned dispatch chain no longer ends in the innermost dispatch condition")
	}
	disp := d[0].pf.ifWithCond(t, d[0].fd, dispatchInnerCond)
	if !within(d[0].node, disp.Body) {
		t.Fatalf("the Cases (2)-(6) call at line %d is not inside the body of the dispatch if at line %d",
			d[0].pf.line(d[0].node), d[0].pf.line(disp))
	}
	for _, v := range []string{"p_sign", "z_sign"} { // q3, q4, C3 and delta are closed by the route anchors below
		d[0].pf.noWriteBetween(t, d[0].fd, v, disp.Cond, d[0].node)
		d[0].pf.pinNoNestedDecl(t, d[0].fd, v)
	}
	// The values the loop's arithmetic reads must be the dispatcher's own, so each is passed by name into the parameter of
	// the same name. res is in the list for the same reason it is in entrySlots: passing C3 into the res slot type-checks
	// and would alias the two.
	loopParams := paramNames(loop)
	for _, kv := range []struct {
		i    int
		name string
	}{{1, "res"}, {6, "p_sign"}, {7, "z_sign"}, {9, "q3"}, {10, "q4"}, {13, "delta"}, {14, "C3"}, {15, "C4"}} {
		if kv.i >= len(loopParams) {
			t.Fatalf("bid_fma_cases_2_to_6 takes %d parameters, so slot %d (%s) is gone", len(loopParams), kv.i, kv.name)
		}
		if loopParams[kv.i] != kv.name || d[0].args[kv.i] != kv.name {
			t.Fatalf("slot %d: caller passes %q into parameter %q, want %q on both sides", kv.i, d[0].args[kv.i], loopParams[kv.i], kv.name)
		}
	}
	// --- anchor: the loop's C3 scaling and the incr_exp normalisation of every rounded C4, the two shapes the k <= 1
	// bounds are read from.
	scaling := body.ifWithCond(t, loop, "scale == 0")
	if ast.Stmt(scaling) != repeat.Body.List[0] {
		t.Fatalf("the C3 scaling at line %d is no longer the first statement of case2_repeat", body.line(scaling))
	}
	body.pinNode(t, "case2_repeat C3 scaling", scaling,
		"if scale == 0 { res.hi = C3.hi res.lo = C3.lo } else if q3 <= 19 { "+
			"if scale <= 19 { *res = __mul_64x64_to_128(C3.lo, bid_ten2k64[scale]) } "+
			"else { *res = __mul_128x64_to_128(C3.lo, bid_ten2k128[scale-20]) } } "+
			"else { *res = __mul_128x64_to_128(bid_ten2k64[scale], *C3) }")
	body.stmtWithText(t, loop, "if incr_exp != 0 { R64 = bid_ten2k64[q4-x0] }")
	for _, w := range [][2]string{{"R128.lo", "R128.hi"}, {"R192.w0", "R192.w1"}, {"R256.w0", "R256.w1"}} {
		body.stmtWithText(t, loop, incrExpNorm(w[0], w[1]))
	}

	// --- premise: C3 has exactly q3 digits with 1 <= q3 <= 34. C3 is loaded from z's coefficient field and then either
	// zeroed or left canonical by the pinned unpacking - the non-canonical zeroing sits in the arm reached when z is
	// finite and not the shifted-exponent form, and the other arm zeroes C3 outright. z is finite at the digit count
	// because the infinity chain's three arms all return, so falling through it means none of x, y, z is inf or NaN. q3
	// is then the port's own digit count of a coefficient in [1, 10^34-1], and a zero C3 returns before the main body.
	rawC3 := fma.stmtWithText(t, extFma, "C3.hi = z.hi & MASK_COEFF_128")
	loC3 := fma.stmtWithText(t, extFma, "C3.lo = z.lo")
	pinnedZUnpack := "if (z.hi & MASK_ANY_INF_128) != MASK_INF_128 { " +
		"if (z.hi & 0x6000000000000000) == 0x6000000000000000 { z_exp = (z.hi << 2) & MASK_EXP_128 C3.hi = 0 C3.lo = 0 } " +
		"else { z_exp = z.hi & MASK_EXP_128 if C3.hi > 0x0001ed09bead87c0 || " +
		"(C3.hi == 0x0001ed09bead87c0 && C3.lo > 0x378d8e63ffffffff) { C3.hi = 0 C3.lo = 0 } } }"
	zUnpack := fma.stmtWithText(t, extFma, pinnedZUnpack)
	if g1, g2 := fma.guardPath(t, extFma, rawC3), fma.guardPath(t, extFma, zUnpack); g1 != "" || g2 != "" ||
		loC3.Pos() < rawC3.End() || loC3.End() > zUnpack.Pos() {
		t.Fatalf("bid128_fma.go: the C3 load (line %d), its low word (line %d) and the unpacking (line %d) are no longer "+
			"consecutive top-level statements (guarded by %q and %q)",
			fma.line(rawC3), fma.line(loC3), fma.line(zUnpack), g1, g2)
	}
	fma.pinWrites(t, extFma, "q3", []string{"q3 = bid128_count_digits(C3)"})
	arm := fma.ifWithCond(t, extFma, "(x.hi & MASK_ANY_INF_128) == MASK_INF_128")
	if got := fma.guardPath(t, extFma, arm); got != "" || arm.Pos() < zUnpack.End() {
		t.Fatalf("bid128_fma.go: the infinity chain at line %d is not a top-level statement after the z unpacking (guarded by %q)",
			fma.line(arm), got)
	}
	infEnd := arm.End()
	for i, want := range []string{"(x.hi & MASK_ANY_INF_128) == MASK_INF_128",
		"(y.hi & MASK_ANY_INF_128) == MASK_INF_128", "(z.hi & MASK_ANY_INF_128) == MASK_INF_128"} {
		fma.pinNode(t, "infinity chain arm", arm.Cond, want)
		if _, ok := arm.Body.List[len(arm.Body.List)-1].(*ast.ReturnStmt); !ok {
			t.Fatalf("bid128_fma.go: infinity arm %d (line %d) no longer ends in a return", i, fma.line(arm))
		}
		if i < 2 {
			arm = fma.elseIf(t, "infinity chain", arm)
			continue
		}
		if arm.Else != nil {
			t.Fatalf("bid128_fma.go: the infinity chain gained a trailing else at line %d", fma.line(arm))
		}
	}
	// --- residual: C3 is written only by the pinned load and unpacking, by the top-level zero-C3 block that always
	// returns, and by the address-of at the main-body call itself. So the C3 (and hence q3) reaching bid_fma_main_body
	// is the counted one: no nonzero-path mutation and no pointer escape can change it in between.
	zeroC3 := fma.ifWithCond(t, extFma, "C3.hi == 0x0 && C3.lo == 0x0")
	mb := portCallsTo(t, files, "bid_fma_main_body")
	if len(mb) != 1 || mb[0].pf.name != fma.name || mb[0].fd.Name.Name != extFma.Name.Name || mb[0].guards != "" {
		t.Fatalf("bid_fma_main_body is no longer called exactly once, unguarded, from bid128_ext_fma: %v", mb)
	}
	if _, ok := zeroC3.Body.List[len(zeroC3.Body.List)-1].(*ast.ReturnStmt); !ok {
		t.Fatalf("bid128_ext_fma: the zero-C3 block at line %d no longer returns", fma.line(zeroC3))
	}
	q3Count := fma.stmtWithText(t, extFma, "q3 = bid128_count_digits(C3)")
	if g := fma.guardPath(t, extFma, zeroC3); g != "" || q3Count.Pos() < infEnd || mb[0].node.Pos() < zeroC3.End() {
		t.Fatalf("bid128_fma.go: the infinity chain, the digit count (line %d), the zero-C3 block (line %d, guarded by %q) "+
			"and the main-body call (line %d) are out of order",
			fma.line(q3Count), fma.line(zeroC3), g, fma.line(mb[0].node))
	}
	for _, w := range fma.writes(t, extFma, "C3") {
		switch {
		case w.pos >= rawC3.Pos() && w.pos < zUnpack.End(): // the pinned load and unpacking, ahead of the count
		case w.pos >= zeroC3.Pos() && w.pos < zeroC3.End(): // inside the block that returns
		case w.pos >= mb[0].node.Pos(): // the address handed over at the call itself
		default:
			t.Fatalf("bid128_ext_fma: C3 is written at line %d outside the pinned load, unpacking and zero-C3 block: %s",
				w.line, w.text)
		}
	}
	fma.pinNoNestedDecl(t, extFma, "C3")
	fma.pinNoNestedDecl(t, extFma, "q3")

	// --- premise: the port's own digit count is exact for every C in [1, 10^34-1]. Its pinned body picks a bid_nr_digits
	// row from the value's bit length - each float64 in it converts an integer below 2^53 (C.lo>>32 < 2^32 on one arm,
	// C.lo < 2^53 on the next, C.hi < 2^49 for a coefficient below 2^113 on the last), so the exponent field is exactly
	// the bit length minus one - and inside a row it answers with a constant or with one threshold split. Checking each
	// row's endpoints, and where it splits the two values either side of the threshold, against the decade boundaries
	// therefore closes every value of the interval rather than a sample of it.
	helpers := loadPortFile(t, "bid128_fma_helpers.go")
	helpers.pinNode(t, "bid128_count_digits", helpers.funcDecl(t, "bid128_count_digits").Body, pinnedCountDigits)
	if maxC := new(big.Int).Sub(pow10big(34), big.NewInt(1)); maxC.BitLen() != 113 || len(bid_nr_digits) < 113 {
		t.Fatalf("10^34-1 spans %d bits and bid_nr_digits has %d rows; the row sweep is out of range",
			maxC.BitLen(), len(bid_nr_digits))
	}
	one := big.NewInt(1)
	for b := 1; b <= 113; b++ {
		lo, hi := pow2big(b-1), new(big.Int).Sub(pow2big(b), one)
		e := bid_nr_digits[b-1]
		probes, wants := []*big.Int{lo, hi}, []int{int(e.digits), int(e.digits)}
		if e.digits != 0 {
			if q := int(e.digits); pow10big(q-1).Cmp(lo) > 0 || pow10big(q).Cmp(hi) <= 0 {
				t.Fatalf("bid_nr_digits[%d].digits = %d, but [2^%d, 2^%d) is not inside [10^%d, 10^%d)", b-1, q, b-1, b, q-1, q)
			}
		} else {
			q := int(e.digits1)
			th := wordsBig([]uint64{e.threshold_lo, e.threshold_hi})
			if th.Cmp(pow10big(q)) != 0 || th.Cmp(lo) <= 0 || th.Cmp(hi) > 0 ||
				pow10big(q-1).Cmp(lo) > 0 || pow10big(q+1).Cmp(hi) <= 0 {
				t.Fatalf("bid_nr_digits[%d] splits [2^%d, 2^%d) at %s, not at a 10^%d boundary inside [10^%d, 10^%d)",
					b-1, b-1, b, th, q, q-1, q+1)
			}
			probes = append(probes, new(big.Int).Sub(th, one), th)
			wants = []int{q, q + 1, q, q + 1}
		}
		for i, c := range probes {
			if got := bid128_count_digits(big128(c)); got != wants[i] {
				t.Fatalf("bid128_count_digits(%s) = %d, want %d", c, got, wants[i])
			}
		}
	}
	if q := bid128_count_digits(big128(new(big.Int).Sub(pow10big(34), big.NewInt(1)))); q != P34 {
		t.Fatalf("bid128_count_digits(10^34-1) = %d, want %d; q3 <= p34 is broken", q, P34)
	}

	// --- premise: 10^(q4-1) <= C4 < 10^q4 for the pair that reaches the loop on the delta >= 0 route. The rounding
	// helpers round a q-digit C to q-x digits, so the subtrahend bound the k <= 1 step uses is only about C4 if q4 really
	// is C4's digit count; a q4 one too small would let the helper return more than 10^(q4-x0) and the bound would say
	// nothing. C1 and C2 are unpacked exactly like C3 and are non-zero once the pinned zero-operand block has returned, so
	// q1 and q2 are the digit counts of two coefficients in [1, 10^34-1]; hence 10^(s-2) <= C1*C2 < 10^s for s = q1+q2.
	// Each pinned arm below computes C4 = C1*C2 exactly - the products stay inside the words the arm writes, and the
	// multiply helpers are exact by the pinned bodies further down - and then answers q4 = s-1 when C4 < 10^(s-1) and
	// q4 = s otherwise. Both answers satisfy the two bounds: the comparison supplies the tight side and the product range
	// the other.
	unpackAs := func(v, c, e string) string {
		return strings.NewReplacer("z.hi", v+".hi", "z_exp", e, "C3", c).Replace(pinnedZUnpack)
	}
	if unpackAs("z", "C3", "z_exp") != pinnedZUnpack {
		t.Fatal("the unpacking template is not the identity on z, so the C1 and C2 forms it derives are not the pinned shape")
	}
	// The bound the unpacking enforces is 10^34-1, which is what puts q1, q2 and q3 in [1, p34]: the pinned text reads the
	// constant as two hex words and nothing else in this file checks what they mean.
	if !strings.Contains(pinnedZUnpack,
		"C3.hi > 0x0001ed09bead87c0 || (C3.hi == 0x0001ed09bead87c0 && C3.lo > 0x378d8e63ffffffff)") {
		t.Fatal("the pinned unpacking no longer zeroes a coefficient above its canonical bound")
	}
	if got := wordsBig([]uint64{0x378d8e63ffffffff, 0x0001ed09bead87c0}); got.Cmp(new(big.Int).Sub(pow10big(P34), one)) != 0 {
		t.Fatalf("the unpacking's canonical bound is %s, not 10^%d-1", got, P34)
	}
	var opCount [2]ast.Stmt
	for i, o := range []struct{ v, c, e, q string }{{"x", "C1", "x_exp", "q1"}, {"y", "C2", "y_exp", "q2"}} {
		hi := fma.stmtWithText(t, extFma, o.c+".hi = "+o.v+".hi & MASK_COEFF_128")
		lo := fma.stmtWithText(t, extFma, o.c+".lo = "+o.v+".lo")
		up := fma.stmtWithText(t, extFma, unpackAs(o.v, o.c, o.e))
		if g1, g2 := fma.guardPath(t, extFma, hi), fma.guardPath(t, extFma, up); g1 != "" || g2 != "" ||
			lo.Pos() < hi.End() || lo.End() > up.Pos() {
			t.Fatalf("bid128_fma.go: the %s load (line %d), its low word (line %d) and the unpacking (line %d) are no "+
				"longer consecutive top-level statements (guarded by %q and %q)",
				o.c, fma.line(hi), fma.line(lo), fma.line(up), g1, g2)
		}
		// The count is the only write to q1/q2 and sits after the unpacking, so it counts the canonicalised coefficient.
		opCount[i] = fma.stmtWithText(t, extFma,
			"if "+o.c+".hi != 0 || "+o.c+".lo != 0 { "+o.q+" = bid128_count_digits("+o.c+") }")
		if g := fma.guardPath(t, extFma, opCount[i]); g != "" || opCount[i].Pos() < up.End() {
			t.Fatalf("bid128_fma.go: the %s digit count at line %d is not a top-level statement after the unpacking (guarded by %q)",
				o.q, fma.line(opCount[i]), g)
		}
		fma.pinWrites(t, extFma, o.q, []string{o.q + " = bid128_count_digits(" + o.c + ")"})
		for _, w := range fma.writes(t, extFma, o.c) {
			if w.pos < hi.Pos() || w.pos >= up.End() {
				t.Fatalf("bid128_ext_fma: %s is written at line %d outside the pinned load and unpacking: %s",
					o.c, w.line, w.text)
			}
		}
		fma.pinNoNestedDecl(t, extFma, o.c)
		fma.pinNoNestedDecl(t, extFma, o.q)
	}
	// --- anchor: the zero-operand block always returns, so both coefficients are non-zero from here on and q1, q2 are
	// in [1, 34] by the digit-count premise above.
	zeroOp := fma.ifWithCond(t, extFma, "(C1.hi == 0 && C1.lo == 0) || (C2.hi == 0 && C2.lo == 0)")
	if _, ok := zeroOp.Body.List[len(zeroOp.Body.List)-1].(*ast.ReturnStmt); !ok {
		t.Fatalf("bid128_ext_fma: the zero-operand block at line %d no longer returns", fma.line(zeroOp))
	}
	if g := fma.guardPath(t, extFma, zeroOp); g != "" ||
		zeroOp.Pos() < opCount[0].End() || zeroOp.Pos() < opCount[1].End() {
		t.Fatalf("bid128_fma.go: the zero-operand block at line %d is not a top-level statement after both digit counts (guarded by %q)",
			fma.line(zeroOp), g)
	}
	// --- anchor: the product chain, its zeroing prologue and the closed world of writes that keeps (C4, q4) the counted
	// pair all the way to the main-body call.
	zi := fma.stmtIndex(t, extFma.Body, "C4.w0 = 0")
	if zi < 3 {
		t.Fatalf("bid128_ext_fma: `C4.w0 = 0` sits at top-level index %d, too early for a four-statement zeroing prologue", zi)
	}
	fma.pinRun(t, "C4 zeroing prologue", extFma.Body, zi-3,
		[]string{"C4.w3 = 0", "C4.w2 = 0", "C4.w1 = 0", "C4.w0 = 0"})
	c4Zero := extFma.Body.List[zi-3]
	c4Chain := fma.asIf(t, "C4 = C1*C2 chain", fma.stmtAt(t, "C4 = C1*C2 chain", extFma.Body, zi+1))
	if g := fma.guardPath(t, extFma, c4Zero); g != "" || c4Zero.Pos() < zeroOp.End() || c4Chain.End() > zeroC3.Pos() {
		t.Fatalf("bid128_fma.go: the C4 prologue (line %d, guarded by %q) and chain (line %d) are not top-level statements "+
			"between the zero-operand block (line %d) and the zero-C3 block (line %d)",
			fma.line(c4Zero), g, fma.line(c4Chain), fma.line(zeroOp), fma.line(zeroC3))
	}
	for _, w := range fma.writes(t, extFma, "C4") {
		if w.pos < c4Zero.Pos() || w.pos >= c4Chain.End() {
			t.Fatalf("bid128_ext_fma: C4 is written at line %d outside the pinned prologue and chain: %s", w.line, w.text)
		}
	}
	for _, w := range fma.writes(t, extFma, "q4") {
		switch {
		case w.pos >= c4Chain.Pos() && w.pos < c4Chain.End(): // the chain's own answer
		case w.pos >= zeroC3.Pos() && w.pos < zeroC3.End(): // inside the block that returns
		default:
			t.Fatalf("bid128_ext_fma: q4 is written at line %d outside the pinned chain and the zero-C3 block: %s",
				w.line, w.text)
		}
	}
	fma.pinNoNestedDecl(t, extFma, "C4")
	fma.pinNoNestedDecl(t, extFma, "q4")

	// The seven arms of the chain: the pinned condition, the pinned body, the numeric reading of the condition, and the
	// side conditions that body needs. `owns` is consulted in order, exactly as the else-if chain evaluates, and an empty
	// cond marks the final else.
	fits := func(what string, v *big.Int, width, s int) {
		if v.Cmp(pow2big(width)) >= 0 {
			t.Fatalf("q1+q2=%d: %s is %s, which does not fit %d bits", s, what, v, width)
		}
	}
	thr := func(got *big.Int, s int) {
		if got.Cmp(pow10big(s-1)) != 0 {
			t.Fatalf("q1+q2=%d: the arm compares C4 against %s, not against 10^%d", s, got, s-1)
		}
	}
	ten64 := func(i, s int) *big.Int {
		if i < 0 || i >= len(bid_ten2k64) {
			t.Fatalf("q1+q2=%d: bid_ten2k64[%d] is outside the table", s, i)
		}
		return new(big.Int).SetUint64(bid_ten2k64[i])
	}
	ten128 := func(i, s int) *big.Int {
		if i < 0 || i >= len(bid_ten2k128) {
			t.Fatalf("q1+q2=%d: bid_ten2k128[%d] is outside the table", s, i)
		}
		return wordsBig([]uint64{bid_ten2k128[i].lo, bid_ten2k128[i].hi})
	}
	ten256 := func(i, s int) *big.Int {
		if i < 0 || i >= len(bid_ten2k256) {
			t.Fatalf("q1+q2=%d: bid_ten2k256[%d] is outside the table", s, i)
		}
		e := bid_ten2k256[i]
		return wordsBig([]uint64{e.w0, e.w1, e.w2, e.w3})
	}
	c4Arms := []struct {
		cond, body string
		owns       func(s int) bool
		check      func(s int)
	}{
		{"q1+q2 <= 19", "{ C4.w0 = C1.lo * C2.lo if C4.w0 < bid_ten2k64[q1+q2-1] { q4 = q1 + q2 - 1 } " +
			"else { q4 = q1 + q2 } }",
			func(s int) bool { return s <= 19 },
			func(s int) {
				// q1, q2 >= 1 so each is at most s-1 digits: both fit 64 bits and C1.lo, C2.lo are the whole values.
				fits("the widest operand", pow10big(s-1), 64, s)
				// The product fits one word, so `C1.lo * C2.lo` cannot wrap and w1..w3 keep their zeroed values.
				fits("the product C1*C2", pow10big(s), 64, s)
				thr(ten64(s-1, s), s)
			}},
		{"q1+q2 == 20", "{ tmp128 := __mul_64x64_to_128(C1.lo, C2.lo) C4.w0 = tmp128.lo C4.w1 = tmp128.hi " +
			"if C4.w1 == 0 && C4.w0 < bid_ten2k64[19] { q4 = 19 } else { q4 = 20 } }",
			func(s int) bool { return s == 20 },
			func(s int) {
				fits("the widest operand", pow10big(19), 64, s)
				fits("the product C1*C2", pow10big(20), 128, s)
				thr(ten64(19, s), s)
				// 10^19 fits one word, so `w1 == 0 && w0 < 10^19` is the full compare against it.
				fits("the threshold", pow10big(19), 64, s)
			}},
		{"q1+q2 <= 38", "{ var tmp128 BID_UINT128 if q1 <= 19 { _, tmp128 = __mul_64x128_full(C1.lo, C2) } " +
			"else { _, tmp128 = __mul_64x128_full(C2.lo, C1) } C4.w0 = tmp128.lo C4.w1 = tmp128.hi " +
			"if C4.w1 < bid_ten2k128[q1+q2-21].hi || (C4.w1 == bid_ten2k128[q1+q2-21].hi && " +
			"C4.w0 < bid_ten2k128[q1+q2-21].lo) { q4 = q1 + q2 - 1 } else { q4 = q1 + q2 } }",
			func(s int) bool { return s <= 38 },
			func(s int) {
				// Whichever side the `q1 <= 19` test picks, the operand it puts in the 64-bit slot has at most 19
				// digits: two operands of 20 or more digits each would put s at 40 or above, outside this arm.
				if 2*20 <= s {
					t.Fatalf("q1+q2=%d: both operands could exceed 19 digits, so the 64x128 slot could truncate one", s)
				}
				fits("the narrow operand", pow10big(19), 64, s)
				// The product fits 128 bits, so the low half __mul_64x128_full returns is the whole product and
				// w2, w3 keep their zeroed values.
				fits("the product C1*C2", pow10big(s), 128, s)
				thr(ten128(s-21, s), s)
			}},
		{"q1+q2 == 39", "{ C4 = __mul_128x128_to_256(C1, C2) if C4.w2 == 0 && (C4.w1 < bid_ten2k128[18].hi || " +
			"(C4.w1 == bid_ten2k128[18].hi && C4.w0 < bid_ten2k128[18].lo)) { q4 = 38 } else { q4 = 39 } }",
			func(s int) bool { return s == 39 },
			func(s int) {
				// The product fits 192 bits, so C4.w3 is zero and the arm's `w2 == 0` plus the 128-bit compare is the
				// full compare against 10^38.
				fits("the product C1*C2", pow10big(39), 192, s)
				thr(ten128(18, s), s)
				fits("the threshold", pow10big(38), 128, s)
			}},
		{"q1+q2 <= 57", "{ C4 = __mul_128x128_to_256(C1, C2) if C4.w2 < bid_ten2k256[q1+q2-40].w2 || " +
			"(C4.w2 == bid_ten2k256[q1+q2-40].w2 && (C4.w1 < bid_ten2k256[q1+q2-40].w1 || " +
			"(C4.w1 == bid_ten2k256[q1+q2-40].w1 && C4.w0 < bid_ten2k256[q1+q2-40].w0))) { q4 = q1 + q2 - 1 } " +
			"else { q4 = q1 + q2 } }",
			func(s int) bool { return s <= 57 },
			func(s int) {
				// Product and threshold both fit 192 bits, so the three-word compare the arm writes is complete.
				fits("the product C1*C2", pow10big(s), 192, s)
				thr(ten256(s-40, s), s)
				fits("the threshold", pow10big(s-1), 192, s)
			}},
		{"q1+q2 == 58", "{ C4 = __mul_128x128_to_256(C1, C2) if C4.w3 == 0 && (C4.w2 < bid_ten2k256[18].w2 || " +
			"(C4.w2 == bid_ten2k256[18].w2 && (C4.w1 < bid_ten2k256[18].w1 || (C4.w1 == bid_ten2k256[18].w1 && " +
			"C4.w0 < bid_ten2k256[18].w0)))) { q4 = 57 } else { q4 = 58 } }",
			func(s int) bool { return s == 58 },
			func(s int) {
				// Here the product can spill into w3, which is why this arm tests it; the threshold still fits 192 bits,
				// so `w3 == 0` plus the three-word compare is the full compare against 10^57.
				fits("the product C1*C2", pow10big(58), 256, s)
				thr(ten256(18, s), s)
				fits("the threshold", pow10big(57), 192, s)
			}},
		{"", "{ C4 = __mul_128x128_to_256(C1, C2) if C4.w3 < bid_ten2k256[q1+q2-40].w3 || " +
			"(C4.w3 == bid_ten2k256[q1+q2-40].w3 && (C4.w2 < bid_ten2k256[q1+q2-40].w2 || " +
			"(C4.w2 == bid_ten2k256[q1+q2-40].w2 && (C4.w1 < bid_ten2k256[q1+q2-40].w1 || " +
			"(C4.w1 == bid_ten2k256[q1+q2-40].w1 && C4.w0 < bid_ten2k256[q1+q2-40].w0))))) { q4 = q1 + q2 - 1 } " +
			"else { q4 = q1 + q2 } }",
			nil,
			func(s int) {
				// The full four-word compare, so only the widths matter.
				fits("the product C1*C2", pow10big(s), 256, s)
				thr(ten256(s-40, s), s)
			}},
	}
	cur := c4Chain
	for i, a := range c4Arms {
		if a.cond == "" {
			fma.pinNode(t, "C4 chain final arm", fma.elseBlock(t, "C4 chain tail", cur), a.body)
			continue
		}
		fma.pinNode(t, "C4 chain arm "+a.cond, cur.Cond, a.cond)
		fma.pinNode(t, "C4 chain arm "+a.cond+" body", cur.Body, a.body)
		if c4Arms[i+1].cond != "" {
			cur = fma.elseIf(t, "C4 chain", cur)
		}
	}
	// The product range, per (q1, q2) pair rather than per s, because that is where it comes from.
	for q1 := 1; q1 <= P34; q1++ {
		for q2 := 1; q2 <= P34; q2++ {
			s := q1 + q2
			min := new(big.Int).Mul(pow10big(q1-1), pow10big(q2-1))
			max := new(big.Int).Mul(new(big.Int).Sub(pow10big(q1), one), new(big.Int).Sub(pow10big(q2), one))
			if min.Cmp(pow10big(s-2)) != 0 || max.Cmp(pow10big(s)) >= 0 {
				t.Fatalf("q1=%d q2=%d: C1*C2 ranges over [%s, %s], not inside [10^%d, 10^%d)", q1, q2, min, max, s-2, s)
			}
		}
	}
	// Every s the chain can see, routed to the arm the pinned conditions select.
	for s := 2; s <= 2*P34; s++ {
		hit := -1
		for i, a := range c4Arms {
			if a.owns == nil || a.owns(s) {
				hit = i
				break
			}
		}
		if hit < 0 {
			t.Fatalf("q1+q2=%d selects no arm of the pinned chain", s)
		}
		c4Arms[hit].check(s)
	}

	// --- anchor: the two routes into the loop. bid_fma_delta_ge_zero has exactly two call sites - bid_fma_main_body under
	// delta >= 0, and the swap arm of bid_fma_delta_lt_zero - and inside it every write to q3, q4, C3, C4 and delta sits
	// after the Cases (2)-(6) call, in a function with no loop and only forward gotos, so the loop is handed the values
	// bid_fma_delta_ge_zero itself received. bid_fma_main_body only forwards, so route 1 carries bid128_ext_fma's own
	// counted pairs and its delta >= 0.
	ltz := body.funcDecl(t, "bid_fma_delta_lt_zero")
	gezDecl := body.funcDecl(t, "bid_fma_delta_ge_zero")
	mbDecl := body.funcDecl(t, "bid_fma_main_body")
	for _, v := range []string{"q3", "q4", "C3", "C4", "delta"} {
		for _, w := range body.writes(t, gezDecl, v) {
			if w.pos < d[0].node.End() {
				t.Fatalf("bid_fma_delta_ge_zero writes %s at line %d, at or before the Cases (2)-(6) call: %s",
					v, w.line, w.text)
			}
		}
		body.pinNoNestedDecl(t, gezDecl, v)
		body.pinWrites(t, mbDecl, v, nil)
		body.pinNoNestedDecl(t, mbDecl, v)
	}
	// --- residual: pinPointerUseContexts, used on four functions below, admits `C3.f` as a field select and rejects
	// `C3.m(...)` by spotting that the selector is the callee of a call. The method *value* `C3.m` escapes both: it is
	// never the callee of anything here, it renders as the same AST shape as a field read, and it binds C3 as a receiver,
	// so `f(C3.m)` or `g := C3.m` would hand *C3 a second name that the by-context census passes as a field select. C3 is
	// a *BID_UINT128, so such a method value can exist only if BID_UINT128 has a method at all - and the package-wide
	// census below fails the moment one is declared in any file this certificate reads, in either receiver form, directly
	// or through an alias or an embedded field. That is what makes the field-select admission below closed rather than
	// fail-open; the direct-call check alone does not.
	pinTypeIsMethodFree(t, files, "BID_UINT128", pinnedUint128Shape,
		"a method value spelled `C3.<that method>` is the AST shape the by-context census admits as a field select, "+
			"and the C3 aliasing closure below stops being closed")

	// --- residual: C3 is a pointer in bid_fma_delta_ge_zero, so a call that hands the bare pointer on can rewrite the
	// coefficient with no statement here naming C3 - invisible to the write census just above, which sees only
	// assignments, inc/dec and address-of. Passing `*C3` or a field copies a value and cannot; passing `C3` can. Close
	// it structurally, starting from the function's four-arm dispatch chain: Case (1')/(1''A), Case (1''B), the
	// Cases (2)-(6) arm the loop is called from, and the final else. The three conditions pinned here are the same ones
	// the dispatch's own guard path already carries, so the chain the exclusion argument uses is the chain the pinned
	// guards describe.
	const gezCase1Cond = "p34 <= delta-1 || (p34 == delta && e3+6176 < p34-q3)"
	if want := "!(" + gezCase1Cond + ") && !(p34 == delta) && " + dispatchInnerCond; dispatchGuards != want {
		t.Fatalf("the pinned dispatch guards are not the negations of the pinned chain arms\n got: %s\nwant: %s",
			dispatchGuards, want)
	}
	gezChain := body.ifWithCond(t, gezDecl, gezCase1Cond)
	gezCase1ppB := body.elseIf(t, "delta_ge_zero Case (1''B) arm", gezChain)
	body.pinNode(t, "delta_ge_zero Case (1''B) guard", gezCase1ppB.Cond, "p34 == delta")
	gezDispatchArm := body.elseIf(t, "delta_ge_zero Cases (2)-(6) arm", gezCase1ppB)
	body.pinNode(t, "delta_ge_zero Cases (2)-(6) guard", gezDispatchArm.Cond, dispatchInnerCond)
	if gezDispatchArm.Pos() != disp.Pos() {
		t.Fatalf("bid_fma_delta_ge_zero: the third chain arm (line %d) is not the dispatch if the guards were read from (line %d)",
			body.line(gezDispatchArm), d[0].pf.line(disp))
	}
	gezTail := body.elseBlock(t, "delta_ge_zero final arm", gezDispatchArm)
	// The closed world of calls handing over the bare pointer: the dispatch itself, plus calls the chain excludes. An
	// excluded call sits in an arm that cannot run in the same execution as the Cases (2)-(6) arm, so it cannot reach
	// *C3 before the loop does; anything else - a new call ahead of the chain, or one inside the dispatch arm itself -
	// fails here. Each excluded call is censused directly as well, so the exclusion is not the only thing holding.
	var c3Passed []*ast.CallExpr
	ast.Inspect(gezDecl, func(n ast.Node) bool {
		ce, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		for _, a := range ce.Args {
			if body.flat(t, a) == "C3" {
				c3Passed = append(c3Passed, ce)
				break
			}
		}
		return true
	})
	// c3AllowedInGez records, per call, the argument indices that carry the bare pointer, so the by-context census below
	// can admit exactly these hand-offs and nothing else.
	c3AllowedInGez := map[token.Pos][]int{}
	excluded := 0
	for _, ce := range c3Passed {
		var args []string
		slot := map[int]string{}
		for i, a := range ce.Args {
			args = append(args, body.flat(t, a))
			if args[i] == "C3" {
				slot[i] = "C3"
				c3AllowedInGez[ce.Pos()] = append(c3AllowedInGez[ce.Pos()], i)
			}
		}
		if ce.Pos() == d[0].node.Pos() { // the dispatch itself; the loop's own C3 world is pinned above
			continue
		}
		if !within(ce, gezChain.Body) && !within(ce, gezCase1ppB.Body) && !within(ce, gezTail) {
			t.Fatalf("bid128_fma_body.go: the call to %s at line %d hands over the bare pointer C3 from outside the arms "+
				"the Cases (2)-(6) arm excludes, so it could rewrite *C3 before the dispatch",
				body.flat(t, ce.Fun), body.line(ce))
		}
		excluded++
		// Census the callee: it must take the pointer in a parameter of the same name, never write through it, never
		// rebind or shadow it, and never pass it on - the last conjunct is what keeps this census closed at one hop.
		name := body.flat(t, ce.Fun)
		var callee *ast.FuncDecl
		var cp *portFile
		for _, p := range files {
			for _, fd := range p.decls() {
				if fd.Name.Name == name {
					callee, cp = fd, p
				}
			}
		}
		if callee == nil {
			t.Fatalf("bid128_fma_body.go: the C3-passing callee %q at line %d has no declaration in the package",
				name, body.line(ce))
		}
		pinSlots(t, "C3-passing call to "+name, callee, args, slot)
		cp.pinWrites(t, callee, "C3", nil)
		cp.pinNoNestedDecl(t, callee, "C3")
		cp.pinNoPointerRebind(t, callee, "C3")
		// Residual: the pass-on scan below reads argument text, so the callee could copy the pointer first
		// (`alias := C3`) and hand `alias` on, or write through the copy, with neither the write census nor that scan
		// naming C3. No context at all is admitted here: the callee may only read fields of *C3 or dereference it.
		cp.pinPointerUseContexts(t, callee, "C3", nil)
		ast.Inspect(callee, func(n ast.Node) bool {
			inner, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			for _, a := range inner.Args {
				if cp.flat(t, a) == "C3" {
					t.Fatalf("%s: %s passes the pointer C3 on at line %d, so this one-hop census stops being closed",
						cp.name, name, cp.line(inner))
				}
			}
			return true
		})
	}
	if len(c3Passed) != 2 || excluded != 1 {
		t.Fatalf("bid_fma_delta_ge_zero hands the bare pointer C3 to %d calls, %d of them in excluded arms; pinned 2 and 1",
			len(c3Passed), excluded)
	}
	// --- anchor: res and C3 stay two variables all the way down. bid128_ext_fma passes the addresses of two distinct
	// locals; every hop below carries each into the parameter of the same name (entrySlots holds both slots); and no hop
	// rebinds either pointer, so `*res` and `*C3` never become the same storage. Writing through them is expected and is
	// left to the per-function write pins.
	if rl := extFma.Type.Results.List; len(rl) == 0 || len(rl[0].Names) != 1 || rl[0].Names[0].Name != "res" ||
		fma.flat(t, rl[0].Type) != "BID_UINT128" {
		t.Fatalf("bid128_ext_fma's first result is no longer the named local `res BID_UINT128`: %s", fma.flat(t, extFma.Type))
	}
	fma.stmtWithText(t, extFma, "var C1, C2, C3 BID_UINT128")
	if pinnedMainBodyArgs[1] != "&res" || pinnedMainBodyArgs[15] != "&C3" {
		t.Fatalf("the pinned main-body call no longer hands over &res and &C3: %v", pinnedMainBodyArgs)
	}
	for _, fd := range []*ast.FuncDecl{mbDecl, gezDecl, ltz, loop} {
		for _, v := range []string{"res", "C3"} {
			body.pinNoPointerRebind(t, fd, v)
			body.pinNoNestedDecl(t, fd, v)
		}
	}
	// --- anchor: source position is execution order in every function whose statements this certificate orders. None of
	// them legitimately loops, so an added loop or a backward goto is a real structural change and fails here rather than
	// silently invalidating an ordering argument. bid_fma_cases_2_to_6 is deliberately absent: it does loop, and its back
	// edge is closed by the narrower continue/break/goto accounting above instead.
	fma.pinForwardOnly(t, extFma)
	for _, name := range []string{"bid_fma_delta_ge_zero", "bid_fma_delta_lt_zero", "bid_fma_case7",
		"bid_fma_cases_11_12", "bid_add_and_round"} {
		body.pinForwardOnly(t, body.funcDecl(t, name))
	}
	if got := strings.Join(mb[0].args, ", "); got != strings.Join(pinnedMainBodyArgs, ", ") {
		t.Fatalf("the bid_fma_main_body call arguments drifted\n got: %s\nwant: %s",
			got, strings.Join(pinnedMainBodyArgs, ", "))
	}
	pinSlots(t, "main-body call", mbDecl, mb[0].args, entrySlots)
	gez := portCallsTo(t, files, "bid_fma_delta_ge_zero")
	if len(gez) != 2 {
		t.Fatalf("bid_fma_delta_ge_zero has %d call sites, pinned 2 (delta >= 0 and the swap arm): %v", len(gez), gez)
	}
	route1, swapCall := -1, -1
	for i, s := range gez {
		switch s.fd.Name.Name {
		case mbDecl.Name.Name:
			route1 = i
		case ltz.Name.Name:
			swapCall = i
		}
	}
	if route1 < 0 || swapCall < 0 {
		t.Fatalf("the two bid_fma_delta_ge_zero routes are no longer main-body and swap-arm: %v", gez)
	}
	if gez[route1].guards != "delta >= 0" {
		t.Fatalf("the delta >= 0 route drifted: %s guarded by %q", gez[route1].key, gez[route1].guards)
	}
	pinSlots(t, "delta >= 0 route", gezDecl, gez[route1].args, entrySlots)

	// --- premise: route 2 reaches no wide rounding helper inside the loop, so the k bound below only has to hold on
	// route 1. bid_fma_delta_lt_zero is entered only from bid_fma_main_body with delta < 0; the swap arm is the only place
	// it writes q3 or q4; and the pinned arm body runs `ind = q3; q3 = q4; q4 = ind`, so the callee's q4 is this
	// function's incoming q3 - bid128_ext_fma's digit count of C3, at most p34 by the premise above. Both loop calls carry
	// the pinned guard `!(q4 <= 38)`, which is then false. The arm's selection guard is irrelevant to that conclusion and
	// is deliberately not pinned: whatever picks the arm, its body hands over q4 = the incoming q3.
	ltzCalls := portCallsTo(t, files, "bid_fma_delta_lt_zero")
	if len(ltzCalls) != 1 {
		t.Fatalf("bid_fma_delta_lt_zero has %d call sites, pinned 1: %v", len(ltzCalls), ltzCalls)
	}
	if ltzCalls[0].fd.Name.Name != mbDecl.Name.Name || ltzCalls[0].guards != "!(delta >= 0)" {
		t.Fatalf("the delta < 0 entry drifted: %s guarded by %q", ltzCalls[0].key, ltzCalls[0].guards)
	}
	pinSlots(t, "delta < 0 entry", ltz, ltzCalls[0].args, entrySlots)

	// --- residual: every anchor above that keeps *C3 the coefficient bid128_ext_fma counted is keyed on a statement or
	// an argument that spells C3 - the write census, the pointer-rebind pin, and the bare-argument census of
	// bid_fma_delta_ge_zero. A copy defeats all three at once: `aliasC3 := C3` assigns to aliasC3, so writes() records
	// nothing; C3 itself is never rebound; and a later `aliasC3.hi = ...` or `f(aliasC3)` rewrites *C3 with no statement
	// naming C3. Close it by classifying every C3 identifier in the three functions that hold the pointer while the
	// certificate's premise has to survive, admitting only field selects, dereferences, and the hand-offs already pinned
	// argument by argument: the two main-body routes at the C3 slot entrySlots fixes on both sides, and the two
	// bare-pointer calls of bid_fma_delta_ge_zero the census above censused. bid_fma_cases_2_to_6 hands the pointer to
	// nobody, so no call argument is admitted there at all.
	c3Slot := -1
	for i, n := range entrySlots {
		if n == "C3" {
			c3Slot = i
		}
	}
	if c3Slot < 0 {
		t.Fatal("entrySlots no longer pins a C3 slot, so the main-body hand-offs cannot be tied to one")
	}
	body.pinPointerUseContexts(t, mbDecl, "C3", map[token.Pos][]int{
		gez[route1].node.Pos(): {c3Slot},
		ltzCalls[0].node.Pos(): {c3Slot},
	})
	body.pinPointerUseContexts(t, gezDecl, "C3", c3AllowedInGez)
	body.pinPointerUseContexts(t, loop, "C3", nil)

	// --- anchor: the swap arm's whole body - the (C3,C4) and (q3,q4) exchange and the call it ends in - and the closed
	// world of writes that keeps the exchanged values the incoming ones.
	negIdx := body.stmtIndex(t, ltz.Body, "delta = -delta")
	chain := body.asIf(t, "delta_lt_zero dispatch", body.stmtAt(t, "delta_lt_zero dispatch", ltz.Body, negIdx+1))
	body.pinNode(t, "delta_lt_zero Case (7) head", chain.Cond, "p34 < q4 && q4 <= delta")
	swapArm := body.elseIf(t, "delta_lt_zero swap arm", chain)
	body.pinNode(t, "delta_lt_zero swap arm body", swapArm.Body, swapArmBody)
	for _, w := range []struct {
		v    string
		want []string
	}{
		{"delta", []string{"delta = -delta", "delta = q3 + e3 - q4 - e4"}},
		{"q3", []string{"q3 = q4"}}, {"q4", []string{"q4 = ind"}},
		{"C3", []string{"C3.hi = C4.w1", "C3.lo = C4.w0"}},
		{"C4", []string{"C4.w1 = P128.hi", "C4.w0 = P128.lo"}},
	} {
		for i, got := range body.pinWrites(t, ltz, w.v, w.want) {
			ok := got.pos >= swapArm.Body.Pos() && got.pos < swapArm.Body.End()
			if w.v == "delta" && i == 0 {
				ok = got.pos < chain.Pos() // the negation, the only write ahead of the guard
			}
			if !ok {
				t.Fatalf("bid128_fma_body.go: %s is written at line %d, outside the pinned swap arm: %s", w.v, got.line, got.text)
			}
		}
		body.pinNoNestedDecl(t, ltz, w.v)
	}
	if !within(gez[swapCall].node, swapArm.Body) {
		t.Fatalf("the swap-arm call at line %d is outside the pinned swap arm at line %d",
			body.line(gez[swapCall].node), body.line(swapArm))
	}
	pinSlots(t, "swap-arm call", gezDecl, gez[swapCall].args, entrySlots)
	if P34 > 38 {
		t.Fatalf("p34 = %d exceeds 38, so the swapped q4 could clear the `!(q4 <= 38)` guard of a wide helper", P34)
	}

	// --- anchor: the power-of-ten tables and the two multiply helpers the scaling and the incr_exp normalisation read.
	// bits.Mul64 is the full 64x64 product by the Go standard library's contract, and __mul_64x128_short adds the low 64
	// bits of a*b.hi into the high word of the full a*b.lo product, i.e. it returns a*b modulo 2^128. So each of the
	// scaling block's three forms is the exact product whenever that product fits 128 bits.
	for i := 0; i < len(bid_ten2k64); i++ {
		if new(big.Int).SetUint64(bid_ten2k64[i]).Cmp(pow10big(i)) != 0 {
			t.Fatalf("bid_ten2k64[%d] = %d, want 10^%d", i, bid_ten2k64[i], i)
		}
	}
	for i := 0; i < len(bid_ten2k128); i++ {
		if got := wordsBig([]uint64{bid_ten2k128[i].lo, bid_ten2k128[i].hi}); got.Cmp(pow10big(i+20)) != 0 {
			t.Fatalf("bid_ten2k128[%d] = %s, want 10^%d", i, got, i+20)
		}
	}
	// The last four pin the exact 192- and 256-bit products the C4 chain uses: __mul_64x128_full splits b into limbs,
	// multiplies each by a with the exact 64x64 helper and adds the overlap with the pinned carry-propagating 128+64 add,
	// so (ph, ql) is the true 192-bit product; __mul_128x128_to_256 composes two of those with the pinned carry-out adds
	// and so is the true 256-bit product, its top word never overflowing because a 128x128 product fits 256 bits.
	for _, m := range []struct{ file, fn, want string }{
		{"internal.go", "__mul_64x64_to_128", "{ hi, lo := bits.Mul64(cx, cy) return BID_UINT128{lo: lo, hi: hi} }"},
		{"to_uint64_support.go", "__mul_128x64_to_128", "{ return __mul_64x128_short(a, b) }"},
		{"inline_round64.go", "__mul_64x128_short", "{ var ql BID_UINT128 var ALBH_L uint64 " +
			"ALBH_L = __mul_64x64_to_64(a, b.hi) ql = __mul_64x64_to_128(a, b.lo) ql.hi += ALBH_L return ql }"},
		{"inline_round64.go", "__mul_64x64_to_64", "{ return cx * cy }"},
		{"internal.go", "__mul_64x128_full", "{ albh := __mul_64x64_to_128(a, b.hi) albl := __mul_64x64_to_128(a, b.lo) " +
			"ql.lo = albl.lo qm2 := __add_128_64(albh, albl.hi) ql.hi = qm2.lo ph = qm2.hi return }"},
		{"internal.go", "__add_128_64", "{ var r BID_UINT128 r64h := a.hi r.lo = b + a.lo if r.lo < b { r64h++ } " +
			"r.hi = r64h return r }"},
		{"internal.go", "__mul_128x128_to_256", "{ var p256 BID_UINT256 var cy1, cy2 uint64 " +
			"phl, qll := __mul_64x128_full(a.lo, b) phh, qlh := __mul_64x128_full(a.hi, b) p256.w0 = qll.lo " +
			"p256.w1, cy1 = __add_carry_out(qlh.lo, qll.hi) p256.w2, cy2 = __add_carry_in_out(qlh.hi, phl, cy1) " +
			"p256.w3 = phh + cy2 return p256 }"},
		{"internal.go", "__add_carry_out", "{ s = x + y if s < x { cy = 1 } return }"},
		{"internal.go", "__add_carry_in_out", "{ x1 := x + ci s = x1 + y if (s < x1) || (x1 < ci) { cy = 1 } return }"},
	} {
		mp := loadPortFile(t, m.file)
		mp.pinNode(t, m.fn, mp.funcDecl(t, m.fn).Body, m.want)
	}

	// --- the k <= 1 conclusion. A second continue would have to fire at iteration 1, where the loop recomputes
	// res = C3 * 10^(p34-q3+1): that product is at least 10^34 and stays inside 128 bits, so it is exact by the helper
	// contracts above. The subtrahend is C4 rounded to q4-x0 = p34+1-delta digits, at most 10^(p34+1-delta) by the
	// helper bound below and by the incr_exp arms' own table value, and 2 <= delta <= p34-1 on this route; both the
	// R192 and R256 forms then fit the two low words the loop copies into R128. res - R128 therefore stays above the
	// back edge's 10^33 threshold, and its `res == 10^33` midpoint alternative is out of reach too. If x0 has instead
	// fallen to 0 the back edge's own `x0 >= 1` conjunct fails, so either way there is no second continue.
	tenPow33 := wordsBig([]uint64{0x38c15b0a00000000, 0x0000314dc6448d93})
	if tenPow33.Cmp(pow10big(33)) != 0 {
		t.Fatalf("the back edge's pinned constant is %s, not 10^33", tenPow33)
	}
	if pow10big(19).Cmp(pow2big(64)) >= 0 {
		t.Fatal("10^19 no longer fits 64 bits, so the q3 <= 19 arms could read a truncated C3.lo")
	}
	for q3 := 1; q3 <= P34; q3++ {
		scale := P34 - q3 + maxContinues
		// The arm the pinned block selects must index inside the pinned tables: 10^scale from bid_ten2k64 when
		// scale <= 19 or q3 > 19, from bid_ten2k128 otherwise.
		if scale < 1 || !((q3 <= 19 && (scale <= 19 || scale-20 < len(bid_ten2k128))) ||
			(q3 > 19 && scale < len(bid_ten2k64))) {
			t.Fatalf("q3=%d: iteration %d scales by 10^%d, outside the arm's pinned table", q3, maxContinues, scale)
		}
		if got := new(big.Int).Mul(pow10big(q3-1), pow10big(scale)); got.Cmp(pow10big(33+maxContinues)) != 0 {
			t.Fatalf("q3=%d: the smallest scaled C3 at iteration %d is %s, want 10^%d", q3, maxContinues, got, 33+maxContinues)
		}
		if got := new(big.Int).Mul(pow10big(q3), pow10big(scale)); got.Cmp(pow2big(128)) >= 0 {
			t.Fatalf("q3=%d: the scaled C3 at iteration %d can reach %s, which does not fit 128 bits", q3, maxContinues, got)
		}
	}
	for delta := minDeltaOpposite; delta < P34; delta++ {
		sub := pow10big(P34 + maxContinues - delta)
		if sub.Cmp(pow2big(128)) >= 0 {
			t.Fatalf("delta=%d: the subtrahend bound %s does not fit the two words copied into R128", delta, sub)
		}
		if res := new(big.Int).Sub(pow10big(33+maxContinues), sub); res.Cmp(tenPow33) <= 0 {
			t.Fatalf("delta=%d: after one continue res can fall to %s, at or below 10^33; a second continue is possible",
				delta, res)
		}
	}

	checkRoundHelperBound(t) // second sampled premise: Cstar <= 10^(q-x)
}

// swapArmBody is the pinned delta < 0 swap arm of bid_fma_delta_lt_zero: the exchange of (C3,C4), (q3,q4), (e3,e4),
// (z_sign,p_sign) and (z_exp,p_exp) that hands bid_fma_delta_ge_zero the other operand pair. What the certificate reads
// out of it is the `ind = q3; q3 = q4; q4 = ind` triple: the callee's q4 is this function's incoming q3.
const swapArmBody = "{ P128.hi = C3.hi P128.lo = C3.lo C3.hi = C4.w1 C3.lo = C4.w0 C4.w1 = P128.hi C4.w0 = P128.lo " +
	"ind = q3 q3 = q4 q4 = ind ind = e3 e3 = e4 e4 = ind tmp_sign = z_sign z_sign = p_sign p_sign = tmp_sign " +
	"tmp64 := z_exp z_exp = p_exp p_exp = tmp64 delta = q3 + e3 - q4 - e4 " +
	"bid_fma_delta_ge_zero(p34, res, &is_midpoint_lt_even, &is_midpoint_gt_even, &is_inexact_lt_midpoint, " +
	"&is_inexact_gt_midpoint, p_sign, z_sign, &z_exp, &p_exp, q3, q4, &e3, &e4, delta, C3, C4, rnd_mode, pfpsf) }"

// pinnedCountDigits is the whole body of bid128_count_digits. The certificate reads two things out of it: the
// bid_nr_digits row is selected by the value's bit length, and within a row the answer is either the row's constant or
// its digits1 raised by one exactly at the row's threshold.
const pinnedCountDigits = "{ var x_nr_bits int if C.hi == 0 { if C.lo == 0 { return 0 } " +
	"if C.lo >= 0x0020000000000000 { tmp := math.Float64bits(float64(C.lo >> 32)) " +
	"x_nr_bits = 33 + int(((tmp>>52)&0x7ff)-0x3ff) } else { tmp := math.Float64bits(float64(C.lo)) " +
	"x_nr_bits = 1 + int(((tmp>>52)&0x7ff)-0x3ff) } } else { tmp := math.Float64bits(float64(C.hi)) " +
	"x_nr_bits = 65 + int(((tmp>>52)&0x7ff)-0x3ff) } q := int(bid_nr_digits[x_nr_bits-1].digits) if q == 0 { " +
	"q = int(bid_nr_digits[x_nr_bits-1].digits1) if C.hi > bid_nr_digits[x_nr_bits-1].threshold_hi || " +
	"(C.hi == bid_nr_digits[x_nr_bits-1].threshold_hi && C.lo >= bid_nr_digits[x_nr_bits-1].threshold_lo) { q++ } } " +
	"return q }"

func wordsBig(w []uint64) *big.Int {
	out := new(big.Int)
	for i := len(w) - 1; i >= 0; i-- {
		out.Lsh(out, 64).Add(out, new(big.Int).SetUint64(w[i]))
	}
	return out
}

func bigWords(v *big.Int, n int) []uint64 {
	out, tmp := make([]uint64, n), new(big.Int).Set(v)
	mask := new(big.Int).SetUint64(^uint64(0))
	for i := range out {
		out[i] = new(big.Int).And(tmp, mask).Uint64()
		tmp.Rsh(tmp, 64)
	}
	if tmp.Sign() != 0 {
		panic("value wider than the requested word count")
	}
	return out
}

// checkRoundHelperBound checks the sampled postcondition the case2_repeat argument rests on: every production rounding
// helper the loop can call returns Cstar <= 10^(q-x), and sets incr_exp only together with Cstar = 10^(q-x-1), which the
// callers' pinned arms replace by exactly 10^(q-x). It runs the production helpers over deterministic boundaries of
// each digit class - the extremes, a rounding-tie probe (lo + 5*10^(x-1), whose discarded tail is exactly half an ulp of
// the retained digits; it is a tie probe, not the digit class's midpoint) and the two coefficients straddling the
// rounding-overflow edge - plus a fixed-seed corpus, across each helper's documented domain (2 <= q <= 76,
// 1 <= x <= q-1). This is a sample, not
// an enumeration; the file header names it as the second sampled premise.
func checkRoundHelperBound(t *testing.T) {
	t.Helper()
	rng := rand.New(rand.NewSource(3174304613))
	run := func(q, x int, c *big.Int) (*big.Int, int) {
		switch w := bigWords(c, 4); {
		case q <= 18:
			var incr, a, b, cc, d int
			return new(big.Int).SetUint64(bid_round64_2_18(q, x, w[0], &incr, &a, &b, &cc, &d)), incr
		case q <= 38:
			cs, incr, _, _, _, _ := bid_round128_19_38(q, x, BID_UINT128{lo: w[0], hi: w[1]})
			return wordsBig([]uint64{cs.lo, cs.hi}), incr
		case q <= 57:
			cs, incr, _, _, _, _ := bid_round192_39_57(q, x, BID_UINT192{w0: w[0], w1: w[1], w2: w[2]})
			return wordsBig([]uint64{cs.w0, cs.w1, cs.w2}), incr
		default:
			cs, incr, _, _, _, _ := bid_round256_58_76(q, x, BID_UINT256{w0: w[0], w1: w[1], w2: w[2], w3: w[3]})
			return wordsBig([]uint64{cs.w0, cs.w1, cs.w2, cs.w3}), incr
		}
	}
	one := big.NewInt(1)
	for q := 2; q <= 76; q++ {
		lo, hi := pow10big(q-1), new(big.Int).Sub(pow10big(q), one)
		span := new(big.Int).Sub(hi, lo)
		for x := 1; x < q; x++ {
			half := new(big.Int).Mul(big.NewInt(5), pow10big(x-1))
			edge := new(big.Int).Sub(pow10big(q), half) // rounds up to 10^(q-x)
			probes := []*big.Int{lo, new(big.Int).Add(lo, one), hi, new(big.Int).Add(lo, half),
				edge, new(big.Int).Sub(edge, one)}
			for i := 0; i < 4; i++ {
				probes = append(probes, new(big.Int).Add(lo, new(big.Int).Rand(rng, span)))
			}
			bound := pow10big(q - x)
			for _, c := range probes {
				if c.Cmp(lo) < 0 || c.Cmp(hi) > 0 {
					continue
				}
				cstar, incr := run(q, x, c)
				if cstar.Cmp(bound) > 0 {
					t.Fatalf("q=%d x=%d C=%s: Cstar = %s exceeds 10^(q-x) = %s", q, x, c, cstar, bound)
				}
				if incr != 0 && cstar.Cmp(pow10big(q-x-1)) != 0 {
					t.Fatalf("q=%d x=%d C=%s: incr_exp set but Cstar = %s, want 10^%d", q, x, c, cstar, q-x-1)
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------
// certificate 6: bid128_sqrt.go directed-rounding undershoot arms
//                (audit site 16 = :423, audit site 19 = :661)
// ---------------------------------------------------------------------------

// The certified regions are the round-up arms the directed-rounding path of each
// sqrt entry point takes when the estimator has *undershot* isqrt(C256) - the
// bodies of `if (CS+1)^2 <= C256 { CS.lo++; if CS.lo == 0 { CS.hi++ } }` at
// Bid128Sqrt :415-427 (site16 is the CS.lo++ at :423) and at Bid128dSqrt :652-663
// (site19 is the CS.hi++ at :661). Both mutants are aor:inc->dec.
//
// Sites 17 (bid128_sqrt.go:571) and 18 (bid128_sqrt.go:646) are NOT certified
// here and nothing below claims them: both additionally require the estimator to
// return CS with CS.lo == 2^64-1, which Theorem S does not exclude. They are dead
// on separate evidence - an exhaustive sweep for a reachable A with
// isqrt(A) = -1 mod 2^64, 0 hits over all 487,890,977,618,477 candidate j - which
// is recorded in the closure note above pinnedSqrtCSWrites but not re-derived by
// anything here. Both lines are pinned exactly by the write worlds below, so both
// aor:inc->dec mutants fail this file; those kills are structural, and the
// reachability verdict on the two sites is the sweep, not this certificate.
//
// # 1. Firing predicates, read off the pinned source
//
// A := C256 as a mathematical integer; M := isqrt(A); CS := the raw 128-bit return
// of bid_long_sqrt128(C256) (:311 / :554), which both sites read before any
// correction.
//
//	:313  (rnd_mode & 3) == 0 false                      -> the directed path
//	:357  M256 = CS^2                                     (__sqr128_to_256, exact)
//	:360  M256 > C256 false                               -> CS^2 <= A
//	:401-414 M256 <- CS^2+2*CS+1 = (CS+1)^2               (three summands: CS^2 is
//	         the M256 of :357; C8 = 2*CS is built at :358-359 and added in at
//	         :401-404; the +1 is the carry cascade at :405-414 / :642-651)
//	:415-421 M256 <= C256, final comparator "<=" at :421  -> (CS+1)^2 <= A
//
// so **site16 fires <=> (rnd_mode & 3) != 0 and (CS+1)^2 <= A**. The :360
// condition is subsumed, and the predicate is the pure undershoot statement
// **CS <= M-1**. site19 is the carry of the same arm in Bid128dSqrt (:556 mode
// split, :595 CS^2, :598 not greater, :638-651 the (CS+1)^2 build, :652-658 the
// guard with its "<=" at :658), reached when :659 CS.lo++ additionally wraps:
//
//	**site19 fires <=> (rnd_mode & 3) != 0 and (CS+1)^2 <= A and CS.lo == 2^64-1**,
//
// a subset of site16's set. The mode split only places the sites on the directed
// path; everything after it is mode-independent, which is why the non-standard
// BID_ROUNDING_NEAREST_DOWN = 5 (also directed under & 3) needs no separate case.
//
// # 2. Reachable band: A = C256 in [10^66, 10^68)
//
// Both callers build C256 the same way (Bid128Sqrt :257-303, Bid128dSqrt
// :499-546), every step pinned below:
//
//	be' := binary exponent of the binary32 estimate fl32(CX)   (noFmaMulAddF32)
//	digits := E[be'] + (CX >= P[be'] ? 1 : 0)                  (E, P are port tables)
//	scale := 67 - digits ; exponent_q := exponent_x - scale ; scale += exponent_q & 1
//	C256 := CX * 10^scale
//
// Lemma D1. be' in {be, be+1} with be = floor(log2 CX): f64_d = 0x5f800000 = 2^64
// exactly in binary32 and noFmaMulAddF32(a,b,c) = fl32(fl32(a*b)+c), so monotone
// rounding gives be' >= be, and fx_d <= CX*(1+2^-24)^2 < 2^(be+2) gives be' <= be+1.
// CX.hi < 2^49 (SMALL_COEFF_MASK128) keeps fl32(CX.hi)*2^64 < 2^113, so binary32
// cannot overflow.
//
// Lemma D2. digits is the *exact* decimal digit count D(CX). For be' = be that is
// E[be] = D(2^be) plus the pinned decade correction; for be' = be+1 the correction
// is 0 and the code yields E[be+1], which equals D(CX) as soon as
// 2^j / 10^(D(2^j)-1) >= (1+2^-24)^2 for j = be+1 <= 114. That minimum is
// 1.01412048... at j = 103 against a required 1.000000119..., a 1.4% margin. The
// table facts (E[i] = D(2^i) for every entry, P[i] = 10^E[i] for i <= 113) and the
// minimum are recomputed mechanically below.
//
// Index range. Bid128dSqrt has CX < 10^16 < 2^54 (unpack_BID64 maps a non-canonical
// coefficient to 0, returning valid=false; the small path masks with
// SMALL_COEFF_MASK64 < 10^16), so be <= 53; Bid128Sqrt has CX < 10^34
// (unpack_BID128_value zeroes any coefficient >= 10^34, returning valid=false), so
// be <= 112. With be' <= be+1 the used indices are <= 113, which is exactly the
// range the P[i] = 10^E[i] check below covers. The identity is not a property of the
// whole table - 118, 119, 121 and 122 fail it - so the bound on CX that keeps the
// indices at or below 113 is load-bearing, and the closure of CX pinned below is
// what keeps that bound attached to the value the band run reads.
//
// Scale. digits in [1,34] gives scale in [33,67]; digits in [1,16] gives scale in
// [51,67]. Both branches multiply by a bid_power10_table_128 entry, and every
// index either branch can reach is value-checked below (P5). The scale > 38
// branch goes through the *truncating* __mul_128x128_low - whose body is pinned
// under P10, since "truncating" is a statement about which product bits that
// primitive keeps - and does not truncate: scale-37 = 30-D+p with p in {0,1}, so
// CX*10^(scale-37) < 10^(30+p) <= 10^31 < 2^128. With CX in [10^(D-1), 10^D),
//
//	**A = C256 = CX*10^scale in [10^(66+p), 10^(67+p)) subset [10^66, 10^68)**
//
// for both entry points. The set over-approximates what is truly reachable (inputs
// whose short-path A10 is a perfect square return at :277 / :519 and never reach the
// estimator); a deadness proof over a superset is still a deadness proof.
// A < 10^68 < 2^226 also zeroes the top two bits of C256.w3, so C4 = 4*A at
// :306-309 / :549-552 is exact - not needed by sites 16/19, but it is what
// separates this arm from its nearest-mode sibling.
//
// # 3. Theorem S: isqrt(A) <= CS <= isqrt(A)+1 on the whole band
//
// Exact symbolic model of the pinned body (:106-203); u := 2^-53, fl(x) := the
// correctly rounded binary64 value of x.
//
// Step 1, lx (:114-121). float64(w3) is exact (w3 < 2^34) and the scalings by
// 2^64 / 2^128 are exact, so the accumulator starts at v3 = w3*2^192 exactly;
// p2 := fl(w2)*2^128 carries |p2 - w2*2^128| <= 2^138 (|fl(w2) - w2| <= 2^10 is the
// half-ulp of a 64-bit integer rounded to 53 bits). Since A >= 10^66 > 2^219,
// ulp(fl(v3+p2)) >= 2^167, so p1 = fl(w1)*2^64 and p0 = fl(w0) together are
// < 2^128 + 2^64 < 1/2 ulp and get absorbed: lx = fl(v3+p2). With d := (lx-A)/A and
// 2^kappa <= v3+p2 < 2^(kappa+1), the final rounding contributes 2^(kappa-53) and
// the absorbed exact terms contribute w1*2^64 + w0 < 2^128 + 2^64 < 2^129, so
// |lx - A| <= 2^(kappa-53) + 2^138 + 2^129. With A >= 10^66 > 2^219.2 the additive
// part is (2^138+2^129)/A < 2^-81:
//
//	(E1) |d| <= 2^(kappa-53)/A + 2^-81.
//
// Step 2, the seed (:123-127). RS := ly_d = fl(1/fl(sqrt(lx))) = MY*2^(-ey-52)
// exactly, MY in [2^52, 2^53). With 1+e := RS^2*A = (1+d2)^2/((1+d1)^2 (1+d)),
//
//	(E2) |e| <= 4u + |d| + 13u^2.
//
// Step 3, eps (:130-138). ARS0 = MY*A (320-bit, exact), ARS = MY*ARS0 = MY^2*A
// (384-bit, exact) = (1+e)*2^(2ey+104). The word expressions extract
// W = floor(ARS/2^(2ey-24)) mod 2^128 and arithmetic-shift ES.hi right by one, so
// the signed 128-bit value is g = floor(e*2^127) =: e*2^127 - phi, phi in [0,1).
//
// Step 4 (:141-172). a1 := ARS1 = floor(P/2^192), a0 := ARS00 = floor(P/2^64) with
// P = MY*A; both sign branches compute the same integer S1 = a0 - g*a1.
// Step 5 (:175-184). h := |g| >> 64 = floor(H) (or floor(H)+1 in the e<0 edge case)
// with H := |e|*2^63; ES32 = h + h>>1, ES2 = ES32*h = 1.5h^2 - psi*h/2 with
// psi = h mod 2; S = S1 + ES2*a1. The edge case is not just named but bounded, and
// T3's positive side below is the only thing that reads the bound. For e < 0 step 3's
// g is floor(e*2^127) = -ceil(X) with X := |e|*2^127, and the negation at :175-178
// makes h = floor(-g / 2^64) = floor(ceil(X)/2^64). That exceeds floor(X/2^64) =
// floor(H) only when some multiple m*2^64 lies in (X, ceil(X)], an interval of length
// < 1, which forces m - H = (m*2^64 - X)/2^64 < 2^-64; and then h is that m, since
// (m+1)*2^64 > X + 2^64 > ceil(X). So **h > floor(H) only when h = ceil(H) and
// |H - h| < 2^-64**, and zeta := H - h lies in [0,1) on the ordinary path and in
// (-2^-64, 0] on the edge one - never in (-1, 0).
//
// Step 6 (:187-200). Q := floor(S/2^(ey-13)), CS = floor((Q+1)/2) - the pinned
// body spells the divisor as a word re-index plus a shift by k = ey-77, and the
// word step contributes the remaining 64. With
// sigma := S/2^(ey-12), integer algebra gives
//
//	(E3) CS >= M   <=  sigma >= M - 1/2      (E4) CS <= M+1 <=  sigma < M + 3/2
//
// Decomposition, with a0 = P/2^64 - alpha and a1 = P/2^192 - beta in [0,1):
//
//	sigma - sqrt(A) = T1 + T2 + T3 + T4
//	T1 = sqrt(A)*( sqrt(1+e)(1 - e/2 + 3e^2/8) - 1 )
//	T2 = A*RS*phi / 2^128
//	T3 = A*RS*( ES2 - 1.5H^2 ) / 2^128
//	T4 = -( alpha - g*beta + ES2*beta ) / 2^(ey-12)
//
// Bounds, all recomputed as exact rationals in the error-bound test below:
//
//	T3 (dominant): ES2 - 1.5H^2 = -1.5*zeta*(2H-zeta) - psi*h/2 with zeta = H-h in
//	  [0,1), so -3.5H <= ES2 - 1.5H^2 <= 3H*2^-64. The positive side is where step 5's
//	  edge-case bound is spent: the expression can only be positive when zeta < 0, i.e.
//	  when h = ceil(H), and step 5 shows that happens only with |zeta| < 2^-64, giving
//	  1.5*|zeta|*(2H+|zeta|) < 3H*2^-64 + 1.5*2^-128. The residue is dropped from the
//	  stated constant rather than carried, which the coded check pays for many times
//	  over: it evaluates 3H*2^-64 at hMax = |e|max*2^63 + 1, an over-estimate of H by a
//	  whole unit against a residue of 0.5*2^-64 units. Without the 2^-64 bound zeta
//	  would only be known to be in (-1,1) and the positive side would be ~3H, twenty
//	  orders of magnitude past the constant. With A*RS <= sqrt(A)(1+2^-51) and
//	  (E1)+(E2), |T3| <= 3.5*2^-118*(1+2^-51)*F where F = |e|*2^53*sqrt(A) <=
//	  4 sqrt(A) + (1+2^-80)*2^kappa/sqrt(A) + 2^-28 sqrt(A) + 13*2^-53 sqrt(A); the
//	  second term is u*(2^kappa/A)*2^53*sqrt(A) = 2^kappa/sqrt(A) exactly, and the
//	  (1+2^-80) written in front of it is pure extra conservatism, not a slack the
//	  step needs; the third is (E1)'s additive 2^-81 times 2^53.
//	  **F's maximum is joint in (A,kappa), not a sum of per-term maxima**: kappa is
//	  the binade of v3+p2, and 2^kappa/sqrt(A) taken alone peaks at 2^112.5 =
//	  7.343e33 at a binade's left edge, 36% above the value the maximising pair
//	  produces - adding per-term maxima gives 4.734e34 and would carry |T3| to
//	  0.4986, past the bound below. Inside one binade F increases in A (its
//	  derivative is positive once (1+2^-80)*2^kappa <= 4*A, which the band gives),
//	  so F_kappa is F at that binade's right end A_r = min(2^(kappa+1), 10^68)
//	  widened by the same 2^-80 slack that relates A to v3+p2, and the maximum over
//	  the admissible kappa in {219..225} is F_225 = 4.5392e34, attained where the
//	  *band* cap 10^68 - not 2^226 - ends binade 225. Hence |T3| <= 0.478088 and,
//	  on the positive side, T3 <= 2.5e-20. The binade set, the monotonicity
//	  inequality, the per-binade maxima and the kappa-slack corner where v3+p2 sits
//	  one binade above A are all recomputed as exact rationals below; the square
//	  roots are bracketed by integer square roots with an explicit 2^-128 slack, so
//	  no irrational value enters the chain.
//	T2: 0 <= T2 <= sqrt(A)(1+2^-51)/2^128 <= 2.939e-5.
//	T4: |alpha - g*beta + ES2*beta| <= 1 + |g| + ES2 < 2^76.4, 2^(ey-12) >= 2^98,
//	  so |T4| <= 3.1e-7.
//	T1: sqrt(1+e)(1-e/2+3e^2/8) - 1 = sqrt(1+e)(5e^3/16 - 35e^4/128 + ...), so
//	  |T1| <= 0.3126*|e|^3*sqrt(A) <= 5.5e-13. The constant 0.3126 is the plan's
//	  round-up of 5/16 for the alternating tail; the certificate checks the
//	  arithmetic that uses it, not the series estimate itself.
//
// Range of ey. RS ~ A^(-1/2) and correctly rounded sqrt/divide put ly_d within a
// factor (1 +- 3u) of 1/sqrt(A), pinning ey to {110,111,112,113}; every shift count
// in the pinned body is therefore inside (0,64) - k = 2ey-216 in {4,6,8,10} at
// :136-137, k = ey-77 in {33..36} at :189-190. Both endpoints are checked below.
//
//	**Theorem S.** For every A in [10^66,10^68) - in particular every A the two
//	callers can build - sigma - sqrt(A) lies in [-0.478089, +2.970e-5], so
//	sigma >= M - 0.478089 > M - 1/2 and sigma <= (M+1) + 2.97e-5 < M + 3/2; by
//	(E3),(E4)  **M <= CS <= M+1**.
//
// The load-bearing lower bound has margin 0.5 - 0.478089 = 0.021911, i.e. 4.4%.
// **Consequence.** Both arms need (CS+1)^2 <= A, i.e. CS <= M-1, contradicting
// CS >= M: they are unreachable, so no input can distinguish CS.lo++ from CS.lo--
// at :423 or CS.hi++ from CS.hi-- at :661.
//
// # 4. Premises
//
//	P1 STANDARD - Go float64 +, -, *, / are IEEE-754 binary64 round-to-nearest-even
//	   and uint64 -> float64 rounds to nearest even (Go spec, "Floating-point
//	   operators", "Conversions").
//	P2 STANDARD - math.Sqrt is correctly rounded (IEEE-754 5.4.1).
//	P3 STANDARD - Go never evaluates float64 expressions in a wider format, and
//	   math.Float64frombits(math.Float64bits(x)) is the identity that forces an
//	   explicit rounding and so suppresses FMA contraction. noFmaMulAddF64 relies on
//	   exactly that; its body is pinned below. If a toolchain ever contracted
//	   through that round trip, d and e in (E1)/(E2) change and section 3 has to be
//	   redone.
//	P4 STANDARD - binary32 semantics for noFmaMulAddF32, used only in Lemma D1; its
//	   body is pinned below too.
//	P5 PROVEN here - bid_estimate_decimal_digits[i] = D(2^i) for i <= 114, checked in
//	   this certificate's own band test rather than inherited from the sibling table
//	   test, so a coordinated edit of E[i] and P[i] together fails here too;
//	   bid_power10_index_binexp_128[i] = 10^E[i] for i <= 113, which is every index
//	   section 2's range argument leaves reachable; the check stops there because the
//	   identity is not a property of the whole table - indices 118, 119, 121 and 122
//	   do not satisfy it, and only the CX bound keeps them out of reach - and
//	   bid_power10_table_128[i] = 10^i for every index the band code can reach:
//	   [33,38] on the scale <= 38 branch, [2,30] and 37 on the scale > 38 branch,
//	   and 34 inside unpack_BID128_value, whose comparison against that entry is
//	   __unsigned_compare_ge_128 - pinned under P10, since the entry's value decides
//	   nothing on its own if the comparison that reads it can answer anything else.
//	   The check below runs over i = 0..38, the
//	   whole table, which is a superset of those. Without it "C256 = CX*10^scale" is
//	   a claim about a table entry nobody read, and section 2's band bound - the
//	   premise the whole of Theorem S is stated over - would rest on it.
//	P6 PROVEN here - Lemma D2's minimum 1.01412048... vs the required 1.000000119...
//	P7 PROVEN here - scale ranges, non-truncation of the __mul_128x128_low branch,
//	   and A in [10^66,10^68).
//	P8 PROVEN here - ey in {110..113} and every shift count in (0,64).
//	P9 PROVEN here - the four term bounds and the resulting interval, as exact
//	   rationals.
//	P10 PINNED, not re-proven - the integer primitives the exactness claims of steps
//	   3-5 rest on (__mul_64x256_to_320, __mul_64x320_to_384, __mul_128x128_to_256,
//	   __sqr128_to_256, __mul_64x64_to_128, the carry/borrow helpers, and the
//	   __mul_64x128_full / __add_128_64 that two of them call) compute their
//	   schoolbook limb decompositions exactly, plus two primitives section 2 rather
//	   than steps 3-5 reads. __mul_128x128_low: its "does not truncate" step is a
//	   statement about which bits of the 256-bit product that primitive keeps, so a
//	   body that kept different ones would leave C256 = CX*10^scale unproven while
//	   every other pin held. __unsigned_compare_ge_128: it is the comparison
//	   unpack_BID128_value's canonicality test runs against entry 34, so "CX < 10^34"
//	   - the bound that holds the digit-table indices at or below 113, and with them
//	   the whole band - is a claim about its body on the one return that yields the
//	   coefficient the band reads, and about nothing else pinned here. A body that
//	   answered false would widen CX to [1, 2^113) with every other pin still holding.
//	   That unpacker's other returns are outside this claim: what holds them is the
//	   literal false valid flag on each, pinned by the return census of section 5, and
//	   neither pin stands in for the other. Boundary choice, stated explicitly:
//	   this file pins their *bodies*, not merely their call identity, so any token
//	   change in them fails this certificate - but it does not re-derive their
//	   exactness, which is the port-wide property the differential vector gates
//	   exercise. Their transitive dependency is math/bits.Mul64, a standard-library
//	   primitive, and that is where this scan stops.
//	P11 NOT CLAIMED - sites 17 (:571) and 18 (:646) as behavioural questions;
//	   Theorem S does not decide either, and neither does this certificate. Both are
//	   dead on the separate exhaustive-sweep evidence recorded in the closure note
//	   above pinnedSqrtCSWrites. The disclaimer holds of the claims only, not of the
//	   pins: :571 and the :646 cascade are both pinned exactly in both callers, the
//	   latter because site19's guard reads M256 = (CS+1)^2 exactly on the wrap
//	   inputs site19 can fire on, so a relaxed cascade would reduce a certified
//	   site's firing predicate to the site-18 question. The resulting mutant kills at
//	   both lines are token/premise integrity, not this file claiming them dead.
//
// # 5. What is pinned, and where each scan stops
//
//   - the whole statement sequence and signature of bid_long_sqrt128, which
//     Theorem S analyses operation by operation. That pin also closes the world of
//     calls the estimator makes, since a new call would be a new token. Likewise
//     the bodies and signatures of noFmaMulAddF64, noFmaMulAddF32 and the P10
//     primitives.
//   - the digits/scale/C256 construction of both callers as one contiguous run
//     each, plus the coefficient bounds inside unpack_BID128_value and
//     unpack_BID64 and, on top of those, the closed return census of each unpacker.
//     The bound runs pin one exit apiece; the census is what covers the rest, and it
//     is a census rather than a text pin per exit because two of
//     unpack_BID128_value's returns render identically.
//   - CX itself, in both callers: the closed world of writes to it, no nested
//     declaration that could shadow that world's binding, the text of the unpack
//     statement that produces it, the `if !validBool` early exit pinned to sit
//     immediately after that statement and to return on every path out of its body,
//     and no write to CX from the unpack to the estimator call. Those are what make
//     A = CX*10^scale a statement about the unpacker's output rather than about
//     whatever happens to be in CX at the band run; without them the unpacker bounds
//     pinned just above are attached to nothing. exponent_x is deliberately left
//     open: the band reads it only through `scale += exponent_q & 1`, whose value is
//     in {0,1} for every int, and P7 quantifies over both. The scan stops at the two
//     unpackers, and what it stops on is this: their returns are a closed world, and on
//     every one of them either the coefficient is held by the clamp pinned above it
//     (< 10^34 in unpack_BID128_value, < 10^16 in unpack_BID64) or the valid flag is the
//     literal false, which the `if !validBool` exit turns into a return before the band
//     run. Nothing upstream of the unpackers can produce a third case, which is why the
//     scan does not have to follow x any further back.
//   - the package-wide census of calls to bid_long_sqrt128: exactly the two pinned
//     sites, both unguarded, both passing C256. A third caller is not covered by
//     the band premise and fails here.
//   - the package-wide no-method censuses of BID_UINT128 and BID_UINT256, the types
//     of CS and C256. Every write world and no-write scan in this certificate is an
//     AST scan for assignments, inc/dec, ValueSpec initialisers and address-of, and
//     a pointer-receiver method call is none of those shapes: `C256.m()` on a
//     `func (c *BID_UINT256) m()` rewrites C256 invisibly to all of them. Both
//     censuses run here rather than being inherited from certificate 5, so this
//     certificate fails independently of it. They pin each type's exact struct
//     shape too, which is what rules out a promoted method from an embedded field,
//     and resolve receivers through package-level aliases. They stop at the
//     package: Go requires a method's receiver base type to be declared in the same
//     package, so no import can add one.
//   - per site: the (rnd_mode & 3) == 0 mode split and the site's placement in its
//     else; the CS^2 / +2CS / +1 cascade feeding the guard; the guard's condition
//     text *and* its comparison operators in source order, whose last element is
//     the "<=" the predicate derivation uses; the increment tokens, checked to be
//     token.INC on the pinned operand so the aor:inc->dec mutant flips the
//     certificate; the closed world of writes to CS and C256 in each caller; the absence
//     of any loop or goto in either caller, without which that closed world would still
//     not say *which* of those writes the guard reads; and no-write scans over the two
//     stretches where the proof needs CS and M256 unchanged (estimator return -> mode
//     split, and the (CS+1)^2 build -> guard). Those scans are scoped to the else block
//     rather than the whole function because the sibling arm of the M256 > C256 test does
//     write CS.

// pinnedSqrtEstimatorBody is the complete statement sequence of bid_long_sqrt128
// (bid128_sqrt.go:106-203) - the function Theorem S models operation by operation.
var pinnedSqrtEstimatorBody = []string{
	"var S BID_UINT256",
	"var ES, ARS1, ES2 BID_UINT128",
	"var ARS00 BID_UINT256",
	"var AE, AE2 BID_UINT256",
	"var CY, MY, ES32 uint64",
	"l64 := math.Float64frombits(0x43f0000000000000)",
	"l128 := l64 * l64",
	"lx := math.Float64frombits(math.Float64bits(float64(C256.w3) * l64 * l128))",
	"lx = noFmaMulAddF64(float64(C256.w2), l128, lx)",
	"lx = noFmaMulAddF64(float64(C256.w1), l64, lx)",
	"l0 := float64(C256.w0)",
	"lx = lx + l0",
	"ly_d := 1.0 / math.Sqrt(lx)",
	"ly_i := math.Float64bits(ly_d)",
	"MY = (ly_i & 0x000fffffffffffff) | 0x0010000000000000",
	"ey := int(0x3ff - (ly_i >> 52))",
	"ARS0 := __mul_64x256_to_320(MY, C256)",
	"ARS := __mul_64x320_to_384(MY, ARS0)",
	"k := (ey << 1) + 104 - 128 - 192",
	"k2 := 64 - k",
	"ES.lo = (ARS.w3 >> uint(k+1)) | (ARS.w4 << uint(k2-1))",
	"ES.hi = (ARS.w4 >> uint(k)) | (ARS.w5 << uint(k2))",
	"ES.hi = uint64(int64(ES.hi) >> 1)",
	"ARS1.lo = ARS0.w3",
	"ARS1.hi = ARS0.w4",
	"ARS00.w0 = ARS0.w1",
	"ARS00.w1 = ARS0.w2",
	"ARS00.w2 = ARS0.w3",
	"ARS00.w3 = ARS0.w4",
	"if int64(ES.hi) < 0 { ES.lo = -ES.lo ES.hi = -ES.hi if ES.lo != 0 { ES.hi-- } " +
		"AE = __mul_128x128_to_256(ES, ARS1) " +
		"S.w0, CY = __add_carry_out(ARS00.w0, AE.w0) " +
		"S.w1, CY = __add_carry_in_out(ARS00.w1, AE.w1, CY) " +
		"S.w2, CY = __add_carry_in_out(ARS00.w2, AE.w2, CY) " +
		"S.w3 = ARS00.w3 + AE.w3 + CY } else { " +
		"AE = __mul_128x128_to_256(ES, ARS1) " +
		"S.w0, CY = __sub_borrow_out(ARS00.w0, AE.w0) " +
		"S.w1, CY = __sub_borrow_in_out(ARS00.w1, AE.w1, CY) " +
		"S.w2, CY = __sub_borrow_in_out(ARS00.w2, AE.w2, CY) " +
		"S.w3 = ARS00.w3 - AE.w3 - CY }",
	"ES32 = ES.hi + (ES.hi >> 1)",
	"ES2 = __mul_64x64_to_128(ES32, ES.hi)",
	"AE2 = __mul_128x128_to_256(ES2, ARS1)",
	"S.w0, CY = __add_carry_out(S.w0, AE2.w0)",
	"S.w1, CY = __add_carry_in_out(S.w1, AE2.w1, CY)",
	"S.w2, CY = __add_carry_in_out(S.w2, AE2.w2, CY)",
	"S.w3 = S.w3 + AE2.w3 + CY",
	"k = ey + 51 - 128",
	"k2 = 64 - k",
	"S.w0 = (S.w1 >> uint(k)) | (S.w2 << uint(k2))",
	"S.w1 = (S.w2 >> uint(k)) | (S.w3 << uint(k2))",
	"S.w0++",
	"if S.w0 == 0 { S.w1++ }",
	"var CS BID_UINT128",
	"CS.lo = (S.w1 << 63) | (S.w0 >> 1)",
	"CS.hi = S.w1 >> 1",
	"return CS",
}

// pinnedSqrtHelper is one function whose body a premise of Theorem S reads: the
// two FMA-suppressing wrappers of P3/P4, and the integer primitives of P10.
type pinnedSqrtHelper struct {
	file, fn, sig string
	body          []string
}

var pinnedSqrtHelpers = []pinnedSqrtHelper{
	{"internal.go", "noFmaMulAddF64", "func(a, b, c float64) float64", []string{
		"return math.Float64frombits(math.Float64bits(a*b)) + c",
	}},
	{"internal.go", "noFmaMulAddF32", "func(a float32, b float32, c float32) float32", []string{
		"return math.Float32frombits(math.Float32bits(a*b)) + c",
	}},
	{"internal.go", "__mul_64x256_to_320", "func(A uint64, B BID_UINT256) BID_UINT320", []string{
		"var P BID_UINT320",
		"var c uint64",
		"lP0 := __mul_64x64_to_128(A, B.w0)",
		"lP1 := __mul_64x64_to_128(A, B.w1)",
		"lP2 := __mul_64x64_to_128(A, B.w2)",
		"lP3 := __mul_64x64_to_128(A, B.w3)",
		"P.w0 = lP0.lo",
		"P.w1, c = __add_carry_out(lP1.lo, lP0.hi)",
		"P.w2, c = __add_carry_in_out(lP2.lo, lP1.hi, c)",
		"P.w3, c = __add_carry_in_out(lP3.lo, lP2.hi, c)",
		"P.w4 = lP3.hi + c",
		"return P",
	}},
	{"internal.go", "__mul_64x320_to_384", "func(A uint64, B BID_UINT320) BID_UINT384", []string{
		"var P BID_UINT384",
		"var c uint64",
		"lP0 := __mul_64x64_to_128(A, B.w0)",
		"lP1 := __mul_64x64_to_128(A, B.w1)",
		"lP2 := __mul_64x64_to_128(A, B.w2)",
		"lP3 := __mul_64x64_to_128(A, B.w3)",
		"lP4 := __mul_64x64_to_128(A, B.w4)",
		"P.w0 = lP0.lo",
		"P.w1, c = __add_carry_out(lP1.lo, lP0.hi)",
		"P.w2, c = __add_carry_in_out(lP2.lo, lP1.hi, c)",
		"P.w3, c = __add_carry_in_out(lP3.lo, lP2.hi, c)",
		"P.w4, c = __add_carry_in_out(lP4.lo, lP3.hi, c)",
		"P.w5 = lP4.hi + c",
		"return P",
	}},
	{"internal.go", "__mul_128x128_to_256", "func(a, b BID_UINT128) BID_UINT256", []string{
		"var p256 BID_UINT256",
		"var cy1, cy2 uint64",
		"phl, qll := __mul_64x128_full(a.lo, b)",
		"phh, qlh := __mul_64x128_full(a.hi, b)",
		"p256.w0 = qll.lo",
		"p256.w1, cy1 = __add_carry_out(qlh.lo, qll.hi)",
		"p256.w2, cy2 = __add_carry_in_out(qlh.hi, phl, cy1)",
		"p256.w3 = phh + cy2",
		"return p256",
	}},
	{"internal.go", "__sqr128_to_256", "func(A BID_UINT128) BID_UINT256", []string{
		"var P256 BID_UINT256",
		"var c1, c2 uint64",
		"Qhh := __mul_64x64_to_128(A.hi, A.hi)",
		"Qlh := __mul_64x64_to_128(A.lo, A.hi)",
		"Qhh.hi += (Qlh.hi >> 63)",
		"Qlh.hi = (Qlh.hi + Qlh.hi) | (Qlh.lo >> 63)",
		"Qlh.lo += Qlh.lo",
		"Qll := __mul_64x64_to_128(A.lo, A.lo)",
		"P256.w1, c1 = __add_carry_out(Qlh.lo, Qll.hi)",
		"P256.w0 = Qll.lo",
		"P256.w2, c2 = __add_carry_in_out(Qlh.hi, Qhh.lo, c1)",
		"P256.w3 = Qhh.hi + c2",
		"return P256",
	}},
	{"internal.go", "__mul_64x64_to_128", "func(cx, cy uint64) BID_UINT128", []string{
		"hi, lo := bits.Mul64(cx, cy)",
		"return BID_UINT128{lo: lo, hi: hi}",
	}},
	// Read by section 2, not by steps 3-5: the scale > 38 branch of both callers
	// builds CX1 with this primitive, and the band argument's "does not truncate"
	// step says the product it drops is zero. That is a claim about which bits the
	// body keeps - the low 128 of the 256-bit product, with the high half never
	// formed - so the body is pinned here alongside the exactness primitives.
	{"bid128_div.go", "__mul_128x128_low", "func(A, B BID_UINT128) BID_UINT128", []string{
		"var Ql BID_UINT128",
		"ALBL := __mul_64x64_to_128(A.lo, B.lo)",
		"QM64 := B.lo*A.hi + A.lo*B.hi",
		"Ql.lo = ALBL.lo",
		"Ql.hi = QM64 + ALBL.hi",
		"return Ql",
	}},
	// Read by section 2 as well, on the other side of the band argument: this is the
	// comparison unpack_BID128_value's canonicality test runs against 10^34, so the
	// CX < 10^34 clamp - and with it the be <= 112 that keeps the digit-table indices
	// inside the range P5 checks - is a claim about this body. A body that answered
	// false would leave the clamp dead, CX as wide as [1, 2^113), and every other pin
	// in this file still holding.
	{"internal.go", "__unsigned_compare_ge_128", "func(a, b BID_UINT128) bool", []string{
		"return (a.hi > b.hi) || ((a.hi == b.hi) && (a.lo >= b.lo))",
	}},
	{"internal.go", "__mul_64x128_full", "func(a uint64, b BID_UINT128) (ph uint64, ql BID_UINT128)", []string{
		"albh := __mul_64x64_to_128(a, b.hi)",
		"albl := __mul_64x64_to_128(a, b.lo)",
		"ql.lo = albl.lo",
		"qm2 := __add_128_64(albh, albl.hi)",
		"ql.hi = qm2.lo",
		"ph = qm2.hi",
		"return",
	}},
	{"internal.go", "__add_128_64", "func(a BID_UINT128, b uint64) BID_UINT128", []string{
		"var r BID_UINT128",
		"r64h := a.hi",
		"r.lo = b + a.lo",
		"if r.lo < b { r64h++ }",
		"r.hi = r64h",
		"return r",
	}},
	{"internal.go", "__add_carry_out", "func(x, y uint64) (s uint64, cy uint64)", []string{
		"s = x + y",
		"if s < x { cy = 1 }",
		"return",
	}},
	{"internal.go", "__add_carry_in_out", "func(x, y, ci uint64) (s uint64, cy uint64)", []string{
		"x1 := x + ci",
		"s = x1 + y",
		"if (s < x1) || (x1 < ci) { cy = 1 }",
		"return",
	}},
	{"internal.go", "__sub_borrow_out", "func(x, y uint64) (s uint64, cy uint64)", []string{
		"s = x - y",
		"if s > x { cy = 1 }",
		"return",
	}},
	{"bid128_div.go", "__sub_borrow_in_out", "func(x, y, ci uint64) (s uint64, co uint64)", []string{
		"x1 := x - ci",
		"if x1 > x { co = 1 }",
		"s = x1 - y",
		"if s > x1 { co = 1 }",
		"return",
	}},
}

// sqrtBandRun is the digits / scale / C256 / estimator-call run of one caller.
// The two callers differ in exactly two tokens: the exponent-bias constant the
// perfect-square early exit repacks with, and whether the parity correction
// parenthesises its operand. Both are parameters here rather than a wildcard, so
// each caller is still pinned token for token.
func sqrtBandRun(bias, parity string) []string {
	return []string{
		"f64_i := uint32(0x5f800000)",
		"f64_d := math.Float32frombits(f64_i)",
		"fx_d := noFmaMulAddF32(float32(CX.hi), f64_d, float32(CX.lo))",
		"fx_i := math.Float32bits(fx_d)",
		"bin_expon_cx = int((fx_i>>23)&0xff) - 0x7f",
		"digits = bid_estimate_decimal_digits[bin_expon_cx]",
		"A10 = CX",
		"if (exponent_x & 1) != 0 { A10.hi = (CX.hi << 3) | (CX.lo >> 61) A10.lo = CX.lo << 3 " +
			"CX2.hi = (CX.hi << 1) | (CX.lo >> 63) CX2.lo = CX.lo << 1 A10 = __add_128_128(A10, CX2) }",
		"CS.lo = short_sqrt128(A10)",
		"CS.hi = 0",
		"if CS.lo*CS.lo == A10.lo { S2 = __mul_64x64_to_128_fast(CS.lo, CS.lo) " +
			"if S2.hi == A10.hi { res = very_fast_get_BID128(0, (exponent_x+" + bias + ")>>1, CS) return res, pfpsf } }",
		"D = int64(CX.hi) - int64(bid_power10_index_binexp_128[bin_expon_cx].hi)",
		"if D > 0 || (D == 0 && CX.lo >= bid_power10_index_binexp_128[bin_expon_cx].lo) { digits++ }",
		"scale = 67 - digits",
		"exponent_q = exponent_x - scale",
		"scale += " + parity,
		"if scale > 38 { T128 = bid_power10_table_128[scale-37] CX1 = __mul_128x128_low(CX, T128) " +
			"TP128 = bid_power10_table_128[37] C256 = __mul_128x128_to_256(CX1, TP128) } " +
			"else { T128 = bid_power10_table_128[scale] C256 = __mul_128x128_to_256(CX, T128) }",
		"C4.w3 = (C256.w3 << 2) | (C256.w2 >> 62)",
		"C4.w2 = (C256.w2 << 2) | (C256.w1 >> 62)",
		"C4.w1 = (C256.w1 << 2) | (C256.w0 >> 62)",
		"C4.w0 = C256.w0 << 2",
		"CS = bid_long_sqrt128(C256)",
	}
}

// Closure note for sites 17 and 18. Both are *behavioural* questions Theorem S
// does not decide, because both need the estimator to return CS.lo = 2^64-1 and
// Theorem S does not exclude that. They are decided - both DEAD - by the exhaustive
// sweep recorded below. That verdict is external evidence this file does not
// re-derive: nothing here re-runs the sweep, and the pins below are exact-token
// pins whose failures are structural, so the two must be read apart.
//
// # The sweep, and why it decides both sites
//
// Firing predicates, read off the pinned source, with A := C256 as a mathematical
// integer, M := isqrt(A), CS := the raw 128-bit return of bid_long_sqrt128 and
// B := 2^64:
//
//	site17 (:571, the CS.hi++ of the nearest-path round-up) fires
//	  <=> (rnd_mode & 3) == 0 and CS = -1 mod B and A >= CS^2 + CS + 1.
//	      The guard at :562-568 is 4A > (2CS+1)^2; 4A = 0 mod 4 sharpens that to
//	      A >= CS^2+CS+1, and :569-570 is the CS.lo++ wrap.
//	site18 (:646, the M256.w2++ of the (CS+1)^2 carry cascade) fires
//	  <=> (rnd_mode & 3) != 0 and CS = -1 mod B and CS^2 <= A.
//	      The w0/w1 wraps at :642-645 say (CS+1)^2 = 0 mod 2^128, and with
//	      CS+1 <= 10^34+1 < 2^114 that is v2(CS+1) >= 64, i.e. CS = -1 mod B.
//
// Both force M = CS: site18 requires CS^2 <= A and site17 requires
// A >= CS^2+CS+1 > CS^2, so M >= CS either way, while Theorem S bounds the
// estimator from the other side with CS >= M. Hence if any input fires either
// site, its A is reachable and satisfies isqrt(A) = -1 mod B, so enumerating
// {A reachable : isqrt(A) = -1 mod B} DECIDES both sites.
//
// Parametrise M+1 = j*B. Then M^2 = j^2B^2-2jB+1 and isqrt(A) = M <=> A lies in
// [M^2, M^2+2M] = [j^2B^2-2jB+1, j^2B^2-1], a window of width 2M+1 <= 2*10^34.
// M in [10^33, 10^34-1], from the reachable band A in [10^66,10^68) of section 2,
// gives
//
//	j in [54210108624276, 542101086242752]  -  487,890,977,618,477 values.
//
// A reachable A is a multiple of P := 10^q, and where that comes from is
// load-bearing rather than incidental. Both sites live in Bid128dSqrt, whose
// CX < 10^16 bounds the digit count by D <= 16 (section 2), so scale = 67-D+p with
// the band parity p in {0,1} gives scale >= 51+p, and A = CX*10^scale is a multiple
// of 10^(51+p). Write q := 51+p, so q = 51 on [10^66,10^67) and q = 52 on
// [10^67,10^68). The same step in Bid128Sqrt would only reach D <= 34, i.e. an
// exponent floor of 33+p in place of 51+p, far too small for what follows - which
// is why this argument runs on Bid128dSqrt's coefficient bound and not on one
// shared by both entry points.
//
// The window is 5.0e16 times narrower than P, so it holds at most one multiple of
// P. With G := (j^2*B^2) mod P, the largest multiple of P below j^2B^2 is j^2B^2-G
// and the next one down is a whole P lower, far outside the window, so exactly:
//
//	a reachable A with isqrt(A) = jB-1 exists  <=>  1 <= G <= 2jB-1,
//	and then A = j^2*2^128 - G.
//
// (The G = 0 escape, where the next multiple down could still land in the window,
// needs P <= 2jB-1 and is impossible here.) Since P = 2^q*5^q and B^2 = 2^128 = 0
// mod 2^q, the test runs on half the limbs as G = 2^q*G' with
// G' = (j^2 * 2^(128-q)) mod 5^q.
//
// # Result
//
// Every j in that range was tested, in two legs split at the band boundary, with
// the boundary j tested under both moduli:
//
//	leg 1  A in [10^66,10^67)  P=10^51  27,293 blocks  117,217,306,833,571 cands  0 hits
//	leg 2  A in [10^67,10^68)  P=10^52  86,305 blocks  370,673,670,784,907 cands  0 hits
//
// 113,598 blocks, and no j admitted a reachable A. The two candidate counts sum to
// 487,890,977,618,478, one more than the 487,890,977,618,477 distinct j, because
// the shared boundary j is counted once per modulus.
//
// So no reachable input satisfies site17's or site18's firing predicate, and both
// sites are unreachable. Stated exactly, because the difference matters: what the
// sweep empties is {A reachable : isqrt(A) = -1 mod B}, and it is the CS = M
// reduction above that turns emptiness of that set into the sites' CS = -1 mod B
// condition. The sweep does *not* establish that the estimator never returns
// CS = -1 mod B at all - Theorem S also permits CS = M+1, so a reachable A with
// M = -2 mod B would produce such a CS without being enumerated here - and neither
// site fires on that branch, since both require M >= CS while CS = M+1.
//
// The enumerator behind those counts carries its own validation. Its scaled-down
// analogues (B = 2^8..2^14 against P = 10^7..10^13) reproduce brute force exactly
// on ranges where brute force finds 9882, 41021, 1645 and 53 real hits, so the
// production accept path is known to report hits when hits exist rather than being
// vacuously empty, and its recurrence is cross-checked against math/big over 80,000
// consecutive j through the production loop. Both of those are self-test modes, run
// against the same code and not timestamped against the sweep; the one check the
// binary does perform on every invocation before any sweep work starts is the
// acceptance-bound overflow audit at j_hi. This matters more than the hit count
// looks: on the density heuristic that G is equidistributed mod P - an estimate,
// not a claim this file makes - the expected number of hits over the whole range is
// ~1e-3, so 0 was the near-certain outcome and the run's discriminating power rests
// on the enumerator being right rather than on the count being zero.
//
// The sweep says nothing about site19, and is not a second argument for it: the
// enumeration is over M = -1 mod B, which coincides with a site's CS = -1 mod B
// condition only through the CS = M reduction above, and that reduction is exactly
// what fails for site19, whose guard requires CS <= M-1. Site19 stays closed by
// Theorem S alone, as certified below.
//
// # What the pins do, which is a different thing
//
// Site 17's write-world entry is now exact in both callers, like every other entry.
// The one-token ++/-- relaxation it used to carry was scope honesty while the
// question was open; with the question closed it would be a hole in the closed
// write world, so it is gone, and with it the machinery that expressed it. Site
// 18's :646 cascade was already pinned exactly in both callers, for a reason that
// never depended on the sweep: site19 fires only when CS.lo = 2^64-1, precisely
// where the +1 carries out of w0 and w1 and reaches the w2 step, so an `M256.w2--`
// rendering would stop the guard at :652-658 from reading M256 = (CS+1)^2 on any
// input site19 could fire on, reducing site19's deadness claim to the site-18
// question. Bid128Sqrt's cascade at :401-414, which builds the (CS+1)^2 that
// site16's guard at :415-421 reads, is exact for the same reason.
//
// Consequence for mutation accounting, stated rather than left to be inferred: the
// aor:inc->dec mutants at :571 and :646 both FAIL this certificate, and both
// failures are **structural** - a pinned token changed - not reachability verdicts
// this file derives. The reachability verdict for those two lines is the sweep
// above and nothing else.

// pinnedSqrtCSWrites is the closed world of writes to CS in one caller: the
// short-path seed, the estimator result, and the ten corrections of the two
// rounding paths. Both callers have the same shape and every entry is exact in
// both. The site 17 relaxation this list used to carry for Bid128dSqrt is gone,
// retired by the sweep recorded in the closure note above.
var pinnedSqrtCSWrites = []string{
	"CS.lo = short_sqrt128(A10)",
	"CS.hi = 0",
	"CS = bid_long_sqrt128(C256)",
	"CS.lo++", "CS.hi++", // nearest path, round up; the CS.hi++ is site 17 (:571)
	"CS.hi--", "CS.lo--", // nearest path, round down
	"CS.hi--", "CS.lo--", // directed path, CS^2 > C256, first decrement
	"CS.hi--", "CS.lo--", // directed path, CS^2 > C256, second decrement
	"CS.lo++", "CS.hi++", // directed path, the certified undershoot arm
	"CS.lo++", "CS.hi++", // directed path, BID_ROUNDING_UP tail
}

// pinnedSqrtC256Writes is the closed world of writes to C256 in one caller: the
// two arms of the scale dispatch and nothing after them, which is what lets the
// guard at :415 / :652 read the same A the estimator was handed.
var pinnedSqrtC256Writes = []string{
	"C256 = __mul_128x128_to_256(CX1, TP128)",
	"C256 = __mul_128x128_to_256(CX, T128)",
}

const pinnedSqrtOvershootCond = "M256.w3 > C256.w3 || (M256.w3 == C256.w3 && " +
	"(M256.w2 > C256.w2 || (M256.w2 == C256.w2 && " +
	"(M256.w1 > C256.w1 || (M256.w1 == C256.w1 && M256.w0 > C256.w0)))))"

// pinnedSqrtUndershootCond is the guard both certified arms sit behind: with
// M256 = (CS+1)^2 it is exactly (CS+1)^2 <= C256. The final comparator is the
// "<=" at :421 / :658; pinSqrtComparisonOps re-checks the operator sequence so a
// cmp mutation on it cannot slip through a text pin alone.
const pinnedSqrtUndershootCond = "M256.w3 < C256.w3 || (M256.w3 == C256.w3 && " +
	"(M256.w2 < C256.w2 || (M256.w2 == C256.w2 && " +
	"(M256.w1 < C256.w1 || (M256.w1 == C256.w1 && M256.w0 <= C256.w0)))))"

// pinnedSqrtPlusOneBuild is the M256 <- CS^2 + 2*CS part of the cascade that feeds
// the guard, taken as one contiguous run of the else block it opens. The +1 carry
// cascade that follows it is pinnedSqrtPlusOneCascade.
var pinnedSqrtPlusOneBuild = []string{
	"M256.w0, Carry = __add_carry_out(M256.w0, C8.w0)",
	"M256.w1, Carry = __add_carry_in_out(M256.w1, C8.w1, Carry)",
	"M256.w2, Carry = __add_carry_in_out(M256.w2, 0, Carry)",
	"M256.w3 = M256.w3 + Carry",
	"M256.w0++",
}

// pinnedSqrtPlusOneCascade is the statement the +1 of (CS+1)^2 carries through, in
// both callers, as one exact text. It is deliberately *not* relaxed for the :646
// M256.w2++ inside Bid128dSqrt's copy, even though site 18 is disclaimed: the
// cascade is the premise that makes M256 equal (CS+1)^2 at the guard, and site19 -
// which this file does certify - fires only when CS.lo == 2^64-1, which is exactly
// the region where the w0/w1 wrap reaches the w2 step. Admitting `M256.w2--` there
// would therefore leave site19's own firing predicate undecided. See the closure
// note above pinnedSqrtCSWrites for what the resulting :646 kill does and does not
// mean.
const pinnedSqrtPlusOneCascade = "if M256.w0 == 0 { M256.w1++ if M256.w1 == 0 { M256.w2++ if M256.w2 == 0 { M256.w3++ } } }"

// comparisonOps returns e's comparison operators in source order. A text pin
// already fails on a changed operator, but recording the sequence separately is
// what lets the failure name the comparator the firing predicate was read off.
func comparisonOps(e ast.Expr) []token.Token {
	var out []token.Token
	ast.Inspect(e, func(n ast.Node) bool {
		if be, ok := n.(*ast.BinaryExpr); ok {
			switch be.Op {
			case token.LSS, token.LEQ, token.GTR, token.GEQ, token.EQL, token.NEQ:
				out = append(out, be.Op)
			}
		}
		return true
	})
	return out
}

func (p *portFile) pinComparisonOps(t *testing.T, label string, e ast.Expr, want []token.Token) {
	t.Helper()
	got := comparisonOps(e)
	if len(got) != len(want) {
		t.Fatalf("%s: %s (line %d): %d comparison operators, pinned %d: got %v want %v",
			p.name, label, p.line(e), len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s: %s (line %d): comparison operator %d is %v, pinned %v",
				p.name, label, p.line(e), i, got[i], want[i])
		}
	}
}

// pinIncrement fails unless s is exactly `operand++`. The token check is the
// half that the aor:inc->dec mutant flips: a decrement at the same position
// renders differently and carries token.DEC.
func (p *portFile) pinIncrement(t *testing.T, label string, s ast.Stmt, operand string) {
	t.Helper()
	inc, ok := s.(*ast.IncDecStmt)
	if !ok {
		t.Fatalf("%s: %s: statement at line %d is %T, want `%s++`", p.name, label, p.line(s), s, operand)
	}
	if got := p.flat(t, inc.X); got != operand {
		t.Fatalf("%s: %s (line %d): operand is %s, pinned %s", p.name, label, p.line(inc), got, operand)
	}
	if inc.Tok != token.INC {
		t.Fatalf("%s: %s (line %d): operator is %v, pinned ++ - an aor:inc->dec mutation here is exactly what "+
			"this certificate exists to fail on", p.name, label, p.line(inc), inc.Tok)
	}
}

// theoremSNotice is appended to every label of a pin that Theorem S reads
// operation by operation, so a drift report names the consequence rather than
// only the token: the bound in section 3 is a statement about this exact source,
// and re-pinning it without redoing the proof would leave the certificate
// attached to an algorithm nobody has analysed.
const theoremSNotice = " [the certificate reads this source operation by operation - Theorem S in section 3, or " +
	"section 2's band argument for the primitive it names; that proof must be redone before this pin is updated]"

// TestBid128SqrtEstimatorSourceIsPinned anchors everything Theorem S models: the
// estimator body itself, the FMA-suppressing wrappers of premises P3/P4, the
// integer primitives of P10, and the closed world of calls to the estimator.
func TestBid128SqrtEstimatorSourceIsPinned(t *testing.T) {
	p := loadPortFile(t, "bid128_sqrt.go")
	fn := p.funcDecl(t, "bid_long_sqrt128")
	p.pinNode(t, "bid_long_sqrt128 signature"+theoremSNotice, fn.Type, "func(C256 BID_UINT256) BID_UINT128")
	p.pinStmtCount(t, "bid_long_sqrt128 body"+theoremSNotice, fn.Body, len(pinnedSqrtEstimatorBody))
	p.pinRun(t, "bid_long_sqrt128 body"+theoremSNotice, fn.Body, 0, pinnedSqrtEstimatorBody)
	p.pinForwardOnly(t, fn)

	for _, h := range pinnedSqrtHelpers {
		hp := loadPortFile(t, h.file)
		hfn := hp.funcDecl(t, h.fn)
		hp.pinNode(t, h.fn+" signature"+theoremSNotice, hfn.Type, h.sig)
		hp.pinStmtCount(t, h.fn+" body"+theoremSNotice, hfn.Body, len(h.body))
		hp.pinRun(t, h.fn+" body"+theoremSNotice, hfn.Body, 0, h.body)
	}

	// --- census: the estimator's closed world of callers. A third one would not
	// be covered by the reachable-band premise of section 2.
	files := portPkg(t)
	calls := portCallsTo(t, files, "bid_long_sqrt128")
	want := []struct{ key, guards string }{
		{"bid128_sqrt.go/Bid128Sqrt/bid_long_sqrt128(C256)", ""},
		{"bid128_sqrt.go/Bid128dSqrt/bid_long_sqrt128(C256)", ""},
	}
	if len(calls) != len(want) {
		var keys []string
		for _, c := range calls {
			keys = append(keys, c.key)
		}
		t.Fatalf("bid_long_sqrt128 has %d call sites, pinned %d: %v", len(calls), len(want), keys)
	}
	for i, w := range want {
		if calls[i].key != w.key || calls[i].guards != w.guards {
			t.Fatalf("bid_long_sqrt128 call site %d drifted\n got: %s under %q\nwant: %s under %q",
				i, calls[i].key, calls[i].guards, w.key, w.guards)
		}
		pinSlots(t, "bid_long_sqrt128 call "+w.key, fn, calls[i].args, map[int]string{0: "C256"})
	}
}

// TestBid128SqrtReachableBandIsPinned anchors and then re-derives the band
// premise: C256 in [10^66, 10^68) for every input that reaches the estimator.
func TestBid128SqrtReachableBandIsPinned(t *testing.T) {
	p := loadPortFile(t, "bid128_sqrt.go")
	for _, c := range []struct {
		fn, bias, parity string
		unpack           string
		cxWrites         []string
	}{
		{
			fn: "Bid128Sqrt", bias: "EXPONENT_BIAS128", parity: "(exponent_q & 1)",
			unpack:   "sign_x, exponent_x, CX, validBool := unpack_BID128_value(x)",
			cxWrites: []string{"sign_x, exponent_x, CX, validBool := unpack_BID128_value(x)"},
		},
		{
			fn: "Bid128dSqrt", bias: "DECIMAL_EXPONENT_BIAS_128", parity: "exponent_q & 1",
			unpack:   "sign_x, exponent_x, CX.lo, validBool = unpack_BID64(x)",
			cxWrites: []string{"CX.hi = 0", "sign_x, exponent_x, CX.lo, validBool = unpack_BID64(x)"},
		},
	} {
		fn := p.funcDecl(t, c.fn)
		blk := p.blockWithRun(t, fn, sqrtBandRun(c.bias, c.parity))
		if blk != fn.Body {
			t.Fatalf("bid128_sqrt.go: %s: the pinned band run is not at the top level of the function body (line %d)",
				c.fn, p.line(blk))
		}
		p.pinWrites(t, fn, "C256", pinnedSqrtC256Writes)
		p.pinNoNestedDecl(t, fn, "C256")
		p.pinNoGoto(t, fn)

		// --- anchor: CX, the operand the whole band premise is stated over. A = CX*10^scale
		// is a claim about the coefficient the unpacker returned, and the unpacker bounds
		// pinned below say nothing about any other value that could be sitting in CX by the
		// time the band run reads it. Four pins close that: the whole world of writes to CX
		// in the caller, no nested declaration that could shadow the binding those writes
		// describe, the unpack statement's own text, and - since the unpacker's bound holds
		// only for the coefficients it calls valid - the `if !validBool` early exit, pinned
		// to sit immediately after the unpack (so nothing can re-decide validBool in
		// between) and to return on every path out of its body. Without them a single
		// `CX = x` before the digit estimate leaves every other pin in this file holding
		// while CX carries exponent bits, indexes the part of bid_power10_index_binexp_128
		// the P5 check below does not cover, and collapses the band Theorem S is stated over.
		p.pinWrites(t, fn, "CX", c.cxWrites)
		p.pinNoNestedDecl(t, fn, "CX")
		topIndex := func(what string, s ast.Stmt) int {
			for i, st := range fn.Body.List {
				if st == s {
					return i
				}
			}
			t.Fatalf("bid128_sqrt.go: %s: %s (line %d) is not a top-level statement of the function body",
				c.fn, what, p.line(s))
			return -1
		}
		unpack := p.stmtWithText(t, fn, c.unpack)
		invalid := p.ifWithCond(t, fn, "!validBool")
		ui := topIndex("the unpack", unpack)
		if ui+1 >= len(fn.Body.List) || fn.Body.List[ui+1] != ast.Stmt(invalid) {
			t.Fatalf("bid128_sqrt.go: %s: the `if !validBool` early exit (line %d) does not immediately follow the "+
				"unpack (line %d), so validBool can be re-decided in between",
				c.fn, p.line(invalid), p.line(unpack))
		}
		p.pinBlockAlwaysReturns(t, c.fn+": the !validBool early exit", invalid.Body)
		if bi := topIndex("the band run", p.stmtWithText(t, fn, "f64_i := uint32(0x5f800000)")); bi <= ui+1 {
			t.Fatalf("bid128_sqrt.go: %s: the band run starts at top-level statement %d, not after the !validBool "+
				"early exit at %d", c.fn, bi, ui+1)
		}
		// Nothing may change CX from the unpack to the estimator call, which is what makes
		// the C256 the estimator is handed the CX*10^scale of section 2. exponent_x is
		// deliberately *not* closed the same way: the band derivation reads it only through
		// `scale += exponent_q & 1`, and that conjunction yields p in {0,1} for every int
		// value it could hold, both of which premise P7 below quantifies over. No bound on
		// exponent_x enters the band.
		p.noWriteBetween(t, fn, "CX", unpack, p.stmtWithText(t, fn, "CS = bid_long_sqrt128(C256)"))
	}

	// --- residual for the C256 write world just pinned: writes() sees assignments, inc/dec, ValueSpec initialisers and
	// address-of, and a pointer-receiver method is none of those. `C256.trim()` on a `func (c *BID_UINT256) trim()` takes
	// &C256 implicitly and can rewrite the value this section calls A, with no statement in either caller that the write
	// census can see. The census below rules that out for the whole package.
	pinTypeIsMethodFree(t, portPkg(t), "BID_UINT256", pinnedUint256Shape,
		"a call `C256.<that method>()` on a pointer receiver rewrites C256 with no assignment, inc/dec or address-of "+
			"for the pinned write world above to see, so A stops being CX*10^scale at the estimator call")

	// --- anchor: the two unpackers, which is where the CX bounds come from and
	// where this scan stops.
	pi := loadPortFile(t, "bid128_internal.go")
	unpack128 := pi.funcDecl(t, "unpack_BID128_value")
	pi.blockWithRun(t, unpack128, []string{
		"coeff := BID_UINT128{lo: x.lo, hi: x.hi & SMALL_COEFF_MASK128}",
		"T34 := bid_power10_table_128[34]",
		"if __unsigned_compare_ge_128(coeff, T34) { coeff.lo = 0 coeff.hi = 0 }",
		"coefficient_x.lo = coeff.lo",
		"coefficient_x.hi = coeff.hi",
		"ex := x.hi >> 49",
		"exponent_x = int(ex) & EXPONENT_MASK128",
		"return sign_x, exponent_x, coefficient_x, (coeff.lo | coeff.hi) != 0",
	})
	pu := loadPortFile(t, "internal.go")
	unpack64 := pu.funcDecl(t, "unpack_BID64")
	pu.blockWithRun(t, unpack64, []string{
		"if coefficient >= 10000000000000000 { coefficient = 0 }",
		"tmp := x >> EXPONENT_SHIFT_LARGE64",
		"exponent = int(tmp & EXPONENT_MASK64)",
		"return sign, exponent, coefficient, coefficient != 0",
	})
	pu.blockWithRun(t, unpack64, []string{
		"tmp := x >> EXPONENT_SHIFT_SMALL64",
		"exponent = int(tmp & EXPONENT_MASK64)",
		"coefficient = x & SMALL_COEFF_MASK64",
		"return sign, exponent, coefficient, coefficient != 0",
	})

	// --- and the closed return census of each unpacker, which is what makes the two runs
	// above a statement about every path out of them. Each run pins one coefficient-yielding
	// exit; neither says anything about the special-encoding exits beside it, which hand back
	// a coefficient no clamp has touched - unpack_BID64's NaN/Inf return carries `x &
	// 0xfe03ffffffffffff`, far above 10^16 - and rely entirely on the valid flag being the
	// literal false for the caller's pinned `if !validBool` exit to fire. Flipping one of those
	// literals to true is a one-token edit that leaves every other pin in this file holding
	// while CX carries an unclamped payload into the digit estimate and the band premise of
	// section 2 collapses. The census is taken over the whole function rather than statement by
	// statement because two of unpack_BID128_value's returns render identically, so no text
	// locator can name either, and because a return added later has to fail here rather than
	// be an exit nobody pinned.
	pi.pinReturnWorld(t, unpack128, []string{
		"return sign_x, exponent_x, coefficient_x, false",
		"return sign_x, exponent_x, coefficient_x, false",
		"return sign_x, exponent_x, coefficient_x, (coeff.lo | coeff.hi) != 0",
	})
	pu.pinReturnWorld(t, unpack64, []string{
		"return sign, exponent, coefficient, false",
		"return sign, exponent, coefficient, coefficient != 0",
		"return sign, exponent, coefficient, coefficient != 0",
	})

	// --- premise P5, both halves over the two tables the digit correction reads:
	// E[b] = D(2^b) and P[b] = 10^E[b]. Lemma D2 needs both - E alone decides the
	// uncorrected digit count and P alone decides the decade correction - and a
	// *coordinated* edit of the pair (E[b] moved, P[b] moved with it to 10^E[b])
	// satisfies the second identity while breaking the first, so checking only
	// P[b] = 10^E[b] would leave the band premise resting on a digit count nobody
	// verified. The E half deliberately duplicates TestPortDigitEstimateTablesAreExact
	// - the sibling table test that owns it port-wide - rather than being inherited
	// from it, so this certificate fails on its own evidence if either table drifts.
	//
	// The P half is checked over exactly the reachable range and no further, because
	// that identity does not hold table-wide: entries 118, 119, 121 and 122 fail it,
	// and what keeps them out of reach is the CX bound pinned above, so widening the
	// loop past maxUsedIndex would fail here rather than prove anything more. The E
	// half does hold table-wide and is run one index further, to maxUsedIndex+1, so
	// the check covers the whole j-range Lemma D2's minimum below is taken over.
	const maxUsedIndex = 113 // BID128: CX < 10^34 < 2^113, and be' <= be+1
	if len(bid_power10_index_binexp_128) <= maxUsedIndex || len(bid_estimate_decimal_digits) <= maxUsedIndex+1 {
		t.Fatalf("digit tables are too short: estimate=%d (need > %d) index_binexp_128=%d (need > %d)",
			len(bid_estimate_decimal_digits), maxUsedIndex+1, len(bid_power10_index_binexp_128), maxUsedIndex)
	}
	for b := 0; b <= maxUsedIndex+1; b++ {
		if got, want := bid_estimate_decimal_digits[b], decDigits(pow2big(b)); got != want {
			t.Fatalf("bid_estimate_decimal_digits[%d] = %d, but 2^%d has %d decimal digits", b, got, b, want)
		}
	}
	for b := 0; b <= maxUsedIndex; b++ {
		want := pow10big(bid_estimate_decimal_digits[b])
		got := new(big.Int).Lsh(new(big.Int).SetUint64(bid_power10_index_binexp_128[b].hi), 64)
		got.Or(got, new(big.Int).SetUint64(bid_power10_index_binexp_128[b].lo))
		if got.Cmp(want) != 0 {
			t.Fatalf("bid_power10_index_binexp_128[%d] = %s, want 10^%d = %s",
				b, got, bid_estimate_decimal_digits[b], want)
		}
	}

	// --- premise P4/P6, Lemma D2: the digit formula is exact whenever
	// 2^j / 10^(D(2^j)-1) >= (1+2^-24)^2 for every j the callers can reach.
	need := new(big.Rat).Mul(
		new(big.Rat).Add(new(big.Rat).SetInt64(1), ratPow2(-24)),
		new(big.Rat).Add(new(big.Rat).SetInt64(1), ratPow2(-24)))
	worst, worstJ := (*big.Rat)(nil), 0
	for j := 1; j <= maxUsedIndex+1; j++ {
		v := new(big.Rat).SetFrac(pow2big(j), pow10big(decDigits(pow2big(j))-1))
		if worst == nil || v.Cmp(worst) < 0 {
			worst, worstJ = v, j
		}
	}
	if worst.Cmp(need) < 0 {
		t.Fatalf("Lemma D2 fails: min 2^j/10^(D(2^j)-1) = %s at j = %d, below (1+2^-24)^2 = %s",
			worst.FloatString(9), worstJ, need.FloatString(9))
	}

	// --- premise P4: CX.hi < 2^49, so float32(CX.hi)*2^64 < 2^113 cannot
	// overflow binary32 in the digit estimate.
	if SMALL_COEFF_MASK128 != (uint64(1)<<49)-1 {
		t.Fatalf("SMALL_COEFF_MASK128 = %#x, the certificate needs 2^49-1", uint64(SMALL_COEFF_MASK128))
	}
	// --- premise: the BID64 small path can only carry CX < 10^16.
	if new(big.Int).SetUint64(SMALL_COEFF_MASK64).Cmp(pow10big(16)) >= 0 {
		t.Fatalf("SMALL_COEFF_MASK64 = %#x is not below 10^16", uint64(SMALL_COEFF_MASK64))
	}

	// --- premise P5, third clause: the values behind every bid_power10_table_128
	// index the band code can reach. The scale arithmetic under P7 below re-derives
	// which indices are read - [33,38] on the scale <= 38 branch, [2,30] and the
	// literal 37 on the scale > 38 branch - but an index range says nothing about
	// what sits at those indices, and "C256 = CX*10^scale" is a claim about the
	// values. unpack_BID128_value's canonicality test reads entry 34 on top of that.
	// Checked over i = 0..38, the whole table and a superset of all three uses, so
	// a single corrupted entry fails here rather than silently widening the band.
	if len(bid_power10_table_128) <= 38 {
		t.Fatalf("bid_power10_table_128 has %d entries, too few for the indices the band code reaches",
			len(bid_power10_table_128))
	}
	for i := 0; i <= 38; i++ {
		if want := big128(pow10big(i)); bid_power10_table_128[i] != want {
			t.Fatalf("bid_power10_table_128[%d] = %016x%016x, want 10^%d = %016x%016x",
				i, bid_power10_table_128[i].hi, bid_power10_table_128[i].lo, i, want.hi, want.lo)
		}
	}

	// --- premise P7: scale range, no truncation in the scale > 38 branch, and
	// the band itself. digits is the exact digit count D by Lemma D2, and CX
	// ranges over [10^(D-1), 10^D) for that D.
	lo, hi := pow10big(66), pow10big(68)
	for _, entry := range []struct {
		name   string
		maxDig int
	}{{"Bid128Sqrt", 34}, {"Bid128dSqrt", 16}} {
		for d := 1; d <= entry.maxDig; d++ {
			for pp := 0; pp <= 1; pp++ {
				scale := 67 - d + pp
				if scale < 33 || scale > 67 {
					t.Fatalf("%s: D = %d, p = %d gives scale = %d outside [33,67]", entry.name, d, pp, scale)
				}
				if scale > 38 {
					if idx := scale - 37; idx < 2 || idx >= len(bid_power10_table_128) {
						t.Fatalf("%s: scale = %d indexes bid_power10_table_128[%d]", entry.name, scale, idx)
					}
					// __mul_128x128_low truncates above 2^128; the largest
					// intermediate is (10^d - 1) * 10^(scale-37) < 10^(30+p).
					top := new(big.Int).Mul(new(big.Int).Sub(pow10big(d), big.NewInt(1)), pow10big(scale-37))
					if top.Cmp(pow2big(128)) >= 0 {
						t.Fatalf("%s: D = %d, scale = %d: CX1 = %s overflows 128 bits, so __mul_128x128_low truncates",
							entry.name, d, scale, top)
					}
				} else if scale >= len(bid_power10_table_128) {
					t.Fatalf("%s: scale = %d indexes bid_power10_table_128 out of range", entry.name, scale)
				}
				bandLo := new(big.Int).Mul(pow10big(d-1), pow10big(scale))
				bandHi := new(big.Int).Mul(new(big.Int).Sub(pow10big(d), big.NewInt(1)), pow10big(scale))
				if bandLo.Cmp(lo) < 0 || bandHi.Cmp(hi) >= 0 {
					t.Fatalf("%s: D = %d, p = %d gives C256 in [%s, %s], outside [10^66, 10^68)",
						entry.name, d, pp, bandLo, bandHi)
				}
			}
		}
	}

	// --- a corollary of P7's band rather than a premise of its own, so it carries no
	// number here: the plan files it as its P9, but P9 in section 4 of this file is
	// the term-bound rationals, and one number per claim is what makes the premise
	// list readable. With C256 < 10^68 < 2^226 the top two bits of C256.w3 are zero,
	// so C4 = 4*C256 at :306-309 / :549-552 is exact. Sites 16 and 19 do not read C4;
	// it is checked because it is what separates this arm from its nearest-mode
	// sibling, not because the deadness argument rests on it.
	if hi.Cmp(pow2big(226)) >= 0 {
		t.Fatalf("10^68 is not below 2^226; the C4 = 4*C256 shift can truncate")
	}
}

// ratPow2 and ratPow10 build exact rationals for the error chain, so no step of
// TestBid128SqrtEstimatorErrorBoundExcludesUndershoot is evaluated in the same
// arithmetic it is reasoning about.
func ratPow2(n int) *big.Rat {
	if n >= 0 {
		return new(big.Rat).SetInt(pow2big(n))
	}
	return new(big.Rat).SetFrac(big.NewInt(1), pow2big(-n))
}

func ratPow10(n int) *big.Rat { return new(big.Rat).SetInt(pow10big(n)) }

func ratDec(t *testing.T, s string) *big.Rat {
	t.Helper()
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		t.Fatalf("cannot parse rational %q", s)
	}
	return r
}

func ratAtMost(t *testing.T, label string, got, want *big.Rat) {
	t.Helper()
	if got.Cmp(want) > 0 {
		t.Fatalf("%s: bound recomputes to %s, above the certificate's %s",
			label, got.FloatString(24), want.FloatString(24))
	}
}

// ratSqrtBounds brackets sqrt(v) between two exact rationals for a positive
// rational v. Naming the technique because the choice is load-bearing: sqrt(p/q)
// = sqrt(p*q)/q, so a single big.Int.Sqrt - a floor - plus an explicit +1 slack
// at the 2^-g scale gives lo <= sqrt(v) < hi with both ends exact rationals. No
// irrational and no float64 enters the chain, and every use below picks the end
// that turns its inequality into an over-estimate: hi where sqrt(A) appears in a
// numerator, lo where it appears in a denominator. g = 128 makes the bracket at
// most 2^-128 wide - an absolute width, in the units of sqrt(v) rather than a
// fraction of v; the exact width is 1/(q*2^128) for v = p/q - which is twenty
// orders of magnitude below the
// tightest margin any comparison in this test has to survive (|T3| clears its
// pinned bound by 7.7e-8).
func ratSqrtBounds(v *big.Rat) (lo, hi *big.Rat) {
	const g = 128
	scaled := new(big.Int).Mul(v.Num(), v.Denom())
	s := new(big.Int).Sqrt(scaled.Lsh(scaled, 2*g))
	den := new(big.Int).Mul(v.Denom(), pow2big(g))
	return new(big.Rat).SetFrac(s, den), new(big.Rat).SetFrac(new(big.Int).Add(s, big.NewInt(1)), den)
}

// TestBid128SqrtEstimatorErrorBoundExcludesUndershoot recomputes the four term
// bounds of section 3 as exact rationals and checks that the resulting interval
// for sigma - sqrt(A) keeps CS at or above isqrt(A) on the whole reachable band.
// Everything here is arithmetic over the plan's symbolic decomposition; the
// decomposition itself is the hand proof the comment transcribes, and the source
// it describes is pinned by TestBid128SqrtEstimatorSourceIsPinned.
func TestBid128SqrtEstimatorErrorBoundExcludesUndershoot(t *testing.T) {
	one := new(big.Rat).SetInt64(1)
	u := ratPow2(-53)

	// Band endpoint: A < 10^68, so sqrt(A) <= 10^34 exactly. Used by every term
	// whose bound depends on A alone; T3 below needs the joint (A,kappa) treatment
	// instead and does not read this.
	sqrtAMax := ratPow10(34)

	// (E1)+(E2): |e| <= 4u + |d| + 13u^2 with |d| <= 2^(kappa-53)/A + 2^-81, and
	// 2^kappa/A <= 1 + 2^-80 from |v3+p2 - A| <= 2^138 + 2^129 and A > 2^219.2.
	eMax := new(big.Rat).Mul(new(big.Rat).SetInt64(4), u)
	eMax.Add(eMax, new(big.Rat).Mul(new(big.Rat).Add(one, ratPow2(-80)), u))
	eMax.Add(eMax, ratPow2(-81))
	eMax.Add(eMax, new(big.Rat).Mul(new(big.Rat).SetInt64(13), new(big.Rat).Mul(u, u)))
	ratAtMost(t, "|e|", eMax, ratDec(t, "5.6e-16"))

	// P8: ey in {110..113}, hence every shift count in the pinned body lies in
	// (0,64). ly_d is within (1 +- 3u) of 1/sqrt(A) (one rounding in sqrt, one in
	// the divide, plus |d| <= ~u), so it suffices that
	// (1+3u)^2 * 2^218 < 10^66 and (1-3u)^2 * 2^226 >= 10^68.
	up := new(big.Rat).Add(one, new(big.Rat).Mul(new(big.Rat).SetInt64(3), u))
	dn := new(big.Rat).Sub(one, new(big.Rat).Mul(new(big.Rat).SetInt64(3), u))
	if new(big.Rat).Mul(new(big.Rat).Mul(up, up), ratPow2(218)).Cmp(ratPow10(66)) >= 0 {
		t.Fatal("ey can fall below 110; the shift counts at :136-137 are no longer pinned to (0,64)")
	}
	if new(big.Rat).Mul(new(big.Rat).Mul(dn, dn), ratPow2(226)).Cmp(ratPow10(68)) < 0 {
		t.Fatal("ey can exceed 113; the shift counts at :189-190 are no longer pinned to (0,64)")
	}
	for ey := 110; ey <= 113; ey++ {
		for _, sh := range []int{2*ey - 216 + 1, 64 - (2*ey - 216) - 1, 2*ey - 216, 64 - (2*ey - 216), ey - 77, 64 - (ey - 77)} {
			if sh <= 0 || sh >= 64 {
				t.Fatalf("ey = %d yields a shift count of %d, outside (0,64)", ey, sh)
			}
		}
	}

	// T3, the dominant term. |T3| <= 3.5 * 2^-118 * (1+2^-51) * F with
	// F = |e|*2^53*sqrt(A). Writing (E1)+(E2) out and keeping (E1)'s 2^kappa/A
	// slack as an explicit (1+2^-80) factor rather than dropping it:
	//
	//	F(A,kappa) <= 4*sqrt(A) + (1+2^-80)*2^kappa/sqrt(A)
	//	              + 2^-28*sqrt(A) + 13*2^-53*sqrt(A)
	//
	// - 4 sqrt(A) from the 4u of (E2); the second term is u*(2^kappa/A)*2^53*sqrt(A),
	// which is exactly 2^kappa/sqrt(A), so the (1+2^-80) in front of it is pure extra
	// conservatism rather than a step the derivation needs - that same slack is
	// load-bearing further down, where it widens each binade's A-range - and the last
	// two carry (E1)'s additive 2^-81 times 2^53 and (E2)'s 13u^2 explicitly instead
	// of absorbing them into a (1+2^-50) factor as the plan's write-up does.
	//
	// This maximum is JOINT in (A,kappa) and has to be taken that way. Term by term
	// it is not reachable: 2^kappa/sqrt(A) alone peaks at 2^112.5 = 7.343e33 at a
	// binade's left edge, 36% above the 5.392e33 the maximising pair actually
	// produces, and summing per-term maxima gives 4.734e34, which would carry |T3|
	// to 0.4986 and break both the pinned bound and the load-bearing margin. The
	// two leading terms move against each other, so the maximum is taken per binade
	// and then over binades:
	//
	//	F_kappa = F(A_r(kappa), kappa),  A_r(kappa) = the right end of kappa's A-range
	//
	// which is where F sits because F is *increasing* in A inside a binade:
	// dF/dA = (4 + 2^-28 + 13*2^-53)/(2 sqrt(A)) - (1+2^-80)*2^kappa/(2 A^(3/2)) is
	// positive as soon as (1+2^-80)*2^kappa <= 4*A, and A never drops below
	// 2^kappa*(1-2^-80). That inequality is checked per binade below, not asserted.
	//
	// kappa indexes the binade of v3+p2, not of A, and step 1 gives
	// |v3+p2 - A| <= 2^138+2^129 < 2^-80*A on the band. So the admissible kappa are
	// those whose binade meets [10^66*(1-2^-80), 10^68*(1+2^-80)) - derived below,
	// not assumed - and for each of them A ranges over
	// [2^kappa*(1-2^-80), 2^(kappa+1)/(1-2^-80)) intersected with the band - the low
	// end below 2^kappa and the high end above 2^(kappa+1) by that same slack, which
	// is what covers the kappa-slack corner where v3+p2 lands one binade above A's
	// own. 1/(1-2^-80) is used through its over-bound 1+2^-79, checked below.
	slackUp := new(big.Rat).Add(one, ratPow2(-80))
	slackDn := new(big.Rat).Sub(one, ratPow2(-80))
	widen := new(big.Rat).Add(one, ratPow2(-79)) // 1/(1-2^-80) < 1+2^-79
	if new(big.Rat).Mul(slackDn, widen).Cmp(one) < 0 {
		t.Fatal("(1-2^-80)*(1+2^-79) < 1, so 1+2^-79 is not an over-bound of 1/(1-2^-80)")
	}
	var kappas []int
	for k := 0; k <= 512; k++ {
		if ratPow2(k).Cmp(new(big.Rat).Mul(ratPow10(68), slackUp)) < 0 &&
			ratPow2(k+1).Cmp(new(big.Rat).Mul(ratPow10(66), slackDn)) > 0 {
			kappas = append(kappas, k)
		}
	}
	if len(kappas) == 0 || kappas[0] != 219 || kappas[len(kappas)-1] != 225 {
		t.Fatalf("the binades v3+p2 can occupy over [10^66,10^68) came out as %v, not {219..225}", kappas)
	}
	// The sub-leading pair of F, which rides on sqrt(A) exactly like the 4u term.
	fSub := new(big.Rat).Add(ratPow2(-28), new(big.Rat).Mul(new(big.Rat).SetInt64(13), ratPow2(-53)))
	// fOf evaluates the F bound at the right end aR of an A-interval on which F is
	// increasing, so it is the maximum over that interval. The sqrt bracket is used
	// in the direction that over-estimates: its high end multiplies (4+sub), its low
	// end divides 2^kappa.
	fOf := func(aR *big.Rat, kappa int) *big.Rat {
		sLo, sHi := ratSqrtBounds(aR)
		f := new(big.Rat).Mul(new(big.Rat).Add(new(big.Rat).SetInt64(4), fSub), sHi)
		return f.Add(f, new(big.Rat).Quo(new(big.Rat).Mul(slackUp, ratPow2(kappa)), sLo))
	}
	fMax, fMaxKappa, cornerMax := (*big.Rat)(nil), 0, (*big.Rat)(nil)
	for _, k := range kappas {
		aLo := new(big.Rat).Mul(ratPow2(k), slackDn)
		if aLo.Cmp(ratPow10(66)) < 0 {
			aLo = ratPow10(66)
		}
		aHi := new(big.Rat).Mul(ratPow2(k+1), widen)
		if aHi.Cmp(ratPow10(68)) > 0 {
			aHi = ratPow10(68)
		}
		// Monotonicity of F in A over [aLo, aHi], the step that puts the per-binade
		// maximum at aHi. Checked in the stronger form (1+2^-80)*2^kappa <= 4*aLo.
		if new(big.Rat).Mul(slackUp, ratPow2(k)).Cmp(new(big.Rat).Mul(new(big.Rat).SetInt64(4), aLo)) > 0 {
			t.Fatalf("F is not increasing in A on binade kappa = %d: (1+2^-80)*2^kappa exceeds 4*A at the low end %s",
				k, aLo.FloatString(0))
		}
		f := fOf(aHi, k)
		if fMax == nil || f.Cmp(fMax) > 0 {
			fMax, fMaxKappa = f, k
		}
		// The kappa-slack corner, checked rather than argued: A just below 2^kappa,
		// where v3+p2 has rounded up into the next binade and kappa sits one above A's
		// own. That is where the 2^kappa/sqrt(A) term is largest *relative* to
		// sqrt(A), and F degenerates to the (4 + 1 + 2^-80 + 2^-28 + 13*2^-53)*sqrt(A)
		// shape. The corner interval [2^kappa*(1-2^-80), 2^kappa] lies inside
		// [aLo,aHi] and F is increasing on it by the same check, so its maximum is F
		// at A = 2^kappa; where the band's own lower end has clamped aLo above it the
		// interval is empty and this is an over-estimate of nothing. It is evaluated
		// separately because it is the only place the binade index and A's own binade
		// disagree, and its value has to stay under the maximum found here for the
		// monotonicity step to have covered it.
		corner := fOf(ratPow2(k), k)
		if cornerMax == nil || corner.Cmp(cornerMax) > 0 {
			cornerMax = corner
		}
	}
	if fMaxKappa != 225 {
		t.Fatalf("the F maximum moved to binade kappa = %d; section 3's constants are stated for kappa = 225 "+
			"with A capped at 10^68 by the band, not by 2^226", fMaxKappa)
	}
	ratAtMost(t, "F at the kappa-slack corner", cornerMax, fMax)
	t3 := new(big.Rat).Mul(ratDec(t, "3.5"), fMax)
	t3.Mul(t3, new(big.Rat).Add(one, ratPow2(-51)))
	t3.Quo(t3, ratPow2(118))
	ratAtMost(t, "|T3|", t3, ratDec(t, "0.478088"))

	// T2 = A*RS*phi/2^128 with phi in [0,1), so 0 <= T2 <= sqrt(A)(1+2^-51)/2^128.
	t2 := new(big.Rat).Mul(sqrtAMax, new(big.Rat).Add(one, ratPow2(-51)))
	t2.Quo(t2, ratPow2(128))
	ratAtMost(t, "T2", t2, ratDec(t, "2.939e-5"))

	// T4 = -(alpha - g*beta + ES2*beta)/2^(ey-12), |alpha|,|beta| < 1, ey >= 110.
	h := new(big.Rat).Mul(eMax, ratPow2(63))
	// Step 5 bounds the e < 0 edge case at h <= H + 2^-64; +1 is that over-bound rounded
	// up to a whole unit, which is also the slack T3's positive side below spends.
	hMax := new(big.Rat).Add(h, one)
	es2Max := new(big.Rat).Mul(ratDec(t, "1.5"), new(big.Rat).Mul(hMax, hMax))
	gMax := new(big.Rat).Mul(eMax, ratPow2(127))
	t4 := new(big.Rat).Add(one, gMax)
	t4.Add(t4, es2Max)
	t4.Quo(t4, ratPow2(98))
	ratAtMost(t, "|T4|", t4, ratDec(t, "3.1e-7"))

	// T1 = sqrt(A)*(sqrt(1+e)(1-e/2+3e^2/8) - 1), tail bounded by 0.3126|e|^3.
	t1 := new(big.Rat).Mul(eMax, new(big.Rat).Mul(eMax, eMax))
	t1.Mul(t1, ratDec(t, "0.3126"))
	t1.Mul(t1, sqrtAMax)
	ratAtMost(t, "|T1|", t1, ratDec(t, "5.5e-13"))

	// T3's positive side: ES2 - 1.5H^2 <= 3H*2^-64, the 2^-64 coming from step 5's bound
	// |H - h| < 2^-64 on the e < 0 edge case - the only case in which the expression is
	// positive at all. Evaluating it at hMax rather than H covers the 1.5*2^-128 residue
	// section 3 drops from the stated constant, with a whole unit of H to spare.
	t3pos := new(big.Rat).Mul(sqrtAMax, new(big.Rat).Add(one, ratPow2(-51)))
	t3pos.Mul(t3pos, new(big.Rat).Mul(new(big.Rat).SetInt64(3), hMax))
	t3pos.Quo(t3pos, ratPow2(192))
	ratAtMost(t, "T3 (positive side)", t3pos, ratDec(t, "2.5e-20"))

	// The interval for sigma - sqrt(A).
	lower := new(big.Rat).Neg(new(big.Rat).Add(t3, new(big.Rat).Add(t4, t1)))
	upper := new(big.Rat).Add(t2, new(big.Rat).Add(t4, new(big.Rat).Add(t1, t3pos)))
	if lower.Cmp(new(big.Rat).Neg(ratDec(t, "0.478089"))) < 0 {
		t.Fatalf("sigma - sqrt(A) can reach %s, below the certificate's -0.478089", lower.FloatString(24))
	}
	ratAtMost(t, "sigma - sqrt(A), upper end", upper, ratDec(t, "2.970e-5"))

	// (E3): sigma >= sqrt(A) + lower >= M + lower > M - 1/2, so CS >= M.
	half := new(big.Rat).SetFrac64(1, 2)
	margin := new(big.Rat).Sub(half, new(big.Rat).Neg(lower))
	if margin.Sign() <= 0 {
		t.Fatalf("the undershoot bound has no margin left: |lower| = %s vs 1/2", new(big.Rat).Neg(lower).FloatString(24))
	}
	// Section 3 states the margin as 0.021911; the floor checked here is the rounded-
	// down 0.0219, and the message quotes that floor rather than the stated value.
	const marginFloor = "0.0219"
	if margin.Cmp(ratDec(t, marginFloor)) < 0 {
		t.Fatalf("the load-bearing margin shrank to %s, below the certificate's floor of %s",
			margin.FloatString(9), marginFloor)
	}
	// (E4): sigma <= sqrt(A) + upper < (M+1) + upper < M + 3/2, so CS <= M+1.
	if upper.Cmp(half) >= 0 {
		t.Fatalf("the overshoot bound %s no longer keeps sigma below M + 3/2", upper.FloatString(24))
	}
}

// TestBid128SqrtDirectedUndershootArmsAreUnreachable certifies audit sites 16
// (bid128_sqrt.go:423, CS.lo++ in Bid128Sqrt) and 19 (bid128_sqrt.go:661,
// CS.hi++ in Bid128dSqrt): both sit behind the directed-rounding guard
// (CS+1)^2 <= C256, which Theorem S makes false on every reachable input.
//
// Sites 17 (:571) and 18 (:646) are deliberately outside this certificate and are
// not claimed by it. They are dead on separate evidence - the exhaustive sweep
// recorded in the closure note above pinnedSqrtCSWrites - which this test does not
// re-run. Both lines are pinned exactly here, like every other write, so a token
// change at either fails this file structurally rather than as a reachability
// verdict.
func TestBid128SqrtDirectedUndershootArmsAreUnreachable(t *testing.T) {
	p := loadPortFile(t, "bid128_sqrt.go")

	// --- residual for every write world and no-write scan below: they are AST scans for assignments, inc/dec, ValueSpec
	// initialisers and address-of, and a pointer-receiver method call is none of those shapes. `CS.m()` on a
	// `func (c *BID_UINT128) m()` takes &CS implicitly and can rewrite CS between the estimator call and the guard with
	// no statement here for either scan to record. Both censuses run in this certificate rather than being inherited
	// from certificate 5, so this one fails on its own: CS is a BID_UINT128 and C256 is a BID_UINT256.
	files := portPkg(t)
	pinTypeIsMethodFree(t, files, "BID_UINT128", pinnedUint128Shape,
		"a call `CS.<that method>()` on a pointer receiver rewrites CS with no assignment, inc/dec or address-of for "+
			"the pinned write world and the no-write scans below to see")
	pinTypeIsMethodFree(t, files, "BID_UINT256", pinnedUint256Shape,
		"a call `C256.<that method>()` on a pointer receiver rewrites C256 with no assignment, inc/dec or address-of "+
			"for the no-write scans below to see, so the guard stops reading the A the estimator was handed")

	for _, c := range []struct {
		fn        string
		siteLabel string
	}{
		{"Bid128Sqrt", "site16 (bid128_sqrt.go:423)"},
		{"Bid128dSqrt", "site19 (bid128_sqrt.go:661)"},
	} {
		fn := p.funcDecl(t, c.fn)

		// --- anchor: the closed world of writes to CS, so no correction outside
		// the pinned rounding paths can reach the guard's operand.
		csWrites := p.pinWrites(t, fn, "CS", pinnedSqrtCSWrites)
		p.pinNoNestedDecl(t, fn, "CS")
		p.pinNoGoto(t, fn)

		// --- boundary: single-pass execution of the whole caller. Theorem S bounds
		// the *raw* estimator return, and the deadness argument needs the guard to
		// evaluate (CS+1)^2 on exactly that value. The closed CS-write world above
		// says which statements may write CS and the dominating chain below says
		// under which conditions each of them runs; it takes no-goto plus no-loop to
		// turn that pair into the straight-line dataflow fact the argument actually
		// uses, that the estimator call is the last write to CS the guard can see.
		// The scan is the whole function rather than the site's dominators because
		// guardPath already renders a loop that lexically encloses the certified
		// increment (as a "for" conjunct) and fails closed on a range: what it cannot
		// see is a loop that re-enters the estimator call and the guard-feeding
		// cascade *without* enclosing the increment, which leaves every other pin in
		// this test passing while the guard reads a CS from an earlier pass.
		p.pinNoLoops(t, fn)

		call := p.stmtWithText(t, fn, "CS = bid_long_sqrt128(C256)")

		// --- anchor: the mode split, and the fact that the certified arm lives in
		// its else. The nearest path (rnd_mode & 3 == 0) has its own round-up at
		// :328 / :569, which is a different mutation site and is not certified here.
		modeSplit := p.ifWithCond(t, fn, "(rnd_mode & 3) == 0")
		if p.line(modeSplit) <= p.line(call) {
			t.Fatalf("bid128_sqrt.go: %s: the mode split (line %d) does not follow the estimator call (line %d)",
				c.fn, p.line(modeSplit), p.line(call))
		}
		p.noWriteBetween(t, fn, "CS", call, modeSplit)
		p.noWriteBetween(t, fn, "C256", call, modeSplit)

		directed := p.elseBlock(t, c.siteLabel+": directed-rounding arm", modeSplit)
		p.pinStmtCount(t, c.siteLabel+": directed-rounding arm", directed, 5)
		p.pinRun(t, c.siteLabel+": M256 = CS^2 and C8 = 2*CS", directed, 0, []string{
			"M256 = __sqr128_to_256(CS)",
			"C8.w1 = (CS.hi << 1) | (CS.lo >> 63)",
			"C8.w0 = CS.lo << 1",
		})
		p.pinNode(t, c.siteLabel+": BID_ROUNDING_UP tail", p.stmtAt(t, "directed arm", directed, 4),
			"if rnd_mode == BID_ROUNDING_UP { CS.lo++ if CS.lo == 0 { CS.hi++ } }")

		// --- anchor: the CS^2 > C256 test. The certified arm is its else, i.e.
		// the branch on which CS^2 <= C256 already holds.
		overshoot := p.asIf(t, c.siteLabel+": CS^2 > C256 test", p.stmtAt(t, "directed arm", directed, 3))
		p.pinNode(t, c.siteLabel+": CS^2 > C256 condition", overshoot.Cond, pinnedSqrtOvershootCond)
		p.pinComparisonOps(t, c.siteLabel+": CS^2 > C256 condition", overshoot.Cond, []token.Token{
			token.GTR, token.EQL, token.GTR, token.EQL, token.GTR, token.EQL, token.GTR,
		})
		p.noWriteBetween(t, directed, "M256", p.stmtAt(t, "directed arm", directed, 0), overshoot)
		p.noWriteBetween(t, directed, "CS", p.stmtAt(t, "directed arm", directed, 0), overshoot)

		// --- anchor: the M256 <- CS^2 + 2*CS + 1 cascade and the guard it feeds.
		under := p.elseBlock(t, c.siteLabel+": undershoot arm", overshoot)
		p.pinStmtCount(t, c.siteLabel+": undershoot arm", under, len(pinnedSqrtPlusOneBuild)+2)
		p.pinRun(t, c.siteLabel+": (CS+1)^2 build", under, 0, pinnedSqrtPlusOneBuild)
		cascade := p.stmtAt(t, "undershoot arm", under, len(pinnedSqrtPlusOneBuild))
		p.pinNode(t, c.siteLabel+": (CS+1)^2 carry cascade", cascade, pinnedSqrtPlusOneCascade)
		guard := p.asIf(t, c.siteLabel+": (CS+1)^2 <= C256 guard",
			p.stmtAt(t, "undershoot arm", under, len(pinnedSqrtPlusOneBuild)+1))
		p.pinNode(t, c.siteLabel+": (CS+1)^2 <= C256 guard", guard.Cond, pinnedSqrtUndershootCond)
		p.pinComparisonOps(t, c.siteLabel+": (CS+1)^2 <= C256 guard", guard.Cond, []token.Token{
			token.LSS, token.EQL, token.LSS, token.EQL, token.LSS, token.EQL, token.LEQ,
		})
		// Nothing may change CS or C256 while (CS+1)^2 is being built, and nothing
		// may change M256 between the last cascade statement and the comparison.
		p.noWriteBetween(t, under, "CS", p.stmtAt(t, "undershoot arm", under, 0), guard)
		p.noWriteBetween(t, under, "C256", p.stmtAt(t, "undershoot arm", under, 0), guard)
		p.noWriteBetween(t, under, "M256", cascade, guard)

		// --- the dead region itself.
		p.pinStmtCount(t, c.siteLabel+": dead region", guard.Body, 2)
		p.pinIncrement(t, c.siteLabel+": low-word round up", p.stmtAt(t, "dead region", guard.Body, 0), "CS.lo")
		carry := p.asIf(t, c.siteLabel+": carry into the high word", p.stmtAt(t, "dead region", guard.Body, 1))
		p.pinNode(t, c.siteLabel+": carry condition", carry.Cond, "CS.lo == 0")
		p.pinStmtCount(t, c.siteLabel+": carry body", carry.Body, 1)
		p.pinIncrement(t, c.siteLabel+": high-word carry", p.stmtAt(t, "carry body", carry.Body, 0), "CS.hi")

		// --- the certified mutation site, located inside the closed write world
		// so that the two increments cannot drift apart from the pinned list.
		site := carry.Body.List[0]
		if c.fn == "Bid128Sqrt" {
			site = guard.Body.List[0]
		}
		found := false
		for _, w := range csWrites {
			if w.pos == site.Pos() {
				found = true
			}
		}
		if !found {
			t.Fatalf("bid128_sqrt.go: %s: the certified increment at line %d is not one of the pinned writes to CS",
				c.fn, p.line(site))
		}

		// --- the guard's full dominating chain, rendered from the AST rather than
		// restated: it is exactly the firing predicate of section 1.
		wantGuards := "!((rnd_mode & 3) == 0) && !(" + pinnedSqrtOvershootCond + ") && " + pinnedSqrtUndershootCond
		if c.fn == "Bid128dSqrt" {
			wantGuards += " && CS.lo == 0"
		}
		if got := p.guardPath(t, fn, site); got != wantGuards {
			t.Fatalf("bid128_sqrt.go: %s: the dominating chain of the certified increment drifted\n got: %s\nwant: %s",
				c.fn, got, wantGuards)
		}
	}

	// --- the arithmetic step the certificate closes with. Theorem S gives CS in
	// {M, M+1} with M = isqrt(A), and for either value (CS+1)^2 >= (M+1)^2 > A, so
	// the pinned guard is false. Checked at the band's boundaries rather than
	// sampled: for each endpoint the residue window [M^2, M^2+2M] on which isqrt is
	// M is taken at its extremes, the last being the largest A that guard could ever
	// be evaluated at for that M.
	one := big.NewInt(1)
	for _, endpoint := range []*big.Int{pow10big(66), pow10big(67), new(big.Int).Sub(pow10big(68), one)} {
		m := new(big.Int).Sqrt(endpoint)
		msq := new(big.Int).Mul(m, m)
		for _, off := range []*big.Int{new(big.Int), m, new(big.Int).Lsh(m, 1)} {
			a := new(big.Int).Add(msq, off)
			if got := new(big.Int).Sqrt(a); got.Cmp(m) != 0 {
				t.Fatalf("isqrt(%s) = %s, expected the window of %s", a, got, m)
			}
			for d := 1; d <= 2; d++ { // CS = M and CS = M+1, so CS+1 = M+d
				next := new(big.Int).Add(m, big.NewInt(int64(d)))
				if next.Mul(next, next); next.Cmp(a) <= 0 {
					t.Fatalf("(CS+1)^2 <= A for A = %s with CS = isqrt(A)+%d; Theorem S would not exclude the arm",
						a, d-1)
				}
			}
		}
	}
}
