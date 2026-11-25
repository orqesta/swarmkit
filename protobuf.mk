# Protobuf generation runs only in the isolated Docker build stage.
ROOTDIR := $(CURDIR)
PROTO_GOGO_DIR = $(shell go list -m -f '{{.Dir}}' github.com/gogo/protobuf)
PROTO_RAFT_DIR = $(shell go list -m -f '{{.Dir}}' go.etcd.io/raft/v3)
PROTO_INCLUDES = --proto_path=github.com/docker/swarmkit=$(ROOTDIR) --proto_path=github.com/gogo/protobuf=$(PROTO_GOGO_DIR) --proto_path=go.etcd.io/raft/v3=$(PROTO_RAFT_DIR) --proto_path=$(PROTO_GOGO_DIR) --proto_path=/usr/include
PROTO_IMPORT_MAPS = Mgogoproto/gogo.proto=github.com/gogo/protobuf/gogoproto,Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor,Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,Mgithub.com/docker/swarmkit/protobuf/plugin/plugin.proto=github.com/moby/swarmkit/v2/protobuf/plugin
PROTO_OPTIONS = plugins=grpc+deepcopy+storeobject+raftproxy+authenticatedwrapper,$(PROTO_IMPORT_MAPS)
PROTO_API = $(addprefix github.com/docker/swarmkit/,$(wildcard api/*.proto))
PROTO_PLUGIN = github.com/docker/swarmkit/protobuf/plugin/plugin.proto
PROTO_DEEPCOPY = github.com/docker/swarmkit/protobuf/plugin/deepcopy/testdata/test/deepcopy.proto
PROTO_RAFTPROXY = github.com/docker/swarmkit/protobuf/plugin/raftproxy/testdata/test/service.proto
PROTO_GENERATOR = --plugin=protoc-gen-gogoswarm=/tmp/protoc-gen-gogoswarm
PROTO_OUTPUT = /tmp/swarmkit-protobuf

.PHONY: generate
generate:
	go mod download
	go build -o /tmp/protoc-gen-gogoswarm ./cmd/protoc-gen-gogoswarm
	test "$$(protoc --version)" = "libprotoc 3.21.12"
	mkdir -p "$(PROTO_OUTPUT)"
	protoc $(PROTO_INCLUDES) $(PROTO_GENERATOR) --gogoswarm_out="$(PROTO_OPTIONS),import_path=github.com/moby/swarmkit/v2/api:$(PROTO_OUTPUT)" $(PROTO_API)
	protoc $(PROTO_INCLUDES) $(PROTO_GENERATOR) --gogoswarm_out="$(PROTO_OPTIONS),import_path=github.com/moby/swarmkit/v2/protobuf/plugin:$(PROTO_OUTPUT)" $(PROTO_PLUGIN)
	protoc $(PROTO_INCLUDES) $(PROTO_GENERATOR) --gogoswarm_out="$(PROTO_OPTIONS),import_path=github.com/moby/swarmkit/v2/protobuf/plugin/deepcopy/testdata/test:$(PROTO_OUTPUT)" $(PROTO_DEEPCOPY)
	protoc $(PROTO_INCLUDES) $(PROTO_GENERATOR) --gogoswarm_out="$(PROTO_OPTIONS),import_path=github.com/moby/swarmkit/v2/protobuf/plugin/raftproxy/testdata/test:$(PROTO_OUTPUT)" $(PROTO_RAFTPROXY)
	protoc $(PROTO_INCLUDES) --include_imports --descriptor_set_out=/tmp/swarmkit-api.pb $(PROTO_API)
	rm -f api/*.pb.go protobuf/plugin/plugin.pb.go
	cp -a "$(PROTO_OUTPUT)/github.com/docker/swarmkit/." .
