import type { ExtractedCompany, LookupResponse, SyncResponse } from "../shared/types";
import { ApiError } from "../shared/types";

export const DEFAULT_BACKEND_URL = "http://localhost:8080";
const BACKEND_URL_STORAGE_KEY = "hubleadBackendUrl";

export async function getBackendBaseUrl(): Promise<string> {
  const storage = getChromeSyncStorage();
  if (!storage) {
    return DEFAULT_BACKEND_URL;
  }

  const result = await storage.get(BACKEND_URL_STORAGE_KEY);
  const stored = result[BACKEND_URL_STORAGE_KEY];
  return typeof stored === "string" && stored.trim() ? trimTrailingSlash(stored) : DEFAULT_BACKEND_URL;
}

export async function setBackendBaseUrl(url: string): Promise<string> {
  const normalized = trimTrailingSlash(url.trim() || DEFAULT_BACKEND_URL);
  const storage = getChromeSyncStorage();
  if (storage) {
    await storage.set({ [BACKEND_URL_STORAGE_KEY]: normalized });
  }
  return normalized;
}

export async function lookupCompany(baseUrl: string, company: ExtractedCompany): Promise<LookupResponse> {
  const url = new URL(`${trimTrailingSlash(baseUrl)}/api/companies/lookup`);
  url.searchParams.set("linkedin_url", company.linkedin_url);
  if (company.domain) {
    url.searchParams.set("domain", company.domain);
  }
  if (company.name) {
    url.searchParams.set("name", company.name);
  }

  const response = await request(url, { method: "GET" });
  if (response.status === 404) {
    return { status: "not_found" };
  }

  return parseJsonResponse<LookupResponse>(response);
}

export async function syncCompany(baseUrl: string, company: ExtractedCompany): Promise<SyncResponse> {
  const response = await request(`${trimTrailingSlash(baseUrl)}/api/companies/sync`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(company)
  });

  return parseJsonResponse<SyncResponse>(response);
}

async function request(input: RequestInfo | URL, init: RequestInit): Promise<Response> {
  try {
    return await fetch(input, init);
  } catch (error) {
    throw new ApiError(error instanceof Error ? error.message : "Unable to reach backend.", "network");
  }
}

async function parseJsonResponse<T>(response: Response): Promise<T> {
  const payload = await readJsonBody(response);

  if (!response.ok) {
    const message = errorMessageFromPayload(payload) ?? `Backend returned HTTP ${response.status}.`;
    if (response.status === 400 || response.status === 422) {
      throw new ApiError(message, "validation", response.status);
    }
    if (response.status === 404) {
      throw new ApiError(message, "not_found", response.status);
    }
    if (response.status >= 500) {
      throw new ApiError(message, "server", response.status);
    }
    throw new ApiError(message, "unknown", response.status);
  }

  return payload as T;
}

async function readJsonBody(response: Response): Promise<unknown> {
  const text = await response.text();
  if (!text) {
    return {};
  }

  try {
    return JSON.parse(text);
  } catch {
    throw new ApiError("Backend returned invalid JSON.", "server", response.status);
  }
}

function errorMessageFromPayload(payload: unknown): string | undefined {
  if (!payload || typeof payload !== "object") {
    return undefined;
  }

  const record = payload as Record<string, unknown>;
  if (typeof record.error === "string") {
    return record.error;
  }
  if (typeof record.message === "string") {
    return record.message;
  }

  return undefined;
}

function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, "");
}

function getChromeSyncStorage(): chrome.storage.StorageArea | undefined {
  return typeof chrome !== "undefined" ? chrome.storage?.sync : undefined;
}
