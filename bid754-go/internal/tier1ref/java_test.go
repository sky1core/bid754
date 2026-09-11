package tier1ref

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func javaCommand(t *testing.T, input string) ([]string, error) {
	t.Helper()
	java, classes := os.Getenv("BID754_BIGDECIMAL_JAVA"), os.Getenv("BID754_BIGDECIMAL_CLASSES")
	if java == "" && classes == "" {
		t.Skip("Java probe requires BID754_BIGDECIMAL_JAVA and BID754_BIGDECIMAL_CLASSES")
	}
	if java == "" || classes == "" {
		t.Fatal("both Java probe environment values are required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, java, "-cp", classes, "Tier1BigDecimalProbe")
	cmd.Stdin = strings.NewReader(input)
	output, err := cmd.CombinedOutput()
	lines := strings.Split(strings.TrimSuffix(string(output), "\n"), "\n")
	if len(lines) == 0 {
		t.Fatal("missing Java handshake")
	}
	fields := strings.Split(lines[0], "\t")
	if len(fields) != 5 || fields[0] != "1" || fields[1] != "ready" {
		t.Fatalf("invalid Java handshake: %q", lines[0])
	}
	for _, encoded := range fields[2:4] {
		if data, err := base64.StdEncoding.DecodeString(encoded); err != nil || len(data) == 0 {
			t.Fatal("invalid Java runtime/vendor")
		}
	}
	class, readErr := os.ReadFile(filepath.Join(classes, "Tier1BigDecimalProbe.class"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if fields[4] != fmt.Sprintf("%x", sha256.Sum256(class)) {
		t.Fatal("Java class identity differs")
	}
	if ctx.Err() != nil {
		t.Fatalf("Java probe exceeded execution budget: %v", ctx.Err())
	}
	return lines[1:], err
}

func TestJavaProbeWitnesses(t *testing.T) {
	type witness struct{ request, response string }
	var rows []witness
	add := func(request, response string) { rows = append(rows, witness{request, response}) }
	for _, width := range []int{32, 64, 128} {
		for _, mode := range modes {
			prefix := fmt.Sprintf("1\t%d\t", width)
			add(prefix+"rem\t"+mode+"\t\t0\t7\t0\t2\t0", "1\tok\tdecimal\t-1\t0")
			add(prefix+"fmod\t"+mode+"\t\t0\t7\t0\t2\t0", "1\tok\tdecimal\t1\t0")
			add(prefix+"rem\t"+mode+"\t\t0\t5\t0\t2\t0", "1\tok\tdecimal\t1\t0")
			add(prefix+"rem\t"+mode+"\t\t0\t-7\t0\t2\t0", "1\tok\tdecimal\t1\t0")
			add(prefix+"fmod\t"+mode+"\t\t0\t-7\t0\t2\t0", "1\tok\tdecimal\t-1\t0")
			add(prefix+"rem\t"+mode+"\t\t0\t1\t0\t0\t0", "1\texcluded\tdivision-by-zero")
			add(prefix+"fmod\t"+mode+"\t\t0\t0\t0\t0\t0", "1\texcluded\tdivision-by-zero")
			add(prefix+"minnum\t"+mode+"\t\t0\t-2\t0\t1\t0", "1\tok\tdecimal\t-2\t0")
			add(prefix+"maxnum\t"+mode+"\t\t0\t-2\t0\t1\t0", "1\tok\tdecimal\t1\t0")
			add(prefix+"scaleb\t"+mode+"\t2\t0\t123\t-2", "1\tok\tdecimal\t123\t0")
			for _, shift := range []string{"9223372036854775807", "-9223372036854775808"} {
				add(prefix+"scaleb\t"+mode+"\t"+shift+"\t0\t1\t0", "1\texcluded\texponent-range")
			}
			for _, bits := range []int{8, 16, 32, 64} {
				for _, op := range []string{"to_int", "to_uint", "to_int_exact", "to_uint_exact"} {
					value := "2"
					if mode == "nearest_away" || mode == "toward_positive" {
						value = "3"
					}
					add(prefix+op+"\t"+mode+"\t\t"+strconv.Itoa(bits)+"\t25\t-1", "1\tok\tinteger\t"+value)
				}
			}
			for _, target := range []int{32, 64, 128} {
				if width == target {
					continue
				}
				add(prefix+"convert\t"+mode+"\t\t"+strconv.Itoa(target)+"\t123\t-2", "1\tok\tdecimal\t123\t-2")
			}
			for _, row := range []struct{ op, want string }{{"quiet_equal", "true"}, {"quiet_not_equal", "false"}, {"quiet_greater", "false"}, {"quiet_greater_equal", "true"}, {"quiet_greater_unordered", "false"}, {"quiet_less", "false"}, {"quiet_less_equal", "true"}, {"quiet_less_unordered", "false"}, {"quiet_not_greater", "true"}, {"quiet_not_less", "true"}, {"quiet_ordered", "true"}, {"quiet_unordered", "false"}} {
				add(prefix+row.op+"\t"+mode+"\t\t0\t1\t0\t100\t-2", "1\tok\tbool\t"+row.want)
			}
			if width == 32 {
				value := "1234566"
				if mode == "nearest_away" || mode == "toward_positive" {
					value = "1234567"
				}
				add(prefix+"from_int\t"+mode+"\t12345665\t32", "1\tok\tdecimal\t"+value+"\t1")
			} else if width == 64 {
				value := "1234567890123456"
				if mode == "nearest_away" || mode == "toward_positive" {
					value = "1234567890123457"
				}
				add(prefix+"from_int\t"+mode+"\t12345678901234565\t64", "1\tok\tdecimal\t"+value+"\t1")
			} else {
				add(prefix+"from_uint\t"+mode+"\t18446744073709551615\t64", "1\tok\tdecimal\t18446744073709551615\t0")
			}
		}
	}
	for _, row := range []witness{
		{"1\t128\tto_uint_exact\tnearest_even\t\t64\t18446744073709551615\t0", "1\tok\tinteger\t18446744073709551615"},
		{"1\t128\tto_uint\tnearest_even\t\t64\t18446744073709551616\t0", "1\texcluded\tinvalid-integer"},
		{"1\t128\tto_int\tnearest_even\t\t64\t-9223372036854775808\t0", "1\tok\tinteger\t-9223372036854775808"},
		{"1\t128\tto_int\tnearest_even\t\t64\t9223372036854775808\t0", "1\texcluded\tinvalid-integer"},
		{"1\t128\tto_uint\tnearest_even\t\t8\t-5\t-1", "1\tok\tinteger\t0"},
		{"1\t128\tto_uint\tnearest_away\t\t8\t-5\t-1", "1\texcluded\tinvalid-integer"},
		{"1\t32\tminnum\tnearest_even\t\t0\t1\t-101\t2\t0", "1\texcluded\texponent-range"},
		{"1\t64\tconvert\tnearest_even\t\t32\t99999995\t-103", "1\tok\tdecimal\t1000000\t-101"},
		{"1\t64\tconvert\ttoward_zero\t\t32\t99999995\t-103", "1\texcluded\texponent-range"},
		{"1\t32\tscaleb\tnearest_even\t9223372036854775807\t0\t0\t0", "1\tok\tdecimal\t0\t90"},
		{"1\t128\tscaleb\tnearest_even\t-9223372036854775808\t0\t0\t0", "1\tok\tdecimal\t0\t-6176"},
		{"1\t128\tto_uint\ttoward_zero\t\t64\t1\t6111", "1\texcluded\tinvalid-integer"},
		{"1\t128\trem\tnearest_even\t\t0\t1\t6111\t3\t0", "1\tok\tdecimal\t1\t0"},
	} {
		add(row.request, row.response)
	}
	requests := make([]string, len(rows))
	for i, row := range rows {
		requests[i] = row.request
	}
	lines, err := javaCommand(t, strings.Join(requests, "\n")+"\n")
	if err != nil {
		t.Fatalf("Java failed: %v\n%s", err, strings.Join(lines, "\n"))
	}
	if len(lines) != len(rows) {
		t.Fatalf("Java produced %d rows, want %d", len(lines), len(rows))
	}
	for i, line := range lines {
		if line == rows[i].response {
			continue
		}
		got, want := strings.Split(line, "\t"), strings.Split(rows[i].response, "\t")
		if len(got) == 5 && len(want) == 5 && got[0] == "1" && got[1] == "ok" && got[2] == "decimal" && want[2] == "decimal" {
			parse := func(fields []string) *big.Rat {
				n, ok := new(big.Int).SetString(fields[3], 10)
				if !ok || len(fields[3]) > 36 {
					t.Fatalf("invalid coefficient %q", fields[3])
				}
				e, err := strconv.Atoi(fields[4])
				if err != nil || e < -20000 || e > 20000 {
					t.Fatal("invalid Java exponent")
				}
				power := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(max(e, -e))), nil)
				if e < 0 {
					return new(big.Rat).SetFrac(n, power)
				}
				return new(big.Rat).SetInt(n.Mul(n, power))
			}
			if parse(got).Cmp(parse(want)) == 0 {
				continue
			}
		}
		t.Errorf("request %s: got %s, want %s", rows[i].request, line, rows[i].response)
	}
	t.Logf("Java independent witnesses: %d", len(rows))
}

func TestJavaProbeRejectsMalformedRequests(t *testing.T) {
	valid := "1\t32\trem\tnearest_even\t\t0\t7\t0\t2\t0"
	for _, input := range []string{
		"", strings.Replace(valid, "1\t32", "2\t32", 1), strings.Replace(valid, "32", "16", 1),
		strings.Replace(valid, "nearest_even", "bad", 1), strings.Replace(valid, "rem", "unknown", 1),
		strings.Replace(valid, "\t7\t", "\t07\t", 1), strings.Replace(valid, "\t7\t", "\t-0\t", 1),
		strings.Replace(valid, "\t7\t", "\t10000000\t", 1), strings.Replace(valid, "\t7\t0", "\t7\t-102", 1),
		valid + "\t0", strings.Repeat("9", 2000),
		"1\t32\tconvert\tnearest_even\t\t32\t1\t0",
		"1\t32\tto_int\tnearest_even\t\t128\t1\t0",
		"1\t32\tscaleb\tnearest_even\t9223372036854775808\t0\t1\t0",
		"1\t32\tscaleb\tnearest_even\t-0\t0\t1\t0",
		"1\t32\tscaleb\tnearest_even\t1\t32\t1\t0",
		"1\t32\tfrom_int\tnearest_even\t2147483648\t32",
		"1\t32\tfrom_uint\tnearest_even\t18446744073709551616\t64",
		"1\t32\tfrom_uint\tnearest_even\t-1\t64",
		"1\t32\tfrom_int\tnearest_even\t1\t8",
		"1\t32\tminnum\tnearest_even\t1\t0\t1\t0\t2\t0",
	} {
		lines, err := javaCommand(t, input+"\n")
		if err == nil {
			t.Fatalf("accepted malformed input length=%d", len(input))
		}
		for _, line := range lines {
			if strings.HasPrefix(line, "1\tok\t") || strings.HasPrefix(line, "1\texcluded\t") {
				t.Fatalf("malformed input treated as observation: %s", line)
			}
		}
	}
}
