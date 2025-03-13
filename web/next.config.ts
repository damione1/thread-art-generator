import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  images: {
    domains: [
      "localhost",
      "tag.local",
      "auth0.com",
      "s.gravatar.com",
      "lh3.googleusercontent.com",
      "avatars.githubusercontent.com",
    ],
  },
};

export default nextConfig;
