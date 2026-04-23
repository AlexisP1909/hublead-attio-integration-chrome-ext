chrome.runtime.onMessage.addListener((message, _sender, sendResponse) => {
  if (message?.type !== "hubleadApiRequest") {
    return false;
  }

  console.debug("[Hublead background] backend request", message.method, message.url);

  fetch(message.url, {
    method: message.method,
    headers: message.headers,
    body: message.body
  })
    .then(async (response) => {
      sendResponse({
        ok: true,
        status: response.status,
        statusText: response.statusText,
        body: await response.text()
      });
    })
    .catch((error) => {
      console.error("[Hublead background] backend request failed", error);
      sendResponse({
        ok: false,
        error: error instanceof Error ? error.message : "Unable to reach backend."
      });
    });

  return true;
});
