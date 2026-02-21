# Cosign

Cosign is a small Go service and CLI for managing public signature campaigns.

It follows the Pollinator style architecture:

- `CLI -> service -> database` for API and automation workflows
- `dashboard -> service API -> database` for admin UI workflows
- `internal/app` owns the server-rendered admin dashboard and UI handlers
- `internal/service` owns business logic and HTTP handlers
- `internal/database` owns SQLite access and migrations
- `command-go` packages provide CLI parsing, envelope I/O, API keys, and CORS

## Build And Test

```bash
make build
make test
```

## Quick Start (Local)

Initialize local credentials:

```bash
make init
```

Run API only:

```bash
make run
```

Run API + dashboard together:

```bash
make run-dashboard
```

## Server

```bash
cosign serve \
  --db-path /var/lib/cosign/cosign.db \
  --port 8080 \
  --cors-allowed-origins http://localhost:3000 \
  --credentials-directory /etc/cosign
```

Required credential file:

- `/etc/cosign/api_key` (or `<credentials-directory>/api_key`)
- format: `{id}.{secret}`

The bootstrap token is used only when the key store is empty.

## Admin Dashboard

Run the API server and the dashboard in separate processes.

```bash
cosign dashboard \
  --port 3000 \
  --api-base-url http://localhost:8080 \
  --api-prefix /api/v1 \
  --credentials-directory ./secrets
```

Dashboard defaults:

- port: `3000`
- API base URL: `http://localhost:8080`
- API prefix: `/api/v1`
- credentials directory: `./secrets`
- API key file: `api_key`

The dashboard server reads `./secrets/api_key` and uses it as a trusted server key when calling admin API routes.

## API Prefix

All routes are mounted at `/api/v1`.

### Public Routes

- `GET /health`
- `GET /campaigns/{campaign_id}`
- `OPTIONS /campaigns/{campaign_id}`
- `GET /campaigns/{campaign_id}/locations`
- `OPTIONS /campaigns/{campaign_id}/locations`
- `GET /campaigns/{campaign_id}/signatures`
- `POST /campaigns/{campaign_id}/signatures`
- `OPTIONS /campaigns/{campaign_id}/signatures`

Public campaign/signature routes enforce CORS whitelist checks.

`POST /campaigns/{campaign_id}/signatures` is IP rate-limited.

### Admin Routes (API Key Required)

- `GET /admin/campaigns`
- `POST /admin/campaigns`
- `GET /admin/campaigns/{campaign_id}`
- `PUT /admin/campaigns/{campaign_id}`
- `DELETE /admin/campaigns/{campaign_id}`
- `GET /admin/campaigns/{campaign_id}/locations`
- `PUT /admin/campaigns/{campaign_id}/locations`
- `GET /admin/campaigns/{campaign_id}/signatures`
- `DELETE /admin/campaigns/{campaign_id}/signatures/{signature_id}`

### Settings Routes (API Key Required)

Managed by `command-go` packages:

- `POST /settings/keys`
- `DELETE /settings/keys/{id}`
- `GET /settings/cors`
- `PUT /settings/cors`

## CLI

Root command includes:

- `serve`
- `dashboard`
- `api`
- `env`
- `status`
- `version`

Global options:

- `--config-dir`
- `--env`
- `--base-url`
- `--api-key`
- `--campaign-id`
- `-v, --verbose`

### Environment Management

```bash
cosign env create local --base-url http://localhost:8080
cosign env key set <token>
```

### Campaign Commands

```bash
cosign api campaign list
cosign api campaign create "Open Letter"
cosign --campaign-id <id> api campaign get
cosign --campaign-id <id> api campaign update "Open Letter 2026" --strict
cosign --campaign-id <id> api campaign locations set --location "New York" --location "Boston"
```

### Signature Commands

```bash
cosign --campaign-id <id> api signatures list --limit 100 --offset 0
cosign --campaign-id <id> api signatures export -o signatures.csv
```

### Settings Commands

```bash
cosign api settings keys create
cosign api settings keys delete <id>
cosign api settings cors get
cosign api settings cors set --url=http://localhost:3000 --url=https://example.org
```
