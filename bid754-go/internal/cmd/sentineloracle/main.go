// Command sentineloracle answers pin-time oracle queries for the devtools
// Tier 1 routing-sentinel codegen over a stdin/stdout line protocol.
//
// The sentinel codegen (devtools/internal/testgen/tier1_sentinel_codegen.go
// and tier1_sentinel_cc_codegen.go) pins expected (result bits, raw flags)
// rows computed through the public bid754-go API, which the publicroute gate
// proves routes through the Go mechanical port. devtools must not require
// any public module (docs/SPEC.md inter-component dependency rules), so the
// codegen runs this command as a subprocess inside the bid754-go module
// directory — a filesystem relationship, not a module dependency — and
// receives the oracle answers over the protocol below.
//
// Protocol: one request per line, one response per line, in request order.
// <value> operands use the canonical sentinel width text (d32 "%08x",
// d64 "%016x", d128 "%016x:%016x" as hi:lo); <mode> is the native Intel
// rounding-mode integer; <exact> is 0 or 1; <register> is a decimal uint64.
//
//	str <width> <value>                       -> ok <decimal string>
//	rounded <width> <op> <mode> <x> <y>       -> ok <bits>/<rawflags>
//	mixed <function> <mode> <x> <y>           -> ok <bits>/<rawflags>
//	unrounded <width> <op> <x> <y>            -> ok <bits>/<rawflags>
//	fma <width> <mode> <x> <y> <z>            -> ok <bits>/<rawflags>
//	sqrt <width> <mode> <x>                   -> ok <bits>/<rawflags>
//	roundintexact <width> <mode> <x>          -> ok <bits>/<rawflags>
//	roundint <width> <variant> <x>            -> ok <bits>/<rawflags>
//	next <width> <up|down> <x>                -> ok <bits>/<rawflags>
//	logb <width> <x>                          -> ok <bits>/<rawflags>
//	scaleb <width> <mode> <n> <x>             -> ok <bits>/<rawflags>
//	quiet <width> <op> <x> <y>                -> ok <00|01>/<rawflags>
//	minmax <width> <op> <x> <y>               -> ok <bits>/<rawflags>
//	toint <width> <kind> <exact> <mode> <x>   -> ok <register>/<rawflags>
//	widthconv <source> <dest> <mode> <x>      -> ok <bits>/<rawflags>
//	binaryconv <source> <dest> <mode> <x>     -> ok <bits>/<rawflags>
//	constructor <dest> <kind> <mode> <register> -> ok <bits>[/<rawflags>]
//
// Any failure answers "err <single-line message>"; the codegen fails its
// generation run on any err or transport failure (no fallback, no partial
// output). EOF on stdin ends the process with exit 0.
package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	bid754 "github.com/sky1core/bid754/bid754-go"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	for {
		line, readErr := in.ReadString('\n')
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		if line != "" {
			payload, err := evalRequest(line)
			if err != nil {
				fmt.Fprintf(out, "err %s\n", strings.ReplaceAll(err.Error(), "\n", "; "))
			} else {
				fmt.Fprintf(out, "ok %s\n", payload)
			}
			if err := out.Flush(); err != nil {
				fmt.Fprintf(os.Stderr, "sentineloracle: write response: %v\n", err)
				os.Exit(1)
			}
		}
		if readErr == io.EOF {
			return
		}
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "sentineloracle: read request: %v\n", readErr)
			os.Exit(1)
		}
	}
}

// oracleValue carries one decoded operand as raw width bits (lo/hi words,
// matching the devtools bid128BidCodecValue layout).
type oracleValue struct {
	lo, hi uint64
}

func parseHexWord(text string, bits int) (uint64, error) {
	if len(text) != bits/4 {
		return 0, fmt.Errorf("value %q is not %d hex digits", text, bits/4)
	}
	value, err := strconv.ParseUint(text, 16, bits)
	if err != nil {
		return 0, fmt.Errorf("value %q is not valid hex: %v", text, err)
	}
	return value, nil
}

func parseValue(width int, text string) (oracleValue, error) {
	switch width {
	case 32:
		lo, err := parseHexWord(text, 32)
		return oracleValue{lo: lo}, err
	case 64:
		lo, err := parseHexWord(text, 64)
		return oracleValue{lo: lo}, err
	case 128:
		hiText, loText, ok := strings.Cut(text, ":")
		if !ok {
			return oracleValue{}, fmt.Errorf("width-128 value %q is not <hi>:<lo>", text)
		}
		hi, err := parseHexWord(hiText, 64)
		if err != nil {
			return oracleValue{}, err
		}
		lo, err := parseHexWord(loText, 64)
		if err != nil {
			return oracleValue{}, err
		}
		return oracleValue{lo: lo, hi: hi}, nil
	default:
		return oracleValue{}, fmt.Errorf("unsupported tier1 sentinel width %d", width)
	}
}

func parseWidth(text string) (int, error) {
	width, err := strconv.Atoi(text)
	if err != nil {
		return 0, fmt.Errorf("width %q is not an integer", text)
	}
	switch width {
	case 32, 64, 128:
		return width, nil
	default:
		return 0, fmt.Errorf("unsupported tier1 sentinel width %d", width)
	}
}

