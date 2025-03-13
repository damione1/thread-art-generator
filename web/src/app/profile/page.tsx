import { auth0 } from "@/lib/auth0";
import { redirect } from "next/navigation";
import Link from "next/link";
import Image from "next/image";

export default async function ProfilePage() {
  const { user } = (await auth0.getSession()) || { user: null };

  if (!user) {
    redirect("/auth/login");
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
            <span className="text-slate-200">Welcome, {user.name}</span>
            <a
              href="/auth/logout"
              className="px-4 py-2 rounded-md border border-dark-300 text-slate-200 hover:bg-dark-300/50 transition"
            >
              Logout
            </a>
          </div>
        </div>
      </header>

      <div className="container mx-auto px-4 py-12">
        <div className="max-w-3xl mx-auto bg-dark-200 rounded-lg p-8 shadow-xl">
          <h1 className="text-3xl font-bold mb-8 text-slate-100">
            Your Profile
          </h1>

          <div className="flex flex-col md:flex-row gap-8">
            {user.picture && (
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
              <div className="mb-4">
                <h2 className="text-sm text-slate-400">Name</h2>
                <p className="text-xl text-slate-100">{user.name}</p>
              </div>

              <div className="mb-4">
                <h2 className="text-sm text-slate-400">Email</h2>
                <p className="text-xl text-slate-100">{user.email}</p>
              </div>

              {user.email_verified && (
                <div className="inline-block px-2 py-1 bg-accent-teal/20 text-accent-teal text-sm rounded">
                  Email verified
                </div>
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
