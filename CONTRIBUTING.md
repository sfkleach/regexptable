# Contributing to RegexTable

Thank you for your interest in contributing to RegexTable! This document provides guidelines for contributing to this Go library for efficient multi-pattern regex classification.

## Getting Started

### Prerequisites

- Go 1.24.2 or later
- [Just](https://github.com/casey/just) (optional, for running tasks)

### Setting Up Development Environment

1. Fork and clone the repository:
   ```bash
   git clone https://github.com/sfkleach/regextable.git
   cd regextable
   ```

2. Verify your setup:
   ```bash
   go mod tidy
   go test ./...
   ```

   Or if you have `just` installed:
   ```bash
   just test
   ```

3. Run the example:
   ```bash
   cd builder_example
   go run main.go
   ```

## Development Guidelines

### Dependency Policy

**Important**: This package is intended to depend only on the Go standard library. This design decision ensures:

- **Minimal dependencies**: Users don't inherit transitive dependencies they don't need
- **Broad compatibility**: No version conflicts with user applications
- **Easy adoption**: Simple `go get` without dependency management concerns
- **Lightweight footprint**: Core functionality remains minimal and focused

### Regex Engine Implementations

Different regex engines (like `regexp2`, `re2`, etc.) should be implemented in **separate repositories** to maintain the zero-dependency principle:

```
github.com/sfkleach/regextable           # Core package (stdlib only)
github.com/your-org/regextable-regexp2   # regexp2 integration (separate repo)
github.com/your-org/regextable-re2       # re2 integration (separate repo)
```

This allows users to:
- Use the core library without any external dependencies
- Selectively add only the regex engines they need
- Avoid pulling in heavy dependencies for features they don't use

When contributing regex engine support, please create companion packages rather than adding dependencies to this core repository.

### Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Comments should be proper sentences with correct grammar and punctuation, including capitalization and periods
- Use meaningful variable and function names
- Keep functions focused and single-purpose

### Defensive Programming

When adding defensive checks (especially around interface boundaries), include a comment explaining why the check is appropriate. For example:

```go
// Defensive check: ensure we don't exceed matches slice bounds
// (SubexpNames and matches should have same length, but we use pluggable engines)
if name != "" && i < len(matches) {
    // ...
}
```

This is particularly important for:
- Interface implementations where behavior can't be guaranteed
- Pluggable components (like `RegexEngine` implementations)
- Boundary conditions and edge cases

### API Design Principles

- **Type Safety**: Use generics (`RegexTable[T]`) for compile-time safety
- **Builder Pattern**: Provide fluent APIs that hide complexity
- **Lazy Compilation**: Defer expensive operations until needed
- **Multiple Access Patterns**: Support different error handling styles (`Lookup`, `TryLookup`, `LookupOrElse`)

## Testing

### Running Tests

```bash
go test ./...
```

Or with just:
```bash
just test
```

### Writing Tests

- Write comprehensive tests for new functionality
- Include edge cases and error conditions
- Test both the core API and builder pattern
- For regex engine implementations, test with various pattern types

Example test structure:
```go
func TestNewFeature(t *testing.T) {
    table := regextable.NewRegexTableBuilder[string]().
        AddPattern(`test_pattern`, "expected_value").
        MustBuild()
    
    value, matches, ok := table.TryLookup("test_input")
    if !ok {
        t.Error("Expected match")
    }
    // ... assertions
}
```

## Adding New Regex Engines

The library supports pluggable regex engines through the `RegexEngine` interface. However, **implementations for non-standard engines should be created in separate repositories** to maintain our zero-dependency policy (see Dependency Policy above).

### Implementation Steps

1. **Create a separate repository** with clear naming (e.g., `regextable-enginename`)

2. **Import this core package** as a dependency:
   ```go
   import "github.com/sfkleach/regextable"
   ```

3. **Implement the interfaces** in your package:
   ```go
   type MyEngine struct{}
   
   func (e *MyEngine) Compile(pattern string) (regextable.CompiledRegex, error) {
       // Your implementation using external library
   }
   
   func (e *MyEngine) FormatNamedGroup(groupName, pattern string) string {
       // Your engine's named group syntax
   }
   ```

4. **Add comprehensive tests** that verify compatibility with the core regextable API

5. **Provide documentation and examples** showing usage patterns

### Example Package Structure

```
github.com/your-org/regextable-regexp2/
├── go.mod                    # Contains regexp2 dependency
├── regexp2_engine.go         # Engine implementation
├── regexp2_engine_test.go    # Comprehensive tests
├── examples/                 # Usage examples
└── README.md                 # Documentation
```

See `docs/integrating_with_regexp2.md` for a detailed implementation example.

## Documentation

### Code Documentation

- Add package-level documentation with examples
- Document all public types, functions, and methods
- Include usage examples in comments
- Use `Example*` functions for testable documentation

### README Updates

When adding significant features:
- Update the feature list
- Add usage examples
- Update the API reference section
- Consider adding performance notes

## Submitting Changes

### Pull Request Process

1. Create a feature branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following the guidelines above

3. Ensure all tests pass:
   ```bash
   just test
   go vet ./...
   gofmt -d .
   ```

4. Commit with clear, descriptive messages:
   ```bash
   git commit -m "Add support for custom regex engines
   
   - Implement RegexEngine interface
   - Add comprehensive tests
   - Update documentation with examples"
   ```

5. Push and create a pull request

### Pull Request Guidelines

- **Clear Description**: Explain what the PR does and why
- **Small, Focused Changes**: Keep PRs focused on a single feature or fix
- **Tests Included**: All new code should have corresponding tests
- **Documentation**: Update docs for user-facing changes
- **Backward Compatibility**: Avoid breaking existing APIs unless absolutely necessary
- **Dependency Policy**: Ensure no external dependencies are added to the core package

**Note**: Pull requests that add external dependencies to the core package will not be accepted. Please create separate companion packages for regex engines that require external dependencies.

### Commit Message Format

Use clear, imperative mood commit messages:
- `Add feature for X`
- `Fix issue with Y`
- `Update documentation for Z`
- `Refactor method to improve performance`

## Performance Considerations

This library is designed for performance. When contributing:

- **Benchmark significant changes**: Use `go test -bench=.` for performance-critical code
- **Consider memory allocation**: Minimize allocations in hot paths
- **Lazy compilation**: Maintain the principle of deferred compilation
- **Profile when needed**: Use `go tool pprof` for complex performance issues

## Architecture Notes

### Core Components

- **`RegexTable[T]`**: Main data structure with pluggable engine support
- **`RegexTableBuilder[T]`**: Fluent API for table construction
- **`RegexEngine` interface**: Abstraction for different regex implementations
- **`CompiledRegex` interface**: Abstraction for compiled regex objects

### Key Design Decisions

- **Single compiled regex**: All patterns are combined into one regex for O(n) matching
- **Named capture groups**: Use reserved namespace (`__REGEXTABLE_N__`) to avoid conflicts
- **Pluggable engines**: Support different regex syntaxes and features
- **Generic types**: Type-safe values without runtime casting

## Getting Help

- **Issues**: Use GitHub issues for bug reports and feature requests
- **Discussions**: Use GitHub discussions for questions and ideas
- **Code Review**: Maintainers will review PRs and provide feedback

## Code of Conduct

Please be respectful and constructive in all interactions. This project aims to be welcoming to contributors of all backgrounds and experience levels.

## Release Process

Releases follow semantic versioning:
- **Patch** (0.0.X): Bug fixes and minor improvements
- **Minor** (0.X.0): New features that maintain backward compatibility
- **Major** (X.0.0): Breaking changes to the public API

Thank you for contributing to RegexTable! 🚀
