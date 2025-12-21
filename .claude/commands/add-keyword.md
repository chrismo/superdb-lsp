Add the keyword or function "$ARGUMENTS" to the grammar.

## Steps

1. Determine the type (keyword, operator, function, aggregate, or type)
2. Add to `lsp/completion.go` in the appropriate list with a description
3. Add to `supersql/spq.tmb/Syntaxes/spq.tmLanguage.json` in the appropriate pattern
4. Run tests: `cd lsp && go test -v`
5. Report what was added and where
