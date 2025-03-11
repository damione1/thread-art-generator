import { Auth0Client } from "@auth0/nextjs-auth0/server";

// Create Auth0 client using environment variables
export const auth0 = new Auth0Client();
