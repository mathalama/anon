# Authentication

The platform uses an anonymous-first flow with optional persistent login.

Typical flow:

1. the client boots a device or session identifier
2. the client requests an anonymous session or login
3. the backend validates credentials or bootstrap data
4. the backend issues access and refresh state
5. the gateway forwards authenticated requests

Document here:

- login
- anonymous bootstrap
- refresh
- logout
- token invalidation
