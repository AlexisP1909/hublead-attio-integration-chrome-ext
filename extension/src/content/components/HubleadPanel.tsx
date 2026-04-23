import { useEffect, useMemo, useState } from "react";
import { ApiError, type CompanySummary, type ExtractedCompany } from "../../shared/types";
import {
  DEFAULT_BACKEND_URL,
  getBackendBaseUrl,
  lookupCompany,
  setBackendBaseUrl,
  syncCompany
} from "../apiClient";

type LookupState =
  | { status: "loading" }
  | { status: "found"; company: CompanySummary }
  | { status: "not_found" }
  | { status: "error"; message: string; kind: string };

type SyncState =
  | { status: "idle" }
  | { status: "syncing" }
  | { status: "synced"; company: CompanySummary }
  | { status: "error"; message: string; kind: string };

type Props = {
  company: ExtractedCompany;
};

export function HubleadPanel({ company }: Props) {
  const [backendUrl, setBackendUrlState] = useState(DEFAULT_BACKEND_URL);
  const [draftBackendUrl, setDraftBackendUrl] = useState(DEFAULT_BACKEND_URL);
  const [lookupState, setLookupState] = useState<LookupState>({ status: "loading" });
  const [syncState, setSyncState] = useState<SyncState>({ status: "idle" });

  const isMissingDomain = !company.domain;
  const canSync = Boolean(company.name && company.linkedin_url && company.domain);
  const aboutUrl = `${company.linkedin_url.replace(/\/+$/, "")}/about/`;
  const syncBlocker = useMemo(() => {
    if (isMissingDomain) {
      return "A website/domain is required. LinkedIn usually shows it on the About tab.";
    }
    if (!company.name) {
      return "A company name is required before syncing.";
    }
    return undefined;
  }, [company.name, isMissingDomain]);

  useEffect(() => {
    let cancelled = false;

    async function loadConfigAndLookup() {
      const baseUrl = await getBackendBaseUrl();
      if (cancelled) {
        return;
      }

      setBackendUrlState(baseUrl);
      setDraftBackendUrl(baseUrl);
      await runLookup(baseUrl);
    }

    void loadConfigAndLookup();

    return () => {
      cancelled = true;
    };
  }, [company.linkedin_url, company.domain, company.name]);

  async function runLookup(baseUrl = backendUrl) {
    setLookupState({ status: "loading" });
    try {
      const response = await lookupCompany(baseUrl, company);
      setLookupState(response.status === "found" ? { status: "found", company: response.company } : { status: "not_found" });
    } catch (error) {
      setLookupState(errorToPanelState(error));
    }
  }

  async function handleSync() {
    if (isMissingDomain) {
      window.location.assign(aboutUrl);
      return;
    }

    setSyncState({ status: "syncing" });
    try {
      const response = await syncCompany(backendUrl, company);
      setSyncState({ status: "synced", company: response.company });
      setLookupState({ status: "found", company: response.company });
    } catch (error) {
      setSyncState(errorToPanelState(error));
    }
  }

  async function handleSaveBackendUrl() {
    const saved = await setBackendBaseUrl(draftBackendUrl);
    setBackendUrlState(saved);
    await runLookup(saved);
  }

  const visibleCompany = syncState.status === "synced" ? syncState.company : lookupState.status === "found" ? lookupState.company : undefined;

  return (
    <aside className="hublead-panel" aria-label="Hublead Attio sync panel">
      <header className="hublead-header">
        <div>
          <p className="hublead-eyebrow">Hublead</p>
          <h2>Attio bridge</h2>
        </div>
        <button className="hublead-icon-button" type="button" title="Refresh lookup" onClick={() => void runLookup()}>
          &#8635;
        </button>
      </header>

      <section className="hublead-section">
        <dl className="hublead-facts">
          <div>
            <dt>Name</dt>
            <dd>{company.name}</dd>
          </div>
          <div>
            <dt>Domain</dt>
            <dd>{company.domain ?? "Not detected"}</dd>
          </div>
        </dl>
      </section>

      <section className="hublead-section">
        {lookupState.status === "loading" && <p className="hublead-muted">Checking backend...</p>}

        {lookupState.status === "error" && (
          <StatusMessage tone="warning" title="Backend unavailable" message={lookupState.message} />
        )}

        {visibleCompany && (
          <div className="hublead-record">
            <p className="hublead-status-text">Known company</p>
            <a href={visibleCompany.attio_web_url} target="_blank" rel="noreferrer">
              Open Attio record
            </a>
            <p className="hublead-muted">Last sync: {formatDate(visibleCompany.last_synced_at)}</p>
          </div>
        )}

        {lookupState.status === "not_found" && !visibleCompany && (
          <p className="hublead-muted">This company is not synced yet.</p>
        )}

        {syncState.status === "error" && <StatusMessage tone="danger" title="Sync failed" message={syncState.message} />}

        {syncBlocker && <StatusMessage tone="warning" title="Needs review" message={syncBlocker} />}

        <button
          className="hublead-primary"
          type="button"
          disabled={(!canSync && !isMissingDomain) || syncState.status === "syncing"}
          onClick={() => void handleSync()}
        >
          {isMissingDomain ? "Open About tab" : syncState.status === "syncing" ? "Syncing..." : visibleCompany ? "Resync company" : "Add to Attio"}
        </button>
      </section>

      <details className="hublead-details">
        <summary>Settings and debug</summary>
        <label className="hublead-label">
          Backend URL
          <span className="hublead-url-row">
            <input value={draftBackendUrl} onChange={(event) => setDraftBackendUrl(event.target.value)} />
            <button type="button" onClick={() => void handleSaveBackendUrl()}>
              Save
            </button>
          </span>
        </label>
        <pre>{JSON.stringify(company.extraction_debug, null, 2)}</pre>
      </details>
    </aside>
  );
}

function StatusMessage({ tone, title, message }: { tone: "warning" | "danger"; title: string; message: string }) {
  return (
    <div className={`hublead-message hublead-message-${tone}`}>
      <strong>{title}</strong>
      <span>{message}</span>
    </div>
  );
}

function errorToPanelState(error: unknown): { status: "error"; message: string; kind: string } {
  if (error instanceof ApiError) {
    return { status: "error", message: error.message, kind: error.kind };
  }
  if (error instanceof Error) {
    return { status: "error", message: error.message, kind: "unknown" };
  }
  return { status: "error", message: "Unexpected error.", kind: "unknown" };
}

function formatDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short"
  }).format(date);
}
