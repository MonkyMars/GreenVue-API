# GreenVue.eu Backend Tests

This directory contains test files for various components of the GreenVue.eu backend application.

## Test Structure

The test directory is organized by component:

- `auth/`: Tests for authentication-related functionality
- `db/`: Tests for database repositories and mock implementations
- `validation/`: Tests for validation utilities
- `security/`: Tests for security-related functionality

## Running Tests

To run all tests, from the project root directory:

```bash
go test ./test/...
```

To run tests for a specific package:

```bash
go test ./test/validation/
go test ./test/auth/
go test ./test/db/
go test ./test/security/
```

To run a specific test:

```bash
go test ./test/validation/ -run TestUsernameValidation
```

To run tests with verbose output:

```bash
go test -v ./test/...
```

## Test Coverage

To get test coverage:

```bash
go test -cover ./test/...
```

For a detailed coverage report:

```bash
go test -coverprofile=coverage.out ./test/...
go tool cover -html=coverage.out
```

## Mock Implementations

The `db/mock_repository.go` file contains mock implementations of the repository interfaces. These mocks provide in-memory storage for testing without requiring a real database connection.

## Environmental Variables for Testing

Some tests, especially those for JWT functionality, require environmental variables to be set. These are automatically set within the tests and cleaned up afterward.
