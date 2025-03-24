import { NextResponse } from "next/server";

// With client-side Auth0 SPA, the logout is handled by the client
// This is just a convenience endpoint that redirects to home after logout
export async function GET() {
    return NextResponse.redirect(new URL("/", process.env.NEXT_PUBLIC_APP_BASE_URL || "http://localhost:3000"));
}
