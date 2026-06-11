package testgen

import (
	"encoding/json"
	"fmt"
	"os"
)

type Manifest struct {
	Output          string                `json:"output"`
	DectestSuites   []DectestSuiteSpec    `json:"dectest_suites"`
	ReadTests       []ReadTestSpec        `json:"readtests"`
	ReadTestGroups  []ReadTestGroupSpec   `json:"readtest_groups"`
	ReadProfiles    []ReadTestProfileSpec `json:"readtest_profiles"`
	FuzzTests       []FuzzTestSpec        `json:"fuzztests"`
	FFITests        []FFITestSpec         `json:"ffi_tests"`
	BidCodecVectors *BidCodecVectorSpec   `json:"bid_codec_vectors,omitempty"`
}

type DectestSuiteSpec struct {
	Name                string   `json:"name"`
	Pattern             string   `json:"pattern"`
	TestType            string   `json:"test_type"`
	Directory           string   `json:"directory"`
	SupportedOperations []string `json:"supported_operations"`
	IgnoredOperations   []string `json:"ignored_operations,omitempty"`
	ExcludePrefixes     []string `json:"exclude_prefixes,omitempty"`
}

type ReadTestSpec struct {
	Name                    string   `json:"name"`
	Group                   string   `json:"group,omitempty"`
	Format                  string   `json:"format"`
	Header                  string   `json:"header"`
	Source                  string   `json:"source"`
	Function                string   `json:"function"`
	Kind                    string   `json:"kind"`
	OutputType              string   `json:"output_type,omitempty"`
	InputTypes              []string `json:"input_types,omitempty"`
	CompareGroup            string   `json:"compare_group,omitempty"`
	NativeCompareSkipReason string   `json:"native_compare_skip_reason,omitempty"`
	Statuses                []string `json:"statuses"`
	RoundingModes           []int    `json:"rounding_modes"`
	Limit                   int      `json:"limit,omitempty"`
}

type ReadTestGroupSpec struct {
	Name          string              `json:"name"`
	Format        string              `json:"format"`
	Header        string              `json:"header"`
	Source        string              `json:"source"`
	Statuses      []string            `json:"statuses"`
	RoundingModes []int               `json:"rounding_modes"`
	Cases         []ReadTestEntrySpec `json:"cases"`
}

type ReadTestProfileSpec struct {
	Name          string   `json:"name"`
	Header        string   `json:"header"`
	Source        string   `json:"source"`
	Formats       []string `json:"formats"`
	Statuses      []string `json:"statuses"`
	RoundingModes []int    `json:"rounding_modes"`
	Selection     string   `json:"selection"`
}

type ReadTestEntrySpec struct {
	Name     string `json:"name"`
	Function string `json:"function"`
	Kind     string `json:"kind"`
	Limit    int    `json:"limit,omitempty"`
}

type FuzzTestSpec struct {
	Name       string   `json:"name"`
	TestType   string   `json:"test_type"`
	Sources    []string `json:"sources"`
	Operations []string `json:"operations"`
	Limit      int      `json:"limit"`
}

type FFITestSpec struct {
	Name             string   `json:"name"`
	Symbols          string   `json:"symbols"`
	Functions        []string `json:"functions"`
	FunctionPatterns []string `json:"function_patterns,omitempty"`
	CasesPerFunction int      `json:"cases_per_function"`
	Seed             uint64   `json:"seed"`
}

type BidCodecVectorSpec struct {
	Output               string `json:"output"`
	Seed                 int64  `json:"seed"`
	RandomCasesPerFormat int    `json:"random_cases_per_format"`
}

type SharedSpec struct {
	DectestSuites           []GeneratedDectestSuite            `json:"dectest_suites"`
	DectestFileAudits       []GeneratedDectestFileAudit        `json:"dectest_file_audits,omitempty"`
	DectestRuntimeSkipAudit []GeneratedDectestRuntimeSkipAudit `json:"dectest_runtime_skip_audit,omitempty"`
	ReadCases               []GeneratedReadCase                `json:"read_cases"`
	ReadtestProfileAudit    []GeneratedReadtestProfileAudit    `json:"readtest_profile_audit,omitempty"`
	FuzzCases               []GeneratedFuzzCase                `json:"fuzz_cases"`
	FFICases                []GeneratedFFICase                 `json:"ffi_cases"`
}

// SpecIndex is the on-disk root of the split generated test spec layout.
// It carries every SharedSpec section except read_cases/ffi_cases, which are
// stored as per-suite/per-function shard files listed here in generation order.
type SpecIndex struct {
	DectestSuites           []GeneratedDectestSuite            `json:"dectest_suites"`
	DectestFileAudits       []GeneratedDectestFileAudit        `json:"dectest_file_audits,omitempty"`
	DectestRuntimeSkipAudit []GeneratedDectestRuntimeSkipAudit `json:"dectest_runtime_skip_audit,omitempty"`
	ReadtestProfileAudit    []GeneratedReadtestProfileAudit    `json:"readtest_profile_audit,omitempty"`
	FuzzCases               []GeneratedFuzzCase                `json:"fuzz_cases"`
	ReadtestShardFiles      []string                           `json:"readtest_shard_files"`
	FFIShardFiles           []string                           `json:"ffi_shard_files"`
}

