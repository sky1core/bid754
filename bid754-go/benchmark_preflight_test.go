//go:build cgo && bid754_native
// +build cgo,bid754_native

package bid754

// Untimed cross-layer benchmark row preflight.
//
// The hand-pinned descriptor (testdata/benchmark_rows.json) declares every
// benchmark row of the three Go-side measured layers (public Go API, Intel C
// direct, Go mechanical port) plus the generated-Rust layer verified in
// bid754-rs/ffi-verify. This file closed-world-compares the descriptor
// against the per-layer row tables and then executes every public-API and
// Go-port row exactly once per wiring fixture, exact-comparing the observed
// bits/flags against the Intel C benchmark leg itself as the anchor. There is
// deliberately no parallel expected-value switch: the anchor is the executed
// Intel C row, so a wiring mistake cannot skew the observed and expected legs
// the same way.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/benchrows"
)

const (
	benchmarkRowsDescriptorPath = "testdata/benchmark_rows.json"
	intelCBenchmarkPoisonFlags  = ^uint32(0)
	publicFlagsUntouchedPoison  = ExceptionFlags(0x55)
)

type benchmarkWiringFixture struct {
	name   string
	inputs benchmarkInputs
}

func benchmarkWiringFixtures(t *testing.T) []benchmarkWiringFixture {
	t.Helper()
	production := loadBenchmarkInputs(t)
	return []benchmarkWiringFixture{
		{
			name:   "production",
			inputs: production,
		},
		{
			name: "x_less_than_y",
			inputs: benchmarkInputs{
				FormatVersion:  2,
				IntegerOperand: production.IntegerOperand,
				ScaleExponent:  production.ScaleExponent,
				Decimal32:      benchmarkInputPair{X: "4.25", Y: "33", Z: "2.5"},
				Decimal64:      benchmarkInputPair{X: "7.125", Y: "55", Z: "3.25"},
				Decimal128:     benchmarkInputPair{X: "11.0625", Y: "96", Z: "5.5"},
			},
		},
		{
			name: "x_greater_than_y",
			inputs: benchmarkInputs{
				FormatVersion:  2,
				IntegerOperand: production.IntegerOperand,
				ScaleExponent:  production.ScaleExponent,
				Decimal32:      benchmarkInputPair{X: "33", Y: "4.25", Z: "2.5"},
				Decimal64:      benchmarkInputPair{X: "55", Y: "7.125", Z: "3.25"},
				Decimal128:     benchmarkInputPair{X: "96", Y: "11.0625", Z: "5.5"},
			},
		},
	}
}

func loadBenchmarkRowDescriptor(t *testing.T) benchrows.Descriptor {
	t.Helper()
	descriptor, err := benchrows.LoadDescriptor(benchmarkRowsDescriptorPath)
	if err != nil {
		t.Fatal(err)
	}
	return descriptor
}

func benchrowsInputsFromRoot(inputs benchmarkInputs) benchrows.Inputs {
	return benchrows.Inputs{
		FormatVersion:  inputs.FormatVersion,
		IntegerOperand: inputs.IntegerOperand,
		ScaleExponent:  inputs.ScaleExponent,
		Decimal32:      benchrows.InputPair{X: inputs.Decimal32.X, Y: inputs.Decimal32.Y, Z: inputs.Decimal32.Z},
		Decimal64:      benchrows.InputPair{X: inputs.Decimal64.X, Y: inputs.Decimal64.Y, Z: inputs.Decimal64.Z},
		Decimal128:     benchrows.InputPair{X: inputs.Decimal128.X, Y: inputs.Decimal128.Y, Z: inputs.Decimal128.Z},
	}
}

func benchrowsPreparedForInputs(t *testing.T, inputs benchmarkInputs) benchrows.Prepared {
	t.Helper()
	prepared, err := benchrows.Prepare(benchrowsInputsFromRoot(inputs))
	if err != nil {
		t.Fatal(err)
	}
	return prepared
}

