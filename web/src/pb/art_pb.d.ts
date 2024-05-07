import * as jspb from 'google-protobuf'

import * as google_protobuf_timestamp_pb from 'google-protobuf/google/protobuf/timestamp_pb'; // proto import: "google/protobuf/timestamp.proto"
import * as google_protobuf_field_mask_pb from 'google-protobuf/google/protobuf/field_mask_pb'; // proto import: "google/protobuf/field_mask.proto"
import * as google_api_field_behavior_pb from './google/api/field_behavior_pb'; // proto import: "google/api/field_behavior.proto"
import * as google_api_resource_pb from './google/api/resource_pb'; // proto import: "google/api/resource.proto"


export class Art extends jspb.Message {
  getName(): string;
  setName(value: string): Art;

  getTitle(): string;
  setTitle(value: string): Art;

  getImageUrl(): string;
  setImageUrl(value: string): Art;

  getAuthor(): string;
  setAuthor(value: string): Art;

  getCreateTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setCreateTime(value?: google_protobuf_timestamp_pb.Timestamp): Art;
  hasCreateTime(): boolean;
  clearCreateTime(): Art;

  getUpdateTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setUpdateTime(value?: google_protobuf_timestamp_pb.Timestamp): Art;
  hasUpdateTime(): boolean;
  clearUpdateTime(): Art;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Art.AsObject;
  static toObject(includeInstance: boolean, msg: Art): Art.AsObject;
  static serializeBinaryToWriter(message: Art, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Art;
  static deserializeBinaryFromReader(message: Art, reader: jspb.BinaryReader): Art;
}

export namespace Art {
  export type AsObject = {
    name: string,
    title: string,
    imageUrl: string,
    author: string,
    createTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    updateTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
  }
}

export class CreateArtRequest extends jspb.Message {
  getParent(): string;
  setParent(value: string): CreateArtRequest;

  getArt(): Art | undefined;
  setArt(value?: Art): CreateArtRequest;
  hasArt(): boolean;
  clearArt(): CreateArtRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateArtRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CreateArtRequest): CreateArtRequest.AsObject;
  static serializeBinaryToWriter(message: CreateArtRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateArtRequest;
  static deserializeBinaryFromReader(message: CreateArtRequest, reader: jspb.BinaryReader): CreateArtRequest;
}

export namespace CreateArtRequest {
  export type AsObject = {
    parent: string,
    art?: Art.AsObject,
  }
}

export class UpdateArtRequest extends jspb.Message {
  getArt(): Art | undefined;
  setArt(value?: Art): UpdateArtRequest;
  hasArt(): boolean;
  clearArt(): UpdateArtRequest;

  getUpdateMask(): google_protobuf_field_mask_pb.FieldMask | undefined;
  setUpdateMask(value?: google_protobuf_field_mask_pb.FieldMask): UpdateArtRequest;
  hasUpdateMask(): boolean;
  clearUpdateMask(): UpdateArtRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UpdateArtRequest.AsObject;
  static toObject(includeInstance: boolean, msg: UpdateArtRequest): UpdateArtRequest.AsObject;
  static serializeBinaryToWriter(message: UpdateArtRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UpdateArtRequest;
  static deserializeBinaryFromReader(message: UpdateArtRequest, reader: jspb.BinaryReader): UpdateArtRequest;
}

export namespace UpdateArtRequest {
  export type AsObject = {
    art?: Art.AsObject,
    updateMask?: google_protobuf_field_mask_pb.FieldMask.AsObject,
  }
}

export class GetArtRequest extends jspb.Message {
  getName(): string;
  setName(value: string): GetArtRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetArtRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetArtRequest): GetArtRequest.AsObject;
  static serializeBinaryToWriter(message: GetArtRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetArtRequest;
  static deserializeBinaryFromReader(message: GetArtRequest, reader: jspb.BinaryReader): GetArtRequest;
}

export namespace GetArtRequest {
  export type AsObject = {
    name: string,
  }
}

export class ListArtsRequest extends jspb.Message {
  getParent(): string;
  setParent(value: string): ListArtsRequest;

  getPageSize(): number;
  setPageSize(value: number): ListArtsRequest;

  getPageToken(): number;
  setPageToken(value: number): ListArtsRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ListArtsRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ListArtsRequest): ListArtsRequest.AsObject;
  static serializeBinaryToWriter(message: ListArtsRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ListArtsRequest;
  static deserializeBinaryFromReader(message: ListArtsRequest, reader: jspb.BinaryReader): ListArtsRequest;
}

export namespace ListArtsRequest {
  export type AsObject = {
    parent: string,
    pageSize: number,
    pageToken: number,
  }
}

export class ListArtsResponse extends jspb.Message {
  getArtsList(): Array<Art>;
  setArtsList(value: Array<Art>): ListArtsResponse;
  clearArtsList(): ListArtsResponse;
  addArts(value?: Art, index?: number): Art;

  getNextPageToken(): number;
  setNextPageToken(value: number): ListArtsResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ListArtsResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ListArtsResponse): ListArtsResponse.AsObject;
  static serializeBinaryToWriter(message: ListArtsResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ListArtsResponse;
  static deserializeBinaryFromReader(message: ListArtsResponse, reader: jspb.BinaryReader): ListArtsResponse;
}

export namespace ListArtsResponse {
  export type AsObject = {
    artsList: Array<Art.AsObject>,
    nextPageToken: number,
  }
}

export class DeleteArtRequest extends jspb.Message {
  getName(): string;
  setName(value: string): DeleteArtRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DeleteArtRequest.AsObject;
  static toObject(includeInstance: boolean, msg: DeleteArtRequest): DeleteArtRequest.AsObject;
  static serializeBinaryToWriter(message: DeleteArtRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DeleteArtRequest;
  static deserializeBinaryFromReader(message: DeleteArtRequest, reader: jspb.BinaryReader): DeleteArtRequest;
}

export namespace DeleteArtRequest {
  export type AsObject = {
    name: string,
  }
}

export class UploadArtRequest extends jspb.Message {
  getName(): string;
  setName(value: string): UploadArtRequest;

  getData(): Uint8Array | string;
  getData_asU8(): Uint8Array;
  getData_asB64(): string;
  setData(value: Uint8Array | string): UploadArtRequest;

  getMimetype(): string;
  setMimetype(value: string): UploadArtRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UploadArtRequest.AsObject;
  static toObject(includeInstance: boolean, msg: UploadArtRequest): UploadArtRequest.AsObject;
  static serializeBinaryToWriter(message: UploadArtRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UploadArtRequest;
  static deserializeBinaryFromReader(message: UploadArtRequest, reader: jspb.BinaryReader): UploadArtRequest;
}

export namespace UploadArtRequest {
  export type AsObject = {
    name: string,
    data: Uint8Array | string,
    mimetype: string,
  }
}

