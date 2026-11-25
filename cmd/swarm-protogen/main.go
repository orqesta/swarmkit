package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/plugin"
	"github.com/gogo/protobuf/vanity"
	"github.com/gogo/protobuf/vanity/command"
	_ "github.com/moby/swarmkit/v2/protobuf/plugin/authenticatedwrapper"
	_ "github.com/moby/swarmkit/v2/protobuf/plugin/deepcopy"
	_ "github.com/moby/swarmkit/v2/protobuf/plugin/raftproxy"
	_ "github.com/moby/swarmkit/v2/protobuf/plugin/storeobject"
)

const (
	sourcePrefix = "github.com/docker/swarmkit/"
	generatorOptions = "plugins=grpc+deepcopy+storeobject+raftproxy+authenticatedwrapper," +
		"Mgogoproto/gogo.proto=github.com/gogo/protobuf/gogoproto," +
		"Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types," +
		"Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor," +
		"Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types," +
		"Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types," +
		"Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types," +
		"Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types," +
		"Mgithub.com/docker/swarmkit/protobuf/plugin/plugin.proto=github.com/moby/swarmkit/v2/protobuf/plugin"
)

func main() {
	root := flag.String("root", "", "SwarmKit source directory")
	gogoDir := flag.String("gogo-dir", "", "selected Gogo Go module directory")
	raftDir := flag.String("raft-dir", "", "selected Raft Go module directory")
	importPath := flag.String("import-path", "", "Go package import path for generated files")
	descriptorPath := flag.String("descriptor", "", "relative path for the API descriptor set")
	flag.Parse()

	if *root == "" || *gogoDir == "" || *raftDir == "" || *importPath == "" || flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	if err := run(context.Background(), *root, *gogoDir, *raftDir, *importPath, *descriptorPath, flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, root, gogoDir, raftDir, importPath, descriptorPath string, names []string) error {
	set, err := compile(ctx, schemaResolver(root, gogoDir, raftDir), names)
	if err != nil {
		return err
	}

	if descriptorPath != "" {
		snapshot := proto.Clone(set).(*descriptor.FileDescriptorSet)
		for _, file := range snapshot.File {
			file.SourceCodeInfo = nil
		}
		if err := os.WriteFile(filepath.Join(root, descriptorPath), []byte(proto.MarshalTextString(snapshot)), 0644); err != nil {
			return err
		}
	}

	req := &plugin.CodeGeneratorRequest{
		FileToGenerate: names,
		ProtoFile: set.File,
		Parameter: proto.String(generatorOptions + ",import_path=" + importPath),
	}

	files := vanity.FilterFiles(req.GetProtoFile(), vanity.NotGoogleProtobufDescriptorProto)
	for _, opt := range []func(*descriptor.FileDescriptorProto){
		vanity.TurnOffGoGettersAll,
		vanity.TurnOffGoStringerAll,
		vanity.TurnOnMarshalerAll,
		vanity.TurnOnStringerAll,
		vanity.TurnOnUnmarshalerAll,
		vanity.TurnOnSizerAll,
		vanity.TurnOffGoUnrecognizedAll,
		vanity.TurnOffGoUnkeyedAll,
		vanity.TurnOffGoSizecacheAll,
		CustomNameID,
	} {
		vanity.ForEachFile(files, opt)
	}

	resp := command.Generate(req)
	if resp.GetError() != "" {
		return fmt.Errorf("generate Go bindings: %s", resp.GetError())
	}
	for _, output := range resp.File {
		name, ok := strings.CutPrefix(output.GetName(), sourcePrefix)
		if !ok || !filepath.IsLocal(name) || output.InsertionPoint != nil {
			return fmt.Errorf("unexpected generated protobuf path %q", output.GetName())
		}
		if err := os.WriteFile(filepath.Join(root, name), []byte(output.GetContent()), 0644); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}
	return nil
}
