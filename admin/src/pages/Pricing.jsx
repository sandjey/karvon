import { useState, useEffect } from 'react'
import { api } from '../api'
import Modal from '../components/Modal'
import { IcoTag, IcoPlus, IcoPencil, IcoTrash } from '../icons'

const GROUPS = [
  { key: 'tokens',  label: 'Пакеты токенов' },
  { key: 'sub',     label: 'Подписки' },
  { key: 'boost',   label: 'Продвижение' },
  { key: 'listing', label: 'Объявления' },
  { key: 'other',   label: 'Прочее' },
]

const lbl = { display: 'block', fontSize: 12, fontWeight: 600, color: '#374151', marginBottom: 7 }

function PricingForm({ form, setForm, isCreate }) {
  const field = (label, name, type = 'text', placeholder = '') => (
    <div>
      <label style={lbl}>{label}</label>
      <input
        type={type} className="input" placeholder={placeholder}
        value={form[name] ?? ''}
        onChange={e => setForm(f => ({ ...f, [name]: type === 'number' ? (parseFloat(e.target.value) || 0) : e.target.value }))}
      />
    </div>
  )
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
      {isCreate && field('Ключ *', 'key', 'text', 'tokens_premium')}
      {field('Название', 'label', 'text', 'Пакет токенов')}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
        {field('Цена UZS', 'price_uzs', 'number', '50000')}
        {field('Цена USD', 'price_usd', 'number', '5')}
        {field('Токенов', 'tokens_amount', 'number', '20')}
        {field('Дней', 'duration_days', 'number', '0')}
      </div>
      <label style={{ display: 'flex', alignItems: 'center', gap: 10, cursor: 'pointer' }}>
        <input
          type="checkbox" checked={form.is_active !== false}
          onChange={e => setForm(f => ({ ...f, is_active: e.target.checked }))}
          style={{ width: 16, height: 16, accentColor: '#2563eb' }}
        />
        <span style={{ fontSize: 13, color: '#374151' }}>Активен</span>
      </label>
    </div>
  )
}

