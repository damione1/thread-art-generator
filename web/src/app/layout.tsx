import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "../styles/globals.css";
import { Toaster } from "react-hot-toast";
import ClientProviders from "../components/auth/ClientProviders";

const inter = Inter({
  subsets: ["latin"],
  weight: ["300", "400", "500", "600", "700"],
  display: "swap",
  variable: "--font-inter",
});

export const metadata: Metadata = {
  title: "Thread Art Generator",
  description: "Transform images into thread art portraits",
};

// Root layout is a server component and can't contain client-side code
// So we use a ClientProviders component to wrap the entire app
export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body
        className={`min-h-screen bg-dark-100 text-slate-200 ${inter.className}`}
      >
        <ClientProviders>{children}</ClientProviders>
        <Toaster
          position="top-right"
          toastOptions={{
            style: {
              background: "#333",
              color: "#fff",
              borderRadius: "8px",
            },
          }}
        />
      </body>
    </html>
  );
}
