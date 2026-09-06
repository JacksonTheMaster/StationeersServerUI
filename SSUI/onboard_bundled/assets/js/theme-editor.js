/**
 * SSUI Theme Editor
 * Renders the theme editing UI inside the theme config tab.
 * Depends on SSUITheme (theme-engine.js) being loaded first.
 */

(function () {
    'use strict';

    function renderThemeEditor() {
        const container = document.getElementById('theme-editor-root');
        if (!container) return;

        const currentTheme = SSUITheme.getCurrentTheme();

        // Group variables by their group
        const groups = {};
        SSUITheme.THEME_VARS.forEach(v => {
            if (!groups[v.group]) groups[v.group] = [];
            groups[v.group].push(v);
        });

        let html = '';

        // ── Preset Buttons ──
        html += '<div class="theme-editor-section">';
        html += '<h3>Theme Presets</h3>';
        html += '<div class="theme-presets">';
        Object.entries(SSUITheme.PRESETS).forEach(([name, preset]) => {
            const swatches = [preset['--primary'], preset['--bg-dark'], preset['--accent'], preset['--danger'], preset['--success']];
            html += `<button class="theme-preset-btn" onclick="applyPresetTheme('${name}')" title="${name}">`;
            html += `<div class="theme-preset-colors">`;
            swatches.forEach(c => {
                html += `<span class="theme-preset-swatch" style="background:${c}"></span>`;
            });
            html += `</div>`;
            html += `<span>${name}</span>`;
            html += `</button>`;
        });
        html += '</div>';
        html += '</div>';

        // ── Color Editors per Group ──
        Object.entries(groups).forEach(([groupName, vars]) => {
            html += `<div class="theme-editor-section">`;
            html += `<h3>${groupName}</h3>`;
            html += `<div class="theme-color-grid">`;
            vars.forEach(v => {
                const val = currentTheme[v.variable] || '#000000';
                // Normalize to hex for the color picker
                const hexVal = toHex(val);
                html += `<div class="theme-color-item">`;
                html += `  <label>${v.label}</label>`;
                html += `  <div class="color-input-wrapper">`;
                html += `    <input type="color" value="${hexVal}" data-var="${v.variable}" onchange="onThemeColorChange(this)" />`;
                html += `    <input type="text" class="color-hex-input" value="${hexVal}" data-var="${v.variable}" onchange="onThemeHexChange(this)" placeholder="#RRGGBB" />`;
                html += `  </div>`;
                html += `  <span class="color-description">${v.description}</span>`;
                html += `</div>`;
            });
            html += `</div>`;
            html += `</div>`;
        });

        // ── Preview Strip ──
        html += `<div class="theme-editor-section">`;
        html += `<h3>Live Preview</h3>`;
        html += `<div class="theme-preview-strip" id="theme-preview-strip">`;
        SSUITheme.THEME_VARS.filter(v => v.group !== 'Console Colors').forEach(v => {
            html += `<div class="preview-swatch-group">`;
            html += `  <div class="preview-swatch" style="background: var(${v.variable})" title="${v.label}"></div>`;
            html += `  <span class="preview-swatch-label">${v.label}</span>`;
            html += `</div>`;
        });
        html += `</div>`;
        html += `</div>`;

        // ── Actions ──
        html += `<div class="theme-editor-section">`;
        html += `<h3>Actions</h3>`;
        html += `<div class="theme-actions">`;
        html += `  <button onclick="saveCurrentTheme()">💾 Save Theme</button>`;
        html += `  <button onclick="resetToDefaults()">↺ Reset to Defaults</button>`;
        html += `  <button onclick="exportCurrentTheme()">📤 Export Theme</button>`;
        html += `  <button onclick="showImportTheme()">📥 Import Theme</button>`;
        html += `</div>`;

        // Import area (hidden by default)
        html += `<div id="theme-import-area" style="display:none; margin-top:15px;">`;
        html += `  <div class="theme-import-export">`;
        html += `    <textarea id="theme-import-text" placeholder="Paste theme JSON here..."></textarea>`;
        html += `  </div>`;
        html += `  <div class="theme-actions" style="margin-top: 10px;">`;
        html += `    <button onclick="doImportTheme()">✓ Apply Imported Theme</button>`;
        html += `    <button onclick="hideImportTheme()">✕ Cancel</button>`;
        html += `  </div>`;
        html += `</div>`;

        // Export area (hidden by default)
        html += `<div id="theme-export-area" style="display:none; margin-top:15px;">`;
        html += `  <div class="theme-import-export">`;
        html += `    <textarea id="theme-export-text" readonly></textarea>`;
        html += `  </div>`;
        html += `  <div class="theme-actions" style="margin-top: 10px;">`;
        html += `    <button onclick="copyExportedTheme()">📋 Copy to Clipboard</button>`;
        html += `    <button onclick="hideExportTheme()">✕ Close</button>`;
        html += `  </div>`;
        html += `</div>`;

        html += `</div>`;

        container.innerHTML = html;
    }

    // ── Color conversion helpers ──

    function toHex(colorStr) {
        if (!colorStr) return '#000000';
        colorStr = colorStr.trim();

        // Already hex
        if (/^#[0-9a-f]{6}$/i.test(colorStr)) return colorStr;
        if (/^#[0-9a-f]{3}$/i.test(colorStr)) {
            return '#' + colorStr[1] + colorStr[1] + colorStr[2] + colorStr[2] + colorStr[3] + colorStr[3];
        }
        // 8-char hex with alpha - just take RGB
        if (/^#[0-9a-f]{8}$/i.test(colorStr)) return colorStr.slice(0, 7);

        // rgb/rgba
        const rgbMatch = colorStr.match(/rgba?\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)/);
        if (rgbMatch) {
            const r = parseInt(rgbMatch[1]).toString(16).padStart(2, '0');
            const g = parseInt(rgbMatch[2]).toString(16).padStart(2, '0');
            const b = parseInt(rgbMatch[3]).toString(16).padStart(2, '0');
            return `#${r}${g}${b}`;
        }

        // Modern browsers may serialize color-mix() results as color(srgb ...).
        const srgbMatch = colorStr.match(/color\(srgb\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)/i);
        if (srgbMatch) {
            const channels = srgbMatch.slice(1, 4).map(channel =>
                Math.max(0, Math.min(255, Math.round(parseFloat(channel) * 255)))
                    .toString(16).padStart(2, '0')
            );
            return `#${channels.join('')}`;
        }

        // Resolve color-mix(), var(), and other supported CSS colors in document context.
        try {
            const resolver = document.createElement('span');
            resolver.style.color = colorStr;
            resolver.style.display = 'none';
            document.body.appendChild(resolver);
            const resolved = getComputedStyle(resolver).color;
            resolver.remove();
            if (resolved && resolved !== colorStr) return toHex(resolved);
        } catch {
            // Fall through to the canvas/named-color resolver.
        }

        // Named colors - use a canvas to resolve
        try {
            const ctx = document.createElement('canvas').getContext('2d');
            ctx.fillStyle = colorStr;
            return ctx.fillStyle; // returns hex
        } catch {
            return '#000000';
        }
    }

    // ── Event Handlers (global scope for inline onclick) ──

    window.onThemeColorChange = function (input) {
        const varName = input.dataset.var;
        const value = input.value;
        document.documentElement.style.setProperty(varName, value);
        // Sync hex text input
        const hexInput = input.parentElement.querySelector('.color-hex-input');
        if (hexInput) hexInput.value = value;
    };

    window.onThemeHexChange = function (input) {
        const varName = input.dataset.var;
        let value = input.value.trim();
        // Validate hex
        if (/^#[0-9a-f]{3,8}$/i.test(value)) {
            document.documentElement.style.setProperty(varName, value);
            // Sync color picker
            const colorInput = input.parentElement.querySelector('input[type="color"]');
            if (colorInput) colorInput.value = toHex(value);
        }
    };

    window.applyPresetTheme = function (presetName) {
        const preset = SSUITheme.PRESETS[presetName];
        if (!preset) return;
        SSUITheme.applyTheme(preset);
        SSUITheme.saveTheme(preset);
        // Re-render to update inputs
        renderThemeEditor();
        showThemeNotification('Theme "' + presetName + '" applied and saved!', 'success');
    };

    window.saveCurrentTheme = function () {
        const theme = SSUITheme.getCurrentTheme();
        SSUITheme.saveTheme(theme);
        showThemeNotification('Theme saved to browser storage!', 'success');
    };

    window.resetToDefaults = function () {
        SSUITheme.clearTheme();
        renderThemeEditor();
        showThemeNotification('Theme reset to defaults.', 'info');
    };

    window.exportCurrentTheme = function () {
        const area = document.getElementById('theme-export-area');
        const text = document.getElementById('theme-export-text');
        if (area && text) {
            text.value = SSUITheme.exportTheme();
            area.style.display = 'block';
        }
        // hide import if open
        const imp = document.getElementById('theme-import-area');
        if (imp) imp.style.display = 'none';
    };

    window.hideExportTheme = function () {
        const area = document.getElementById('theme-export-area');
        if (area) area.style.display = 'none';
    };

    window.copyExportedTheme = function () {
        const text = document.getElementById('theme-export-text');
        if (text) {
            navigator.clipboard.writeText(text.value).then(() => {
                showThemeNotification('Theme JSON copied to clipboard!', 'success');
            }).catch(() => {
                // Fallback
                showThemeNotification('Theme copy failed!', 'error');
            });
        }
    };

    window.showImportTheme = function () {
        const area = document.getElementById('theme-import-area');
        if (area) area.style.display = 'block';
        // hide export if open
        const exp = document.getElementById('theme-export-area');
        if (exp) exp.style.display = 'none';
    };

    window.hideImportTheme = function () {
        const area = document.getElementById('theme-import-area');
        if (area) area.style.display = 'none';
    };

    window.doImportTheme = function () {
        const text = document.getElementById('theme-import-text');
        if (!text || !text.value.trim()) {
            showThemeNotification('Please paste a theme JSON first.', 'error');
            return;
        }
        if (SSUITheme.importTheme(text.value.trim())) {
            renderThemeEditor();
            showThemeNotification('Theme imported and applied!', 'success');
        } else {
            showThemeNotification('Invalid theme JSON. Please check the format.', 'error');
        }
    };

    function showThemeNotification(message, type) {
        // Reuse existing notification element or create a temporary one
        let notif = document.getElementById('theme-notification');
        if (!notif) {
            notif = document.createElement('div');
            notif.id = 'theme-notification';
            notif.className = 'notification';
            notif.style.position = 'fixed';
            notif.style.top = '20px';
            notif.style.right = '20px';
            notif.style.zIndex = '9999';
            notif.style.maxWidth = '350px';
            document.body.appendChild(notif);
        }
        notif.className = 'notification notification-' + type;
        notif.textContent = message;
        notif.style.display = 'block';
        clearTimeout(notif._timer);
        notif._timer = setTimeout(() => {
            notif.style.display = 'none';
        }, 3000);
    }

    // Initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', renderThemeEditor);
    } else {
        renderThemeEditor();
    }
})();
