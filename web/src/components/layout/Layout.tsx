"use client";

import React, { useEffect, useState } from "react";
import Head from "next/head";
import Header from "./Header";
import Footer from "./Footer";
import { User } from "../../types/user";
import { useAuth } from "@/hooks/useAuth";

interface LayoutProps {
  children: React.ReactNode;
  title?: string;
  user?: User | null;
}

const Layout: React.FC<LayoutProps> = ({
  children,
  title = "Thread Art Generator",
  user: initialUser = null,
}) => {
  const { user: authUser, isLoading } = useAuth();
  const [user, setUser] = useState<User | null>(initialUser);

  // Update user if Auth0 user becomes available
  useEffect(() => {
    if (authUser && !isLoading) {
      setUser({
        id: authUser.sub || "",
        name: authUser.name || "User",
        email: authUser.email || "",
      });
    }
  }, [authUser, isLoading]);

  return (
    <div className="min-h-screen bg-dark-100 text-white flex flex-col">
      <Head>
        <title>{title}</title>
        <meta
          name="description"
          content="Generate beautiful thread art patterns"
        />
        <link rel="icon" href="/favicon.ico" />
        <link
          href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap"
          rel="stylesheet"
        />
      </Head>

      <Header user={user} />

      <main className="flex-grow">{children}</main>

      <Footer />
    </div>
  );
};

export default Layout;
