import { Auth0Provider } from "@auth0/auth0-react";
import { useRouter } from "next/navigation";
import React, { ReactNode } from "react";

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

  const onRedirectCallback = (appState?: AppState) => {
    router.push(appState?.returnTo || "/dashboard");
  };

  if (!domain || !clientId) {
    return <>{children}</>;
  }

  return (
    <Auth0Provider
      domain={domain}
      clientId={clientId}
      authorizationParams={{
        redirect_uri:
          typeof window !== "undefined"
            ? window.location.origin + "/dashboard"
            : "",
        audience: audience,
      }}
      onRedirectCallback={onRedirectCallback}
      useRefreshTokens={true}
      cacheLocation="localstorage"
    >
      {children}
    </Auth0Provider>
  );
};
