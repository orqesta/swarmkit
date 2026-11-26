# Protobuf API

The `.proto` files are the source of truth for the SwarmKit API. Do not edit
the generated `.pb.go` files directly.

```sh
make generate
make generate-validate
```

Generation uses the official compiler in an isolated Docker build and exports
only the Go bindings required by downstream Go module consumers. The API
descriptor is compiled during generation and is not stored in Git. Historical
API contracts can be reconstructed from the versioned schemas at a release
revision.
