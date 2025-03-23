import Link from "next/link";
import Layout from "../../components/layout/Layout";
import { auth0 } from "@/lib/auth0";
import { User } from "../../types/user";

export default async function DashboardPage() {
  // This will now use Auth0 to get the user session
  let user: User = {
    id: "",
    name: "User",
    email: "",
  };

  try {
    const session = await auth0.getSession();
    if (session?.user) {
      user = {
        id: session.user.sub || "",
        name: session.user.name || "User",
        email: session.user.email || "",
      };
    }
  } catch (error) {
    console.error("Failed to get user session:", error);
  }

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

          {/* Sample Project Card */}
          <div className="bg-dark-200 rounded-lg overflow-hidden shadow-lg">
            <div className="h-40 bg-dark-300 flex items-center justify-center">
              <div className="w-24 h-24 rounded-full border border-primary-400 relative">
                <div
                  className="absolute inset-0"
                  style={{
                    background:
                      "radial-gradient(circle, transparent 50%, transparent 56%), conic-gradient(from 0deg, rgba(90, 127, 255, 0) 0%, rgba(90, 127, 255, 0.1) 20%, rgba(90, 127, 255, 0.3) 40%, rgba(90, 127, 255, 0.7) 60%, rgba(90, 127, 255, 0.3) 80%, rgba(90, 127, 255, 0) 100%)",
                  }}
                ></div>
              </div>
            </div>
            <div className="p-4">
              <h3 className="text-lg font-semibold text-slate-100">
                Sample Project
              </h3>
              <p className="text-slate-400 text-sm mt-1">
                Created on {new Date().toLocaleDateString()}
              </p>
              <div className="flex justify-between mt-4">
                <Link
                  href="/project/sample"
                  className="text-primary-400 hover:text-primary-300 text-sm"
                >
                  View details
                </Link>
                <span className="text-accent-teal text-sm">Completed</span>
              </div>
            </div>
          </div>

          {/* Recent Activity Card */}
          <div className="bg-dark-200 rounded-lg p-6 shadow-lg">
            <h3 className="text-lg font-semibold text-slate-100 mb-4">
              Recent Activity
            </h3>
            <div className="space-y-4">
              <div className="flex items-start">
                <div className="w-2 h-2 rounded-full bg-primary-500 mt-2 mr-3"></div>
                <div>
                  <p className="text-slate-300 text-sm">
                    You created a new project
                  </p>
                  <p className="text-slate-400 text-xs mt-1">2 days ago</p>
                </div>
              </div>
              <div className="flex items-start">
                <div className="w-2 h-2 rounded-full bg-accent-teal mt-2 mr-3"></div>
                <div>
                  <p className="text-slate-300 text-sm">
                    Project &quot;Sunset&quot; completed
                  </p>
                  <p className="text-slate-400 text-xs mt-1">1 week ago</p>
                </div>
              </div>
              <div className="flex items-start">
                <div className="w-2 h-2 rounded-full bg-accent-purple mt-2 mr-3"></div>
                <div>
                  <p className="text-slate-300 text-sm">
                    You updated your profile
                  </p>
                  <p className="text-slate-400 text-xs mt-1">2 weeks ago</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
}
