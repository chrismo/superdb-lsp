package main

import (
	"fmt"
	"strings"
)

// getHover returns hover information for the word at the given position
func getHover(text string, pos Position) *Hover {
	word := getWordAtPosition(text, pos)
	if word == "" {
		return nil
	}

	b := Builtins.Lookup(word)
	if b == nil {
		return nil
	}

	return &Hover{
		Contents: MarkupContent{
			Kind:  MarkupKindMarkdown,
			Value: formatHoverContent(b),
		},
	}
}

// formatHoverContent formats a Builtin into markdown hover content
func formatHoverContent(b *Builtin) string {
	switch b.Kind {
	case KindFunction, KindAggregate:
		if b.Signature != "" {
			doc := b.Doc
			if doc == "" {
				doc = b.Brief
			}
			return fmt.Sprintf("```spq\n%s\n```\n\n%s", b.Signature, doc)
		}
		kindName := "function"
		if b.Kind == KindAggregate {
			kindName = "aggregate"
		}
		return fmt.Sprintf("**%s** (%s)\n\n%s", b.Name, kindName, b.Brief)

	case KindKeyword:
		return fmt.Sprintf("**%s** (keyword)\n\n%s", b.Name, b.Brief)

	case KindOperator:
		return fmt.Sprintf("**%s** (operator)\n\n%s", b.Name, b.Brief)

	case KindType:
		return fmt.Sprintf("**%s** (type)\n\n%s", b.Name, b.Brief)

	default:
		return fmt.Sprintf("**%s**\n\n%s", b.Name, b.Brief)
	}
}

// getWordAtPosition extracts the word at the given position
func getWordAtPosition(text string, pos Position) string {
	lines := strings.Split(text, "\n")
	if pos.Line >= len(lines) {
		return ""
	}

	line := lines[pos.Line]
	if pos.Character > len(line) {
		return ""
	}

	// Find word boundaries
	start := pos.Character
	end := pos.Character

	// Move start backward to find word start
	for start > 0 && isIdentifierChar(line[start-1]) {
		start--
	}

	// Move end forward to find word end
	for end < len(line) && isIdentifierChar(line[end]) {
		end++
	}

	if start == end {
		return ""
	}

	return line[start:end]
}
