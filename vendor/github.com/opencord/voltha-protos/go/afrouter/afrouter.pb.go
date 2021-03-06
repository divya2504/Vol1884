// Code generated by protoc-gen-go. DO NOT EDIT.
// source: voltha_protos/afrouter.proto

package afrouter

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	common "github.com/opencord/voltha-protos/go/common"
	grpc "google.golang.org/grpc"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Result struct {
	Success              bool     `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Error                string   `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	Info                 string   `protobuf:"bytes,3,opt,name=info,proto3" json:"info,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Result) Reset()         { *m = Result{} }
func (m *Result) String() string { return proto.CompactTextString(m) }
func (*Result) ProtoMessage()    {}
func (*Result) Descriptor() ([]byte, []int) {
	return fileDescriptor_be7e2f565b9e1c96, []int{0}
}

func (m *Result) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Result.Unmarshal(m, b)
}
func (m *Result) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Result.Marshal(b, m, deterministic)
}
func (m *Result) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Result.Merge(m, src)
}
func (m *Result) XXX_Size() int {
	return xxx_messageInfo_Result.Size(m)
}
func (m *Result) XXX_DiscardUnknown() {
	xxx_messageInfo_Result.DiscardUnknown(m)
}

var xxx_messageInfo_Result proto.InternalMessageInfo

func (m *Result) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *Result) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func (m *Result) GetInfo() string {
	if m != nil {
		return m.Info
	}
	return ""
}

type Empty struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Empty) Reset()         { *m = Empty{} }
func (m *Empty) String() string { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()    {}
func (*Empty) Descriptor() ([]byte, []int) {
	return fileDescriptor_be7e2f565b9e1c96, []int{1}
}

func (m *Empty) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Empty.Unmarshal(m, b)
}
func (m *Empty) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Empty.Marshal(b, m, deterministic)
}
func (m *Empty) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Empty.Merge(m, src)
}
func (m *Empty) XXX_Size() int {
	return xxx_messageInfo_Empty.Size(m)
}
func (m *Empty) XXX_DiscardUnknown() {
	xxx_messageInfo_Empty.DiscardUnknown(m)
}

var xxx_messageInfo_Empty proto.InternalMessageInfo

type Count struct {
	Count                uint32   `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Count) Reset()         { *m = Count{} }
func (m *Count) String() string { return proto.CompactTextString(m) }
func (*Count) ProtoMessage()    {}
func (*Count) Descriptor() ([]byte, []int) {
	return fileDescriptor_be7e2f565b9e1c96, []int{2}
}

func (m *Count) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Count.Unmarshal(m, b)
}
func (m *Count) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Count.Marshal(b, m, deterministic)
}
func (m *Count) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Count.Merge(m, src)
}
func (m *Count) XXX_Size() int {
	return xxx_messageInfo_Count.Size(m)
}
func (m *Count) XXX_DiscardUnknown() {
	xxx_messageInfo_Count.DiscardUnknown(m)
}

var xxx_messageInfo_Count proto.InternalMessageInfo

func (m *Count) GetCount() uint32 {
	if m != nil {
		return m.Count
	}
	return 0
}

