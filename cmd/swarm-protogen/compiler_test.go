package main

import (
	"context"
	"strings"
	"testing"

	"github.com/bufbuild/protocompile"
	"github.com/gogo/protobuf/gogoproto"
)

func TestCompileDescriptorGraph(t *testing.T) {
	const dependency = "gogoproto/gogo.proto"
	const target = "github.com/docker/swarmkit/api/test.proto"
	const descriptorSchema = "google/protobuf/descriptor.proto"
	resolver := &protocompile.SourceResolver{
		Accessor: protocompile.SourceAccessorFromMap(map[string]string{
			descriptorSchema: `syntax = "proto2";
package google.protobuf;
message FieldOptions { extensions 1000 to max; }`,
			dependency: `syntax = "proto2";
package gogoproto;
import "google/protobuf/descriptor.proto";
extend google.protobuf.FieldOptions { optional bool nullable = 65001; }`,
			target: `syntax = "proto3";
package test;
import "gogoproto/gogo.proto";
// Message documentation.
message Example {
  string label = 7 [(gogoproto.nullable) = false];
}`,
		}),
	}
	set, err := compile(context.Background(), resolver, []string{target})
	if err != nil {
		t.Fatal(err)
	}
	if len(set.File) != 3 || set.File[0].GetName() != descriptorSchema ||
		set.File[1].GetName() != dependency || set.File[2].GetName() != target {
		t.Fatalf("incorrect descriptor order: %v", set.File)
	}
	file := set.File[2]
	field := file.GetMessageType()[0].GetField()[0]
	if field.GetNumber() != 7 || gogoproto.IsNullable(field) {
		t.Fatal("field number or custom Gogo option changed during compilation")
	}
	for _, location := range file.GetSourceCodeInfo().GetLocation() {
		if strings.Contains(location.GetLeadingComments(), "Message documentation.") {
			return
		}
	}
	t.Fatal("source documentation was lost")
}
