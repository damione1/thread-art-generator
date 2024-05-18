"use client";
import "./globals.css";
import "./data-tables-css.css";
import "./satoshi.css";
import ThemeProvider from "./template-provider";
import { SessionProvider } from "next-auth/react";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="dark:bg-boxdark-2 dark:text-bodydark">
     <SessionProvider><ThemeProvider>{children}</ThemeProvider></SessionProvider>
    </div>
  );
}
