package goboundary

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

var approvedDefaultProductCompilationUnits = map[string][]string{
	modulePath: {
		"api_v2.go",
		"bid754.go",
		"context_v2.go",
		"format_helpers.go",
		"generated_types.go",
		"mixed_arithmetic.go",
		"mixed_fma_sqrt.go",
		"native_support.go",
		"native_support_disabled.go",
		"optimized_format_helpers.go",
		"serialization_json.go",
		"serialization_text.go",
		"serialization_xml.go",
		"types.go",
		"types_bid_class.go",
		"types_bid_compare.go",
		"types_bid_intconv.go",
		"types_bid_methods.go",
		"types_bid_nan_payload.go",
		"types_bidgo_invalid_mode.go",
		"types_bidgo_mode_arith.go",
		"types_bidgo_runtime.go",
		"types_bidgo_runtime_compare.go",
		"types_bidgo_runtime_intconv.go",
		"types_layout_check.go",
		"version.go",
	},
	modulePath + "/internal/bidgo": {
		"add128_inline.go",
		"add64.go",
		"add64_inline.go",
		"bid128_add.go",
		"bid128_compare.go",
		"bid128_conversions.go",
		"bid128_div.go",
		"bid128_fma.go",
		"bid128_fma_body.go",
		"bid128_fma_helpers.go",
		"bid128_frexp.go",
		"bid128_from_int.go",
		"bid128_internal.go",
		"bid128_ldexp.go",
		"bid128_minmax.go",
		"bid128_misc.go",
		"bid128_modf.go",
		"bid128_mul.go",
		"bid128_nearbyint.go",
		"bid128_next.go",
		"bid128_noncomp.go",
		"bid128_quantize.go",
		"bid128_rem.go",
		"bid128_round.go",
		"bid128_round_integral.go",
		"bid128_sqrt.go",
		"bid128_string.go",
		"bid128_to_binary.go",
		"bid128_to_int.go",
		"bid128_words.go",
		"bid32_add.go",
		"bid32_div.go",
		"bid32_exports.go",
		"bid32_fma.go",
		"bid32_from_int.go",
		"bid32_internal.go",
		"bid32_logb.go",
		"bid32_minmax.go",
		"bid32_misc.go",
		"bid32_mul.go",
		"bid32_next.go",
		"bid32_noncomp.go",
		"bid32_quantize.go",
		"bid32_rem.go",
		"bid32_round_integral.go",
		"bid32_scalb.go",
		"bid32_sqrt.go",
		"bid32_status.go",
		"bid32_string.go",
		"bid32_to_bid64.go",
		"bid32_to_int.go",
		"bid64_from_string.go",
		"compare32.go",
		"compare64.go",
		"convert64.go",
		"div64.go",
		"fdim64.go",
		"flag_operations.go",
		"fma64.go",
		"fmod64.go",
		"frexp64.go",
		"inline_round64.go",
		"internal.go",
		"logb64.go",
		"lrint64.go",
		"minmax64.go",
		"modf64.go",
		"mul64.go",
		"next64.go",
		"nexttoward64.go",
		"noncomp64.go",
		"quantize64.go",
		"rem64.go",
		"round_integral64.go",
		"scalb64.go",
		"sqrt64.go",
		"string64.go",
		"tables.go",
		"tables_binarydecimal.go",
		"tables_intconv.go",
		"tables_round.go",
		"tables_round128_fma.go",
		"tables_round_const128.go",
		"tables_tostring.go",
		"to_bid12864.go",
		"to_bid3264.go",
		"to_binary64.go",
		"to_int32.go",
		"to_int32_ceil.go",
		"to_int32_floor.go",
		"to_int32_int.go",
		"to_int32_rninta.go",
		"to_int64_ceil.go",
		"to_int64_floor.go",
		"to_int64_int.go",
		"to_int64_signed.go",
		"to_int_small.go",
		"to_uint32_ceil.go",
		"to_uint32_floor.go",
		"to_uint32_int.go",
		"to_uint32_rnint.go",
		"to_uint32_rninta.go",
		"to_uint64_ceil.go",
		"to_uint64_floor.go",
		"to_uint64_int.go",
		"to_uint64_rnint.go",
		"to_uint64_rninta.go",
		"to_uint64_support.go",
		"to_uint_small.go",
		"types.go",
	},
}

