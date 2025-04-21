import { NextResponse } from "next/server";

// The auth is now handled client-side with Auth0 SPA
// This endpoint is kept as a placeholder for potential backend session management

export async function GET() {
    return NextResponse.json({
        message: "Authentication is now handled client-side with Auth0 SPA",
    });
}
