# ADR-0002 — Use a single typed in-process event transport

- **Recorded:** 2026-11-19
- **Status:** Accepted and CI-validated
- **Scope:** `orqesta/swarmkit`

## Context

SwarmKit previously used `github.com/docker/go-events` for local event delivery. The Orqesta architecture requires a single in-process, typed event mechanism with explicit cancellation and bounded delivery, without a parallel legacy event bus.

## Decision

Use `github.com/jmeiracorbal/go-events` as the sole typed local event transport. Keep events as native Go values; they are not transported via NATS. Certificate retry jitter remains an independent certificate-authority concern.

## Implementation

The typed transport migration began in commit `eb52204`; subsequent commits synchronized the dependency graph, corrected typed channel consumers and validated compilation. `go.mod` pins `github.com/jmeiracorbal/go-events v0.1.2-0.20261115081746-6704f54534c5`.

## Consequences

There is one local event transport, with no second broadcaster, compatibility fallback or bridge. Cancellation closes watches and must be handled explicitly by consumers (ADR-0003). The repository still contains `vendor/`; its removal is a separate outstanding change.

## Validation

The final SwarmKit head `0d604ab` passed [CI 43931243961](https://github.com/orqesta/swarmkit/actions/runs/43931243961) and [validation 43931244130](https://github.com/orqesta/swarmkit/actions/runs/43931244130), including builds, regressions and vendoring checks.

## References

- [Migration commit eb52204](https://github.com/orqesta/swarmkit/commit/eb522047a3)
- [Current validated implementation](https://github.com/orqesta/swarmkit/commit/0d604ab7e0e18a84d05f9c667c9a82e896e25652)
- [ADR-0003 — watch closure](ADR-0003-watch-channel-closure.md)
- [Orqesta ADR-0009](https://docs.google.com/document/d/1-INdw3aEy3BIucm9pVEg0CLn5FivSVpZwODpsAhQy-Y/edit)
