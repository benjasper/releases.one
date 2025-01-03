// @generated by protoc-gen-es v2.2.3 with parameter "target=ts"
// @generated from file api/v1/api.proto (package api.v1, syntax proto3)
/* eslint-disable */

import type { GenFile, GenMessage, GenService } from "@bufbuild/protobuf/codegenv1";
import { fileDesc, messageDesc, serviceDesc } from "@bufbuild/protobuf/codegenv1";
import type { Timestamp } from "@bufbuild/protobuf/wkt";
import { file_google_protobuf_timestamp } from "@bufbuild/protobuf/wkt";
import type { Message } from "@bufbuild/protobuf";

/**
 * Describes the file api/v1/api.proto.
 */
export const file_api_v1_api: GenFile = /*@__PURE__*/
  fileDesc("ChBhcGkvdjEvYXBpLnByb3RvEgZhcGkudjEiTQoHUmVsZWFzZRIMCgRuYW1lGAEgASgJEhMKC2Rlc2NyaXB0aW9uGAIgASgJEg8KB3ZlcnNpb24YAyABKAkSDgoGYXV0aG9yGAQgASgJIk8KClJlcG9zaXRvcnkSDAoEbmFtZRgBIAEoCRITCgtkZXNjcmlwdGlvbhgCIAEoCRILCgN1cmwYAyABKAkSEQoJaW1hZ2VfdXJsGAQgASgJIpACCg1UaW1lbGluZUVudHJ5EgoKAmlkGAEgASgFEhUKDXJlcG9zaXRvcnlfaWQYAiABKAUSDAoEbmFtZRgDIAEoCRILCgN1cmwYBCABKAkSEAoIdGFnX25hbWUYBSABKAkSEwoLZGVzY3JpcHRpb24YBiABKAkSFQoNaXNfcHJlcmVsZWFzZRgHIAEoCBIvCgtyZWxlYXNlZF9hdBgIIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXASFwoPcmVwb3NpdG9yeV9uYW1lGAkgASgJEhEKCWltYWdlX3VybBgKIAEoCRIOCgZhdXRob3IYCyABKAkSFgoOcmVwb3NpdG9yeV91cmwYDCABKAkiHwoLU3luY1JlcXVlc3QSEAoIdXNlcm5hbWUYASABKAkiNwoMU3luY1Jlc3BvbnNlEicKCHRpbWVsaW5lGAEgAygLMhUuYXBpLnYxLlRpbWVsaW5lRW50cnkiGAoWR2V0UmVwb3NpdG9yaWVzUmVxdWVzdCJCChdHZXRSZXBvc2l0b3JpZXNSZXNwb25zZRInCgh0aW1lbGluZRgBIAMoCzIVLmFwaS52MS5UaW1lbGluZUVudHJ5Ii4KG1Rvb2dsZVVzZXJQdWJsaWNGZWVkUmVxdWVzdBIPCgdlbmFibGVkGAEgASgIIjEKHFRvb2dsZVVzZXJQdWJsaWNGZWVkUmVzcG9uc2USEQoJcHVibGljX2lkGAEgASgJIhIKEEdldE15VXNlclJlcXVlc3QihwEKEUdldE15VXNlclJlc3BvbnNlEgoKAmlkGAEgASgFEjIKDmxhc3Rfc3luY2VkX2F0GAIgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcBIRCglpc19wdWJsaWMYAyABKAgSEQoJcHVibGljX2lkGAQgASgJEgwKBG5hbWUYBSABKAkiDwoNTG9nb3V0UmVxdWVzdCIQCg5Mb2dvdXRSZXNwb25zZSIVChNSZWZyZXNoVG9rZW5SZXF1ZXN0Ir4BChRSZWZyZXNoVG9rZW5SZXNwb25zZRIUCgxhY2Nlc3NfdG9rZW4YASABKAkSFQoNcmVmcmVzaF90b2tlbhgCIAEoCRI7ChdhY2Nlc3NfdG9rZW5fZXhwaXJlc19hdBgDIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXASPAoYcmVmcmVzaF90b2tlbl9leHBpcmVzX2F0GAQgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcDLxAgoKQXBpU2VydmljZRIxCgRTeW5jEhMuYXBpLnYxLlN5bmNSZXF1ZXN0GhQuYXBpLnYxLlN5bmNSZXNwb25zZRJSCg9HZXRSZXBvc2l0b3JpZXMSHi5hcGkudjEuR2V0UmVwb3NpdG9yaWVzUmVxdWVzdBofLmFwaS52MS5HZXRSZXBvc2l0b3JpZXNSZXNwb25zZRJhChRUb29nbGVVc2VyUHVibGljRmVlZBIjLmFwaS52MS5Ub29nbGVVc2VyUHVibGljRmVlZFJlcXVlc3QaJC5hcGkudjEuVG9vZ2xlVXNlclB1YmxpY0ZlZWRSZXNwb25zZRJACglHZXRNeVVzZXISGC5hcGkudjEuR2V0TXlVc2VyUmVxdWVzdBoZLmFwaS52MS5HZXRNeVVzZXJSZXNwb25zZRI3CgZMb2dvdXQSFS5hcGkudjEuTG9nb3V0UmVxdWVzdBoWLmFwaS52MS5Mb2dvdXRSZXNwb25zZTJYCgtBdXRoU2VydmljZRJJCgxSZWZyZXNoVG9rZW4SGy5hcGkudjEuUmVmcmVzaFRva2VuUmVxdWVzdBocLmFwaS52MS5SZWZyZXNoVG9rZW5SZXNwb25zZUI9WjtnaXRodWIuY29tL2Jlbmphc3Blci9yZWxlYXNlcy5vbmUvaW50ZXJuYWwvZ2VuL2FwaS92MTthcGl2MWIGcHJvdG8z", [file_google_protobuf_timestamp]);

