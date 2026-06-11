package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sky1core/bid754/internal/readtestspec"
)

func main() {
	root := findProjectRoot()
	registry, err := readtestspec.LoadFromProjectRoot(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load readtest specs: %v\n", err)
		os.Exit(1)
	}

	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal readtest specs: %v\n", err)
		os.Exit(1)
	}
	data = append(data, '\n')
	if _, err := os.Stdout.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "write readtest specs: %v\n", err)
		os.Exit(1)
	}
}

func findProjectRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			fmt.Fprintln(os.Stderr, "cannot find project root")
			os.Exit(1)
		}
		dir = parent
	}
}
