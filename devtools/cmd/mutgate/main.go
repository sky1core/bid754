// Command mutgate is a mutation-testing audit harness for the bidgo Go
// mechanical port (bid754-go/internal/bidgo).
//
// It measures how well the existing generated verification gates (portable
// goport readtest / goport decTest / public-API parity, plus the native FFI
// bit-compare and decNumber third-oracle differential stages) detect real
// code mutations in the port. It is an offline audit tool, not a CI gate:
// a full run compiles and executes hundreds of mutants and takes tens of
// minutes.
//
// Isolation contract: mutations are applied ONLY inside a detached git
// worktree that this tool creates (or is pointed at). The primary checkout
// is never written. Every mutated file is restored from its pristine bytes
// after each mutant, and the tool re-verifies the worktree is clean at exit.
//
// Mutation categories (single-edit, AST-located, text-applied):
//
//	aor     arithmetic operator swap: + <-> -, * -> +, / -> *, % -> /,
//	        also on op= assignment forms and ++ <-> --
//	bit     bitwise/shift swap: & <-> |, ^ -> &, &^ -> &, << <-> >>,
//	        also on op= assignment forms
//	cmp     comparison swap: < <-> <=, > <-> >=, == <-> !=
//	const   integer literal value +1 / -1, and unary-minus sign drop on
//	        integer literals
//	negcond if-condition inversion: `if c` -> `if !(c)`
//	branch  unlabeled break <-> continue
//
// Deliberately excluded: && <-> || swaps and `for` condition inversion
// (both mostly manufacture hangs/no-ops instead of semantic probes), labeled
// branch statements, and string/float literals.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

const bidgoRel = "bid754-go/internal/bidgo"

// arithCorePreset is the deterministic "arithmetic core" sample target:
// the bid128/bid64/bid32 add/mul/div/sqrt/fma/quantize/round-integral family
// of the mechanical port. All files in this set carry mechanical-port
// provenance (Intel BID C); no bid754-authored helper files are included.
var arithCorePreset = []string{
	"bid128_add.go", "bid128_mul.go", "bid128_div.go", "bid128_sqrt.go",
	"bid128_fma.go", "bid128_fma_body.go", "bid128_fma_helpers.go",
	"bid128_quantize.go", "bid128_round.go", "bid128_round_integral.go",
	"bid128_nearbyint.go",
	"add64.go", "add64_inline.go", "add128_inline.go", "mul64.go",
	"div64.go", "sqrt64.go", "fma64.go", "quantize64.go",
	"round_integral64.go", "inline_round64.go",
	"bid32_add.go", "bid32_mul.go", "bid32_div.go", "bid32_sqrt.go",
	"bid32_fma.go", "bid32_quantize.go", "bid32_round_integral.go",
}

// stage describes one kill-suite stage, executed in order; the first failing
// stage kills the mutant and later stages are skipped.
type stage struct {
	Name    string
	Binary  string // key into the lazily built per-mutant test binaries
	RunExpr string
}

var stageCatalog = map[string]stage{
	"readtest": {Name: "readtest", Binary: "portable", RunExpr: "^TestGeneratedReadCasesGoPort$"},
	"dectest":  {Name: "dectest", Binary: "portable", RunExpr: "^TestGeneratedDectestSuitesGoPort$"},
	"parity": {Name: "parity", Binary: "portable",
		RunExpr: "^(TestGeneratedPublicAPIParity|TestGeneratedPublicAPIFlaglessSiblingEquivalence)$"},
	"native": {Name: "native", Binary: "native", RunExpr: "^(TestGeneratedReadCases|TestGeneratedDectestSuites|TestGeneratedFFIBitCompareSubset)$"},
	"decnumber": {Name: "decnumber", Binary: "decnumber",
		RunExpr: "^TestGeneratedDecnumberDifferential(CorpusContract|RoutingSentinels|Structured|DeterministicRandom)$"},
	// Secondary-analysis stage (not part of the regular audited gate chain):
	// the Tier 1 arithmetic long differential, structured leg only, used to
	// probe whether survivors of the regular chain fall to the long corpus.
	"tier1arith": {Name: "tier1arith", Binary: "tier1long",
		RunExpr: "^TestTier1ArithmeticStructuredNativeDifferential$"},
	// Secondary-analysis stage: the hand-written in-package bidgo tests
	// (flagless-variant equivalence, targeted regression tests). Part of the
	// repo's test-go-modules gate but outside the generated verification
	// domains, so it is measured as a separate increment.
	"bidgopkg": {Name: "bidgopkg", Binary: "bidgopkg", RunExpr: "."},
}

var stageOrderDefault = []string{"readtest", "dectest", "parity"}

type mutationSite struct {
	File     string `json:"file"`     // bidgo-relative file name
	Category string `json:"category"` // aor|bit|cmp|const|negcond|branch
	Offset   int    `json:"offset"`   // byte offset of the replaced span start
	End      int    `json:"end"`      // byte offset of the replaced span end
	Line     int    `json:"line"`
	Col      int    `json:"col"`
	Orig     string `json:"orig"`    // original span text (condition text truncated)
	Mutated  string `json:"mutated"` // replacement text (truncated for negcond)
	Repl     string `json:"-"`       // full replacement bytes
	OrigFull string `json:"-"`       // full original span bytes
	Func     string `json:"func"`    // enclosing function name, if any
	Variant  string `json:"variant"` // e.g. "+1", "-1", "sign-drop", "+->-"
}

func (s mutationSite) ID() string {
	return fmt.Sprintf("%s:%d:%s:%s", s.File, s.Offset, s.Category, s.Variant)
}

