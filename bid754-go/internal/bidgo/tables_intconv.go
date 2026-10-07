package bidgo

// DEC_DIGITS is ported mechanically from Intel bid_internal.h.
type DEC_DIGITS struct {
	digits       uint32
	threshold_hi uint64
	threshold_lo uint64
	digits1      uint32
}
