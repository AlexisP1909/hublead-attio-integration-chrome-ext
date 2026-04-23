import type { ExtractedCompany, ExtractionDebug } from "../shared/types";

type Candidate = {
  value: string;
  source: string;
};

const LINKEDIN_COMPANY_PATH = /^\/company\/([^/?#]+)/;

export function extractLinkedInCompany(doc: Document = document, locationHref = window.location.href): ExtractedCompany {
  const debug: ExtractionDebug = { warnings: [] };
  const linkedinUrl = normalizeLinkedInCompanyUrl(locationHref);

  const nameCandidate = firstTextCandidate(doc, [
    { selector: "main h1", source: "main h1" },
    { selector: "h1", source: "h1" },
    { selector: "[data-test-org-name]", source: "[data-test-org-name]" },
    { selector: "meta[property='og:title']", source: "meta og:title", attr: "content" },
    { selector: "title", source: "document title" }
  ]);

  const name = cleanCompanyName(nameCandidate?.value ?? "");
  debug.nameSource = nameCandidate?.source;

  const websiteCandidate = firstWebsiteCandidate(doc);
  const websiteUrl = websiteCandidate?.value;
  const domain = websiteUrl ? normalizeDomain(websiteUrl) : undefined;
  debug.websiteSource = websiteCandidate?.source;
  debug.domainSource = domain ? "website_url" : undefined;

  const descriptionCandidate = firstTextCandidate(doc, [
    { selector: "meta[property='og:description']", source: "meta og:description", attr: "content" },
    { selector: "meta[name='description']", source: "meta description", attr: "content" },
    { selector: "section p", source: "section p" }
  ]);
  const description = descriptionCandidate ? cleanDescription(descriptionCandidate.value) : undefined;
  debug.descriptionSource = descriptionCandidate?.source;

  if (!name) {
    debug.warnings.push("Could not confidently extract a company name from the page.");
  }

  if (!domain) {
    debug.warnings.push("Could not extract a company website/domain; backend sync will likely require manual correction.");
  }

  return {
    name: name || fallbackNameFromUrl(linkedinUrl),
    linkedin_url: linkedinUrl,
    ...(websiteUrl ? { website_url: websiteUrl } : {}),
    ...(domain ? { domain } : {}),
    ...(description ? { description } : {}),
    extraction_debug: debug
  };
}

export function normalizeLinkedInCompanyUrl(rawUrl: string): string {
  const url = new URL(rawUrl);
  const match = url.pathname.match(LINKEDIN_COMPANY_PATH);
  const slug = match?.[1] ?? "";
  return `https://www.linkedin.com/company/${slug}`;
}

export function normalizeDomain(rawUrlOrDomain: string): string | undefined {
  const trimmed = rawUrlOrDomain.trim();
  if (!trimmed) {
    return undefined;
  }

  try {
    const withProtocol = /^[a-z][a-z0-9+.-]*:\/\//i.test(trimmed) ? trimmed : `https://${trimmed}`;
    const url = new URL(withProtocol);
    return url.hostname.replace(/^www\./i, "").toLowerCase();
  } catch {
    return undefined;
  }
}

function firstTextCandidate(
  doc: Document,
  selectors: Array<{ selector: string; source: string; attr?: string }>
): Candidate | undefined {
  for (const item of selectors) {
    const element = doc.querySelector(item.selector);
    if (!element) {
      continue;
    }

    const rawValue = item.attr ? element.getAttribute(item.attr) : element.textContent;
    const value = normalizeWhitespace(rawValue ?? "");
    if (value) {
      return { value, source: item.source };
    }
  }

  return undefined;
}

function firstWebsiteCandidate(doc: Document): Candidate | undefined {
  const anchors = Array.from(doc.querySelectorAll<HTMLAnchorElement>("a[href]"));
  const websiteAnchor = anchors.find((anchor) => {
    const text = normalizeWhitespace(anchor.textContent ?? "").toLowerCase();
    const href = anchor.href;
    const ariaLabel = normalizeWhitespace(anchor.getAttribute("aria-label") ?? "").toLowerCase();
    const dataControlName = anchor.getAttribute("data-control-name")?.toLowerCase() ?? "";
    const dataTestId = anchor.getAttribute("data-test-id")?.toLowerCase() ?? "";

    return (
      text.includes("website") ||
      ariaLabel.includes("website") ||
      dataControlName.includes("website") ||
      dataTestId.includes("website") ||
      href.includes("/redir/redirect") ||
      (/^https?:\/\//.test(href) && !href.includes("linkedin.com"))
    );
  });

  if (!websiteAnchor) {
    return undefined;
  }

  const cleaned = cleanWebsiteUrl(websiteAnchor.href);
  return cleaned ? { value: cleaned, source: "anchor website link" } : undefined;
}

function cleanWebsiteUrl(rawUrl: string): string | undefined {
  try {
    const url = new URL(rawUrl);
    const nestedUrl = url.searchParams.get("url");
    if (url.hostname.includes("linkedin.com") && nestedUrl) {
      return new URL(nestedUrl).toString();
    }
    return url.toString();
  } catch {
    return undefined;
  }
}

function cleanCompanyName(rawName: string): string {
  return normalizeWhitespace(rawName)
    .replace(/\s*\|\s*LinkedIn$/i, "")
    .replace(/\s*:\s*Overview$/i, "")
    .trim();
}

function cleanDescription(rawDescription: string): string | undefined {
  const description = normalizeWhitespace(rawDescription);
  return description || undefined;
}

function fallbackNameFromUrl(linkedinUrl: string): string {
  const slug = new URL(linkedinUrl).pathname.split("/").filter(Boolean).at(1) ?? "Unknown company";
  return slug
    .split("-")
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function normalizeWhitespace(value: string): string {
  return value.replace(/\s+/g, " ").trim();
}
