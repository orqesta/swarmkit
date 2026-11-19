# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25
ARG BASE_DEBIAN_DISTRO="bookworm"
ARG GOLANG_IMAGE="golang:${GO_VERSION}-${BASE_DEBIAN_DISTRO}"

ARG PROTOC_VERSION=3.14.0
ARG GOLANGCI_LINT_VERSION=v2.12.2

# gobase
FROM --platform=$BUILDPLATFORM ${GOLANG_IMAGE} AS gobase
RUN apt-get update && apt-get install -y --no-install-recommends git make rsync
WORKDIR /go/src/github.com/docker/swarmkit
RUN git config --global --add safe.directory /go/src/github.com/docker/swarmkit

FROM gobase AS module-validate
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/go/pkg/mod \
    make --file=direct.mk module-validate

FROM gobase AS protoc-gen-gogoswarm
RUN --mount=type=bind,target=.,rw \
    --mount=type=cache,target=/root/.cache \
    make bin/protoc-gen-gogoswarm && mv bin/protoc-gen-gogoswarm /usr/local/bin/

FROM gobase AS generate-base
RUN apt-get --no-install-recommends install -y unzip
ARG PROTOC_VERSION
ARG TARGETOS
ARG TARGETARCH
RUN <<EOT
  set -e
  arch=$(echo $TARGETARCH | sed -e s/amd64/x86_64/ -e s/arm64/aarch_64/)
  wget -q https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-${TARGETOS}-${arch}.zip
  unzip protoc-${PROTOC_VERSION}-${TARGETOS}-${arch}.zip -d /usr/local
EOT

FROM generate-base AS generate-build
RUN --mount=type=bind,target=.,rw \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=from=protoc-gen-gogoswarm,source=/usr/local/bin/protoc-gen-gogoswarm,target=/usr/bin/protoc-gen-gogoswarm <<EOT
  set -ex
  make generate
  mkdir /out
  git ls-files -m --others -- '**/*.pb.go' 'api/api.pb.txt' | tar -cf - --files-from - | tar -C /out -xf -
EOT

FROM scratch AS generate-update
COPY --from=generate-build /out /

FROM gobase AS generate-validate
RUN --mount=type=bind,target=.,rw \
    --mount=type=bind,from=generate-build,source=/out,target=/generated <<EOT
  set -e
  git add -A
  if [ "$(ls -A /generated)" ]; then
    cp -rf /generated/* .
  fi
  diff=$(git status --porcelain -- '**/*.pb.go' 'api/api.pb.txt')
  if [ -n "$diff" ]; then
    echo >&2 'ERROR: The result of "go generate" differs. Please update with "make generate"'
    echo "$diff"
    exit 1
  fi
EOT

FROM golangci/golangci-lint:${GOLANGCI_LINT_VERSION} AS golangci-lint
FROM gobase AS lint
RUN apt-get install -y --no-install-recommends libgcc-11-dev libc6-dev
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=from=golangci-lint,source=/usr/bin/golangci-lint,target=/usr/bin/golangci-lint <<EOT
  set -e
  config=$(pwd)/.golangci.yml
  golangci-lint run --config "$config" ./...
  (
    cd swarmd
    GOWORK=off golangci-lint run --config "$config" ./...
  )
EOT

FROM gobase AS fmt-proto
RUN --mount=type=bind,target=. \
    make fmt-proto

# use generate-base to have protoc available
FROM generate-base
ENV GO111MODULE=on
# install the dependencies from `make setup`
# we only copy `direct.mk` to avoid busting the cache too easily
COPY direct.mk .
COPY go.* .
RUN make --file=direct.mk setup
# now we can copy the rest
COPY . .
# default to just `make`. If you want to change the default command, change the
# default make command, not this command.
CMD ["make"]
