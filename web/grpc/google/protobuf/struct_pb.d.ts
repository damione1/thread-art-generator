import * as jspb from 'google-protobuf'



export class Struct extends jspb.Message {
  getFieldsMap(): jspb.Map<string, Value>;
  clearFieldsMap(): Struct;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Struct.AsObject;
  static toObject(includeInstance: boolean, msg: Struct): Struct.AsObject;
  static serializeBinaryToWriter(message: Struct, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Struct;
  static deserializeBinaryFromReader(message: Struct, reader: jspb.BinaryReader): Struct;
}

export namespace Struct {
  export type AsObject = {
    fieldsMap: Array<[string, Value.AsObject]>,
  }
}

export class Value extends jspb.Message {
  getNullValue(): NullValue;
  setNullValue(value: NullValue): Value;

  getNumberValue(): number;
  setNumberValue(value: number): Value;

  getStringValue(): string;
  setStringValue(value: string): Value;

  getBoolValue(): boolean;
  setBoolValue(value: boolean): Value;

  getStructValue(): Struct | undefined;
  setStructValue(value?: Struct): Value;
  hasStructValue(): boolean;
  clearStructValue(): Value;

  getListValue(): ListValue | undefined;
  setListValue(value?: ListValue): Value;
  hasListValue(): boolean;
  clearListValue(): Value;

  getKindCase(): Value.KindCase;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Value.AsObject;
  static toObject(includeInstance: boolean, msg: Value): Value.AsObject;
  static serializeBinaryToWriter(message: Value, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Value;
  static deserializeBinaryFromReader(message: Value, reader: jspb.BinaryReader): Value;
}

export namespace Value {
  export type AsObject = {
    nullValue: NullValue,
    numberValue: number,
    stringValue: string,
    boolValue: boolean,
    structValue?: Struct.AsObject,
    listValue?: ListValue.AsObject,
  }

  export enum KindCase { 
    KIND_NOT_SET = 0,
    NULL_VALUE = 1,
    NUMBER_VALUE = 2,
    STRING_VALUE = 3,
    BOOL_VALUE = 4,
    STRUCT_VALUE = 5,
    LIST_VALUE = 6,
  }
}

export class ListValue extends jspb.Message {
  getValuesList(): Array<Value>;
  setValuesList(value: Array<Value>): ListValue;
  clearValuesList(): ListValue;
  addValues(value?: Value, index?: number): Value;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ListValue.AsObject;
  static toObject(includeInstance: boolean, msg: ListValue): ListValue.AsObject;
  static serializeBinaryToWriter(message: ListValue, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ListValue;
  static deserializeBinaryFromReader(message: ListValue, reader: jspb.BinaryReader): ListValue;
}

export namespace ListValue {
  export type AsObject = {
    valuesList: Array<Value.AsObject>,
  }
}

export enum NullValue { 
  NULL_VALUE = 0,
}
