// Package regextable provides efficient multi-pattern regex classification.
//
// A regex table is an associative array whose keys are regular expressions
// that map to arbitrary values. Lookup operations match strings against
// the regex-keys to find a match and return the corresponding value and
// match groups.
//
// Core to the implementation is the compilation of regex-keys into a single
// regular expression with named capture groups for each key. This allows
// efficient matching against the combined regex.
//
// Example usage:
//
//	table, err := regextable.NewRegexTableBuilder[string]().
//		AddPattern(`\d+`, "number").
//		AddPattern(`[a-zA-Z]+`, "word").
//		Build()
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	value, matches, ok := table.TryLookup("123")
//	// value: "number", matches: ["123"], ok: true
package regextable

import (
	"fmt"
	"strings"
)

// RegexTable provides efficient multi-pattern regex classification using a pluggable regex engine.
// It compiles multiple regex patterns into a single automaton for optimal performance.
type RegexTable[T any] struct {
	engine         RegexEngine
	compiled       CompiledRegex
	values         map[string]T
	patterns       map[string]string // Maps group names to original patterns
	patternNames   []string
	nextGroupID    int
	needsRecompile bool
	anchorStart    bool // Whether to anchor patterns to start of string with ^
	anchorEnd      bool // Whether to anchor patterns to end of string with $
}

// NewRegexTable creates a new empty RegexTable using the standard regex engine.
func NewRegexTable[T any](anchorStart, anchorEnd bool) *RegexTable[T] {
	return NewRegexTableWithEngine[T](NewStandardRegexEngine(), anchorStart, anchorEnd)
}

// NewRegexTableWithEngine creates a new empty RegexTable with a custom regex engine.
func NewRegexTableWithEngine[T any](engine RegexEngine, anchorStart, anchorEnd bool) *RegexTable[T] {
	return &RegexTable[T]{
		engine:         engine,
		values:         make(map[string]T),
		patterns:       make(map[string]string),
		patternNames:   make([]string, 0),
		nextGroupID:    1,
		needsRecompile: false,
		anchorStart:    anchorStart,
		anchorEnd:      anchorEnd,
	}
}

// AddPattern adds a new regex pattern with its associated value to the table.
// This method defers recompilation until Lookup is called for better performance.
func (rt *RegexTable[T]) AddPattern(pattern string, value T) error {
	// Auto-generate a unique internal name
	groupName := fmt.Sprintf("__REGEXTABLE_%d__", rt.nextGroupID)
	rt.nextGroupID++

	// Create a unique capture group name using the engine's syntax
	namedPattern := rt.engine.FormatNamedGroup(groupName, pattern)

	rt.patternNames = append(rt.patternNames, namedPattern)
	rt.values[groupName] = value
	rt.patterns[groupName] = pattern // Store the original pattern
	rt.needsRecompile = true

	return nil
}

// AddPatternThenRecompile is like AddPattern but immediately recompiles the regex.
// Use this when you need immediate validation of the pattern or when you're only adding one pattern.
func (rt *RegexTable[T]) AddPatternThenRecompile(pattern string, value T) error {
	err := rt.AddPattern(pattern, value)
	if err != nil {
		return err
	}

	err = rt.Recompile()
	if err != nil {
		return err
	}

	return nil
}

// HasPatterns returns true if the table has any patterns configured.
func (rt *RegexTable[T]) HasPatterns() bool {
	return len(rt.patternNames) > 0
}

// anchorPattern applies start/end anchoring to a pattern based on the table's settings.
func (rt *RegexTable[T]) anchorPattern(pattern string) string {
	result := pattern
	if rt.anchorStart {
		result = "^(?:" + result + ")"
	} else {
		result = "(?:" + result + ")"
	}
	if rt.anchorEnd {
		result = result + "$"
	}
	return result
}

