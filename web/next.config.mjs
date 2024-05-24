/** @type {import('next').NextConfig} */
const nextConfig = {
    images: {
        remotePatterns: [
            {
                hostname: "storage.googleapis.com",
            },
            {
                hostname: "storage.cloud.google.com",
            },
            {
                hostname: "www.gravatar.com",
            },
            {
                hostname: "minio",
            }
        ],
    },
    reactStrictMode: false,
};


export default nextConfig;
