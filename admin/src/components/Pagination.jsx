export default function Pagination({ page, total, perPage = 20, onChange }) {
  const totalPages = Math.ceil(total / perPage)
  if (totalPages <= 1) return null

  const pages = []
  for (let i = Math.max(1, page - 2); i <= Math.min(totalPages, page + 2); i++) pages.push(i)

  const btn = (active) => ({
    padding: '5px 10px', borderRadius: 6, fontSize: 13, cursor: 'pointer',
    border: `1px solid ${active ? '#2563eb' : '#e2e8f0'}`,
    background: active ? '#2563eb' : '#fff',
    color: active ? '#fff' : '#64748b',
    fontFamily: 'inherit', transition: 'all .12s',
  })

  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', paddingTop: 16 }}>
      <span style={{ fontSize: 12, color: '#94a3b8' }}>
        {(page - 1) * perPage + 1}–{Math.min(page * perPage, total)} из {total}
      </span>
      <div style={{ display: 'flex', gap: 4 }}>
        <button style={{ ...btn(false), opacity: page === 1 ? .4 : 1 }} onClick={() => onChange(page - 1)} disabled={page === 1}>←</button>
        {pages.map(p => <button key={p} style={btn(p === page)} onClick={() => onChange(p)}>{p}</button>)}
        <button style={{ ...btn(false), opacity: page === totalPages ? .4 : 1 }} onClick={() => onChange(page + 1)} disabled={page === totalPages}>→</button>
      </div>
    </div>
  )
}
