package bid754

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

type bigDecimalResult struct {
	Status      string `json:"status"`
	Coefficient string `json:"coefficient,omitempty"`
	Exp         int    `json:"exponent"`
	Reason      string `json:"reason,omitempty"`
}

type bigDecimalReference struct {
	client *finiteProcessClient
	source map[string]string
}

var bigDecimalInteger = regexp.MustCompile(`^(0|-?[1-9][0-9]*)$`)

func startBigDecimal(t *testing.T) *bigDecimalReference {
	t.Helper()
	classes, executable := os.Getenv("BID754_BIGDECIMAL_CLASSES"), os.Getenv("BID754_BIGDECIMAL_JAVA")
	if classes == "" && executable == "" {
		return nil
	}
	if classes == "" || executable == "" {
		t.Fatal("BigDecimal requires both BID754_BIGDECIMAL_CLASSES and BID754_BIGDECIMAL_JAVA")
	}
	ctx := finiteTestContext(t)
	client, err := finiteStartProcess(ctx, executable, "-cp", classes, "BigDecimalProbe")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		err := client.close()
		if ctx.Err() == context.DeadlineExceeded {
			t.Log(`FINITE-PATH-EXECUTION {"version":1,"language":"java","status":"timeout"}`)
		}
		if err != nil {
			t.Error(err)
		}
	})
	reference, err := connectBigDecimal(client, executable)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("BIGDECIMAL-ORACLE runtime=%q vendor=%q class_sha256=%s", reference.source["java_runtime"], reference.source["java_vendor"], reference.source["bigdecimal_class_sha256"])
	return reference
}

func connectBigDecimal(client *finiteProcessClient, executable string) (*bigDecimalReference, error) {
	line, err := client.readLine()
	if err != nil {
		return nil, fmt.Errorf("BigDecimal handshake: %w", err)
	}
	fields := strings.Split(strings.TrimSuffix(string(line), "\n"), "\t")
	if len(fields) != 5 || fields[0] != "1" || fields[1] != "ready" || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(fields[4]) {
		return nil, fmt.Errorf("invalid BigDecimal handshake %q", line)
	}
	source := map[string]string{"bigdecimal_class_sha256": fields[4]}
	for i, key := range []string{"java_runtime", "java_vendor"} {
		value, err := base64.StdEncoding.DecodeString(fields[i+2])
		if err != nil || len(value) == 0 {
			return nil, fmt.Errorf("invalid Java identity %q: %v", fields[i+2], err)
		}
		source[key] = string(value)
	}
	binary, err := os.ReadFile(executable)
	if err != nil {
		return nil, err
	}
	source["java_executable_sha256"] = fmt.Sprintf("%x", sha256.Sum256(binary))
	return &bigDecimalReference{client, source}, nil
}

func (reference *bigDecimalReference) evaluate(c decimalref.Case) (bigDecimalResult, error) {
	fields := []string{"1", strconv.Itoa(c.Width), c.Op, c.Mode}
	for _, raw := range c.Operands {
		d, err := decimalref.Decode(c.Width, raw)
		if err != nil {
			return bigDecimalResult{}, err
		}
		if d.Kind != "finite" {
			return bigDecimalResult{}, fmt.Errorf("BigDecimal input must be finite")
		}
		coeff := d.Coeff.String()
		if d.Negative && d.Coeff.Sign() != 0 {
			coeff = "-" + coeff
		}
		fields = append(fields, coeff, strconv.Itoa(d.Exp))
	}
	line, err := reference.client.exchangeLine([]byte(strings.Join(fields, "\t")))
	if err != nil {
		return bigDecimalResult{}, err
	}
	return decodeBigDecimalResult(c.Width, line)
}

