"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Layout from "../../components/layout/Layout";
import Image from "next/image";
import ProfileEditor from "@/components/profile/ProfileEditor";
import { useUser } from "@/contexts/UserContext";

export default function ProfilePage() {
  const router = useRouter();
  const { user, profile, loading, error, refreshUserData } = useUser();
  const [showEditor, setShowEditor] = useState(false);

  // If not loaded yet, show loading state
  if (loading) {
    return (
      <Layout title="Profile - ThreadArt">
        <div className="container mx-auto p-6">
          <div className="flex justify-center items-center min-h-[200px]">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary-500"></div>
          </div>
        </div>
      </Layout>
    );
  }

  // If error occurred, show error state
  if (error) {
    return (
      <Layout title="Profile - ThreadArt">
        <div className="container mx-auto p-6">
          <div className="bg-red-100 text-red-700 p-4 rounded mb-4">
            {error}
          </div>
          <button
            onClick={() => refreshUserData()}
            className="px-4 py-2 bg-blue-600 text-white rounded"
          >
            Try Again
          </button>
        </div>
      </Layout>
    );
  }

  return (
    <Layout user={user} title="Profile - ThreadArt">
      <div className="container mx-auto px-4 py-12">
        <div className="max-w-4xl mx-auto bg-dark-200 rounded-lg shadow-lg overflow-hidden">
          {/* Profile header */}
          <div className="p-6 sm:p-8 bg-dark-300">
            <div className="flex flex-col sm:flex-row items-center gap-6">
              <div className="w-32 h-32 rounded-full overflow-hidden border-4 border-dark-100">
                {profile?.avatar || user.picture ? (
                  <Image
                    src={profile?.avatar || user.picture || ""}
                    alt={user.name}
                    width={128}
                    height={128}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full bg-primary-900/30 flex items-center justify-center text-primary-500 text-3xl font-bold">
                    {user.name?.charAt(0) || "U"}
                  </div>
                )}
              </div>
              <div className="text-center sm:text-left">
                <h1 className="text-2xl sm:text-3xl font-bold text-white">
                  {profile?.firstName && profile?.lastName
                    ? `${profile.firstName} ${profile.lastName}`
                    : user.name}
                </h1>
                <p className="text-slate-400 mt-2">{user.email}</p>

                {!showEditor && (
                  <button
                    onClick={() => setShowEditor(true)}
                    className="mt-4 px-4 py-2 bg-primary-600 text-white rounded hover:bg-primary-500 transition"
                  >
                    Edit Profile
                  </button>
                )}
              </div>
            </div>
          </div>

          {/* Profile content */}
          <div className="p-6 sm:p-8">
            {showEditor ? (
              <ProfileEditor
                userData={profile || undefined}
                onCancel={() => setShowEditor(false)}
                onSuccess={() => {
                  setShowEditor(false);
                  refreshUserData();
                }}
              />
            ) : (
              <div className="space-y-6">
                <div>
                  <h2 className="text-xl font-semibold text-white mb-3">
                    Profile Information
                  </h2>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div className="p-4 bg-dark-300 rounded">
                      <p className="text-slate-400 text-sm">First Name</p>
                      <p className="text-white mt-1">
                        {profile?.firstName || "-"}
                      </p>
                    </div>
                    <div className="p-4 bg-dark-300 rounded">
                      <p className="text-slate-400 text-sm">Last Name</p>
                      <p className="text-white mt-1">
                        {profile?.lastName || "-"}
                      </p>
                    </div>
                    <div className="p-4 bg-dark-300 rounded">
                      <p className="text-slate-400 text-sm">Email</p>
                      <p className="text-white mt-1">{user.email}</p>
                    </div>
                    <div className="p-4 bg-dark-300 rounded">
                      <p className="text-slate-400 text-sm">User ID</p>
                      <p className="text-white mt-1 truncate">
                        {user.id || "-"}
                      </p>
                    </div>
                  </div>
                </div>

                {/* Account Actions */}
                <div>
                  <h2 className="text-xl font-semibold text-white mb-3">
                    Account Actions
                  </h2>
                  <div className="flex flex-wrap gap-3">
                    <button
                      onClick={() => router.push("/dashboard")}
                      className="px-4 py-2 bg-dark-300 text-white rounded hover:bg-dark-400 transition"
                    >
                      Back to Dashboard
                    </button>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </Layout>
  );
}