// parseMode maps the native Intel rounding-mode integer onto the public
// RoundingMode. This pairing is the oracle-owned half of the native<->public
// correspondence; the devtools mode table carries the same native order, and
// the generated runners re-verify every pinned row against live Intel C, so
// a miswired pairing fails the runtime comparison (false-fail direction,
// never false-pass).
func parseMode(text string) (bid754.RoundingMode, error) {
	native, err := strconv.Atoi(text)
	if err != nil {
		return 0, fmt.Errorf("rounding mode %q is not an integer", text)
	}
	switch native {
	case 0:
		return bid754.RoundNearestEven, nil
	case 4:
		return bid754.RoundNearestAway, nil
	case 3:
		return bid754.RoundTowardZero, nil
	case 2:
		return bid754.RoundTowardPositive, nil
	case 1:
		return bid754.RoundTowardNegative, nil
	default:
		return 0, fmt.Errorf("unknown native rounding mode %d", native)
	}
}

// rawFlags mirrors the generated runners' public_raw_flags mapping onto the
// Intel raw flag bits. Any public flag outside the five Intel-visible bits
// fails the request (and with it the generation run).
func rawFlags(flags bid754.ExceptionFlags) (uint32, error) {
	var raw uint32
	if flags&bid754.FlagInvalidOperation != 0 {
		raw |= 0x01
	}
	if flags&bid754.FlagDivisionByZero != 0 {
		raw |= 0x04
	}
	if flags&bid754.FlagOverflow != 0 {
		raw |= 0x08
	}
	if flags&bid754.FlagUnderflow != 0 {
		raw |= 0x10
	}
	if flags&bid754.FlagInexact != 0 {
		raw |= 0x20
	}
	known := bid754.FlagInvalidOperation | bid754.FlagDivisionByZero |
		bid754.FlagOverflow | bid754.FlagUnderflow | bid754.FlagInexact
	if unknown := flags &^ known; unknown != 0 {
		return 0, fmt.Errorf("tier1 sentinel oracle produced flags outside the Intel raw set: %s", unknown)
	}
	return raw, nil
}

func decimal128(v oracleValue) bid754.Decimal128BID {
	var out bid754.Decimal128BID
	binary.LittleEndian.PutUint64(out[0:8], v.lo)
	binary.LittleEndian.PutUint64(out[8:16], v.hi)
	return out
}

func result32(value bid754.Decimal32BID, flags bid754.ExceptionFlags) (string, error) {
	raw, err := rawFlags(flags)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%08x/%08x", value.ToUint32(), raw), nil
}

func result64(value bid754.Decimal64BID, flags bid754.ExceptionFlags) (string, error) {
	raw, err := rawFlags(flags)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%016x/%08x", value.ToUint64(), raw), nil
}

func result128(value bid754.Decimal128BID, flags bid754.ExceptionFlags) (string, error) {
	raw, err := rawFlags(flags)
	if err != nil {
		return "", err
	}
	valueBytes := value.ToBytes()
	lo := binary.LittleEndian.Uint64(valueBytes[0:8])
	hi := binary.LittleEndian.Uint64(valueBytes[8:16])
	return fmt.Sprintf("%016x:%016x/%08x", hi, lo, raw), nil
}

func boolText(value bool) string {
	if value {
		return "01"
	}
	return "00"
}

// requireFields checks the per-request argument count after the leading verb.
func requireFields(fields []string, want int) error {
	if len(fields)-1 != want {
		return fmt.Errorf("request %q wants %d arguments, got %d", fields[0], want, len(fields)-1)
	}
	return nil
}

