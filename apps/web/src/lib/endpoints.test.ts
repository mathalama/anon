import { describe, it, expect } from 'vitest';
import { API_ENDPOINTS } from './endpoints';

describe('API Endpoints Constants', () => {
  it('should have correct users paths', () => {
    expect(API_ENDPOINTS.USERS.ME).toBe('/users/me');
    expect(API_ENDPOINTS.USERS.ANONYMOUS).toBe('/users/anonymous');
  });

  it('should have correct match paths', () => {
    expect(API_ENDPOINTS.MATCH.SEARCH).toBe('/match/search');
    expect(API_ENDPOINTS.MATCH.STATUS).toBe('/match/status');
    expect(API_ENDPOINTS.MATCH.STATUS_EVENTS).toBe('/match/status/events');
  });

  it('should have correct report paths', () => {
    expect(API_ENDPOINTS.REPORT.REPORT).toBe('/report/report');
  });
});
