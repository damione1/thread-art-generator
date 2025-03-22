import { Auth0Client } from "@auth0/nextjs-auth0/server"

export const auth0 = new Auth0Client({
    signInReturnToPath: "/dashboard",
    authorizationParameters: {
        audience: process.env.AUTH0_AUDIENCE,
        scope: "openid profile email read:current_user",
    },
})
