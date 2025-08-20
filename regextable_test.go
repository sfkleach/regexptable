package regexptable

import (
	"testing"
)

// TokenType represents different types of tokens for testing
type TokenType int

const (
	TokenFormStart TokenType = iota
	TokenFormEnd
	TokenSimpleLabel
	TokenVariable
)

func TestRegexpTable_Basic(t *testing.T) {
	table := NewRegexpTable[TokenType](true, false) // Start anchoring, no end anchoring

	// Add some test patterns using deferred compilation
	err := table.AddPattern(`form\w*`, TokenFormStart)
	if err != nil {
		t.Fatalf("Failed to add form_start pattern: %v", err)
	}

	err = table.AddPattern(`end\w*`, TokenFormEnd)
	if err != nil {
		t.Fatalf("Failed to add form_end pattern: %v", err)
	}

	err = table.AddPattern(`[a-z]+:`, TokenSimpleLabel)
	if err != nil {
		t.Fatalf("Failed to add simple_label pattern: %v", err)
	}

	// Test successful matches (compilation happens here)
	testCases := []struct {
		input       string
		expected    TokenType
		shouldMatch bool
	}{
		{"form", TokenFormStart, true},
		{"formData", TokenFormStart, true},
		{"endform", TokenFormEnd, true},
		{"endif", TokenFormEnd, true},
		{"else:", TokenSimpleLabel, true},
		{"nomatch", TokenVariable, false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			value, match, err := table.Lookup(tc.input)

			if tc.shouldMatch {
				if err != nil {
					t.Errorf("Expected match for '%s', but got error: %v", tc.input, err)
					return
				}
				if value != tc.expected {
					t.Errorf("Expected %v for '%s', got %v", tc.expected, tc.input, value)
				}
				if match == nil {
					t.Errorf("Expected non-nil match for '%s'", tc.input)
				}
			} else {
				if err == nil {
					t.Errorf("Expected no match for '%s', but got value: %v", tc.input, value)
				}
				if match != nil {
					t.Errorf("Expected nil match for '%s', but got: %v", tc.input, match)
				}
			}
		})
	}
}

func TestRegexpTable_TryLookup(t *testing.T) {
	table := NewRegexpTable[string](true, false) // Start anchoring, no end anchoring

	err := table.AddPattern(`hello`, "greeting")
	if err != nil {
		t.Fatalf("Failed to add pattern: %v", err)
	}

	// Test successful match
	value, match, err := table.Lookup("hello")
	if err != nil {
		t.Error("Expected successful match for 'hello'")
	}
	if value != "greeting" {
		t.Errorf("Expected 'greeting', got '%s'", value)
	}
	if match == nil {
		t.Error("Expected non-nil match")
	}

	// Test unsuccessful match
	value, match, err = table.Lookup("goodbye")
	if err == nil {
		t.Error("Expected unsuccessful match for 'goodbye'")
	}
	if value != "" {
		t.Errorf("Expected empty string for no match, got '%s'", value)
	}
	if match != nil {
		t.Error("Expected nil match for no match")
	}
}

func TestRegexpTable_LazyVsImmediateCompilation(t *testing.T) {
	// Test lazy compilation
	lazy := NewRegexpTable[string](true, false) // Start anchoring, no end anchoring

	// These should succeed without compilation
	err := lazy.AddPattern("valid", "value1")
	if err != nil {
		t.Errorf("Lazy AddPattern should succeed: %v", err)
	}

	// This should fail at classification time, not add time
	err = lazy.AddPattern("[invalid", "value2") // Invalid regex
	if err != nil {
		t.Errorf("Lazy AddPattern should defer validation: %v", err)
	}

	// Classification should fail due to invalid regex
	_, _, err = lazy.Lookup("test")
	if err == nil {
		t.Error("Expected lookup to fail due to invalid regexp")
	}

	// Test immediate compilation
	immediate := NewRegexpTable[string](true, false) // Start anchoring, no end anchoring

	// Valid pattern should succeed
	err = immediate.AddAndCheckPattern("valid", "value1")
	if err != nil {
		t.Errorf("Immediate AddPattern should succeed: %v", err)
	}

	// Invalid pattern should fail immediately
	err = immediate.AddAndCheckPattern("[invalid", "value2") // Invalid regex
	if err == nil {
		t.Error("Expected immediate AddPattern to fail with invalid regexp")
	}
}

