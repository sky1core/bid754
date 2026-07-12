// Ported from: Intel bid32_noncomp.c
// Mechanical translation - all logic preserved exactly.

package bidgo

var bid32_mult_factor = [7]uint64{1, 10, 100, 1000, 10000, 100000, 1000000}

// Bid32IsSigned returns 1 if x has sign bit set.
func Bid32IsSigned(x uint32) int {
	if (x & MASK_SIGN32) == MASK_SIGN32 {
		return 1
	}
	return 0
}

// Bid32IsNormal returns 1 if x is a normal number.
func Bid32IsNormal(x uint32) int {
	var res int
	var sig_x_prime uint64
	var sig_x, exp_x uint32

	if (x & MASK_INF32) == MASK_INF32 {
		res = 0
	} else {
		if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
			sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
			if sig_x > 9999999 || sig_x == 0 {
				return 0
			}
			exp_x = (x & MASK_BINARY_EXPONENT2_32) >> 21
		} else {
			sig_x = (x & MASK_BINARY_SIG1_32)
			if sig_x == 0 {
				return 0
			}
			exp_x = (x & MASK_BINARY_EXPONENT1_32) >> 23
		}
		if exp_x < 6 {
			sig_x_prime = uint64(sig_x) * bid32_mult_factor[exp_x]
			if sig_x_prime < 1000000 {
				res = 0
			} else {
				res = 1
			}
		} else {
			res = 1
		}
	}
	return res
}

// Bid32IsSubnormal returns 1 if x is a subnormal number.
func Bid32IsSubnormal(x uint32) int {
	var res int
	var sig_x_prime uint64
	var sig_x, exp_x uint32

	if (x & MASK_INF32) == MASK_INF32 {
		res = 0
	} else {
		if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
			sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
			if sig_x > 9999999 || sig_x == 0 {
				return 0
			}
			exp_x = (x & MASK_BINARY_EXPONENT2_32) >> 21
		} else {
			sig_x = (x & MASK_BINARY_SIG1_32)
			if sig_x == 0 {
				return 0
			}
			exp_x = (x & MASK_BINARY_EXPONENT1_32) >> 23
		}
		if exp_x < 6 {
			sig_x_prime = uint64(sig_x) * bid32_mult_factor[exp_x]
			if sig_x_prime < 1000000 {
				res = 1
			} else {
				res = 0
			}
		} else {
			res = 0
		}
	}
	return res
}

// Bid32IsFinite returns 1 if x is finite (not infinity or NaN).
func Bid32IsFinite(x uint32) int {
	if (x & MASK_INF32) != MASK_INF32 {
		return 1
	}
	return 0
}

// Bid32IsZero returns 1 if x is zero.
func Bid32IsZero32(x uint32) int {
	var res int
	if (x & MASK_INF32) == MASK_INF32 {
		res = 0
	} else if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		if ((x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32) > 9999999 {
			res = 1
		} else {
			res = 0
		}
	} else {
		if (x & MASK_BINARY_SIG1_32) == 0 {
			res = 1
		} else {
			res = 0
		}
	}
	return res
}

// Bid32IsInf returns 1 if x is infinity.
func Bid32IsInf32(x uint32) int {
	if ((x & MASK_INF32) == MASK_INF32) && ((x & MASK_NAN32) != MASK_NAN32) {
		return 1
	}
	return 0
}

// Bid32IsSignaling returns 1 if x is a signaling NaN.
func Bid32IsSignaling(x uint32) int {
	if (x & MASK_SNAN32) == MASK_SNAN32 {
		return 1
	}
	return 0
}

// Bid32IsNaN32 returns 1 if x is NaN.
func Bid32IsNaN32(x uint32) int {
	if (x & MASK_NAN32) == MASK_NAN32 {
		return 1
	}
	return 0
}

// Bid32IsCanonical returns 1 if x is canonical.
func Bid32IsCanonical(x uint32) int {
	if (x & MASK_NAN32) == MASK_NAN32 {
		if (x & 0x01f00000) != 0 {
			return 0
		}
		if (x & 0x000fffff) > 999999 {
			return 0
		}
		return 1
	}
	if (x & MASK_INF32) == MASK_INF32 {
		if (x & 0x03ffffff) != 0 {
			return 0
		}
		return 1
	}
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		if ((x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32) > 9999999 {
			return 0
		}
	}
	return 1
}

// Bid32Copy returns a copy of x.
func Bid32Copy(x uint32) uint32 {
	return x
}

// Bid32Negate returns -x.
// Already in bid32_exports.go via Decimal32Pure.Neg but re-implementing mechanically.

// Bid32CopySign returns x with the sign of y.
func Bid32CopySign(x, y uint32) uint32 {
	return (x & 0x7fffffff) | (y & MASK_SIGN32)
}

