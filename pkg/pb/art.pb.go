// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v3.21.12
// source: art.proto

package pb

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
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

type Art struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ID is the unique identifier for the art.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Title is the art's title.
	Title string `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	// ImageURL is the art's image URL.
	ImageUrl string `protobuf:"bytes,3,opt,name=image_url,json=imageUrl,proto3" json:"image_url,omitempty"`
	// Author is the art's author.
	AuthorId string `protobuf:"bytes,4,opt,name=author_id,json=authorId,proto3" json:"author_id,omitempty"`
	// Author is the art's author.
	Author *User `protobuf:"bytes,5,opt,name=author,proto3" json:"author,omitempty"`
	// CreatedAt is the art's creation time. Output only.
	CreateTime *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty"`
	// UpdatedAt is the art's last update time.
	UpdateTime *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=update_time,json=updateTime,proto3" json:"update_time,omitempty"`
}

func (x *Art) Reset() {
	*x = Art{}
	if protoimpl.UnsafeEnabled {
		mi := &file_art_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Art) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Art) ProtoMessage() {}

func (x *Art) ProtoReflect() protoreflect.Message {
	mi := &file_art_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Art.ProtoReflect.Descriptor instead.
func (*Art) Descriptor() ([]byte, []int) {
	return file_art_proto_rawDescGZIP(), []int{0}
}

func (x *Art) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Art) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Art) GetImageUrl() string {
	if x != nil {
		return x.ImageUrl
	}
	return ""
}

func (x *Art) GetAuthorId() string {
	if x != nil {
		return x.AuthorId
	}
	return ""
}

func (x *Art) GetAuthor() *User {
	if x != nil {
		return x.Author
	}
	return nil
}

func (x *Art) GetCreateTime() *timestamppb.Timestamp {
	if x != nil {
		return x.CreateTime
	}
	return nil
}

func (x *Art) GetUpdateTime() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdateTime
	}
	return nil
}

type CreateArtRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Art is the art.
	Art *Art `protobuf:"bytes,1,opt,name=art,proto3" json:"art,omitempty"`
}

func (x *CreateArtRequest) Reset() {
	*x = CreateArtRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_art_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateArtRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateArtRequest) ProtoMessage() {}

func (x *CreateArtRequest) ProtoReflect() protoreflect.Message {
	mi := &file_art_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateArtRequest.ProtoReflect.Descriptor instead.
func (*CreateArtRequest) Descriptor() ([]byte, []int) {
	return file_art_proto_rawDescGZIP(), []int{1}
}

func (x *CreateArtRequest) GetArt() *Art {
	if x != nil {
		return x.Art
	}
	return nil
}

type CreateArtResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Art is the art.
	Art *Art `protobuf:"bytes,1,opt,name=art,proto3" json:"art,omitempty"`
}

func (x *CreateArtResponse) Reset() {
	*x = CreateArtResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_art_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateArtResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateArtResponse) ProtoMessage() {}

func (x *CreateArtResponse) ProtoReflect() protoreflect.Message {
	mi := &file_art_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateArtResponse.ProtoReflect.Descriptor instead.
func (*CreateArtResponse) Descriptor() ([]byte, []int) {
	return file_art_proto_rawDescGZIP(), []int{2}
}

func (x *CreateArtResponse) GetArt() *Art {
	if x != nil {
		return x.Art
	}
	return nil
}

type UpdateArtRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Art is the art.
	Art *Art `protobuf:"bytes,1,opt,name=art,proto3" json:"art,omitempty"`
}

func (x *UpdateArtRequest) Reset() {
	*x = UpdateArtRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_art_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateArtRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateArtRequest) ProtoMessage() {}

func (x *UpdateArtRequest) ProtoReflect() protoreflect.Message {
	mi := &file_art_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateArtRequest.ProtoReflect.Descriptor instead.
func (*UpdateArtRequest) Descriptor() ([]byte, []int) {
	return file_art_proto_rawDescGZIP(), []int{3}
}

func (x *UpdateArtRequest) GetArt() *Art {
	if x != nil {
		return x.Art
	}
	return nil
}

type UpdateArtResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Art is the art.
	Art *Art `protobuf:"bytes,1,opt,name=art,proto3" json:"art,omitempty"`
}

