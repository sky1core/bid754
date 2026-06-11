package bid754

import "unsafe"

// Compile-time layout guards for the SPEC.md value-type contract:
// Decimal32BID, Decimal64BID, and Decimal128BID stay fixed-width value types
// with 1:1 byte correspondence and no backend pointer or extra metadata.
// If a change alters any of these sizes, the declarations below stop
// compiling, so every build acts as the gate.
var (
	_ = [1]struct{}{}[unsafe.Sizeof(Decimal32BID(0))-4]
	_ = [1]struct{}{}[unsafe.Sizeof(Decimal64BID(0))-8]
	_ = [1]struct{}{}[unsafe.Sizeof(Decimal128BID{})-16]
)
