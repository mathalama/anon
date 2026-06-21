import { API_ENDPOINTS } from './endpoints';

function normalizeApiBase(raw?: string) {
  if (!raw) return '';
  const trimmed = raw.replace(/\/+$/, '');
  return trimmed.endsWith('/api/v1') ? trimmed : `${trimmed}/api/v1`;
}

const API_BASE = normalizeApiBase(process.env.NEXT_PUBLIC_API_URL);
if (!API_BASE) console.error('NEXT_PUBLIC_API_URL is not defined in environment!');

function isBrowser() {
  return typeof window !== 'undefined' && typeof localStorage !== 'undefined';
}

function getAccessToken() {
  if (!isBrowser()) return null;
  try {
    return localStorage.getItem('access_token');
  } catch {
    return null;
  }
}

function clearAccessToken() {
  if (!isBrowser()) return;
  try {
    localStorage.removeItem('access_token');
  } catch {
    // ignore
  }
}

function buildHeaders(options: RequestInit, path: string) {
  const headers = new Headers(options.headers);

  const token = getAccessToken();
  const isPublic = path === API_ENDPOINTS.USERS.ANONYMOUS;
  if (token && !isPublic) headers.set('Authorization', `Bearer ${token}`);

  const hasContentType = headers.has('Content-Type');
  const body = options.body as unknown;
  const isFormData =
    typeof FormData !== 'undefined' && body instanceof FormData;
  const isBlob = typeof Blob !== 'undefined' && body instanceof Blob;
  const isArrayBuffer =
    typeof ArrayBuffer !== 'undefined' && body instanceof ArrayBuffer;

  if (!hasContentType && body != null && !isFormData && !isBlob && !isArrayBuffer) {
    headers.set('Content-Type', 'application/json');
  }

  return headers;
}

async function parseResponse(res: Response) {
  if (res.status === 204) return null;
  const contentType = res.headers.get('content-type') || '';
  if (contentType.includes('application/json')) return res.json();
  return res.text();
}

export async function fetchWithAuth(path: string, options: RequestInit = {}) {
  if (!API_BASE) throw new Error('API base URL is not configured (NEXT_PUBLIC_API_URL).');

  const headers = buildHeaders(options, path);
  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });

  if (!res.ok) {
    if (res.status === 401) {
      console.warn('Unauthorized! Clearing token...');
      clearAccessToken();
    }

    const body = await res.text().catch(() => '');
    const bodyPreview = body.length > 500 ? `${body.slice(0, 500)}...` : body;
    throw new Error(
      `API Error: ${res.status} ${res.statusText}${bodyPreview ? ` - ${bodyPreview}` : ''}`,
    );
  }

  return parseResponse(res);
}

type SearchFilter = {
  my_gender: 'male' | 'female' | '';
  gender: 'any' | 'male' | 'female' | '';
  mode: 'text' | 'voice' | '';
  room_topic?: string;
};

export const api = {
  createAnonymous: (deviceId: string, turnstileToken?: string) => 
    fetchWithAuth(API_ENDPOINTS.USERS.ANONYMOUS, {
      method: 'POST',
      body: JSON.stringify({ device_id: deviceId, 'cf-turnstile-response': turnstileToken }),
    }),
  getMe: () => fetchWithAuth(API_ENDPOINTS.USERS.ME),
  updateMe: (gender: string, interests: string[]) =>
    fetchWithAuth(API_ENDPOINTS.USERS.ME, {
      method: 'PUT',
      body: JSON.stringify({ gender, interests }),
    }),
  
  search: (filter: SearchFilter, signal?: AbortSignal) =>
    fetchWithAuth(API_ENDPOINTS.MATCH.SEARCH, {
      method: 'POST',
      body: JSON.stringify({ filter }),
      signal,
    }),
  
  getStatus: () => fetchWithAuth(API_ENDPOINTS.MATCH.STATUS),
  
  cancelSearch: () => fetchWithAuth(API_ENDPOINTS.MATCH.SEARCH, { method: 'DELETE' }),

  next: () => fetchWithAuth(API_ENDPOINTS.MATCH.NEXT, { method: 'POST' }),

  reportUser: (roomId: string, reportedUserId: string, reason: string) =>
    fetchWithAuth(API_ENDPOINTS.REPORT.REPORT, {
      method: 'POST',
      body: JSON.stringify({
        room_id: roomId,
        reported_user_id: reportedUserId,
        reason,
      }),
    }),

  
  getMatchSSEUrl: (token: string) =>
    `${API_BASE}${API_ENDPOINTS.MATCH.STATUS_EVENTS}?token=${encodeURIComponent(token || '')}`,
};

