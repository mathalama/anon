const API_BASE = process.env.NEXT_PUBLIC_API_URL;
if (!API_BASE) {
  console.error('NEXT_PUBLIC_API_URL is not defined in environment!');
}

export async function fetchWithAuth(path: string, options: RequestInit = {}) {
  const token = localStorage.getItem('access_token');
  const isPublic = path === '/users/anonymous';
  const headers = {
    'Content-Type': 'application/json',
    ...(token && !isPublic ? { Authorization: `Bearer ${token}` } : {}),
    ...options.headers,
  };

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  if (!res.ok) {
    if (res.status === 401) {
      console.warn('Unauthorized! Clearing token...');
      localStorage.removeItem('access_token');
    }
    throw new Error(`API Error: ${res.statusText}`);
  }

  const text = await res.text();
  return text ? JSON.parse(text) : {};
}

export const api = {
  createAnonymous: (deviceId: string) => 
    fetchWithAuth('/users/anonymous', {
      method: 'POST',
      body: JSON.stringify({ device_id: deviceId }),
    }),
  getMe: () => fetchWithAuth('/users/me'),
  updateMe: (gender: string, interests: string[]) =>
    fetchWithAuth('/users/me', {
      method: 'PUT',
      body: JSON.stringify({ gender, interests }),
    }),
  
  search: (filter: any) =>
    fetchWithAuth('/match/search', {
      method: 'POST',
      body: JSON.stringify({ filter }),
    }),
  
  getStatus: () => fetchWithAuth('/match/status'),
  
  cancelSearch: () => fetchWithAuth('/match/search', { method: 'DELETE' }),
  
  getMatchSSEUrl: (token: string) => `${API_BASE}/match/status/events?token=${token || ''}`,
};
