package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func finalizeRustGenerated(projectRoot string) {
	generatedDir := filepath.Join(projectRoot, "..", "bid754-rs", "src", "generated")

	applyRustGeneratedRewrites(generatedDir)

	fmt.Println("Rust generated postprocess complete")
}

func applyRustGeneratedRewrites(generatedDir string) {
	optimizeBid32MiscAliases(filepath.Join(generatedDir, "bid32_misc.rs"))
	optimizeBid64NextTowardAlias(filepath.Join(generatedDir, "nexttoward64.rs"))
	optimizeBid128Misc(filepath.Join(generatedDir, "bid128_misc.rs"))
	optimizeBid128Sqrt(filepath.Join(generatedDir, "bid128_sqrt.rs"))
}

func optimizeBid32MiscAliases(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fatal("read %s: %v", path, err)
	}
	src := string(data)

	alias := `
pub fn bid32_nexttoward(mut x: u32, mut y: BID_UINT128) -> (u32, u32) {
    bid32_next_toward(x, y)
}
`
	if strings.Contains(src, "pub fn bid32_nexttoward(") {
		fmt.Printf("  optimized %s: nexttoward alias already applied\n", filepath.Base(path))
		return
	}
	src += alias
	writeFile(path, src)
	fmt.Printf("  optimized %s: nexttoward alias lowering\n", filepath.Base(path))
}

func optimizeBid64NextTowardAlias(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fatal("read %s: %v", path, err)
	}
	src := string(data)

	alias := `
pub fn bid64_nexttoward(mut x: u64, mut y: BID_UINT128) -> (u64, u32) {
    bid64_next_toward(x, y)
}
`
	if strings.Contains(src, "pub fn bid64_nexttoward(") {
		fmt.Printf("  optimized %s: nexttoward alias already applied\n", filepath.Base(path))
		return
	}
	src += alias
	writeFile(path, src)
	fmt.Printf("  optimized %s: nexttoward alias lowering\n", filepath.Base(path))
}

func optimizeBid128Misc(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fatal("read %s: %v", path, err)
	}
	src := string(data)

	if strings.Contains(src, "i64::from(n1)") {
		fmt.Printf("  optimized %s: scalbln closure lowering already applied\n", filepath.Base(path))
		return
	}

	oldBlock := `pub fn bid128_scalbln(mut x: BID_UINT128, mut n: i64, mut rnd_mode: i64, pfpsf: &mut u32) -> BID_UINT128 {
    let mut n1 = (n as i32);
    n1 = (|| -> i32 {
    if ((n1 as i64) < n) {
        return (0x7fffffff as i32);
    }
    if ((n1 as i64) > n) {
        return ((-0x80000000) as i32);
    }
    return n1;
})();
    return bid128_scalbn(x, (n1 as i64), rnd_mode, pfpsf);
}
`
	newBlock := `pub fn bid128_scalbln(mut x: BID_UINT128, mut n: i64, mut rnd_mode: i64, pfpsf: &mut u32) -> BID_UINT128 {
    let mut n1 = (n as i32);
    n1 = if ((i64::from(n1)) < n) {
        i32::MAX
    } else if ((i64::from(n1)) > n) {
        i32::MIN
    } else {
        n1
    };
    return bid128_scalbn(x, (n1 as i64), rnd_mode, pfpsf);
}
`
	if !strings.Contains(src, oldBlock) {
		fatal("rewrite %s: expected structured bid128_scalbln immediate-closure block; rerun go2rs or update the converter instead of relying on fallback TODO markers", path)
	}
	src = strings.Replace(src, oldBlock, newBlock, 1)

	writeFile(path, src)
	fmt.Printf("  optimized %s: scalbln closure lowering\n", filepath.Base(path))
}

func optimizeBid128Sqrt(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fatal("read %s: %v", path, err)
	}
	src := string(data)

	if (strings.Contains(src, "wrapping_sub(ES)") && strings.Contains(src, "wrapping_sub(ES.lo)")) ||
		(strings.Contains(src, "ES.wrapping_neg()") && strings.Contains(src, "ES.lo.wrapping_neg()")) {
		fmt.Printf("  optimized %s: unsigned negation lowering already applied\n", filepath.Base(path))
		return
	}

	src = mustReplaceString(src, `        ES = (-ES);
`, `        ES = (0u64).wrapping_sub(ES);
`, path)
	src = mustReplaceString(src, `        ES.lo = (-ES.lo);
`, `        ES.lo = (0u64).wrapping_sub(ES.lo);
`, path)
	src = mustReplaceString(src, `        ES.hi = (-ES.hi);
`, `        ES.hi = (0u64).wrapping_sub(ES.hi);
`, path)

	writeFile(path, src)
	fmt.Printf("  optimized %s: unsigned negation lowering\n", filepath.Base(path))
}

func writeFile(path, content string) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fatal("mkdir %s: %v", filepath.Dir(path), err)
	}
	content = strings.TrimRight(content, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fatal("write %s: %v", path, err)
	}
}

func mustReplaceString(src, old, new, path string) string {
	if !strings.Contains(src, old) {
		fatal("rewrite %s: expected pattern not found", path)
	}
	return strings.Replace(src, old, new, 1)
}
