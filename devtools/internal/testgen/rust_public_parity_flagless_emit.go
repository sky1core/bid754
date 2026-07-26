package testgen

import (
	"fmt"
	"strings"
)

// ---- generated flagless-sibling equivalence leg (Rust public parity gate) ----
//
// Rust mirror of the Go leg emitted by emitPublicParityFlaglessSibling
// (public_parity_runner_emit.go). Both legs are emitted from the same
// generator-owned tables in mutation_witness_corpus.go
// (parityFlaglessSiblingTargets, parityFlaglessWitnessRows, the seeds and the
// random-pair counts) and from the same shared bit-literal corpus
// (buildPublicParityCorpus), so the two languages run the identical operand
// stream by construction rather than by two hand-kept copies.
//
// Why the Rust leg is needed at all: the go2rs-generated flagless port bodies
// have no oracle in the Rust regular chain either. The generated Rust readtest
// runner dispatches the *_with_flags variant of these operations (readtest
// compares status flags), and the Rust public wrapper for the value-only
// method calls the very same flagless port function this gate's left-hand side
// calls -- so the wrapper-vs-port parity check is tautological for that body.
// This leg pins flagless(x, y, mode) == value(with_flags(x, y, mode)) on the
// generated Rust side, which is the same contract the Go leg pins for the
// mechanical port.
//
// The leg is accounted separately from EXPECTED_PARITY_CASES (like the
// trait-parity, NaN round-trip, and constants legs) with its own
// EXPECTED_FLAGLESS_SIBLING_* self-checks, so the wrapper-census anchor stays
// exactly traceable to rust_api_surface_inventory.json's emitted row count.

// rustFlaglessWidth carries the per-width literals the emitted Rust needs. Every
// field is an explicit literal (never derived by string-splicing at a call
// site), matching the parityWidth convention in rust_public_parity_emit.go.
type rustFlaglessWidth struct {
	Width       int
	UintType    string
	Corpus      string
	OpFmt       string
	RandomPairs int
	Seed        uint64
}

var rustFlaglessWidths = []rustFlaglessWidth{
	{Width: 32, UintType: "u32", Corpus: "CORPUS_32", OpFmt: "{:#010x}", RandomPairs: parityFlaglessRandomPairs32, Seed: parityFlaglessSeed32},
	{Width: 64, UintType: "u64", Corpus: "CORPUS_64", OpFmt: "{:#018x}", RandomPairs: parityFlaglessRandomPairs64, Seed: parityFlaglessSeed64},
}

// witnessLiteral renders a witness operand at the width's hex literal shape so
// the Rust rows read the same as the Go leg's.
func (w rustFlaglessWidth) witnessLiteral(v uint64) string {
	if w.Width == 32 {
		return fmt.Sprintf("0x%08x", uint32(v))
	}
	return fmt.Sprintf("0x%016x", v)
}

// emitRustParityFlaglessSibling renders the Rust flagless-sibling equivalence
// leg. Port addresses are resolved through apiemit.PortPathFor (resolvePort),
// the same routing metadata the wrapper emitter used, so a missing table entry
// is a generation-time failure rather than a silently skipped target. The
// function-pointer field types in the emitted target tables are the structural
// check that each resolved pair really has the flagless/with-flags shape: a
// port function whose Rust signature drifts fails to compile here.
func emitRustParityFlaglessSibling(b *strings.Builder) error {
	widthOfTarget := map[string]rustFlaglessWidth{}
	for _, w := range rustFlaglessWidths {
		if err := emitRustFlaglessTargetTable(b, w, widthOfTarget); err != nil {
			return err
		}
	}
	for _, target := range parityFlaglessSiblingTargets {
		if _, ok := widthOfTarget[target.Flagless]; !ok {
			return fmt.Errorf("rust public parity flagless leg: target %q has width %d, which has no emitted Rust width record; add it to rustFlaglessWidths rather than dropping the target", target.Flagless, target.Width)
		}
	}
	for _, row := range parityFlaglessWitnessRows {
		if _, ok := widthOfTarget[row.Target]; !ok {
			return fmt.Errorf("rust public parity flagless leg: witness row %s references undeclared target %q", row.MutantID, row.Target)
		}
	}
	for _, w := range rustFlaglessWidths {
		emitRustFlaglessWitnessTable(b, w, widthOfTarget)
	}
	for _, w := range rustFlaglessWidths {
		emitRustFlaglessCheckFn(b, w)
	}
	emitRustFlaglessDriver(b)
	return nil
}