/**
 * @generated from message api.v1.Release
 */
export type Release = Message<"api.v1.Release"> & {
  /**
   * @generated from field: string name = 1;
   */
  name: string;

  /**
   * @generated from field: string description = 2;
   */
  description: string;

  /**
   * @generated from field: string version = 3;
   */
  version: string;

  /**
   * @generated from field: string author = 4;
   */
  author: string;
};

/**
 * Describes the message api.v1.Release.
 * Use `create(ReleaseSchema)` to create a new message.
 */
export const ReleaseSchema: GenMessage<Release> = /*@__PURE__*/
  messageDesc(file_api_v1_api, 0);

/**
 * @generated from message api.v1.Repository
 */
export type Repository = Message<"api.v1.Repository"> & {
  /**
   * @generated from field: string name = 1;
   */
  name: string;

  /**
   * @generated from field: string description = 2;
   */
  description: string;

  /**
   * @generated from field: string url = 3;
   */
  url: string;

  /**
   * @generated from field: string image_url = 4;
   */
  imageUrl: string;
};

/**
 * Describes the message api.v1.Repository.
 * Use `create(RepositorySchema)` to create a new message.
 */
export const RepositorySchema: GenMessage<Repository> = /*@__PURE__*/
  messageDesc(file_api_v1_api, 1);

/**
 * @generated from message api.v1.TimelineEntry
 */
export type TimelineEntry = Message<"api.v1.TimelineEntry"> & {
  /**
   * @generated from field: int32 id = 1;
   */
  id: number;

  /**
   * @generated from field: int32 repository_id = 2;
   */
  repositoryId: number;

  /**
   * @generated from field: string name = 3;
   */
  name: string;

  /**
   * @generated from field: string url = 4;
   */
  url: string;

  /**
   * @generated from field: string tag_name = 5;
   */
  tagName: string;

  /**
   * @generated from field: string description = 6;
   */
  description: string;

  /**
   * @generated from field: bool is_prerelease = 7;
   */
  isPrerelease: boolean;

  /**
   * @generated from field: google.protobuf.Timestamp released_at = 8;
   */
  releasedAt?: Timestamp;

  /**
   * @generated from field: string repository_name = 9;
   */
  repositoryName: string;

  /**
   * @generated from field: string image_url = 10;
   */
  imageUrl: string;

  /**
   * @generated from field: string author = 11;
   */
  author: string;

  /**
   * @generated from field: string repository_url = 12;
   */
  repositoryUrl: string;
};

/**
 * Describes the message api.v1.TimelineEntry.
 * Use `create(TimelineEntrySchema)` to create a new message.
 */
export const TimelineEntrySchema: GenMessage<TimelineEntry> = /*@__PURE__*/
  messageDesc(file_api_v1_api, 2);

/**
 * @generated from message api.v1.SyncRequest
 */
export type SyncRequest = Message<"api.v1.SyncRequest"> & {
  /**
   * @generated from field: string username = 1;
   */
  username: string;
};

