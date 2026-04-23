# Tradeoffs

The two goals are to ship something that works end to end and learn a lot while building it.

## API Key Instead of OAuth

Decision: use one backend-held Attio API key for the MVP.

Why: this project targets a single Attio workspace, so an API key matches the immediate use case and keeps the first version focused on LinkedIn extraction, the backend contract, persistence, and Attio record sync.

Tradeoff: OAuth is the better production path for many workspaces or many users. Deferring OAuth means the MVP is not multi-tenant, but it avoids spending the assessment window on account and session infrastructure.

## Chrome Extension and Backend in One Repo

Decision: keep the Chrome extension and Go backend in one repository, but keep their responsibilities separate.

Why: the assessment can be run end to end with one checkout and `docker compose up --build`, while the extension still never owns Attio credentials or calls Attio directly.

Tradeoff: this repo is broader than an extension-only handoff. The boundary is kept through `extension/`, `backend/`, and the HTTP API contract rather than by separate repositories.

## React + TypeScript Instead of Vanilla JavaScript

Decision: use React, TypeScript, and Vite.

Why: the panel has enough state to benefit from components and typed API contracts. TypeScript also makes the extension/backend boundary easier to understand.

Tradeoff: this adds build tooling and dependencies. A vanilla content script would be smaller, but would teach less about typed frontend integration.

## Injected Page Panel Instead of Popup-Only UX

Decision: inject a compact panel directly into LinkedIn company pages.

Why: the assessment asks the extension to recognize companies when the user lands on a LinkedIn company page. A visible page panel makes that flow obvious.

Tradeoff: injected UI must avoid conflicts with LinkedIn CSS and layout. Shadow DOM reduces that risk.

## Service Worker Backend Proxy

Decision: route backend requests through the Manifest V3 service worker instead of calling the local API directly from the content script.

Why: content scripts can still run into page/CORS constraints and are harder to distinguish from LinkedIn page noise. The service worker has clearer extension privileges for host-permissioned requests, and the backend logs include an `X-Hublead-Client: chrome-extension` marker.

Tradeoff: this adds message-passing complexity and another debug surface. Failures now need to be checked in both the LinkedIn page console and the extension service worker console.

## Local Backend Database as Recognition Source

Decision: the extension asks the backend database whether a company is known.

Why: local lookup is fast, debuggable, and avoids querying Attio on every page load.

Tradeoff: if someone changes or deletes the Attio record outside this integration, the local cache may be stale until the company is resynced.

## Require Domain for Attio Company Sync

Decision: sync requires a domain.

Why: Attio company assertion uniquely matches by `domains`, so domain-based matching prevents accidental duplicates.

Tradeoff: some LinkedIn pages do not expose a website or domain on the main company page. The extension routes the user to LinkedIn's About tab when the domain is missing instead of guessing.

## Best-Effort LinkedIn DOM Extraction

Decision: extract from common DOM and metadata selectors, retry after LinkedIn renders, watch SPA navigation, and include debug metadata.

Why: LinkedIn markup changes often, so the extractor should be understandable and easy to improve.

Tradeoff: extraction may still miss fields on some pages. Debug metadata and route-aware re-extraction make those misses easier to learn from and fix.

## About Tab for Website Detection

Decision: use the LinkedIn About tab as the preferred source for the company website/domain.

Why: LinkedIn usually displays the website in the About page's `Website` field, and that is the most useful source for deriving the domain required by Attio.

Tradeoff: the user may need one extra click when they land on the main company overview page. The main action becomes `Open About tab` when the domain is missing, then switches back to sync once the domain is detected.

## Docker Startup Over External Migration Image

Decision: run idempotent database setup from the backend on startup instead of using a separate Goose container image.

Why: pulling `ghcr.io/pressly/goose:latest` can fail in restricted environments. Embedding the simple MVP migration keeps `docker compose up --build` reliable with only Postgres and the local backend image.

Tradeoff: this is less flexible than a full migration tool for long-term schema evolution. For the MVP's single-table schema, the simpler startup migration is acceptable.
