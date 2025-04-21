"use client";

import { useAuth } from "@/hooks/useAuth";
import { useEffect } from "react";

export default function RequireAuth({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, isLoading, loginRedirect } = useAuth();

  useEffect(() => {
    // Only redirect if we're not loading and not authenticated
    if (!isLoading && !isAuthenticated) {
      loginRedirect({
        appState: {
          returnTo: window.location.pathname,
        },
      });
    }
  }, [isAuthenticated, isLoading, loginRedirect]);

  // Show nothing while loading or when redirecting
  if (isLoading || !isAuthenticated) {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary-500"></div>
      </div>
    );
  }

  // If authenticated, render children
  return <>{children}</>;
}
