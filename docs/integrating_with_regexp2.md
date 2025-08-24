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

To use regexp2 with regexptable, implement the `RegexpEngine` and `CompiledRegexp` interfaces:

```go
package main

import (
	"fmt"

	"github.com/dlclark/regexp2"
	"github.com/sfkleach/regexptable"
)

// Regexp2Engine implements regexptable.RegexpEngine using the regexp2 package for .NET-compatible regular expressions.
// This provides advanced features like lookbehind, named groups, and other functionality not available
// in Go's standard regexp package.
type Regexp2Engine struct {
	options regexp2.RegexOptions
}

// NewRegexp2Engine creates a new Regexp2Engine with default options.
func NewRegexp2Engine() *Regexp2Engine {
	return &Regexp2Engine{options: regexp2.None}
}

// NewRegexp2EngineWithOptions creates a new Regexp2Engine with custom options.
func NewRegexp2EngineWithOptions(options regexp2.RegexOptions) *Regexp2Engine {
	return &Regexp2Engine{options: options}
}

// Compile compiles a regexp pattern using regexp2.Compile.
func (e *Regexp2Engine) Compile(pattern string) (regexptable.CompiledRegexp, error) {
	compiled, err := regexp2.Compile(pattern, e.options)
	if err != nil {
		return nil, err
	}
	return NewRegexp2CompiledRegexp(compiled), nil
}

// FormatNamedGroup formats a named capture group using .NET's (?<name>pattern) syntax.
func (e *Regexp2Engine) FormatNamedGroup(groupName, pattern string) string {
	return fmt.Sprintf("(?<%s>%s)", groupName, pattern)
}

// Regexp2CompiledRegexp wraps a regexp2.Regexp to implement regexptable.CompiledRegexp.
type Regexp2CompiledRegexp struct {
	regexp *regexp2.Regexp
}

// NewRegexp2CompiledRegexp creates a new Regexp2CompiledRegexp wrapping the given regexp2.Regexp.
func NewRegexp2CompiledRegexp(regexp *regexp2.Regexp) *Regexp2CompiledRegexp {
	return &Regexp2CompiledRegexp{regexp: regexp}
}

// FindStringSubmatch finds the first match and returns all submatches.
// Adapts regexp2's (*Match, error) API to Go's []string format for compatibility.
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
		// Note: Empty groups are left as empty strings (zero value)
	}
	return result
}

// SubexpNames returns the names of the capturing groups.
// Uses regexp2's GetGroupNames() to get the actual group names,
// which is essential for regexptable to determine which named groups matched.
func (r *Regexp2CompiledRegexp) SubexpNames() []string {
	// Get the group names from regexp2
	groupNames := r.regexp.GetGroupNames()
	
	// The first element should be empty (like Go's regexp) for the full match
	result := make([]string, len(groupNames))
	for i, name := range groupNames {
		if i == 0 {
			// Group 0 is the full match and should have no name
			result[i] = ""
		} else {
			result[i] = name
		}
	}
	return result
}
```

## Testing the Implementation

Here's a complete test that verifies the SubexpNames functionality works correctly:

```go
package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dlclark/regexp2"
	"github.com/sfkleach/regexptable"
)

func TestRegexp2SubexpNames(t *testing.T) {
	// Create the regexp2 engine
	engine := NewRegexp2Engine()
	
	// Test that our SubexpNames implementation works correctly
	// by creating a regexp with named groups and checking the names
	pattern := `(?<__REGEXPTABLE_1__>\d+)|(?<__REGEXPTABLE_2__>[a-zA-Z]+)`
	compiled, err := engine.Compile(pattern)
	if err != nil {
		t.Fatalf("Failed to compile pattern: %v", err)
	}
	
	// Get the subexp names
	names := compiled.SubexpNames()
	
	// Should have: ["", "__REGEXPTABLE_1__", "__REGEXPTABLE_2__"]
	expectedNames := []string{"", "__REGEXPTABLE_1__", "__REGEXPTABLE_2__"}
	
	if len(names) != len(expectedNames) {
		t.Fatalf("Expected %d names, got %d", len(expectedNames), len(names))
	}
	
	for i, expected := range expectedNames {
		if names[i] != expected {
			t.Errorf("Expected name[%d] = %q, got %q", i, expected, names[i])
		}
	}
	
	// Test that matches work correctly with named groups
	matches := compiled.FindStringSubmatch("123")
	if matches == nil {
		t.Fatal("Expected match for '123'")
	}
	
	// Should have: ["123", "123", ""]
	// Group 0: full match, Group 1: digit match, Group 2: empty (didn't match letters)
	if len(matches) != 3 {
		t.Fatalf("Expected 3 match groups, got %d", len(matches))
	}
	
	if matches[0] != "123" || matches[1] != "123" || matches[2] != "" {
		t.Errorf("Unexpected matches: %v", matches)
	}
}

func TestRegexp2WithRegexpTable(t *testing.T) {
	// Test the full integration with regexptable
	engine := NewRegexp2Engine()
	
	table, err := regexptable.NewRegexpTableBuilderWithEngine[string](engine).
		AddPattern(`\d+`, "number").
		AddPattern(`[a-zA-Z]+`, "word").
		Build(true, false) // Anchor to start
	
	if err != nil {
		t.Fatalf("Failed to build table: %v", err)
	}
	
	// Test number matching
	value, matches, ok := table.TryLookup("123")
	if !ok {
		t.Fatal("Expected match for '123'")
	}
	if value != "number" {
		t.Errorf("Expected 'number', got %q", value)
	}
	if len(matches) != 1 || matches[0] != "123" {
		t.Errorf("Expected matches ['123'], got %v", matches)
	}
	
	// Test word matching  
	value, matches, ok = table.TryLookup("hello")
	if !ok {
		t.Fatal("Expected match for 'hello'")
	}
	if value != "word" {
		t.Errorf("Expected 'word', got %q", value)
	}
	if len(matches) != 1 || matches[0] != "hello" {
		t.Errorf("Expected matches ['hello'], got %v", matches)
	}
}
```

