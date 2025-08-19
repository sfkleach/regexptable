package regextable

import (
	"strings"
	"testing"
)

func TestRegexTableBuilder_Basic(t *testing.T) {
	builder := NewRegexTableBuilder[string]()

	// Chain pattern additions
	table, err := builder.
		AddPattern("hello", "greeting").
		AddPattern("world", "place").
		AddPattern(`\d+`, "number").
		Build(true, false) // Default anchoring: start but not end

	if err != nil {
		t.Fatalf("Failed to build table: %v", err)
	}

	// Test the built table
	testCases := []struct {
		input       string
		expected    string
		shouldMatch bool
	}{
		{"hello", "greeting", true},
		{"world", "place", true},
		{"123", "number", true},
		{"unknown", "", false},
	}

	for _, tc := range testCases {
		value, _, ok := table.TryLookup(tc.input)
		if tc.shouldMatch {
			if !ok {
				t.Errorf("Expected match for '%s'", tc.input)
			} else if value != tc.expected {
				t.Errorf("Expected '%s' for '%s', got '%s'", tc.expected, tc.input, value)
			}
		} else {
			if ok {
				t.Errorf("Expected no match for '%s', got '%s'", tc.input, value)
			}
		}
	}
}

func TestRegexTableBuilder_InvalidPattern(t *testing.T) {
	builder := NewRegexTableBuilder[string]()

	// Add valid and invalid patterns
	table, err := builder.
		AddPattern("valid", "good").
		AddPattern("[invalid", "bad"). // Invalid regex
		Build(true, false)

	if err == nil {
		t.Error("Expected build to fail with invalid pattern")
	}

	if table != nil {
		t.Error("Expected nil table when build fails")
	}

	// Error should mention the invalid pattern or compilation failure
	if !strings.Contains(err.Error(), "invalid pattern") && !strings.Contains(err.Error(), "failed to compile") {
		t.Errorf("Expected error to mention invalid pattern or compilation failure, got: %v", err)
	}
}

func TestRegexTableBuilder_MustBuild(t *testing.T) {
	// Test successful MustBuild
	builder := NewRegexTableBuilder[int]()
	table := builder.
		AddPattern("test", 42).
		MustBuild(true, false)

	value, _, ok := table.TryLookup("test")
	if !ok || value != 42 {
		t.Error("MustBuild should create working table")
	}

	// Test panic on invalid pattern
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected MustBuild to panic with invalid pattern")
		}
	}()

	invalidBuilder := NewRegexTableBuilder[int]()
	invalidBuilder.AddPattern("[invalid", 1).MustBuild(true, false) // Should panic
}

func TestRegexTableBuilder_Clone(t *testing.T) {
	original := NewRegexTableBuilder[string]()
	original.AddPattern("test1", "value1")
	original.AddPattern("test2", "value2")

	clone := original.Clone()

	// Add to clone and verify original is unchanged
	clone.AddPattern("test3", "value3")

	// Build both and verify they work correctly
	originalTable, err := original.Build(true, false)
	if err != nil {
		t.Fatalf("Failed to build original: %v", err)
	}

	cloneTable, err := clone.Build(true, false)
	if err != nil {
		t.Fatalf("Failed to build clone: %v", err)
	}

	// Original should not have test3
	_, _, ok := originalTable.TryLookup("test3")
	if ok {
		t.Error("Original table should not have test3 pattern")
	}

	// Clone should have test3
	value, _, ok := cloneTable.TryLookup("test3")
	if !ok || value != "value3" {
		t.Error("Clone table should have test3 pattern")
	}
}

func TestRegexTableBuilder_EmptyBuild(t *testing.T) {
	builder := NewRegexTableBuilder[string]()

	// Build empty table
	table, err := builder.Build(true, false)
	if err != nil {
		t.Fatalf("Building empty table should not fail: %v", err)
	}

	// Empty table should return error on lookup
	_, _, err = table.Lookup("anything")
	if err == nil {
		t.Error("Empty table should return error on lookup")
	}
}