// ReadtestShardHeader holds the eleven GeneratedReadCase fields that are
// constant for every case of one readtest suite.
type ReadtestShardHeader struct {
	Suite                   string   `json:"suite"`
	Group                   string   `json:"group,omitempty"`
	Format                  string   `json:"format"`
	Header                  string   `json:"header"`
	Source                  string   `json:"source"`
	Function                string   `json:"function"`
	Kind                    string   `json:"kind"`
	OutputType              string   `json:"output_type,omitempty"`
	InputTypes              []string `json:"input_types,omitempty"`
	CompareGroup            string   `json:"compare_group,omitempty"`
	NativeCompareSkipReason string   `json:"native_compare_skip_reason,omitempty"`
}

// ReadtestShard is one generated/testspec/readtest/<suite>.json file.
type ReadtestShard struct {
	ReadtestShardHeader
	Cases []ReadtestShardCase `json:"cases"`
}

// ReadtestShardCase holds the per-case GeneratedReadCase fields.
type ReadtestShardCase struct {
	ID       string   `json:"id"`
	Line     int      `json:"line"`
	Operands []string `json:"operands"`
	Expected string   `json:"expected"`
	Status   string   `json:"status"`
	Rounding int      `json:"rounding"`
}

// FFIShardHeader holds the seven GeneratedFFICase fields that are constant
// for every case of one FFI function.
type FFIShardHeader struct {
	Suite       string `json:"suite"`
	Format      string `json:"format"`
	Operation   string `json:"operation"`
	Function    string `json:"function"`
	LinkName    string `json:"link_name"`
	Declaration string `json:"declaration"`
	Source      string `json:"source"`
}

// FFIShard is one generated/testspec/ffi/<function>.json file.
type FFIShard struct {
	FFIShardHeader
	Cases []FFIShardCase `json:"cases"`
}

// FFIShardCase holds the per-case GeneratedFFICase fields.
type FFIShardCase struct {
	ID       string   `json:"id"`
	Rounding int      `json:"rounding"`
	Operands []string `json:"operands"`
}

type GeneratedDectestSuite struct {
	Name                string   `json:"name"`
	Pattern             string   `json:"pattern"`
	TestType            string   `json:"test_type"`
	Files               []string `json:"files"`
	SupportedOperations []string `json:"supported_operations,omitempty"`
	IgnoredOperations   []string `json:"ignored_operations,omitempty"`
}

type GeneratedDectestFileAudit struct {
	File                              string                       `json:"file"`
	Operations                        []string                     `json:"operations"`
	SelectedSuites                    []string                     `json:"selected_suites,omitempty"`
	UnsupportedBySuite                map[string][]string          `json:"unsupported_by_suite,omitempty"`
	UnsupportedReasonsBySuite         map[string]map[string]string `json:"unsupported_reasons_by_suite,omitempty"`
	UnsupportedClassificationsBySuite map[string]map[string]string `json:"unsupported_classifications_by_suite,omitempty"`
}

type GeneratedDectestRuntimeSkipAudit struct {
	Suite       string         `json:"suite"`
	Cases       int            `json:"cases"`
	SkipReasons map[string]int `json:"skip_reasons,omitempty"`
}

type GeneratedReadCase struct {
	Suite                   string   `json:"suite"`
	Group                   string   `json:"group,omitempty"`
	Format                  string   `json:"format"`
	Header                  string   `json:"header"`
	Source                  string   `json:"source"`
	ID                      string   `json:"id"`
	Line                    int      `json:"line"`
	Function                string   `json:"function"`
	Kind                    string   `json:"kind"`
	OutputType              string   `json:"output_type,omitempty"`
	InputTypes              []string `json:"input_types,omitempty"`
	CompareGroup            string   `json:"compare_group,omitempty"`
	NativeCompareSkipReason string   `json:"native_compare_skip_reason,omitempty"`
	Operands                []string `json:"operands"`
	Expected                string   `json:"expected"`
	Status                  string   `json:"status"`
	Rounding                int      `json:"rounding"`
}

type GeneratedReadtestProfileAudit struct {
	Profile           string                           `json:"profile"`
	Header            string                           `json:"header"`
	Source            string                           `json:"source"`
	Selection         string                           `json:"selection"`
	TotalFunctions    int                              `json:"total_functions"`
	SelectedFunctions int                              `json:"selected_functions"`
	ExcludedFunctions int                              `json:"excluded_functions"`
	Functions         []GeneratedReadtestFunctionAudit `json:"functions"`
}

