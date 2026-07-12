package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRustReadtestDispatchInventoryIsManifestBacked(t *testing.T) {
	projectRoot := filepath.Clean(filepath.Join("..", ".."))
	inventoryPath := filepath.Join(projectRoot, "generated", "testspec", "rust_readtest_dispatch_inventory.json")
	data, err := os.ReadFile(inventoryPath)
	if err != nil {
		t.Fatalf("read Rust readtest dispatch inventory: %v", err)
	}
	var inventory RustReadtestDispatchInventory
	if err := json.Unmarshal(data, &inventory); err != nil {
		t.Fatalf("parse Rust readtest dispatch inventory: %v", err)
	}
	if inventory.SkipManifest != filepath.Join("tools", "registry", "rust_readtest_skip_manifest.json") {
		t.Fatalf("skip manifest = %q", inventory.SkipManifest)
	}
	if inventory.Dispatched != 521 || inventory.Skipped != 0 {
		t.Fatalf("Rust readtest dispatch counts = dispatched %d skipped %d, want 521/0", inventory.Dispatched, inventory.Skipped)
	}
	skipped := 0
	for _, row := range inventory.Functions {
		if row.Status == "skipped" {
			skipped++
			if row.Function == "" || row.Compare == "" || row.ReasonCode == "" || row.Reason == "" || row.Classification == "" {
				t.Fatalf("incomplete Rust readtest skip inventory row: %+v", row)
			}
			continue
		}
		switch row.Route {
		case "generic":
			if row.RustFunction == "" {
				t.Fatalf("generic Rust readtest dispatch row lacks rust_function: %+v", row)
			}
		case "custom":
			if row.RustFunction != "" {
				t.Fatalf("custom Rust readtest dispatch row carries rust_function: %+v", row)
			}
		default:
			t.Fatalf("Rust readtest dispatch row has unknown route: %+v", row)
		}
	}
	if skipped != inventory.Skipped {
		t.Fatalf("counted skipped rows = %d, inventory skipped = %d", skipped, inventory.Skipped)
	}
	if len(inventory.SuitePasses) != len(readtestSuiteFilters) {
		t.Fatalf("inventory suite pass rows = %d, want %d", len(inventory.SuitePasses), len(readtestSuiteFilters))
	}
	for i, suite := range readtestSuiteFilters {
		row := inventory.SuitePasses[i]
		if row.Suite != suite.Display || row.Filter != suite.Filter {
			t.Fatalf("inventory suite pass row %d = %+v, want suite %q filter %q", i, row, suite.Display, suite.Filter)
		}
		if row.ExpectedPasses <= 0 {
			t.Fatalf("inventory suite pass row %d has non-positive expected passes: %+v", i, row)
		}
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
