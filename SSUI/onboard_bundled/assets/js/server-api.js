// /static/server-api.js

let backupFetchSequence = 0;

// Server control functions
function startServer() {
    toggleServer('/api/v3/server/start');
}

function stopServer() {
    toggleServer('/api/v3/server/stop');
}

function toggleServer(endpoint) {
    const action = endpoint.endsWith('/stop') ? 'stop' : 'start';
    if (!window.SSUIAccess.require('server.control', `You don't have permission to ${action} the server.`)) return;
    const status = document.getElementById('status');
    fetch(endpoint, { method: 'POST' })
        .then(async response => {
            const data = await response.json();
            if (!response.ok) throw new Error(data.error || 'Server action failed');
            return data;
        })
        .then(data => {
            const message = data.message || 'Request completed';
            status.hidden = false;
            typeTextWithCallback(status, message, 20, () => {
                setTimeout(() => status.hidden = true, 10000);
            });
        })
        .catch(err => window.SSUIAccess.notify(err.message));
}

function triggerSteamCMD() {
    if (!window.SSUIAccess.require('steamcmd.run', "You don't have permission to update the game server.")) return;
    const status = document.getElementById('status');
    status.hidden = false;
    typeTextWithCallback(status, 'Running SteamCMD, please wait... ', 20, () => {
        fetch('/api/v3/steamcmd/run', { method: 'POST' })
            .then(async response => {
                const data = await response.json();
                if (!response.ok) throw new Error(data.error || 'SteamCMD failed');
                return data;
            })
            .then(data => {
                showPopup("info", data.message);
            })
            .catch(err => {
                typeTextWithCallback(status, 'Error: Failed to trigger SteamCMD', 20, () => {
                    setTimeout(() => status.hidden = true, 10000);
                });
                console.error(`Failed to trigger SteamCMD:`, err);
            });
    });
}

function fetchBackups() {
    const requestSequence = ++backupFetchSequence;
    const limit = '3';
    const url = `/api/v3/backups?limit=${limit}&include=summary`;
    if (!window.SSUIAccess.can('backups.view')) {
        document.getElementById('backupList').innerHTML = '<li class="no-backups">You don\'t have permission to view backups.</li>';
        updateLatestBackupDisplay(undefined);
        return Promise.resolve();
    }
    
    return fetch(url)
        .then(response => {
            const contentType = response.headers.get('Content-Type');
            if (contentType && contentType.includes('application/json')) {
                return response.json().then(data => ({ status: response.ok, data }));
            } else {
                return response.text().then(text => ({ status: response.ok, text }));
            }
        })
        .then(result => {
            if (requestSequence !== backupFetchSequence) return;
            const backupList = document.getElementById('backupList');
            backupList.innerHTML = '';
            
            if (!result.status || result.text) {
                backupList.innerHTML = `<li class="backuperror">${result.text || 'Failed to load backups'}</li>`;
                updateLatestBackupDisplay(undefined);
                return;
            }
            
            const data = result.data?.items || [];
            if (data.length === 0) {
                backupList.innerHTML = '<li class="no-backups">No valid backup files found.</li>';
                updateLatestBackupDisplay(null);
                return;
            }

            updateLatestBackupDisplay(data[0]);
            
            let animationCount = 0;
            data.forEach((backup) => {
                const li = createBackupItem(backup);
                backupList.appendChild(li);
                
                if (animationCount < 20) {
                    setTimeout(() => {
                        li.classList.add('animate-in');
                    }, animationCount * 50);
                    animationCount++;
                }
            });
        })
        .catch(err => {
            if (requestSequence !== backupFetchSequence) return;
            console.error("Failed to fetch backups:", err);
            document.getElementById('backupList').innerHTML = '<li class="backuperror">Failed to load backups</li>';
            updateLatestBackupDisplay(undefined);
        });
}

function createBackupItem(backup) {
    const li = document.createElement('li');
    li.className = 'backup-item';
    const text = getBackupUIText();
    const summary = backup.summary || {};
    const name = backup.name;
    const title = summary.worldName || name;
    const gameVersion = summary.gameVersion
        ? `<span>${escapeBackupHTML(text.gameVersion)}: ${escapeBackupHTML(summary.gameVersion)}</span>`
        : '';
    li.innerHTML = `
        <div class="backup-row-main">
            <div class="backup-info">
                <div class="backup-header">
                    <span class="backup-name">${escapeBackupHTML(title)}</span>
                    <span class="backup-filename">${escapeBackupHTML(name)}</span>
                </div>
                <div class="backup-date">
                    <span>${escapeBackupHTML(text.created)}: ${new Date(backup.saveTime).toLocaleString()}</span>
                    ${gameVersion}
                </div>
                ${renderBackupSummaryStrip(summary)}
            </div>
            <div class="backup-actions">
                <button class="download-btn">Download</button>
                <button class="restore-btn">Restore</button>
            </div>
        </div>
    `;
    li.querySelector('.download-btn').addEventListener('click', () => downloadBackup(name));
    li.querySelector('.restore-btn').addEventListener('click', () => restoreBackup(name));
    return li;
}

