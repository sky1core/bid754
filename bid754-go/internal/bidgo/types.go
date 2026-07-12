// types.go - Type definitions for bidgo package
package bidgo

// RoundingMode represents IEEE 754-2019 rounding modes
type RoundingMode int

const (
	// Intel BID compatible values
	RoundNearestEven    RoundingMode = 0 // BID_ROUNDING_TO_NEAREST: IEEE 754 default (half_even)
	RoundTowardNegative RoundingMode = 1 // BID_ROUNDING_DOWN: toward -infinity (floor)
	RoundTowardPositive RoundingMode = 2 // BID_ROUNDING_UP: toward +infinity (ceiling)
	RoundTowardZero     RoundingMode = 3 // BID_ROUNDING_TO_ZERO: toward zero (truncation)
	RoundNearestAway    RoundingMode = 4 // BID_ROUNDING_TIES_AWAY: away from zero (half_up)

	// Additional mode for decTest compatibility
	RoundNearestDown RoundingMode = 5 // half_down: ties toward zero
)
