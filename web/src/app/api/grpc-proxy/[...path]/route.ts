import { NextRequest } from 'next/server';

export async function GET(
    request: NextRequest,
    context: { params: { path: string[] } }
) {
    return handleRequest(request, context.params.path);
}

export async function POST(
    request: NextRequest,
    context: { params: { path: string[] } }
) {
    return handleRequest(request, context.params.path);
}

export async function OPTIONS(
    request: NextRequest,
    context: { params: { path: string[] } }
) {
    // These aren't used but including them to avoid TypeScript errors
    void request;
    void context;

    return new Response(null, {
        status: 200,
        headers: {
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
            'Access-Control-Allow-Headers': 'Content-Type, Authorization, x-grpc-web, grpc-timeout',
            'Access-Control-Max-Age': '86400',
        },
    });
}

async function handleRequest(request: NextRequest, pathParts: string[]) {
    // Determine the correct base URL based on environment
    const isProduction = process.env.NODE_ENV === 'production';
    // In production, we proxy to Envoy internally
    // In development, we go through the TLS termination provided by Traefik
    const baseUrl = isProduction
        ? 'http://envoy:8080'
        : 'https://tag.local';

    const path = `/${pathParts.join('/')}`;
    const targetUrl = `${baseUrl}${path}`;

    // Clone the request headers to forward them
    const headers = new Headers();
    request.headers.forEach((value, key) => {
        headers.append(key, value);
    });

    try {
        const response = await fetch(targetUrl, {
            method: request.method,
            headers,
            body: request.body,
        });

        // Construct a response that includes the necessary gRPC-Web headers
        const responseHeaders = new Headers();
        response.headers.forEach((value, key) => {
            responseHeaders.append(key, value);
        });

        // Ensure CORS and gRPC-Web specific headers are present
        responseHeaders.set('Access-Control-Allow-Origin', '*');

        return new Response(response.body, {
            status: response.status,
            statusText: response.statusText,
            headers: responseHeaders,
        });
    } catch (error) {
        console.error('Error proxying gRPC-Web request:', error);
        return new Response(JSON.stringify({ error: 'Failed to proxy request' }), {
            status: 500,
            headers: {
                'Content-Type': 'application/json',
                'Access-Control-Allow-Origin': '*',
            },
        });
    }
}
