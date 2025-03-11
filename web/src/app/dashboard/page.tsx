import { auth0 } from "@/lib/auth0";
import { redirect } from "next/navigation";
import Link from "next/link";

export default async function DashboardPage() {
  const { user } = (await auth0.getSession()) || { user: null };

  // If user is not logged in, redirect to login
  if (!user) {
    redirect("/login");
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
              href="/dashboard"
              className="text-primary-400 hover:text-primary-300 transition"
            >
              Dashboard
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
              href="/profile"
              className="text-slate-300 hover:text-white transition"
            >
              Profile
            </Link>
          </nav>
          <div className="flex items-center space-x-4">
            <span className="text-slate-200">Welcome, {user.name}</span>
            <a
              href="/api/auth/logout"
              className="px-4 py-2 rounded-md border border-dark-300 text-slate-200 hover:bg-dark-300/50 transition"
            >
              Logout
            </a>
          </div>
        </div>
      </header>

      <div className="container mx-auto px-4 py-12">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-slate-100">Dashboard</h1>
          <p className="text-slate-400 mt-2">Manage your thread art projects</p>
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
                  d="M12 4v16m8-8H4"
                />
              </svg>
            </div>
            <h3 className="text-xl font-semibold text-slate-100">
              Create New Project
            </h3>
            <p className="text-slate-400 mt-2">
              Start a new thread art creation
            </p>
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
    </main>
  );
}
