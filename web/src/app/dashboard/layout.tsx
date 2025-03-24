"use client";

import { ProtectedRoute } from "@/components/auth/ProtectedRoute";
import React from "react";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <ProtectedRoute>{children}</ProtectedRoute>;
}
