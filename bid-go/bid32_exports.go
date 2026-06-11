package bidgo

// Decimal32 exported wrappers for mechanically ported bid32 functions

func Bid32Add(x, y uint32, rndMode int) uint32 {
	return bid32_add_pure(x, y, rndMode)
}

func Bid32Sub(x, y uint32, rndMode int) uint32 {
	return bid32_sub_pure(x, y, rndMode)
}

func Bid32Mul(x, y uint32, rndMode int) uint32 {
	return bid32_mul_pure(x, y, rndMode)
}

func Bid32Div(x, y uint32, rndMode int) uint32 {
	return bid32_div_pure(x, y, rndMode)
}

func Bid32MinNum(x, y uint32) uint32 {
	return bid32_minnum_pure(x, y)
}

func Bid32MaxNum(x, y uint32) uint32 {
	return bid32_maxnum_pure(x, y)
}

func Bid32MinNumMag(x, y uint32) uint32 {
	return bid32_minnum_mag_pure(x, y)
}

func Bid32MaxNumMag(x, y uint32) uint32 {
	return bid32_maxnum_mag_pure(x, y)
}

func Bid32SameQuantum(x, y uint32) bool {
	return bid32_sameQuantum_pure(x, y)
}

func Bid32Quantum(x uint32) uint32 {
	return bid32_quantum_pure(x)
}

// Classification - mechanical ports from bid32_noncomp.go

func Bid32IsNaN(x uint32) bool {
	return Bid32IsNaN32(x) != 0
}

func Bid32IsInf(x uint32) bool {
	return Bid32IsInf32(x) != 0
}

func Bid32IsZero(x uint32) bool {
	return Bid32IsZero32(x) != 0
}

func Bid32Abs(x uint32) uint32 {
	return x & 0x7fffffff
}

func Bid32Negate(x uint32) uint32 {
	return x ^ MASK_SIGN32
}

func Bid32ToString(x uint32) string {
	return Bid32ToStringRaw(x)
}

func Bid32FromString(s string) uint32 {
	r, _ := Bid32FromStringRaw(s, 0)
	return r
}
