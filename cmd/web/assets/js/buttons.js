function handleViewChange(button) {
    button.parentElement.querySelectorAll('.view-btn').forEach(e => e.classList.remove('active'));
    button.classList.add('active');
}
