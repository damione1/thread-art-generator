"use client";

import Link from "next/link";
import { User } from "../../types/user";
import { useAuth } from "@/hooks/useAuth";

interface UserMenuProps {
  user: User;
}

export default function UserMenu({ user }: UserMenuProps) {
  const { logoutUser } = useAuth();

  return (
    <>
      <span className="text-slate-200">Welcome, {user.name}</span>
      <Link
        href="/dashboard"
        className="px-4 py-2 rounded-md bg-primary-600 text-white hover:bg-primary-500 transition shadow-lg shadow-primary-900/20"
      >
        Dashboard
      </Link>
      <button
        onClick={logoutUser}
        className="px-4 py-2 rounded-md border border-dark-300 text-slate-200 hover:bg-dark-300/50 transition"
      >
        Logout
      </button>
    </>
  );
}
