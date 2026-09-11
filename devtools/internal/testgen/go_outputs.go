package testgen

import (
	"fmt"
	"strings"
)

func GenerateGoTestOutputs(repoRoot string, manifest Manifest, spec SharedSpec) (map[string][]byte, error) {
	generators := []func() (map[string][]byte, error){
		func() (map[string][]byte, error) { return GenerateDectestTestOutputs(repoRoot, spec) },
		func() (map[string][]byte, error) { return GenerateDectestGoportOutputs(repoRoot, spec) },
		func() (map[string][]byte, error) { return GenerateReadtestDispatchOutputs(repoRoot, manifest) },
		func() (map[string][]byte, error) { return GenerateReadtestTestOutputs(spec) },
		func() (map[string][]byte, error) { return GenerateReadtestGoportOutputs(repoRoot, manifest, spec) },
		func() (map[string][]byte, error) { return GenerateFFITestOutputs(spec) },
		GenerateTier1ArithmeticLongOutputs,
		GenerateTier1CompareConversionLongOutputs,
		func() (map[string][]byte, error) { return GenerateDecnumberDifferentialOutputs(manifest) },
		GenerateD32ExhaustiveOutputs,
		GenerateFinitePathsOutputs,
		GenerateTier1Outputs,
		GenerateBidCodecVectorTestOutputs,
		func() (map[string][]byte, error) { return GenerateBidStringVectorTestOutputs(spec), nil },
		func() (map[string][]byte, error) { return GeneratePublicParityOutputs(repoRoot, manifest) },
		func() (map[string][]byte, error) { return GenerateTestspecPackageOutputs(repoRoot) },
	}
	outputs := map[string][]byte{}
	for _, generate := range generators {
		files, err := generate()
		if err != nil {
			return nil, err
		}
		for path, data := range files {
			if !(strings.HasPrefix(path, "../bid754-go/") && strings.HasSuffix(path, ".go")) && path != publicParityInventoryPath {
				continue
			}
			if _, exists := outputs[path]; exists {
				return nil, fmt.Errorf("duplicate Go verification output %s", path)
			}
			outputs[path] = data
		}
	}
	return outputs, nil
}
