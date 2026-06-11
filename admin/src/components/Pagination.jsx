export default function Pagination({ page, total, perPage = 20, onChange }) {
  const totalPages = Math.ceil(total / perPage)
  if (totalPages <= 1) return null

  const pages = []
  for (let i = Math.max(1, page - 2); i <= Math.min(totalPages, page + 2); i++) pages.push(i)

  const btnBase = {
    padding: '5px 10px', borderRadius: 6, fontSize: 13, cursor: 'pointer',
    border: '1px solid #19273d', background: 'transparent', color: '#4d6d90',
    transition: 'all .13s', fontFamily: 'inherit',
  }
  const btnActive = { ...btnBase, background: '#1d56d4', color: '#fff', borderColor: '#1d56d4' }

  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', paddingTop: 16 }}>
      <span style={{ fontSize: 12, color: '#2d4a6e' }}>
        {(page - 1) * perPage + 1}–{Math.min(page * perPage, total)} из {total}
      </span>
      <div style={{ display: 'flex', gap: 4 }}>
        <button
          style={{ ...btnBase, opacity: page === 1 ? 0.35 : 1 }}
          onClick={() => onChange(page - 1)} disabled={page === 1}
        >←</button>
        {pages.map(p => (
          <button key={p} style={p === page ? btnActive : btnBase} onClick={() => onChange(p)}>{p}</button>
        ))}
        <button
          style={{ ...btnBase, opacity: page === totalPages ? 0.35 : 1 }}
          onClick={() => onChange(page + 1)} disabled={page === totalPages}
        >→</button>
      </div>
    </div>
  )
}
