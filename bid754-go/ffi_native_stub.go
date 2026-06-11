//go:build !cgo || !bid754_native

package bid754

func nativeFFIBID32Binary(function string, a uint32, b uint32) uint32 {
	return 0
}

func nativeFFIBID32Unary(function string, a uint32) uint32 {
	return 0
}

func nativeFFIBID64Binary(function string, a uint64, b uint64) uint64 {
	return 0
}

func nativeFFIBID64Unary(function string, a uint64) uint64 {
	return 0
}
