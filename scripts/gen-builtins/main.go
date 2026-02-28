// gen-builtins generates grammar_generated.go from the upstream brimdata/super
// PEG grammar. It also reports differences in functions and aggregates between
// upstream Go source and the local builtins.go, which require manual edits.
//
// Usage (via go generate from lsp/):
//
//	cd lsp && go generate ./...
//
// Or directly (from the scripts/gen-builtins directory):
//
//	go run . /path/to/lsp
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func main() {
	if len(os.Args) < 2 {
		fatalf("usage: gen-builtins <lsp-dir>")
	}
	lspDir := os.Args[1]

	// Find upstream source directory via go list run in the lsp module.
	dir, version, err := findUpstreamDir(lspDir)
	if err != nil {
		fatalf("finding upstream module: %v", err)
	}

	// Read the PEG grammar.
	pegFile := filepath.Join(dir, "compiler/parser/parser.peg")
	pegText, err := os.ReadFile(pegFile)
	if err != nil {
		fatalf("reading PEG grammar: %v", err)
	}
	peg := string(pegText)

	// Extract grammar elements.
	keywords := extractPEGKeywords(peg)
	operators := extractPEGOperators(peg)
	primitiveTypes := extractPEGPrimitiveTypes(peg)
	sqlTypes := extractPEGSQLTypes(peg)

	// Deduplicate: operator > type > keyword.
	operatorSet := toSet(operators)
	allTypes := append(primitiveTypes, sqlTypes...)
	typeSet := toSet(allTypes)
	var filteredKeywords []string
	for _, kw := range keywords {
		if !operatorSet[kw] && !typeSet[kw] {
			filteredKeywords = append(filteredKeywords, kw)
		}
	}

	// Generate grammar_generated.go in the lsp directory.
	outFile := filepath.Join(lspDir, "grammar_generated.go")
	if err := generateFile(outFile, filteredKeywords, operators, primitiveTypes, sqlTypes, version); err != nil {
		fatalf("generating file: %v", err)
	}

	// Parse upstream Go source for diff report.
	funcFile := filepath.Join(dir, "runtime/sam/expr/function/function.go")
	aggFile := filepath.Join(dir, "runtime/sam/expr/agg/agg.go")

	upstreamFuncs, err := extractFunctions(funcFile)
	if err != nil {
		fatalf("parsing function.go: %v", err)
	}
	upstreamAggs, err := extractAggregates(aggFile)
	if err != nil {
		fatalf("parsing agg.go: %v", err)
	}

	// Parse local builtins.go for comparison.
	builtinsFile := filepath.Join(lspDir, "builtins.go")
	localBuiltins, err := extractLocalBuiltins(builtinsFile)
	if err != nil {
		fatalf("parsing builtins.go: %v", err)
	}

	// Print diff report.
	printDiffReport(upstreamFuncs, upstreamAggs, localBuiltins, version)
}

// ---------------------------------------------------------------------------
// Upstream module discovery
// ---------------------------------------------------------------------------

type moduleInfo struct {
	Dir     string `json:"Dir"`
	Version string `json:"Version"`
}

func findUpstreamDir(lspDir string) (dir, version string, err error) {
	cmd := exec.Command("go", "list", "-m", "-json", "github.com/brimdata/super")
	cmd.Dir = lspDir
	out, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("go list: %w", err)
	}
	var info moduleInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return "", "", fmt.Errorf("parsing go list output: %w", err)
	}
	if info.Dir == "" {
		return "", "", fmt.Errorf("module not downloaded; run: go mod download")
	}
	return info.Dir, info.Version, nil
}

// ---------------------------------------------------------------------------
// PEG grammar extraction
//
// These functions parse the PEG text directly using Go string operations.
// No regex or external dependencies needed — the patterns are regular enough.
// ---------------------------------------------------------------------------

