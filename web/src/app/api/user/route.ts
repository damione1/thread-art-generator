import { auth0 } from "@/lib/auth0";
import { NextResponse } from "next/server";

export async function GET() {
    try {
        const session = await auth0.getSession();

        if (!session?.user) {
            return NextResponse.json(null, { status: 401 });
        }

        // Return the user data
        return NextResponse.json(session.user);
    } catch (error) {
        console.error("Error getting user data:", error);
        return NextResponse.json(
            { error: "Failed to get user data" },
            { status: 500 }
        );
    }
}
