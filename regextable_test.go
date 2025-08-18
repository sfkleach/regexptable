package regextable

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

func TestRegexTable_Basic(t *testing.T) {
	table := NewRegexTable[TokenType](true, false) // Start anchoring, no end anchoring

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

func TestRegexTable_TryLookup(t *testing.T) {
	table := NewRegexTable[string](true, false) // Start anchoring, no end anchoring

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

func TestRegexTable_LazyVsImmediateCompilation(t *testing.T) {
	// Test lazy compilation
	lazy := NewRegexTable[string](true, false) // Start anchoring, no end anchoring

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
		t.Error("Expected lookup to fail due to invalid regex")
	}

	// Test immediate compilation
	immediate := NewRegexTable[string](true, false) // Start anchoring, no end anchoring

	// Valid pattern should succeed
	err = immediate.AddPatternThenRecompile("valid", "value1")
	if err != nil {
		t.Errorf("Immediate AddPattern should succeed: %v", err)
	}

	// Invalid pattern should fail immediately
	err = immediate.AddPatternThenRecompile("[invalid", "value2") // Invalid regex
	if err == nil {
		t.Error("Expected immediate AddPattern to fail with invalid regex")
	}
}

func TestRegexTable_ManualRecompile(t *testing.T) {
	table := NewRegexTable[string](true, false) // Start anchoring, no end anchoring

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

func TestRegexTable_AnchoringOptions(t *testing.T) {
	// Test default anchoring (start anchored, not end anchored)
	t.Run("DefaultAnchoring", func(t *testing.T) {
		table := NewRegexTable[string](true, false) // Start anchoring, no end anchoring

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
		table := NewRegexTableWithEngine[string](&StandardRegexEngine{}, false, false)

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
		table := NewRegexTableWithEngine[string](&StandardRegexEngine{}, true, true)

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
		table := NewRegexTableWithEngine[string](&StandardRegexEngine{}, false, true)

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
