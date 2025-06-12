# ADR 004: CI workflow considerations

**State**: Accepted.

**Date**: 2025-06-09.

**Author**: Oleksandr Prokhorov.

## Context

To ensure a consistent code style across all developers and catch bad practices, we need an automated linting step in our CI pipeline.

## Considered Options

### GitHub Actions
**Pros:**
- Native to our code host — no extra infrastructure.
- Easy to configure.
- Built-in caching for Go modules and linters.
- Visibility in PR UI with pass/fail status.

**Cons:**
- Tied to GitHub (migrating to another host would require rewriting workflows).
- CI runs add a few seconds of latency to each push.

### Jenkins
**Pros:**
- Full control over build agents and environment.
- Can integrate with on-prem resources.

**Cons:**
- Requires maintaining separate infrastructure.
- Longer setup and onboarding time.
- Overkill for simple lint checks.

### GitLab CI
**Pros:**
- Similar ease of use if we were on GitLab.
- Shared runners handle scaling automatically.

**Cons:**
- We’re not using GitLab — duplicate tooling.
- Unnecessary complexity given GitHub hosting.

### Local pre-commit hook
**Pros:**
- Immediate feedback before commit.
- No CI runtime cost.

**Cons:**
- Developers can bypass or disable hooks.
- Enforcement is inconsistent across contributors.

## Chosen Solution

Use **GitHub Actions** to run `golangci-lint` on every push request.

## Consequences

**Positive:**
- Enforced, consistent linting on all branches and PRs.
- No additional servers or maintenance burden.
- Clear pass/fail indicators in GitHub UI.

**Negative:**
- Slight increase in CI cycle time (2–5s per run).
- Dependency on GitHub Actions availability and rate limits.
