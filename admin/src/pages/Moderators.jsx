import { useState, useEffect } from 'react'
import { api } from '../api'
import Modal from '../components/Modal'

export default function Moderators() {
  const [moderators, setModerators] = useState([])
  const [loading, setLoading]       = useState(true)
  const [showModal, setShowModal]   = useState(false)
  const [form, setForm]             = useState({ phone: '', name: '' })
  const [saving, setSaving]         = useState(false)
  const [error, setError]           = useState('')

  const load = () => {
    setLoading(true)
    api.users('', 'moderator', 1)
      .then(r => setModerators(r?.data?.items || []))
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  const handleCreate = async () => {
    if (!form.phone) return
    setError('')
    setSaving(true)
    try {
      await api.createModerator(form.phone, form.name || undefined)
      setShowModal(false)
      setForm({ phone: '', name: '' })
      load()
    } catch (err) {
      setError(err.message || 'Ошибка создания')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (m) => {
    if (!confirm(`Удалить модератора ${m.phone}? Он потеряет доступ к панели.`)) return
    await api.deleteModerator(m.id)
    load()
  }

  return (
    <div className="p-6 space-y-5">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-800">Модераторы</h1>
          <p className="text-sm text-slate-500">Пользователи с доступом к очереди верификации</p>
        </div>
        <button className="btn-primary" onClick={() => setShowModal(true)}>+ Добавить модератора</button>
      </div>

      <div className="card overflow-x-auto">
        {loading ? (
          <div className="text-center py-12 text-slate-400">Загрузка...</div>
        ) : moderators.length === 0 ? (
          <div className="text-center py-16">
            <div className="text-4xl mb-3">🛡️</div>
            <div className="text-slate-500 font-medium">Нет модераторов</div>
            <p className="text-slate-400 text-sm mt-1">Нажмите «Добавить модератора» чтобы создать</p>
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-slate-500 border-b border-slate-100">
                <th className="pb-3 font-medium">Модератор</th>
                <th className="pb-3 font-medium">Телефон</th>
                <th className="pb-3 font-medium">Дата создания</th>
                <th className="pb-3 font-medium text-right">Действия</th>
              </tr>
            </thead>
            <tbody>
              {moderators.map(m => (
                <tr key={m.id} className="border-b border-slate-50 hover:bg-slate-50 group">
                  <td className="py-3">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center text-blue-600 text-sm font-semibold">
                        {(m.name || m.phone).slice(0,2).toUpperCase()}
                      </div>
                      <div className="font-medium text-slate-800">{m.name || '—'}</div>
                    </div>
                  </td>
                  <td className="py-3 text-slate-500">{m.phone}</td>
                  <td className="py-3 text-slate-400 text-xs">{new Date(m.created_at).toLocaleDateString('ru')}</td>
                  <td className="py-3 text-right">
                    <button
                      onClick={() => handleDelete(m)}
                      className="opacity-0 group-hover:opacity-100 transition px-3 py-1 bg-red-50 text-red-600 hover:bg-red-100 rounded-lg text-xs font-medium"
                    >Удалить</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {showModal && (
        <Modal title="Новый модератор" onClose={() => { setShowModal(false); setError('') }}>
          <div className="space-y-4">
            {error && <div className="bg-red-50 text-red-600 rounded-lg px-4 py-2.5 text-sm">{error}</div>}
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Телефон <span className="text-red-500">*</span></label>
              <input
                className="input" placeholder="998901234567"
                value={form.phone}
                onChange={e => setForm(f => ({...f, phone: e.target.value}))}
                autoFocus
              />
              <p className="text-xs text-slate-400 mt-1">Формат: 998XXXXXXXXX. Модератор войдёт через OTP-вход.</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Имя (необязательно)</label>
              <input
                className="input" placeholder="Алишер Каримов"
                value={form.name}
                onChange={e => setForm(f => ({...f, name: e.target.value}))}
              />
            </div>
            <div className="flex gap-3 justify-end pt-2">
              <button className="btn-secondary" onClick={() => setShowModal(false)}>Отмена</button>
              <button className="btn-primary" onClick={handleCreate} disabled={saving}>
                {saving ? 'Создание...' : 'Создать'}
              </button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
