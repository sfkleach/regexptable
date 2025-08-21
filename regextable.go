// Package regexptable provides efficient multi-pattern regexp classification.
//
// A regexp table is an associative array whose keys are regular expressions
// that map to arbitrary values. Lookup operations match strings against
// the regexp-keys to find a match and return the corresponding value and
// match groups.
//
// Core to the implementation is the compilation of regexp-keys into a single
// regular expression with named capture groups for each key. This allows
// efficient matching against the combined regexp.
//
// Example usage:
//
//	table, err := regexptable.NewRegexpTableBuilder[string]().
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
package regexptable

import (
	"fmt"
	"strings"
)

// ValueAndPattern holds both the value and original pattern for a regexp group.
type ValueAndPattern[T any] struct {
	GroupName       string // e.g. __REGEXPTABLE_1
	namedPattern    string // e.g. (?P<__REGEXPTABLE_1>pattern)
	Value           T
	Pattern         string         // e.g. pattern
	compiledPattern CompiledRegexp // Cached compiled pattern for disambiguation
}

// RegexpTable provides efficient multi-pattern regexp classification using a pluggable regexp engine.
// It compiles multiple regexp patterns into a single automaton for optimal performance.
type RegexpTable[T any] struct {
	engine         RegexpEngine
	compiled       CompiledRegexp
	lookup         []*ValueAndPattern[T]
	maplets        []*ValueAndPattern[T]
	nextGroupID    int
	needsRecompile bool
	anchorStart    bool // Whether to anchor patterns to start of string with ^
	anchorEnd      bool // Whether to anchor patterns to end of string with $
}

// NewRegexpTable creates a new empty RegexpTable using the standard regexp engine.
func NewRegexpTable[T any](anchorStart, anchorEnd bool) *RegexpTable[T] {
	return NewRegexpTableWithEngine[T](NewStandardRegexpEngine(), anchorStart, anchorEnd)
}

// NewRegexpTableWithEngine creates a new empty RegexpTable with a custom regexp engine.
func NewRegexpTableWithEngine[T any](engine RegexpEngine, anchorStart, anchorEnd bool) *RegexpTable[T] {
	return &RegexpTable[T]{
		engine:         engine,
		maplets:        make([]*ValueAndPattern[T], 0),
		nextGroupID:    1,
		needsRecompile: false,
		anchorStart:    anchorStart,
		anchorEnd:      anchorEnd,
	}
}

// AddPattern adds a new regexp pattern with its associated value to the table.
// This method defers recompilation until Lookup is called for better performance.
func (rt *RegexpTable[T]) AddPattern(pattern string, value T) error {
	// Auto-generate a unique internal name
	groupName := fmt.Sprintf("__REGEXPTABLE_%d__", rt.nextGroupID)
	rt.nextGroupID++

	// Create a unique capture group name using the engine's syntax
	namedPattern := rt.engine.FormatNamedGroup(groupName, pattern)

	rt.maplets = append(rt.maplets,
		&ValueAndPattern[T]{
			GroupName:    groupName,
			namedPattern: namedPattern,
			Value:        value,
			Pattern:      pattern,
		},
	)

	rt.needsRecompile = true

	return nil
}