// extractPEGKeywords finds rules matching: ALL_CAPS = "..."i !IdentifierRest
func extractPEGKeywords(peg string) []string {
	var keywords []string
	for _, line := range strings.Split(peg, "\n") {
		// Keyword rules start at column 0 with an ALL_CAPS identifier.
		if len(line) == 0 || !unicode.IsUpper(rune(line[0])) {
			continue
		}
		name := scanIdent(line)
		if name == "" || !isAllCaps(name) {
			continue
		}
		if !strings.Contains(line, "!IdentifierRest") {
			continue
		}
		// Extract the quoted value: "..."i
		val := extractCaseInsensitiveLiteral(line)
		if val != "" {
			keywords = append(keywords, strings.ToLower(val))
		}
	}
	sort.Strings(keywords)
	return keywords
}

// extractPEGOperators collects rule references ending in "Op" from the
// Operator rule's alternatives.
func extractPEGOperators(peg string) []string {
	block := findRuleBlock(peg, "Operator")
	if block == "" {
		return nil
	}
	seen := make(map[string]bool)
	var ops []string
	for _, word := range strings.Fields(block) {
		// Strip label prefix (e.g., "op:SQLOp" → "SQLOp").
		if idx := strings.IndexByte(word, ':'); idx >= 0 {
			word = word[idx+1:]
		}
		if !strings.HasSuffix(word, "Op") {
			continue
		}
		if !unicode.IsUpper(rune(word[0])) {
			continue
		}
		if word == "SQLOp" || word == "EndOfOp" {
			continue
		}
		op := strings.ToLower(strings.TrimSuffix(word, "Op"))
		if !seen[op] {
			seen[op] = true
			ops = append(ops, op)
		}
	}
	sort.Strings(ops)
	return ops
}

// extractPEGPrimitiveTypes collects non-case-insensitive string literals
// from the PrimitiveType rule (e.g., "uint8", "string", "time").
func extractPEGPrimitiveTypes(peg string) []string {
	block := findRuleBlock(peg, "PrimitiveType")
	if block == "" {
		return nil
	}
	types := extractLiterals(block, false)
	sort.Strings(types)
	return types
}

// extractPEGSQLTypes collects the first case-insensitive string literal from
// each alternative in the PostgreSQLPrimitiveType rule. Only the leading
// literal of each alternative is taken, so negative lookaheads like !"a"i
// are not included.
func extractPEGSQLTypes(peg string) []string {
	block := findRuleBlock(peg, "PostgreSQLPrimitiveType")
	if block == "" {
		return nil
	}
	seen := make(map[string]bool)
	var result []string
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		// Each alternative starts with "=" or "/".
		if len(line) == 0 {
			continue
		}
		if line[0] != '=' && line[0] != '/' {
			continue
		}
		// Extract first "..."i literal on this line.
		val := extractCaseInsensitiveLiteral(line)
		if val == "" {
			continue
		}
		t := strings.ToLower(val)
		if !seen[t] {
			seen[t] = true
			result = append(result, t)
		}
	}
	sort.Strings(result)
	return result
}

// findRuleBlock returns the text of a named PEG rule (from its name line
// through the last indented continuation line before the next rule).
func findRuleBlock(peg, ruleName string) string {
	lines := strings.Split(peg, "\n")
	start := -1
	for i, line := range lines {
		if start < 0 {
			// Look for the rule name at the start of a line.
			ident := scanIdent(line)
			if ident == ruleName {
				start = i
			}
		} else if i > start && len(line) > 0 && !isSpace(line[0]) && unicode.IsLetter(rune(line[0])) {
			// New rule begins — end of block.
			return strings.Join(lines[start:i], "\n")
		}
	}
	if start >= 0 {
		return strings.Join(lines[start:], "\n")
	}
	return ""
}