func TestRegexpTable_ManualRecompile(t *testing.T) {
	table := NewRegexpTable[string](true, false) // Start anchoring, no end anchoring

	// Add patterns without compilation
	err := table.AddPattern("hello", "greeting")
	if err != nil {
		t.Fatalf("Failed to add pattern: %v", err)
	}

	err = table.AddPattern("world", "place")
	if err != nil {
		t.Fatalf("Failed to add pattern: %v", err)
	}

	// Manually trigger compilation
	err = table.Recompile()
	if err != nil {
		t.Fatalf("Manual recompile failed: %v", err)
	}

	// Now lookup should work
	value, _, err := table.Lookup("hello")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	if value != "greeting" {
		t.Errorf("Expected 'greeting', got '%s'", value)
	}
}

func TestRegexpTable_AnchoringOptions(t *testing.T) {
	// Test default anchoring (start anchored, not end anchored)
	t.Run("DefaultAnchoring", func(t *testing.T) {
		table := NewRegexpTable[string](true, false) // Start anchoring, no end anchoring

		err := table.AddPattern("hello", "greeting")
		if err != nil {
			t.Fatalf("Failed to add pattern: %v", err)
		}

		// Should match at start
		value, _, err := table.Lookup("hello world")
		if err != nil {
			t.Error("Expected match for 'hello world' with start anchoring")
		} else if value != "greeting" {
			t.Errorf("Expected 'greeting', got '%s'", value)
		}

		// Should match exact string
		value, _, err = table.Lookup("hello")
		if err != nil {
			t.Error("Expected match for 'hello'")
		} else if value != "greeting" {
			t.Errorf("Expected 'greeting', got '%s'", value)
		}

		// Should not match in middle
		_, _, err = table.Lookup("say hello")
		if err == nil {
			t.Error("Expected no match for 'say hello' with start anchoring")
		}
	})

	// Test no anchoring
	t.Run("NoAnchoring", func(t *testing.T) {
		table := NewRegexpTableWithEngine[string](&StandardRegexpEngine{}, false, false)

		err := table.AddPattern("hello", "greeting")
		if err != nil {
			t.Fatalf("Failed to add pattern: %v", err)
		}

		// Should match anywhere
		testCases := []string{"hello", "hello world", "say hello", "say hello world"}
		for _, input := range testCases {
			value, _, err := table.Lookup(input)
			if err != nil {
				t.Errorf("Expected match for '%s' with no anchoring", input)
			} else if value != "greeting" {
				t.Errorf("Expected 'greeting' for '%s', got '%s'", input, value)
			}
		}
	})

	// Test full anchoring (both start and end)
	t.Run("FullAnchoring", func(t *testing.T) {
		table := NewRegexpTableWithEngine[string](&StandardRegexpEngine{}, true, true)

		err := table.AddPattern("hello", "greeting")
		if err != nil {
			t.Fatalf("Failed to add pattern: %v", err)
		}

		// Should only match exact string
		value, _, err := table.Lookup("hello")
		if err != nil {
			t.Error("Expected match for exact 'hello' with full anchoring")
		} else if value != "greeting" {
			t.Errorf("Expected 'greeting', got '%s'", value)
		}

		// Should not match with extra text
		testCases := []string{"hello world", "say hello", "say hello world"}
		for _, input := range testCases {
			_, _, err := table.Lookup(input)
			if err == nil {
				t.Errorf("Expected no match for '%s' with full anchoring", input)
			}
		}
	})

	// Test end anchoring only
	t.Run("EndAnchoringOnly", func(t *testing.T) {
		table := NewRegexpTableWithEngine[string](&StandardRegexpEngine{}, false, true)

		err := table.AddPattern("world", "place")
		if err != nil {
			t.Fatalf("Failed to add pattern: %v", err)
		}

		// Should match at end
		value, _, err := table.Lookup("hello world")
		if err != nil {
			t.Error("Expected match for 'hello world' with end anchoring")
		} else if value != "place" {
			t.Errorf("Expected 'place', got '%s'", value)
		}

		// Should match exact string
		value, _, err = table.Lookup("world")
		if err != nil {
			t.Error("Expected match for 'world'")
		} else if value != "place" {
			t.Errorf("Expected 'place', got '%s'", value)
		}

		// Should not match in middle
		_, _, err = table.Lookup("world peace")
		if err == nil {
			t.Error("Expected no match for 'world peace' with end anchoring")
		}
	})
}