func decodeBigDecimalResult(width int, line []byte) (bigDecimalResult, error) {
	p, err := decimalref.ParametersFor(width)
	if err != nil {
		return bigDecimalResult{}, err
	}
	parts := strings.Split(strings.TrimSuffix(string(line), "\n"), "\t")
	if len(parts) == 3 && parts[0] == "1" && parts[1] == "excluded" {
		switch parts[2] {
		case "division-by-zero", "exponent-range", "quantize-precision":
			return bigDecimalResult{Status: "excluded", Reason: parts[2]}, nil
		}
	}
	if len(parts) == 4 && parts[0] == "1" && parts[1] == "ok" && bigDecimalInteger.MatchString(parts[2]) {
		exp, err := strconv.Atoi(parts[3])
		if err == nil && strconv.Itoa(exp) == parts[3] && exp >= 2*p.MinExp && exp <= -2*p.MinExp && len(strings.TrimPrefix(parts[2], "-")) <= p.Precision {
			return bigDecimalResult{Status: "ok", Coefficient: parts[2], Exp: exp}, nil
		}
	}
	return bigDecimalResult{}, fmt.Errorf("invalid BigDecimal response %q", line)
}

func bigDecimalValue(coeff *big.Int, exponent int) *big.Rat {
	value := new(big.Rat).SetInt(coeff)
	if coeff.Sign() == 0 {
		return value
	}
	power := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(max(exponent, -exponent))), nil)
	if exponent < 0 {
		return value.Quo(value, new(big.Rat).SetInt(power))
	}
	return value.Mul(value, new(big.Rat).SetInt(power))
}

func compareBigDecimal(width int, want bigDecimalResult, observations []finiteObservation) error {
	if want.Status != "ok" {
		return fmt.Errorf("BigDecimal comparison requires an eligible result")
	}
	coeff, ok := new(big.Int).SetString(want.Coefficient, 10)
	if !ok {
		return fmt.Errorf("invalid BigDecimal coefficient")
	}
	expected := bigDecimalValue(coeff, want.Exp)
	if len(observations) == 0 {
		return fmt.Errorf("missing BigDecimal comparison paths")
	}
	for _, observation := range observations {
		actual, err := decimalref.Decode(width, observation.Bits)
		if err != nil {
			return err
		}
		if actual.Kind != "finite" {
			return fmt.Errorf("%s: BigDecimal value differs: got %s, want %sE%d", observation.Path, actual.Kind, want.Coefficient, want.Exp)
		}
		coefficient := new(big.Int).Set(actual.Coeff)
		if actual.Negative {
			coefficient.Neg(coefficient)
		}
		if bigDecimalValue(coefficient, actual.Exp).Cmp(expected) != 0 {
			return fmt.Errorf("%s: BigDecimal value differs: got %sE%d, want %sE%d", observation.Path, coefficient, actual.Exp, want.Coefficient, want.Exp)
		}
	}
	return nil
}

