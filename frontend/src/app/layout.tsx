import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import Link from "next/link";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Rate Limiter Playground",
  description:
    "An interactive dashboard for exploring the Go rate limiter API.",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col bg-slate-50 text-slate-900 dark:bg-slate-950 dark:text-slate-100">
        <nav className="border-b border-slate-200 bg-white px-6 py-3 dark:border-slate-800 dark:bg-slate-900">
          <div className="mx-auto flex w-full max-w-6xl gap-4 text-sm font-medium">
            <Link href="/" className="text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white">
              Playground
            </Link>
            <Link
              href="/observability"
              className="text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white"
            >
              Observability
            </Link>
          </div>
        </nav>
        {children}
      </body>
    </html>
  );
}
