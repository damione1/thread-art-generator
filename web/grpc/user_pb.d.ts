import * as jspb from 'google-protobuf'

import * as google_protobuf_timestamp_pb from 'google-protobuf/google/protobuf/timestamp_pb'; // proto import: "google/protobuf/timestamp.proto"
import * as google_protobuf_field_mask_pb from 'google-protobuf/google/protobuf/field_mask_pb'; // proto import: "google/protobuf/field_mask.proto"
import * as google_api_field_behavior_pb from './google/api/field_behavior_pb'; // proto import: "google/api/field_behavior.proto"
import * as google_api_resource_pb from './google/api/resource_pb'; // proto import: "google/api/resource.proto"


export class User extends jspb.Message {
  getName(): string;
  setName(value: string): User;

  getFirstName(): string;
  setFirstName(value: string): User;

  getLastName(): string;
  setLastName(value: string): User;

  getEmail(): string;
  setEmail(value: string): User;

  getPassword(): string;
  setPassword(value: string): User;

  getAvatar(): string;
  setAvatar(value: string): User;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): User.AsObject;
  static toObject(includeInstance: boolean, msg: User): User.AsObject;
  static serializeBinaryToWriter(message: User, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): User;
  static deserializeBinaryFromReader(message: User, reader: jspb.BinaryReader): User;
}

export namespace User {
  export type AsObject = {
    name: string,
    firstName: string,
    lastName: string,
    email: string,
    password: string,
    avatar: string,
  }
}

export class CreateUserRequest extends jspb.Message {
  getUser(): User | undefined;
  setUser(value?: User): CreateUserRequest;
  hasUser(): boolean;
  clearUser(): CreateUserRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateUserRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CreateUserRequest): CreateUserRequest.AsObject;
  static serializeBinaryToWriter(message: CreateUserRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateUserRequest;
  static deserializeBinaryFromReader(message: CreateUserRequest, reader: jspb.BinaryReader): CreateUserRequest;
}

export namespace CreateUserRequest {
  export type AsObject = {
    user?: User.AsObject,
  }
}

export class GetUserRequest extends jspb.Message {
  getName(): string;
  setName(value: string): GetUserRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetUserRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetUserRequest): GetUserRequest.AsObject;
  static serializeBinaryToWriter(message: GetUserRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetUserRequest;
  static deserializeBinaryFromReader(message: GetUserRequest, reader: jspb.BinaryReader): GetUserRequest;
}

export namespace GetUserRequest {
  export type AsObject = {
    name: string,
  }
}

export class UpdateUserRequest extends jspb.Message {
  getUser(): User | undefined;
  setUser(value?: User): UpdateUserRequest;
  hasUser(): boolean;
  clearUser(): UpdateUserRequest;

  getUpdateMask(): google_protobuf_field_mask_pb.FieldMask | undefined;
  setUpdateMask(value?: google_protobuf_field_mask_pb.FieldMask): UpdateUserRequest;
  hasUpdateMask(): boolean;
  clearUpdateMask(): UpdateUserRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UpdateUserRequest.AsObject;
  static toObject(includeInstance: boolean, msg: UpdateUserRequest): UpdateUserRequest.AsObject;
  static serializeBinaryToWriter(message: UpdateUserRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UpdateUserRequest;
  static deserializeBinaryFromReader(message: UpdateUserRequest, reader: jspb.BinaryReader): UpdateUserRequest;
}

export namespace UpdateUserRequest {
  export type AsObject = {
    user?: User.AsObject,
    updateMask?: google_protobuf_field_mask_pb.FieldMask.AsObject,
  }
}

export class ListUsersRequest extends jspb.Message {
  getPageToken(): string;
  setPageToken(value: string): ListUsersRequest;

  getPageSize(): number;
  setPageSize(value: number): ListUsersRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ListUsersRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ListUsersRequest): ListUsersRequest.AsObject;
  static serializeBinaryToWriter(message: ListUsersRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ListUsersRequest;
  static deserializeBinaryFromReader(message: ListUsersRequest, reader: jspb.BinaryReader): ListUsersRequest;
}

export namespace ListUsersRequest {
  export type AsObject = {
    pageToken: string,
    pageSize: number,
  }
}

export class ListUsersResponse extends jspb.Message {
  getUsersList(): Array<User>;
  setUsersList(value: Array<User>): ListUsersResponse;
  clearUsersList(): ListUsersResponse;
  addUsers(value?: User, index?: number): User;

