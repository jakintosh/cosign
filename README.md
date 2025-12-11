# Cosign

A lightweight public letter sign-on management system with HTTP API and CLI tooling.

## Features

- **Public sign-on submission** via REST API with rate limiting
- **Configurable location field** - strict dropdown, dropdown with custom option, or fully open
- **Email validation** and optional duplicate prevention
- **CORS whitelist** for web form domains
- **API key authentication** for admin operations
- **CLI tool** for management and data export
- **SQLite database** with automatic migrations

## Quick Start

### Build

```bash
go build -o cosign ./cmd/cosign
```

### Start the Server

```bash
# Start with default settings (creates ./cosign.db)
./cosign serve

# Start with custom database path and WAL mode
./cosign serve --db /path/to/database.db --wal

# Create a bootstrap API key on startup
./cosign serve --bootstrap-key admin
```

The server will start on `:8080` by default. Use `--port` to customize.

### Create Your First API Key

Once the server is running, you can create additional API keys:

```bash
./cosign --key {bootstrap-key-here} api keys create my-admin-key
```

Save the returned full key (format: `id.secret`) - it won't be shown again!

## API Endpoints

### Public Endpoints (CORS + Rate Limited)

**Submit a sign-on:**
```bash
POST /api/v1/signons
Content-Type: application/json

{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "location": "New York, NY"
}
```

**List all sign-ons:**
```bash
GET /api/v1/signons?limit=100&offset=0
```

**Get location configuration:**
```bash
GET /api/v1/location-config
```

### Admin Endpoints (Require Authentication)

All admin endpoints require an `Authorization: Bearer {key}` header.

**Manage sign-ons:**
```bash
GET    /api/v1/admin/signons          # List with pagination
DELETE /api/v1/admin/signons/{id}     # Delete by ID
```

**Configure location field:**
```bash
GET /api/v1/admin/location-config
PUT /api/v1/admin/location-config
    {"allow_custom_text": true}

GET    /api/v1/admin/location-config/options      # List preset options
POST   /api/v1/admin/location-config/options      # Add option
       {"value": "California", "display_order": 1}
PUT    /api/v1/admin/location-config/options/{id} # Update option
DELETE /api/v1/admin/location-config/options/{id} # Remove option
```

**Manage API keys:**
```bash
GET    /api/v1/admin/keys         # List all keys (IDs only)
POST   /api/v1/admin/keys         # Create new key
       {"id": "optional-key-id"}
DELETE /api/v1/admin/keys/{id}    # Delete key
```

**Manage CORS origins:**
```bash
GET    /api/v1/admin/cors            # List allowed origins
POST   /api/v1/admin/cors            # Add origin
       {"origin": "https://example.com"}
DELETE /api/v1/admin/cors/{origin}  # Remove origin
```

**Health check:**
```bash
GET /api/v1/health
```

## CLI Usage

All commands support global options:
- `--db` - Database file path (default: `./cosign.db`)
- `--url` - API base URL (default: `http://localhost:8080`)
- `--key` - API key for authenticated operations

### Sign-on Management

```bash
# List sign-ons
./cosign --key {your-key} api signons list --limit 50

# Export to CSV
./cosign --key {your-key} api signons export -o signons.csv
```

### Location Configuration

```bash
# View current configuration
./cosign api location-config get

# Allow custom text (dropdown + "Other" option)
./cosign --key {your-key} api location-config set --allow-custom

# Strict dropdown only
./cosign --key {your-key} api location-config set --strict

# Manage preset options
./cosign --key {your-key} api location-options list
./cosign --key {your-key} api location-options add "California" --order 1
./cosign --key {your-key} api location-options remove {id}
```

### API Key Management

```bash
# Create a new key
./cosign --key {your-key} api keys create my-new-key

# Delete a key
./cosign --key {your-key} api keys delete my-old-key
```

### CORS Management

```bash
# List allowed origins
./cosign --key {your-key} api cors list

# Add an origin
./cosign --key {your-key} api cors add https://myform.example.com

# Remove an origin
./cosign --key {your-key} api cors remove https://old-site.com
```

## Architecture

Follows the "API, Database, Service" pattern from STYLE.md:

