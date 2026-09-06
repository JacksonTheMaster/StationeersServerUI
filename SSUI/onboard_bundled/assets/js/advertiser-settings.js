(function () {
    const modal = document.getElementById('advertiser-modal');
    const addressInput = document.getElementById('advertiserAddress');
    const addressField = document.getElementById('advertiser-address-field');
    const currentValue = document.getElementById('advertiser-current-value');
    if (!modal || !addressInput || !addressField || !currentValue) return;

    const ipv4Pattern = /^(?:25[0-5]|2[0-4]\d|1?\d?\d)(?:\.(?:25[0-5]|2[0-4]\d|1?\d?\d)){3}$/;

    function modeForValue(value) {
        if (!value) return 'disabled';
        if (value.toLowerCase() === 'auto') return 'auto';
        return ipv4Pattern.test(value) ? 'ipv4' : 'dns';
    }

    function selectedMode() {
        return document.querySelector('input[name="advertiserMode"]:checked')?.value || 'disabled';
    }

    function updateAddressField() {
        const mode = selectedMode();
        const needsAddress = mode === 'ipv4' || mode === 'dns';
        addressField.classList.toggle('is-hidden', !needsAddress);
        addressInput.disabled = !needsAddress;
        addressInput.placeholder = mode === 'dns' ? 'server.example.com' : '203.0.113.10';
        document.getElementById('advertiser-validation').textContent = '';
    }

    function updateCurrentValue() {
        const value = currentValue.dataset.value.trim();
        const mode = modeForValue(value);
        const label = currentValue.dataset[mode];
        currentValue.textContent = value && mode !== 'auto' ? `${label}: ${value}` : label;
        currentValue.classList.toggle('is-active', mode !== 'disabled');
    }

    window.openAdvertiserModal = function () {
        const value = currentValue.dataset.value.trim();
        const mode = modeForValue(value);
        const radio = document.querySelector(`input[name="advertiserMode"][value="${mode}"]`);
        if (radio) radio.checked = true;
        addressInput.value = mode === 'ipv4' || mode === 'dns' ? value : '';
        updateAddressField();
        modal.classList.add('show');
        modal.setAttribute('aria-hidden', 'false');
        document.body.classList.add('modal-open');
    };

    window.closeAdvertiserModal = function () {
        if (document.getElementById('advertiser-save-button').disabled) return;
        modal.classList.remove('show');
        modal.setAttribute('aria-hidden', 'true');
        document.body.classList.remove('modal-open');
    };

    window.saveAdvertiserOverride = async function () {
        const mode = selectedMode();
        const validation = document.getElementById('advertiser-validation');
        const saveButton = document.getElementById('advertiser-save-button');
        const value = addressInput.value.trim();

        if (mode === 'ipv4' && !ipv4Pattern.test(value)) {
            validation.textContent = 'Enter a valid IPv4 address.';
            return;
        }
        if (mode === 'dns' && !value) {
            validation.textContent = 'Enter a DNS hostname.';
            return;
        }

        saveButton.disabled = true;
        validation.textContent = '';
        try {
            const response = await fetch('/api/v2/advertiser/override', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ mode, value })
            });
            const result = await response.json();
            if (!response.ok) throw new Error(result.message || 'Failed to save advertiser configuration.');

            document.getElementById('advertiser-restarting').hidden = false;
            document.querySelector('.advertiser-modal-actions').hidden = true;
            setTimeout(() => window.location.href = '/', 4000);
        } catch (error) {
            validation.textContent = error.message;
            saveButton.disabled = false;
        }
    };

    document.querySelectorAll('input[name="advertiserMode"]').forEach(radio => {
        radio.addEventListener('change', updateAddressField);
    });
    modal.addEventListener('click', event => {
        if (event.target === modal) window.closeAdvertiserModal();
    });
    document.addEventListener('keydown', event => {
        if (event.key === 'Escape' && modal.classList.contains('show')) window.closeAdvertiserModal();
    });

    updateCurrentValue();
})();
