document.addEventListener("click", (event) => {
  const button = event.target.closest(".copy");
  if (!button) return;
  const block = button.closest("[data-copy]");
  const code = block && block.querySelector("code");
  if (!code) return;
  navigator.clipboard.writeText(code.innerText).then(() => {
    const original = button.textContent;
    button.textContent = "copied";
    setTimeout(() => { button.textContent = original; }, 1400);
  });
});