  getNextPageToken(): string;
  setNextPageToken(value: string): ListUsersResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ListUsersResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ListUsersResponse): ListUsersResponse.AsObject;
  static serializeBinaryToWriter(message: ListUsersResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ListUsersResponse;
  static deserializeBinaryFromReader(message: ListUsersResponse, reader: jspb.BinaryReader): ListUsersResponse;
}

export namespace ListUsersResponse {
  export type AsObject = {
    usersList: Array<User.AsObject>,
    nextPageToken: string,
  }
}

export class DeleteUserRequest extends jspb.Message {
  getName(): string;
  setName(value: string): DeleteUserRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DeleteUserRequest.AsObject;
  static toObject(includeInstance: boolean, msg: DeleteUserRequest): DeleteUserRequest.AsObject;
  static serializeBinaryToWriter(message: DeleteUserRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DeleteUserRequest;
  static deserializeBinaryFromReader(message: DeleteUserRequest, reader: jspb.BinaryReader): DeleteUserRequest;
}

export namespace DeleteUserRequest {
  export type AsObject = {
    name: string,
  }
}

export class CreateSessionRequest extends jspb.Message {
  getEmail(): string;
  setEmail(value: string): CreateSessionRequest;

  getPassword(): string;
  setPassword(value: string): CreateSessionRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateSessionRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CreateSessionRequest): CreateSessionRequest.AsObject;
  static serializeBinaryToWriter(message: CreateSessionRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateSessionRequest;
  static deserializeBinaryFromReader(message: CreateSessionRequest, reader: jspb.BinaryReader): CreateSessionRequest;
}

export namespace CreateSessionRequest {
  export type AsObject = {
    email: string,
    password: string,
  }
}

export class CreateSessionResponse extends jspb.Message {
  getUser(): User | undefined;
  setUser(value?: User): CreateSessionResponse;
  hasUser(): boolean;
  clearUser(): CreateSessionResponse;

  getSessionId(): string;
  setSessionId(value: string): CreateSessionResponse;

  getAccessToken(): string;
  setAccessToken(value: string): CreateSessionResponse;

  getRefreshToken(): string;
  setRefreshToken(value: string): CreateSessionResponse;

  getAccessTokenExpireTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setAccessTokenExpireTime(value?: google_protobuf_timestamp_pb.Timestamp): CreateSessionResponse;
  hasAccessTokenExpireTime(): boolean;
  clearAccessTokenExpireTime(): CreateSessionResponse;

  getRefreshTokenExpireTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setRefreshTokenExpireTime(value?: google_protobuf_timestamp_pb.Timestamp): CreateSessionResponse;
  hasRefreshTokenExpireTime(): boolean;
  clearRefreshTokenExpireTime(): CreateSessionResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateSessionResponse.AsObject;
  static toObject(includeInstance: boolean, msg: CreateSessionResponse): CreateSessionResponse.AsObject;
  static serializeBinaryToWriter(message: CreateSessionResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateSessionResponse;
  static deserializeBinaryFromReader(message: CreateSessionResponse, reader: jspb.BinaryReader): CreateSessionResponse;
}

export namespace CreateSessionResponse {
  export type AsObject = {
    user?: User.AsObject,
    sessionId: string,
    accessToken: string,
    refreshToken: string,
    accessTokenExpireTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    refreshTokenExpireTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
  }
}

export class RefreshTokenRequest extends jspb.Message {
  getRefreshToken(): string;
  setRefreshToken(value: string): RefreshTokenRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RefreshTokenRequest.AsObject;
  static toObject(includeInstance: boolean, msg: RefreshTokenRequest): RefreshTokenRequest.AsObject;
  static serializeBinaryToWriter(message: RefreshTokenRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RefreshTokenRequest;
  static deserializeBinaryFromReader(message: RefreshTokenRequest, reader: jspb.BinaryReader): RefreshTokenRequest;
}

export namespace RefreshTokenRequest {
  export type AsObject = {
    refreshToken: string,
  }
}

export class RefreshTokenResponse extends jspb.Message {
  getAccessToken(): string;
  setAccessToken(value: string): RefreshTokenResponse;

