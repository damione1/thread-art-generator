import { createClient } from "@connectrpc/connect";
import { createGrpcWebTransport } from "@connectrpc/connect-web";
import { ArtGeneratorService } from "@/lib/pb/services_connect";
import {
    User,
    GetUserRequest,
    UpdateUserRequest
} from "@/lib/pb/user_pb";
import { FieldMask } from "@bufbuild/protobuf";
import { withAuth } from "@/lib/auth/authService";
import { ConnectError, Code } from "@connectrpc/connect";

// Create a gRPC-Web transport for the browser environment
const transport = createGrpcWebTransport({
    // Connect directly to the gRPC service without any prefix
    baseUrl: process.env.NEXT_PUBLIC_FRONTEND_URL || "https://tag.local",
    credentials: "include", // Include cookies for auth
    useBinaryFormat: true, // Use binary format for gRPC-Web (more compatible)
});

// Log the API URL for debugging
console.log("gRPC-Web API URL:", process.env.NEXT_PUBLIC_FRONTEND_URL || "https://tag.local");
console.log("Using gRPC-Web protocol with binary format");

// Create a gRPC client for the ArtGeneratorService
export const artGeneratorClient = createClient(ArtGeneratorService, transport);

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
    console.log("Fetching user profile with ID:", userId);

    // Log token structure to debug JWT format issues
    if (!accessToken) {
        console.error("Access token is empty or undefined");
        throw new Error("Missing access token. Please re-authenticate.");
    }

    const request = new GetUserRequest({
        name: `users/${userId}`,
    });

    console.log("GetUserRequest:", JSON.stringify(request, null, 2));

    try {
        // Add additional logging to debug the gRPC request
        const authHeaders = withAuth(accessToken);
        console.log("Making getUser gRPC call with auth headers");

        const response = await artGeneratorClient.getUser(request, authHeaders);
        console.log("GetUser response received");
        return response;
    } catch (error) {
        // Enhanced error logging to debug the issue
        console.error("Error getting user profile:", error);

        // Handle known error codes
        if (error instanceof ConnectError) {
            switch (error.code) {
                case Code.Unavailable:
                    throw new Error("The API service is currently unavailable. Please try again later.");
                case Code.NotFound:
                    throw new Error("User profile not found. Please create a profile first.");
                case Code.Unauthenticated:
                    throw new Error("Your session has expired. Please log in again.");
                default:
                    throw new Error(`API Error: ${error.message}`);
            }
        }

        throw error;
    }
}
