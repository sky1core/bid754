package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func finalizeRustGenerated(projectRoot string) {
	generatedDir := filepath.Join(projectRoot, "..", "bid754-rs", "src", "generated")

	applyRustGeneratedRewrites(generatedDir)

	fmt.Println("Rust generated postprocess complete")
}

func applyRustGeneratedRewrites(generatedDir string) {
	optimizeByteCursorParser(filepath.Join(generatedDir, "bid64_from_string.rs"), "pub(crate) fn bid64_from_string")
	optimizeByteCursorParser(filepath.Join(generatedDir, "bid128_string.rs"), "pub fn bid128_from_string")
	optimizeBid32String(filepath.Join(generatedDir, "bid32_string.rs"))
	optimizeBid32MiscAliases(filepath.Join(generatedDir, "bid32_misc.rs"))
	optimizeBid64NextTowardAlias(filepath.Join(generatedDir, "nexttoward64.rs"))
	optimizeBid128Misc(filepath.Join(generatedDir, "bid128_misc.rs"))
	optimizeBid128Sqrt(filepath.Join(generatedDir, "bid128_sqrt.rs"))
}

func optimizeByteCursorParser(path string, fnSignature string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fatal("read %s: %v", path, err)
	}
	src := string(data)

	newHeader := "    let ps = str.as_bytes();\n    let mut ps_idx: usize = 0;\n    macro_rules! ps_at {\n        ($offset:expr) => {\n            *ps.get(ps_idx + ($offset as usize)).unwrap_or(&0)\n        };\n    }\n"
	headerRe := regexp.MustCompile(`(?m)^    let mut ps = \(?str\)?\.(?:into_bytes\(\)|as_bytes\(\)\.to_vec\(\));\n    ps = go_append\(ps, 0(?: as u8)?\);\n`)
	if !headerRe.MatchString(src) {
		if strings.Contains(src, "    let ps = str.as_bytes();\n    let mut ps_idx: usize = 0;\n    macro_rules! ps_at {") {
			fmt.Printf("  optimized %s: byte-cursor parser lowering already applied\n", filepath.Base(path))
			return
		}
		fatal("rewrite %s: expected byte parser header not found", path)
	}
	src = headerRe.ReplaceAllLiteralString(src, newHeader)

	advanceCount := 0
	for _, advanceRe := range []*regexp.Regexp{
		regexp.MustCompile(`ps = \(&mut ps\[(\d+) as usize\.\.\]\)\.to_vec\(\);`),
		regexp.MustCompile(`ps = &mut ps\[(\d+) as usize\.\.\];`),
	} {
		src = advanceRe.ReplaceAllStringFunc(src, func(m string) string {
			parts := advanceRe.FindStringSubmatch(m)
			advanceCount++
			return fmt.Sprintf("ps_idx += %s;", parts[1])
		})
	}
	if advanceCount == 0 {
		fatal("optimize %s: no cursor advance rewrites applied", path)
	}

	indexRe := regexp.MustCompile(`ps\[(\d+)(?: as usize)?\]`)
	indexCount := 0
	src = indexRe.ReplaceAllStringFunc(src, func(m string) string {
		parts := indexRe.FindStringSubmatch(m)
		indexCount++
		return fmt.Sprintf("ps_at!(%s)", parts[1])
	})
	if indexCount == 0 {
		fatal("optimize %s: no indexed access rewrites applied", path)
	}

	if !strings.Contains(src, fnSignature) {
		fatal("optimize %s: missing expected function signature %q", path, fnSignature)
	}

	writeFile(path, src)
	fmt.Printf("  optimized %s: byte-cursor parser lowering\n", filepath.Base(path))
}

