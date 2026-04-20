import { createRoot } from "react-dom/client";
import { HubleadPanel } from "./components/HubleadPanel";
import { extractLinkedInCompany } from "./extractLinkedInCompany";
import styles from "./styles.css?inline";

const HOST_ID = "hublead-attio-bridge-root";

function mountPanel() {
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

  const company = extractLinkedInCompany();
  createRoot(mount).render(<HubleadPanel company={company} />);
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", mountPanel, { once: true });
} else {
  mountPanel();
}