func TestRegexTableBuilder_ReuseAfterBuild(t *testing.T) {
	builder := NewRegexTableBuilder[string]()

	// Build first table
	table1, err := builder.
		AddPattern("test1", "value1").
		Build(true, false)
	if err != nil {
		t.Fatalf("Failed to build first table: %v", err)
	}

	// Add more patterns and build second table
	table2, err := builder.
		AddPattern("test2", "value2").
		Build(true, false)
	if err != nil {
		t.Fatalf("Failed to build second table: %v", err)
	}

	// Both tables should work
	value1, _, ok := table1.TryLookup("test1")
	if !ok || value1 != "value1" {
		t.Error("First table should work")
	}

	// Second table should have both patterns
	value1, _, ok = table2.TryLookup("test1")
	if !ok || value1 != "value1" {
		t.Error("Second table should have first pattern")
	}

	value2, _, ok := table2.TryLookup("test2")
	if !ok || value2 != "value2" {
		t.Error("Second table should have second pattern")
	}
}

func TestRegexTableBuilder_Anchoring(t *testing.T) {
	// Test builder with default anchoring
	t.Run("DefaultAnchoring", func(t *testing.T) {
		table, err := NewRegexTableBuilder[string]().
			AddPattern("hello", "greeting").
			Build(true, false) // Start anchoring, no end anchoring

		if err != nil {
			t.Fatalf("Failed to build table: %v", err)
		}

		// Should match at start (default behavior)
		value, _, ok := table.TryLookup("hello world")
		if !ok || value != "greeting" {
			t.Error("Expected match for 'hello world' with default start anchoring")
		}

		// Should not match in middle
		_, _, ok = table.TryLookup("say hello")
		if ok {
			t.Error("Expected no match for 'say hello' with start anchoring")
		}
	})

	// Test builder with custom anchoring
	t.Run("CustomAnchoring", func(t *testing.T) {
		table, err := NewRegexTableBuilder[string]().
			AddPattern("world", "place").
			Build(false, true) // No start anchor, end anchor

		if err != nil {
			t.Fatalf("Failed to build table: %v", err)
		}

		// Should match at end
		value, _, ok := table.TryLookup("hello world")
		if !ok || value != "place" {
			t.Error("Expected match for 'hello world' with end anchoring")
		}

		// Should not match in middle
		_, _, ok = table.TryLookup("world peace")
		if ok {
			t.Error("Expected no match for 'world peace' with end anchoring")
		}
	})

	// Test builder with no anchoring
	t.Run("NoAnchoring", func(t *testing.T) {
		table, err := NewRegexTableBuilder[string]().
			AddPattern("test", "value").
			Build(false, false) // No anchoring

		if err != nil {
			t.Fatalf("Failed to build table: %v", err)
		}

		// Should match anywhere with no anchoring
		testCases := []string{"test", "testing", "pretest", "pretesting"}
		for _, input := range testCases {
			value, _, ok := table.TryLookup(input)
			if !ok || value != "value" {
				t.Errorf("Expected match for '%s' with no anchoring", input)
			}
		}
	})

	// Test builder with full anchoring
	t.Run("FullAnchoring", func(t *testing.T) {
		table, err := NewRegexTableBuilder[string]().
			AddPattern("exact", "match").
			Build(true, true) // Full anchoring

		if err != nil {
			t.Fatalf("Failed to build table: %v", err)
		}

		// Should only match exact string with full anchoring
		value, _, ok := table.TryLookup("exact")
		if !ok || value != "match" {
			t.Error("Expected exact match with full anchoring")
		}

		// Should not match with extra text
		_, _, ok = table.TryLookup("exact match")
		if ok {
			t.Error("Expected no match for 'exact match' with full anchoring")
		}
	})
}

