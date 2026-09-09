package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var rustPortSignature = regexp.MustCompile(`(?m)^(pub(?:\(crate\))? )fn (\w+)\(([^\n]*?)\)(?: -> ([^\n]+?))? \{`)
var rustRoundingToken = regexp.MustCompile(`(?s)//[^\n]*|/\*.*?\*/|"(?:\\.|[^"\\])*"|\b[A-Za-z_][A-Za-z0-9_]*\b`)
var rustRoundingParam = regexp.MustCompile(`\b(?:rndMode|rnd_mode|rounding_mode):`)

type roundingBoundary struct {
	name, params, result, flags string
	modes, args                 []string
	maxMode                     int
}

func isRoundingControlWordFunction(name, params, result string) (bool, error) {
	count := 0
	switch name {
	case "bid_get_decimal_rounding_direction":
		count = 1
	case "bid_set_decimal_rounding_direction":
		count = 2
	default:
		return false, nil
	}
	fields := strings.Split(params, ", ")
	if len(fields) != count || result != "u32" {
		return false, fmt.Errorf("rounding control word: unexpected signature for %s", name)
	}
	for _, field := range fields {
		_, typ, ok := strings.Cut(field, ": ")
		if !ok || typ != "u32" {
			return false, fmt.Errorf("rounding control word: unexpected parameter in %s: %s", name, field)
		}
	}
	return true, nil
}

