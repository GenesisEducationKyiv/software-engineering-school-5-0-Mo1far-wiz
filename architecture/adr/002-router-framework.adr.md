# ADR 002: Router framework consideration

**State**: Accepted.

**Date**: 2025-06-09.

**Author**: Oleksandr Prokhorov.

## Context

It is necessary to chose routing framework for:
- Running server and accepting connections.
- Handling requests and send responses.
- Maintaining clean, easy-to-use and understand code.
- Supporting middleware for cross-cutting concerns (logging, CORS, recovery).

## Considered Options

### Gin
**Pros:**
- High performance and low latency.
- Built-in routing groups and path parameter parsing.
- Easy JSON binding & validation.
- Rich ecosystem of middleware (logging, recovery, CORS).
- Large community and good documentation.

**Cons:**
- Larger binary size.
- Uses reflection under the hood (some overhead).
- Less modular control compared to minimal routers.

### Chi
**Pros:**
- Lightweight, composable router.
- No reflection—predictable performance.
- Middleware chaining using standard `net/http` handlers.
- Small binary and fast compile times.

**Cons:**
- Requires more boilerplate code for binding/validation.
- Fewer built-in features out-of-the-box.
- Smaller community than Gin.

### net/http
**Pros:**
- Part of Go standard library.
- Full control over request lifecycle and performance tuning.
- Depending binary footprint.

**Cons:**
- Manual routing setup (mux or custom logic needed).
- No built-in middleware support.
- More boilerplate for JSON binding, validation and error handling.

## Chosen Solution
Gin was chosen.

## API Schema
You can access postman collection [on this url](https://www.postman.com/avionics-operator-63001856/workspace/genesis-weather).

## Consequences
**Positive:**
- Fast development with built-in features.
- Consistent error handling and recovery across all routes.
- Clear route grouping and middleware application.
- Strong community support for plugins and extensions.

**Negative:**
- Increased binary size.
- Reflection-based internals may impact some benchmark results.
- Slightly less granular control than using `net/http` directly.