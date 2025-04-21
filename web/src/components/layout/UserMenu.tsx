"use client";

import Link from "next/link";
import { User } from "../../types/user";
import { useAuth } from "@/hooks/useAuth";
import { useState, useRef, useEffect } from "react";

interface UserMenuProps {
  user: User;
}

export default function UserMenu({ user }: UserMenuProps) {
  const { logoutUser } = useAuth();
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    }

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 px-4 py-2 rounded-md bg-primary-600 text-white hover:bg-primary-500 transition shadow-lg shadow-primary-900/20"
      >
        <span>{user.name}</span>
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className={`h-4 w-4 transition-transform ${
            isOpen ? "rotate-180" : ""
          }`}
          viewBox="0 0 20 20"
          fill="currentColor"
        >
          <path
            fillRule="evenodd"
            d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
            clipRule="evenodd"
          />
        </svg>
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-48 py-2 bg-dark-100 border border-dark-300 rounded-md shadow-xl z-50">
          <Link
            href="/dashboard"
            className="block px-4 py-2 text-slate-200 hover:bg-dark-300/50 transition"
            onClick={() => setIsOpen(false)}
          >
            Dashboard
          </Link>
          <Link
            href="/dashboard/arts/new"
            className="block px-4 py-2 text-slate-200 hover:bg-dark-300/50 transition"
            onClick={() => setIsOpen(false)}
          >
            Create New Art
          </Link>
          <Link
            href="/profile"
            className="block px-4 py-2 text-slate-200 hover:bg-dark-300/50 transition"
            onClick={() => setIsOpen(false)}
          >
            Profile
          </Link>
          <hr className="my-1 border-dark-300" />
          <button
            onClick={() => {
              logoutUser();
              setIsOpen(false);
            }}
            className="block w-full text-left px-4 py-2 text-slate-200 hover:bg-dark-300/50 transition"
          >
            Logout
          </button>
        </div>
      )}
    </div>
  );
}
