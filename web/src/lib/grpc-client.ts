import { createGrpcWebTransport } from "@connectrpc/connect-web";
import { createClient } from "@connectrpc/connect";
import { ArtGeneratorService } from "./pb/services_connect";

// Create a transport that connects to the gRPC service
export const createTransport = () => {
    return createGrpcWebTransport({
        baseUrl: "https://tag.local",
        useBinaryFormat: true,
        credentials: 'include',
    });
};

// Create a client with the transport
export const createGrpcClient = (accessToken?: string) => {
    const transport = createTransport();
    const client = createClient(ArtGeneratorService, transport);

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
