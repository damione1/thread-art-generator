import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  reactStrictMode: true,
  images: {
    domains: [
      "localhost",
      "tag.local",
      "auth0.com",
      "s.gravatar.com",
      "lh3.googleusercontent.com",
      "avatars.githubusercontent.com",
      "www.gravatar.com",
      "storage.tag.local"
    ],
    unoptimized: process.env.NODE_ENV === 'development', // Only in dev
  },
  // For gRPC-Web we need to configure rewrites to proxy API requests
  async rewrites() {
    return [
      {
        // Forward API requests to the Envoy proxy
        source: "/api.v1/:path*",
        destination: "/api/grpc-proxy/:path*",
      },
    ];
  },
  // Configure headers for gRPC-Web
  async headers() {
    return [
      {
        source: "/api.v1/:path*",
        headers: [
          { key: "Access-Control-Allow-Origin", value: "*" },
          { key: "Access-Control-Allow-Methods", value: "GET, POST, OPTIONS" },
          {
            key: "Access-Control-Allow-Headers",
            value: "Content-Type, Authorization, x-grpc-web",
          },
        ],
      },
    ];
  },
};

export default nextConfig;