// extractLiterals scans a PEG rule block for string literals, skipping
// code blocks (brace-delimited). If wantCaseInsensitive is true, only
// literals with the "i" suffix are returned; otherwise only those without.
func extractLiterals(block string, wantCaseInsensitive bool) []string {
	var results []string
	braceDepth := 0
	inGoStr := false
	i := 0
	for i < len(block) {
		ch := block[i]

		// Track code blocks via brace depth.
		if !inGoStr {
			if ch == '{' {
				braceDepth++
				i++
				continue
			}
			if ch == '}' && braceDepth > 0 {
				braceDepth--
				i++
				continue
			}
		}

		// Inside a code block: track Go strings so that braces within
		// them don't confuse the depth counter.
		if braceDepth > 0 {
			if ch == '"' {
				inGoStr = !inGoStr
			} else if ch == '\\' && inGoStr {
				i++ // skip escaped char
			}
			i++
			continue
		}

		// Outside code blocks: look for PEG string literals ("...").
		if ch == '"' {
			end := strings.IndexByte(block[i+1:], '"')
			if end < 0 {
				break
			}
			val := block[i+1 : i+1+end]
			nextPos := i + 1 + end + 1
			isCaseInsensitive := nextPos < len(block) && block[nextPos] == 'i'
			if isCaseInsensitive == wantCaseInsensitive && val != "" {
				results = append(results, val)
			}
			i = nextPos
			if isCaseInsensitive {
				i++
			}
			continue
		}
		i++
	}
	return results
}

// extractCaseInsensitiveLiteral finds the first "..."i literal on a line.
func extractCaseInsensitiveLiteral(line string) string {
	for i := 0; i < len(line); i++ {
		if line[i] != '"' {
			continue
		}
		end := strings.IndexByte(line[i+1:], '"')
		if end < 0 {
			return ""
		}
		val := line[i+1 : i+1+end]
		nextPos := i + 1 + end + 1
		if nextPos < len(line) && line[nextPos] == 'i' {
			return val
		}
		i = nextPos
	}
	return ""
}

// scanIdent reads the leading identifier from a string.
func scanIdent(s string) string {
	i := 0
	for i < len(s) {
		r := rune(s[i])
		if unicode.IsLetter(r) || r == '_' || (i > 0 && unicode.IsDigit(r)) {
			i++
		} else {
			break
		}
	}
	return s[:i]
}

func isAllCaps(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r != '_' && !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t'
}

// ---------------------------------------------------------------------------
// Go AST: function.go extraction
// ---------------------------------------------------------------------------

type funcInfo struct {
	Name   string
	ArgMin int
	ArgMax int
}

func extractFunctions(filename string) ([]funcInfo, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, err
	}
	fn := findFuncDecl(file, "New")
	if fn == nil {
		return nil, fmt.Errorf("func New not found in %s", filename)
	}
	sw := findSwitch(fn.Body)
	if sw == nil {
		return nil, fmt.Errorf("switch statement not found in func New")
	}
	return extractCaseFunctions(sw), nil
}

func extractCaseFunctions(sw *ast.SwitchStmt) []funcInfo {
	var funcs []funcInfo
	for _, stmt := range sw.Body.List {
		cc, ok := stmt.(*ast.CaseClause)
		if !ok || cc.List == nil {
			continue
		}
		names := extractStringLiterals(cc.List)
		argmin, argmax := extractArity(cc.Body)
		for _, name := range names {
			funcs = append(funcs, funcInfo{Name: name, ArgMin: argmin, ArgMax: argmax})
		}
	}
	return funcs
}

func extractArity(stmts []ast.Stmt) (argmin, argmax int) {
	argmin, argmax = 1, 1
	for _, stmt := range stmts {
		assign, ok := stmt.(*ast.AssignStmt)
		if !ok {
			continue
		}
		if len(assign.Lhs) == 1 && len(assign.Rhs) == 1 {
			if ident, ok := assign.Lhs[0].(*ast.Ident); ok {
				switch ident.Name {
				case "argmin":
					argmin = intVal(assign.Rhs[0])
				case "argmax":
					argmax = intVal(assign.Rhs[0])
				}
			}
		}
		if len(assign.Lhs) == 2 && len(assign.Rhs) == 2 {
			for i, lhs := range assign.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					switch ident.Name {
					case "argmin":
						argmin = intVal(assign.Rhs[i])
					case "argmax":
						argmax = intVal(assign.Rhs[i])
					}
				}
			}
		}
	}
	return
}

