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

Generation requires `protoc` 3.21.12 and its standard schemas, supplied by
Debian's pinned `protobuf-compiler` and `libprotobuf-dev` packages
(3.21.12-3+deb12u1) in the development image. The generation target rejects
other compiler versions.
There are no manual binary downloads or installer scripts. The custom
`protoc-gen-gogoswarm` compiler plugin is built from this repository using
Go Modules. The gogo/protobuf schema directory is resolved from the selected
Go module, not copied into the source tree.

```sh
make generate
```

Regeneration retains the historical protobuf import names and checks the
generated Go sources and API descriptor.

## Debugging

Delve is a pinned Go tool in `tools/debug/`. Run `hack/debug` with the test
package and arguments; it builds Delve using the committed module manifests
and starts a headless debugging session without installing a global binary.

## Container build

```sh
DOCKER_SWARMKIT_USE_CONTAINER=1 make image
DOCKER_SWARMKIT_USE_CONTAINER=1 make all
```

The development image is explicitly built using the digest-pinned base and
version-pinned system packages. The Makefile tags it with the current Git
revision; container runs never pull or build an image implicitly. Rebuild the
image explicitly when editing its build inputs locally.

The `docker-bake.hcl` validation targets run linting, module validation and
generation validation. GolangCI-Lint is built from the pinned Go tool module
in `tools/`, using its committed `go.mod` and `go.sum`. CI uses Go Modules for compilation and tests. The
container runner mounts source directly through Docker; it does not require
`docker-sync` or shell-built command strings.