type mutantResult struct {
	ID        string           `json:"id"`
	Site      mutationSite     `json:"site"`
	Status    string           `json:"status"` // killed|survived|invalid
	KilledBy  string           `json:"killed_by,omitempty"`
	Reason    string           `json:"reason,omitempty"` // fail|timeout|compile_error|native_build_error
	FailLines []string         `json:"fail_lines,omitempty"`
	StageMS   map[string]int64 `json:"stage_ms,omitempty"`
	BuildMS   map[string]int64 `json:"build_ms,omitempty"`
	Passed    []string         `json:"passed_stages,omitempty"`
	Note      string           `json:"note,omitempty"`
}

type config struct {
	mode         string
	repo         string
	worktree     string
	commit       string
	files        string
	perFile      int
	seed         int64
	stages       string
	stageTimeout time.Duration
	buildTimeout time.Duration
	jsonlPath    string
	maxMutants   int
	logEvery     int
	strata       string
	failfast     bool
	idsFile      string
}

func main() {
	var cfg config
	flag.StringVar(&cfg.mode, "mode", "run", "run|list|selfcheck|setup|teardown")
	flag.StringVar(&cfg.repo, "repo", "", "primary repo root (default: git toplevel of cwd)")
	flag.StringVar(&cfg.worktree, "worktree", "", "isolated worktree path (required; created by setup/run if absent)")
	flag.StringVar(&cfg.commit, "commit", "HEAD", "commit for worktree creation")
	flag.StringVar(&cfg.files, "files", "arith-core", "'arith-core' preset or CSV of bidgo file names")
	flag.IntVar(&cfg.perFile, "per-file", 20, "sampled mutants per file")
	flag.Int64Var(&cfg.seed, "seed", 754, "deterministic sampling seed")
	flag.StringVar(&cfg.stages, "stages", strings.Join(stageOrderDefault, ","), "ordered kill-suite stages (readtest,dectest,parity,native,decnumber)")
	flag.DurationVar(&cfg.stageTimeout, "stage-timeout", 300*time.Second, "per-stage execution timeout")
	flag.DurationVar(&cfg.buildTimeout, "build-timeout", 300*time.Second, "per-build timeout")
	flag.StringVar(&cfg.jsonlPath, "jsonl", "", "JSONL result output path (required for run/selfcheck)")
	flag.IntVar(&cfg.maxMutants, "max", 0, "cap on total sampled mutants (0 = no cap)")
	flag.IntVar(&cfg.logEvery, "log-every", 10, "progress log interval")
	flag.StringVar(&cfg.strata, "strata", "aor=5,cmp=4,const=4,negcond=3,bit=3,branch=1",
		"per-file per-category quotas; leftover slots backfill uniformly ('' = pure uniform)")
	flag.BoolVar(&cfg.failfast, "failfast", true, "pass -test.failfast to stage runs (kill verdict unchanged, much faster)")
	flag.StringVar(&cfg.idsFile, "ids-file", "", "run mode: newline-separated mutant IDs to evaluate instead of sampling")
	flag.Parse()

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "mutgate: %v\n", err)
		os.Exit(1)
	}
}

func run(cfg config) error {
	var err error
	if cfg.repo == "" {
		cfg.repo, err = gitTopLevel(".")
		if err != nil {
			return fmt.Errorf("resolve repo root: %w", err)
		}
	}
	cfg.repo, err = filepath.Abs(cfg.repo)
	if err != nil {
		return err
	}
	switch cfg.mode {
	case "setup":
		return setupWorktree(cfg)
	case "teardown":
		return teardownWorktree(cfg)
	case "list":
		sites, files, err := enumerateAll(cfg, cfg.repo)
		if err != nil {
			return err
		}
		return printSiteInventory(sites, files)
	case "run":
		return runMutants(cfg)
	case "selfcheck":
		return runSelfCheck(cfg)
	default:
		return fmt.Errorf("unknown -mode %q", cfg.mode)
	}
}

// ---------- worktree management ----------

func gitTopLevel(dir string) (string, error) {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// assertWorktreeIsolated enforces the isolation contract structurally: the
// mutation worktree must not be the primary checkout itself nor any path
// inside it. Paths are absolutized and symlink-resolved (where they exist)
// before comparison, so an aliased path cannot slip through.
func assertWorktreeIsolated(repo, worktree string) error {
	if worktree == "" {
		return errors.New("-worktree is required")
	}
	canon := func(p string) (string, error) {
		abs, err := filepath.Abs(p)
		if err != nil {
			return "", err
		}
		if resolved, err := filepath.EvalSymlinks(abs); err == nil {
			return resolved, nil
		}
		// Path does not exist yet (fresh worktree target): resolve the
		// closest existing ancestor so a symlinked parent cannot alias into
		// the primary checkout.
		dir, base := filepath.Dir(abs), filepath.Base(abs)
		if resolved, err := filepath.EvalSymlinks(dir); err == nil {
			return filepath.Join(resolved, base), nil
		}
		return abs, nil
	}
	repoC, err := canon(repo)
	if err != nil {
		return fmt.Errorf("resolve repo path: %w", err)
	}
	wtC, err := canon(worktree)
	if err != nil {
		return fmt.Errorf("resolve worktree path: %w", err)
	}
	rel, err := filepath.Rel(repoC, wtC)
	if err == nil && (rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))) {
		return fmt.Errorf("refusing worktree %q: it is the primary checkout %q or a path inside it; mutations may only run in a separate detached worktree", worktree, repo)
	}
	return nil
}

