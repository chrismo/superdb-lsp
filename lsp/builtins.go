package main

//go:generate sh -c "cd ../scripts/gen-builtins && go run . $PWD"

// builtins.go - Registry of SuperSQL language elements
// Keywords, operators, and types are auto-generated in grammar_generated.go.
// Functions and aggregates (with docs/signatures) are maintained here.

// BuiltinKind categorizes language elements
type BuiltinKind int

const (
	KindKeyword BuiltinKind = iota
	KindOperator
	KindFunction
	KindAggregate
	KindType
)

// Builtin represents a language element with all its metadata
type Builtin struct {
	Name       string
	Kind       BuiltinKind
	Brief      string       // Short description for completion
	Doc        string       // Full documentation for hover
	Signature  string       // Function signature (for functions/aggregates)
	Parameters []ParamDef   // Parameter definitions (for signature help)
}

// ParamDef defines a function parameter
type ParamDef struct {
	Name string
	Doc  string
}

// Registry holds all builtins indexed for fast lookup
type Registry struct {
	byName map[string]*Builtin
	byKind map[BuiltinKind][]*Builtin
}

// Lookup finds a builtin by name (case-insensitive)
func (r *Registry) Lookup(name string) *Builtin {
	return r.byName[toLower(name)]
}

// ByKind returns all builtins of a given kind
func (r *Registry) ByKind(kind BuiltinKind) []*Builtin {
	return r.byKind[kind]
}

// Keywords returns all keywords
func (r *Registry) Keywords() []*Builtin { return r.byKind[KindKeyword] }

// Operators returns all operators
func (r *Registry) Operators() []*Builtin { return r.byKind[KindOperator] }

// Functions returns all functions
func (r *Registry) Functions() []*Builtin { return r.byKind[KindFunction] }

// Aggregates returns all aggregates
func (r *Registry) Aggregates() []*Builtin { return r.byKind[KindAggregate] }

// Types returns all types
func (r *Registry) Types() []*Builtin { return r.byKind[KindType] }

// Builtins is the global registry instance
var Builtins = buildRegistry()

func buildRegistry() *Registry {
	r := &Registry{
		byName: make(map[string]*Builtin),
		byKind: make(map[BuiltinKind][]*Builtin),
	}
	// Generated: keywords, types, operators
	for i := range grammarBuiltins {
		b := &grammarBuiltins[i]
		r.byName[toLower(b.Name)] = b
		r.byKind[b.Kind] = append(r.byKind[b.Kind], b)
	}
	// Manual: functions, aggregates (with docs/signatures)
	for i := range allBuiltins {
		b := &allBuiltins[i]
		r.byName[toLower(b.Name)] = b
		r.byKind[b.Kind] = append(r.byKind[b.Kind], b)
	}
	return r
}

