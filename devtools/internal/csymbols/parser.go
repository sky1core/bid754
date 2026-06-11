package csymbols

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

var (
	functionDeclPattern = regexp.MustCompile(`^extern\s+(.+?)\s+([A-Za-z_][A-Za-z0-9_]*)\s*\((.*)\)$`)
	identPattern        = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

type Generated struct {
	JSON []byte
}

type SymbolFile struct {
	Library string   `json:"library"`
	Headers []string `json:"headers"`
	Symbols []Symbol `json:"symbols"`
}

type Symbol struct {
	Name        string   `json:"name"`
	LinkName    string   `json:"link_name"`
	Header      string   `json:"header"`
	ReturnType  string   `json:"return_type"`
	Parameters  []string `json:"parameters"`
	Declaration string   `json:"declaration"`
}

type extractor struct {
	macros  map[string]string
	aliases map[string]string
}

func Generate(repoRoot string, manifest Manifest) (Generated, error) {
	e := extractor{
		macros:  map[string]string{},
		aliases: map[string]string{},
	}

	var headers []string
	var symbols []Symbol
	for _, header := range manifest.Headers {
		source, err := e.processHeader(repoRoot, header)
		if err != nil {
			return Generated{}, err
		}
		if !header.ExtractSymbols {
			continue
		}
		headers = append(headers, header.Path)
		parsed, err := e.extractSymbols(header.Path, source)
		if err != nil {
			return Generated{}, err
		}
		symbols = append(symbols, parsed...)
	}

	sort.Slice(symbols, func(i, j int) bool {
		return symbols[i].Name < symbols[j].Name
	})

	payload := SymbolFile{
		Library: manifest.Library,
		Headers: headers,
		Symbols: symbols,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return Generated{}, fmt.Errorf("marshal symbols json: %w", err)
	}
	data = append(data, '\n')
	return Generated{JSON: data}, nil
}

func WriteOutputs(repoRoot string, manifest Manifest, generated Generated) error {
	path := filepath.Join(repoRoot, manifest.Output)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output dir for %q: %w", path, err)
	}
	if err := os.WriteFile(path, generated.JSON, 0o644); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}
	return nil
}

func (e *extractor) processHeader(repoRoot string, spec HeaderSpec) (string, error) {
	sourcePath := filepath.Join(repoRoot, spec.Path)
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", fmt.Errorf("read header %q: %w", spec.Path, err)
	}

	content := stripComments(string(data))
	lines := strings.Split(content, "\n")
	state := newConditionalState()
	var active []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			if err := e.handleDirective(trimmed, state); err != nil {
				return "", fmt.Errorf("process directives in %q: %w", spec.Path, err)
			}
			continue
		}
		if !state.active() {
			continue
		}
		if spec.ExtractSymbols {
			active = append(active, line)
		}
	}
	return strings.Join(active, "\n"), nil
}

func (e *extractor) handleDirective(line string, state *conditionalState) error {
	body := strings.TrimSpace(strings.TrimPrefix(line, "#"))
	switch {
	case strings.HasPrefix(body, "if "):
		expr := strings.TrimSpace(strings.TrimPrefix(body, "if"))
		state.push(evalExpr(expr, e.macros))
	case strings.HasPrefix(body, "ifdef "):
		name := strings.TrimSpace(strings.TrimPrefix(body, "ifdef"))
		_, ok := e.macros[name]
		state.push(ok)
	case strings.HasPrefix(body, "ifndef "):
		name := strings.TrimSpace(strings.TrimPrefix(body, "ifndef"))
		_, ok := e.macros[name]
		state.push(!ok)
	case strings.HasPrefix(body, "elif "):
		expr := strings.TrimSpace(strings.TrimPrefix(body, "elif"))
		state.elif(evalExpr(expr, e.macros))
	case body == "else":
		state.elseBranch()
	case body == "endif":
		state.pop()
	case strings.HasPrefix(body, "define "):
		if !state.active() {
			return nil
		}
		e.handleDefine(strings.TrimSpace(strings.TrimPrefix(body, "define")))
	case strings.HasPrefix(body, "undef "):
		if !state.active() {
			return nil
		}
		name := strings.TrimSpace(strings.TrimPrefix(body, "undef"))
		delete(e.macros, name)
		delete(e.aliases, name)
	default:
	}
	return nil
}

func (e *extractor) handleDefine(body string) {
	if body == "" {
		return
	}
	parts := strings.Fields(body)
	if len(parts) == 0 {
		return
	}
	name := parts[0]
	if strings.Contains(name, "(") {
		return
	}
	value := strings.TrimSpace(strings.TrimPrefix(body, name))
	if strings.HasPrefix(name, "bid") && identPattern.MatchString(value) && strings.HasPrefix(value, "__") {
		e.aliases[name] = value
		return
	}
	e.macros[name] = value
}

