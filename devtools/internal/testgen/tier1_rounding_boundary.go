package testgen

import "math/big"

func tier1NativeBoundaryPower(n int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}

func tier1NativeBoundaryCoefficients(precision, shift int) []*big.Int {
	low, high := tier1NativeBoundaryPower(precision-1), tier1NativeBoundaryPower(precision)
	pattern := new(big.Int).Mul(big.NewInt(1234566), tier1NativeBoundaryPower(precision-7))
	retained := []*big.Int{
		new(big.Int).Sub(low, big.NewInt(1)), low, new(big.Int).Add(low, big.NewInt(1)),
		pattern, new(big.Int).Add(pattern, big.NewInt(1)),
		new(big.Int).Sub(high, big.NewInt(2)), new(big.Int).Sub(high, big.NewInt(1)),
	}
	scale := tier1NativeBoundaryPower(shift)
	half := new(big.Int).Quo(scale, big.NewInt(2))
	var out []*big.Int
	for _, q := range retained {
		base := new(big.Int).Mul(q, scale)
		out = append(out, new(big.Int).Set(base))
		for delta := int64(-2); delta <= 2; delta++ {
			n := new(big.Int).Add(base, half)
			out = append(out, n.Add(n, big.NewInt(delta)))
		}
	}
	return out
}

func tier1NativeBoundaryIntegers(bits int, unsigned bool) []*big.Int {
	lo := new(big.Int)
	hi := new(big.Int).Lsh(big.NewInt(1), uint(bits))
	if !unsigned {
		hi.Rsh(hi, 1)
		lo.Neg(hi)
	}
	hi.Sub(hi, big.NewInt(1))
	var out []*big.Int
	seen := map[string]bool{}
	add := func(n *big.Int) {
		if n.Cmp(lo) >= 0 && n.Cmp(hi) <= 0 && !seen[n.String()] {
			seen[n.String()] = true
			out = append(out, new(big.Int).Set(n))
		}
	}
	for _, limit := range []*big.Int{lo, hi, big.NewInt(0)} {
		for delta := int64(-2); delta <= 2; delta++ {
			add(new(big.Int).Add(limit, big.NewInt(delta)))
		}
	}
	for digits := 1; digits <= len(hi.String()); digits++ {
		for delta := int64(-1); delta <= 1; delta++ {
			n := new(big.Int).Add(tier1NativeBoundaryPower(digits), big.NewInt(delta))
			add(n)
			add(new(big.Int).Neg(n))
		}
	}
	for _, precision := range []int{7, 16} {
		for shift := 1; shift <= len(hi.String())-precision; shift++ {
			for _, n := range tier1NativeBoundaryCoefficients(precision, shift) {
				add(n)
				add(new(big.Int).Neg(n))
			}
		}
	}
	return out
}

func tier1AppendConstructorBoundaries[T ~int32 | ~uint32 | ~int64 | ~uint64](base []T, bits int, unsigned bool) []T {
	seen := make(map[T]bool, len(base))
	for _, n := range base {
		seen[n] = true
	}
	for _, n := range tier1NativeBoundaryIntegers(bits, unsigned) {
		var value T
		if unsigned {
			value = T(n.Uint64())
		} else {
			value = T(n.Int64())
		}
		if !seen[value] {
			seen[value] = true
			base = append(base, value)
		}
	}
	return base
}

func tier1NativeNarrowingBoundaries(sourcePrecision int) []tier1DecimalInputSpec {
	var out []tier1DecimalInputSpec
	for _, dst := range []struct{ precision, minExp, maxExp int }{{7, -101, 90}, {16, -398, 369}} {
		gap := sourcePrecision - dst.precision
		shifts := map[int]bool{}
		for _, shift := range []int{1, 2, gap - 1, gap, 16 - dst.precision, 17 - dst.precision} {
			if shift < 1 || shift > gap || shifts[shift] {
				continue
			}
			shifts[shift] = true
			for _, coeff := range tier1NativeBoundaryCoefficients(dst.precision, shift) {
				for _, exp := range []int{0, -shift, dst.minExp - shift, dst.minExp - shift - 1, dst.maxExp - shift} {
					for _, negative := range []bool{false, true} {
						out = append(out, tier1DecimalInputSpec{negative: negative, coefficient: coeff.String(), exponent: int32(exp)})
					}
				}
			}
		}
	}
	return out
}
