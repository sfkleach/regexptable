# Change Log for RegexTable

Following the style in https://keepachangelog.com/en/1.0.0/

## [Unreleased]

Nothing yet.

## [0.1.0] - 2025-08-19

Initial release of the RegexTable library for Go.

### Added

- **Core RegexTable[T]**: Generic regex table with efficient multi-pattern matching using named capture groups
- **Builder Pattern**: `RegexTableBuilder[T]` with fluent API for easy construction and method chaining
- **Multiple Lookup Methods**: 
  - `Lookup(input)` - returns value, matches, and error
  - `TryLookup(input)` - returns value, matches, and boolean success
  - `LookupOrElse(input, default)` - returns value/matches or default/empty slice
- **Pluggable Regex Engines**: Interface-based system supporting different regex implementations
- **Standard Regex Engine**: Built-in implementation using Go's `regexp` package
- **Lazy Compilation**: Deferred regex compilation until first lookup for optimal performance
- **Dynamic Pattern Management**: Add/remove patterns with auto-generated IDs and lazy recompilation
- **Named Capture Groups**: Uses reserved `__REGEXTABLE_N__` namespace to avoid conflicts with user patterns
- **Type Safety**: Full generic support for any value type T with compile-time type checking
- **Zero Dependencies**: Core package only depends on Go standard library
- **Builder Utilities**: 
  - `Clone()` - create independent copies of builders
  - `Clear()` - reset builder to empty state
  - `MustBuild()` - panic on compilation errors for static configurations
- **Comprehensive Examples**:
  - Basic usage examples in README
  - Complete builder example demonstrating tokenization use case
  - Integration guide for regexp2 library with advanced features like lookbehind
- **Development Infrastructure**:
  - Comprehensive test suite with 100% coverage
  - GitHub Actions CI/CD pipeline
  - Contributing guidelines and project documentation

