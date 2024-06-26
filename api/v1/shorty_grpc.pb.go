// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: api/v1/shorty.proto

package api

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	ShortenerService_ShortRequest_FullMethodName      = "/api.v1.ShortenerService/ShortRequest"
	ShortenerService_ShortID_FullMethodName           = "/api.v1.ShortenerService/ShortID"
	ShortenerService_ShortRequestBatch_FullMethodName = "/api.v1.ShortenerService/ShortRequestBatch"
	ShortenerService_GetStats_FullMethodName          = "/api.v1.ShortenerService/GetStats"
)

// ShortenerServiceClient is the client API for ShortenerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ShortenerServiceClient interface {
	// Create a new shortened URL
	ShortRequest(ctx context.Context, in *ShortRequestRequest, opts ...grpc.CallOption) (*ShortRequestResponse, error)
	// Get the real URL for a shortened URL
	ShortID(ctx context.Context, in *ShortIDRequest, opts ...grpc.CallOption) (*ShortIDResponse, error)
	// Creates a batch of URLs and returns their shortened versions
	ShortRequestBatch(ctx context.Context, in *ShortRequestBatchRequest, opts ...grpc.CallOption) (*ShortRequestBatchResponse, error)
	// Get statistics
	GetStats(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetStatsResponse, error)
}

type shortenerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewShortenerServiceClient(cc grpc.ClientConnInterface) ShortenerServiceClient {
	return &shortenerServiceClient{cc}
}

func (c *shortenerServiceClient) ShortRequest(ctx context.Context, in *ShortRequestRequest, opts ...grpc.CallOption) (*ShortRequestResponse, error) {
	out := new(ShortRequestResponse)
	err := c.cc.Invoke(ctx, ShortenerService_ShortRequest_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) ShortID(ctx context.Context, in *ShortIDRequest, opts ...grpc.CallOption) (*ShortIDResponse, error) {
	out := new(ShortIDResponse)
	err := c.cc.Invoke(ctx, ShortenerService_ShortID_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) ShortRequestBatch(ctx context.Context, in *ShortRequestBatchRequest, opts ...grpc.CallOption) (*ShortRequestBatchResponse, error) {
	out := new(ShortRequestBatchResponse)
	err := c.cc.Invoke(ctx, ShortenerService_ShortRequestBatch_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) GetStats(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetStatsResponse, error) {
	out := new(GetStatsResponse)
	err := c.cc.Invoke(ctx, ShortenerService_GetStats_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ShortenerServiceServer is the server API for ShortenerService service.
// All implementations must embed UnimplementedShortenerServiceServer
// for forward compatibility
type ShortenerServiceServer interface {
	// Create a new shortened URL
	ShortRequest(context.Context, *ShortRequestRequest) (*ShortRequestResponse, error)
	// Get the real URL for a shortened URL
	ShortID(context.Context, *ShortIDRequest) (*ShortIDResponse, error)
	// Creates a batch of URLs and returns their shortened versions
	ShortRequestBatch(context.Context, *ShortRequestBatchRequest) (*ShortRequestBatchResponse, error)
	// Get statistics
	GetStats(context.Context, *emptypb.Empty) (*GetStatsResponse, error)
	mustEmbedUnimplementedShortenerServiceServer()
}

// UnimplementedShortenerServiceServer must be embedded to have forward compatible implementations.
type UnimplementedShortenerServiceServer struct {
}

func (UnimplementedShortenerServiceServer) ShortRequest(context.Context, *ShortRequestRequest) (*ShortRequestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ShortRequest not implemented")
}
func (UnimplementedShortenerServiceServer) ShortID(context.Context, *ShortIDRequest) (*ShortIDResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ShortID not implemented")
}
func (UnimplementedShortenerServiceServer) ShortRequestBatch(context.Context, *ShortRequestBatchRequest) (*ShortRequestBatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ShortRequestBatch not implemented")
}
func (UnimplementedShortenerServiceServer) GetStats(context.Context, *emptypb.Empty) (*GetStatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStats not implemented")
}
func (UnimplementedShortenerServiceServer) mustEmbedUnimplementedShortenerServiceServer() {}

// UnsafeShortenerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ShortenerServiceServer will
// result in compilation errors.
type UnsafeShortenerServiceServer interface {
	mustEmbedUnimplementedShortenerServiceServer()
}

func RegisterShortenerServiceServer(s grpc.ServiceRegistrar, srv ShortenerServiceServer) {
	s.RegisterService(&ShortenerService_ServiceDesc, srv)
}

func _ShortenerService_ShortRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ShortRequestRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).ShortRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_ShortRequest_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).ShortRequest(ctx, req.(*ShortRequestRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortenerService_ShortID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ShortIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).ShortID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_ShortID_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).ShortID(ctx, req.(*ShortIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortenerService_ShortRequestBatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ShortRequestBatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).ShortRequestBatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_ShortRequestBatch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).ShortRequestBatch(ctx, req.(*ShortRequestBatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortenerService_GetStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).GetStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_GetStats_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).GetStats(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// ShortenerService_ServiceDesc is the grpc.ServiceDesc for ShortenerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ShortenerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.v1.ShortenerService",
	HandlerType: (*ShortenerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ShortRequest",
			Handler:    _ShortenerService_ShortRequest_Handler,
		},
		{
			MethodName: "ShortID",
			Handler:    _ShortenerService_ShortID_Handler,
		},
		{
			MethodName: "ShortRequestBatch",
			Handler:    _ShortenerService_ShortRequestBatch_Handler,
		},
		{
			MethodName: "GetStats",
			Handler:    _ShortenerService_GetStats_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/v1/shorty.proto",
}