func setupWorktree(cfg config) error {
	if err := assertWorktreeIsolated(cfg.repo, cfg.worktree); err != nil {
		return err
	}
	if _, err := os.Stat(cfg.worktree); err == nil {
		// Reusing a worktree left behind by an interrupted or crashed run
		// would evaluate every mutant on top of a leftover mutation, so a
		// reused worktree must prove it is pristine before anything runs.
		// The tracked tree must be clean; the gitignored setup inputs
		// (decTest copies, Intel DFP symlinks) do not appear in porcelain
		// output and stay permitted.
		dirty, err := gitStatusPorcelain(cfg.worktree)
		if err != nil {
			return fmt.Errorf("verify reused worktree %s is clean: %w", cfg.worktree, err)
		}
		if dirty != "" {
			return fmt.Errorf("refusing to reuse dirty worktree %s (leftover state would poison every mutant verdict):\n%s", cfg.worktree, dirty)
		}
		fmt.Printf("worktree already present (verified clean): %s\n", cfg.worktree)
	} else {
		cmd := exec.Command("git", "-C", cfg.repo, "worktree", "add", "--detach", cfg.worktree, cfg.commit)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git worktree add: %v\n%s", err, out)
		}
		fmt.Printf("worktree created: %s @ %s\n", cfg.worktree, cfg.commit)
	}
	// Copy the pinned (gitignored) decTest inputs so the goport decTest gate
	// executes instead of taking its input-absence skip.
	srcDir := filepath.Join(cfg.repo, "devtools", "tests")
	dstDir := filepath.Join(cfg.worktree, "devtools", "tests")
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("read decTest inputs: %w", err)
	}
	copied := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".decTest") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(srcDir, e.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dstDir, e.Name()), b, 0o644); err != nil {
			return err
		}
		copied++
	}
	fmt.Printf("decTest inputs copied: %d\n", copied)

	// Link the pinned Intel DFP build artifacts (gitignored, absent from a
	// fresh worktree) from the primary checkout so the ${SRCDIR}-relative
	// cgo directives of the native gates resolve. Read-only inputs: mutants
	// never write under third_party.
	intelSrc := filepath.Join(cfg.repo, "devtools", "third_party", "intel_dfp")
	intelDst := filepath.Join(cfg.worktree, "devtools", "third_party", "intel_dfp")
	for _, name := range []string{"src", "include", "lib", "LIBRARY"} {
		target := filepath.Join(intelSrc, name)
		if _, err := os.Stat(target); err != nil {
			continue // native prerequisites not installed; portable stages unaffected
		}
		link := filepath.Join(intelDst, name)
		if _, err := os.Lstat(link); err == nil {
			continue
		}
		if err := os.Symlink(target, link); err != nil {
			return fmt.Errorf("symlink %s: %w", link, err)
		}
	}
	return nil
}

func teardownWorktree(cfg config) error {
	if err := assertWorktreeIsolated(cfg.repo, cfg.worktree); err != nil {
		return err
	}
	cmd := exec.Command("git", "-C", cfg.repo, "worktree", "remove", "--force", cfg.worktree)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree remove: %v\n%s", err, out)
	}
	fmt.Printf("worktree removed: %s\n", cfg.worktree)
	return nil
}

