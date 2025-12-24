package main

import (
	"strings"
)

// getSignatureHelp returns signature help for the current position
func getSignatureHelp(text string, pos Position) *SignatureHelp {
	// Find the function call context
	funcName, paramIndex := findFunctionContext(text, pos)
	if funcName == "" {
		return nil
	}

	b := Builtins.Lookup(funcName)
	if b == nil || (b.Kind != KindFunction && b.Kind != KindAggregate) {
		return nil
	}

	if b.Signature == "" {
		return nil
	}

	return buildSignatureHelp(b, paramIndex)
}

// buildSignatureHelp creates a SignatureHelp from a Builtin
func buildSignatureHelp(b *Builtin, activeParam int) *SignatureHelp {
	params := make([]ParameterInformation, len(b.Parameters))

	// Calculate parameter label offsets
	labelOffset := strings.Index(b.Signature, "(") + 1
	currentOffset := labelOffset

	for i, p := range b.Parameters {
		// Find this parameter in the label
		paramStart := strings.Index(b.Signature[currentOffset:], p.Name)
		if paramStart == -1 {
			continue
		}
		paramStart += currentOffset

		// Find the end of this parameter (comma or closing paren)
		paramEnd := paramStart + len(p.Name)
		for paramEnd < len(b.Signature) && b.Signature[paramEnd] != ',' && b.Signature[paramEnd] != ')' {
			paramEnd++
		}

		params[i] = ParameterInformation{
			Label: [2]int{paramStart, paramEnd},
			Documentation: &MarkupContent{
				Kind:  MarkupKindPlainText,
				Value: p.Doc,
			},
		}

		currentOffset = paramEnd + 1
	}

	if activeParam >= len(params) {
		activeParam = len(params) - 1
	}
	if activeParam < 0 {
		activeParam = 0
	}

	doc := b.Doc
	if doc == "" {
		doc = b.Brief
	}

	return &SignatureHelp{
		Signatures: []SignatureInformation{
			{
				Label: b.Signature,
				Documentation: &MarkupContent{
					Kind:  MarkupKindPlainText,
					Value: doc,
				},
				Parameters: params,
			},
		},
		ActiveSignature: 0,
		ActiveParameter: activeParam,
	}
}

// findFunctionContext finds the function name and parameter index at position
func findFunctionContext(text string, pos Position) (string, int) {
	lines := strings.Split(text, "\n")
	if pos.Line >= len(lines) {
		return "", 0
	}

	// Get text up to cursor position
	var textToCursor strings.Builder
	for i := 0; i <= pos.Line && i < len(lines); i++ {
		if i == pos.Line {
			if pos.Character <= len(lines[i]) {
				textToCursor.WriteString(lines[i][:pos.Character])
			} else {
				textToCursor.WriteString(lines[i])
			}
		} else {
			textToCursor.WriteString(lines[i])
			textToCursor.WriteByte('\n')
		}
	}

	content := textToCursor.String()

	// Walk backward to find matching open paren
	parenDepth := 0
	funcEnd := -1

	for i := len(content) - 1; i >= 0; i-- {
		ch := content[i]
		switch ch {
		case ')':
			parenDepth++
		case '(':
			if parenDepth == 0 {
				funcEnd = i
				break
			}
			parenDepth--
		}
		if funcEnd >= 0 {
			break
		}
	}

	if funcEnd < 0 {
		return "", 0
	}

	// Extract function name
	funcStart := funcEnd - 1
	for funcStart >= 0 && isIdentifierChar(content[funcStart]) {
		funcStart--
	}
	funcStart++

	if funcStart >= funcEnd {
		return "", 0
	}

	funcName := content[funcStart:funcEnd]

	// Count commas to determine parameter index
	paramIndex := 0
	parenDepth = 0
	for i := funcEnd + 1; i < len(content); i++ {
		ch := content[i]
		switch ch {
		case '(':
			parenDepth++
		case ')':
			parenDepth--
		case ',':
			if parenDepth == 0 {
				paramIndex++
			}
		}
	}

	return funcName, paramIndex
}
