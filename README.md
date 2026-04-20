# Hublead LinkedIn to Attio Chrome Extension

This repo contains the Chrome extension side of the Hublead junior assessment. The backend is intentionally not implemented here; see [BACKEND.md](./BACKEND.md) for the backend handoff plan.

## What It Does

- Detects LinkedIn company pages.
- Injects a compact Hublead panel into the page.
- Extracts best-effort company details from the LinkedIn DOM.
- Looks up whether the company has already been synced through the backend.
- Lets the user ask the backend to sync the company to Attio.

The extension never talks to Attio directly and never stores Attio credentials.

## Local Setup

```bash
npm install
npm run check
```

To load the extension:

1. Run `npm run extension:build`.
2. Open `chrome://extensions`.
3. Enable Developer Mode.
4. Click **Load unpacked**.
5. Select the `extension/` folder.

The backend URL defaults to `http://localhost:8080`. You can change it from the Hublead panel on a LinkedIn company page.

## Expected Backend

The extension expects:

- `GET /api/companies/lookup?linkedin_url=...&domain=...&name=...`
- `POST /api/companies/sync`

Full backend details are in [BACKEND.md](./BACKEND.md).
