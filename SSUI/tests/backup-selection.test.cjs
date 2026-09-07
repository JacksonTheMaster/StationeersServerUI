const assert = require('node:assert/strict');
const { readFileSync } = require('node:fs');
const path = require('node:path');
const { test } = require('node:test');
const vm = require('node:vm');

// Exercise the shipped scripts without a server, DOM dependency or network access.
function element() {
    const selectors = new Map();
    return {
        dataset: {}, children: [], events: {}, hidden: true, isConnected: true,
        classList: { add() {}, toggle() {} }, value: '20',
        set innerHTML(value) { this.html = value; this.children = []; },
        get innerHTML() { return this.html || ''; },
        appendChild(child) { this.children.push(child); },
        addEventListener(event, fn) { this.events[event] = fn; },
        querySelector(selector) {
            if (!selectors.has(selector)) selectors.set(selector, element());
            return selectors.get(selector);
        },
        querySelectorAll() { return []; },
        setAttribute() {}, click() { this.events.click?.(); }, remove() {}
    };
}

function harness(page) {
    const elements = new Map();
    const requests = [];
    const popups = [];
    const body = element();
    const state = {
        backups: [{ name: 'nested/ä + # <world>.save', saveTime: '2026-09-04T12:00:00Z' }],
        restoreOK: true
    };
    const document = {
        body, createElement: element,
        getElementById(id) {
            if (!elements.has(id)) elements.set(id, element());
            return elements.get(id);
        }
    };
    const browserURL = { createObjectURL() { return 'blob:test'; }, revokeObjectURL() {} };
    const context = vm.createContext({
        document, URLSearchParams, console: { error() {} },
        URL: browserURL,
        setTimeout() {}, window: { URL: browserURL, setTimeout() {} }, requestAnimationFrame: fn => fn(),
        typeTextWithCallback(_element, _text, _delay, callback) { callback(); },
        showPopup(type, message) { popups.push({ type, message }); },
        async fetch(url, options) {
            requests.push({ url, options });
            const target = new URL(url, 'http://local');
            return {
                ok: target.pathname.endsWith('/restore') ? state.restoreOK : true,
                headers: { get(name) { return name === 'Content-Type' ? 'application/json' : ''; } },
                async json() { return target.pathname === '/api/v3/backups' ? state.backups : {}; },
                async text() { return state.restoreOK ? 'Restored' : 'Backup missing'; },
                async blob() { return {}; }
            };
        }
    });
    const script = page === 'dashboard' ? 'server-api.js' : 'backups-page.js';
    vm.runInContext(readFileSync(path.join(__dirname, '../onboard_bundled/assets/js', script), 'utf8'), context);
    const refresh = () => page === 'dashboard'
        ? vm.runInContext('fetchBackups()', context)
        : document.getElementById('backupPageRefresh').click();
    const rows = () => document.getElementById(page === 'dashboard' ? 'backupList' : 'backupPageList').children;
    return { state, requests, popups, document, refresh, rows, body };
}

const settle = () => new Promise(resolve => setImmediate(resolve));

for (const page of ['dashboard', 'workspace']) {
    test(`${page}: actions retain names after list changes`, async () => {
        const h = harness(page);
        await h.refresh();
        await settle();
        const selected = h.rows()[0];
        const name = h.state.backups[0].name;
        assert.ok(selected.innerHTML.includes('&lt;world&gt;.save'));
        assert.ok(!selected.innerHTML.includes('<world>'));

        h.state.backups = [{ name: 'other.save', saveTime: '2026-09-05T12:00:00Z' }, ...h.state.backups];
        await h.refresh();
        await settle();
        selected.querySelector(page === 'dashboard' ? '.download-btn' : '.backup-download').click();
        selected.querySelector(page === 'dashboard' ? '.restore-btn' : '.backup-restore').click();
        await settle();

        const download = h.requests.find(request => request.url.endsWith('/download'));
        assert.deepEqual(JSON.parse(download.options.body), { name });
        const restore = h.requests.find(request => request.url.includes('/restore?'));
        assert.deepEqual([...new URL(restore.url, 'http://local').searchParams], [['name', name]]);
        assert.equal(h.body.children.at(-1).download, name.split('/').pop());
        if (page === 'workspace') {
            const analysis = h.requests.find(request => request.url.includes('/analysis?'));
            assert.deepEqual([...new URL(analysis.url, 'http://local').searchParams], [['name', name]]);
        }
    });

    test(`${page}: missing restore is shown as an error`, async () => {
        const h = harness(page);
        await h.refresh();
        await settle();
        h.state.restoreOK = false;
        h.rows()[0].querySelector(page === 'dashboard' ? '.restore-btn' : '.backup-restore').click();
        await settle();
        if (page === 'dashboard') {
            assert.deepEqual(h.popups, [{ type: 'error', message: 'Backup missing' }]);
        } else {
            assert.equal(h.document.getElementById('backupPageNotice').className, 'backup-page-notice is-error');
            assert.equal(h.document.getElementById('backupPageNotice').textContent, 'Backup missing');
        }
    });
}
