package regexptable

// RegexpEngine defines the minimal interface needed by RegexpTable for regexp operations.
// This abstraction allows different regexp engines to be used with RegexpTable.
type RegexpEngine interface {

	// Compile compiles a regexp pattern and returns a CompiledRegexp or an error.
	Compile(pattern string) (CompiledRegexp, error)

	// FormatNamedGroup formats a pattern with a named capture group using the engine's syntax.
	// For example: Go uses (?P<name>pattern), .NET uses (?<name>pattern), etc.
	FormatNamedGroup(groupName, pattern string) string
}

// CompiledRegexp represents a compiled regexp pattern that can perform matches.
// This interface abstracts over different regexp engine implementations, allowing
// RegexpTable to work with any regexp engine (Go's standard regexp, regexp2, etc.)
// by wrapping their compiled regexp objects to provide a uniform interface.
type CompiledRegexp interface {

	// FindStringSubmatch finds the first match and returns all submatches.
	// Returns nil if no match is found.
	// The first element is the full match, subsequent elements are capture groups
	// (including empty strings for groups that didn't match). The returned slice
	// has the same length as SubexpNames() and corresponds 1:1 with its elements.
	// Each capture group corresponds to an open parenthesis in the regexp pattern
	// (excluding only escaped parentheses like \().
	FindStringSubmatch(s string) []string

	// SubexpNames returns the names of the capturing groups.
	// This method behaves like Go's regexp.SubexpNames(): it returns a slice of strings
	// whose length equals the number of capture groups (including non-capturing groups)
	// in the regexp. Each entry that corresponds to a named capture group contains the
	// name of that group; all other entries are the empty string.
	//
	// The first element (index 0) always corresponds to the entire match and is always
	// the empty string. Subsequent elements correspond to capture groups in the order
	// they appear in the pattern.
	//
	// For example, given the pattern `(?P<name>\w+):(?P<value>\d+)(\s*)`:
	// - Index 0: "" (full match)
	// - Index 1: "name" (first named group)
	// - Index 2: "value" (second named group)
	// - Index 3: "" (third group, unnamed)
	SubexpNames() []string
}
