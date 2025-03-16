// This file exists just to prevent 404 errors for the NextAuth API path
// We're using Auth0 React SDK (@auth0/auth0-react) directly instead of NextAuth
// Auth0 is handling authentication via the Auth0Provider in /web/src/app/providers.tsx

export async function GET() {
    return new Response('Auth is handled by Auth0 React SDK', { status: 200 });
}

export async function POST() {
    return new Response('Auth is handled by Auth0 React SDK', { status: 200 });
}