func evalRequest(line string) (string, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", fmt.Errorf("empty request")
	}
	switch fields[0] {
	case "str":
		if err := requireFields(fields, 2); err != nil {
			return "", err
		}
		width, err := parseWidth(fields[1])
		if err != nil {
			return "", err
		}
		v, err := parseValue(width, fields[2])
		if err != nil {
			return "", err
		}
		switch width {
		case 32:
			return bid754.Decimal32BID(uint32(v.lo)).String(), nil
		case 64:
			return bid754.Decimal64BID(v.lo).String(), nil
		default:
			return decimal128(v).String(), nil
		}
	case "rounded":
		if err := requireFields(fields, 5); err != nil {
			return "", err
		}
		width, err := parseWidth(fields[1])
		if err != nil {
			return "", err
		}
		mode, err := parseMode(fields[3])
		if err != nil {
			return "", err
		}
		x, err := parseValue(width, fields[4])
		if err != nil {
			return "", err
		}
		y, err := parseValue(width, fields[5])
		if err != nil {
			return "", err
		}
		return evalRounded(width, fields[2], x, y, mode)
	case "mixed":
		if err := requireFields(fields, 4); err != nil {
			return "", err
		}
		mode, err := parseMode(fields[2])
		if err != nil {
			return "", err
		}
		return evalMixed(fields[1], fields[3], fields[4], mode)
	case "unrounded":
		if err := requireFields(fields, 4); err != nil {
			return "", err
		}
		width, err := parseWidth(fields[1])
		if err != nil {
			return "", err
		}
		x, err := parseValue(width, fields[3])
		if err != nil {
			return "", err
		}
		y, err := parseValue(width, fields[4])
		if err != nil {
			return "", err
		}
		return evalUnrounded(width, fields[2], x, y)
	case "fma":
		if err := requireFields(fields, 5); err != nil {
			return "", err
		}
		width, err := parseWidth(fields[1])
		if err != nil {
			return "", err
		}
		mode, err := parseMode(fields[2])
		if err != nil {
			return "", err
		}
		x, err := parseValue(width, fields[3])
		if err != nil {
			return "", err
		}
		y, err := parseValue(width, fields[4])
		if err != nil {
			return "", err
		}
		z, err := parseValue(width, fields[5])
		if err != nil {
			return "", err
		}
		return evalFma(width, x, y, z, mode)
	case "sqrt":
		if err := requireFields(fields, 3); err != nil {
			return "", err
		}
		width, err := parseWidth(fields[1])
		if err != nil {
			return "", err
		}
		mode, err := parseMode(fields[2])
		if err != nil {
			return "", err
		}
		x, err := parseValue(width, fields[3])
		if err != nil {
			return "", err
		}
		return evalSqrt(width, x, mode)
	case "roundintexact":
		if err := requireFields(fields, 3); err != nil {
			return "", err
		}
		width, err := parseWidth(fields[1])
		if err != nil {
			return "", err
		}
		mode, err := parseMode(fields[2])
		if err != nil {
			return "", err
		}
		x, err := parseValue(width, fields[3])
		if err != nil {
			return "", err
		}
		return evalRoundIntegralExact(width, x, mode)
	case "roundint":
		if err := requireFields(fields, 3); err != nil {
			return "", err
		}
		width, err := parseWidth(fields[1])
		if err != nil {
			return "", err
		}
		x, err := parseValue(width, fields[3])
		if err != nil {
			return "", err
		}
		return evalRoundIntegralFixed(width, fields[2], x)
	case "next":
		if err := requireFields(fields, 3); err != nil {
			return "", err
		}
		width, err := parseWidth(fields[1])
		if err != nil {
			return "", err
		}
		x, err := parseValue(width, fields[3])
		if err != nil {
			return "", err
		}
		return evalNext(width, fields[2], x)
	case "logb":
		if err := requireFields(fields, 2); err != nil {
			return "", err
		}
		width, err := parseWidth(fields[1])
		if err != nil {
			return "", err
		}
		x, err := parseValue(width, fields[2])
		if err != nil {
			return "", err
		}
		return evalLogB(width, x)
	case "scaleb":
		if err := requireFields(fields, 4); err != nil {
			return "", err
		}
		width, err := parseWidth(fields[1])
		if err != nil {
			return "", err
		}
		mode, err := parseMode(fields[2])
		if err != nil {
			return "", err
		}
		n, err := strconv.ParseInt(fields[3], 10, 64)
		if err != nil {
			return "", fmt.Errorf("scaleb exponent %q is not an integer", fields[3])
		}
		x, err := parseValue(width, fields[4])
		if err != nil {
			return "", err
		}
		return evalScale(width, x, n, mode)
	case "quiet":
		if err := requireFields(fields, 4); err != nil {
			return "", err
		}
		width, err := parseWidth(fields[1])
		if err != nil {
			return "", err
		}
		x, err := parseValue(width, fields[3])
		if err != nil {
			return "", err
		}
		y, err := parseValue(width, fields[4])
		if err != nil {
			return "", err
		}
		return evalQuiet(width, fields[2], x, y)
	case "minmax":
		if err := requireFields(fields, 4); err != nil {
			return "", err
		}
		width, err := parseWidth(fields[1])
		if err != nil {
			return "", err
		}
		x, err := parseValue(width, fields[3])
		if err != nil {
			return "", err
		}
		y, err := parseValue(width, fields[4])
		if err != nil {
			return "", err
		}
		return evalMinMax(width, fields[2], x, y)
	case "toint":
		if err := requireFields(fields, 5); err != nil {
			return "", err
		}
		width, err := parseWidth(fields[1])
		if err != nil {
			return "", err
		}
		exact, err := parseExact(fields[3])
		if err != nil {
			return "", err
		}
		mode, err := parseMode(fields[4])
		if err != nil {
			return "", err
		}
		x, err := parseValue(width, fields[5])
		if err != nil {
			return "", err
		}
		return evalToInt(width, fields[2], exact, mode, x)
	case "widthconv":
		if err := requireFields(fields, 4); err != nil {
			return "", err
		}
		source, err := parseWidth(fields[1])
		if err != nil {
			return "", err
		}
		dest, err := parseWidth(fields[2])
		if err != nil {
			return "", err
		}
		mode, err := parseMode(fields[3])
		if err != nil {
			return "", err
		}
		x, err := parseValue(source, fields[4])
		if err != nil {
			return "", err
		}
		return evalWidthConversion(source, dest, x, mode)
	case "binaryconv":
		if err := requireFields(fields, 4); err != nil {
			return "", err
		}
		source, err := parseWidth(fields[1])
		if err != nil {
			return "", err
		}
		dest, err := parseWidth(fields[2])
		if err != nil {
			return "", err
		}
		mode, err := parseMode(fields[3])
		if err != nil {
			return "", err
		}
		x, err := parseValue(source, fields[4])
		if err != nil {
			return "", err
		}
		return evalBinaryConversion(source, dest, x, mode)
	case "constructor":
		if err := requireFields(fields, 4); err != nil {
			return "", err
		}
		dest, err := parseWidth(fields[1])
		if err != nil {
			return "", err
		}
		mode, err := parseMode(fields[3])
		if err != nil {
			return "", err
		}
		register, err := strconv.ParseUint(fields[4], 10, 64)
		if err != nil {
			return "", fmt.Errorf("constructor register %q is not a uint64", fields[4])
		}
		return evalConstructor(dest, fields[2], mode, register)
	default:
		return "", fmt.Errorf("unknown request %q", fields[0])
	}
}

func parseExact(text string) (bool, error) {
	switch text {
	case "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return false, fmt.Errorf("exact marker %q is not 0 or 1", text)
	}
}

