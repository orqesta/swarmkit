# Building SwarmKit

SwarmKit is a Go 1.25 module. Build and test with Go Modules; dependencies are
declared in `go.mod` and verified with `go.sum`. Neither the repository nor
the build process uses a `vendor/` directory. Public modules are verified against
Go's checksum database (`sum.golang.org`) and the committed `go.sum` files.

## Local build

```sh
go mod download
go build ./...
go test ./...
cd swarmd && GOWORK=off go mod download && GOWORK=off go test ./...
```

The standard `Makefile` provides the `build`, `test`, `integration-tests`,
`swarmd-tests` and `module-validate` targets. `module-validate` checks
dependency integrity and ensures that the application and tooling Go modules are tidy without editing
the manifests. CI builds each module with `GOWORK=off`, so the committed
module manifests are the source of truth. The development workspace is optional.

## Protobuf generation

Generation requires the `protoc` compiler, installed from the operating
system package manager (Debian's `protobuf-compiler` and `libprotobuf-dev` in the development image).
There are no manual binary downloads or installer scripts. The custom
`protoc-gen-gogoswarm` compiler plugin is built from this repository using
Go Modules. The gogo/protobuf schema directory is resolved from the selected
Go module, not copied into the source tree.

```sh
make generate
```

Regeneration retains the historical protobuf import names and checks the
generated Go sources and API descriptor.

## Container build

```sh
DOCKER_SWARMKIT_USE_CONTAINER=1 make
```

The `docker-bake.hcl` validation targets run linting, module validation and
generation validation. GolangCI-Lint is built from the pinned Go tool module
in `tools/`, using its committed `go.mod` and `go.sum`. CI uses Go Modules for compilation and tests. The
container runner mounts source directly through Docker; it does not require
`docker-sync` or shell-built command strings.
