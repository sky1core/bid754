//go:build cgo && bid754_native

package main

/*
#cgo CFLAGS: -I${SRCDIR}/../../../../devtools/third_party/intel_dfp/src -I${SRCDIR}/../../../../devtools/third_party/intel_dfp/include
#cgo LDFLAGS: -L${SRCDIR}/../../../../devtools/third_party/intel_dfp/lib -lbid -lm

#include <stdint.h>
#include "bid_conf.h"
#include "bid_functions.h"

static BID_UINT128 explore_mk128(BID_UINT64 lo, BID_UINT64 hi) {
	BID_UINT128 r;
	r.w[BID_LOW_128W] = lo;
	r.w[BID_HIGH_128W] = hi;
	return r;
}
static BID_UINT64 explore_lo128(BID_UINT128 x) { return x.w[BID_LOW_128W]; }
static BID_UINT64 explore_hi128(BID_UINT128 x) { return x.w[BID_HIGH_128W]; }
*/
import "C"

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	bidgo "github.com/sky1core/bid754/bid754-go/internal/bidgo"
)

// ---------- JSONL records ----------

type configRecord struct {
	Type          string         `json:"type"`
	Tool          string         `json:"tool"`
	FormatVersion int            `json:"format_version"`
	Seed          string         `json:"seed"` // decimal uint64 as string (JSON number precision)
	CasesPerTgt   int            `json:"cases_per_target"`
	Bias          float64        `json:"bias"`
	Ops           []string       `json:"ops"`
	Widths        []int          `json:"widths"`
	Modes         []string       `json:"modes"`
	PoolSizes     map[string]int `json:"pool_sizes"`
	Commit        string         `json:"commit,omitempty"`
	GoOS          string         `json:"go_os"`
	GoArch        string         `json:"go_arch"`
}

type mismatchRecord struct {
	Type      string `json:"type"`
	Target    string `json:"target"`
	Op        string `json:"op"`
	Width     int    `json:"width"`
	Mode      string `json:"mode"`
	Rnd       int    `json:"rnd"`
	CaseIndex int    `json:"case_index"`
	X         string `json:"x"`
	Y         string `json:"y,omitempty"`
	Z         string `json:"z,omitempty"`
	CBits     string `json:"c_bits"`
	CFlags    string `json:"c_flags"`
	GoBits    string `json:"go_bits"`
	GoFlags   string `json:"go_flags"`
}

type targetSummaryRecord struct {
	Type        string `json:"type"`
	Target      string `json:"target"`
	Cases       int    `json:"cases"`
	Comparisons int    `json:"comparisons"`
	Mismatches  int    `json:"mismatches"`
	ElapsedMS   int64  `json:"elapsed_ms"`
}

type summaryRecord struct {
	Type        string `json:"type"`
	Targets     int    `json:"targets"`
	Cases       int    `json:"cases"`
	Comparisons int    `json:"comparisons"`
	Mismatches  int    `json:"mismatches"`
	ElapsedMS   int64  `json:"elapsed_ms"`
}

func main() {
	seedText := flag.String("seed", "", "case-stream seed, decimal or 0x hex uint64 (required)")
	cases := flag.Int("cases", 20000, "cases per (width, op) target; each case runs under every selected mode")
	bias := flag.Float64("bias", 0.25, "probability in [0,1] that a case is boundary-biased (pool draw + exponent correlation)")
	opsText := flag.String("ops", strings.Join(allOps, ","), "CSV of Tier 1 arithmetic ops")
	widthsText := flag.String("widths", "32,64,128", "CSV of decimal widths")
	modesText := flag.String("modes", modeNamesCSV(), "CSV of rounding-mode names")
	commit := flag.String("commit", "", "commit descriptor recorded in the config record")
	flag.Parse()

	if err := run(*seedText, *cases, *bias, *opsText, *widthsText, *modesText, *commit); err != nil {
		fmt.Fprintf(os.Stderr, "explorenative: %v\n", err)
		os.Exit(1)
	}
}

func modeNamesCSV() string {
	names := make([]string, len(allModes))
	for i, m := range allModes {
		names[i] = m.name
	}
	return strings.Join(names, ",")
}

