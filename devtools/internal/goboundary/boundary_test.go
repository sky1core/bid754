package goboundary

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/build/constraint"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/sky1core/bid754/devtools/internal/testgen"
)

const modulePath = "github.com/sky1core/bid754/bid754-go"

func runGo(dir, cgo string, args ...string) ([]byte, error) {
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=", "GOPROXY=off", "GOSUMDB=off", "CGO_ENABLED="+cgo)
	return cmd.CombinedOutput()
}

func TestExternalConsumer(t *testing.T) {
	root, err := filepath.Abs("../../../bid754-go")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	write := func(name, text string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module boundaryconsumer\n\ngo 1.23\nrequire "+modulePath+" v0.0.0\nreplace "+modulePath+" => "+strconv.Quote(root)+"\n")
	var source strings.Builder
	source.WriteString("package main\nimport (d " + strconv.Quote(modulePath) + "; \"unsafe\"; \"fmt\")\nfunc main() {\n")
	for _, width := range []int{32, 64, 128} {
		fmt.Fprintf(&source, "check%d()\n", width)
	}
	source.WriteString("fmt.Println(\"PASS external value contracts: widths=3\")\n}\n")
	for _, width := range []int{32, 64, 128} {
		typ := fmt.Sprintf("d.Decimal%dBID", width)
		fmt.Fprintf(&source, `func check%d() {
 var z %s
 if unsafe.Sizeof(z) != %d || !z.IsZero() { panic("zero/layout") }
 a, err := d.NewDecimal%d("1"); if err != nil { panic(err) }
 cohort, err := d.NewDecimal%d("1.0"); if err != nil { panic(err) }
 equal, flags := a.QuietEqual(cohort)
 if a == cohort || !equal || flags != 0 { panic("representation/numeric equality") }
 pos, err := d.NewDecimal%d("0"); if err != nil { panic(err) }
 neg, err := d.NewDecimal%d("-0"); if err != nil { panic(err) }
 equal, flags = pos.QuietEqual(neg)
 if pos == neg || !equal || flags != 0 { panic("signed zero equality") }
 nan, err := d.NewDecimal%d("NaN"); if err != nil { panic(err) }
 equal, flags = nan.QuietEqual(nan)
 if nan != nan || equal || flags != 0 { panic("NaN equality") }
 sum, flags := a.AddWithMode(a, d.RoundNearestEven)
 expected, err := d.NewDecimal%d("2"); if err != nil { panic(err) }
 equal, compareFlags := sum.QuietEqual(expected)
 if !equal || flags != 0 || compareFlags != 0 { panic("addition") }
 equal, flags = a.Sub(a).QuietEqual(pos)
 if !equal || flags != 0 { panic("subtraction") }
 equal, flags = a.Mul(a).QuietEqual(a)
 if !equal || flags != 0 { panic("multiplication") }
 equal, flags = a.Div(a).QuietEqual(a)
 if !equal || flags != 0 { panic("division") }
 less, flags := a.QuietLess(sum)
 if !less || flags != 0 { panic("numeric ordering") }
 n, flags := a.ConvertToInt64(d.RoundNearestEven)
 if n != 1 || flags != 0 { panic("numeric conversion") }
 values := map[%s]int{a: 1, cohort: 2}
 if len(values) != 2 { panic("map representation identity") }
 if d.One%dBID() != a { panic("constant accessor") }
`, width, typ, width/8, width, width, width, width, width, width, typ, width)
		if width == 128 {
			source.WriteString(`for bit:=0;bit<128;bit++ { var raw [16]byte; raw[bit/8]=1<<uint(bit%8); v:=d.Decimal128BIDFromBytes(raw); if v.ToBytes()!=raw {panic("raw roundtrip")}; copied:=v.ToBytes(); copied[bit/8]^=255; if v.ToBytes()!=raw {panic("raw alias")}}
`)
		} else {
			fmt.Fprintf(&source, "for bit:=0;bit<%d;bit++ { raw:=uint%d(1)<<uint(bit); v:=d.Decimal%dBIDFromBits(raw); if v.ToUint%d()!=raw {panic(\"raw roundtrip\")}}\n", width, width, width, width)
		}
		source.WriteString("}\n")
	}
	write("main.go", source.String())
	for _, cgo := range []string{"0", "1"} {
		out, err := runGo(dir, cgo, "run", ".")
		if err != nil {
			t.Fatalf("baseline CGO=%s: %v\n%s", cgo, err, out)
		}
		t.Log(strings.TrimSpace(string(out)))
	}
	for _, width := range []int{32, 64, 128} {
		typ := fmt.Sprintf("d.Decimal%dBID", width)
		probes := map[string]string{}
		for _, op := range []string{"+", "-", "*", "/", "%", "<", ">", "<=", ">=", "&", "|", "^", "&^"} {
			probes["binary_"+op] = "_ = a " + op + " b"
		}
		for _, op := range []string{"<<", ">>"} {
			probes["shift_"+op] = "_ = a " + op + " 1"
		}
		for _, op := range []string{"+", "-", "^"} {
			probes["unary_"+op] = "_ = " + op + "a"
		}
		probes["increment"] = "a++"
		probes["decrement"] = "a--"
		probes["field_read"] = "_ = a.raw"
		probes["field_write"] = "a.raw = b.raw"
		probes["field_literal"] = "_ = " + typ + "{raw: 0}"
		probes["unkeyed_literal"] = "_ = " + typ + "{0}"
		if width == 128 {
			probes["index"] = "_ = a[0]"
			probes["raw_cast"] = "_ = [16]byte(a)"
			probes["decimal_cast"] = "_ = d.Decimal128BID([16]byte{})"
		} else {
			probes["raw_cast"] = "_ = uint" + strconv.Itoa(width) + "(a)"
			probes["decimal_cast"] = "_ = " + typ + "(uint" + strconv.Itoa(width) + "(0))"
		}
		names := make([]string, 0, len(probes))
		for name := range probes {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			body := probes[name]
			t.Run(fmt.Sprintf("%d/%s", width, name), func(t *testing.T) {
				write("main.go", "package main\nimport d "+strconv.Quote(modulePath)+"\nfunc main(){var a,b "+typ+"; _=a; _=b; "+body+"}\n")
				out, err := runGo(dir, "0", "build", "-o", filepath.Join(dir, "probe"), ".")
				if err == nil {
					t.Fatalf("forbidden operation compiled: %s", body)
				}
				diagnostic := string(out)
				want := "not defined"
				switch {
				case strings.HasPrefix(name, "field_") || name == "unkeyed_literal":
					want = "unexported"
				case strings.HasPrefix(name, "shift_"):
					want = "must be integer"
				case name == "raw_cast" || name == "decimal_cast":
					want = "cannot convert"
				case name == "index":
					want = "cannot index"
				case name == "increment" || name == "decrement":
					want = "non-numeric type"
				}
				if !strings.Contains(diagnostic, want) || !strings.Contains(diagnostic, "main.go:") {
					t.Fatalf("wrong rejection, want %q: %v\n%s", want, err, out)
				}
			})
		}
	}
}

type listedPackage struct {
	Dir            string
	ImportPath     string
	Standard       bool
	GoFiles        []string
	CgoFiles       []string
	TestGoFiles    []string
	XTestGoFiles   []string
	IgnoredGoFiles []string
}

func decodePackages(data []byte) ([]listedPackage, error) {
	var packages []listedPackage
	dec := json.NewDecoder(bytes.NewReader(data))
	for {
		var p listedPackage
		err := dec.Decode(&p)
		if err == io.EOF {
			return packages, nil
		}
		if err != nil {
			return nil, err
		}
		packages = append(packages, p)
	}
}

func TestDefaultBuildExcludesGeneratedVerification(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := testgen.LoadManifest(filepath.Join(root, "testgen_manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	spec, err := testgen.LoadGenerated(filepath.Join(root, manifest.Output))
	if err != nil {
		t.Fatal(err)
	}
	outputs, err := testgen.GenerateGoTestOutputs(root, manifest, spec)
	if err != nil {
		t.Fatal(err)
	}
	generated := map[string]bool{}
	for path, data := range outputs {
		if !strings.HasSuffix(path, ".go") {
			continue
		}
		full := filepath.Clean(filepath.Join(root, path))
		generated[full] = true
		actual, err := os.ReadFile(full)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(actual, data) {
			t.Errorf("generated output is stale: %s", path)
		}
		file, err := parser.ParseFile(token.NewFileSet(), full, data, parser.ImportsOnly|parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		hasC := false
		for _, imp := range file.Imports {
			if imp.Path.Value == `"C"` {
				hasC = true
			}
		}
		if hasC && strings.HasSuffix(path, "_test.go") {
			t.Errorf("cgo import in test file: %s", path)
		}
		if strings.HasSuffix(path, "_test.go") || strings.Contains(path, "/internal/testspec/") {
			continue
		}
		var tag constraint.Expr
		for _, group := range file.Comments {
			for _, comment := range group.List {
				if constraint.IsGoBuild(comment.Text) {
					tag, err = constraint.Parse(comment.Text)
					if err != nil {
						t.Fatal(err)
					}
				}
			}
		}
		if tag == nil {
			t.Errorf("verification support has no build boundary: %s", path)
			continue
		}
		for _, cgo := range []bool{false, true} {
			if tag.Eval(func(name string) bool { return name == "cgo" && cgo }) {
				t.Errorf("verification support enabled by default cgo=%t: %s", cgo, path)
			}
		}
	}
	module := filepath.Join(root, "../bid754-go")
	for _, cgo := range []string{"0", "1"} {
		out, err := runGo(module, cgo, "list", "-deps", "-json", ".")
		if err != nil {
			t.Fatalf("production graph: %v\n%s", err, out)
		}
		packages, err := decodePackages(out)
		if err != nil {
			t.Fatal(err)
		}
		if err := checkDefaultProductSourceOwnership(packages); err != nil {
			t.Errorf("CGO=%s: %v", cgo, err)
		}
		count := 0
		for _, p := range packages {
			if p.ImportPath == "testing" || strings.HasPrefix(p.ImportPath, "testing/") {
				t.Errorf("test dependency in production graph: %s", p.ImportPath)
			}
			for _, name := range append(append([]string{}, p.GoFiles...), p.CgoFiles...) {
				full := filepath.Join(p.Dir, name)
				if generated[full] {
					t.Errorf("generated verification in production graph: %s", full)
				}
				if p.ImportPath == modulePath {
					count++
				}
			}
		}
		out, err = runGo(module, cgo, "list", "-deps", "-test", "-json", ".")
		if err != nil {
			t.Fatalf("test graph: %v\n%s", err, out)
		}
		packages, err = decodePackages(out)
		if err != nil {
			t.Fatal(err)
		}
		testGraph := map[string]bool{}
		for _, p := range packages {
			for _, names := range [][]string{p.GoFiles, p.CgoFiles, p.TestGoFiles, p.XTestGoFiles, p.IgnoredGoFiles} {
				for _, name := range names {
					testGraph[filepath.Join(p.Dir, name)] = true
				}
			}
		}
		for path := range generated {
			if !testGraph[path] {
				t.Errorf("generated verification output absent from test graph: %s", path)
			}
		}
		t.Logf("CGO=%s production root files=%d generated verification files=%d graph separation checked", cgo, count, len(generated))
	}
}
