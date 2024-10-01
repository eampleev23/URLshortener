// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.1
// source: proto/grpc.proto

package proto

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
	Shortener_AddShortURL_FullMethodName = "/grpc.Shortener/AddShortURL"
)

// ShortenerClient is the client API for Shortener service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ShortenerClient interface {
	AddShortURL(ctx context.Context, in *AddShortURLRequest, opts ...grpc.CallOption) (*AddShortURLResponse, error)
}

type shortenerClient struct {
	cc grpc.ClientConnInterface
}

func NewShortenerClient(cc grpc.ClientConnInterface) ShortenerClient {
	return &shortenerClient{cc}
}

func (c *shortenerClient) AddShortURL(ctx context.Context, in *AddShortURLRequest, opts ...grpc.CallOption) (*AddShortURLResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AddShortURLResponse)
	err := c.cc.Invoke(ctx, Shortener_AddShortURL_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ShortenerServer is the server API for Shortener service.
// All implementations must embed UnimplementedShortenerServer
// for forward compatibility.
type ShortenerServer interface {
	AddShortURL(context.Context, *AddShortURLRequest) (*AddShortURLResponse, error)
	mustEmbedUnimplementedShortenerServer()
}

// UnimplementedShortenerServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedShortenerServer struct{}

func (UnimplementedShortenerServer) AddShortURL(context.Context, *AddShortURLRequest) (*AddShortURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddShortURL not implemented")
}
func (UnimplementedShortenerServer) mustEmbedUnimplementedShortenerServer() {}
func (UnimplementedShortenerServer) testEmbeddedByValue()                   {}

// UnsafeShortenerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ShortenerServer will
// result in compilation errors.
type UnsafeShortenerServer interface {
	mustEmbedUnimplementedShortenerServer()
}

func RegisterShortenerServer(s grpc.ServiceRegistrar, srv ShortenerServer) {
	// If the following call pancis, it indicates UnimplementedShortenerServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Shortener_ServiceDesc, srv)
}

func _Shortener_AddShortURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddShortURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServer).AddShortURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shortener_AddShortURL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServer).AddShortURL(ctx, req.(*AddShortURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Shortener_ServiceDesc is the grpc.ServiceDesc for Shortener service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Shortener_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpc.Shortener",
	HandlerType: (*ShortenerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddShortURL",
			Handler:    _Shortener_AddShortURL_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/grpc.proto",
}