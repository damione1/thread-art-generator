"use client";

import { Auth0ProviderWithRedirect } from "@/components/auth/Auth0Provider";
import { ProtectedRoute } from "@/components/auth/ProtectedRoute";
import React from "react";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <Auth0ProviderWithRedirect>
      <ProtectedRoute>{children}</ProtectedRoute>
    </Auth0ProviderWithRedirect>
  );
}
