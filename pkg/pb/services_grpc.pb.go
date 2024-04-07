// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.12
// source: services.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	ArtGeneratorService_LoginUser_FullMethodName      = "/pb.ArtGeneratorService/LoginUser"
	ArtGeneratorService_LogoutUser_FullMethodName     = "/pb.ArtGeneratorService/LogoutUser"
	ArtGeneratorService_RefreshToken_FullMethodName   = "/pb.ArtGeneratorService/RefreshToken"
	ArtGeneratorService_CreateUser_FullMethodName     = "/pb.ArtGeneratorService/CreateUser"
	ArtGeneratorService_UpdateUser_FullMethodName     = "/pb.ArtGeneratorService/UpdateUser"
	ArtGeneratorService_GetUser_FullMethodName        = "/pb.ArtGeneratorService/GetUser"
	ArtGeneratorService_ResetPassword_FullMethodName  = "/pb.ArtGeneratorService/ResetPassword"
	ArtGeneratorService_ChangePassword_FullMethodName = "/pb.ArtGeneratorService/ChangePassword"
	ArtGeneratorService_CreateArt_FullMethodName      = "/pb.ArtGeneratorService/CreateArt"
	ArtGeneratorService_UpdateArt_FullMethodName      = "/pb.ArtGeneratorService/UpdateArt"
	ArtGeneratorService_GetArt_FullMethodName         = "/pb.ArtGeneratorService/GetArt"
	ArtGeneratorService_ListArts_FullMethodName       = "/pb.ArtGeneratorService/ListArts"
	ArtGeneratorService_DeleteArt_FullMethodName      = "/pb.ArtGeneratorService/DeleteArt"
)

// ArtGeneratorServiceClient is the client API for ArtGeneratorService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ArtGeneratorServiceClient interface {
	LoginUser(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error)
	LogoutUser(ctx context.Context, in *LogoutRequest, opts ...grpc.CallOption) (*LogoutResponse, error)
	RefreshToken(ctx context.Context, in *RefreshTokenRequest, opts ...grpc.CallOption) (*RefreshTokenResponse, error)
	CreateUser(ctx context.Context, in *CreateUserRequest, opts ...grpc.CallOption) (*CreateUserResponse, error)
	UpdateUser(ctx context.Context, in *UpdateUserRequest, opts ...grpc.CallOption) (*UpdateUserResponse, error)
	GetUser(ctx context.Context, in *GetUserRequest, opts ...grpc.CallOption) (*GetUserResponse, error)
	ResetPassword(ctx context.Context, in *ResetPasswordRequest, opts ...grpc.CallOption) (*ResetPasswordResponse, error)
	ChangePassword(ctx context.Context, in *ChangePasswordRequest, opts ...grpc.CallOption) (*ChangePasswordResponse, error)
	CreateArt(ctx context.Context, in *CreateArtRequest, opts ...grpc.CallOption) (*CreateArtResponse, error)
	UpdateArt(ctx context.Context, in *UpdateArtRequest, opts ...grpc.CallOption) (*UpdateArtResponse, error)
	GetArt(ctx context.Context, in *GetArtRequest, opts ...grpc.CallOption) (*GetArtResponse, error)
	ListArts(ctx context.Context, in *ListArtRequest, opts ...grpc.CallOption) (*ListArtResponse, error)
	DeleteArt(ctx context.Context, in *DeleteArtRequest, opts ...grpc.CallOption) (*DeleteArtResponse, error)
}

type artGeneratorServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewArtGeneratorServiceClient(cc grpc.ClientConnInterface) ArtGeneratorServiceClient {
	return &artGeneratorServiceClient{cc}
}

