const BASE = ''

function getCSRFToken(): string {
  const m = document.cookie.match(/(?:^| )csrf_token=([^;]+)/)
  return m ? m[1] : ''
}

async function request(method: string, url: string, body?: unknown) {
  const headers: Record<string, string> = {
    'X-CSRF-Token': getCSRFToken(),
  }
  const token = localStorage.getItem('token')
  if (token) headers['Authorization'] = `Bearer ${token}`
  if (body && !(body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
  }
  const res = await fetch(`${BASE}${url}`, {
    method,
    headers,
    body: body instanceof FormData ? body : body ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || 'Request failed')
  }
  if (res.status === 204) return null
  return res.json()
}

export const api = {
  get: (url: string) => request('GET', url),
  post: (url: string, body?: unknown) => request('POST', url, body),
  put: (url: string, body?: unknown) => request('PUT', url, body),
  delete: (url: string) => request('DELETE', url),
}
