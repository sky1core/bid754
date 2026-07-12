// Package tablecrosscheck cross-checks the c-tablegen output package
// (devtools/generated/go, generated from pinned Intel BID C) against the
// hand-ported table literals inside bid754-go/internal/bidgo.
//
// bid754-go is zero-dependency and bidgo is an internal package, so the bidgo
// values cannot be imported here. Instead this test type-checks the bidgo
// package from source with go/types and evaluates the table initializer
// literals as exact constants. The orphaned c-tablegen Go output thereby
// becomes the verification anchor for the hand-ported tables (including
// tables_binarydecimal.go, which carries a generated-code marker but has no
// in-repo generator).
//
// The check is exhaustive in both directions:
//   - every tablegen_manifest.json table must be anchored by at least one
//     mapping entry, and
//   - every bidgo package-level var initialized with a composite literal must
//     have a mapping entry (a value comparison or a documented exclusion).
//     No size threshold is applied: the smallest real Intel table in the
//     package has 7 scalar leaves, so any threshold would create a coverage gap
//     for hand-ported copies of small Intel tables.
package tablecrosscheck

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	generatedtables "github.com/sky1core/bid754/devtools/generated/go"
	"github.com/sky1core/bid754/devtools/internal/cgen"
)

const (
	bidgoDirRel         = "../../../bid754-go/internal/bidgo"
	tablegenManifestRel = "../../tablegen_manifest.json"
)

// tableMapping links one bidgo package-level table var (the map key in
// bidgoTableMappings) to the c-tablegen output value that anchors it.
type tableMapping struct {
	// goName is the tablegen_manifest.json go_name whose generated value
	// anchors this bidgo var. Several bidgo vars may share one go_name when
	// bidgo carries multiple hand-ported copies of the same Intel table.
	// The manifest completeness check trusts this string as-is: a mapping
	// entry whose goName names table X while generated holds table Y's value
	// is only caught when the value comparison fails (leaf count or values
	// differ); it cannot be caught structurally because Go offers no way to
	// resolve a package-level var from its name at run time.
	goName string
	// generated is the c-tablegen output value; nil only for excluded entries.
	generated any
	// excludedReason documents bidgo vars whose value cannot be statically
	// compared, and why. Empty for compared tables.
	excludedReason string
	// pinnedGeneratedLeaves / pinnedBidgoLeaves must both be set when the two
	// sides intentionally hold a different number of scalar leaves. The common
	// prefix is compared and both totals are asserted exactly. When zero, the
	// two sides must have identical leaf counts.
	pinnedGeneratedLeaves int
	pinnedBidgoLeaves     int
	// pinnedGeneratedRowLen / pinnedBidgoRowLen must both be set (together
	// with the pinned leaf totals) when the two sides slice the same Intel
	// rows at different row widths, so a flat prefix comparison would
	// misalign. Rows are then compared pairwise over the common column
	// prefix.
	pinnedGeneratedRowLen int
	pinnedBidgoRowLen     int
}

