package verification

import (
	"path/filepath"
	"strings"
	"testing"
)

func repoPlan(t *testing.T) Plan {
	t.Helper()
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := Load(filepath.Join(repo, "devtools/verification_plan.json"))
	if err != nil {
		t.Fatalf("load plan: %v", err)
	}
	return plan
}

func selectedGates(t *testing.T, plan Plan, profile, platform string) map[string]Gate {
	t.Helper()
	gates, err := plan.Select(profile, platform)
	if err != nil {
		t.Fatalf("select %s/%s: %v", profile, platform, err)
	}
	byID := map[string]Gate{}
	for _, g := range gates {
		byID[g.ID] = g
	}
	return byID
}

func hasPattern(g Gate, substr string) bool {
	for _, p := range g.Evidence.Patterns {
		if strings.Contains(p, substr) {
			return true
		}
	}
	return false
}

func TestTier1BigDecimalRequiredAcrossPortableProfiles(t *testing.T) {
	plan := repoPlan(t)
	const deterministic = "tier1-bigdecimal"
	fuzz := []string{"bigdecimal-fuzz", "tier1-bigdecimal-fuzz"}

	for _, profile := range []string{"ci-portable", "full", "portable"} {
		p, ok := plan.Profiles[profile]
		if !ok {
			t.Fatalf("profile %s missing", profile)
		}
		for _, platform := range p.Platforms {
			gates := selectedGates(t, plan, profile, platform)
			g, ok := gates[deterministic]
			if !ok {
				t.Errorf("%s/%s: deterministic %s gate not selected", profile, platform, deterministic)
			} else if !hasPattern(g, "TIER1-BIGDECIMAL") || len(g.Evidence.Passes) == 0 {
				t.Errorf("%s/%s: %s gate lost its deterministic evidence", profile, platform, deterministic)
			}
			for _, f := range fuzz {
				if _, present := gates[f]; present {
					t.Errorf("%s/%s: long fuzz gate %s must stay optional, not required", profile, platform, f)
				}
			}
		}
	}
}

func TestNumericProfileKeepsDeterministicAndFuzz(t *testing.T) {
	plan := repoPlan(t)
	p, ok := plan.Profiles["numeric"]
	if !ok {
		t.Fatal("numeric profile missing")
	}
	for _, platform := range p.Platforms {
		gates := selectedGates(t, plan, "numeric", platform)
		if _, ok := gates["tier1-bigdecimal"]; !ok {
			t.Errorf("numeric/%s: tier1-bigdecimal not selected", platform)
		}
		for _, f := range []string{"bigdecimal-fuzz", "tier1-bigdecimal-fuzz"} {
			g, ok := gates[f]
			if !ok {
				t.Errorf("numeric/%s: fuzz gate %s not selected", platform, f)
			} else if g.Evidence.GoFuzz == "" {
				t.Errorf("numeric/%s: %s is not a Go-fuzz discovery gate", platform, f)
			}
		}
	}
}

func TestTier1DecnumberRequiredAcrossNativeProfiles(t *testing.T) {
	plan := repoPlan(t)
	for _, profile := range []string{"native", "full"} {
		for _, platform := range plan.Profiles[profile].Platforms {
			gates := selectedGates(t, plan, profile, platform)
			gate, ok := gates["tier1-reference-decnumber"]
			if !ok || !hasPattern(gate, "TIER1-DECNUMBER width=32") || !hasPattern(gate, "TIER1-DECNUMBER width=64") || !hasPattern(gate, "TIER1-DECNUMBER width=128") {
				t.Errorf("%s/%s: missing three-width Tier1 decNumber calibration evidence", profile, platform)
			}
		}
	}
}