export default function Pricing() {
  const [items, setItems]           = useState([])
  const [loading, setLoading]       = useState(true)
  const [editItem, setEditItem]     = useState(null)
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm]             = useState({})
  const [saving, setSaving]         = useState(false)

  const load = () => {
    setLoading(true)
    api.pricing()
      .then(r => setItems(r?.data || []))
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  const openEdit = (item) => {
    setEditItem(item)
    setForm({ label: item.label, price_uzs: item.price_uzs, price_usd: item.price_usd, tokens_amount: item.tokens_amount, duration_days: item.duration_days, is_active: item.is_active })
  }

  const handleSave = async () => {
    setSaving(true)
    try { await api.updatePricing(editItem.key, form); setEditItem(null); load() }
    catch (err) { alert(err.message) }
    finally { setSaving(false) }
  }

  const handleCreate = async () => {
    setSaving(true)
    try { await api.createPricing(form); setShowCreate(false); setForm({}); load() }
    catch (err) { alert(err.message) }
    finally { setSaving(false) }
  }

  const handleDelete = async (item) => {
    if (!confirm(`Удалить тариф "${item.label}"?`)) return
    await api.deletePricing(item.key); load()
  }

  const grouped = GROUPS.map(g => ({
    ...g,
    items: g.key === 'other'
      ? items.filter(i => !i.key.match(/^(tokens|sub|boost|listing)/))
      : items.filter(i => i.key.startsWith(g.key)),
  })).filter(g => g.items.length > 0)

  return (
    <div style={{ padding: 28, display: 'flex', flexDirection: 'column', gap: 24 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <div style={{ fontSize: 20, fontWeight: 800, color: '#0f172a' }}>Тарифы</div>
          <div style={{ fontSize: 13, color: '#64748b', marginTop: 3 }}>Цены и пакеты платформы</div>
        </div>
        <button className="btn-primary" onClick={() => { setShowCreate(true); setForm({ is_active: true }) }}>
          <IcoPlus size={14} /> Новый тариф
        </button>
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: '60px 0', color: '#94a3b8' }}>Загрузка...</div>
      ) : grouped.length === 0 ? (
        <div className="card" style={{ textAlign: 'center', padding: '60px 0', color: '#94a3b8' }}>
          <IcoTag size={36} style={{ margin: '0 auto 12px', display: 'block', color: '#e2e8f0' }} />
          <div style={{ fontWeight: 600, color: '#64748b' }}>Нет тарифов</div>
          <div style={{ fontSize: 13, marginTop: 6 }}>Нажмите «Новый тариф» чтобы добавить</div>
        </div>
      ) : (
        grouped.map(group => (
          <div key={group.key}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 7, marginBottom: 10 }}>
              <IcoTag size={13} style={{ color: '#94a3b8' }} />
              <span style={{ fontSize: 11, fontWeight: 700, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '.08em' }}>
                {group.label}
              </span>
            </div>
            <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
              <table className="data-table">
                <thead>
                  <tr>
                    <th style={{ paddingLeft: 20 }}>Тариф</th>
                    <th>Ключ</th><th>Цена UZS</th><th>Цена USD</th>
                    <th>Токены</th><th>Дней</th><th>Статус</th>
                    <th style={{ paddingRight: 20 }}></th>
                  </tr>
                </thead>
                <tbody>
                  {group.items.map(item => (
                    <tr key={item.key}>
                      <td style={{ paddingLeft: 20, color: '#0f172a', fontWeight: 600 }}>{item.label}</td>
                      <td>
                        <code style={{ fontSize: 11.5, color: '#64748b', background: '#f8fafc', padding: '2px 6px', borderRadius: 4, border: '1px solid #e2e8f0' }}>{item.key}</code>
                      </td>
                      <td style={{ color: '#0f172a', fontWeight: 600 }}>{item.price_uzs ? item.price_uzs.toLocaleString() + ' сум' : '—'}</td>
                      <td>{item.price_usd ? '$' + item.price_usd : '—'}</td>
                      <td style={{ color: '#2563eb', fontWeight: 600 }}>{item.tokens_amount || '—'}</td>
                      <td>{item.duration_days || '—'}</td>
                      <td>
                        {item.is_active
                          ? <span className="badge badge-green">Активен</span>
                          : <span className="badge badge-gray">Откл.</span>
                        }
                      </td>
                      <td style={{ paddingRight: 20 }}>
                        <div style={{ display: 'flex', gap: 6, justifyContent: 'flex-end' }}>
                          <button
                            onClick={() => openEdit(item)}
                            style={{
                              display: 'inline-flex', alignItems: 'center', gap: 4,
                              padding: '4px 10px', fontSize: 12, fontWeight: 600,
                              background: '#eff6ff', color: '#2563eb',
                              border: '1px solid #bfdbfe', borderRadius: 6, cursor: 'pointer',
                            }}
                          ><IcoPencil size={11} /> Изменить</button>
                          <button
                            onClick={() => handleDelete(item)}
                            style={{
                              display: 'inline-flex', alignItems: 'center', gap: 4,
                              padding: '4px 10px', fontSize: 12, fontWeight: 600,
                              background: '#fef2f2', color: '#dc2626',
                              border: '1px solid #fecaca', borderRadius: 6, cursor: 'pointer',
                            }}
                          ><IcoTrash size={11} /> Удалить</button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        ))
      )}

      {editItem && (
        <Modal title={`Изменить: ${editItem.label}`} onClose={() => setEditItem(null)}>
          <PricingForm form={form} setForm={setForm} isCreate={false} />
          <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 20 }}>
            <button className="btn-secondary" onClick={() => setEditItem(null)}>Отмена</button>
            <button className="btn-primary" onClick={handleSave} disabled={saving}>{saving ? 'Сохранение...' : 'Сохранить'}</button>
          </div>
        </Modal>
      )}

      {showCreate && (
        <Modal title="Новый тариф" onClose={() => setShowCreate(false)}>
          <PricingForm form={form} setForm={setForm} isCreate />
          <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 20 }}>
            <button className="btn-secondary" onClick={() => setShowCreate(false)}>Отмена</button>
            <button className="btn-primary" onClick={handleCreate} disabled={saving}>{saving ? 'Создание...' : 'Создать'}</button>
          </div>
        </Modal>
      )}
    </div>
  )
}