func evalRounded(width int, op string, x, y oracleValue, mode bid754.RoundingMode) (string, error) {
	switch width {
	case 32:
		left, right := bid754.Decimal32BID(uint32(x.lo)), bid754.Decimal32BID(uint32(y.lo))
		var value bid754.Decimal32BID
		var flags bid754.ExceptionFlags
		switch op {
		case "add":
			value, flags = left.AddWithMode(right, mode)
		case "sub":
			value, flags = left.SubWithMode(right, mode)
		case "mul":
			value, flags = left.MulWithMode(right, mode)
		case "div":
			value, flags = left.DivWithMode(right, mode)
		case "quantize":
			value, flags = left.QuantizeWithMode(right, mode)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel rounded operation %q", op)
		}
		return result32(value, flags)
	case 64:
		left, right := bid754.Decimal64BID(x.lo), bid754.Decimal64BID(y.lo)
		var value bid754.Decimal64BID
		var flags bid754.ExceptionFlags
		switch op {
		case "add":
			value, flags = left.AddWithMode(right, mode)
		case "sub":
			value, flags = left.SubWithMode(right, mode)
		case "mul":
			value, flags = left.MulWithMode(right, mode)
		case "div":
			value, flags = left.DivWithMode(right, mode)
		case "quantize":
			value, flags = left.QuantizeWithMode(right, mode)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel rounded operation %q", op)
		}
		return result64(value, flags)
	case 128:
		left, right := decimal128(x), decimal128(y)
		var value bid754.Decimal128BID
		var flags bid754.ExceptionFlags
		switch op {
		case "add":
			value, flags = left.AddWithMode(right, mode)
		case "sub":
			value, flags = left.SubWithMode(right, mode)
		case "mul":
			value, flags = left.MulWithMode(right, mode)
		case "div":
			value, flags = left.DivWithMode(right, mode)
		case "quantize":
			value, flags = left.QuantizeWithMode(right, mode)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel rounded operation %q", op)
		}
		return result128(value, flags)
	default:
		return "", fmt.Errorf("unsupported tier1 sentinel width %d", width)
	}
}

// mixedRoutingSentinelOperandWidth returns the operand width for one supported
// mixed-format routing-sentinel function. Only the four equal-width
// non-commutative CORE functions are wired (bid64qq_sub/div read two
// Decimal128 operands, bid128dd_sub/div read two Decimal64 operands); any
// other function fails the request and, with it, the generation run — never a
// silent fallback.
func mixedRoutingSentinelOperandWidth(function string) (int, error) {
	switch function {
	case "bid64qq_sub", "bid64qq_div":
		return 128, nil
	case "bid128dd_sub", "bid128dd_div":
		return 64, nil
	default:
		return 0, fmt.Errorf("unsupported mixed routing-sentinel function %q", function)
	}
}

// evalMixed answers `mixed <function> <mode> <x> <y>` through the public mixed
// bid754-go API (the publicroute gate proves that surface routes through the
// Go mechanical port). Operands are decoded at the function's operand width;
// bid64qq results are Decimal64, bid128dd results are Decimal128.
func evalMixed(function, xText, yText string, mode bid754.RoundingMode) (string, error) {
	operandWidth, err := mixedRoutingSentinelOperandWidth(function)
	if err != nil {
		return "", err
	}
	x, err := parseValue(operandWidth, xText)
	if err != nil {
		return "", err
	}
	y, err := parseValue(operandWidth, yText)
	if err != nil {
		return "", err
	}
	switch function {
	case "bid64qq_sub":
		value, flags := bid754.Sub64QQBIDWithMode(decimal128(x), decimal128(y), mode)
		return result64(value, flags)
	case "bid64qq_div":
		value, flags := bid754.Div64QQBIDWithMode(decimal128(x), decimal128(y), mode)
		return result64(value, flags)
	case "bid128dd_sub":
		value, flags := bid754.Sub128DDBIDWithMode(bid754.Decimal64BID(x.lo), bid754.Decimal64BID(y.lo), mode)
		return result128(value, flags)
	case "bid128dd_div":
		value, flags := bid754.Div128DDBIDWithMode(bid754.Decimal64BID(x.lo), bid754.Decimal64BID(y.lo), mode)
		return result128(value, flags)
	default:
		return "", fmt.Errorf("unsupported mixed routing-sentinel function %q", function)
	}
}

func evalUnrounded(width int, op string, x, y oracleValue) (string, error) {
	switch width {
	case 32:
		left, right := bid754.Decimal32BID(uint32(x.lo)), bid754.Decimal32BID(uint32(y.lo))
		var value bid754.Decimal32BID
		var flags bid754.ExceptionFlags
		switch op {
		case "remainder":
			value, flags = left.Remainder(right)
		case "fmod":
			value, flags = left.Fmod(right)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel unrounded operation %q", op)
		}
		return result32(value, flags)
	case 64:
		left, right := bid754.Decimal64BID(x.lo), bid754.Decimal64BID(y.lo)
		var value bid754.Decimal64BID
		var flags bid754.ExceptionFlags
		switch op {
		case "remainder":
			value, flags = left.Remainder(right)
		case "fmod":
			value, flags = left.Fmod(right)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel unrounded operation %q", op)
		}
		return result64(value, flags)
	case 128:
		left, right := decimal128(x), decimal128(y)
		var value bid754.Decimal128BID
		var flags bid754.ExceptionFlags
		switch op {
		case "remainder":
			value, flags = left.Remainder(right)
		case "fmod":
			value, flags = left.Fmod(right)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel unrounded operation %q", op)
		}
		return result128(value, flags)
	default:
		return "", fmt.Errorf("unsupported tier1 sentinel width %d", width)
	}
}

