export default function Pagination({ page, total, perPage = 20, onChange }) {
  const totalPages = Math.ceil(total / perPage)
  if (totalPages <= 1) return null

  const pages = []
  for (let i = Math.max(1, page - 2); i <= Math.min(totalPages, page + 2); i++) pages.push(i)

  return (
    <div className="flex items-center justify-between pt-4">
      <span className="text-sm text-slate-500">
        {(page - 1) * perPage + 1}–{Math.min(page * perPage, total)} из {total}
      </span>
      <div className="flex gap-1">
        <button
          onClick={() => onChange(page - 1)} disabled={page === 1}
          className="px-3 py-1.5 rounded-lg text-sm border border-slate-200 hover:bg-slate-50 disabled:opacity-40"
        >←</button>
        {pages.map(p => (
          <button
            key={p} onClick={() => onChange(p)}
            className={`px-3 py-1.5 rounded-lg text-sm border ${p === page ? 'bg-blue-600 text-white border-blue-600' : 'border-slate-200 hover:bg-slate-50'}`}
          >{p}</button>
        ))}
        <button
          onClick={() => onChange(page + 1)} disabled={page === totalPages}
          className="px-3 py-1.5 rounded-lg text-sm border border-slate-200 hover:bg-slate-50 disabled:opacity-40"
        >→</button>
      </div>
    </div>
  )
}