// gitStatusPorcelain returns the porcelain status of a worktree, optionally
// restricted to pathspecs. Empty output means clean (gitignored files do not
// appear).
func gitStatusPorcelain(worktree string, pathspec ...string) (string, error) {
	args := []string{"-C", worktree, "status", "--porcelain"}
	if len(pathspec) > 0 {
		args = append(args, "--")
		args = append(args, pathspec...)
	}
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func worktreeDirty(worktree string) (string, error) {
	return gitStatusPorcelain(worktree, "bid754-go")
}

// ---------- mutation site enumeration ----------

var binarySwaps = map[token.Token]struct {
	repl string
	cat  string
}{
	token.ADD:     {"-", "aor"},
	token.SUB:     {"+", "aor"},
	token.MUL:     {"+", "aor"},
	token.QUO:     {"*", "aor"},
	token.REM:     {"/", "aor"},
	token.AND:     {"|", "bit"},
	token.OR:      {"&", "bit"},
	token.XOR:     {"&", "bit"},
	token.AND_NOT: {"&", "bit"},
	token.SHL:     {">>", "bit"},
	token.SHR:     {"<<", "bit"},
	token.LSS:     {"<=", "cmp"},
	token.LEQ:     {"<", "cmp"},
	token.GTR:     {">=", "cmp"},
	token.GEQ:     {">", "cmp"},
	token.EQL:     {"!=", "cmp"},
	token.NEQ:     {"==", "cmp"},
}

var assignSwaps = map[token.Token]struct {
	repl string
	cat  string
}{
	token.ADD_ASSIGN: {"-=", "aor"},
	token.SUB_ASSIGN: {"+=", "aor"},
	token.MUL_ASSIGN: {"+=", "aor"},
	token.QUO_ASSIGN: {"*=", "aor"},
	token.REM_ASSIGN: {"/=", "aor"},
	token.AND_ASSIGN: {"|=", "bit"},
	token.OR_ASSIGN:  {"&=", "bit"},
	token.XOR_ASSIGN: {"&=", "bit"},
	token.SHL_ASSIGN: {">>=", "bit"},
	token.SHR_ASSIGN: {"<<=", "bit"},
}

func enumerateFile(root, relFile string) ([]mutationSite, []byte, error) {
	path := filepath.Join(root, bidgoRel, relFile)
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.SkipObjectResolution)
	if err != nil {
		return nil, nil, fmt.Errorf("parse %s: %w", relFile, err)
	}
	offsetOf := func(p token.Pos) int { return fset.Position(p).Offset }
	var sites []mutationSite

	// Track enclosing function names for locator-based self-checks.
	type funcSpan struct {
		name       string
		start, end int
	}
	var funcs []funcSpan
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Body != nil {
			funcs = append(funcs, funcSpan{fd.Name.Name, offsetOf(fd.Pos()), offsetOf(fd.End())})
		}
	}
	enclosing := func(off int) string {
		for _, fs := range funcs {
			if off >= fs.start && off < fs.end {
				return fs.name
			}
		}
		return ""
	}
	add := func(start, end int, repl, cat, variant string) {
		orig := string(src[start:end])
		s := mutationSite{
			File: relFile, Category: cat, Offset: start, End: end,
			Line: fset.Position(fset.File(f.Pos()).Pos(start)).Line,
			Col:  fset.Position(fset.File(f.Pos()).Pos(start)).Column,
			Orig: truncate(orig, 60), Mutated: truncate(repl, 66),
			Repl: repl, OrigFull: orig,
			Func: enclosing(start), Variant: variant,
		}
		sites = append(sites, s)
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.BinaryExpr:
			if sw, ok := binarySwaps[x.Op]; ok {
				start := offsetOf(x.OpPos)
				add(start, start+len(x.Op.String()), sw.repl, sw.cat, x.Op.String()+"->"+sw.repl)
			}
		case *ast.AssignStmt:
			if sw, ok := assignSwaps[x.Tok]; ok {
				start := offsetOf(x.TokPos)
				add(start, start+len(x.Tok.String()), sw.repl, sw.cat, x.Tok.String()+"->"+sw.repl)
			}
		case *ast.IncDecStmt:
			start := offsetOf(x.TokPos)
			if x.Tok == token.INC {
				add(start, start+2, "--", "aor", "inc->dec")
			} else {
				add(start, start+2, "++", "aor", "dec->inc")
			}
		case *ast.BasicLit:
			if x.Kind == token.INT {
				start, end := offsetOf(x.Pos()), offsetOf(x.End())
				for _, delta := range []int64{1, -1} {
					if repl, ok := intLitDelta(x.Value, delta); ok {
						variant := "+1"
						if delta < 0 {
							variant = "-1"
						}
						add(start, end, repl, "const", variant)
					}
				}
			}
		case *ast.UnaryExpr:
			if x.Op == token.SUB {
				if lit, ok := x.X.(*ast.BasicLit); ok && lit.Kind == token.INT {
					start := offsetOf(x.OpPos)
					add(start, offsetOf(lit.Pos()), "", "const", "sign-drop")
				}
			}
		case *ast.IfStmt:
			start, end := offsetOf(x.Cond.Pos()), offsetOf(x.Cond.End())
			cond := string(src[start:end])
			add(start, end, "!("+cond+")", "negcond", "negate")
		case *ast.BranchStmt:
			if x.Label == nil {
				start := offsetOf(x.TokPos)
				switch x.Tok {
				case token.BREAK:
					add(start, start+len("break"), "continue", "branch", "break->continue")
				case token.CONTINUE:
					add(start, start+len("continue"), "break", "branch", "continue->break")
				}
			}
		}
		return true
	})

	sort.Slice(sites, func(i, j int) bool {
		if sites[i].Offset != sites[j].Offset {
			return sites[i].Offset < sites[j].Offset
		}
		return sites[i].Variant < sites[j].Variant
	})
	return sites, src, nil
}

// intLitDelta returns the literal text for value+delta preserving the base
// prefix. It reports false when the literal cannot be shifted (overflow,
// unparseable, or unsupported form).
func intLitDelta(lit string, delta int64) (string, bool) {
	clean := strings.ReplaceAll(lit, "_", "")
	base := 10
	prefix := ""
	digits := clean
	lower := strings.ToLower(clean)
	switch {
	case strings.HasPrefix(lower, "0x"):
		base, prefix, digits = 16, clean[:2], clean[2:]
	case strings.HasPrefix(lower, "0b"):
		base, prefix, digits = 2, clean[:2], clean[2:]
	case strings.HasPrefix(lower, "0o"):
		base, prefix, digits = 8, clean[:2], clean[2:]
	case len(clean) > 1 && clean[0] == '0':
		base, prefix, digits = 8, clean[:1], clean[1:]
	}
	v, err := strconv.ParseUint(digits, base, 64)
	if err != nil {
		return "", false
	}
	if delta > 0 {
		if v == ^uint64(0) {
			return "", false
		}
		v++
	} else {
		if v == 0 {
			// 0 - 1: emit a parenthesized negative decimal literal; often a
			// compile error in constant contexts (counted invalid), but a
			// legitimate semantic probe where it compiles.
			return "(-1)", true
		}
		v--
	}
	out := prefix + strconv.FormatUint(v, base)
	if base == 8 && prefix == "0" && v == 0 {
		out = "0"
	}
	return out, true
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func targetFiles(cfg config) []string {
	if cfg.files == "arith-core" {
		out := make([]string, len(arithCorePreset))
		copy(out, arithCorePreset)
		return out
	}
	var out []string
	for _, f := range strings.Split(cfg.files, ",") {
		f = strings.TrimSpace(f)
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

func enumerateAll(cfg config, root string) (map[string][]mutationSite, []string, error) {
	files := targetFiles(cfg)
	sort.Strings(files)
	sites := make(map[string][]mutationSite, len(files))
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			return nil, nil, fmt.Errorf("test file %s is not a mutation target", f)
		}
		s, _, err := enumerateFile(root, f)
		if err != nil {
			return nil, nil, err
		}
		sites[f] = s
	}
	return sites, files, nil
}

func printSiteInventory(sites map[string][]mutationSite, files []string) error {
	total := 0
	catTotal := map[string]int{}
	fmt.Printf("%-28s %6s  %s\n", "file", "sites", "by category")
	for _, f := range files {
		byCat := map[string]int{}
		for _, s := range sites[f] {
			byCat[s.Category]++
			catTotal[s.Category]++
		}
		total += len(sites[f])
		fmt.Printf("%-28s %6d  %v\n", f, len(sites[f]), formatCatMap(byCat))
	}
	fmt.Printf("TOTAL %d sites, by category %v\n", total, formatCatMap(catTotal))
	return nil
}

func formatCatMap(m map[string]int) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteString(" ")
		}
		fmt.Fprintf(&b, "%s=%d", k, m[k])
	}
	return b.String()
}

