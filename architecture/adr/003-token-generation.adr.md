# ADR 003: Token generation algorithm consideration

**State**: Accepted.

**Date**: 2025-06-09.

**Author**: Oleksandr Prokhorov.

## Context

It is necessary to chose algorithm for Token generation. It should be:
- Deterministic.
- With low or no Collision Rate.
- Efficient.
- Have Avalanche Effect.
- Irreversible.

## Considered Options

### SHA256
**Pros:**
- Strong collision- and preimage-resistance.
- Support in standard library (`crypto/sha256`).
- Good performance/security trade-off for web services.
- Excellent avalanche properties (tiny input change → unpredictable output).

**Cons:**
- Produces a 32-byte digest, so encoded tokens are relatively long (e.g. 64 hex chars).
- Slightly higher CPU overhead than non-cryptographic hashes.

### KECCAK256
**Pros:**
- Basis of the SHA-3 standard with high security margins.
- Strong collision- and preimage-resistance.
- Comparable avalanche effect to SHA-2.

**Cons:**
- Not in Go’s stdlib (requires external dependency).
- Marginally slower than SHA256 on some platforms.
- Same long digest size as SHA256.

### MD5
**Pros:**
- Very fast hashing.
- 16-byte digest yields shorter tokens (e.g. 32 hex chars).

**Cons:**
- Cryptographically broken—high collision risk.
- Unsuitable for security-sensitive token generation.
- Weaker avalanche effect.

## Chosen Solution
SHA256 was chosen.

## Consequences
**Positive:**
- Tokens are collision-resistant and secure.
- Native Go support simplifies implementation and reduces dependencies.
- Good balance of speed and security.

**Negative:**
- Longer encoded tokens (64 hex chars).
- Higher CPU cost per hash compared to MD5.