func run(seedText string, cases int, bias float64, opsText, widthsText, modesText, commit string) error {
	if seedText == "" {
		return fmt.Errorf("-seed is required (the driver always passes the resolved seed)")
	}
	seed, err := strconv.ParseUint(seedText, 0, 64)
	if err != nil {
		return fmt.Errorf("invalid -seed %q: %v", seedText, err)
	}
	if cases <= 0 {
		return fmt.Errorf("-cases must be positive, got %d", cases)
	}
	if math.IsNaN(bias) || bias < 0 || bias > 1 {
		return fmt.Errorf("-bias must be in [0,1], got %g", bias)
	}
	var ops []string
	for _, op := range splitCSV(opsText) {
		known := false
		for _, k := range allOps {
			if op == k {
				known = true
			}
		}
		if !known {
			return fmt.Errorf("unknown op %q (known: %s)", op, strings.Join(allOps, ","))
		}
		ops = append(ops, op)
	}
	var widths []widthParams
	var widthBits []int
	for _, text := range splitCSV(widthsText) {
		n, err := strconv.Atoi(text)
		if err != nil {
			return fmt.Errorf("invalid width %q", text)
		}
		w, ok := widthByBits(n)
		if !ok {
			return fmt.Errorf("unknown width %d (known: 32,64,128)", n)
		}
		widths = append(widths, w)
		widthBits = append(widthBits, n)
	}
	var modes []modeSpec
	var modeNames []string
	for _, name := range splitCSV(modesText) {
		m, ok := modeByName(name)
		if !ok {
			return fmt.Errorf("unknown mode %q (known: %s)", name, modeNamesCSV())
		}
		modes = append(modes, m)
		modeNames = append(modeNames, name)
	}
	if len(ops) == 0 || len(widths) == 0 || len(modes) == 0 {
		return fmt.Errorf("ops, widths, and modes must each select at least one entry")
	}

	out := bufio.NewWriterSize(os.Stdout, 1<<16)
	defer out.Flush()
	emit := func(record any) error {
		b, err := json.Marshal(record)
		if err != nil {
			return err
		}
		if _, err := out.Write(append(b, '\n')); err != nil {
			return err
		}
		return nil
	}

	pools := map[int][]poolEntry{}
	poolSizes := map[string]int{}
	for _, w := range widths {
		pool := buildPool(w)
		pools[w.width] = pool
		poolSizes[strconv.Itoa(w.width)] = len(pool)
	}
	if err := emit(configRecord{
		Type: "config", Tool: "explorediff", FormatVersion: 1,
		Seed: strconv.FormatUint(seed, 10), CasesPerTgt: cases, Bias: bias,
		Ops: ops, Widths: widthBits, Modes: modeNames, PoolSizes: poolSizes,
		Commit: commit, GoOS: runtime.GOOS, GoArch: runtime.GOARCH,
	}); err != nil {
		return err
	}

	start := time.Now()
	totalComparisons, totalMismatches, totalCases, targets := 0, 0, 0, 0
	for _, w := range widths {
		for _, op := range ops {
			targetStart := time.Now()
			targetName := fmt.Sprintf("d%d/%s", w.width, op)
			rng := &splitMix64{state: targetSeed(seed, w, op)}
			comparisons, mismatches := 0, 0
			for i := 0; i < cases; i++ {
				tup := drawTuple(rng, w, op, bias, pools[w.width])
				for _, mode := range modes {
					cBits, cFlags := cEval(w, op, tup, mode.native)
					gBits, gFlags := goEval(w, op, tup, mode.native)
					comparisons++
					if cBits != gBits || cFlags != gFlags {
						mismatches++
						rec := mismatchRecord{
							Type: "mismatch", Target: targetName, Op: op, Width: w.width,
							Mode: mode.name, Rnd: mode.native, CaseIndex: i,
							X:     formatValue(w, tup[0]),
							CBits: formatValue(w, cBits), CFlags: fmt.Sprintf("%08x", cFlags),
							GoBits: formatValue(w, gBits), GoFlags: fmt.Sprintf("%08x", gFlags),
						}
						if opArity(op) >= 2 {
							rec.Y = formatValue(w, tup[1])
						}
						if opArity(op) >= 3 {
							rec.Z = formatValue(w, tup[2])
						}
						if err := emit(rec); err != nil {
							return err
						}
					}
				}
			}
			if err := emit(targetSummaryRecord{
				Type: "target_summary", Target: targetName, Cases: cases,
				Comparisons: comparisons, Mismatches: mismatches,
				ElapsedMS: time.Since(targetStart).Milliseconds(),
			}); err != nil {
				return err
			}
			totalComparisons += comparisons
			totalMismatches += mismatches
			totalCases += cases
			targets++
		}
	}
	return emit(summaryRecord{
		Type: "summary", Targets: targets, Cases: totalCases,
		Comparisons: totalComparisons, Mismatches: totalMismatches,
		ElapsedMS: time.Since(start).Milliseconds(),
	})
}

