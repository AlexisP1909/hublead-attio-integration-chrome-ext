(() => {
  const hostId = "hublead-placeholder-test-root";

  if (document.getElementById(hostId)) {
    return;
  }

  const host = document.createElement("div");
  host.id = hostId;
  document.documentElement.append(host);

  const shadowRoot = host.attachShadow({ mode: "open" });
  shadowRoot.innerHTML = `
    <style>
      :host {
        all: initial;
        color-scheme: light;
        font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      }

      * {
        box-sizing: border-box;
      }

      .placeholder {
        position: fixed;
        right: 20px;
        top: 92px;
        z-index: 2147483647;
        width: min(360px, calc(100vw - 32px));
        background: #ffffff;
        border: 2px dashed #4f3bd9;
        border-radius: 8px;
        box-shadow: 0 18px 48px rgba(24, 31, 54, 0.18);
        color: #17172e;
        display: grid;
        gap: 8px;
        padding: 16px;
      }

      .eyebrow {
        color: #5b46e8;
        font-size: 11px;
        font-weight: 700;
        letter-spacing: 0;
        text-transform: uppercase;
      }

      .title {
        font-size: 17px;
        font-weight: 750;
        line-height: 1.3;
      }

      .body {
        color: #4b5470;
        font-size: 13px;
        line-height: 1.45;
      }

      @media (max-width: 720px) {
        .placeholder {
          bottom: 12px;
          right: 12px;
          top: auto;
          width: calc(100vw - 24px);
        }
      }
    </style>
    <aside class="placeholder" aria-label="Hublead extension placeholder">
      <div class="eyebrow">Hublead test extension</div>
      <div class="title">Actual extension panel should appear here</div>
      <div class="body">This placeholder uses the same fixed position as the real Hublead panel.</div>
    </aside>
  `;
})();
