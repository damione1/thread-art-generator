"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { User } from "../../types/user";

interface NavigationProps {
  user?: User | null;
}

export default function Navigation({ user }: NavigationProps) {
  const pathname = usePathname();

  const getLinkClass = (path: string) => {
    const isActive = pathname === path;
    return isActive
      ? "text-primary-400 hover:text-primary-300 transition"
      : "text-slate-300 hover:text-white transition";
  };

  return (
    <nav className="hidden space-x-8 md:flex">
      <Link href="/" className={getLinkClass("/")}>
        Home
      </Link>
      <Link href="/gallery" className={getLinkClass("/gallery")}>
        Gallery
      </Link>
      {user ? (
        <Link
          href="/dashboard/arts/new"
          className={getLinkClass("/dashboard/arts/new")}
        >
          Create
        </Link>
      ) : (
        <Link href="/about" className={getLinkClass("/about")}>
          About
        </Link>
      )}
    </nav>
  );
}
