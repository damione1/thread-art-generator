"use client";

import Header from "@/components/Header";
import Sidebar from "@/components/Sidebar";
import Loader from "@/components/common/Loader";
import { SessionProvider, useSession } from "next-auth/react";
import { createContext, useEffect, useState } from "react";

export const ThemeContext = createContext({});

export default function ThemeProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const { data: session, status } = useSession({ required: true })
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    // if (status === "authenticated") {
    //   setLoading(false);
    // }else if (status === "loading") {
    //   setLoading(true);
    // }
    setTimeout(() => setLoading(false), 1000);
  }, [status]);
  return (
<>
      {loading ? (
        <Loader />
      ) : (
        <div className="flex h-screen overflow-hidden">
          {/* <!-- ===== Sidebar Start ===== --> */}
          <Sidebar sidebarOpen={sidebarOpen} setSidebarOpen={setSidebarOpen} />
          {/* <!-- ===== Sidebar End ===== --> */}

          {/* <!-- ===== Content Area Start ===== --> */}
          <div className="relative flex flex-1 flex-col overflow-y-auto overflow-x-hidden">
            {/* <!-- ===== Header Start ===== --> */}
            <Header sidebarOpen={sidebarOpen} setSidebarOpen={setSidebarOpen} />
            {/* <!-- ===== Header End ===== --> */}

            {/* <!-- ===== Main Content Start ===== --> */}
            <main>
              <div className="mx-auto max-w-screen-2xl p-4 md:p-6 2xl:p-10">
                {children}
              </div>
            </main>
            {/* <!-- ===== Main Content End ===== --> */}
          </div>
          {/* <!-- ===== Content Area End ===== --> */}
        </div>
      )}
    </>

  );
}
