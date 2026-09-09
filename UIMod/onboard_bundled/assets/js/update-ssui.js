(function () {
    let updateStatus = null;
    let selectedVersion = '';
    let reconnecting = false;

    function installableCandidates(status) {
        const useful = [status.compatibleStable, status.latestStable, status.latestPrerelease, status.latestMajor];
        const seen = new Set();
        return useful.filter(candidate => {
            if (!candidate?.assetAvailable || seen.has(candidate.version)) return false;
            seen.add(candidate.version);
            return true;
        });
    }

    function preferredCandidate(status, candidates) {
        return status.compatibleStable?.assetAvailable
            ? status.compatibleStable
            : status.latestStable?.assetAvailable
                ? status.latestStable
                : candidates[0];
    }

    function releaseLabel(candidate) {
        const flags = [];
        if (candidate.major) flags.push('major');
        if (candidate.prerelease) flags.push('prerelease');
        if (!flags.length) flags.push('stable');
        return `${candidate.version} - ${flags.join(', ')}`;
    }

    function renderStatus(status) {
        updateStatus = status;
        const candidates = installableCandidates(status);
        const updateButton = document.getElementById('update-button');
        const select = document.getElementById('update-version-select');
        const previous = selectedVersion;
        updateButton.style.display = candidates.length ? 'block' : 'none';
        updateButton.classList.toggle('bounce', candidates.length > 0);
        select.replaceChildren(...candidates.map(candidate => {
            const option = document.createElement('option');
            option.value = candidate.version;
            option.textContent = releaseLabel(candidate);
            return option;
        }));
        const preferred = candidates.find(candidate => candidate.version === previous) || preferredCandidate(status, candidates);
        selectedVersion = preferred?.version || '';
        select.value = selectedVersion;
        selectUpdateVersion();
        if (status.installing) {
            selectedVersion = status.installingVersion || selectedVersion;
            showInstalling();
            monitorInstall();
        } else if (status.lastError && document.getElementById('update-modal').classList.contains('show')) {
            showUpdateFailure(status.lastError);
        }
    }

    async function pollUpdateStatus() {
        try {
            const response = await fetch('/api/v2/update/check', { cache: 'no-store' });
            if (!response.ok) return;
            renderStatus(await response.json());
        } catch (error) {
            if (!reconnecting) console.warn('Failed to check for updates:', error);
        }
    }

    function selectedCandidate() {
        return updateStatus?.candidates?.find(candidate => candidate.version === selectedVersion) || null;
    }

    window.selectUpdateVersion = function () {
        const select = document.getElementById('update-version-select');
        if (select.value) selectedVersion = select.value;
        const candidate = selectedCandidate();
        const major = Boolean(candidate?.major);
        const prerelease = Boolean(candidate?.prerelease);
        document.getElementById('major-update-warning').classList.toggle('show', major);
        document.getElementById('prerelease-update-warning').classList.toggle('show', prerelease);
        document.querySelector('.update-modal-content').classList.toggle('major', major || prerelease);
        document.getElementById('major-update-confirm').checked = false;
        document.getElementById('prerelease-update-confirm').checked = false;
        const releaseLink = document.getElementById('update-release-notes');
        releaseLink.hidden = !candidate?.releaseNotes;
        releaseLink.href = candidate?.releaseNotes || '#';
        updateInstallButton();
    };

    window.updateInstallButton = function () {
        const candidate = selectedCandidate();
        const majorConfirmed = !candidate?.major || document.getElementById('major-update-confirm').checked;
        const prereleaseConfirmed = !candidate?.prerelease || document.getElementById('prerelease-update-confirm').checked;
        document.getElementById('update-now-btn').disabled = !candidate || !majorConfirmed || !prereleaseConfirmed;
    };

    window.openUpdateModal = function () {
        if (!selectedCandidate()) return;
        resetUpdateModal();
        document.getElementById('update-modal').classList.add('show');
    };

    window.closeUpdateModal = function () {
        if (updateStatus?.installing) return;
        document.getElementById('update-modal').classList.remove('show');
    };

    window.resetUpdateModal = function () {
        reconnecting = false;
        document.getElementById('update-status-running').className = 'update-status-message';
        document.getElementById('update-status-failed').className = 'update-status-message';
        document.getElementById('update-error-detail').textContent = '';
        document.getElementById('update-modal-buttons').style.display = '';
        document.querySelectorAll('.major-update-warning').forEach(item => item.style.display = '');
        selectUpdateVersion();
    };

    function showInstalling() {
        document.getElementById('update-modal').classList.add('show');
        document.getElementById('update-modal-buttons').style.display = 'none';
        document.querySelectorAll('.major-update-warning').forEach(item => item.style.display = 'none');
        document.getElementById('update-status-failed').className = 'update-status-message';
        document.getElementById('update-status-running').className = 'update-status-message running';
        document.getElementById('update-running-text').textContent = `Installing ${selectedVersion}. SSUI will reconnect after the restart.`;
    }

    function showUpdateFailure(message) {
        reconnecting = false;
        document.getElementById('update-status-running').className = 'update-status-message';
        document.getElementById('update-status-failed').className = 'update-status-message failed';
        document.getElementById('update-error-detail').textContent = message || 'The update could not be installed. Check the backend log for details.';
    }

    window.startUpdate = async function () {
        const candidate = selectedCandidate();
        if (!candidate) return;
        showInstalling();
        try {
            const response = await fetch('/api/v2/update/trigger', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    action: 'install',
                    version: candidate.version,
                    confirmMajor: document.getElementById('major-update-confirm').checked,
                    confirmPrerelease: document.getElementById('prerelease-update-confirm').checked
                })
            });
            const result = await response.json();
            if (!response.ok) throw new Error(result?.error || 'The update request was rejected.');
            updateStatus.installing = true;
            updateStatus.installingVersion = candidate.version;
            monitorInstall();
        } catch (error) {
            showUpdateFailure(error.message);
        }
    };

    async function monitorInstall() {
        if (reconnecting) return;
        reconnecting = true;
        const deadline = Date.now() + 5 * 60 * 1000;
        while (reconnecting) {
            if (Date.now() > deadline) {
                showUpdateFailure('SSUI did not come back after the update. Check the backend log and start the new executable manually if needed.');
                return;
            }
            await new Promise(resolve => setTimeout(resolve, 3000));
            try {
                const response = await fetch('/api/v2/update/check', { cache: 'no-store' });
                if (!response.ok) continue;
                const status = await response.json();
                if (sameVersion(status.currentVersion, selectedVersion)) {
                    window.location.reload();
                    return;
                }
                if (status.lastError) {
                    showUpdateFailure(status.lastError);
                    return;
                }
                if (!status.installing) {
                    showUpdateFailure('The updater stopped before the new version started. Check the backend log for details.');
                    return;
                }
            } catch (_) {}
        }
    }

    function sameVersion(left, right) {
        return String(left || '').replace(/^v/, '') === String(right || '').replace(/^v/, '');
    }

    document.getElementById('update-modal').addEventListener('click', event => {
        if (event.target === event.currentTarget) closeUpdateModal();
    });
    document.addEventListener('DOMContentLoaded', () => {
        pollUpdateStatus();
        window.setInterval(pollUpdateStatus, 60000);
    });
})();