// bidgoTableMappings maps every bidgo package-level composite-literal table
// var to its c-tablegen anchor. The test fails if this map ever diverges from
// either tablegen_manifest.json or the set of composite-literal vars in the
// bidgo package.
func bidgoTableMappings() map[string]tableMapping {
	return map[string]tableMapping{
		"bid_bid_bid_recip_scale32":   {goName: "BidBidBidRecipScale32", generated: generatedtables.BidBidBidRecipScale32},
		"bid_bid_reciprocals10_32":    {goName: "BidBidReciprocals10_32", generated: generatedtables.BidBidReciprocals10_32},
		"bid_convert_table":           {goName: "BidConvertTable", generated: generatedtables.BidConvertTable},
		"bid_estimate_bin_expon":      {goName: "BidEstimateBinExpon", generated: generatedtables.BidEstimateBinExpon},
		"bid_estimate_decimal_digits": {goName: "BidEstimateDecimalDigits", generated: generatedtables.BidEstimateDecimalDigits},
		"bid_Ex128m128":               {goName: "BidEx128M128", generated: generatedtables.BidEx128M128},
		"bid_Ex192m192":               {goName: "BidEx192M192", generated: generatedtables.BidEx192M192},
		"bid_Ex256m256":               {goName: "BidEx256M256", generated: generatedtables.BidEx256M256},
		"bid_Ex64m64":                 {goName: "BidEx64M64", generated: generatedtables.BidEx64M64},
		// Intel C declares 1024 rows; bidgo declares [1025] with the same 1024
		// literal rows, leaving one zero-valued check row at index 1024.
		"bid_factors": {goName: "BidFactors", generated: generatedtables.BidFactors, pinnedGeneratedLeaves: 2048, pinnedBidgoLeaves: 2050},
		// bid_factors32 in bid32_div.go is a second hand-ported copy of the
		// same Intel C bid_factors[1024][2] table (declared as [][2]int
		// instead of int8 pairs), used on the live bid32_div trailing-zero
		// path.
		"bid_factors32":   {goName: "BidFactors", generated: generatedtables.BidFactors},
		"bid_half128":     {goName: "BidHalf128", generated: generatedtables.BidHalf128},
		"bid_half192":     {goName: "BidHalf192", generated: generatedtables.BidHalf192},
		"bid_half256":     {goName: "BidHalf256", generated: generatedtables.BidHalf256},
		"bid_half64":      {goName: "BidHalf64", generated: generatedtables.BidHalf64},
		"bid_Kx128":       {goName: "BidKx128", generated: generatedtables.BidKx128},
		"bid_Kx192":       {goName: "BidKx192", generated: generatedtables.BidKx192},
		"bid_Kx256":       {goName: "BidKx256", generated: generatedtables.BidKx256},
		"bid_Kx64":        {goName: "BidKx64", generated: generatedtables.BidKx64},
		"bid_maskhigh128": {goName: "BidMaskHigh128", generated: generatedtables.BidMaskHigh128},
		"bid_mask128":     {goName: "BidMask128", generated: generatedtables.BidMask128},
		"bid_mask192":     {goName: "BidMask192", generated: generatedtables.BidMask192},
		"bid_mask256":     {goName: "BidMask256", generated: generatedtables.BidMask256},
		"bid_mask64":      {goName: "BidMask64", generated: generatedtables.BidMask64},
		"bid_midi_tbl":    {goName: "BidMidiTbl", generated: generatedtables.BidMidiTbl},
		"bid_midpoint128": {goName: "BidMidpoint128", generated: generatedtables.BidMidpoint128},
		"bid_midpoint192": {goName: "BidMidpoint192", generated: generatedtables.BidMidpoint192},
		"bid_midpoint256": {goName: "BidMidpoint256", generated: generatedtables.BidMidpoint256},
		"bid_midpoint64":  {goName: "BidMidpoint64", generated: generatedtables.BidMidpoint64},
		// Intel C ends the table with a commented-out 114th row ("114-bit n <
		// 10^35"); the active C array has 113 rows. bidgo ported that
		// commented-out row as a live 114th entry. Indexing never reaches it
		// (nr_bits <= 113 for coefficients below 10^34), so the prefix
		// comparison plus both pinned lengths covers the live surface.
		"bid_nr_digits":  {goName: "BidNrDigits", generated: generatedtables.BidNrDigits, pinnedGeneratedLeaves: 452, pinnedBidgoLeaves: 456},
		"mod10_18_tbl":   {goName: "Mod10_18Tbl", generated: generatedtables.Mod10_18Tbl},
		"bid_onehalf128": {goName: "BidOneHalf128", generated: generatedtables.BidOneHalf128},
		// bid_onehalf128_round64 / bid_shiftright128_round64 /
		// bid_maskhigh128_round64 in round_integral64.go are local copies of
		// the first 22 entries of the Intel bid128.c tables (the 64-bit
		// round-integral path only indexes the 128-bit rows).
		"bid_onehalf128_round64":    {goName: "BidOneHalf128", generated: generatedtables.BidOneHalf128, pinnedGeneratedLeaves: 34, pinnedBidgoLeaves: 22},
		"bid_shiftright128_round64": {goName: "BidShiftRight128", generated: generatedtables.BidShiftRight128, pinnedGeneratedLeaves: 34, pinnedBidgoLeaves: 22},
		"bid_maskhigh128_round64":   {goName: "BidMaskHigh128", generated: generatedtables.BidMaskHigh128, pinnedGeneratedLeaves: 34, pinnedBidgoLeaves: 22},
		// Intel declares bid_mult_factor in four C files; bidgo ports each one
		// under a distinct name.
		"bid_mult_factor":        {goName: "BidMultFactor", generated: generatedtables.BidMultFactor},
		"bid32_mult_factor":      {goName: "Bid32MultFactor", generated: generatedtables.Bid32MultFactor},
		"bid_mult_factor_minmax": {goName: "BidMultFactorMinmax", generated: generatedtables.BidMultFactorMinmax},
		"bid_mult_factor32":      {goName: "BidMultFactor32", generated: generatedtables.BidMultFactor32},
		// bid64MultFactor and bidMultFactor64 in decimal64.go are two more
		// hand-ported copies of the Intel bid64_minmax.c bid_mult_factor
		// table (used by the Decimal64Pure Min/Max and MinNum/MaxNum paths).
		"bid64MultFactor":              {goName: "BidMultFactorMinmax", generated: generatedtables.BidMultFactorMinmax},
		"bidMultFactor64":              {goName: "BidMultFactorMinmax", generated: generatedtables.BidMultFactorMinmax},
		"bid_packed_10000_zeros":       {goName: "BidPacked10000Zeros", generated: generatedtables.BidPacked10000Zeros},
		"bid_power10_index_binexp":     {goName: "BidPower10IndexBinExp", generated: generatedtables.BidPower10IndexBinExp},
		"bid_power10_index_binexp_128": {goName: "BidPower10IndexBinExp128", generated: generatedtables.BidPower10IndexBinExp128},
		"bid_power10_table_128":        {goName: "BidPower10Table128", generated: generatedtables.BidPower10Table128},
		"bid_reciprocals10_128":        {goName: "BidReciprocals10_128", generated: generatedtables.BidReciprocals10_128},
		"bid_reciprocals10_64":         {goName: "BidReciprocals10_64", generated: generatedtables.BidReciprocals10_64},
		// bid64Reciprocals10 in decimal64.go is a second hand-ported copy of
		// Intel's bid_reciprocals10_64 (used by the Decimal64Pure quantize
		// path).
		"bid64Reciprocals10": {goName: "BidReciprocals10_64", generated: generatedtables.BidReciprocals10_64},
		"bid_recip_scale":    {goName: "BidRecipScale", generated: generatedtables.BidRecipScale},
		// bidgo declares one extra trailing rounding-mode row.
		"bid_round_const_table": {goName: "BidRoundConstTable", generated: generatedtables.BidRoundConstTable, pinnedGeneratedLeaves: 95, pinnedBidgoLeaves: 114},
		// bid64RoundConstTable in decimal64.go copies Intel's
		// bid_round_const_table with 18-wide rows (extra-digits 0..17; the
		// quantize path never rounds off more than 17 digits) and one extra
		// trailing rounding-mode row, so the rows must be compared pairwise:
		// generated [5][19] vs bidgo [6][18], common 5x18 prefix. The 19th
		// generated column is fully covered by the bid_round_const_table
		// entry above.
		"bid64RoundConstTable": {
			goName:                "BidRoundConstTable",
			generated:             generatedtables.BidRoundConstTable,
			pinnedGeneratedLeaves: 95,
			pinnedBidgoLeaves:     108,
			pinnedGeneratedRowLen: 19,
			pinnedBidgoRowLen:     18,
		},
		"bid_round_const_table_128": {
			goName:         "BidRoundConstTable128",
			excludedReason: "bidgo computes this table at init via make_bid_round_const_table_128(); there is no static literal to extract. Its only input, bid_power10_table_128, is value-compared above.",
		},
		"bid_shiftright128":     {goName: "BidShiftRight128", generated: generatedtables.BidShiftRight128},
		"bid_short_recip_scale": {goName: "BidShortRecipScale", generated: generatedtables.BidShortRecipScale},
		// bid64ShortRecipScale in decimal64.go is a second hand-ported copy
		// of Intel's bid_short_recip_scale.
		"bid64ShortRecipScale": {goName: "BidShortRecipScale", generated: generatedtables.BidShortRecipScale},
		"bid_ten2k64":          {goName: "BidTen2K64", generated: generatedtables.BidTen2K64},
		// pow10 in decimal64.go is a Go-side powers-of-ten table rather than
		// a labeled Intel transcription, but its content is exactly Intel's
		// bid_ten2k64 (10^0..10^19), so it is anchored to the same generated
		// table to check against digit typos.
		"pow10":              {goName: "BidTen2K64", generated: generatedtables.BidTen2K64},
		"bid_ten2k256":       {goName: "BidTen2K256", generated: generatedtables.BidTen2K256},
		"bid_ten2mxtrunc128": {goName: "BidTen2MxTrunc128", generated: generatedtables.BidTen2MxTrunc128},
		"bid_ten2mxtrunc192": {goName: "BidTen2MxTrunc192", generated: generatedtables.BidTen2MxTrunc192},
		"bid_ten2mxtrunc256": {goName: "BidTen2MxTrunc256", generated: generatedtables.BidTen2MxTrunc256},
		"bid_ten2mxtrunc64":  {goName: "BidTen2MxTrunc64", generated: generatedtables.BidTen2MxTrunc64},
		"bid_ten2mk128":      {goName: "BidTen2MK128", generated: generatedtables.BidTen2MK128},
		"bid_ten2mk128trunc": {goName: "BidTen2MK128Trunc", generated: generatedtables.BidTen2MK128Trunc},
		// bid_ten2mk64 in tables_intconv.go is declared as an alias of
		// bid_ten2mk64_round64; comparing the literal here anchors both
		// names.
		"bid_ten2mk64_round64": {goName: "BidTen2MK64", generated: generatedtables.BidTen2MK64},
		// bid_round128_19_38_for64 in convert64.go carries local copies of
		// the first 19 rows of the Intel bid_round.c bid_round128_19_38
		// tables (the uint64 input path only needs 1 <= x <= 19).
		"bid_Kx128_for64":           {goName: "BidKx128", generated: generatedtables.BidKx128, pinnedGeneratedLeaves: 74, pinnedBidgoLeaves: 38},
		"bid_ten2mxtrunc128_for64":  {goName: "BidTen2MxTrunc128", generated: generatedtables.BidTen2MxTrunc128, pinnedGeneratedLeaves: 74, pinnedBidgoLeaves: 38},
		"bid_Ex128m128_for64":       {goName: "BidEx128M128", generated: generatedtables.BidEx128M128, pinnedGeneratedLeaves: 37, pinnedBidgoLeaves: 19},
		"bid_half128_for64":         {goName: "BidHalf128", generated: generatedtables.BidHalf128, pinnedGeneratedLeaves: 37, pinnedBidgoLeaves: 19},
		"bid_mask128_for64":         {goName: "BidMask128", generated: generatedtables.BidMask128, pinnedGeneratedLeaves: 37, pinnedBidgoLeaves: 19},
		"bid_roundbound_128":        {goName: "BidRoundbound128", generated: generatedtables.BidRoundbound128},
		"bid_breakpoints_binary32":  {goName: "BidBreakpointsBinary32", generated: generatedtables.BidBreakpointsBinary32},
		"bid_exponents_binary32":    {goName: "BidExponentsBinary32", generated: generatedtables.BidExponentsBinary32},
		"bid_multipliers1_binary32": {goName: "BidMultipliers1Binary32", generated: generatedtables.BidMultipliers1Binary32},
		"bid_multipliers2_binary32": {goName: "BidMultipliers2Binary32", generated: generatedtables.BidMultipliers2Binary32},
		"bid_breakpoints_binary64":  {goName: "BidBreakpointsBinary64", generated: generatedtables.BidBreakpointsBinary64},
		"bid_exponents_binary64":    {goName: "BidExponentsBinary64", generated: generatedtables.BidExponentsBinary64},
		"bid_multipliers1_binary64": {goName: "BidMultipliers1Binary64", generated: generatedtables.BidMultipliers1Binary64},
		"bid_multipliers2_binary64": {goName: "BidMultipliers2Binary64", generated: generatedtables.BidMultipliers2Binary64},
	}
}