func evalFma(width int, x, y, z oracleValue, mode bid754.RoundingMode) (string, error) {
	switch width {
	case 32:
		value, flags := bid754.Decimal32BID(uint32(x.lo)).FMAWithMode(bid754.Decimal32BID(uint32(y.lo)), bid754.Decimal32BID(uint32(z.lo)), mode)
		return result32(value, flags)
	case 64:
		value, flags := bid754.Decimal64BID(x.lo).FMAWithMode(bid754.Decimal64BID(y.lo), bid754.Decimal64BID(z.lo), mode)
		return result64(value, flags)
	case 128:
		value, flags := decimal128(x).FMAWithMode(decimal128(y), decimal128(z), mode)
		return result128(value, flags)
	default:
		return "", fmt.Errorf("unsupported tier1 sentinel width %d", width)
	}
}

func evalSqrt(width int, x oracleValue, mode bid754.RoundingMode) (string, error) {
	switch width {
	case 32:
		value, flags := bid754.Decimal32BID(uint32(x.lo)).SqrtWithMode(mode)
		return result32(value, flags)
	case 64:
		value, flags := bid754.Decimal64BID(x.lo).SqrtWithMode(mode)
		return result64(value, flags)
	case 128:
		value, flags := decimal128(x).SqrtWithMode(mode)
		return result128(value, flags)
	default:
		return "", fmt.Errorf("unsupported tier1 sentinel width %d", width)
	}
}

// evalRoundIntegralExact answers `roundintexact <width> <mode> <x>` through
// the public mode-taking round-to-integral-exact surface. The d32 exhaustive
// sentinel codegen is the only current consumer, so only width 32 is wired;
// widening it is an explicit follow-up, not a silent fallback.
func evalRoundIntegralExact(width int, x oracleValue, mode bid754.RoundingMode) (string, error) {
	switch width {
	case 32:
		value, flags := bid754.Decimal32BID(uint32(x.lo)).RoundIntegralExactWithMode(mode)
		return result32(value, flags)
	default:
		return "", fmt.Errorf("roundintexact oracle supports width 32 only (d32 exhaustive sentinel scope), got %d", width)
	}
}

// evalRoundIntegralFixed answers `roundint <width> <variant> <x>` through the
// public fixed-attribute round-to-integral surfaces (the Intel
// bid32_round_integral_<variant> family). Width scope matches
// evalRoundIntegralExact.
func evalRoundIntegralFixed(width int, variant string, x oracleValue) (string, error) {
	if width != 32 {
		return "", fmt.Errorf("roundint oracle supports width 32 only (d32 exhaustive sentinel scope), got %d", width)
	}
	value := bid754.Decimal32BID(uint32(x.lo))
	switch variant {
	case "nearest_even":
		result, flags := value.RoundIntegralNearestEven()
		return result32(result, flags)
	case "nearest_away":
		result, flags := value.RoundIntegralNearestAway()
		return result32(result, flags)
	case "zero":
		result, flags := value.RoundIntegralZero()
		return result32(result, flags)
	case "positive":
		result, flags := value.RoundIntegralPositive()
		return result32(result, flags)
	case "negative":
		result, flags := value.RoundIntegralNegative()
		return result32(result, flags)
	default:
		return "", fmt.Errorf("unknown roundint variant %q", variant)
	}
}

// evalNext answers `next <width> <up|down> <x>` through the public
// NextPlus/NextMinus surfaces (the Intel bid32_nextup/bid32_nextdown family).
// The d32 exhaustive sentinel codegen is the only current consumer, so only
// width 32 is wired; widening it is an explicit follow-up, not a silent
// fallback.
func evalNext(width int, direction string, x oracleValue) (string, error) {
	if width != 32 {
		return "", fmt.Errorf("next oracle supports width 32 only (d32 exhaustive sentinel scope), got %d", width)
	}
	value := bid754.Decimal32BID(uint32(x.lo))
	switch direction {
	case "up":
		result, flags := value.NextPlus()
		return result32(result, flags)
	case "down":
		result, flags := value.NextMinus()
		return result32(result, flags)
	default:
		return "", fmt.Errorf("unknown next direction %q (want up or down)", direction)
	}
}

// evalLogB answers `logb <width> <x>` through the public LogB surface (the
// Intel bid32_logb family). Width scope matches evalNext.
func evalLogB(width int, x oracleValue) (string, error) {
	if width != 32 {
		return "", fmt.Errorf("logb oracle supports width 32 only (d32 exhaustive sentinel scope), got %d", width)
	}
	value, flags := bid754.Decimal32BID(uint32(x.lo)).LogB()
	return result32(value, flags)
}

func evalScale(width int, x oracleValue, n int64, mode bid754.RoundingMode) (string, error) {
	switch width {
	case 32:
		value, flags := bid754.Decimal32BID(uint32(x.lo)).ScaleBWithMode(int(n), mode)
		return result32(value, flags)
	case 64:
		value, flags := bid754.Decimal64BID(x.lo).ScaleBWithMode(int(n), mode)
		return result64(value, flags)
	case 128:
		value, flags := decimal128(x).ScaleBWithMode(int(n), mode)
		return result128(value, flags)
	default:
		return "", fmt.Errorf("unsupported tier1 sentinel width %d", width)
	}
}

