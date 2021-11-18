// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package npool

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

// CloudHashingApisClient is the client API for CloudHashingApis service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CloudHashingApisClient interface {
	// Method Version
	Version(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*VersionResponse, error)
}

type cloudHashingApisClient struct {
	cc grpc.ClientConnInterface
}

func NewCloudHashingApisClient(cc grpc.ClientConnInterface) CloudHashingApisClient {
	return &cloudHashingApisClient{cc}
}

func (c *cloudHashingApisClient) Version(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*VersionResponse, error) {
	out := new(VersionResponse)
	err := c.cc.Invoke(ctx, "/cloud.hashing.apis.v1.CloudHashingApis/Version", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CloudHashingApisServer is the server API for CloudHashingApis service.
// All implementations must embed UnimplementedCloudHashingApisServer
// for forward compatibility
type CloudHashingApisServer interface {
	// Method Version
	Version(context.Context, *emptypb.Empty) (*VersionResponse, error)
	mustEmbedUnimplementedCloudHashingApisServer()
}

// UnimplementedCloudHashingApisServer must be embedded to have forward compatible implementations.
type UnimplementedCloudHashingApisServer struct {
}

func (UnimplementedCloudHashingApisServer) Version(context.Context, *emptypb.Empty) (*VersionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Version not implemented")
}
func (UnimplementedCloudHashingApisServer) mustEmbedUnimplementedCloudHashingApisServer() {}

// UnsafeCloudHashingApisServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CloudHashingApisServer will
// result in compilation errors.
type UnsafeCloudHashingApisServer interface {
	mustEmbedUnimplementedCloudHashingApisServer()
}

func RegisterCloudHashingApisServer(s grpc.ServiceRegistrar, srv CloudHashingApisServer) {
	s.RegisterService(&CloudHashingApis_ServiceDesc, srv)
}

func _CloudHashingApis_Version_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudHashingApisServer).Version(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cloud.hashing.apis.v1.CloudHashingApis/Version",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudHashingApisServer).Version(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// CloudHashingApis_ServiceDesc is the grpc.ServiceDesc for CloudHashingApis service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CloudHashingApis_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "cloud.hashing.apis.v1.CloudHashingApis",
	HandlerType: (*CloudHashingApisServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Version",
			Handler:    _CloudHashingApis_Version_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "npool/cloud-hashing-apis.proto",
}
