# Sirius Future — referral service

Backend for an online school's referral program: users get personal referral links with a usage limit, new students register through them, payments are tracked per user. Includes Redis caching and Prometheus metrics.

![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white)
![Fiber](https://img.shields.io/badge/Fiber-v2-00ACD7)
![Redis](https://img.shields.io/badge/Redis-cache-DC382D?logo=redis&logoColor=white)
![Prometheus](https://img.shields.io/badge/Prometheus-metrics-E6522C?logo=prometheus&logoColor=white)
[![ci](https://github.com/HowkaCoder/sirius_future/actions/workflows/ci.yml/badge.svg)](https://github.com/HowkaCoder/sirius_future/actions/workflows/ci.yml)

## How it works

1. An admin creates a referral link for a user with a limit (`link_limit`).
2. A new student registers with this link: the link counter is increased and the student gets `referrer_id`.
3. When the limit is reached or the link is turned off, registration by this link returns `409`.

The counter is increased with a single conditional `UPDATE … WHERE count < limit` inside the same transaction as the user insert, so the limit holds even when many people register at the same moment (covered by a concurrent test), and a failed registration does not spend a use.

## Architecture

```
cmd/                 entrypoint, routes, Prometheus middleware
internal/
  app.go             database setup
  app/
    entity/          GORM models and validation
    repository/      database access
    usecase/         business rules, caching
    service/         Redis, logger, referral link generation
    handler/         HTTP handlers
```

Redis is optional: without `REDIS_ADDR` the service works without cache. Cached lists are invalidated on every change.

## Getting started

```bash
cp .env.example .env
docker compose up --build     # app, Redis, Prometheus (:9090), Grafana (:3000)
```

or `go run ./cmd` (needs a C compiler for SQLite).

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP port |
| `DB_PATH` | `database/sirius.db` | SQLite file |
| `REDIS_ADDR` | — | Redis address, cache is off if empty |
| `ADMIN_TOKEN` | — | Bearer token for admin endpoints |

## API

| Method | Path | Access |
|---|---|---|
| `POST` | `/register` | public, optional `url` field with referral link |
| `GET` | `/api/check-link/:url` | public, `{"result": true}` if the link can still be used |
| `GET` | `/api/get-referrer/:url` | public |
| `POST` | `/api/create-link` | admin, `{"user_id", "link_limit"}` |
| `GET` | `/api/users`, `/api/links` | admin |
| `POST`, `GET` | `/api/payments` | admin |
| `GET` | `/api/payments/user/:id` | admin |
| `PATCH` | `/api/payments/:id` | admin |
| `GET` | `/metrics` | Prometheus |

Passwords are hashed with bcrypt and never returned. Public registration always creates role `user`.

## Tests

```bash
go test -race ./...
```