func alignedBenchmarkGroupRows(t *testing.T, inputs benchmarkInputs) map[string][]alignedBenchmarkRow {
	t.Helper()
	return map[string][]alignedBenchmarkRow{
		"bid32":        alignedBID32BenchmarkRowsForInputs(t, inputs),
		"bid64":        alignedBID64BenchmarkRowsForInputs(t, inputs),
		"bid128":       alignedBID128BenchmarkRowsForInputs(t, inputs),
		"bid64_mixed":  alignedMixedBID64BenchmarkRowsForInputs(t, inputs),
		"bid128_mixed": alignedMixedBID128BenchmarkRowsForInputs(t, inputs),
	}
}

func benchrowsGroupRows(t *testing.T, prepared benchrows.Prepared) map[string][]benchrows.Row {
	t.Helper()
	groups := make(map[string][]benchrows.Row, 5)
	for _, group := range []string{"bid32", "bid64", "bid128", "bid64_mixed", "bid128_mixed"} {
		rows, err := benchrows.GroupRows(prepared, group)
		if err != nil {
			t.Fatal(err)
		}
		groups[group] = rows
	}
	return groups
}

func intelCBenchmarkGroupRows() map[string][]intelCBenchmarkRow {
	return map[string][]intelCBenchmarkRow{
		"bid32":        intelCBID32BenchmarkRows,
		"bid64":        intelCBID64BenchmarkRows,
		"bid128":       intelCBID128BenchmarkRows,
		"bid64_mixed":  intelCMixedBID64BenchmarkRows,
		"bid128_mixed": intelCMixedBID128BenchmarkRows,
	}
}

func initIntelCBenchmarkNativeInputs(t *testing.T, inputs benchmarkInputs) {
	t.Helper()
	if !nativeBenchCBIDInit(
		inputs.Decimal32.X,
		inputs.Decimal32.Y,
		inputs.Decimal32.Z,
		inputs.Decimal64.X,
		inputs.Decimal64.Y,
		inputs.Decimal64.Z,
		inputs.Decimal128.X,
		inputs.Decimal128.Y,
		inputs.Decimal128.Z,
		inputs.IntegerOperand,
		inputs.ScaleExponent,
	) {
		t.Fatal("Intel C benchmark inputs are not finite and exact")
	}
}

// intelCResultKindForDescriptor maps a descriptor result kind onto the Intel C
// row-table result metadata so the closed-world check can cross-compare them.
func intelCResultKindForDescriptor(spec benchrows.DescriptorRow) (intelCBenchmarkResultKind, error) {
	switch spec.Result {
	case benchrows.ResultD32:
		return intelCBenchmarkBID32, nil
	case benchrows.ResultD64:
		return intelCBenchmarkBID64, nil
	case benchrows.ResultD128:
		return intelCBenchmarkBID128, nil
	case benchrows.ResultI64:
		return intelCBenchmarkInt64, nil
	case benchrows.ResultText:
		return intelCBenchmarkString, nil
	case benchrows.ResultPredicate:
		if spec.Group == "bid32" {
			return intelCBenchmarkBool32, nil
		}
		return intelCBenchmarkBool64, nil
	default:
		return 0, fmt.Errorf("descriptor row %s/%s has unmapped result kind %q", spec.Group, spec.Name, spec.Result)
	}
}

// anchorRowGroup is one (op, operands) anchor group: exactly one Intel C row
// whose one-shot execution is the expected observation, plus every public-API
// and Go-port row that must match it.
type anchorRowGroup struct {
	intel   benchrows.DescriptorRow
	members []benchrows.DescriptorRow
}

// benchmarkAnchorGroups folds the descriptor into anchor groups ordered by
// the Intel C rows' file order. The Rust layer is verified against the same
// descriptor by bid754-rs/ffi-verify/tests/benchmark_wiring.rs.
func benchmarkAnchorGroups(descriptor benchrows.Descriptor) []anchorRowGroup {
	index := make(map[string]int)
	groups := make([]anchorRowGroup, 0, len(descriptor.LayerRows(benchrows.LayerIntelC)))
	for _, spec := range descriptor.LayerRows(benchrows.LayerIntelC) {
		index[spec.AnchorKey()] = len(groups)
		groups = append(groups, anchorRowGroup{intel: spec})
	}
	for _, spec := range descriptor.Rows {
		if spec.Layer != benchrows.LayerPublicAPI && spec.Layer != benchrows.LayerGoport {
			continue
		}
		at := index[spec.AnchorKey()]
		groups[at].members = append(groups[at].members, spec)
	}
	return groups
}

