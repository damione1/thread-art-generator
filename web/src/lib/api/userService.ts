import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { ArtGeneratorService } from "@/lib/pb/services_connect";
import {
    User,
    GetUserRequest,
    UpdateUserRequest
} from "@/lib/pb/user_pb";
import { FieldMask } from "@bufbuild/protobuf";

// Create a Connect transport for the browser environment
const transport = createConnectTransport({
    baseUrl: process.env.NEXT_PUBLIC_API_URL || "https://tag.local/grpc-api",
    credentials: "include", // Include cookies for auth
});

// Create a gRPC client for the ArtGeneratorService
export const artGeneratorClient = createPromiseClient(ArtGeneratorService, transport);

// Helper function to add auth token to metadata
export const withAuth = (token: string) => ({
    headers: {
        Authorization: `Bearer ${token}`,
    },
});

// Function to update user profile
export async function updateUserProfile(
    accessToken: string,
    userId: string,
    updates: {
        firstName?: string;
        lastName?: string;
        email?: string;
    }
) {
    // Create paths array for field mask based on provided updates
    const paths: string[] = [];
    if (updates.firstName !== undefined) paths.push("first_name");
    if (updates.lastName !== undefined) paths.push("last_name");
    if (updates.email !== undefined) paths.push("email");

    // Create field mask
    const updateMask = new FieldMask({
        paths: paths,
    });

    // Create user object
    const user = new User({
        name: `users/${userId}`,
        firstName: updates.firstName ?? "",
        lastName: updates.lastName ?? "",
        email: updates.email ?? "",
        avatar: "", // Required field in the proto but we're not updating it
    });

    // Create update request with user and field mask
    const request = new UpdateUserRequest({
        user: user,
        updateMask: updateMask,
    });

    // Make the gRPC call with authentication
    return artGeneratorClient.updateUser(request, withAuth(accessToken));
}

// Function to get user profile
export async function getUserProfile(accessToken: string, userId: string) {
    const request = new GetUserRequest({
        name: `users/${userId}`,
    });

    return artGeneratorClient.getUser(request, withAuth(accessToken));
}
