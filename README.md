# HivemKV

Identity-authenticated KV service with scoped access and cross-device sync.

- Designed for small pieces of per-user or identity-locked data.
- e.g Bookmarks, preferences, configuration etc.
- 1,000 short bookmark IDs can fit in under 10 KB.

## Features

- Identity-based authentication
- Frontend and backend access roles
- `public`, `private`, and `unsafe` key scopes
- Real-time updates over server-sent events
- [pebble kv](https://github.com/cockroachdb/pebble) backed storage

## Key Scopes

| Scope | Frontend | Backend |
| --- | --- | --- |
| `public` | Read | Read and write |
| `private` | No access | Read and write |
| `unsafe` | Read and write | Read and write |

Keys use the following format:

```text
identity:<scope>:<key>
```

## Usage

See [`service/.env.example`](./service/.env.example) for configuration.
