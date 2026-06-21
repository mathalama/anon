export const API_ENDPOINTS = {
  USERS: {
    ANONYMOUS: '/users/anonymous',
    ME: '/users/me',
  },
  MATCH: {
    SEARCH: '/match/search',
    STATUS: '/match/status',
    STATUS_EVENTS: '/match/status/events',
    NEXT: '/match/next',
  },
  REPORT: {
    REPORT: '/report/report',
  }
} as const;
