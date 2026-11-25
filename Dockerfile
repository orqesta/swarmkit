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

FROM gobase AS generate-build
RUN --mount=type=bind,target=.,rw \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache <<EOT
  set -ex
  make generate
  mkdir /out
  git ls-files -m --others --exclude-standard -z -- '*.pb.go' 'api/api.pb.txt' > /tmp/generated-files
  tar --null -T /tmp/generated-files -cf /tmp/generated.tar
  tar -C /out -xf /tmp/generated.tar
EOT

FROM scratch AS generate-update
COPY --from=generate-build /out /

FROM gobase AS generate-validate
RUN --mount=type=bind,target=.,rw \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache \
    make checkprotos

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
