# Contributing to SublimeGo

Thank you for your interest in contributing! This document covers everything you need to get started.

---

## Prerequisites

- Go 1.24 or later
- Git
- [Templ CLI](https://templ.guide/)  `go install github.com/a-h/templ/cmd/templ@latest`
- [golangci-lint](https://golangci-lint.run/)  `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- [air](https://github.com/air-verse/air) (optional, for hot reload)  `go install github.com/air-verse/air@latest`

Or install everything at once:

```bash
make install-tools
```

---

## Development Setup

```bash
# 1. Fork and clone
git clone https://github.com/bozz33/sublimego.git
cd sublimego

# 2. Download dependencies
go mod download

# 3. Download Tailwind CSS standalone CLI
make tailwind-download

# 4. Run tests to verify the setup
go test ./...

# 5. Start in development mode
make dev
```

---

## Branching Strategy

| Branch | Purpose |
|--------|---------|
| `main` | Stable release  only merge from `SublimeGo-Dev` via PR |
| `SublimeGo-Dev` | Active development  integration branch |
| `refactor/go-standards` | Ongoing Go standards refactoring |
| `feature/*` | New features  branch from `SublimeGo-Dev` |
| `fix/*` | Bug fixes  branch from `SublimeGo-Dev` |
| `docs/*` | Documentation only |

```bash
# Start a new feature
git checkout SublimeGo-Dev
git pull origin SublimeGo-Dev
git checkout -b feature/my-feature

# Start a bug fix
git checkout -b fix/issue-description
```

---

## Workflow

1. Create a branch from `SublimeGo-Dev`
2. Make your changes
3. Write or update tests
4. Run the full check suite:

```bash
go test ./...           # All tests must pass
go test -race ./...     # No data races
golangci-lint run       # No lint errors
go vet ./...            # No vet issues
templ generate          # Regenerate templates if .templ files changed
```

5. Commit using [Conventional Commits](#commit-message-format)
6. Open a Pull Request against `SublimeGo-Dev`

---

## Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short description>

[optional body]

[optional footer]
```

### Types

| Type | When to use |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation changes only |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `test` | Adding or updating tests |
| `chore` | Maintenance (deps, CI, tooling) |
| `perf` | Performance improvement |
| `style` | Formatting, missing semicolons  no logic change |

### Examples

```bash
git commit -m "feat(form): add ColorPicker and Slider fields"
git commit -m "fix(notifications): close SSE channel on context cancellation"
git commit -m "docs: rewrite ARCHITECTURE.md with current module layout"
git commit -m "refactor(engine): extract tenant middleware to separate file"
git commit -m "test(table): add grouping unit tests"
git commit -m "chore: remove mattn/go-sqlite3 direct import"
```

---

## Code Style

### Go Code

- Run `gofmt` before committing (enforced by CI)
- Follow [Effective Go](https://go.dev/doc/effective_go) and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Export only what needs to be exported
- Document all exported types, functions, and methods
- Handle every error explicitly  never use `_` for errors in production code
- Prefer table-driven tests
- Use `context.Context` as the first parameter for functions that perform I/O

### Naming Conventions

```go
// Interfaces  describe behaviour, not the implementor
type NotificationStore interface { ... }   // good
type INotificationStore interface { ... }  // bad  no I prefix in Go

// Constructors
func NewStore(...) *Store { ... }          // New prefix for constructors

// Builders  With* prefix for option setters
func (t *Table) WithColumns(...) *Table { ... }

// Errors  package apperrors, not errors
import apperrors "github.com/bozz33/sublimego/errors"
```

### Package Structure

- One responsibility per package
- No circular imports
- `internal/` for packages that must not be imported externally
- Avoid `utils/`, `helpers/`, `common/`  name packages by what they do

---

## Testing

- Write unit tests for all new functionality
- Maintain or improve existing coverage
- Use table-driven tests where appropriate
- Mock external dependencies via interfaces

```go
// Table-driven test example
func TestGroupRows(t *testing.T) {
    tests := []struct {
        name     string
        rows     []any
        wantLen  int
    }{
        {"empty input", nil, 0},
        {"single group", []any{row1, row2}, 1},
        {"two groups", []any{row1, row3}, 2},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := GroupRows(tt.rows, col, grouping)
            if len(got) != tt.wantLen {
                t.Errorf("got %d groups, want %d", len(got), tt.wantLen)
            }
        })
    }
}
```

---

## Reporting Issues

### Bug Reports

Please include:

- Go version (`go version`)
- Operating system and architecture
- Steps to reproduce
- Expected vs actual behaviour
- Relevant error messages or stack traces

### Feature Requests

Please describe:

- The problem you are trying to solve
- Your proposed solution
- Alternatives you have considered
- Whether you are willing to implement it

---

## Pull Request Checklist

- [ ] Tests pass (`go test ./...`)
- [ ] No race conditions (`go test -race ./...`)
- [ ] No lint errors (`golangci-lint run`)
- [ ] No vet issues (`go vet ./...`)
- [ ] Templ templates regenerated if needed (`templ generate`)
- [ ] Documentation updated if behaviour changed
- [ ] Commit messages follow Conventional Commits

---

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
