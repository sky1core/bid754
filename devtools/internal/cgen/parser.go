package cgen

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var declPattern = regexp.MustCompile(`(?s)(?:^|[\s;])(?:const\s+)?([A-Za-z_][A-Za-z0-9_]*(?:\s+[A-Za-z_][A-Za-z0-9_]*)?)\s+([A-Za-z_][A-Za-z0-9_]*)\s*((?:\[[^\]]*\])+)\s*=\s*\{`)

type Table struct {
	Spec      TableSpec
	CType     string
	Dims      []int
	Value     Value
	SourceRel string
}

type Value struct {
	Elements []Value
	Number   *big.Int
}

func (v Value) IsScalar() bool {
	return v.Number != nil
}

func ParseTableFile(repoRoot string, spec TableSpec) (Table, error) {
	sourcePath := filepath.Join(repoRoot, spec.Source)
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return Table{}, fmt.Errorf("read source %q: %w", spec.Source, err)
	}

	clean := stripComments(string(data))
	if err := ensureUnconditionalDeclaration(clean, spec.Name); err != nil {
		return Table{}, fmt.Errorf("parse table %q: %w", spec.Name, err)
	}
	ctype, dims, init, err := extractDeclaration(clean, spec.Name)
	if err != nil {
		return Table{}, fmt.Errorf("parse table %q: %w", spec.Name, err)
	}

	value, err := parseInitializer(init)
	if err != nil {
		return Table{}, fmt.Errorf("parse initializer for %q: %w", spec.Name, err)
	}
	if arity := fixedWordArity(ctype); arity > 0 {
		value = unwrapFixedWordScalars(value, arity)
	}

	inferred, err := inferShape(value)
	if err != nil {
		return Table{}, fmt.Errorf("infer shape for %q: %w", spec.Name, err)
	}
	declShape := inferred
	if arity := fixedWordArity(ctype); arity > 0 {
		if len(inferred) == 0 || inferred[len(inferred)-1] != arity {
			return Table{}, fmt.Errorf("initializer for %q is not a %s word tuple", spec.Name, ctype)
		}
		declShape = inferred[:len(inferred)-1]
	} else if ctype == "DEC_DIGITS" {
		if len(inferred) == 0 || inferred[len(inferred)-1] != 4 {
			return Table{}, fmt.Errorf("initializer for %q is not a DEC_DIGITS field tuple", spec.Name)
		}
		declShape = inferred[:len(inferred)-1]
	}
	if len(dims) != len(declShape) {
		return Table{}, fmt.Errorf("declaration for %q has %d dims, initializer has %d", spec.Name, len(dims), len(declShape))
	}
	for i := range dims {
		if dims[i] == 0 {
			dims[i] = declShape[i]
			continue
		}
		if dims[i] != declShape[i] {
			return Table{}, fmt.Errorf("declaration for %q dimension %d says %d, initializer has %d", spec.Name, i, dims[i], declShape[i])
		}
	}

	if len(spec.ExpectedShape) > 0 && !equalInts(dims, spec.ExpectedShape) {
		return Table{}, fmt.Errorf("parse table %q: parsed shape %v does not match manifest expected_shape %v; a truncated or mis-parsed initializer is rejected here", spec.Name, dims, spec.ExpectedShape)
	}

	return Table{
		Spec:  spec,
		CType: ctype,
		Dims:  dims,
		Value: value,
		// Provenance shown in generated-file comments, expressed from the
		// repository root. spec.Source is read relative to the devtools module
		// (the generator's working directory); the checked-in comment points
		// readers at the same file from the repo root.
		SourceRel: filepath.Join("devtools", spec.Source),
	}, nil
}

func fixedWordArity(ctype string) int {
	switch ctype {
	case "BID_UINT128":
		return 2
	case "BID_UINT192":
		return 3
	case "BID_UINT256":
		return 4
	default:
		return 0
	}
}

func unwrapFixedWordScalars(v Value, arity int) Value {
	if v.IsScalar() {
		return v
	}
	normalized := make([]Value, len(v.Elements))
	for i, elem := range v.Elements {
		normalized[i] = unwrapFixedWordScalars(elem, arity)
	}
	v.Elements = normalized
	if len(v.Elements) == 1 && !v.Elements[0].IsScalar() && len(v.Elements[0].Elements) == arity &&
		allScalar(v.Elements[0].Elements) {
		return v.Elements[0]
	}
	return v
}

