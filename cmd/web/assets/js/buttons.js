function handleViewChange(button) {
    button.parentElement.querySelectorAll('.view-btn').forEach(e => e.classList.remove('active'));
    button.classList.add('active');
}

function setActiveTab(button) {
    document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
    button.classList.add('active');
}

function toggleSkill(header) {
    let item = header.parentElement;
    let content = header.nextElementSibling;

    item.classList.toggle("active");

    if (content.style.maxHeight){
        content.style.maxHeight = null;
    } else {
        content.style.maxHeight = content.scrollHeight + "px";
    }
}
