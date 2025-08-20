# Case-study: integrating with regexp2

To show how to integrate alternative regexp engines into regexptable, this
case-study demonstrates how to integrate Doug Clark's `regexp2` library. The
`regexp2` library provides .NET-compatible regular expressions with advanced
features like lookbehind, named groups, and other functionality not available in
Go's standard `regexp` package.


## Installation

Add the regexp2 dependency to your application:

```bash
go get github.com/dlclark/regexp2
```

## Implementation

To use regexp2 with regexptable, implement the `regexpEngine` and `Compiledregexp` interfaces:

```go
package main

import (
	"fmt"
	"github.com/dlclark/regexp2"
	"github.com/sfkleach/regexptable"
)

// Regexp2Engine wraps regexp2 to implement regexptable.regexpEngine
// 
// regexpEngine interface requires:
// - Compile(pattern string) (Compiledregexp, error)
// - FormatNamedGroup(groupName, pattern string) string
type Regexp2Engine struct {
	options regexp2.RegexOptions
}

func NewRegexp2Engine() *Regexp2Engine {
	return &Regexp2Engine{options: regexp2.None}
}

// Compile compiles a regexp pattern using regexp2
func (e *Regexp2Engine) Compile(pattern string) (regexptable.CompiledRegexp, error) {
	compiled, err := regexp2.Compile(pattern, e.options)
	if err != nil {
		return nil, err
	}
	return &Regexp2CompiledRegexp{regexp: compiled}, nil
}

// FormatNamedGroup formats a pattern with a named capture group using .NET syntax
func (e *Regexp2Engine) FormatNamedGroup(groupName, pattern string) string {
	return fmt.Sprintf("(?<%s>%s)", groupName, pattern) // .NET syntax: (?<name>pattern)
}

// Regexp2CompiledRegexp wraps regexp2.Regexp to implement regexptable.CompiledRegex
//
// Compiledregexp interface requires:
// - FindStringSubmatch(s string) []string
// - SubexpNames() []string
type Regexp2CompiledRegexp struct {
	regexp *regexp2.Regexp
}

// FindStringSubmatch finds the first match and returns all submatches
// Adapts regexp2's (*Match, error) API to Go's []string format
func (r *Regexp2CompiledRegexp) FindStringSubmatch(s string) []string {
	match, err := r.regexp.FindStringMatch(s)
	if err != nil || match == nil {
		return nil
	}

	groups := match.Groups()
	result := make([]string, len(groups))
	for i, group := range groups {
		if group.Length > 0 {
			result[i] = group.String()
		}
	}
	return result
}

// SubexpNames returns the names of capturing groups
// regexp2 doesn't expose group names the same way as Go's regexp,
// but regexptable generates its own names, so we return empty strings
func (r *Regexp2CompiledRegexp) SubexpNames() []string {
	groupCount := r.regexp.GetGroupNumbers()
	names := make([]string, len(groupCount))
	return names // Return empty names - regexptable generates its own
}
```

## Usage

Use your regexp2 engine with regexptable:

```go
func main() {
	engine := NewRegexp2Engine()

	table, err := regexptable.NewRegexpTableBuilderWithEngine[string](engine).
		AddPattern(`(?<=\bclass\s+)\w+`, "class_name"). // Lookbehind
		AddPattern(`\w+(?=\s*\()`, "function_name").    // Lookahead
		AddPattern(`\b\w+\b`, "identifier").
		Build(true, false) // Start anchoring, no end anchoring

	if err != nil {
		panic(err)
	}

	// Test with advanced regexp2 features
	if value, _, ok := table.TryLookup("MyClass"); ok {
		fmt.Printf("Matched: %s\n", value)
	}
}
```

## Key Differences

The main API differences to handle:

- **Matching**: regexp2 returns `(*Match, error)` vs Go's `[]string`
- **Error handling**: regexp2 can fail during matching, not just compilation
- **Group names**: regexp2 doesn't expose group names like Go's regexp
- **Syntax**: regexp2 uses .NET-style named groups `(?<name>pattern)`