func allScalar(elements []Value) bool {
	for _, elem := range elements {
		if !elem.IsScalar() {
			return false
		}
	}
	return true
}

func extractDeclaration(content, name string) (string, []int, string, error) {
	matches := declPattern.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		declName := content[match[4]:match[5]]
		if declName != name {
			continue
		}
		ctype := strings.Join(strings.Fields(content[match[2]:match[3]]), " ")
		dims, err := parseDims(content[match[6]:match[7]])
		if err != nil {
			return "", nil, "", err
		}

		braceStart := match[1] - 1
		init, err := extractBraceBlock(content, braceStart)
		if err != nil {
			return "", nil, "", err
		}
		return ctype, dims, init, nil
	}
	return "", nil, "", fmt.Errorf("declaration not found")
}

// ensureUnconditionalDeclaration fails when the requested table is either
// declared more than once in the same source file or sits inside a preprocessor
// conditional block. extractDeclaration takes the first textual `= {` match and
// the tokenizer never sees `#if`/`#else`/`#endif`, so an upstream file that
// defines the same table twice under endian/config branches would be silently
// resolved to the first branch. This is the strict validation for that case: the
// currently pinned Intel sources declare every manifest table exactly once and
// outside any conditional, so a future conditional/duplicate definition must be
// resolved deliberately rather than guessed.
func ensureUnconditionalDeclaration(clean, name string) error {
	matches := declPattern.FindAllStringSubmatchIndex(clean, -1)
	var offsets []int
	for _, match := range matches {
		if clean[match[4]:match[5]] == name {
			offsets = append(offsets, match[0])
		}
	}
	if len(offsets) == 0 {
		// Absence is reported by extractDeclaration with its own message.
		return nil
	}
	if len(offsets) > 1 {
		return fmt.Errorf("declaration appears %d times in one source file; c-tablegen takes the first `= {` match and cannot tell endian/config branches apart, so a duplicated (e.g. #if/#else) definition must be resolved in the manifest, not guessed", len(offsets))
	}
	if depth := preprocessorConditionalDepthAt(clean, offsets[0]); depth != 0 {
		return fmt.Errorf("declaration sits inside a preprocessor conditional (#if/#ifdef/#ifndef ... #else/#elif/#endif) block; c-tablegen does not evaluate the preprocessor and could silently take a wrong endian/config branch")
	}
	return nil
}

// preprocessorConditionalDepthAt reports the net `#if`-nesting depth at byte
// offset within comment-stripped C source. A positive depth means the offset is
// checked by an unclosed `#if`/`#ifdef`/`#ifndef`. `#else`/`#elif` do not change
// the depth; `#endif` closes one level.
func preprocessorConditionalDepthAt(clean string, offset int) int {
	if offset > len(clean) {
		offset = len(clean)
	}
	depth := 0
	for pos := 0; pos < offset; {
		nl := strings.IndexByte(clean[pos:], '\n')
		lineEnd := len(clean)
		if nl >= 0 {
			lineEnd = pos + nl + 1
		}
		trimmed := strings.TrimLeft(clean[pos:min(lineEnd, len(clean))], " \t")
		if strings.HasPrefix(trimmed, "#") {
			directive := strings.TrimSpace(trimmed[1:])
			switch {
			case strings.HasPrefix(directive, "endif"):
				if depth > 0 {
					depth--
				}
			case strings.HasPrefix(directive, "if"): // if / ifdef / ifndef
				depth++
			}
		}
		if nl < 0 {
			break
		}
		pos = lineEnd
	}
	return depth
}

func parseDims(raw string) ([]int, error) {
	matches := regexp.MustCompile(`\[([^\]]*)\]`).FindAllStringSubmatch(raw, -1)
	dims := make([]int, 0, len(matches))
	for _, match := range matches {
		part := strings.TrimSpace(match[1])
		if part == "" {
			dims = append(dims, 0)
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("unsupported dimension %q", part)
		}
		dims = append(dims, n)
	}
	if len(dims) == 0 {
		return nil, fmt.Errorf("no dimensions found in %q", raw)
	}
	return dims, nil
}

func extractBraceBlock(content string, start int) (string, error) {
	if start < 0 || start >= len(content) || content[start] != '{' {
		return "", fmt.Errorf("invalid brace start")
	}
	depth := 0
	for i := start; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return content[start : i+1], nil
			}
		}
	}
	return "", fmt.Errorf("unterminated initializer")
}