// ---------- sampling ----------

// parseStrata parses "aor=5,cmp=4,..." into per-category quotas.
func parseStrata(s string) (map[string]int, []string, error) {
	if s == "" {
		return nil, nil, nil
	}
	quotas := map[string]int{}
	var order []string
	for _, part := range strings.Split(s, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			return nil, nil, fmt.Errorf("bad strata entry %q", part)
		}
		n, err := strconv.Atoi(kv[1])
		if err != nil {
			return nil, nil, fmt.Errorf("bad strata count %q: %w", part, err)
		}
		quotas[kv[0]] = n
		order = append(order, kv[0])
	}
	return quotas, order, nil
}

// sampleSites picks cfg.perFile mutants per file, deterministically from
// cfg.seed. With strata quotas, each category is sampled up to its quota
// first; leftover slots are backfilled uniformly from the remaining sites,
// so every file still contributes exactly min(perFile, len(sites)) mutants.
func sampleSites(cfg config, sites map[string][]mutationSite, files []string) ([]mutationSite, error) {
	quotas, order, err := parseStrata(cfg.strata)
	if err != nil {
		return nil, err
	}
	rng := rand.New(rand.NewSource(cfg.seed))
	var picked []mutationSite
	for _, f := range files { // files sorted: deterministic rng consumption order
		fs := sites[f]
		n := cfg.perFile
		if n > len(fs) {
			n = len(fs)
		}
		chosen := map[int]bool{}
		if quotas != nil {
			byCat := map[string][]int{}
			for i, s := range fs {
				byCat[s.Category] = append(byCat[s.Category], i)
			}
			for _, cat := range order {
				pool := byCat[cat]
				perm := rng.Perm(len(pool))
				take := quotas[cat]
				if take > len(pool) {
					take = len(pool)
				}
				for _, pi := range perm[:take] {
					if len(chosen) < n {
						chosen[pool[pi]] = true
					}
				}
			}
		}
		if len(chosen) < n {
			var rest []int
			for i := range fs {
				if !chosen[i] {
					rest = append(rest, i)
				}
			}
			perm := rng.Perm(len(rest))
			for _, pi := range perm {
				if len(chosen) >= n {
					break
				}
				chosen[rest[pi]] = true
			}
		}
		idx := make([]int, 0, len(chosen))
		for i := range chosen {
			idx = append(idx, i)
		}
		sort.Ints(idx)
		for _, i := range idx {
			picked = append(picked, fs[i])
		}
	}
	if cfg.maxMutants > 0 && len(picked) > cfg.maxMutants {
		picked = picked[:cfg.maxMutants]
	}
	return picked, nil
}

// sitesByIDFile resolves an explicit mutant ID list against the enumerated
// sites (IDs are stable for a fixed worktree commit).
func sitesByIDFile(path string, sites map[string][]mutationSite, files []string) ([]mutationSite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	index := map[string]mutationSite{}
	for _, f := range files {
		for _, s := range sites[f] {
			index[s.ID()] = s
		}
	}
	var out []mutationSite
	for _, line := range strings.Split(string(data), "\n") {
		id := strings.TrimSpace(line)
		if id == "" {
			continue
		}
		s, ok := index[id]
		if !ok {
			return nil, fmt.Errorf("mutant ID %q not found in enumerated sites (worktree commit drift?)", id)
		}
		out = append(out, s)
	}
	return out, nil
}

// ---------- execution engine ----------

type engine struct {
	cfg      config
	worktree string
	repo     string
	goDir    string // worktree/bid754-go
	binDir   string
	jsonl    *os.File
}

func newEngine(cfg config) (*engine, error) {
	// Structural isolation guard on the exact path evaluateMutant will write
	// under; setupWorktree re-checks the same invariant.
	if err := assertWorktreeIsolated(cfg.repo, cfg.worktree); err != nil {
		return nil, err
	}
	if cfg.jsonlPath == "" {
		return nil, errors.New("-jsonl is required")
	}
	if err := setupWorktree(cfg); err != nil {
		return nil, err
	}
	binDir, err := os.MkdirTemp("", "mutgate-bin-")
	if err != nil {
		return nil, err
	}
	jf, err := os.Create(cfg.jsonlPath)
	if err != nil {
		return nil, err
	}
	return &engine{
		cfg: cfg, worktree: cfg.worktree, repo: cfg.repo,
		goDir: filepath.Join(cfg.worktree, "bid754-go"), binDir: binDir, jsonl: jf,
	}, nil
}

func (e *engine) close() {
	if e.jsonl != nil {
		e.jsonl.Close()
	}
	os.RemoveAll(e.binDir)
}

