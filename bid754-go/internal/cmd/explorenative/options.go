package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/sky1core/bid754/bid754-go/internal/decimalprobe"
)

func validateFlagArgs(args []string, names map[string]bool) error {
	seen := map[string]bool{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			return fmt.Errorf("unexpected argument %q", arg)
		}
		name, _, inline := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		takesValue, ok := names[name]
		if !ok || seen[name] {
			return fmt.Errorf("unknown or duplicate flag %q", name)
		}
		seen[name] = true
		if takesValue && !inline {
			i++
			if i >= len(args) {
				return fmt.Errorf("missing value for -%s", name)
			}
		}
	}
	return nil
}

func strictCSV(text string) ([]string, error) {
	parts := strings.Split(text, ",")
	seen := map[string]bool{}
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			return nil, fmt.Errorf("empty or duplicate CSV entry %q", p)
		}
		parts[i] = p
		seen[p] = true
	}
	return parts, nil
}

func validateOptions(campaign string, cases int, bias float64, ops, widths, modes string, shrink int) error {
	if err := validateCampaign(campaign); err != nil {
		return err
	}
	if cases <= 0 || shrink < 0 || math.IsNaN(bias) || bias < 0 || bias > 1 {
		return fmt.Errorf("cases must be positive, shrink-attempts nonnegative, bias in [0,1]")
	}
	selected, err := strictCSV(ops)
	if err != nil {
		return err
	}
	for _, op := range selected {
		if !supportedReferenceOp(op) && !(campaign == campaignLegacy && op == "sqrt") {
			return fmt.Errorf("unsupported op %q for %s", op, campaign)
		}
	}
	ws, err := strictCSV(widths)
	if err != nil {
		return err
	}
	for _, w := range ws {
		if w != "32" && w != "64" && w != "128" {
			return fmt.Errorf("unknown width %q", w)
		}
	}
	ms, err := strictCSV(modes)
	if err != nil {
		return err
	}
	for _, m := range ms {
		if !knownMode(m) {
			return fmt.Errorf("unknown mode %q", m)
		}
	}
	return nil
}

func referenceTargets(campaign, text string) ([]string, []string, error) {
	ops, err := strictCSV(text)
	if err != nil {
		return nil, nil, err
	}
	selected := map[string]bool{}
	for _, op := range ops {
		if !supportedReferenceOp(op) {
			return nil, nil, fmt.Errorf("unsupported reference op %q", op)
		}
		selected[op] = true
	}
	var families []string
	if campaign == campaignRel {
		for _, fam := range decimalprobe.Families() {
			op, err := decimalprobe.FamilyOperation(fam)
			if err != nil {
				return nil, nil, err
			}
			if selected[op] {
				families = append(families, fam)
			}
		}
		if len(families) == 0 {
			return nil, nil, fmt.Errorf("no relation families selected")
		}
	}
	return ops, families, nil
}

func validateSeed(text string) error {
	_, err := strconv.ParseUint(text, 0, 64)
	if err != nil {
		return fmt.Errorf("invalid seed %q: %w", text, err)
	}
	return nil
}