  getRefreshToken(): string;
  setRefreshToken(value: string): RefreshTokenResponse;

  getAccessTokenExpireTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setAccessTokenExpireTime(value?: google_protobuf_timestamp_pb.Timestamp): RefreshTokenResponse;
  hasAccessTokenExpireTime(): boolean;
  clearAccessTokenExpireTime(): RefreshTokenResponse;

  getRefreshTokenExpireTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setRefreshTokenExpireTime(value?: google_protobuf_timestamp_pb.Timestamp): RefreshTokenResponse;
  hasRefreshTokenExpireTime(): boolean;
  clearRefreshTokenExpireTime(): RefreshTokenResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RefreshTokenResponse.AsObject;
  static toObject(includeInstance: boolean, msg: RefreshTokenResponse): RefreshTokenResponse.AsObject;
  static serializeBinaryToWriter(message: RefreshTokenResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RefreshTokenResponse;
  static deserializeBinaryFromReader(message: RefreshTokenResponse, reader: jspb.BinaryReader): RefreshTokenResponse;
}

export namespace RefreshTokenResponse {
  export type AsObject = {
    accessToken: string,
    refreshToken: string,
    accessTokenExpireTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    refreshTokenExpireTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
  }
}

export class DeleteSessionRequest extends jspb.Message {
  getRefreshToken(): string;
  setRefreshToken(value: string): DeleteSessionRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DeleteSessionRequest.AsObject;
  static toObject(includeInstance: boolean, msg: DeleteSessionRequest): DeleteSessionRequest.AsObject;
  static serializeBinaryToWriter(message: DeleteSessionRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DeleteSessionRequest;
  static deserializeBinaryFromReader(message: DeleteSessionRequest, reader: jspb.BinaryReader): DeleteSessionRequest;
}

export namespace DeleteSessionRequest {
  export type AsObject = {
    refreshToken: string,
  }
}

export class ResetPasswordRequest extends jspb.Message {
  getEmail(): string;
  setEmail(value: string): ResetPasswordRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ResetPasswordRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ResetPasswordRequest): ResetPasswordRequest.AsObject;
  static serializeBinaryToWriter(message: ResetPasswordRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ResetPasswordRequest;
  static deserializeBinaryFromReader(message: ResetPasswordRequest, reader: jspb.BinaryReader): ResetPasswordRequest;
}

export namespace ResetPasswordRequest {
  export type AsObject = {
    email: string,
  }
}

export class ValidateEmailRequest extends jspb.Message {
  getEmail(): string;
  setEmail(value: string): ValidateEmailRequest;

  getValidationnumber(): number;
  setValidationnumber(value: number): ValidateEmailRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ValidateEmailRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ValidateEmailRequest): ValidateEmailRequest.AsObject;
  static serializeBinaryToWriter(message: ValidateEmailRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ValidateEmailRequest;
  static deserializeBinaryFromReader(message: ValidateEmailRequest, reader: jspb.BinaryReader): ValidateEmailRequest;
}

export namespace ValidateEmailRequest {
  export type AsObject = {
    email: string,
    validationnumber: number,
  }
}

export class SendValidationEmailRequest extends jspb.Message {
  getEmail(): string;
  setEmail(value: string): SendValidationEmailRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SendValidationEmailRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SendValidationEmailRequest): SendValidationEmailRequest.AsObject;
  static serializeBinaryToWriter(message: SendValidationEmailRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SendValidationEmailRequest;
  static deserializeBinaryFromReader(message: SendValidationEmailRequest, reader: jspb.BinaryReader): SendValidationEmailRequest;
}

export namespace SendValidationEmailRequest {
  export type AsObject = {
    email: string,
  }
}

