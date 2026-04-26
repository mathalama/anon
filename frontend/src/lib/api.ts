const API_BASE = typeof window !== 'undefined' && window.location.hostname !== 'localhost' 
  ? `http://${window.location.hostname}:8080/api/v1`
  : 'http://localhost:8080/api/v1';

export async function fetchWithAuth(path: string, options: RequestInit = {}) {
  const token = localStorage.getItem('access_token');
  const headers = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options.headers,
  };

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  if (!res.ok) {
    if (res.status === 401) {
      // Handle token expiration?
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
  loginTelegram: (data: any) =>
    fetchWithAuth('/users/auth/telegram', {
      method: 'POST',
      body: JSON.stringify(data),
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
};
