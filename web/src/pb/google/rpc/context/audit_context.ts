/**
 * Generated by the protoc-gen-ts.  DO NOT EDIT!
 * compiler version: 5.26.1
 * source: google/rpc/context/audit_context.proto
 * git: https://github.com/thesayyn/protoc-gen-ts */
import * as dependency_1 from "./../../protobuf/struct";
import * as pb_1 from "google-protobuf";
export namespace google.rpc.context {
    export class AuditContext extends pb_1.Message {
        #one_of_decls: number[][] = [];
        constructor(data?: any[] | {
            audit_log?: Uint8Array;
            scrubbed_request?: dependency_1.google.protobuf.Struct;
            scrubbed_response?: dependency_1.google.protobuf.Struct;
            scrubbed_response_item_count?: number;
            target_resource?: string;
        }) {
            super();
            pb_1.Message.initialize(this, Array.isArray(data) ? data : [], 0, -1, [], this.#one_of_decls);
            if (!Array.isArray(data) && typeof data == "object") {
                if ("audit_log" in data && data.audit_log != undefined) {
                    this.audit_log = data.audit_log;
                }
                if ("scrubbed_request" in data && data.scrubbed_request != undefined) {
                    this.scrubbed_request = data.scrubbed_request;
                }
                if ("scrubbed_response" in data && data.scrubbed_response != undefined) {
                    this.scrubbed_response = data.scrubbed_response;
                }
                if ("scrubbed_response_item_count" in data && data.scrubbed_response_item_count != undefined) {
                    this.scrubbed_response_item_count = data.scrubbed_response_item_count;
                }
                if ("target_resource" in data && data.target_resource != undefined) {
                    this.target_resource = data.target_resource;
                }
            }
        }
        get audit_log() {
            return pb_1.Message.getFieldWithDefault(this, 1, new Uint8Array(0)) as Uint8Array;
        }
        set audit_log(value: Uint8Array) {
            pb_1.Message.setField(this, 1, value);
        }
        get scrubbed_request() {
            return pb_1.Message.getWrapperField(this, dependency_1.google.protobuf.Struct, 2) as dependency_1.google.protobuf.Struct;
        }
        set scrubbed_request(value: dependency_1.google.protobuf.Struct) {
            pb_1.Message.setWrapperField(this, 2, value);
        }
        get has_scrubbed_request() {
            return pb_1.Message.getField(this, 2) != null;
        }
        get scrubbed_response() {
            return pb_1.Message.getWrapperField(this, dependency_1.google.protobuf.Struct, 3) as dependency_1.google.protobuf.Struct;
        }
        set scrubbed_response(value: dependency_1.google.protobuf.Struct) {
            pb_1.Message.setWrapperField(this, 3, value);
        }
        get has_scrubbed_response() {
            return pb_1.Message.getField(this, 3) != null;
        }
        get scrubbed_response_item_count() {
            return pb_1.Message.getFieldWithDefault(this, 4, 0) as number;
        }
        set scrubbed_response_item_count(value: number) {
            pb_1.Message.setField(this, 4, value);
        }
        get target_resource() {
            return pb_1.Message.getFieldWithDefault(this, 5, "") as string;
        }
        set target_resource(value: string) {
            pb_1.Message.setField(this, 5, value);
        }
        static fromObject(data: {
            audit_log?: Uint8Array;
            scrubbed_request?: ReturnType<typeof dependency_1.google.protobuf.Struct.prototype.toObject>;
            scrubbed_response?: ReturnType<typeof dependency_1.google.protobuf.Struct.prototype.toObject>;
            scrubbed_response_item_count?: number;
            target_resource?: string;
        }): AuditContext {
            const message = new AuditContext({});
            if (data.audit_log != null) {
                message.audit_log = data.audit_log;
            }
            if (data.scrubbed_request != null) {
                message.scrubbed_request = dependency_1.google.protobuf.Struct.fromObject(data.scrubbed_request);
            }
            if (data.scrubbed_response != null) {
                message.scrubbed_response = dependency_1.google.protobuf.Struct.fromObject(data.scrubbed_response);
            }
            if (data.scrubbed_response_item_count != null) {
                message.scrubbed_response_item_count = data.scrubbed_response_item_count;
            }
            if (data.target_resource != null) {
                message.target_resource = data.target_resource;
            }
            return message;
        }
        toObject() {
            const data: {
                audit_log?: Uint8Array;
                scrubbed_request?: ReturnType<typeof dependency_1.google.protobuf.Struct.prototype.toObject>;
                scrubbed_response?: ReturnType<typeof dependency_1.google.protobuf.Struct.prototype.toObject>;
                scrubbed_response_item_count?: number;
                target_resource?: string;
            } = {};
            if (this.audit_log != null) {
                data.audit_log = this.audit_log;
            }
            if (this.scrubbed_request != null) {
                data.scrubbed_request = this.scrubbed_request.toObject();
            }
            if (this.scrubbed_response != null) {
                data.scrubbed_response = this.scrubbed_response.toObject();
            }
            if (this.scrubbed_response_item_count != null) {
                data.scrubbed_response_item_count = this.scrubbed_response_item_count;
            }
            if (this.target_resource != null) {
                data.target_resource = this.target_resource;
            }
            return data;
        }
        serialize(): Uint8Array;
        serialize(w: pb_1.BinaryWriter): void;
        serialize(w?: pb_1.BinaryWriter): Uint8Array | void {
            const writer = w || new pb_1.BinaryWriter();
            if (this.audit_log.length)
                writer.writeBytes(1, this.audit_log);
            if (this.has_scrubbed_request)
                writer.writeMessage(2, this.scrubbed_request, () => this.scrubbed_request.serialize(writer));
            if (this.has_scrubbed_response)
                writer.writeMessage(3, this.scrubbed_response, () => this.scrubbed_response.serialize(writer));
            if (this.scrubbed_response_item_count != 0)
                writer.writeInt32(4, this.scrubbed_response_item_count);
            if (this.target_resource.length)
                writer.writeString(5, this.target_resource);
            if (!w)
                return writer.getResultBuffer();
        }
        static deserialize(bytes: Uint8Array | pb_1.BinaryReader): AuditContext {
            const reader = bytes instanceof pb_1.BinaryReader ? bytes : new pb_1.BinaryReader(bytes), message = new AuditContext();
            while (reader.nextField()) {
                if (reader.isEndGroup())
                    break;
                switch (reader.getFieldNumber()) {
                    case 1:
                        message.audit_log = reader.readBytes();
                        break;
                    case 2:
                        reader.readMessage(message.scrubbed_request, () => message.scrubbed_request = dependency_1.google.protobuf.Struct.deserialize(reader));
                        break;
                    case 3:
                        reader.readMessage(message.scrubbed_response, () => message.scrubbed_response = dependency_1.google.protobuf.Struct.deserialize(reader));
                        break;
                    case 4:
                        message.scrubbed_response_item_count = reader.readInt32();
                        break;
                    case 5:
                        message.target_resource = reader.readString();
                        break;
                    default: reader.skipField();
                }
            }
            return message;
        }
        serializeBinary(): Uint8Array {
            return this.serialize();
        }
        static deserializeBinary(bytes: Uint8Array): AuditContext {
            return AuditContext.deserialize(bytes);
        }
    }
}
