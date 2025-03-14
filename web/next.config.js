/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  reactStrictMode: true,
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

module.exports = nextConfig;
