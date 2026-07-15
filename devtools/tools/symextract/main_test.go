package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractSymbolsRejectsReceiverMethods(t *testing.T) {
	dir := t.TempDir()
	src := []byte(`package sample

type decimal uint64

func (d decimal) bits() uint64 { return uint64(d) }
`)
	if err := os.WriteFile(filepath.Join(dir, "receiver.go"), src, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	reg, err := extractSymbols(dir)
	if err == nil {
		t.Fatalf("extractSymbols silently accepted a receiver method: %#v", reg)
	}
	if !strings.Contains(err.Error(), "receiver method bits") {
		t.Fatalf("extractSymbols error = %q, want receiver method identity", err)
	}
}

func TestExtractSymbolsAcceptsPackageFunctions(t *testing.T) {
	dir := t.TempDir()
	src := []byte(`package sample

type decimal uint64

func bits(d decimal) uint64 { return uint64(d) }
`)
	if err := os.WriteFile(filepath.Join(dir, "function.go"), src, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	reg, err := extractSymbols(dir)
	if err != nil {
		t.Fatalf("extractSymbols: %v", err)
	}
	if _, ok := reg.Functions["bits"]; !ok {
		t.Fatalf("extractSymbols omitted package function bits: %#v", reg.Functions)
	}
}
