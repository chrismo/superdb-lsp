Check if the LSP and TextMate grammar are in sync with the latest brimdata/super PEG grammar.

## Steps

1. **Fetch the latest sources** from brimdata/zed (main branch):
   - `compiler/parser/parser.peg` - for keywords, operators, types
   - `runtime/sam/expr/function/function.go` - for built-in functions
   - `runtime/sam/expr/agg/agg.go` - for aggregate functions

2. **Compare against local files**:
   - `lsp/completion.go` - LSP completion lists
   - `supersql/spq.tmb/Syntaxes/spq.tmLanguage.json` - TextMate grammar

3. **Report differences**:
   - List any missing keywords, functions, operators, or types
   - List any items we have that aren't in the upstream grammar

4. **If there are differences**:
   - Update `lsp/completion.go` with missing items
   - Update `spq.tmLanguage.json` to match
   - Run `cd lsp && go test -v` to verify tests pass
   - Update the "Last synchronized" date in `lsp/README.md`

5. **Commit changes** if any updates were made
