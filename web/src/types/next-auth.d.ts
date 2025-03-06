import { DefaultSession, NextAuth } from "next-auth";
import { JWT as NextAuthJWT } from "next-auth/jwt";

declare module "next-auth" {
    interface Session extends DefaultSession {
        user: {
            id: number;
            email: string;
            name: string;
            image: string;
        };
        backendTokens: {
            accessToken: string;
            accessTokenExpires: Date;
            refreshToken: string;
            refreshTokenExpires: Date;
        };
    }

    interface User {
        id: number;
        email: string;
        name: string;
        accessToken: string;
        accessTokenExpires: Date;
        refreshToken: string;
        refreshTokenExpires: Date;
        image?: string;
    }
}

declare module "next-auth/jwt" {
    interface JWT extends NextAuthJWT {
        user: {
            id: number;
            email: string;
            name: string;
            image?: string;
        };
        backendTokens: {
            accessToken: string;
            accessTokenExpires: Date;
            refreshToken: string;
            refreshTokenExpires: Date;
        };
        error?: "RefreshAccessTokenError" | "RefreshTokenExpired";
    }
}