func intelCBenchmarkRawFlags(t *testing.T, flags ExceptionFlags) uint32 {
	t.Helper()
	var raw uint32
	if flags&FlagInvalidOperation != 0 {
		raw |= 0x01
	}
	if flags&FlagDivisionByZero != 0 {
		raw |= 0x04
	}
	if flags&FlagOverflow != 0 {
		raw |= 0x08
	}
	if flags&FlagUnderflow != 0 {
		raw |= 0x10
	}
	if flags&FlagInexact != 0 {
		raw |= 0x20
	}
	known := FlagInvalidOperation | FlagDivisionByZero | FlagOverflow | FlagUnderflow | FlagInexact
	if unknown := flags &^ known; unknown != 0 {
		t.Fatalf("benchmark preflight observed flags outside the Intel raw set: %s", unknown)
	}
	return raw
}

func exceptionFlagsFromRaw(t *testing.T, raw uint32) ExceptionFlags {
	t.Helper()
	var flags ExceptionFlags
	if raw&0x01 != 0 {
		flags |= FlagInvalidOperation
	}
	if raw&0x04 != 0 {
		flags |= FlagDivisionByZero
	}
	if raw&0x08 != 0 {
		flags |= FlagOverflow
	}
	if raw&0x10 != 0 {
		flags |= FlagUnderflow
	}
	if raw&0x20 != 0 {
		flags |= FlagInexact
	}
	if raw&^uint32(0x3d) != 0 {
		t.Fatalf("Intel C anchor flags %#x are outside the Intel raw set", raw)
	}
	return flags
}

func decimal128BIDFromBits(lo, hi uint64) Decimal128BID {
	var raw [16]byte
	binary.LittleEndian.PutUint64(raw[0:8], lo)
	binary.LittleEndian.PutUint64(raw[8:16], hi)
	return Decimal128BID(raw)
}

func decimal128BIDBits(d Decimal128BID) (lo, hi uint64) {
	raw := d.ToBytes()
	return binary.LittleEndian.Uint64(raw[0:8]), binary.LittleEndian.Uint64(raw[8:16])
}

// intelCObservation runs one Intel C benchmark row once against poisoned C
// sinks and normalizes the sink snapshot into the shared observation form.
func intelCObservation(t *testing.T, spec benchrows.DescriptorRow, row intelCBenchmarkRow) benchrows.Observation {
	t.Helper()
	nativeBenchCBIDResetSinks()
	row.run(1)
	snapshot := nativeBenchCBIDSnapshot()

	observation := benchrows.Observation{Kind: spec.Result}
	switch spec.Result {
	case benchrows.ResultD32:
		observation.Bits32 = snapshot.BID32
	case benchrows.ResultD64:
		observation.Bits64 = snapshot.BID64
	case benchrows.ResultD128:
		observation.Bits128Lo = snapshot.BID128Low
		observation.Bits128Hi = snapshot.BID128High
	case benchrows.ResultI64:
		observation.Int64 = snapshot.Int64
	case benchrows.ResultPredicate:
		if row.result == intelCBenchmarkBool32 {
			observation.Predicate = snapshot.BID32 != 0
		} else {
			observation.Predicate = snapshot.BID64 != 0
		}
	case benchrows.ResultText:
		if snapshot.String == "" || snapshot.String == "<unset>" {
			t.Fatalf("Intel C row %s/%s did not capture its to_string result: %q", spec.Group, spec.Name, snapshot.String)
		}
		observation.Text = snapshot.String
	default:
		t.Fatalf("Intel C row %s/%s has unmapped result kind %q", spec.Group, spec.Name, spec.Result)
	}

	switch spec.Status {
	case benchrows.StatusFlagsObserved:
		if snapshot.Flags == intelCBenchmarkPoisonFlags {
			t.Fatalf("Intel C row %s/%s did not write the status-flag sink", spec.Group, spec.Name)
		}
		observation.Flags = snapshot.Flags
		observation.HasFlags = true
	case benchrows.StatusValueOnly:
		if snapshot.Flags != intelCBenchmarkPoisonFlags {
			t.Errorf("Intel C row %s/%s is value-only but wrote status flags %#x", spec.Group, spec.Name, snapshot.Flags)
		}
	default:
		t.Fatalf("Intel C row %s/%s has unobservable status %q", spec.Group, spec.Name, spec.Status)
	}
	return observation
}

