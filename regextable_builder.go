package regexptable

import (
	"fmt"
	"strings"
)

// RegexpTableBuilder provides a convenient builder pattern for creating RegexpTable instances.
// It accumulates patterns and builds the final RegexpTable with a single compilation step.
type RegexpTableBuilder[T any] struct {
	patterns []patternEntry[T]
	engine   RegexpEngine
}

// patternEntry holds a pattern and its associated value during building
type patternEntry[T any] struct {
	pattern string
	value   T
}

// RegexpTableSubBuilder provides a type-safe fluent interface for building alternation patterns.
// It is returned by BeginAddSubPatterns() and ensures proper method sequencing.
type RegexpTableSubBuilder[T any] struct {
	parent      *RegexpTableBuilder[T]
	subPatterns []string
}

// NewRegexpTableBuilder creates a new RegexpTableBuilder with the standard regexp engine.
func NewRegexpTableBuilder[T any]() *RegexpTableBuilder[T] {
	return &RegexpTableBuilder[T]{
		patterns: make([]patternEntry[T], 0),
		engine:   &StandardRegexpEngine{},
	}
}

// NewRegexpTableBuilderWithEngine creates a new RegexpTableBuilder with a custom engine.
func NewRegexpTableBuilderWithEngine[T any](engine RegexpEngine) *RegexpTableBuilder[T] {
	return &RegexpTableBuilder[T]{
		patterns: make([]patternEntry[T], 0),
		engine:   engine,
	}
}

// AddPattern adds a pattern to be included in the final RegexpTable.
// This method never fails - validation happens during Build().
func (b *RegexpTableBuilder[T]) AddPattern(pattern string, value T) *RegexpTableBuilder[T] {
	b.patterns = append(b.patterns, patternEntry[T]{
		pattern: pattern,
		value:   value,
	})
	return b
}

// AddPatterns adds multiple patterns as a single alternation pattern with a shared value.
// The patterns are combined using alternation syntax (?:pattern1|pattern2|...) and
// treated as a single regexp key that maps to the given value. Note that anchoring
// does not apply to this construction, as it is simply a longhand way to add
// a single pattern entry.
func (b *RegexpTableBuilder[T]) AddSubPatterns(patterns []string, value T) *RegexpTableBuilder[T] {
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

// Build creates the final RegexpTable with all accumulated patterns.
// This is when compilation and validation occur.
func (b *RegexpTableBuilder[T]) Build(anchorStart, anchorEnd bool) (*RegexpTable[T], error) {
	table := NewRegexpTableWithEngine[T](b.engine, anchorStart, anchorEnd)

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
		return nil, fmt.Errorf("failed to compile regexp table: %w", err)
	}

	return table, nil
}

// MustBuild is like Build but panics on error. Useful for static configurations
// where patterns are known to be valid.
func (b *RegexpTableBuilder[T]) MustBuild(anchorStart, anchorEnd bool) *RegexpTable[T] {
	table, err := b.Build(anchorStart, anchorEnd)
	if err != nil {
		panic(fmt.Sprintf("RegexpTableBuilder.MustBuild failed: %v", err))
	}
	return table
}

// Clear removes all patterns from the builder, allowing it to be reused.
func (b *RegexpTableBuilder[T]) Clear() *RegexpTableBuilder[T] {
	b.patterns = b.patterns[:0] // Reset slice but keep capacity
	return b
}

// Clone creates a copy of the builder with the same patterns and engine.
func (b *RegexpTableBuilder[T]) Clone() *RegexpTableBuilder[T] {
	clone := NewRegexpTableBuilderWithEngine[T](b.engine)
	clone.patterns = make([]patternEntry[T], len(b.patterns))
	copy(clone.patterns, b.patterns)
	return clone
}

// BeginAddSubPatterns starts building an alternation pattern with a type-safe fluent interface.
// Returns a RegexpTableSubBuilder that only allows AddSubPattern() and EndAddSubPatterns() calls.
// This prevents calling methods out of order and ensures proper alternation construction.
// Usage: BeginAddSubPatterns() -> AddSubPattern(...) -> EndAddSubPatterns(value).
func (b *RegexpTableBuilder[T]) BeginAddSubPatterns() *RegexpTableSubBuilder[T] {
	return &RegexpTableSubBuilder[T]{
		parent:      b,
		subPatterns: make([]string, 0),
	}
}

// AddSubPattern adds a pattern to the current alternation being built.
// Must be called between BeginAddSubPatterns() and EndAddSubPatterns().
// Returns RegexpTableSubBuilder to maintain type-safe method chaining.
func (sb *RegexpTableSubBuilder[T]) AddSubPattern(pattern string) *RegexpTableSubBuilder[T] {
	sb.subPatterns = append(sb.subPatterns, pattern)
	return sb
}

// EndAddSubPatterns completes the alternation pattern and adds it to the builder with the given value.
// The accumulated sub-patterns are combined using alternation syntax (?:pattern1|pattern2|...).
// Returns the parent RegexpTableBuilder to continue the fluent interface.
func (sb *RegexpTableSubBuilder[T]) EndAddSubPatterns(value T) *RegexpTableBuilder[T] {
	// Use AddSubPatterns to handle the alternation logic. Note that anchoring
	// is not applied at this level (it would make no sense).
	sb.parent.AddSubPatterns(sb.subPatterns, value)

	// Clear the sub-patterns after use
	sb.subPatterns = sb.subPatterns[:0]
	return sb.parent
}
