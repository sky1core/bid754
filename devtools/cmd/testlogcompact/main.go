// testlogcompact filters verbose Go test output without hiding failures.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/sky1core/bid754/devtools/internal/testlogcompact"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, input io.Reader, output, errorOutput io.Writer) int {
	flags := flag.NewFlagSet("testlogcompact", flag.ContinueOnError)
	flags.SetOutput(errorOutput)
	rootTest := flags.String("root", "", "top-level Go test whose successful/skipped subtest lifecycle lines should be suppressed")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(errorOutput, "testlogcompact: positional arguments are not supported")
		return 2
	}

	stats, err := testlogcompact.Filter(input, output, *rootTest)
	if err != nil {
		fmt.Fprintf(errorOutput, "testlogcompact: %v\n", err)
		return 1
	}
	if _, err := fmt.Fprintf(
		output,
		"testlogcompact: suppressed %d subtest lifecycle lines (run=%d pass=%d skip=%d) for %s\n",
		stats.Runs+stats.Passes+stats.Skips,
		stats.Runs,
		stats.Passes,
		stats.Skips,
		*rootTest,
	); err != nil {
		fmt.Fprintf(errorOutput, "testlogcompact: write summary: %v\n", err)
		return 1
	}
	return 0
}