func toLower(s string) string {
	// Fast ASCII lowercase
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

// allBuiltins contains functions and aggregates with human-written
// docs and signatures. Keywords, operators, and types are in
// grammar_generated.go (auto-generated from the PEG grammar).
var allBuiltins = []Builtin{
	// =========================================================================
	// FUNCTIONS (scalar functions)
	// =========================================================================

	{
		Name: "abs", Kind: KindFunction,
		Brief: "Absolute value", Doc: "Returns the absolute value of a number",
		Signature: "abs(value: number) -> number",
		Parameters: []ParamDef{{Name: "value", Doc: "Numeric value"}},
	},
	{
		Name: "base64", Kind: KindFunction,
		Brief: "Base64 encode/decode", Doc: "Encode or decode base64 data",
		Signature: "base64(value: bytes|string) -> string",
		Parameters: []ParamDef{{Name: "value", Doc: "Value to encode/decode"}},
	},
	{
		Name: "bucket", Kind: KindFunction,
		Brief: "Bucket values into ranges", Doc: "Bucket numeric values into fixed-size ranges",
		Signature: "bucket(value: number, size: number) -> number",
		Parameters: []ParamDef{{Name: "value", Doc: "Value to bucket"}, {Name: "size", Doc: "Bucket size"}},
	},
	{
		Name: "ceil", Kind: KindFunction,
		Brief: "Ceiling function", Doc: "Round up to the nearest integer",
		Signature: "ceil(value: number) -> number",
		Parameters: []ParamDef{{Name: "value", Doc: "Numeric value"}},
	},
	{
		Name: "cidr_match", Kind: KindFunction,
		Brief: "Match IP against CIDR", Doc: "Check if an IP address matches a CIDR network",
		Signature: "cidr_match(network: net, ip: ip) -> bool",
		Parameters: []ParamDef{{Name: "network", Doc: "CIDR network"}, {Name: "ip", Doc: "IP address to check"}},
	},
	{
		Name: "coalesce", Kind: KindFunction,
		Brief: "First non-null value", Doc: "Return the first non-null value from arguments",
		Signature: "coalesce(value: any, ...) -> any",
		Parameters: []ParamDef{{Name: "value", Doc: "Values to check"}},
	},
	{
		Name: "concat", Kind: KindFunction,
		Brief: "Concatenate strings", Doc: "Concatenate multiple strings into one",
		Signature: "concat(values: string, ...) -> string",
		Parameters: []ParamDef{{Name: "values", Doc: "Strings to concatenate"}},
	},
	{
		Name: "compare", Kind: KindFunction,
		Brief: "Compare two values", Doc: "Compare two values, returning -1, 0, or 1",
		Signature: "compare(a: any, b: any) -> int64",
		Parameters: []ParamDef{{Name: "a", Doc: "First value"}, {Name: "b", Doc: "Second value"}},
	},
	{
		Name: "date_part", Kind: KindFunction,
		Brief: "Extract date component", Doc: "Extract a component (year, month, day, etc.) from a timestamp",
		Signature: "date_part(part: string, time: time) -> int64",
		Parameters: []ParamDef{{Name: "part", Doc: "Part name (year, month, day, hour, minute, second)"}, {Name: "time", Doc: "Timestamp value"}},
	},
	{
		Name: "fields", Kind: KindFunction,
		Brief: "Get record field names", Doc: "Return the field names of a record as an array",
		Signature: "fields(record: record) -> [string]",
		Parameters: []ParamDef{{Name: "record", Doc: "Record value"}},
	},
	{
		Name: "flatten", Kind: KindFunction,
		Brief: "Flatten nested records", Doc: "Flatten nested record structure into dotted field names",
		Signature: "flatten(record: record) -> record",
		Parameters: []ParamDef{{Name: "record", Doc: "Record to flatten"}},
	},
	{
		Name: "floor", Kind: KindFunction,
		Brief: "Floor function", Doc: "Round down to the nearest integer",
		Signature: "floor(value: number) -> number",
		Parameters: []ParamDef{{Name: "value", Doc: "Numeric value"}},
	},
	{
		Name: "grep", Kind: KindFunction,
		Brief: "Search with pattern", Doc: "Search for a pattern in a value",
		Signature: "grep(pattern: string|regexp, value: any) -> bool",
		Parameters: []ParamDef{{Name: "pattern", Doc: "Search pattern"}, {Name: "value", Doc: "Value to search"}},
	},
	{
		Name: "grok", Kind: KindFunction,
		Brief: "Parse with grok pattern", Doc: "Parse a string using a grok pattern",
		Signature: "grok(pattern: string, value: string) -> record",
		Parameters: []ParamDef{{Name: "pattern", Doc: "Grok pattern"}, {Name: "value", Doc: "String to parse"}},
	},
	{
		Name: "has", Kind: KindFunction,
		Brief: "Check if field exists", Doc: "Check if a record has a specific field",
		Signature: "has(record: record, field: string) -> bool",
		Parameters: []ParamDef{{Name: "record", Doc: "Record to check"}, {Name: "field", Doc: "Field name"}},
	},
	{
		Name: "has_error", Kind: KindFunction,
		Brief: "Check for error", Doc: "Check if a value contains a nested error",
		Signature: "has_error(value: any) -> bool",
		Parameters: []ParamDef{{Name: "value", Doc: "Value to check"}},
	},
	{
		Name: "hex", Kind: KindFunction,
		Brief: "Hexadecimal conversion", Doc: "Convert bytes or string to hexadecimal",
		Signature: "hex(value: bytes|string) -> string",
		Parameters: []ParamDef{{Name: "value", Doc: "Value to convert"}},
	},
	{
		Name: "is", Kind: KindFunction,
		Brief: "Type check function", Doc: "Check if a value is of a specific type",
		Signature: "is(value: any, type: type) -> bool",
		Parameters: []ParamDef{{Name: "value", Doc: "Value to check"}, {Name: "type", Doc: "Type to check against"}},
	},
	{
		Name: "is_error", Kind: KindFunction,
		Brief: "Check if value is error", Doc: "Check if a value is an error",
		Signature: "is_error(value: any) -> bool",
		Parameters: []ParamDef{{Name: "value", Doc: "Value to check"}},
	},
	{
		Name: "join", Kind: KindFunction,
		Brief: "Join strings", Doc: "Join an array of strings with a separator",
		Signature: "join(array: [string], sep: string) -> string",
		Parameters: []ParamDef{{Name: "array", Doc: "Array of strings"}, {Name: "sep", Doc: "Separator"}},
	},
	{
		Name: "kind", Kind: KindFunction,
		Brief: "Get value kind", Doc: "Return the kind of a value (primitive, record, array, etc.)",
		Signature: "kind(value: any) -> string",
		Parameters: []ParamDef{{Name: "value", Doc: "Value to check"}},
	},
	{
		Name: "ksuid", Kind: KindFunction,
		Brief: "Generate KSUID", Doc: "Generate a K-Sortable Unique Identifier",
		Signature: "ksuid() -> string",
		Parameters: []ParamDef{},
	},
	{
		Name: "len", Kind: KindFunction,
		Brief: "Length of value", Doc: "Return the length of a string, bytes, or array",
		Signature: "len(value: string|bytes|array) -> int64",
		Parameters: []ParamDef{{Name: "value", Doc: "Value to measure"}},
	},
	{
		Name: "length", Kind: KindFunction,
		Brief: "Length of value (alias)", Doc: "Return the length of a string, bytes, or array (alias for len)",
		Signature: "length(value: string|bytes|array) -> int64",
		Parameters: []ParamDef{{Name: "value", Doc: "Value to measure"}},
	},
	{
		Name: "levenshtein", Kind: KindFunction,
		Brief: "Levenshtein distance", Doc: "Calculate the Levenshtein edit distance between two strings",
		Signature: "levenshtein(a: string, b: string) -> int64",
		Parameters: []ParamDef{{Name: "a", Doc: "First string"}, {Name: "b", Doc: "Second string"}},
	},
	{
		Name: "log", Kind: KindFunction,
		Brief: "Logarithm", Doc: "Calculate the logarithm of a number",
		Signature: "log(value: number, base?: number) -> float64",
		Parameters: []ParamDef{{Name: "value", Doc: "Numeric value"}, {Name: "base", Doc: "Log base (default: e)"}},
	},
	{
		Name: "lower", Kind: KindFunction,
		Brief: "Convert to lowercase", Doc: "Convert a string to lowercase",
		Signature: "lower(value: string) -> string",
		Parameters: []ParamDef{{Name: "value", Doc: "String to convert"}},
	},
	{
		Name: "missing", Kind: KindFunction,
		Brief: "Create missing value", Doc: "Create a missing value of optional type",
		Signature: "missing(type?: type) -> missing",
		Parameters: []ParamDef{{Name: "type", Doc: "Optional type"}},
	},
	{
		Name: "nameof", Kind: KindFunction,
		Brief: "Get type name", Doc: "Return the name of a value's type",
		Signature: "nameof(value: any) -> string",
		Parameters: []ParamDef{{Name: "value", Doc: "Value to check"}},
	},
	{
		Name: "nest_dotted", Kind: KindFunction,
		Brief: "Nest dotted field names", Doc: "Convert dotted field names into nested records",
		Signature: "nest_dotted(record: record) -> record",
		Parameters: []ParamDef{{Name: "record", Doc: "Record with dotted names"}},
	},
	{
		Name: "network_of", Kind: KindFunction,
		Brief: "Get network from IP", Doc: "Get the network address from an IP and mask",
		Signature: "network_of(ip: ip, mask: net) -> net",
		Parameters: []ParamDef{{Name: "ip", Doc: "IP address"}, {Name: "mask", Doc: "Network mask"}},
	},
	{
		Name: "now", Kind: KindFunction,
		Brief: "Current timestamp", Doc: "Return the current timestamp",
		Signature: "now() -> time",
		Parameters: []ParamDef{},
	},
	{
		Name: "nullif", Kind: KindFunction,
		Brief: "Return null if equal", Doc: "Return null if two values are equal, otherwise return the first value",
		Signature: "nullif(a: any, b: any) -> any",
		Parameters: []ParamDef{{Name: "a", Doc: "First value"}, {Name: "b", Doc: "Value to compare"}},
	},
	{
		Name: "parse_sup", Kind: KindFunction,
		Brief: "Parse Super format", Doc: "Parse a string in Super format",
		Signature: "parse_sup(value: string) -> any",
		Parameters: []ParamDef{{Name: "value", Doc: "String to parse"}},
	},
	{
		Name: "parse_uri", Kind: KindFunction,
		Brief: "Parse URI string", Doc: "Parse a URI string into its components",
		Signature: "parse_uri(uri: string) -> record",
		Parameters: []ParamDef{{Name: "uri", Doc: "URI to parse"}},
	},
	{
		Name: "position", Kind: KindFunction,
		Brief: "Find substring position", Doc: "Find the position of a substring in a string",
		Signature: "position(substr: string, str: string) -> int64",
		Parameters: []ParamDef{{Name: "substr", Doc: "Substring to find"}, {Name: "str", Doc: "String to search"}},
	},
	{
		Name: "pow", Kind: KindFunction,
		Brief: "Power function", Doc: "Calculate base raised to the power of exponent",
		Signature: "pow(base: number, exp: number) -> number",
		Parameters: []ParamDef{{Name: "base", Doc: "Base value"}, {Name: "exp", Doc: "Exponent"}},
	},
	{
		Name: "quiet", Kind: KindFunction,
		Brief: "Suppress errors", Doc: "Suppress errors and return null instead",
		Signature: "quiet(value: any) -> any",
		Parameters: []ParamDef{{Name: "value", Doc: "Value to quiet"}},
	},
	{
		Name: "regexp", Kind: KindFunction,
		Brief: "Regular expression match", Doc: "Match a string against a regular expression",
		Signature: "regexp(pattern: string, value: string) -> bool",
		Parameters: []ParamDef{{Name: "pattern", Doc: "Regex pattern"}, {Name: "value", Doc: "String to match"}},
	},
	{
		Name: "regexp_replace", Kind: KindFunction,
		Brief: "Regex replacement", Doc: "Replace matches of a regex pattern",
		Signature: "regexp_replace(value: string, pattern: string, replacement: string) -> string",
		Parameters: []ParamDef{{Name: "value", Doc: "Input string"}, {Name: "pattern", Doc: "Regex pattern"}, {Name: "replacement", Doc: "Replacement string"}},
	},
	{
		Name: "replace", Kind: KindFunction,
		Brief: "String replacement", Doc: "Replace occurrences of a substring",
		Signature: "replace(value: string, old: string, new: string) -> string",
		Parameters: []ParamDef{{Name: "value", Doc: "Input string"}, {Name: "old", Doc: "String to replace"}, {Name: "new", Doc: "Replacement string"}},
	},
	{
		Name: "round", Kind: KindFunction,
		Brief: "Round to precision", Doc: "Round a number to a specified precision",
		Signature: "round(value: number, precision?: int64) -> number",
		Parameters: []ParamDef{{Name: "value", Doc: "Numeric value"}, {Name: "precision", Doc: "Decimal places (default: 0)"}},
	},
	{
		Name: "split", Kind: KindFunction,
		Brief: "Split string", Doc: "Split a string by a separator",
		Signature: "split(value: string, sep: string) -> [string]",
		Parameters: []ParamDef{{Name: "value", Doc: "String to split"}, {Name: "sep", Doc: "Separator"}},
	},
	{
		Name: "sqrt", Kind: KindFunction,
		Brief: "Square root", Doc: "Calculate the square root of a number",
		Signature: "sqrt(value: number) -> float64",
		Parameters: []ParamDef{{Name: "value", Doc: "Numeric value"}},
	},
	{
		Name: "strftime", Kind: KindFunction,
		Brief: "Format time as string", Doc: "Format a timestamp as a string using a format specifier",
		Signature: "strftime(format: string, time: time) -> string",
		Parameters: []ParamDef{{Name: "format", Doc: "Format string"}, {Name: "time", Doc: "Timestamp value"}},
	},
	{
		Name: "trim", Kind: KindFunction,
		Brief: "Trim whitespace", Doc: "Remove leading and trailing whitespace from a string",
		Signature: "trim(value: string) -> string",
		Parameters: []ParamDef{{Name: "value", Doc: "String to trim"}},
	},
	{
		Name: "typename", Kind: KindFunction,
		Brief: "Get type name", Doc: "Return the name of a value's type as a string",
		Signature: "typename(value: any) -> string",
		Parameters: []ParamDef{{Name: "value", Doc: "Value to check"}},
	},
	{
		Name: "typeof", Kind: KindFunction,
		Brief: "Get type of value", Doc: "Return the type of a value",
		Signature: "typeof(value: any) -> type",
		Parameters: []ParamDef{{Name: "value", Doc: "Value to check"}},
	},
	{
		Name: "under", Kind: KindFunction,
		Brief: "Get underlying value", Doc: "Unwrap a value to get its underlying representation",
		Signature: "under(value: any) -> any",
		Parameters: []ParamDef{{Name: "value", Doc: "Value to unwrap"}},
	},
	{
		Name: "unflatten", Kind: KindFunction,
		Brief: "Unflatten records", Doc: "Convert dotted field names back into nested records",
		Signature: "unflatten(record: record) -> record",
		Parameters: []ParamDef{{Name: "record", Doc: "Record to unflatten"}},
	},
	{
		Name: "upper", Kind: KindFunction,
		Brief: "Convert to uppercase", Doc: "Convert a string to uppercase",
		Signature: "upper(value: string) -> string",
		Parameters: []ParamDef{{Name: "value", Doc: "String to convert"}},
	},

	// Additional functions that need signatures
	{
		Name: "cast", Kind: KindFunction,
		Brief: "Cast value to type", Doc: "Convert a value to a specified type",
		Signature: "cast(value: any, type: type) -> any",
		Parameters: []ParamDef{{Name: "value", Doc: "Value to cast"}, {Name: "type", Doc: "Target type"}},
	},
	{
		Name: "error", Kind: KindFunction,
		Brief: "Create error value", Doc: "Create an error value with a message",
		Signature: "error(message: string) -> error",
		Parameters: []ParamDef{{Name: "message", Doc: "Error message"}},
	},
	{
		Name: "max", Kind: KindFunction,
		Brief: "Maximum of values", Doc: "Return the maximum of two values",
		Signature: "max(a: number, b: number) -> number",
		Parameters: []ParamDef{{Name: "a", Doc: "First value"}, {Name: "b", Doc: "Second value"}},
	},
	{
		Name: "min", Kind: KindFunction,
		Brief: "Minimum of values", Doc: "Return the minimum of two values",
		Signature: "min(a: number, b: number) -> number",
		Parameters: []ParamDef{{Name: "a", Doc: "First value"}, {Name: "b", Doc: "Second value"}},
	},

	// =========================================================================
	// AGGREGATES
	// =========================================================================

	{
		Name: "count", Kind: KindAggregate,
		Brief: "Count records", Doc: "Count the number of records in a group",
		Signature: "count() -> int64",
		Parameters: []ParamDef{},
	},
	{
		Name: "sum", Kind: KindAggregate,
		Brief: "Sum of values", Doc: "Calculate the sum of numeric values",
		Signature: "sum(value: number) -> number",
		Parameters: []ParamDef{{Name: "value", Doc: "Numeric values"}},
	},
	{
		Name: "avg", Kind: KindAggregate,
		Brief: "Average of values", Doc: "Calculate the average of numeric values",
		Signature: "avg(value: number) -> float64",
		Parameters: []ParamDef{{Name: "value", Doc: "Numeric values"}},
	},
	{
		Name: "collect", Kind: KindAggregate,
		Brief: "Collect values into array", Doc: "Collect all values into an array",
		Signature: "collect(value: any) -> [any]",
		Parameters: []ParamDef{{Name: "value", Doc: "Values to collect"}},
	},
	{
		Name: "collect_map", Kind: KindAggregate,
		Brief: "Collect into map", Doc: "Collect key-value pairs into a map",
		Signature: "collect_map(key: any, value: any) -> map",
		Parameters: []ParamDef{{Name: "key", Doc: "Map keys"}, {Name: "value", Doc: "Map values"}},
	},
	{
		Name: "dcount", Kind: KindAggregate,
		Brief: "Distinct count", Doc: "Count the number of distinct values",
		Signature: "dcount(value: any) -> int64",
		Parameters: []ParamDef{{Name: "value", Doc: "Values to count"}},
	},
	{
		Name: "any", Kind: KindAggregate,
		Brief: "Any value from group", Doc: "Return any arbitrary value from a group",
		Signature: "any(value: any) -> any",
		Parameters: []ParamDef{{Name: "value", Doc: "Values to choose from"}},
	},
	{
		Name: "union", Kind: KindAggregate,
		Brief: "Union of values", Doc: "Create a set union of all values",
		Signature: "union(value: any) -> set",
		Parameters: []ParamDef{{Name: "value", Doc: "Values to union"}},
	},
	{
		Name: "fuse", Kind: KindAggregate,
		Brief: "Fuse schemas in group", Doc: "Fuse schemas together within a group",
		Signature: "fuse(value: any) -> type",
		Parameters: []ParamDef{{Name: "value", Doc: "Values to fuse"}},
	},
	{
		Name: "and", Kind: KindAggregate,
		Brief: "Logical AND aggregate", Doc: "Returns true if all values in the group are true",
		Signature: "and(value: bool) -> bool",
		Parameters: []ParamDef{{Name: "value", Doc: "Boolean values"}},
	},
	{
		Name: "or", Kind: KindAggregate,
		Brief: "Logical OR aggregate", Doc: "Returns true if any value in the group is true",
		Signature: "or(value: bool) -> bool",
		Parameters: []ParamDef{{Name: "value", Doc: "Boolean values"}},
	},
	{
		Name: "first", Kind: KindAggregate,
		Brief: "First value in group", Doc: "Return the first value encountered in a group",
		Signature: "first(value: any) -> any",
		Parameters: []ParamDef{{Name: "value", Doc: "Values to select from"}},
	},
	{
		Name: "last", Kind: KindAggregate,
		Brief: "Last value in group", Doc: "Return the last value encountered in a group",
		Signature: "last(value: any) -> any",
		Parameters: []ParamDef{{Name: "value", Doc: "Values to select from"}},
	},
}

