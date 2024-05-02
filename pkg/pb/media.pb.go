// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v3.21.12
// source: media.proto

package pb

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/known/fieldmaskpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Media struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of the Media resource.
	// For example: "medias/123"
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Bucket is the media's bucket.
	Bucket string `protobuf:"bytes,2,opt,name=bucket,proto3" json:"bucket,omitempty"`
	// ImageURL is the media's image URL.
	ImageUrl string `protobuf:"bytes,3,opt,name=image_url,json=imageUrl,proto3" json:"image_url,omitempty"`
	// The name of the User resource who is the author of the media.
	// For example: "users/456"
	Author string `protobuf:"bytes,5,opt,name=author,proto3" json:"author,omitempty"`
	// CreatedAt is the media's creation time. Output only.
	CreateTime *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty"`
	// UpdatedAt is the media's last update time.
	UpdateTime *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=update_time,json=updateTime,proto3" json:"update_time,omitempty"`
}

func (x *Media) Reset() {
	*x = Media{}
	if protoimpl.UnsafeEnabled {
		mi := &file_media_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Media) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Media) ProtoMessage() {}

func (x *Media) ProtoReflect() protoreflect.Message {
	mi := &file_media_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Media.ProtoReflect.Descriptor instead.
func (*Media) Descriptor() ([]byte, []int) {
	return file_media_proto_rawDescGZIP(), []int{0}
}

func (x *Media) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Media) GetBucket() string {
	if x != nil {
		return x.Bucket
	}
	return ""
}

func (x *Media) GetImageUrl() string {
	if x != nil {
		return x.ImageUrl
	}
	return ""
}

func (x *Media) GetAuthor() string {
	if x != nil {
		return x.Author
	}
	return ""
}

func (x *Media) GetCreateTime() *timestamppb.Timestamp {
	if x != nil {
		return x.CreateTime
	}
	return nil
}

func (x *Media) GetUpdateTime() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdateTime
	}
	return nil
}

// MediaChunk represents a chunk of a media file.
type MediaChunk struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The bytes of the chunk.
	Chunk []byte `protobuf:"bytes,1,opt,name=chunk,proto3" json:"chunk,omitempty"`
	// The name of the media.
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *MediaChunk) Reset() {
	*x = MediaChunk{}
	if protoimpl.UnsafeEnabled {
		mi := &file_media_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MediaChunk) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MediaChunk) ProtoMessage() {}

func (x *MediaChunk) ProtoReflect() protoreflect.Message {
	mi := &file_media_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MediaChunk.ProtoReflect.Descriptor instead.
func (*MediaChunk) Descriptor() ([]byte, []int) {
	return file_media_proto_rawDescGZIP(), []int{1}
}

func (x *MediaChunk) GetChunk() []byte {
	if x != nil {
		return x.Chunk
	}
	return nil
}

func (x *MediaChunk) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type UploadMediaRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The media to upload.
	MediaChunk *MediaChunk `protobuf:"bytes,1,opt,name=media_chunk,json=mediaChunk,proto3" json:"media_chunk,omitempty"`
}

func (x *UploadMediaRequest) Reset() {
	*x = UploadMediaRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_media_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadMediaRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadMediaRequest) ProtoMessage() {}

func (x *UploadMediaRequest) ProtoReflect() protoreflect.Message {
	mi := &file_media_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadMediaRequest.ProtoReflect.Descriptor instead.
func (*UploadMediaRequest) Descriptor() ([]byte, []int) {
	return file_media_proto_rawDescGZIP(), []int{2}
}

func (x *UploadMediaRequest) GetMediaChunk() *MediaChunk {
	if x != nil {
		return x.MediaChunk
	}
	return nil
}

type GetMediaRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of the Media resource.
	// For example: "users/{user}/medias/{media}"
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *GetMediaRequest) Reset() {
	*x = GetMediaRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_media_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetMediaRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetMediaRequest) ProtoMessage() {}

