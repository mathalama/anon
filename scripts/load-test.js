import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export const options = {
  stages: [
    { duration: '1m', target: 50 }, // ramp up to 50 users
    { duration: '3m', target: 50 }, // stay at 50 users
    { duration: '1m', target: 0 },  // ramp down
  ],
};

const BASE_URL = 'http://127.0.0.1:8080/api/v1';

export default function () {
  const email = `user_${randomString(8)}@example.com`;
  const password = 'Password123!';

  // 1. Register
  let res = http.post(`${BASE_URL}/users/register`, JSON.stringify({
    email: email,
    password: password,
  }), { headers: { 'Content-Type': 'application/json' } });
  
  if (res.status !== 201) {
    console.log(`Registration failed: ${res.status} ${res.body}`);
  }
  check(res, { 'registered': (r) => r.status === 201 });
  if (res.status !== 201) return;

  // 2. Login
  res = http.post(`${BASE_URL}/users/login`, JSON.stringify({
    email: email,
    password: password,
  }), { headers: { 'Content-Type': 'application/json' } });

  check(res, { 'logged in': (r) => r.status === 200 });
  if (res.status !== 200) {
    console.log(`Login failed: ${res.status} ${res.body}`);
    return;
  }
  const authToken = res.json('access_token');
  if (!authToken) {
    console.log(`Login missing access_token: ${res.status} ${res.body}`);
    return;
  }

  // 3. Get Profile
  res = http.get(`${BASE_URL}/users/me`, {
    headers: { 'Authorization': `Bearer ${authToken}` },
  });
  check(res, { 'got profile': (r) => r.status === 200 });
  if (res.status !== 200) {
    console.log(`Get profile failed: ${res.status} ${res.body}`);
    return;
  }

  // 4. Start Matchmaking (Simulated HTTP call)
  res = http.post(`${BASE_URL}/match/search`, JSON.stringify({
    filter: {
      mode: 'text',
      my_gender: 'male',
      gender: 'any',
    },
  }), {
    headers: { 
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json' 
    },
  });
  if (res.status !== 202) {
    console.log(`Match search failed: ${res.status} ${res.body}`);
  }
  check(res, { 'match search started': (r) => r.status === 202 });

  sleep(1);
}