func emitRustFlaglessTargetTable(b *strings.Builder, w rustFlaglessWidth, widthOfTarget map[string]rustFlaglessWidth) error {
	fmt.Fprintf(b, `
/// One flagless/with-flags sibling pair of the generated Rust port at width
/// %d. The two function-pointer field types are the leg's structural
/// signature check: a port function that stops matching the flagless or the
/// with-flags shape fails to compile in the table below.
struct FlaglessSiblingTarget%d {
    name: &'static str,
    flagless: fn(%s, %s, i64) -> %s,
    with_flags: fn(%s, %s, i64) -> (%s, u32),
}

`, w.Width, w.Width, w.UintType, w.UintType, w.UintType, w.UintType, w.UintType, w.UintType)

	fmt.Fprintf(b, "const FLAGLESS_SIBLING_TARGETS_%d: &[FlaglessSiblingTarget%d] = &[\n", w.Width, w.Width)
	for _, target := range parityFlaglessSiblingTargets {
		if target.Width != w.Width {
			continue
		}
		flaglessModule, flaglessFn, err := resolvePort(target.Flagless, "flagless-sibling equivalence leg")
		if err != nil {
			return err
		}
		withFlagsModule, withFlagsFn, err := resolvePort(target.WithFlags, "flagless-sibling equivalence leg")
		if err != nil {
			return err
		}
		widthOfTarget[target.Flagless] = w
		fmt.Fprintf(b, `    // %s
    FlaglessSiblingTarget%d {
        name: %q,
        flagless: bid754::generated::%s::%s,
        with_flags: bid754::generated::%s::%s,
    },
`, target.Reason, w.Width, target.Flagless, flaglessModule, flaglessFn, withFlagsModule, withFlagsFn)
	}
	b.WriteString("];\n\n")
	return nil
}

func emitRustFlaglessWitnessTable(b *strings.Builder, w rustFlaglessWidth, widthOfTarget map[string]rustFlaglessWidth) {
	fmt.Fprintf(b, `/// Pinned distinguishing inputs of the mutation-audit survivors that lived on
/// these flagless bodies (mutation_witness_corpus.go). They are re-run here on
/// the generated Rust port so a witness stays exercised in both languages.
struct FlaglessWitnessRow%d {
    target: &'static str,
    x: %s,
    y: %s,
    mode: i64,
}

`, w.Width, w.UintType, w.UintType)

	fmt.Fprintf(b, "const FLAGLESS_WITNESS_ROWS_%d: &[FlaglessWitnessRow%d] = &[\n", w.Width, w.Width)
	for _, row := range parityFlaglessWitnessRows {
		if widthOfTarget[row.Target].Width != w.Width {
			continue
		}
		fmt.Fprintf(b, "    FlaglessWitnessRow%d { target: %q, x: %s, y: %s, mode: %d }, // mutant %s\n",
			w.Width, row.Target, w.witnessLiteral(row.X), w.witnessLiteral(row.Y), row.Mode, row.MutantID)
	}
	b.WriteString("];\n\n")
}

func emitRustFlaglessCheckFn(b *strings.Builder, w rustFlaglessWidth) {
	fmt.Fprintf(b, `fn flagless_sibling_check_%d(target: &FlaglessSiblingTarget%d, x: %s, y: %s, mode: i64) {
    let got = (target.flagless)(x, y, mode);
    let (want, _) = (target.with_flags)(x, y, mode);
    if got != want {
        panic!(
            "public parity flagless sibling {}(%s, %s, mode {}) = %s, want with-flags value %s",
            target.name, x, y, mode, got, want
        );
    }
}

`, w.Width, w.Width, w.UintType, w.UintType, w.OpFmt, w.OpFmt, w.OpFmt, w.OpFmt)
}