func checkDefaultProductSourceOwnership(packages []listedPackage) error {
	foundRoot := false
	for _, p := range packages {
		if p.ImportPath == "testing" || strings.HasPrefix(p.ImportPath, "testing/") {
			return fmt.Errorf("test dependency in production graph: %s", p.ImportPath)
		}
		if p.Standard {
			continue
		}
		approved, ok := approvedDefaultProductCompilationUnits[p.ImportPath]
		if !ok {
			return fmt.Errorf("package has no approved product source ownership: %s", p.ImportPath)
		}
		foundRoot = foundRoot || p.ImportPath == modulePath
		for _, names := range [][]string{p.GoFiles, p.CgoFiles} {
			for _, name := range names {
				if !slices.Contains(approved, name) {
					return fmt.Errorf("unapproved product compilation unit: %s/%s", p.ImportPath, name)
				}
			}
		}
	}
	if !foundRoot {
		return fmt.Errorf("public library absent from production graph: %s", modulePath)
	}
	return nil
}

func copyGoModuleSources(t *testing.T, source, target string) {
	t.Helper()
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if rel == "internal/generated" || entry.Name() == "test_results" || strings.HasPrefix(entry.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" && rel != "go.mod" && rel != "go.sum" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		dest := filepath.Join(target, rel)
		if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0600)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDefaultProductSourceOwnershipRejectsRenamedGeneratedSupport(t *testing.T) {
	source, err := filepath.Abs("../../../bid754-go")
	if err != nil {
		t.Fatal(err)
	}
	module := t.TempDir()
	copyGoModuleSources(t, source, module)
	buildAndList := func(t *testing.T, cgo, phase string) []listedPackage {
		t.Helper()
		out, err := runGo(module, cgo, "build", ".")
		if err != nil {
			t.Fatalf("%s CGO=%s BUILD FAILED: %v\n%s", phase, cgo, err, out)
		}
		t.Logf("%s CGO=%s go build .: BUILD SUCCESS", phase, cgo)
		out, err = runGo(module, cgo, "list", "-deps", "-json", ".")
		if err != nil {
			t.Fatalf("%s CGO=%s graph: %v\n%s", phase, cgo, err, out)
		}
		packages, err := decodePackages(out)
		if err != nil {
			t.Fatal(err)
		}
		return packages
	}
	for _, cgo := range []string{"0", "1"} {
		if err := checkDefaultProductSourceOwnership(buildAndList(t, cgo, "baseline")); err != nil {
			t.Fatal(err)
		}
		t.Logf("baseline CGO=%s: ownership PASS", cgo)
	}
	const exposed = "auxiliary.go"
	if err := os.Rename(filepath.Join(module, "generated_readtest_shared_test.go"), filepath.Join(module, exposed)); err != nil {
		t.Fatal(err)
	}
	for _, cgo := range []string{"0", "1"} {
		t.Run("mutant_cgo_"+cgo, func(t *testing.T) {
			packages := buildAndList(t, cgo, "mutant")
			included := false
			for _, p := range packages {
				if p.ImportPath == modulePath && slices.Contains(p.GoFiles, exposed) {
					included = true
				}
			}
			if !included {
				t.Fatal("mutant support absent from actual default compilation graph")
			}
			err := checkDefaultProductSourceOwnership(packages)
			want := "unapproved product compilation unit: " + modulePath + "/" + exposed
			if err == nil || err.Error() != want {
				t.Fatalf("want ownership rejection %q, got %v", want, err)
			}
			t.Logf("mutant CGO=%s: REJECTED: %v", cgo, err)
		})
	}
}
