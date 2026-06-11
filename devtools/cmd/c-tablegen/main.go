package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sky1core/bid754/devtools/internal/cgen"
)

func main() {
	manifestPath := flag.String("manifest", "tablegen_manifest.json", "path to the table generation manifest")
	flag.Parse()

	repoRoot, err := os.Getwd()
	if err != nil {
		exitf("get working directory: %v", err)
	}
	manifestAbs := filepath.Join(repoRoot, *manifestPath)
	manifest, err := cgen.LoadManifest(manifestAbs)
	if err != nil {
		exitf("%v", err)
	}

	generated, err := cgen.Generate(repoRoot, manifest)
	if err != nil {
		exitf("%v", err)
	}
	if err := cgen.WriteOutputs(repoRoot, manifest, generated); err != nil {
		exitf("%v", err)
	}
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