func (e *engine) buildEnv(binary string) []string {
	env := append([]string{}, os.Environ()...)
	env = append(env, "GOFLAGS=")
	if binary == "portable" || binary == "bidgopkg" {
		env = append(env, "CGO_ENABLED=0")
		return env
	}
	// Mirror .env.sh: decNumber from $HOME/local when present, Intel DFP from
	// the PRIMARY checkout's pinned third_party build (read-only; the
	// worktree does not carry the gitignored lib artifacts).
	home, _ := os.UserHomeDir()
	var cflags, ldflags []string
	if home != "" {
		if _, err := os.Stat(filepath.Join(home, "local", "include", "libdecnumber")); err == nil {
			cflags = append(cflags, "-I"+filepath.Join(home, "local", "include", "libdecnumber"),
				"-I"+filepath.Join(home, "local", "include"))
			ldflags = append(ldflags, "-L"+filepath.Join(home, "local", "lib"))
		}
	}
	intel := filepath.Join(e.repo, "devtools", "third_party", "intel_dfp")
	cflags = append(cflags, "-I"+filepath.Join(intel, "include"))
	ldflags = append(ldflags, "-L"+filepath.Join(intel, "lib"))
	if runtime.GOOS == "darwin" {
		ldflags = append(ldflags, "-framework", "CoreFoundation")
	}
	env = append(env,
		"CGO_ENABLED=1",
		"CGO_CFLAGS="+strings.Join(cflags, " "),
		"CGO_LDFLAGS="+strings.Join(ldflags, " "),
	)
	return env
}

func (e *engine) buildArgs(binary, outPath string) []string {
	args := []string{"test", "-count=1", "-c", "-vet=off", "-o", outPath}
	pkg := "."
	switch binary {
	case "portable":
	case "bidgopkg":
		pkg = "./internal/bidgo"
	case "native":
		args = append(args, "-tags", "bid754_native")
	case "decnumber":
		args = append(args, "-tags", "bid754_native,bid754_decnumber_diff")
	case "tier1long":
		args = append(args, "-tags", "bid754_native,bid754_tier1_long")
	}
	return append(args, pkg)
}

// buildBinary compiles the bid754-go test binary for the given tag set.
// Returns (ok, output, duration).
func (e *engine) buildBinary(binary string) (bool, string, time.Duration, string) {
	outPath := filepath.Join(e.binDir, "mut_"+binary+".test")
	ctx, cancel := context.WithTimeout(context.Background(), e.cfg.buildTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", e.buildArgs(binary, outPath)...)
	cmd.Dir = e.goDir
	cmd.Env = e.buildEnv(binary)
	start := time.Now()
	out, err := cmd.CombinedOutput()
	dur := time.Since(start)
	if err != nil {
		return false, tail(string(out), 2000), dur, outPath
	}
	return true, "", dur, outPath
}

// runStage executes one stage against a prebuilt binary.
// verdict: "pass", "fail", "timeout", "nomatch".
//
// Before running, the stage's -test.run expression is resolved with
// -test.list against the same binary: an expression that selects zero tests
// exits 0 without executing anything, which previously counted as a silent
// stage pass (and a survived mutant). Zero selected tests is therefore a
// stage error ("nomatch"), never a pass — in baseline and per-mutant runs
// alike, since baseline reuses this path.
func (e *engine) runStage(st stage, binPath string) (string, string, time.Duration) {
	dir := e.goDir
	if st.Binary == "bidgopkg" {
		dir = filepath.Join(e.goDir, "internal", "bidgo")
	}

	start := time.Now()
	matched, listOut, err := listMatchedTests(binPath, dir, st.RunExpr, e.cfg.stageTimeout)
	if err != nil {
		return "nomatch", "test selection preflight failed: " + err.Error() + "\n" + tail(listOut, 500), time.Since(start)
	}
	if matched == 0 {
		return "nomatch", fmt.Sprintf("-test.run %q selects zero tests in binary %s; a run would pass without executing anything", st.RunExpr, filepath.Base(binPath)), time.Since(start)
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.cfg.stageTimeout)
	defer cancel()
	args := []string{"-test.run", st.RunExpr, "-test.count=1",
		"-test.timeout", e.cfg.stageTimeout.String()}
	if e.cfg.failfast {
		args = append(args, "-test.failfast")
	}
	cmd := exec.CommandContext(ctx, binPath, args...)
	cmd.Dir = dir
	cmd.Env = append(append([]string{}, os.Environ()...), "GOFLAGS=")
	out, err := cmd.CombinedOutput()
	dur := time.Since(start)
	if ctx.Err() == context.DeadlineExceeded {
		return "timeout", tail(string(out), 1500), dur
	}
	if err != nil {
		return "fail", tail(string(out), 1500), dur
	}
	return "pass", "", dur
}

// listMatchedTests runs `binary -test.list <expr>` and counts the selected
// top-level tests. The list output is one identifier per line; counting those
// lines measures "how many tests the expression selects" directly instead of
// pattern-matching the run output.
func listMatchedTests(binPath, dir, runExpr string, timeout time.Duration) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, binPath, "-test.list", runExpr)
	cmd.Dir = dir
	cmd.Env = append(append([]string{}, os.Environ()...), "GOFLAGS=")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, string(out), fmt.Errorf("-test.list %q: %w", runExpr, err)
	}
	return countListedTests(string(out)), string(out), nil
}

// countListedTests counts the test identifiers in -test.list output (one
// non-empty line per selected test).
func countListedTests(out string) int {
	n := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return "..." + s[len(s)-n:]
}

func failLines(out string) []string {
	var lines []string
	for _, l := range strings.Split(out, "\n") {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "--- FAIL") || strings.Contains(t, ".go:") && strings.Contains(t, "expected") {
			lines = append(lines, truncate(t, 220))
			if len(lines) >= 4 {
				break
			}
		}
	}
	return lines
}

