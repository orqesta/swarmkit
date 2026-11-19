# ADR-0003 — Handle closed watch channels explicitly

- **Recorded:** 2026-11-19
- **Status:** Accepted and CI-validated
- **Scope:** `orqesta/swarmkit`

## Context

After migrating the local event transport, cancelled subscriptions close their channels. Single-value receives could interpret the zero value from a closed channel as an actual event, causing invalid type conversions or unintended loop behavior.

## Decision

Use two-value channel receives (`event, ok := <-ch`) in affected watch consumers. A closed channel terminates observation or produces an explicit error where an RPC/stream requires continuity. No alternative or fallback event transport is added.

## Implementation

Commit `5d10408` updated the manager, service updates, jobs and Raft watchers. Commit `0d604ab` updated dispatcher, log broker and CSI consumers and cancellation handling.

## Consequences

Consumers terminate predictably on closure instead of interpreting zero values as real messages. This does not establish functional validation for a Docker Engine built with the latest fork.

## Validation

[SwarmKit CI 43931243961](https://github.com/orqesta/swarmkit/actions/runs/43931243961) and [validation 43931244130](https://github.com/orqesta/swarmkit/actions/runs/43931244130) passed on commit `0d604ab`.

## References

- [Commit 5d10408](https://github.com/orqesta/swarmkit/commit/5d10408d38dabe9ab9370c61ea9ce477670d8fea)
- [Commit 0d604ab](https://github.com/orqesta/swarmkit/commit/0d604ab7e0e18a84d05f9c667c9a82e896e25652)
- [ADR-0002 — local events](ADR-0002-typed-local-events.md)
- [Orqesta ADR-0009](https://docs.google.com/document/d/1-INdw3aEy3BIucm9pVEg0CLn5FivSVpZwODpsAhQy-Y/edit)
