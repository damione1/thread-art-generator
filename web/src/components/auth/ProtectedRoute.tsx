import { useAuth0 } from "@auth0/auth0-react";
import { usePathname } from "next/navigation";
import { ReactNode, useEffect } from "react";

interface ProtectedRouteProps {
  children: ReactNode;
}

export const ProtectedRoute = ({ children }: ProtectedRouteProps) => {
  const { isAuthenticated, isLoading, loginWithRedirect } = useAuth0();
  const pathname = usePathname();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      loginWithRedirect({
        appState: { returnTo: pathname || "/" },
      });
    }
  }, [isAuthenticated, isLoading, loginWithRedirect, pathname]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary-500"></div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <p className="text-lg text-slate-200 mb-4">Redirecting to login...</p>
          <div className="animate-pulse rounded-full h-8 w-8 border-t-2 border-b-2 border-primary-500 mx-auto"></div>
        </div>
      </div>
    );
  }

  return <>{children}</>;
};