func TestRegexTableBuilder_AddSubPatterns(t *testing.T) {
	// Test empty patterns slice
	t.Run("EmptyPatterns", func(t *testing.T) {
		builder := NewRegexTableBuilder[string]()

		// Add empty patterns slice - should do nothing
		table, err := builder.
			AddSubPatterns([]string{}, "empty").
			AddPattern("test", "value").
			Build(true, false)

		if err != nil {
			t.Fatalf("Failed to build table: %v", err)
		}

		// Should only match the "test" pattern, not the empty patterns
		value, _, ok := table.TryLookup("test")
		if !ok || value != "value" {
			t.Error("Expected match for 'test' pattern")
		}

		// Should not match anything that would match an empty pattern
		_, _, ok = table.TryLookup("empty")
		if ok {
			t.Error("Expected no match for empty patterns")
		}
	})

	// Test single pattern in slice
	t.Run("SinglePattern", func(t *testing.T) {
		builder := NewRegexTableBuilder[string]()

		table, err := builder.
			AddSubPatterns([]string{"hello"}, "greeting").
			Build(true, false)

		if err != nil {
			t.Fatalf("Failed to build table: %v", err)
		}

		value, _, ok := table.TryLookup("hello")
		if !ok || value != "greeting" {
			t.Error("Expected match for single pattern in AddSubPatterns")
		}
	})

	// Test multiple patterns creating alternation
	t.Run("MultiplePatterns", func(t *testing.T) {
		builder := NewRegexTableBuilder[string]()

		// Create alternation of greetings
		table, err := builder.
			AddSubPatterns([]string{"hello", "hi", "hey"}, "greeting").
			AddSubPatterns([]string{"bye", "goodbye", "farewell"}, "farewell").
			Build(true, false)

		if err != nil {
			t.Fatalf("Failed to build table: %v", err)
		}

		// Test all greeting patterns
		greetings := []string{"hello", "hi", "hey"}
		for _, greeting := range greetings {
			value, _, ok := table.TryLookup(greeting)
			if !ok || value != "greeting" {
				t.Errorf("Expected match for greeting '%s'", greeting)
			}
		}

		// Test all farewell patterns
		farewells := []string{"bye", "goodbye", "farewell"}
		for _, farewell := range farewells {
			value, _, ok := table.TryLookup(farewell)
			if !ok || value != "farewell" {
				t.Errorf("Expected match for farewell '%s'", farewell)
			}
		}

		// Test non-matching patterns
		_, _, ok := table.TryLookup("unknown")
		if ok {
			t.Error("Expected no match for 'unknown'")
		}
	})

	// Test complex patterns with regex syntax
	t.Run("ComplexPatterns", func(t *testing.T) {
		builder := NewRegexTableBuilder[string]()

		// Patterns with different regex features - creates (?:\d+|[a-z]+|[A-Z]+)
		// With start anchoring, this becomes ^(?:(?:\d+|[a-z]+|[A-Z]+))
		table, err := builder.
			AddSubPatterns([]string{`\d+`, `[a-z]+`, `[A-Z]+`}, "pattern_match").
			Build(true, false)

		if err != nil {
			t.Fatalf("Failed to build table: %v", err)
		}

		testCases := []struct {
			input    string
			expected bool
		}{
			{"123", true},    // matches \d+ at start
			{"abc", true},    // matches [a-z]+ at start
			{"ABC", true},    // matches [A-Z]+ at start
			{"123abc", true}, // matches \d+ at start (anchoring allows partial match)
			{"abc123", true}, // matches [a-z]+ at start
			{"", false},      // no match
			{"!abc", false},  // doesn't start with any of the patterns
		}

		for _, tc := range testCases {
			value, _, ok := table.TryLookup(tc.input)
			if tc.expected {
				if !ok || value != "pattern_match" {
					t.Errorf("Expected match for '%s'", tc.input)
				}
			} else {
				if ok {
					t.Errorf("Expected no match for '%s', got '%s'", tc.input, value)
				}
			}
		}
	})

	// Test patterns with special regex characters that need proper grouping
	t.Run("PatternsWithSpecialChars", func(t *testing.T) {
		builder := NewRegexTableBuilder[string]()

		// These patterns contain pipe characters, which will be treated as alternation
		// AddSubPatterns(["a|b", "c|d", "e"]) creates (?:a|b|c|d|e)
		// This means it will match "a", "b", "c", "d", or "e" - not the literal strings "a|b" or "c|d"
		table, err := builder.
			AddSubPatterns([]string{"a|b", "c|d", "e"}, "alternation_test").
			Build(true, false)

		if err != nil {
			t.Fatalf("Failed to build table: %v", err)
		}

		// The alternation (?:a|b|c|d|e) will match individual characters
		testCases := []struct {
			input    string
			expected bool
		}{
			{"a", true},   // matches "a" alternative
			{"b", true},   // matches "b" alternative
			{"c", true},   // matches "c" alternative
			{"d", true},   // matches "d" alternative
			{"e", true},   // matches "e" alternative
			{"a|b", true}, // matches "a" at start due to anchoring
			{"c|d", true}, // matches "c" at start due to anchoring
			{"f", false},  // not in alternation
			{"|a", false}, // doesn't start with any of a,b,c,d,e
		}

		for _, tc := range testCases {
			value, _, ok := table.TryLookup(tc.input)
			if tc.expected {
				if !ok || value != "alternation_test" {
					t.Errorf("Expected match for '%s'", tc.input)
				}
			} else {
				if ok {
					t.Errorf("Expected no match for '%s', got '%s'", tc.input, value)
				}
			}
		}
	})

	// Test patterns with literal special characters (properly escaped)
	t.Run("LiteralSpecialChars", func(t *testing.T) {
		builder := NewRegexTableBuilder[string]()

		// To match literal pipe characters, they need to be escaped
		// AddSubPatterns(["a\\|b", "c\\|d", "e"]) creates (?:a\|b|c\|d|e)
		table, err := builder.
			AddSubPatterns([]string{`a\|b`, `c\|d`, "e"}, "literal_test").
			Build(true, false)

		if err != nil {
			t.Fatalf("Failed to build table: %v", err)
		}

		testCases := []struct {
			input    string
			expected bool
		}{
			{"a|b", true}, // matches literal "a|b"
			{"c|d", true}, // matches literal "c|d"
			{"e", true},   // matches "e"
			{"a", false},  // doesn't match just "a"
			{"b", false},  // doesn't match just "b"
			{"c", false},  // doesn't match just "c"
			{"d", false},  // doesn't match just "d"
		}

		for _, tc := range testCases {
			value, _, ok := table.TryLookup(tc.input)
			if tc.expected {
				if !ok || value != "literal_test" {
					t.Errorf("Expected match for '%s'", tc.input)
				}
			} else {
				if ok {
					t.Errorf("Expected no match for '%s', got '%s'", tc.input, value)
				}
			}
		}
	})

	// Test mixing AddSubPatterns with regular AddPattern
	t.Run("MixWithAddPattern", func(t *testing.T) {
		builder := NewRegexTableBuilder[string]()

		table, err := builder.
			AddPattern("single", "single_value").
			AddSubPatterns([]string{"multi1", "multi2"}, "multi_value").
			AddPattern("another", "another_value").
			Build(true, false)

		if err != nil {
			t.Fatalf("Failed to build table: %v", err)
		}

		// Test all patterns work correctly
		testCases := []struct {
			input    string
			expected string
		}{
			{"single", "single_value"},
			{"multi1", "multi_value"},
			{"multi2", "multi_value"},
			{"another", "another_value"},
		}

		for _, tc := range testCases {
			value, _, ok := table.TryLookup(tc.input)
			if !ok || value != tc.expected {
				t.Errorf("Expected '%s' for input '%s', got '%s' (ok=%v)", tc.expected, tc.input, value, ok)
			}
		}
	})

	// Test with different value types
	t.Run("DifferentValueTypes", func(t *testing.T) {
		// Test with integer values
		intBuilder := NewRegexTableBuilder[int]()

		intTable, err := intBuilder.
			AddSubPatterns([]string{"one", "1"}, 1).
			AddSubPatterns([]string{"two", "2"}, 2).
			Build(true, false)

		if err != nil {
			t.Fatalf("Failed to build int table: %v", err)
		}

		value, _, ok := intTable.TryLookup("one")
		if !ok || value != 1 {
			t.Error("Expected integer value 1 for 'one'")
		}

		value, _, ok = intTable.TryLookup("2")
		if !ok || value != 2 {
			t.Error("Expected integer value 2 for '2'")
		}
	})

	// Test invalid patterns in AddSubPatterns
	t.Run("InvalidPatterns", func(t *testing.T) {
		builder := NewRegexTableBuilder[string]()

		// Include an invalid regex pattern
		_, err := builder.
			AddSubPatterns([]string{"valid", "[invalid"}, "test").
			Build(true, false)

		if err == nil {
			t.Error("Expected build to fail with invalid pattern in AddSubPatterns")
		}

		// Error should mention the invalid pattern or compilation failure
		if !strings.Contains(err.Error(), "invalid pattern") && !strings.Contains(err.Error(), "failed to compile") {
			t.Errorf("Expected error to mention invalid pattern or compilation failure, got: %v", err)
		}
	})
}