func TestCTablegenOutputMatchesBidgoPortedTables(t *testing.T) {
	manifest, err := cgen.LoadManifest(tablegenManifestRel)
	if err != nil {
		t.Fatalf("load tablegen manifest: %v", err)
	}
	mappings := bidgoTableMappings()

	manifestGoNames := map[string]bool{}
	for _, table := range manifest.Tables {
		if manifestGoNames[table.GoName] {
			t.Fatalf("duplicate go_name %q in tablegen manifest", table.GoName)
		}
		manifestGoNames[table.GoName] = true
	}

	referencedGoNames := map[string]bool{}
	for bidgoVar, mapping := range mappings {
		if mapping.excludedReason == "" && (mapping.generated == nil || mapping.goName == "") {
			t.Errorf("mapping entry %q must either compare a generated value under a manifest go_name or carry an exclusion reason", bidgoVar)
			continue
		}
		if mapping.excludedReason != "" && mapping.generated != nil {
			t.Errorf("mapping entry %q has both a generated value and an exclusion reason", bidgoVar)
			continue
		}
		if mapping.goName != "" {
			if !manifestGoNames[mapping.goName] {
				t.Errorf("mapping entry %q references go_name %q which does not exist in tablegen_manifest.json", bidgoVar, mapping.goName)
			}
			referencedGoNames[mapping.goName] = true
		}
	}
	for goName := range manifestGoNames {
		if !referencedGoNames[goName] {
			t.Errorf("tablegen manifest table %q has no bidgo mapping entry; add a comparison or a documented exclusion", goName)
		}
	}

	pkg := bidgoPkg(t)

	// Exhaustive set over bidgo: every package-level composite-literal var is a
	// potential hand-ported table and must be mapped or excluded; every
	// mapping key must still exist as a bidgo package-level var.
	for _, varName := range pkg.compositeLitVarNames() {
		if _, ok := mappings[varName]; !ok {
			t.Errorf("bidgo package-level composite-literal var %q has no cross-check mapping entry; add a value comparison against the c-tablegen output or a documented exclusion", varName)
		}
	}
	for bidgoVar := range mappings {
		if _, ok := pkg.decls[bidgoVar]; !ok {
			t.Errorf("mapping entry %q does not exist as a bidgo package-level var; remove the stale entry", bidgoVar)
		}
	}
	if t.Failed() {
		t.FailNow()
	}

	bidgoVars := make([]string, 0, len(mappings))
	for bidgoVar := range mappings {
		bidgoVars = append(bidgoVars, bidgoVar)
	}
	sort.Strings(bidgoVars)

	for _, bidgoVar := range bidgoVars {
		mapping := mappings[bidgoVar]
		t.Run(bidgoVar, func(t *testing.T) {
			if mapping.excludedReason != "" {
				t.Skipf("excluded from static value comparison: %s", mapping.excludedReason)
			}
			genLeaves := flattenGeneratedValue(t, reflect.ValueOf(mapping.generated))
			portLeaves := pkg.flattenVar(t, bidgoVar)

			if mapping.pinnedGeneratedLeaves != 0 || mapping.pinnedBidgoLeaves != 0 {
				if len(genLeaves) != mapping.pinnedGeneratedLeaves {
					t.Fatalf("generated %s leaf count = %d, pinned %d", mapping.goName, len(genLeaves), mapping.pinnedGeneratedLeaves)
				}
				if len(portLeaves) != mapping.pinnedBidgoLeaves {
					t.Fatalf("bidgo %s leaf count = %d, pinned %d", bidgoVar, len(portLeaves), mapping.pinnedBidgoLeaves)
				}
			} else if len(genLeaves) != len(portLeaves) {
				t.Fatalf("leaf count mismatch: generated %s has %d scalars, bidgo %s has %d scalars",
					mapping.goName, len(genLeaves), bidgoVar, len(portLeaves))
			}

			if mapping.pinnedGeneratedRowLen != 0 || mapping.pinnedBidgoRowLen != 0 {
				compareRowPrefix(t, mapping, bidgoVar, genLeaves, portLeaves)
				return
			}

			n := len(genLeaves)
			if len(portLeaves) < n {
				n = len(portLeaves)
			}
			for i := 0; i < n; i++ {
				if genLeaves[i].Cmp(portLeaves[i]) != 0 {
					t.Fatalf("value mismatch at scalar index %d: generated %s = %s, bidgo %s = %s",
						i, mapping.goName, genLeaves[i], bidgoVar, portLeaves[i])
				}
			}
		})
	}
}

