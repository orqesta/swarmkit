# Architecture decision records

This directory documents decisions implemented and tested in the Orqesta SwarmKit fork.

- [ADR-0001 — Respect stop-first grace periods](ADR-0001-stop-first-grace-period.md)
- [ADR-0002 — Use a single typed in-process event transport](ADR-0002-typed-local-events.md)
- [ADR-0003 — Handle closed watch channels explicitly](ADR-0003-watch-channel-closure.md)

**Current limitation:** `go.mod` and `go.sum` pin dependencies, but this repository still contains a tracked `vendor/` directory and inherited vendoring validation. Removing it is pending and must be implemented and validated separately; these ADRs do not claim otherwise.

The cross-repository history is also held in [Orqesta's Google Drive ADR register](https://drive.google.com/drive/folders/1Y-ErdX024KOy4zicZ1G35YSvYLUvTyno).
