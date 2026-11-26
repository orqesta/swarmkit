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

The official `protoc` compiler and its standard schemas are provisioned only
in the isolated Docker Buildx generation stage from version-pinned Debian
packages and a dated Debian Snapshot archive. The snapshot is the only APT
source in that stage. APT verifies signed release metadata and package hashes;
only the expiration check for historical metadata is disabled. The project-owned Go plugin is compiled in that stage using Go
Modules. Gogo and Raft schemas are resolved directly from their selected
modules. The development image does not contain the Protobuf compiler.

```sh
make generate
make generate-validate
```

Generation exports only the API and plugin Go bindings required by consumers
of the SwarmKit Go module. Go does not generate source files when importing a
module, so these bindings remain versioned. The API descriptor, generator
binary and generated test fixtures exist only inside the generation build.
The validation target checks the distributed bindings and executes the
generated deepcopy and raftproxy tests in that same isolated environment.

## Debugging

Delve is a pinned Go tool in `tools/debug/`. Run `hack/debug` with the test
package and arguments; it builds Delve using the committed module manifests
and starts a headless debugging session. Both Delve and its generated test
executable are created in a temporary directory outside the checkout and
removed when the session ends. No global binary is installed.

## Coverage

The `coverage` and `coverage-integration` targets run the existing coverage
tests with temporary profiles outside the source tree. The coverage percentages
are reported in test output; profiles are removed when the targets exit.

## Container build

```sh
DOCKER_SWARMKIT_USE_CONTAINER=1 make image
DOCKER_SWARMKIT_USE_CONTAINER=1 make all
```

The development image is explicitly built using the digest-pinned base and
Go build inputs. The Makefile tags it with the current Git
revision; container runs never pull or build an image implicitly. Rebuild the
image explicitly when editing its build inputs locally.

The `docker-bake.hcl` validation targets run linting, module validation and
generation validation. GolangCI-Lint is built from the pinned Go tool module
in `tools/`, using its committed `go.mod` and `go.sum`. CI uses Go Modules for compilation and tests. The
container runner mounts source directly through Docker; it does not require
`docker-sync` or shell-built command strings.
