import { ALGORITHMS } from "@/lib/algorithms";

export default function ComparisonTable() {
  return (
    <div className="overflow-x-auto rounded-lg border border-slate-200 dark:border-slate-800">
      <table className="min-w-full divide-y divide-slate-200 text-sm dark:divide-slate-800">
        <thead className="bg-slate-50 dark:bg-slate-800/60">
          <tr>
            <th scope="col" className="px-3 py-2 text-left font-medium text-slate-500">Algorithm</th>
            <th scope="col" className="px-3 py-2 text-left font-medium text-slate-500">Summary</th>
            <th scope="col" className="px-3 py-2 text-left font-medium text-slate-500">Parameters</th>
            <th scope="col" className="px-3 py-2 text-left font-medium text-slate-500">Pros</th>
            <th scope="col" className="px-3 py-2 text-left font-medium text-slate-500">Cons</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-100 align-top dark:divide-slate-800">
          {ALGORITHMS.map((algorithm) => (
            <tr key={algorithm.id}>
              <td className="px-3 py-3 font-medium text-slate-800 dark:text-slate-100">{algorithm.name}</td>
              <td className="px-3 py-3 text-slate-600 dark:text-slate-300">{algorithm.summary}</td>
              <td className="px-3 py-3 text-slate-500 dark:text-slate-400">
                {algorithm.fields.map((f) => f.label).join(", ")}
              </td>
              <td className="px-3 py-3">
                <ul className="space-y-1 text-emerald-700 dark:text-emerald-400">
                  {algorithm.pros.map((pro) => (
                    <li key={pro}>+ {pro}</li>
                  ))}
                </ul>
              </td>
              <td className="px-3 py-3">
                <ul className="space-y-1 text-rose-700 dark:text-rose-400">
                  {algorithm.cons.map((con) => (
                    <li key={con}>− {con}</li>
                  ))}
                </ul>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