func TestRegexpTable_LookupOrElse(t *testing.T) {
	table := NewRegexpTable[string](true, false) // Start anchoring, no end anchoring

	// Add test patterns
	err := table.AddPattern("hello", "greeting")
	if err != nil {
		t.Fatalf("Failed to add pattern: %v", err)
	}

	err = table.AddPattern(`\d+`, "number")
	if err != nil {
		t.Fatalf("Failed to add pattern: %v", err)
	}

	// Test successful matches - should return matched value and matches
	t.Run("successful match", func(t *testing.T) {
		value, matches := table.LookupOrElse("hello", "default")
		if value != "greeting" {
			t.Errorf("Expected 'greeting', got '%s'", value)
		}
		if len(matches) < 1 || matches[0] != "hello" {
			t.Errorf("Expected matches to start with 'hello', got %v", matches)
		}
	})

	t.Run("successful match with number", func(t *testing.T) {
		value, matches := table.LookupOrElse("123", "default")
		if value != "number" {
			t.Errorf("Expected 'number', got '%s'", value)
		}
		if len(matches) < 1 || matches[0] != "123" {
			t.Errorf("Expected matches to start with '123', got %v", matches)
		}
	})

	// Test no match - should return default value and empty matches
	t.Run("no match returns default", func(t *testing.T) {
		value, matches := table.LookupOrElse("nomatch", "default_value")
		if value != "default_value" {
			t.Errorf("Expected 'default_value', got '%s'", value)
		}
		if len(matches) != 0 {
			t.Errorf("Expected empty matches, got %v", matches)
		}
	})

	// Test with different default value types
	t.Run("different default values", func(t *testing.T) {
		// String default
		value, _ := table.LookupOrElse("nomatch", "fallback")
		if value != "fallback" {
			t.Errorf("Expected 'fallback', got '%s'", value)
		}

		// Empty string default
		value, _ = table.LookupOrElse("nomatch", "")
		if value != "" {
			t.Errorf("Expected empty string, got '%s'", value)
		}
	})

	// Test with typed table using different types
	t.Run("typed table with int values", func(t *testing.T) {
		intTable := NewRegexpTable[int](true, false)
		err := intTable.AddPattern("one", 1)
		if err != nil {
			t.Fatalf("Failed to add pattern: %v", err)
		}

		// Successful match
		value, matches := intTable.LookupOrElse("one", 999)
		if value != 1 {
			t.Errorf("Expected 1, got %d", value)
		}
		if len(matches) < 1 || matches[0] != "one" {
			t.Errorf("Expected matches to start with 'one', got %v", matches)
		}

		// No match with default
		value, matches = intTable.LookupOrElse("nomatch", 999)
		if value != 999 {
			t.Errorf("Expected 999, got %d", value)
		}
		if len(matches) != 0 {
			t.Errorf("Expected empty matches, got %v", matches)
		}
	})
}