func roundingBoundaries(src string) ([]roundingBoundary, error) {
	var out []roundingBoundary
	for _, sig := range rustPortSignature.FindAllStringSubmatch(src, -1) {
		if sig[1] != "pub " && sig[2] != "bid64_from_string" {
			continue
		}
		controlWord, err := isRoundingControlWordFunction(sig[2], sig[3], sig[4])
		if err != nil {
			return nil, err
		}
		if controlWord {
			continue
		}
		contract, ok := roundingContracts[sig[2]]
		if !ok {
			if rustRoundingParam.MatchString(sig[3]) {
				return nil, fmt.Errorf("rounding boundary: %s carries a rounding-mode parameter but has no contract entry", sig[2])
			}
			continue
		}
		b, err := contract.boundary(sig[2], sig[3], sig[4])
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, nil
}

func (c roundingContract) boundary(name, params, result string) (roundingBoundary, error) {
	b := roundingBoundary{name: name, params: params, result: result, maxMode: c.maxMode}
	fields := strings.Split(params, ", ")
	if len(fields) != len(c.params) {
		return b, fmt.Errorf("rounding boundary: %s has %d parameters, contract pins %d", name, len(fields), len(c.params))
	}
	modeAt := make(map[int]bool, len(c.modes))
	for _, i := range c.modes {
		if i < 0 || i >= len(c.params) {
			return b, fmt.Errorf("rounding boundary: %s contract mode index %d out of range", name, i)
		}
		modeAt[i] = true
	}
	for i, field := range fields {
		parts := strings.SplitN(strings.TrimPrefix(field, "mut "), ": ", 2)
		if len(parts) != 2 {
			return b, fmt.Errorf("rounding boundary: unsupported parameter %q in %s", field, name)
		}
		if parts[1] != c.params[i] {
			return b, fmt.Errorf("rounding boundary: %s parameter %d is %q, contract pins %q", name, i, parts[1], c.params[i])
		}
		b.args = append(b.args, parts[0])
		if modeAt[i] {
			if parts[1] != "i64" && parts[1] != "u32" {
				return b, fmt.Errorf("rounding boundary: unsupported mode type in %s: %s", name, field)
			}
			b.modes = append(b.modes, parts[0])
		}
		if parts[1] == "&mut u32" {
			if b.flags != "" {
				return b, fmt.Errorf("rounding boundary: ambiguous flags in %s", name)
			}
			b.flags = parts[0]
		}
	}
	return b, nil
}

func (b roundingBoundary) invalidValue(result string) (string, error) {
	if strings.Contains(b.name, "_to_binary") {
		switch result {
		case "u32":
			return "0x7fc00000", nil
		case "u64":
			return "0x7ff8000000000000", nil
		case "BID_UINT128":
			return "BID_UINT128 { lo: 0, hi: 0x7fff800000000000 }", nil
		case "f32":
			return "f32::NAN", nil
		case "f64":
			return "f64::NAN", nil
		}
	}
	switch result {
	case "u32":
		return "0x7c000000", nil
	case "u64":
		return "0x7c00000000000000", nil
	case "BID_UINT128":
		return "BID_UINT128 { lo: 0, hi: 0x7c00000000000000 }", nil
	case "i64":
		return "i64::MIN", nil
	default:
		return "", fmt.Errorf("rounding boundary: no invalid result contract for %s -> %s", b.name, result)
	}
}

func (b roundingBoundary) wrapper() (string, error) {
	result := b.result
	call := b.name + "_port(" + strings.Join(b.args, ", ") + ")"
	var failure string
	if strings.HasPrefix(result, "(") && strings.HasSuffix(result, ", u32)") {
		value, err := b.invalidValue(strings.TrimSuffix(strings.TrimPrefix(result, "("), ", u32)"))
		if err != nil {
			return "", err
		}
		failure = "return (" + value + ", 0x01);"
	} else if b.flags != "" {
		value, err := b.invalidValue(result)
		if err != nil {
			return "", err
		}
		failure = "*" + b.flags + " |= 0x01; return " + value + ";"
	} else {
		result = "Result<" + result + ", &'static str>"
		failure = `return Err("unsupported rounding mode");`
		call = "Ok(" + call + ")"
	}
	var checks []string
	for _, mode := range b.modes {
		checks = append(checks, fmt.Sprintf("!(0..=%d).contains(&%s)", b.maxMode, mode))
	}
	return fmt.Sprintf("\n#[inline]\npub fn %s(%s) -> %s {\n    if %s { %s }\n    %s\n}\n", b.name, b.params, result, strings.Join(checks, " || "), failure, call), nil
}

func checkedRoundingSources(sources map[string]string) (map[string]string, error) {
	boundaries := make(map[string][]roundingBoundary)
	renames := make(map[string]string)
	var names []string
	for name, src := range sources {
		names = append(names, name)
		bs, err := roundingBoundaries(src)
		if err != nil {
			return nil, err
		}
		boundaries[name] = bs
		for _, b := range bs {
			if _, duplicate := renames[b.name]; duplicate {
				return nil, fmt.Errorf("rounding boundary: duplicate port %s", b.name)
			}
			renames[b.name] = b.name + "_port"
		}
	}
	for name := range roundingContracts {
		if _, ok := renames[name]; !ok {
			return nil, fmt.Errorf("rounding boundary: contract entry %s matched no generated function", name)
		}
	}
	sort.Strings(names)
	out := make(map[string]string)
	for _, name := range names {
		var rewritten strings.Builder
		original := sources[name]
		last := 0
		for _, span := range rustRoundingToken.FindAllStringIndex(original, -1) {
			id := original[span[0]:span[1]]
			renamed, ok := renames[id]
			if !ok || strings.HasPrefix(strings.TrimSpace(original[span[1]:]), "::") ||
				strings.HasSuffix(original[:span[0]], "mod ") {
				continue
			}
			rewritten.WriteString(original[last:span[0]])
			rewritten.WriteString(renamed)
			last = span[1]
		}
		rewritten.WriteString(original[last:])
		src := rewritten.String()
		for _, b := range boundaries[name] {
			src = strings.ReplaceAll(src, "pub fn "+b.name+"_port(", "pub(crate) fn "+b.name+"_port(")
			wrapper, err := b.wrapper()
			if err != nil {
				return nil, err
			}
			src += wrapper
		}
		out[name] = src
	}
	return out, nil
}

func emitCheckedRoundingBoundaries(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.rs"))
	if err != nil {
		return err
	}
	sources := make(map[string]string)
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sources[path] = string(data)
	}
	checked, err := checkedRoundingSources(sources)
	if err != nil {
		return err
	}
	for _, path := range files {
		if err := os.WriteFile(path, []byte(checked[path]), 0o644); err != nil {
			return err
		}
	}
	return nil
}
