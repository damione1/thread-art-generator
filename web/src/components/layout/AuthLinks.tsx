"use client";

import { useAuth } from "@/hooks/useAuth";
import { usePathname } from "next/navigation";

export default function AuthLinks() {
  const { loginRedirect } = useAuth();
  const pathname = usePathname();

  const handleLogin = () => {
    // Log the attempt to help with debugging
    try {
      loginRedirect({
        authorizationParams: {
          scope: "openid profile email",
        },
        appState: {
          returnTo: pathname || "/dashboard",
        },
      });
    } catch (error) {
      console.error("Failed to redirect to login:", error);
    }
  };

  // Always show buttons - loading skeleton removed
  return (
    <>
      <button
        onClick={handleLogin}
        className="px-4 py-2 rounded-md text-slate-200 hover:text-primary-300 transition"
      >
        Log in
      </button>
      <button
        onClick={handleLogin}
        className="px-4 py-2 rounded-md bg-primary-600 text-white hover:bg-primary-500 transition shadow-lg shadow-primary-900/20"
      >
        Sign up
      </button>
    </>
  );
}
