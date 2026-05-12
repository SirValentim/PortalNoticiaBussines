(() => {
  function request(el) {
    const url = el.getAttribute("hx-get");
    const targetSelector = el.getAttribute("hx-target");
    const swap = el.getAttribute("hx-swap") || "innerHTML";
    if (!url || !targetSelector) return;
    const target = document.querySelector(targetSelector);
    if (!target) return;
    el.setAttribute("aria-busy", "true");
    el.disabled = true;
    fetch(url, { headers: { "HX-Request": "true" } })
      .then((response) => {
        if (!response.ok) throw new Error("request failed");
        return response.text();
      })
      .then((html) => {
        if (swap === "outerHTML") {
          target.outerHTML = html;
        } else if (swap === "beforeend") {
          target.insertAdjacentHTML("beforeend", html);
        } else {
          target.innerHTML = html;
        }
      })
      .catch(() => {
        el.disabled = false;
        el.removeAttribute("aria-busy");
        el.textContent = "Tentar novamente";
      });
  }

  document.addEventListener("click", (event) => {
    const el = event.target.closest("[hx-get]");
    if (!el) return;
    event.preventDefault();
    request(el);
  });
})();