// evaluateMutant applies one mutation in the worktree, compiles, and walks the
// kill-suite stages. The target file is restored before returning.
func (e *engine) evaluateMutant(site mutationSite, pristine []byte, stages []stage) mutantResult {
	res := mutantResult{ID: site.ID(), Site: site, StageMS: map[string]int64{}, BuildMS: map[string]int64{}}
	path := filepath.Join(e.worktree, bidgoRel, site.File)

	mutated := make([]byte, 0, len(pristine)+len(site.Repl))
	mutated = append(mutated, pristine[:site.Offset]...)
	mutated = append(mutated, site.Repl...)
	mutated = append(mutated, pristine[site.End:]...)
	if err := os.WriteFile(path, mutated, 0o644); err != nil {
		res.Status = "invalid"
		res.Reason = "write_error"
		res.Note = err.Error()
		return res
	}
	defer func() {
		if err := os.WriteFile(path, pristine, 0o644); err != nil {
			// A restore failure poisons every later mutant: stop hard.
			fmt.Fprintf(os.Stderr, "FATAL: restore %s failed: %v\n", path, err)
			os.Exit(2)
		}
	}()

	built := map[string]string{}
	for _, st := range stages {
		binPath, ok := built[st.Binary]
		if !ok {
			okBuild, out, dur, p := e.buildBinary(st.Binary)
			res.BuildMS[st.Binary] = dur.Milliseconds()
			if !okBuild {
				if st.Binary == "portable" {
					res.Status = "invalid"
					res.Reason = "compile_error"
					res.Note = truncate(out, 400)
					return res
				}
				// Native build failure for a portable-valid mutant is an
				// infrastructure fault, not a mutant property. The mutant did
				// survive every stage that actually ran.
				res.Status = "survived"
				res.Reason = "native_build_error"
				res.Note = "native build failed: " + truncate(out, 300)
				return res
			}
			built[st.Binary] = p
			binPath = p
		}
		verdict, out, dur := e.runStage(st, binPath)
		res.StageMS[st.Name] = dur.Milliseconds()
		switch verdict {
		case "pass":
			res.Passed = append(res.Passed, st.Name)
		case "fail", "timeout":
			res.Status = "killed"
			res.KilledBy = st.Name
			res.Reason = map[string]string{"fail": "fail", "timeout": "timeout"}[verdict]
			res.FailLines = failLines(out)
			return res
		case "nomatch":
			// A zero-test selection is a harness error, not a mutant verdict:
			// counting it as survived would silently deflate the kill rate.
			res.Status = "invalid"
			res.Reason = "no_test_match"
			res.Note = truncate("stage "+st.Name+": "+out, 400)
			return res
		}
	}
	res.Status = "survived"
	return res
}

func (e *engine) writeResult(r mutantResult) {
	b, err := json.Marshal(r)
	if err == nil {
		e.jsonl.Write(append(b, '\n'))
	}
}

func resolveStages(cfg config) ([]stage, error) {
	var out []stage
	for _, name := range strings.Split(cfg.stages, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		st, ok := stageCatalog[name]
		if !ok {
			return nil, fmt.Errorf("unknown stage %q", name)
		}
		out = append(out, st)
	}
	if len(out) == 0 {
		return nil, errors.New("no stages selected")
	}
	return out, nil
}

// baseline verifies every requested stage passes on the pristine worktree.
func (e *engine) baseline(stages []stage) error {
	fmt.Println("baseline: building and running all stages on pristine worktree...")
	built := map[string]string{}
	for _, st := range stages {
		binPath, ok := built[st.Binary]
		if !ok {
			okBuild, out, dur, p := e.buildBinary(st.Binary)
			if !okBuild {
				return fmt.Errorf("baseline build %s failed:\n%s", st.Binary, out)
			}
			fmt.Printf("baseline build %-9s %6dms\n", st.Binary, dur.Milliseconds())
			built[st.Binary] = p
			binPath = p
		}
		verdict, out, dur := e.runStage(st, binPath)
		fmt.Printf("baseline stage %-9s %-7s %6dms\n", st.Name, verdict, dur.Milliseconds())
		if verdict != "pass" {
			return fmt.Errorf("baseline stage %s did not pass (%s):\n%s", st.Name, verdict, out)
		}
	}
	return nil
}

func runMutants(cfg config) error {
	e, err := newEngine(cfg)
	if err != nil {
		return err
	}
	defer e.close()

	stages, err := resolveStages(cfg)
	if err != nil {
		return err
	}
	sites, files, err := enumerateAll(cfg, e.worktree)
	if err != nil {
		return err
	}
	pristine := map[string][]byte{}
	for _, f := range files {
		b, err := os.ReadFile(filepath.Join(e.worktree, bidgoRel, f))
		if err != nil {
			return err
		}
		pristine[f] = b
	}
	if err := e.baseline(stages); err != nil {
		return err
	}
	var picked []mutationSite
	if cfg.idsFile != "" {
		picked, err = sitesByIDFile(cfg.idsFile, sites, files)
		if err != nil {
			return err
		}
		fmt.Printf("resolved %d mutants from -ids-file %s\n", len(picked), cfg.idsFile)
	} else {
		picked, err = sampleSites(cfg, sites, files)
		if err != nil {
			return err
		}
		fmt.Printf("sampled %d mutants (seed=%d, per-file=%d, strata=%q, files=%d)\n",
			len(picked), cfg.seed, cfg.perFile, cfg.strata, len(files))
	}

	head, _ := exec.Command("git", "-C", e.worktree, "rev-parse", "HEAD").Output()
	fmt.Printf("worktree commit: %s", head)

	summary := map[string]int{}
	killedByStage := map[string]int{}
	byCat := map[string]map[string]int{}
	start := time.Now()
	for i, site := range picked {
		r := e.evaluateMutant(site, pristine[site.File], stages)
		e.writeResult(r)
		summary[r.Status]++
		if r.Status == "killed" {
			killedByStage[r.KilledBy]++
		}
		if byCat[site.Category] == nil {
			byCat[site.Category] = map[string]int{}
		}
		byCat[site.Category][r.Status]++
		if (i+1)%cfg.logEvery == 0 || i == len(picked)-1 {
			fmt.Printf("[%d/%d] killed=%d survived=%d invalid=%d elapsed=%s\n",
				i+1, len(picked), summary["killed"], summary["survived"], summary["invalid"],
				time.Since(start).Round(time.Second))
		}
	}

	fmt.Println("---- SUMMARY ----")
	valid := summary["killed"] + summary["survived"]
	fmt.Printf("total=%d valid=%d killed=%d survived=%d invalid=%d\n",
		len(picked), valid, summary["killed"], summary["survived"], summary["invalid"])
	if valid > 0 {
		fmt.Printf("kill rate (killed/valid) = %.1f%%\n", 100*float64(summary["killed"])/float64(valid))
	}
	fmt.Printf("killed by stage: %v\n", formatCatMap(killedByStage))
	for cat, m := range byCat {
		fmt.Printf("category %-8s killed=%d survived=%d invalid=%d\n", cat, m["killed"], m["survived"], m["invalid"])
	}
	return e.reportWorktreeClean()
}

