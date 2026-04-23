import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import manifest from "../manifest.json";

type ManifestLike = {
  manifest_version?: number;
  background?: {
    service_worker?: string;
    scripts?: string[];
    page?: string;
    persistent?: boolean;
  };
  browser_action?: unknown;
  page_action?: unknown;
  content_security_policy?: string | { extension_pages?: string; sandbox?: string };
  content_scripts?: Array<{ js?: string[]; matches?: string[]; run_at?: string }>;
  host_permissions?: string[];
  permissions?: string[];
};

const mv2OnlyKeys = ["browser_action", "page_action"] as const;
const forbiddenCspTokens = ["'unsafe-eval'", "wasm-unsafe-eval"] as const;

describe("extension manifest", () => {
  const chromeManifest = manifest as ManifestLike;

  it("uses Manifest V3", () => {
    expect(chromeManifest.manifest_version).toBe(3);
  });

  it("does not use Manifest V2 background or action fields", () => {
    for (const key of mv2OnlyKeys) {
      expect(chromeManifest[key]).toBeUndefined();
    }

    expect(chromeManifest.background?.scripts).toBeUndefined();
    expect(chromeManifest.background?.page).toBeUndefined();
    expect(chromeManifest.background?.persistent).toBeUndefined();

    if (chromeManifest.background) {
      expect(chromeManifest.background.service_worker).toMatch(/\.js$/);
    }
  });

  it("uses a Manifest V3 service worker for backend requests", () => {
    expect(chromeManifest.background?.service_worker).toBe("dist/background.js");
  });

  it("does not relax extension page CSP for dynamic code execution", () => {
    const csp =
      typeof chromeManifest.content_security_policy === "string"
        ? chromeManifest.content_security_policy
        : chromeManifest.content_security_policy?.extension_pages;

    for (const token of forbiddenCspTokens) {
      expect(csp ?? "").not.toContain(token);
    }
  });

  it("loads local bundled content scripts on LinkedIn company pages only", () => {
    expect(chromeManifest.content_scripts).toEqual([
      expect.objectContaining({
        matches: ["https://www.linkedin.com/company/*", "https://linkedin.com/company/*"],
        js: ["dist/content.js"],
        run_at: "document_idle"
      })
    ]);

    for (const script of chromeManifest.content_scripts ?? []) {
      for (const file of script.js ?? []) {
        expect(file).not.toMatch(/^https?:\/\//);
      }
    }
  });

  it("keeps permissions explicit and MV3-compatible", () => {
    expect(chromeManifest.permissions).toEqual(["storage"]);
    expect(chromeManifest.host_permissions).toEqual([
      "http://host.docker.internal/*",
      "http://localhost/*",
      "http://127.0.0.1/*",
      "https://www.linkedin.com/company/*",
      "https://linkedin.com/company/*"
    ]);
  });

  it("inlines Node environment checks for browser content scripts", () => {
    const config = readFileSync(resolve(__dirname, "../vite.config.ts"), "utf8");

    expect(config).toContain('"process.env.NODE_ENV": JSON.stringify("production")');
  });
});
