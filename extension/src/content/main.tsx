import { createRoot } from "react-dom/client";
import type { Root } from "react-dom/client";
import { HubleadPanel } from "./components/HubleadPanel";
import { extractLinkedInCompany } from "./extractLinkedInCompany";
import styles from "./styles.css?inline";

const HOST_ID = "hublead-attio-bridge-root";
const EXTRACTION_RETRIES = 8;
const EXTRACTION_RETRY_DELAY_MS = 500;
const ROUTE_POLL_INTERVAL_MS = 500;

let root: Root | undefined;
let lastRenderedSignature = "";
let lastLocationHref = window.location.href;

async function mountPanel() {
  if (document.getElementById(HOST_ID)) {
    return;
  }

  const host = document.createElement("div");
  host.id = HOST_ID;
  document.documentElement.append(host);

  const shadowRoot = host.attachShadow({ mode: "open" });
  const style = document.createElement("style");
  style.textContent = styles;

  const mount = document.createElement("div");
  shadowRoot.append(style, mount);
  root = createRoot(mount);

  await renderExtractedCompany();
  watchLinkedInNavigation();
}

async function extractLinkedInCompanyWhenReady() {
  let company = extractLinkedInCompany();

  for (let attempt = 0; attempt < EXTRACTION_RETRIES && !company.domain; attempt += 1) {
    await delay(EXTRACTION_RETRY_DELAY_MS);
    company = extractLinkedInCompany();
  }

  return company;
}

async function renderExtractedCompany() {
  const company = await extractLinkedInCompanyWhenReady();
  const signature = JSON.stringify({
    href: window.location.href,
    name: company.name,
    domain: company.domain,
    website_url: company.website_url
  });

  if (signature === lastRenderedSignature) {
    return;
  }

  lastRenderedSignature = signature;
  console.debug("[Hublead] extracted LinkedIn company", company);
  root?.render(<HubleadPanel company={company} />);
}

function watchLinkedInNavigation() {
  window.setInterval(() => {
    if (window.location.href !== lastLocationHref) {
      lastLocationHref = window.location.href;
      void renderExtractedCompany();
    }
  }, ROUTE_POLL_INTERVAL_MS);

  const observer = new MutationObserver(() => {
    void renderExtractedCompany();
  });
  observer.observe(document.body, { childList: true, subtree: true });
}

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", () => void mountPanel(), { once: true });
} else {
  void mountPanel();
}
