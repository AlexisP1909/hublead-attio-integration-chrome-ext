# Tradeoffs

The two goals are to ship something that works end to end and learn a lot while building it.

## API Key Instead of OAuth

Decision: use one backend-held Attio API key for the MVP.

Why: this project targets a single Attio workspace, so an API key matches the immediate use case and keeps the first version focused on LinkedIn extraction, the backend contract, persistence, and Attio record sync.

Tradeoff: OAuth is the better production path for many workspaces or many users. Deferring OAuth means the MVP is not multi-tenant, but it avoids spending the assessment window on account and session infrastructure.

## Chrome Extension First, Backend Documented Separately

Decision: implement the Chrome integration now and document the backend contract separately.

Why: the extension can be built and tested against a stable API contract before the backend exists. This separates responsibilities clearly.

Tradeoff: the extension will show network errors until the backend is running. That is acceptable because the UI handles those failures explicitly.

## React + TypeScript Instead of Vanilla JavaScript

Decision: use React, TypeScript, and Vite.

Why: the panel has enough state to benefit from components and typed API contracts. TypeScript also makes the extension/backend boundary easier to understand.

Tradeoff: this adds build tooling and dependencies. A vanilla content script would be smaller, but would teach less about typed frontend integration.

## Injected Page Panel Instead of Popup-Only UX

Decision: inject a compact panel directly into LinkedIn company pages.

Why: the assessment asks the extension to recognize companies when the user lands on a LinkedIn company page. A visible page panel makes that flow obvious.

Tradeoff: injected UI must avoid conflicts with LinkedIn CSS and layout. Shadow DOM reduces that risk.

## Local Backend Database as Recognition Source

Decision: the extension asks the backend database whether a company is known.

Why: local lookup is fast, debuggable, and avoids querying Attio on every page load.

Tradeoff: if someone changes or deletes the Attio record outside this integration, the local cache may be stale until the company is resynced.

## Require Domain for Attio Company Sync

Decision: sync requires a domain.

Why: Attio company assertion uniquely matches by `domains`, so domain-based matching prevents accidental duplicates.

Tradeoff: some LinkedIn pages may not expose a website or domain clearly. The extension surfaces this as a review-needed state instead of guessing.

## Best-Effort LinkedIn DOM Extraction

Decision: extract from common DOM and metadata selectors and include debug metadata.

Why: LinkedIn markup changes often, so the extractor should be understandable and easy to improve.

Tradeoff: extraction may miss fields on some pages. Debug metadata makes those misses easier to learn from and fix.