func TestRegexTableBuilder_FluentSubPatterns(t *testing.T) {
	// Test the fluent interface for building sub-patterns
	t.Run("BasicFluentInterface", func(t *testing.T) {
		builder := NewRegexTableBuilder[string]()

		table, err := builder.
			BeginAddSubPatterns().
			AddSubPattern("hello").
			AddSubPattern("hi").
			AddSubPattern("hey").
			EndAddSubPatterns("greeting").
			BeginAddSubPatterns().
			AddSubPattern("bye").
			AddSubPattern("goodbye").
			EndAddSubPatterns("farewell").
			Build(true, false)

		if err != nil {
			t.Fatalf("Failed to build table: %v", err)
		}

		// Test greeting patterns
		greetings := []string{"hello", "hi", "hey"}
		for _, greeting := range greetings {
			value, _, ok := table.TryLookup(greeting)
			if !ok || value != "greeting" {
				t.Errorf("Expected match for greeting '%s'", greeting)
			}
		}

		// Test farewell patterns
		farewells := []string{"bye", "goodbye"}
		for _, farewell := range farewells {
			value, _, ok := table.TryLookup(farewell)
			if !ok || value != "farewell" {
				t.Errorf("Expected match for farewell '%s'", farewell)
			}
		}
	})

	// Test empty fluent sub-patterns
	t.Run("EmptyFluentSubPatterns", func(t *testing.T) {
		builder := NewRegexTableBuilder[string]()

		table, err := builder.
			BeginAddSubPatterns().
			EndAddSubPatterns("empty").
			AddPattern("test", "value").
			Build(true, false)

		if err != nil {
			t.Fatalf("Failed to build table: %v", err)
		}

		// Should only match the "test" pattern
		value, _, ok := table.TryLookup("test")
		if !ok || value != "value" {
			t.Error("Expected match for 'test' pattern")
		}
	})

	// Test single pattern in fluent interface
	t.Run("SingleFluentPattern", func(t *testing.T) {
		builder := NewRegexTableBuilder[string]()

		table, err := builder.
			BeginAddSubPatterns().
			AddSubPattern("hello").
			EndAddSubPatterns("greeting").
			Build(true, false)

		if err != nil {
			t.Fatalf("Failed to build table: %v", err)
		}

		value, _, ok := table.TryLookup("hello")
		if !ok || value != "greeting" {
			t.Error("Expected match for single fluent pattern")
		}
	})

	// Test mixing fluent interface with regular methods
	t.Run("MixFluentWithRegular", func(t *testing.T) {
		builder := NewRegexTableBuilder[string]()

		table, err := builder.
			AddPattern("single", "single_value").
			BeginAddSubPatterns().
			AddSubPattern("multi1").
			AddSubPattern("multi2").
			EndAddSubPatterns("multi_value").
			AddSubPatterns([]string{"direct1", "direct2"}, "direct_value").
			Build(true, false)

		if err != nil {
			t.Fatalf("Failed to build table: %v", err)
		}

		testCases := []struct {
			input    string
			expected string
		}{
			{"single", "single_value"},
			{"multi1", "multi_value"},
			{"multi2", "multi_value"},
			{"direct1", "direct_value"},
			{"direct2", "direct_value"},
		}

		for _, tc := range testCases {
			value, _, ok := table.TryLookup(tc.input)
			if !ok || value != tc.expected {
				t.Errorf("Expected '%s' for input '%s', got '%s' (ok=%v)", tc.expected, tc.input, value, ok)
			}
		}
	})

	// Test complex patterns in fluent interface
	t.Run("ComplexFluentPatterns", func(t *testing.T) {
		builder := NewRegexTableBuilder[int]()

		table, err := builder.
			BeginAddSubPatterns().
			AddSubPattern(`\d+`).
			AddSubPattern(`[a-z]+`).
			AddSubPattern(`[A-Z]+`).
			EndAddSubPatterns(42).
			Build(false, false) // No anchoring for this test

		if err != nil {
			t.Fatalf("Failed to build table: %v", err)
		}

		testCases := []struct {
			input    string
			expected bool
		}{
			{"123", true},
			{"abc", true},
			{"ABC", true},
			{"mixed123", true}, // No anchoring, should match \d+
			{"", false},
		}

		for _, tc := range testCases {
			value, _, ok := table.TryLookup(tc.input)
			if tc.expected {
				if !ok || value != 42 {
					t.Errorf("Expected match for '%s'", tc.input)
				}
			} else {
				if ok {
					t.Errorf("Expected no match for '%s', got %d", tc.input, value)
				}
			}
		}
	})
}