/**
 * Describes the message api.v1.SyncRequest.
 * Use `create(SyncRequestSchema)` to create a new message.
 */
export const SyncRequestSchema: GenMessage<SyncRequest> = /*@__PURE__*/
  messageDesc(file_api_v1_api, 3);

/**
 * @generated from message api.v1.SyncResponse
 */
export type SyncResponse = Message<"api.v1.SyncResponse"> & {
  /**
   * @generated from field: repeated api.v1.TimelineEntry timeline = 1;
   */
  timeline: TimelineEntry[];
};

/**
 * Describes the message api.v1.SyncResponse.
 * Use `create(SyncResponseSchema)` to create a new message.
 */
export const SyncResponseSchema: GenMessage<SyncResponse> = /*@__PURE__*/
  messageDesc(file_api_v1_api, 4);

/**
 * @generated from message api.v1.GetRepositoriesRequest
 */
export type GetRepositoriesRequest = Message<"api.v1.GetRepositoriesRequest"> & {
};

/**
 * Describes the message api.v1.GetRepositoriesRequest.
 * Use `create(GetRepositoriesRequestSchema)` to create a new message.
 */
export const GetRepositoriesRequestSchema: GenMessage<GetRepositoriesRequest> = /*@__PURE__*/
  messageDesc(file_api_v1_api, 5);

/**
 * @generated from message api.v1.GetRepositoriesResponse
 */
export type GetRepositoriesResponse = Message<"api.v1.GetRepositoriesResponse"> & {
  /**
   * @generated from field: repeated api.v1.TimelineEntry timeline = 1;
   */
  timeline: TimelineEntry[];
};

/**
 * Describes the message api.v1.GetRepositoriesResponse.
 * Use `create(GetRepositoriesResponseSchema)` to create a new message.
 */
export const GetRepositoriesResponseSchema: GenMessage<GetRepositoriesResponse> = /*@__PURE__*/
  messageDesc(file_api_v1_api, 6);

/**
 * @generated from message api.v1.ToogleUserPublicFeedRequest
 */
export type ToogleUserPublicFeedRequest = Message<"api.v1.ToogleUserPublicFeedRequest"> & {
  /**
   * @generated from field: bool enabled = 1;
   */
  enabled: boolean;
};

/**
 * Describes the message api.v1.ToogleUserPublicFeedRequest.
 * Use `create(ToogleUserPublicFeedRequestSchema)` to create a new message.
 */
export const ToogleUserPublicFeedRequestSchema: GenMessage<ToogleUserPublicFeedRequest> = /*@__PURE__*/
  messageDesc(file_api_v1_api, 7);

/**
 * @generated from message api.v1.ToogleUserPublicFeedResponse
 */
export type ToogleUserPublicFeedResponse = Message<"api.v1.ToogleUserPublicFeedResponse"> & {
  /**
   * @generated from field: string public_id = 1;
   */
  publicId: string;
};

/**
 * Describes the message api.v1.ToogleUserPublicFeedResponse.
 * Use `create(ToogleUserPublicFeedResponseSchema)` to create a new message.
 */
export const ToogleUserPublicFeedResponseSchema: GenMessage<ToogleUserPublicFeedResponse> = /*@__PURE__*/
  messageDesc(file_api_v1_api, 8);

/**
 * @generated from message api.v1.GetMyUserRequest
 */
export type GetMyUserRequest = Message<"api.v1.GetMyUserRequest"> & {
};

/**
 * Describes the message api.v1.GetMyUserRequest.
 * Use `create(GetMyUserRequestSchema)` to create a new message.
 */
export const GetMyUserRequestSchema: GenMessage<GetMyUserRequest> = /*@__PURE__*/
  messageDesc(file_api_v1_api, 9);

/**
 * @generated from message api.v1.GetMyUserResponse
 */
export type GetMyUserResponse = Message<"api.v1.GetMyUserResponse"> & {
  /**
   * @generated from field: int32 id = 1;
   */
  id: number;

  /**
   * @generated from field: google.protobuf.Timestamp last_synced_at = 2;
   */
  lastSyncedAt?: Timestamp;

  /**
   * @generated from field: bool is_public = 3;
   */
  isPublic: boolean;

  /**
   * @generated from field: string public_id = 4;
   */
  publicId: string;

  /**
   * @generated from field: string name = 5;
   */
  name: string;
};

