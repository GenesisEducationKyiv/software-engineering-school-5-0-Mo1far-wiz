**Testing Instructions**

This document explains how to run both unit and integration tests for the project.

---

## Prerequisites

1. **Docker & Docker Compose**

   * Used to spin up the test PostgreSQL database.
2. **Taskfile (`task`)**

   * Simplifies command runners for linting, building, and testing.
3. **Environment file (`.test.env`)**

   * ENVs are not loaded from this file, it is just an example of data that is used for tests
---

## Environment Variables

```dotenv
# Test database configuration
test_db_name=test_weather
test_db_user=test
test_db_password=password
test_db_host=127.0.0.1
test_db_port=5433
test_db_ssl_mode=disable
```

---

## Taskfile Overview

The `Taskfile.yml` includes two test-related tasks:

* **`test-unit`**: Runs fast unit tests (no external dependencies).
* **`test-integration`**: Spins up a Docker test database, runs tests against `./internal/api/...`, then tears down the database.
* **`test`**: Runs both tests (dependencies same as for integration tests).

---

## Running Tests

From the project root, execute:

```bash
task test
```

---

## Troubleshooting

* **Database connection errors**: Verify Docker is running and port `5433` is free.
* **Missing Task binary**: Install Taskfile via Homebrew or from [https://taskfile.dev](https://taskfile.dev).