type GeneratedReadtestFunctionAudit struct {
	Function       string   `json:"function"`
	OutputType     string   `json:"output_type,omitempty"`
	InputTypes     []string `json:"input_types,omitempty"`
	CompareGroup   string   `json:"compare_group,omitempty"`
	Selected       bool     `json:"selected"`
	Format         string   `json:"format,omitempty"`
	Kind           string   `json:"kind,omitempty"`
	Group          string   `json:"group,omitempty"`
	Reason         string   `json:"reason"`
	Classification string   `json:"classification"`
}

type GeneratedFuzzCase struct {
	Suite        string   `json:"suite"`
	TestType     string   `json:"test_type"`
	Source       string   `json:"source"`
	ID           string   `json:"id"`
	Operation    string   `json:"operation"`
	Operands     []string `json:"operands"`
	Expected     string   `json:"expected"`
	Precision    int      `json:"precision"`
	RoundingMode string   `json:"rounding_mode"`
	MaxExponent  int      `json:"max_exponent"`
	MinExponent  int      `json:"min_exponent"`
	Clamp        int      `json:"clamp"`
}

type GeneratedFFICase struct {
	Suite       string   `json:"suite"`
	ID          string   `json:"id"`
	Format      string   `json:"format"`
	Operation   string   `json:"operation"`
	Function    string   `json:"function"`
	LinkName    string   `json:"link_name"`
	Declaration string   `json:"declaration"`
	Source      string   `json:"source"`
	Rounding    int      `json:"rounding"`
	Operands    []string `json:"operands"`
}

func LoadManifest(path string) (Manifest, error) {
	var manifest Manifest

	data, err := os.ReadFile(path)
	if err != nil {
		return manifest, fmt.Errorf("read manifest %q: %w", path, err)
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return manifest, fmt.Errorf("parse manifest %q: %w", path, err)
	}
	if manifest.Output == "" {
		return manifest, fmt.Errorf("manifest %q: output is required", path)
	}
	if len(manifest.DectestSuites) == 0 {
		return manifest, fmt.Errorf("manifest %q: dectest_suites must not be empty", path)
	}
	for i, suite := range manifest.DectestSuites {
		if suite.Name == "" || suite.Pattern == "" || suite.TestType == "" || suite.Directory == "" || len(suite.SupportedOperations) == 0 {
			return manifest, fmt.Errorf("manifest %q: dectest_suites[%d] is incomplete", path, i)
		}
	}
	for i, read := range manifest.ReadTests {
		if read.Name == "" || read.Format == "" || read.Header == "" || read.Source == "" || read.Function == "" || read.Kind == "" || len(read.Statuses) == 0 || len(read.RoundingModes) == 0 || read.Limit < 0 {
			return manifest, fmt.Errorf("manifest %q: readtests[%d] is incomplete", path, i)
		}
	}
	for i, group := range manifest.ReadTestGroups {
		if group.Name == "" || group.Format == "" || group.Header == "" || group.Source == "" || len(group.Statuses) == 0 || len(group.RoundingModes) == 0 || len(group.Cases) == 0 {
			return manifest, fmt.Errorf("manifest %q: readtest_groups[%d] is incomplete", path, i)
		}
		for j, tc := range group.Cases {
			if tc.Name == "" || tc.Function == "" || tc.Kind == "" || tc.Limit < 0 {
				return manifest, fmt.Errorf("manifest %q: readtest_groups[%d].cases[%d] is incomplete", path, i, j)
			}
		}
	}
	for i, profile := range manifest.ReadProfiles {
		if profile.Name == "" || profile.Header == "" || profile.Source == "" || len(profile.Formats) == 0 || profile.Selection == "" {
			return manifest, fmt.Errorf("manifest %q: readtest_profiles[%d] is incomplete", path, i)
		}
	}
	for i, fuzz := range manifest.FuzzTests {
		if fuzz.Name == "" || fuzz.TestType == "" || len(fuzz.Sources) == 0 || len(fuzz.Operations) == 0 || fuzz.Limit <= 0 {
			return manifest, fmt.Errorf("manifest %q: fuzztests[%d] is incomplete", path, i)
		}
	}
	for i, ffi := range manifest.FFITests {
		if ffi.Name == "" || ffi.Symbols == "" || (len(ffi.Functions) == 0 && len(ffi.FunctionPatterns) == 0) || ffi.CasesPerFunction <= 0 {
			return manifest, fmt.Errorf("manifest %q: ffi_tests[%d] is incomplete", path, i)
		}
	}
	if manifest.BidCodecVectors == nil {
		return manifest, fmt.Errorf("manifest %q: bid_codec_vectors is required", path)
	}
	if manifest.BidCodecVectors.Output == "" || manifest.BidCodecVectors.RandomCasesPerFormat <= 0 {
		return manifest, fmt.Errorf("manifest %q: bid_codec_vectors is incomplete", path)
	}

	return manifest, nil
}
