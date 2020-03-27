// Code generated by protoc-gen-go. DO NOT EDIT.
// source: envoy/config/accesslog/v2/file.proto

package envoy_config_accesslog_v2

import (
	fmt "fmt"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	proto "github.com/golang/protobuf/proto"
	_struct "github.com/golang/protobuf/ptypes/struct"
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

type FileAccessLog struct {
	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	// Types that are valid to be assigned to AccessLogFormat:
	//	*FileAccessLog_Format
	//	*FileAccessLog_JsonFormat
	AccessLogFormat      isFileAccessLog_AccessLogFormat `protobuf_oneof:"access_log_format"`
	XXX_NoUnkeyedLiteral struct{}                        `json:"-"`
	XXX_unrecognized     []byte                          `json:"-"`
	XXX_sizecache        int32                           `json:"-"`
}

func (m *FileAccessLog) Reset()         { *m = FileAccessLog{} }
func (m *FileAccessLog) String() string { return proto.CompactTextString(m) }
func (*FileAccessLog) ProtoMessage()    {}
func (*FileAccessLog) Descriptor() ([]byte, []int) {
	return fileDescriptor_bb42a04cfa71ce3c, []int{0}
}

func (m *FileAccessLog) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FileAccessLog.Unmarshal(m, b)
}
func (m *FileAccessLog) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FileAccessLog.Marshal(b, m, deterministic)
}
func (m *FileAccessLog) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FileAccessLog.Merge(m, src)
}
func (m *FileAccessLog) XXX_Size() int {
	return xxx_messageInfo_FileAccessLog.Size(m)
}
func (m *FileAccessLog) XXX_DiscardUnknown() {
	xxx_messageInfo_FileAccessLog.DiscardUnknown(m)
}

var xxx_messageInfo_FileAccessLog proto.InternalMessageInfo

func (m *FileAccessLog) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

type isFileAccessLog_AccessLogFormat interface {
	isFileAccessLog_AccessLogFormat()
}

type FileAccessLog_Format struct {
	Format string `protobuf:"bytes,2,opt,name=format,proto3,oneof"`
}

type FileAccessLog_JsonFormat struct {
	JsonFormat *_struct.Struct `protobuf:"bytes,3,opt,name=json_format,json=jsonFormat,proto3,oneof"`
}

func (*FileAccessLog_Format) isFileAccessLog_AccessLogFormat() {}

func (*FileAccessLog_JsonFormat) isFileAccessLog_AccessLogFormat() {}

func (m *FileAccessLog) GetAccessLogFormat() isFileAccessLog_AccessLogFormat {
	if m != nil {
		return m.AccessLogFormat
	}
	return nil
}

func (m *FileAccessLog) GetFormat() string {
	if x, ok := m.GetAccessLogFormat().(*FileAccessLog_Format); ok {
		return x.Format
	}
	return ""
}

func (m *FileAccessLog) GetJsonFormat() *_struct.Struct {
	if x, ok := m.GetAccessLogFormat().(*FileAccessLog_JsonFormat); ok {
		return x.JsonFormat
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*FileAccessLog) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*FileAccessLog_Format)(nil),
		(*FileAccessLog_JsonFormat)(nil),
	}
}

func init() {
	proto.RegisterType((*FileAccessLog)(nil), "envoy.config.accesslog.v2.FileAccessLog")
}

func init() {
	proto.RegisterFile("envoy/config/accesslog/v2/file.proto", fileDescriptor_bb42a04cfa71ce3c)
}

var fileDescriptor_bb42a04cfa71ce3c = []byte{
	// 252 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x8f, 0xb1, 0x4e, 0xc3, 0x30,
	0x10, 0x86, 0xeb, 0x52, 0x15, 0xd5, 0x15, 0x03, 0x61, 0x68, 0x28, 0x0c, 0x11, 0x42, 0xa2, 0xd3,
	0x59, 0x0a, 0x12, 0x03, 0x1b, 0x1e, 0xaa, 0x0e, 0x0c, 0x55, 0x78, 0x80, 0xca, 0x0d, 0x8e, 0x31,
	0x32, 0xb9, 0xc8, 0x71, 0x23, 0xfa, 0x24, 0xbc, 0x2b, 0x13, 0xb2, 0x9d, 0xb0, 0x75, 0xf3, 0xf9,
	0xff, 0xee, 0x3e, 0xfd, 0xf4, 0x5e, 0xd6, 0x1d, 0x1e, 0x59, 0x89, 0x75, 0xa5, 0x15, 0x13, 0x65,
	0x29, 0xdb, 0xd6, 0xa0, 0x62, 0x5d, 0xce, 0x2a, 0x6d, 0x24, 0x34, 0x16, 0x1d, 0x26, 0xd7, 0x81,
	0x82, 0x48, 0xc1, 0x3f, 0x05, 0x5d, 0xbe, 0x5c, 0x74, 0xc2, 0xe8, 0x77, 0xe1, 0x24, 0x1b, 0x1e,
	0x71, 0x67, 0x79, 0xab, 0x10, 0x95, 0x91, 0x2c, 0x4c, 0xfb, 0x43, 0xc5, 0x5a, 0x67, 0x0f, 0xa5,
	0x8b, 0xe9, 0xdd, 0x0f, 0xa1, 0x17, 0x6b, 0x6d, 0xe4, 0x4b, 0xb8, 0xf5, 0x8a, 0x2a, 0xb9, 0xa1,
	0x93, 0x46, 0xb8, 0x8f, 0x94, 0x64, 0x64, 0x35, 0xe3, 0xe7, 0xbf, 0x7c, 0x62, 0xc7, 0x19, 0x29,
	0xc2, 0x67, 0x92, 0xd2, 0x69, 0x85, 0xf6, 0x4b, 0xb8, 0x74, 0xec, 0xe3, 0xcd, 0xa8, 0xe8, 0xe7,
	0xe4, 0x99, 0xce, 0x3f, 0x5b, 0xac, 0x77, 0x7d, 0x7c, 0x96, 0x91, 0xd5, 0x3c, 0x5f, 0x40, 0x94,
	0xc3, 0x20, 0x87, 0xb7, 0x20, 0xdf, 0x8c, 0x0a, 0xea, 0xe9, 0x75, 0x80, 0xf9, 0x15, 0xbd, 0x8c,
	0x5d, 0x76, 0x06, 0x55, 0x7f, 0x81, 0x3f, 0xd1, 0x07, 0x8d, 0x10, 0x0a, 0x37, 0x16, 0xbf, 0x8f,
	0x70, 0xb2, 0x3b, 0x9f, 0xf9, 0x06, 0x5b, 0xaf, 0xd8, 0x92, 0xfd, 0x34, 0xb8, 0x1e, 0xff, 0x02,
	0x00, 0x00, 0xff, 0xff, 0xc1, 0xdb, 0x63, 0x77, 0x52, 0x01, 0x00, 0x00,
}