function getBackupUIText() {
    const data = document.getElementById('backups')?.dataset || {};
    return {
        created: data.created || 'Created',
        daysPlayed: data.daysPlayed || 'Days played',
        things: data.things || 'Things',
        atmospheres: data.atmospheres || 'Atmospheres',
        rooms: data.rooms || 'Rooms',
        pipeNetworks: data.pipeNetworks || 'Pipe networks',
        cableNetworks: data.cableNetworks || 'Cable networks',
        players: data.players || 'Players',
        playersAlive: data.playersAlive || 'Alive',
        playersUnconscious: data.playersUnconscious || 'Unconscious',
        furnaces: data.furnaces || 'Furnaces',
        destroyedFurnaces: data.destroyedFurnaces || 'Destroyed furnaces',
        expand: data.expand || 'Show details',
        collapse: data.collapse || 'Hide details',
        analysisLoading: data.analysisLoading || 'Analyzing save file...',
        analysisFailed: data.analysisFailed || 'The save file could not be analyzed.',
        retry: data.retry || 'Retry',
        world: data.world || 'World',
        gameVersion: data.gameVersion || 'Game version',
        archiveSize: data.archiveSize || 'Archive size'
    };
}

function escapeBackupHTML(value) {
    return String(value ?? '').replace(/[&<>"']/g, character => ({
        '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#039;'
    })[character]);
}

function formatBackupNumber(value) {
    const number = Number(value);
    return Number.isFinite(number) ? number.toLocaleString() : '-';
}

function renderBackupSummaryStrip(summary) {
    if (!summary || Object.keys(summary).length === 0) return '';
    const text = getBackupUIText();
    const stats = [
        [text.daysPlayed, summary.daysPlayed],
        [text.things, summary.things],
        [text.atmospheres, summary.atmospheres]
    ];
    return `<div class="backup-summary-strip">${stats.map(([label, value]) => `
        <span class="backup-summary-value"><strong>${formatBackupNumber(value)}</strong>${escapeBackupHTML(label)}</span>
    `).join('')}</div>`;
}

function fetchPlayers() {
    const playersDiv = document.getElementById('players');
    const playerList = document.getElementById('playerList');
    const emptyState = document.getElementById('players-empty');
    if (!window.SSUIAccess.can('server.view')) {
        playersDiv.classList.add('is-empty');
        emptyState.textContent = "You don't have permission to view connected players.";
        emptyState.style.display = 'block';
        updateWorkspacePlayerState(false);
        updatePlayerCount(null);
        return Promise.resolve();
    }
    
    const playerImages = [
        "/static/playerimages/anna.webp",
        "/static/playerimages/dan.webp",
        "/static/playerimages/darragh.webp",
        "/static/playerimages/david.webp",
        "/static/playerimages/dean.webp",
        "/static/playerimages/garrison.webp",
        "/static/playerimages/ivette.webp",
        "/static/playerimages/john.webp",
        "/static/playerimages/julia.webp",
        "/static/playerimages/ove.webp",
        "/static/playerimages/pierre.webp",
        "/static/playerimages/rolf.webp",
        "/static/playerimages/ronald.webp",
    ];

    return fetch('/api/v3/server/players')
        .then(response => response.json())
        .then(data => {
            data = data.players || [];
            playerList.innerHTML = '';
            updatePlayerCount(Array.isArray(data) ? data.length : null);
            
            if (!Array.isArray(data) || data.length === 0) {
                playersDiv.classList.add('is-empty');
                emptyState.textContent = emptyState.dataset.empty;
                emptyState.style.display = 'block';
                updateWorkspacePlayerState(false);
                return;
            }

            playersDiv.classList.remove('is-empty');
            emptyState.style.display = 'none';
            updateWorkspacePlayerState(true);
            let animationCount = 0;
            data.forEach(player => {
                const li = document.createElement('li');
                li.className = 'player-item';
                
                // Create player item content
                const playerContent = document.createElement('div');
                playerContent.className = 'player-content';
                
                // Avatar
                const avatar = document.createElement('img');
                let persistedImage = sessionStorage.getItem(`playerImage_${player.steamId}`);
                if (!persistedImage) {
                    // Assign rnd image and persist it until page reload
                    persistedImage = playerImages[Math.floor(Math.random() * playerImages.length)];
                    sessionStorage.setItem(`playerImage_${player.steamId}`, persistedImage);
                }
                avatar.src = persistedImage;
                avatar.alt = `${player.username}'s avatar`;
                avatar.className = 'player-avatar';
                avatar.title = player.steamId;
                avatar.addEventListener('click', () => {
                    window.open(`https://steamcommunity.com/profiles/${player.steamId}`, '_blank');
                });
                
                const name = document.createElement('span');
                name.textContent = player.username;
                name.className = 'player-name';
                
                playerContent.appendChild(avatar);
                playerContent.appendChild(name);
                li.appendChild(playerContent);
                playerList.appendChild(li);
                
                // Animation
                if (animationCount < 20) {
                    setTimeout(() => {
                        li.classList.add('animate-in');
                    }, animationCount * 100);
                    animationCount++;
                }
            });
        })
        .catch(err => {
            console.error("Failed to fetch players:", err);
            playersDiv.classList.add('is-empty');
            emptyState.textContent = emptyState.dataset.error;
            emptyState.style.display = 'block';
            updateWorkspacePlayerState(false);
            updatePlayerCount(null);
        });
}

