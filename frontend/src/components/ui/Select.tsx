import type { SelectHTMLAttributes } from "react";

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label?: string;
}

export default function Select({ label, id, className = "", ...rest }: SelectProps) {
  return (
    <div>
      {label && (
        <label htmlFor={id} className="mb-1 block text-xs font-medium text-slate-500">
          {label}
        </label>
      )}
      <select
        id={id}
        className={`rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm dark:border-slate-700 dark:bg-slate-800 ${className}`}
        {...rest}
      />
    </div>
  );
}
