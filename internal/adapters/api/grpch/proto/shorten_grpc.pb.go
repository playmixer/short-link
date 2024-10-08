// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.0--rc2
// source: shorten.proto

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
	Shorten_Login_FullMethodName          = "/grpch.proto.Shorten/Login"
	Shorten_NewShort_FullMethodName       = "/grpch.proto.Shorten/NewShort"
	Shorten_NewShorts_FullMethodName      = "/grpch.proto.Shorten/NewShorts"
	Shorten_GetURLByShort_FullMethodName  = "/grpch.proto.Shorten/GetURLByShort"
	Shorten_GetUserURLs_FullMethodName    = "/grpch.proto.Shorten/GetUserURLs"
	Shorten_DeleteUserURLs_FullMethodName = "/grpch.proto.Shorten/DeleteUserURLs"
	Shorten_GetStatus_FullMethodName      = "/grpch.proto.Shorten/GetStatus"
)

// ShortenClient is the client API for Shorten service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ShortenClient interface {
	Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error)
	NewShort(ctx context.Context, in *NewShortRequest, opts ...grpc.CallOption) (*NewShortResponse, error)
	NewShorts(ctx context.Context, in *NewShortsRequest, opts ...grpc.CallOption) (*NewShortsResponse, error)
	GetURLByShort(ctx context.Context, in *GetUrlByShortRequest, opts ...grpc.CallOption) (*GetURLByShortResponse, error)
	GetUserURLs(ctx context.Context, in *GetUserURLsRequest, opts ...grpc.CallOption) (*GetUserURLsResponse, error)
	DeleteUserURLs(ctx context.Context, in *DeleteUserURLsRequest, opts ...grpc.CallOption) (*DeleteUserURLsRespons, error)
	GetStatus(ctx context.Context, in *GetStatusRequest, opts ...grpc.CallOption) (*GetStatusResponse, error)
}

type shortenClient struct {
	cc grpc.ClientConnInterface
}

func NewShortenClient(cc grpc.ClientConnInterface) ShortenClient {
	return &shortenClient{cc}
}