function updateWorkspacePlayerState(hasPlayers) {
    const workspace = document.getElementById('control-panel-workspace');
    if (!workspace) return;
    workspace.classList.toggle('has-players', hasPlayers);
    workspace.classList.toggle('no-players', !hasPlayers);
}

function updatePlayerCount(count) {
    const display = document.getElementById('player-count-display');
    if (!display) return;
    display.textContent = Number.isInteger(count) ? count : '-';
    display.title = Number.isInteger(count) ? `${count} connected player${count === 1 ? '' : 's'}` : 'Player count unavailable';
}

function updateLatestBackupDisplay(backup) {
    const display = document.getElementById('latest-backup-display');
    if (!display) return;

    if (!backup || !backup.saveTime) {
        display.textContent = backup === null ? 'None found' : 'Unavailable';
        display.title = '';
        return;
    }

    const created = new Date(backup.saveTime);
    if (Number.isNaN(created.getTime())) {
        display.textContent = 'Available';
        return;
    }

    const elapsedSeconds = Math.max(0, Math.floor((Date.now() - created.getTime()) / 1000));
    let age;
    if (elapsedSeconds < 60) age = 'Just now';
    else if (elapsedSeconds < 3600) age = `${Math.floor(elapsedSeconds / 60)}m ago`;
    else if (elapsedSeconds < 86400) age = `${Math.floor(elapsedSeconds / 3600)}h ago`;
    else age = `${Math.floor(elapsedSeconds / 86400)}d ago`;

    display.textContent = age;
    display.title = `${backup.name} · ${created.toLocaleString()}`;
}

function restoreBackup(name) {
    if (!window.SSUIAccess.require('backups.restore', "You don't have permission to restore backups.")) return;
    const status = document.getElementById('status');
    fetch('/api/v3/backups/restore', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name })
    })
        .then(response => response.text().then(message => ({ ok: response.ok, message })))
        .then(result => {
            if (!result.ok) throw new Error(result.message || 'Restore failed');
            return result.message || 'Backup restored successfully';
        })
        .then(data => {
            status.hidden = false;
            typeTextWithCallback(status, data, 20, () => {
                setTimeout(() => status.hidden = true, 30000);
            });
            showPopup('info', data);
        })
        .catch(err => {
            console.error(`Failed to restore backup ${name}:`, err);
            showPopup('error', err.message);
        });
}

function downloadBackup(name) {
    if (!window.SSUIAccess.require('backups.download', "You don't have permission to download backups.")) return;
    const status = document.getElementById('status');
    status.hidden = false;
    typeTextWithCallback(status, 'Preparing download...', 20, () => {});
    
    const selection = new URLSearchParams({ name });
    fetch(`/api/v3/backups/download?${selection}`)
    .then(response => {
        if (!response.ok) {
            return response.json().then(err => { throw new Error(err.error || 'Download failed'); });
        }
        const disposition = response.headers.get('Content-Disposition');
        let filename = name.split("/").pop();
        if (disposition) {
            const match = disposition.match(/filename="(.+)"/);
            if (match) filename = match[1];
        }
        return response.blob().then(blob => ({ blob, filename }));
    })
    .then(({ blob, filename }) => {
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        a.remove();
        status.hidden = true;
    })
    .catch(err => {
        console.error(`Failed to download backup ${name}:`, err);
        showPopup('error', 'Download failed: ' + err.message);
        status.hidden = true;
    });
}

