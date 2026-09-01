(() => {
    let fetchSequence = 0;

    const workspace = document.getElementById('backup-workspace');
    const list = document.getElementById('backupPageList');
    const notice = document.getElementById('backupPageNotice');
    const limit = document.getElementById('backupPageLimit');
    const refresh = document.getElementById('backupPageRefresh');
    const text = workspace?.dataset || {};

    function escapeHTML(value) {
        return String(value ?? '').replace(/[&<>"']/g, character => ({
            '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#039;'
        })[character]);
    }

    function formatNumber(value) {
        const number = Number(value);
        return Number.isFinite(number) ? number.toLocaleString() : '—';
    }

    function formatBytes(value) {
        const bytes = Number(value);
        if (!Number.isFinite(bytes) || bytes < 0) return '—';
        const units = ['B', 'KiB', 'MiB', 'GiB'];
        let amount = bytes;
        let unit = 0;
        while (amount >= 1024 && unit < units.length - 1) {
            amount /= 1024;
            unit++;
        }
        return `${amount.toLocaleString(undefined, { maximumFractionDigits: unit === 0 ? 0 : 1 })} ${units[unit]}`;
    }

    function iconMarkup(icon, fallback, className) {
        const encodedIcon = encodeURIComponent(icon);
        return `<span class="backup-ui-icon ${escapeHTML(className)}" aria-hidden="true">
            <span class="backup-icon-letter" hidden>${escapeHTML(fallback)}</span>
            <img src="/static/backupstats/${encodedIcon}.png"
                data-alternate-src="/static/backupstats/${encodedIcon}.webp"
                alt="" draggable="false">
        </span>`;
    }

    function wireIconFallbacks(root) {
        root.querySelectorAll('.backup-ui-icon img').forEach(image => {
            image.addEventListener('error', () => {
                if (image.dataset.alternateTried !== 'true') {
                    image.dataset.alternateTried = 'true';
                    image.src = image.dataset.alternateSrc;
                    return;
                }

                image.hidden = true;
                const fallback = image.previousElementSibling;
                if (fallback) fallback.hidden = false;
            });
        });
    }

    function summaryMetric(icon, fallback, label, value) {
        return `<span class="backup-core-metric">
            ${iconMarkup(icon, fallback, 'backup-core-icon')}
            <span><strong>${formatNumber(value)}</strong><small>${escapeHTML(label)}</small></span>
        </span>`;
    }

    function createBackupRow(backup, position) {
        const summary = backup.Summary || {};
        const index = Number(backup.Index);
        const worldName = summary.worldName || `${text.backupIndex} ${index}`;
        const item = document.createElement('li');
        item.className = 'backup-page-item';
        item.dataset.backupIndex = String(index);

        item.innerHTML = `
            <div class="backup-page-row">
                <div class="backup-page-identity">
                    <div class="backup-page-title-line">
                        <h2>${escapeHTML(worldName)}</h2>
                        <span>${escapeHTML(text.backupIndex)} ${index}</span>
                    </div>
                    <div class="backup-page-meta">
                        <span><small>${escapeHTML(text.created)}</small><strong>${new Date(backup.SaveTime).toLocaleString()}</strong></span>
                        <span><small>${escapeHTML(text.gameVersion)}</small><strong>${escapeHTML(summary.gameVersion || '—')}</strong></span>
                    </div>
                </div>
                <div class="backup-page-core-metrics">
                    ${summaryMetric('days', 'D', text.daysPlayed, summary.daysPlayed)}
                    ${summaryMetric('things', 'T', text.things, summary.things)}
                    ${summaryMetric('atmospheres', 'A', text.atmospheres, summary.atmospheres)}
                </div>
                <div class="backup-page-actions">
                    <button type="button" class="backup-download">${escapeHTML(text.download)}</button>
                    <button type="button" class="backup-restore">${escapeHTML(text.restore)}</button>
                </div>
                <button type="button" class="backup-page-chevron" aria-expanded="false"
                    aria-controls="backup-page-analysis-${index}" aria-label="${escapeHTML(text.expand)}">
                    <span aria-hidden="true">›</span>
                </button>
            </div>
            <section id="backup-page-analysis-${index}" class="backup-page-analysis" data-state="idle" hidden></section>`;

        const row = item.querySelector('.backup-page-row');
        row.addEventListener('click', event => {
            if (event.target.closest('button, a, input, select, textarea')) return;
            toggleAnalysis(item);
        });
        item.querySelector('.backup-download').addEventListener('click', () => downloadBackup(index));
        item.querySelector('.backup-restore').addEventListener('click', () => restoreBackup(index));
        item.querySelector('.backup-page-chevron').addEventListener('click', () => toggleAnalysis(item));
        wireIconFallbacks(item);

        if (position === 0) requestAnimationFrame(() => toggleAnalysis(item, true));
        return item;
    }

    function toggleAnalysis(item, forceOpen = false) {
        const toggle = item.querySelector('.backup-page-chevron');
        const panel = item.querySelector('.backup-page-analysis');
        const open = forceOpen || panel.hidden;
        panel.hidden = !open;
        item.classList.toggle('is-expanded', open);
        toggle.setAttribute('aria-expanded', String(open));
        toggle.setAttribute('aria-label', open ? text.collapse : text.expand);
        if (open) loadAnalysis(item);
    }

    function loadAnalysis(item) {
        const panel = item.querySelector('.backup-page-analysis');
        if (panel.dataset.state === 'loading' || panel.dataset.state === 'loaded') return;

        const index = Number(item.dataset.backupIndex);
        panel.dataset.state = 'loading';
        panel.innerHTML = `<div class="backup-analysis-message"><span class="backup-analysis-spinner"></span>${escapeHTML(text.analysisLoading)}</div>`;

        fetch(`/api/v2/backups/analyze?index=${index}`)
            .then(response => {
                if (!response.ok) return response.text().then(message => { throw new Error(message || text.analysisFailed); });
                return response.json();
            })
            .then(analysis => {
                if (!item.isConnected) return;
                panel.dataset.state = 'loaded';
                panel.innerHTML = renderAnalysis(analysis);
                wireIconFallbacks(panel);
            })
            .catch(error => {
                if (!item.isConnected) return;
                panel.dataset.state = 'error';
                panel.innerHTML = `<div class="backup-analysis-message is-error">
                    <span>${escapeHTML(text.analysisFailed)} ${escapeHTML(error.message)}</span>
                    <button type="button" class="backup-analysis-retry">${escapeHTML(text.retry)}</button>
                </div>`;
                panel.querySelector('.backup-analysis-retry').addEventListener('click', () => loadAnalysis(item));
            });
    }

    function statCard(icon, fallback, label, displayValue) {
        return `<div class="backup-bento-stat backup-bento-stat--${escapeHTML(icon)}">
            ${iconMarkup(icon, fallback, 'backup-deep-icon')}
            <span><strong>${escapeHTML(displayValue)}</strong><small>${escapeHTML(label)}</small></span>
        </div>`;
    }

    function renderAnalysis(analysis) {
        return `<section class="backup-bento-panel">
            <div class="backup-bento-grid">
                ${statCard('days', 'D', text.daysPlayed, formatNumber(analysis.daysPlayed))}
                ${statCard('things', 'T', text.things, formatNumber(analysis.things))}
                ${statCard('atmospheres', 'A', text.atmospheres, formatNumber(analysis.atmospheres))}
                ${statCard('archive-size', 'S', text.archiveSize, formatBytes(analysis.archiveSize))}
                ${statCard('players', 'P', text.players, formatNumber(analysis.players))}
                ${statCard('players-alive', 'A', text.playersAlive, formatNumber(analysis.playersAlive))}
                ${statCard('players-unconscious', 'U', text.playersUnconscious, formatNumber(analysis.playersUnconscious))}
                ${statCard('rooms', 'R', text.rooms, formatNumber(analysis.rooms))}
                ${statCard('furnaces', 'F', text.furnaces, formatNumber(analysis.furnaces))}
                ${statCard('destroyed-furnaces', 'D', text.destroyedFurnaces, formatNumber(analysis.destroyedFurnaces))}
                ${statCard('pipe-networks', 'P', text.pipeNetworks, formatNumber(analysis.pipeNetworks))}
                ${statCard('cable-networks', 'C', text.cableNetworks, formatNumber(analysis.cableNetworks))}
            </div>
        </section>`;
    }

    function showNotice(message, type = 'info') {
        notice.textContent = message;
        notice.className = `backup-page-notice is-${type}`;
        notice.hidden = false;
        window.setTimeout(() => { notice.hidden = true; }, 12000);
    }

    function fetchBackups() {
        const sequence = ++fetchSequence;
        const params = new URLSearchParams({ include: 'summary' });
        if (limit.value) params.set('limit', limit.value);
        list.innerHTML = `<li class="backup-page-empty">${escapeHTML(text.loading)}</li>`;
        refresh.disabled = true;

        fetch(`/api/v2/backups?${params}`)
            .then(response => {
                if (!response.ok) return response.text().then(message => { throw new Error(message); });
                return response.json();
            })
            .then(backups => {
                if (sequence !== fetchSequence) return;
                list.innerHTML = '';
                if (!Array.isArray(backups) || backups.length === 0) {
                    list.innerHTML = `<li class="backup-page-empty">${escapeHTML(text.none)}</li>`;
                    return;
                }
                backups.forEach((backup, position) => list.appendChild(createBackupRow(backup, position)));
            })
            .catch(error => {
                if (sequence !== fetchSequence) return;
                list.innerHTML = `<li class="backup-page-empty is-error">${escapeHTML(error.message)}</li>`;
            })
            .finally(() => {
                if (sequence === fetchSequence) refresh.disabled = false;
            });
    }

    function restoreBackup(index) {
        fetch(`/api/v2/backups/restore?index=${index}`)
            .then(response => response.text().then(message => ({ ok: response.ok, message })))
            .then(result => {
                if (!result.ok) throw new Error(result.message);
                showNotice(result.message, 'success');
            })
            .catch(error => showNotice(error.message, 'error'));
    }

    function downloadBackup(index) {
        fetch('/api/v2/backups/download', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ index })
        })
            .then(async response => {
                if (!response.ok) {
                    const error = await response.json().catch(() => ({}));
                    throw new Error(error.error || text.downloadFailed);
                }
                const disposition = response.headers.get('Content-Disposition') || '';
                const filename = disposition.match(/filename="([^"]+)"/)?.[1] || `backup_${index}.save`;
                return { blob: await response.blob(), filename };
            })
            .then(({ blob, filename }) => {
                const url = URL.createObjectURL(blob);
                const anchor = document.createElement('a');
                anchor.href = url;
                anchor.download = filename;
                document.body.appendChild(anchor);
                anchor.click();
                anchor.remove();
                URL.revokeObjectURL(url);
            })
            .catch(error => showNotice(error.message, 'error'));
    }

    refresh.addEventListener('click', fetchBackups);
    limit.addEventListener('change', fetchBackups);
    fetchBackups();
})();
