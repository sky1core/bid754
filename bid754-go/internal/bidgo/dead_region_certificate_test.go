// Hand-written unreachability certificates for ten mutation sites of the
// mechanical port. It lives outside every generation path and must stay
// hand-written.
//
// Why this gate exists. The 2026-07 mutation audit left ten mutants that no
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
// to pin it separately: TestBid128RoundBoundaryArmsAreUnreachable carries the package-wide receiver census that keeps
// BID_UINT128 - the type of the C3 pointer it censuses - method-free, and the field-select admission here rests on it.
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
	// a *BID_UINT128, so such a method value can exist only if BID_UINT128 has a method at all - and this census fails the
	// moment one is declared in any file of the package this certificate reads, in either receiver form. That is what
	// makes the field-select admission below closed rather than fail-open; the direct-call check alone does not.
	//
	// Two ways a method reaches *C3 without any receiver spelling BID_UINT128, so a match on the written receiver name
	// alone would be fail-open. First, promotion: if BID_UINT128 embedded a type, that type's methods would be
	// BID_UINT128's methods with no receiver naming BID_UINT128 anywhere. Pin the struct's exact shape, so it has two
	// explicitly named scalar fields and no embedded field to promote from. Second, a package-level type alias:
	// `type A = BID_UINT128` makes `func (a A) m()` a method on BID_UINT128 itself, so the name match has to run on the
	// alias-resolved receiver root rather than the written one.
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
	u128, ok := typeSpecs["BID_UINT128"]
	if !ok {
		t.Fatal("the package declares no package-level BID_UINT128, so the C3 no-method pin has no type to close")
	}
	if u128.spec.Assign.IsValid() || u128.spec.TypeParams != nil {
		t.Fatalf("%s: BID_UINT128 is not a plain non-generic defined type, so its method set cannot be read off its own "+
			"declaration: %s", u128.file.name, u128.file.flat(t, u128.spec))
	}
	// The exact rendering is the pin: it admits exactly two fields, both with explicit names, so no embedded field - and
	// so no promoted method reachable as `C3.m` - can exist. Adding, embedding or retyping a field fails here.
	if got, want := u128.file.flat(t, u128.spec.Type), "struct { lo uint64 hi uint64 }"; got != want {
		t.Fatalf("%s: BID_UINT128 is no longer the pinned embedded-field-free shape, so promoted methods on it are no "+
			"longer ruled out\n got: %s\nwant: %s", u128.file.name, got, want)
	}
	// resolveTypeName walks the alias graph from a written type name to the type it finally names. `type A = B` is
	// followed; `type A B` is a distinct defined type whose methods are not BID_UINT128's, so it ends the walk, as does a
	// name the package does not declare. Anything the walk cannot decide - a cycle, a generic alias, an alias whose
	// right-hand side has no root identifier this file can name - fails rather than returning a name that would pass the
	// BID_UINT128 check below.
	resolveTypeName := func(what, name string) string {
		seen := map[string]bool{}
		for {
			ts, ok := typeSpecs[name]
			if !ok || !ts.spec.Assign.IsValid() {
				return name
			}
			if seen[name] {
				t.Fatalf("%s: the alias chain from %s revisits %s, so it cannot be resolved and BID_UINT128 cannot be "+
					"ruled out", ts.file.name, what, name)
			}
			seen[name] = true
			if ts.spec.TypeParams != nil {
				t.Fatalf("%s: %s, reached from %s, is a generic alias, so the chain resolves to no single type and "+
					"BID_UINT128 cannot be ruled out", ts.file.name, name, what)
			}
			next := rootIdentName(ts.spec.Type)
			if next == "" {
				t.Fatalf("%s: alias %s, reached from %s, has the unrecognised right-hand side %s, so the chain cannot "+
					"rule out BID_UINT128", ts.file.name, name, what, ts.file.flat(t, ts.spec.Type))
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
			// anything else unexpected - fails here instead of being skipped as "not BID_UINT128".
			if len(fd.Recv.List) != 1 {
				t.Fatalf("%s: %s has a %d-field receiver list, so the no-method pin cannot read its receiver type",
					p.name, fd.Name.Name, len(fd.Recv.List))
			}
			recv := fd.Recv.List[0].Type
			// rootIdentName peels *, (), [] and selectors, so both `BID_UINT128` and `*BID_UINT128` land on the same
			// root name, and an unpeelable shape returns "" and fails just above the check rather than below it.
			root := rootIdentName(recv)
			if root == "" {
				t.Fatalf("%s: %s has the unrecognised receiver type %s, so the no-method pin cannot rule out BID_UINT128",
					p.name, fd.Name.Name, p.flat(t, recv))
			}
			// Resolve through the alias graph before matching: `type A = BID_UINT128` puts a method on BID_UINT128 under
			// a receiver that spells A. Value, pointer and parenthesised receivers all peel to the same root above.
			if root = resolveTypeName("the receiver of "+fd.Name.Name, root); root == "BID_UINT128" {
				t.Fatalf("%s: %s is a method with receiver %s, so `C3.%s` is a method value the by-context census admits "+
					"as a field select and the C3 aliasing closure below stops being closed",
					p.name, fd.Name.Name, p.flat(t, recv), fd.Name.Name)
			}
		}
	}

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
