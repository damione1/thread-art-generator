"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Layout from "../../components/layout/Layout";
import { User } from "../../types/user";
import Image from "next/image";
import { User as ProtoUser } from "@/lib/pb/user_pb";
import { getCurrentUser } from "@/lib/grpc-client";
import ProfileEditor from "@/components/profile/ProfileEditor";

export default function ProfilePage() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [user, setUser] = useState<User>({ id: "", name: "User", email: "" });
  const [profileData, setProfileData] = useState<ProtoUser | null>(null);
  const [showEditor, setShowEditor] = useState(false);

  useEffect(() => {
    // Get the session data from the auth cookie
    async function getSessionData() {
      try {
        const response = await fetch("/api/auth/session");
        if (!response.ok) {
          throw new Error("Failed to get session");
        }

        const data = await response.json();

        if (!data.user) {
          // Not authenticated, redirect to login
          router.push("/api/auth/login");
          return;
        }

        setUser({
          id: data.user.sub || "",
          name: data.user.name || "User",
          email: data.user.email || "",
        });

        try {
          // Use our new gRPC client with token caching
          const userData = await getCurrentUser();
          setProfileData(userData);
        } catch (err) {
          console.error("Error fetching user profile:", err);
          setError(
            `Error: ${err instanceof Error ? err.message : "Unknown error"}`
          );
        }
      } catch (err) {
        console.error("Error fetching session:", err);
        setError(
          `Error: ${err instanceof Error ? err.message : "Unknown error"}`
        );
      } finally {
        setLoading(false);
      }
    }

    getSessionData();
  }, [router]);

  return (
    <Layout user={user} title="Profile - ThreadArt">
      <div className="container mx-auto p-6">
        <h1 className="text-3xl font-bold mb-6">User Profile</h1>

        {loading && (
          <div className="bg-dark-200 p-8 rounded-lg text-center shadow-inner">
            <div className="flex justify-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-slate-100"></div>
            </div>
            <p className="text-slate-400 mt-4">Loading profile data...</p>
          </div>
        )}

        {error && (
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-6">
            <span className="block sm:inline">{error}</span>
          </div>
        )}

        {profileData && !loading && (
          <div className="bg-dark-200 rounded-lg p-6 shadow-lg">
            <div className="flex items-center justify-between mb-6">
              <div className="flex items-center">
                {profileData.avatar && (
                  <Image
                    src={profileData.avatar}
                    alt={`${profileData.firstName}'s avatar`}
                    className="w-24 h-24 rounded-full mr-6 object-cover"
                    width={96}
                    height={96}
                  />
                )}
                <div>
                  <h2 className="text-2xl font-semibold text-slate-100">
                    {profileData.firstName} {profileData.lastName}
                  </h2>
                  <p className="text-slate-400">{profileData.email}</p>
                </div>
              </div>
              <button
                onClick={() => setShowEditor(!showEditor)}
                className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
              >
                {showEditor ? "Cancel" : "Edit Profile"}
              </button>
            </div>

            {showEditor ? (
              <ProfileEditor
                userData={profileData}
                onUpdate={(updatedUser) => {
                  setProfileData(updatedUser);
                  setShowEditor(false);
                }}
              />
            ) : (
              <div className="border-t border-dark-300 pt-4">
                <h3 className="text-lg font-semibold mb-3 text-slate-100">
                  User Details
                </h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <p className="text-slate-500 text-sm">First Name</p>
                    <p className="text-slate-100">{profileData.firstName}</p>
                  </div>
                  <div>
                    <p className="text-slate-500 text-sm">Last Name</p>
                    <p className="text-slate-100">{profileData.lastName}</p>
                  </div>
                  <div>
                    <p className="text-slate-500 text-sm">Email</p>
                    <p className="text-slate-100">{profileData.email}</p>
                  </div>
                  <div>
                    <p className="text-slate-500 text-sm">User ID</p>
                    <p className="text-sm break-all text-slate-100">
                      {profileData.name}
                    </p>
                  </div>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </Layout>
  );
}
