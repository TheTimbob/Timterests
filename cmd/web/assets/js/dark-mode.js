// JavaScript for toggling theme
function toggleTheme() {
    const html = document.documentElement;
    const darkModeSwitch = document.getElementById('dark-mode-switch');

    if (!darkModeSwitch.checked) {
        html.classList.add('dark');
        localStorage.setItem('theme', 'dark');
    } else {
        html.classList.remove('dark');
        localStorage.setItem('theme', 'light');
    }
}

// Add event listener for theme switch
document.addEventListener('DOMContentLoaded', function () {
    const darkModeSwitch = document.getElementById('dark-mode-switch');

    // Load stored theme or system preference on page load
    const storedTheme = localStorage.getItem('theme') || window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    document.documentElement.classList.toggle('dark', storedTheme === 'dark');

    //Sync switch
    darkModeSwitch.checked = storedTheme === 'light';

    // Add event listener to toggle theme
    darkModeSwitch.addEventListener('change', toggleTheme);
});
