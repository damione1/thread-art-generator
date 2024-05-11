import "./globals.css";
import "./data-tables-css.css";
import "./satoshi.css";
import ThemeProvider from "./template-provider";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="dark:bg-boxdark-2 dark:text-bodydark">
      <ThemeProvider>{children}</ThemeProvider>
    </div>
  );
}
