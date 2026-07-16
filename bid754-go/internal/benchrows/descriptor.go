// Package benchrows carries the Go mechanical-port benchmark row table, the
// shared exact-operand benchmark input contract, and the loader for the
// hand-pinned cross-layer benchmark row descriptor
// (testdata/benchmark_rows.json). The descriptor is a hand-maintained pin:
// no generator, template, manifest, or emitting script may read or write it.
//
// The package exists so the Go-port benchmark rows can be executed from the
// module root (Go test files cannot be imported across packages): the native
// benchmark preflight runs every row once, untimed, and exact-compares the
// observed bits/flags against the Intel C benchmark leg.
package benchrows

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// DescriptorFormatVersion pins the accepted benchmark_rows.json format.
const DescriptorFormatVersion = 1

// Descriptor layer identifiers.
const (
	LayerPublicAPI = "public_api"
	LayerIntelC    = "intel_c"
	LayerGoport    = "goport"
	LayerRust      = "rust"
)

// Descriptor result kinds.
const (
	ResultD32       = "d32"
	ResultD64       = "d64"
	ResultD128      = "d128"
	ResultI64       = "i64"
	ResultPredicate = "predicate"
	ResultText      = "text"
)

// Descriptor status contracts.
const (
	StatusFlagsObserved = "flags_observed"
	StatusValueOnly     = "value_only"
	StatusErrorChannel  = "error_channel"
)

// Descriptor rounding contracts.
const (
	RoundingExplicitNearestEven = "explicit_nearest_even"
	RoundingFixedNearestEven    = "fixed_nearest_even"
	RoundingNotApplicable       = "not_applicable"
)

var descriptorLayers = []string{LayerPublicAPI, LayerIntelC, LayerGoport, LayerRust}

var descriptorGroups = map[string]bool{
	"bid32":        true,
	"bid64":        true,
	"bid128":       true,
	"bid64_mixed":  true,
	"bid128_mixed": true,
}

var descriptorResults = map[string]bool{
	ResultD32:       true,
	ResultD64:       true,
	ResultD128:      true,
	ResultI64:       true,
	ResultPredicate: true,
	ResultText:      true,
}

var descriptorRoundings = map[string]bool{
	RoundingExplicitNearestEven: true,
	RoundingFixedNearestEven:    true,
	RoundingNotApplicable:       true,
}

var descriptorStatuses = map[string]bool{
	StatusFlagsObserved: true,
	StatusValueOnly:     true,
	StatusErrorChannel:  true,
}

var descriptorOperandTokens = map[string]bool{
	"x32": true, "y32": true, "z32": true, "integer32": true,
	"x64": true, "y64": true, "z64": true, "integer64": true,
	"x128": true, "y128": true, "z128": true, "integer128": true,
	"integer_operand": true, "scale_exponent": true,
	"decimal32_x_text": true, "decimal64_x_text": true, "decimal128_x_text": true,
}

// DescriptorRow is one benchmark row identity pin: which pinned Intel BID C
// operation a named benchmark row of one measured layer must execute, with
// its exact operand binding, result kind, rounding contract, and status
// observation contract.
type DescriptorRow struct {
	Layer    string   `json:"layer"`
	Group    string   `json:"group"`
	Name     string   `json:"name"`
	Op       string   `json:"op"`
	Operands []string `json:"operands"`
	Result   string   `json:"result"`
	Rounding string   `json:"rounding"`
	Status   string   `json:"status"`
}

// AnchorKey groups rows that must observe identical results: rows of every
// layer that execute the same pinned Intel C operation on the same operand
// binding fold into one anchor group whose expected observation is the Intel
// C benchmark leg itself.
func (r DescriptorRow) AnchorKey() string {
	return r.Op + "(" + strings.Join(r.Operands, ",") + ")"
}

// Descriptor is the parsed, structurally validated benchmark row pin.
type Descriptor struct {
	Comment       []string        `json:"comment"`
	FormatVersion int             `json:"format_version"`
	LayerCounts   map[string]int  `json:"layer_counts"`
	Rows          []DescriptorRow `json:"rows"`
}

// LayerRows returns the descriptor rows of one layer in file order.
func (d Descriptor) LayerRows(layer string) []DescriptorRow {
	rows := make([]DescriptorRow, 0, len(d.Rows))
	for _, row := range d.Rows {
		if row.Layer == layer {
			rows = append(rows, row)
		}
	}
	return rows
}

