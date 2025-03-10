import React from "react";
import Link from "next/link";
import Head from "next/head";

interface LayoutProps {
  children: React.ReactNode;
  title?: string;
}

const Layout: React.FC<LayoutProps> = ({
  children,
  title = "Thread Art Generator",
}) => {
  return (
    <div className="min-h-screen bg-black text-white">
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

      <header className="sticky top-0 z-50 border-b border-dark-300 bg-black/80 backdrop-blur-md">
        <div className="container mx-auto flex items-center justify-between p-4">
          <div className="flex items-center gap-3">
            <Link
              href="/"
              className="text-2xl font-bold tracking-tight text-white"
            >
              Thread<span className="text-primary-400">Art</span>
            </Link>
          </div>
          <nav className="hidden space-x-8 md:flex">
            <Link
              href="/"
              className="text-gray-300 hover:text-white transition"
            >
              Home
            </Link>
            <Link
              href="/gallery"
              className="text-gray-300 hover:text-white transition"
            >
              Gallery
            </Link>
            <Link
              href="/create"
              className="text-gray-300 hover:text-white transition"
            >
              Create
            </Link>
            <Link
              href="/about"
              className="text-gray-300 hover:text-white transition"
            >
              About
            </Link>
          </nav>
          <div className="flex items-center space-x-4">
            <Link
              href="/login"
              className="px-4 py-2 rounded-md text-white hover:text-primary-300 transition"
            >
              Log in
            </Link>
            <Link
              href="/register"
              className="px-4 py-2 rounded-md bg-primary-600 text-white hover:bg-primary-500 transition"
            >
              Sign up
            </Link>
          </div>
        </div>
      </header>

      <main>{children}</main>

      <footer className="border-t border-dark-300 bg-black">
        <div className="container mx-auto p-6">
          <div className="flex flex-col md:flex-row justify-between items-center">
            <div className="mb-4 md:mb-0">
              <Link
                href="/"
                className="text-xl font-bold tracking-tight text-white"
              >
                Thread<span className="text-primary-400">Art</span>
              </Link>
              <p className="mt-2 text-sm text-gray-400">
                Generate beautiful thread art from images
              </p>
            </div>
            <div className="flex space-x-6">
              <a href="#" className="text-gray-400 hover:text-white transition">
                Privacy
              </a>
              <a href="#" className="text-gray-400 hover:text-white transition">
                Terms
              </a>
              <a href="#" className="text-gray-400 hover:text-white transition">
                Contact
              </a>
            </div>
          </div>
          <div className="mt-8 border-t border-dark-300 pt-8 text-center text-sm text-gray-400">
            © {new Date().getFullYear()} ThreadArt. All rights reserved.
          </div>
        </div>
      </footer>
    </div>
  );
};

export default Layout;
