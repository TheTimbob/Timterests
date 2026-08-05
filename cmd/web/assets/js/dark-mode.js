// Get stored theme or system preference
function getStoredTheme() {
  return localStorage.getItem("theme") ||
    (window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light");
}

// Apply the theme before first paint. This must be able to remove the class as
// well as add it, or a light-mode visitor renders dark until the DOM is ready.
function applyStoredTheme() {
  document.documentElement.classList.toggle("dark", getStoredTheme() === "dark");
}

applyStoredTheme();

// Toggle theme function
function toggleTheme() {
  const isDark = document.documentElement.classList.toggle("dark");
  localStorage.setItem("theme", isDark ? "dark" : "light");
}

// Initialize theme switch
function initThemeSwitch() {
  const darkModeSwitch = document.getElementById("dark-mode-switch");
  if (!darkModeSwitch) return;

  applyStoredTheme();

  // Add event listener (remove first to prevent duplicates)
  darkModeSwitch.removeEventListener("change", toggleTheme);
  darkModeSwitch.addEventListener("change", toggleTheme);
}

// Initialize on page load and navigation
document.addEventListener("DOMContentLoaded", initThemeSwitch);
window.addEventListener("pageshow", initThemeSwitch);
document.addEventListener("htmx:afterSwap", initThemeSwitch);