func intVal(expr ast.Expr) int {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.INT {
			n, _ := strconv.Atoi(e.Value)
			return n
		}
	case *ast.UnaryExpr:
		if e.Op == token.SUB {
			if lit, ok := e.X.(*ast.BasicLit); ok && lit.Kind == token.INT {
				n, _ := strconv.Atoi(lit.Value)
				return -n
			}
		}
	}
	return 0
}

// ---------------------------------------------------------------------------
// Go AST: agg.go extraction
// ---------------------------------------------------------------------------

func extractAggregates(filename string) ([]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, err
	}
	fn := findFuncDecl(file, "NewPattern")
	if fn == nil {
		return nil, fmt.Errorf("func NewPattern not found in %s", filename)
	}
	sw := findSwitch(fn.Body)
	if sw == nil {
		return nil, fmt.Errorf("switch statement not found in func NewPattern")
	}
	var aggs []string
	for _, stmt := range sw.Body.List {
		cc, ok := stmt.(*ast.CaseClause)
		if !ok || cc.List == nil {
			continue
		}
		aggs = append(aggs, extractStringLiterals(cc.List)...)
	}
	sort.Strings(aggs)
	return aggs, nil
}

// ---------------------------------------------------------------------------
// Go AST: builtins.go extraction
// ---------------------------------------------------------------------------

type localBuiltin struct {
	Name       string
	Kind       string
	ParamCount int
}

func extractLocalBuiltins(filename string) ([]localBuiltin, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, err
	}
	compLit := findVarCompositeLit(file, "allBuiltins")
	if compLit == nil {
		return nil, fmt.Errorf("var allBuiltins not found in %s", filename)
	}
	var builtins []localBuiltin
	for _, elt := range compLit.Elts {
		cl, ok := elt.(*ast.CompositeLit)
		if !ok {
			continue
		}
		var lb localBuiltin
		for _, field := range cl.Elts {
			kv, ok := field.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := kv.Key.(*ast.Ident)
			if !ok {
				continue
			}
			switch key.Name {
			case "Name":
				if lit, ok := kv.Value.(*ast.BasicLit); ok {
					lb.Name, _ = strconv.Unquote(lit.Value)
				}
			case "Kind":
				if ident, ok := kv.Value.(*ast.Ident); ok {
					lb.Kind = ident.Name
				}
			case "Parameters":
				if cl2, ok := kv.Value.(*ast.CompositeLit); ok {
					lb.ParamCount = len(cl2.Elts)
				}
			}
		}
		builtins = append(builtins, lb)
	}
	return builtins, nil
}

// ---------------------------------------------------------------------------
// Go AST helpers
// ---------------------------------------------------------------------------

func findFuncDecl(file *ast.File, name string) *ast.FuncDecl {
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == name {
			return fn
		}
	}
	return nil
}

func findSwitch(block *ast.BlockStmt) *ast.SwitchStmt {
	var sw *ast.SwitchStmt
	ast.Inspect(block, func(n ast.Node) bool {
		if s, ok := n.(*ast.SwitchStmt); ok && sw == nil {
			sw = s
			return false
		}
		return true
	})
	return sw
}

func findVarCompositeLit(file *ast.File, name string) *ast.CompositeLit {
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok || len(vs.Names) == 0 || vs.Names[0].Name != name {
				continue
			}
			if len(vs.Values) > 0 {
				if cl, ok := vs.Values[0].(*ast.CompositeLit); ok {
					return cl
				}
			}
		}
	}
	return nil
}

