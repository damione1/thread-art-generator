import { NextAuthOptions, Session } from "next-auth";
import { JWT } from "next-auth/jwt";
import { signOut } from "next-auth/react";

// nextauth.d.ts
declare module "next-auth" {
  interface Session {
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
  }
}

export const authOptions: NextAuthOptions = {
  providers: [
    {
      id: "credentials",
      name: "Credentials",
      type: "credentials",
      credentials: {
        email: {
          label: "Email",
          type: "email",
        },
        password: {
          label: "Password",
          type: "password",
        },
      },
      async authorize(credentials) {
        if (!credentials || !credentials.email || !credentials.password) {
          return null;
        }

        const authResponse = await fetch(`http://api:9091/v1/sessions`, {
          method: "POST",
          body: JSON.stringify({
            email: credentials.email,
            password: credentials.password,
          }),
          headers: {
            "Content-Type": "application/json",
            Accept: "application/json",
          },
        });
        const authToken = await authResponse.json();

        if (authResponse.ok && authToken) {
          return {
            id: authToken.user.id,
            email: authToken.user.email,
            name: `${authToken.user.first_name} ${authToken.user.last_name}`,
            firstName: authToken.user.first_name,
            lastName: authToken.user.last_name || "",
            image: authToken.user.avatar,
            accessToken: authToken.access_token,
            refreshToken: authToken.refresh_token,
            accessTokenExpires: new Date(authToken.access_token_expire_time),
            refreshTokenExpires: new Date(authToken.refresh_token_expire_time),
          };
        }
        return null;
      },
    },
  ],
  pages: {
    signIn: "/auth/signin",
  },
  callbacks: {
    async redirect({ url, baseUrl }) {
      return url;
    },
    async session({ session, token }) {
      session.user = {
        ...session.user,
        id: token.user.id,
        email: token.user.email,
        name: token.user.name,
        image: token.user.image,
      };
      session.backendTokens = {
        accessToken: token.backendTokens.accessToken,
        accessTokenExpires: token.backendTokens.accessTokenExpires,
        refreshToken: token.backendTokens.refreshToken,
        refreshTokenExpires: token.backendTokens.refreshTokenExpires,
      };
      return session;
    },
    async jwt({ token, user }) {
      // Initial sign-in
      if (user) {
        token.user = user;
        token.backendTokens = {
          accessToken: user.accessToken,
          accessTokenExpires: user.accessTokenExpires,
          refreshToken: user.refreshToken,
          refreshTokenExpires: user.refreshTokenExpires,
        };
      }

      // If refresh token Expired, log out user
      if (Date.now() > new Date(token.backendTokens.refreshTokenExpires).getTime()) {
        console.error("Refresh token expired");
        return {
          ...token,
          error: "RefreshTokenExpired",
        };
      }

      // If access token has not expired yet
      if (Date.now() < new Date(token.backendTokens.accessTokenExpires).getTime()) {
        return token;
      }

      // Access token has expired, try to refresh it
      return await refreshAccessToken(token);
    },
  },
};

async function refreshAccessToken(token: any) {
  try {
    console.log("Refreshing access token");
    const response = await fetch(`http://api:9091/v1/tokens:refresh`, {
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        refreshToken: token.backendTokens.refreshToken,
      }),
      method: "POST",
    });

    const refreshedTokens = await response.json();

    if (!response.ok) {
      throw new Error("Failed to refresh access token");
    }

    return {
      ...token,
      backendTokens: {
        accessToken: refreshedTokens.access_token,
        accessTokenExpires: refreshedTokens.access_token_expire_time,
        refreshToken: refreshedTokens.refresh_token,
        refreshTokenExpires: refreshedTokens.refresh_token_expire_time,
      },
    };
  } catch (error) {
    console.error("RefreshAccessTokenError", error);
    return {
      ...token,
      error: "RefreshAccessTokenError",
    };
  }
}
