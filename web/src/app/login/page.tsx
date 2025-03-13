import { auth0 } from "@/lib/auth0";
import { redirect } from "next/navigation";
import Link from "next/link";

export default async function LoginPage() {
  const { user } = (await auth0.getSession()) || { user: null };

  // If user is already logged in, redirect to dashboard
  if (user) {
    redirect("/dashboard");
  }

  return (
    <main className="min-h-screen bg-dark-100">
      <div className="container mx-auto px-4 py-16">
        <div className="max-w-md mx-auto">
          <div className="text-center mb-8">
            <Link href="/" className="inline-block">
              <h1 className="text-3xl font-bold text-slate-100">
                Thread<span className="text-primary-500">Art</span>
              </h1>
            </Link>
            <p className="mt-2 text-slate-300">Sign in to your account</p>
          </div>

          <div className="bg-dark-200 rounded-lg shadow-xl p-8">
            <div className="space-y-6">
              <div className="space-y-4">
                <a
                  href="/auth/login"
                  className="flex items-center justify-center w-full px-4 py-3 rounded-md bg-primary-600 text-white hover:bg-primary-500 transition shadow-lg shadow-primary-900/20"
                >
                  <span className="text-center font-medium">Sign in</span>
                </a>

                <div className="relative flex items-center justify-center">
                  <div className="border-t border-dark-300 w-full"></div>
                  <div className="absolute bg-dark-200 px-3 text-sm text-slate-400">
                    or
                  </div>
                </div>

                <a
                  href="/auth/login?screen_hint=signup"
                  className="flex items-center justify-center w-full px-4 py-3 rounded-md border border-dark-300 text-slate-200 hover:bg-dark-300/50 transition"
                >
                  <span className="text-center font-medium">
                    Create an account
                  </span>
                </a>
              </div>

              <div className="text-center text-sm text-slate-400">
                <p>
                  By signing in, you agree to our{" "}
                  <Link
                    href="/terms"
                    className="text-primary-400 hover:text-primary-300"
                  >
                    Terms of Service
                  </Link>{" "}
                  and{" "}
                  <Link
                    href="/privacy"
                    className="text-primary-400 hover:text-primary-300"
                  >
                    Privacy Policy
                  </Link>
                </p>
              </div>
            </div>
          </div>

          <div className="mt-8 text-center">
            <Link
              href="/"
              className="text-slate-400 hover:text-slate-300 text-sm"
            >
              ← Back to home
            </Link>
          </div>
        </div>
      </div>
    </main>
  );
}
