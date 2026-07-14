const BASE = ''

function getCSRFToken(): string {
  const m = document.cookie.match(/(?:^| )csrf_token=([^;]+)/)
  return m ? m[1] : ''
}

async function request(method: string, url: string, body?: unknown) {
  const headers: Record<string, string> = {
    'X-CSRF-Token': getCSRFToken(),
  }
  if (body && !(body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
  }
  const res = await fetch(`${BASE}${url}`, {
    method,
    headers,
    body: body instanceof FormData ? body : body ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) {
    const text = await res.text().catch(() => '')
    try { const err = JSON.parse(text); throw new Error(err.error || 'Request failed') }
    catch (e: any) { throw new Error(e.message || res.statusText || 'Request failed') }
  }
  const text = await res.text()
  if (!text) return null
  try { return JSON.parse(text) } catch { return null }
}

export const api = {
  get: (url: string) => request('GET', url),
  post: (url: string, body?: unknown) => request('POST', url, body),
  put: (url: string, body?: unknown) => request('PUT', url, body),
  delete: (url: string) => request('DELETE', url),
}