func TestRegexTableBuilder_Clear(t *testing.T) {
	// Build a table with some patterns
	builder := NewRegexTableBuilder[string]().
		AddPattern("hello", "greeting").
		AddPattern("world", "place").
		AddPattern(`\d+`, "number")

	// Verify patterns were added
	table, err := builder.Build(true, false)
	if err != nil {
		t.Fatalf("Failed to build initial table: %v", err)
	}

	// Verify the table works
	if value, _, ok := table.TryLookup("hello"); !ok || value != "greeting" {
		t.Errorf("Initial table should match 'hello'")
	}

	// Clear the builder
	builder.Clear()

	// Build new table after clear - should be empty
	emptyTable, err := builder.Build(true, false)
	if err != nil {
		t.Fatalf("Failed to build table after clear: %v", err)
	}

	// Verify the cleared table doesn't match anything
	if _, _, ok := emptyTable.TryLookup("hello"); ok {
		t.Errorf("Cleared table should not match 'hello'")
	}
	if _, _, ok := emptyTable.TryLookup("world"); ok {
		t.Errorf("Cleared table should not match 'world'")
	}
	if _, _, ok := emptyTable.TryLookup("123"); ok {
		t.Errorf("Cleared table should not match '123'")
	}

	// Verify we can add new patterns after clear
	newTable, err := builder.
		AddPattern("foo", "bar").
		Build(true, false)
	if err != nil {
		t.Fatalf("Failed to build new table after clear: %v", err)
	}

	// Verify the new table works
	if value, _, ok := newTable.TryLookup("foo"); !ok || value != "bar" {
		t.Errorf("New table should match 'foo' -> 'bar'")
	}

	// Verify old patterns are still gone
	if _, _, ok := newTable.TryLookup("hello"); ok {
		t.Errorf("New table should not match old pattern 'hello'")
	}
}
