package publicroute

import (
	"fmt"
	"strings"
	"testing"
)

func TestVariantEquivalenceArgumentFlow(t *testing.T) {
	for _, layout := range []string{"shared_delegate", "flagless_calls_flagged", "flagged_calls_flagless"} {
		t.Run(layout, func(t *testing.T) {
			callee := "delegate"
			callArgs := "x, y, rndMode, false"
			binding := "v, _"
			result := "v"
			wrapper := "Bid64Add"
			if layout == "flagless_calls_flagged" {
				callee = "Bid64AddWithFlags"
				callArgs = "x, y, rndMode"
			}
			if layout == "flagged_calls_flagless" {
				callee = "Bid64Add"
				callArgs = "x, y, rndMode"
				binding = "v"
				result = "v, uint32(0)"
				wrapper = "Bid64AddWithFlags"
			}
			boundBody := func(args string) string {
				return fmt.Sprintf("%s := %s(%s); return %s", binding, callee, args, result)
			}
			type argumentCase struct {
				name    string
				body    string
				want    bool
				refusal string
			}
			cases := []argumentCase{
				{"renamed_parameters", boundBody(callArgs), true, ""},
				{"parenthesized_parameters", boundBody(strings.ReplaceAll(callArgs, "x, y", "(x), (y)")), true, ""},
				{"var_binding", "var " + strings.Replace(boundBody(callArgs), ":=", "=", 1), true, ""},
				{"swapped_operands", boundBody(strings.ReplaceAll(callArgs, "x, y", "y, x")), false, "arguments"},
				{"duplicated_operand", boundBody(strings.ReplaceAll(callArgs, "x, y", "y, y")), false, "arguments"},
				{"constant_rounding_mode", boundBody(strings.ReplaceAll(callArgs, "rndMode", "0")), false, "arguments"},
				{"changed_rounding_expression", boundBody(strings.ReplaceAll(callArgs, "rndMode", "rndMode + 1")), false, "wrapper"},
				{"argument_call", boundBody(strings.ReplaceAll(callArgs, "x, y", "bump(&x), y")), false, "wrapper"},
				{"global_argument", boundBody(strings.ReplaceAll(callArgs, "x, y", "state, y")), false, "wrapper"},
				{"parameter_assignment", "x = y; " + boundBody(callArgs), false, "wrapper"},
				{"parameter_compound_assignment", "x += y; " + boundBody(callArgs), false, "wrapper"},
				{"parameter_increment", "x++; " + boundBody(callArgs), false, "wrapper"},
				{"parameter_shadow", "{ x := y; " + boundBody(callArgs) + " }", false, "wrapper"},
				{"parameter_address_escape", "bump(&x); " + boundBody(callArgs), false, "wrapper"},
				{"deferred_mutation", "defer bump(&x); " + boundBody(callArgs), false, "wrapper"},
				{"local_alias", "a := x; " + boundBody(strings.ReplaceAll(callArgs, "x, y", "a, y")), false, "wrapper"},
				{"multiple_delegate_calls", fmt.Sprintf("%s = %s(%s); ", strings.ReplaceAll(binding, "v", "_"), callee, callArgs) + boundBody(callArgs), false, "wrapper"},
			}
			if layout == "shared_delegate" {
				cases = append(cases, argumentCase{"changed_boolean_constant", boundBody(strings.ReplaceAll(callArgs, "false", "true")), false, "arguments"})
			}
			if layout == "flagged_calls_flagless" {
				cases = append(cases, argumentCase{"status_side_effect", strings.ReplaceAll(boundBody(callArgs), "uint32(0)", "uint32(bump(&x))"), false, "wrapper"})
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					src := `package corpus
var state uint64
var bump = func(p *uint64) uint64 { *p++; return *p }
func delegate(x, y uint64, mode int, check bool) (uint64, uint32) {
	if check { return y, 1 }
	return x - y + uint64(mode), 0
}

`
					switch layout {
					case "shared_delegate":
						src += "func Bid64Add(x, y uint64, rndMode int) uint64 { " + tc.body + " }\n"
						src += "func Bid64AddWithFlags(a, b uint64, mode int) (uint64, uint32) { return delegate(a, b, mode, false) }\n"
					case "flagless_calls_flagged":
						src += "func Bid64Add(x, y uint64, rndMode int) uint64 { " + tc.body + " }\n"
						src += "func Bid64AddWithFlags(a, b uint64, mode int) (uint64, uint32) { return a - b + uint64(mode), 0 }\n"
					case "flagged_calls_flagless":
						src += "func Bid64Add(a, b uint64, mode int) uint64 { return a - b + uint64(mode) }\n"
						src += "func Bid64AddWithFlags(x, y uint64, rndMode int) (uint64, uint32) { " + tc.body + " }\n"
					}
					env := newCorpusEnv(t)
					pkg, info, file := env.typecheck(tc.name, src)
					r := env.newResolver(pkg, info, file)
					decl := r.funcByObj[r.lookupFunc(wrapper)]
					calleeObj := r.lookupFunc(callee)
					callees := r.bidgoCalleeFuncs(decl)
					if len(callees) != 1 || !callees[calleeObj] || !r.firstResultFlowsFromCall(decl, calleeObj) {
						t.Fatal("fixture must retain sole-callee identity and first-result flow; refusal must come from the argument proof")
					}
					_, extracted := r.variantCallArguments(decl, calleeObj)
					if tc.refusal == "arguments" && !extracted {
						t.Fatal("mismatch must reach argument comparison, not fail wrapper extraction")
					}
					if tc.refusal == "wrapper" && extracted {
						t.Fatal("unsafe or unsupported wrapper unexpectedly has provable argument flow")
					}
					for _, fn := range []string{"Bid64Add", "Bid64AddWithFlags"} {
						how, got := r.proveVariantEquivalence(fn, func(name string) bool { return name != fn })
						if got != tc.want || (got && how == "") || (!got && how != "") {
							t.Errorf("proveVariantEquivalence(%q) = (%q, %v), want proof=%v", fn, how, got, tc.want)
						}
					}
				})
			}
		})
	}
}

