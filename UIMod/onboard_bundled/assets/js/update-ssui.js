let currentUpdateVersion = null;
let currentUpdateIsMajor = false;
let updateStarted = false;

function isTrue(value) {
    return value === true || value === 'true';
}

function pollUpdateStatus() {
    fetch('/api/v2/update/check', { cache: 'no-store' })
        .then(response => response.json())
        .then(data => {
            if (data.operationState === 'installing') {
                showInstallingState();
                return;
            }

            if (data.operationState === 'failed') {
                setAvailableUpdate(data);
                if (updateStarted) {
                    showUpdateFailure(data.message);
                }
                return;
            }

            if (updateStarted && !isTrue(data.updateAvailable)) {
                location.reload();
                return;
            }

            setAvailableUpdate(data);
        })
        .catch(error => {
            if (!updateStarted) {
                console.warn('Failed to check for updates:', error);
            }
        });
}

function setAvailableUpdate(data) {
    const updateButton = document.getElementById('update-button');
    if (!isTrue(data.updateAvailable) || !data.version) {
        currentUpdateVersion = null;
        currentUpdateIsMajor = false;
        updateButton.style.display = 'none';
        updateButton.classList.remove('bounce');
        return;
    }

    currentUpdateVersion = data.version;
    currentUpdateIsMajor = isTrue(data.majorUpdate);
    document.getElementById('modal-version-text').textContent = data.version;
    updateButton.style.display = 'block';
    updateButton.classList.add('bounce');
}

function openUpdateModal() {
    if (!currentUpdateVersion) {
        return;
    }
    document.getElementById('update-modal').classList.add('show');
    resetUpdateModal();
}

function resetUpdateModal() {
    updateStarted = false;
    document.getElementById('modal-version-text').textContent = currentUpdateVersion || '';
    document.getElementById('update-status-running').className = 'update-status-message';
    document.getElementById('update-status-failed').className = 'update-status-message';
    document.getElementById('update-error-detail').textContent = '';
    document.getElementById('update-later-btn').style.display = '';
    document.getElementById('update-now-btn').style.display = '';

    const warning = document.getElementById('major-update-warning');
    const confirmation = document.getElementById('major-update-confirm');
    warning.classList.toggle('show', currentUpdateIsMajor);
    document.querySelector('.update-modal-content').classList.toggle('major', currentUpdateIsMajor);
    confirmation.checked = false;
    confirmation.onchange = updateInstallButton;

    const releaseLink = document.getElementById('update-release-notes');
    releaseLink.href = 'https://github.com/SteamServerUI/StationeersServerUI/releases/tag/' + encodeURIComponent(currentUpdateVersion);
    updateInstallButton();
}

function updateInstallButton() {
    const confirmed = document.getElementById('major-update-confirm').checked;
    document.getElementById('update-now-btn').disabled = currentUpdateIsMajor && !confirmed;
}

function closeUpdateModal() {
    if (!updateStarted) {
        document.getElementById('update-modal').classList.remove('show');
    }
}

function startUpdate() {
    const majorApproved = currentUpdateIsMajor && document.getElementById('major-update-confirm').checked;
    if (currentUpdateIsMajor && !majorApproved) {
        return;
    }

    updateStarted = true;
    showInstallingState();

    fetch('/api/v2/update/trigger', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            allowUpdate: true,
            allowMajorUpdate: majorApproved,
            version: currentUpdateVersion
        })
    })
        .then(async response => {
            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.message || 'Update request failed');
            }
        })
        .catch(error => showUpdateFailure(error.message));
}

function showInstallingState() {
    updateStarted = true;
    document.getElementById('update-modal').classList.add('show');
    document.getElementById('update-later-btn').style.display = 'none';
    document.getElementById('update-now-btn').style.display = 'none';
    document.getElementById('update-status-failed').className = 'update-status-message';
    document.getElementById('update-status-running').className = 'update-status-message running';
}

function showUpdateFailure(message) {
    updateStarted = false;
    document.getElementById('update-modal').classList.add('show');
    document.getElementById('update-status-running').className = 'update-status-message';
    document.getElementById('update-status-failed').className = 'update-status-message failed';
    document.getElementById('update-error-detail').textContent = message || '';
}

document.getElementById('update-modal').addEventListener('click', event => {
    if (event.target === event.currentTarget) {
        closeUpdateModal();
    }
});

document.addEventListener('DOMContentLoaded', () => {
    pollUpdateStatus();
    setInterval(pollUpdateStatus, 1500);
});
