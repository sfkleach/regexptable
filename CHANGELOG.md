# Change Log for RegexTable

Following the style in https://keepachangelog.com/en/1.0.0/

## Unreleased 

### Added

- **Core RegexTable[T]**: Generic regex table with efficient multi-pattern matching
- **Builder Pattern**: `RegexTableBuilder[T]` with fluent API for easy construction
- **Multiple Lookup Methods**: 
  - `Lookup(input)` - returns value, matches, and error
  - `TryLookup(input)` - returns value, matches, and boolean success
  - `LookupOrElse(input, default)` - returns value/matches or default/empty slice
- **Pluggable Regex Engines**: Interface-based system supporting different regex implementations
- **Standard Regex Engine**: Built-in implementation using Go's `regexp` package
- **Lazy Compilation**: Deferred regex compilation until lookup for optimal performance
- **Pattern Management**: Add/remove patterns with auto-generated IDs and lazy recompilation
- **Named Capture Groups**: Uses reserved `__REGEXTABLE_N__` namespace to avoid conflicts
- **Type Safety**: Full generic support for any value type T
- **Zero Dependencies**: Core package only depends on Go standard library
- **Builder Utilities**: Clone, Clear, PatternCount, HasPatterns methods
- **Immediate Compilation Options**: `AddPatternThenRecompile` and `RemovePatternThenRecompile`
- **Comprehensive Documentation**: README, CONTRIBUTING.md, and integration examples
- **regexp2 Integration Guide**: Complete example for using advanced regex features like lookbehind

