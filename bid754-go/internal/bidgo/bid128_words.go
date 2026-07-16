// bid754-authored word-level constructor/accessor pair for the BID_UINT128
// mechanical-port representation. The struct fields stay unexported so the
// port boundary remains closed; these two functions expose the explicit
// (hi, lo) word view that the public package and the platform-digest tool use
// to convert between the little-endian [16]byte value-type image and the port
// representation. Keeping the conversion word-explicit removes the
// native-endian pointer reinterpretation that byte-swapped every 128-bit word
// on big-endian platforms.

package bidgo

// Bid128FromWords constructs a BID_UINT128 from its high and low 64-bit words.
func Bid128FromWords(hi, lo uint64) BID_UINT128 {
	return BID_UINT128{lo: lo, hi: hi}
}

// Bid128Words returns the high and low 64-bit words of x.
func Bid128Words(x BID_UINT128) (hi, lo uint64) {
	return x.hi, x.lo
}
