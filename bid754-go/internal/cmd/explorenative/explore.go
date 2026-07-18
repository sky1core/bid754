// Command explorenative is the in-module execution leg of the fresh-seed
// exploration tool devtools/cmd/explorediff (a discovery/audit tool like
// devtools/cmd/mutgate, not a verification domain).
//
// It generates seed-derived random Tier 1 arithmetic cases (add/sub/mul/div/
// fma/sqrt/quantize across Decimal32/64/128 and the five rounding modes),
// runs each case through both the Go mechanical port (internal/bidgo) and
// pinned Intel BID C (cgo), and emits one JSONL record per exact
// (result bits, raw flags) mismatch on stdout, plus config/summary records.
// Findings are inputs to the existing manual promotion procedures
// (regression vectors, sentinels, corpus); this tool only discovers and
// records them.
//
// It lives inside the bid754-go module because devtools is a stdlib-only
// module that must not require bid754-go (docs/SPEC.md inter-component
// dependency rules); devtools/cmd/explorediff runs it as a subprocess from
// the bid754-go module directory, the same filesystem relationship the
// sentinel codegen uses for internal/cmd/sentineloracle.
//
// This file holds the portable core: the splitmix64 case stream, the
// boundary-bias operand pool, and the BID encoders. The cgo differential
// leg is in main_native.go (build tags cgo && bid754_native); main_stub.go
// explains the required build otherwise.
package main

import (
	"fmt"
	"math/bits"
)

// ---------- deterministic case stream ----------

// splitMix64 is Vigna's SplitMix64 generator. One instance drives one
// (width, op) target's whole case stream, so a fixed -seed reproduces every
// drawn operand tuple exactly.
type splitMix64 struct{ state uint64 }