func (e *extractor) extractSymbols(headerPath, source string) ([]Symbol, error) {
	stmts := splitStatements(source)
	symbols := make([]Symbol, 0, len(stmts))
	seen := map[string]bool{}
	for _, stmt := range stmts {
		normalized := normalizeWhitespace(expandMacros(stmt, e.macros))
		if !strings.HasPrefix(normalized, "extern ") || !strings.Contains(normalized, "(") {
			continue
		}
		matches := functionDeclPattern.FindStringSubmatch(normalized)
		if matches == nil {
			continue
		}
		name := matches[2]
		params := splitParams(matches[3])
		symbol := Symbol{
			Name:        name,
			LinkName:    e.linkName(name),
			Header:      headerPath,
			ReturnType:  strings.TrimSpace(matches[1]),
			Parameters:  params,
			Declaration: fmt.Sprintf("%s %s(%s)", strings.TrimSpace(matches[1]), name, strings.Join(params, ", ")),
		}
		key := symbol.Name + "\x00" + symbol.Declaration
		if seen[key] {
			continue
		}
		seen[key] = true
		symbols = append(symbols, symbol)
	}
	if len(symbols) == 0 {
		return nil, fmt.Errorf("no symbols extracted from %q", headerPath)
	}
	return symbols, nil
}

func (e *extractor) linkName(name string) string {
	if alias, ok := e.aliases[name]; ok {
		return alias
	}
	return name
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

func splitStatements(source string) []string {
	var out []string
	var current strings.Builder
	depthParen := 0
	for _, r := range source {
		switch r {
		case '(':
			depthParen++
		case ')':
			if depthParen > 0 {
				depthParen--
			}
		case ';':
			if depthParen == 0 {
				stmt := strings.TrimSpace(current.String())
				if stmt != "" {
					out = append(out, stmt)
				}
				current.Reset()
				continue
			}
		}
		current.WriteRune(r)
	}
	stmt := strings.TrimSpace(current.String())
	if stmt != "" {
		out = append(out, stmt)
	}
	return out
}

func splitParams(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "void" {
		return nil
	}
	var params []string
	start := 0
	depth := 0
	for i, r := range raw {
		switch r {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				params = append(params, normalizeWhitespace(raw[start:i]))
				start = i + 1
			}
		}
	}
	params = append(params, normalizeWhitespace(raw[start:]))
	return params
}

func expandMacros(s string, macros map[string]string) string {
	const maxPasses = 8
	for pass := 0; pass < maxPasses; pass++ {
		var b strings.Builder
		changed := false
		for i := 0; i < len(s); {
			if isIdentStart(rune(s[i])) {
				j := i + 1
				for j < len(s) && isIdentPart(rune(s[j])) {
					j++
				}
				name := s[i:j]
				if replacement, ok := macros[name]; ok {
					b.WriteString(replacement)
					changed = true
				} else {
					b.WriteString(name)
				}
				i = j
				continue
			}
			b.WriteByte(s[i])
			i++
		}
		s = b.String()
		if !changed {
			return s
		}
	}
	return s
}

func normalizeWhitespace(s string) string {
	var b strings.Builder
	lastSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !lastSpace {
				b.WriteByte(' ')
				lastSpace = true
			}
			continue
		}
		b.WriteRune(r)
		lastSpace = false
	}
	s = strings.TrimSpace(b.String())
	replacer := strings.NewReplacer(
		" *", "*",
		"( ", "(",
		" )", ")",
		" ,", ",",
	)
	s = replacer.Replace(s)
	s = strings.ReplaceAll(s, "_IDEC_flags* pfpsf", "_IDEC_flags*pfpsf")
	return s
}

func isIdentStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentPart(r rune) bool {
	return isIdentStart(r) || unicode.IsDigit(r)
}

type conditionalState struct {
	stack []branchState
}

type branchState struct {
	parentActive bool
	branchTaken  bool
	active       bool
}

func newConditionalState() *conditionalState {
	return &conditionalState{}
}

func (s *conditionalState) active() bool {
	if len(s.stack) == 0 {
		return true
	}
	return s.stack[len(s.stack)-1].active
}

func (s *conditionalState) push(cond bool) {
	parent := s.active()
	s.stack = append(s.stack, branchState{
		parentActive: parent,
		branchTaken:  parent && cond,
		active:       parent && cond,
	})
}

func (s *conditionalState) elif(cond bool) {
	if len(s.stack) == 0 {
		return
	}
	top := &s.stack[len(s.stack)-1]
	if !top.parentActive || top.branchTaken {
		top.active = false
		return
	}
	top.active = cond
	if cond {
		top.branchTaken = true
	}
}

