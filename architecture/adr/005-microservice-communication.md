# ADR 005: Splitting the Monolith—Extracting the Email Sending Service and Choosing a Communication Protocol

**State**: Accepted.

**Date**: 2025-07-07.

**Author**: Oleksandr Prokhorov.

## Context

Our application is currently a monolith. We want to extract the email-sending functionality (via SMTP) into its own microservice, `Weather Sender Service`. We must choose how the existing `Scheduler` component will communicate with this new service.

Key requirements:

- **Low latency**: emails must be sent immediately after a trigger.
- **Quick implementation**: minimal ramp-up time to start delivering value.
- **Scalability**: both services should scale independently.
- **Minimal external dependencies**: avoid introducing heavy infra (such external components).

## Considered Options

### 1. gRPC
**Pros:**
- High performance via HTTP/2 and efficient binary serialization (Protocol Buffers).
- Strongly typed service definitions.
- Built-in deadlines, timeouts, streaming, and metadata support.

**Cons:**  
- Harder to debug compared to plain HTTP/JSON.
- Not directly consumable by typical browsers or REST tools without a proxy.
- Requires HTTP/2 support in proxies/gateways.

### 2. HTTP/REST
**Pros:**
- Ubiquitous and easy for developers to adopt.
- Simple to debug, monitor, and log with existing tools (curl, Postman).
- Broad client support (browsers, mobile apps).

**Cons:**
- Textual JSON payloads add overhead for large messages.
- No strict schema enforcement (though OpenAPI/Swagger can mitigate).
- Lacks built-in streaming semantics.

### 3. Message Queue (e.g. Kafka, RabbitMQ).
**Pros:**
- High reliability: messages persist even if the consumer is down
- Asynchronous decoupling, good for high-throughput event streams
- Independent scaling of producer/consumer

**Cons:**
- More infrastructure to maintain (broker, producers, consumers)
- Potential message delivery latency
- Not ideal for RPC-style immediate request/response

## Chosen Solution

We will use **gRPC** for synchronous communication between `Scheduler` and `Weather Sender Service`.

## Consequences

**Positive:**
- Very low latency between `Scheduler` and email service.
- Clear, strongly typed contract via `.proto` definitions.
- Easy to extend with new RPC methods without changing transport.

**Negative:**  
- More complex monitoring and tracing compared to HTTP/JSON.
- Need to maintain and version `.proto` schema files.
- All components must support HTTP/2.