func (r *splitMix64) next() uint64 {
	r.state += 0x9e3779b97f4a7c15
	z := r.state
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

func (r *splitMix64) intn(n int) int {
	return int(r.next() % uint64(n))
}

// chance reports true with probability p (53-bit uniform draw).
func (r *splitMix64) chance(p float64) bool {
	if p <= 0 {
		return false
	}
	return float64(r.next()>>11)/(1<<53) < p
}

// span returns a uniform draw in [lo, hi] (inclusive, lo <= hi).
func (r *splitMix64) span(lo, hi int64) int64 {
	return lo + int64(r.next()%uint64(hi-lo+1))
}

// fnv64a hashes the target label into a per-target seed offset.
func fnv64a(text string) uint64 {
	h := uint64(0xcbf29ce484222325)
	for i := 0; i < len(text); i++ {
		h ^= uint64(text[i])
		h *= 0x100000001b3
	}
	return h
}

// ---------- 128-bit value plumbing ----------

// words carries one operand or result as raw BID bits: Decimal32 in the low
// 32 bits of lo, Decimal64 in lo, Decimal128 as lo/hi.
type words struct {
	lo uint64
	hi uint64
}

func u128FromUint64(v uint64) words { return words{lo: v} }

func u128Mul10(v words) words {
	hi, lo := bits.Mul64(v.lo, 10)
	return words{lo: lo, hi: v.hi*10 + hi}
}

func u128Sub64(v words, d uint64) words {
	lo, borrow := bits.Sub64(v.lo, d, 0)
	return words{lo: lo, hi: v.hi - borrow}
}

// pow10u128 returns 10^n for 0 <= n <= 38.
func pow10u128(n int) words {
	v := u128FromUint64(1)
	for i := 0; i < n; i++ {
		v = u128Mul10(v)
	}
	return v
}

// ---------- width parameters and encoding ----------

type widthParams struct {
	width        int
	precision    int64
	expBias      int64 // biased exponent of 1E0
	maxBiasedExp int64 // largest biased exponent field value
}

var width32 = widthParams{width: 32, precision: 7, expBias: 101, maxBiasedExp: 191}
var width64 = widthParams{width: 64, precision: 16, expBias: 398, maxBiasedExp: 767}
var width128 = widthParams{width: 128, precision: 34, expBias: 6176, maxBiasedExp: 12287}

func widthByBits(width int) (widthParams, bool) {
	switch width {
	case 32:
		return width32, true
	case 64:
		return width64, true
	case 128:
		return width128, true
	}
	return widthParams{}, false
}

// encodeBID assembles the BID interchange encoding of
// (-1)^sign * coeff * 10^(biasedExp - bias) for the width, using the normal
// form when the coefficient fits and the special (steering-bits 11) form
// otherwise for the 32/64-bit widths. Coefficients beyond the encodable
// range report ok=false. Non-canonical coefficients (> 10^p - 1) encode
// fine on purpose: they are legitimate boundary inputs.
func encodeBID(w widthParams, sign uint64, biasedExp int64, coeff words) (words, bool) {
	if biasedExp < 0 || biasedExp > w.maxBiasedExp {
		return words{}, false
	}
	e := uint64(biasedExp)
	switch w.width {
	case 32:
		if coeff.hi != 0 {
			return words{}, false
		}
		c := coeff.lo
		if c < 1<<23 {
			return words{lo: sign<<31 | e<<23 | c}, true
		}
		if c <= 1<<23+1<<21-1 {
			return words{lo: sign<<31 | 0x6000_0000 | e<<21 | (c & (1<<21 - 1))}, true
		}
		return words{}, false
	case 64:
		if coeff.hi != 0 {
			return words{}, false
		}
		c := coeff.lo
		if c < 1<<53 {
			return words{lo: sign<<63 | e<<53 | c}, true
		}
		if c <= 1<<53+1<<51-1 {
			return words{lo: sign<<63 | 0x6000_0000_0000_0000 | e<<51 | (c & (1<<51 - 1))}, true
		}
		return words{}, false
	case 128:
		if coeff.hi >= 1<<49 {
			return words{}, false
		}
		return words{lo: coeff.lo, hi: sign<<63 | e<<49 | coeff.hi}, true
	}
	return words{}, false
}

// flipSign toggles the width's sign bit of a materialized value.
func flipSign(w widthParams, v words, sign uint64) words {
	switch w.width {
	case 32:
		v.lo ^= sign << 31
	case 64:
		v.lo ^= sign << 63
	case 128:
		v.hi ^= sign << 63
	}
	return v
}

func formatValue(w widthParams, v words) string {
	switch w.width {
	case 32:
		return fmt.Sprintf("%08x", uint32(v.lo))
	case 64:
		return fmt.Sprintf("%016x", v.lo)
	default:
		return fmt.Sprintf("%016x:%016x", v.hi, v.lo)
	}
}

// ---------- boundary-bias operand pool ----------

// poolEntry is one boundary operand template. Finite entries keep their
// (coeff, biased exponent) so correlation can re-encode them at a retargeted
// exponent; special entries are fixed bit patterns. Signs are applied at
// draw time, so every entry covers both signs.
type poolEntry struct {
	finite bool
	coeff  words
	exp    int64 // biased
	bits   words // materialized with sign 0
}

// buildPool assembles the per-width boundary pool: coefficient extremes
// (including non-canonical probes) crossed with exponent extremes, plus
// special encodings (infinities, NaN payload variants, non-canonical
// steering patterns).
func buildPool(w widthParams) []poolEntry {
	p := int(w.precision)
	coeffs := []words{
		u128FromUint64(0),
		u128FromUint64(1),
		u128FromUint64(2),
		u128FromUint64(5),
		u128FromUint64(9),
		u128Sub64(pow10u128(p-1), 1),    // 10^(p-1)-1: largest (p-1)-digit value
		pow10u128(p - 1),                // 10^(p-1): smallest p-digit value
		u128MulSmall(pow10u128(p-1), 5), // 5*10^(p-1): nearest-even tie coefficient
		u128Sub64(pow10u128(p), 2),      // 10^p-2
		u128Sub64(pow10u128(p), 1),      // 10^p-1: max canonical coefficient
		pow10u128(p),                    // 10^p: non-canonical probe
	}

	exps := dedupInt64(clampInt64All([]int64{
		0, 1, 2,
		w.expBias - 2*w.precision, w.expBias - w.precision, w.expBias - 1,
		w.expBias, w.expBias + 1,
		w.expBias + w.precision - 1, w.expBias + w.precision, w.expBias + 2*w.precision,
		w.maxBiasedExp - w.precision, w.maxBiasedExp - 2, w.maxBiasedExp - 1, w.maxBiasedExp,
	}, 0, w.maxBiasedExp))

	var pool []poolEntry
	for _, c := range coeffs {
		for _, e := range exps {
			bits, ok := encodeBID(w, 0, e, c)
			if !ok {
				continue
			}
			pool = append(pool, poolEntry{finite: true, coeff: c, exp: e, bits: bits})
		}
	}
	pool = append(pool, buildSpecials(w)...)
	return pool
}

// u128MulSmall returns v * m for small m (never overflows in pool range).
func u128MulSmall(v words, m uint64) words {
	hi, lo := bits.Mul64(v.lo, m)
	return words{lo: lo, hi: v.hi*m + hi}
}

func buildSpecials(w widthParams) []poolEntry {
	fixed := func(v words) poolEntry { return poolEntry{bits: v} }
	switch w.width {
	case 32:
		payloadMax := uint64(999999) // 10^(p-1)-1: largest canonical NaN payload
		return []poolEntry{
			fixed(words{lo: 0x7800_0000}),                      // +Inf
			fixed(words{lo: 0x7800_0001}),                      // Inf, non-canonical trailing bits
			fixed(words{lo: 0x7810_0000}),                      // Inf, non-canonical exponent-area bits
			fixed(words{lo: 0x7c00_0000}),                      // qNaN
			fixed(words{lo: 0x7c00_0001}),                      // qNaN payload 1
			fixed(words{lo: 0x7c00_0000 | payloadMax}),         // qNaN max canonical payload
			fixed(words{lo: 0x7c0f_ffff}),                      // qNaN over-limit payload bits
			fixed(words{lo: 0x7e00_0000}),                      // sNaN
			fixed(words{lo: 0x7e00_0001}),                      // sNaN payload 1
			fixed(words{lo: 0x6c00_0000}),                      // steering-11 form, coefficient 2^23 (canonical probe)
			fixed(words{lo: 0x6000_0000 | 191<<21 | 0x1fffff}), // steering-11 non-canonical coefficient at max exponent
		}
	case 64:
		payloadMax := uint64(999_999_999_999_999) // 10^(p-1)-1
		return []poolEntry{
			fixed(words{lo: 0x7800_0000_0000_0000}),
			fixed(words{lo: 0x7800_0000_0000_0001}),
			fixed(words{lo: 0x7810_0000_0000_0000}),
			fixed(words{lo: 0x7c00_0000_0000_0000}),
			fixed(words{lo: 0x7c00_0000_0000_0001}),
			fixed(words{lo: 0x7c00_0000_0000_0000 | payloadMax}),
			fixed(words{lo: 0x7c03_ffff_ffff_ffff}),
			fixed(words{lo: 0x7e00_0000_0000_0000}),
			fixed(words{lo: 0x7e00_0000_0000_0001}),
			fixed(words{lo: 0x6c00_0000_0000_0000}),
			fixed(words{lo: 0x6000_0000_0000_0000 | 767<<51 | (1<<51 - 1)}),
		}
	default:
		nines33 := u128Sub64(pow10u128(33), 1)
		return []poolEntry{
			fixed(words{hi: 0x7800_0000_0000_0000}),
			fixed(words{hi: 0x7800_0000_0000_0000, lo: 1}),
			fixed(words{hi: 0x7810_0000_0000_0000}),
			fixed(words{hi: 0x7c00_0000_0000_0000}),
			fixed(words{hi: 0x7c00_0000_0000_0000, lo: 1}),
			fixed(words{hi: 0x7c00_0000_0000_0000 | nines33.hi, lo: nines33.lo}), // qNaN max canonical payload
			fixed(words{hi: 0x7c00_ffff_ffff_ffff, lo: ^uint64(0)}),              // qNaN over-limit payload bits
			fixed(words{hi: 0x7e00_0000_0000_0000}),
			fixed(words{hi: 0x7e00_0000_0000_0000, lo: 1}),
			fixed(words{hi: 0x6000_0000_0000_0000}),                             // steering-11: non-canonical zero
			fixed(words{hi: 0x6000_0000_0000_0000 | 12287<<47, lo: ^uint64(0)}), // steering-11 junk at max field bits
			fixed(words{hi: uint64(12287)<<49 | (1<<49 - 1), lo: ^uint64(0)}),   // normal form, coefficient 2^113-1 (non-canonical)
		}
	}
}

func clampInt64All(vs []int64, lo, hi int64) []int64 {
	out := make([]int64, 0, len(vs))
	for _, v := range vs {
		if v < lo {
			v = lo
		}
		if v > hi {
			v = hi
		}
		out = append(out, v)
	}
	return out
}

func dedupInt64(vs []int64) []int64 {
	seen := make(map[int64]bool, len(vs))
	out := make([]int64, 0, len(vs))
	for _, v := range vs {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

// ---------- operations, modes, targets ----------

var allOps = []string{"add", "sub", "mul", "div", "fma", "sqrt", "quantize"}

func opArity(op string) int {
	switch op {
	case "sqrt":
		return 1
	case "fma":
		return 3
	default:
		return 2
	}
}

type modeSpec struct {
	name   string
	native int // Intel BID_ROUNDING_* value
}

var allModes = []modeSpec{
	{"nearest_even", 0},
	{"toward_negative", 1},
	{"toward_positive", 2},
	{"toward_zero", 3},
	{"nearest_away", 4},
}

func modeByName(name string) (modeSpec, bool) {
	for _, m := range allModes {
		if m.name == name {
			return m, true
		}
	}
	return modeSpec{}, false
}

// targetSeed derives the per-(width, op) stream seed from the run seed.
func targetSeed(runSeed uint64, w widthParams, op string) uint64 {
	return runSeed ^ fnv64a(fmt.Sprintf("d%d/%s", w.width, op))
}

// ---------- tuple drawing ----------

type drawnOperand struct {
	bits  words
	entry *poolEntry
	sign  uint64
}

func (o drawnOperand) finiteParam() bool { return o.entry != nil && o.entry.finite }

func rawRandom(rng *splitMix64, w widthParams) words {
	switch w.width {
	case 32:
		return words{lo: rng.next() & 0xffff_ffff}
	case 64:
		return words{lo: rng.next()}
	default:
		return words{lo: rng.next(), hi: rng.next()}
	}
}

// drawTuple draws one operand tuple for (width, op). With probability bias
// the case is boundary-biased: operands come mostly from the pool (sign
// randomized), and exponents of later slots are retargeted with probability
// 1/2 into the op's interaction window (alignment span for add/sub/quantize,
// overflow/underflow cap span for mul/div and the fma product, and a
// product-adjacent span for the fma addend) so carries, clamps, and sticky
// paths at the format boundaries are exercised far more often than uniform
// bits would manage.
func drawTuple(rng *splitMix64, w widthParams, op string, bias float64, pool []poolEntry) [3]words {
	arity := opArity(op)
	var ops [3]drawnOperand
	biased := rng.chance(bias)
	for i := 0; i < arity; i++ {
		if !biased || rng.chance(0.25) {
			ops[i] = drawnOperand{bits: rawRandom(rng, w)}
			continue
		}
		entry := &pool[rng.intn(len(pool))]
		sign := rng.next() & 1
		ops[i] = drawnOperand{bits: flipSign(w, entry.bits, sign), entry: entry, sign: sign}
	}
	if biased {
		applyExponentCorrelation(rng, w, op, &ops)
	}
	return [3]words{ops[0].bits, ops[1].bits, ops[2].bits}
}

// applyExponentCorrelation retargets the biased exponent of the second (and
// for fma, third) operand relative to the first so the tuple lands in the
// op's boundary-interaction window. Only finite pool-derived operands are
// retargeted; specials and raw-random operands stay as drawn.
func applyExponentCorrelation(rng *splitMix64, w widthParams, op string, ops *[3]drawnOperand) {
	p := w.precision
	minU := -w.expBias
	maxU := w.maxBiasedExp - w.expBias
	unbiased := func(o drawnOperand) int64 { return o.entry.exp - w.expBias }
	retarget := func(i int, eU int64) {
		if eU < minU {
			eU = minU
		}
		if eU > maxU {
			eU = maxU
		}
		bits, ok := encodeBID(w, ops[i].sign, eU+w.expBias, ops[i].entry.coeff)
		if !ok {
			return
		}
		ops[i].bits = bits
		entry := *ops[i].entry
		entry.exp = eU + w.expBias
		ops[i].entry = &entry
	}
	// The product's result exponent is the operand exponent sum plus the
	// coefficient-digit carry (0..p digits beyond one-operand precision), so
	// the overflow crossing region for the SUM spans [maxU-2p, maxU+p].
	capSpan := func() int64 {
		if rng.next()&1 == 0 {
			return rng.span(maxU-2*p, maxU+p+2) // overflow cap crossing region
		}
		return rng.span(minU-2*p, minU+p) // underflow / subnormal crossing region
	}
	switch op {
	case "add", "sub", "quantize":
		if ops[0].finiteParam() && ops[1].finiteParam() && rng.chance(0.5) {
			retarget(1, unbiased(ops[0])+rng.span(-(p+2), p+2))
		}
	case "mul", "div", "fma":
		if ops[0].finiteParam() && ops[1].finiteParam() && rng.chance(0.5) {
			t := capSpan()
			if op == "div" {
				retarget(1, unbiased(ops[0])-t)
			} else {
				retarget(1, t-unbiased(ops[0]))
			}
		}
		if op == "fma" && ops[0].finiteParam() && ops[1].finiteParam() &&
			ops[2].finiteParam() && rng.chance(0.5) {
			sum := unbiased(ops[0]) + unbiased(ops[1])
			retarget(2, sum+rng.span(-(2*p+2), 2*p+2))
		}
	}
}
