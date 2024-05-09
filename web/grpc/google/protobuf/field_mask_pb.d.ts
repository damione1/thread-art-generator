import * as jspb from 'google-protobuf'



export class FieldMask extends jspb.Message {
  getPathsList(): Array<string>;
  setPathsList(value: Array<string>): FieldMask;
  clearPathsList(): FieldMask;
  addPaths(value: string, index?: number): FieldMask;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FieldMask.AsObject;
  static toObject(includeInstance: boolean, msg: FieldMask): FieldMask.AsObject;
  static serializeBinaryToWriter(message: FieldMask, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FieldMask;
  static deserializeBinaryFromReader(message: FieldMask, reader: jspb.BinaryReader): FieldMask;
}

export namespace FieldMask {
  export type AsObject = {
    pathsList: Array<string>,
  }
}