func stripComments(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); {
		if i+1 < len(s) && s[i] == '/' && s[i+1] == '*' {
			i += 2
			for i+1 < len(s) && !(s[i] == '*' && s[i+1] == '/') {
				if s[i] == '\n' {
					b.WriteByte('\n')
				}
				i++
			}
			if i+1 < len(s) {
				i += 2
			}
			continue
		}
		if i+1 < len(s) && s[i] == '/' && s[i+1] == '/' {
			i += 2
			for i < len(s) && s[i] != '\n' {
				i++
			}
			continue
		}
		b.WriteByte(s[i])
		i++
	}

	return b.String()
}

type parser struct {
	tokens []token
	pos    int
}

type tokenType int

const (
	tokenLBrace tokenType = iota
	tokenRBrace
	tokenComma
	tokenPlus
	tokenMinus
	tokenTilde
	tokenShl
	tokenLParen
	tokenRParen
	tokenNumber
	tokenString
)

type token struct {
	typ tokenType
	raw string
}

func parseInitializer(init string) (Value, error) {
	tokens, err := tokenize(init)
	if err != nil {
		return Value{}, err
	}
	p := &parser{tokens: tokens}
	value, err := p.parseValue()
	if err != nil {
		return Value{}, err
	}
	if p.pos != len(p.tokens) {
		return Value{}, fmt.Errorf("unexpected trailing tokens")
	}
	return value, nil
}

func tokenize(input string) ([]token, error) {
	tokens := make([]token, 0, len(input)/2)
	for i := 0; i < len(input); {
		switch ch := input[i]; {
		case unicode.IsSpace(rune(ch)):
			i++
		case ch == '{':
			tokens = append(tokens, token{typ: tokenLBrace, raw: "{"})
			i++
		case ch == '}':
			tokens = append(tokens, token{typ: tokenRBrace, raw: "}"})
			i++
		case ch == ',':
			tokens = append(tokens, token{typ: tokenComma, raw: ","})
			i++
		case ch == '+':
			tokens = append(tokens, token{typ: tokenPlus, raw: "+"})
			i++
		case ch == '-':
			tokens = append(tokens, token{typ: tokenMinus, raw: "-"})
			i++
		case ch == '~':
			tokens = append(tokens, token{typ: tokenTilde, raw: "~"})
			i++
		case ch == '(':
			tokens = append(tokens, token{typ: tokenLParen, raw: "("})
			i++
		case ch == ')':
			tokens = append(tokens, token{typ: tokenRParen, raw: ")"})
			i++
		case ch == '<' && i+1 < len(input) && input[i+1] == '<':
			tokens = append(tokens, token{typ: tokenShl, raw: "<<"})
			i += 2
		case ch == '"':
			raw, next, err := readCStringLiteral(input, i)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token{typ: tokenString, raw: raw})
			i = next
		case isNumberStart(input[i:]):
			j := i + 1
			for j < len(input) && isNumberPart(input[j]) {
				j++
			}
			tokens = append(tokens, token{typ: tokenNumber, raw: input[i:j]})
			i = j
		default:
			return nil, fmt.Errorf("unsupported token near %q", input[i:min(i+16, len(input))])
		}
	}
	return tokens, nil
}

func readCStringLiteral(input string, start int) (string, int, error) {
	var b strings.Builder
	for i := start + 1; i < len(input); i++ {
		switch input[i] {
		case '\\':
			if i+1 >= len(input) {
				return "", 0, fmt.Errorf("unterminated escape sequence")
			}
			i++
			switch input[i] {
			case '\\', '"':
				b.WriteByte(input[i])
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case '0':
				b.WriteByte(0)
			default:
				return "", 0, fmt.Errorf("unsupported escape \\%c", input[i])
			}
		case '"':
			return b.String(), i + 1, nil
		default:
			b.WriteByte(input[i])
		}
	}
	return "", 0, fmt.Errorf("unterminated string literal")
}

func isNumberStart(s string) bool {
	if s == "" {
		return false
	}
	ch := s[0]
	return ch >= '0' && ch <= '9'
}

func isNumberPart(ch byte) bool {
	return (ch >= '0' && ch <= '9') ||
		(ch >= 'a' && ch <= 'f') ||
		(ch >= 'A' && ch <= 'F') ||
		ch == 'x' || ch == 'X' || ch == 'u' || ch == 'U' || ch == 'l' || ch == 'L'
}

