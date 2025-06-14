// proto/mcp.proto

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: mcp.proto

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	MeshContext_GetMeshGraph_FullMethodName             = "/mcp.v1.MeshContext/GetMeshGraph"
	MeshContext_ApplyAuthorizationPolicy_FullMethodName = "/mcp.v1.MeshContext/ApplyAuthorizationPolicy"
)

// MeshContextClient is the client API for MeshContext service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// Placeholder service definition
type MeshContextClient interface {
	GetMeshGraph(ctx context.Context, in *GetMeshGraphRequest, opts ...grpc.CallOption) (*GetMeshGraphResponse, error)
	ApplyAuthorizationPolicy(ctx context.Context, in *ApplyAuthorizationPolicyRequest, opts ...grpc.CallOption) (*ApplyAuthorizationPolicyResponse, error)
}

type meshContextClient struct {
	cc grpc.ClientConnInterface
}

func NewMeshContextClient(cc grpc.ClientConnInterface) MeshContextClient {
	return &meshContextClient{cc}
}

func (c *meshContextClient) GetMeshGraph(ctx context.Context, in *GetMeshGraphRequest, opts ...grpc.CallOption) (*GetMeshGraphResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetMeshGraphResponse)
	err := c.cc.Invoke(ctx, MeshContext_GetMeshGraph_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *meshContextClient) ApplyAuthorizationPolicy(ctx context.Context, in *ApplyAuthorizationPolicyRequest, opts ...grpc.CallOption) (*ApplyAuthorizationPolicyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ApplyAuthorizationPolicyResponse)
	err := c.cc.Invoke(ctx, MeshContext_ApplyAuthorizationPolicy_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MeshContextServer is the server API for MeshContext service.
// All implementations must embed UnimplementedMeshContextServer
// for forward compatibility.
//
// Placeholder service definition
type MeshContextServer interface {
	GetMeshGraph(context.Context, *GetMeshGraphRequest) (*GetMeshGraphResponse, error)
	ApplyAuthorizationPolicy(context.Context, *ApplyAuthorizationPolicyRequest) (*ApplyAuthorizationPolicyResponse, error)
	mustEmbedUnimplementedMeshContextServer()
}

// UnimplementedMeshContextServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedMeshContextServer struct{}

func (UnimplementedMeshContextServer) GetMeshGraph(context.Context, *GetMeshGraphRequest) (*GetMeshGraphResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMeshGraph not implemented")
}
func (UnimplementedMeshContextServer) ApplyAuthorizationPolicy(context.Context, *ApplyAuthorizationPolicyRequest) (*ApplyAuthorizationPolicyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApplyAuthorizationPolicy not implemented")
}
func (UnimplementedMeshContextServer) mustEmbedUnimplementedMeshContextServer() {}
func (UnimplementedMeshContextServer) testEmbeddedByValue()                     {}

// UnsafeMeshContextServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MeshContextServer will
// result in compilation errors.
type UnsafeMeshContextServer interface {
	mustEmbedUnimplementedMeshContextServer()
}

func RegisterMeshContextServer(s grpc.ServiceRegistrar, srv MeshContextServer) {
	// If the following call pancis, it indicates UnimplementedMeshContextServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&MeshContext_ServiceDesc, srv)
}

func _MeshContext_GetMeshGraph_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMeshGraphRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MeshContextServer).GetMeshGraph(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MeshContext_GetMeshGraph_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MeshContextServer).GetMeshGraph(ctx, req.(*GetMeshGraphRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MeshContext_ApplyAuthorizationPolicy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ApplyAuthorizationPolicyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MeshContextServer).ApplyAuthorizationPolicy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MeshContext_ApplyAuthorizationPolicy_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MeshContextServer).ApplyAuthorizationPolicy(ctx, req.(*ApplyAuthorizationPolicyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MeshContext_ServiceDesc is the grpc.ServiceDesc for MeshContext service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MeshContext_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mcp.v1.MeshContext",
	HandlerType: (*MeshContextServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetMeshGraph",
			Handler:    _MeshContext_GetMeshGraph_Handler,
		},
		{
			MethodName: "ApplyAuthorizationPolicy",
			Handler:    _MeshContext_ApplyAuthorizationPolicy_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "mcp.proto",
}
