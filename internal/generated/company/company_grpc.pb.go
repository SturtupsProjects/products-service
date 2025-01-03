// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: user/company.proto

package company

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

// CompanyServiceClient is the client API for CompanyService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CompanyServiceClient interface {
	CreateCompany(ctx context.Context, in *CreateCompanyRequest, opts ...grpc.CallOption) (*CompanyResponse, error)
	GetCompany(ctx context.Context, in *GetCompanyRequest, opts ...grpc.CallOption) (*CompanyResponse, error)
	UpdateCompany(ctx context.Context, in *UpdateCompanyRequest, opts ...grpc.CallOption) (*CompanyResponse, error)
	DeleteCompany(ctx context.Context, in *DeleteCompanyRequest, opts ...grpc.CallOption) (*Message, error)
	ListCompanies(ctx context.Context, in *ListCompaniesRequest, opts ...grpc.CallOption) (*ListCompaniesResponse, error)
	ListCompanyUsers(ctx context.Context, in *ListCompanyUsersRequest, opts ...grpc.CallOption) (*ListCompanyUsersResponse, error)
	CreateUserToCompany(ctx context.Context, in *CreateUserToCompanyRequest, opts ...grpc.CallOption) (*Id, error)
}

type companyServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCompanyServiceClient(cc grpc.ClientConnInterface) CompanyServiceClient {
	return &companyServiceClient{cc}
}

