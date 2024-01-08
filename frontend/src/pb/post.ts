/* eslint-disable */
import * as _m0 from "protobufjs/minimal";
import { Timestamp } from "./google/protobuf/timestamp";

export const protobufPackage = "pb";

export interface Art {
  /** ID is the unique identifier for the art. */
  id: string;
  /** Title is the art's title. */
  title: string;
  /** ImageURL is the art's image URL. */
  imageUrl: string;
  /** CreatedAt is the art's creation time. */
  createTime?:
    | Date
    | undefined;
  /** UpdatedAt is the art's last update time. */
  updateTime?: Date | undefined;
}

export interface CreateArtRequest {
  /** Art is the art. */
  art?: Art | undefined;
}

export interface CreateArtResponse {
  /** Art is the art. */
  art?: Art | undefined;
}

export interface UpdateArtRequest {
  /** Art is the art. */
  art?: Art | undefined;
}

export interface UpdateArtResponse {
  /** Art is the art. */
  art?: Art | undefined;
}

export interface GetArtRequest {
  /** ID is the unique identifier for the art. */
  id: string;
}

export interface GetArtResponse {
  /** Art is the art. */
  art?: Art | undefined;
}

export interface ListArtRequest {
  /** PageToken is the page token. */
  pageToken: number;
  /** PageSize is the page size. */
  pageSize: number;
}

export interface ListArtResponse {
  /** Arts is the list of arts. */
  arts: Art[];
  /** NextPageToken is the next page token. */
  nextPageToken: number;
}

export interface DeleteArtRequest {
  /** ID is the unique identifier for the art. */
  id: string;
}

export interface DeleteArtResponse {
  /** Art is the art. */
  art?: Art | undefined;
}

function createBaseArt(): Art {
  return { id: "", title: "", imageUrl: "", createTime: undefined, updateTime: undefined };
}