// benchmarkPreflightOperands carries the fixture's public-API X operands for
// the to_string parse-back check.
type benchmarkPreflightOperands struct {
	x32  Decimal32BID
	x64  Decimal64BID
	x128 Decimal128BID
}

func newBenchmarkPreflightOperands(t *testing.T, inputs benchmarkInputs) benchmarkPreflightOperands {
	t.Helper()
	return benchmarkPreflightOperands{
		x32:  exactBenchmarkDecimal32(t, inputs.Decimal32.X),
		x64:  exactBenchmarkDecimal64(t, inputs.Decimal64.X),
		x128: exactBenchmarkDecimal128(t, inputs.Decimal128.X),
	}
}

// checkIntelCToStringAnchor re-derives the C to_string anchor through the
// independent one-shot C helper and requires the anchor text to parse back to
// the fixture's X operand bits through the public parser.
func checkIntelCToStringAnchor(t *testing.T, spec benchrows.DescriptorRow, anchor benchrows.Observation, operands benchmarkPreflightOperands) {
	t.Helper()
	switch spec.Group {
	case "bid32":
		if once := nativeBenchCBID32ToStringOnce(); anchor.Text != once {
			t.Errorf("Intel C bid32 to_string sink = %q, one-shot C result = %q", anchor.Text, once)
		}
		value, err := NewDecimal32BIDDirect(anchor.Text)
		if err != nil {
			t.Fatalf("parse Intel C Decimal32 to_string anchor %q: %v", anchor.Text, err)
		}
		if value.ToUint32() != operands.x32.ToUint32() {
			t.Errorf("to_string anchor %q round-tripped to %#08x, want operand %#08x", anchor.Text, value.ToUint32(), operands.x32.ToUint32())
		}
	case "bid64":
		if once := nativeBenchCBID64ToStringOnce(); anchor.Text != once {
			t.Errorf("Intel C bid64 to_string sink = %q, one-shot C result = %q", anchor.Text, once)
		}
		value, err := NewDecimal64BIDDirect(anchor.Text)
		if err != nil {
			t.Fatalf("parse Intel C Decimal64 to_string anchor %q: %v", anchor.Text, err)
		}
		if value.ToUint64() != operands.x64.ToUint64() {
			t.Errorf("to_string anchor %q round-tripped to %#016x, want operand %#016x", anchor.Text, value.ToUint64(), operands.x64.ToUint64())
		}
	case "bid128":
		if once := nativeBenchCBID128ToStringOnce(); anchor.Text != once {
			t.Errorf("Intel C bid128 to_string sink = %q, one-shot C result = %q", anchor.Text, once)
		}
		value, err := NewDecimal128BIDDirect(anchor.Text)
		if err != nil {
			t.Fatalf("parse Intel C Decimal128 to_string anchor %q: %v", anchor.Text, err)
		}
		gotLo, gotHi := decimal128BIDBits(value)
		wantLo, wantHi := decimal128BIDBits(operands.x128)
		if gotLo != wantLo || gotHi != wantHi {
			t.Errorf("to_string anchor %q round-tripped to %016x:%016x, want operand %016x:%016x", anchor.Text, gotHi, gotLo, wantHi, wantLo)
		}
	default:
		t.Fatalf("to_string anchor in unexpected group %q", spec.Group)
	}
}

