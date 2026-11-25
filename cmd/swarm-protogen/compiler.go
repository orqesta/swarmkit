package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	legacyproto "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"google.golang.org/protobuf/proto"
)

// schemaResolver assigns each supported import namespace to exactly one source
// directory. Standard schemas come from the same Gogo release as the generator.
func schemaResolver(root, gogoDir, raftDir string) protocompile.Resolver {
	return &protocompile.SourceResolver{
		Accessor: func(name string) (io.ReadCloser, error) {
			if !fs.ValidPath(name) {
				return nil, fmt.Errorf("invalid schema path %q", name)
			}
			switch {
			case strings.HasPrefix(name, sourcePrefix):
				return os.Open(filepath.Join(root, strings.TrimPrefix(name, sourcePrefix)))
			case strings.HasPrefix(name, "github.com/gogo/protobuf/"):
				return os.Open(filepath.Join(gogoDir, strings.TrimPrefix(name, "github.com/gogo/protobuf/")))
			case strings.HasPrefix(name, "gogoproto/"):
				return os.Open(filepath.Join(gogoDir, name))
			case strings.HasPrefix(name, "google/protobuf/"):
				return os.Open(filepath.Join(gogoDir, "protobuf", name))
			case strings.HasPrefix(name, "go.etcd.io/raft/v3/"):
				return os.Open(filepath.Join(raftDir, strings.TrimPrefix(name, "go.etcd.io/raft/v3/")))
			default:
				return nil, fmt.Errorf("unmapped schema %q: %w", name, fs.ErrNotExist)
			}
		},
	}
}

// compile returns the descriptor graph in dependency-first order, as required
// by the Gogo generator. The wire encoding transfers custom extension options
// and source locations between the compiler and generator descriptor types.
func compile(ctx context.Context, resolver protocompile.Resolver, names []string) (*descriptor.FileDescriptorSet, error) {
	compiler := protocompile.Compiler{
		Resolver: resolver,
		SourceInfoMode: protocompile.SourceInfoStandard,
	}
	files, err := compiler.Compile(ctx, names...)
	if err != nil {
		return nil, fmt.Errorf("compile protobuf schemas: %w", err)
	}

	set := &descriptor.FileDescriptorSet{}
	seen := make(map[string]bool)
	var visit func(linker.File) error
	visit = func(file linker.File) error {
		if file == nil {
			return fmt.Errorf("missing compiled dependency")
		}
		if seen[file.Path()] {
			return nil
		}
		result, ok := file.(linker.Result)
		if !ok {
			return fmt.Errorf("schema %q was not compiled from source", file.Path())
		}
		fd := result.FileDescriptorProto()
		for _, name := range fd.GetDependency() {
			if err := visit(file.FindImportByPath(name)); err != nil {
				return fmt.Errorf("resolve %s: %w", name, err)
			}
		}
		data, err := (proto.MarshalOptions{Deterministic: true}).Marshal(fd)
		if err != nil {
			return fmt.Errorf("encode descriptor %s: %w", file.Path(), err)
		}
		converted := &descriptor.FileDescriptorProto{}
		if err := legacyproto.Unmarshal(data, converted); err != nil {
			return fmt.Errorf("decode Gogo descriptor %s: %w", file.Path(), err)
		}
		set.File = append(set.File, converted)
		seen[file.Path()] = true
		return nil
	}
	for _, file := range files {
		if err := visit(file); err != nil {
			return nil, err
		}
	}
	return set, nil
}
