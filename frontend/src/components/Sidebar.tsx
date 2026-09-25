"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";
import ThemeToggle from "./ThemeToggle";

const LINKS = [
  { href: "/", label: "Playground" },
  { href: "/observability", label: "Observability" },
  { href: "/config", label: "Configuration" },
  { href: "/keys", label: "API Keys" },
  { href: "/algorithms", label: "Algorithms" },
  { href: "/system", label: "System" },
] as const;

// Solid navy in light mode; a deeper navy gradient with a glowing blue
// active state in dark mode, per the user's white + navy blue palette.
const SIDEBAR_SURFACE =
  "bg-blue-900 dark:bg-gradient-to-b dark:from-blue-950 dark:via-slate-950 dark:to-blue-900";

function SidebarLink({
  href,
  label,
  active,
  onClick,
}: {
  href: string;
  label: string;
  active: boolean;
  onClick?: () => void;
}) {
  return (
    <Link
      href={href}
      onClick={onClick}
      aria-current={active ? "page" : undefined}
      className={`rounded-md px-3 py-2 text-sm font-medium transition-colors ${
        active
          ? "bg-blue-500/30 text-white ring-1 ring-blue-300/60 shadow-[0_0_10px_rgba(59,130,246,0.45)] dark:bg-blue-500/25 dark:text-blue-100 dark:shadow-[0_0_18px_rgba(59,130,246,0.55)] dark:ring-1 dark:ring-blue-400/60"
          : "text-blue-100 hover:bg-white/10 hover:text-white dark:text-blue-200/70 dark:hover:bg-blue-500/10 dark:hover:text-blue-100"
      }`}
    >
      {label}
    </Link>
  );
}

export default function Sidebar() {
  const pathname = usePathname();
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <>
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:rounded-md focus:bg-blue-800 focus:px-4 focus:py-2 focus:text-sm focus:font-medium focus:text-white"
      >
        Skip to content
      </a>

      {/* Mobile top bar */}
      <div className={`flex items-center justify-between px-4 py-3 text-white md:hidden ${SIDEBAR_SURFACE}`}>
        <Link href="/" className="text-sm font-semibold">
          Rate Limiter Admin
        </Link>
        <div className="flex items-center gap-1">
          <ThemeToggle />
          <button
            type="button"
            onClick={() => setMobileOpen((open) => !open)}
            aria-expanded={mobileOpen}
            aria-controls="mobile-sidebar"
            aria-label={mobileOpen ? "Close menu" : "Open menu"}
            className="inline-flex h-9 w-9 items-center justify-center rounded-lg text-white/90 hover:bg-white/10 dark:hover:bg-blue-500/10"
          >
            <svg aria-hidden viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="h-5 w-5">
              {mobileOpen ? (
                <path d="M6 6l12 12M18 6L6 18" strokeLinecap="round" />
              ) : (
                <path d="M4 6h16M4 12h16M4 18h16" strokeLinecap="round" />
              )}
            </svg>
          </button>
        </div>
      </div>

      {/* Mobile drawer */}
      {mobileOpen && (
        <div className="fixed inset-0 z-40 md:hidden">
          <button
            type="button"
            aria-label="Close menu"
            onClick={() => setMobileOpen(false)}
            className="absolute inset-0 bg-black/40"
          />
          <nav
            id="mobile-sidebar"
            aria-label="Main navigation"
            className={`relative z-50 flex h-full w-64 flex-col gap-1 p-4 text-white ${SIDEBAR_SURFACE}`}
          >
            {LINKS.map((link) => (
              <SidebarLink
                key={link.href}
                {...link}
                active={pathname === link.href}
                onClick={() => setMobileOpen(false)}
              />
            ))}
          </nav>
        </div>
      )}

      {/* Desktop sidebar */}
      <nav
        aria-label="Main navigation"
        className={`hidden md:sticky md:top-0 md:flex md:h-screen md:w-60 md:flex-none md:flex-col md:justify-between ${SIDEBAR_SURFACE}`}
      >
        <div>
          <div className="px-4 py-5">
            <Link href="/" className="text-base font-semibold text-white">
              Rate Limiter Admin
            </Link>
          </div>
          <div className="flex flex-col gap-1 px-3">
            {LINKS.map((link) => (
              <SidebarLink key={link.href} {...link} active={pathname === link.href} />
            ))}
          </div>
        </div>

        <div className="border-t border-white/10 px-3 py-4 dark:border-blue-400/10">
          <ThemeToggle />
        </div>
      </nav>
    </>
  );
}
