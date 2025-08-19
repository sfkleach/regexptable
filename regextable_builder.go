package regextable

import (
	"fmt"
	"strings"
)

// RegexTableBuilder provides a convenient builder pattern for creating RegexTable instances.
// It accumulates patterns and builds the final RegexTable with a single compilation step.
type RegexTableBuilder[T any] struct {
	patterns []patternEntry[T]
	engine   RegexEngine
}

// patternEntry holds a pattern and its associated value during building
type patternEntry[T any] struct {
	pattern string
	value   T
}

// RegexTableSubBuilder provides a type-safe fluent interface for building alternation patterns.
// It is returned by BeginAddSubPatterns() and ensures proper method sequencing.
type RegexTableSubBuilder[T any] struct {
	parent      *RegexTableBuilder[T]
	subPatterns []string
}

// NewRegexTableBuilder creates a new RegexTableBuilder with the standard regex engine.
func NewRegexTableBuilder[T any]() *RegexTableBuilder[T] {
	return &RegexTableBuilder[T]{
		patterns: make([]patternEntry[T], 0),
		engine:   &StandardRegexEngine{},
	}
}

// NewRegexTableBuilderWithEngine creates a new RegexTableBuilder with a custom engine.
func NewRegexTableBuilderWithEngine[T any](engine RegexEngine) *RegexTableBuilder[T] {
	return &RegexTableBuilder[T]{
		patterns: make([]patternEntry[T], 0),
		engine:   engine,
	}
}

// AddPattern adds a pattern to be included in the final RegexTable.
// This method never fails - validation happens during Build().
func (b *RegexTableBuilder[T]) AddPattern(pattern string, value T) *RegexTableBuilder[T] {
	b.patterns = append(b.patterns, patternEntry[T]{
		pattern: pattern,
		value:   value,
	})
	return b
}

// AddPatterns adds multiple patterns as a single alternation pattern with a shared value.
// The patterns are combined using alternation syntax (?:pattern1|pattern2|...) and
// treated as a single regex key that maps to the given value.
func (b *RegexTableBuilder[T]) AddSubPatterns(patterns []string, value T) *RegexTableBuilder[T] {
	if len(patterns) == 0 {
		return b // No patterns to add, return unchanged
	}

	if len(patterns) == 1 {
		// Single pattern, no need for alternation syntax
		return b.AddPattern(patterns[0], value)
	}

	// Create alternation pattern with proper grouping
	var alternation strings.Builder
	alternation.WriteString("(?:")
	for i, pattern := range patterns {
		if i > 0 {
			alternation.WriteString("|")
		}
		alternation.WriteString(pattern)
	}
	alternation.WriteString(")")

	return b.AddPattern(alternation.String(), value)
}

// Build creates the final RegexTable with all accumulated patterns.
// This is when compilation and validation occur.
func (b *RegexTableBuilder[T]) Build(anchorStart, anchorEnd bool) (*RegexTable[T], error) {
	table := NewRegexTableWithEngine[T](b.engine, anchorStart, anchorEnd)

	// Add all patterns to the table (using lazy compilation)
	for _, entry := range b.patterns {
		err := table.AddPattern(entry.pattern, entry.value)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern '%s': %w", entry.pattern, err)
		}
	}

	// Trigger compilation once at the end
	err := table.Recompile()
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex table: %w", err)
	}

	return table, nil
}

// MustBuild is like Build but panics on error. Useful for static configurations
// where patterns are known to be valid.
func (b *RegexTableBuilder[T]) MustBuild(anchorStart, anchorEnd bool) *RegexTable[T] {
	table, err := b.Build(anchorStart, anchorEnd)
	if err != nil {
		panic(fmt.Sprintf("RegexTableBuilder.MustBuild failed: %v", err))
	}
	return table
}

// Clear removes all patterns from the builder, allowing it to be reused.
func (b *RegexTableBuilder[T]) Clear() *RegexTableBuilder[T] {
	b.patterns = b.patterns[:0] // Reset slice but keep capacity
	return b
}

// Clone creates a copy of the builder with the same patterns and engine.
func (b *RegexTableBuilder[T]) Clone() *RegexTableBuilder[T] {
	clone := NewRegexTableBuilderWithEngine[T](b.engine)
	clone.patterns = make([]patternEntry[T], len(b.patterns))
	copy(clone.patterns, b.patterns)
	return clone
}

// BeginAddSubPatterns starts building an alternation pattern with a type-safe fluent interface.
// Returns a RegexTableSubBuilder that only allows AddSubPattern() and EndAddSubPatterns() calls.
// This prevents calling methods out of order and ensures proper alternation construction.
// Usage: BeginAddSubPatterns() -> AddSubPattern(...) -> EndAddSubPatterns(value).
func (b *RegexTableBuilder[T]) BeginAddSubPatterns() *RegexTableSubBuilder[T] {
	return &RegexTableSubBuilder[T]{
		parent:      b,
		subPatterns: make([]string, 0),
	}
}

// AddSubPattern adds a pattern to the current alternation being built.
// Must be called between BeginAddSubPatterns() and EndAddSubPatterns().
// Returns RegexTableSubBuilder to maintain type-safe method chaining.
func (sb *RegexTableSubBuilder[T]) AddSubPattern(pattern string) *RegexTableSubBuilder[T] {
	sb.subPatterns = append(sb.subPatterns, pattern)
	return sb
}

// EndAddSubPatterns completes the alternation pattern and adds it to the builder with the given value.
// The accumulated sub-patterns are combined using alternation syntax (?:pattern1|pattern2|...).
// Returns the parent RegexTableBuilder to continue the fluent interface.
func (sb *RegexTableSubBuilder[T]) EndAddSubPatterns(value T) *RegexTableBuilder[T] {
	// Use AddSubPatterns to handle the alternation logic
	sb.parent.AddSubPatterns(sb.subPatterns, value)

	// Clear the sub-patterns after use
	sb.subPatterns = sb.subPatterns[:0]
	return sb.parent
}
