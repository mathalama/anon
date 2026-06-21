import { test, expect } from '@playwright/test';

test.describe.serial('Mathalama Talk API Tests', () => {
  const API_BASE = 'http://127.0.0.1:8080/api/v1'; 
  let accessToken: string;
  const deviceId = `test-device-${Date.now()}`;

  test('Create anonymous user', async ({ request }) => {
    const response = await request.post(`${API_BASE}/users/anonymous`, {
      data: { device_id: deviceId }
    });
    
    expect(response.status()).toBe(200);
    const body = await response.json();
    
    expect(body).toHaveProperty('access_token');
    
    accessToken = body.access_token;
  });

  test('Get current user profile (Me)', async ({ request }) => {
    expect(accessToken).toBeDefined();

    const response = await request.get(`${API_BASE}/users/me`, {
      headers: { Authorization: `Bearer ${accessToken}` }
    });

    expect(response.status()).toBe(200);
    const user = await response.json();
    expect(user).toHaveProperty('id');
  });

  test('Search for partner and cancel', async ({ request }) => {
    expect(accessToken).toBeDefined();

    // Start search
    const searchRes = await request.post(`${API_BASE}/match/search`, {
      headers: { Authorization: `Bearer ${accessToken}` },
      data: {
        filter: {
          my_gender: 'male',
          gender: 'any',
          mode: 'text'
        }
      }
    });
    // 202 Accepted means search started
    expect(searchRes.status()).toBe(202);

    // Get status
    const statusRes = await request.get(`${API_BASE}/match/status`, {
      headers: { Authorization: `Bearer ${accessToken}` }
    });
    expect(statusRes.status()).toBe(200);
    const statusBody = await statusRes.json();
    expect(statusBody.status).toBe('searching'); // Because we just started and have no partner

    // Cancel search
    const cancelRes = await request.delete(`${API_BASE}/match/search`, {
      headers: { Authorization: `Bearer ${accessToken}` }
    });
    expect(cancelRes.status()).toBe(200);
  });

  test('Gateway Health check', async ({ request }) => {
    const response = await request.get(`${API_BASE}/health`);
    expect(response.status()).toBe(200);
    const text = await response.text();
    expect(text).toContain('OK');
  });
});
