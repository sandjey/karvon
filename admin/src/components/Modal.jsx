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
        style={{ position: 'absolute', inset: 0, background: 'rgba(2,6,15,0.75)' }}
        onClick={onClose}
      />
      <div style={{
        position: 'relative', width: '100%', maxWidth: widths[size],
        maxHeight: '90vh', overflowY: 'auto',
        background: '#0e1a2d',
        border: '1px solid #1e3050',
        borderRadius: 14,
        boxShadow: '0 30px 80px rgba(0,0,0,0.7)',
      }}>
        {/* Header */}
        <div style={{
          display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          padding: '16px 20px', borderBottom: '1px solid #19273d',
        }}>
          <span style={{ color: '#dce8f5', fontWeight: 600, fontSize: 15 }}>{title}</span>
          <button
            onClick={onClose}
            style={{
              background: 'rgba(255,255,255,0.04)', border: '1px solid #19273d',
              borderRadius: 7, width: 28, height: 28, cursor: 'pointer',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              color: '#4d6d90', transition: 'all .13s',
            }}
            onMouseEnter={e => { e.currentTarget.style.background='rgba(232,64,64,0.1)'; e.currentTarget.style.color='#e84040' }}
            onMouseLeave={e => { e.currentTarget.style.background='rgba(255,255,255,0.04)'; e.currentTarget.style.color='#4d6d90' }}
          >
            <IcoX size={13} />
          </button>
        </div>
        {/* Body */}
        <div style={{ padding: '20px' }}>{children}</div>
      </div>
    </div>
  )
}
