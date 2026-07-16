package benchrows

import (
	"testing"
)

const (
	benchmarkInputsPath = "../../testdata/benchmark_inputs.json"
	benchmarkRowsPath   = "../../testdata/benchmark_rows.json"
)

func loadProductionPrepared(t *testing.T) Prepared {
	t.Helper()
	inputs, err := LoadInputs(benchmarkInputsPath)
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := Prepare(inputs)
	if err != nil {
		t.Fatal(err)
	}
	return prepared
}

func loadCheckedDescriptor(t *testing.T) Descriptor {
	t.Helper()
	descriptor, err := LoadDescriptor(benchmarkRowsPath)
	if err != nil {
		t.Fatal(err)
	}
	return descriptor
}

// TestDescriptorLoadsAndValidates pins the structural gate on the checked-in
// hand-maintained descriptor: format version, vocabulary closure, layer
// counts, and anchor-group shape.
func TestDescriptorLoadsAndValidates(t *testing.T) {
	descriptor := loadCheckedDescriptor(t)
	if got := len(descriptor.Rows); got != 340 {
		t.Fatalf("descriptor rows = %d, want 340", got)
	}
}

// TestGoportRowTablesMatchDescriptor closed-world-compares the Go-port row
// tables against the descriptor's goport layer, group by group: same row-name
// sets, same counts, no duplicates.
func TestGoportRowTablesMatchDescriptor(t *testing.T) {
	descriptor := loadCheckedDescriptor(t)
	prepared := loadProductionPrepared(t)

	descriptorNames := make(map[string]map[string]bool)
	for _, row := range descriptor.LayerRows(LayerGoport) {
		if descriptorNames[row.Group] == nil {
			descriptorNames[row.Group] = make(map[string]bool)
		}
		descriptorNames[row.Group][row.Name] = true
	}

	for group, wantCount := range map[string]int{
		"bid32":        23,
		"bid64":        19,
		"bid128":       19,
		"bid64_mixed":  12,
		"bid128_mixed": 12,
	} {
		group := group
		wantCount := wantCount
		t.Run(group, func(t *testing.T) {
			rows, err := GroupRows(prepared, group)
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != wantCount {
				t.Fatalf("table row count = %d, want %d", len(rows), wantCount)
			}
			if len(descriptorNames[group]) != wantCount {
				t.Fatalf("descriptor goport/%s row count = %d, want %d", group, len(descriptorNames[group]), wantCount)
			}
			seen := make(map[string]bool, len(rows))
			for _, row := range rows {
				if seen[row.Name] {
					t.Fatalf("duplicate table row %q", row.Name)
				}
				seen[row.Name] = true
				if !descriptorNames[group][row.Name] {
					t.Errorf("table row %q is missing from the descriptor goport layer", row.Name)
				}
			}
			for name := range descriptorNames[group] {
				if !seen[name] {
					t.Errorf("descriptor goport row %q is missing from the table", name)
				}
			}
		})
	}
}

// TestGoportRowsHonorDeclaredSinkDiscipline executes every Go-port row once
// against poisoned sinks and requires each row to write exactly the sink its
// descriptor row declares, with the declared status-flag behavior.
func TestGoportRowsHonorDeclaredSinkDiscipline(t *testing.T) {
	descriptor := loadCheckedDescriptor(t)
	prepared := loadProductionPrepared(t)

	tableRows := make(map[string]map[string]Row)
	for _, group := range []string{"bid32", "bid64", "bid128", "bid64_mixed", "bid128_mixed"} {
		rows, err := GroupRows(prepared, group)
		if err != nil {
			t.Fatal(err)
		}
		tableRows[group] = make(map[string]Row, len(rows))
		for _, row := range rows {
			tableRows[group][row.Name] = row
		}
	}

	for _, spec := range descriptor.LayerRows(LayerGoport) {
		spec := spec
		t.Run(spec.Group+"/"+spec.Name, func(t *testing.T) {
			row, ok := tableRows[spec.Group][spec.Name]
			if !ok {
				t.Fatalf("descriptor goport row %s/%s has no table row", spec.Group, spec.Name)
			}
			if _, err := ObserveOnce(row, spec); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBenchmarkInputsAreFiniteAndExact(t *testing.T) {
	loadProductionPrepared(t)
}

func TestBenchmarkScaleExponentFitsCInt(t *testing.T) {
	for _, tc := range []struct {
		name     string
		exponent int64
		want     bool
	}{
		{name: "min", exponent: -1 << 31, want: true},
		{name: "max", exponent: 1<<31 - 1, want: true},
		{name: "below_min", exponent: -1<<31 - 1, want: false},
		{name: "above_max", exponent: 1 << 31, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := scaleExponentFitsCInt(tc.exponent); got != tc.want {
				t.Fatalf("scaleExponentFitsCInt(%d) = %t, want %t", tc.exponent, got, tc.want)
			}
		})
	}
}

func TestBenchmarkExactOperandsRejectCohortCoercion(t *testing.T) {
	for _, tc := range []struct {
		name  string
		parse func(string) error
		input string
	}{
		{
			name: "decimal32_precision_plus_one_trailing_zero",
			parse: func(input string) error {
				_, err := parseExactDecimal32(input)
				return err
			},
			input: "1000000.0",
		},
		{
			name: "decimal64_precision_plus_one_trailing_zero",
			parse: func(input string) error {
				_, err := parseExactDecimal64(input)
				return err
			},
			input: "1000000000000000.0",
		},
		{
			name: "decimal128_precision_plus_one_trailing_zero",
			parse: func(input string) error {
				_, err := parseExactDecimal128(input)
				return err
			},
			input: "1000000000000000000000000000000000.0",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.parse(tc.input); err == nil {
				t.Fatalf("benchmark input %q accepted after exact numeric parsing silently changed its cohort", tc.input)
			}
		})
	}
}

func TestBenchmarkExactOperandsAcceptCohortBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name  string
		parse func(string) error
		input string
	}{
		{
			name: "decimal32_minimum_quantum",
			parse: func(input string) error {
				_, err := parseExactDecimal32(input)
				return err
			},
			input: "1E-101",
		},
		{
			name: "decimal32_maximum_quantum_and_precision",
			parse: func(input string) error {
				_, err := parseExactDecimal32(input)
				return err
			},
			input: "9999999E+90",
		},
		{
			name: "decimal64_minimum_quantum",
			parse: func(input string) error {
				_, err := parseExactDecimal64(input)
				return err
			},
			input: "1E-398",
		},
		{
			name: "decimal64_maximum_quantum_and_precision",
			parse: func(input string) error {
				_, err := parseExactDecimal64(input)
				return err
			},
			input: "9999999999999999E+369",
		},
		{
			name: "decimal128_minimum_quantum",
			parse: func(input string) error {
				_, err := parseExactDecimal128(input)
				return err
			},
			input: "1E-6176",
		},
		{
			name: "decimal128_maximum_quantum_and_precision",
			parse: func(input string) error {
				_, err := parseExactDecimal128(input)
				return err
			},
			input: "9999999999999999999999999999999999E+6111",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.parse(tc.input); err != nil {
				t.Fatalf("benchmark input %q rejected at a valid cohort boundary: %v", tc.input, err)
			}
		})
	}
}