export const Art = {
  encode(message: Art, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.id !== "") {
      writer.uint32(10).string(message.id);
    }
    if (message.title !== "") {
      writer.uint32(18).string(message.title);
    }
    if (message.imageUrl !== "") {
      writer.uint32(26).string(message.imageUrl);
    }
    if (message.createTime !== undefined) {
      Timestamp.encode(toTimestamp(message.createTime), writer.uint32(50).fork()).ldelim();
    }
    if (message.updateTime !== undefined) {
      Timestamp.encode(toTimestamp(message.updateTime), writer.uint32(58).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Art {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseArt();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.id = reader.string();
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.title = reader.string();
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.imageUrl = reader.string();
          continue;
        case 6:
          if (tag !== 50) {
            break;
          }

          message.createTime = fromTimestamp(Timestamp.decode(reader, reader.uint32()));
          continue;
        case 7:
          if (tag !== 58) {
            break;
          }

          message.updateTime = fromTimestamp(Timestamp.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): Art {
    return {
      id: isSet(object.id) ? globalThis.String(object.id) : "",
      title: isSet(object.title) ? globalThis.String(object.title) : "",
      imageUrl: isSet(object.imageUrl) ? globalThis.String(object.imageUrl) : "",
      createTime: isSet(object.createTime) ? fromJsonTimestamp(object.createTime) : undefined,
      updateTime: isSet(object.updateTime) ? fromJsonTimestamp(object.updateTime) : undefined,
    };
  },

  toJSON(message: Art): unknown {
    const obj: any = {};
    if (message.id !== "") {
      obj.id = message.id;
    }
    if (message.title !== "") {
      obj.title = message.title;
    }
    if (message.imageUrl !== "") {
      obj.imageUrl = message.imageUrl;
    }
    if (message.createTime !== undefined) {
      obj.createTime = message.createTime.toISOString();
    }
    if (message.updateTime !== undefined) {
      obj.updateTime = message.updateTime.toISOString();
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<Art>, I>>(base?: I): Art {
    return Art.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<Art>, I>>(object: I): Art {
    const message = createBaseArt();
    message.id = object.id ?? "";
    message.title = object.title ?? "";
    message.imageUrl = object.imageUrl ?? "";
    message.createTime = object.createTime ?? undefined;
    message.updateTime = object.updateTime ?? undefined;
    return message;
  },
};

function createBaseCreateArtRequest(): CreateArtRequest {
  return { art: undefined };
}

export const CreateArtRequest = {
  encode(message: CreateArtRequest, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.art !== undefined) {
      Art.encode(message.art, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CreateArtRequest {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseCreateArtRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.art = Art.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): CreateArtRequest {
    return { art: isSet(object.art) ? Art.fromJSON(object.art) : undefined };
  },

  toJSON(message: CreateArtRequest): unknown {
    const obj: any = {};
    if (message.art !== undefined) {
      obj.art = Art.toJSON(message.art);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<CreateArtRequest>, I>>(base?: I): CreateArtRequest {
    return CreateArtRequest.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<CreateArtRequest>, I>>(object: I): CreateArtRequest {
    const message = createBaseCreateArtRequest();
    message.art = (object.art !== undefined && object.art !== null) ? Art.fromPartial(object.art) : undefined;
    return message;
  },
};

function createBaseCreateArtResponse(): CreateArtResponse {
  return { art: undefined };
}

export const CreateArtResponse = {
  encode(message: CreateArtResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.art !== undefined) {
      Art.encode(message.art, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CreateArtResponse {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseCreateArtResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.art = Art.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): CreateArtResponse {
    return { art: isSet(object.art) ? Art.fromJSON(object.art) : undefined };
  },

  toJSON(message: CreateArtResponse): unknown {
    const obj: any = {};
    if (message.art !== undefined) {
      obj.art = Art.toJSON(message.art);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<CreateArtResponse>, I>>(base?: I): CreateArtResponse {
    return CreateArtResponse.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<CreateArtResponse>, I>>(object: I): CreateArtResponse {
    const message = createBaseCreateArtResponse();
    message.art = (object.art !== undefined && object.art !== null) ? Art.fromPartial(object.art) : undefined;
    return message;
  },
};

function createBaseUpdateArtRequest(): UpdateArtRequest {
  return { art: undefined };
}

export const UpdateArtRequest = {
  encode(message: UpdateArtRequest, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.art !== undefined) {
      Art.encode(message.art, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): UpdateArtRequest {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseUpdateArtRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.art = Art.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): UpdateArtRequest {
    return { art: isSet(object.art) ? Art.fromJSON(object.art) : undefined };
  },

  toJSON(message: UpdateArtRequest): unknown {
    const obj: any = {};
    if (message.art !== undefined) {
      obj.art = Art.toJSON(message.art);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<UpdateArtRequest>, I>>(base?: I): UpdateArtRequest {
    return UpdateArtRequest.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<UpdateArtRequest>, I>>(object: I): UpdateArtRequest {
    const message = createBaseUpdateArtRequest();
    message.art = (object.art !== undefined && object.art !== null) ? Art.fromPartial(object.art) : undefined;
    return message;
  },
};

function createBaseUpdateArtResponse(): UpdateArtResponse {
  return { art: undefined };
}

export const UpdateArtResponse = {
  encode(message: UpdateArtResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.art !== undefined) {
      Art.encode(message.art, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): UpdateArtResponse {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseUpdateArtResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.art = Art.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): UpdateArtResponse {
    return { art: isSet(object.art) ? Art.fromJSON(object.art) : undefined };
  },

  toJSON(message: UpdateArtResponse): unknown {
    const obj: any = {};
    if (message.art !== undefined) {
      obj.art = Art.toJSON(message.art);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<UpdateArtResponse>, I>>(base?: I): UpdateArtResponse {
    return UpdateArtResponse.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<UpdateArtResponse>, I>>(object: I): UpdateArtResponse {
    const message = createBaseUpdateArtResponse();
    message.art = (object.art !== undefined && object.art !== null) ? Art.fromPartial(object.art) : undefined;
    return message;
  },
};

function createBaseGetArtRequest(): GetArtRequest {
  return { id: "" };
}

export const GetArtRequest = {
  encode(message: GetArtRequest, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.id !== "") {
      writer.uint32(10).string(message.id);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetArtRequest {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseGetArtRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.id = reader.string();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): GetArtRequest {
    return { id: isSet(object.id) ? globalThis.String(object.id) : "" };
  },

  toJSON(message: GetArtRequest): unknown {
    const obj: any = {};
    if (message.id !== "") {
      obj.id = message.id;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<GetArtRequest>, I>>(base?: I): GetArtRequest {
    return GetArtRequest.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<GetArtRequest>, I>>(object: I): GetArtRequest {
    const message = createBaseGetArtRequest();
    message.id = object.id ?? "";
    return message;
  },
};

function createBaseGetArtResponse(): GetArtResponse {
  return { art: undefined };
}

export const GetArtResponse = {
  encode(message: GetArtResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.art !== undefined) {
      Art.encode(message.art, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetArtResponse {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseGetArtResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.art = Art.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): GetArtResponse {
    return { art: isSet(object.art) ? Art.fromJSON(object.art) : undefined };
  },

  toJSON(message: GetArtResponse): unknown {
    const obj: any = {};
    if (message.art !== undefined) {
      obj.art = Art.toJSON(message.art);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<GetArtResponse>, I>>(base?: I): GetArtResponse {
    return GetArtResponse.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<GetArtResponse>, I>>(object: I): GetArtResponse {
    const message = createBaseGetArtResponse();
    message.art = (object.art !== undefined && object.art !== null) ? Art.fromPartial(object.art) : undefined;
    return message;
  },
};

function createBaseListArtRequest(): ListArtRequest {
  return { pageToken: 0, pageSize: 0 };
}

export const ListArtRequest = {
  encode(message: ListArtRequest, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.pageToken !== 0) {
      writer.uint32(8).int32(message.pageToken);
    }
    if (message.pageSize !== 0) {
      writer.uint32(16).int32(message.pageSize);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ListArtRequest {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseListArtRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break;
          }

          message.pageToken = reader.int32();
          continue;
        case 2:
          if (tag !== 16) {
            break;
          }

          message.pageSize = reader.int32();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ListArtRequest {
    return {
      pageToken: isSet(object.pageToken) ? globalThis.Number(object.pageToken) : 0,
      pageSize: isSet(object.pageSize) ? globalThis.Number(object.pageSize) : 0,
    };
  },

  toJSON(message: ListArtRequest): unknown {
    const obj: any = {};
    if (message.pageToken !== 0) {
      obj.pageToken = Math.round(message.pageToken);
    }
    if (message.pageSize !== 0) {
      obj.pageSize = Math.round(message.pageSize);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ListArtRequest>, I>>(base?: I): ListArtRequest {
    return ListArtRequest.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ListArtRequest>, I>>(object: I): ListArtRequest {
    const message = createBaseListArtRequest();
    message.pageToken = object.pageToken ?? 0;
    message.pageSize = object.pageSize ?? 0;
    return message;
  },
};

function createBaseListArtResponse(): ListArtResponse {
  return { arts: [], nextPageToken: 0 };
}

export const ListArtResponse = {
  encode(message: ListArtResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    for (const v of message.arts) {
      Art.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.nextPageToken !== 0) {
      writer.uint32(16).int32(message.nextPageToken);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ListArtResponse {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseListArtResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.arts.push(Art.decode(reader, reader.uint32()));
          continue;
        case 2:
          if (tag !== 16) {
            break;
          }

          message.nextPageToken = reader.int32();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ListArtResponse {
    return {
      arts: globalThis.Array.isArray(object?.arts) ? object.arts.map((e: any) => Art.fromJSON(e)) : [],
      nextPageToken: isSet(object.nextPageToken) ? globalThis.Number(object.nextPageToken) : 0,
    };
  },

  toJSON(message: ListArtResponse): unknown {
    const obj: any = {};
    if (message.arts?.length) {
      obj.arts = message.arts.map((e) => Art.toJSON(e));
    }
    if (message.nextPageToken !== 0) {
      obj.nextPageToken = Math.round(message.nextPageToken);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ListArtResponse>, I>>(base?: I): ListArtResponse {
    return ListArtResponse.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ListArtResponse>, I>>(object: I): ListArtResponse {
    const message = createBaseListArtResponse();
    message.arts = object.arts?.map((e) => Art.fromPartial(e)) || [];
    message.nextPageToken = object.nextPageToken ?? 0;
    return message;
  },
};

function createBaseDeleteArtRequest(): DeleteArtRequest {
  return { id: "" };
}

export const DeleteArtRequest = {
  encode(message: DeleteArtRequest, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.id !== "") {
      writer.uint32(10).string(message.id);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): DeleteArtRequest {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseDeleteArtRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.id = reader.string();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): DeleteArtRequest {
    return { id: isSet(object.id) ? globalThis.String(object.id) : "" };
  },

  toJSON(message: DeleteArtRequest): unknown {
    const obj: any = {};
    if (message.id !== "") {
      obj.id = message.id;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<DeleteArtRequest>, I>>(base?: I): DeleteArtRequest {
    return DeleteArtRequest.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<DeleteArtRequest>, I>>(object: I): DeleteArtRequest {
    const message = createBaseDeleteArtRequest();
    message.id = object.id ?? "";
    return message;
  },
};

function createBaseDeleteArtResponse(): DeleteArtResponse {
  return { art: undefined };
}

export const DeleteArtResponse = {
  encode(message: DeleteArtResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.art !== undefined) {
      Art.encode(message.art, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): DeleteArtResponse {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseDeleteArtResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.art = Art.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): DeleteArtResponse {
    return { art: isSet(object.art) ? Art.fromJSON(object.art) : undefined };
  },

  toJSON(message: DeleteArtResponse): unknown {
    const obj: any = {};
    if (message.art !== undefined) {
      obj.art = Art.toJSON(message.art);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<DeleteArtResponse>, I>>(base?: I): DeleteArtResponse {
    return DeleteArtResponse.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<DeleteArtResponse>, I>>(object: I): DeleteArtResponse {
    const message = createBaseDeleteArtResponse();
    message.art = (object.art !== undefined && object.art !== null) ? Art.fromPartial(object.art) : undefined;
    return message;
  },
};

type Builtin = Date | Function | Uint8Array | string | number | boolean | undefined;

export type DeepPartial<T> = T extends Builtin ? T
  : T extends globalThis.Array<infer U> ? globalThis.Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>>
  : T extends {} ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

type KeysOfUnion<T> = T extends T ? keyof T : never;
export type Exact<P, I extends P> = P extends Builtin ? P
  : P & { [K in keyof P]: Exact<P[K], I[K]> } & { [K in Exclude<keyof I, KeysOfUnion<P>>]: never };

function toTimestamp(date: Date): Timestamp {
  const seconds = Math.trunc(date.getTime() / 1_000);
  const nanos = (date.getTime() % 1_000) * 1_000_000;
  return { seconds, nanos };
}

function fromTimestamp(t: Timestamp): Date {
  let millis = (t.seconds || 0) * 1_000;
  millis += (t.nanos || 0) / 1_000_000;
  return new globalThis.Date(millis);
}

function fromJsonTimestamp(o: any): Date {
  if (o instanceof globalThis.Date) {
    return o;
  } else if (typeof o === "string") {
    return new globalThis.Date(o);
  } else {
    return fromTimestamp(Timestamp.fromJSON(o));
  }
}

function isSet(value: any): boolean {
  return value !== null && value !== undefined;
}
