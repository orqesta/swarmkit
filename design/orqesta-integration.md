# Orqesta integration

This document describes implementation-specific behavior in the Orqesta SwarmKit fork. The underlying scheduler, orchestration, Raft and task models remain those of SwarmKit.

## Stop-first updates

The updater accounts for the outgoing task's `StopGracePeriod` instead of unconditionally using the previous fixed one-minute wait. The effective wait is at least the existing timeout, and for longer grace periods it includes an additional five seconds.

This behavior is covered by the `TestUpdaterStopGracePeriod` regression test. Its successful CI run does not replace a real-cluster test of a long shutdown.

## Typed local event delivery

The fork uses [`github.com/jmeiracorbal/go-events`](https://github.com/jmeiracorbal/go-events) for in-process watches and notifications, including tasks, manager state, leadership, certificates and logs. Events remain typed Go values. They are not sent over a separate NATS bus.

The local transport has one lifecycle: consumers subscribe, process events, and cancel or close their subscriptions. The transport can close watch channels on cancellation; consumers must check the `ok` value of channel receives. Closure terminates observation or produces an explicit error where a stream requires continued delivery. This avoids interpreting a zero value from a closed channel as an event.

Certificate retry jitter remains independent of the event transport in `ca/backoff.go`. There is no parallel legacy local event bus or compatibility fallback.

## Go dependencies and repository state

The repository declares dependencies in `go.mod` and records their checksums in `go.sum`. The local event dependency is pinned there. **The fork currently still tracks a `vendor/` directory and related inherited validation.** A move to exclusively module-cache dependency resolution requires its own implementation and CI verification; it must not be inferred from the event transport migration.

## Integration boundary

Moby currently pins an earlier Orqesta SwarmKit revision. Therefore the newer local event transport and subscription-close changes in this fork have not yet been proven in an Orqesta-built Docker Engine. Full runtime validation is separate from the SwarmKit unit and integration test suite.
