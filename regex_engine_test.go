package regexptable

import (
	"fmt"
	"testing"
)

// MockRegexpEngine implements RegexpEngine for testing with different naming conventions.
type MockRegexpEngine struct {
	compiledRegexps map[string]*MockCompiledRegexp
	groupSyntax     string // e.g., "(?P<%s>%s)" for Go, "(?<%s>%s)" for .NET
}

// NewMockRegexpEngine creates a new mock engine with the specified group syntax.
func NewMockRegexpEngine(groupSyntax string) *MockRegexpEngine {
	return &MockRegexpEngine{
		compiledRegexps: make(map[string]*MockCompiledRegexp),
		groupSyntax:     groupSyntax,
	}
}

// Compile returns a pre-configured mock or creates a simple one.
func (e *MockRegexpEngine) Compile(pattern string) (CompiledRegexp, error) {
	if compiled, exists := e.compiledRegexps[pattern]; exists {
		return compiled, nil
	}
	// Return a simple mock that doesn't actually match anything
	return &MockCompiledRegexp{pattern: pattern}, nil
}

// FormatNamedGroup uses the configured group syntax.
func (e *MockRegexpEngine) FormatNamedGroup(groupName, pattern string) string {
	return fmt.Sprintf(e.groupSyntax, groupName, pattern)
}

// SetCompiledRegexp allows tests to configure what a pattern should return.
func (e *MockRegexpEngine) SetCompiledRegexp(pattern string, compiled *MockCompiledRegexp) {
	e.compiledRegexps[pattern] = compiled
}

// MockCompiledRegexp implements CompiledRegexp for testing.
type MockCompiledRegexp struct {
	pattern     string
	matchResult []string
	subexpNames []string
	shouldMatch bool
}

// SetMatchResult configures what FindStringSubmatch should return.
func (r *MockCompiledRegexp) SetMatchResult(matches []string, subexpNames []string) {
	r.matchResult = matches
	r.subexpNames = subexpNames
	r.shouldMatch = matches != nil
}

// FindStringSubmatch returns the configured match result.
func (r *MockCompiledRegexp) FindStringSubmatch(s string) []string {
	if r.shouldMatch {
		return r.matchResult
	}
	return nil
}

// SubexpNames returns the configured subexpression names.
func (r *MockCompiledRegexp) SubexpNames() []string {
	return r.subexpNames
}

func TestRegexpEngine_DifferentNamingConventions(t *testing.T) {
	testCases := []struct {
		name        string
		groupSyntax string
		expected    string
	}{
		{
			name:        "Go/PCRE syntax",
			groupSyntax: "(?P<%s>%s)",
			expected:    "(?P<test_group>hello)",
		},
		{
			name:        ".NET syntax",
			groupSyntax: "(?<%s>%s)",
			expected:    "(?<test_group>hello)",
		},
		{
			name:        "Custom syntax",
			groupSyntax: "(?'%s'%s)",
			expected:    "(?'test_group'hello)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			engine := NewMockRegexpEngine(tc.groupSyntax)
			result := engine.FormatNamedGroup("test_group", "hello")
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestRegexpTable_WithCustomEngine(t *testing.T) {
	// Create a mock engine with .NET-style syntax
	mockEngine := NewMockRegexpEngine("(?<%s>%s)")

	// Create a mock compiled regexp that will match "hello"
	mockCompiled := &MockCompiledRegexp{}
	mockCompiled.SetMatchResult(
		[]string{"hello", "hello"},        // full match and capture group
		[]string{"", "__REGEXPTABLE_1__"}, // subexp names
	)

	// Configure the engine to return our mock when compiling the union pattern
	// This is a bit artificial but demonstrates the concept
	table := NewRegexpTableWithEngine[string](mockEngine, true, false)

	// Add a pattern - this should use the .NET syntax
	err := table.AddPattern("hello", "greeting")
	if err != nil {
		t.Fatalf("Failed to add pattern: %v", err)
	}

	// Verify the pattern was formatted with .NET syntax
	expectedPattern := "(?<__REGEXPTABLE_1__>hello)"
	if len(table.patternNames) != 1 || table.patternNames[0] != expectedPattern {
		t.Errorf("Expected pattern %q, got %q", expectedPattern, table.patternNames[0])
	}
}

func TestStandardRegexpEngine_FormatNamedGroup(t *testing.T) {
	engine := NewStandardRegexpEngine()
	result := engine.FormatNamedGroup("testgroup", "pattern")
	expected := "(?P<testgroup>pattern)"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
