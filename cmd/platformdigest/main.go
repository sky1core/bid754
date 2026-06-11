// Command platformdigest computes a deterministic SHA-256 digest of Go
// mechanical-port outputs over the generated testspec inputs, for the direct
// cross-platform bit comparison required by PLATFORM_SPEC section 4 item 2.
//
// It replays every generated readtest case whose function belongs to the
// seed-sensitive core set (add/sub/mul/div/fma/sqrt in all three widths)
// through bid-go, recording parsed operand bits, result bits, and flags.
// Two platforms agree bit-for-bit on these operations iff their digests match.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unsafe"

	bidgo "github.com/sky1core/bid754/bid-go"
	"github.com/sky1core/bid754/internal/testgen"
)

func main() {
	spec, err := testgen.LoadGenerated("generated/testspec/spec_index.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "platformdigest: load testspec: %v\n", err)
		os.Exit(1)
	}

	var lines []string
	for _, c := range spec.ReadCases {
		line, ok := digestLine(c)
		if !ok {
			continue
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		fmt.Fprintln(os.Stderr, "platformdigest: no digest cases produced")
		os.Exit(1)
	}
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	fmt.Printf("PLATFORM-DIGEST goos=%s goarch=%s cases=%d sha256=%s\n",
		runtime.GOOS, runtime.GOARCH, len(lines), hex.EncodeToString(sum[:]))
}

func digestLine(c testgen.GeneratedReadCase) (string, bool) {
	out, ok := runCase(c)
	if !ok {
		return "", false
	}
	return fmt.Sprintf("%s|%s|rnd=%d|%s", c.Function, c.ID, c.Rounding, out), true
}

func runCase(c testgen.GeneratedReadCase) (string, bool) {
	rnd := c.Rounding
	switch c.Function {
	case "bid32_add", "bid32_sub", "bid32_mul", "bid32_div":
		x, ok1 := operand32(c.Operands, 0)
		y, ok2 := operand32(c.Operands, 1)
		if !ok1 || !ok2 {
			return "", false
		}
		var r, f uint32
		switch c.Function {
		case "bid32_add":
			r, f = bidgo.Bid32AddWithFlags(x, y, rnd)
		case "bid32_sub":
			r, f = bidgo.Bid32SubWithFlags(x, y, rnd)
		case "bid32_mul":
			r, f = bidgo.Bid32MulWithFlags(x, y, rnd)
		case "bid32_div":
			r, f = bidgo.Bid32DivWithFlags(x, y, rnd)
		}
		return fmt.Sprintf("in=%08x,%08x|out=%08x|fl=%02x", x, y, r, f), true
	case "bid32_sqrt":
		x, ok := operand32(c.Operands, 0)
		if !ok {
			return "", false
		}
		r, f := bidgo.Bid32Sqrt(x, rnd)
		return fmt.Sprintf("in=%08x|out=%08x|fl=%02x", x, r, f), true
	case "bid32_fma":
		x, ok1 := operand32(c.Operands, 0)
		y, ok2 := operand32(c.Operands, 1)
		z, ok3 := operand32(c.Operands, 2)
		if !ok1 || !ok2 || !ok3 {
			return "", false
		}
		r, f := bidgo.Bid32Fma(x, y, z, rnd)
		return fmt.Sprintf("in=%08x,%08x,%08x|out=%08x|fl=%02x", x, y, z, r, f), true

	case "bid64_add", "bid64_sub", "bid64_mul", "bid64_div":
		x, ok1 := operand64(c.Operands, 0)
		y, ok2 := operand64(c.Operands, 1)
		if !ok1 || !ok2 {
			return "", false
		}
		var r uint64
		var f uint32
		switch c.Function {
		case "bid64_add":
			r, f = bidgo.Bid64AddWithFlags(x, y, rnd)
		case "bid64_sub":
			r, f = bidgo.Bid64SubWithFlags(x, y, rnd)
		case "bid64_mul":
			r, f = bidgo.Bid64MulWithFlags(x, y, rnd)
		case "bid64_div":
			r, f = bidgo.Bid64DivWithFlags(x, y, rnd)
		}
		return fmt.Sprintf("in=%016x,%016x|out=%016x|fl=%02x", x, y, r, f), true
	case "bid64_sqrt":
		x, ok := operand64(c.Operands, 0)
		if !ok {
			return "", false
		}
		r, f := bidgo.Bid64Sqrt(x, rnd)
		return fmt.Sprintf("in=%016x|out=%016x|fl=%02x", x, r, f), true
	case "bid64_fma":
		x, ok1 := operand64(c.Operands, 0)
		y, ok2 := operand64(c.Operands, 1)
		z, ok3 := operand64(c.Operands, 2)
		if !ok1 || !ok2 || !ok3 {
			return "", false
		}
		r, f := bidgo.Bid64Fma(x, y, z, rnd)
		return fmt.Sprintf("in=%016x,%016x,%016x|out=%016x|fl=%02x", x, y, z, r, f), true

	case "bid128_add", "bid128_sub", "bid128_mul", "bid128_div":
		x, ok1 := operand128(c.Operands, 0)
		y, ok2 := operand128(c.Operands, 1)
		if !ok1 || !ok2 {
			return "", false
		}
		var r bidgo.BID_UINT128
		var f uint32
		switch c.Function {
		case "bid128_add":
			r = bidgo.Bid128Add(x, y, rnd, &f)
		case "bid128_sub":
			r = bidgo.Bid128Sub(x, y, rnd, &f)
		case "bid128_mul":
			r, f = bidgo.Bid128Mul(x, y, rnd)
		case "bid128_div":
			r, f = bidgo.Bid128Div(x, y, rnd)
		}
		return fmt.Sprintf("in=%s,%s|out=%s|fl=%02x", hex128(x), hex128(y), hex128(r), f), true
	case "bid128_sqrt":
		x, ok := operand128(c.Operands, 0)
		if !ok {
			return "", false
		}
		r, f := bidgo.Bid128Sqrt(x, rnd)
		return fmt.Sprintf("in=%s|out=%s|fl=%02x", hex128(x), hex128(r), f), true
	case "bid128_fma":
		x, ok1 := operand128(c.Operands, 0)
		y, ok2 := operand128(c.Operands, 1)
		z, ok3 := operand128(c.Operands, 2)
		if !ok1 || !ok2 || !ok3 {
			return "", false
		}
		r, f := bidgo.Bid128Fma(x, y, z, rnd)
		return fmt.Sprintf("in=%s,%s,%s|out=%s|fl=%02x", hex128(x), hex128(y), hex128(z), hex128(r), f), true
	}
	return "", false
}

// operandN parses a readtest operand: "[hex]" is a direct bit pattern, and
// anything else goes through the bid-go from_string path (whose parsing
// determinism is therefore part of the digest as well).
func operand32(ops []string, i int) (uint32, bool) {
	if i >= len(ops) {
		return 0, false
	}
	if h, ok := bracketHex(ops[i]); ok {
		v, err := strconv.ParseUint(h, 16, 32)
		if err != nil {
			return 0, false
		}
		return uint32(v), true
	}
	r, _ := bidgo.Bid32FromStringRaw(ops[i], 0)
	return r, true
}

func operand64(ops []string, i int) (uint64, bool) {
	if i >= len(ops) {
		return 0, false
	}
	if h, ok := bracketHex(ops[i]); ok {
		v, err := strconv.ParseUint(h, 16, 64)
		if err != nil {
			return 0, false
		}
		return v, true
	}
	r, _ := bidgo.Bid64FromString(ops[i], 0)
	return r, true
}

func operand128(ops []string, i int) (bidgo.BID_UINT128, bool) {
	var zero bidgo.BID_UINT128
	if i >= len(ops) {
		return zero, false
	}
	if h, ok := bracketHex(ops[i]); ok {
		if len(h) != 32 {
			return zero, false
		}
		hi, err1 := strconv.ParseUint(h[:16], 16, 64)
		lo, err2 := strconv.ParseUint(h[16:], 16, 64)
		if err1 != nil || err2 != nil {
			return zero, false
		}
		var raw [16]byte
		for b := 0; b < 8; b++ {
			raw[b] = byte(lo >> (8 * b))
			raw[8+b] = byte(hi >> (8 * b))
		}
		return *(*bidgo.BID_UINT128)(unsafe.Pointer(&raw)), true
	}
	r, _ := bidgo.Bid128FromString(ops[i], 0)
	return r, true
}

func bracketHex(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		return strings.ToLower(strings.Trim(s, "[]")), true
	}
	return "", false
}

func hex128(x bidgo.BID_UINT128) string {
	raw := *(*[16]byte)(unsafe.Pointer(&x))
	var hi, lo uint64
	for b := 0; b < 8; b++ {
		lo |= uint64(raw[b]) << (8 * b)
		hi |= uint64(raw[8+b]) << (8 * b)
	}
	return fmt.Sprintf("%016x%016x", hi, lo)
}
