"use client";

import { useTheme } from "@/hooks/useTheme";
import { MoonIcon, SunIcon } from "./ThemeToggle";

/** A clearly labeled light/dark switch for the dashboard content area
 * itself (as opposed to the compact icon-only ThemeToggle tucked into
 * the sidebar) — same underlying state via useTheme, just a more
 * visible, textual control at the top of every page. */
export default function ThemeModeButton() {
  const { theme, toggle } = useTheme();
  const isDark = theme === "dark";

  return (
    <button
      type="button"
      onClick={toggle}
      aria-label={isDark ? "Switch to light mode" : "Switch to dark mode"}
      className="inline-flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-sm font-medium text-slate-700 shadow-sm hover:bg-slate-50 dark:border-blue-400/20 dark:bg-slate-900 dark:text-blue-100 dark:shadow-[0_0_10px_rgba(59,130,246,0.15)] dark:hover:bg-slate-800"
    >
      {isDark ? <MoonIcon /> : <SunIcon />}
      {isDark ? "Dark mode" : "Light mode"}
    </button>
  );
}