function pollRecurringTasks() {
    window.gamserverstate = false;

    const fetchServerStatus = () => {
        fetch('/api/v3/server/status')
            .then(response => response.json())
            .then(data => {
                updateStatusIndicator(data.running, false, formatAPIUptime(data.uptimeSeconds), data.state);
                if (data.serverId) {
                    localStorage.setItem('gameserverrunID', data.serverId);
                }
            })
            .catch(err => {
                console.error("Failed to fetch server status:", err);
                updateStatusIndicator(false, true); // Set error state
            });
    };

    // Fetch immediately, then poll server status every 3.5 seconds
    if (window.SSUIAccess.can('server.view')) {
        fetchServerStatus();
        setInterval(fetchServerStatus, 3500);

        fetchConnectivityStatus();
        setInterval(fetchConnectivityStatus, 10000);
    } else {
        updateStatusIndicator(false, true);
        document.getElementById('server-state-label').textContent = 'No permission';
    }

    // Poll connectred players every 10 seconds
    if (window.SSUIAccess.can('server.view')) {
        setInterval(() => {
            fetchPlayers().catch(err => console.error("Failed to fetch connectedplayers:", err));
        }, 10000);
    }

    // Poll backups every 30 seconds
    if (window.SSUIAccess.can('backups.view')) {
        setInterval(() => {
            fetchBackups().catch(err => console.error("Failed to fetch backups:", err));
        }, 30000);
    }
}

function fetchConnectivityStatus() {
    fetch('/api/v3/server/connectivity')
        .then(response => response.json().then(data => ({ ok: response.ok, data })))
        .then(result => {
            if (!result.ok) throw new Error(result.data?.error || 'Connectivity status unavailable');
            updateConnectivityPanel(result.data);
        })
        .catch(err => console.debug('Failed to fetch connectivity status:', err));
}

function runConnectivityCheck() {
    const panel = document.getElementById('connectivity-panel');
    const permissionMessage = panel?.dataset.permission || '';
    if (!window.SSUIAccess.require('server.control', permissionMessage)) return;

    const button = document.getElementById('connectivity-retry');
    if (button) button.disabled = true;
    updateConnectivityPanel({ state: 'checking' });

    fetch('/api/v3/server/connectivity/check', { method: 'POST' })
        .then(response => response.json().then(data => ({ ok: response.ok, data })))
        .then(result => {
            if (!result.ok && result.data?.state !== 'checking') {
                throw new Error(result.data?.error || 'Connectivity check failed');
            }
            updateConnectivityPanel(result.data);
        })
        .catch(() => updateConnectivityPanel({ state: 'service_unavailable' }))
        .finally(() => {
            if (button) button.disabled = false;
        });
}

const connectivityWarningStorageKey = 'ssui-connectivity-warning-dismissed';

function isConnectivityWarningDismissed() {
    try {
        return localStorage.getItem(connectivityWarningStorageKey) === 'true';
    } catch (error) {
        return false;
    }
}

function clearConnectivityWarningDismissal() {
    try {
        localStorage.removeItem(connectivityWarningStorageKey);
    } catch (error) {
        // Storage can be unavailable in private or restricted browser contexts.
    }
}

function dismissConnectivityWarning() {
    try {
        localStorage.setItem(connectivityWarningStorageKey, 'true');
    } catch (error) {
        // The warning still disappears for this render if storage is unavailable.
    }
    const panel = document.getElementById('connectivity-panel');
    if (panel) {
        panel.hidden = true;
        panel.classList.remove('open');
    }
}

function placeConnectivityPanel(state) {
    const panel = document.getElementById('connectivity-panel');
    const controls = document.getElementById('controls');
    const legacyNotice = document.getElementById('legacy-info-notice');
    if (!panel || !controls || !legacyNotice) return;

    if (state === 'reachable') {
        legacyNotice.after(panel);
        return;
    }

    controls.after(panel);
}

