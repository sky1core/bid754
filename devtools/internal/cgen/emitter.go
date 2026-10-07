package cgen

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

type Generated struct {
	Go              []byte
	Rust            []byte
	Bidgo           []byte
	Runtime         []byte
	RuntimeRound128 []byte
}

func Generate(repoRoot string, manifest Manifest) (Generated, error) {
	tables, err := loadTables(repoRoot, manifest.Tables)
	if err != nil {
		return Generated{}, err
	}

	goData, err := renderGo(manifest.GoPackage, tables)
	if err != nil {
		return Generated{}, err
	}
	rustData, err := renderRust(tables)
	if err != nil {
		return Generated{}, err
	}
	var bidgoTables []Table
	for _, table := range tables {
		if table.Spec.Source == manifest.BidgoSource {
			if len(table.Spec.Runtime) != 1 || table.Spec.Runtime[0].Name != table.Spec.Name || table.Spec.Runtime[0].Scalar != bidgoScalarFor(table.CType) || table.Spec.Runtime[0].Prefix != 0 || table.Spec.Runtime[0].ExtraNearestRow {
				return Generated{}, fmt.Errorf("binary runtime target for %s does not match its generated declaration", table.Spec.Name)
			}
			bidgoTables = append(bidgoTables, table)
		}
	}
	if len(bidgoTables) == 0 {
		return Generated{}, fmt.Errorf("bidgo_source %q matched no manifest tables", manifest.BidgoSource)
	}
	bidgoData, err := renderBidgo(bidgoTables)
	if err != nil {
		return Generated{}, err
	}

	runtimeData, err := renderRuntimeTables(tables, manifest.BidgoSource, false)
	if err != nil {
		return Generated{}, err
	}
	round128Data, err := renderRuntimeTables(tables, manifest.BidgoSource, true)
	if err != nil {
		return Generated{}, err
	}
	return Generated{Go: goData, Rust: rustData, Bidgo: bidgoData, Runtime: runtimeData, RuntimeRound128: round128Data}, nil
}

func WriteOutputs(repoRoot string, manifest Manifest, generated Generated) error {
	if err := writeFile(filepath.Join(repoRoot, manifest.GoOutput), generated.Go); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(repoRoot, manifest.RustOutput), generated.Rust); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(repoRoot, manifest.BidgoOutput), generated.Bidgo); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(repoRoot, manifest.RuntimeOutput), generated.Runtime); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(repoRoot, manifest.RuntimeRound128Output), generated.RuntimeRound128); err != nil {
		return err
	}
	return nil
}

func renderRuntimeTables(tables []Table, binarySource string, round128Only bool) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(genmarker.Line("c-tablegen") + "\n")
	b.WriteString("// Source tables come from pinned Intel DFP C files.\n\npackage bidgo\n\n")
	for _, table := range tables {
		if table.Spec.Source == binarySource {
			continue
		}
		for _, target := range table.Spec.Runtime {
			if (target.Name == "bid_round_const_table_128") != round128Only {
				continue
			}
			value := table.Value
			dims := append([]int(nil), table.Dims...)
			if target.Prefix > 0 {
				if len(dims) != 1 || target.Prefix > dims[0] {
					return nil, fmt.Errorf("invalid prefix %d for %s", target.Prefix, target.Name)
				}
				value.Elements = value.Elements[:target.Prefix]
				dims[0] = target.Prefix
			}
			if target.ExtraNearestRow {
				if len(dims) != 2 || dims[0] != 5 {
					return nil, fmt.Errorf("invalid extra rounding row for %s", target.Name)
				}
				value.Elements = append(append([]Value(nil), value.Elements...), value.Elements[0])
				dims[0]++
			}
			if err := checkRuntimeScalar(table, target); err != nil {
				return nil, err
			}
			runtimeTable := table
			runtimeTable.Dims = dims
			runtimeTable.Value = value
			b.WriteString(fmt.Sprintf("// %s comes from %s:%s.\n", target.Name, table.SourceRel, table.Spec.Name))
			base := target.Scalar
			for i := len(dims) - 1; i >= 0; i-- {
				base = fmt.Sprintf("[%d]%s", dims[i], base)
			}
			b.WriteString(fmt.Sprintf("var %s = %s%s\n\n", target.Name, base, renderBidgoValue(runtimeTable, value, 0, 0)))
		}
	}
	formatted, err := format.Source(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format runtime tables: %w", err)
	}
	return formatted, nil
}