// primePublicBenchmarkSinks poisons every sink the public row is expected to
// write with the inverse of the anchor observation, so a row that fails to
// write its sink (or writes the wrong one) cannot match the anchor.
func primePublicBenchmarkSinks(t *testing.T, spec benchrows.DescriptorRow, anchor benchrows.Observation) {
	t.Helper()
	switch spec.Result {
	case benchrows.ResultD32:
		alignedSink32 = Decimal32BID(^anchor.Bits32)
	case benchrows.ResultD64:
		alignedSink64 = Decimal64BID(^anchor.Bits64)
	case benchrows.ResultD128:
		alignedSink128 = decimal128BIDFromBits(^anchor.Bits128Lo, ^anchor.Bits128Hi)
	case benchrows.ResultI64:
		alignedSinkInt64 = ^anchor.Int64
	case benchrows.ResultPredicate:
		alignedSinkBool = !anchor.Predicate
	case benchrows.ResultText:
		alignedSinkString = anchor.Text + "#unwritten"
	default:
		t.Fatalf("public row %s/%s has unmapped result kind %q", spec.Group, spec.Name, spec.Result)
	}
	switch spec.Status {
	case benchrows.StatusFlagsObserved:
		alignedSinkFlags = exceptionFlagsFromRaw(t, anchor.Flags) ^ FlagInexact
	case benchrows.StatusValueOnly:
		alignedSinkFlags = publicFlagsUntouchedPoison
	case benchrows.StatusErrorChannel:
		alignedSinkFlags = publicFlagsUntouchedPoison
		alignedSinkErr = errors.New("benchmark row did not write the error sink")
	default:
		t.Fatalf("public row %s/%s has unknown status %q", spec.Group, spec.Name, spec.Status)
	}
}

// snapshotPublicBenchmarkObservation reads the public sinks back into the
// shared observation form and enforces the row's status contract.
func snapshotPublicBenchmarkObservation(t *testing.T, spec benchrows.DescriptorRow) benchrows.Observation {
	t.Helper()
	observation := benchrows.Observation{Kind: spec.Result}
	switch spec.Result {
	case benchrows.ResultD32:
		observation.Bits32 = alignedSink32.ToUint32()
	case benchrows.ResultD64:
		observation.Bits64 = alignedSink64.ToUint64()
	case benchrows.ResultD128:
		observation.Bits128Lo, observation.Bits128Hi = decimal128BIDBits(alignedSink128)
	case benchrows.ResultI64:
		observation.Int64 = alignedSinkInt64
	case benchrows.ResultPredicate:
		observation.Predicate = alignedSinkBool
	case benchrows.ResultText:
		observation.Text = alignedSinkString
	}
	switch spec.Status {
	case benchrows.StatusFlagsObserved:
		observation.Flags = intelCBenchmarkRawFlags(t, alignedSinkFlags)
		observation.HasFlags = true
	case benchrows.StatusValueOnly:
		if alignedSinkFlags != publicFlagsUntouchedPoison {
			t.Errorf("public row %s/%s is value-only but wrote status flags %v", spec.Group, spec.Name, alignedSinkFlags)
		}
	case benchrows.StatusErrorChannel:
		if alignedSinkErr != nil {
			t.Errorf("public row %s/%s error = %v, want nil", spec.Group, spec.Name, alignedSinkErr)
		}
		if alignedSinkFlags != publicFlagsUntouchedPoison {
			t.Errorf("public row %s/%s reports through the error channel but wrote status flags %v", spec.Group, spec.Name, alignedSinkFlags)
		}
	}
	return observation
}

