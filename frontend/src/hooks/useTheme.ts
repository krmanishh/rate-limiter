"use client";

import { useSyncExternalStore } from "react";

export type Theme = "light" | "dark";

const STORAGE_KEY = "theme";

function subscribe(onChange: () => void) {
  const observer = new MutationObserver(onChange);
  observer.observe(document.documentElement, { attributes: true, attributeFilter: ["class"] });
  return () => observer.disconnect();
}

function getSnapshot(): Theme {
  return document.documentElement.classList.contains("dark") ? "dark" : "light";
}

// The layout.tsx inline script sets the real class before hydration, so
// the server can never know which theme that will be — "light" here is
// just a deterministic placeholder for the one SSR pass; useSyncExternalStore
// re-renders with the real client snapshot immediately after hydration,
// which is its intended way of bridging externally-owned state like this
// without a setState-in-effect (and the flash that pattern causes).
function getServerSnapshot(): Theme {
  return "light";
}

/** Shared by every theme-switching control (the sidebar's icon button
 * and the dashboard's labeled button) so there's exactly one place that
 * reads/writes the `.dark` class and localStorage. */
export function useTheme() {
  const theme = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);

  function toggle() {
    const next: Theme = theme === "dark" ? "light" : "dark";
    document.documentElement.classList.toggle("dark", next === "dark");
    window.localStorage.setItem(STORAGE_KEY, next);
  }

  return { theme, toggle };
}
