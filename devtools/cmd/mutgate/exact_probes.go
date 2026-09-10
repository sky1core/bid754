package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

type exactProbe struct {
	Name        string
	File        string
	Function    string
	Original    string
	Replacement string
	Body        bool
}

func exactProbes(set string) ([]exactProbe, error) {
	switch set {
	case "calibration":
		return []exactProbe{
			{Name: "quantize-midpoint", File: "bid32_quantize.go", Function: "Bid32Quantize", Original: "C64&1 != 0", Replacement: "C64&1 == 0"},
			{Name: "mul-inexact", File: "bid32_mul.go", Function: "bid32_mul_core", Original: "flags |= BID_INEXACT_EXCEPTION", Replacement: "flags |= 0"},
			{Name: "unfused32", File: "bid32_fma.go", Function: "Bid32Fma", Body: true, Replacement: "{ { p, pf := bid32_mul_core(x, y, rnd_mode); r, rf := bid32_add_core(p, z, rnd_mode); return r, pf | rf }"},
		}, nil
	case "heldout":
		return []exactProbe{
			{Name: "mul-midpoint", File: "bid32_mul.go", Function: "bid32_mul_core", Original: "R == 0", Replacement: "R != 0"},
			{Name: "add-inexact", File: "bid32_add.go", Function: "bid32_add_core", Original: "flags |= BID_INEXACT_EXCEPTION", Replacement: "flags |= 0"},
			{Name: "unfused64", File: "fma64.go", Function: "Bid64Fma", Body: true, Replacement: "{ { p, pf := Bid64MulWithFlags(x, y, rndMode); r, rf := Bid64AddWithFlags(p, z, rndMode); return r, pf | rf }"},
		}, nil
	default:
		return nil, fmt.Errorf("unknown exact probe set %q", set)
	}
}

func exactProbeSite(root string, p exactProbe) (mutationSite, []byte, error) {
	path := filepath.Join(root, bidgoRel, p.File)
	src, err := os.ReadFile(path)
	if err != nil {
		return mutationSite{}, nil, err
	}
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, path, src, 0)
	if err != nil {
		return mutationSite{}, nil, err
	}
	var candidates []ast.Node
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != p.Function {
			continue
		}
		if p.Body {
			candidates = append(candidates, fn.Body)
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			switch n.(type) {
			case *ast.BinaryExpr, *ast.AssignStmt:
				if string(src[fs.Position(n.Pos()).Offset:fs.Position(n.End()).Offset]) == p.Original {
					candidates = append(candidates, n)
				}
			}
			return true
		})
	}
	if len(candidates) != 1 {
		return mutationSite{}, nil, fmt.Errorf("probe %s requires one exact AST site, found %d", p.Name, len(candidates))
	}
	n := candidates[0]
	pos, end := fs.Position(n.Pos()), fs.Position(n.End())
	if p.Body {
		end.Offset = pos.Offset + 1
	}
	original := string(src[pos.Offset:end.Offset])
	candidate := append([]byte(nil), src[:pos.Offset]...)
	candidate = append(candidate, p.Replacement...)
	candidate = append(candidate, src[end.Offset:]...)
	if bytes.Equal(candidate, src) {
		return mutationSite{}, nil, fmt.Errorf("probe %s is a byte no-op", p.Name)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), path, candidate, 0); err != nil {
		return mutationSite{}, nil, err
	}
	return mutationSite{File: p.File, Func: p.Function, Category: "exactprobe", Variant: p.Name, Offset: pos.Offset, End: end.Offset, Line: pos.Line, Col: pos.Column, Orig: truncate(original, 180), OrigFull: original, Mutated: truncate(p.Replacement, 180), Repl: p.Replacement}, src, nil
}

func intendedExactFinding(p exactProbe, run exactRun) bool {
	if run.Config.Campaign != "witness" || run.Config.Witness != p.Name || run.Status != "killed" || run.Report == nil || len(run.Report.Findings) != 1 {
		return false
	}
	f := run.Report.Findings[0]
	if !f.IntendedWitness || f.ForbiddenRaw == "" {
		return false
	}
	switch p.Name {
	case "quantize-midpoint", "mul-midpoint":
		return f.Sample.Case.Mode == "nearest_even" && f.Rounding == "tie" && f.ValueMismatch && !f.FlagsMismatch
	case "mul-inexact", "add-inexact":
		return !f.ValueMismatch && f.FlagsMismatch && f.ExpectedFlags&0x20 != 0 && f.ActualFlags == f.ExpectedFlags&^0x20
	case "unfused32", "unfused64":
		return f.Sample.Case.Op == "fma" && f.ValueMismatch && f.ExpectedFlags == 0 && f.ActualFlags == 0x20
	}
	return false
}

func runExactCheck(cfg config) error {
	if cfg.stages != "exactprobe" {
		return fmt.Errorf("exactcheck requires -stages exactprobe")
	}
	probes, err := exactProbes(cfg.exactProbeSet)
	if err != nil {
		return err
	}
	configs, err := exactConfigs(cfg)
	if err != nil {
		return err
	}
	tuningSeeds, err := exactSeedList(cfg.exactTuningSeeds)
	if err != nil {
		return err
	}
	if cfg.exactProbeSet == "heldout" {
		for _, c := range configs {
			for _, seed := range tuningSeeds {
				if c.Seed == seed {
					return fmt.Errorf("heldout campaign seed %d overlaps declared tuning seeds", c.Seed)
				}
			}
		}
	}
	e, err := newEngine(cfg)
	if err != nil {
		return err
	}
	defer e.close()
	e.exact.ProbeSet = cfg.exactProbeSet
	e.exact.TuningSeeds = tuningSeeds
	for _, p := range probes {
		e.exact.Configs = append(e.exact.Configs, exactConfig{Campaign: "witness", Cases: 1, Witness: p.Name})
	}
	stages := []stage{stageCatalog["exactprobe"]}
	if err := e.baseline(stages); err != nil {
		return err
	}
	failures := 0
	for _, p := range probes {
		site, src, err := exactProbeSite(e.worktree, p)
		if err != nil {
			return err
		}
		result := e.evaluateMutant(site, src, stages)
		proven := false
		if result.Exact != nil {
			for _, run := range result.Exact.Runs {
				if intendedExactFinding(p, run) {
					proven = true
				}
			}
		}
		if !proven {
			failures++
			result.Note = "intended non-equivalent witness diagnostic absent"
		}
		if err := e.writeResult(result); err != nil {
			return err
		}
		fmt.Printf("exactcheck %s: status=%s intended_witness=%t\n", p.Name, result.Status, proven)
	}
	if err := e.reportWorktreeClean(); err != nil {
		return err
	}
	if failures > 0 {
		return fmt.Errorf("%d exact probes lacked intended arithmetic evidence", failures)
	}
	return nil
}
