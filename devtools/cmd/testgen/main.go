package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/sky1core/bid754/devtools/internal/testgen"
)

func main() {
	manifestPath := flag.String("manifest", "testgen_manifest.json", "path to the shared test generation manifest")
	printSentinelAnchors := flag.Bool("print-sentinel-anchors", false,
		"print the proposed verification_sentinels.json routing-sentinel rows to stdout and exit; writes no file and reads no anchor")
	printDecnumberSentinelAnchors := flag.Bool("print-decnumber-sentinel-anchors", false,
		"print the proposed verification_sentinels.json decNumber differential sentinel rows to stdout and exit; writes no file and reads no anchor")
	printD32ExhaustiveSentinelAnchors := flag.Bool("print-d32-exhaustive-sentinel-anchors", false,
		"print the proposed verification_sentinels.json d32 exhaustive sentinel rows to stdout and exit; writes no file and reads no anchor")
	printMixedFFISentinelAnchors := flag.Bool("print-mixed-ffi-sentinel-anchors", false,
		"print the proposed verification_sentinels.json mixed-format FFI routing sentinel rows to stdout and exit; writes no file and reads no anchor")
	flag.Parse()

	if *printSentinelAnchors {
		proposal, err := testgen.Tier1SentinelAnchorProposal()
		if err != nil {
			log.Fatalf("compute sentinel anchor proposal: %v", err)
		}
		fmt.Print(proposal)
		return
	}
	if *printDecnumberSentinelAnchors {
		proposal, err := testgen.DecnumberDifferentialSentinelAnchorProposal()
		if err != nil {
			log.Fatalf("compute decNumber differential sentinel anchor proposal: %v", err)
		}
		fmt.Print(proposal)
		return
	}
	if *printD32ExhaustiveSentinelAnchors {
		proposal, err := testgen.D32ExhaustiveSentinelAnchorProposal()
		if err != nil {
			log.Fatalf("compute d32 exhaustive sentinel anchor proposal: %v", err)
		}
		fmt.Print(proposal)
		return
	}
	if *printMixedFFISentinelAnchors {
		proposal, err := testgen.MixedFFIRoutingSentinelAnchorProposal()
		if err != nil {
			log.Fatalf("compute mixed-format FFI routing sentinel anchor proposal: %v", err)
		}
		fmt.Print(proposal)
		return
	}

	repoRoot, err := filepath.Abs(".")
	if err != nil {
		log.Fatalf("resolve repo root: %v", err)
	}
	manifestAbs := filepath.Join(repoRoot, *manifestPath)
	manifest, err := testgen.LoadManifest(manifestAbs)
	if err != nil {
		log.Fatalf("load manifest: %v", err)
	}
	spec, err := testgen.Generate(repoRoot, manifest)
	if err != nil {
		log.Fatalf("generate shared test spec: %v", err)
	}
	if err := testgen.WriteOutput(repoRoot, manifest, spec); err != nil {
		log.Fatalf("write generated spec: %v", err)
	}
	if err := testgen.WriteReadtestDispatchOutputs(repoRoot, manifest); err != nil {
		log.Fatalf("write generated readtest dispatch: %v", err)
	}
	if err := testgen.WriteReadtestTestOutputs(repoRoot, spec); err != nil {
		log.Fatalf("write generated readtest test: %v", err)
	}
	if err := testgen.WriteReadtestGoportOutputs(repoRoot, manifest, spec); err != nil {
		log.Fatalf("write generated readtest goport outputs: %v", err)
	}
	// Write the decTest executor/dispatch set before generators that compile the
	// sibling bid754-go module. Removing a generated executor and its dispatch
	// row must be able to bootstrap from the previous checked-in artifact set.
	if err := testgen.WriteDectestTestOutputs(repoRoot, spec); err != nil {
		log.Fatalf("write generated dectest test: %v", err)
	}
	if err := testgen.WriteDectestGoportOutputs(repoRoot, spec); err != nil {
		log.Fatalf("write generated dectest goport outputs: %v", err)
	}
	if err := testgen.WriteDectestRustOutputs(repoRoot, spec); err != nil {
		log.Fatalf("write generated dectest rust outputs: %v", err)
	}
	if err := testgen.WritePublicParityOutputs(repoRoot, manifest); err != nil {
		log.Fatalf("write generated public parity outputs: %v", err)
	}
	if err := testgen.WriteRustPublicParityOutputs(repoRoot); err != nil {
		log.Fatalf("write generated rust public parity outputs: %v", err)
	}
	if err := testgen.WriteFFITestOutputs(repoRoot, spec); err != nil {
		log.Fatalf("write generated ffi test: %v", err)
	}
	if err := testgen.WriteTier1ArithmeticLongOutputs(repoRoot); err != nil {
		log.Fatalf("write generated Tier 1 arithmetic long test: %v", err)
	}
	if err := testgen.WriteTier1CompareConversionLongOutputs(repoRoot); err != nil {
		log.Fatalf("write generated Tier 1 compare/conversion long test: %v", err)
	}
	if err := testgen.WriteDecnumberDifferentialOutputs(repoRoot, manifest); err != nil {
		log.Fatalf("write generated decNumber differential outputs: %v", err)
	}
	if err := testgen.WriteD32ExhaustiveOutputs(repoRoot); err != nil {
		log.Fatalf("write generated d32 exhaustive outputs: %v", err)
	}
	if err := testgen.WriteBidCodecVectorDataOutput(repoRoot, *manifest.BidCodecVectors); err != nil {
		log.Fatalf("write generated BID codec vectors: %v", err)
	}
	if err := testgen.WriteBidCodecVectorTestOutputs(repoRoot); err != nil {
		log.Fatalf("write generated BID codec vector tests: %v", err)
	}
	if err := testgen.WriteBidStringVectorTestOutputs(repoRoot, spec); err != nil {
		log.Fatalf("write generated BID string vector tests: %v", err)
	}
	if err := testgen.WriteTestspecPackageOutputs(repoRoot); err != nil {
		log.Fatalf("write generated testspec package: %v", err)
	}
}
