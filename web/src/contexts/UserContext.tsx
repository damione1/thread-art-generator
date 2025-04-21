import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from "react";
import { useAuth } from "@/hooks/useAuth";
import { getCurrentUser } from "@/lib/grpc-client";
import { User as ProtoUser } from "@/lib/pb/user_pb";
import { User } from "@/types/user";
import { Auth0User } from "@/types/auth0";
import { getErrorMessage } from "@/utils/errorUtils";

interface UserContextType {
  // Auth0 user profile
  authUser: Auth0User | null;
  // Our backend user data
  user: User;
  // Full user profile from our API
  profile: ProtoUser | null;
  // Loading state
  loading: boolean;
  // Error state
  error: string | null;
  // Refresh user data
  refreshUserData: () => Promise<void>;
}

const defaultUser: User = {
  id: "",
  name: "",
  email: "",
};

const UserContext = createContext<UserContextType>({
  authUser: null,
  user: defaultUser,
  profile: null,
  loading: false,
  error: null,
  refreshUserData: async () => {},
});

export function useUser() {
  return useContext(UserContext);
}

interface UserProviderProps {
  children: ReactNode;
}

export function UserProvider({ children }: UserProviderProps) {
  const {
    user: auth0User,
    isAuthenticated,
    isLoading: authLoading,
  } = useAuth();
  const [authUser, setAuthUser] = useState<Auth0User | null>(null);
  const [user, setUser] = useState<User>(defaultUser);
  const [profile, setProfile] = useState<ProtoUser | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Update authUser state when auth0 user changes
  useEffect(() => {
    if (auth0User) {
      setAuthUser(auth0User as Auth0User);

      // Update user from auth0 data
      setUser({
        id: auth0User.sub || "",
        name: auth0User.name || "",
        email: auth0User.email || "",
        picture: auth0User.picture,
      });
    } else {
      // Reset user when not authenticated
      setAuthUser(null);
      setUser(defaultUser);
      setProfile(null);
    }
  }, [auth0User]);

  // Fetch user data from backend when authenticated
  useEffect(() => {
    if (isAuthenticated && !authLoading && user.id && !profile) {
      fetchUserData();
    }
  }, [isAuthenticated, authLoading, user.id, profile]);

  // Function to fetch user data from API
  const fetchUserData = async () => {
    if (!isAuthenticated || !user.id) {
      return;
    }

    try {
      setLoading(true);
      setError(null);

      // Fetch complete user profile from API
      const userData = await getCurrentUser();
      setProfile(userData);

      // Update user with more complete data if available
      if (userData) {
        setUser((prev) => ({
          ...prev,
          name:
            userData.firstName && userData.lastName
              ? `${userData.firstName} ${userData.lastName}`
              : prev.name,
          // Update any other fields as needed
        }));
      }
    } catch (err) {
      console.error("Error fetching user data:", err);

      // Don't display auth errors to the user - they'll be handled by redirection
      if (
        typeof err === "object" &&
        err !== null &&
        "isAuthError" in err &&
        (err as { isAuthError: boolean }).isAuthError
      ) {
        console.warn(
          "Auth error during user data fetch - this is handled automatically"
        );
      } else {
        // Only set non-auth errors
        setError(getErrorMessage(err));
      }
    } finally {
      setLoading(false);
    }
  };

  // Public function to manually refresh user data
  const refreshUserData = async () => {
    await fetchUserData();
  };

  const value: UserContextType = {
    authUser,
    user,
    profile,
    loading,
    error,
    refreshUserData,
  };

  return <UserContext.Provider value={value}>{children}</UserContext.Provider>;
}
