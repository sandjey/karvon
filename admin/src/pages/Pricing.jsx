import { useState, useEffect } from 'react'
import { api } from '../api'
import Modal from '../components/Modal'

export default function Pricing() {
  const [items, setItems]   = useState([])
  const [loading, setLoading] = useState(true)
  const [editItem, setEditItem] = useState(null)
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm]     = useState({})
  const [saving, setSaving] = useState(false)

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
    setForm({
      label: item.label,
      price_uzs: item.price_uzs,
      price_usd: item.price_usd,
      tokens_amount: item.tokens_amount,
      duration_days: item.duration_days,
      is_active: item.is_active,
    })
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      await api.updatePricing(editItem.key, form)
      setEditItem(null)
      load()
    } catch (err) {
      alert(err.message)
    } finally {
      setSaving(false)
    }
  }

  const handleCreate = async () => {
    setSaving(true)
    try {
      await api.createPricing(form)
      setShowCreate(false)
      setForm({})
      load()
    } catch (err) {
      alert(err.message)
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (item) => {
    if (!confirm(`Удалить тариф "${item.label}"?`)) return
    await api.deletePricing(item.key)
    load()
  }

  const PricingForm = () => (
    <div className="space-y-3">
      {showCreate && (
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">Ключ <span className="text-red-500">*</span></label>
          <input className="input" placeholder="tokens_premium" value={form.key || ''} onChange={e => setForm(f=>({...f,key:e.target.value}))} />
        </div>
      )}
      <div>
        <label className="block text-sm font-medium text-slate-700 mb-1">Название</label>
        <input className="input" value={form.label || ''} onChange={e => setForm(f=>({...f,label:e.target.value}))} />
      </div>
      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">Цена UZS</label>
          <input type="number" className="input" value={form.price_uzs || ''} onChange={e => setForm(f=>({...f,price_uzs:parseFloat(e.target.value)||0}))} />
        </div>
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">Цена USD</label>
          <input type="number" className="input" value={form.price_usd || ''} onChange={e => setForm(f=>({...f,price_usd:parseFloat(e.target.value)||0}))} />
        </div>
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">Токенов</label>
          <input type="number" className="input" value={form.tokens_amount || ''} onChange={e => setForm(f=>({...f,tokens_amount:parseInt(e.target.value)||0}))} />
        </div>
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">Дней</label>
          <input type="number" className="input" value={form.duration_days || ''} onChange={e => setForm(f=>({...f,duration_days:parseInt(e.target.value)||0}))} />
        </div>
      </div>
      <label className="flex items-center gap-2 cursor-pointer">
        <input type="checkbox" className="w-4 h-4 accent-blue-600" checked={form.is_active !== false} onChange={e => setForm(f=>({...f,is_active:e.target.checked}))} />
        <span className="text-sm text-slate-700">Активен</span>
      </label>
    </div>
  )

  const groups = {
    'Токены': items.filter(i => i.key.startsWith('tokens')),
    'Подписки': items.filter(i => i.key.startsWith('sub')),
    'Буст': items.filter(i => i.key.startsWith('boost')),
    'Объявления': items.filter(i => i.key.startsWith('listing')),
    'Прочее': items.filter(i => !i.key.match(/^(tokens|sub|boost|listing)/)),
  }

  return (
    <div className="p-6 space-y-5">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-800">Тарифы</h1>
          <p className="text-sm text-slate-500">Цены и пакеты платформы</p>
        </div>
        <button className="btn-primary" onClick={() => { setShowCreate(true); setForm({ is_active: true }) }}>+ Новый тариф</button>
      </div>

      {loading ? (
        <div className="text-center py-12 text-slate-400">Загрузка...</div>
      ) : (
        Object.entries(groups).filter(([,v]) => v.length > 0).map(([group, groupItems]) => (
          <div key={group}>
            <h2 className="text-sm font-semibold text-slate-500 uppercase tracking-wider mb-3">{group}</h2>
            <div className="card overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-left text-slate-500 border-b border-slate-100">
                    <th className="pb-3 font-medium">Тариф</th>
                    <th className="pb-3 font-medium">Ключ</th>
                    <th className="pb-3 font-medium">Цена UZS</th>
                    <th className="pb-3 font-medium">Цена USD</th>
                    <th className="pb-3 font-medium">Токены</th>
                    <th className="pb-3 font-medium">Дней</th>
                    <th className="pb-3 font-medium">Статус</th>
                    <th className="pb-3"></th>
                  </tr>
                </thead>
                <tbody>
                  {groupItems.map(item => (
                    <tr key={item.key} className="border-b border-slate-50 hover:bg-slate-50 group">
                      <td className="py-3 font-medium text-slate-800">{item.label}</td>
                      <td className="py-3 font-mono text-xs text-slate-400">{item.key}</td>
                      <td className="py-3">{item.price_uzs ? item.price_uzs.toLocaleString() + ' сум' : '—'}</td>
                      <td className="py-3">{item.price_usd ? '$' + item.price_usd : '—'}</td>
                      <td className="py-3">{item.tokens_amount || '—'}</td>
                      <td className="py-3">{item.duration_days || '—'}</td>
                      <td className="py-3">
                        {item.is_active
                          ? <span className="inline-flex px-2 py-0.5 rounded-full text-xs bg-emerald-100 text-emerald-700 font-medium">Активен</span>
                          : <span className="inline-flex px-2 py-0.5 rounded-full text-xs bg-slate-100 text-slate-500 font-medium">Отключён</span>
                        }
                      </td>
                      <td className="py-3">
                        <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition">
                          <button onClick={() => openEdit(item)} className="px-3 py-1 bg-blue-50 text-blue-600 hover:bg-blue-100 rounded text-xs font-medium">Изменить</button>
                          <button onClick={() => handleDelete(item)} className="px-3 py-1 bg-red-50 text-red-600 hover:bg-red-100 rounded text-xs font-medium">Удалить</button>
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
        <Modal title={`Редактировать: ${editItem.label}`} onClose={() => setEditItem(null)}>
          <PricingForm />
          <div className="flex gap-3 justify-end mt-5">
            <button className="btn-secondary" onClick={() => setEditItem(null)}>Отмена</button>
            <button className="btn-primary" onClick={handleSave} disabled={saving}>{saving ? 'Сохранение...' : 'Сохранить'}</button>
          </div>
        </Modal>
      )}

      {showCreate && (
        <Modal title="Новый тариф" onClose={() => setShowCreate(false)}>
          <PricingForm />
          <div className="flex gap-3 justify-end mt-5">
            <button className="btn-secondary" onClick={() => setShowCreate(false)}>Отмена</button>
            <button className="btn-primary" onClick={handleCreate} disabled={saving}>{saving ? 'Создание...' : 'Создать'}</button>
          </div>
        </Modal>
      )}
    </div>
  )
}
