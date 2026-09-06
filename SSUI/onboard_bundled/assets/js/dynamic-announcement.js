// dynamic-announcement.js
(function () {
    // Configuration
    const CONTAINER_ID = 'dynamic-announcement-list';
    const JSON_URL = 'https://steamserverui.github.io/StationeersServerUI/dynamic-announcement-list.json';
    const FETCH_TIMEOUT = 8000; // ms

    const container = document.getElementById(CONTAINER_ID);
    if (!container) {
        console.warn(`[Dynamic Announcement] Container #${CONTAINER_ID} not found on page.`);
        return;
    }

    container.innerHTML = '';

    const fetchWithTimeout = (url, options = {}, timeout = FETCH_TIMEOUT) => {
        return Promise.race([
            fetch(url, options),
            new Promise((_, reject) =>
                setTimeout(() => reject(new Error('Fetch timeout')), timeout)
            )
        ]);
    };

    function attachCollapsibleHandlers() {
        document.querySelectorAll('.info-notice h3').forEach(header => {
            // Avoid adding multiple listeners if called repeatedly
            header.removeEventListener('click', handleClick);
            header.addEventListener('click', handleClick);
        });
    }

    function handleClick(event) {
        const notice = event.currentTarget.parentElement;
        notice.classList.toggle('open');
    }

    fetchWithTimeout(JSON_URL, { method: 'GET', cache: 'no-cache' })
        .then(response => {
            if (!response.ok) {
                if (response.status === 404) {
                    console.info('[Dynamic Announcement] No announcement file (404).');
                    return null;
                }
                throw new Error(`HTTP ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            if (!data || (Array.isArray(data) && data.length === 0)) {
                console.info('[Dynamic Announcement] No announcements defined.');
                return;
            }

            const announcements = Array.isArray(data) ? data : [data];

            const now = Date.now();
            const validAnnouncements = announcements.filter(ann => {
                if (!ann.headline || !ann.bodyHtml) return false;

                const start = ann.validFrom ? new Date(ann.validFrom).getTime() : null;
                const end = ann.validUntil ? new Date(ann.validUntil).getTime() : null;

                if ((start !== null && now < start) || (end !== null && now > end)) {
                    return false;
                }

                return true;
            });

            if (validAnnouncements.length === 0) {
                console.info('[Dynamic Announcement] No currently valid announcements.');
                return;
            }

            validAnnouncements.sort((a, b) => {
                const timeA = a.validFrom ? new Date(a.validFrom).getTime() : 0;
                const timeB = b.validFrom ? new Date(b.validFrom).getTime() : 0;
                return timeB - timeA;
            });

            // Clear container
            container.innerHTML = '';

            validAnnouncements.forEach(ann => {
                const notice = document.createElement('div');
                notice.className = 'info-notice';

                // Build inner HTML
                let contentHTML = '';

                if (ann.shortDescription) {
                    contentHTML += `<p>${escapeHtml(ann.shortDescription)}</p>`;
                }

                contentHTML += `<p>${ann.bodyHtml}</p>`;

                if (ann.warningHtml) {
                    contentHTML += `<p class="status-bad">${ann.warningHtml}</p>`;
                }

                contentHTML += `<p></p>`; // spacer

                if (ann.author || ann.authorRole) {
                    const author = ann.author ? escapeHtml(ann.author) : '';
                    const role = ann.authorRole ? escapeHtml(ann.authorRole) : '';
                    contentHTML += `<p><i>${author}${author && role ? ' - ' : ''}${role}</i></p>`;
                }

                notice.innerHTML = `
                    <h3>
                        <span class="notice-icon">📢</span>
                        ${escapeHtml(ann.headline)}
                    </h3>
                    <div class="collapsible-content">
                        ${contentHTML}
                    </div>
                `;

                container.appendChild(notice);
            });

            attachCollapsibleHandlers();

            console.info(`[Dynamic Announcement] ${validAnnouncements.length} announcement(s) loaded and collapsible handlers attached.`);
        })
        .catch(err => {
            console.info('[Dynamic Announcement] Failed to load announcements:', err.message);
        });

    function escapeHtml(text) {
        if (typeof text !== 'string') return text;
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
})();