import NextAuth from "next-auth/next";
import { authOptions } from "./options";

// @ts-ignore - Type error but works at runtime
const handler = NextAuth(authOptions);

export { handler as GET, handler as POST }