function updateConnectivityPanel(status) {
    const panel = document.getElementById('connectivity-panel');
    const summary = document.getElementById('connectivity-summary');
    const title = document.getElementById('connectivity-title');
    const message = document.getElementById('connectivity-message');
    const probes = document.getElementById('connectivity-probes');
    const retry = document.getElementById('connectivity-retry');
    const dismiss = document.getElementById('connectivity-dismiss');
    if (!panel || !summary || !title || !message || !probes) return;

    const copy = panel.dataset;
    const titles = {
        checking: copy.checking,
        reachable: copy.reachable,
        partial: copy.partial,
        unreachable: copy.unreachable,
        bind_failed: copy.bindFailed,
        configuration_error: copy.configError,
        service_unavailable: copy.serviceError
    };
    const messages = {
        checking: copy.messageChecking,
        reachable: copy.messageReachable,
        partial: copy.messagePartial,
        unreachable: copy.messageUnreachable,
        bind_failed: copy.messageBindFailed,
        configuration_error: copy.messageConfigError,
        service_unavailable: copy.messageServiceError
    };

    const state = status?.state || 'not_checked';
    if (state === 'not_checked' || state === 'disabled' || state === 'skipped_running') {
        panel.hidden = true;
        return;
    }

    placeConnectivityPanel(state);

    if (state === 'reachable') {
        clearConnectivityWarningDismissal();
    }

    const warning = state !== 'checking';
    panel.hidden = warning && isConnectivityWarningDismissed();
    panel.classList.remove(
        'connectivity-checking',
        'connectivity-reachable',
        'connectivity-partial',
        'connectivity-unreachable',
        'connectivity-bind_failed',
        'connectivity-configuration_error',
        'connectivity-service_unavailable'
    );
    panel.classList.add(`connectivity-${state}`);
    const stateTitle = titles[state] || copy.defaultTitle;
    summary.textContent = `${copy.notice}: ${state === 'reachable' ? copy.allCrisp : stateTitle}`;
    title.textContent = stateTitle;
    message.textContent = messages[state] || copy.messageServiceError;
    probes.replaceChildren();

    if (dismiss) dismiss.hidden = !warning;

    (status.probes || []).forEach(probe => {
        const item = document.createElement('span');
        item.className = probe.received ? 'probe-ok' : 'probe-failed';
        item.textContent = `${probe.size} ${copy.bytes}: ${probe.received ? copy.received : copy.lost}`;
        probes.appendChild(item);
    });

    if (retry) retry.hidden = !window.SSUIAccess.can('server.control');
}

function formatAPIUptime(seconds) {
    seconds = Math.max(0, Number(seconds) || 0);
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const remaining = Math.floor(seconds % 60);
    return [days && `${days}d`, (hours || days) && `${hours}h`, (minutes || hours || days) && `${minutes}m`, `${remaining}s`]
        .filter(Boolean)
        .join('');
}

function updateStatusIndicator(isRunning, isError = false, uptime = '', state = 'uncertain') {
    const indicator = document.getElementById('status-indicator');
    const uptimeDisplay = document.getElementById('uptime-display');
    const stateLabel = document.getElementById('server-state-label');
    const startButton = document.getElementById('start-server-button');
    const stopButton = document.getElementById('stop-server-button');
    
    if (isError) {
        indicator.className = 'status-indicator error';
        indicator.title = 'Server status temporarily unavailable';
        // A failed HTTP poll is not evidence that the game server state is
        // uncertain. Keep the last lifecycle label, uptime and controls until
        // the next successful status response.
        return;
    }
    
    const normalizedState = state || (isRunning ? 'uncertain' : 'stopped');
    if (!isRunning || normalizedState === 'stopped') {
        indicator.className = 'status-indicator offline';
        indicator.title = 'Server is offline';
        window.gamserverstate = false;
    } else if (normalizedState === 'running') {
        indicator.className = 'status-indicator online';
        indicator.title = 'Server is running';
        window.gamserverstate = true;
    } else {
        indicator.className = `status-indicator ${normalizedState === 'uncertain' ? 'uncertain' : 'starting'}`;
        indicator.title = `Server state: ${normalizedState}`;
        window.gamserverstate = true;
    }

    if (stateLabel) {
        const stateDataKey = normalizedState.replace(/-([a-z])/g, (_, letter) => letter.toUpperCase());
        stateLabel.textContent = stateLabel.dataset[stateDataKey] || stateLabel.dataset.uncertain;
    }

    if (startButton) startButton.disabled = isRunning;
    if (stopButton) stopButton.disabled = !isRunning;

    // Show uptime only when server is running and uptime is not "0s"
    if (uptimeDisplay) {
        if (isRunning && uptime && uptime !== '0s') {
            uptimeDisplay.textContent = uptime;
        } else {
            uptimeDisplay.textContent = '-';
        }
    }
}