func checkRuntimeScalar(table Table, target RuntimeSpec) error {
	if target.Scalar == "DEC_DIGITS" {
		if table.CType != "DEC_DIGITS" {
			return fmt.Errorf("%s cannot use DEC_DIGITS for %s", target.Name, table.CType)
		}
		return nil
	}
	if fixedWordArity(table.CType) > 0 {
		if target.Scalar != table.CType {
			return fmt.Errorf("%s cannot use %s for %s", target.Name, target.Scalar, table.CType)
		}
		return nil
	}
	if target.Scalar != "int" && target.Scalar != "int8" && target.Scalar != "uint8" && target.Scalar != "byte" && target.Scalar != "uint32" && target.Scalar != "uint64" {
		return fmt.Errorf("unsupported runtime scalar %q for %s", target.Scalar, target.Name)
	}
	return nil
}

func loadTables(repoRoot string, specs []TableSpec) ([]Table, error) {
	tables := make([]Table, 0, len(specs))
	for _, spec := range specs {
		table, err := ParseTableFile(repoRoot, spec)
		if err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	sort.Slice(tables, func(i, j int) bool {
		return tables[i].Spec.GoName < tables[j].Spec.GoName
	})
	return tables, nil
}

func renderGo(pkg string, tables []Table) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(genmarker.Line("c-tablegen") + "\n")
	b.WriteString("// Source tables come from Intel DFP C files under devtools/third_party/intel_dfp/src.\n\n")
	b.WriteString("package " + pkg + "\n\n")
	b.WriteString("type UInt128Words = [2]uint64\n\n")
	b.WriteString("type UInt192Words = [3]uint64\n\n")
	b.WriteString("type UInt256Words = [4]uint64\n\n")
	if hasCType(tables, "DEC_DIGITS") {
		b.WriteString("type DecDigitsWords struct {\n")
		b.WriteString("\tDigits      uint32\n")
		b.WriteString("\tThresholdHi uint64\n")
		b.WriteString("\tThresholdLo uint64\n")
		b.WriteString("\tDigits1     uint32\n")
		b.WriteString("}\n\n")
	}

	for _, table := range tables {
		b.WriteString(fmt.Sprintf("// %s comes from %s:%s.\n", table.Spec.GoName, table.SourceRel, table.Spec.Name))
		b.WriteString(fmt.Sprintf("var %s = %s%s\n\n", table.Spec.GoName, goTypeFor(table), renderGoValue(table, table.Value, 0, 0)))
	}

	formatted, err := format.Source(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated go: %w", err)
	}
	return formatted, nil
}

func renderRust(tables []Table) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(genmarker.Line("c-tablegen") + "\n")
	b.WriteString("// Source tables come from Intel DFP C files under devtools/third_party/intel_dfp/src.\n\n")
	b.WriteString("pub type UInt128Words = [u64; 2];\n\n")
	b.WriteString("pub type UInt192Words = [u64; 3];\n\n")
	b.WriteString("pub type UInt256Words = [u64; 4];\n\n")
	if hasCType(tables, "DEC_DIGITS") {
		b.WriteString("#[derive(Clone, Copy, Debug, PartialEq, Eq)]\n")
		b.WriteString("pub struct DecDigitsWords {\n")
		b.WriteString("    pub digits: u32,\n")
		b.WriteString("    pub threshold_hi: u64,\n")
		b.WriteString("    pub threshold_lo: u64,\n")
		b.WriteString("    pub digits1: u32,\n")
		b.WriteString("}\n\n")
	}

	for _, table := range tables {
		b.WriteString(fmt.Sprintf("// %s comes from %s:%s.\n", table.Spec.RustName, table.SourceRel, table.Spec.Name))
		b.WriteString(fmt.Sprintf("pub const %s: %s = %s;\n\n", table.Spec.RustName, rustTypeFor(table), renderRustValue(table, table.Value, 0, 0)))
	}

	return b.Bytes(), nil
}