func evalQuiet(width int, op string, x, y oracleValue) (string, error) {
	var got bool
	var flags bid754.ExceptionFlags
	switch width {
	case 32:
		left, right := bid754.Decimal32BID(uint32(x.lo)), bid754.Decimal32BID(uint32(y.lo))
		switch op {
		case "quiet_equal":
			got, flags = left.QuietEqual(right)
		case "quiet_not_equal":
			got, flags = left.QuietNotEqual(right)
		case "quiet_less":
			got, flags = left.QuietLess(right)
		case "quiet_less_equal":
			got, flags = left.QuietLessEqual(right)
		case "quiet_greater":
			got, flags = left.QuietGreater(right)
		case "quiet_greater_equal":
			got, flags = left.QuietGreaterEqual(right)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel quiet predicate %q", op)
		}
	case 64:
		left, right := bid754.Decimal64BID(x.lo), bid754.Decimal64BID(y.lo)
		switch op {
		case "quiet_equal":
			got, flags = left.QuietEqual(right)
		case "quiet_not_equal":
			got, flags = left.QuietNotEqual(right)
		case "quiet_less":
			got, flags = left.QuietLess(right)
		case "quiet_less_equal":
			got, flags = left.QuietLessEqual(right)
		case "quiet_greater":
			got, flags = left.QuietGreater(right)
		case "quiet_greater_equal":
			got, flags = left.QuietGreaterEqual(right)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel quiet predicate %q", op)
		}
	case 128:
		left, right := decimal128(x), decimal128(y)
		switch op {
		case "quiet_equal":
			got, flags = left.QuietEqual(right)
		case "quiet_not_equal":
			got, flags = left.QuietNotEqual(right)
		case "quiet_less":
			got, flags = left.QuietLess(right)
		case "quiet_less_equal":
			got, flags = left.QuietLessEqual(right)
		case "quiet_greater":
			got, flags = left.QuietGreater(right)
		case "quiet_greater_equal":
			got, flags = left.QuietGreaterEqual(right)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel quiet predicate %q", op)
		}
	default:
		return "", fmt.Errorf("unsupported tier1 sentinel width %d", width)
	}
	raw, err := rawFlags(flags)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%08x", boolText(got), raw), nil
}

func evalMinMax(width int, op string, x, y oracleValue) (string, error) {
	switch width {
	case 32:
		left, right := bid754.Decimal32BID(uint32(x.lo)), bid754.Decimal32BID(uint32(y.lo))
		var value bid754.Decimal32BID
		var flags bid754.ExceptionFlags
		switch op {
		case "minnum":
			value, flags = left.MinNum(right)
		case "maxnum":
			value, flags = left.MaxNum(right)
		case "minnum_mag":
			value, flags = left.MinNumMag(right)
		case "maxnum_mag":
			value, flags = left.MaxNumMag(right)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel min/max operation %q", op)
		}
		return result32(value, flags)
	case 64:
		left, right := bid754.Decimal64BID(x.lo), bid754.Decimal64BID(y.lo)
		var value bid754.Decimal64BID
		var flags bid754.ExceptionFlags
		switch op {
		case "minnum":
			value, flags = left.MinNum(right)
		case "maxnum":
			value, flags = left.MaxNum(right)
		case "minnum_mag":
			value, flags = left.MinNumMag(right)
		case "maxnum_mag":
			value, flags = left.MaxNumMag(right)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel min/max operation %q", op)
		}
		return result64(value, flags)
	case 128:
		left, right := decimal128(x), decimal128(y)
		var value bid754.Decimal128BID
		var flags bid754.ExceptionFlags
		switch op {
		case "minnum":
			value, flags = left.MinNum(right)
		case "maxnum":
			value, flags = left.MaxNum(right)
		case "minnum_mag":
			value, flags = left.MinNumMag(right)
		case "maxnum_mag":
			value, flags = left.MaxNumMag(right)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel min/max operation %q", op)
		}
		return result128(value, flags)
	default:
		return "", fmt.Errorf("unsupported tier1 sentinel width %d", width)
	}
}

