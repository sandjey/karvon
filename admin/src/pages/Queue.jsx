import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import Pagination from '../components/Pagination'
import Modal from '../components/Modal'
import { statusBadge } from '../components/Badge'
import { IcoCheck, IcoX, IcoDoc, IcoAlert } from '../icons'

const TABS = [
  { value: 'pending',        label: 'На проверке' },
  { value: 'docs_requested', label: 'Доп. документы' },
  { value: 'approved',       label: 'Одобренные' },
  { value: 'rejected',       label: 'Отклонённые' },
]

export default function Queue() {
  const [items, setItems]   = useState([])
  const [total, setTotal]   = useState(0)
  const [page, setPage]     = useState(1)
  const [status, setStatus] = useState('pending')
  const [loading, setLoading] = useState(true)
  const [detail, setDetail] = useState(null)
  const [rejectModal, setRejectModal] = useState(null)
  const [docsModal, setDocsModal]     = useState(null)
  const [reason, setReason]   = useState('')
  const [message, setMessage] = useState('')
  const [saving, setSaving]   = useState(false)

  const load = useCallback(() => {
    setLoading(true)
    api.queue(status, page)
      .then(r => { setItems(r?.data?.items || []); setTotal(r?.data?.total || 0) })
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [status, page])

  useEffect(() => { load() }, [load])

  const handleApprove = async (item) => {
    if (!confirm(`Одобрить компанию "${item.company_name}"?`)) return
    setSaving(true)
    try { await api.approve(item.id); load(); if (detail?.id === item.id) setDetail(null) }
    finally { setSaving(false) }
  }

  const handleReject = async () => {
    if (!reason.trim()) return
    setSaving(true)
    try { await api.reject(rejectModal.id, reason); setRejectModal(null); setReason(''); load() }
    finally { setSaving(false) }
  }

  const handleRequestDocs = async () => {
    if (!message.trim()) return
    setSaving(true)
    try { await api.requestDocs(docsModal.id, message); setDocsModal(null); setMessage(''); load() }
    finally { setSaving(false) }
  }

  const openDetail = async (item) => {
    const r = await api.queueItem(item.id)
    setDetail(r?.data || item)
  }

  const isPending = status === 'pending' || status === 'docs_requested'

  const tabBtn = (t) => ({
    padding: '7px 16px', borderRadius: 7, fontSize: 13, fontWeight: 600, cursor: 'pointer',
    border: 'none', fontFamily: 'inherit', transition: 'all .13s',
    background: status === t.value ? '#1d56d4' : 'transparent',
    color: status === t.value ? '#fff' : '#334d6e',
  })

  return (
    <div style={{ padding: 28, display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div>
        <div style={{ fontSize: 20, fontWeight: 800, color: '#dce8f5' }}>Очередь верификации</div>
        <div style={{ fontSize: 13, color: '#2d4a6e', marginTop: 3 }}>Компании, ожидающие проверки документов</div>
      </div>

      {/* Tabs */}
      <div style={{
        display: 'flex', gap: 2, background: '#0c1526',
        border: '1px solid #19273d', borderRadius: 9, padding: 3, width: 'fit-content',
      }}>
        {TABS.map(t => (
          <button key={t.value} style={tabBtn(t)} onClick={() => { setStatus(t.value); setPage(1) }}>
            {t.label}
          </button>
        ))}
      </div>

      {/* Cards grid */}
      <div style={{ display: 'grid', gridTemplateColumns: detail ? '1fr' : 'repeat(2, 1fr)', gap: 14 }}>
        {loading ? (
          <div style={{ gridColumn: '1/-1', textAlign: 'center', padding: '48px 0', color: '#1e3350' }}>Загрузка...</div>
        ) : items.length === 0 ? (
          <div className="card" style={{ gridColumn: '1/-1', textAlign: 'center', padding: '60px 0' }}>
            <IcoCheck size={36} style={{ color: '#1e3350', margin: '0 auto 12px', display: 'block' }} />
            <div style={{ color: '#2d4a6e', fontWeight: 600 }}>Очередь пуста</div>
            <div style={{ color: '#1a2a3d', fontSize: 13, marginTop: 6 }}>
              Нет компаний со статусом «{TABS.find(t => t.value === status)?.label}»
            </div>
          </div>
        ) : items.map(item => (
          <div
            key={item.id}
            className="card"
            onClick={() => openDetail(item)}
            style={{
              cursor: 'pointer', transition: 'border-color .13s',
              borderColor: detail?.id === item.id ? 'rgba(59,127,245,0.5)' : '#19273d',
              borderLeftWidth: item.is_urgent ? 3 : 1,
              borderLeftColor: item.is_urgent ? '#e84040' : (detail?.id === item.id ? 'rgba(59,127,245,0.5)' : '#19273d'),
            }}
          >
            <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 10 }}>
              <div style={{ flex: 1 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
                  <span style={{ fontWeight: 700, color: '#dce8f5' }}>{item.company_name}</span>
                  {item.is_urgent && (
                    <span style={{ fontSize: 11, fontWeight: 600, padding: '2px 7px', borderRadius: 5, background: 'rgba(232,64,64,0.1)', color: '#e84040', border: '1px solid rgba(232,64,64,0.2)' }}>
                      СРОЧНО
                    </span>
                  )}
                </div>
                <div style={{ fontSize: 12.5, color: '#4d6d90', marginTop: 4 }}>
                  {item.user_name} · {item.user_phone}
                </div>
                <div style={{ display: 'flex', gap: 10, marginTop: 8, flexWrap: 'wrap', fontSize: 12, color: '#2d4a6e' }}>
                  <span>{item.org_type?.toUpperCase()}</span>
                  <span>·</span>
                  <span>{item.country}</span>
                  {item.inn && <><span>·</span><span>ИНН: {item.inn}</span></>}
                </div>
              </div>
              <div style={{ flexShrink: 0 }}>{statusBadge(item.status)}</div>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginTop: 14, paddingTop: 12, borderTop: '1px solid #0f1c2f' }}>
              <div style={{ fontSize: 12, color: '#1e3350' }}>
                {new Date(item.created_at).toLocaleDateString('ru', { day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit' })}
              </div>
              {isPending && (
                <div style={{ display: 'flex', gap: 6 }} onClick={e => e.stopPropagation()}>
                  <button
                    onClick={() => handleApprove(item)} disabled={saving}
                    style={{
                      display: 'inline-flex', alignItems: 'center', gap: 4,
                      padding: '4px 10px', fontSize: 12, fontWeight: 600, cursor: 'pointer',
                      background: 'rgba(18,198,120,0.12)', color: '#12c678',
                      border: '1px solid rgba(18,198,120,0.22)', borderRadius: 6,
                    }}
                  ><IcoCheck size={11} /> Одобрить</button>
                  <button
                    onClick={() => setDocsModal(item)}
                    style={{ padding: '4px 10px', fontSize: 12, fontWeight: 600, cursor: 'pointer', background: 'rgba(59,127,245,0.1)', color: '#60a5fa', border: '1px solid rgba(59,127,245,0.2)', borderRadius: 6 }}
                  ><IcoDoc size={11} /></button>
                  <button
                    onClick={() => setRejectModal(item)}
                    style={{ padding: '4px 10px', fontSize: 12, fontWeight: 600, cursor: 'pointer', background: 'rgba(232,64,64,0.1)', color: '#e84040', border: '1px solid rgba(232,64,64,0.2)', borderRadius: 6 }}
                  ><IcoX size={11} /></button>
                </div>
              )}
            </div>
          </div>
        ))}
      </div>

      <Pagination page={page} total={total} onChange={setPage} />

      {/* Detail Side Panel */}
      {detail && (
        <div style={{
          position: 'fixed', top: 0, right: 0, bottom: 0, width: 380, zIndex: 40,
          background: '#0c1526', borderLeft: '1px solid #19273d',
          display: 'flex', flexDirection: 'column',
          boxShadow: '-8px 0 40px rgba(0,0,0,0.5)',
        }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '18px 20px', borderBottom: '1px solid #19273d' }}>
            <span style={{ fontWeight: 700, color: '#dce8f5', fontSize: 15 }}>{detail.company_name}</span>
            <button
              onClick={() => setDetail(null)}
              style={{ background: 'rgba(255,255,255,0.04)', border: '1px solid #19273d', borderRadius: 7, width: 28, height: 28, cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#4d6d90' }}
            ><IcoX size={13} /></button>
          </div>

          <div style={{ flex: 1, overflowY: 'auto', padding: 20, display: 'flex', flexDirection: 'column', gap: 20 }}>
            {[
              { title: 'Компания', rows: [
                ['Тип', detail.org_type?.toUpperCase()],
                ['Страна', detail.country],
                ['Город', `${detail.city}, ${detail.region}`],
                ['Адрес', detail.street],
                ['ИНН', detail.inn],
              ]},
              { title: 'Контакты', rows: [
                ['Email', detail.email || '—'],
                ['Телефон', detail.phone || '—'],
              ]},
              { title: 'Владелец', rows: [
                ['Имя', detail.user_name],
                ['Телефон', detail.user_phone],
              ]},
            ].map(section => (
              <div key={section.title}>
                <div style={{ fontSize: 10.5, fontWeight: 700, color: '#1e3350', textTransform: 'uppercase', letterSpacing: '.08em', marginBottom: 10 }}>
                  {section.title}
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                  {section.rows.filter(([, v]) => v).map(([k, v]) => (
                    <div key={k} style={{ display: 'flex', gap: 8 }}>
                      <span style={{ fontSize: 12.5, color: '#2d4a6e', minWidth: 70 }}>{k}:</span>
                      <span style={{ fontSize: 12.5, color: '#7393b5' }}>{v}</span>
                    </div>
                  ))}
                </div>
              </div>
            ))}

            {detail.reg_doc_url && (
              <div>
                <div style={{ fontSize: 10.5, fontWeight: 700, color: '#1e3350', textTransform: 'uppercase', letterSpacing: '.08em', marginBottom: 10 }}>Документы</div>
                <a href={detail.reg_doc_url} target="_blank" rel="noreferrer" style={{
                  display: 'flex', alignItems: 'center', gap: 8, padding: '10px 12px',
                  background: 'rgba(59,127,245,0.08)', color: '#60a5fa',
                  border: '1px solid rgba(59,127,245,0.18)', borderRadius: 8,
                  textDecoration: 'none', fontSize: 13, fontWeight: 500,
                }}>
                  <IcoDoc size={15} /> Регистрационный документ
                </a>
              </div>
            )}

            {detail.rejection_reason && (
              <div style={{ background: 'rgba(232,64,64,0.08)', border: '1px solid rgba(232,64,64,0.18)', borderRadius: 8, padding: 12 }}>
                <div style={{ fontSize: 12, fontWeight: 600, color: '#e84040', marginBottom: 4 }}>Причина отклонения:</div>
                <div style={{ fontSize: 13, color: '#f87171' }}>{detail.rejection_reason}</div>
              </div>
            )}
          </div>

          {isPending && (
            <div style={{ padding: '14px 16px', borderTop: '1px solid #19273d', display: 'flex', gap: 8 }}>
              <button
                onClick={() => handleApprove(detail)}
                style={{ flex: 1, padding: '9px', background: '#0fa05e', color: '#fff', border: 'none', borderRadius: 8, cursor: 'pointer', fontSize: 13, fontWeight: 700, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6 }}
              ><IcoCheck size={14} /> Одобрить</button>
              <button
                onClick={() => setDocsModal(detail)}
                style={{ padding: '9px 14px', background: 'rgba(59,127,245,0.1)', color: '#60a5fa', border: '1px solid rgba(59,127,245,0.2)', borderRadius: 8, cursor: 'pointer', display: 'flex', alignItems: 'center' }}
              ><IcoDoc size={15} /></button>
              <button
                onClick={() => setRejectModal(detail)}
                style={{ padding: '9px 14px', background: 'rgba(232,64,64,0.1)', color: '#e84040', border: '1px solid rgba(232,64,64,0.2)', borderRadius: 8, cursor: 'pointer', display: 'flex', alignItems: 'center' }}
              ><IcoX size={15} /></button>
            </div>
          )}
        </div>
      )}

      {/* Reject Modal */}
      {rejectModal && (
        <Modal title={`Отклонить: ${rejectModal.company_name}`} onClose={() => { setRejectModal(null); setReason('') }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <div>
              <label style={{ display: 'block', fontSize: 12, fontWeight: 600, color: '#3d5a7d', marginBottom: 7, textTransform: 'uppercase', letterSpacing: '.05em' }}>
                Причина отклонения <span style={{ color: '#e84040' }}>*</span>
              </label>
              <textarea
                className="input" style={{ minHeight: 100 }}
                placeholder="Укажите причину отказа..."
                value={reason}
                onChange={e => setReason(e.target.value)}
                autoFocus
              />
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <button className="btn-secondary" onClick={() => setRejectModal(null)}>Отмена</button>
              <button className="btn-danger" onClick={handleReject} disabled={saving || !reason.trim()}>Отклонить</button>
            </div>
          </div>
        </Modal>
      )}

      {/* Request Docs Modal */}
      {docsModal && (
        <Modal title={`Запросить документы: ${docsModal.company_name}`} onClose={() => { setDocsModal(null); setMessage('') }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <div>
              <label style={{ display: 'block', fontSize: 12, fontWeight: 600, color: '#3d5a7d', marginBottom: 7, textTransform: 'uppercase', letterSpacing: '.05em' }}>
                Сообщение <span style={{ color: '#e84040' }}>*</span>
              </label>
              <textarea
                className="input" style={{ minHeight: 100 }}
                placeholder="Какие документы нужно предоставить..."
                value={message}
                onChange={e => setMessage(e.target.value)}
                autoFocus
              />
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <button className="btn-secondary" onClick={() => setDocsModal(null)}>Отмена</button>
              <button className="btn-primary" onClick={handleRequestDocs} disabled={saving || !message.trim()}>Отправить запрос</button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
