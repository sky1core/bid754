package gotypegen

import (
	"encoding/json"
	"fmt"
	"os"
)

type Manifest struct {
	Package     string           `json:"package"`
	Output      string           `json:"output"`
	TypeAliases []TypeAliasSpec  `json:"type_aliases"`
	ConstGroups []ConstGroupSpec `json:"const_groups"`
	ValueTypes  []ValueTypeSpec  `json:"value_types"`
}

type TypeAliasSpec struct {
	Name       string `json:"name"`
	Underlying string `json:"underlying"`
	Comment    string `json:"comment,omitempty"`
}

type ConstGroupSpec struct {
	Type    string      `json:"type"`
	Comment string      `json:"comment,omitempty"`
	Values  []ConstSpec `json:"values"`
}

type ConstSpec struct {
	Name    string `json:"name"`
	Expr    string `json:"expr"`
	Comment string `json:"comment,omitempty"`
}

type ValueTypeSpec struct {
	Name            string `json:"name"`
	Underlying      string `json:"underlying"`
	Comment         string `json:"comment,omitempty"`
	AccessorName    string `json:"accessor_name,omitempty"`
	AccessorType    string `json:"accessor_type,omitempty"`
	AccessorComment string `json:"accessor_comment,omitempty"`
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
	if manifest.Package == "" {
		return manifest, fmt.Errorf("manifest %q: package is required", path)
	}
	if manifest.Output == "" {
		return manifest, fmt.Errorf("manifest %q: output is required", path)
	}
	if len(manifest.TypeAliases) == 0 && len(manifest.ConstGroups) == 0 && len(manifest.ValueTypes) == 0 {
		return manifest, fmt.Errorf("manifest %q: at least one type_aliases, const_groups, or value_types entry is required", path)
	}
	for i, spec := range manifest.TypeAliases {
		if spec.Name == "" || spec.Underlying == "" {
			return manifest, fmt.Errorf("manifest %q: type_aliases[%d] is incomplete", path, i)
		}
	}
	for i, group := range manifest.ConstGroups {
		if group.Type == "" || len(group.Values) == 0 {
			return manifest, fmt.Errorf("manifest %q: const_groups[%d] is incomplete", path, i)
		}
		for j, value := range group.Values {
			if value.Name == "" || value.Expr == "" {
				return manifest, fmt.Errorf("manifest %q: const_groups[%d].values[%d] is incomplete", path, i, j)
			}
		}
	}
	for i, spec := range manifest.ValueTypes {
		if spec.Name == "" || spec.Underlying == "" {
			return manifest, fmt.Errorf("manifest %q: value_types[%d] is incomplete", path, i)
		}
		if (spec.AccessorName == "") != (spec.AccessorType == "") {
			return manifest, fmt.Errorf("manifest %q: value_types[%d] accessor_name/accessor_type must be set together", path, i)
		}
	}

	return manifest, nil
}
