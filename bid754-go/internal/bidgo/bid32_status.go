package bidgo

func bid32_flags_via_bid64_binary(x, y uint32, rndMode int, op64 func(uint64, uint64, int) (uint64, uint32)) uint32 {
	x64, f1 := Bid32ToBid64(x)
	y64, f2 := Bid32ToBid64(y)
	r64, f3 := op64(x64, y64, rndMode)
	_, f4 := Bid64ToBid32(r64, rndMode)
	return f1 | f2 | f3 | f4
}

func bid32_flags_via_bid64_binary_nornd(x, y uint32, op64 func(uint64, uint64) (uint64, uint32)) uint32 {
	x64, f1 := Bid32ToBid64(x)
	y64, f2 := Bid32ToBid64(y)
	r64, f3 := op64(x64, y64)
	_, f4 := Bid64ToBid32(r64, 0)
	return f1 | f2 | f3 | f4
}

func Bid32AddWithFlags(x, y uint32, rndMode int) (uint32, uint32) {
	return bid32_add_pure(x, y, rndMode), bid32_flags_via_bid64_binary(x, y, rndMode, Bid64AddWithFlags)
}

func Bid32SubWithFlags(x, y uint32, rndMode int) (uint32, uint32) {
	return bid32_sub_pure(x, y, rndMode), bid32_flags_via_bid64_binary(x, y, rndMode, Bid64SubWithFlags)
}

func Bid32MulWithFlags(x, y uint32, rndMode int) (uint32, uint32) {
	return bid32_mul_pure(x, y, rndMode), bid32_flags_via_bid64_binary(x, y, rndMode, Bid64MulWithFlags)
}

func Bid32DivWithFlags(x, y uint32, rndMode int) (uint32, uint32) {
	return bid32_div_pure(x, y, rndMode), bid32_flags_via_bid64_binary(x, y, rndMode, Bid64DivWithFlags)
}

func Bid32MinNumWithFlags(x, y uint32) (uint32, uint32) {
	return bid32_minnum_pure(x, y), bid32_flags_via_bid64_binary_nornd(x, y, Bid64MinNum)
}

func Bid32MaxNumWithFlags(x, y uint32) (uint32, uint32) {
	return bid32_maxnum_pure(x, y), bid32_flags_via_bid64_binary_nornd(x, y, Bid64MaxNum)
}

func Bid32MinNumMagWithFlags(x, y uint32) (uint32, uint32) {
	return bid32_minnum_mag_pure(x, y), bid32_flags_via_bid64_binary_nornd(x, y, Bid64MinNumMag)
}

func Bid32MaxNumMagWithFlags(x, y uint32) (uint32, uint32) {
	return bid32_maxnum_mag_pure(x, y), bid32_flags_via_bid64_binary_nornd(x, y, Bid64MaxNumMag)
}

func Bid32ScalbnWithFlags(x uint32, n int, rndMode int) (uint32, uint32) {
	res, f0 := Bid32Scalbn(x, n, rndMode)
	x64, f1 := Bid32ToBid64(x)
	r64, f2 := Bid64Scalbn(x64, n, rndMode)
	_, f3 := Bid64ToBid32(r64, rndMode)
	return res, f0 | f1 | f2 | f3
}

func Bid32ScalblnWithFlags(x uint32, n int64, rndMode int) (uint32, uint32) {
	res, f0 := Bid32Scalbln(x, n, rndMode)
	x64, f1 := Bid32ToBid64(x)
	r64, f2 := Bid64Scalbln(x64, n, rndMode)
	_, f3 := Bid64ToBid32(r64, rndMode)
	return res, f0 | f1 | f2 | f3
}

func Bid32LdexpWithFlags(x uint32, n int, rndMode int) (uint32, uint32) {
	res, f0 := Bid32Ldexp(x, n, rndMode)
	x64, f1 := Bid32ToBid64(x)
	r64, f2 := Bid64Ldexp(x64, n, rndMode)
	_, f3 := Bid64ToBid32(r64, rndMode)
	return res, f0 | f1 | f2 | f3
}