func extractStringLiterals(exprs []ast.Expr) []string {
	var result []string
	for _, expr := range exprs {
		if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			if s, err := strconv.Unquote(lit.Value); err == nil {
				result = append(result, s)
			}
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// Code generation
// ---------------------------------------------------------------------------

func generateFile(outFile string, keywords, operators, primitiveTypes, sqlTypes []string, version string) error {
	var b strings.Builder

	fmt.Fprintf(&b, "// Code generated by gen-builtins; DO NOT EDIT.\n")
	fmt.Fprintf(&b, "// Source: brimdata/super@%s\n", version)
	fmt.Fprintf(&b, "// Generated: %s\n\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "package main\n\n")
	fmt.Fprintf(&b, "var grammarBuiltins = []Builtin{\n")

	fmt.Fprintf(&b, "\t// Keywords\n")
	for _, kw := range keywords {
		fmt.Fprintf(&b, "\t{Name: %q, Kind: KindKeyword},\n", kw)
	}
	fmt.Fprintf(&b, "\n")

	fmt.Fprintf(&b, "\t// Operators\n")
	for _, op := range operators {
		fmt.Fprintf(&b, "\t{Name: %q, Kind: KindOperator},\n", op)
	}
	fmt.Fprintf(&b, "\n")

	fmt.Fprintf(&b, "\t// Primitive types\n")
	for _, t := range primitiveTypes {
		fmt.Fprintf(&b, "\t{Name: %q, Kind: KindType},\n", t)
	}
	fmt.Fprintf(&b, "\n")

	fmt.Fprintf(&b, "\t// SQL type aliases\n")
	for _, t := range sqlTypes {
		fmt.Fprintf(&b, "\t{Name: %q, Kind: KindType},\n", t)
	}

	fmt.Fprintf(&b, "}\n")

	return os.WriteFile(outFile, []byte(b.String()), 0644)
}

// ---------------------------------------------------------------------------
// Diff report
// ---------------------------------------------------------------------------

func printDiffReport(upstreamFuncs []funcInfo, upstreamAggs []string, local []localBuiltin, version string) {
	localFuncs := make(map[string]int) // name → param count
	localAggs := make(map[string]bool)
	for _, lb := range local {
		switch lb.Kind {
		case "KindFunction":
			localFuncs[lb.Name] = lb.ParamCount
		case "KindAggregate":
			localAggs[lb.Name] = true
		}
	}

	upFuncMap := make(map[string]funcInfo)
	for _, f := range upstreamFuncs {
		upFuncMap[f.Name] = f
	}
	upAggSet := toSet(upstreamAggs)

	fmt.Printf("=== Diff Report (brimdata/super@%s) ===\n\n", version)

	// Functions.
	fmt.Println("--- Functions ---")
	hasDiff := false
	for _, f := range upstreamFuncs {
		if _, ok := localFuncs[f.Name]; !ok {
			fmt.Printf("  NEW:     %s (argmin=%d, argmax=%d)\n", f.Name, f.ArgMin, f.ArgMax)
			hasDiff = true
		}
	}
	for name := range localFuncs {
		if _, ok := upFuncMap[name]; !ok {
			fmt.Printf("  REMOVED: %s\n", name)
			hasDiff = true
		}
	}
	for name, paramCount := range localFuncs {
		if f, ok := upFuncMap[name]; ok {
			if f.ArgMin != paramCount || f.ArgMax != paramCount {
				fmt.Printf("  ARITY:   %s (local params=%d, upstream argmin=%d, argmax=%d)\n",
					name, paramCount, f.ArgMin, f.ArgMax)
				hasDiff = true
			}
		}
	}
	if !hasDiff {
		fmt.Println("  (no changes)")
	}
	fmt.Println()

	// Aggregates.
	fmt.Println("--- Aggregates ---")
	hasDiff = false
	for _, name := range upstreamAggs {
		if !localAggs[name] {
			fmt.Printf("  NEW:     %s\n", name)
			hasDiff = true
		}
	}
	for name := range localAggs {
		if !upAggSet[name] {
			fmt.Printf("  REMOVED: %s\n", name)
			hasDiff = true
		}
	}
	if !hasDiff {
		fmt.Println("  (no changes)")
	}
	fmt.Println()
}

// ---------------------------------------------------------------------------
// Utility
// ---------------------------------------------------------------------------

func toSet(ss []string) map[string]bool {
	m := make(map[string]bool, len(ss))
	for _, s := range ss {
		m[s] = true
	}
	return m
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "gen-builtins: "+format+"\n", args...)
	os.Exit(1)
}
