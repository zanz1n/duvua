// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.28.3
// source: api/proto/davinci/davinci.proto

package davinci

import (
	_ "github.com/srikrsna/protoc-gen-gotag/tagger"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ImageSendData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ChannelId uint64 `protobuf:"fixed64,1,opt,name=channel_id,json=channelId,proto3" json:"channel_id,omitempty" validate:"required"`
	Message   string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	FileName  string `protobuf:"bytes,3,opt,name=file_name,json=fileName,proto3" json:"file_name,omitempty" validate:"required"`
}

func (x *ImageSendData) Reset() {
	*x = ImageSendData{}
	mi := &file_api_proto_davinci_davinci_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ImageSendData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ImageSendData) ProtoMessage() {}

func (x *ImageSendData) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_davinci_davinci_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ImageSendData.ProtoReflect.Descriptor instead.
func (*ImageSendData) Descriptor() ([]byte, []int) {
	return file_api_proto_davinci_davinci_proto_rawDescGZIP(), []int{0}
}

func (x *ImageSendData) GetChannelId() uint64 {
	if x != nil {
		return x.ChannelId
	}
	return 0
}

func (x *ImageSendData) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *ImageSendData) GetFileName() string {
	if x != nil {
		return x.FileName
	}
	return ""
}

type WelcomeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Username     string         `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	ImageUrl     string         `protobuf:"bytes,2,opt,name=image_url,json=imageUrl,proto3" json:"image_url,omitempty"`
	GreetingText string         `protobuf:"bytes,3,opt,name=greeting_text,json=greetingText,proto3" json:"greeting_text,omitempty"`
	Data         *ImageSendData `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty" validate:"required"`
}

func (x *WelcomeRequest) Reset() {
	*x = WelcomeRequest{}
	mi := &file_api_proto_davinci_davinci_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *WelcomeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WelcomeRequest) ProtoMessage() {}

func (x *WelcomeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_davinci_davinci_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WelcomeRequest.ProtoReflect.Descriptor instead.
func (*WelcomeRequest) Descriptor() ([]byte, []int) {
	return file_api_proto_davinci_davinci_proto_rawDescGZIP(), []int{1}
}

func (x *WelcomeRequest) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *WelcomeRequest) GetImageUrl() string {
	if x != nil {
		return x.ImageUrl
	}
	return ""
}

func (x *WelcomeRequest) GetGreetingText() string {
	if x != nil {
		return x.GreetingText
	}
	return ""
}

func (x *WelcomeRequest) GetData() *ImageSendData {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_api_proto_davinci_davinci_proto protoreflect.FileDescriptor

var file_api_proto_davinci_davinci_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x64, 0x61, 0x76, 0x69,
	0x6e, 0x63, 0x69, 0x2f, 0x64, 0x61, 0x76, 0x69, 0x6e, 0x63, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x07, 0x64, 0x61, 0x76, 0x69, 0x6e, 0x63, 0x69, 0x1a, 0x13, 0x74, 0x61, 0x67, 0x67,
	0x65, 0x72, 0x2f, 0x74, 0x61, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x99, 0x01, 0x0a,
	0x0d, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x53, 0x65, 0x6e, 0x64, 0x44, 0x61, 0x74, 0x61, 0x12, 0x37,
	0x0a, 0x0a, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x06, 0x42, 0x18, 0x9a, 0x84, 0x9e, 0x03, 0x13, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x65, 0x3a, 0x22, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x22, 0x52, 0x09, 0x63, 0x68,
	0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x12, 0x35, 0x0a, 0x09, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x18, 0x9a, 0x84, 0x9e, 0x03, 0x13, 0x76, 0x61, 0x6c, 0x69, 0x64,
	0x61, 0x74, 0x65, 0x3a, 0x22, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x22, 0x52, 0x08,
	0x66, 0x69, 0x6c, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x22, 0xb4, 0x01, 0x0a, 0x0e, 0x57, 0x65, 0x6c,
	0x63, 0x6f, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x75,
	0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x75,
	0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x69, 0x6d, 0x61, 0x67, 0x65,
	0x5f, 0x75, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x69, 0x6d, 0x61, 0x67,
	0x65, 0x55, 0x72, 0x6c, 0x12, 0x23, 0x0a, 0x0d, 0x67, 0x72, 0x65, 0x65, 0x74, 0x69, 0x6e, 0x67,
	0x5f, 0x74, 0x65, 0x78, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x67, 0x72, 0x65,
	0x65, 0x74, 0x69, 0x6e, 0x67, 0x54, 0x65, 0x78, 0x74, 0x12, 0x44, 0x0a, 0x04, 0x64, 0x61, 0x74,
	0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x64, 0x61, 0x76, 0x69, 0x6e, 0x63,
	0x69, 0x2e, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x53, 0x65, 0x6e, 0x64, 0x44, 0x61, 0x74, 0x61, 0x42,
	0x18, 0x9a, 0x84, 0x9e, 0x03, 0x13, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x3a, 0x22,
	0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x22, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x32,
	0x49, 0x0a, 0x07, 0x44, 0x61, 0x76, 0x69, 0x6e, 0x63, 0x69, 0x12, 0x3e, 0x0a, 0x0b, 0x53, 0x65,
	0x6e, 0x64, 0x57, 0x65, 0x6c, 0x63, 0x6f, 0x6d, 0x65, 0x12, 0x17, 0x2e, 0x64, 0x61, 0x76, 0x69,
	0x6e, 0x63, 0x69, 0x2e, 0x57, 0x65, 0x6c, 0x63, 0x6f, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x42, 0x0b, 0x5a, 0x09, 0x2e, 0x2f,
	0x64, 0x61, 0x76, 0x69, 0x6e, 0x63, 0x69, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_davinci_davinci_proto_rawDescOnce sync.Once
	file_api_proto_davinci_davinci_proto_rawDescData = file_api_proto_davinci_davinci_proto_rawDesc
)

func file_api_proto_davinci_davinci_proto_rawDescGZIP() []byte {
	file_api_proto_davinci_davinci_proto_rawDescOnce.Do(func() {
		file_api_proto_davinci_davinci_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_davinci_davinci_proto_rawDescData)
	})
	return file_api_proto_davinci_davinci_proto_rawDescData
}

var file_api_proto_davinci_davinci_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_davinci_davinci_proto_goTypes = []any{
	(*ImageSendData)(nil),  // 0: davinci.ImageSendData
	(*WelcomeRequest)(nil), // 1: davinci.WelcomeRequest
	(*emptypb.Empty)(nil),  // 2: google.protobuf.Empty
}
var file_api_proto_davinci_davinci_proto_depIdxs = []int32{
	0, // 0: davinci.WelcomeRequest.data:type_name -> davinci.ImageSendData
	1, // 1: davinci.Davinci.SendWelcome:input_type -> davinci.WelcomeRequest
	2, // 2: davinci.Davinci.SendWelcome:output_type -> google.protobuf.Empty
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_api_proto_davinci_davinci_proto_init() }
func file_api_proto_davinci_davinci_proto_init() {
	if File_api_proto_davinci_davinci_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_proto_davinci_davinci_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_proto_davinci_davinci_proto_goTypes,
		DependencyIndexes: file_api_proto_davinci_davinci_proto_depIdxs,
		MessageInfos:      file_api_proto_davinci_davinci_proto_msgTypes,
	}.Build()
	File_api_proto_davinci_davinci_proto = out.File
	file_api_proto_davinci_davinci_proto_rawDesc = nil
	file_api_proto_davinci_davinci_proto_goTypes = nil
	file_api_proto_davinci_davinci_proto_depIdxs = nil
}