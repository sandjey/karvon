const BASE = import.meta.env.VITE_API_URL || '/api/v1'

function getToken() {
  return localStorage.getItem('admin_token')
}

async function request(method, path, body) {
  const token = getToken()
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(`${BASE}${path}`, {
    method,
    headers,
    body: body != null ? JSON.stringify(body) : undefined,
  })

  if (res.status === 401) {
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_role')
    const base = import.meta.env.BASE_URL || '/'
    window.location.href = base + 'login'
    return
  }

  const data = await res.json()
  if (!data.success && data.error) throw new Error(data.error.message || data.error.code)
  return data
}

const get  = (path)        => request('GET',    path)
const post = (path, body)  => request('POST',   path, body)
const put  = (path, body)  => request('PUT',    path, body)
const patch= (path, body)  => request('PATCH',  path, body)
const del  = (path)        => request('DELETE', path)

export const api = {
  // Auth
  adminLogin: (login, password) => post('/admin/login', { login, password }),

  // Dashboard
  dashboard: (period = '30d') => get(`/admin/dashboard?period=${period}`),

  // Users
  users: (q = '', role = '', page = 1) =>
    get(`/admin/users?q=${q}&role=${role}&page=${page}&per_page=20`),
  user: (id) => get(`/admin/users/${id}`),
  blockUser: (id, blocked) => patch(`/admin/users/${id}/block`, { blocked }),
  topupTokens: (id, amount) => post(`/admin/users/${id}/topup`, { amount }),

  // Moderators
  createModerator: (phone, name, login, password) => post('/admin/moderators', { phone, name, login, password }),
  deleteModerator: (id) => del(`/admin/moderators/${id}`),

  // Listings
  listings: (type = 'cargo', page = 1) =>
    get(`/admin/listings?type=${type}&page=${page}&per_page=20`),
  deleteListing: (type, id) => del(`/admin/listings/${type}/${id}`),
  blockListing: (type, id) => patch(`/admin/listings/${type}/${id}/block`),

  // Companies
  companies: (status = '', page = 1) =>
    get(`/admin/companies?status=${status}&page=${page}&per_page=20`),

  // Payments
  payments: (page = 1) => get(`/admin/payments?page=${page}&per_page=20`),

  // Pricing
  pricing: () => get('/admin/pricing'),
  updatePricing: (key, body) => put(`/admin/pricing/${key}`, body),
  createPricing: (body) => post('/admin/pricing', body),
  deletePricing: (key) => del(`/admin/pricing/${key}`),

  // Moderator queue
  queue: (status = 'pending', page = 1) =>
    get(`/moderator/queue?status=${status}&page=${page}&per_page=20`),
  queueItem: (id) => get(`/moderator/queue/${id}`),
  approve: (id) => post(`/moderator/queue/${id}/approve`),
  reject: (id, reason) => post(`/moderator/queue/${id}/reject`, { reason }),
  requestDocs: (id, message) => post(`/moderator/queue/${id}/request-docs`, { message }),
  history: (page = 1) => get(`/moderator/history?page=${page}&per_page=20`),
}