func (c *companyServiceClient) CreateCompany(ctx context.Context, in *CreateCompanyRequest, opts ...grpc.CallOption) (*CompanyResponse, error) {
	out := new(CompanyResponse)
	err := c.cc.Invoke(ctx, "/company.CompanyService/CreateCompany", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *companyServiceClient) GetCompany(ctx context.Context, in *GetCompanyRequest, opts ...grpc.CallOption) (*CompanyResponse, error) {
	out := new(CompanyResponse)
	err := c.cc.Invoke(ctx, "/company.CompanyService/GetCompany", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *companyServiceClient) UpdateCompany(ctx context.Context, in *UpdateCompanyRequest, opts ...grpc.CallOption) (*CompanyResponse, error) {
	out := new(CompanyResponse)
	err := c.cc.Invoke(ctx, "/company.CompanyService/UpdateCompany", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *companyServiceClient) DeleteCompany(ctx context.Context, in *DeleteCompanyRequest, opts ...grpc.CallOption) (*Message, error) {
	out := new(Message)
	err := c.cc.Invoke(ctx, "/company.CompanyService/DeleteCompany", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *companyServiceClient) ListCompanies(ctx context.Context, in *ListCompaniesRequest, opts ...grpc.CallOption) (*ListCompaniesResponse, error) {
	out := new(ListCompaniesResponse)
	err := c.cc.Invoke(ctx, "/company.CompanyService/ListCompanies", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *companyServiceClient) ListCompanyUsers(ctx context.Context, in *ListCompanyUsersRequest, opts ...grpc.CallOption) (*ListCompanyUsersResponse, error) {
	out := new(ListCompanyUsersResponse)
	err := c.cc.Invoke(ctx, "/company.CompanyService/ListCompanyUsers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *companyServiceClient) CreateUserToCompany(ctx context.Context, in *CreateUserToCompanyRequest, opts ...grpc.CallOption) (*Id, error) {
	out := new(Id)
	err := c.cc.Invoke(ctx, "/company.CompanyService/CreateUserToCompany", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CompanyServiceServer is the server API for CompanyService service.
// All implementations must embed UnimplementedCompanyServiceServer
// for forward compatibility
type CompanyServiceServer interface {
	CreateCompany(context.Context, *CreateCompanyRequest) (*CompanyResponse, error)
	GetCompany(context.Context, *GetCompanyRequest) (*CompanyResponse, error)
	UpdateCompany(context.Context, *UpdateCompanyRequest) (*CompanyResponse, error)
	DeleteCompany(context.Context, *DeleteCompanyRequest) (*Message, error)
	ListCompanies(context.Context, *ListCompaniesRequest) (*ListCompaniesResponse, error)
	ListCompanyUsers(context.Context, *ListCompanyUsersRequest) (*ListCompanyUsersResponse, error)
	CreateUserToCompany(context.Context, *CreateUserToCompanyRequest) (*Id, error)
	mustEmbedUnimplementedCompanyServiceServer()
}

// UnimplementedCompanyServiceServer must be embedded to have forward compatible implementations.
type UnimplementedCompanyServiceServer struct {
}

func (UnimplementedCompanyServiceServer) CreateCompany(context.Context, *CreateCompanyRequest) (*CompanyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateCompany not implemented")
}
func (UnimplementedCompanyServiceServer) GetCompany(context.Context, *GetCompanyRequest) (*CompanyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCompany not implemented")
}
func (UnimplementedCompanyServiceServer) UpdateCompany(context.Context, *UpdateCompanyRequest) (*CompanyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateCompany not implemented")
}
func (UnimplementedCompanyServiceServer) DeleteCompany(context.Context, *DeleteCompanyRequest) (*Message, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteCompany not implemented")
}
func (UnimplementedCompanyServiceServer) ListCompanies(context.Context, *ListCompaniesRequest) (*ListCompaniesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListCompanies not implemented")
}
func (UnimplementedCompanyServiceServer) ListCompanyUsers(context.Context, *ListCompanyUsersRequest) (*ListCompanyUsersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListCompanyUsers not implemented")
}
func (UnimplementedCompanyServiceServer) CreateUserToCompany(context.Context, *CreateUserToCompanyRequest) (*Id, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateUserToCompany not implemented")
}
func (UnimplementedCompanyServiceServer) mustEmbedUnimplementedCompanyServiceServer() {}

// UnsafeCompanyServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CompanyServiceServer will
// result in compilation errors.
type UnsafeCompanyServiceServer interface {
	mustEmbedUnimplementedCompanyServiceServer()
}

func RegisterCompanyServiceServer(s grpc.ServiceRegistrar, srv CompanyServiceServer) {
	s.RegisterService(&CompanyService_ServiceDesc, srv)
}

func _CompanyService_CreateCompany_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateCompanyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CompanyServiceServer).CreateCompany(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/company.CompanyService/CreateCompany",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CompanyServiceServer).CreateCompany(ctx, req.(*CreateCompanyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CompanyService_GetCompany_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCompanyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CompanyServiceServer).GetCompany(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/company.CompanyService/GetCompany",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CompanyServiceServer).GetCompany(ctx, req.(*GetCompanyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CompanyService_UpdateCompany_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateCompanyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CompanyServiceServer).UpdateCompany(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/company.CompanyService/UpdateCompany",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CompanyServiceServer).UpdateCompany(ctx, req.(*UpdateCompanyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CompanyService_DeleteCompany_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteCompanyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CompanyServiceServer).DeleteCompany(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/company.CompanyService/DeleteCompany",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CompanyServiceServer).DeleteCompany(ctx, req.(*DeleteCompanyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CompanyService_ListCompanies_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListCompaniesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CompanyServiceServer).ListCompanies(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/company.CompanyService/ListCompanies",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CompanyServiceServer).ListCompanies(ctx, req.(*ListCompaniesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CompanyService_ListCompanyUsers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListCompanyUsersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CompanyServiceServer).ListCompanyUsers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/company.CompanyService/ListCompanyUsers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CompanyServiceServer).ListCompanyUsers(ctx, req.(*ListCompanyUsersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CompanyService_CreateUserToCompany_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateUserToCompanyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CompanyServiceServer).CreateUserToCompany(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/company.CompanyService/CreateUserToCompany",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CompanyServiceServer).CreateUserToCompany(ctx, req.(*CreateUserToCompanyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CompanyService_ServiceDesc is the grpc.ServiceDesc for CompanyService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CompanyService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "company.CompanyService",
	HandlerType: (*CompanyServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateCompany",
			Handler:    _CompanyService_CreateCompany_Handler,
		},
		{
			MethodName: "GetCompany",
			Handler:    _CompanyService_GetCompany_Handler,
		},
		{
			MethodName: "UpdateCompany",
			Handler:    _CompanyService_UpdateCompany_Handler,
		},
		{
			MethodName: "DeleteCompany",
			Handler:    _CompanyService_DeleteCompany_Handler,
		},
		{
			MethodName: "ListCompanies",
			Handler:    _CompanyService_ListCompanies_Handler,
		},
		{
			MethodName: "ListCompanyUsers",
			Handler:    _CompanyService_ListCompanyUsers_Handler,
		},
		{
			MethodName: "CreateUserToCompany",
			Handler:    _CompanyService_CreateUserToCompany_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "user/company.proto",
}
