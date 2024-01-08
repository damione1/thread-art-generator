import * as jspb from 'google-protobuf'

import * as google_protobuf_timestamp_pb from 'google-protobuf/google/protobuf/timestamp_pb';
import * as google_api_field_behavior_pb from './google/api/field_behavior_pb';


export class Art extends jspb.Message {
  getId(): string;
  setId(value: string): Art;

  getTitle(): string;
  setTitle(value: string): Art;

  getImageUrl(): string;
  setImageUrl(value: string): Art;

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
    id: string,
    title: string,
    imageUrl: string,
    createTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    updateTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
  }
}

export class CreateArtRequest extends jspb.Message {
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
    art?: Art.AsObject,
  }
}

export class CreateArtResponse extends jspb.Message {
  getArt(): Art | undefined;
  setArt(value?: Art): CreateArtResponse;
  hasArt(): boolean;
  clearArt(): CreateArtResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateArtResponse.AsObject;
  static toObject(includeInstance: boolean, msg: CreateArtResponse): CreateArtResponse.AsObject;
  static serializeBinaryToWriter(message: CreateArtResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateArtResponse;
  static deserializeBinaryFromReader(message: CreateArtResponse, reader: jspb.BinaryReader): CreateArtResponse;
}

export namespace CreateArtResponse {
  export type AsObject = {
    art?: Art.AsObject,
  }
}

export class UpdateArtRequest extends jspb.Message {
  getArt(): Art | undefined;
  setArt(value?: Art): UpdateArtRequest;
  hasArt(): boolean;
  clearArt(): UpdateArtRequest;

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
  }
}

export class UpdateArtResponse extends jspb.Message {
  getArt(): Art | undefined;
  setArt(value?: Art): UpdateArtResponse;
  hasArt(): boolean;
  clearArt(): UpdateArtResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UpdateArtResponse.AsObject;
  static toObject(includeInstance: boolean, msg: UpdateArtResponse): UpdateArtResponse.AsObject;
  static serializeBinaryToWriter(message: UpdateArtResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UpdateArtResponse;
  static deserializeBinaryFromReader(message: UpdateArtResponse, reader: jspb.BinaryReader): UpdateArtResponse;
}

export namespace UpdateArtResponse {
  export type AsObject = {
    art?: Art.AsObject,
  }
}

export class GetArtRequest extends jspb.Message {
  getId(): string;
  setId(value: string): GetArtRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetArtRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetArtRequest): GetArtRequest.AsObject;
  static serializeBinaryToWriter(message: GetArtRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetArtRequest;
  static deserializeBinaryFromReader(message: GetArtRequest, reader: jspb.BinaryReader): GetArtRequest;
}

export namespace GetArtRequest {
  export type AsObject = {
    id: string,
  }
}

export class GetArtResponse extends jspb.Message {
  getArt(): Art | undefined;
  setArt(value?: Art): GetArtResponse;
  hasArt(): boolean;
  clearArt(): GetArtResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetArtResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetArtResponse): GetArtResponse.AsObject;
  static serializeBinaryToWriter(message: GetArtResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetArtResponse;
  static deserializeBinaryFromReader(message: GetArtResponse, reader: jspb.BinaryReader): GetArtResponse;
}

export namespace GetArtResponse {
  export type AsObject = {
    art?: Art.AsObject,
  }
}

export class ListArtRequest extends jspb.Message {
  getPageToken(): number;
  setPageToken(value: number): ListArtRequest;

  getPageSize(): number;
  setPageSize(value: number): ListArtRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ListArtRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ListArtRequest): ListArtRequest.AsObject;
  static serializeBinaryToWriter(message: ListArtRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ListArtRequest;
  static deserializeBinaryFromReader(message: ListArtRequest, reader: jspb.BinaryReader): ListArtRequest;
}

export namespace ListArtRequest {
  export type AsObject = {
    pageToken: number,
    pageSize: number,
  }
}

export class ListArtResponse extends jspb.Message {
  getArtsList(): Array<Art>;
  setArtsList(value: Array<Art>): ListArtResponse;
  clearArtsList(): ListArtResponse;
  addArts(value?: Art, index?: number): Art;

  getNextPageToken(): number;
  setNextPageToken(value: number): ListArtResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ListArtResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ListArtResponse): ListArtResponse.AsObject;
  static serializeBinaryToWriter(message: ListArtResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ListArtResponse;
  static deserializeBinaryFromReader(message: ListArtResponse, reader: jspb.BinaryReader): ListArtResponse;
}

export namespace ListArtResponse {
  export type AsObject = {
    artsList: Array<Art.AsObject>,
    nextPageToken: number,
  }
}

export class DeleteArtRequest extends jspb.Message {
  getId(): string;
  setId(value: string): DeleteArtRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DeleteArtRequest.AsObject;
  static toObject(includeInstance: boolean, msg: DeleteArtRequest): DeleteArtRequest.AsObject;
  static serializeBinaryToWriter(message: DeleteArtRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DeleteArtRequest;
  static deserializeBinaryFromReader(message: DeleteArtRequest, reader: jspb.BinaryReader): DeleteArtRequest;
}

export namespace DeleteArtRequest {
  export type AsObject = {
    id: string,
  }
}

export class DeleteArtResponse extends jspb.Message {
  getArt(): Art | undefined;
  setArt(value?: Art): DeleteArtResponse;
  hasArt(): boolean;
  clearArt(): DeleteArtResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DeleteArtResponse.AsObject;
  static toObject(includeInstance: boolean, msg: DeleteArtResponse): DeleteArtResponse.AsObject;
  static serializeBinaryToWriter(message: DeleteArtResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DeleteArtResponse;
  static deserializeBinaryFromReader(message: DeleteArtResponse, reader: jspb.BinaryReader): DeleteArtResponse;
}

export namespace DeleteArtResponse {
  export type AsObject = {
    art?: Art.AsObject,
  }
}