/**
 * Describes the message api.v1.GetMyUserResponse.
 * Use `create(GetMyUserResponseSchema)` to create a new message.
 */
export const GetMyUserResponseSchema: GenMessage<GetMyUserResponse> = /*@__PURE__*/
  messageDesc(file_api_v1_api, 10);

/**
 * @generated from message api.v1.LogoutRequest
 */
export type LogoutRequest = Message<"api.v1.LogoutRequest"> & {
};

/**
 * Describes the message api.v1.LogoutRequest.
 * Use `create(LogoutRequestSchema)` to create a new message.
 */
export const LogoutRequestSchema: GenMessage<LogoutRequest> = /*@__PURE__*/
  messageDesc(file_api_v1_api, 11);

/**
 * @generated from message api.v1.LogoutResponse
 */
export type LogoutResponse = Message<"api.v1.LogoutResponse"> & {
};

/**
 * Describes the message api.v1.LogoutResponse.
 * Use `create(LogoutResponseSchema)` to create a new message.
 */
export const LogoutResponseSchema: GenMessage<LogoutResponse> = /*@__PURE__*/
  messageDesc(file_api_v1_api, 12);

/**
 * @generated from message api.v1.RefreshTokenRequest
 */
export type RefreshTokenRequest = Message<"api.v1.RefreshTokenRequest"> & {
};

/**
 * Describes the message api.v1.RefreshTokenRequest.
 * Use `create(RefreshTokenRequestSchema)` to create a new message.
 */
export const RefreshTokenRequestSchema: GenMessage<RefreshTokenRequest> = /*@__PURE__*/
  messageDesc(file_api_v1_api, 13);

/**
 * @generated from message api.v1.RefreshTokenResponse
 */
export type RefreshTokenResponse = Message<"api.v1.RefreshTokenResponse"> & {
  /**
   * @generated from field: string access_token = 1;
   */
  accessToken: string;

  /**
   * @generated from field: string refresh_token = 2;
   */
  refreshToken: string;

  /**
   * @generated from field: google.protobuf.Timestamp access_token_expires_at = 3;
   */
  accessTokenExpiresAt?: Timestamp;

  /**
   * @generated from field: google.protobuf.Timestamp refresh_token_expires_at = 4;
   */
  refreshTokenExpiresAt?: Timestamp;
};

/**
 * Describes the message api.v1.RefreshTokenResponse.
 * Use `create(RefreshTokenResponseSchema)` to create a new message.
 */
export const RefreshTokenResponseSchema: GenMessage<RefreshTokenResponse> = /*@__PURE__*/
  messageDesc(file_api_v1_api, 14);

/**
 * @generated from service api.v1.ApiService
 */
export const ApiService: GenService<{
  /**
   * @generated from rpc api.v1.ApiService.Sync
   */
  sync: {
    methodKind: "unary";
    input: typeof SyncRequestSchema;
    output: typeof SyncResponseSchema;
  },
  /**
   * @generated from rpc api.v1.ApiService.GetRepositories
   */
  getRepositories: {
    methodKind: "unary";
    input: typeof GetRepositoriesRequestSchema;
    output: typeof GetRepositoriesResponseSchema;
  },
  /**
   * @generated from rpc api.v1.ApiService.ToogleUserPublicFeed
   */
  toogleUserPublicFeed: {
    methodKind: "unary";
    input: typeof ToogleUserPublicFeedRequestSchema;
    output: typeof ToogleUserPublicFeedResponseSchema;
  },
  /**
   * @generated from rpc api.v1.ApiService.GetMyUser
   */
  getMyUser: {
    methodKind: "unary";
    input: typeof GetMyUserRequestSchema;
    output: typeof GetMyUserResponseSchema;
  },
  /**
   * @generated from rpc api.v1.ApiService.Logout
   */
  logout: {
    methodKind: "unary";
    input: typeof LogoutRequestSchema;
    output: typeof LogoutResponseSchema;
  },
}> = /*@__PURE__*/
  serviceDesc(file_api_v1_api, 0);

/**
 * @generated from service api.v1.AuthService
 */
export const AuthService: GenService<{
  /**
   * @generated from rpc api.v1.AuthService.RefreshToken
   */
  refreshToken: {
    methodKind: "unary";
    input: typeof RefreshTokenRequestSchema;
    output: typeof RefreshTokenResponseSchema;
  },
}> = /*@__PURE__*/
  serviceDesc(file_api_v1_api, 1);

