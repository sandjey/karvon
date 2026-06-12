import { useEffect } from 'react'
import { IcoX } from '../icons'

export default function Modal({ title, onClose, children, size = 'md' }) {
  const widths = { sm: 380, md: 460, lg: 560, xl: 680 }

  useEffect(() => {
    const onKey = (e) => { if (e.key === 'Escape') onClose() }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [onClose])

  return (
    <div style={{
      position: 'fixed', inset: 0, zIndex: 50,
      display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 16,
    }}>
      <div
        style={{ position: 'absolute', inset: 0, background: 'rgba(15,23,42,.45)', backdropFilter: 'blur(2px)' }}
        onClick={onClose}
      />
      <div style={{
        position: 'relative', width: '100%', maxWidth: widths[size],
        maxHeight: '90vh', overflowY: 'auto',
        background: '#ffffff',
        border: '1px solid #e2e8f0',
        borderRadius: 16,
        boxShadow: '0 20px 60px rgba(15,23,42,.15)',
      }}>
        <div style={{
          display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          padding: '16px 20px', borderBottom: '1px solid #f1f5f9',
        }}>
          <span style={{ color: '#0f172a', fontWeight: 700, fontSize: 15 }}>{title}</span>
          <button
            onClick={onClose}
            style={{
              background: 'none', border: '1px solid #e2e8f0',
              borderRadius: 7, width: 28, height: 28, cursor: 'pointer',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              color: '#94a3b8', transition: 'all .12s',
            }}
            onMouseEnter={e => { e.currentTarget.style.background = '#fee2e2'; e.currentTarget.style.borderColor = '#fecaca'; e.currentTarget.style.color = '#dc2626' }}
            onMouseLeave={e => { e.currentTarget.style.background = 'none'; e.currentTarget.style.borderColor = '#e2e8f0'; e.currentTarget.style.color = '#94a3b8' }}
          >
            <IcoX size={13} />
          </button>
        </div>
        <div style={{ padding: '20px' }}>{children}</div>
      </div>
    </div>
  )
}