func TestBigDecimalOracle(t *testing.T) {
	java := startBigDecimal(t)
	if java == nil {
		t.Skip("requires the BigDecimal comparison target")
	}
	for _, name := range []string{"quantize-midpoint", "quantize-midpoint64", "quantize-midpoint128", "unfused32", "unfused64", "unfused128"} {
		sample, good, _, bad, _, err := finiteCampaignWitness(name)
		if err != nil {
			t.Fatal(err)
		}
		want, err := java.evaluate(sample.Case)
		if err != nil {
			t.Fatal(err)
		}
		if err := compareBigDecimal(sample.Case.Width, want, []finiteObservation{{Path: "good", Bits: good}}); err != nil {
			t.Fatalf("%s pinned result: %v", name, err)
		}
		if compareBigDecimal(sample.Case.Width, want, []finiteObservation{{Path: "bad", Bits: bad}}) == nil {
			t.Fatalf("%s wrong rounding accepted by Java comparison", name)
		}
	}
	for _, width := range []int{32, 64, 128} {
		p, _ := decimalref.ParametersFor(width)
		encode := func(coefficient string, exponent int) string {
			value, ok := new(big.Int).SetString(coefficient, 10)
			if !ok {
				t.Fatal("invalid test coefficient")
			}
			raw, err := decimalref.Encode(width, decimalref.Decimal{Kind: "finite", Negative: value.Sign() < 0, Coeff: new(big.Int).Abs(value), Exp: exponent})
			if err != nil {
				t.Fatal(err)
			}
			return raw
		}
		for _, negative := range []bool{false, true} {
			coefficient := "1" + strings.Repeat("0", p.Precision-2) + "5"
			if negative {
				coefficient = "-" + coefficient
			}
			for _, mode := range finiteModes {
				result, err := java.evaluate(decimalref.Case{Width: width, Op: "quantize", Mode: mode, Operands: []string{encode(coefficient, 0), encode("1", 1)}})
				if err != nil {
					t.Fatal(err)
				}
				upper := mode == "nearest_away" || mode == "toward_positive" && !negative || mode == "toward_negative" && negative
				want := "1" + strings.Repeat("0", p.Precision-2)
				if upper {
					want = want[:len(want)-1] + "1"
				}
				if negative {
					want = "-" + want
				}
				if result.Status != "ok" || result.Coefficient != want || result.Exp != 1 {
					t.Fatalf("Java mode mapping width=%d negative=%t mode=%s got=%+v want=%sE1", width, negative, mode, result, want)
				}
			}
		}
		for _, tc := range []struct {
			op, a  string
			exp    int
			b      string
			target int
			reason string
		}{
			{"div", "1", 0, "0", 0, "division-by-zero"},
			{"mul", "1", p.MinExp, "1", -1, "exponent-range"},
			{"mul", strings.Repeat("9", p.Precision), p.MaxExp, "1", 1, "exponent-range"},
			{"quantize", strings.Repeat("9", p.Precision), 0, "1", -1, "quantize-precision"},
		} {
			result, err := java.evaluate(decimalref.Case{Width: width, Op: tc.op, Mode: "nearest_even", Operands: []string{encode(tc.a, tc.exp), encode(tc.b, tc.target)}})
			if err != nil || result.Status != "excluded" || result.Reason != tc.reason {
				t.Fatalf("Java exclusion width=%d case=%+v got=%+v err=%v", width, tc, result, err)
			}
		}
	}
	for _, line := range []string{
		"", "2\t32\tadd\tnearest_even\t1\t0\t1\t0", "1\t256\tadd\tnearest_even\t1\t0\t1\t0",
		"1\t32\tmissing\tnearest_even\t1\t0\t1\t0", "1\t32\tadd\tmissing\t1\t0\t1\t0",
		"1\t32\tadd\tnearest_even\t1\t0", "1\t32\tadd\tnearest_even\t10000000\t0\t1\t0",
		"1\t32\tadd\tnearest_even\t1\t-102\t1\t0", "1\t32\tadd\tnearest_even\t01\t0\t1\t0",
	} {
		cmd := exec.CommandContext(finiteTestContext(t), os.Getenv("BID754_BIGDECIMAL_JAVA"), "-cp", os.Getenv("BID754_BIGDECIMAL_CLASSES"), "BigDecimalProbe")
		cmd.WaitDelay = 250 * time.Millisecond
		cmd.Stdin = strings.NewReader(line + "\n")
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		err := cmd.Run()
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() == 0 || strings.Contains(stdout.String(), "\tok\t") || strings.Contains(stdout.String(), "\texcluded\t") {
			t.Fatalf("invalid Java request was not a process failure: %q exit=%v stdout=%q stderr=%q", line, err, stdout.String(), stderr.String())
		}
	}
	for _, line := range []string{"", "1\texcluded\tunknown\n", "1\tok\t00\t0\n", "1\tok\t1\t+0\n", "1\tok\t1\t999999\n", "2\tok\t1\t0\n"} {
		if _, err := decodeBigDecimalResult(32, []byte(line)); err == nil {
			t.Fatalf("invalid Java result accepted: %q", line)
		}
	}
}
