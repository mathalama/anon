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

export function setup() {
  const password = 'Password123!';
  const users = [];

  for (let i = 0; i < 50; i++) {
    const email = `user_${randomString(12)}@example.com`;

    // 1) Register
    let res = http.post(
      `${BASE_URL}/users/register`,
      JSON.stringify({ email, password }),
      { headers: { 'Content-Type': 'application/json' } }
    );
    if (res.status !== 201) {
      console.log(`Setup registration failed: ${res.status} ${res.body}`);
      continue;
    }

    // 2) Login
    res = http.post(
      `${BASE_URL}/users/login`,
      JSON.stringify({ email, password }),
      { headers: { 'Content-Type': 'application/json' } }
    );
    if (res.status !== 200) {
      console.log(`Setup login failed: ${res.status} ${res.body}`);
      continue;
    }

    const token = res.json('access_token');
    if (!token) {
      console.log(`Setup missing access_token: ${res.status} ${res.body}`);
      continue;
    }

    users.push({ email, token });
  }

  return { users };
}

export default function (data) {
  const users = data.users || [];
  if (users.length === 0) {
    console.log('No users created in setup()');
    sleep(1);
    return;
  }

  const user = users[Math.floor(Math.random() * users.length)];
  const authToken = user.token;

  const mode = Math.random() < 0.8 ? 'text' : 'voice';
  const myGender = Math.random() < 0.5 ? 'male' : 'female';

  // 1) Get Profile
  let res = http.get(`${BASE_URL}/users/me`, {
    headers: { 'Authorization': `Bearer ${authToken}` },
  });
  if (res.status !== 200) {
    console.log(`Get profile failed: ${res.status} ${res.body}`);
  }
  check(res, { 'got profile': (r) => r.status === 200 });

  // 2) Start Matchmaking
  res = http.post(`${BASE_URL}/match/search`, JSON.stringify({
    filter: {
      mode,
      my_gender: myGender,
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

  sleep(0.5);
}