func optimizeBid32String(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fatal("read %s: %v", path, err)
	}
	src := string(data)

	if strings.Contains(src, "String::with_capacity(digits.len() + 12)") &&
		(strings.Contains(src, `let s = (ps).trim_start_matches(|c| " \t".contains(c)).as_bytes().to_vec();`) ||
			(strings.Contains(src, `let s = (ps).trim_start_matches(|c| " \t".contains(c)).as_bytes();`) &&
				strings.Contains(src, `s.eq_ignore_ascii_case(b"inf")`))) {
		fmt.Printf("  optimized %s: string assembly lowering already applied\n", filepath.Base(path))
		return
	}

	if !strings.Contains(src, "use core::fmt::Write as _;") {
		src = mustReplaceString(src, "use super::prelude::*;\n", "use super::prelude::*;\nuse core::fmt::Write as _;\n", path)
	}

	oldBlocks := []string{
		`    let mut digits = go_string_from_bytes(&mut ps[1 as usize..istart as usize]);
    let mut adjustedExp = (((exponent_x.wrapping_sub(101)).wrapping_add((digits.len() as i64))).wrapping_sub(1));
    let mut out = (go_string_from_bytes(&mut ps[..1 as usize]) + &mut digits[..1 as usize]);
    if ((digits.len() as i64) > 1) {
        out += ("." + &mut digits[1 as usize..]);
    }
    if (adjustedExp != 0) {
        out += ("e" + format!("{}", adjustedExp));
    }
    return out;
`,
		`    let mut digits = go_string_from_bytes(&mut ps[1 as usize..istart as usize]);
    let mut adjustedExp = (((exponent_x - 101) + (digits.len() as i64)) - 1);
    let mut out = (go_string_from_bytes(&mut ps[..1 as usize]) + &mut digits[..1 as usize]);
    if ((digits.len() as i64) > 1) {
        out += ("." + &mut digits[1 as usize..]);
    }
    if (adjustedExp != 0) {
        out += ("e" + format!("{}", adjustedExp));
    }
    return out;
`,
		`    let mut digits = go_string_from_bytes(&mut ps[1 as usize..istart as usize]);
    let mut adjustedExp = (((exponent_x - 101) + (digits.len() as i64)) - 1);
    let mut out = (go_string_from_bytes(&mut ps[..1 as usize]) + &mut digits[..1 as usize]);
    if ((digits.len() as i64) > 1) {
        out += &(".".to_string() + &digits[1 as usize..]);
    }
    if (adjustedExp != 0) {
        out += &format!("e{}", adjustedExp);
    }
    return out;
`,
	}
	newBlock := `    let mut digits = go_string_from_bytes(&mut ps[1 as usize..istart as usize]);
    let mut adjustedExp = (((exponent_x - 101) + (digits.len() as i64)) - 1);
    let mut out = String::with_capacity(digits.len() + 12);
    out.push(ps[0] as char);
    out.push_str(&digits[..1 as usize]);
    if ((digits.len() as i64) > 1) {
        out.push('.');
        out.push_str(&digits[1 as usize..]);
    }
    if (adjustedExp != 0) {
        out.push('e');
        let _ = write!(&mut out, "{}", adjustedExp);
    }
    return out;
`
	replaced := false
	for _, oldBlock := range oldBlocks {
		if strings.Contains(src, oldBlock) {
			src = strings.Replace(src, oldBlock, newBlock, 1)
			replaced = true
			break
		}
	}
	if !replaced {
		fatal("rewrite %s: expected bid32 string assembly pattern not found", path)
	}

	parserHeaderRe := regexp.MustCompile(`(?m)^    let mut s = \(ps\)\.trim_start_matches\(\|c\| " \\t"\.contains\(c\)\)\.to_string\(\);\n    if \(\(s\.len\(\) as i64\) == 0\) \{\n        return \(0x7c000000, 0\);\n    \}\n    let mut c = s\[0\];\n`)
	parserHeaderNew := `    let s = (ps).trim_start_matches(|c| " \t".contains(c)).as_bytes().to_vec();
    if ((s.len() as i64) == 0) {
        return (0x7c000000, 0);
    }
    let mut c = s[0];
`
	if !parserHeaderRe.MatchString(src) {
		if strings.Contains(src, `let s = (ps).trim_start_matches(|c| " \t".contains(c)).as_bytes();`) &&
			strings.Contains(src, `s.eq_ignore_ascii_case(b"inf")`) {
			writeFile(path, src)
			fmt.Printf("  optimized %s: string assembly lowering\n", filepath.Base(path))
			return
		}
		fatal("rewrite %s: expected bid32 parser header not found", path)
	}
	src = parserHeaderRe.ReplaceAllString(src, parserHeaderNew)
	src = mustReplaceString(src, `    let mut sl = (s).to_ascii_lowercase();
`, `    let mut sl = String::from_utf8_lossy(&s).to_ascii_lowercase();
`, path)
	src = mustReplaceString(src, `        let mut sl1 = (&mut s[1 as usize..]).to_ascii_lowercase();
`, `        let mut sl1 = String::from_utf8_lossy(&s[1 as usize..]).to_ascii_lowercase();
`, path)

	writeFile(path, src)
	fmt.Printf("  optimized %s: string assembly lowering\n", filepath.Base(path))
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

	if (strings.Contains(src, "wrapping_sub(ES)") && strings.Contains(src, "wrapping_sub(ES.w[0])")) ||
		(strings.Contains(src, "ES.wrapping_neg()") && strings.Contains(src, "ES.w[0].wrapping_neg()")) {
		fmt.Printf("  optimized %s: unsigned negation lowering already applied\n", filepath.Base(path))
		return
	}

	src = mustReplaceString(src, `        ES = (-ES);
`, `        ES = (0u64).wrapping_sub(ES);
`, path)
	src = mustReplaceString(src, `        ES.w[0] = (-ES.w[0]);
`, `        ES.w[0] = (0u64).wrapping_sub(ES.w[0]);
`, path)
	src = mustReplaceString(src, `        ES.w[1] = (-ES.w[1]);
`, `        ES.w[1] = (0u64).wrapping_sub(ES.w[1]);
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
