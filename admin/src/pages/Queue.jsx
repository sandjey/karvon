import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import Pagination from '../components/Pagination'
import Modal from '../components/Modal'
import { statusBadge } from '../components/Badge'

const TABS = [
  { value: 'pending',        label: '⏳ На проверке' },
  { value: 'docs_requested', label: '📄 Доп. документы' },
  { value: 'approved',       label: '✅ Одобренные' },
  { value: 'rejected',       label: '❌ Отклонённые' },
]

export default function Queue() {
  const [items, setItems]       = useState([])
  const [total, setTotal]       = useState(0)
  const [page, setPage]         = useState(1)
  const [status, setStatus]     = useState('pending')
  const [loading, setLoading]   = useState(true)
  const [detail, setDetail]     = useState(null)
  const [rejectModal, setRejectModal] = useState(null)
  const [docsModal, setDocsModal]     = useState(null)
  const [reason, setReason]     = useState('')
  const [message, setMessage]   = useState('')
  const [saving, setSaving]     = useState(false)

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
    try {
      await api.approve(item.id)
      load()
      if (detail?.id === item.id) setDetail(null)
    } finally { setSaving(false) }
  }

  const handleReject = async () => {
    if (!reason.trim()) return
    setSaving(true)
    try {
      await api.reject(rejectModal.id, reason)
      setRejectModal(null)
      setReason('')
      load()
    } finally { setSaving(false) }
  }

  const handleRequestDocs = async () => {
    if (!message.trim()) return
    setSaving(true)
    try {
      await api.requestDocs(docsModal.id, message)
      setDocsModal(null)
      setMessage('')
      load()
    } finally { setSaving(false) }
  }

  const openDetail = async (item) => {
    const r = await api.queueItem(item.id)
    setDetail(r?.data || item)
  }

  const isPending = status === 'pending' || status === 'docs_requested'

  return (
    <div className="p-6 space-y-5">
      <div>
        <h1 className="text-xl font-bold text-slate-800">Очередь верификации</h1>
        <p className="text-sm text-slate-500">Компании, ожидающие проверки документов</p>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 bg-white rounded-xl border border-slate-200 p-1 w-fit flex-wrap">
        {TABS.map(t => (
          <button
            key={t.value}
            onClick={() => { setStatus(t.value); setPage(1) }}
            className={`px-4 py-1.5 rounded-lg text-sm font-medium transition ${status === t.value ? 'bg-blue-600 text-white' : 'text-slate-600 hover:bg-slate-50'}`}
          >{t.label}</button>
        ))}
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
        {loading ? (
          <div className="col-span-2 text-center py-12 text-slate-400">Загрузка...</div>
        ) : items.length === 0 ? (
          <div className="col-span-2 card text-center py-16">
            <div className="text-4xl mb-3">🎉</div>
            <div className="text-slate-500 font-medium">Очередь пуста</div>
            <p className="text-slate-400 text-sm mt-1">Нет компаний со статусом «{TABS.find(t=>t.value===status)?.label}»</p>
          </div>
        ) : items.map(item => (
          <div
            key={item.id}
            className={`card cursor-pointer hover:shadow-md transition border-2 ${detail?.id === item.id ? 'border-blue-500' : 'border-transparent'} ${item.is_urgent ? 'border-l-4 border-l-red-400' : ''}`}
            onClick={() => openDetail(item)}
          >
            <div className="flex items-start justify-between gap-3">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="font-semibold text-slate-800">{item.company_name}</span>
                  {item.is_urgent && <span className="text-xs bg-red-100 text-red-600 px-2 py-0.5 rounded-full font-medium">⚡ Срочно</span>}
                </div>
                <div className="text-sm text-slate-500 mt-1">{item.user_name} · {item.user_phone}</div>
                <div className="flex gap-2 mt-2 flex-wrap text-xs text-slate-500">
                  <span>🏢 {item.org_type?.toUpperCase()}</span>
                  <span>📍 {item.country}</span>
                  {item.inn && <span>ИНН: {item.inn}</span>}
                </div>
              </div>
              <div className="flex-shrink-0">{statusBadge(item.status)}</div>
            </div>

            <div className="flex items-center justify-between mt-4 pt-4 border-t border-slate-100">
              <div className="text-xs text-slate-400">
                {new Date(item.created_at).toLocaleDateString('ru', { day:'numeric', month:'long', hour:'2-digit', minute:'2-digit' })}
                {item.deadline && ` · До: ${new Date(item.deadline).toLocaleDateString('ru')}`}
              </div>
              {isPending && (
                <div className="flex gap-2" onClick={e => e.stopPropagation()}>
                  <button
                    onClick={() => handleApprove(item)}
                    disabled={saving}
                    className="px-3 py-1 bg-emerald-500 hover:bg-emerald-600 text-white rounded-lg text-xs font-medium transition"
                  >✓ Одобрить</button>
                  <button
                    onClick={() => setDocsModal(item)}
                    className="px-3 py-1 bg-blue-50 hover:bg-blue-100 text-blue-600 rounded-lg text-xs font-medium"
                  >📄 Доки</button>
                  <button
                    onClick={() => setRejectModal(item)}
                    className="px-3 py-1 bg-red-50 hover:bg-red-100 text-red-600 rounded-lg text-xs font-medium"
                  >✕ Отклонить</button>
                </div>
              )}
            </div>
          </div>
        ))}
      </div>

      <Pagination page={page} total={total} onChange={setPage} />

      {/* Detail Panel */}
      {detail && (
        <div className="fixed inset-y-0 right-0 w-96 bg-white shadow-2xl flex flex-col z-40 border-l border-slate-200">
          <div className="flex items-center justify-between p-5 border-b border-slate-100">
            <h3 className="font-semibold text-slate-800">{detail.company_name}</h3>
            <button onClick={() => setDetail(null)} className="text-slate-400 hover:text-slate-600 text-xl">×</button>
          </div>
          <div className="flex-1 overflow-y-auto p-5 space-y-4 text-sm">
            <section>
              <div className="font-medium text-slate-500 text-xs uppercase tracking-wider mb-2">Компания</div>
              <div className="space-y-1.5 text-slate-700">
                <div><span className="text-slate-400">Тип:</span> {detail.org_type?.toUpperCase()}</div>
                <div><span className="text-slate-400">Страна:</span> {detail.country}</div>
                <div><span className="text-slate-400">Город:</span> {detail.city}, {detail.region}</div>
                <div><span className="text-slate-400">Адрес:</span> {detail.street}</div>
                <div><span className="text-slate-400">ИНН:</span> {detail.inn} {detail.inn_verified && '✅'}</div>
              </div>
            </section>
            <section>
              <div className="font-medium text-slate-500 text-xs uppercase tracking-wider mb-2">Контакты</div>
              <div className="space-y-1.5 text-slate-700">
                <div><span className="text-slate-400">Email:</span> {detail.email || '—'}</div>
                <div><span className="text-slate-400">Телефон:</span> {detail.phone || '—'}</div>
              </div>
            </section>
            <section>
              <div className="font-medium text-slate-500 text-xs uppercase tracking-wider mb-2">Владелец</div>
              <div className="space-y-1.5 text-slate-700">
                <div>{detail.user_name}</div>
                <div className="text-slate-400">{detail.user_phone}</div>
              </div>
            </section>
            {detail.reg_doc_url && (
              <section>
                <div className="font-medium text-slate-500 text-xs uppercase tracking-wider mb-2">Документы</div>
                <a href={detail.reg_doc_url} target="_blank" rel="noreferrer"
                  className="flex items-center gap-2 p-3 bg-blue-50 rounded-lg text-blue-600 hover:bg-blue-100">
                  📄 <span>Регистрационный документ</span>
                </a>
              </section>
            )}
            {detail.rejection_reason && (
              <div className="bg-red-50 rounded-lg p-3 text-red-700 text-sm">
                <div className="font-medium">Причина отклонения:</div>
                <div>{detail.rejection_reason}</div>
              </div>
            )}
            {detail.docs_request_note && (
              <div className="bg-blue-50 rounded-lg p-3 text-blue-700 text-sm">
                <div className="font-medium">Запрос документов:</div>
                <div>{detail.docs_request_note}</div>
              </div>
            )}
          </div>
          {isPending && (
            <div className="p-5 border-t border-slate-100 flex gap-2">
              <button onClick={() => handleApprove(detail)} className="flex-1 bg-emerald-500 hover:bg-emerald-600 text-white rounded-lg py-2 text-sm font-medium">✓ Одобрить</button>
              <button onClick={() => setDocsModal(detail)} className="px-4 bg-blue-50 text-blue-600 rounded-lg text-sm font-medium">📄</button>
              <button onClick={() => setRejectModal(detail)} className="px-4 bg-red-50 text-red-600 rounded-lg text-sm font-medium">✕</button>
            </div>
          )}
        </div>
      )}

      {/* Reject Modal */}
      {rejectModal && (
        <Modal title={`Отклонить: ${rejectModal.company_name}`} onClose={() => { setRejectModal(null); setReason('') }}>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Причина отклонения <span className="text-red-500">*</span></label>
              <textarea
                className="input min-h-[100px] resize-none"
                placeholder="Укажите причину отказа..."
                value={reason}
                onChange={e => setReason(e.target.value)}
                autoFocus
              />
            </div>
            <div className="flex gap-3 justify-end">
              <button className="btn-secondary" onClick={() => setRejectModal(null)}>Отмена</button>
              <button className="btn-danger" onClick={handleReject} disabled={saving || !reason.trim()}>Отклонить</button>
            </div>
          </div>
        </Modal>
      )}

      {/* Request Docs Modal */}
      {docsModal && (
        <Modal title={`Запросить документы: ${docsModal.company_name}`} onClose={() => { setDocsModal(null); setMessage('') }}>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Сообщение <span className="text-red-500">*</span></label>
              <textarea
                className="input min-h-[100px] resize-none"
                placeholder="Какие документы нужно предоставить..."
                value={message}
                onChange={e => setMessage(e.target.value)}
                autoFocus
              />
            </div>
            <div className="flex gap-3 justify-end">
              <button className="btn-secondary" onClick={() => setDocsModal(null)}>Отмена</button>
              <button className="btn-primary" onClick={handleRequestDocs} disabled={saving || !message.trim()}>Отправить запрос</button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
