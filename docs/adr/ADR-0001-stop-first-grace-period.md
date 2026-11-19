# ADR-0001 — Respect stop-first grace periods

- **Recorded:** 2026-11-19
- **Status:** Accepted; CI-validated, not verified by a live cluster shutdown test
- **Scope:** `orqesta/swarmkit`

## Context

During stop-first service updates the updater could start replacement tasks after a fixed one-minute timeout, even when the outgoing task specified a longer `StopGracePeriod`.

## Decision

Wait for the maximum of the existing timeout and the outgoing task's declared grace period plus five seconds. Keep one updater and scheduler implementation; do not introduce alternate execution paths.

## Implementation

Changes to the updater and the `TestUpdaterStopGracePeriod` regression test were merged through [PR #1](https://github.com/orqesta/swarmkit/pull/1) in commit `cf6d063`.

## Consequences

Stop-first updates respect longer declared shutdown windows. This does not independently validate the fork's behavior on a live multinode deployment or fix unrelated overlay networking behavior.

## Validation

The fork's [CI](https://github.com/orqesta/swarmkit/actions/runs/43265651355) and [validation](https://github.com/orqesta/swarmkit/actions/runs/43265651192) workflows passed after the merge. A real-cluster 90-second shutdown reproduction was not performed.

## References

- [Merge commit cf6d063](https://github.com/orqesta/swarmkit/commit/cf6d063b76dfb48e9c22fc21d57c95181357dac3)
- [Orqesta ADR-0002](https://docs.google.com/document/d/1fpQTDDSFAWJYi3bR__sdR_udaa8lfNfJK0YFLlhf0SY/edit)
