import "./globals.css";
import "./data-tables-css.css";
import "./satoshi.css";
import ThemeProvider from "./template-provider";
import { GrpcProvider } from "@/lib/grpc-context";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="dark:bg-boxdark-2 dark:text-bodydark">
      <GrpcProvider><ThemeProvider>{children}</ThemeProvider></GrpcProvider>
    </div>
  );
}
