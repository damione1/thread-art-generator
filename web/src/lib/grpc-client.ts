import { createGrpcWebTransport } from "@connectrpc/connect-web";
import { createClient, ConnectError, Code } from "@connectrpc/connect";
import { ArtGeneratorService } from "./pb/services_connect";
import { User } from "./pb/user_pb";
import { Art } from "./pb/art_pb";

// Cache for the access token
type TokenCache = {
    token: string | null;
    expiresAt: number; // Unix timestamp when token expires
};

// Default token cache with no token
const tokenCache: TokenCache = {
    token: null,
    expiresAt: 0,
};

// Configuration for the gRPC client
const CONFIG = {
    baseUrl: process.env.NEXT_PUBLIC_APP_BASE_URL || "http://localhost:3000",
    tokenExpiryBufferMs: 5 * 60 * 1000, // 5 minutes buffer before token expiry
};

/**
 * Creates a transport for gRPC web
 */
export const createTransport = () => {
    return createGrpcWebTransport({
        baseUrl: CONFIG.baseUrl,
        useBinaryFormat: true,
        credentials: 'include',
    });
};

/**
 * Fetches a fresh access token from the server
 */
export const fetchAccessToken = async (): Promise<string> => {
    try {
        const response = await fetch("/api/auth/session");
        if (!response.ok) {
            throw new Error("Failed to get session");
        }

        const data = await response.json();
        if (!data.accessToken) {
            throw new Error("No access token available");
        }

        // Set the token in the cache with an estimated expiry time
        // This assumes the token is valid for 1 hour, adjust as needed
        tokenCache.token = data.accessToken;
        tokenCache.expiresAt = Date.now() + 55 * 60 * 1000; // 55 minutes from now

        return data.accessToken;
    } catch (error) {
        console.error("Error fetching access token:", error);
        throw error;
    }
};

/**
 * Gets a valid access token, either from cache or by fetching a new one
 */
export const getAccessToken = async (): Promise<string> => {
    // Check if we have a cached token that's not expired
    if (tokenCache.token && Date.now() < tokenCache.expiresAt - CONFIG.tokenExpiryBufferMs) {
        return tokenCache.token;
    }

    // Fetch a new token
    return fetchAccessToken();
};

/**
 * Creates a gRPC client with the current access token
 * If no token is provided, it will attempt to fetch one
 */
export const createGrpcClient = async (providedToken?: string) => {
    const transport = createTransport();
    const client = createClient(ArtGeneratorService, transport);

    // If a token is explicitly provided, use it
    let accessToken = providedToken;

    // Otherwise, get one from cache or fetch a new one
    if (!accessToken) {
        try {
            accessToken = await getAccessToken();
        } catch (error) {
            console.error("Failed to get access token:", error);
            // Return client without token - will work for public endpoints
        }
    }

    return {
        client,
        // Helper to use with each call
        callOptions: accessToken ? {
            headers: {
                Authorization: `Bearer ${accessToken}`,
            },
        } : undefined
    };
};

/**
 * Utility class for making service calls with automatic token refresh
 * and error handling
 */
export class GrpcService {
    /**
     * Makes a gRPC service call with proper error handling and token refresh
     * @param serviceCall - Function that makes the actual gRPC call
     * @param forceFetchToken - Whether to force fetch a new token
     */
    static async call<T>(
        serviceCall: (token: string | undefined) => Promise<T>,
        forceFetchToken = false
    ): Promise<T> {
        try {
            // Get the current token (or force fetch a new one)
            const token = forceFetchToken ? await fetchAccessToken() : await getAccessToken();

            // Make the call with the token
            return await serviceCall(token);

        } catch (error) {
            // Handle auth errors by refreshing the token and retrying once
            if (
                error instanceof ConnectError &&
                (error.code === Code.Unauthenticated || error.code === Code.PermissionDenied) &&
                !forceFetchToken
            ) {
                // Try one more time with a fresh token
                return GrpcService.call(serviceCall, true);
            }

            // Otherwise rethrow
            throw error;
        }
    }
}

// Wrapper functions for common operations

/**
 * Get the current authenticated user
 */
export const getCurrentUser = async () => {
    const { GetCurrentUserRequest } = await import("./pb/user_pb");

    return GrpcService.call(async (token) => {
        const { client, callOptions } = await createGrpcClient(token);
        return client.getCurrentUser(new GetCurrentUserRequest(), callOptions);
    });
};

/**
 * Get a user by ID
 */
