<!-- Context: project-intelligence/workflow | Priority: high | Version: 1.0 | Updated: 2026-02-16 -->

# Code Workflow

**Purpose**: Development workflow and code quality practices for finsplitter.
**Last Updated**: 2026-02-16

## Clean Code Principles
- **Constants**: No magic numbers, use named constants
- **Names**: Descriptive, explain purpose (avoid unclear abbreviations)
- **Comments**: Explain "why", not "what" (code should be self-documenting)
- **DRY**: Extract repeated code into reusable functions
- **SRP**: Each function does one thing

## Development Workflow
```bash
make start-dev        # Hot reload dev environment
make generate         # Run sqlc + mocks
make test            # Run tests
make code-check      # Format + lint
```

## Testing
- Write tests before fixing bugs
- Table-driven tests with clear cases
- Use testify for assertions
- Mock external dependencies

## Code Review Guidelines
- Verify information before presenting
- File-by-file changes
- No unnecessary confirmations
- Preserve existing code
- Single chunk edits

## Quality Commands
```bash
make format           # golangci-lint fmt
make code-check      # lint + format check
make test -v         # verbose test output
```
