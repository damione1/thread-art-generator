// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// @generated by protoc-gen-es v1.4.1 with parameter "target=ts,import_extension=none"
// @generated from file google/rpc/error_details.proto (package google.rpc, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Duration, Message, proto3 } from "@bufbuild/protobuf";

/**
 * Describes the cause of the error with structured details.
 *
 * Example of an error when contacting the "pubsub.googleapis.com" API when it
 * is not enabled:
 *
 *     { "reason": "API_DISABLED"
 *       "domain": "googleapis.com"
 *       "metadata": {
 *         "resource": "projects/123",
 *         "service": "pubsub.googleapis.com"
 *       }
 *     }
 *
 * This response indicates that the pubsub.googleapis.com API is not enabled.
 *
 * Example of an error that is returned when attempting to create a Spanner
 * instance in a region that is out of stock:
 *
 *     { "reason": "STOCKOUT"
 *       "domain": "spanner.googleapis.com",
 *       "metadata": {
 *         "availableRegions": "us-central1,us-east2"
 *       }
 *     }
 *
 * @generated from message google.rpc.ErrorInfo
 */
export class ErrorInfo extends Message<ErrorInfo> {
  /**
   * The reason of the error. This is a constant value that identifies the
   * proximate cause of the error. Error reasons are unique within a particular
   * domain of errors. This should be at most 63 characters and match a
   * regular expression of `[A-Z][A-Z0-9_]+[A-Z0-9]`, which represents
   * UPPER_SNAKE_CASE.
   *
   * @generated from field: string reason = 1;
   */
  reason = "";

  /**
   * The logical grouping to which the "reason" belongs. The error domain
   * is typically the registered service name of the tool or product that
   * generates the error. Example: "pubsub.googleapis.com". If the error is
   * generated by some common infrastructure, the error domain must be a
   * globally unique value that identifies the infrastructure. For Google API
   * infrastructure, the error domain is "googleapis.com".
   *
   * @generated from field: string domain = 2;
   */
  domain = "";

  /**
   * Additional structured details about this error.
   *
   * Keys should match /[a-zA-Z0-9-_]/ and be limited to 64 characters in
   * length. When identifying the current value of an exceeded limit, the units
   * should be contained in the key, not the value.  For example, rather than
   * {"instanceLimit": "100/request"}, should be returned as,
   * {"instanceLimitPerRequest": "100"}, if the client exceeds the number of
   * instances that can be created in a single (batch) request.
   *
   * @generated from field: map<string, string> metadata = 3;
   */
  metadata: { [key: string]: string } = {};