// compareRowPrefix compares two tables that slice the same Intel rows at
// different row widths: each common row is compared over the common column
// prefix. Both row lengths and both leaf totals must be pinned.
func compareRowPrefix(t *testing.T, mapping tableMapping, bidgoVar string, genLeaves, portLeaves []*big.Int) {
	t.Helper()
	if mapping.pinnedGeneratedRowLen <= 0 || mapping.pinnedBidgoRowLen <= 0 ||
		mapping.pinnedGeneratedLeaves == 0 || mapping.pinnedBidgoLeaves == 0 {
		t.Fatalf("row-prefix mapping for %q must pin both row lengths and both leaf totals", bidgoVar)
	}
	if mapping.pinnedGeneratedLeaves%mapping.pinnedGeneratedRowLen != 0 {
		t.Fatalf("pinned generated leaves %d not divisible by pinned generated row length %d",
			mapping.pinnedGeneratedLeaves, mapping.pinnedGeneratedRowLen)
	}
	if mapping.pinnedBidgoLeaves%mapping.pinnedBidgoRowLen != 0 {
		t.Fatalf("pinned bidgo leaves %d not divisible by pinned bidgo row length %d",
			mapping.pinnedBidgoLeaves, mapping.pinnedBidgoRowLen)
	}
	genRows := mapping.pinnedGeneratedLeaves / mapping.pinnedGeneratedRowLen
	portRows := mapping.pinnedBidgoLeaves / mapping.pinnedBidgoRowLen
	rows := genRows
	if portRows < rows {
		rows = portRows
	}
	cols := mapping.pinnedGeneratedRowLen
	if mapping.pinnedBidgoRowLen < cols {
		cols = mapping.pinnedBidgoRowLen
	}
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			gen := genLeaves[r*mapping.pinnedGeneratedRowLen+c]
			port := portLeaves[r*mapping.pinnedBidgoRowLen+c]
			if gen.Cmp(port) != 0 {
				t.Fatalf("value mismatch at row %d column %d: generated %s = %s, bidgo %s = %s",
					r, c, mapping.goName, gen, bidgoVar, port)
			}
		}
	}
}

