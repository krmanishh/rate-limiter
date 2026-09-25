import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import Sidebar from "@/components/Sidebar";
import ThemeModeButton from "@/components/ThemeModeButton";
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
  title: "Rate Limiter Admin Console",
  description:
    "An interactive admin console for the Go rate limiter API: playground, observability, configuration, and system health.",
};

// Sets the `.dark` class before React hydrates, using the same storage
// key ThemeToggle reads/writes, so there's no flash of the wrong theme
// on load. Falls back to the OS preference when nothing's been chosen
// yet. Inlined (not an external script) so it runs before first paint.
const themeInitScript = `
(function () {
  try {
    var stored = window.localStorage.getItem("theme");
    var dark = stored ? stored === "dark" : window.matchMedia("(prefers-color-scheme: dark)").matches;
    document.documentElement.classList.toggle("dark", dark);
  } catch (e) {}
})();
`;

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
      suppressHydrationWarning
    >
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeInitScript }} />
      </head>
      <body className="min-h-full flex flex-col bg-white text-slate-900 dark:bg-slate-950 dark:text-slate-100 md:flex-row">
        <Sidebar />
        <div id="main-content" tabIndex={-1} className="flex flex-1 flex-col outline-none">
          <div className="flex justify-end border-b border-slate-200 bg-white px-4 py-2 dark:border-slate-800 dark:bg-slate-950 sm:px-6 lg:px-8">
            <ThemeModeButton />
          </div>
          {children}
        </div>
      </body>
    </html>
  );
}
