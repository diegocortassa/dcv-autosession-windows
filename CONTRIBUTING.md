## Contributing

Thank you for considering contributing to dcv-autosession.

### Getting Started

1. Fork the repository.
2. Ensure you have the build requirements from the README.
3. Run `go mod tidy` and `make build` to verify your environment works.

### Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go) conventions.
- Avoid adding comments unless the logic is non-obvious. Let the code speak.

### Structure

- `cmd/dcv-autosession/main.go` - application entry point.
- `internal/` - all application logic, organized by concern:
  - `config/` - INI configuration parsing.
  - `dcv/` - Interface to DCV commands.
  - `logger/` - logging setup.
  - `reaper/` - periodic cleanup of idle sessions with no connections.
  - `service/` - business logic glue and Windows service management.
  - `version/` - build version metadata.

- Shared logic must be extracted into the appropriate package, no duplication.

### Pull Requests

- Keep changes focused, one feature or fix per PR.
- Run `make build` before submitting to confirm the project compiles.
- Write clear commit messages following the existing style.
- Ensure new code follows the package conventions.