// Bid32Radix returns 10.
func Bid32Radix() int {
	return 10
}

// Bid32TotalOrder is ported mechanically from bid32_noncomp.c: bid32_totalOrder.
func Bid32TotalOrder(x, y uint32) int {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y, pyld_y, pyld_x uint32
	var sig_n_prime uint64
	var x_is_zero, y_is_zero int

	// NaN (CASE1)
	if (x & MASK_NAN32) == MASK_NAN32 {
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			if (y&MASK_NAN32) != MASK_NAN32 || (y&MASK_SIGN32) != MASK_SIGN32 {
				res = 1
				return res
			} else {
				if !(((y & MASK_SNAN32) == MASK_SNAN32) != ((x & MASK_SNAN32) == MASK_SNAN32)) {
					pyld_y = y & 0x000fffff
					pyld_x = x & 0x000fffff
					if pyld_y > 999999 || pyld_y == 0 {
						res = 1
						return res
					}
					if pyld_x > 999999 || pyld_x == 0 {
						res = 0
						return res
					}
					res = 0
					if pyld_x >= pyld_y {
						res = 1
					}
					return res
				} else {
					res = 0
					if (y & MASK_SNAN32) == MASK_SNAN32 {
						res = 1
					}
					return res
				}
			}
		} else {
			if (y&MASK_NAN32) != MASK_NAN32 || (y&MASK_SIGN32) == MASK_SIGN32 {
				res = 0
				return res
			} else {
				if !(((y & MASK_SNAN32) == MASK_SNAN32) != ((x & MASK_SNAN32) == MASK_SNAN32)) {
					pyld_y = y & 0x000fffff
					pyld_x = x & 0x000fffff
					if pyld_x > 999999 || pyld_x == 0 {
						res = 1
						return res
					}
					if pyld_y > 999999 || pyld_y == 0 {
						res = 0
						return res
					}
					res = 0
					if pyld_x <= pyld_y {
						res = 1
					}
					return res
				} else {
					res = 0
					if (x & MASK_SNAN32) == MASK_SNAN32 {
						res = 1
					}
					return res
				}
			}
		}
	} else if (y & MASK_NAN32) == MASK_NAN32 {
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res
	}
	if x == y {
		res = 1
		return res
	}
	if ((x & MASK_SIGN32) == MASK_SIGN32) != ((y & MASK_SIGN32) == MASK_SIGN32) {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res
	}
	if (x & MASK_INF32) == MASK_INF32 {
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
			return res
		} else {
			res = 0
			if (y & MASK_INF32) == MASK_INF32 {
				res = 1
			}
			return res
		}
	} else if (y & MASK_INF32) == MASK_INF32 {
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res
	}
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 || sig_x == 0 {
			x_is_zero = 1
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = (x & MASK_BINARY_SIG1_32)
		if sig_x == 0 {
			x_is_zero = 1
		}
	}
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 || sig_y == 0 {
			y_is_zero = 1
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = (y & MASK_BINARY_SIG1_32)
		if sig_y == 0 {
			y_is_zero = 1
		}
	}
	if x_is_zero != 0 && y_is_zero != 0 {
		if exp_x == exp_y {
			res = 1
			return res
		}
		res = 0
		if (exp_x <= exp_y) != ((x & MASK_SIGN32) == MASK_SIGN32) {
			res = 1
		}
		return res
	}
	if x_is_zero != 0 {
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res
	}
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res
	}
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res
	}
	if exp_x-exp_y > 6 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res
	}
	if exp_y-exp_x > 6 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res
	}
	if exp_x > exp_y {
		sig_n_prime = uint64(sig_x) * bid32_mult_factor[exp_x-exp_y]
		if sig_n_prime == uint64(sig_y) {
			res = 0
			if (exp_x <= exp_y) != ((x & MASK_SIGN32) == MASK_SIGN32) {
				res = 1
			}
			return res
		}
		res = 0
		if (sig_n_prime < uint64(sig_y)) != ((x & MASK_SIGN32) == MASK_SIGN32) {
			res = 1
		}
		return res
	}
	sig_n_prime = uint64(sig_y) * bid32_mult_factor[exp_y-exp_x]
	if sig_n_prime == uint64(sig_x) {
		res = 0
		if (exp_x <= exp_y) != ((x & MASK_SIGN32) == MASK_SIGN32) {
			res = 1
		}
		return res
	}
	res = 0
	if (uint64(sig_x) < sig_n_prime) != ((x & MASK_SIGN32) == MASK_SIGN32) {
		res = 1
	}
	return res
}

// Bid32TotalOrderMag is TotalOrder(|x|, |y|).
func Bid32TotalOrderMag(x, y uint32) int {
	return Bid32TotalOrder(x&0x7fffffff, y&0x7fffffff)
}
