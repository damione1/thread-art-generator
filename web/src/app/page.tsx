import { auth0 } from "@/lib/auth0";
import Link from "next/link";

export default async function Home() {
  const { user } = (await auth0.getSession()) || { user: null };

  return (
    <main>
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
            {!user ? (
              <>
                <Link
                  href="/login"
                  className="px-4 py-2 rounded-md text-slate-200 hover:text-primary-300 transition"
                >
                  Log in
                </Link>
                <Link
                  href="/login"
                  className="px-4 py-2 rounded-md bg-primary-600 text-white hover:bg-primary-500 transition shadow-lg shadow-primary-900/20"
                >
                  Sign up
                </Link>
              </>
            ) : (
              <>
                <span className="text-slate-200">Welcome, {user.name}</span>
                <Link
                  href="/dashboard"
                  className="px-4 py-2 rounded-md bg-primary-600 text-white hover:bg-primary-500 transition shadow-lg shadow-primary-900/20"
                >
                  Dashboard
                </Link>
                <a
                  href="/api/auth/logout"
                  className="px-4 py-2 rounded-md border border-dark-300 text-slate-200 hover:bg-dark-300/50 transition"
                >
                  Logout
                </a>
              </>
            )}
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="relative overflow-hidden">
        {/* Background gradient */}
        <div className="absolute inset-0 bg-gradient-to-b from-dark-100 via-dark-200 to-primary-950"></div>

        <div className="container mx-auto px-4 py-20 relative z-10">
          <div className="flex flex-col lg:flex-row items-center gap-12">
            <div className="w-full lg:w-1/2 text-center lg:text-left">
              <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold mb-6 leading-tight">
                Transform Images into Stunning{" "}
                <span className="text-primary-500 bg-gradient-to-r from-primary-500 to-accent-purple bg-clip-text text-transparent">
                  Thread Art
                </span>
              </h1>
              <p className="text-xl text-slate-300 mb-8 max-w-2xl">
                Our algorithm converts your photos into beautiful thread
                patterns that can be created on a circular frame with just a
                single thread.
              </p>
              <div className="flex flex-col sm:flex-row gap-4 justify-center lg:justify-start">
                {user ? (
                  <Link
                    href="/dashboard"
                    className="px-8 py-3 rounded-md bg-primary-600 text-white hover:bg-primary-500 transition text-lg font-medium shadow-lg shadow-primary-900/20"
                  >
                    Go to Dashboard
                  </Link>
                ) : (
                  <Link
                    href="/login"
                    className="px-8 py-3 rounded-md bg-primary-600 text-white hover:bg-primary-500 transition text-lg font-medium shadow-lg shadow-primary-900/20"
                  >
                    Get Started
                  </Link>
                )}
                <Link
                  href="/gallery"
                  className="px-8 py-3 rounded-md border border-dark-300 text-slate-200 hover:bg-dark-300/50 transition text-lg font-medium"
                >
                  View Gallery
                </Link>
              </div>
            </div>
            <div className="w-full lg:w-1/2">
              <div className="relative h-[400px] lg:h-[500px] w-full rounded-lg overflow-hidden shadow-2xl shadow-primary-900/20">
                {/* Placeholder for thread art image */}
                <div className="absolute inset-0 bg-dark-300 flex items-center justify-center">
                  <div className="w-[80%] h-[80%] rounded-full border border-primary-400 relative">
                    <div className="absolute inset-0 flex items-center justify-center">
                      <span className="text-primary-300 opacity-50">
                        Thread Art Preview
                      </span>
                    </div>
                    {/* Simulated thread lines */}
                    <div
                      className="absolute inset-0"
                      style={{
                        background:
                          "radial-gradient(circle, transparent 50%, transparent 56%), conic-gradient(from 0deg, rgba(90, 127, 255, 0) 0%, rgba(90, 127, 255, 0.1) 20%, rgba(90, 127, 255, 0.3) 40%, rgba(90, 127, 255, 0.7) 60%, rgba(90, 127, 255, 0.3) 80%, rgba(90, 127, 255, 0) 100%)",
                      }}
                    ></div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-dark-300 bg-dark-200">
        <div className="container mx-auto p-6">
          <div className="flex flex-col md:flex-row justify-between items-center">
            <div className="mb-4 md:mb-0">
              <Link
                href="/"
                className="text-xl font-bold tracking-tight text-slate-100"
              >
                Thread<span className="text-primary-500">Art</span>
              </Link>
              <p className="mt-2 text-sm text-slate-400">
                Generate beautiful thread art from images
              </p>
            </div>
            <div className="flex space-x-6">
              <a
                href="#"
                className="text-slate-400 hover:text-white transition"
              >
                Privacy
              </a>
              <a
                href="#"
                className="text-slate-400 hover:text-white transition"
              >
                Terms
              </a>
              <a
                href="#"
                className="text-slate-400 hover:text-white transition"
              >
                Contact
              </a>
            </div>
          </div>
          <div className="mt-8 border-t border-dark-300 pt-8 text-center text-sm text-slate-400">
            © {new Date().getFullYear()} ThreadArt. All rights reserved.
          </div>
        </div>
      </footer>
    </main>
  );
}
