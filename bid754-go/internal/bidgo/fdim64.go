package bidgo

// Bid64Fdim returns x-y if x > y, and +0 if x <= y.
// Ported mechanically from Intel bid64_fdimd.c.
func Bid64Fdim(x, y uint64, rndMode int) (uint64, uint32) {
	var res uint64
	var cmpres int
	var tmp_pfpsf uint32

	tmp_pfpsf = 0
	_ = tmp_pfpsf

	cmpres, _ = Bid64QuietGreater(x, y)
	if ((x & MASK_NAN64) != MASK_NAN64) && ((y & MASK_NAN64) != MASK_NAN64) &&
		(cmpres == 0) { // if x != NaN and y != NaN and x <= y return +0
		res = 0x31c0000000000000
		return res, 0
	}

	// else if x = NaN or y = NaN or x > y return x - y
	res, pfpsf := Bid64SubWithFlags(x, y, rndMode)
	return res, pfpsf
}
