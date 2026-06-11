package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/sky1core/bid754/internal/testgen"
)

func main() {
	manifestPath := flag.String("manifest", "testgen_manifest.json", "path to the shared test generation manifest")
	flag.Parse()

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
	if err := testgen.WriteFFITestOutputs(repoRoot, spec); err != nil {
		log.Fatalf("write generated ffi test: %v", err)
	}
	if err := testgen.WriteDectestTestOutputs(repoRoot, spec); err != nil {
		log.Fatalf("write generated dectest test: %v", err)
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
}