func (c *shortenClient) Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LoginResponse)
	err := c.cc.Invoke(ctx, Shorten_Login_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenClient) NewShort(ctx context.Context, in *NewShortRequest, opts ...grpc.CallOption) (*NewShortResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(NewShortResponse)
	err := c.cc.Invoke(ctx, Shorten_NewShort_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenClient) NewShorts(ctx context.Context, in *NewShortsRequest, opts ...grpc.CallOption) (*NewShortsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(NewShortsResponse)
	err := c.cc.Invoke(ctx, Shorten_NewShorts_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenClient) GetURLByShort(ctx context.Context, in *GetUrlByShortRequest, opts ...grpc.CallOption) (*GetURLByShortResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetURLByShortResponse)
	err := c.cc.Invoke(ctx, Shorten_GetURLByShort_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenClient) GetUserURLs(ctx context.Context, in *GetUserURLsRequest, opts ...grpc.CallOption) (*GetUserURLsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetUserURLsResponse)
	err := c.cc.Invoke(ctx, Shorten_GetUserURLs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenClient) DeleteUserURLs(ctx context.Context, in *DeleteUserURLsRequest, opts ...grpc.CallOption) (*DeleteUserURLsRespons, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeleteUserURLsRespons)
	err := c.cc.Invoke(ctx, Shorten_DeleteUserURLs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenClient) GetStatus(ctx context.Context, in *GetStatusRequest, opts ...grpc.CallOption) (*GetStatusResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetStatusResponse)
	err := c.cc.Invoke(ctx, Shorten_GetStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ShortenServer is the server API for Shorten service.
// All implementations must embed UnimplementedShortenServer
// for forward compatibility.
type ShortenServer interface {
	Login(context.Context, *LoginRequest) (*LoginResponse, error)
	NewShort(context.Context, *NewShortRequest) (*NewShortResponse, error)
	NewShorts(context.Context, *NewShortsRequest) (*NewShortsResponse, error)
	GetURLByShort(context.Context, *GetUrlByShortRequest) (*GetURLByShortResponse, error)
	GetUserURLs(context.Context, *GetUserURLsRequest) (*GetUserURLsResponse, error)
	DeleteUserURLs(context.Context, *DeleteUserURLsRequest) (*DeleteUserURLsRespons, error)
	GetStatus(context.Context, *GetStatusRequest) (*GetStatusResponse, error)
	mustEmbedUnimplementedShortenServer()
}

// UnimplementedShortenServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedShortenServer struct{}

func (UnimplementedShortenServer) Login(context.Context, *LoginRequest) (*LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}
func (UnimplementedShortenServer) NewShort(context.Context, *NewShortRequest) (*NewShortResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewShort not implemented")
}
func (UnimplementedShortenServer) NewShorts(context.Context, *NewShortsRequest) (*NewShortsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewShorts not implemented")
}
func (UnimplementedShortenServer) GetURLByShort(context.Context, *GetUrlByShortRequest) (*GetURLByShortResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetURLByShort not implemented")
}
func (UnimplementedShortenServer) GetUserURLs(context.Context, *GetUserURLsRequest) (*GetUserURLsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserURLs not implemented")
}
func (UnimplementedShortenServer) DeleteUserURLs(context.Context, *DeleteUserURLsRequest) (*DeleteUserURLsRespons, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUserURLs not implemented")
}
func (UnimplementedShortenServer) GetStatus(context.Context, *GetStatusRequest) (*GetStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStatus not implemented")
}
func (UnimplementedShortenServer) mustEmbedUnimplementedShortenServer() {}
func (UnimplementedShortenServer) testEmbeddedByValue()                 {}

// UnsafeShortenServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ShortenServer will
// result in compilation errors.
type UnsafeShortenServer interface {
	mustEmbedUnimplementedShortenServer()
}

func RegisterShortenServer(s grpc.ServiceRegistrar, srv ShortenServer) {
	// If the following call pancis, it indicates UnimplementedShortenServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Shorten_ServiceDesc, srv)
}

func _Shorten_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shorten_Login_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenServer).Login(ctx, req.(*LoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Shorten_NewShort_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NewShortRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenServer).NewShort(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shorten_NewShort_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenServer).NewShort(ctx, req.(*NewShortRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Shorten_NewShorts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NewShortsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenServer).NewShorts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shorten_NewShorts_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenServer).NewShorts(ctx, req.(*NewShortsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Shorten_GetURLByShort_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUrlByShortRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenServer).GetURLByShort(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shorten_GetURLByShort_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenServer).GetURLByShort(ctx, req.(*GetUrlByShortRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Shorten_GetUserURLs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserURLsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenServer).GetUserURLs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shorten_GetUserURLs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenServer).GetUserURLs(ctx, req.(*GetUserURLsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Shorten_DeleteUserURLs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteUserURLsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenServer).DeleteUserURLs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shorten_DeleteUserURLs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenServer).DeleteUserURLs(ctx, req.(*DeleteUserURLsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Shorten_GetStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenServer).GetStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shorten_GetStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenServer).GetStatus(ctx, req.(*GetStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Shorten_ServiceDesc is the grpc.ServiceDesc for Shorten service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Shorten_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpch.proto.Shorten",
	HandlerType: (*ShortenServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Login",
			Handler:    _Shorten_Login_Handler,
		},
		{
			MethodName: "NewShort",
			Handler:    _Shorten_NewShort_Handler,
		},
		{
			MethodName: "NewShorts",
			Handler:    _Shorten_NewShorts_Handler,
		},
		{
			MethodName: "GetURLByShort",
			Handler:    _Shorten_GetURLByShort_Handler,
		},
		{
			MethodName: "GetUserURLs",
			Handler:    _Shorten_GetUserURLs_Handler,
		},
		{
			MethodName: "DeleteUserURLs",
			Handler:    _Shorten_DeleteUserURLs_Handler,
		},
		{
			MethodName: "GetStatus",
			Handler:    _Shorten_GetStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "shorten.proto",
}
