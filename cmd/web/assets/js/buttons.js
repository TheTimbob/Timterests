function handleViewChange(button) {
    button.parentElement.querySelectorAll('.view-btn').forEach(e => e.classList.remove('active'));
    button.classList.add('active');
}

function setActiveTab(button) {
    document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
    button.classList.add('active');
}

document.body.addEventListener('htmx:afterSwap', function (evt) {
    if (evt.detail.target.id === 'main-content') {
        var toggle = document.getElementById('nav-toggle');
        if (toggle) {
            toggle.checked = false;
        }
    }
});
