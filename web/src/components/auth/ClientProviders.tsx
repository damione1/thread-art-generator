"use client";

import { Auth0ProviderWithRedirect } from "./Auth0Provider";
import { UserProvider } from "@/contexts/UserContext";
import { ReactNode } from "react";

export default function ClientProviders({ children }: { children: ReactNode }) {
  return (
    <Auth0ProviderWithRedirect>
      <UserProvider>{children}</UserProvider>
    </Auth0ProviderWithRedirect>
  );
}