func (c *artGeneratorServiceClient) LoginUser(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error) {
	out := new(LoginResponse)
	err := c.cc.Invoke(ctx, ArtGeneratorService_LoginUser_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *artGeneratorServiceClient) LogoutUser(ctx context.Context, in *LogoutRequest, opts ...grpc.CallOption) (*LogoutResponse, error) {
	out := new(LogoutResponse)
	err := c.cc.Invoke(ctx, ArtGeneratorService_LogoutUser_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *artGeneratorServiceClient) RefreshToken(ctx context.Context, in *RefreshTokenRequest, opts ...grpc.CallOption) (*RefreshTokenResponse, error) {
	out := new(RefreshTokenResponse)
	err := c.cc.Invoke(ctx, ArtGeneratorService_RefreshToken_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *artGeneratorServiceClient) CreateUser(ctx context.Context, in *CreateUserRequest, opts ...grpc.CallOption) (*CreateUserResponse, error) {
	out := new(CreateUserResponse)
	err := c.cc.Invoke(ctx, ArtGeneratorService_CreateUser_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *artGeneratorServiceClient) UpdateUser(ctx context.Context, in *UpdateUserRequest, opts ...grpc.CallOption) (*UpdateUserResponse, error) {
	out := new(UpdateUserResponse)
	err := c.cc.Invoke(ctx, ArtGeneratorService_UpdateUser_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *artGeneratorServiceClient) GetUser(ctx context.Context, in *GetUserRequest, opts ...grpc.CallOption) (*GetUserResponse, error) {
	out := new(GetUserResponse)
	err := c.cc.Invoke(ctx, ArtGeneratorService_GetUser_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *artGeneratorServiceClient) ResetPassword(ctx context.Context, in *ResetPasswordRequest, opts ...grpc.CallOption) (*ResetPasswordResponse, error) {
	out := new(ResetPasswordResponse)
	err := c.cc.Invoke(ctx, ArtGeneratorService_ResetPassword_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *artGeneratorServiceClient) ChangePassword(ctx context.Context, in *ChangePasswordRequest, opts ...grpc.CallOption) (*ChangePasswordResponse, error) {
	out := new(ChangePasswordResponse)
	err := c.cc.Invoke(ctx, ArtGeneratorService_ChangePassword_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *artGeneratorServiceClient) CreateArt(ctx context.Context, in *CreateArtRequest, opts ...grpc.CallOption) (*CreateArtResponse, error) {
	out := new(CreateArtResponse)
	err := c.cc.Invoke(ctx, ArtGeneratorService_CreateArt_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *artGeneratorServiceClient) UpdateArt(ctx context.Context, in *UpdateArtRequest, opts ...grpc.CallOption) (*UpdateArtResponse, error) {
	out := new(UpdateArtResponse)
	err := c.cc.Invoke(ctx, ArtGeneratorService_UpdateArt_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *artGeneratorServiceClient) GetArt(ctx context.Context, in *GetArtRequest, opts ...grpc.CallOption) (*GetArtResponse, error) {
	out := new(GetArtResponse)
	err := c.cc.Invoke(ctx, ArtGeneratorService_GetArt_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *artGeneratorServiceClient) ListArts(ctx context.Context, in *ListArtRequest, opts ...grpc.CallOption) (*ListArtResponse, error) {
	out := new(ListArtResponse)
	err := c.cc.Invoke(ctx, ArtGeneratorService_ListArts_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *artGeneratorServiceClient) DeleteArt(ctx context.Context, in *DeleteArtRequest, opts ...grpc.CallOption) (*DeleteArtResponse, error) {
	out := new(DeleteArtResponse)
	err := c.cc.Invoke(ctx, ArtGeneratorService_DeleteArt_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ArtGeneratorServiceServer is the server API for ArtGeneratorService service.
// All implementations must embed UnimplementedArtGeneratorServiceServer
// for forward compatibility
type ArtGeneratorServiceServer interface {
	LoginUser(context.Context, *LoginRequest) (*LoginResponse, error)
	LogoutUser(context.Context, *LogoutRequest) (*LogoutResponse, error)
	RefreshToken(context.Context, *RefreshTokenRequest) (*RefreshTokenResponse, error)
	CreateUser(context.Context, *CreateUserRequest) (*CreateUserResponse, error)
	UpdateUser(context.Context, *UpdateUserRequest) (*UpdateUserResponse, error)
	GetUser(context.Context, *GetUserRequest) (*GetUserResponse, error)
	ResetPassword(context.Context, *ResetPasswordRequest) (*ResetPasswordResponse, error)
	ChangePassword(context.Context, *ChangePasswordRequest) (*ChangePasswordResponse, error)
	CreateArt(context.Context, *CreateArtRequest) (*CreateArtResponse, error)
	UpdateArt(context.Context, *UpdateArtRequest) (*UpdateArtResponse, error)
	GetArt(context.Context, *GetArtRequest) (*GetArtResponse, error)
	ListArts(context.Context, *ListArtRequest) (*ListArtResponse, error)
	DeleteArt(context.Context, *DeleteArtRequest) (*DeleteArtResponse, error)
	mustEmbedUnimplementedArtGeneratorServiceServer()
}

// UnimplementedArtGeneratorServiceServer must be embedded to have forward compatible implementations.
type UnimplementedArtGeneratorServiceServer struct {
}

func (UnimplementedArtGeneratorServiceServer) LoginUser(context.Context, *LoginRequest) (*LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LoginUser not implemented")
}
func (UnimplementedArtGeneratorServiceServer) LogoutUser(context.Context, *LogoutRequest) (*LogoutResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LogoutUser not implemented")
}
func (UnimplementedArtGeneratorServiceServer) RefreshToken(context.Context, *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RefreshToken not implemented")
}
func (UnimplementedArtGeneratorServiceServer) CreateUser(context.Context, *CreateUserRequest) (*CreateUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateUser not implemented")
}
func (UnimplementedArtGeneratorServiceServer) UpdateUser(context.Context, *UpdateUserRequest) (*UpdateUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUser not implemented")
}
func (UnimplementedArtGeneratorServiceServer) GetUser(context.Context, *GetUserRequest) (*GetUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUser not implemented")
}
func (UnimplementedArtGeneratorServiceServer) ResetPassword(context.Context, *ResetPasswordRequest) (*ResetPasswordResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResetPassword not implemented")
}
func (UnimplementedArtGeneratorServiceServer) ChangePassword(context.Context, *ChangePasswordRequest) (*ChangePasswordResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangePassword not implemented")
}
func (UnimplementedArtGeneratorServiceServer) CreateArt(context.Context, *CreateArtRequest) (*CreateArtResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateArt not implemented")
}
func (UnimplementedArtGeneratorServiceServer) UpdateArt(context.Context, *UpdateArtRequest) (*UpdateArtResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateArt not implemented")
}
func (UnimplementedArtGeneratorServiceServer) GetArt(context.Context, *GetArtRequest) (*GetArtResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetArt not implemented")
}
func (UnimplementedArtGeneratorServiceServer) ListArts(context.Context, *ListArtRequest) (*ListArtResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListArts not implemented")
}
func (UnimplementedArtGeneratorServiceServer) DeleteArt(context.Context, *DeleteArtRequest) (*DeleteArtResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteArt not implemented")
}
func (UnimplementedArtGeneratorServiceServer) mustEmbedUnimplementedArtGeneratorServiceServer() {}

// UnsafeArtGeneratorServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ArtGeneratorServiceServer will
// result in compilation errors.
type UnsafeArtGeneratorServiceServer interface {
	mustEmbedUnimplementedArtGeneratorServiceServer()
}

func RegisterArtGeneratorServiceServer(s grpc.ServiceRegistrar, srv ArtGeneratorServiceServer) {
	s.RegisterService(&ArtGeneratorService_ServiceDesc, srv)
}

func _ArtGeneratorService_LoginUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArtGeneratorServiceServer).LoginUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ArtGeneratorService_LoginUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArtGeneratorServiceServer).LoginUser(ctx, req.(*LoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArtGeneratorService_LogoutUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogoutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArtGeneratorServiceServer).LogoutUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ArtGeneratorService_LogoutUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArtGeneratorServiceServer).LogoutUser(ctx, req.(*LogoutRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArtGeneratorService_RefreshToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RefreshTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArtGeneratorServiceServer).RefreshToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ArtGeneratorService_RefreshToken_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArtGeneratorServiceServer).RefreshToken(ctx, req.(*RefreshTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArtGeneratorService_CreateUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArtGeneratorServiceServer).CreateUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ArtGeneratorService_CreateUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArtGeneratorServiceServer).CreateUser(ctx, req.(*CreateUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArtGeneratorService_UpdateUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArtGeneratorServiceServer).UpdateUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ArtGeneratorService_UpdateUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArtGeneratorServiceServer).UpdateUser(ctx, req.(*UpdateUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArtGeneratorService_GetUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArtGeneratorServiceServer).GetUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ArtGeneratorService_GetUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArtGeneratorServiceServer).GetUser(ctx, req.(*GetUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArtGeneratorService_ResetPassword_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResetPasswordRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArtGeneratorServiceServer).ResetPassword(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ArtGeneratorService_ResetPassword_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArtGeneratorServiceServer).ResetPassword(ctx, req.(*ResetPasswordRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArtGeneratorService_ChangePassword_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChangePasswordRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArtGeneratorServiceServer).ChangePassword(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ArtGeneratorService_ChangePassword_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArtGeneratorServiceServer).ChangePassword(ctx, req.(*ChangePasswordRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArtGeneratorService_CreateArt_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateArtRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArtGeneratorServiceServer).CreateArt(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ArtGeneratorService_CreateArt_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArtGeneratorServiceServer).CreateArt(ctx, req.(*CreateArtRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArtGeneratorService_UpdateArt_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateArtRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArtGeneratorServiceServer).UpdateArt(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ArtGeneratorService_UpdateArt_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArtGeneratorServiceServer).UpdateArt(ctx, req.(*UpdateArtRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArtGeneratorService_GetArt_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetArtRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArtGeneratorServiceServer).GetArt(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ArtGeneratorService_GetArt_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArtGeneratorServiceServer).GetArt(ctx, req.(*GetArtRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArtGeneratorService_ListArts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListArtRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArtGeneratorServiceServer).ListArts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ArtGeneratorService_ListArts_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArtGeneratorServiceServer).ListArts(ctx, req.(*ListArtRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArtGeneratorService_DeleteArt_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteArtRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArtGeneratorServiceServer).DeleteArt(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ArtGeneratorService_DeleteArt_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArtGeneratorServiceServer).DeleteArt(ctx, req.(*DeleteArtRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ArtGeneratorService_ServiceDesc is the grpc.ServiceDesc for ArtGeneratorService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ArtGeneratorService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.ArtGeneratorService",
	HandlerType: (*ArtGeneratorServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "LoginUser",
			Handler:    _ArtGeneratorService_LoginUser_Handler,
		},
		{
			MethodName: "LogoutUser",
			Handler:    _ArtGeneratorService_LogoutUser_Handler,
		},
		{
			MethodName: "RefreshToken",
			Handler:    _ArtGeneratorService_RefreshToken_Handler,
		},
		{
			MethodName: "CreateUser",
			Handler:    _ArtGeneratorService_CreateUser_Handler,
		},
		{
			MethodName: "UpdateUser",
			Handler:    _ArtGeneratorService_UpdateUser_Handler,
		},
		{
			MethodName: "GetUser",
			Handler:    _ArtGeneratorService_GetUser_Handler,
		},
		{
			MethodName: "ResetPassword",
			Handler:    _ArtGeneratorService_ResetPassword_Handler,
		},
		{
			MethodName: "ChangePassword",
			Handler:    _ArtGeneratorService_ChangePassword_Handler,
		},
		{
			MethodName: "CreateArt",
			Handler:    _ArtGeneratorService_CreateArt_Handler,
		},
		{
			MethodName: "UpdateArt",
			Handler:    _ArtGeneratorService_UpdateArt_Handler,
		},
		{
			MethodName: "GetArt",
			Handler:    _ArtGeneratorService_GetArt_Handler,
		},
		{
			MethodName: "ListArts",
			Handler:    _ArtGeneratorService_ListArts_Handler,
		},
		{
			MethodName: "DeleteArt",
			Handler:    _ArtGeneratorService_DeleteArt_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "services.proto",
}
