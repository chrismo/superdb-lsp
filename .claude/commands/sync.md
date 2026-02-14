Perform a full grammar synchronization with upstream brimdata/super.

Use WebFetch or gh instead of curl.

## Do all of this autonomously:

### 1. Fetch Latest Release & Source Files

**Get the latest release** from brimdata/super:
- `https://api.github.com/repos/brimdata/super/releases/latest`
- Extract the tag name (e.g., `v0.1.0`) and release date
- Extract the commit SHA from the release tag

**Fetch these files** from the release tag (not main) to discover new names:

| File | Provides | Why |
|------|----------|-----|
| `compiler/parser/parser.peg` | Keywords, operators, types | The PEG grammar defines the language syntax |
| `runtime/sam/expr/function/function.go` | Built-in scalar functions | Function names are registered at runtime, not in grammar |
| `runtime/sam/expr/agg/agg.go` | Aggregate functions | Aggregate names are registered separately from scalar functions |

### 2. Review Release Changes

**Get the current synced version** from `lsp/version.go` (the `Version` constant).

**Review what changed** between releases:
- Check release notes at https://github.com/brimdata/super/releases for high-level changes
- Fetch commits between old and new tags: `https://api.github.com/repos/brimdata/super/compare/<old-tag>...<new-tag>`

Look for changes that affect:
- Function/aggregate signatures (return types, parameters)
- Renamed or removed functions
- New functions not in the registry files

### 3. Compare & Update

Compare against local files and update if needed:
- `lsp/builtins.go` - add any missing keywords/functions/operators/types, update signatures
- `supersql/spq.tmbundle/Syntaxes/spq.tmLanguage.json` - keep TextMate grammar in sync

### 4. Update Version & Dependencies

**Get the latest release version** from brimdata/super:
- Check https://github.com/brimdata/super/releases for the latest release tag
- Use that version number (e.g., `0.1.0`, `0.2.0`)

Update version in:
- `lsp/version.go`:
  - `Version` constant (match upstream, e.g., `0.1.0`)
  - `LSPPatch` constant (reset to `0` on sync)
  - `SuperCommit` constant (short SHA)
- `supersql/spq.tmbundle/info.plist` - the version string (include LSP patch, e.g., `0.1.0.0`)

**Update Go dependency** to the release tag:
```bash
cd lsp && go get github.com/brimdata/super@<release-tag> && go mod tidy
```
Example: `go get github.com/brimdata/super@v0.1.0`

This ensures the parser used for diagnostics matches the released upstream version.

### 5. Test

Run the full test suite:
```bash
cd lsp && go build -v && go test -v
```
Fix any test failures.

### 6. Build

Build the binary and verify it works:
```bash
cd lsp && go build -o superdb-lsp .
./superdb-lsp --version
```

### 7. Update Docs

Update `lsp/README.md` with:
- New "Last synchronized" date (use the release date)
- Any new keywords/functions added to the reference section

Update `CHANGELOG.md` with a new version entry listing:
- Added items (keywords, functions, aggregates, types)
- Changed items (signature changes, behavior changes)
- Fixed items (bug fixes)

### 8. Commit & Push

If changes were made:
- Stage all changes
- Commit with a descriptive message listing what was added/changed
- Include the new version number in the commit message
- Push to the current branch

### 9. Report

Summarize what was done:
- New version number
- Number of new items added (by category)
- Any signature changes
- Test results
- Binary size
- Commit hash
