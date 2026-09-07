(() => {
    const root = document.getElementById('access-config-tab');
    if (!root) return;

    const state = {
        session: null,
        users: [],
        groups: [],
        permissions: [],
        selectedGroup: 'new'
    };

    const presets = {
        viewer: ['server.view', 'backups.view'],
        operator: [
            'server.view', 'server.control', 'console.read', 'console.write',
            'backups.view', 'backups.download', 'steamcmd.run'
        ],
        backups: ['server.view', 'backups.view', 'backups.analyze', 'backups.download', 'backups.restore'],
        administrator: null
    };

    function escapeHTML(value) {
        return String(value ?? '').replace(/[&<>"']/g, character => ({
            '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#039;'
        })[character]);
    }

    async function request(url, options) {
        const response = await fetch(url, options);
        if (response.status === 204) return null;
        const result = await response.json().catch(() => ({}));
        if (!response.ok) throw new Error(result.error || result.message || 'The request failed.');
        return result;
    }

    function notify(message, type = 'error') {
        window.SSUIAccess.notify(message, type);
    }

    function can(permission) {
        return window.SSUIAccess.can(permission);
    }

    function showPanel(id) {
        root.querySelectorAll('.access-panel').forEach(panel => panel.classList.toggle('active', panel.id === id));
        root.querySelectorAll('.access-nav-button').forEach(button => button.classList.toggle('active', button.dataset.accessPanel === id));
    }

    function groupName(id) {
        return state.groups.find(group => group.id === id)?.name || id;
    }

    function canGrantGroup(group) {
        return group.permissions.every(permission => can(permission)) &&
            (group.id !== 'system-owner' || can('security.manage'));
    }

    function canEditUserGroups(user) {
        return user.groupIds.every(id => state.groups.some(group => group.id === id && canGrantGroup(group)));
    }

    function userName(id) {
        if (state.session?.user.id === id) return state.session.user.username;
        return state.users.find(user => user.id === id)?.username || id;
    }

    function formatDate(value) {
        if (!value) return 'Never';
        const date = new Date(value);
        return Number.isNaN(date.getTime()) ? 'Unknown' : date.toLocaleString();
    }

    async function loadSession() {
        state.session = await request('/api/v3/auth/session');
        document.getElementById('access-current-user').textContent = state.session.user.username;
    }

    async function changeOwnPassword(event) {
        event.preventDefault();
        const form = event.currentTarget;
        const currentPassword = form.elements.currentPassword.value;
        const newPassword = form.elements.newPassword.value;
        if (newPassword !== form.elements.confirmPassword.value) {
            notify('The new passwords do not match.');
            return;
        }
        try {
            await request('/api/v3/auth/password', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ currentPassword, newPassword })
            });
            notify('Password changed. Please sign in again.', 'success');
            window.setTimeout(() => window.location.href = '/login', 1200);
        } catch (error) {
            notify(error.message);
        }
    }

    async function loadSessions() {
        const list = document.getElementById('access-session-list');
        try {
            const result = await request('/api/v3/auth/sessions');
            if (!result.sessions.length) {
                list.innerHTML = '<p class="access-empty">No active sessions.</p>';
                return;
            }
            list.innerHTML = result.sessions.map(session => `
                <div class="access-list-item">
                    <div><strong>${escapeHTML(userName(session.userId))}</strong>
                        <small>${session.id === state.session.credentialId ? 'This session' : `Last used ${escapeHTML(formatDate(session.lastUsedAt))}`}</small></div>
                    <div class="access-list-actions"><button type="button" data-revoke-session="${escapeHTML(session.id)}">Sign out</button></div>
                </div>`).join('');
        } catch (error) {
            list.innerHTML = `<p class="access-empty">${escapeHTML(error.message)}</p>`;
        }
    }

    async function revokeSession(id) {
        try {
            await request(`/api/v3/auth/sessions/${encodeURIComponent(id)}`, { method: 'DELETE' });
            if (id === state.session.credentialId) {
                window.location.href = '/login';
                return;
            }
            notify('Session signed out.', 'success');
            await loadSessions();
        } catch (error) {
            notify(error.message);
        }
    }

    async function loadGroups() {
        if (!can('groups.manage') && !can('users.manage')) return;
        const result = await request('/api/v3/auth/groups');
        state.groups = result.groups;
        state.permissions = result.permissions.filter(permission => can(permission));
        populateGroupSelectors();
        if (can('groups.manage')) {
            renderPermissionGrid('access-permission-grid', state.permissions);
            if (state.selectedGroup !== 'new' && !state.groups.some(group => group.id === state.selectedGroup)) {
                state.selectedGroup = 'new';
            }
            selectGroup(state.selectedGroup);
        }
    }

    function populateGroupSelectors() {
        const groupSelect = document.getElementById('access-group-select');
        const createSelect = document.querySelector('#access-create-user select[name="groupId"]');
        if (can('groups.manage')) {
            groupSelect.innerHTML = '<option value="new">Create a new group</option>' + state.groups.map(group =>
                `<option value="${escapeHTML(group.id)}">${escapeHTML(group.name)}${group.system ? ' (system)' : ''}</option>`
            ).join('');
        }
        const grantableGroups = state.groups.filter(canGrantGroup);
        createSelect.innerHTML = '<option value="">No access group</option>' + grantableGroups.map(group =>
            `<option value="${escapeHTML(group.id)}">${escapeHTML(group.name)}</option>`
        ).join('');
    }

    function renderPermissionGrid(targetId, permissions, selected = []) {
        const target = document.getElementById(targetId);
        const groups = new Map();
        permissions.forEach(permission => {
            const category = permission.split('.')[0];
            if (!groups.has(category)) groups.set(category, []);
            groups.get(category).push(permission);
        });
        target.innerHTML = Array.from(groups.entries()).map(([category, entries]) => `
            <div class="access-permission-group">
                <h5>${escapeHTML(category)}</h5>
                ${entries.map(permission => `<label class="access-permission-option">
                    <input type="checkbox" value="${escapeHTML(permission)}" ${selected.includes(permission) ? 'checked' : ''}>
                    <span>${escapeHTML(permission.slice(category.length + 1).replaceAll('.', ' '))}</span>
                </label>`).join('')}
            </div>`).join('');
    }

    function checkedPermissions(targetId) {
        return Array.from(document.querySelectorAll(`#${targetId} input:checked`)).map(input => input.value);
    }

    function setCheckedPermissions(targetId, permissions) {
        const selected = new Set(permissions);
        document.querySelectorAll(`#${targetId} input`).forEach(input => input.checked = selected.has(input.value));
    }

    function selectGroup(id) {
        const form = document.getElementById('access-group-form');
        const deleteButton = document.getElementById('access-delete-group');
        const saveButton = form.querySelector('button[type="submit"]');
        const preset = document.getElementById('access-preset-select');
        state.selectedGroup = id;
        document.getElementById('access-group-select').value = id;
        preset.value = 'custom';

        const group = state.groups.find(item => item.id === id);
        form.elements.name.value = group?.name || '';
        form.elements.description.value = group?.description || '';
        setCheckedPermissions('access-permission-grid', group?.permissions || []);
        const exceedsAccess = Boolean(group?.permissions.some(permission => !can(permission)));
        const locked = Boolean(group?.system || exceedsAccess);
        form.elements.name.disabled = locked;
        form.elements.description.disabled = locked;
        form.querySelectorAll('#access-permission-grid input').forEach(input => input.disabled = locked);
        preset.disabled = locked;
        document.getElementById('access-clear-permissions').disabled = locked;
        saveButton.hidden = locked;
        deleteButton.hidden = !group || locked;
    }

    function applyPreset(name) {
        if (name === 'custom') return;
        setCheckedPermissions('access-permission-grid', name === 'administrator' ? state.permissions : presets[name]);
    }

    async function saveGroup(event) {
        event.preventDefault();
        const form = event.currentTarget;
        const payload = {
            name: form.elements.name.value.trim(),
            description: form.elements.description.value.trim(),
            permissions: checkedPermissions('access-permission-grid')
        };
        try {
            if (state.selectedGroup === 'new') {
                await request('/api/v3/auth/groups', {
                    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload)
                });
            } else {
                await request(`/api/v3/auth/groups/${encodeURIComponent(state.selectedGroup)}`, {
                    method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload)
                });
            }
            notify('Access group saved.', 'success');
            state.selectedGroup = 'new';
            await loadGroups();
            if (can('users.manage')) await loadUsers();
        } catch (error) {
            notify(error.message);
        }
    }

    async function deleteGroup() {
        const group = state.groups.find(item => item.id === state.selectedGroup);
        if (!group || !window.confirm(`Delete the access group "${group.name}"?`)) return;
        try {
            await request(`/api/v3/auth/groups/${encodeURIComponent(group.id)}`, { method: 'DELETE' });
            notify('Access group deleted.', 'success');
            state.selectedGroup = 'new';
            await loadGroups();
            if (can('users.manage')) await loadUsers();
        } catch (error) {
            notify(error.message);
        }
    }

    async function loadUsers() {
        if (!can('users.manage')) return;
        const result = await request('/api/v3/auth/users');
        state.users = result.users;
        renderUsers();
    }

    function userGroupOptions(user) {
        const grantableGroups = state.groups.filter(canGrantGroup);
        if (!canEditUserGroups(user)) {
            return '<p class="access-empty">This user belongs to an access group above your own access. Group membership cannot be changed.</p>';
        }
        if (!grantableGroups.length) return '';
        return `<div class="access-user-groups">${grantableGroups.map(group => `<label class="access-permission-option">
            <input type="checkbox" value="${escapeHTML(group.id)}" ${user.groupIds.includes(group.id) ? 'checked' : ''}>
            <span>${escapeHTML(group.name)}</span>
        </label>`).join('')}</div>`;
    }

    function renderUsers() {
        const list = document.getElementById('access-user-list');
        list.innerHTML = state.users.map(user => `
            <div class="access-list-item access-user-item" data-user-id="${escapeHTML(user.id)}">
                <div><strong>${escapeHTML(user.username)}</strong>
                    <small>${user.enabled ? 'Enabled' : 'Disabled'} · ${escapeHTML(user.groupIds.map(groupName).join(', ') || 'No access group')}</small></div>
                <div class="access-list-actions"><button type="button" data-edit-user>Manage</button></div>
                <form class="access-user-editor" data-can-edit-groups="${canEditUserGroups(user)}" hidden>
                    <label class="access-toggle"><input type="checkbox" name="enabled" ${user.enabled ? 'checked' : ''}> Account enabled</label>
                    ${userGroupOptions(user)}
                    <label>New password <small>Leave empty to keep it unchanged</small>
                        <input type="password" name="password" minlength="10" autocomplete="new-password"></label>
                    <div class="access-form-actions">
                        <button type="submit">Save user</button>
                        <button type="button" class="access-danger-button" data-delete-user>Delete user</button>
                    </div>
                </form>
            </div>`).join('');
    }

    async function createUser(event) {
        event.preventDefault();
        const form = event.currentTarget;
        const groupId = form.elements.groupId.value;
        try {
            await request('/api/v3/auth/users', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    username: form.elements.username.value.trim(),
                    password: form.elements.password.value,
                    groupIds: groupId ? [groupId] : []
                })
            });
            form.reset();
            notify('User added.', 'success');
            await loadUsers();
        } catch (error) {
            notify(error.message);
        }
    }

    async function saveUser(form) {
        const item = form.closest('[data-user-id]');
        const user = state.users.find(entry => entry.id === item.dataset.userId);
        const password = form.elements.password.value;
        const payload = { enabled: form.elements.enabled.checked };
        if (form.dataset.canEditGroups === 'true') {
            payload.groupIds = Array.from(form.querySelectorAll('.access-user-groups input:checked')).map(input => input.value);
        }
        if (password) payload.password = password;
        try {
            await request(`/api/v3/auth/users/${encodeURIComponent(user.id)}`, {
                method: 'PATCH', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload)
            });
            if (user.id === state.session.user.id && (password || !payload.enabled)) {
                window.location.href = '/login';
                return;
            }
            notify('User updated.', 'success');
            await loadUsers();
        } catch (error) {
            notify(error.message);
        }
    }

    async function deleteUser(id) {
        const user = state.users.find(entry => entry.id === id);
        if (!user || !window.confirm(`Delete the user "${user.username}"?`)) return;
        try {
            await request(`/api/v3/auth/users/${encodeURIComponent(id)}`, { method: 'DELETE' });
            if (id === state.session.user.id) {
                window.location.href = '/login';
                return;
            }
            notify('User deleted.', 'success');
            await loadUsers();
        } catch (error) {
            notify(error.message);
        }
    }

    async function loadTokens() {
        if (!can('tokens.manage')) return;
        const result = await request('/api/v3/auth/tokens');
        const list = document.getElementById('access-token-list');
        const active = result.tokens.filter(token => !token.revokedAt);
        list.innerHTML = active.length ? active.map(token => `
            <div class="access-list-item">
                <div><strong>${escapeHTML(token.name)}</strong>
                    <small>${escapeHTML(userName(token.ownerId))} · ${token.expiresAt ? `Expires ${escapeHTML(formatDate(token.expiresAt))}` : 'Never expires'}<br>${escapeHTML(token.scopes.join(', '))}</small></div>
                <div class="access-list-actions"><button type="button" data-revoke-token="${escapeHTML(token.id)}">Revoke</button></div>
            </div>`).join('') : '<p class="access-empty">No active API tokens.</p>';
    }

    async function createToken(event) {
        event.preventDefault();
        const form = event.currentTarget;
        const days = form.elements.expires.value;
        const expiresAt = days === 'never' ? null : new Date(Date.now() + Number(days) * 86400000).toISOString();
        const scopes = checkedPermissions('access-token-scopes');
        if (!scopes.length) {
            notify('Choose at least one permission for the token.');
            return;
        }
        try {
            const result = await request('/api/v3/auth/tokens', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name: form.elements.name.value.trim(), scopes, expiresAt })
            });
            const secret = document.getElementById('access-token-secret');
            secret.querySelector('code').textContent = result.secret;
            secret.hidden = false;
            form.reset();
            setCheckedPermissions('access-token-scopes', []);
            notify('API token created.', 'success');
            await loadTokens();
        } catch (error) {
            notify(error.message);
        }
    }

    async function revokeToken(id) {
        if (!window.confirm('Revoke this API token?')) return;
        try {
            await request(`/api/v3/auth/tokens/${encodeURIComponent(id)}`, { method: 'DELETE' });
            notify('API token revoked.', 'success');
            await loadTokens();
        } catch (error) {
            notify(error.message);
        }
    }

    function setupAvailability() {
        const peopleAllowed = can('users.manage');
        document.getElementById('access-people-denied').hidden = peopleAllowed;
        document.getElementById('access-people-content').hidden = !peopleAllowed;

        const groupsAllowed = can('groups.manage');
        document.getElementById('access-groups-denied').hidden = groupsAllowed;
        document.getElementById('access-groups-content').hidden = !groupsAllowed;

        const apiAllowed = can('tokens.manage');
        document.getElementById('access-api-denied').hidden = apiAllowed;
        document.getElementById('access-api-content').hidden = !apiAllowed;
    }

    function bindEvents() {
        root.querySelectorAll('.access-nav-button').forEach(button => button.addEventListener('click', () => showPanel(button.dataset.accessPanel)));
        document.getElementById('own-password-form').addEventListener('submit', changeOwnPassword);
        document.getElementById('access-create-user').addEventListener('submit', createUser);
        document.getElementById('access-group-form').addEventListener('submit', saveGroup);
        document.getElementById('access-group-select').addEventListener('change', event => selectGroup(event.target.value));
        document.getElementById('access-preset-select').addEventListener('change', event => applyPreset(event.target.value));
        document.getElementById('access-clear-permissions').addEventListener('click', () => {
            document.getElementById('access-preset-select').value = 'custom';
            setCheckedPermissions('access-permission-grid', []);
        });
        document.getElementById('access-delete-group').addEventListener('click', deleteGroup);
        document.getElementById('access-token-form').addEventListener('submit', createToken);
        document.getElementById('access-token-secret').querySelector('button').addEventListener('click', async () => {
            const value = document.getElementById('access-token-secret').querySelector('code').textContent;
            try {
                await navigator.clipboard.writeText(value);
                notify('Token copied.', 'success');
            } catch (_) {
                notify('Could not copy the token. Select it and copy it manually.');
            }
        });

        root.addEventListener('click', event => {
            const edit = event.target.closest('[data-edit-user]');
            if (edit) {
                const editor = edit.closest('.access-user-item').querySelector('.access-user-editor');
                editor.hidden = !editor.hidden;
            }
            const deleteButton = event.target.closest('[data-delete-user]');
            if (deleteButton) deleteUser(deleteButton.closest('[data-user-id]').dataset.userId);
            const sessionButton = event.target.closest('[data-revoke-session]');
            if (sessionButton) revokeSession(sessionButton.dataset.revokeSession);
            const tokenButton = event.target.closest('[data-revoke-token]');
            if (tokenButton) revokeToken(tokenButton.dataset.revokeToken);
        });
        root.addEventListener('submit', event => {
            if (event.target.matches('.access-user-editor')) {
                event.preventDefault();
                saveUser(event.target);
            }
        });
    }

    async function initialize() {
        bindEvents();
        setupAvailability();
        try {
            await loadSession();
            if (can('groups.manage') || can('users.manage')) await loadGroups();
            if (can('users.manage')) await loadUsers();
            await loadSessions();
            if (can('tokens.manage')) {
                renderPermissionGrid('access-token-scopes', (document.body.dataset.permissions || '').split(',').filter(Boolean));
                await loadTokens();
            }
        } catch (error) {
            notify(error.message);
        }
    }

    initialize();
})();
