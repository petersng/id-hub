// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: pkg/api/idhub.proto

package api

import (
	fmt "fmt"
	math "math"
	proto "github.com/golang/protobuf/proto"
	context "context"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type IdHubGQLServer struct{ Service IdHubServer }

func (s *IdHubGQLServer) IDHubGetDID(ctx context.Context, in *Did) (*DidDocument, error) {
	return s.Service.GetDID(ctx, in)
}