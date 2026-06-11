package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRustReadtestDispatchAuditIsManifestBacked(t *testing.T) {
	projectRoot := filepath.Clean(filepath.Join("..", ".."))
	auditPath := filepath.Join(projectRoot, "generated", "testspec", "rust_readtest_dispatch_audit.json")
	data, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("read Rust readtest dispatch audit: %v", err)
	}
	var audit RustReadtestDispatchAudit
	if err := json.Unmarshal(data, &audit); err != nil {
		t.Fatalf("parse Rust readtest dispatch audit: %v", err)
	}
	if audit.SkipManifest != filepath.Join("tools", "registry", "rust_readtest_skip_manifest.json") {
		t.Fatalf("skip manifest = %q", audit.SkipManifest)
	}
	if audit.Dispatched != 521 || audit.Skipped != 0 {
		t.Fatalf("Rust readtest dispatch counts = dispatched %d skipped %d, want 521/0", audit.Dispatched, audit.Skipped)
	}
	skipped := 0
	for _, row := range audit.Functions {
		if row.Status != "skipped" {
			continue
		}
		skipped++
		if row.Function == "" || row.Compare == "" || row.ReasonCode == "" || row.Reason == "" || row.Classification == "" {
			t.Fatalf("incomplete Rust readtest skip audit row: %+v", row)
		}
	}
	if skipped != audit.Skipped {
		t.Fatalf("counted skipped rows = %d, audit skipped = %d", skipped, audit.Skipped)
	}
}

func TestRustRoundingParamAcceptsGeneratedIntWidth(t *testing.T) {
	sig, ok := parseRustFuncSigLine("pub fn bid128_add(mut x: BID_UINT128, mut y: BID_UINT128, mut rnd_mode: i64, pfpsf: &mut u32) -> BID_UINT128 {")
	if !ok {
		t.Fatal("failed to parse generated Rust signature")
	}
	if !rustSigHasRounding(sig) {
		t.Fatalf("generated Rust signature should expose rounding parameter: %+v", sig.Params)
	}
}
