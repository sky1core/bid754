package testgen

import "testing"

func TestDectestNonSignalingAdaptersCheckBIDFlags(t *testing.T) {
	want := map[string]bool{
		"class": true, "samequantum": true, "copy": true,
		"copyabs": true, "copynegate": true, "copysign": true,
	}
	for _, spec := range dectestDispatchSpecs {
		for _, op := range spec.Operations {
			if !want[op] {
				continue
			}
			if spec.FlagCheck != "generatedDectestFlagCheckBIDFive" {
				t.Errorf("%s flag check = %s", op, spec.FlagCheck)
			}
			delete(want, op)
		}
	}
	for op := range want {
		t.Errorf("missing adapter %s", op)
	}
}