func splitCSV(text string) []string {
	var out []string
	for _, part := range strings.Split(text, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

// ---------- pinned Intel C leg ----------

func cEval(w widthParams, op string, tup [3]words, rnd int) (words, uint32) {
	switch w.width {
	case 32:
		bits, flags := cEval32(op, uint32(tup[0].lo), uint32(tup[1].lo), uint32(tup[2].lo), rnd)
		return words{lo: uint64(bits)}, flags
	case 64:
		bits, flags := cEval64(op, tup[0].lo, tup[1].lo, tup[2].lo, rnd)
		return words{lo: bits}, flags
	default:
		return cEval128(op, tup[0], tup[1], tup[2], rnd)
	}
}

func cEval32(op string, x, y, z uint32, rnd int) (uint32, uint32) {
	var flags C._IDEC_flags
	mode := C._IDEC_round(rnd)
	var res C.BID_UINT32
	switch op {
	case "add":
		res = C.bid32_add(C.BID_UINT32(x), C.BID_UINT32(y), mode, &flags)
	case "sub":
		res = C.bid32_sub(C.BID_UINT32(x), C.BID_UINT32(y), mode, &flags)
	case "mul":
		res = C.bid32_mul(C.BID_UINT32(x), C.BID_UINT32(y), mode, &flags)
	case "div":
		res = C.bid32_div(C.BID_UINT32(x), C.BID_UINT32(y), mode, &flags)
	case "fma":
		res = C.bid32_fma(C.BID_UINT32(x), C.BID_UINT32(y), C.BID_UINT32(z), mode, &flags)
	case "sqrt":
		res = C.bid32_sqrt(C.BID_UINT32(x), mode, &flags)
	case "quantize":
		res = C.bid32_quantize(C.BID_UINT32(x), C.BID_UINT32(y), mode, &flags)
	default:
		panic("unsupported explore op " + op)
	}
	return uint32(res), uint32(flags)
}

func cEval64(op string, x, y, z uint64, rnd int) (uint64, uint32) {
	var flags C._IDEC_flags
	mode := C._IDEC_round(rnd)
	var res C.BID_UINT64
	switch op {
	case "add":
		res = C.bid64_add(C.BID_UINT64(x), C.BID_UINT64(y), mode, &flags)
	case "sub":
		res = C.bid64_sub(C.BID_UINT64(x), C.BID_UINT64(y), mode, &flags)
	case "mul":
		res = C.bid64_mul(C.BID_UINT64(x), C.BID_UINT64(y), mode, &flags)
	case "div":
		res = C.bid64_div(C.BID_UINT64(x), C.BID_UINT64(y), mode, &flags)
	case "fma":
		res = C.bid64_fma(C.BID_UINT64(x), C.BID_UINT64(y), C.BID_UINT64(z), mode, &flags)
	case "sqrt":
		res = C.bid64_sqrt(C.BID_UINT64(x), mode, &flags)
	case "quantize":
		res = C.bid64_quantize(C.BID_UINT64(x), C.BID_UINT64(y), mode, &flags)
	default:
		panic("unsupported explore op " + op)
	}
	return uint64(res), uint32(flags)
}

func cEval128(op string, x, y, z words, rnd int) (words, uint32) {
	var flags C._IDEC_flags
	mode := C._IDEC_round(rnd)
	cx := C.explore_mk128(C.BID_UINT64(x.lo), C.BID_UINT64(x.hi))
	cy := C.explore_mk128(C.BID_UINT64(y.lo), C.BID_UINT64(y.hi))
	cz := C.explore_mk128(C.BID_UINT64(z.lo), C.BID_UINT64(z.hi))
	var res C.BID_UINT128
	switch op {
	case "add":
		res = C.bid128_add(cx, cy, mode, &flags)
	case "sub":
		res = C.bid128_sub(cx, cy, mode, &flags)
	case "mul":
		res = C.bid128_mul(cx, cy, mode, &flags)
	case "div":
		res = C.bid128_div(cx, cy, mode, &flags)
	case "fma":
		res = C.bid128_fma(cx, cy, cz, mode, &flags)
	case "sqrt":
		res = C.bid128_sqrt(cx, mode, &flags)
	case "quantize":
		res = C.bid128_quantize(cx, cy, mode, &flags)
	default:
		panic("unsupported explore op " + op)
	}
	return words{lo: uint64(C.explore_lo128(res)), hi: uint64(C.explore_hi128(res))}, uint32(flags)
}

// ---------- Go mechanical-port leg ----------

func goEval(w widthParams, op string, tup [3]words, rnd int) (words, uint32) {
	switch w.width {
	case 32:
		bits, flags := goEval32(op, uint32(tup[0].lo), uint32(tup[1].lo), uint32(tup[2].lo), rnd)
		return words{lo: uint64(bits)}, flags
	case 64:
		bits, flags := goEval64(op, tup[0].lo, tup[1].lo, tup[2].lo, rnd)
		return words{lo: bits}, flags
	default:
		return goEval128(op, tup[0], tup[1], tup[2], rnd)
	}
}

func goEval32(op string, x, y, z uint32, rnd int) (uint32, uint32) {
	switch op {
	case "add":
		return bidgo.Bid32AddWithFlags(x, y, rnd)
	case "sub":
		return bidgo.Bid32SubWithFlags(x, y, rnd)
	case "mul":
		return bidgo.Bid32MulWithFlags(x, y, rnd)
	case "div":
		return bidgo.Bid32DivWithFlags(x, y, rnd)
	case "fma":
		return bidgo.Bid32Fma(x, y, z, rnd)
	case "sqrt":
		return bidgo.Bid32Sqrt(x, rnd)
	case "quantize":
		return bidgo.Bid32Quantize(x, y, rnd)
	default:
		panic("unsupported explore op " + op)
	}
}

func goEval64(op string, x, y, z uint64, rnd int) (uint64, uint32) {
	switch op {
	case "add":
		return bidgo.Bid64AddWithFlags(x, y, rnd)
	case "sub":
		return bidgo.Bid64SubWithFlags(x, y, rnd)
	case "mul":
		return bidgo.Bid64MulWithFlags(x, y, rnd)
	case "div":
		return bidgo.Bid64DivWithFlags(x, y, rnd)
	case "fma":
		return bidgo.Bid64Fma(x, y, z, rnd)
	case "sqrt":
		return bidgo.Bid64Sqrt(x, rnd)
	case "quantize":
		return bidgo.Bid64Quantize(x, y, rnd)
	default:
		panic("unsupported explore op " + op)
	}
}

func goEval128(op string, x, y, z words, rnd int) (words, uint32) {
	gx := bidgo.Bid128FromWords(x.hi, x.lo)
	gy := bidgo.Bid128FromWords(y.hi, y.lo)
	gz := bidgo.Bid128FromWords(z.hi, z.lo)
	var res bidgo.BID_UINT128
	var flags uint32
	switch op {
	case "add":
		res = bidgo.Bid128Add(gx, gy, rnd, &flags)
	case "sub":
		res = bidgo.Bid128Sub(gx, gy, rnd, &flags)
	case "mul":
		res, flags = bidgo.Bid128Mul(gx, gy, rnd)
	case "div":
		res, flags = bidgo.Bid128Div(gx, gy, rnd)
	case "fma":
		res, flags = bidgo.Bid128Fma(gx, gy, gz, rnd)
	case "sqrt":
		res, flags = bidgo.Bid128Sqrt(gx, rnd)
	case "quantize":
		res, flags = bidgo.Bid128Quantize(gx, gy, rnd)
	default:
		panic("unsupported explore op " + op)
	}
	hi, lo := bidgo.Bid128Words(res)
	return words{lo: lo, hi: hi}, flags
}
