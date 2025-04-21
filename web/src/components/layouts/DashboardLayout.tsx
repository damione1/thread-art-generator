import { useAuth0 } from "@auth0/auth0-react";
import { useRouter } from "next/router";
import { ReactNode, useEffect } from "react";

interface DashboardLayoutProps {
  children: ReactNode;
}

export const DashboardLayout = ({ children }: DashboardLayoutProps) => {
  const { isAuthenticated, isLoading, loginWithRedirect } = useAuth0();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      loginWithRedirect({
        appState: { returnTo: router.asPath },
      });
    }
  }, [isAuthenticated, isLoading, loginWithRedirect, router.asPath]);

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (!isAuthenticated) {
    return <div>Redirecting to login...</div>;
  }

  return <>{children}</>;
};
