/**
 * Generated by the protoc-gen-ts.  DO NOT EDIT!
 * compiler version: 5.26.1
 * source: google/rpc/http.proto
 * git: https://github.com/thesayyn/protoc-gen-ts */
import * as pb_1 from "google-protobuf";
export namespace google.rpc {
    export class HttpRequest extends pb_1.Message {
        #one_of_decls: number[][] = [];
        constructor(data?: any[] | {
            method?: string;
            uri?: string;
            headers?: HttpHeader[];
            body?: Uint8Array;
        }) {
            super();
            pb_1.Message.initialize(this, Array.isArray(data) ? data : [], 0, -1, [3], this.#one_of_decls);
            if (!Array.isArray(data) && typeof data == "object") {
                if ("method" in data && data.method != undefined) {
                    this.method = data.method;
                }
                if ("uri" in data && data.uri != undefined) {
                    this.uri = data.uri;
                }
                if ("headers" in data && data.headers != undefined) {
                    this.headers = data.headers;
                }
                if ("body" in data && data.body != undefined) {
                    this.body = data.body;
                }
            }
        }
        get method() {
            return pb_1.Message.getFieldWithDefault(this, 1, "") as string;
        }
        set method(value: string) {
            pb_1.Message.setField(this, 1, value);
        }
        get uri() {
            return pb_1.Message.getFieldWithDefault(this, 2, "") as string;
        }
        set uri(value: string) {
            pb_1.Message.setField(this, 2, value);
        }
        get headers() {
            return pb_1.Message.getRepeatedWrapperField(this, HttpHeader, 3) as HttpHeader[];
        }
        set headers(value: HttpHeader[]) {
            pb_1.Message.setRepeatedWrapperField(this, 3, value);
        }
        get body() {
            return pb_1.Message.getFieldWithDefault(this, 4, new Uint8Array(0)) as Uint8Array;
        }
        set body(value: Uint8Array) {
            pb_1.Message.setField(this, 4, value);
        }
        static fromObject(data: {
            method?: string;
            uri?: string;
            headers?: ReturnType<typeof HttpHeader.prototype.toObject>[];
            body?: Uint8Array;
        }): HttpRequest {
            const message = new HttpRequest({});
            if (data.method != null) {
                message.method = data.method;
            }
            if (data.uri != null) {
                message.uri = data.uri;
            }
            if (data.headers != null) {
                message.headers = data.headers.map(item => HttpHeader.fromObject(item));
            }
            if (data.body != null) {
                message.body = data.body;
            }
            return message;
        }
        toObject() {
            const data: {
                method?: string;
                uri?: string;
                headers?: ReturnType<typeof HttpHeader.prototype.toObject>[];
                body?: Uint8Array;
            } = {};
            if (this.method != null) {
                data.method = this.method;
            }
            if (this.uri != null) {
                data.uri = this.uri;
            }
            if (this.headers != null) {
                data.headers = this.headers.map((item: HttpHeader) => item.toObject());
            }
            if (this.body != null) {
                data.body = this.body;
            }
            return data;
        }
        serialize(): Uint8Array;
        serialize(w: pb_1.BinaryWriter): void;
        serialize(w?: pb_1.BinaryWriter): Uint8Array | void {
            const writer = w || new pb_1.BinaryWriter();
            if (this.method.length)
                writer.writeString(1, this.method);
            if (this.uri.length)
                writer.writeString(2, this.uri);
            if (this.headers.length)
                writer.writeRepeatedMessage(3, this.headers, (item: HttpHeader) => item.serialize(writer));
            if (this.body.length)
                writer.writeBytes(4, this.body);
            if (!w)
                return writer.getResultBuffer();
        }
        static deserialize(bytes: Uint8Array | pb_1.BinaryReader): HttpRequest {
            const reader = bytes instanceof pb_1.BinaryReader ? bytes : new pb_1.BinaryReader(bytes), message = new HttpRequest();
            while (reader.nextField()) {
                if (reader.isEndGroup())
                    break;
                switch (reader.getFieldNumber()) {
                    case 1:
                        message.method = reader.readString();
                        break;
                    case 2:
                        message.uri = reader.readString();
                        break;
                    case 3:
                        reader.readMessage(message.headers, () => pb_1.Message.addToRepeatedWrapperField(message, 3, HttpHeader.deserialize(reader), HttpHeader));
                        break;
                    case 4:
                        message.body = reader.readBytes();
                        break;
                    default: reader.skipField();
                }
            }
            return message;
        }
        serializeBinary(): Uint8Array {
            return this.serialize();
        }
        static deserializeBinary(bytes: Uint8Array): HttpRequest {
            return HttpRequest.deserialize(bytes);
        }
    }
    export class HttpResponse extends pb_1.Message {
        #one_of_decls: number[][] = [];
        constructor(data?: any[] | {
            status?: number;
            reason?: string;
            headers?: HttpHeader[];
            body?: Uint8Array;
        }) {
            super();
            pb_1.Message.initialize(this, Array.isArray(data) ? data : [], 0, -1, [3], this.#one_of_decls);
            if (!Array.isArray(data) && typeof data == "object") {
                if ("status" in data && data.status != undefined) {
                    this.status = data.status;
                }
                if ("reason" in data && data.reason != undefined) {
                    this.reason = data.reason;
                }
                if ("headers" in data && data.headers != undefined) {
                    this.headers = data.headers;
                }
                if ("body" in data && data.body != undefined) {
                    this.body = data.body;
                }
            }
        }
        get status() {
            return pb_1.Message.getFieldWithDefault(this, 1, 0) as number;
        }
        set status(value: number) {
            pb_1.Message.setField(this, 1, value);
        }
        get reason() {
            return pb_1.Message.getFieldWithDefault(this, 2, "") as string;
        }
        set reason(value: string) {
            pb_1.Message.setField(this, 2, value);
        }
        get headers() {
            return pb_1.Message.getRepeatedWrapperField(this, HttpHeader, 3) as HttpHeader[];
        }
        set headers(value: HttpHeader[]) {
            pb_1.Message.setRepeatedWrapperField(this, 3, value);
        }
        get body() {
            return pb_1.Message.getFieldWithDefault(this, 4, new Uint8Array(0)) as Uint8Array;
        }
        set body(value: Uint8Array) {
            pb_1.Message.setField(this, 4, value);
        }
        static fromObject(data: {
            status?: number;
            reason?: string;
            headers?: ReturnType<typeof HttpHeader.prototype.toObject>[];
            body?: Uint8Array;
        }): HttpResponse {
            const message = new HttpResponse({});
            if (data.status != null) {
                message.status = data.status;
            }
            if (data.reason != null) {
                message.reason = data.reason;
            }
            if (data.headers != null) {
                message.headers = data.headers.map(item => HttpHeader.fromObject(item));
            }
            if (data.body != null) {
                message.body = data.body;
            }
            return message;
        }
        toObject() {
            const data: {
                status?: number;
                reason?: string;
                headers?: ReturnType<typeof HttpHeader.prototype.toObject>[];
                body?: Uint8Array;
            } = {};
            if (this.status != null) {
                data.status = this.status;
            }
            if (this.reason != null) {
                data.reason = this.reason;
            }
            if (this.headers != null) {
                data.headers = this.headers.map((item: HttpHeader) => item.toObject());
            }
            if (this.body != null) {
                data.body = this.body;
            }
            return data;
        }
        serialize(): Uint8Array;
        serialize(w: pb_1.BinaryWriter): void;
        serialize(w?: pb_1.BinaryWriter): Uint8Array | void {
            const writer = w || new pb_1.BinaryWriter();
            if (this.status != 0)
                writer.writeInt32(1, this.status);
            if (this.reason.length)
                writer.writeString(2, this.reason);
            if (this.headers.length)
                writer.writeRepeatedMessage(3, this.headers, (item: HttpHeader) => item.serialize(writer));
            if (this.body.length)
                writer.writeBytes(4, this.body);
            if (!w)
                return writer.getResultBuffer();
        }
        static deserialize(bytes: Uint8Array | pb_1.BinaryReader): HttpResponse {
            const reader = bytes instanceof pb_1.BinaryReader ? bytes : new pb_1.BinaryReader(bytes), message = new HttpResponse();
            while (reader.nextField()) {
                if (reader.isEndGroup())
                    break;
                switch (reader.getFieldNumber()) {
                    case 1:
                        message.status = reader.readInt32();
                        break;
                    case 2:
                        message.reason = reader.readString();
                        break;
                    case 3:
                        reader.readMessage(message.headers, () => pb_1.Message.addToRepeatedWrapperField(message, 3, HttpHeader.deserialize(reader), HttpHeader));
                        break;
                    case 4:
                        message.body = reader.readBytes();
                        break;
                    default: reader.skipField();
                }
            }
            return message;
        }
        serializeBinary(): Uint8Array {
            return this.serialize();
        }
        static deserializeBinary(bytes: Uint8Array): HttpResponse {
            return HttpResponse.deserialize(bytes);
        }
    }
    export class HttpHeader extends pb_1.Message {
        #one_of_decls: number[][] = [];
        constructor(data?: any[] | {
            key?: string;
            value?: string;
        }) {
            super();
            pb_1.Message.initialize(this, Array.isArray(data) ? data : [], 0, -1, [], this.#one_of_decls);
            if (!Array.isArray(data) && typeof data == "object") {
                if ("key" in data && data.key != undefined) {
                    this.key = data.key;
                }
                if ("value" in data && data.value != undefined) {
                    this.value = data.value;
                }
            }
        }
        get key() {
            return pb_1.Message.getFieldWithDefault(this, 1, "") as string;
        }
        set key(value: string) {
            pb_1.Message.setField(this, 1, value);
        }
        get value() {
            return pb_1.Message.getFieldWithDefault(this, 2, "") as string;
        }
        set value(value: string) {
            pb_1.Message.setField(this, 2, value);
        }
        static fromObject(data: {
            key?: string;
            value?: string;
        }): HttpHeader {
            const message = new HttpHeader({});
            if (data.key != null) {
                message.key = data.key;
            }
            if (data.value != null) {
                message.value = data.value;
            }
            return message;
        }
        toObject() {
            const data: {
                key?: string;
                value?: string;
            } = {};
            if (this.key != null) {
                data.key = this.key;
            }
            if (this.value != null) {
                data.value = this.value;
            }
            return data;
        }
        serialize(): Uint8Array;
        serialize(w: pb_1.BinaryWriter): void;
        serialize(w?: pb_1.BinaryWriter): Uint8Array | void {
            const writer = w || new pb_1.BinaryWriter();
            if (this.key.length)
                writer.writeString(1, this.key);
            if (this.value.length)
                writer.writeString(2, this.value);
            if (!w)
                return writer.getResultBuffer();
        }
        static deserialize(bytes: Uint8Array | pb_1.BinaryReader): HttpHeader {
            const reader = bytes instanceof pb_1.BinaryReader ? bytes : new pb_1.BinaryReader(bytes), message = new HttpHeader();
            while (reader.nextField()) {
                if (reader.isEndGroup())
                    break;
                switch (reader.getFieldNumber()) {
                    case 1:
                        message.key = reader.readString();
                        break;
                    case 2:
                        message.value = reader.readString();
                        break;
                    default: reader.skipField();
                }
            }
            return message;
        }
        serializeBinary(): Uint8Array {
            return this.serialize();
        }
        static deserializeBinary(bytes: Uint8Array): HttpHeader {
            return HttpHeader.deserialize(bytes);
        }
    }
}
