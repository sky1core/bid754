// Legacy cexport placeholders are intentionally quarantined.
//
// The historical placeholder implementations live in stubs.c.quarantined so
// they cannot be linked by normal CGo builds. This C file stays in place as an
// explicit build guard for anyone attempting to revive
// bid754-go/internal/bidgo/cexport without first replacing the placeholders
// with real generated/mechanical paths.

#error "bid754: bid754-go/internal/bidgo/cexport legacy stubs are quarantined; do not link them as verification or public API behavior"