// evalToInt normalizes every integer conversion to the cross-language
// canonical form: the result as a 64-bit two's-complement register (Rust's
// `as i64 as u64` convention) plus the raw flag word.
func evalToInt(width int, kind string, exact bool, mode bid754.RoundingMode, x oracleValue) (string, error) {
	var register uint64
	var flags bid754.ExceptionFlags
	eval32 := func(value bid754.Decimal32BID) error {
		switch {
		case kind == "int8" && !exact:
			r, f := value.ConvertToInt8(mode)
			register, flags = uint64(int64(r)), f
		case kind == "int8":
			r, f := value.ConvertToInt8Exact(mode)
			register, flags = uint64(int64(r)), f
		case kind == "int16" && !exact:
			r, f := value.ConvertToInt16(mode)
			register, flags = uint64(int64(r)), f
		case kind == "int16":
			r, f := value.ConvertToInt16Exact(mode)
			register, flags = uint64(int64(r)), f
		case kind == "int32" && !exact:
			r, f := value.ConvertToInt32(mode)
			register, flags = uint64(int64(r)), f
		case kind == "int32":
			r, f := value.ConvertToInt32Exact(mode)
			register, flags = uint64(int64(r)), f
		case kind == "int64" && !exact:
			r, f := value.ConvertToInt64(mode)
			register, flags = uint64(r), f
		case kind == "int64":
			r, f := value.ConvertToInt64Exact(mode)
			register, flags = uint64(r), f
		case kind == "uint8" && !exact:
			r, f := value.ConvertToUint8(mode)
			register, flags = uint64(r), f
		case kind == "uint8":
			r, f := value.ConvertToUint8Exact(mode)
			register, flags = uint64(r), f
		case kind == "uint16" && !exact:
			r, f := value.ConvertToUint16(mode)
			register, flags = uint64(r), f
		case kind == "uint16":
			r, f := value.ConvertToUint16Exact(mode)
			register, flags = uint64(r), f
		case kind == "uint32" && !exact:
			r, f := value.ConvertToUint32(mode)
			register, flags = uint64(r), f
		case kind == "uint32":
			r, f := value.ConvertToUint32Exact(mode)
			register, flags = uint64(r), f
		case kind == "uint64" && !exact:
			r, f := value.ConvertToUint64(mode)
			register, flags = r, f
		case kind == "uint64":
			r, f := value.ConvertToUint64Exact(mode)
			register, flags = r, f
		default:
			return fmt.Errorf("unknown tier1 sentinel to-int kind %q", kind)
		}
		return nil
	}
	eval64 := func(value bid754.Decimal64BID) error {
		switch {
		case kind == "int8" && !exact:
			r, f := value.ConvertToInt8(mode)
			register, flags = uint64(int64(r)), f
		case kind == "int8":
			r, f := value.ConvertToInt8Exact(mode)
			register, flags = uint64(int64(r)), f
		case kind == "int16" && !exact:
			r, f := value.ConvertToInt16(mode)
			register, flags = uint64(int64(r)), f
		case kind == "int16":
			r, f := value.ConvertToInt16Exact(mode)
			register, flags = uint64(int64(r)), f
		case kind == "int32" && !exact:
			r, f := value.ConvertToInt32(mode)
			register, flags = uint64(int64(r)), f
		case kind == "int32":
			r, f := value.ConvertToInt32Exact(mode)
			register, flags = uint64(int64(r)), f
		case kind == "int64" && !exact:
			r, f := value.ConvertToInt64(mode)
			register, flags = uint64(r), f
		case kind == "int64":
			r, f := value.ConvertToInt64Exact(mode)
			register, flags = uint64(r), f
		case kind == "uint8" && !exact:
			r, f := value.ConvertToUint8(mode)
			register, flags = uint64(r), f
		case kind == "uint8":
			r, f := value.ConvertToUint8Exact(mode)
			register, flags = uint64(r), f
		case kind == "uint16" && !exact:
			r, f := value.ConvertToUint16(mode)
			register, flags = uint64(r), f
		case kind == "uint16":
			r, f := value.ConvertToUint16Exact(mode)
			register, flags = uint64(r), f
		case kind == "uint32" && !exact:
			r, f := value.ConvertToUint32(mode)
			register, flags = uint64(r), f
		case kind == "uint32":
			r, f := value.ConvertToUint32Exact(mode)
			register, flags = uint64(r), f
		case kind == "uint64" && !exact:
			r, f := value.ConvertToUint64(mode)
			register, flags = r, f
		case kind == "uint64":
			r, f := value.ConvertToUint64Exact(mode)
			register, flags = r, f
		default:
			return fmt.Errorf("unknown tier1 sentinel to-int kind %q", kind)
		}
		return nil
	}
	eval128 := func(value bid754.Decimal128BID) error {
		switch {
		case kind == "int8" && !exact:
			r, f := value.ConvertToInt8(mode)
			register, flags = uint64(int64(r)), f
		case kind == "int8":
			r, f := value.ConvertToInt8Exact(mode)
			register, flags = uint64(int64(r)), f
		case kind == "int16" && !exact:
			r, f := value.ConvertToInt16(mode)
			register, flags = uint64(int64(r)), f
		case kind == "int16":
			r, f := value.ConvertToInt16Exact(mode)
			register, flags = uint64(int64(r)), f
		case kind == "int32" && !exact:
			r, f := value.ConvertToInt32(mode)
			register, flags = uint64(int64(r)), f
		case kind == "int32":
			r, f := value.ConvertToInt32Exact(mode)
			register, flags = uint64(int64(r)), f
		case kind == "int64" && !exact:
			r, f := value.ConvertToInt64(mode)
			register, flags = uint64(r), f
		case kind == "int64":
			r, f := value.ConvertToInt64Exact(mode)
			register, flags = uint64(r), f
		case kind == "uint8" && !exact:
			r, f := value.ConvertToUint8(mode)
			register, flags = uint64(r), f
		case kind == "uint8":
			r, f := value.ConvertToUint8Exact(mode)
			register, flags = uint64(r), f
		case kind == "uint16" && !exact:
			r, f := value.ConvertToUint16(mode)
			register, flags = uint64(r), f
		case kind == "uint16":
			r, f := value.ConvertToUint16Exact(mode)
			register, flags = uint64(r), f
		case kind == "uint32" && !exact:
			r, f := value.ConvertToUint32(mode)
			register, flags = uint64(r), f
		case kind == "uint32":
			r, f := value.ConvertToUint32Exact(mode)
			register, flags = uint64(r), f
		case kind == "uint64" && !exact:
			r, f := value.ConvertToUint64(mode)
			register, flags = r, f
		case kind == "uint64":
			r, f := value.ConvertToUint64Exact(mode)
			register, flags = r, f
		default:
			return fmt.Errorf("unknown tier1 sentinel to-int kind %q", kind)
		}
		return nil
	}
	var err error
	switch width {
	case 32:
		err = eval32(bid754.Decimal32BID(uint32(x.lo)))
	case 64:
		err = eval64(bid754.Decimal64BID(x.lo))
	case 128:
		err = eval128(decimal128(x))
	default:
		return "", fmt.Errorf("unsupported tier1 sentinel width %d", width)
	}
	if err != nil {
		return "", err
	}
	raw, err := rawFlags(flags)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%016x/%08x", register, raw), nil
}