type Conn struct {
	Server               string   `protobuf:"bytes,1,opt,name=server,proto3" json:"server,omitempty"`
	Pkg                  string   `protobuf:"bytes,2,opt,name=pkg,proto3" json:"pkg,omitempty"`
	Svc                  string   `protobuf:"bytes,3,opt,name=svc,proto3" json:"svc,omitempty"`
	Cluster              string   `protobuf:"bytes,4,opt,name=cluster,proto3" json:"cluster,omitempty"`
	Backend              string   `protobuf:"bytes,5,opt,name=backend,proto3" json:"backend,omitempty"`
	Connection           string   `protobuf:"bytes,6,opt,name=connection,proto3" json:"connection,omitempty"`
	Addr                 string   `protobuf:"bytes,7,opt,name=addr,proto3" json:"addr,omitempty"`
	Port                 uint64   `protobuf:"varint,8,opt,name=port,proto3" json:"port,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Conn) Reset()         { *m = Conn{} }
func (m *Conn) String() string { return proto.CompactTextString(m) }
func (*Conn) ProtoMessage()    {}
func (*Conn) Descriptor() ([]byte, []int) {
	return fileDescriptor_be7e2f565b9e1c96, []int{3}
}

func (m *Conn) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Conn.Unmarshal(m, b)
}
func (m *Conn) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Conn.Marshal(b, m, deterministic)
}
func (m *Conn) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Conn.Merge(m, src)
}
func (m *Conn) XXX_Size() int {
	return xxx_messageInfo_Conn.Size(m)
}
func (m *Conn) XXX_DiscardUnknown() {
	xxx_messageInfo_Conn.DiscardUnknown(m)
}

var xxx_messageInfo_Conn proto.InternalMessageInfo

func (m *Conn) GetServer() string {
	if m != nil {
		return m.Server
	}
	return ""
}

func (m *Conn) GetPkg() string {
	if m != nil {
		return m.Pkg
	}
	return ""
}

func (m *Conn) GetSvc() string {
	if m != nil {
		return m.Svc
	}
	return ""
}

func (m *Conn) GetCluster() string {
	if m != nil {
		return m.Cluster
	}
	return ""
}

func (m *Conn) GetBackend() string {
	if m != nil {
		return m.Backend
	}
	return ""
}

func (m *Conn) GetConnection() string {
	if m != nil {
		return m.Connection
	}
	return ""
}

func (m *Conn) GetAddr() string {
	if m != nil {
		return m.Addr
	}
	return ""
}

func (m *Conn) GetPort() uint64 {
	if m != nil {
		return m.Port
	}
	return 0
}

type Affinity struct {
	Router               string   `protobuf:"bytes,1,opt,name=router,proto3" json:"router,omitempty"`
	Route                string   `protobuf:"bytes,2,opt,name=route,proto3" json:"route,omitempty"`
	Cluster              string   `protobuf:"bytes,3,opt,name=cluster,proto3" json:"cluster,omitempty"`
	Backend              string   `protobuf:"bytes,4,opt,name=backend,proto3" json:"backend,omitempty"`
	Id                   string   `protobuf:"bytes,5,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Affinity) Reset()         { *m = Affinity{} }
func (m *Affinity) String() string { return proto.CompactTextString(m) }
func (*Affinity) ProtoMessage()    {}
func (*Affinity) Descriptor() ([]byte, []int) {
	return fileDescriptor_be7e2f565b9e1c96, []int{4}
}

func (m *Affinity) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Affinity.Unmarshal(m, b)
}
func (m *Affinity) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Affinity.Marshal(b, m, deterministic)
}
func (m *Affinity) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Affinity.Merge(m, src)
}
func (m *Affinity) XXX_Size() int {
	return xxx_messageInfo_Affinity.Size(m)
}
func (m *Affinity) XXX_DiscardUnknown() {
	xxx_messageInfo_Affinity.DiscardUnknown(m)
}

var xxx_messageInfo_Affinity proto.InternalMessageInfo

func (m *Affinity) GetRouter() string {
	if m != nil {
		return m.Router
	}
	return ""
}

func (m *Affinity) GetRoute() string {
	if m != nil {
		return m.Route
	}
	return ""
}

func (m *Affinity) GetCluster() string {
	if m != nil {
		return m.Cluster
	}
	return ""
}

func (m *Affinity) GetBackend() string {
	if m != nil {
		return m.Backend
	}
	return ""
}

func (m *Affinity) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func init() {
	proto.RegisterType((*Result)(nil), "afrouter.Result")
	proto.RegisterType((*Empty)(nil), "afrouter.Empty")
	proto.RegisterType((*Count)(nil), "afrouter.Count")
	proto.RegisterType((*Conn)(nil), "afrouter.Conn")
	proto.RegisterType((*Affinity)(nil), "afrouter.Affinity")
}

func init() { proto.RegisterFile("voltha_protos/afrouter.proto", fileDescriptor_be7e2f565b9e1c96) }

