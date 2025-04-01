"use client";

import React from "react";
import Head from "next/head";
import Header from "./Header";
import Footer from "./Footer";
import { User } from "../../types/user";
import { useUser } from "@/contexts/UserContext";

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
  // Use centralized user context instead of directly using auth0
  const { user: contextUser } = useUser();

  // Prefer context user over initialUser prop
  const user = contextUser.id ? contextUser : initialUser;

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
