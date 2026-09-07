(function () {
    const nativeFetch = window.fetch.bind(window);
    const permissions = new Set((document.body?.dataset.permissions || '').split(',').filter(Boolean));
    const authenticatedPage = document.body?.dataset.authenticated === 'true';
    let notificationTimer;

    function notify(message, type = 'error') {
        let notification = document.getElementById('ssui-notification');
        if (!notification) {
            notification = document.createElement('div');
            notification.id = 'ssui-notification';
            notification.setAttribute('role', 'status');
            notification.setAttribute('aria-live', 'polite');
            document.body.appendChild(notification);
        }
        window.clearTimeout(notificationTimer);
        notification.textContent = message;
        notification.className = `ssui-notification is-${type} show`;
        notificationTimer = window.setTimeout(() => notification.classList.remove('show'), type === 'error' ? 7000 : 4000);
    }

    window.SSUIAccess = {
        can(permission) {
            return permissions.has(permission);
        },
        require(permission, message) {
            if (permissions.has(permission)) return true;
            notify(message || "You don't have permission to use this action.");
            return false;
        },
        notify
    };

    const denied = new URLSearchParams(window.location.search).get('denied');
    if (denied) {
        notify(denied);
        const url = new URL(window.location.href);
        url.searchParams.delete('denied');
        window.history.replaceState({}, '', url);
    }

    function cookie(name) {
        const prefix = `${name}=`;
        const item = document.cookie.split(';').map(value => value.trim()).find(value => value.startsWith(prefix));
        return item ? decodeURIComponent(item.slice(prefix.length)) : '';
    }

    window.fetch = async function (input, options = {}) {
        const request = input instanceof Request ? input : null;
        const url = new URL(request ? request.url : String(input), window.location.href);
        const method = String(options.method || request?.method || 'GET').toUpperCase();
        const next = { ...options, credentials: options.credentials || 'same-origin' };

        if (url.origin === window.location.origin && !['GET', 'HEAD', 'OPTIONS'].includes(method)) {
            const csrf = cookie('SSUICSRF');
            if (csrf) {
                const headers = new Headers(request?.headers || options.headers || {});
                headers.set('X-SSUI-CSRF', csrf);
                next.headers = headers;
            }
        }

        const response = await nativeFetch(input, next);
        if (authenticatedPage && response.status === 401) {
            notify('Your session has expired. Please sign in again.');
        } else if (response.status === 403 && !['GET', 'HEAD', 'OPTIONS'].includes(method)) {
            const error = await response.clone().json().catch(() => null);
            notify(error?.error?.message || "You don't have permission to use this action.");
        }
        const readJSON = response.json.bind(response);
        const readText = response.text.bind(response);
        response.json = async function () {
            const envelope = await readJSON();
            if (!envelope || typeof envelope !== 'object') return envelope;
            if (Object.prototype.hasOwnProperty.call(envelope, 'data')) return envelope.data;
            if (envelope.error && typeof envelope.error === 'object') {
                return { error: envelope.error.message, errorCode: envelope.error.code, details: envelope.error.details };
            }
            return envelope;
        };
        response.text = async function () {
            const text = await readText();
            try {
                const envelope = JSON.parse(text);
                if (envelope?.error?.message) return envelope.error.message;
                if (envelope?.data?.message) return envelope.data.message;
            } catch (_) {
                // This was a regular text response.
            }
            return text;
        };
        return response;
    };
})();
