package testgen

import "strings"

// NormalizeReadtestFuncName is the single shared normalization used to match
// Intel readtest function names against port implementation names (Go
// exported CamelCase, generated Rust snake_case): lowercase with all
// underscores removed. Both the goport readtest generator and the Rust
// readtest generator (devtools/tools/codegen) resolve their dispatch targets
// through this normalization, and the dispatch cross-check test compares the
// two resolved surfaces under it.
func NormalizeReadtestFuncName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, "_", ""))
}