// LoadDescriptor reads and structurally validates the hand-pinned benchmark
// row descriptor: format version, closed vocabulary, per-layer counts against
// the pinned layer_counts, row uniqueness, and anchor-group shape (every
// anchor group carries exactly one intel_c row and exactly one rust row, at
// least one public_api and one goport row, and one consistent result kind).
func LoadDescriptor(path string) (Descriptor, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Descriptor{}, fmt.Errorf("read benchmark row descriptor: %w", err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var descriptor Descriptor
	if err := decoder.Decode(&descriptor); err != nil {
		return Descriptor{}, fmt.Errorf("parse benchmark row descriptor: %w", err)
	}
	if err := descriptor.validate(); err != nil {
		return Descriptor{}, fmt.Errorf("benchmark row descriptor %s: %w", path, err)
	}
	return descriptor, nil
}

func (d Descriptor) validate() error {
	if d.FormatVersion != DescriptorFormatVersion {
		return fmt.Errorf("format_version = %d, want %d", d.FormatVersion, DescriptorFormatVersion)
	}
	if len(d.LayerCounts) != len(descriptorLayers) {
		return fmt.Errorf("layer_counts has %d layers, want %d", len(d.LayerCounts), len(descriptorLayers))
	}
	layerCounts := make(map[string]int, len(descriptorLayers))
	for _, layer := range descriptorLayers {
		pinned, ok := d.LayerCounts[layer]
		if !ok {
			return fmt.Errorf("layer_counts is missing layer %q", layer)
		}
		if pinned <= 0 {
			return fmt.Errorf("layer_counts[%q] = %d, want > 0", layer, pinned)
		}
		layerCounts[layer] = 0
	}

	seen := make(map[string]bool, len(d.Rows))
	layerGroups := make(map[string]map[string]bool, len(descriptorLayers))
	anchorGroups := make(map[string][]DescriptorRow)
	for index, row := range d.Rows {
		if _, ok := layerCounts[row.Layer]; !ok {
			return fmt.Errorf("row %d: unknown layer %q", index, row.Layer)
		}
		if !descriptorGroups[row.Group] {
			return fmt.Errorf("row %d: unknown group %q", index, row.Group)
		}
		if row.Name == "" {
			return fmt.Errorf("row %d: empty name", index)
		}
		if row.Op == "" {
			return fmt.Errorf("row %d (%s/%s/%s): empty op", index, row.Layer, row.Group, row.Name)
		}
		if len(row.Operands) < 1 || len(row.Operands) > 3 {
			return fmt.Errorf("row %s/%s/%s: %d operands, want 1..3", row.Layer, row.Group, row.Name, len(row.Operands))
		}
		for _, operand := range row.Operands {
			if !descriptorOperandTokens[operand] {
				return fmt.Errorf("row %s/%s/%s: unknown operand token %q", row.Layer, row.Group, row.Name, operand)
			}
		}
		if !descriptorResults[row.Result] {
			return fmt.Errorf("row %s/%s/%s: unknown result %q", row.Layer, row.Group, row.Name, row.Result)
		}
		if !descriptorRoundings[row.Rounding] {
			return fmt.Errorf("row %s/%s/%s: unknown rounding %q", row.Layer, row.Group, row.Name, row.Rounding)
		}
		if !descriptorStatuses[row.Status] {
			return fmt.Errorf("row %s/%s/%s: unknown status %q", row.Layer, row.Group, row.Name, row.Status)
		}
		if row.Status == StatusErrorChannel && row.Layer != LayerPublicAPI {
			return fmt.Errorf("row %s/%s/%s: status %q is only valid on layer %q", row.Layer, row.Group, row.Name, StatusErrorChannel, LayerPublicAPI)
		}
		identity := row.Layer + "/" + row.Group + "/" + row.Name
		if seen[identity] {
			return fmt.Errorf("duplicate row %s", identity)
		}
		seen[identity] = true
		layerCounts[row.Layer]++
		if layerGroups[row.Layer] == nil {
			layerGroups[row.Layer] = make(map[string]bool, len(descriptorGroups))
		}
		layerGroups[row.Layer][row.Group] = true
		key := row.AnchorKey()
		anchorGroups[key] = append(anchorGroups[key], row)
	}

	for _, layer := range descriptorLayers {
		if layerCounts[layer] != d.LayerCounts[layer] {
			return fmt.Errorf("layer %q has %d rows, layer_counts pins %d", layer, layerCounts[layer], d.LayerCounts[layer])
		}
		if len(layerGroups[layer]) != len(descriptorGroups) {
			return fmt.Errorf("layer %q covers %d groups, want %d", layer, len(layerGroups[layer]), len(descriptorGroups))
		}
	}

	if len(anchorGroups) != d.LayerCounts[LayerIntelC] {
		return fmt.Errorf("%d anchor groups, want %d (one per intel_c row)", len(anchorGroups), d.LayerCounts[LayerIntelC])
	}
	for key, rows := range anchorGroups {
		perLayer := make(map[string]int, len(descriptorLayers))
		for _, row := range rows {
			perLayer[row.Layer]++
			if row.Result != rows[0].Result {
				return fmt.Errorf("anchor group %s mixes result kinds %q and %q", key, rows[0].Result, row.Result)
			}
		}
		if perLayer[LayerIntelC] != 1 {
			return fmt.Errorf("anchor group %s has %d intel_c rows, want exactly 1", key, perLayer[LayerIntelC])
		}
		if perLayer[LayerRust] != 1 {
			return fmt.Errorf("anchor group %s has %d rust rows, want exactly 1", key, perLayer[LayerRust])
		}
		if perLayer[LayerPublicAPI] < 1 {
			return fmt.Errorf("anchor group %s has no public_api row", key)
		}
		if perLayer[LayerGoport] < 1 {
			return fmt.Errorf("anchor group %s has no goport row", key)
		}
	}
	return nil
}