```
cosign/
├── cmd/cosign/         # CLI entry point and commands
├── internal/
│   ├── api/            # HTTP handlers and routing
│   ├── service/        # Business logic with store interfaces
│   ├── database/       # SQLite persistence implementations
│   └── util/           # Shared utilities
├── init/               # Deployment configs (systemd, etc.)
└── scripts/            # Build and deployment automation
```

**Dependency flow:** API → Service → Database

## Configuration

### Location Field Modes

1. **Strict dropdown** (`allow_custom_text: false` with options configured)
   - Users must select from preset location options
   - Validation rejects any custom text

2. **Dropdown with custom option** (`allow_custom_text: true` with options)
   - Users can select from preset options OR enter custom text
   - Provides structure while allowing flexibility

3. **Free text** (`allow_custom_text: true` with no options OR no options configured)
   - Users can enter any location string
   - Maximum flexibility, minimal structure

### Working with Location Config and Options

- **Defaults:** On first run the config is `allow_custom_text=true` with no options, so any location is accepted.
- **Make it a strict dropdown:**
  - Set `allow_custom_text` to false: `PUT /api/v1/admin/location-config {"allow_custom_text":false}` or `./cosign --key {admin} api location-config set --strict`
  - Add options with order: `POST /api/v1/admin/location-config/options {"value":"New York, NY","display_order":1}` or `./cosign --key {admin} api location-options add "New York, NY" --order 1`
  - Update/remove with `PUT/DELETE /api/v1/admin/location-config/options/{id}`; options are ordered by `display_order` ascending.
- **Allow a dropdown + “Other”:** Keep options, but set `allow_custom_text` to true (`./cosign --key {admin} api location-config set --allow-custom`). Validation accepts either a preset value or any custom text.
- **Front-end usage:** Before rendering a form, call the public endpoint `GET /api/v1/location-config` (CORS applies). The response shape is:
  ```json
  {
    "config": {"allow_custom_text": true},
    "options": [{"id":1,"value":"New York, NY","display_order":1}]
  }
  ```
  - If `options` is non-empty, render them in a dropdown ordered by `display_order`.
  - If `allow_custom_text` is true, include an “Other”/free-text input; if false, block custom entries.
  - Submit sign-ons to `POST /api/v1/signons` with the chosen `location` value.
- **CORS:** Add your form’s origin via admin CORS endpoints (`./cosign --key {admin} api cors add https://your-form.example`) so the browser can call the public config and sign-on endpoints.

### Rate Limiting

Public endpoints are rate limited to:
- 10 requests per second per IP
- Burst allowance of 20 requests

### Database

- **Engine:** SQLite via modernc.org/sqlite (pure Go, no CGo)
- **Migrations:** Automatic on startup via PRAGMA user_version
- **WAL mode:** Optional (use `--wal` flag for better concurrency)
- **Connection pooling:** Max 1 connection (SQLite best practice for writes)

## Security

- **API keys:** SHA256 hashed with random salt, constant-time comparison
- **CORS:** Whitelist-based origin validation, returns 403 for disallowed origins
- **Rate limiting:** IP-based to prevent abuse
- **Email validation:** Regex-based format checking
- **SQL injection:** Prevented via parameterized queries

## Production Deployment

### Systemd Service

Create `/etc/systemd/system/cosign.service`:

```ini
[Unit]
Description=Cosign Public Letter Sign-On Service
After=network.target

[Service]
Type=simple
User=cosign
Group=cosign
WorkingDirectory=/var/lib/cosign
ExecStart=/usr/local/bin/cosign serve --db /var/lib/cosign/cosign.db --wal --port :8080
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable cosign
sudo systemctl start cosign
```

### Reverse Proxy (nginx)

```nginx
server {
    listen 443 ssl http2;
    server_name api.example.com;

    ssl_certificate /etc/letsencrypt/live/api.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.example.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Development

```bash
# Run tests (TODO: add tests)
go test ./...

# Format code
go fmt ./...

# Run linter
golangci-lint run

# Hot reload during development
find . -name "*.go" | entr -r go run ./cmd/cosign serve
```

## License

[Add your license here]