// compareBenchmarkObservation exact-compares one member-layer observation
// against the Intel C anchor observation.
func compareBenchmarkObservation(t *testing.T, spec benchrows.DescriptorRow, got, anchor benchrows.Observation) {
	t.Helper()
	if got.Kind != anchor.Kind {
		t.Errorf("%s row %s/%s observation kind = %q, anchor kind = %q", spec.Layer, spec.Group, spec.Name, got.Kind, anchor.Kind)
		return
	}
	switch anchor.Kind {
	case benchrows.ResultD32:
		if got.Bits32 != anchor.Bits32 {
			t.Errorf("%s row %s/%s bits = %#08x, anchor = %#08x", spec.Layer, spec.Group, spec.Name, got.Bits32, anchor.Bits32)
		}
	case benchrows.ResultD64:
		if got.Bits64 != anchor.Bits64 {
			t.Errorf("%s row %s/%s bits = %#016x, anchor = %#016x", spec.Layer, spec.Group, spec.Name, got.Bits64, anchor.Bits64)
		}
	case benchrows.ResultD128:
		if got.Bits128Lo != anchor.Bits128Lo || got.Bits128Hi != anchor.Bits128Hi {
			t.Errorf("%s row %s/%s bits = %016x:%016x, anchor = %016x:%016x",
				spec.Layer, spec.Group, spec.Name, got.Bits128Hi, got.Bits128Lo, anchor.Bits128Hi, anchor.Bits128Lo)
		}
	case benchrows.ResultI64:
		if got.Int64 != anchor.Int64 {
			t.Errorf("%s row %s/%s int64 = %d, anchor = %d", spec.Layer, spec.Group, spec.Name, got.Int64, anchor.Int64)
		}
	case benchrows.ResultPredicate:
		if got.Predicate != anchor.Predicate {
			t.Errorf("%s row %s/%s predicate = %t, anchor = %t", spec.Layer, spec.Group, spec.Name, got.Predicate, anchor.Predicate)
		}
	case benchrows.ResultText:
		if got.Text != anchor.Text {
			t.Errorf("%s row %s/%s text = %q, anchor = %q", spec.Layer, spec.Group, spec.Name, got.Text, anchor.Text)
		}
	}
	switch spec.Status {
	case benchrows.StatusFlagsObserved:
		if !anchor.HasFlags {
			t.Errorf("%s row %s/%s observes flags but its Intel C anchor is value-only", spec.Layer, spec.Group, spec.Name)
		} else if got.Flags != anchor.Flags {
			t.Errorf("%s row %s/%s flags = %#08x, anchor = %#08x", spec.Layer, spec.Group, spec.Name, got.Flags, anchor.Flags)
		}
	case benchrows.StatusErrorChannel:
		if !anchor.HasFlags || anchor.Flags != 0 {
			t.Errorf("%s row %s/%s reports through the error channel but its Intel C anchor flags are %#08x (has=%t), want clean 0",
				spec.Layer, spec.Group, spec.Name, anchor.Flags, anchor.HasFlags)
		}
	}
}