func (s *conditionalState) elseBranch() {
	if len(s.stack) == 0 {
		return
	}
	top := &s.stack[len(s.stack)-1]
	top.active = top.parentActive && !top.branchTaken
	top.branchTaken = true
}

func (s *conditionalState) pop() {
	if len(s.stack) == 0 {
		return
	}
	s.stack = s.stack[:len(s.stack)-1]
}

func evalExpr(expr string, macros map[string]string) bool {
	parser := exprParser{tokens: tokenizeExpr(expr), macros: macros}
	value, ok := parser.parse()
	return ok && value != 0
}

type exprParser struct {
	tokens []string
	pos    int
	macros map[string]string
}

func tokenizeExpr(expr string) []string {
	var tokens []string
	for i := 0; i < len(expr); {
		switch {
		case unicode.IsSpace(rune(expr[i])):
			i++
		case i+1 < len(expr) && (expr[i:i+2] == "&&" || expr[i:i+2] == "||" || expr[i:i+2] == "==" || expr[i:i+2] == "!="):
			tokens = append(tokens, expr[i:i+2])
			i += 2
		case strings.ContainsRune("!()", rune(expr[i])):
			tokens = append(tokens, expr[i:i+1])
			i++
		case isIdentStart(rune(expr[i])):
			j := i + 1
			for j < len(expr) && isIdentPart(rune(expr[j])) {
				j++
			}
			tokens = append(tokens, expr[i:j])
			i = j
		case unicode.IsDigit(rune(expr[i])):
			j := i + 1
			for j < len(expr) && (unicode.IsDigit(rune(expr[j])) || unicode.IsLetter(rune(expr[j]))) {
				j++
			}
			tokens = append(tokens, expr[i:j])
			i = j
		default:
			i++
		}
	}
	return tokens
}

func (p *exprParser) parse() (int64, bool) {
	return p.parseOr()
}

func (p *exprParser) parseOr() (int64, bool) {
	left, ok := p.parseAnd()
	if !ok {
		return 0, false
	}
	for p.match("||") {
		right, ok := p.parseAnd()
		if !ok {
			return 0, false
		}
		if left != 0 || right != 0 {
			left = 1
		} else {
			left = 0
		}
	}
	return left, true
}

func (p *exprParser) parseAnd() (int64, bool) {
	left, ok := p.parseUnary()
	if !ok {
		return 0, false
	}
	for p.match("&&") {
		right, ok := p.parseUnary()
		if !ok {
			return 0, false
		}
		if left != 0 && right != 0 {
			left = 1
		} else {
			left = 0
		}
	}
	return left, true
}

func (p *exprParser) parseUnary() (int64, bool) {
	if p.match("!") {
		value, ok := p.parseUnary()
		if !ok {
			return 0, false
		}
		if value == 0 {
			return 1, true
		}
		return 0, true
	}
	if p.match("defined") {
		if p.match("(") {
			name := p.next()
			p.match(")")
			if _, ok := p.macros[name]; ok {
				return 1, true
			}
			return 0, true
		}
		name := p.next()
		if _, ok := p.macros[name]; ok {
			return 1, true
		}
		return 0, true
	}
	if p.match("(") {
		value, ok := p.parseOr()
		if !ok {
			return 0, false
		}
		p.match(")")
		return value, true
	}
	token := p.next()
	if token == "" {
		return 0, false
	}
	if n, err := parseIntToken(token); err == nil {
		return n, true
	}
	if value, ok := p.macros[token]; ok {
		fields := strings.Fields(value)
		if len(fields) > 0 {
			if n, err := parseIntToken(fields[0]); err == nil {
				return n, true
			}
		}
		if value == "" {
			return 0, true
		}
	}
	return 0, true
}

func (p *exprParser) match(want string) bool {
	if p.pos >= len(p.tokens) || p.tokens[p.pos] != want {
		return false
	}
	p.pos++
	return true
}

func (p *exprParser) next() string {
	if p.pos >= len(p.tokens) {
		return ""
	}
	token := p.tokens[p.pos]
	p.pos++
	return token
}

func parseIntToken(token string) (int64, error) {
	token = strings.TrimSpace(token)
	token = strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(token, "u"), "U"), "l")
	token = strings.TrimSuffix(strings.TrimSuffix(token, "L"), "ll")
	token = strings.TrimSuffix(token, "LL")
	if token == "" {
		return 0, fmt.Errorf("empty token")
	}
	if strings.HasPrefix(token, "0x") || strings.HasPrefix(token, "0X") {
		return strconv.ParseInt(token[2:], 16, 64)
	}
	return strconv.ParseInt(token, 10, 64)
}
