# Change Log for RegexpTable

Following the style in https://keepachangelog.com/en/1.0.0/

## [0.1.2]

### Fixed

- Match groups inside key-regexes are now correctly matched.
- Documentation for integrating with other regexp engines is accurate.

### Improved

- Enhanced documentation for `CompiledRegexp` interface methods with detailed explanations:
  - `FindStringSubmatch()` now explains the 1:1 correspondence with `SubexpNames()`
  - `SubexpNames()` now includes comprehensive behavior description matching Go's `regexp.SubexpNames()`
  - Added practical guidance on counting capture groups using parentheses in patterns
- Clarified that `CompiledRegexp` interface enables pluggable regexp engines

## [0.1.1] - 2025-08-19

Initial release of the RegexpTable library for Go.

### Added

- **Core RegexpTable[T]**: Generic regexp table with efficient multi-pattern matching using named capture groups
- **Builder Pattern**: `RegexpTableBuilder[T]` with fluent API for easy construction and method chaining
- **Multiple Lookup Methods**: 
  - `Lookup(input)` - returns value, matches, and error
  - `TryLookup(input)` - returns value, matches, and boolean success
  - `LookupOrElse(input, default)` - returns value/matches or default/empty slice
- **Pluggable Regexp Engines**: Interface-based system supporting different regexp implementations
- **Standard Regexp Engine**: Built-in implementation using Go's `regexp` package
- **Lazy Compilation**: Deferred regexp compilation until first lookup for optimal performance
- **Dynamic Pattern Management**: Add/remove patterns with auto-generated IDs and lazy recompilation
- **Named Capture Groups**: Uses reserved `__REGEXPTABLE_N__` namespace to avoid conflicts with user patterns
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

