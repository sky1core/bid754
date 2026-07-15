// bid754-authored status-aware wrapper glue around the BID32 mechanical ports.
// The min/max status projection factorizes the identical signaling-NaN side
// effect from bid32_minmax.c; scale wrappers return the flags produced by
// bid32_scalb.c, bid32_scalbl.c, and bid32_ldexp.c directly.

package bidgo

// bid32_minmax_flags is the common status projection of the four canonical
// min/max operations: each raises invalid exactly when either operand is a
// signaling NaN. The result/cohort logic remains in the line-for-line ports.
func bid32_minmax_flags(x, y uint32) uint32 {
	if (x&MASK_SNAN32) == MASK_SNAN32 || (y&MASK_SNAN32) == MASK_SNAN32 {
		return BID_INVALID_EXCEPTION
	}
	return 0
}

func Bid32AddWithFlags(x, y uint32, rndMode int) (uint32, uint32) {
	return bid32_add_core(x, y, rndMode)
}

func Bid32SubWithFlags(x, y uint32, rndMode int) (uint32, uint32) {
	return bid32_sub_core(x, y, rndMode)
}

func Bid32MulWithFlags(x, y uint32, rndMode int) (uint32, uint32) {
	return bid32_mul_core(x, y, rndMode)
}

func Bid32DivWithFlags(x, y uint32, rndMode int) (uint32, uint32) {
	return bid32_div_core(x, y, rndMode)
}

func Bid32MinNumWithFlags(x, y uint32) (uint32, uint32) {
	return bid32_minnum_pure(x, y), bid32_minmax_flags(x, y)
}

func Bid32MaxNumWithFlags(x, y uint32) (uint32, uint32) {
	return bid32_maxnum_pure(x, y), bid32_minmax_flags(x, y)
}

func Bid32MinNumMagWithFlags(x, y uint32) (uint32, uint32) {
	return bid32_minnum_mag_pure(x, y), bid32_minmax_flags(x, y)
}

func Bid32MaxNumMagWithFlags(x, y uint32) (uint32, uint32) {
	return bid32_maxnum_mag_pure(x, y), bid32_minmax_flags(x, y)
}

func Bid32ScalbnWithFlags(x uint32, n int, rndMode int) (uint32, uint32) {
	return Bid32Scalbn(x, n, rndMode)
}

func Bid32ScalblnWithFlags(x uint32, n int64, rndMode int) (uint32, uint32) {
	return Bid32Scalbln(x, n, rndMode)
}

func Bid32LdexpWithFlags(x uint32, n int, rndMode int) (uint32, uint32) {
	return Bid32Ldexp(x, n, rndMode)
}
