import type { JWT } from 'next-auth/jwt'
import CredentialsProvider from "next-auth/providers/credentials";
import type { Session } from "next-auth";

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

export const authOptions = {
  providers: [
    CredentialsProvider({
      id: "credentials",
      name: "Credentials",
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
      async authorize(credentials: Record<"email" | "password", string> | undefined) {
        if (!credentials || !credentials.email || !credentials.password) {
          throw new Error("Email and password are required");
        }

        try {
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

          const data = await authResponse.json();

          if (authResponse.ok && data) {
            return {
              id: data.user.id,
              email: data.user.email,
              name: `${data.user.first_name} ${data.user.last_name}`,
              firstName: data.user.first_name,
              lastName: data.user.last_name || "",
              image: data.user.avatar,
              accessToken: data.access_token,
              refreshToken: data.refresh_token,
              accessTokenExpires: new Date(data.access_token_expire_time),
              refreshTokenExpires: new Date(data.refresh_token_expire_time),
            };
          }

          // Pass through the backend error message
          throw new Error(data.error || data.message || "Authentication failed");
        } catch (error: any) {
          // Pass through any error messages from the backend
          throw new Error(error.message || "Authentication failed");
        }
      },
    }),
  ],
  pages: {
    signIn: "/auth",
  },
  callbacks: {
    async redirect({ url, baseUrl }: { url: string; baseUrl: string }) {
      return url;
    },
    async session({ session, token }: { session: Session; token: JWT }) {
      session.user = {
        ...session.user,
        id: token.user.id,
        email: token.user.email,
        name: token.user.name,
        image: token.user.image || "",
      };
      session.backendTokens = {
        accessToken: token.backendTokens.accessToken,
        accessTokenExpires: token.backendTokens.accessTokenExpires,
        refreshToken: token.backendTokens.refreshToken,
        refreshTokenExpires: token.backendTokens.refreshTokenExpires,
      };
      return session;
    },
    async jwt({ token, user }: { token: JWT; user: any }) {
      // Initial sign-in
      if (user) {
        token.user = {
          id: user.id,
          email: user.email,
          name: user.name,
          image: user.image || undefined,
        };
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
