import { useState, useEffect } from 'react'
import { api } from '../api'
import Modal from '../components/Modal'
import { IcoShield, IcoPlus, IcoTrash } from '../icons'

const lbl = {
  display: 'block', fontSize: 12, fontWeight: 600, color: '#374151', marginBottom: 8,
}

export default function Moderators() {
  const [moderators, setModerators] = useState([])
  const [loading, setLoading]       = useState(true)
  const [showModal, setShowModal]   = useState(false)
  const [form, setForm]             = useState({ phone: '', name: '', login: '', password: '' })
  const [saving, setSaving]         = useState(false)
  const [error, setError]           = useState('')

  const load = () => {
    setLoading(true)
    api.users('', 'moderator', 1)
      .then(r => setModerators(r?.data || []))
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  const openModal = () => { setShowModal(true); setError(''); setForm({ phone: '', name: '', login: '', password: '' }) }

  const handleCreate = async () => {
    if (!form.phone || !form.login || !form.password) return
    setError('')
    setSaving(true)
    try {
      await api.createModerator(form.phone, form.name || undefined, form.login, form.password)
      setShowModal(false); load()
    } catch (err) {
      setError(err.message || 'Ошибка создания')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (m) => {
    if (!confirm(`Удалить модератора ${m.phone}?`)) return
    await api.deleteModerator(m.id)
    load()
  }

  return (
    <div style={{ padding: 28, display: 'flex', flexDirection: 'column', gap: 20 }}>

      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <div style={{ fontSize: 20, fontWeight: 800, color: '#0f172a' }}>Модераторы</div>
          <div style={{ fontSize: 13, color: '#64748b', marginTop: 3 }}>Доступ к очереди верификации</div>
        </div>
        <button className="btn-primary" onClick={openModal}>
          <IcoPlus size={14} /> Добавить
        </button>
      </div>

      <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
        {loading ? (
          <div style={{ textAlign: 'center', padding: '48px 0', color: '#94a3b8' }}>Загрузка...</div>
        ) : moderators.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '60px 0' }}>
            <IcoShield size={36} style={{ color: '#e2e8f0', margin: '0 auto 12px', display: 'block' }} />
            <div style={{ color: '#64748b', fontWeight: 600 }}>Нет модераторов</div>
            <div style={{ color: '#94a3b8', fontSize: 13, marginTop: 6 }}>Нажмите «Добавить» чтобы создать</div>
          </div>
        ) : (
          <table className="data-table">
            <thead>
              <tr>
                <th style={{ paddingLeft: 20 }}>Модератор</th>
                <th>Логин</th>
                <th>Телефон</th>
                <th>Добавлен</th>
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
                        background: '#eff6ff', color: '#2563eb',
                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                        fontSize: 12, fontWeight: 700,
                      }}>
                        {(m.name || m.phone).slice(0, 2).toUpperCase()}
                      </div>
                      <span style={{ color: '#0f172a', fontWeight: 600 }}>{m.name || '—'}</span>
                    </div>
                  </td>
                  <td>
                    <span style={{ fontFamily: 'monospace', fontSize: 13, color: '#2563eb' }}>
                      {m.admin_login || '—'}
                    </span>
                  </td>
                  <td style={{ fontFamily: 'monospace', fontSize: 13 }}>{m.phone}</td>
                  <td style={{ fontSize: 12 }}>{new Date(m.created_at).toLocaleDateString('ru')}</td>
                  <td style={{ paddingRight: 20, textAlign: 'right' }}>
                    <button
                      onClick={() => handleDelete(m)}
                      style={{
                        display: 'inline-flex', alignItems: 'center', gap: 5,
                        padding: '4px 10px', fontSize: 12, fontWeight: 600,
                        background: '#fef2f2', color: '#dc2626',
                        border: '1px solid #fecaca', borderRadius: 6, cursor: 'pointer',
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
        <Modal title="Новый модератор" onClose={() => setShowModal(false)}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            {error && (
              <div style={{ background: '#fef2f2', border: '1px solid #fecaca', color: '#dc2626', borderRadius: 8, padding: '10px 14px', fontSize: 13 }}>
                {error}
              </div>
            )}

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
              <div>
                <label style={lbl}>Телефон <span style={{ color: '#ef4444' }}>*</span></label>
                <input className="input" placeholder="998901234567" value={form.phone}
                  onChange={e => setForm(f => ({ ...f, phone: e.target.value }))} autoFocus />
              </div>
              <div>
                <label style={lbl}>Имя</label>
                <input className="input" placeholder="Алишер Каримов" value={form.name}
                  onChange={e => setForm(f => ({ ...f, name: e.target.value }))} />
              </div>
            </div>

            <div style={{ height: 1, background: '#f1f5f9' }} />
            <div style={{ fontSize: 12, color: '#64748b' }}>Учётные данные для входа в панель (логин + пароль)</div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
              <div>
                <label style={lbl}>Логин <span style={{ color: '#ef4444' }}>*</span></label>
                <input className="input" placeholder="moderator1" value={form.login}
                  onChange={e => setForm(f => ({ ...f, login: e.target.value }))} />
              </div>
              <div>
                <label style={lbl}>Пароль <span style={{ color: '#ef4444' }}>*</span></label>
                <input type="password" className="input" placeholder="Мин. 6 символов" value={form.password}
                  onChange={e => setForm(f => ({ ...f, password: e.target.value }))} />
              </div>
            </div>

            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', paddingTop: 4 }}>
              <button className="btn-secondary" onClick={() => setShowModal(false)}>Отмена</button>
              <button className="btn-primary" onClick={handleCreate}
                disabled={saving || !form.phone || !form.login || !form.password}>
                {saving ? 'Создание...' : 'Создать'}
              </button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