// AddAndCheckPattern is like AddPattern but immediately recompiles the regexp.
// Use this when you need immediate validation of the pattern or when you're only adding one pattern.
func (rt *RegexpTable[T]) AddAndCheckPattern(pattern string, value T) error {
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

// anchorPattern applies start/end anchoring to a pattern based on the table's settings.
func (rt *RegexpTable[T]) anchorPattern(pattern string) string {
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
func (rt *RegexpTable[T]) validatePatterns() []string {
	var invalidPatterns []string

	for _, valueAndPattern := range rt.maplets {
		// Try to compile this pattern individually with proper anchoring
		anchoredPattern := rt.anchorPattern(valueAndPattern.Pattern)
		_, err := rt.engine.Compile(anchoredPattern)
		if err != nil {
			invalidPatterns = append(invalidPatterns, fmt.Sprintf("group %s (pattern: %s): %v", valueAndPattern.GroupName, valueAndPattern.Pattern, err))
		}
	}

	return invalidPatterns
}

// Recompile rebuilds the union regexp from all registered patterns.
// This is exposed to allow manual control over when recompilation occurs.
func (rt *RegexpTable[T]) Recompile() error {
	if len(rt.maplets) == 0 {
		rt.compiled = nil
		rt.needsRecompile = false
		return nil
	}

	// Create union pattern with proper anchoring
	var unionPattern strings.Builder
	for i, entry := range rt.maplets {
		if i > 0 {
			unionPattern.WriteString("|")
		}
		unionPattern.WriteString(entry.namedPattern)
	}
	anchoredUnionPattern := rt.anchorPattern(unionPattern.String())

	var err error
	rt.compiled, err = rt.engine.Compile(anchoredUnionPattern)
	if err != nil {
		// Try to identify which specific patterns are invalid
		invalidPatterns := rt.validatePatterns()
		if len(invalidPatterns) > 0 {
			return fmt.Errorf("failed to compile union regexp due to invalid patterns:\n%s", strings.Join(invalidPatterns, "\n"))
		}
		// Fallback to original error if we can't identify specific patterns
		return fmt.Errorf("failed to compile union regexp: %w", err)
	}

	// We now need to build the lookup slice. For each name in the SubexpNames
	// we use the corresponding ValueAndPattern from the maplets slice OR nil
	// if the name is "". The result is congruent to the strings returned by a match.
	names := rt.compiled.SubexpNames()
	n := 0
	rt.lookup = make([]*ValueAndPattern[T], 0)
	for _, name := range names {
		// Note that the SubexpNames will include the prefixed names in
		// the set order they were generated in. So we can rely on simply
		// walking the maplets slice.
		if strings.HasPrefix(name, "__REGEXPTABLE_") {
			rt.lookup = append(rt.lookup, rt.maplets[n]) // Skip the first empty name
			n++
		} else {
			rt.lookup = append(rt.lookup, nil)
		}
	}
	// for x, name := range names {
	// 	fmt.Println("subexpnames", x, name)
	// }
	// fmt.Println("lookup", len(rt.lookup), rt.lookup) // Debugging output to see lookup

	rt.needsRecompile = false
	return nil
}

// ensureCompiled ensures the regexp is compiled before use, recompiling if necessary.
func (rt *RegexpTable[T]) ensureCompiled() error {
	if rt.needsRecompile || rt.compiled == nil {
		return rt.Recompile()
	}
	return nil
}

// Lookup attempts to match the input string against all registered patterns.
// Returns the value, submatch slice, and error. If no patterns match, returns zero value, nil, error.
// This method automatically recompiles the regexp if patterns have been added/removed since last compilation.
func (rt *RegexpTable[T]) Lookup(input string) (T, []string, error) {
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
	// for x, m := range matches {
	// 	fmt.Println("match", x, m)
	// }

	// Note that rt.lookup and matches will be congruent (we force this in Recompile).
	for i, valueAndPattern := range rt.lookup {
		// fmt.Println("valueAndPattern", i, valueAndPattern) // Debugging output to see lookup and matches
		if valueAndPattern != nil && i < len(matches) && matches[i] != "" {
			// Now find the set of matches that applies for this lookup.
			our_matches := make([]string, 1)
			our_matches[0] = matches[i]
			for j := i + 1; j < len(rt.lookup); j++ {
				if rt.lookup[j] != nil {
					// Stop at the next __REGEXPTABLE capture group.
					break
				}
				// This must be a capture group that is part of the matching key.
				our_matches = append(our_matches, matches[j])
			}
			return valueAndPattern.Value, our_matches, nil
		}
	}

	// If all matches are empty strings, we need to disambiguate by testing individual patterns
	// This handles the case where multiple patterns could match empty strings or when alternation
	// makes it impossible to distinguish which group actually matched.
	for _, valueAndPattern := range rt.maplets {
		// Use cached compiled pattern or compile on-demand
		var individualRegexp CompiledRegexp
		if valueAndPattern.compiledPattern != nil {
			individualRegexp = valueAndPattern.compiledPattern
		} else {
			// Compile and cache the pattern
			individualPattern := rt.anchorPattern(valueAndPattern.Pattern)
			compiledRegexp, err := rt.engine.Compile(individualPattern)
			if err != nil {
				continue // Skip invalid patterns (should never happen)
			}
			// Cache the compiled pattern (note: this modifies the map entry)
			valueAndPattern.compiledPattern = compiledRegexp
			individualRegexp = compiledRegexp
		}

		// Test if this individual pattern matches
		if individualMatches := individualRegexp.FindStringSubmatch(input); individualMatches != nil {
			return valueAndPattern.Value, individualMatches, nil
		}
	}

	return zero, nil, fmt.Errorf("internal error: match found but no capture group matched")
}

func (rt *RegexpTable[T]) TryLookup(input string) (T, []string, bool) {
	value, matches, err := rt.Lookup(input)
	return value, matches, err == nil
}

func (rt *RegexpTable[T]) LookupOrElse(input string, defaultValue T) (T, []string) {
	value, matches, err := rt.Lookup(input)
	if err != nil {
		return defaultValue, []string{}
	}
	return value, matches
}
