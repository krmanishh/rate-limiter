export default function EmptyState({
  title,
  description,
}: {
  title: string;
  description?: string;
}) {
  return (
    <div className="flex flex-col items-center justify-center rounded-lg border border-dashed border-slate-300 px-4 py-8 text-center dark:border-slate-700">
      <p className="text-sm font-medium text-slate-600 dark:text-slate-300">
        {title}
      </p>
      {description && (
        <p className="mt-1 text-xs text-slate-400 dark:text-slate-500">
          {description}
        </p>
      )}
    </div>
  );
}
