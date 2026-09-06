function showPopup(status, message) {
    const popup = document.getElementById('universalPopup');
    const popupTitle = document.getElementById('popupTitle');
    const popupMessage = document.getElementById('popupMessage');
    
    popup.className = 'popup';
    popupTitle.textContent = '';
    popupMessage.innerHTML = message.replace(/\n/g, '<br>');

    switch(status.toLowerCase()) {
        case 'error':
            popup.classList.add('error');
            popupTitle.textContent = 'Error';
            break;
        case 'success':
            popup.classList.add('success');
            popupTitle.textContent = 'Success';
            break;
        case 'info':
            popup.classList.add('info');
            popupTitle.textContent = 'Info';
            break;
        default:
            popup.classList.add('info');
            popupTitle.textContent = 'Info';
    }

    popup.style.display = 'flex';
}

function closePopup() {
    const popup = document.getElementById('universalPopup');
    popup.style.display = 'none';
}