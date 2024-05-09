import * as jspb from 'google-protobuf'



export class HttpRequest extends jspb.Message {
  getMethod(): string;
  setMethod(value: string): HttpRequest;

  getUri(): string;
  setUri(value: string): HttpRequest;

  getHeadersList(): Array<HttpHeader>;
  setHeadersList(value: Array<HttpHeader>): HttpRequest;
  clearHeadersList(): HttpRequest;
  addHeaders(value?: HttpHeader, index?: number): HttpHeader;

  getBody(): Uint8Array | string;
  getBody_asU8(): Uint8Array;
  getBody_asB64(): string;
  setBody(value: Uint8Array | string): HttpRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): HttpRequest.AsObject;
  static toObject(includeInstance: boolean, msg: HttpRequest): HttpRequest.AsObject;
  static serializeBinaryToWriter(message: HttpRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): HttpRequest;
  static deserializeBinaryFromReader(message: HttpRequest, reader: jspb.BinaryReader): HttpRequest;
}

export namespace HttpRequest {
  export type AsObject = {
    method: string,
    uri: string,
    headersList: Array<HttpHeader.AsObject>,
    body: Uint8Array | string,
  }
}

export class HttpResponse extends jspb.Message {
  getStatus(): number;
  setStatus(value: number): HttpResponse;

  getReason(): string;
  setReason(value: string): HttpResponse;

  getHeadersList(): Array<HttpHeader>;
  setHeadersList(value: Array<HttpHeader>): HttpResponse;
  clearHeadersList(): HttpResponse;
  addHeaders(value?: HttpHeader, index?: number): HttpHeader;

  getBody(): Uint8Array | string;
  getBody_asU8(): Uint8Array;
  getBody_asB64(): string;
  setBody(value: Uint8Array | string): HttpResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): HttpResponse.AsObject;
  static toObject(includeInstance: boolean, msg: HttpResponse): HttpResponse.AsObject;
  static serializeBinaryToWriter(message: HttpResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): HttpResponse;
  static deserializeBinaryFromReader(message: HttpResponse, reader: jspb.BinaryReader): HttpResponse;
}

export namespace HttpResponse {
  export type AsObject = {
    status: number,
    reason: string,
    headersList: Array<HttpHeader.AsObject>,
    body: Uint8Array | string,
  }
}

export class HttpHeader extends jspb.Message {
  getKey(): string;
  setKey(value: string): HttpHeader;

  getValue(): string;
  setValue(value: string): HttpHeader;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): HttpHeader.AsObject;
  static toObject(includeInstance: boolean, msg: HttpHeader): HttpHeader.AsObject;
  static serializeBinaryToWriter(message: HttpHeader, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): HttpHeader;
  static deserializeBinaryFromReader(message: HttpHeader, reader: jspb.BinaryReader): HttpHeader;
}

export namespace HttpHeader {
  export type AsObject = {
    key: string,
    value: string,
  }
}