export const getUser = async (userId: string) => {
    const { GetUserRequest } = await import("./pb/user_pb");

    return GrpcService.call(async (token) => {
        const { client, callOptions } = await createGrpcClient(token);
        return client.getUser(new GetUserRequest({ name: userId }), callOptions);
    });
};

/**
 * Update a user with partial fields
 * Following Google API Design guidelines with field masks
 */
export const updateUser = async (
    userData: Partial<{
        name: string;
        firstName: string;
        lastName: string;
        email: string;
        avatar: string;
    }>,
    updateMask: string[] = []
) => {
    const { UpdateUserRequest, User } = await import("./pb/user_pb");
    const { FieldMask } = await import("./pb/google/protobuf/field_mask_pb");

    // If updateMask is empty, automatically generate it from userData keys
    if (updateMask.length === 0 && userData) {
        updateMask = Object.keys(userData).filter(key => key !== 'name'); // name is identifier, not updatable field
    }

    return GrpcService.call(async (token) => {
        const { client, callOptions } = await createGrpcClient(token);

        const request = new UpdateUserRequest({
            user: new User(userData),
            updateMask: new FieldMask({ paths: updateMask }),
        });

        return client.updateUser(request, callOptions);
    });
};

// Add more API methods as needed for Art, etc.

/**
 * Create a new art piece
 */
export const createArt = async (art: Partial<Art>, parent: string) => {
    const { CreateArtRequest } = await import("./pb/art_pb");

    return GrpcService.call(async (token) => {
        const { client, callOptions } = await createGrpcClient(token);
        const request = new CreateArtRequest({
            parent,
            art: art as Art
        });
        return client.createArt(request, callOptions);
    });
};

/**
 * Get an art piece by ID
 */
export const getArt = async (artId: string) => {
    const { GetArtRequest } = await import("./pb/art_pb");

    return GrpcService.call(async (token) => {
        const { client, callOptions } = await createGrpcClient(token);
        return client.getArt(new GetArtRequest({ name: artId }), callOptions);
    });
};

/**
 * Update an art piece
 */
export const updateArt = async (art: Partial<Art>, updateMask: string[] = []) => {
    const { UpdateArtRequest } = await import("./pb/art_pb");
    const { FieldMask } = await import("./pb/google/protobuf/field_mask_pb");

    return GrpcService.call(async (token) => {
        const { client, callOptions } = await createGrpcClient(token);
        const request = new UpdateArtRequest({
            art: art as Art,
            updateMask: new FieldMask({ paths: updateMask })
        });
        return client.updateArt(request, callOptions);
    });
};

/**
 * List arts for a parent (user)
 */
export const listArts = async (parent: string, pageSize: number = 10, pageToken?: number) => {
    const { ListArtsRequest } = await import("./pb/art_pb");

    return GrpcService.call(async (token) => {
        const { client, callOptions } = await createGrpcClient(token);
        const request = new ListArtsRequest({
            parent,
            pageSize,
            pageToken
        });
        return client.listArts(request, callOptions);
    });
};

/**
 * Delete an art piece
 */
export const deleteArt = async (artId: string) => {
    const { DeleteArtRequest } = await import("./pb/art_pb");

    return GrpcService.call(async (token) => {
        const { client, callOptions } = await createGrpcClient(token);
        return client.deleteArt(new DeleteArtRequest({ name: artId }), callOptions);
    });
};

/**
 * List users
 */
export const listUsers = async (pageSize: number = 10, pageToken?: string) => {
    const { ListUsersRequest } = await import("./pb/user_pb");

    return GrpcService.call(async (token) => {
        const { client, callOptions } = await createGrpcClient(token);
        const request = new ListUsersRequest({
            pageSize,
            pageToken
        });
        return client.listUsers(request, callOptions);
    });
};

/**
 * Create a new user
 */
export const createUser = async (user: Partial<User>) => {
    const { CreateUserRequest } = await import("./pb/user_pb");

    return GrpcService.call(async (token) => {
        const { client, callOptions } = await createGrpcClient(token);
        const request = new CreateUserRequest({
            user: user as User
        });
        return client.createUser(request, callOptions);
    });
};

/**
 * Delete a user
 */
export const deleteUser = async (userId: string) => {
    const { DeleteUserRequest } = await import("./pb/user_pb");

    return GrpcService.call(async (token) => {
        const { client, callOptions } = await createGrpcClient(token);
        return client.deleteUser(new DeleteUserRequest({ name: userId }), callOptions);
    });
};