// validatePatterns checks each pattern individually and returns details about any invalid patterns.
func (rt *RegexTable[T]) validatePatterns() []string {
	var invalidPatterns []string

	for groupName, originalPattern := range rt.patterns {
		// Try to compile this pattern individually with proper anchoring
		anchoredPattern := rt.anchorPattern(originalPattern)
		_, err := rt.engine.Compile(anchoredPattern)
		if err != nil {
			invalidPatterns = append(invalidPatterns, fmt.Sprintf("group %s (pattern: %s): %v", groupName, originalPattern, err))
		}
	}

	return invalidPatterns
}

// Recompile rebuilds the union regex from all registered patterns.
// This is exposed to allow manual control over when recompilation occurs.
func (rt *RegexTable[T]) Recompile() error {
	if len(rt.patternNames) == 0 {
		rt.compiled = nil
		rt.needsRecompile = false
		return nil
	}

	// Create union pattern with proper anchoring
	unionPattern := strings.Join(rt.patternNames, "|")
	anchoredUnionPattern := rt.anchorPattern(unionPattern)

	var err error
	rt.compiled, err = rt.engine.Compile(anchoredUnionPattern)
	if err != nil {
		// Try to identify which specific patterns are invalid
		invalidPatterns := rt.validatePatterns()
		if len(invalidPatterns) > 0 {
			return fmt.Errorf("failed to compile union regex due to invalid patterns:\n%s", strings.Join(invalidPatterns, "\n"))
		}
		// Fallback to original error if we can't identify specific patterns
		return fmt.Errorf("failed to compile union regex: %w", err)
	}

	rt.needsRecompile = false
	return nil
}

// ensureCompiled ensures the regex is compiled before use, recompiling if necessary.
func (rt *RegexTable[T]) ensureCompiled() error {
	if rt.needsRecompile || rt.compiled == nil {
		return rt.Recompile()
	}
	return nil
}

// Lookup attempts to match the input string against all registered patterns.
// Returns the value, submatch slice, and error. If no patterns match, returns zero value, nil, error.
// This method automatically recompiles the regex if patterns have been added/removed since last compilation.
func (rt *RegexTable[T]) Lookup(input string) (T, []string, error) {
	var zero T

	err := rt.ensureCompiled()
	if err != nil {
		return zero, nil, err
	}

	if rt.compiled == nil {
		return zero, nil, fmt.Errorf("no patterns configured")
	}

	matches := rt.compiled.FindStringSubmatch(input)
	if matches == nil {
		return zero, nil, fmt.Errorf("no pattern matched")
	}

	// Find which named group matched by checking submatches
	subexpNames := rt.compiled.SubexpNames()
	for i, name := range subexpNames {
		// Defensive check: ensure we don't exceed matches slice bounds
		// (SubexpNames and matches should have same length, but we use pluggable engines)
		if name != "" && i < len(matches) && matches[i] != "" {
			if value, exists := rt.values[name]; exists {
				return value, matches, nil
			}
		}
	}

	// If all matches are empty strings, we need to disambiguate by testing individual patterns
	// This handles the case where multiple patterns could match empty strings or when alternation
	// makes it impossible to distinguish which group actually matched
	for i, name := range subexpNames {
		if name != "" && i < len(matches) {
			if value, exists := rt.values[name]; exists {
				// Get the original pattern for this group
				if originalPattern, patternExists := rt.patterns[name]; patternExists {
					// Create a regex for just this pattern with proper anchoring
					individualPattern := rt.anchorPattern(originalPattern)
					individualRegex, err := rt.engine.Compile(individualPattern)
					if err != nil {
						continue // Skip invalid patterns
					}

					// Test if this individual pattern matches
					if individualMatches := individualRegex.FindStringSubmatch(input); individualMatches != nil {
						return value, matches, nil
					}
				}
			}
		}
	}

	return zero, nil, fmt.Errorf("internal error: match found but no capture group matched")
}

func (rt *RegexTable[T]) TryLookup(input string) (T, []string, bool) {
	value, matches, err := rt.Lookup(input)
	return value, matches, err == nil
}

func (rt *RegexTable[T]) LookupOrElse(input string, defaultValue T) (T, []string) {
	value, matches, err := rt.Lookup(input)
	if err != nil {
		return defaultValue, []string{}
	}
	return value, matches
}
