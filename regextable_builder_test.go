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
