// bid754-authored implementation (no originating Intel C file): the Intel
// BID library has no reduce operation; this implements the decTest
// trailing-zero reduction for Decimal64. The NaN/Inf/unpack boilerplate
// follows the Intel idiom used across the ported files.

package bidgo

// Bid64Reduce removes trailing zeros from the coefficient and adjusts the exponent.
func Bid64Reduce(x uint64) (uint64, uint32) {
	var xSign uint64
	var xExp uint64
	var c1 uint64

	if (x & MASK_NAN64) == MASK_NAN64 {
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x = x & 0xfe00000000000000
		} else {
			x = x & 0xfe03ffffffffffff
		}
		if (x & MASK_SNAN64) == MASK_SNAN64 {
			return x & QUIET_MASK64, BID_INVALID_EXCEPTION
		}
		return x, 0
	}

	if (x & MASK_INF64) == MASK_INF64 {
		return (x & MASK_SIGN64) | MASK_INF64, 0
	}

	xSign = x & MASK_SIGN64
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		c1 = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		xExp = (x & MASK_BINARY_EXPONENT2_64) >> 51
		if c1 > 9999999999999999 || c1 == 0 {
			return xSign | (uint64(DECIMAL_EXPONENT_BIAS) << 53), 0
		}
	} else {
		c1 = x & MASK_BINARY_SIG1_64
		xExp = (x & MASK_BINARY_EXPONENT1_64) >> 53
		if c1 == 0 {
			return xSign | (uint64(DECIMAL_EXPONENT_BIAS) << 53), 0
		}
	}

	for c1 != 0 && (c1%10) == 0 && xExp < DECIMAL_MAX_EXPON_64 {
		c1 = c1 / 10
		xExp = xExp + 1
	}

	if c1 < MASK_BINARY_OR2_64 {
		return xSign | (xExp << 53) | c1, 0
	}
	return xSign | MASK_STEERING_BITS64 | (xExp << 51) | (c1 & MASK_BINARY_SIG2_64), 0
}
