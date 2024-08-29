// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.12.4
// source: internal/grpc/proto/shortener.proto

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
	Shortener_EncodeURL_FullMethodName      = "/shortener.Shortener/EncodeURL"
	Shortener_DecodeURL_FullMethodName      = "/shortener.Shortener/DecodeURL"
	Shortener_Batch_FullMethodName          = "/shortener.Shortener/Batch"
	Shortener_UserURLs_FullMethodName       = "/shortener.Shortener/UserURLs"
	Shortener_DeleteUserURLs_FullMethodName = "/shortener.Shortener/DeleteUserURLs"
	Shortener_Stats_FullMethodName          = "/shortener.Shortener/Stats"
	Shortener_Ping_FullMethodName           = "/shortener.Shortener/Ping"
)

// ShortenerClient is the client API for Shortener service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ShortenerClient interface {
	EncodeURL(ctx context.Context, in *EncodeURLRequest, opts ...grpc.CallOption) (*EncodeURLResponse, error)
	DecodeURL(ctx context.Context, in *DecodeURLRequest, opts ...grpc.CallOption) (*DecodeURLResponse, error)
	Batch(ctx context.Context, in *BatchRequest, opts ...grpc.CallOption) (*BatchResponse, error)
	UserURLs(ctx context.Context, in *UserURLsRequest, opts ...grpc.CallOption) (*UserURLsResponse, error)
	DeleteUserURLs(ctx context.Context, in *DelUserRequest, opts ...grpc.CallOption) (*DeleteUserURLsResponse, error)
	Stats(ctx context.Context, in *StatsRequest, opts ...grpc.CallOption) (*StatsResponse, error)
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
}

type shortenerClient struct {
	cc grpc.ClientConnInterface
}

func NewShortenerClient(cc grpc.ClientConnInterface) ShortenerClient {
	return &shortenerClient{cc}
}

func (c *shortenerClient) EncodeURL(ctx context.Context, in *EncodeURLRequest, opts ...grpc.CallOption) (*EncodeURLResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EncodeURLResponse)
	err := c.cc.Invoke(ctx, Shortener_EncodeURL_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerClient) DecodeURL(ctx context.Context, in *DecodeURLRequest, opts ...grpc.CallOption) (*DecodeURLResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DecodeURLResponse)
	err := c.cc.Invoke(ctx, Shortener_DecodeURL_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerClient) Batch(ctx context.Context, in *BatchRequest, opts ...grpc.CallOption) (*BatchResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BatchResponse)
	err := c.cc.Invoke(ctx, Shortener_Batch_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerClient) UserURLs(ctx context.Context, in *UserURLsRequest, opts ...grpc.CallOption) (*UserURLsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UserURLsResponse)
	err := c.cc.Invoke(ctx, Shortener_UserURLs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerClient) DeleteUserURLs(ctx context.Context, in *DelUserRequest, opts ...grpc.CallOption) (*DeleteUserURLsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeleteUserURLsResponse)
	err := c.cc.Invoke(ctx, Shortener_DeleteUserURLs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerClient) Stats(ctx context.Context, in *StatsRequest, opts ...grpc.CallOption) (*StatsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(StatsResponse)
	err := c.cc.Invoke(ctx, Shortener_Stats_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, Shortener_Ping_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ShortenerServer is the server API for Shortener service.
// All implementations must embed UnimplementedShortenerServer
// for forward compatibility.
type ShortenerServer interface {
	EncodeURL(context.Context, *EncodeURLRequest) (*EncodeURLResponse, error)
	DecodeURL(context.Context, *DecodeURLRequest) (*DecodeURLResponse, error)
	Batch(context.Context, *BatchRequest) (*BatchResponse, error)
	UserURLs(context.Context, *UserURLsRequest) (*UserURLsResponse, error)
	DeleteUserURLs(context.Context, *DelUserRequest) (*DeleteUserURLsResponse, error)
	Stats(context.Context, *StatsRequest) (*StatsResponse, error)
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	mustEmbedUnimplementedShortenerServer()
}

// UnimplementedShortenerServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedShortenerServer struct{}

func (UnimplementedShortenerServer) EncodeURL(context.Context, *EncodeURLRequest) (*EncodeURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EncodeURL not implemented")
}
func (UnimplementedShortenerServer) DecodeURL(context.Context, *DecodeURLRequest) (*DecodeURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DecodeURL not implemented")
}
func (UnimplementedShortenerServer) Batch(context.Context, *BatchRequest) (*BatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Batch not implemented")
}
func (UnimplementedShortenerServer) UserURLs(context.Context, *UserURLsRequest) (*UserURLsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UserURLs not implemented")
}
func (UnimplementedShortenerServer) DeleteUserURLs(context.Context, *DelUserRequest) (*DeleteUserURLsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUserURLs not implemented")
}
func (UnimplementedShortenerServer) Stats(context.Context, *StatsRequest) (*StatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stats not implemented")
}
func (UnimplementedShortenerServer) Ping(context.Context, *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
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

func _Shortener_EncodeURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EncodeURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServer).EncodeURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shortener_EncodeURL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServer).EncodeURL(ctx, req.(*EncodeURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Shortener_DecodeURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DecodeURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServer).DecodeURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shortener_DecodeURL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServer).DecodeURL(ctx, req.(*DecodeURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Shortener_Batch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServer).Batch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shortener_Batch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServer).Batch(ctx, req.(*BatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Shortener_UserURLs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserURLsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServer).UserURLs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shortener_UserURLs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServer).UserURLs(ctx, req.(*UserURLsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Shortener_DeleteUserURLs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DelUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServer).DeleteUserURLs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shortener_DeleteUserURLs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServer).DeleteUserURLs(ctx, req.(*DelUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Shortener_Stats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServer).Stats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shortener_Stats_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServer).Stats(ctx, req.(*StatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Shortener_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shortener_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Shortener_ServiceDesc is the grpc.ServiceDesc for Shortener service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Shortener_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "shortener.Shortener",
	HandlerType: (*ShortenerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "EncodeURL",
			Handler:    _Shortener_EncodeURL_Handler,
		},
		{
			MethodName: "DecodeURL",
			Handler:    _Shortener_DecodeURL_Handler,
		},
		{
			MethodName: "Batch",
			Handler:    _Shortener_Batch_Handler,
		},
		{
			MethodName: "UserURLs",
			Handler:    _Shortener_UserURLs_Handler,
		},
		{
			MethodName: "DeleteUserURLs",
			Handler:    _Shortener_DeleteUserURLs_Handler,
		},
		{
			MethodName: "Stats",
			Handler:    _Shortener_Stats_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _Shortener_Ping_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "internal/grpc/proto/shortener.proto",
}
