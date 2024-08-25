import Link from "next/link";

export default function PublicLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="dark:bg-boxdark-2 dark:text-bodydark min-h-screen">
      <header className="text-gray-900 dark:text-gray-500 body-font">
        <div className="container mx-auto flex flex-wrap p-5 flex-col md:flex-row items-center">
          <nav className="md:ml-auto md:mr-auto flex flex-wrap items-center text-base justify-center">
            <Link
              href="/"
              className="border-0 py-1 px-3 mr-5 hover:text-gray-100 dark:target:text-gray-100"
            >
              Home
            </Link>
            <Link
              href="/dashboard"
              className="border-0 py-1 px-3 mr-5 hover:text-gray-100 dark:target:text-gray-100"
            >
              Dashboard
            </Link>
          </nav>
        </div>
      </header>
      <div className="container mx-auto max-w-screen-lg xl:max-w-screen-xl">
        {children}
      </div>
    </div>
  );
}
