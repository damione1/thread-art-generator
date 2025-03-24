"use client";

import Link from "next/link";
import Layout from "../../components/layout/Layout";
import { User } from "../../types/user";
import { listArts } from "@/lib/grpc-client";
import { Art } from "@/lib/pb/art_pb";
import { formatDistanceToNow } from "date-fns";
import Image from "next/image";
import { useState, useEffect } from "react";
import { useAuth } from "@/hooks/useAuth";

// Helper function to get status display information
function getStatusInfo(status: number) {
  switch (status) {
    case 3: // ART_STATUS_COMPLETE
      return { text: "Completed", color: "text-accent-teal" };
    case 2: // ART_STATUS_PROCESSING
      return { text: "Processing", color: "text-amber-400" };
    case 1: // ART_STATUS_PENDING_IMAGE
      return { text: "Pending Image", color: "text-primary-400" };
    case 4: // ART_STATUS_FAILED
      return { text: "Failed", color: "text-red-500" };
    case 5: // ART_STATUS_ARCHIVED
      return { text: "Archived", color: "text-slate-400" };
    default:
      return { text: "Unknown", color: "text-slate-400" };
  }
}

export default function DashboardPage() {
  const { user: authUser } = useAuth();
  const [user, setUser] = useState<User>({
    id: "",
    name: "User",
    email: "",
  });
  const [userArts, setUserArts] = useState<Art[]>([]);
  const [errorMessage, setErrorMessage] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (authUser) {
      setUser({
        id: authUser.sub || "",
        name: authUser.name || "User",
        email: authUser.email || "",
      });
    }
  }, [authUser]);

  useEffect(() => {
    async function fetchData() {
      try {
        if (!user.id) return;

        // Fetch user's arts if we have a user ID
        const parentResource = `users/${user.id}`;
        const artsResponse = await listArts(parentResource, 10);
        setUserArts(artsResponse.arts || []);
      } catch (error) {
        console.error("Failed to get user data or arts:", error);
        setErrorMessage(
          error instanceof Error ? error.message : "Unknown error occurred"
        );
      } finally {
        setLoading(false);
      }
    }

    fetchData();
  }, [user.id]);

  return (
    <Layout user={user} title="Dashboard - ThreadArt">
      <div className="container mx-auto px-4 py-12">
        <div className="mb-8">
          <h1 className="text-3xl font-bold mb-4">Welcome, {user.name}!</h1>
          <p className="text-gray-600 mb-6">
            This is your dashboard where you can manage your Thread Art
            projects.
          </p>
          <div className="flex flex-wrap gap-4">
            <Link
              href="/dashboard/arts/new"
              className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
            >
              Create New Art
            </Link>
            <Link
              href="/profile"
              className="px-4 py-2 bg-gray-100 text-gray-800 rounded hover:bg-gray-200 transition-colors"
            >
              View Profile
            </Link>
          </div>
        </div>

        {errorMessage && (
          <div className="mb-6 p-4 bg-red-100 text-red-700 rounded">
            Error loading arts: {errorMessage}
          </div>
        )}

        {loading ? (
          <div className="flex justify-center items-center min-h-[200px]">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary-500"></div>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {/* Project Card - Create New */}
            <div className="bg-dark-200 rounded-lg p-6 border border-dark-300 border-dashed flex flex-col items-center justify-center text-center h-64 hover:bg-dark-300/50 transition cursor-pointer">
              <div className="w-16 h-16 rounded-full bg-primary-900/30 flex items-center justify-center mb-4">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-8 w-8 text-primary-500"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 6v6m0 0v6m0-6h6m-6 0H6"
                  />
                </svg>
              </div>
              <h3 className="text-xl font-semibold text-slate-100 mb-2">
                Create New Art
              </h3>
              <p className="text-slate-400 mb-6">
                Start a new thread art project
              </p>
              <Link
                href="/dashboard/arts/new"
                className="text-primary-500 font-medium hover:text-primary-400 transition"
              >
                Get Started →
              </Link>
            </div>

            {/* Dynamic Art Projects */}
            {userArts.length > 0 ? (
              userArts.map((art) => {
                const statusInfo = getStatusInfo(art.status);
                const created = art.createTime
                  ? new Date(Number(art.createTime.seconds) * 1000)
                  : new Date();
                const resourceId = art.name.split("/").pop();

                return (
                  <div
                    key={art.name}
                    className="bg-dark-200 rounded-lg overflow-hidden shadow-lg"
                  >
                    <div className="h-40 bg-dark-300 flex items-center justify-center">
                      {art.imageUrl ? (
                        <Image
                          src={art.imageUrl}
                          alt={art.title}
                          className="h-full w-full object-cover"
                          width={100}
                          height={100}
                        />
                      ) : (
                        <div className="w-24 h-24 rounded-full border border-primary-400 relative">
                          <div
                            className="absolute inset-0"
                            style={{
                              background:
                                "radial-gradient(circle, transparent 50%, transparent 56%), conic-gradient(from 0deg, rgba(90, 127, 255, 0) 0%, rgba(90, 127, 255, 0.1) 20%, rgba(90, 127, 255, 0.3) 40%, rgba(90, 127, 255, 0.7) 60%, rgba(90, 127, 255, 0.3) 80%, rgba(90, 127, 255, 0) 100%)",
                            }}
                          ></div>
                        </div>
                      )}
                    </div>
                    <div className="p-4">
                      <h3 className="text-lg font-semibold text-slate-100">
                        {art.title}
                      </h3>
                      <p className="text-slate-400 text-sm mt-1">
                        Created{" "}
                        {formatDistanceToNow(created, { addSuffix: true })}
                      </p>
                      <div className="flex justify-between mt-4">
                        <Link
                          href={`/dashboard/arts/${resourceId}`}
                          className="text-primary-400 hover:text-primary-300 text-sm"
                        >
                          View details
                        </Link>
                        <span className={`${statusInfo.color} text-sm`}>
                          {statusInfo.text}
                        </span>
                      </div>
                    </div>
                  </div>
                );
              })
            ) : (
              <div className="bg-dark-200 rounded-lg p-6 col-span-2">
                <p className="text-slate-300 text-center">
                  No art projects found. Create your first project to get
                  started!
                </p>
              </div>
            )}

            {/* Recent Activity Card */}
            <div className="bg-dark-200 rounded-lg p-6 shadow-lg">
              <h3 className="text-lg font-semibold text-slate-100 mb-4">
                Recent Activity
              </h3>
              {userArts.length > 0 ? (
                <div className="space-y-4">
                  {userArts.slice(0, 3).map((art, index) => {
                    const created = art.createTime
                      ? new Date(Number(art.createTime.seconds) * 1000)
                      : new Date();
                    const timeAgo = formatDistanceToNow(created, {
                      addSuffix: true,
                    });
                    const colorClasses = [
                      "bg-primary-500",
                      "bg-accent-teal",
                      "bg-accent-purple",
                    ];

                    return (
                      <div
                        key={`activity-${art.name}`}
                        className="flex items-start"
                      >
                        <div
                          className={`w-2 h-2 rounded-full ${
                            colorClasses[index % 3]
                          } mt-2 mr-3`}
                        ></div>
                        <div>
                          <p className="text-slate-300 text-sm">
                            You created &ldquo;{art.title}&rdquo;
                          </p>
                          <p className="text-slate-400 text-xs mt-1">
                            {timeAgo}
                          </p>
                        </div>
                      </div>
                    );
                  })}
                </div>
              ) : (
                <p className="text-slate-400 text-sm">No recent activity</p>
              )}
            </div>
          </div>
        )}
      </div>
    </Layout>
  );
}
