import { auth0 } from "@/lib/auth0";
import { NextResponse } from "next/server";

export async function GET() {
    try {
        const session = await auth0.getSession();

        // Return the session data
        return NextResponse.json({
            user: session?.user || null,
            accessToken: session?.tokenSet?.accessToken || null,
        });
    } catch (error) {
        console.error("Error getting session:", error);
        return NextResponse.json(
            { error: "Failed to get session" },
            { status: 500 }
        );
    }
}