// TestBenchmarkRowDescriptorMatchesLayerTables closed-world-compares the
// hand-pinned descriptor against the actual per-layer row tables: the
// public-API tables (93 rows), the Intel C tables (81 rows, including their
// result-kind and hasFlags metadata), and the Go-port tables (85 rows).
func TestBenchmarkRowDescriptorMatchesLayerTables(t *testing.T) {
	descriptor := loadBenchmarkRowDescriptor(t)
	production := loadBenchmarkInputs(t)

	descriptorNames := func(layer string) map[string]map[string]benchrows.DescriptorRow {
		byGroup := make(map[string]map[string]benchrows.DescriptorRow)
		for _, spec := range descriptor.LayerRows(layer) {
			if byGroup[spec.Group] == nil {
				byGroup[spec.Group] = make(map[string]benchrows.DescriptorRow)
			}
			byGroup[spec.Group][spec.Name] = spec
		}
		return byGroup
	}

	t.Run("public_api", func(t *testing.T) {
		specs := descriptorNames(benchrows.LayerPublicAPI)
		total := 0
		for group, rows := range alignedBenchmarkGroupRows(t, production) {
			seen := make(map[string]bool, len(rows))
			for _, row := range rows {
				if seen[row.name] {
					t.Errorf("public table %s has duplicate row %q", group, row.name)
				}
				seen[row.name] = true
				if _, ok := specs[group][row.name]; !ok {
					t.Errorf("public table row %s/%s is missing from the descriptor", group, row.name)
				}
			}
			for name := range specs[group] {
				if !seen[name] {
					t.Errorf("descriptor public_api row %s/%s is missing from the table", group, name)
				}
			}
			total += len(rows)
		}
		if total != descriptor.LayerCounts[benchrows.LayerPublicAPI] {
			t.Errorf("public table rows = %d, descriptor pins %d", total, descriptor.LayerCounts[benchrows.LayerPublicAPI])
		}
	})

	t.Run("intel_c", func(t *testing.T) {
		specs := descriptorNames(benchrows.LayerIntelC)
		total := 0
		for group, rows := range intelCBenchmarkGroupRows() {
			seen := make(map[string]bool, len(rows))
			for _, row := range rows {
				if seen[row.name] {
					t.Errorf("Intel C table %s has duplicate row %q", group, row.name)
				}
				seen[row.name] = true
				spec, ok := specs[group][row.name]
				if !ok {
					t.Errorf("Intel C table row %s/%s is missing from the descriptor", group, row.name)
					continue
				}
				wantKind, err := intelCResultKindForDescriptor(spec)
				if err != nil {
					t.Error(err)
					continue
				}
				if row.result != wantKind {
					t.Errorf("Intel C row %s/%s result kind = %d, descriptor result %q maps to %d", group, row.name, row.result, spec.Result, wantKind)
				}
				if wantFlags := spec.Status == benchrows.StatusFlagsObserved; row.hasFlags != wantFlags {
					t.Errorf("Intel C row %s/%s hasFlags = %t, descriptor status %q wants %t", group, row.name, row.hasFlags, spec.Status, wantFlags)
				}
			}
			for name := range specs[group] {
				if !seen[name] {
					t.Errorf("descriptor intel_c row %s/%s is missing from the table", group, name)
				}
			}
			total += len(rows)
		}
		if total != descriptor.LayerCounts[benchrows.LayerIntelC] {
			t.Errorf("Intel C table rows = %d, descriptor pins %d", total, descriptor.LayerCounts[benchrows.LayerIntelC])
		}
	})

	t.Run("goport", func(t *testing.T) {
		specs := descriptorNames(benchrows.LayerGoport)
		prepared := benchrowsPreparedForInputs(t, production)
		total := 0
		for group, rows := range benchrowsGroupRows(t, prepared) {
			seen := make(map[string]bool, len(rows))
			for _, row := range rows {
				if seen[row.Name] {
					t.Errorf("Go-port table %s has duplicate row %q", group, row.Name)
				}
				seen[row.Name] = true
				if _, ok := specs[group][row.Name]; !ok {
					t.Errorf("Go-port table row %s/%s is missing from the descriptor", group, row.Name)
				}
			}
			for name := range specs[group] {
				if !seen[name] {
					t.Errorf("descriptor goport row %s/%s is missing from the table", group, name)
				}
			}
			total += len(rows)
		}
		if total != descriptor.LayerCounts[benchrows.LayerGoport] {
			t.Errorf("Go-port table rows = %d, descriptor pins %d", total, descriptor.LayerCounts[benchrows.LayerGoport])
		}
	})
}

// TestBenchmarkRowPreflightCrossLayer executes every public-API and Go-port
// benchmark row exactly once per wiring fixture and exact-compares the
// observed bits/flags against the Intel C benchmark leg of the same anchor
// group.
func TestBenchmarkRowPreflightCrossLayer(t *testing.T) {
	descriptor := loadBenchmarkRowDescriptor(t)
	anchorGroups := benchmarkAnchorGroups(descriptor)
	intelGroups := intelCBenchmarkGroupRows()

	for _, fixture := range benchmarkWiringFixtures(t) {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			alignedGroups := alignedBenchmarkGroupRows(t, fixture.inputs)
			prepared := benchrowsPreparedForInputs(t, fixture.inputs)
			goportGroups := benchrowsGroupRows(t, prepared)
			operands := newBenchmarkPreflightOperands(t, fixture.inputs)
			initIntelCBenchmarkNativeInputs(t, fixture.inputs)

			for _, group := range anchorGroups {
				group := group
				t.Run(group.intel.Group+"/"+group.intel.Name, func(t *testing.T) {
					intelRow, ok := intelCRowByName(intelGroups[group.intel.Group], group.intel.Name)
					if !ok {
						t.Fatalf("Intel C table has no row %s/%s", group.intel.Group, group.intel.Name)
					}
					anchor := intelCObservation(t, group.intel, intelRow)
					if anchor.Kind == benchrows.ResultText {
						checkIntelCToStringAnchor(t, group.intel, anchor, operands)
					}

					for _, spec := range group.members {
						spec := spec
						t.Run(spec.Layer+"/"+spec.Name, func(t *testing.T) {
							switch spec.Layer {
							case benchrows.LayerPublicAPI:
								row, ok := alignedRowByName(alignedGroups[spec.Group], spec.Name)
								if !ok {
									t.Fatalf("public table has no row %s/%s", spec.Group, spec.Name)
								}
								primePublicBenchmarkSinks(t, spec, anchor)
								row.run(1)
								got := snapshotPublicBenchmarkObservation(t, spec)
								compareBenchmarkObservation(t, spec, got, anchor)
							case benchrows.LayerGoport:
								row, ok := benchrowsRowByName(goportGroups[spec.Group], spec.Name)
								if !ok {
									t.Fatalf("Go-port table has no row %s/%s", spec.Group, spec.Name)
								}
								got, err := benchrows.ObserveOnce(row, spec)
								if err != nil {
									t.Fatal(err)
								}
								compareBenchmarkObservation(t, spec, got, anchor)
							default:
								t.Fatalf("unexpected member layer %q", spec.Layer)
							}
						})
					}
				})
			}
		})
	}
}

