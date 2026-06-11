package main

import "testing"

func TestTypeHasUnsupportedRustFields(t *testing.T) {
	if !typeHasUnsupportedRustFields(TypeDef{
		Fields: []FieldDef{{Name: "coeff", Type: "&mut Int"}},
	}) {
		t.Fatal("expected &mut Int field to be unsupported")
	}
	if typeHasUnsupportedRustFields(TypeDef{
		Fields: []FieldDef{{Name: "w", Type: "[u64; 2]"}},
	}) {
		t.Fatal("fixed-width integer array field should be supported")
	}
}
