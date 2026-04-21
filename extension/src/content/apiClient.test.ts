import { afterEach, describe, expect, it, vi } from "vitest";
import type { ExtractedCompany } from "../shared/types";
import { lookupCompany, syncCompany } from "./apiClient";

const company: ExtractedCompany = {
  name: "Attio",
  linkedin_url: "https://www.linkedin.com/company/attio",
  domain: "attio.com",
  extraction_debug: { warnings: [] }
};

afterEach(() => {
  vi.restoreAllMocks();
});

describe("apiClient", () => {
  it("returns not_found on lookup 404", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: "not found" }), { status: 404 })));

    await expect(lookupCompany("http://localhost:8080", company)).resolves.toEqual({ status: "not_found" });
  });

  it("parses known company lookup responses", async () => {
    const payload = {
      status: "found",
      company: {
        name: "Attio",
        linkedin_url: company.linkedin_url,
        domain: "attio.com",
        attio_record_id: "rec_123",
        attio_web_url: "https://app.attio.com/record/rec_123",
        last_synced_at: "2026-04-20T10:00:00Z"
      }
    };
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify(payload), { status: 200 })));

    await expect(lookupCompany("http://localhost:8080", company)).resolves.toEqual(payload);
  });

  it("sends sync requests to the backend", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          status: "synced",
          company: {
            name: "Attio",
            linkedin_url: company.linkedin_url,
            domain: "attio.com",
            attio_record_id: "rec_123",
            attio_web_url: "https://app.attio.com/record/rec_123",
            last_synced_at: "2026-04-20T10:00:00Z"
          }
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    await syncCompany("http://localhost:8080", company);

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/companies/sync",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify(company)
      })
    );
  });

  it("throws validation errors with backend messages", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: "domain is required" }), { status: 422 })));

    await expect(syncCompany("http://localhost:8080", company)).rejects.toMatchObject({
      kind: "validation",
      message: "domain is required",
      status: 422
    });
  });

  it("throws auth errors when Attio credentials are rejected", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: "Attio credentials were rejected" }), { status: 401 })));

    await expect(syncCompany("http://localhost:8080", company)).rejects.toMatchObject({
      kind: "auth",
      message: "Attio credentials were rejected",
      status: 401
    });
  });

  it("throws network errors when the backend is unreachable", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("Failed to fetch")));

    await expect(lookupCompany("http://localhost:8080", company)).rejects.toMatchObject({ kind: "network" });
  });
});
