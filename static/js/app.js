// Behaviour shared by every page. Loaded after htmx (both deferred).

// A 422 is a form submitted with bad data: swap it in so the form shows its errors.
document.addEventListener("htmx:beforeSwap", (evt) => {
  if (evt.detail.xhr.status === 422) {
    evt.detail.shouldSwap = true;
    evt.detail.isError = false;
  }
});

// Clamped text ([data-clamp]) shows its "Show more" toggle ([data-clamp-toggle],
// aria-controls naming the text) only when the clamp actually cuts it off.
function updateClampToggles(root) {
  root.querySelectorAll("[data-clamp-toggle]").forEach((toggle) => {
    const text = document.getElementById(toggle.getAttribute("aria-controls"));
    if (!text || toggle.getAttribute("aria-expanded") === "true") return;
    toggle.hidden = text.scrollHeight <= text.clientHeight + 1;
  });
}

document.addEventListener("click", (evt) => {
  const toggle = evt.target.closest("[data-clamp-toggle]");
  if (toggle) {
    const text = document.getElementById(toggle.getAttribute("aria-controls"));
    const expanded = toggle.getAttribute("aria-expanded") !== "true";
    toggle.setAttribute("aria-expanded", String(expanded));
    toggle.textContent = expanded ? "Show less" : "Show more";
    text.classList.toggle("clamp", !expanded);
    text.classList.toggle("expanded", expanded);
  }

  // Close an open header menu on a click outside it.
  document.querySelectorAll("details[data-menu][open]").forEach((menu) => {
    if (!menu.contains(evt.target)) menu.open = false;
  });
});

document.addEventListener("keydown", (evt) => {
  if (evt.key !== "Escape") return;
  document.querySelectorAll("details[data-menu][open]").forEach((menu) => {
    menu.open = false;
    menu.querySelector("summary").focus();
  });
});

document.addEventListener("DOMContentLoaded", () => updateClampToggles(document));
document.addEventListener("htmx:afterSettle", () => updateClampToggles(document));
window.addEventListener("resize", () => updateClampToggles(document));
document.fonts.ready.then(() => updateClampToggles(document));
