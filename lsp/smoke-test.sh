#!/usr/bin/env bash
# Quick smoke test for the SuperDB LSP server.
# Sends a few requests and prints the responses.
#
# Usage:
#   make build && lsp/smoke-test.sh
#   make build && cd lsp && ./smoke-test.sh

set -euo pipefail

# Find the LSP binary relative to the script location or cwd
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

for dir in "$SCRIPT_DIR" .; do
  if [[ -x "$dir/superdb-lsp" ]]; then
    LSP="$dir/superdb-lsp"
    break
  elif [[ -x "$dir/lsp" ]]; then
    LSP="$dir/lsp"
    break
  fi
done

if [[ -z "${LSP:-}" ]]; then
  echo "Build first: make build  (from repo root)"
  exit 1
fi

# Helper: wrap a JSON-RPC message with Content-Length header
lsp_msg() {
  local body="$1"
  printf 'Content-Length: %d\r\n\r\n%s' "${#body}" "$body"
}

# Build a full session: initialize, open a doc, ask for completions, shutdown
build_session() {
  # 1. Initialize
  lsp_msg '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"capabilities":{}}}'

  # 2. Initialized notification
  lsp_msg '{"jsonrpc":"2.0","method":"initialized","params":{}}'

  # 3. Open a document with a query
  local doc='{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///tmp/test.spq","languageId":"spq","version":1,"text":"yield {x: 1+2} | where x > 1 | sort x"}}}'
  lsp_msg "$doc"

  # 4. Open a document with a syntax error
  local bad='{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///tmp/bad.spq","languageId":"spq","version":1,"text":"from | where"}}}'
  lsp_msg "$bad"

  # 5. Completion after pipe
  local comp='{"jsonrpc":"2.0","id":2,"method":"textDocument/completion","params":{"textDocument":{"uri":"file:///tmp/test.spq"},"position":{"line":0,"character":41}}}'
  lsp_msg "$comp"

  # 6. Hover over "where"
  local hover='{"jsonrpc":"2.0","id":3,"method":"textDocument/hover","params":{"textDocument":{"uri":"file:///tmp/test.spq"},"position":{"line":0,"character":22}}}'
  lsp_msg "$hover"

  # 7. Format the document
  local fmt='{"jsonrpc":"2.0","id":4,"method":"textDocument/formatting","params":{"textDocument":{"uri":"file:///tmp/test.spq"},"options":{"tabSize":2,"insertSpaces":true}}}'
  lsp_msg "$fmt"

  # 8. Shutdown
  lsp_msg '{"jsonrpc":"2.0","id":99,"method":"shutdown","params":null}'

  # 9. Exit
  lsp_msg '{"jsonrpc":"2.0","method":"exit","params":null}'
}

echo "--- Smoke testing superdb-lsp ---"
echo

# Run the session, pipe through super to parse and summarize responses.
# Super reads raw LSP output as lines (stripping \r from \r\n endings),
# strips Content-Length framing, and parses JSON into structured data.
build_session \
  | "$LSP" 2>/dev/null \
  | super -i line -f table -c '
      values regexp_replace(this, "Content-Length.*", "")
      | where this[0:1] == "{"
      | values parse_sup(this)
      | put id:=coalesce(id, -1)
      | values {
          label: case
            when id==-1 then f"<< {method}"
            when id==1  then "Initialize"
            when id==2  then "Completion"
            when id==3  then "Hover"
            when id==4  then "Formatting"
            when id==99 then "Shutdown"
            else f"Response {id}"
          end,
          detail: case
            when id==1  then f"server={result.serverInfo.name} v{result.serverInfo.version}"
            when id==2  then f"{len(result.items)} completion items"
            when id==3  then result.contents.value
            when id==4  then "reformatted"
            when id==99 then "ok"
            when len(params.diagnostics) > 0 then f"{params.uri}: {len(params.diagnostics)} diagnostic(s)"
            else f"{params.uri}: clean"
          end
        }
    ' -

echo "--- Done ---"