func evalWidthConversion(source, dest int, x oracleValue, mode bid754.RoundingMode) (string, error) {
	switch source {
	case 32:
		value := bid754.Decimal32BID(uint32(x.lo))
		switch dest {
		case 64:
			result, flags := value.ToDecimal64()
			return result64(result, flags)
		case 128:
			result, flags := value.ToDecimal128()
			return result128(result, flags)
		}
	case 64:
		value := bid754.Decimal64BID(x.lo)
		switch dest {
		case 32:
			result, flags := value.ToDecimal32(mode)
			return result32(result, flags)
		case 128:
			result, flags := value.ToDecimal128()
			return result128(result, flags)
		}
	case 128:
		value := decimal128(x)
		switch dest {
		case 32:
			result, flags := value.ToDecimal32(mode)
			return result32(result, flags)
		case 64:
			result, flags := value.ToDecimal64(mode)
			return result64(result, flags)
		}
	}
	return "", fmt.Errorf("unknown tier1 sentinel width conversion source %d dest %d", source, dest)
}

func binary128Text(value bid754.Binary128, raw uint32) string {
	valueBytes := value.ToBytes()
	lo := binary.LittleEndian.Uint64(valueBytes[0:8])
	hi := binary.LittleEndian.Uint64(valueBytes[8:16])
	return fmt.Sprintf("%016x:%016x/%08x", hi, lo, raw)
}

func evalBinaryConversion(source, dest int, x oracleValue, mode bid754.RoundingMode) (string, error) {
	format := func(dest int, f32 float32, f64 float64, f128 bid754.Binary128, flags bid754.ExceptionFlags) (string, error) {
		raw, err := rawFlags(flags)
		if err != nil {
			return "", err
		}
		switch dest {
		case 32:
			return fmt.Sprintf("%08x/%08x", math.Float32bits(f32), raw), nil
		case 64:
			return fmt.Sprintf("%016x/%08x", math.Float64bits(f64), raw), nil
		case 128:
			return binary128Text(f128, raw), nil
		}
		return "", fmt.Errorf("unknown tier1 sentinel binary destination %d", dest)
	}
	switch source {
	case 32:
		value := bid754.Decimal32BID(uint32(x.lo))
		switch dest {
		case 32:
			result, flags := value.ToBinary32(mode)
			return format(32, result, 0, bid754.Binary128{}, flags)
		case 64:
			result, flags := value.ToBinary64(mode)
			return format(64, 0, result, bid754.Binary128{}, flags)
		case 128:
			result, flags := value.ToBinary128(mode)
			return format(128, 0, 0, result, flags)
		}
	case 64:
		value := bid754.Decimal64BID(x.lo)
		switch dest {
		case 32:
			result, flags := value.ToBinary32(mode)
			return format(32, result, 0, bid754.Binary128{}, flags)
		case 64:
			result, flags := value.ToBinary64(mode)
			return format(64, 0, result, bid754.Binary128{}, flags)
		case 128:
			result, flags := value.ToBinary128(mode)
			return format(128, 0, 0, result, flags)
		}
	case 128:
		value := decimal128(x)
		switch dest {
		case 32:
			result, flags := value.ToBinary32(mode)
			return format(32, result, 0, bid754.Binary128{}, flags)
		case 64:
			result, flags := value.ToBinary64(mode)
			return format(64, 0, result, bid754.Binary128{}, flags)
		case 128:
			result, flags := value.ToBinary128(mode)
			return format(128, 0, 0, result, flags)
		}
	}
	return "", fmt.Errorf("unknown tier1 sentinel binary conversion source %d dest %d", source, dest)
}

func evalConstructor(dest int, kind string, mode bid754.RoundingMode, raw uint64) (string, error) {
	switch dest {
	case 32:
		var value bid754.Decimal32BID
		var flags bid754.ExceptionFlags
		switch kind {
		case "int32":
			value, flags = bid754.NewDecimal32FromInt32(int32(uint32(raw)), mode)
		case "uint32":
			value, flags = bid754.NewDecimal32FromUint32(uint32(raw), mode)
		case "int64":
			value, flags = bid754.NewDecimal32FromInt64(int64(raw), mode)
		case "uint64":
			value, flags = bid754.NewDecimal32FromUint64(raw, mode)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel constructor kind %q", kind)
		}
		return result32(value, flags)
	case 64:
		switch kind {
		case "int32":
			return fmt.Sprintf("%016x", bid754.NewDecimal64FromInt32(int32(uint32(raw))).ToUint64()), nil
		case "uint32":
			return fmt.Sprintf("%016x", bid754.NewDecimal64FromUint32(uint32(raw)).ToUint64()), nil
		case "int64":
			value, flags := bid754.NewDecimal64FromInt64(int64(raw), mode)
			return result64(value, flags)
		case "uint64":
			value, flags := bid754.NewDecimal64FromUint64(raw, mode)
			return result64(value, flags)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel constructor kind %q", kind)
		}
	case 128:
		var value bid754.Decimal128BID
		switch kind {
		case "int32":
			value = bid754.NewDecimal128FromInt32(int32(uint32(raw)))
		case "uint32":
			value = bid754.NewDecimal128FromUint32(uint32(raw))
		case "int64":
			value = bid754.NewDecimal128FromInt64(int64(raw))
		case "uint64":
			value = bid754.NewDecimal128FromUint64(raw)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel constructor kind %q", kind)
		}
		valueBytes := value.ToBytes()
		return fmt.Sprintf("%016x:%016x", binary.LittleEndian.Uint64(valueBytes[8:16]), binary.LittleEndian.Uint64(valueBytes[0:8])), nil
	default:
		return "", fmt.Errorf("unknown tier1 sentinel constructor destination %d", dest)
	}
}