func TestVariantEquivalenceTypedArguments(t *testing.T) {
	for _, tc := range []struct {
		name     string
		params   string
		flagless string
		flagged  string
		want     bool
	}{
		{"equal_boolean_constants", "a, b uint64, mode int", "x, y, rndMode, false", "a, b, mode, off", true},
		{"different_boolean_constants", "a, b uint64, mode int", "x, y, rndMode, false", "a, b, mode, true", false},
		{"equal_folded_constants", "a, b uint64, mode int", "x, y, rndMode, 1 << 1", "a, b, mode, 2", true},
		{"different_constant_values", "a, b uint64, mode int", "x, y, rndMode, 1", "a, b, mode, 2", false},
		{"different_constant_types", "a, b uint64, mode int", "x, y, rndMode, int32(1)", "a, b, mode, int64(1)", false},
		{"same_spelling_different_positions", "y, x uint64, rndMode int", "x, y, rndMode, false", "x, y, rndMode, false", false},
		{"same_parameter_named_as_constant", "a, b uint64, off int", "x, y, rndMode, false", "a, b, off, false", true},
		{"constant_shadowed_by_parameter", "a, b uint64, fixed int", "x, y, fixed, false", "a, b, fixed, false", false},
		{"different_parameter_types", "a, b uint64, mode int32", "x, y, rndMode, false", "a, b, mode, false", false},
		{"extra_parameter", "a, b uint64, mode, extra int", "x, y, rndMode, false", "a, b, mode, false", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := fmt.Sprintf(`package corpus
const off = false
const fixed = 0
func delegate(x, y uint64, mode, option any) (uint64, uint32) { return x - y, 0 }
func Bid64Add(x, y uint64, rndMode int) uint64 {
	v, _ := delegate(%s)
	return v
}
func Bid64AddWithFlags(%s) (uint64, uint32) {
	v, flags := delegate(%s)
	return v, flags
}
`, tc.flagless, tc.params, tc.flagged)
			env := newCorpusEnv(t)
			pkg, info, file := env.typecheck(tc.name, src)
			r := env.newResolver(pkg, info, file)
			callee := r.lookupFunc("delegate")
			for _, fn := range []string{"Bid64Add", "Bid64AddWithFlags"} {
				decl := r.funcByObj[r.lookupFunc(fn)]
				if !r.firstResultFlowsFromCall(decl, callee) {
					t.Fatal("fixture must retain first-result flow")
				}
				if _, ok := r.variantCallArguments(decl, callee); !ok {
					t.Fatal("fixture must reach typed argument or signature comparison")
				}
				if how, got := r.proveVariantEquivalence(fn, func(name string) bool { return name != fn }); got != tc.want {
					t.Errorf("proveVariantEquivalence(%q) = (%q, %v), want proof=%v", fn, how, got, tc.want)
				}
			}
		})
	}
}
