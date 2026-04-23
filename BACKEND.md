# Backend Implementation

The backend owns all Attio API calls and all Attio credentials. Do not put Attio credentials in the Chrome extension.

## Product Goal

The backend supports two Chrome extension flows:

1. Recognize a known company: when the extension loads on a LinkedIn company page, it asks the backend whether that LinkedIn URL or domain has already been synced.
2. Save to Attio: when the user clicks sync, the extension sends extracted LinkedIn company data to the backend. The backend validates it, asserts the company in Attio, stores the local mapping in PostgreSQL, and returns the Attio record link.

The MVP optimizes for one working workspace and learning value. It uses a single backend-held Attio API key. OAuth is the production path for many users or many workspaces.

## Runtime and Configuration

- Runtime: Go 1.25 or newer.
- Database: PostgreSQL.
- Startup: `docker compose up` starts the database and backend. The backend runs its idempotent migration on startup.
- Router: `chi`.
- PostgreSQL driver: `pgxpool`.
- Migrations: startup SQL in the backend, mirrored in `backend/migrations` for review.

Required environment variables:

- `ATTIO_API_TOKEN`: backend-only API key for the single Attio workspace.
- `DATABASE_URL`: PostgreSQL connection string. Docker Compose sets this for the backend container.
- `PORT`: backend port, default `8080`.
- `ALLOWED_EXTENSION_ORIGINS`: comma-separated origins for CORS during local development.
- `ATTIO_BASE_URL`: optional, default `https://api.attio.com`.

Never return or log the Attio token.

## API Contract

### `GET /healthz`

Success response:

```json
{
  "status": "ok",
  "database": "ok"
}
```

### `GET /api/companies/lookup`

Query parameters:

- `linkedin_url` required.
- `domain` optional.
- `name` optional and currently ignored.

Matching rules:

1. Normalize and match by LinkedIn company URL.
2. If no LinkedIn match exists and `domain` is present, normalize and match by domain.
3. Return HTTP 404 when no local mapping exists. The MVP does not call Attio for lookup.

Found response:

```json
{
  "status": "found",
  "company": {
    "name": "Attio",
    "linkedin_url": "https://www.linkedin.com/company/attio",
    "domain": "attio.com",
    "attio_record_id": "record-id",
    "attio_web_url": "https://app.attio.com/...",
    "first_synced_at": "2026-04-20T10:00:00Z",
    "last_synced_at": "2026-04-20T10:00:00Z"
  }
}
```

Not found response:

```json
{
  "status": "not_found"
}
```

### `POST /api/companies/sync`

Request body from the extension:

```json
{
  "name": "Attio",
  "linkedin_url": "https://www.linkedin.com/company/attio",
  "website_url": "https://www.attio.com/",
  "domain": "attio.com",
  "description": "Customer relationship magic.",
  "extraction_debug": {
    "nameSource": "main h1",
    "websiteSource": "anchor website link",
    "descriptionSource": "meta description",
    "domainSource": "website_url",
    "warnings": []
  }
}
```

Validation:

- `name` is required.
- `linkedin_url` is required and must normalize to a LinkedIn company URL.
- `domain` is required for MVP sync because Attio company assertion uniquely matches on `domains`.
- `description`, `website_url`, and `extraction_debug` are optional.

Success response:

```json
{
  "status": "synced",
  "company": {
    "name": "Attio",
    "linkedin_url": "https://www.linkedin.com/company/attio",
    "domain": "attio.com",
    "attio_record_id": "record-id",
    "attio_web_url": "https://app.attio.com/...",
    "first_synced_at": "2026-04-20T10:00:00Z",
    "last_synced_at": "2026-04-20T10:00:00Z"
  }
}
```

Error response shape:

```json
{
  "error": "domain is required"
}
```

The backend uses HTTP 400 or 422 for validation errors, 401 for bad Attio credentials, 502 for Attio/network failures, and 500 only for unexpected backend errors.

## Attio Integration

The backend asserts company records with:

```text
PUT https://api.attio.com/v2/objects/companies/records?matching_attribute=domains
```

Request body:

```json
{
  "data": {
    "values": {
      "domains": ["attio.com"],
      "name": "Attio",
      "description": "Customer relationship magic.",
      "linkedin": "https://www.linkedin.com/company/attio"
    }
  }
}
```

The backend stores `data.id.record_id` and `data.web_url` from Attio.

## Local Development

1. Copy `.env.example` to `.env`.
2. Set `ATTIO_API_TOKEN`.
3. Run `docker compose up --build`.
4. Build and load the Chrome extension from the `extension/` folder.
5. Visit a LinkedIn company page.

Use this health check:

```sh
curl http://localhost:8080/healthz
```

## Tests

From the `backend/` directory:

```sh
go test ./...
```

The committed tests cover handlers, normalization, and the Attio client. Repository tests against PostgreSQL are the next useful backend test layer.
