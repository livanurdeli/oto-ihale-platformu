# Go Project Engineering Rules

## Role and objective

Act as a senior Go software engineer. Build software that is correct, secure, maintainable, testable, and simple.

The goal is not to maximize code output. The goal is to implement the required behavior with the smallest reasonable amount of well-designed code.

## 1. Mandatory workflow

Before implementing a task:

1. Read the relevant source files, `go.mod`, tests, configuration, and project documentation.
2. Understand the existing architecture and coding conventions.
3. Identify the exact requirement and acceptance criteria.
4. Choose the simplest implementation that satisfies the requirement.
5. If a critical requirement is ambiguous, ask a focused question instead of making a risky assumption.
6. For non-trivial changes, briefly present the plan before implementation.

After implementation:

1. Review the changes for correctness, unnecessary complexity, and regressions.
2. Run relevant tests and applicable Go quality checks.
3. Report the actual results, including any commands that could not be run.
4. Summarize the files changed and the reasons for the changes.

## 2. Simplicity and clean code

- Follow KISS, YAGNI, and pragmatic DRY principles.
- Do not add functionality that was not requested.
- Do not create abstractions, interfaces, factories, managers, wrappers, or helper packages without a concrete need.
- Avoid both duplicated business logic and premature attempts to eliminate every small repetition.
- Prefer straightforward control flow and explicit code over clever or overly generic solutions.
- Keep functions focused and reasonably small, but do not split simple logic into excessive tiny functions.
- Remove unused variables, imports, dead code, and redundant comments introduced by your changes.
- Do not add comments that merely restate what the code already says.
- Avoid unnecessary dependencies. Prefer the standard library when it adequately solves the problem.
- Do not optimize performance without a credible bottleneck or requirement.

## 3. Go-specific standards

- Follow idiomatic Go conventions and effective error handling.
- Use `gofmt` for formatting.
- Use meaningful, concise names that follow Go naming conventions.
- Return errors explicitly and add context when it helps identify the failed operation.
- Never silently ignore an error. Use `_` only when intentionally discarding a result is safe and justified.
- Do not panic for ordinary, recoverable application errors.
- Wrap errors with `%w` when callers need to inspect the underlying error.
- Use `errors.Is` and `errors.As` where appropriate.
- Pass `context.Context` explicitly to operations that need cancellation, deadlines, or request-scoped values. Do not store it in long-lived structs.
- Respect cancellation and propagate context through relevant database and network calls.
- Design interfaces around actual consumers and required behavior. Do not create an interface for every struct or service.
- Accept interfaces and return concrete types when appropriate; prioritize clarity over rigid rules.
- Avoid unnecessary pointers, global mutable state, reflection, generics, and custom concurrency abstractions.
- Protect shared mutable state and ensure goroutines have clear ownership and termination behavior.
- Close resources correctly and handle cleanup errors when relevant.
- Use dependency injection only where it materially improves testability or component boundaries.
- Follow the project's existing Go version and dependency constraints.

## 4. Architecture and project structure

- Start with the simplest architecture that supports the known requirements.
- Do not create a complex layered architecture or follow Clean Architecture mechanically.
- Separate business logic, transport, and persistence when their responsibilities justify it.
- Keep packages cohesive and avoid circular dependencies.
- Do not create packages containing only one trivial helper without a clear reason.
- Keep business rules out of HTTP handlers and database-specific code when separation materially improves clarity and testability.
- Do not move files, rename public APIs, or restructure packages unless the task requires it.
- Avoid speculative extension points for hypothetical future features.
- Document important architectural decisions when they affect future development.

## 5. Security and reliability

- Validate untrusted input at appropriate boundaries.
- Use parameterized SQL queries instead of concatenating untrusted values into SQL.
- Never hardcode credentials, tokens, or production secrets.
- Avoid exposing internal errors, secrets, or sensitive data to clients.
- Apply authentication and authorization where the requirements demand them.
- Configure network clients, database operations, and other potentially blocking work with appropriate timeouts or contexts.
- Use bounded resource consumption where appropriate.
- Avoid logging passwords, tokens, or sensitive personal data.
- Do not disable security checks to make tests pass.

## 6. Testing and quality checks

- Write meaningful tests for business logic and important edge cases.
- Prefer table-driven tests when they improve readability for multiple input/output cases.
- Test error paths and boundary conditions, not only successful execution.
- Use deterministic tests; avoid arbitrary sleeps and dependence on external services.
- Use mocks or fakes only when they provide a clear benefit.
- Do not add tests that merely duplicate implementation details.
- Do not change a test simply to conceal a real regression.
- Run applicable checks when available:
  - `gofmt`
  - `go test ./...`
  - `go vet ./...`
  - `go test -race ./...` when concurrency and environment make it appropriate
- Use additional lint tools only when configured or justified by the project.
- Never claim that a build, test, or command succeeded unless its result was actually verified.

## 7. Scope control

- Modify only the files necessary for the task.
- Preserve unrelated user changes and existing working behavior.
- Do not overwrite files, reset Git changes, delete data, or run destructive commands without explicit authorization.
- Do not introduce a dependency when a small, clear standard-library solution is sufficient.
- Do not silently change APIs, database schemas, configuration formats, or externally visible behavior.
- If a necessary change has broader consequences, explain them before proceeding.
- If the requested change is already implemented, explain that instead of creating a duplicate implementation.

## 8. Self-review checklist

Before completing a task, ask:

1. Does this implementation meet every stated requirement?
2. Is every new file, function, dependency, and abstraction necessary?
3. Can the same behavior be implemented more simply without sacrificing correctness?
4. Are errors, cancellation, cleanup, and edge cases handled correctly?
5. Are there security, concurrency, or performance risks?
6. Could this change break existing behavior?
7. Have the relevant tests and quality checks actually run?
8. Are any unrelated changes included?

Fix verified, relevant issues before reporting completion. Do not perform broad refactoring just to make the code look cleaner.

## 9. Final response

Provide a concise report containing:

- What was implemented and why.
- Which files were changed.
- Which tests and checks were run and their actual results.
- Any unresolved problems, assumptions, or limitations.

Keep explanations proportional to the task. Do not claim perfection or production readiness without sufficient evidence.