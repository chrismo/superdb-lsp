package main

import (
	"strings"
)

// getCompletions returns completion items based on the current context
func getCompletions(text string, pos Position) []CompletionItem {
	var items []CompletionItem

	// Get the current line and word being typed
	lines := strings.Split(text, "\n")
	if pos.Line >= len(lines) {
		return items
	}

	line := lines[pos.Line]
	prefix := ""
	if pos.Character <= len(line) {
		// Get the word prefix before cursor
		start := pos.Character
		for start > 0 && isIdentifierChar(line[start-1]) {
			start--
		}
		if start < pos.Character {
			prefix = strings.ToLower(line[start:pos.Character])
		}
	}

	// Check context for better completions
	context := getCompletionContext(line, pos.Character)

	// Add completions based on context
	switch context {
	case contextType:
		// After type-related keywords, suggest types
		items = append(items, getTypeCompletions(prefix)...)
	case contextFunction:
		// After opening paren or in function context
		items = append(items, getFunctionCompletions(prefix)...)
		items = append(items, getAggregateCompletions(prefix)...)
	default:
		// General context - suggest everything
		items = append(items, getKeywordCompletions(prefix)...)
		items = append(items, getOperatorCompletions(prefix)...)
		items = append(items, getFunctionCompletions(prefix)...)
		items = append(items, getAggregateCompletions(prefix)...)
		items = append(items, getTypeCompletions(prefix)...)
	}

	return items
}

type completionContext int

const (
	contextGeneral completionContext = iota
	contextType
	contextFunction
)

// getCompletionContext analyzes the line to determine the completion context
func getCompletionContext(line string, col int) completionContext {
	if col > len(line) {
		col = len(line)
	}
	prefix := strings.ToLower(line[:col])

	// Check if we're after a type cast operator
	if strings.Contains(prefix, "cast(") ||
		strings.Contains(prefix, "::") ||
		strings.HasSuffix(strings.TrimSpace(prefix), "<") {
		return contextType
	}

	// Check if we're inside a function call
	openParens := strings.Count(prefix, "(") - strings.Count(prefix, ")")
	if openParens > 0 {
		return contextFunction
	}

	return contextGeneral
}

func isIdentifierChar(b byte) bool {
	return (b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') ||
		b == '_'
}

func getKeywordCompletions(prefix string) []CompletionItem {
	return getCompletionsByKind(KindKeyword, prefix, CompletionItemKindKeyword, "")
}

func getOperatorCompletions(prefix string) []CompletionItem {
	return getCompletionsByKind(KindOperator, prefix, CompletionItemKindFunction, "operator")
}

func getFunctionCompletions(prefix string) []CompletionItem {
	var items []CompletionItem
	for _, fn := range Builtins.Functions() {
		if prefix == "" || strings.HasPrefix(strings.ToLower(fn.Name), prefix) {
			items = append(items, CompletionItem{
				Label:      fn.Name,
				Kind:       CompletionItemKindFunction,
				Detail:     "function: " + fn.Brief,
				InsertText: fn.Name + "($1)",
			})
		}
	}
	return items
}

func getAggregateCompletions(prefix string) []CompletionItem {
	var items []CompletionItem
	for _, agg := range Builtins.Aggregates() {
		if prefix == "" || strings.HasPrefix(strings.ToLower(agg.Name), prefix) {
			items = append(items, CompletionItem{
				Label:      agg.Name,
				Kind:       CompletionItemKindFunction,
				Detail:     "aggregate: " + agg.Brief,
				InsertText: agg.Name + "($1)",
			})
		}
	}
	return items
}

func getTypeCompletions(prefix string) []CompletionItem {
	return getCompletionsByKind(KindType, prefix, CompletionItemKindClass, "type")
}

// getCompletionsByKind is a helper to build completion items from the registry
func getCompletionsByKind(kind BuiltinKind, prefix string, itemKind int, labelPrefix string) []CompletionItem {
	var items []CompletionItem
	for _, b := range Builtins.ByKind(kind) {
		if prefix == "" || strings.HasPrefix(strings.ToLower(b.Name), prefix) {
			detail := b.Brief
			if labelPrefix != "" {
				detail = labelPrefix + ": " + detail
			}
			items = append(items, CompletionItem{
				Label:  b.Name,
				Kind:   itemKind,
				Detail: detail,
			})
		}
	}
	return items
}
