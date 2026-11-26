# TLA+ design specifications

This directory retains the formal specifications and models inherited from
SwarmKit, including the overall model and worker implementation. They are
design documentation, not inputs to the Go build or CI.

The inherited Makefile that acquired an unversioned third-party TLA+ container
has been removed. The repository does not install or provision TLA+ tooling.
The specifications may be inspected independently with an existing TLA+
environment; formal model verification is outside the Go build and validation
pipeline.

See the [TLA+ project](https://github.com/tlaplus/tlaplus) for the specification
language and its tooling.
