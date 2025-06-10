// Get stored theme or system preference
function getStoredTheme() {
  return localStorage.getItem("theme") ||
    (window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light");
}

// Apply theme immediately (runs before DOM is ready)
(function() {
  const theme = getStoredTheme();
  if (theme === "dark") {
    document.documentElement.classList.add("dark");
  }
})();

// Toggle theme function
function toggleTheme() {
  const html = document.documentElement;
  const darkModeSwitch = document.getElementById("dark-mode-switch");

  if (!darkModeSwitch.checked) {
    html.classList.add("dark");
    localStorage.setItem("theme", "dark");
  } else {
    html.classList.remove("dark");
    localStorage.setItem("theme", "light");
  }
}

// Initialize theme switch
function initThemeSwitch() {
  const darkModeSwitch = document.getElementById("dark-mode-switch");
  if (!darkModeSwitch) return;

  const storedTheme = getStoredTheme();

  // Apply theme
  document.documentElement.classList.toggle("dark", storedTheme === "dark");

  // Sync switch state
  darkModeSwitch.checked = storedTheme === "light";

  // Add event listener (remove first to prevent duplicates)
  darkModeSwitch.removeEventListener("change", toggleTheme);
  darkModeSwitch.addEventListener("change", toggleTheme);
}

// Initialize on page load and navigation
document.addEventListener("DOMContentLoaded", initThemeSwitch);
window.addEventListener("pageshow", initThemeSwitch);
document.addEventListener("htmx:afterSwap", initThemeSwitch);
