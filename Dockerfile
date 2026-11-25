# syntax=docker/dockerfile:1@sha256:572cebbc2b15e6f84c077742952674889cb63cd5e9b80abdffcbe587b6f19183

FROM --platform=$BUILDPLATFORM golang:1.25-bookworm@sha256:3b4a11519ad929d1e1d261a12cff056f0c85b735253d7d861346b9c6f8b36437 AS gobase
ENV GOWORK=off GOFLAGS=-mod=readonly GOTOOLCHAIN=local \
    GOPROXY=https://proxy.golang.org GOSUMDB=sum.golang.org \
    GOPRIVATE="" GONOPROXY=none GONOSUMDB=none GOINSECURE=""
WORKDIR /workspace/swarmkit
RUN git config --global --add safe.directory /workspace/swarmkit

FROM gobase AS module-validate
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/go/pkg/mod \
    make --file=direct.mk module-validate

# protoc is a build compiler provided by the base distribution.
# Go dependencies and the generator plugin remain managed through Go Modules.
FROM gobase AS generate-base
RUN apt-get update && apt-get install -y --no-install-recommends \
    protobuf-compiler=3.21.12-3+deb12u1 \
    libprotobuf-dev=3.21.12-3+deb12u1 \
    libprotobuf32=3.21.12-3+deb12u1 \
    libprotobuf-lite32=3.21.12-3+deb12u1 \
    libprotoc32=3.21.12-3+deb12u1 \
    zlib1g-dev=1:1.2.13.dfsg-1 \
    && rm -rf /var/lib/apt/lists/*

FROM generate-base AS generate-build
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache \
    make --file=protobuf.mk generate

FROM generate-build AS generate-export
RUN mkdir /out && cp --parents api/*.pb.go protobuf/plugin/plugin.pb.go /out

FROM scratch AS generate-update
COPY --from=generate-export /out /

FROM generate-build AS generate-validate
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache <<EOT
  set -eu
  changes=$(git status --porcelain --untracked-files=all -- 'api/*.pb.go' 'protobuf/plugin/plugin.pb.go')
  if [ -n "$changes" ]; then
    printf '%s\n' "$changes" >&2
    git --no-pager diff -- 'api/*.pb.go' 'protobuf/plugin/plugin.pb.go'
    exit 1
  fi
  go test ./protobuf/plugin/deepcopy/testdata/test ./protobuf/plugin/raftproxy/testdata/test
EOT

FROM gobase AS lint
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    <<EOT
  set -e
  go -C tools build -o /usr/local/bin/golangci-lint github.com/golangci/golangci-lint/v2/cmd/golangci-lint
  config=$(pwd)/.golangci.yml
  golangci-lint run --config "$config" ./...
  (
    cd swarmd
    golangci-lint run --config "$config" ./...
  )
EOT

FROM gobase AS fmt-proto
RUN --mount=type=bind,target=. \
    make fmt-proto

FROM gobase
COPY direct.mk .
COPY go.* .
RUN make --file=direct.mk setup
COPY . .
CMD ["make"]