// reportWorktreeClean re-verifies at exit that the worktree's bid754-go tree
// carries no leftover mutation. A dirty exit is a hard error (non-zero
// process exit), not a warning: leftover state invalidates this run's
// isolation contract and would poison any later reuse.
func (e *engine) reportWorktreeClean() error {
	if err := checkWorktreeCleanAfterRun(e.worktree); err != nil {
		return err
	}
	fmt.Println("worktree bid754-go clean after run: OK")
	return nil
}

func checkWorktreeCleanAfterRun(worktree string) error {
	dirty, err := worktreeDirty(worktree)
	switch {
	case err != nil:
		return fmt.Errorf("worktree status check failed after run: %w", err)
	case dirty != "":
		return fmt.Errorf("worktree bid754-go not clean after run (leftover mutation; isolation contract violated):\n%s", dirty)
	}
	return nil
}

// ---------- self-check ----------

// selfCheckSpec pins a known mutant by (file, function, category, substring).
type selfCheckSpec struct {
	name     string
	file     string
	funcName string
	category string
	contains string
	expect   string // killed|survived
	why      string
}

var selfChecks = []selfCheckSpec{
	{
		name: "killable-fma-clamp", file: "bid128_fma_body.go",
		funcName: "bid_fma_delta_ge_zero", category: "negcond",
		contains: "bid_fma_case1ppB_psign_ne_zsign(p34, res",
		expect:   "killed",
		why:      "inverts the D2-regression-area overflow-clamp helper branch in Bid128Fma (the goto-done exit restored by commit c08ac65 stops firing on the exact rows that need it and fires on all others); differential gates must fail",
	},
	{
		name: "equivalent-dead-func", file: "to_bid3264.go",
		funcName: "bid32GetNoFlags", category: "cmp",
		contains: "coeff > 9999999",
		expect:   "survived",
		why:      "bid32GetNoFlags is referenced nowhere in package bidgo (definition-only symbol, unexported, so unreachable from any other package); any mutation inside it is semantics-preserving for every gate",
	},
}

// findSelfCheckSite pins a site by category + enclosing function + a source
// window around the site containing the locator substring. First offset-order
// match wins (deterministic).
func findSelfCheckSite(sites []mutationSite, src []byte, sc selfCheckSpec) (mutationSite, bool) {
	for _, s := range sites {
		if s.Category != sc.category {
			continue
		}
		if sc.funcName != "" && s.Func != sc.funcName {
			continue
		}
		lo := s.Offset - 80
		if lo < 0 {
			lo = 0
		}
		hi := s.End + 80
		if hi > len(src) {
			hi = len(src)
		}
		if !strings.Contains(string(src[lo:hi]), sc.contains) {
			continue
		}
		return s, true
	}
	return mutationSite{}, false
}

func runSelfCheck(cfg config) error {
	e, err := newEngine(cfg)
	if err != nil {
		return err
	}
	defer e.close()
	stages, err := resolveStages(cfg)
	if err != nil {
		return err
	}
	if err := e.baseline(stages); err != nil {
		return err
	}
	failed := 0
	for _, sc := range selfChecks {
		sites, pristine, err := enumerateFile(e.worktree, sc.file)
		if err != nil {
			return err
		}
		site, ok := findSelfCheckSite(sites, pristine, sc)
		if !ok {
			return fmt.Errorf("self-check %s: site not found (file drifted? re-pin locator)", sc.name)
		}
		fmt.Printf("self-check %s: %s:%d %s %q -> %q\n  rationale: %s\n", sc.name, site.File, site.Line, site.Category, site.Orig, site.Mutated, sc.why)
		r := e.evaluateMutant(site, pristine, stages)
		e.writeResult(r)
		verdictNote := ""
		if r.Status == "killed" {
			verdictNote = fmt.Sprintf(" (killed_by=%s)", r.KilledBy)
		}
		if r.Status != sc.expect {
			failed++
			fmt.Printf("self-check %s: FAIL expected %s got %s%s note=%s\n", sc.name, sc.expect, r.Status, verdictNote, r.Note)
		} else {
			fmt.Printf("self-check %s: OK %s%s\n", sc.name, r.Status, verdictNote)
		}
	}
	if err := e.reportWorktreeClean(); err != nil {
		return err
	}
	if failed > 0 {
		return fmt.Errorf("%d self-check(s) failed", failed)
	}
	fmt.Println("self-check: all OK")
	return nil
}