func renderBidgo(tables []Table) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(genmarker.Line("c-tablegen") + "\n")
	b.WriteString("// Source tables come from pinned Intel DFP C files.\n\n")
	b.WriteString("package bidgo\n\n")

	for _, table := range tables {
		b.WriteString(fmt.Sprintf("// %s comes from %s:%s.\n", table.Spec.Name, table.SourceRel, table.Spec.Name))
		b.WriteString(fmt.Sprintf("var %s = %s%s\n\n", table.Spec.Name, bidgoTypeFor(table), renderBidgoValue(table, table.Value, 0, 0)))
	}

	formatted, err := format.Source(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated bidgo tables: %w", err)
	}
	return formatted, nil
}

func bidgoTypeFor(table Table) string {
	base := bidgoScalarFor(table.CType)
	for i := len(table.Dims) - 1; i >= 0; i-- {
		base = fmt.Sprintf("[%d]%s", table.Dims[i], base)
	}
	return base
}

func bidgoScalarFor(cType string) string {
	switch cType {
	case "BID_UINT8":
		return "uint8"
	case "BID_UINT32":
		return "uint32"
	case "BID_UINT64":
		return "uint64"
	case "BID_UINT128", "BID_UINT192", "BID_UINT256":
		return cType
	case "BID_SINT8":
		return "int8"
	case "char":
		return "uint8"
	case "unsigned int":
		return "uint32"
	case "int":
		return "int"
	default:
		panic("unsupported C type for bidgo: " + cType)
	}
}

func renderBidgoValue(table Table, v Value, indent int, depth int) string {
	if depth == len(table.Dims) {
		switch table.CType {
		case "DEC_DIGITS":
			if len(v.Elements) != 4 || !allScalar(v.Elements) {
				panic("DEC_DIGITS value does not have four scalar fields")
			}
			return fmt.Sprintf("{digits: %s, threshold_hi: %s, threshold_lo: %s, digits1: %s}", v.Elements[0].Number, v.Elements[1].Number, v.Elements[2].Number, v.Elements[3].Number)
		case "BID_UINT128":
			if len(v.Elements) != 2 || !allScalar(v.Elements) {
				panic("BID_UINT128 value does not have two scalar limbs")
			}
			return fmt.Sprintf("{lo: %s, hi: %s}", v.Elements[0].Number.String(), v.Elements[1].Number.String())
		case "BID_UINT192", "BID_UINT256":
			arity := fixedWordArity(table.CType)
			if len(v.Elements) != arity || !allScalar(v.Elements) {
				panic(fmt.Sprintf("%s value does not have %d scalar limbs", table.CType, arity))
			}
			words := make([]string, len(v.Elements))
			for i, elem := range v.Elements {
				words[i] = fmt.Sprintf("w%d: %s", i, elem.Number.String())
			}
			return fmt.Sprintf("{%s}", strings.Join(words, ", "))
		}
	}
	if v.IsScalar() {
		return v.Number.String()
	}
	if len(v.Elements) == 0 {
		return "{}"
	}
	prefix := strings.Repeat("\t", indent)
	childPrefix := strings.Repeat("\t", indent+1)
	parts := make([]string, 0, len(v.Elements))
	for _, elem := range v.Elements {
		parts = append(parts, childPrefix+renderBidgoValue(table, elem, indent+1, depth+1)+",")
	}
	return "{\n" + strings.Join(parts, "\n") + "\n" + prefix + "}"
}

func goTypeFor(table Table) string {
	base := goScalarFor(table.CType)
	for i := len(table.Dims) - 1; i >= 0; i-- {
		base = fmt.Sprintf("[%d]%s", table.Dims[i], base)
	}
	return base
}

func rustTypeFor(table Table) string {
	base := rustScalarFor(table.CType)
	for i := len(table.Dims) - 1; i >= 0; i-- {
		base = fmt.Sprintf("[%s; %d]", base, table.Dims[i])
	}
	return base
}

