import { NextResponse } from "next/server";

// The dashboard is now fully protected with client-side Auth0 SPA authentication.
// This API route is being kept as a placeholder for future backend communication.

export async function GET() {
    return NextResponse.json({ message: "Use client-side Auth0 for user data" });
}
