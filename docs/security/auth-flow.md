# Auth Flow

Document the end-to-end auth path here:

- anonymous bootstrap
- login
- refresh
- logout
- token invalidation
- gateway auth enforcement

Keep the trust boundaries explicit:

- user auth is not the same as service auth
- internal headers are not a substitute for service identity
