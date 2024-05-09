import * as jspb from 'google-protobuf'

import * as google_protobuf_any_pb from 'google-protobuf/google/protobuf/any_pb'; // proto import: "google/protobuf/any.proto"
import * as google_protobuf_duration_pb from 'google-protobuf/google/protobuf/duration_pb'; // proto import: "google/protobuf/duration.proto"
import * as google_protobuf_struct_pb from 'google-protobuf/google/protobuf/struct_pb'; // proto import: "google/protobuf/struct.proto"
import * as google_protobuf_timestamp_pb from 'google-protobuf/google/protobuf/timestamp_pb'; // proto import: "google/protobuf/timestamp.proto"


export class AttributeContext extends jspb.Message {
  getOrigin(): AttributeContext.Peer | undefined;
  setOrigin(value?: AttributeContext.Peer): AttributeContext;
  hasOrigin(): boolean;
  clearOrigin(): AttributeContext;

  getSource(): AttributeContext.Peer | undefined;
  setSource(value?: AttributeContext.Peer): AttributeContext;
  hasSource(): boolean;
  clearSource(): AttributeContext;

  getDestination(): AttributeContext.Peer | undefined;
  setDestination(value?: AttributeContext.Peer): AttributeContext;
  hasDestination(): boolean;
  clearDestination(): AttributeContext;

  getRequest(): AttributeContext.Request | undefined;
  setRequest(value?: AttributeContext.Request): AttributeContext;
  hasRequest(): boolean;
  clearRequest(): AttributeContext;

  getResponse(): AttributeContext.Response | undefined;
  setResponse(value?: AttributeContext.Response): AttributeContext;
  hasResponse(): boolean;
  clearResponse(): AttributeContext;

  getResource(): AttributeContext.Resource | undefined;
  setResource(value?: AttributeContext.Resource): AttributeContext;
  hasResource(): boolean;
  clearResource(): AttributeContext;

  getApi(): AttributeContext.Api | undefined;
  setApi(value?: AttributeContext.Api): AttributeContext;
  hasApi(): boolean;
  clearApi(): AttributeContext;

  getExtensionsList(): Array<google_protobuf_any_pb.Any>;
  setExtensionsList(value: Array<google_protobuf_any_pb.Any>): AttributeContext;
  clearExtensionsList(): AttributeContext;
  addExtensions(value?: google_protobuf_any_pb.Any, index?: number): google_protobuf_any_pb.Any;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AttributeContext.AsObject;
  static toObject(includeInstance: boolean, msg: AttributeContext): AttributeContext.AsObject;
  static serializeBinaryToWriter(message: AttributeContext, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AttributeContext;
  static deserializeBinaryFromReader(message: AttributeContext, reader: jspb.BinaryReader): AttributeContext;
}

export namespace AttributeContext {
  export type AsObject = {
    origin?: AttributeContext.Peer.AsObject,
    source?: AttributeContext.Peer.AsObject,
    destination?: AttributeContext.Peer.AsObject,
    request?: AttributeContext.Request.AsObject,
    response?: AttributeContext.Response.AsObject,
    resource?: AttributeContext.Resource.AsObject,
    api?: AttributeContext.Api.AsObject,
    extensionsList: Array<google_protobuf_any_pb.Any.AsObject>,
  }

  export class Peer extends jspb.Message {
    getIp(): string;
    setIp(value: string): Peer;

    getPort(): number;
    setPort(value: number): Peer;

    getLabelsMap(): jspb.Map<string, string>;
    clearLabelsMap(): Peer;

    getPrincipal(): string;
    setPrincipal(value: string): Peer;

    getRegionCode(): string;
    setRegionCode(value: string): Peer;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Peer.AsObject;
    static toObject(includeInstance: boolean, msg: Peer): Peer.AsObject;
    static serializeBinaryToWriter(message: Peer, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Peer;
    static deserializeBinaryFromReader(message: Peer, reader: jspb.BinaryReader): Peer;
  }

  export namespace Peer {
    export type AsObject = {
      ip: string,
      port: number,
      labelsMap: Array<[string, string]>,
      principal: string,
      regionCode: string,
    }
  }


  export class Api extends jspb.Message {
    getService(): string;
    setService(value: string): Api;

    getOperation(): string;
    setOperation(value: string): Api;

    getProtocol(): string;
    setProtocol(value: string): Api;

    getVersion(): string;
    setVersion(value: string): Api;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Api.AsObject;
    static toObject(includeInstance: boolean, msg: Api): Api.AsObject;
    static serializeBinaryToWriter(message: Api, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Api;
    static deserializeBinaryFromReader(message: Api, reader: jspb.BinaryReader): Api;
  }

  export namespace Api {
    export type AsObject = {
      service: string,
      operation: string,
      protocol: string,
      version: string,
    }
  }


  export class Auth extends jspb.Message {
    getPrincipal(): string;
    setPrincipal(value: string): Auth;

    getAudiencesList(): Array<string>;
    setAudiencesList(value: Array<string>): Auth;
    clearAudiencesList(): Auth;
    addAudiences(value: string, index?: number): Auth;

    getPresenter(): string;
    setPresenter(value: string): Auth;

    getClaims(): google_protobuf_struct_pb.Struct | undefined;
    setClaims(value?: google_protobuf_struct_pb.Struct): Auth;
    hasClaims(): boolean;
    clearClaims(): Auth;

