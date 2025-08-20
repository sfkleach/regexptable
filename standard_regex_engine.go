package regexptable

import (
	"fmt"
	"regexp"
)

// StandardRegexpEngine implements RegexpEngine using Go's built-in regexp package.
type StandardRegexpEngine struct{}

// NewStandardRegexpEngine creates a new StandardRegexpEngine.
func NewStandardRegexpEngine() *StandardRegexpEngine {
	return &StandardRegexpEngine{}
}

// Compile compiles a regexp pattern using Go's regexp.Compile.
func (e *StandardRegexpEngine) Compile(pattern string) (CompiledRegexp, error) {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return NewStandardCompiledRegexp(compiled), nil
}

// FormatNamedGroup formats a named capture group using Go's (?P<name>pattern) syntax.
func (e *StandardRegexpEngine) FormatNamedGroup(groupName, pattern string) string {
	return fmt.Sprintf("(?P<%s>%s)", groupName, pattern)
}

// StandardCompiledRegexp wraps a Go *regexp.Regexp to implement CompiledRegexp.
type StandardCompiledRegexp struct {
	regexp *regexp.Regexp
}

// NewStandardCompiledRegexp creates a new StandardCompiledRegexp wrapping the given regexp.
func NewStandardCompiledRegexp(regexp *regexp.Regexp) *StandardCompiledRegexp {
	return &StandardCompiledRegexp{regexp: regexp}
}

// FindStringSubmatch delegates to the wrapped regexp.
func (r *StandardCompiledRegexp) FindStringSubmatch(s string) []string {
	return r.regexp.FindStringSubmatch(s)
}

// SubexpNames delegates to the wrapped regexp.
func (r *StandardCompiledRegexp) SubexpNames() []string {
	return r.regexp.SubexpNames()
}
