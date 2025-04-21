import "@/styles/globals.css";
import type { AppProps } from "next/app";

export default function App({ Component, pageProps }: AppProps) {
  // All routes in the Pages Router are rendered normally
  return <Component {...pageProps} />;
}