var fileDescriptor_be7e2f565b9e1c96 = []byte{
	// 463 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x53, 0xcb, 0x8e, 0xd3, 0x30,
	0x14, 0x6d, 0xda, 0xf4, 0x31, 0x17, 0xfa, 0xc0, 0x42, 0xc8, 0x8a, 0x00, 0x55, 0x59, 0x75, 0x33,
	0xad, 0xc4, 0x80, 0xd8, 0xb0, 0x81, 0x08, 0x75, 0xd3, 0x55, 0x46, 0x6c, 0xd8, 0xa0, 0xd4, 0x71,
	0x32, 0xd6, 0xb4, 0xbe, 0x91, 0xed, 0x54, 0x1a, 0x89, 0x0f, 0xe2, 0x2b, 0xf8, 0x36, 0x64, 0x3b,
	0xe9, 0x4b, 0xcc, 0xee, 0x9c, 0x73, 0xef, 0xb5, 0xcf, 0xf1, 0x03, 0xde, 0x1e, 0x70, 0x67, 0x1e,
	0xb2, 0x5f, 0x95, 0x42, 0x83, 0x7a, 0x95, 0x15, 0x0a, 0x6b, 0xc3, 0xd5, 0xd2, 0x71, 0x32, 0x6a,
	0x79, 0x14, 0x5d, 0xf6, 0x31, 0xdc, 0xef, 0x51, 0xfa, 0xae, 0x78, 0x03, 0x83, 0x94, 0xeb, 0x7a,
	0x67, 0x08, 0x85, 0xa1, 0xae, 0x19, 0xe3, 0x5a, 0xd3, 0x60, 0x1e, 0x2c, 0x46, 0x69, 0x4b, 0xc9,
	0x6b, 0xe8, 0x73, 0xa5, 0x50, 0xd1, 0xee, 0x3c, 0x58, 0xdc, 0xa4, 0x9e, 0x10, 0x02, 0xa1, 0x90,
	0x05, 0xd2, 0x9e, 0x13, 0x1d, 0x8e, 0x87, 0xd0, 0xff, 0xbe, 0xaf, 0xcc, 0x53, 0xfc, 0x0e, 0xfa,
	0x09, 0xd6, 0xd2, 0xd8, 0x59, 0x66, 0x81, 0x5b, 0x73, 0x9c, 0x7a, 0x12, 0xff, 0x0d, 0x20, 0x4c,
	0x50, 0x4a, 0xf2, 0x06, 0x06, 0x9a, 0xab, 0x03, 0x57, 0xae, 0x7e, 0x93, 0x36, 0x8c, 0xcc, 0xa0,
	0x57, 0x3d, 0x96, 0xcd, 0x86, 0x16, 0x5a, 0x45, 0x1f, 0x58, 0xb3, 0x9b, 0x85, 0xd6, 0x30, 0xdb,
	0xd5, 0xda, 0x70, 0x45, 0x43, 0xa7, 0xb6, 0xd4, 0x56, 0xb6, 0x19, 0x7b, 0xe4, 0x32, 0xa7, 0x7d,
	0x5f, 0x69, 0x28, 0x79, 0x0f, 0xc0, 0x50, 0x4a, 0xce, 0x8c, 0x40, 0x49, 0x07, 0xae, 0x78, 0xa6,
	0xd8, 0x50, 0x59, 0x9e, 0x2b, 0x3a, 0xf4, 0xa1, 0x2c, 0xb6, 0x5a, 0x85, 0xca, 0xd0, 0xd1, 0x3c,
	0x58, 0x84, 0xa9, 0xc3, 0xf1, 0x6f, 0x18, 0x7d, 0x2d, 0x0a, 0x21, 0x85, 0x79, 0xb2, 0x19, 0xfc,
	0x41, 0xb7, 0x19, 0x3c, 0xb3, 0xd1, 0x1d, 0x6a, 0x8f, 0xcd, 0x91, 0x73, 0xd7, 0xbd, 0x67, 0x5d,
	0x87, 0x97, 0xae, 0x27, 0xd0, 0x15, 0x6d, 0x94, 0xae, 0xc8, 0x3f, 0xfc, 0xe9, 0xc2, 0x38, 0x41,
	0x59, 0x88, 0xb2, 0x56, 0x99, 0xf3, 0x7d, 0x07, 0xe3, 0x7b, 0x6e, 0x92, 0x53, 0x90, 0xc9, 0xf2,
	0xf8, 0x1c, 0xac, 0x1a, 0xcd, 0x4e, 0xdc, 0xdf, 0x77, 0xdc, 0x21, 0x9f, 0xe0, 0xc5, 0x3d, 0x37,
	0xc7, 0x1c, 0xe4, 0xd4, 0xd2, 0x6a, 0xff, 0x1d, 0xfb, 0x0c, 0xaf, 0xd6, 0xdc, 0xac, 0xd1, 0xea,
	0x42, 0x72, 0x7f, 0xcf, 0xd3, 0x53, 0xa3, 0x7b, 0x01, 0xd1, 0xf4, 0xdc, 0x80, 0xbd, 0xf3, 0x0e,
	0xf9, 0x08, 0x93, 0x1f, 0x55, 0x9e, 0x19, 0xbe, 0xc1, 0x72, 0xc3, 0x0f, 0x7c, 0x47, 0xa6, 0xcb,
	0xe6, 0x31, 0x6e, 0xb0, 0x2c, 0x85, 0x2c, 0xa3, 0xeb, 0x65, 0xe2, 0x0e, 0xf9, 0x02, 0x2f, 0xd7,
	0xdc, 0xb4, 0x23, 0x9a, 0xd0, 0xab, 0x99, 0x04, 0xf7, 0x15, 0x4a, 0x2e, 0x4d, 0x34, 0xbb, 0xaa,
	0xe8, 0xb8, 0xf3, 0x6d, 0xf5, 0xf3, 0xb6, 0x14, 0xe6, 0xa1, 0xde, 0xda, 0xda, 0x0a, 0x2b, 0x2e,
	0x19, 0xaa, 0x7c, 0xe5, 0x7f, 0xc4, 0x6d, 0xf3, 0x23, 0x4a, 0x3c, 0x7e, 0x9e, 0xed, 0xc0, 0x69,
	0x77, 0xff, 0x02, 0x00, 0x00, 0xff, 0xff, 0xec, 0x41, 0x8b, 0x29, 0x5d, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ConfigurationClient is the client API for Configuration service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ConfigurationClient interface {
	SetConnection(ctx context.Context, in *Conn, opts ...grpc.CallOption) (*Result, error)
	SetAffinity(ctx context.Context, in *Affinity, opts ...grpc.CallOption) (*Result, error)
	GetGoroutineCount(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Count, error)
	UpdateLogLevel(ctx context.Context, in *common.Logging, opts ...grpc.CallOption) (*Empty, error)
	GetLogLevels(ctx context.Context, in *common.LoggingComponent, opts ...grpc.CallOption) (*common.Loggings, error)
}

type configurationClient struct {
	cc *grpc.ClientConn
}

func NewConfigurationClient(cc *grpc.ClientConn) ConfigurationClient {
	return &configurationClient{cc}
}

func (c *configurationClient) SetConnection(ctx context.Context, in *Conn, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := c.cc.Invoke(ctx, "/afrouter.Configuration/SetConnection", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *configurationClient) SetAffinity(ctx context.Context, in *Affinity, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := c.cc.Invoke(ctx, "/afrouter.Configuration/SetAffinity", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *configurationClient) GetGoroutineCount(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Count, error) {
	out := new(Count)
	err := c.cc.Invoke(ctx, "/afrouter.Configuration/GetGoroutineCount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *configurationClient) UpdateLogLevel(ctx context.Context, in *common.Logging, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/afrouter.Configuration/UpdateLogLevel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *configurationClient) GetLogLevels(ctx context.Context, in *common.LoggingComponent, opts ...grpc.CallOption) (*common.Loggings, error) {
	out := new(common.Loggings)
	err := c.cc.Invoke(ctx, "/afrouter.Configuration/GetLogLevels", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ConfigurationServer is the server API for Configuration service.
type ConfigurationServer interface {
	SetConnection(context.Context, *Conn) (*Result, error)
	SetAffinity(context.Context, *Affinity) (*Result, error)
	GetGoroutineCount(context.Context, *Empty) (*Count, error)
	UpdateLogLevel(context.Context, *common.Logging) (*Empty, error)
	GetLogLevels(context.Context, *common.LoggingComponent) (*common.Loggings, error)
}

func RegisterConfigurationServer(s *grpc.Server, srv ConfigurationServer) {
	s.RegisterService(&_Configuration_serviceDesc, srv)
}

func _Configuration_SetConnection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Conn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConfigurationServer).SetConnection(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/afrouter.Configuration/SetConnection",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConfigurationServer).SetConnection(ctx, req.(*Conn))
	}
	return interceptor(ctx, in, info, handler)
}

func _Configuration_SetAffinity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Affinity)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConfigurationServer).SetAffinity(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/afrouter.Configuration/SetAffinity",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConfigurationServer).SetAffinity(ctx, req.(*Affinity))
	}
	return interceptor(ctx, in, info, handler)
}

func _Configuration_GetGoroutineCount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConfigurationServer).GetGoroutineCount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/afrouter.Configuration/GetGoroutineCount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConfigurationServer).GetGoroutineCount(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Configuration_UpdateLogLevel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(common.Logging)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConfigurationServer).UpdateLogLevel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/afrouter.Configuration/UpdateLogLevel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConfigurationServer).UpdateLogLevel(ctx, req.(*common.Logging))
	}
	return interceptor(ctx, in, info, handler)
}

func _Configuration_GetLogLevels_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(common.LoggingComponent)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConfigurationServer).GetLogLevels(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/afrouter.Configuration/GetLogLevels",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConfigurationServer).GetLogLevels(ctx, req.(*common.LoggingComponent))
	}
	return interceptor(ctx, in, info, handler)
}

var _Configuration_serviceDesc = grpc.ServiceDesc{
	ServiceName: "afrouter.Configuration",
	HandlerType: (*ConfigurationServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetConnection",
			Handler:    _Configuration_SetConnection_Handler,
		},
		{
			MethodName: "SetAffinity",
			Handler:    _Configuration_SetAffinity_Handler,
		},
		{
			MethodName: "GetGoroutineCount",
			Handler:    _Configuration_GetGoroutineCount_Handler,
		},
		{
			MethodName: "UpdateLogLevel",
			Handler:    _Configuration_UpdateLogLevel_Handler,
		},
		{
			MethodName: "GetLogLevels",
			Handler:    _Configuration_GetLogLevels_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "voltha_protos/afrouter.proto",
}