func emitRustFlaglessDriver(b *strings.Builder) {
	b.WriteString(`/// flagless_sibling_next is the same splitmix-style deterministic stream the Go
/// leg emits (publicParityFlaglessNext), so both languages consume the identical
/// pseudo-random operand sequence for a given seed and pair count.
fn flagless_sibling_next(state: &mut u64) -> u64 {
    *state = state.wrapping_add(0x9e37_79b9_7f4a_7c15);
    let mut z = *state;
    z ^= z >> 30;
    z = z.wrapping_mul(0xbf58_476d_1ce4_e5b9);
    z ^= z >> 27;
    z = z.wrapping_mul(0x94d0_49bb_1331_11eb);
    z ^= z >> 31;
    z
}

`)

	fmt.Fprintf(b, `/// The flagless-sibling equivalence leg: the %d separately generated flagless
/// port bodies have no direct oracle in the Rust regular chain (the generated
/// readtest runner dispatches their *_with_flags siblings, and the value-only
/// public wrapper calls the very flagless function under test), so this leg
/// pins flagless(x, y, mode) == value(with_flags(x, y, mode)) bit-exactly over
/// the shared parity corpus crossed both ways, a seeded pseudo-random
/// supplement, and the pinned mutation-audit witness rows. Mirrors
/// bid754-go/generated_public_parity_cases_test.go's
/// TestGeneratedPublicAPIFlaglessSiblingEquivalence, and both legs are emitted
/// from the same generator tables so their case counts are equal by
/// construction. Accounted separately from EXPECTED_PARITY_CASES.
const EXPECTED_FLAGLESS_SIBLING_TARGETS: usize = %d;
const EXPECTED_FLAGLESS_SIBLING_CASES: usize = %d;

#[test]
fn generated_public_api_flagless_sibling_equivalence() {
`, len(parityFlaglessSiblingTargets), len(parityFlaglessSiblingTargets), parityFlaglessSiblingCaseCount())

	var lenTerms []string
	for _, w := range rustFlaglessWidths {
		lenTerms = append(lenTerms, fmt.Sprintf("FLAGLESS_SIBLING_TARGETS_%d.len()", w.Width))
	}
	fmt.Fprintf(b, `    assert_eq!(
        %s,
        EXPECTED_FLAGLESS_SIBLING_TARGETS,
        "flagless-sibling target census drifted"
    );
    let mut count = 0usize;
`, strings.Join(lenTerms, " + "))

	for _, w := range rustFlaglessWidths {
		fmt.Fprintf(b, `    for target in FLAGLESS_SIBLING_TARGETS_%d {
        for &x in %s {
            for &y in %s {
                for mode in 0..%d {
                    flagless_sibling_check_%d(target, x, y, mode);
                    count += 1;
                }
            }
        }
        let mut state: u64 = %#x;
        for _ in 0..%d {
            let x = flagless_sibling_next(&mut state) as %s;
            let y = flagless_sibling_next(&mut state) as %s;
            for mode in 0..%d {
                flagless_sibling_check_%d(target, x, y, mode);
                count += 1;
            }
        }
    }
`, w.Width, w.Corpus, w.Corpus, parityFlaglessModeCount, w.Width, w.Seed, w.RandomPairs, w.UintType, w.UintType, parityFlaglessModeCount, w.Width)
	}

	for _, w := range rustFlaglessWidths {
		fmt.Fprintf(b, `    for row in FLAGLESS_WITNESS_ROWS_%d {
        let target = FLAGLESS_SIBLING_TARGETS_%d
            .iter()
            .find(|t| t.name == row.target)
            .unwrap_or_else(|| panic!("flagless witness row targets unknown function {:?}", row.target));
        flagless_sibling_check_%d(target, row.x, row.y, row.mode);
        count += 1;
    }
`, w.Width, w.Width, w.Width)
	}

	b.WriteString(`    assert_eq!(
        count, EXPECTED_FLAGLESS_SIBLING_CASES,
        "flagless-sibling case count drifted"
    );
}

`)
}
