# Users API

The user surface should separate public profile data from private account data.

Document here:

- profile reads and updates
- anonymous session bootstrap
- ban checks
- public vs private fields

Rule of thumb:

- public reads must not leak account-only data
- owner-only reads can return private account fields
