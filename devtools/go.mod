module github.com/sky1core/bid754/devtools

go 1.23

toolchain go1.25.12

// Pin-time oracle for the Tier 1 routing sentinels: the sentinel codegen
// computes expected (result bits, raw flags) through the public bid754-go
// API, which the publicroute gate proves routes through the Go mechanical
// port. Local replace only — no external dependency enters the tree.
require github.com/sky1core/bid754/bid754-go v0.0.0

replace github.com/sky1core/bid754/bid754-go => ../bid754-go
