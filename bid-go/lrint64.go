package bidgo

import (
	"math/bits"
	"runtime"
)

func bidSizeLong() int {
	if runtime.GOOS == "windows" {
		return 4
	}
	return bits.UintSize / 8
}

// Bid64Llrint is ported mechanically from Intel bid64_llrintd.c: bid64_llrint.
func Bid64Llrint(x uint64, rndMode int) (int64, uint32) {
	var res int64
	var pfpsf uint32

	if rndMode == BID_ROUNDING_TO_NEAREST {
		res, pfpsf = Bid64ToInt64Xrnint(x)
	} else if rndMode == BID_ROUNDING_TIES_AWAY {
		res, pfpsf = Bid64ToInt64Xrninta(x)
	} else if rndMode == BID_ROUNDING_DOWN {
		res, pfpsf = Bid64ToInt64Xfloor(x)
	} else if rndMode == BID_ROUNDING_UP {
		res, pfpsf = Bid64ToInt64Xceil(x)
	} else { // if (rnd_mode == BID_ROUNDING_TO_ZERO)
		res, pfpsf = Bid64ToInt64Xint(x)
	}
	return res, pfpsf
}

// Bid64Lrint is ported mechanically from Intel bid64_lrintd.c: bid64_lrint.
func Bid64Lrint(x uint64, rndMode int) (int64, uint32) {
	var res32 int32
	var res64 int64
	var pfpsf uint32

	if bidSizeLong() == 4 {
		if rndMode == BID_ROUNDING_TO_NEAREST {
			res32, pfpsf = Bid64ToInt32Xrnint(x)
		} else if rndMode == BID_ROUNDING_TIES_AWAY {
			res32, pfpsf = Bid64ToInt32Xrninta(x)
		} else if rndMode == BID_ROUNDING_DOWN {
			res32, pfpsf = Bid64ToInt32Xfloor(x)
		} else if rndMode == BID_ROUNDING_UP {
			res32, pfpsf = Bid64ToInt32Xceil(x)
		} else { // if (rnd_mode == BID_ROUNDING_TO_ZERO)
			res32, pfpsf = Bid64ToInt32Xint(x)
		}
		return int64(res32), pfpsf
	}
	// if BID_SIZE_LONG==8
	if rndMode == BID_ROUNDING_TO_NEAREST {
		res64, pfpsf = Bid64ToInt64Xrnint(x)
	} else if rndMode == BID_ROUNDING_TIES_AWAY {
		res64, pfpsf = Bid64ToInt64Xrninta(x)
	} else if rndMode == BID_ROUNDING_DOWN {
		res64, pfpsf = Bid64ToInt64Xfloor(x)
	} else if rndMode == BID_ROUNDING_UP {
		res64, pfpsf = Bid64ToInt64Xceil(x)
	} else { // if (rnd_mode == BID_ROUNDING_TO_ZERO)
		res64, pfpsf = Bid64ToInt64Xint(x)
	}
	return int64(res64), pfpsf
}

// Bid64Llround is ported mechanically from Intel bid64_llround.c: bid64_llround.
func Bid64Llround(x uint64) (int64, uint32) {
	var res int64
	var pfpsf uint32

	res, pfpsf = Bid64ToInt64Rninta(x)
	return int64(res), pfpsf
}

// Bid64Lround is ported mechanically from Intel bid64_lround.c: bid64_lround.
func Bid64Lround(x uint64) (int64, uint32) {
	var res32 int32
	var res64 int64
	var pfpsf uint32

	if bidSizeLong() == 4 {
		res32, pfpsf = Bid64ToInt32Rninta(x)
		return int64(res32), pfpsf
	}
	// if BID_SIZE_LONG==8
	res64, pfpsf = Bid64ToInt64Rninta(x)
	return int64(res64), pfpsf
}
