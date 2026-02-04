# Contributing to SublimeGo

Thank you for your interest in contributing to SublimeGo! This document provides guidelines and information for contributors.

## Getting Started

### Prerequisites

- Go 1.24 or later
- Git

### Setting Up the Development Environment

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/bozz33/sublimego.git
   cd sublimego
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Run tests to ensure everything works:
   ```bash
   go test ./...
   ```

## Development Workflow

### Branching Strategy

- `main` - Stable release branch
- `dev` - Development branch for integration
- Feature branches should be created from `dev`

### Making Changes

1. Create a new branch for your feature or fix:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following the code style guidelines

3. Write or update tests as needed

4. Run tests and ensure they pass:
   ```bash
   go test ./... -count=1
   ```

5. Run the linter:
   ```bash
   golangci-lint run
   ```

6. Commit your changes with a clear message:
   ```bash
   git commit -m "feat: add new feature description"
   ```

### Commit Message Format

We follow conventional commits:

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks

### Pull Request Process

1. Push your branch to your fork
2. Open a Pull Request against the `dev` branch
3. Fill in the PR template with relevant information
4. Wait for review and address any feedback

## Code Style Guidelines

### Go Code

- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and small
- Handle errors explicitly

### Project Structure

```
sublimego/
├── actions/        # Action system
├── auth/           # Authentication
├── engine/         # Core panel engine
├── form/           # Form builder
├── table/          # Table builder
├── middleware/     # HTTP middlewares
├── ui/             # UI components
├── validation/     # Validation rules
├── widget/         # Dashboard widgets
├── internal/       # Private packages
├── cmd/            # CLI commands
└── views/          # Templ templates
```

### Testing

- Write unit tests for new functionality
- Maintain or improve test coverage
- Use table-driven tests where appropriate
- Mock external dependencies

## Reporting Issues

### Bug Reports

When reporting bugs, please include:

- Go version (`go version`)
- Operating system
- Steps to reproduce
- Expected behavior
- Actual behavior
- Error messages or logs

### Feature Requests

For feature requests, please describe:

- The problem you're trying to solve
- Your proposed solution
- Any alternatives you've considered

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on the code, not the person
- Help others learn and grow

## Questions?

If you have questions, feel free to:

- Open an issue with the `question` label
- Join discussions in existing issues

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
