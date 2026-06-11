package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sky1core/bid754/devtools/internal/gotypegen"
)

func main() {
	manifestPath := flag.String("manifest", "typegen_manifest.json", "path to the Go type generation manifest")
	flag.Parse()

	repoRoot, err := os.Getwd()
	if err != nil {
		exitf("get working directory: %v", err)
	}
	manifestAbs := filepath.Join(repoRoot, *manifestPath)
	manifest, err := gotypegen.LoadManifest(manifestAbs)
	if err != nil {
		exitf("%v", err)
	}

	generated, err := gotypegen.Generate(manifest)
	if err != nil {
		exitf("%v", err)
	}
	if err := gotypegen.WriteOutput(repoRoot, manifest, generated); err != nil {
		exitf("%v", err)
	}
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
