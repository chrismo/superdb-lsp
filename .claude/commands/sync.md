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

### 2. Review Release Changes for Breaking Changes

**Get the current synced version** from `lsp/version.go` (the `Version` constant).

#### 2a. Scan Release Notes & Commits

**Review what changed** between releases:
- Check release notes at https://github.com/brimdata/super/releases for high-level changes
- Fetch commits between old and new tags: `https://api.github.com/repos/brimdata/super/compare/<old-tag>...<new-tag>`

**Flag commits with breaking-change signals.** Scan each commit message
for keywords (case-insensitive): `remove`, `rename`, `deprecate`,
`breaking`, `no longer`, `drop`, `delete`, `replace`, `incompatible`.
Collect these as potential breaking changes.

Also review commits for:
- Function/aggregate signatures (return types, parameters)
- Renamed or removed functions
- New functions not in the registry files
- Changes to operators or syntax (especially in parser/grammar files)

Scan release body text for the same breaking-change keywords — release
notes often summarize breaking changes more clearly than individual
commits.

#### 2b. Check asdf versions.txt

**Fetch the asdf-superdb versions file** for annotated breaking changes:
- `https://raw.githubusercontent.com/chrismo/asdf-superdb/main/scripts/versions.txt`
- Look for comment lines (starting with `#`) near versions newer than
  the last synced version — these may note known breaking changes

#### 2c. Present Findings

Before proceeding, output a **Breaking Change Review** section listing:
- Flagged commits (SHA, message, why it was flagged)
- Flagged release notes
- Any annotations from versions.txt

If potential breaking changes are found, note them for the CHANGELOG
and continue with the sync. The human will review the report at the end.

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

Update `CHANGELOG.md` with a new version entry. Use these section headers:
- `### Added` — new keywords, functions, aggregates, types
- `### Changed` — non-breaking signature or behavior changes
- `### Breaking` — breaking changes (removed/renamed operators,
  functions, syntax changes that break existing queries). The MCP
  sync downstream scans for this header specifically.
- `### Fixed` — bug fixes

### 8. Commit & Push

If changes were made:
- Stage all changes
- Commit with a descriptive message listing what was added/changed
- Include the new version number in the commit message
- Push to the current branch

### 9. Report

Summarize what was done:

**Version**: new version number and commit hash

**Breaking Change Summary** (most important — put this first):
- List each potential breaking change found in step 2
- For each: commit SHA, one-line description, and affected area
  (syntax, function, operator, type, etc.)
- If none found, say "No breaking changes detected"

**Changes**:
- Number of new items added (by category)
- Any signature changes
- Items removed or renamed

**Build**:
- Test results
- Binary size
