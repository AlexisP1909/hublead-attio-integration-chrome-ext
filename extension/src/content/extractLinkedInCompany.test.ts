import { describe, expect, it } from "vitest";
import { extractLinkedInCompany, normalizeDomain, normalizeLinkedInCompanyUrl } from "./extractLinkedInCompany";

describe("extractLinkedInCompany", () => {
  it("extracts company details from a LinkedIn-like document", () => {
    document.body.innerHTML = `
      <main>
        <h1>Attio</h1>
        <a href="https://www.attio.com/">Website</a>
        <section><p>Customer relationship magic.</p></section>
      </main>
    `;

    const company = extractLinkedInCompany(document, "https://www.linkedin.com/company/attio/about/?viewAsMember=true");

    expect(company).toMatchObject({
      name: "Attio",
      linkedin_url: "https://www.linkedin.com/company/attio",
      website_url: "https://www.attio.com/",
      domain: "attio.com",
      description: "Customer relationship magic."
    });
  });

  it("falls back to the company slug when the page name is missing", () => {
    document.body.innerHTML = "";

    const company = extractLinkedInCompany(document, "https://www.linkedin.com/company/example-company/");

    expect(company.name).toBe("Example Company");
    expect(company.extraction_debug.warnings).toContain("Could not confidently extract a company name from the page.");
  });

  it("extracts website URLs from LinkedIn redirect links", () => {
    document.body.innerHTML = `
      <main>
        <h1>HappyRobot</h1>
        <a href="https://www.linkedin.com/redir/redirect?url=https%3A%2F%2Fwww.happyrobot.ai%2F" aria-label="Website">Visit</a>
      </main>
    `;

    const company = extractLinkedInCompany(document, "https://www.linkedin.com/company/happyrobot/");

    expect(company.website_url).toBe("https://www.happyrobot.ai/");
    expect(company.domain).toBe("happyrobot.ai");
  });
});

describe("normalization helpers", () => {
  it("normalizes LinkedIn company URLs to canonical company URLs", () => {
    expect(normalizeLinkedInCompanyUrl("https://linkedin.com/company/hublead/posts/?foo=bar")).toBe(
      "https://www.linkedin.com/company/hublead"
    );
  });

  it("normalizes website URLs to root hostnames without www", () => {
    expect(normalizeDomain("https://www.example.com/pricing")).toBe("example.com");
    expect(normalizeDomain("example.org/about")).toBe("example.org");
  });
});
