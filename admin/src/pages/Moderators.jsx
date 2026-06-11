import { useState, useEffect } from 'react'
import { api } from '../api'
import Modal from '../components/Modal'
import { IcoShield, IcoPlus, IcoTrash } from '../icons'

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
      setShowModal(false); setForm({ phone: '', name: '' }); load()
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
    <div style={{ padding: 28, display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <div style={{ fontSize: 20, fontWeight: 800, color: '#dce8f5' }}>Модераторы</div>
          <div style={{ fontSize: 13, color: '#2d4a6e', marginTop: 3 }}>Пользователи с доступом к очереди верификации</div>
        </div>
        <button className="btn-primary" onClick={() => { setShowModal(true); setError('') }}>
          <IcoPlus size={14} /> Добавить модератора
        </button>
      </div>

      <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
        {loading ? (
          <div style={{ textAlign: 'center', padding: '48px 0', color: '#1e3350' }}>Загрузка...</div>
        ) : moderators.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '60px 0' }}>
            <div style={{ marginBottom: 12 }}>
              <IcoShield size={36} style={{ color: '#1e3350', margin: '0 auto', display: 'block' }} />
            </div>
            <div style={{ color: '#2d4a6e', fontSize: 14, fontWeight: 600 }}>Нет модераторов</div>
            <div style={{ color: '#1a2a3d', fontSize: 13, marginTop: 6 }}>Нажмите «Добавить модератора» чтобы создать</div>
          </div>
        ) : (
          <table className="data-table">
            <thead>
              <tr>
                <th style={{ paddingLeft: 20 }}>Модератор</th>
                <th>Телефон</th><th>Дата добавления</th>
                <th style={{ paddingRight: 20, textAlign: 'right' }}>Действия</th>
              </tr>
            </thead>
            <tbody>
              {moderators.map(m => (
                <tr key={m.id}>
                  <td style={{ paddingLeft: 20 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                      <div style={{
                        width: 32, height: 32, borderRadius: 8, flexShrink: 0,
                        background: 'rgba(59,127,245,0.12)', color: '#60a5fa',
                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                        fontSize: 12, fontWeight: 700,
                      }}>
                        {(m.name || m.phone).slice(0, 2).toUpperCase()}
                      </div>
                      <span style={{ color: '#c4d8ef', fontWeight: 600 }}>{m.name || '—'}</span>
                    </div>
                  </td>
                  <td style={{ fontFamily: 'monospace', fontSize: 13 }}>{m.phone}</td>
                  <td style={{ fontSize: 12 }}>{new Date(m.created_at).toLocaleDateString('ru')}</td>
                  <td style={{ paddingRight: 20, textAlign: 'right' }}>
                    <button
                      onClick={() => handleDelete(m)}
                      style={{
                        display: 'inline-flex', alignItems: 'center', gap: 5,
                        padding: '4px 10px', fontSize: 12, fontWeight: 600,
                        background: 'rgba(232,64,64,0.1)', color: '#e84040',
                        border: '1px solid rgba(232,64,64,0.2)', borderRadius: 6, cursor: 'pointer',
                      }}
                    ><IcoTrash size={12} /> Удалить</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {showModal && (
        <Modal title="Новый модератор" onClose={() => { setShowModal(false); setError('') }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            {error && (
              <div style={{ background: 'rgba(232,64,64,0.1)', border: '1px solid rgba(232,64,64,0.2)', color: '#f87171', borderRadius: 8, padding: '10px 14px', fontSize: 13 }}>
                {error}
              </div>
            )}
            <div>
              <label style={{ display: 'block', fontSize: 12, fontWeight: 600, color: '#3d5a7d', marginBottom: 7, textTransform: 'uppercase', letterSpacing: '.05em' }}>
                Телефон <span style={{ color: '#e84040' }}>*</span>
              </label>
              <input
                className="input" placeholder="998901234567"
                value={form.phone}
                onChange={e => setForm(f => ({ ...f, phone: e.target.value }))}
                autoFocus
              />
              <div style={{ fontSize: 12, color: '#1e3350', marginTop: 6 }}>
                Формат: 998XXXXXXXXX. Модератор войдёт через OTP-вход.
              </div>
            </div>
            <div>
              <label style={{ display: 'block', fontSize: 12, fontWeight: 600, color: '#3d5a7d', marginBottom: 7, textTransform: 'uppercase', letterSpacing: '.05em' }}>
                Имя (необязательно)
              </label>
              <input
                className="input" placeholder="Алишер Каримов"
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
              />
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', paddingTop: 4 }}>
              <button className="btn-secondary" onClick={() => setShowModal(false)}>Отмена</button>
              <button className="btn-primary" onClick={handleCreate} disabled={saving || !form.phone}>
                {saving ? 'Создание...' : 'Создать'}
              </button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