  constructor(data?: PartialMessage<ErrorInfo>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.rpc.ErrorInfo";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "reason", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "domain", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "metadata", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 9 /* ScalarType.STRING */} },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ErrorInfo {
    return new ErrorInfo().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ErrorInfo {
    return new ErrorInfo().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ErrorInfo {
    return new ErrorInfo().fromJsonString(jsonString, options);
  }

  static equals(a: ErrorInfo | PlainMessage<ErrorInfo> | undefined, b: ErrorInfo | PlainMessage<ErrorInfo> | undefined): boolean {
    return proto3.util.equals(ErrorInfo, a, b);
  }
}

/**
 * Describes when the clients can retry a failed request. Clients could ignore
 * the recommendation here or retry when this information is missing from error
 * responses.
 *
 * It's always recommended that clients should use exponential backoff when
 * retrying.
 *
 * Clients should wait until `retry_delay` amount of time has passed since
 * receiving the error response before retrying.  If retrying requests also
 * fail, clients should use an exponential backoff scheme to gradually increase
 * the delay between retries based on `retry_delay`, until either a maximum
 * number of retries have been reached or a maximum retry delay cap has been
 * reached.
 *
 * @generated from message google.rpc.RetryInfo
 */
export class RetryInfo extends Message<RetryInfo> {
  /**
   * Clients should wait at least this long between retrying the same request.
   *
   * @generated from field: google.protobuf.Duration retry_delay = 1;
   */
  retryDelay?: Duration;

  constructor(data?: PartialMessage<RetryInfo>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.rpc.RetryInfo";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "retry_delay", kind: "message", T: Duration },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): RetryInfo {
    return new RetryInfo().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): RetryInfo {
    return new RetryInfo().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): RetryInfo {
    return new RetryInfo().fromJsonString(jsonString, options);
  }

  static equals(a: RetryInfo | PlainMessage<RetryInfo> | undefined, b: RetryInfo | PlainMessage<RetryInfo> | undefined): boolean {
    return proto3.util.equals(RetryInfo, a, b);
  }
}

/**
 * Describes additional debugging info.
 *
 * @generated from message google.rpc.DebugInfo
 */
export class DebugInfo extends Message<DebugInfo> {
  /**
   * The stack trace entries indicating where the error occurred.
   *
   * @generated from field: repeated string stack_entries = 1;
   */
  stackEntries: string[] = [];

  /**
   * Additional debugging information provided by the server.
   *
   * @generated from field: string detail = 2;
   */
  detail = "";

  constructor(data?: PartialMessage<DebugInfo>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.rpc.DebugInfo";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "stack_entries", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
    { no: 2, name: "detail", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DebugInfo {
    return new DebugInfo().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DebugInfo {
    return new DebugInfo().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DebugInfo {
    return new DebugInfo().fromJsonString(jsonString, options);
  }

  static equals(a: DebugInfo | PlainMessage<DebugInfo> | undefined, b: DebugInfo | PlainMessage<DebugInfo> | undefined): boolean {
    return proto3.util.equals(DebugInfo, a, b);
  }
}

/**
 * Describes how a quota check failed.
 *
 * For example if a daily limit was exceeded for the calling project,
 * a service could respond with a QuotaFailure detail containing the project
 * id and the description of the quota limit that was exceeded.  If the
 * calling project hasn't enabled the service in the developer console, then
 * a service could respond with the project id and set `service_disabled`
 * to true.
 *
 * Also see RetryInfo and Help types for other details about handling a
 * quota failure.
 *
 * @generated from message google.rpc.QuotaFailure
 */
export class QuotaFailure extends Message<QuotaFailure> {
  /**
   * Describes all quota violations.
   *
   * @generated from field: repeated google.rpc.QuotaFailure.Violation violations = 1;
   */
  violations: QuotaFailure_Violation[] = [];

  constructor(data?: PartialMessage<QuotaFailure>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.rpc.QuotaFailure";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "violations", kind: "message", T: QuotaFailure_Violation, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): QuotaFailure {
    return new QuotaFailure().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): QuotaFailure {
    return new QuotaFailure().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): QuotaFailure {
    return new QuotaFailure().fromJsonString(jsonString, options);
  }

  static equals(a: QuotaFailure | PlainMessage<QuotaFailure> | undefined, b: QuotaFailure | PlainMessage<QuotaFailure> | undefined): boolean {
    return proto3.util.equals(QuotaFailure, a, b);
  }
}

/**
 * A message type used to describe a single quota violation.  For example, a
 * daily quota or a custom quota that was exceeded.
 *
 * @generated from message google.rpc.QuotaFailure.Violation
 */
export class QuotaFailure_Violation extends Message<QuotaFailure_Violation> {
  /**
   * The subject on which the quota check failed.
   * For example, "clientip:<ip address of client>" or "project:<Google
   * developer project id>".
   *
   * @generated from field: string subject = 1;
   */
  subject = "";

  /**
   * A description of how the quota check failed. Clients can use this
   * description to find more about the quota configuration in the service's
   * public documentation, or find the relevant quota limit to adjust through
   * developer console.
   *
   * For example: "Service disabled" or "Daily Limit for read operations
   * exceeded".
   *
   * @generated from field: string description = 2;
   */
  description = "";

  constructor(data?: PartialMessage<QuotaFailure_Violation>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.rpc.QuotaFailure.Violation";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "subject", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "description", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): QuotaFailure_Violation {
    return new QuotaFailure_Violation().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): QuotaFailure_Violation {
    return new QuotaFailure_Violation().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): QuotaFailure_Violation {
    return new QuotaFailure_Violation().fromJsonString(jsonString, options);
  }

  static equals(a: QuotaFailure_Violation | PlainMessage<QuotaFailure_Violation> | undefined, b: QuotaFailure_Violation | PlainMessage<QuotaFailure_Violation> | undefined): boolean {
    return proto3.util.equals(QuotaFailure_Violation, a, b);
  }
}

/**
 * Describes what preconditions have failed.
 *
 * For example, if an RPC failed because it required the Terms of Service to be
 * acknowledged, it could list the terms of service violation in the
 * PreconditionFailure message.
 *
 * @generated from message google.rpc.PreconditionFailure
 */
export class PreconditionFailure extends Message<PreconditionFailure> {
  /**
   * Describes all precondition violations.
   *
   * @generated from field: repeated google.rpc.PreconditionFailure.Violation violations = 1;
   */
  violations: PreconditionFailure_Violation[] = [];

  constructor(data?: PartialMessage<PreconditionFailure>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.rpc.PreconditionFailure";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "violations", kind: "message", T: PreconditionFailure_Violation, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PreconditionFailure {
    return new PreconditionFailure().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PreconditionFailure {
    return new PreconditionFailure().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PreconditionFailure {
    return new PreconditionFailure().fromJsonString(jsonString, options);
  }

  static equals(a: PreconditionFailure | PlainMessage<PreconditionFailure> | undefined, b: PreconditionFailure | PlainMessage<PreconditionFailure> | undefined): boolean {
    return proto3.util.equals(PreconditionFailure, a, b);
  }
}

/**
 * A message type used to describe a single precondition failure.
 *
 * @generated from message google.rpc.PreconditionFailure.Violation
 */
export class PreconditionFailure_Violation extends Message<PreconditionFailure_Violation> {
  /**
   * The type of PreconditionFailure. We recommend using a service-specific
   * enum type to define the supported precondition violation subjects. For
   * example, "TOS" for "Terms of Service violation".
   *
   * @generated from field: string type = 1;
   */
  type = "";

  /**
   * The subject, relative to the type, that failed.
   * For example, "google.com/cloud" relative to the "TOS" type would indicate
   * which terms of service is being referenced.
   *
   * @generated from field: string subject = 2;
   */
  subject = "";

  /**
   * A description of how the precondition failed. Developers can use this
   * description to understand how to fix the failure.
   *
   * For example: "Terms of service not accepted".
   *
   * @generated from field: string description = 3;
   */
  description = "";

  constructor(data?: PartialMessage<PreconditionFailure_Violation>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.rpc.PreconditionFailure.Violation";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "type", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "subject", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "description", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PreconditionFailure_Violation {
    return new PreconditionFailure_Violation().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PreconditionFailure_Violation {
    return new PreconditionFailure_Violation().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PreconditionFailure_Violation {
    return new PreconditionFailure_Violation().fromJsonString(jsonString, options);
  }

  static equals(a: PreconditionFailure_Violation | PlainMessage<PreconditionFailure_Violation> | undefined, b: PreconditionFailure_Violation | PlainMessage<PreconditionFailure_Violation> | undefined): boolean {
    return proto3.util.equals(PreconditionFailure_Violation, a, b);
  }
}

/**
 * Describes violations in a client request. This error type focuses on the
 * syntactic aspects of the request.
 *
 * @generated from message google.rpc.BadRequest
 */
export class BadRequest extends Message<BadRequest> {
  /**
   * Describes all violations in a client request.
   *
   * @generated from field: repeated google.rpc.BadRequest.FieldViolation field_violations = 1;
   */
  fieldViolations: BadRequest_FieldViolation[] = [];

  constructor(data?: PartialMessage<BadRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.rpc.BadRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "field_violations", kind: "message", T: BadRequest_FieldViolation, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): BadRequest {
    return new BadRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): BadRequest {
    return new BadRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): BadRequest {
    return new BadRequest().fromJsonString(jsonString, options);
  }

  static equals(a: BadRequest | PlainMessage<BadRequest> | undefined, b: BadRequest | PlainMessage<BadRequest> | undefined): boolean {
    return proto3.util.equals(BadRequest, a, b);
  }
}

/**
 * A message type used to describe a single bad request field.
 *
 * @generated from message google.rpc.BadRequest.FieldViolation
 */
export class BadRequest_FieldViolation extends Message<BadRequest_FieldViolation> {
  /**
   * A path that leads to a field in the request body. The value will be a
   * sequence of dot-separated identifiers that identify a protocol buffer
   * field.
   *
   * Consider the following:
   *
   *     message CreateContactRequest {
   *       message EmailAddress {
   *         enum Type {
   *           TYPE_UNSPECIFIED = 0;
   *           HOME = 1;
   *           WORK = 2;
   *         }
   *
   *         optional string email = 1;
   *         repeated EmailType type = 2;
   *       }
   *
   *       string full_name = 1;
   *       repeated EmailAddress email_addresses = 2;
   *     }
   *
   * In this example, in proto `field` could take one of the following values:
   *
   * * `full_name` for a violation in the `full_name` value
   * * `email_addresses[1].email` for a violation in the `email` field of the
   *   first `email_addresses` message
   * * `email_addresses[3].type[2]` for a violation in the second `type`
   *   value in the third `email_addresses` message.
   *
   * In JSON, the same values are represented as:
   *
   * * `fullName` for a violation in the `fullName` value
   * * `emailAddresses[1].email` for a violation in the `email` field of the
   *   first `emailAddresses` message
   * * `emailAddresses[3].type[2]` for a violation in the second `type`
   *   value in the third `emailAddresses` message.
   *
   * @generated from field: string field = 1;
   */
  field = "";

  /**
   * A description of why the request element is bad.
   *
   * @generated from field: string description = 2;
   */
  description = "";

  constructor(data?: PartialMessage<BadRequest_FieldViolation>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.rpc.BadRequest.FieldViolation";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "field", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "description", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): BadRequest_FieldViolation {
    return new BadRequest_FieldViolation().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): BadRequest_FieldViolation {
    return new BadRequest_FieldViolation().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): BadRequest_FieldViolation {
    return new BadRequest_FieldViolation().fromJsonString(jsonString, options);
  }

  static equals(a: BadRequest_FieldViolation | PlainMessage<BadRequest_FieldViolation> | undefined, b: BadRequest_FieldViolation | PlainMessage<BadRequest_FieldViolation> | undefined): boolean {
    return proto3.util.equals(BadRequest_FieldViolation, a, b);
  }
}

/**
 * Contains metadata about the request that clients can attach when filing a bug
 * or providing other forms of feedback.
 *
 * @generated from message google.rpc.RequestInfo
 */
export class RequestInfo extends Message<RequestInfo> {
  /**
   * An opaque string that should only be interpreted by the service generating
   * it. For example, it can be used to identify requests in the service's logs.
   *
   * @generated from field: string request_id = 1;
   */
  requestId = "";

  /**
   * Any data that was used to serve this request. For example, an encrypted
   * stack trace that can be sent back to the service provider for debugging.
   *
   * @generated from field: string serving_data = 2;
   */
  servingData = "";

  constructor(data?: PartialMessage<RequestInfo>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.rpc.RequestInfo";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "request_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "serving_data", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): RequestInfo {
    return new RequestInfo().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): RequestInfo {
    return new RequestInfo().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): RequestInfo {
    return new RequestInfo().fromJsonString(jsonString, options);
  }

  static equals(a: RequestInfo | PlainMessage<RequestInfo> | undefined, b: RequestInfo | PlainMessage<RequestInfo> | undefined): boolean {
    return proto3.util.equals(RequestInfo, a, b);
  }
}

/**
 * Describes the resource that is being accessed.
 *
 * @generated from message google.rpc.ResourceInfo
 */
export class ResourceInfo extends Message<ResourceInfo> {
  /**
   * A name for the type of resource being accessed, e.g. "sql table",
   * "cloud storage bucket", "file", "Google calendar"; or the type URL
   * of the resource: e.g. "type.googleapis.com/google.pubsub.v1.Topic".
   *
   * @generated from field: string resource_type = 1;
   */
  resourceType = "";

  /**
   * The name of the resource being accessed.  For example, a shared calendar
   * name: "example.com_4fghdhgsrgh@group.calendar.google.com", if the current
   * error is
   * [google.rpc.Code.PERMISSION_DENIED][google.rpc.Code.PERMISSION_DENIED].
   *
   * @generated from field: string resource_name = 2;
   */
  resourceName = "";

  /**
   * The owner of the resource (optional).
   * For example, "user:<owner email>" or "project:<Google developer project
   * id>".
   *
   * @generated from field: string owner = 3;
   */
  owner = "";

  /**
   * Describes what error is encountered when accessing this resource.
   * For example, updating a cloud project may require the `writer` permission
   * on the developer console project.
   *
   * @generated from field: string description = 4;
   */
  description = "";

  constructor(data?: PartialMessage<ResourceInfo>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.rpc.ResourceInfo";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "resource_type", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "resource_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "owner", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "description", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ResourceInfo {
    return new ResourceInfo().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ResourceInfo {
    return new ResourceInfo().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ResourceInfo {
    return new ResourceInfo().fromJsonString(jsonString, options);
  }

  static equals(a: ResourceInfo | PlainMessage<ResourceInfo> | undefined, b: ResourceInfo | PlainMessage<ResourceInfo> | undefined): boolean {
    return proto3.util.equals(ResourceInfo, a, b);
  }
}

/**
 * Provides links to documentation or for performing an out of band action.
 *
 * For example, if a quota check failed with an error indicating the calling
 * project hasn't enabled the accessed service, this can contain a URL pointing
 * directly to the right place in the developer console to flip the bit.
 *
 * @generated from message google.rpc.Help
 */
export class Help extends Message<Help> {
  /**
   * URL(s) pointing to additional information on handling the current error.
   *
   * @generated from field: repeated google.rpc.Help.Link links = 1;
   */
  links: Help_Link[] = [];

  constructor(data?: PartialMessage<Help>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.rpc.Help";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "links", kind: "message", T: Help_Link, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Help {
    return new Help().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Help {
    return new Help().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Help {
    return new Help().fromJsonString(jsonString, options);
  }

  static equals(a: Help | PlainMessage<Help> | undefined, b: Help | PlainMessage<Help> | undefined): boolean {
    return proto3.util.equals(Help, a, b);
  }
}

/**
 * Describes a URL link.
 *
 * @generated from message google.rpc.Help.Link
 */
export class Help_Link extends Message<Help_Link> {
  /**
   * Describes what the link offers.
   *
   * @generated from field: string description = 1;
   */
  description = "";

  /**
   * The URL of the link.
   *
   * @generated from field: string url = 2;
   */
  url = "";

  constructor(data?: PartialMessage<Help_Link>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.rpc.Help.Link";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "description", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "url", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Help_Link {
    return new Help_Link().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Help_Link {
    return new Help_Link().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Help_Link {
    return new Help_Link().fromJsonString(jsonString, options);
  }

  static equals(a: Help_Link | PlainMessage<Help_Link> | undefined, b: Help_Link | PlainMessage<Help_Link> | undefined): boolean {
    return proto3.util.equals(Help_Link, a, b);
  }
}

/**
 * Provides a localized error message that is safe to return to the user
 * which can be attached to an RPC error.
 *
 * @generated from message google.rpc.LocalizedMessage
 */
export class LocalizedMessage extends Message<LocalizedMessage> {
  /**
   * The locale used following the specification defined at
   * https://www.rfc-editor.org/rfc/bcp/bcp47.txt.
   * Examples are: "en-US", "fr-CH", "es-MX"
   *
   * @generated from field: string locale = 1;
   */
  locale = "";

  /**
   * The localized error message in the above locale.
   *
   * @generated from field: string message = 2;
   */
  message = "";

  constructor(data?: PartialMessage<LocalizedMessage>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.rpc.LocalizedMessage";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "locale", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "message", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): LocalizedMessage {
    return new LocalizedMessage().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): LocalizedMessage {
    return new LocalizedMessage().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): LocalizedMessage {
    return new LocalizedMessage().fromJsonString(jsonString, options);
  }

  static equals(a: LocalizedMessage | PlainMessage<LocalizedMessage> | undefined, b: LocalizedMessage | PlainMessage<LocalizedMessage> | undefined): boolean {
    return proto3.util.equals(LocalizedMessage, a, b);
  }
}

