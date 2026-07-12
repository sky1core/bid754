// Hand-written architecture-contract gate (the same standing as
// bid754-go/readtest_comparator_strength_test.go): it lives OUTSIDE every
// generation path and must remain hand-written and unchanged by any
// emitter.
//
// Why this gate exists: some flagless port entrypoints are separately ported
// full implementations, not thin wrappers over their WithFlags siblings
// (BID32 add/sub/mul/div and BID64 mul/div). The goport readtest and FFI
// bit-compare gates exercise the WithFlags variants, so the flagless bodies -
// which the public value-only wrappers actually route through - would
// otherwise run under no oracle at all. This gate closes that hole
// transitively: it asserts bit-exact value agreement flagless(x, y, mode) ==
// value(WithFlags(x, y, mode)) over pinned deterministic corpora (edge bit
// patterns crossed both ways plus seeded pseudo-random pairs, all five
// rounding modes), and the WithFlags side is oracle-verified by the regular
// domains.
//
// The target list is exhaustive: devtools/internal/publicroute cross-checks
// (by AST) that every port function it exempts as
// "sibling-equivalence-covered" appears here as a {flagless, WithFlags} pair,
// so removing a pair from this list fails the publicroute gate, and a new
// separately-ported flagless base reached by a public wrapper fails
// publicroute until it is added here.
package bidgo

import (
	"math/rand"
	"testing"
)

// flaglessVariantEquivalenceTargets32 pins the separately ported flagless
// BID32 entrypoints and their oracle-verified WithFlags siblings. Both
// identifiers must be the real port functions; publicroute verifies each pair
// is present by name.
var flaglessVariantEquivalenceTargets32 = []struct {
	name      string
	flagless  func(x, y uint32, rndMode int) uint32
	withFlags func(x, y uint32, rndMode int) (uint32, uint32)
}{
	{name: "Bid32Add", flagless: Bid32Add, withFlags: Bid32AddWithFlags},
	{name: "Bid32Sub", flagless: Bid32Sub, withFlags: Bid32SubWithFlags},
	{name: "Bid32Mul", flagless: Bid32Mul, withFlags: Bid32MulWithFlags},
	{name: "Bid32Div", flagless: Bid32Div, withFlags: Bid32DivWithFlags},
}

// flaglessVariantEquivalenceTargets64 pins the separately ported flagless
// BID64 entrypoints and their oracle-verified WithFlags siblings. Both
// identifiers must be the real port functions; publicroute verifies the pair
// is present by name.
var flaglessVariantEquivalenceTargets64 = []struct {
	name      string
	flagless  func(x, y uint64, rndMode int) uint64
	withFlags func(x, y uint64, rndMode int) (uint64, uint32)
}{
	{name: "Bid64Div", flagless: Bid64Div, withFlags: Bid64DivWithFlags},
	{name: "Bid64Mul", flagless: Bid64Mul, withFlags: Bid64MulWithFlags},
}

// flaglessEquivalenceEdgeBits32 are exact BID32 bit patterns covering the same
// semantic categories as the BID64 corpus below: signed zeros, finite and
// subnormal boundaries, steering encodings, infinities, and quiet/signaling
// NaNs including payload and noncanonical forms. Every pattern feeds both
// operand positions.
var flaglessEquivalenceEdgeBits32 = []uint32{
	0x32800000, // +0 (canonical, exponent bias)
	0xb2800000, // -0
	0x00000000, // +0 at the minimum exponent
	0x32800001, // +1
	0xb2800001, // -1
	0x3280007b, // +123
	0x3180007b, // +123 at a smaller exponent (1.23)
	0x40800001, // +1e28 (large positive exponent)
	0x24800001, // +1e-28 (large negative exponent)
	0x77f8967f, // +9999999e90 (finite maximum)
	0xf7f8967f, // finite minimum (negative maximum magnitude)
	0x00000001, // +1e-101 (smallest subnormal)
	0x80000001, // -1e-101
	0x6cb8967f, // steered large-coefficient encoding
	0x601fffff, // steered encoding with out-of-range coefficient (noncanonical)
	0x78000000, // +Inf (canonical)
	0xf8000000, // -Inf
	0x780dead0, // +Inf with noncanonical trailing payload bits
	0x7c000000, // +qNaN
	0xfc000000, // -qNaN
	0x7c000abc, // +qNaN with payload
	0x7c1fffff, // +qNaN with out-of-range (noncanonical) payload
	0x7e000000, // +sNaN
	0xfe000000, // -sNaN
	0x7e000abc, // +sNaN with payload
}

