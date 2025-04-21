import { createConnectTransport } from "@connectrpc/connect-web";
import { createPromiseClient } from "@connectrpc/connect";
import { ConnectError, Code } from "@connectrpc/connect";
import { ArtGeneratorService } from "./pb/services_connect";
import { Art, Composition } from "./pb/art_pb";
import { getAccessToken, refreshAccessToken } from "@/utils/auth-token-manager";
import { processApiError } from "@/utils/errorUtils";

// Configuration for the Connect client
const CONFIG = {
    baseUrl: process.env.NEXT_PUBLIC_API_URL || "http://localhost:9090",
};

// Add this type for Connect errors at the top of the file
interface ConnectErrorDetails {
    code: number;
    message: string;
    details?: unknown;
}

/**
 * Creates a transport for Connect
 */
export const createTransport = () => {
    return createConnectTransport({
        baseUrl: CONFIG.baseUrl,
        useBinaryFormat: true,
        credentials: 'include',
    });
};

/**
 * Creates a Connect client with the current access token
 * If no token is provided, it will attempt to fetch one
 */
export const createConnectClient = async (providedToken?: string) => {
    const transport = createTransport();
    const client = createPromiseClient(ArtGeneratorService, transport);

    // If a token is explicitly provided, use it
    let accessToken = providedToken;

    // Otherwise, get one from our token manager
    if (!accessToken) {
        try {
            accessToken = await getAccessToken();
        } catch (error) {
            console.error("Failed to get access token:", error);
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
export class ConnectService {
    /**
     * Makes a Connect service call with proper error handling and token refresh
     * @param serviceCall - Function that makes the actual Connect call
     * @param forceFetchToken - Whether to force fetch a new token
     */
    static async call<T>(
        serviceCall: (token: string | undefined) => Promise<T>,
        forceFetchToken = false
    ): Promise<T> {
        try {
            // Get the current token
            const token = await getAccessToken();

            // Make the call with the token
            return await serviceCall(token);

        } catch (error) {
            // Handle auth errors by refreshing the token and retrying once
            if (
                error instanceof ConnectError &&
                (error.code === Code.Unauthenticated || error.code === Code.PermissionDenied) &&
                !forceFetchToken
            ) {
                // Try one more time with a fresh token by logging out and redirecting
                await refreshAccessToken();

                // The page will reload, so we don't need to return anything
                throw error;
            }

            // Process error to standardized format
            const processedError = processApiError(error);

            // Rethrow the processed error for higher-level handling
            throw processedError;
        }
    }
}

// Wrapper functions for common operations

/**
 * Get the current authenticated user
 */
export const getCurrentUser = async () => {
    const { GetCurrentUserRequest } = await import("./pb/user_pb");

    return ConnectService.call(async (token) => {
        const { client, callOptions } = await createConnectClient(token);
        return client.getCurrentUser(new GetCurrentUserRequest(), callOptions);
    });
};

/**
 * Get a user by ID
 */
export const getUser = async (userId: string) => {
    const { GetUserRequest } = await import("./pb/user_pb");

    return ConnectService.call(async (token) => {
        const { client, callOptions } = await createConnectClient(token);
        return client.getUser(new GetUserRequest({ name: userId }), callOptions);
    });
};

/**
 * Update a user with all fields
 */
export const updateUser = async (
    userData: Partial<{
        name: string;
        firstName: string;
        lastName: string;
        email: string;
        avatar: string;
    }>
) => {
    const { UpdateUserRequest, User } = await import("./pb/user_pb");
    return ConnectService.call(async (token) => {
        try {
            const { client, callOptions } = await createConnectClient(token);
            const request = new UpdateUserRequest({
                user: new User(userData),
            });
            const response = await client.updateUser(request, callOptions);
            return response;
        } catch (error) {
            console.error("UpdateUser error:", error);
            // Log more details if it's a Connect error
            if (error && typeof error === 'object' && 'code' in error) {
                const connectError = error as ConnectErrorDetails;
                console.error("Connect error details:", {
                    code: connectError.code,
                    message: connectError.message,
                    details: connectError.details
                });
            }
            throw error;
        }
    });
};

// Add more API methods as needed for Art, etc.

/**
 * Create a new art piece
 */
export const createArt = async (art: Partial<Art>, parent: string) => {
    const { CreateArtRequest, Art } = await import("./pb/art_pb");

    return ConnectService.call(async (token) => {
        const { client, callOptions } = await createConnectClient(token);
        const request = new CreateArtRequest({
            parent,
            art: new Art(art)
        });
        return client.createArt(request, callOptions);
    });
};

/**
 * Get an upload URL for an art piece image
 */
export const getArtUploadUrl = async (artName: string) => {
    const { GetArtUploadUrlRequest } = await import("./pb/art_pb");

    return ConnectService.call(async (token) => {
        const { client, callOptions } = await createConnectClient(token);
        return client.getArtUploadUrl(new GetArtUploadUrlRequest({ name: artName }), callOptions);
    });
};

/**
 * Confirm that an art image has been uploaded successfully
 */
export const confirmArtImageUpload = async (artName: string) => {
    const { ConfirmArtImageUploadRequest } = await import("./pb/art_pb");

    return ConnectService.call(async (token) => {
        const { client, callOptions } = await createConnectClient(token);
        return client.confirmArtImageUpload(
            new ConfirmArtImageUploadRequest({ name: artName }),
            callOptions
        );
    });
};

/**
 * Get an art piece by ID
 */
export const getArt = async (artId: string) => {
    const { GetArtRequest } = await import("./pb/art_pb");

    return ConnectService.call(async (token) => {
        const { client, callOptions } = await createConnectClient(token);
        return client.getArt(new GetArtRequest({ name: artId }), callOptions);
    });
};

/**
 * Update an art piece
 */
export const updateArt = async (art: Partial<Art>, updateMask: string[] = []) => {
    const { UpdateArtRequest } = await import("./pb/art_pb");
    const { FieldMask } = await import("./pb/google/protobuf/field_mask_pb");

    return ConnectService.call(async (token) => {
        const { client, callOptions } = await createConnectClient(token);
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
export const listArts = async (parent: string, pageSize: number = 10, pageToken?: string) => {
    const { ListArtsRequest } = await import("./pb/art_pb");

    return ConnectService.call(async (token) => {
        const { client, callOptions } = await createConnectClient(token);
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

    return ConnectService.call(async (token) => {
        const { client, callOptions } = await createConnectClient(token);
        return client.deleteArt(new DeleteArtRequest({ name: artId }), callOptions);
    });
};

/**
 * List users
 */
export const listUsers = async (pageSize: number = 10, pageToken?: string) => {
    const { ListUsersRequest } = await import("./pb/user_pb");

    return ConnectService.call(async (token) => {
        const { client, callOptions } = await createConnectClient(token);
        const request = new ListUsersRequest({
            pageSize,
            pageToken
        });
        return client.listUsers(request, callOptions);
    });
};


/**
 * Delete a user
 */
export const deleteUser = async (userId: string) => {
    const { DeleteUserRequest } = await import("./pb/user_pb");

    return ConnectService.call(async (token) => {
        const { client, callOptions } = await createConnectClient(token);
        return client.deleteUser(new DeleteUserRequest({ name: userId }), callOptions);
    });
};

/**
 * Create a new composition
 */
export const createComposition = async (request: {
    parent: string;
    composition: Partial<Composition>;
}) => {
    const { CreateCompositionRequest } = await import("./pb/art_pb");

    return ConnectService.call(async (token) => {
        const { client, callOptions } = await createConnectClient(token);
        const grpcRequest = new CreateCompositionRequest({
            parent: request.parent,
            composition: request.composition as Composition
        });
        return client.createComposition(grpcRequest, callOptions);
    });
};

/**
 * List compositions for an art piece
 */
export const listCompositions = async (request: {
    parent: string;
    pageSize?: number;
    pageToken?: string;
}) => {
    const { ListCompositionsRequest } = await import("./pb/art_pb");

    return ConnectService.call(async (token) => {
        const { client, callOptions } = await createConnectClient(token);
        const grpcRequest = new ListCompositionsRequest({
            parent: request.parent,
            pageSize: request.pageSize,
            pageToken: request.pageToken
        });
        return client.listCompositions(grpcRequest, callOptions);
    });
};

/**
 * Get a composition by ID
 */
export const getComposition = async (name: string) => {
    const { GetCompositionRequest } = await import("./pb/art_pb");

    return ConnectService.call(async (token) => {
        const { client, callOptions } = await createConnectClient(token);
        return client.getComposition(new GetCompositionRequest({ name }), callOptions);
    });
};

/**
 * Delete a composition
 */
export const deleteComposition = async (name: string) => {
    const { DeleteCompositionRequest } = await import("./pb/art_pb");

    return ConnectService.call(async (token) => {
        const { client, callOptions } = await createConnectClient(token);
        return client.deleteComposition(new DeleteCompositionRequest({ name }), callOptions);
    });
};
