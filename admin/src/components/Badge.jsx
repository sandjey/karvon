const variants = {
  green:  'badge-green',
  red:    'badge-red',
  orange: 'badge-orange',
  blue:   'badge-blue',
  gray:   'badge-gray',
  purple: 'badge-purple',
}

export default function Badge({ color = 'gray', children }) {
  return <span className={`badge ${variants[color] || 'badge-gray'}`}>{children}</span>
}

export function statusBadge(status) {
  const map = {
    approved:        { color: 'green',  label: 'Одобрена' },
    pending:         { color: 'orange', label: 'На проверке' },
    rejected:        { color: 'red',    label: 'Отклонена' },
    docs_requested:  { color: 'blue',   label: 'Доп. документы' },
    paid:            { color: 'green',  label: 'Оплачено' },
    pending_payment: { color: 'orange', label: 'Ожидание' },
    failed:          { color: 'red',    label: 'Ошибка' },
    cancelled:       { color: 'gray',   label: 'Отменено' },
    active:          { color: 'green',  label: 'Активно' },
    archived:        { color: 'gray',   label: 'Архив' },
    super_admin:     { color: 'purple', label: 'Super Admin' },
    moderator:       { color: 'blue',   label: 'Модератор' },
    user:            { color: 'gray',   label: 'Пользователь' },
  }
  const v = map[status] || { color: 'gray', label: status }
  return <Badge color={v.color}>{v.label}</Badge>
}
