import * as jspb from 'google-protobuf'

import * as google_protobuf_struct_pb from 'google-protobuf/google/protobuf/struct_pb'; // proto import: "google/protobuf/struct.proto"


export class AuditContext extends jspb.Message {
  getAuditLog(): Uint8Array | string;
  getAuditLog_asU8(): Uint8Array;
  getAuditLog_asB64(): string;
  setAuditLog(value: Uint8Array | string): AuditContext;

  getScrubbedRequest(): google_protobuf_struct_pb.Struct | undefined;
  setScrubbedRequest(value?: google_protobuf_struct_pb.Struct): AuditContext;
  hasScrubbedRequest(): boolean;
  clearScrubbedRequest(): AuditContext;

  getScrubbedResponse(): google_protobuf_struct_pb.Struct | undefined;
  setScrubbedResponse(value?: google_protobuf_struct_pb.Struct): AuditContext;
  hasScrubbedResponse(): boolean;
  clearScrubbedResponse(): AuditContext;

  getScrubbedResponseItemCount(): number;
  setScrubbedResponseItemCount(value: number): AuditContext;

  getTargetResource(): string;
  setTargetResource(value: string): AuditContext;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AuditContext.AsObject;
  static toObject(includeInstance: boolean, msg: AuditContext): AuditContext.AsObject;
  static serializeBinaryToWriter(message: AuditContext, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AuditContext;
  static deserializeBinaryFromReader(message: AuditContext, reader: jspb.BinaryReader): AuditContext;
}

export namespace AuditContext {
  export type AsObject = {
    auditLog: Uint8Array | string,
    scrubbedRequest?: google_protobuf_struct_pb.Struct.AsObject,
    scrubbedResponse?: google_protobuf_struct_pb.Struct.AsObject,
    scrubbedResponseItemCount: number,
    targetResource: string,
  }
}

