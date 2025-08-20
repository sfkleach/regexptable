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
type CompiledRegexp interface {
	// FindStringSubmatch finds the first match and returns all submatches.
	// Returns nil if no match is found.
	// The first element is the full match, subsequent elements are capture groups.
	FindStringSubmatch(s string) []string

	// SubexpNames returns the names of the capturing groups.
	// The first element corresponds to the entire match and is empty.
	SubexpNames() []string
}
