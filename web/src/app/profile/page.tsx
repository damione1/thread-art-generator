"use client";

import { useState, useEffect } from "react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth0 } from "@auth0/auth0-react";
import { updateUserProfile, getUserProfile } from "@/lib/api/userService";
import { toast } from "react-hot-toast";

// Define the user profile response type
interface UserProfileResponse {
  firstName?: string;
  lastName?: string;
  email?: string;
  avatar?: string;
}

export default function ProfilePage() {
  const router = useRouter();
  const {
    user,
    isAuthenticated,
    isLoading: authLoading,
    getAccessTokenSilently,
    logout,
  } = useAuth0();

  const [isLoading, setIsLoading] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [profile, setProfile] = useState<{
    firstName: string;
    lastName: string;
    email: string;
  }>({
    firstName: "",
    lastName: "",
    email: "",
  });
  const [formData, setFormData] = useState({
    firstName: "",
    lastName: "",
    email: "",
  });

  // Check authentication state
  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push("/");
    }
  }, [authLoading, isAuthenticated, router]);

  // Extract user ID from Auth0 sub claim (format: auth0|xyz)
  const getAuth0Id = () => {
    if (!user?.sub) return "";
    return user.sub.split("|")[1];
  };

  // Fetch user profile data from API
  useEffect(() => {
    if (isAuthenticated && user) {
      const fetchProfile = async () => {
        try {
          const accessToken = await getAccessTokenSilently();
          const userId = getAuth0Id();
          const response = (await getUserProfile(
            accessToken,
            userId
          )) as UserProfileResponse;

          // Update profile state with API data
          setProfile({
            firstName: response.firstName || "",
            lastName: response.lastName || "",
            email: response.email || "",
          });

          // Initialize form data
          setFormData({
            firstName: response.firstName || "",
            lastName: response.lastName || "",
            email: response.email || "",
          });
        } catch (error) {
          console.error("Failed to fetch profile:", error);
          toast.error("Failed to load profile data");
        }
      };

      fetchProfile();
    }
  }, [isAuthenticated, user, getAccessTokenSilently]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData({ ...formData, [name]: value });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!isAuthenticated) {
      toast.error("Not authenticated");
      return;
    }

    setIsLoading(true);
    try {
      // Only include changed fields
      const updates: Record<string, string> = {};
      if (formData.firstName !== profile.firstName)
        updates.firstName = formData.firstName;
      if (formData.lastName !== profile.lastName)
        updates.lastName = formData.lastName;
      if (formData.email !== profile.email) updates.email = formData.email;

      // Skip API call if nothing changed
      if (Object.keys(updates).length === 0) {
        setIsEditing(false);
        setIsLoading(false);
        return;
      }

      const accessToken = await getAccessTokenSilently();
      const userId = getAuth0Id();
      await updateUserProfile(accessToken, userId, updates);

      // Update profile state
      setProfile({ ...profile, ...updates });
      setIsEditing(false);
      toast.success("Profile updated successfully");
    } catch (error) {
      console.error("Failed to update profile:", error);
      toast.error("Failed to update profile");
    } finally {
      setIsLoading(false);
    }
  };

  const cancelEdit = () => {
    // Reset form data to current profile
    setFormData({
      firstName: profile.firstName,
      lastName: profile.lastName,
      email: profile.email,
    });
    setIsEditing(false);
  };

  const handleLogout = () => {
    logout({ logoutParams: { returnTo: window.location.origin } });
  };

  if (authLoading) {
    return (
      <div className="min-h-screen bg-dark-100 text-slate-200 flex items-center justify-center">
        Loading...
      </div>
    );
  }

  return (
    <main className="min-h-screen bg-dark-100 text-slate-200">
      <header className="sticky top-0 z-50 border-b border-dark-300 bg-dark-100/80 backdrop-blur-md">
        <div className="container mx-auto flex items-center justify-between p-4">
          <div className="flex items-center gap-3">
            <Link
              href="/"
              className="text-2xl font-bold tracking-tight text-slate-100"
            >
              Thread<span className="text-primary-500">Art</span>
            </Link>
          </div>
          <nav className="hidden space-x-8 md:flex">
            <Link
              href="/"
              className="text-slate-300 hover:text-white transition"
            >
              Home
            </Link>
            <Link
              href="/gallery"
              className="text-slate-300 hover:text-white transition"
            >
              Gallery
            </Link>
            <Link
              href="/create"
              className="text-slate-300 hover:text-white transition"
            >
              Create
            </Link>
            <Link
              href="/about"
              className="text-slate-300 hover:text-white transition"
            >
              About
            </Link>
          </nav>
          <div className="flex items-center space-x-4">
            <span className="text-slate-200">Welcome, {user?.name}</span>
            <button
              onClick={handleLogout}
              className="px-4 py-2 rounded-md border border-dark-300 text-slate-200 hover:bg-dark-300/50 transition"
            >
              Logout
            </button>
          </div>
        </div>
      </header>

      <div className="container mx-auto px-4 py-12">
        <div className="max-w-3xl mx-auto bg-dark-200 rounded-lg p-8 shadow-xl">
          <div className="flex justify-between items-center mb-8">
            <h1 className="text-3xl font-bold text-slate-100">Your Profile</h1>
            {!isEditing && (
              <button
                onClick={() => setIsEditing(true)}
                className="px-4 py-2 rounded-md bg-primary-600 text-white hover:bg-primary-500 transition"
              >
                Edit Profile
              </button>
            )}
          </div>

          <div className="flex flex-col md:flex-row gap-8">
            {user?.picture && (
              <div className="w-32 h-32 rounded-full overflow-hidden border-2 border-primary-500/30 shadow-lg">
                <Image
                  src={user.picture}
                  alt={user.name || "Profile picture"}
                  className="w-full h-full object-cover"
                  width={128}
                  height={128}
                />
              </div>
            )}

            <div className="flex-1">
              {!isEditing ? (
                <>
                  <div className="mb-4">
                    <h2 className="text-sm text-slate-400">Name</h2>
                    <p className="text-xl text-slate-100">
                      {profile.firstName} {profile.lastName}
                    </p>
                  </div>

                  <div className="mb-4">
                    <h2 className="text-sm text-slate-400">Email</h2>
                    <p className="text-xl text-slate-100">{profile.email}</p>
                  </div>

                  {user?.email_verified && (
                    <div className="inline-block px-2 py-1 bg-accent-teal/20 text-accent-teal text-sm rounded">
                      Email verified
                    </div>
                  )}
                </>
              ) : (
                <form onSubmit={handleSubmit} className="space-y-4">
                  <div>
                    <label
                      htmlFor="firstName"
                      className="block text-sm text-slate-400 mb-1"
                    >
                      First Name
                    </label>
                    <input
                      type="text"
                      id="firstName"
                      name="firstName"
                      value={formData.firstName}
                      onChange={handleInputChange}
                      className="w-full px-4 py-2 rounded-md bg-dark-300 border border-dark-400 text-slate-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
                    />
                  </div>

                  <div>
                    <label
                      htmlFor="lastName"
                      className="block text-sm text-slate-400 mb-1"
                    >
                      Last Name
                    </label>
                    <input
                      type="text"
                      id="lastName"
                      name="lastName"
                      value={formData.lastName}
                      onChange={handleInputChange}
                      className="w-full px-4 py-2 rounded-md bg-dark-300 border border-dark-400 text-slate-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
                    />
                  </div>

                  <div>
                    <label
                      htmlFor="email"
                      className="block text-sm text-slate-400 mb-1"
                    >
                      Email
                    </label>
                    <input
                      type="email"
                      id="email"
                      name="email"
                      value={formData.email}
                      onChange={handleInputChange}
                      className="w-full px-4 py-2 rounded-md bg-dark-300 border border-dark-400 text-slate-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
                    />
                  </div>

                  <div className="flex gap-3 pt-2">
                    <button
                      type="submit"
                      disabled={isLoading}
                      className="px-4 py-2 rounded-md bg-primary-600 text-white hover:bg-primary-500 transition disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      {isLoading ? "Saving..." : "Save Changes"}
                    </button>
                    <button
                      type="button"
                      onClick={cancelEdit}
                      className="px-4 py-2 rounded-md border border-dark-300 text-slate-200 hover:bg-dark-300/50 transition"
                    >
                      Cancel
                    </button>
                  </div>
                </form>
              )}
            </div>
          </div>

          <div className="mt-12 pt-8 border-t border-dark-300">
            <h2 className="text-2xl font-bold mb-4 text-slate-100">
              Your Thread Art Projects
            </h2>
            <div className="bg-dark-300 rounded-lg p-8 text-center shadow-inner">
              <p className="text-slate-400 mb-4">
                You haven&apos;t created any thread art projects yet.
              </p>
              <Link
                href="/create"
                className="px-6 py-2 rounded-md bg-primary-600 text-white hover:bg-primary-500 transition inline-block shadow-lg shadow-primary-900/20"
              >
                Create Your First Project
              </Link>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}
