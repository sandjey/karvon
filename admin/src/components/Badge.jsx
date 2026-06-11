const variants = {
  green:  'bg-emerald-100 text-emerald-700',
  red:    'bg-red-100 text-red-700',
  orange: 'bg-orange-100 text-orange-700',
  blue:   'bg-blue-100 text-blue-700',
  gray:   'bg-slate-100 text-slate-600',
  purple: 'bg-purple-100 text-purple-700',
}

export default function Badge({ color = 'gray', children }) {
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${variants[color]}`}>
      {children}
    </span>
  )
}

export function statusBadge(status) {
  const map = {
    approved: { color: 'green',  label: 'Одобрена' },
    pending:  { color: 'orange', label: 'На проверке' },
    rejected: { color: 'red',    label: 'Отклонена' },
    docs_requested: { color: 'blue', label: 'Доп. документы' },
    paid:     { color: 'green',  label: 'Оплачено' },
    pending_payment: { color: 'orange', label: 'Ожидание' },
    failed:   { color: 'red',    label: 'Ошибка' },
    cancelled:{ color: 'gray',   label: 'Отменено' },
    active:   { color: 'green',  label: 'Активно' },
    archived: { color: 'gray',   label: 'Архив' },
    super_admin: { color: 'purple', label: 'Super Admin' },
    moderator:   { color: 'blue',   label: 'Модератор' },
    user:        { color: 'gray',   label: 'Пользователь' },
  }
  const v = map[status] || { color: 'gray', label: status }
  return <Badge color={v.color}>{v.label}</Badge>
}
