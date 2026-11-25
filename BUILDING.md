# Building SwarmKit

SwarmKit is a Go 1.25 module. Build and test with Go Modules; dependencies are
declared in `go.mod` and verified with `go.sum`. Neither the repository nor
the build process uses a `vendor/` directory. Public modules are verified against
Go's checksum database (`sum.golang.org`) and the committed `go.sum` files.

## Local build

```sh
make module-validate
make binaries
make test swarmd-tests
```

The standard `Makefile` provides the `build`, `test`, `integration-tests`,
`swarmd-tests` and `module-validate` targets. `module-validate` checks
dependency integrity and ensures that the application and tooling Go modules are tidy without editing
the manifests. CI builds each module with `GOWORK=off`, so the committed
module manifests are the source of truth; the repository does not commit an
alternative workspace dependency graph. The Makefile, development container,
CI and debugger use proxy.golang.org without a direct-source fallback, verify
public modules against sum.golang.org, and disable automatic toolchain downloads.

## Protobuf generation

The `swarm-protogen` command compiles schemas using
`github.com/bufbuild/protocompile` and calls the existing SwarmKit Gogo
generator in the same Go process. All dependencies are selected by `go.mod`
and verified by `go.sum`. Gogo, Raft and the standard Protobuf schemas are
read from the selected Go module directories; no external Protobuf compiler,
system header directory, binary download or installer is needed.

```sh
make generate
make checkprotos
```

The compiler preserves historical schema names, custom options and source
locations. The generator still uses the project's existing Gogo plugins.
The API descriptor is emitted by Go from the compiled descriptor set, without
source locations, before generator-specific transformations.

## Debugging

Delve is a pinned Go tool in `tools/debug/`. Run `hack/debug` with the test
package and arguments; it builds Delve using the committed module manifests
and starts a headless debugging session without installing a global binary.

## Container build

```sh
DOCKER_SWARMKIT_USE_CONTAINER=1 make image
DOCKER_SWARMKIT_USE_CONTAINER=1 make all
```

The development image is explicitly built using the digest-pinned Go base.
No additional Protobuf packages are installed. The Makefile tags it with the current Git
revision; container runs never pull or build an image implicitly. Rebuild the
image explicitly when editing its build inputs locally.

The `docker-bake.hcl` validation targets run linting, module validation and
generation validation. GolangCI-Lint is built from the pinned Go tool module
in `tools/`, using its committed `go.mod` and `go.sum`. CI uses Go Modules for compilation and tests. The
container runner mounts source directly through Docker; it does not require
`docker-sync` or shell-built command strings.
