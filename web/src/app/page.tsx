import Link from "next/link";
import Layout from "../components/layout/Layout";
import { User } from "../types/user";

export default async function Home() {
  const user: User | null = null; // placeholder for user - will be replaced with auth logic later

  return (
    <Layout user={user} title="ThreadArt - Create Beautiful Thread Art">
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
    </Layout>
  );
}
