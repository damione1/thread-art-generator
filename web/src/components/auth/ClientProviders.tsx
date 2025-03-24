"use client";

import { Auth0ProviderWithRedirect } from "./Auth0Provider";
import { ReactNode } from "react";

export default function ClientProviders({ children }: { children: ReactNode }) {
  return <Auth0ProviderWithRedirect>{children}</Auth0ProviderWithRedirect>;
}
