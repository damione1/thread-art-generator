"use client";

import Link from "next/link";
import Layout from "../../components/layout/Layout";
import { listArts } from "@/lib/grpc-client";
import { Art } from "@/lib/pb/art_pb";
import { formatDistanceToNow } from "date-fns";
import Image from "next/image";
import { useState, useEffect } from "react";
import { useUser } from "@/contexts/UserContext";
import { getErrorMessage } from "@/utils/errorUtils";
import { getStatusInfo } from "@/utils/artUtils";

export default function DashboardPage() {
  // Use centralized user context
  const { user, loading: userLoading } = useUser();
  const [userArts, setUserArts] = useState<Art[]>([]);
  const [errorMessage, setErrorMessage] = useState("");
  const [loading, setLoading] = useState(true);

  // Fetch arts when user is available
  useEffect(() => {
    async function fetchArts() {
      try {
        if (!user.id) return;

        setLoading(true);
        // Fetch user's arts
        const parentResource = `users/${user.id}`;
        const artsResponse = await listArts(parentResource, 10);
        setUserArts(artsResponse.arts || []);
        setErrorMessage("");
      } catch (error) {
        console.error("Failed to get arts:", error);
        setErrorMessage(getErrorMessage(error));
      } finally {
        setLoading(false);
      }
    }

    if (user.id) {
      fetchArts();
    }
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
        </div>

        {errorMessage && (
          <div className="mb-6 p-4 bg-red-100 text-red-700 rounded">
            Error loading arts: {errorMessage}
          </div>
        )}

        {loading || userLoading ? (
          <div className="flex justify-center items-center min-h-[200px]">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary-500"></div>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {/* Project Card - Create New (fully clickable) */}
            <Link
              href="/dashboard/arts/new"
              className="bg-dark-200 rounded-lg p-6 border border-dark-300 border-dashed flex flex-col items-center justify-center text-center h-64 hover:bg-dark-300/50 transition cursor-pointer group"
            >
              <div className="w-16 h-16 rounded-full bg-primary-900/30 flex items-center justify-center mb-4 group-hover:bg-primary-900/40 transition">
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
              <span className="text-primary-500 font-medium group-hover:text-primary-400 transition">
                Get Started →
              </span>
            </Link>

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
                    <div className="h-40 bg-dark-300 flex items-center justify-center p-3">
                      {art.imageUrl ? (
                        <div className="w-full h-full relative flex items-center justify-center">
                          <div className="aspect-square h-full relative rounded-full overflow-hidden border-2 border-primary-400/30">
                            <Image
                              src={art.imageUrl}
                              alt={art.title}
                              className="object-cover"
                              fill
                              sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
                            />
                          </div>
                        </div>
                      ) : (
                        <div className="w-32 h-32 rounded-full border border-primary-400 relative">
                          <div
                            className="absolute inset-0 rounded-full"
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