## Troubleshooting

### Common Issues

1. **Empty SubexpNames**: If `SubexpNames()` returns empty strings, regexptable cannot determine which pattern matched. Ensure you're using `regexp.GetGroupNames()` correctly.

2. **Group numbering mismatch**: regexp2 group numbering might differ from Go's regexp. Always use `GroupNameFromNumber()` to map between numbers and names.

3. **Named group syntax**: Use .NET-style `(?<name>pattern)` syntax, not Go's `(?P<name>pattern)`.

4. **Match vs. FindStringSubmatch**: The regexptable interface expects `[]string` from `FindStringSubmatch`, but regexp2's native API returns `*Match`. The adapter must convert between these formats.

## Usage

Use your custom regexp2 engine with regexptable:

```go
package main

import (
	"fmt"

	"github.com/dlclark/regexp2"
	"github.com/sfkleach/regexptable"
)

func main() {
	// Create your custom regexp2 engine
	engine := NewRegexp2Engine()

	table, err := regexptable.NewRegexpTableBuilderWithEngine[string](engine).
		AddPattern(`(?<=\bclass\s+)\w+`, "class_name"). // Lookbehind - more specific, goes first
		AddPattern(`\w+(?=\s*\()`, "function_name").    // Lookahead
		AddPattern(`\b\w+\b`, "identifier").            // General word - less specific, goes last
		Build(false, false) // No anchoring to allow lookbehind to work properly

	if err != nil {
		panic(err)
	}

	// Test with advanced regexp2 features
	if value, _, ok := table.TryLookup("class MyClass"); ok {
		fmt.Printf("Matched: %s\n", value) // Output: Matched: class_name
	}
	
	// To see the lookbehind in action with match details
	if value, matches, ok := table.TryLookup("class MyClass extends"); ok {
		fmt.Printf("Matched '%s' as: %s\n", matches[0], value) // Output: Matched 'MyClass' as: class_name
	}

	// Example with case-insensitive matching
	caseInsensitiveEngine := NewRegexp2EngineWithOptions(regexp2.IgnoreCase)
	
	caseTable, err := regexptable.NewRegexpTableBuilderWithEngine[string](caseInsensitiveEngine).
		AddPattern(`hello`, "greeting").
		Build(true, false)
	
	if err != nil {
		panic(err)
	}
	
	// Both will match due to case-insensitive option
	if value, _, ok := caseTable.TryLookup("HELLO"); ok {
		fmt.Printf("Matched: %s\n", value) // Output: Matched: greeting
	}
}
```

## Verifying Named Subgroup Matches

To verify which named subgroup matched with regexp2, you can use several approaches:

### Method 1: Using Group Names from Matches

```go
// Get a match and check which groups captured text
match, err := regexp.FindStringMatch("test string")
if err != nil || match == nil {
    return
}

// Iterate through all groups to see which ones matched
groups := match.Groups()
for i, group := range groups {
    if group.Length > 0 {
        groupName := regexp.GroupNameFromNumber(i)
        fmt.Printf("Group %d (%s) matched: %s\n", i, groupName, group.String())
    }
}
```

### Method 2: Direct Group Lookup by Name

```go
// Look up a specific group by name
match, err := regexp.FindStringMatch("test string")
if err != nil || match == nil {
    return
}

// Check if a specific named group matched
if group := match.GroupByName("__REGEXPTABLE_1__"); group != nil && group.Length > 0 {
    fmt.Printf("Group '__REGEXPTABLE_1__' matched: %s\n", group.String())
}
```

### Method 3: Get All Group Names

```go
// Get all group names from the compiled regexp
groupNames := regexp.GetGroupNames()
for i, name := range groupNames {
    fmt.Printf("Group %d has name: %s\n", i, name)
}
```

## Key Differences

The main API differences to handle:

- **Matching**: regexp2 returns `(*Match, error)` vs Go's `[]string`
- **Error handling**: regexp2 can fail during matching, not just compilation
- **Group names**: regexp2 provides `GetGroupNames()`, `GroupNameFromNumber()`, etc. for group introspection
- **Syntax**: regexp2 uses .NET-style named groups `(?<name>pattern)` vs Go's `(?P<name>pattern)`
- **Group access**: regexp2 provides rich `Match.GroupByName()` and `Match.Groups()` methods