func intelCRowByName(rows []intelCBenchmarkRow, name string) (intelCBenchmarkRow, bool) {
	for _, row := range rows {
		if row.name == name {
			return row, true
		}
	}
	return intelCBenchmarkRow{}, false
}

func alignedRowByName(rows []alignedBenchmarkRow, name string) (alignedBenchmarkRow, bool) {
	for _, row := range rows {
		if row.name == name {
			return row, true
		}
	}
	return alignedBenchmarkRow{}, false
}

func benchrowsRowByName(rows []benchrows.Row, name string) (benchrows.Row, bool) {
	for _, row := range rows {
		if row.Name == name {
			return row, true
		}
	}
	return benchrows.Row{}, false
}

func benchmarkObservationFingerprint(observation benchrows.Observation) string {
	return fmt.Sprintf(
		"kind=%s/b32=%08x/b64=%016x/b128=%016x:%016x/i64=%d/pred=%t/text=%q/flags=%08x/has_flags=%t",
		observation.Kind,
		observation.Bits32,
		observation.Bits64,
		observation.Bits128Hi,
		observation.Bits128Lo,
		observation.Int64,
		observation.Predicate,
		observation.Text,
		observation.Flags,
		observation.HasFlags,
	)
}

// TestBenchmarkRowPreflightDiscriminatesAnchorRows requires the Intel C
// anchor observations, chained across every wiring fixture, to be pairwise
// distinct within each benchmark group. A common-mode op-aliasing wiring
// error (every layer running the same wrong operation for two different row
// names) collapses two fingerprints and fails here even though every
// per-layer comparison still matches its anchor.
func TestBenchmarkRowPreflightDiscriminatesAnchorRows(t *testing.T) {
	descriptor := loadBenchmarkRowDescriptor(t)
	intelSpecs := descriptor.LayerRows(benchrows.LayerIntelC)
	intelGroups := intelCBenchmarkGroupRows()
	fixtures := benchmarkWiringFixtures(t)

	chained := make(map[string]map[string]*strings.Builder)
	for _, fixture := range fixtures {
		initIntelCBenchmarkNativeInputs(t, fixture.inputs)
		for _, spec := range intelSpecs {
			row, ok := intelCRowByName(intelGroups[spec.Group], spec.Name)
			if !ok {
				t.Fatalf("Intel C table has no row %s/%s", spec.Group, spec.Name)
			}
			observation := intelCObservation(t, spec, row)
			if chained[spec.Group] == nil {
				chained[spec.Group] = make(map[string]*strings.Builder)
			}
			builder := chained[spec.Group][spec.Name]
			if builder == nil {
				builder = &strings.Builder{}
				chained[spec.Group][spec.Name] = builder
			}
			builder.WriteString(fixture.name)
			builder.WriteByte('=')
			builder.WriteString(benchmarkObservationFingerprint(observation))
			builder.WriteByte('\n')
		}
	}

	for group, rows := range chained {
		fingerprints := make(map[string]string, len(rows))
		for name, builder := range rows {
			fingerprint := builder.String()
			if previous, duplicate := fingerprints[fingerprint]; duplicate {
				t.Errorf("%s rows %q and %q have indistinguishable Intel C anchor observations across all wiring fixtures", group, previous, name)
			} else {
				fingerprints[fingerprint] = name
			}
		}
	}
}
