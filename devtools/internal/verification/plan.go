package verification

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
)

type Plan struct {
	Version  int                `json:"version"`
	Profiles map[string]Profile `json:"profiles"`
	Gates    []Gate             `json:"gates"`
	Matrix   []RequiredRun      `json:"matrix,omitempty"`
}

type RequiredRun struct {
	Profile  string `json:"profile"`
	Platform string `json:"platform"`
}

type Profile struct {
	Groups    []string `json:"groups"`
	Scope     string   `json:"scope"`
	Platforms []string `json:"platforms"`
}

type Gate struct {
	ID            string   `json:"id"`
	Target        string   `json:"target"`
	Group         string   `json:"group"`
	Tools         []string `json:"tools"`
	Prerequisites []string `json:"prerequisites"`
	Comparison    string   `json:"comparison"`
	Evidence      Evidence `json:"evidence"`
}

type Evidence struct {
	Domain         string   `json:"domain,omitempty"`
	Passes         []string `json:"passes,omitempty"`
	Patterns       []string `json:"patterns,omitempty"`
	Codec          bool     `json:"codec,omitempty"`
	ParserResource bool     `json:"parser_resource,omitempty"`
}

var identifier = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
var targetName = regexp.MustCompile(`^[_a-z][a-z0-9_-]*$`)

func ReadJSON(path string, value any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(value); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("%s: trailing JSON data", path)
	}
	return nil
}

func Load(path string) (Plan, error) {
	var p Plan
	if err := ReadJSON(path, &p); err != nil {
		return p, err
	}
	if p.Version != 1 || len(p.Profiles) == 0 || len(p.Gates) == 0 {
		return p, fmt.Errorf("invalid or empty execution plan")
	}
	ids, groups := map[string]bool{}, map[string]bool{}
	for _, g := range p.Gates {
		if !identifier.MatchString(g.ID) || !identifier.MatchString(g.Group) || !targetName.MatchString(g.Target) || ids[g.ID] {
			return p, fmt.Errorf("invalid or duplicate gate %q", g.ID)
		}
		ids[g.ID], groups[g.Group] = true, true
		if g.Comparison == "" || (g.Evidence.Domain == "" && len(g.Evidence.Passes) == 0 && len(g.Evidence.Patterns) == 0 && !g.Evidence.Codec && !g.Evidence.ParserResource) {
			return p, fmt.Errorf("gate %s lacks comparison or execution evidence", g.ID)
		}
		for _, pattern := range g.Evidence.Patterns {
			if _, err := regexp.Compile(pattern); err != nil {
				return p, fmt.Errorf("gate %s: %w", g.ID, err)
			}
		}
	}
	for name, profile := range p.Profiles {
		if !identifier.MatchString(name) || len(profile.Groups) == 0 || len(profile.Platforms) == 0 || (profile.Scope != "full" && profile.Scope != "profile" && profile.Scope != "partial") {
			return p, fmt.Errorf("invalid profile %q", name)
		}
		seen := map[string]bool{}
		for _, group := range profile.Groups {
			if !groups[group] || seen[group] {
				return p, fmt.Errorf("profile %s: invalid or duplicate group %s", name, group)
			}
			seen[group] = true
		}
	}
	seenRuns := map[string]bool{}
	for _, required := range p.Matrix {
		key := required.Profile + "@" + required.Platform
		if seenRuns[key] {
			return p, fmt.Errorf("duplicate matrix requirement %s", key)
		}
		seenRuns[key] = true
		if _, err := p.Select(required.Profile, required.Platform); err != nil {
			return p, err
		}
	}
	return p, nil
}

func Hash(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

func (p Plan) Hash() string { raw, _ := json.Marshal(p); return Hash(raw) }

func (p Plan) Select(name, platform string) ([]Gate, error) {
	profile, ok := p.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("unknown verification profile %q", name)
	}
	allowed := false
	for _, value := range profile.Platforms {
		if value == platform {
			allowed = true
		}
	}
	if !allowed {
		return nil, fmt.Errorf("profile %s does not apply to %s", name, platform)
	}
	var gates []Gate
	for _, group := range profile.Groups {
		for _, gate := range p.Gates {
			if gate.Group == group {
				gates = append(gates, gate)
			}
		}
	}
	return gates, nil
}
