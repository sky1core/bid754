package bid754

// DecimalClass is the IEEE 754-2019 class(x) result for a decimal value
// (clause 5.7.2), using the GDA decTest class spellings.
type DecimalClass string

// The ten IEEE 754-2019 classes a decimal value can belong to.
const (
	DecimalClassSignalingNaN      DecimalClass = "sNaN"
	DecimalClassQuietNaN          DecimalClass = "NaN"
	DecimalClassNegativeInfinity  DecimalClass = "-Infinity"
	DecimalClassNegativeNormal    DecimalClass = "-Normal"
	DecimalClassNegativeSubnormal DecimalClass = "-Subnormal"
	DecimalClassNegativeZero      DecimalClass = "-Zero"
	DecimalClassPositiveZero      DecimalClass = "+Zero"
	DecimalClassPositiveSubnormal DecimalClass = "+Subnormal"
	DecimalClassPositiveNormal    DecimalClass = "+Normal"
	DecimalClassPositiveInfinity  DecimalClass = "+Infinity"
)

// String returns the class spelling, e.g. "+Normal" or "sNaN".
func (c DecimalClass) String() string {
	return string(c)
}

func decimalClassFromBIDClass(class int) DecimalClass {
	switch class {
	case 0:
		return DecimalClassSignalingNaN
	case 1:
		return DecimalClassQuietNaN
	case 2:
		return DecimalClassNegativeInfinity
	case 3:
		return DecimalClassNegativeNormal
	case 4:
		return DecimalClassNegativeSubnormal
	case 5:
		return DecimalClassNegativeZero
	case 6:
		return DecimalClassPositiveZero
	case 7:
		return DecimalClassPositiveSubnormal
	case 8:
		return DecimalClassPositiveNormal
	case 9:
		return DecimalClassPositiveInfinity
	default:
		return DecimalClassQuietNaN
	}
}