func goScalarFor(cType string) string {
	switch cType {
	case "BID_UINT8":
		return "uint8"
	case "BID_UINT32":
		return "uint32"
	case "BID_UINT64":
		return "uint64"
	case "BID_UINT128":
		return "UInt128Words"
	case "BID_UINT192":
		return "UInt192Words"
	case "BID_UINT256":
		return "UInt256Words"
	case "BID_SINT8":
		return "int8"
	case "char":
		return "uint8"
	case "unsigned int":
		return "uint32"
	case "int":
		return "int32"
	case "DEC_DIGITS":
		return "DecDigitsWords"
	default:
		panic("unsupported C type: " + cType)
	}
}

func rustScalarFor(cType string) string {
	switch cType {
	case "BID_UINT8":
		return "u8"
	case "BID_UINT32":
		return "u32"
	case "BID_UINT64":
		return "u64"
	case "BID_UINT128":
		return "UInt128Words"
	case "BID_UINT192":
		return "UInt192Words"
	case "BID_UINT256":
		return "UInt256Words"
	case "BID_SINT8":
		return "i8"
	case "char":
		return "u8"
	case "unsigned int":
		return "u32"
	case "int":
		return "i32"
	case "DEC_DIGITS":
		return "DecDigitsWords"
	default:
		panic("unsupported C type: " + cType)
	}
}

func renderGoValue(table Table, v Value, indent int, depth int) string {
	if table.CType == "DEC_DIGITS" && depth == len(table.Dims) {
		if len(v.Elements) != 4 || !v.Elements[0].IsScalar() || !v.Elements[1].IsScalar() || !v.Elements[2].IsScalar() || !v.Elements[3].IsScalar() {
			panic("DEC_DIGITS value does not have four scalar fields")
		}
		return fmt.Sprintf("{Digits: %s, ThresholdHi: %s, ThresholdLo: %s, Digits1: %s}",
			v.Elements[0].Number.String(),
			v.Elements[1].Number.String(),
			v.Elements[2].Number.String(),
			v.Elements[3].Number.String(),
		)
	}
	if v.IsScalar() {
		return v.Number.String()
	}
	if len(v.Elements) == 0 {
		return "{}"
	}
	prefix := strings.Repeat("\t", indent)
	childPrefix := strings.Repeat("\t", indent+1)
	var parts []string
	for _, elem := range v.Elements {
		parts = append(parts, childPrefix+renderGoValue(table, elem, indent+1, depth+1)+",")
	}
	return "{\n" + strings.Join(parts, "\n") + "\n" + prefix + "}"
}

func renderRustValue(table Table, v Value, indent int, depth int) string {
	if table.CType == "DEC_DIGITS" && depth == len(table.Dims) {
		if len(v.Elements) != 4 || !v.Elements[0].IsScalar() || !v.Elements[1].IsScalar() || !v.Elements[2].IsScalar() || !v.Elements[3].IsScalar() {
			panic("DEC_DIGITS value does not have four scalar fields")
		}
		return fmt.Sprintf("DecDigitsWords { digits: %su32, threshold_hi: %su64, threshold_lo: %su64, digits1: %su32 }",
			v.Elements[0].Number.String(),
			v.Elements[1].Number.String(),
			v.Elements[2].Number.String(),
			v.Elements[3].Number.String(),
		)
	}
	scalarType := rustScalarFor(table.CType)
	if v.IsScalar() {
		if isRustWordAlias(scalarType) {
			return v.Number.String()
		}
		return v.Number.String() + scalarType
	}
	if len(v.Elements) == 0 {
		return "[]"
	}
	prefix := strings.Repeat("    ", indent)
	childPrefix := strings.Repeat("    ", indent+1)
	var parts []string
	for _, elem := range v.Elements {
		parts = append(parts, childPrefix+renderRustValue(table, elem, indent+1, depth+1)+",")
	}
	return "[\n" + strings.Join(parts, "\n") + "\n" + prefix + "]"
}

func isRustWordAlias(name string) bool {
	switch name {
	case "UInt128Words", "UInt192Words", "UInt256Words":
		return true
	default:
		return false
	}
}

func hasCType(tables []Table, cType string) bool {
	for _, table := range tables {
		if table.CType == cType {
			return true
		}
	}
	return false
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output dir for %q: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}
	return nil
}