// --- bidgo source extraction ---

type bidgoPackage struct {
	fset  *token.FileSet
	info  *types.Info
	decls map[string]ast.Expr
}

// loadBidgoPackageOnce parses and type-checks the bidgo package exactly once
// per test binary run; the result is immutable and shared by all tests.
var loadBidgoPackageOnce = sync.OnceValues(loadBidgoPackage)

func bidgoPkg(t *testing.T) *bidgoPackage {
	t.Helper()
	pkg, err := loadBidgoPackageOnce()
	if err != nil {
		t.Fatalf("load bidgo package: %v", err)
	}
	return pkg
}

func loadBidgoPackage() (*bidgoPackage, error) {
	entries, err := os.ReadDir(bidgoDirRel)
	if err != nil {
		return nil, fmt.Errorf("read bidgo dir: %w", err)
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(bidgoDirRel, name), nil, 0)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		files = append(files, file)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no bidgo source files found")
	}

	info := &types.Info{Types: map[ast.Expr]types.TypeAndValue{}}
	conf := types.Config{Importer: importer.ForCompiler(fset, "source", nil)}
	if _, err := conf.Check("bidgo", fset, files, info); err != nil {
		return nil, fmt.Errorf("type-check bidgo package: %w", err)
	}

	decls := map[string]ast.Expr{}
	for _, file := range files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok || len(vs.Names) != len(vs.Values) {
					continue
				}
				for i, name := range vs.Names {
					decls[name.Name] = vs.Values[i]
				}
			}
		}
	}
	return &bidgoPackage{fset: fset, info: info, decls: decls}, nil
}