func (p *parser) parseValue() (Value, error) {
	if p.peek(tokenLBrace) {
		p.pos++
		elements := []Value{}
		for !p.peek(tokenRBrace) {
			item, err := p.parseValue()
			if err != nil {
				return Value{}, err
			}
			elements = append(elements, item)
			if p.peek(tokenComma) {
				p.pos++
				if p.peek(tokenRBrace) {
					break
				}
			} else if !p.peek(tokenRBrace) {
				return Value{}, fmt.Errorf("expected comma or closing brace")
			}
		}
		if !p.peek(tokenRBrace) {
			return Value{}, fmt.Errorf("missing closing brace")
		}
		p.pos++
		return Value{Elements: elements}, nil
	}

	if p.peek(tokenString) {
		raw := p.tokens[p.pos].raw
		p.pos++
		elements := make([]Value, len(raw))
		for i := range raw {
			elements[i] = Value{Number: big.NewInt(int64(raw[i]))}
		}
		return Value{Elements: elements}, nil
	}

	number, err := p.parseExpr()
	if err != nil {
		return Value{}, err
	}
	return Value{Number: number}, nil
}

func (p *parser) parseExpr() (*big.Int, error) {
	left, err := p.parseShift()
	if err != nil {
		return nil, err
	}
	for p.peek(tokenPlus) || p.peek(tokenMinus) {
		op := p.tokens[p.pos].typ
		p.pos++
		right, err := p.parseShift()
		if err != nil {
			return nil, err
		}
		if op == tokenPlus {
			left = new(big.Int).Add(left, right)
		} else {
			left = new(big.Int).Sub(left, right)
		}
	}
	return left, nil
}

func (p *parser) parseShift() (*big.Int, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for p.peek(tokenShl) {
		p.pos++
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		left = new(big.Int).Lsh(left, uint(right.Int64()))
	}
	return left, nil
}

func (p *parser) parseTerm() (*big.Int, error) {
	sign := 1
	if p.peek(tokenPlus) {
		p.pos++
	} else if p.peek(tokenMinus) {
		sign = -1
		p.pos++
	}
	if p.peek(tokenTilde) {
		p.pos++
		inner, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		// Bitwise NOT for 64-bit unsigned: ^x == 0xFFFFFFFFFFFFFFFF - x
		mask := new(big.Int).SetUint64(^uint64(0))
		inner = new(big.Int).Xor(mask, inner)
		if sign < 0 {
			inner.Neg(inner)
		}
		return inner, nil
	}
	if p.peek(tokenLParen) {
		p.pos++
		inner, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if !p.peek(tokenRParen) {
			return nil, fmt.Errorf("expected closing parenthesis")
		}
		p.pos++
		if sign < 0 {
			inner.Neg(inner)
		}
		return inner, nil
	}
	if !p.peek(tokenNumber) {
		return nil, fmt.Errorf("expected numeric literal")
	}
	value, err := parseNumericLiteral(p.tokens[p.pos].raw)
	if err != nil {
		return nil, err
	}
	p.pos++
	if sign < 0 {
		value.Neg(value)
	}
	return value, nil
}

func parseNumericLiteral(raw string) (*big.Int, error) {
	trimmed := strings.TrimRight(raw, "uUlL")
	n := new(big.Int)
	base := 10
	if strings.HasPrefix(trimmed, "0x") || strings.HasPrefix(trimmed, "0X") {
		base = 16
		trimmed = trimmed[2:]
	}
	if _, ok := n.SetString(trimmed, base); !ok {
		return nil, fmt.Errorf("invalid numeric literal %q", raw)
	}
	return n, nil
}

func (p *parser) peek(tt tokenType) bool {
	return p.pos < len(p.tokens) && p.tokens[p.pos].typ == tt
}

func inferShape(v Value) ([]int, error) {
	if v.IsScalar() {
		return nil, nil
	}
	shape := []int{len(v.Elements)}
	if len(v.Elements) == 0 {
		return shape, nil
	}
	childShape, err := inferShape(v.Elements[0])
	if err != nil {
		return nil, err
	}
	for i := 1; i < len(v.Elements); i++ {
		nextShape, err := inferShape(v.Elements[i])
		if err != nil {
			return nil, err
		}
		if !equalInts(childShape, nextShape) {
			return nil, fmt.Errorf("ragged initializer")
		}
	}
	return append(shape, childShape...), nil
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
