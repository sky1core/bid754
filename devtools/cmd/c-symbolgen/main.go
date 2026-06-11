package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sky1core/bid754/internal/csymbols"
)

func main() {
	manifestPath := flag.String("manifest", "symbolgen_manifest.json", "path to the symbol generation manifest")
	flag.Parse()

	repoRoot, err := os.Getwd()
	if err != nil {
		exitf("get working directory: %v", err)
	}
	manifestAbs := filepath.Join(repoRoot, *manifestPath)
	manifest, err := csymbols.LoadManifest(manifestAbs)
	if err != nil {
		exitf("%v", err)
	}

	generated, err := csymbols.Generate(repoRoot, manifest)
	if err != nil {
		exitf("%v", err)
	}
	if err := csymbols.WriteOutputs(repoRoot, manifest, generated); err != nil {
		exitf("%v", err)
	}
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