    getAccessLevelsList(): Array<string>;
    setAccessLevelsList(value: Array<string>): Auth;
    clearAccessLevelsList(): Auth;
    addAccessLevels(value: string, index?: number): Auth;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Auth.AsObject;
    static toObject(includeInstance: boolean, msg: Auth): Auth.AsObject;
    static serializeBinaryToWriter(message: Auth, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Auth;
    static deserializeBinaryFromReader(message: Auth, reader: jspb.BinaryReader): Auth;
  }

  export namespace Auth {
    export type AsObject = {
      principal: string,
      audiencesList: Array<string>,
      presenter: string,
      claims?: google_protobuf_struct_pb.Struct.AsObject,
      accessLevelsList: Array<string>,
    }
  }


  export class Request extends jspb.Message {
    getId(): string;
    setId(value: string): Request;

    getMethod(): string;
    setMethod(value: string): Request;

    getHeadersMap(): jspb.Map<string, string>;
    clearHeadersMap(): Request;

    getPath(): string;
    setPath(value: string): Request;

    getHost(): string;
    setHost(value: string): Request;

    getScheme(): string;
    setScheme(value: string): Request;

    getQuery(): string;
    setQuery(value: string): Request;

    getTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setTime(value?: google_protobuf_timestamp_pb.Timestamp): Request;
    hasTime(): boolean;
    clearTime(): Request;

    getSize(): number;
    setSize(value: number): Request;

    getProtocol(): string;
    setProtocol(value: string): Request;

    getReason(): string;
    setReason(value: string): Request;

    getAuth(): AttributeContext.Auth | undefined;
    setAuth(value?: AttributeContext.Auth): Request;
    hasAuth(): boolean;
    clearAuth(): Request;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Request.AsObject;
    static toObject(includeInstance: boolean, msg: Request): Request.AsObject;
    static serializeBinaryToWriter(message: Request, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Request;
    static deserializeBinaryFromReader(message: Request, reader: jspb.BinaryReader): Request;
  }

  export namespace Request {
    export type AsObject = {
      id: string,
      method: string,
      headersMap: Array<[string, string]>,
      path: string,
      host: string,
      scheme: string,
      query: string,
      time?: google_protobuf_timestamp_pb.Timestamp.AsObject,
      size: number,
      protocol: string,
      reason: string,
      auth?: AttributeContext.Auth.AsObject,
    }
  }


  export class Response extends jspb.Message {
    getCode(): number;
    setCode(value: number): Response;

    getSize(): number;
    setSize(value: number): Response;

    getHeadersMap(): jspb.Map<string, string>;
    clearHeadersMap(): Response;

    getTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setTime(value?: google_protobuf_timestamp_pb.Timestamp): Response;
    hasTime(): boolean;
    clearTime(): Response;

    getBackendLatency(): google_protobuf_duration_pb.Duration | undefined;
    setBackendLatency(value?: google_protobuf_duration_pb.Duration): Response;
    hasBackendLatency(): boolean;
    clearBackendLatency(): Response;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Response.AsObject;
    static toObject(includeInstance: boolean, msg: Response): Response.AsObject;
    static serializeBinaryToWriter(message: Response, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Response;
    static deserializeBinaryFromReader(message: Response, reader: jspb.BinaryReader): Response;
  }

  export namespace Response {
    export type AsObject = {
      code: number,
      size: number,
      headersMap: Array<[string, string]>,
      time?: google_protobuf_timestamp_pb.Timestamp.AsObject,
      backendLatency?: google_protobuf_duration_pb.Duration.AsObject,
    }
  }


  export class Resource extends jspb.Message {
    getService(): string;
    setService(value: string): Resource;

    getName(): string;
    setName(value: string): Resource;

    getType(): string;
    setType(value: string): Resource;

    getLabelsMap(): jspb.Map<string, string>;
    clearLabelsMap(): Resource;

    getUid(): string;
    setUid(value: string): Resource;

    getAnnotationsMap(): jspb.Map<string, string>;
    clearAnnotationsMap(): Resource;

    getDisplayName(): string;
    setDisplayName(value: string): Resource;

    getCreateTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setCreateTime(value?: google_protobuf_timestamp_pb.Timestamp): Resource;
    hasCreateTime(): boolean;
    clearCreateTime(): Resource;

    getUpdateTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setUpdateTime(value?: google_protobuf_timestamp_pb.Timestamp): Resource;
    hasUpdateTime(): boolean;
    clearUpdateTime(): Resource;

    getDeleteTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setDeleteTime(value?: google_protobuf_timestamp_pb.Timestamp): Resource;
    hasDeleteTime(): boolean;
    clearDeleteTime(): Resource;

    getEtag(): string;
    setEtag(value: string): Resource;

    getLocation(): string;
    setLocation(value: string): Resource;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Resource.AsObject;
    static toObject(includeInstance: boolean, msg: Resource): Resource.AsObject;
    static serializeBinaryToWriter(message: Resource, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Resource;
    static deserializeBinaryFromReader(message: Resource, reader: jspb.BinaryReader): Resource;
  }

  export namespace Resource {
    export type AsObject = {
      service: string,
      name: string,
      type: string,
      labelsMap: Array<[string, string]>,
      uid: string,
      annotationsMap: Array<[string, string]>,
      displayName: string,
      createTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
      updateTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
      deleteTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
      etag: string,
      location: string,
    }
  }

}

