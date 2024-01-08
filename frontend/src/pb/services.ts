/* eslint-disable */
import * as _m0 from "protobufjs/minimal";
import {
  CreateArtRequest,
  CreateArtResponse,
  DeleteArtRequest,
  DeleteArtResponse,
  GetArtRequest,
  GetArtResponse,
  ListArtRequest,
  ListArtResponse,
  UpdateArtRequest,
  UpdateArtResponse,
} from "./post";
import {
  ChangePasswordRequest,
  ChangePasswordResponse,
  CreateUserRequest,
  CreateUserResponse,
  GetUserRequest,
  GetUserResponse,
  LoginRequest,
  LoginResponse,
  LogoutRequest,
  LogoutResponse,
  RefreshTokenRequest,
  RefreshTokenResponse,
  ResetPasswordRequest,
  ResetPasswordResponse,
  UpdateUserRequest,
  UpdateUserResponse,
} from "./user";

export const protobufPackage = "pb";

export interface ArtGeneratorService {
  LoginUser(request: LoginRequest): Promise<LoginResponse>;
  LogoutUser(request: LogoutRequest): Promise<LogoutResponse>;
  RefreshToken(request: RefreshTokenRequest): Promise<RefreshTokenResponse>;
  CreateUser(request: CreateUserRequest): Promise<CreateUserResponse>;
  UpdateUser(request: UpdateUserRequest): Promise<UpdateUserResponse>;
  GetUser(request: GetUserRequest): Promise<GetUserResponse>;
  ResetPassword(request: ResetPasswordRequest): Promise<ResetPasswordResponse>;
  ChangePassword(request: ChangePasswordRequest): Promise<ChangePasswordResponse>;
  CreateArt(request: CreateArtRequest): Promise<CreateArtResponse>;
  UpdateArt(request: UpdateArtRequest): Promise<UpdateArtResponse>;
  GetArt(request: GetArtRequest): Promise<GetArtResponse>;
  ListArts(request: ListArtRequest): Promise<ListArtResponse>;
  DeleteArt(request: DeleteArtRequest): Promise<DeleteArtResponse>;
}

export const ArtGeneratorServiceServiceName = "pb.ArtGeneratorService";
export class ArtGeneratorServiceClientImpl implements ArtGeneratorService {
  private readonly rpc: Rpc;
  private readonly service: string;
  constructor(rpc: Rpc, opts?: { service?: string }) {
    this.service = opts?.service || ArtGeneratorServiceServiceName;
    this.rpc = rpc;
    this.LoginUser = this.LoginUser.bind(this);
    this.LogoutUser = this.LogoutUser.bind(this);
    this.RefreshToken = this.RefreshToken.bind(this);
    this.CreateUser = this.CreateUser.bind(this);
    this.UpdateUser = this.UpdateUser.bind(this);
    this.GetUser = this.GetUser.bind(this);
    this.ResetPassword = this.ResetPassword.bind(this);
    this.ChangePassword = this.ChangePassword.bind(this);
    this.CreateArt = this.CreateArt.bind(this);
    this.UpdateArt = this.UpdateArt.bind(this);
    this.GetArt = this.GetArt.bind(this);
    this.ListArts = this.ListArts.bind(this);
    this.DeleteArt = this.DeleteArt.bind(this);
  }
  LoginUser(request: LoginRequest): Promise<LoginResponse> {
    const data = LoginRequest.encode(request).finish();
    const promise = this.rpc.request(this.service, "LoginUser", data);
    return promise.then((data) => LoginResponse.decode(_m0.Reader.create(data)));
  }

  LogoutUser(request: LogoutRequest): Promise<LogoutResponse> {
    const data = LogoutRequest.encode(request).finish();
    const promise = this.rpc.request(this.service, "LogoutUser", data);
    return promise.then((data) => LogoutResponse.decode(_m0.Reader.create(data)));
  }

  RefreshToken(request: RefreshTokenRequest): Promise<RefreshTokenResponse> {
    const data = RefreshTokenRequest.encode(request).finish();
    const promise = this.rpc.request(this.service, "RefreshToken", data);
    return promise.then((data) => RefreshTokenResponse.decode(_m0.Reader.create(data)));
  }

  CreateUser(request: CreateUserRequest): Promise<CreateUserResponse> {
    const data = CreateUserRequest.encode(request).finish();
    const promise = this.rpc.request(this.service, "CreateUser", data);
    return promise.then((data) => CreateUserResponse.decode(_m0.Reader.create(data)));
  }

  UpdateUser(request: UpdateUserRequest): Promise<UpdateUserResponse> {
    const data = UpdateUserRequest.encode(request).finish();
    const promise = this.rpc.request(this.service, "UpdateUser", data);
    return promise.then((data) => UpdateUserResponse.decode(_m0.Reader.create(data)));
  }

  GetUser(request: GetUserRequest): Promise<GetUserResponse> {
    const data = GetUserRequest.encode(request).finish();
    const promise = this.rpc.request(this.service, "GetUser", data);
    return promise.then((data) => GetUserResponse.decode(_m0.Reader.create(data)));
  }

  ResetPassword(request: ResetPasswordRequest): Promise<ResetPasswordResponse> {
    const data = ResetPasswordRequest.encode(request).finish();
    const promise = this.rpc.request(this.service, "ResetPassword", data);
    return promise.then((data) => ResetPasswordResponse.decode(_m0.Reader.create(data)));
  }

  ChangePassword(request: ChangePasswordRequest): Promise<ChangePasswordResponse> {
    const data = ChangePasswordRequest.encode(request).finish();
    const promise = this.rpc.request(this.service, "ChangePassword", data);
    return promise.then((data) => ChangePasswordResponse.decode(_m0.Reader.create(data)));
  }

  CreateArt(request: CreateArtRequest): Promise<CreateArtResponse> {
    const data = CreateArtRequest.encode(request).finish();
    const promise = this.rpc.request(this.service, "CreateArt", data);
    return promise.then((data) => CreateArtResponse.decode(_m0.Reader.create(data)));
  }

  UpdateArt(request: UpdateArtRequest): Promise<UpdateArtResponse> {
    const data = UpdateArtRequest.encode(request).finish();
    const promise = this.rpc.request(this.service, "UpdateArt", data);
    return promise.then((data) => UpdateArtResponse.decode(_m0.Reader.create(data)));
  }

  GetArt(request: GetArtRequest): Promise<GetArtResponse> {
    const data = GetArtRequest.encode(request).finish();
    const promise = this.rpc.request(this.service, "GetArt", data);
    return promise.then((data) => GetArtResponse.decode(_m0.Reader.create(data)));
  }

  ListArts(request: ListArtRequest): Promise<ListArtResponse> {
    const data = ListArtRequest.encode(request).finish();
    const promise = this.rpc.request(this.service, "ListArts", data);
    return promise.then((data) => ListArtResponse.decode(_m0.Reader.create(data)));
  }

  DeleteArt(request: DeleteArtRequest): Promise<DeleteArtResponse> {
    const data = DeleteArtRequest.encode(request).finish();
    const promise = this.rpc.request(this.service, "DeleteArt", data);
    return promise.then((data) => DeleteArtResponse.decode(_m0.Reader.create(data)));
  }
}

interface Rpc {
  request(service: string, method: string, data: Uint8Array): Promise<Uint8Array>;
}
