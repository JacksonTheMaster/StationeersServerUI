(function () {
    const modal = document.getElementById('tls-modal');
    const certInput = document.getElementById('tls-certificate-file');
    const keyInput = document.getElementById('tls-private-key-file');
    const validation = document.getElementById('tls-validation');
    const saveButton = document.getElementById('tls-save-button');
    if (!modal || !certInput || !keyInput) return;

    window.openTLSModal = function () {
        validation.textContent = '';
        modal.classList.add('show');
        modal.setAttribute('aria-hidden', 'false');
        document.body.classList.add('modal-open');
    };
    window.closeTLSModal = function () {
        if (saveButton.disabled) return;
        modal.classList.remove('show');
        modal.setAttribute('aria-hidden', 'true');
        document.body.classList.remove('modal-open');
    };
    window.saveTLSCertificate = async function () {
        if (!certInput.files[0] || !keyInput.files[0]) {
            validation.textContent = 'Select both a certificate and private key file.';
            return;
        }
        const body = new FormData();
        body.append('certificate', certInput.files[0]);
        body.append('privateKey', keyInput.files[0]);
        saveButton.disabled = true;
        validation.textContent = '';
        try {
            const response = await fetch('/api/v2/tls/certificate', { method: 'POST', body });
            const result = await response.json();
            if (!response.ok) throw new Error(result.message || 'Failed to save TLS certificate.');
            document.getElementById('tls-restarting').hidden = false;
            document.getElementById('tls-modal-actions').hidden = true;
            setTimeout(() => window.location.href = '/', 5000);
        } catch (error) {
            validation.textContent = error.message;
            saveButton.disabled = false;
        }
    };
    modal.addEventListener('click', event => { if (event.target === modal) window.closeTLSModal(); });
    document.addEventListener('keydown', event => { if (event.key === 'Escape' && modal.classList.contains('show')) window.closeTLSModal(); });
})();
