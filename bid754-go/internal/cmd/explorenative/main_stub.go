//go:build !(cgo && bid754_native)

package main

import (
	"fmt"
	"os"
)

// The default (portable) build carries no cgo, so the differential leg is
// unavailable. The devtools/cmd/explorediff driver builds this command with
// the required tags and environment; direct builds get the guidance below.
func main() {
	fmt.Fprintln(os.Stderr, "explorenative: built without the native differential leg;")
	fmt.Fprintln(os.Stderr, "run it through devtools/cmd/explorediff, or build with")
	fmt.Fprintln(os.Stderr, "CGO_ENABLED=1 go build -tags bid754_native ./internal/cmd/explorenative")
	fmt.Fprintln(os.Stderr, "after make setup-native has produced devtools/third_party/intel_dfp/lib.")
	os.Exit(2)
}
