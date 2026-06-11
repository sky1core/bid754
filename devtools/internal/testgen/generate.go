package testgen

func Generate(repoRoot string, manifest Manifest) (SharedSpec, error) {
	return buildSpec(repoRoot, manifest)
}
