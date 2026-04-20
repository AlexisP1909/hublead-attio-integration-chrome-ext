export type ExtractionDebug = {
  nameSource?: string;
  websiteSource?: string;
  descriptionSource?: string;
  domainSource?: string;
  warnings: string[];
};

export type ExtractedCompany = {
  name: string;
  linkedin_url: string;
  website_url?: string;
  domain?: string;
  description?: string;
  extraction_debug: ExtractionDebug;
};

export type CompanySummary = {
  name: string;
  linkedin_url: string;
  domain?: string;
  attio_record_id: string;
  attio_web_url: string;
  first_synced_at?: string;
  last_synced_at: string;
};

export type LookupResponse =
  | {
      status: "found";
      company: CompanySummary;
    }
  | {
      status: "not_found";
    };

export type SyncResponse = {
  status: "synced";
  company: CompanySummary;
};

export type ApiErrorKind = "validation" | "not_found" | "network" | "server" | "unknown";

export class ApiError extends Error {
  readonly kind: ApiErrorKind;
  readonly status?: number;

  constructor(message: string, kind: ApiErrorKind, status?: number) {
    super(message);
    this.name = "ApiError";
    this.kind = kind;
    this.status = status;
  }
}
