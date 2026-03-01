module github.com/superdb/superdb-lsp/lsp

go 1.24.7

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/brimdata/super v0.2.0
)

require (
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
)

replace github.com/brimdata/super v0.2.0 => ../_deps/super
