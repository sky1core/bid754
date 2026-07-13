package bidgo

import (
	"reflect"
	"testing"
)

func TestBIDUint128UsesTwoScalarLimbs(t *testing.T) {
	typ := reflect.TypeOf(BID_UINT128{})
	if got, want := typ.Size(), uintptr(16); got != want {
		t.Fatalf("BID_UINT128 size = %d, want %d", got, want)
	}
	if got, want := typ.NumField(), 2; got != want {
		t.Fatalf("BID_UINT128 field count = %d, want %d", got, want)
	}

	want := []struct {
		name   string
		offset uintptr
	}{
		{name: "lo", offset: 0},
		{name: "hi", offset: 8},
	}
	for i, expected := range want {
		field := typ.Field(i)
		if field.Name != expected.name || field.Type.Kind() != reflect.Uint64 || field.Offset != expected.offset {
			t.Fatalf(
				"BID_UINT128 field %d = {%s %s offset=%d}, want {%s uint64 offset=%d}",
				i, field.Name, field.Type, field.Offset, expected.name, expected.offset,
			)
		}
	}
}