// compositeLitVarNames returns the sorted names of all package-level vars
// whose initializer is a composite literal — the exhaustive enumeration of
// potential hand-ported tables.
func (p *bidgoPackage) compositeLitVarNames() []string {
	var names []string
	for name, expr := range p.decls {
		if _, ok := expr.(*ast.CompositeLit); ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func (p *bidgoPackage) flattenVar(t *testing.T, name string) []*big.Int {
	t.Helper()
	expr, ok := p.decls[name]
	if !ok {
		t.Fatalf("bidgo package has no top-level var %q", name)
	}
	return p.flattenExpr(t, expr, 0)
}

func (p *bidgoPackage) flattenExpr(t *testing.T, expr ast.Expr, aliasDepth int) []*big.Int {
	t.Helper()
	if aliasDepth > 4 {
		t.Fatalf("var alias chain too deep at %s", p.fset.Position(expr.Pos()))
	}
	switch e := expr.(type) {
	case *ast.CompositeLit:
		return p.flattenComposite(t, e)
	case *ast.Ident:
		// A var initialized from another var (e.g. bid_ten2mk64 =
		// bid_ten2mk64_round64) resolves to the target's literal. Named
		// constants fall through to the constant leaf path below.
		if target, ok := p.decls[e.Name]; ok {
			return p.flattenExpr(t, target, aliasDepth+1)
		}
	}
	return []*big.Int{p.constLeaf(t, expr)}
}

func (p *bidgoPackage) flattenComposite(t *testing.T, lit *ast.CompositeLit) []*big.Int {
	t.Helper()
	tv, ok := p.info.Types[lit]
	if !ok {
		t.Fatalf("no type info for composite literal at %s", p.fset.Position(lit.Pos()))
	}
	switch typ := tv.Type.Underlying().(type) {
	case *types.Array:
		out := make([]*big.Int, 0, typ.Len())
		for _, elt := range lit.Elts {
			if _, isKV := elt.(*ast.KeyValueExpr); isKV {
				t.Fatalf("indexed array literal entries are not supported at %s", p.fset.Position(elt.Pos()))
			}
			out = append(out, p.flattenExpr(t, elt, 0)...)
		}
		if missing := int(typ.Len()) - len(lit.Elts); missing > 0 {
			out = append(out, zeroLeaves(t, p.fset, lit, typ.Elem(), missing)...)
		}
		return out
	case *types.Slice:
		var out []*big.Int
		for _, elt := range lit.Elts {
			if _, isKV := elt.(*ast.KeyValueExpr); isKV {
				t.Fatalf("indexed slice literal entries are not supported at %s", p.fset.Position(elt.Pos()))
			}
			out = append(out, p.flattenExpr(t, elt, 0)...)
		}
		return out
	case *types.Struct:
		fieldExprs := make([]ast.Expr, typ.NumFields())
		keyed := false
		for i, elt := range lit.Elts {
			if kv, isKV := elt.(*ast.KeyValueExpr); isKV {
				keyed = true
				key, ok := kv.Key.(*ast.Ident)
				if !ok {
					t.Fatalf("unsupported struct literal key at %s", p.fset.Position(kv.Pos()))
				}
				idx := structFieldIndex(typ, key.Name)
				if idx < 0 {
					t.Fatalf("struct literal key %q not found at %s", key.Name, p.fset.Position(kv.Pos()))
				}
				fieldExprs[idx] = kv.Value
				continue
			}
			if keyed {
				t.Fatalf("mixed keyed/positional struct literal at %s", p.fset.Position(elt.Pos()))
			}
			fieldExprs[i] = elt
		}
		var out []*big.Int
		for i := 0; i < typ.NumFields(); i++ {
			if fieldExprs[i] == nil {
				out = append(out, zeroLeaves(t, p.fset, lit, typ.Field(i).Type(), 1)...)
				continue
			}
			out = append(out, p.flattenExpr(t, fieldExprs[i], 0)...)
		}
		return out
	default:
		t.Fatalf("unsupported composite literal type %v at %s", tv.Type, p.fset.Position(lit.Pos()))
		return nil
	}
}

func (p *bidgoPackage) constLeaf(t *testing.T, expr ast.Expr) *big.Int {
	t.Helper()
	tv, ok := p.info.Types[expr]
	if !ok || tv.Value == nil {
		t.Fatalf("expression at %s is not a typed constant; the extractor only supports constant table literals",
			p.fset.Position(expr.Pos()))
	}
	val := constant.ToInt(tv.Value)
	if val.Kind() != constant.Int {
		t.Fatalf("expression at %s is not an integer constant", p.fset.Position(expr.Pos()))
	}
	out, ok := new(big.Int).SetString(val.ExactString(), 10)
	if !ok {
		t.Fatalf("cannot parse constant %q at %s", val.ExactString(), p.fset.Position(expr.Pos()))
	}
	return out
}

func zeroLeaves(t *testing.T, fset *token.FileSet, lit *ast.CompositeLit, typ types.Type, count int) []*big.Int {
	t.Helper()
	per := scalarLeafCount(t, fset, lit, typ)
	out := make([]*big.Int, 0, per*count)
	for i := 0; i < per*count; i++ {
		out = append(out, big.NewInt(0))
	}
	return out
}

func scalarLeafCount(t *testing.T, fset *token.FileSet, lit *ast.CompositeLit, typ types.Type) int {
	t.Helper()
	switch u := typ.Underlying().(type) {
	case *types.Basic:
		return 1
	case *types.Array:
		return int(u.Len()) * scalarLeafCount(t, fset, lit, u.Elem())
	case *types.Struct:
		total := 0
		for i := 0; i < u.NumFields(); i++ {
			total += scalarLeafCount(t, fset, lit, u.Field(i).Type())
		}
		return total
	default:
		t.Fatalf("unsupported zero-fill element type %v at %s", typ, fset.Position(lit.Pos()))
		return 0
	}
}

func structFieldIndex(typ *types.Struct, name string) int {
	for i := 0; i < typ.NumFields(); i++ {
		if typ.Field(i).Name() == name {
			return i
		}
	}
	return -1
}

// --- generated tables flattening ---

func flattenGeneratedValue(t *testing.T, v reflect.Value) []*big.Int {
	t.Helper()
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		var out []*big.Int
		for i := 0; i < v.Len(); i++ {
			out = append(out, flattenGeneratedValue(t, v.Index(i))...)
		}
		return out
	case reflect.Struct:
		var out []*big.Int
		for i := 0; i < v.NumField(); i++ {
			out = append(out, flattenGeneratedValue(t, v.Field(i))...)
		}
		return out
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []*big.Int{new(big.Int).SetUint64(v.Uint())}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []*big.Int{big.NewInt(v.Int())}
	default:
		t.Fatalf("unsupported generated table value kind %v", v.Kind())
		return nil
	}
}
