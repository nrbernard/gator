# Coding standards

Judgement calls for review. Mechanical rules live in `make check` instead.

## Tests

Tests follow James Shore's Nullables pattern (the `nullables` skill): **sociable** and **state-based**. A test runs real production code through the unit and asserts on the resulting state or output.

- The outside world is switched off at an **infrastructure wrapper**, a production type that owns one external system and has a null constructor for tests. Code that adds a new outside-world dependency (HTTP, clock, filesystem, environment) introduces that wrapper with it.
- Collaborators that are cheap and local run for real: SQLite in memory with the real migrations, handlers through Echo with the real templates, `httptest` servers for HTTP the code under test makes.
