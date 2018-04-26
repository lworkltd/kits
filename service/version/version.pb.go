// Code generated by protoc-gen-go. DO NOT EDIT.
// source: version.proto

/*
Package version is a generated protocol buffer package.

It is generated from these files:
	version.proto

It has these top-level messages:
	VersionResponse
*/
package version

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// VersionResponse 版本信息
type VersionResponse struct {
	// 编译版本
	Golang string `protobuf:"bytes,1,opt,name=golang" json:"golang"`
	// 占用的核心数
	Cpus int32 `protobuf:"varint,2,opt,name=cpus" json:"cpus"`
	// 当前协程数
	Routines int32 `protobuf:"varint,3,opt,name=routines" json:"routines"`
	// 应用名称
	Name string `protobuf:"bytes,4,opt,name=name" json:"name"`
	// version 为{branch}.{buildId}
	Version string `protobuf:"bytes,5,opt,name=version" json:"version"`
}

func (m *VersionResponse) Reset()                    { *m = VersionResponse{} }
func (m *VersionResponse) String() string            { return proto.CompactTextString(m) }
func (*VersionResponse) ProtoMessage()               {}
func (*VersionResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *VersionResponse) GetGolang() string {
	if m != nil {
		return m.Golang
	}
	return ""
}

func (m *VersionResponse) GetCpus() int32 {
	if m != nil {
		return m.Cpus
	}
	return 0
}

func (m *VersionResponse) GetRoutines() int32 {
	if m != nil {
		return m.Routines
	}
	return 0
}

func (m *VersionResponse) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *VersionResponse) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func init() {
	proto.RegisterType((*VersionResponse)(nil), "version.VersionResponse")
}

func init() { proto.RegisterFile("version.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 140 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2d, 0x4b, 0x2d, 0x2a,
	0xce, 0xcc, 0xcf, 0xd3, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x87, 0x72, 0x95, 0xda, 0x19,
	0xb9, 0xf8, 0xc3, 0x20, 0xec, 0xa0, 0xd4, 0xe2, 0x82, 0xfc, 0xbc, 0xe2, 0x54, 0x21, 0x31, 0x2e,
	0xb6, 0xf4, 0xfc, 0x9c, 0xc4, 0xbc, 0x74, 0x09, 0x46, 0x05, 0x46, 0x0d, 0xce, 0x20, 0x28, 0x4f,
	0x48, 0x88, 0x8b, 0x25, 0xb9, 0xa0, 0xb4, 0x58, 0x82, 0x49, 0x81, 0x51, 0x83, 0x35, 0x08, 0xcc,
	0x16, 0x92, 0xe2, 0xe2, 0x28, 0xca, 0x2f, 0x2d, 0xc9, 0xcc, 0x4b, 0x2d, 0x96, 0x60, 0x06, 0x8b,
	0xc3, 0xf9, 0x20, 0xf5, 0x79, 0x89, 0xb9, 0xa9, 0x12, 0x2c, 0x60, 0x53, 0xc0, 0x6c, 0x21, 0x09,
	0x2e, 0x98, 0xd5, 0x12, 0xac, 0x60, 0x61, 0x18, 0x37, 0x89, 0x0d, 0xec, 0x32, 0x63, 0x40, 0x00,
	0x00, 0x00, 0xff, 0xff, 0xb7, 0xb6, 0x51, 0x55, 0xaa, 0x00, 0x00, 0x00,
}