func (x *GetMediaRequest) ProtoReflect() protoreflect.Message {
	mi := &file_media_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetMediaRequest.ProtoReflect.Descriptor instead.
func (*GetMediaRequest) Descriptor() ([]byte, []int) {
	return file_media_proto_rawDescGZIP(), []int{3}
}

func (x *GetMediaRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type DeleteMediaRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of the Media resource.
	// For example: "medias/123"
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *DeleteMediaRequest) Reset() {
	*x = DeleteMediaRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_media_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteMediaRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteMediaRequest) ProtoMessage() {}

func (x *DeleteMediaRequest) ProtoReflect() protoreflect.Message {
	mi := &file_media_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteMediaRequest.ProtoReflect.Descriptor instead.
func (*DeleteMediaRequest) Descriptor() ([]byte, []int) {
	return file_media_proto_rawDescGZIP(), []int{4}
}

func (x *DeleteMediaRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

var File_media_proto protoreflect.FileDescriptor

var file_media_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70,
	0x62, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x6d, 0x61, 0x73, 0x6b, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x72, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xa9, 0x02, 0x0a, 0x05, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x12, 0x31, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x1d, 0xe0, 0x41, 0x03, 0xfa, 0x41, 0x17,
	0x0a, 0x15, 0x61, 0x72, 0x74, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a,
	0x06, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x62,
	0x75, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x5f, 0x75,
	0x72, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x55,
	0x72, 0x6c, 0x12, 0x34, 0x0a, 0x06, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x1c, 0xe0, 0x41, 0x03, 0xfa, 0x41, 0x16, 0x0a, 0x14, 0x61, 0x72, 0x74, 0x2e,
	0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x55, 0x73, 0x65, 0x72,
	0x52, 0x06, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x12, 0x40, 0x0a, 0x0b, 0x63, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x0a,
	0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x40, 0x0a, 0x0b, 0x75, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x03, 0xe0, 0x41, 0x03,
	0x52, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x22, 0x36, 0x0a, 0x0a,
	0x4d, 0x65, 0x64, 0x69, 0x61, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x68,
	0x75, 0x6e, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x63, 0x68, 0x75, 0x6e, 0x6b,
	0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x22, 0x45, 0x0a, 0x12, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x65,
	0x64, 0x69, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2f, 0x0a, 0x0b, 0x6d, 0x65,
	0x64, 0x69, 0x61, 0x5f, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0e, 0x2e, 0x70, 0x62, 0x2e, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x52,
	0x0a, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x22, 0x44, 0x0a, 0x0f, 0x47,
	0x65, 0x74, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x31,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x1d, 0xe0, 0x41,
	0x02, 0xfa, 0x41, 0x17, 0x0a, 0x15, 0x61, 0x72, 0x74, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c,
	0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x22, 0x3a, 0x0a, 0x12, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x4d, 0x65, 0x64, 0x69, 0x61,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x24, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x10, 0xe0, 0x41, 0x02, 0xfa, 0x41, 0x0a, 0x0a, 0x08, 0x70,
	0x62, 0x2e, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x42, 0x23, 0x5a,
	0x21, 0x74, 0x68, 0x72, 0x65, 0x61, 0x64, 0x2d, 0x61, 0x72, 0x74, 0x2d, 0x67, 0x65, 0x6e, 0x65,
	0x72, 0x61, 0x74, 0x6f, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x3b,
	0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_media_proto_rawDescOnce sync.Once
	file_media_proto_rawDescData = file_media_proto_rawDesc
)

func file_media_proto_rawDescGZIP() []byte {
	file_media_proto_rawDescOnce.Do(func() {
		file_media_proto_rawDescData = protoimpl.X.CompressGZIP(file_media_proto_rawDescData)
	})
	return file_media_proto_rawDescData
}

var file_media_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_media_proto_goTypes = []interface{}{
	(*Media)(nil),                 // 0: pb.Media
	(*MediaChunk)(nil),            // 1: pb.MediaChunk
	(*UploadMediaRequest)(nil),    // 2: pb.UploadMediaRequest
	(*GetMediaRequest)(nil),       // 3: pb.GetMediaRequest
	(*DeleteMediaRequest)(nil),    // 4: pb.DeleteMediaRequest
	(*timestamppb.Timestamp)(nil), // 5: google.protobuf.Timestamp
}
var file_media_proto_depIdxs = []int32{
	5, // 0: pb.Media.create_time:type_name -> google.protobuf.Timestamp
	5, // 1: pb.Media.update_time:type_name -> google.protobuf.Timestamp
	1, // 2: pb.UploadMediaRequest.media_chunk:type_name -> pb.MediaChunk
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_media_proto_init() }
func file_media_proto_init() {
	if File_media_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_media_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Media); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_media_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MediaChunk); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_media_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadMediaRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_media_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetMediaRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_media_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteMediaRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_media_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_media_proto_goTypes,
		DependencyIndexes: file_media_proto_depIdxs,
		MessageInfos:      file_media_proto_msgTypes,
	}.Build()
	File_media_proto = out.File
	file_media_proto_rawDesc = nil
	file_media_proto_goTypes = nil
	file_media_proto_depIdxs = nil
}
