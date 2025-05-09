import { Auth0Provider } from "@auth0/auth0-react";
import { useRouter } from "next/navigation";
import React, { ReactNode, useEffect } from "react";

interface Auth0ProviderWithNavigateProps {
  children: ReactNode;
}

interface AppState {
  returnTo?: string;
}

export const Auth0ProviderWithRedirect = ({
  children,
}: Auth0ProviderWithNavigateProps) => {
  const router = useRouter();

  const domain = process.env.NEXT_PUBLIC_AUTH0_DOMAIN || "";
  const clientId = process.env.NEXT_PUBLIC_AUTH0_CLIENT_ID || "";
  const audience = process.env.NEXT_PUBLIC_AUTH0_AUDIENCE || "";
  const appBaseUrl = process.env.NEXT_PUBLIC_APP_BASE_URL || "";

  // Debug Auth0 config
  useEffect(() => {
    if (!domain || !clientId) {
      console.error(
        "Auth0 configuration is missing. Check your environment variables:"
      );
      console.error(`Domain: ${domain ? "✅" : "❌"}`);
      console.error(`ClientID: ${clientId ? "✅" : "❌"}`);
      console.error(`Audience: ${audience ? "✅" : "❌"}`);
    }
  }, [domain, clientId, audience]);

  const onRedirectCallback = (appState?: AppState) => {
    // Use App Router's push instead of Pages Router's replace
    router.push(appState?.returnTo || "/dashboard");
  };

  // Fallback if Auth0 isn't configured
  if (!domain || !clientId) {
    console.warn("Auth0 isn't properly configured. Auth will not work!");
    return <>{children}</>;
  }

  return (
    <Auth0Provider
      domain={domain}
      clientId={clientId}
      authorizationParams={{
        redirect_uri: `${appBaseUrl}/dashboard`,
        audience: audience,
        scope: "openid profile email",
      }}
      onRedirectCallback={onRedirectCallback}
      useRefreshTokens={true}
      useRefreshTokensFallback={true}
      cacheLocation="localstorage"
    >
      {children}
    </Auth0Provider>
  );
};