// flaglessEquivalenceEdgeBits64 are exact BID64 bit patterns covering signed
// zeros, small/large finites at several exponents, subnormal boundaries, the
// finite maxima, infinities (canonical and payload-bearing), quiet and
// signaling NaNs (with and without payloads, including a noncanonical
// payload), and noncanonical steered encodings. Every pattern feeds both
// operand positions.
var flaglessEquivalenceEdgeBits64 = []uint64{
	0x31c0000000000000,    // +0 (canonical, exponent bias)
	0xb1c0000000000000,    // -0
	0x0000000000000000,    // +0 at the minimum exponent
	0x31c0000000000001,    // +1
	0xb1c0000000000001,    // -1
	0x31c000000000007b,    // +123
	0x31a000000000007b,    // +123 at a smaller exponent (1.23)
	0x3540000000000001,    // +1e28 (large positive exponent)
	0x2e40000000000001,    // +1e-28 (large negative exponent)
	0x77fb86f26fc0ffff,    // +9999999999999999e369 (finite maximum)
	0xf7fb86f26fc0ffff,    // finite minimum (negative maximum magnitude)
	0x0000000000000001,    // +1e-398 (smallest subnormal)
	0x8000000000000001,    // -1e-398
	0x6c73_86f2_6fc0_ffff, // steered (11 steering bits) large-coefficient encoding
	0x6003ffffffffffff,    // steered encoding with out-of-range coefficient (noncanonical)
	0x7800000000000000,    // +Inf (canonical)
	0xf800000000000000,    // -Inf
	0x7800dead00000000,    // +Inf with noncanonical trailing payload bits
	0x7c00000000000000,    // +qNaN
	0xfc00000000000000,    // -qNaN
	0x7c00000000000abc,    // +qNaN with payload
	0x7c03ffffffffffff,    // +qNaN with out-of-range (noncanonical) payload
	0x7e00000000000000,    // +sNaN
	0xfe00000000000000,    // -sNaN
	0x7e00000000000abc,    // +sNaN with payload
}

// Deterministic pseudo-random supplement: both operands vary independently,
// drawn from a fixed-seed source so every run executes the identical corpus.
const (
	flaglessEquivalenceRandomPairs32 = 1 << 20
	flaglessEquivalenceRandomPairs64 = 4096
	flaglessEquivalenceSeed32        = 0x754d32f1a95
	flaglessEquivalenceSeed64        = 0x754d64f1a95
)

// flaglessEquivalenceRoundingModes are the five Intel BID rounding-mode values
// (BID_ROUNDING_TO_NEAREST..BID_ROUNDING_TIES_AWAY); every pair runs under all
// of them.
var flaglessEquivalenceRoundingModes = []int{0, 1, 2, 3, 4}

func TestFlaglessVariantsMatchWithFlagsValues(t *testing.T) {
	t.Run("bid32", func(t *testing.T) {
		if len(flaglessVariantEquivalenceTargets32) == 0 {
			t.Fatal("BID32 flagless variant equivalence target list is empty; the gate lost its subject")
		}

		expectedPairs := len(flaglessEquivalenceEdgeBits32)*len(flaglessEquivalenceEdgeBits32) + flaglessEquivalenceRandomPairs32
		expected := len(flaglessVariantEquivalenceTargets32) * expectedPairs * len(flaglessEquivalenceRoundingModes)
		executed := 0
		for _, target := range flaglessVariantEquivalenceTargets32 {
			check := func(x, y uint32) {
				for _, mode := range flaglessEquivalenceRoundingModes {
					got := target.flagless(x, y, mode)
					want, _ := target.withFlags(x, y, mode)
					executed++
					if got != want {
						t.Fatalf("%s(%#010x, %#010x, mode %d) = %#010x, but %sWithFlags value = %#010x: the separately ported flagless implementation diverged from its oracle-verified WithFlags sibling",
							target.name, x, y, mode, got, target.name, want)
					}
				}
			}
			for _, x := range flaglessEquivalenceEdgeBits32 {
				for _, y := range flaglessEquivalenceEdgeBits32 {
					check(x, y)
				}
			}
			rng := rand.New(rand.NewSource(flaglessEquivalenceSeed32))
			for i := 0; i < flaglessEquivalenceRandomPairs32; i++ {
				check(rng.Uint32(), rng.Uint32())
			}
		}
		if executed == 0 {
			t.Fatal("BID32 flagless variant equivalence executed zero cases; the corpus is empty")
		}
		if executed != expected {
			t.Fatalf("BID32 flagless variant equivalence executed %d cases, expected %d", executed, expected)
		}
	})

	t.Run("bid64", func(t *testing.T) {
		if len(flaglessVariantEquivalenceTargets64) == 0 {
			t.Fatal("BID64 flagless variant equivalence target list is empty; the gate lost its subject")
		}

		expectedPairs := len(flaglessEquivalenceEdgeBits64)*len(flaglessEquivalenceEdgeBits64) + flaglessEquivalenceRandomPairs64
		expected := len(flaglessVariantEquivalenceTargets64) * expectedPairs * len(flaglessEquivalenceRoundingModes)
		executed := 0
		for _, target := range flaglessVariantEquivalenceTargets64 {
			check := func(x, y uint64) {
				for _, mode := range flaglessEquivalenceRoundingModes {
					got := target.flagless(x, y, mode)
					want, _ := target.withFlags(x, y, mode)
					executed++
					if got != want {
						t.Fatalf("%s(%#018x, %#018x, mode %d) = %#018x, but %sWithFlags value = %#018x: the separately ported flagless implementation diverged from its oracle-verified WithFlags sibling",
							target.name, x, y, mode, got, target.name, want)
					}
				}
			}
			for _, x := range flaglessEquivalenceEdgeBits64 {
				for _, y := range flaglessEquivalenceEdgeBits64 {
					check(x, y)
				}
			}
			rng := rand.New(rand.NewSource(flaglessEquivalenceSeed64))
			for i := 0; i < flaglessEquivalenceRandomPairs64; i++ {
				check(rng.Uint64(), rng.Uint64())
			}
		}
		if executed == 0 {
			t.Fatal("BID64 flagless variant equivalence executed zero cases; the corpus is empty")
		}
		if executed != expected {
			t.Fatalf("BID64 flagless variant equivalence executed %d cases, expected %d", executed, expected)
		}
	})
}
