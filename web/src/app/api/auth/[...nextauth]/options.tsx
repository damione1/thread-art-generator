import { DefaultSession, NextAuthOptions, Session } from "next-auth";
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
        const authResponse = await fetch(`http://api:9091/v1/sessions`, {//using the http api for authentication via the backend
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
            id: authToken.user.name as string,
            email: authToken.user.email as string,
            name: `${authToken.user.first_name} ${authToken.user.last_name}`,
            firstName: authToken.user.last_name as string,
            lastName: authToken.user.last_name as string,
            image: authToken.user.avatar as string,
            accessToken: authToken.access_token as string,
            refreshToken: authToken.refresh_token as string,
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
      return {
        ...session,
        user: {
          ...session.user,
          id: token.id,
        },
        backendTokens: {
          accessToken: token.backendTokens.accessToken,
        },
      };
    },
    async jwt({ token, user, account, profile }) {
      // Persist the OAuth access_token and or the user id to the token right after signin
      if (user) {
        console.log("jwt user", user);
        console.log("jwt token", token);
        //only when user is signing in
        token.id = user.id;
        token.backendTokens = {
          accessToken: user.accessToken,
          accessTokenExpires: user.accessTokenExpires,
          refreshToken: user.refreshToken,
          refreshTokenExpires: user.refreshTokenExpires,
        } as JWT["backendTokens"];
      }
      // Ensure that the token object always has a user and backendTokens properties
      token.user = token.user || {};
      token.backendTokens = token.backendTokens || {};


      // Return previous token if the refresh token has not expired yet
      if (
        Date.now() > new Date(token.backendTokens.refreshTokenExpires).getTime()
      ) {
        console.error("refresh token expired " + new Date(token.backendTokens.refreshTokenExpires).getTime() + " " + Date.now().toLocaleString());
        return {
          ...token,
          error: "RefreshTokenExpired",
        };
      }

      // Return previous token if the access token has not expired yet
      if (
        Date.now() < new Date(token.backendTokens.accessTokenExpires).getTime()
      ) {
        console.log("access token not expired " + new Date(token.backendTokens.accessTokenExpires).getTime() + " " + Date.now());
        return token;
      }
      console.log("access token expired " + new Date(token.backendTokens.accessTokenExpires).getTime() + " " + Date.now());

      // Access token has expired, try to update it
      token = await refreshAccessToken(token);
      return token;
    },
  },
};

async function refreshAccessToken(token: any) {
  console.log("refreshing access token", token);
  const response = await fetch(`http://api:9091/v1/tokens:refresh`, {
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      refreshToken: token.backendTokens.refreshToken,
    }),
    method: "POST",
  });

  const decodedResponse = await response.json();

  console.log("decodedResponse", decodedResponse);
  if (response.ok && decodedResponse) {
    const backendTokens={
      accessToken: decodedResponse.access_token,
      accessTokenExpires: decodedResponse.access_token_expire_time,
      refreshToken: decodedResponse.refresh_token,
      refreshTokenExpires: decodedResponse.refresh_token_expire_time,
    }
    token.backendTokens = backendTokens;
    return token;
  }
  console.error("RefreshAccessTokenError", decodedResponse);

  return {
    ...token,
    error: "RefreshAccessTokenError",
  };
}
