package bidgo

import (
	"reflect"
	"testing"
)

func TestBIDUintTypesUseScalarLimbs(t *testing.T) {
	tests := []struct {
		name       string
		typ        reflect.Type
		fieldNames []string
	}{
		{name: "BID_UINT128", typ: reflect.TypeOf(BID_UINT128{}), fieldNames: []string{"lo", "hi"}},
		{name: "BID_UINT192", typ: reflect.TypeOf(BID_UINT192{}), fieldNames: []string{"w0", "w1", "w2"}},
		{name: "BID_UINT256", typ: reflect.TypeOf(BID_UINT256{}), fieldNames: []string{"w0", "w1", "w2", "w3"}},
		{name: "BID_UINT320", typ: reflect.TypeOf(BID_UINT320{}), fieldNames: []string{"w0", "w1", "w2", "w3", "w4"}},
		{name: "BID_UINT384", typ: reflect.TypeOf(BID_UINT384{}), fieldNames: []string{"w0", "w1", "w2", "w3", "w4", "w5"}},
		{name: "BID_UINT512", typ: reflect.TypeOf(BID_UINT512{}), fieldNames: []string{"w0", "w1", "w2", "w3", "w4", "w5", "w6", "w7"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wantSize := uintptr(len(test.fieldNames) * 8)
			if got := test.typ.Size(); got != wantSize {
				t.Fatalf("size = %d, want %d", got, wantSize)
			}
			if got, want := test.typ.NumField(), len(test.fieldNames); got != want {
				t.Fatalf("field count = %d, want %d", got, want)
			}

			for i, wantName := range test.fieldNames {
				field := test.typ.Field(i)
				wantOffset := uintptr(i * 8)
				if field.Name != wantName || field.Type.Kind() != reflect.Uint64 || field.Offset != wantOffset {
					t.Fatalf(
						"field %d = {%s %s offset=%d}, want {%s uint64 offset=%d}",
						i, field.Name, field.Type, field.Offset, wantName, wantOffset,
					)
				}
			}
		})
	}
}
