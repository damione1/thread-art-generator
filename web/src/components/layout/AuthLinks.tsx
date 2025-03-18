"use client";

import Link from "next/link";

export default function AuthLinks() {
  return (
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
  );
}
