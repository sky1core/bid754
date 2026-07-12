package bid754_test

import (
	"fmt"

	bid754 "github.com/sky1core/bid754/bid754-go"
)

// Decimal strings render in the Intel BID coefficient-exponent form
// (sign, coefficient, E, exponent), matching the pinned C library.
func Example() {
	a, _ := bid754.NewDecimal64("123.45")
	b, _ := bid754.NewDecimal64("0.55")
	fmt.Println(a.Add(b).String())
	// Output: +12400E-2
}

// The *WithFlags variants surface the IEEE 754 exception flags raised by the
// operation.
func ExampleDecimal64BID_DivWithFlags() {
	one, _ := bid754.NewDecimal64("1")
	three, _ := bid754.NewDecimal64("3")
	q, flags := one.DivWithFlags(three)
	fmt.Println(q.String(), flags.String())
	// Output: +3333333333333333E-16 Inexact
}

// Quiet comparisons follow IEEE 754 quiet-predicate semantics: comparing
// quiet NaNs does not raise the invalid-operation flag.
func ExampleDecimal64BID_QuietLess() {
	a, _ := bid754.NewDecimal64("1.5")
	b, _ := bid754.NewDecimal64("2.5")
	less, flags := a.QuietLess(b)
	fmt.Println(less, flags.String())
	// Output: true None
}
