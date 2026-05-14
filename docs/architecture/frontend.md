# Frontend Architecture

The frontend is a Next.js app that focuses on a small set of flows:

- landing and static pages
- anonymous session bootstrap
- matchmaking search
- active chat room
- voice call UI
- rules and privacy pages

The browser should not know about internal service topology.
It talks to the gateway, which decides whether to proxy, aggregate, or call an
internal service.

Frontend state and transport live in:

- `apps/web/src/store`
- `apps/web/src/hooks`
- `apps/web/src/lib`
- `apps/web/src/components`