func (x *UpdateArtResponse) Reset() {
	*x = UpdateArtResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_art_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateArtResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateArtResponse) ProtoMessage() {}

func (x *UpdateArtResponse) ProtoReflect() protoreflect.Message {
	mi := &file_art_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateArtResponse.ProtoReflect.Descriptor instead.
func (*UpdateArtResponse) Descriptor() ([]byte, []int) {
	return file_art_proto_rawDescGZIP(), []int{4}
}

func (x *UpdateArtResponse) GetArt() *Art {
	if x != nil {
		return x.Art
	}
	return nil
}

type GetArtRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ID is the unique identifier for the art.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetArtRequest) Reset() {
	*x = GetArtRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_art_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetArtRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetArtRequest) ProtoMessage() {}

func (x *GetArtRequest) ProtoReflect() protoreflect.Message {
	mi := &file_art_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetArtRequest.ProtoReflect.Descriptor instead.
func (*GetArtRequest) Descriptor() ([]byte, []int) {
	return file_art_proto_rawDescGZIP(), []int{5}
}

func (x *GetArtRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GetArtResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Art is the art.
	Art *Art `protobuf:"bytes,1,opt,name=art,proto3" json:"art,omitempty"`
}

func (x *GetArtResponse) Reset() {
	*x = GetArtResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_art_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetArtResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetArtResponse) ProtoMessage() {}

func (x *GetArtResponse) ProtoReflect() protoreflect.Message {
	mi := &file_art_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetArtResponse.ProtoReflect.Descriptor instead.
func (*GetArtResponse) Descriptor() ([]byte, []int) {
	return file_art_proto_rawDescGZIP(), []int{6}
}

func (x *GetArtResponse) GetArt() *Art {
	if x != nil {
		return x.Art
	}
	return nil
}

type ListArtRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// PageToken is the page token.
	PageToken int32 `protobuf:"varint,1,opt,name=page_token,json=pageToken,proto3" json:"page_token,omitempty"`
	// PageSize is the page size.
	PageSize int32 `protobuf:"varint,2,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
}

func (x *ListArtRequest) Reset() {
	*x = ListArtRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_art_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListArtRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListArtRequest) ProtoMessage() {}

func (x *ListArtRequest) ProtoReflect() protoreflect.Message {
	mi := &file_art_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListArtRequest.ProtoReflect.Descriptor instead.
func (*ListArtRequest) Descriptor() ([]byte, []int) {
	return file_art_proto_rawDescGZIP(), []int{7}
}

func (x *ListArtRequest) GetPageToken() int32 {
	if x != nil {
		return x.PageToken
	}
	return 0
}

func (x *ListArtRequest) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

type ListArtResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Arts is the list of arts.
	Arts []*Art `protobuf:"bytes,1,rep,name=arts,proto3" json:"arts,omitempty"`
	// NextPageToken is the next page token.
	NextPageToken int32 `protobuf:"varint,2,opt,name=next_page_token,json=nextPageToken,proto3" json:"next_page_token,omitempty"`
}

func (x *ListArtResponse) Reset() {
	*x = ListArtResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_art_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListArtResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListArtResponse) ProtoMessage() {}

func (x *ListArtResponse) ProtoReflect() protoreflect.Message {
	mi := &file_art_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListArtResponse.ProtoReflect.Descriptor instead.
func (*ListArtResponse) Descriptor() ([]byte, []int) {
	return file_art_proto_rawDescGZIP(), []int{8}
}

func (x *ListArtResponse) GetArts() []*Art {
	if x != nil {
		return x.Arts
	}
	return nil
}

func (x *ListArtResponse) GetNextPageToken() int32 {
	if x != nil {
		return x.NextPageToken
	}
	return 0
}

type DeleteArtRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ID is the unique identifier for the art.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *DeleteArtRequest) Reset() {
	*x = DeleteArtRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_art_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteArtRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteArtRequest) ProtoMessage() {}

func (x *DeleteArtRequest) ProtoReflect() protoreflect.Message {
	mi := &file_art_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteArtRequest.ProtoReflect.Descriptor instead.
func (*DeleteArtRequest) Descriptor() ([]byte, []int) {
	return file_art_proto_rawDescGZIP(), []int{9}
}

func (x *DeleteArtRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type DeleteArtResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Art is the art.
	Art *Art `protobuf:"bytes,1,opt,name=art,proto3" json:"art,omitempty"`
}

func (x *DeleteArtResponse) Reset() {
	*x = DeleteArtResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_art_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteArtResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteArtResponse) ProtoMessage() {}

func (x *DeleteArtResponse) ProtoReflect() protoreflect.Message {
	mi := &file_art_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteArtResponse.ProtoReflect.Descriptor instead.
func (*DeleteArtResponse) Descriptor() ([]byte, []int) {
	return file_art_proto_rawDescGZIP(), []int{10}
}

func (x *DeleteArtResponse) GetArt() *Art {
	if x != nil {
		return x.Art
	}
	return nil
}

var File_art_proto protoreflect.FileDescriptor

var file_art_proto_rawDesc = []byte{
	0x0a, 0x09, 0x61, 0x72, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70, 0x62, 0x1a,
	0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x65,
	0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x0a, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa2, 0x02,
	0x0a, 0x03, 0x41, 0x72, 0x74, 0x12, 0x13, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x19, 0x0a, 0x05, 0x74, 0x69,
	0x74, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x05,
	0x74, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x5f, 0x75,
	0x72, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x55,
	0x72, 0x6c, 0x12, 0x23, 0x0a, 0x09, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x5f, 0x69, 0x64, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x09, 0x42, 0x06, 0xe0, 0x41, 0x02, 0xe0, 0x41, 0x04, 0x52, 0x08, 0x61,
	0x75, 0x74, 0x68, 0x6f, 0x72, 0x49, 0x64, 0x12, 0x25, 0x0a, 0x06, 0x61, 0x75, 0x74, 0x68, 0x6f,
	0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x08, 0x2e, 0x70, 0x62, 0x2e, 0x55, 0x73, 0x65,
	0x72, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x06, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x12, 0x40,
	0x0a, 0x0b, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42,
	0x03, 0xe0, 0x41, 0x03, 0x52, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65,
	0x12, 0x40, 0x0a, 0x0b, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x69,
	0x6d, 0x65, 0x22, 0x32, 0x0a, 0x10, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x72, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1e, 0x0a, 0x03, 0x61, 0x72, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x07, 0x2e, 0x70, 0x62, 0x2e, 0x41, 0x72, 0x74, 0x42, 0x03, 0xe0, 0x41,
	0x02, 0x52, 0x03, 0x61, 0x72, 0x74, 0x22, 0x2e, 0x0a, 0x11, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x41, 0x72, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x19, 0x0a, 0x03, 0x61,
	0x72, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x07, 0x2e, 0x70, 0x62, 0x2e, 0x41, 0x72,
	0x74, 0x52, 0x03, 0x61, 0x72, 0x74, 0x22, 0x32, 0x0a, 0x10, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x41, 0x72, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1e, 0x0a, 0x03, 0x61, 0x72,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x07, 0x2e, 0x70, 0x62, 0x2e, 0x41, 0x72, 0x74,
	0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x03, 0x61, 0x72, 0x74, 0x22, 0x2e, 0x0a, 0x11, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x41, 0x72, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x19, 0x0a, 0x03, 0x61, 0x72, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x07, 0x2e, 0x70,
	0x62, 0x2e, 0x41, 0x72, 0x74, 0x52, 0x03, 0x61, 0x72, 0x74, 0x22, 0x1f, 0x0a, 0x0d, 0x47, 0x65,
	0x74, 0x41, 0x72, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x2b, 0x0a, 0x0e, 0x47,
	0x65, 0x74, 0x41, 0x72, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x19, 0x0a,
	0x03, 0x61, 0x72, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x07, 0x2e, 0x70, 0x62, 0x2e,
	0x41, 0x72, 0x74, 0x52, 0x03, 0x61, 0x72, 0x74, 0x22, 0x4c, 0x0a, 0x0e, 0x4c, 0x69, 0x73, 0x74,
	0x41, 0x72, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x61,
	0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09,
	0x70, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x61, 0x67,
	0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x61,
	0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x22, 0x56, 0x0a, 0x0f, 0x4c, 0x69, 0x73, 0x74, 0x41, 0x72,
	0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1b, 0x0a, 0x04, 0x61, 0x72, 0x74,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x07, 0x2e, 0x70, 0x62, 0x2e, 0x41, 0x72, 0x74,
	0x52, 0x04, 0x61, 0x72, 0x74, 0x73, 0x12, 0x26, 0x0a, 0x0f, 0x6e, 0x65, 0x78, 0x74, 0x5f, 0x70,
	0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x0d, 0x6e, 0x65, 0x78, 0x74, 0x50, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x22,
	0x0a, 0x10, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x41, 0x72, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x22, 0x2e, 0x0a, 0x11, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x41, 0x72, 0x74, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x19, 0x0a, 0x03, 0x61, 0x72, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x07, 0x2e, 0x70, 0x62, 0x2e, 0x41, 0x72, 0x74, 0x52, 0x03, 0x61,
	0x72, 0x74, 0x42, 0x23, 0x5a, 0x21, 0x74, 0x68, 0x72, 0x65, 0x61, 0x64, 0x2d, 0x61, 0x72, 0x74,
	0x2d, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x3b, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_art_proto_rawDescOnce sync.Once
	file_art_proto_rawDescData = file_art_proto_rawDesc
)

func file_art_proto_rawDescGZIP() []byte {
	file_art_proto_rawDescOnce.Do(func() {
		file_art_proto_rawDescData = protoimpl.X.CompressGZIP(file_art_proto_rawDescData)
	})
	return file_art_proto_rawDescData
}

var file_art_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_art_proto_goTypes = []interface{}{
	(*Art)(nil),                   // 0: pb.Art
	(*CreateArtRequest)(nil),      // 1: pb.CreateArtRequest
	(*CreateArtResponse)(nil),     // 2: pb.CreateArtResponse
	(*UpdateArtRequest)(nil),      // 3: pb.UpdateArtRequest
	(*UpdateArtResponse)(nil),     // 4: pb.UpdateArtResponse
	(*GetArtRequest)(nil),         // 5: pb.GetArtRequest
	(*GetArtResponse)(nil),        // 6: pb.GetArtResponse
	(*ListArtRequest)(nil),        // 7: pb.ListArtRequest
	(*ListArtResponse)(nil),       // 8: pb.ListArtResponse
	(*DeleteArtRequest)(nil),      // 9: pb.DeleteArtRequest
	(*DeleteArtResponse)(nil),     // 10: pb.DeleteArtResponse
	(*User)(nil),                  // 11: pb.User
	(*timestamppb.Timestamp)(nil), // 12: google.protobuf.Timestamp
}
var file_art_proto_depIdxs = []int32{
	11, // 0: pb.Art.author:type_name -> pb.User
	12, // 1: pb.Art.create_time:type_name -> google.protobuf.Timestamp
	12, // 2: pb.Art.update_time:type_name -> google.protobuf.Timestamp
	0,  // 3: pb.CreateArtRequest.art:type_name -> pb.Art
	0,  // 4: pb.CreateArtResponse.art:type_name -> pb.Art
	0,  // 5: pb.UpdateArtRequest.art:type_name -> pb.Art
	0,  // 6: pb.UpdateArtResponse.art:type_name -> pb.Art
	0,  // 7: pb.GetArtResponse.art:type_name -> pb.Art
	0,  // 8: pb.ListArtResponse.arts:type_name -> pb.Art
	0,  // 9: pb.DeleteArtResponse.art:type_name -> pb.Art
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_art_proto_init() }
func file_art_proto_init() {
	if File_art_proto != nil {
		return
	}
	file_user_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_art_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Art); i {
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
		file_art_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateArtRequest); i {
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
		file_art_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateArtResponse); i {
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
		file_art_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateArtRequest); i {
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
		file_art_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateArtResponse); i {
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
		file_art_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetArtRequest); i {
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
		file_art_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetArtResponse); i {
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
		file_art_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListArtRequest); i {
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
		file_art_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListArtResponse); i {
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
		file_art_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteArtRequest); i {
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
		file_art_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteArtResponse); i {
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
			RawDescriptor: file_art_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_art_proto_goTypes,
		DependencyIndexes: file_art_proto_depIdxs,
		MessageInfos:      file_art_proto_msgTypes,
	}.Build()
	File_art_proto = out.File
	file_art_proto_rawDesc = nil
	file_art_proto_goTypes = nil
	file_art_proto_depIdxs = nil
}
