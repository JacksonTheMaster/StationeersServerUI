/**
 * SSUI Theme Engine
 * Manages user-customizable themes via CSS variables and localStorage.
 */

const SSUITheme = (() => {
    const STORAGE_KEY = 'ssui-theme';

    // ── Themeable variables ──
    // Each entry: { variable, label, description, group }
    const THEME_VARS = [
        // Core Colors
        { variable: '--primary',     label: 'Primary',       description: 'Main accent / brand color',         group: 'Core Colors' },
        { variable: '--bg-dark',     label: 'Background',    description: 'Page background color',             group: 'Core Colors' },
        { variable: '--bg-panel',    label: 'Panel BG',      description: 'Panel / card background',           group: 'Core Colors' },
        { variable: '--accent',      label: 'Accent',        description: 'Secondary accent (links, badges)',  group: 'Core Colors' },
        { variable: '--danger',      label: 'Danger',        description: 'Errors, destructive actions',       group: 'Core Colors' },
        { variable: '--success',     label: 'Success',       description: 'Success states',                    group: 'Core Colors' },
        { variable: '--warning',     label: 'Warning',       description: 'Warning states',                    group: 'Core Colors' },
        // Text
        { variable: '--text-header', label: 'Header Text',   description: 'Headings, bright text',             group: 'Text Colors' },
        { variable: '--text-bright', label: 'Bright Text',   description: 'Primary readable text',             group: 'Text Colors' },
        { variable: '--text-dim',    label: 'Dim Text',      description: 'Subdued text, hints',               group: 'Text Colors' },
        { variable: '--text-muted',  label: 'Muted Text',    description: 'Timestamps, least-important',       group: 'Text Colors' },
        // Surfaces
        { variable: '--surface-dark',    label: 'Surface',       description: 'Button / elevated backgrounds', group: 'Surfaces' },
        { variable: '--surface-hover',   label: 'Surface Hover', description: 'Hover state for surfaces',     group: 'Surfaces' },
        { variable: '--surface-overlay', label: 'Overlay',       description: 'Dropdowns, overlays',          group: 'Surfaces' },
        { variable: '--panel-frame',     label: 'Panel Frame',   description: 'Control panel edges',          group: 'Control Panel' },
        { variable: '--panel-surface',   label: 'Panel Surface', description: 'Control panel modules',        group: 'Control Panel' },
        { variable: '--panel-inset',     label: 'Panel Inset',   description: 'Inset console background',     group: 'Control Panel' },
        { variable: '--panel-lamp',      label: 'Status Lamp',   description: 'Connected stream indicator',   group: 'Control Panel' },
        // Console
        { variable: '--console-info',    label: 'Info',    description: 'Informational console messages',  group: 'Console Colors' },
        { variable: '--console-warning', label: 'Warning', description: 'Warning console messages',        group: 'Console Colors' },
        { variable: '--console-error',   label: 'Error',   description: 'Error console messages',          group: 'Console Colors' },
        { variable: '--console-success', label: 'Success', description: 'Success / boot complete messages',group: 'Console Colors' },
    ];

    // ── Preset Themes ──
    const PRESETS = {
        'StationeersServerUI (default)': {
            '--primary': '#00FFAB',
            '--bg-dark': '#0a0a14',
            '--bg-panel': '#1b1b2f8f',
            '--accent': '#0084ff',
            '--danger': '#ff3860',
            '--success': '#48c774',
            '--warning': '#ffdd57',
            '--text-header': '#ffffff',
            '--text-bright': '#e0ffe9',
            '--text-dim': '#aaaaaa',
            '--text-muted': '#888888',
            '--surface-dark': '#232338',
            '--surface-hover': '#333350',
            '--surface-overlay': '#1b1b2f',
            '--console-info': '#0af',
            '--console-warning': '#ff0',
            '--console-error': '#ff3333',
            '--console-success': '#0f0',
        },
        'Neon Blue': {
            '--primary': '#00D4FF',
            '--bg-dark': '#080818',
            '--bg-panel': '#12122a8f',
            '--accent': '#7B68EE',
            '--danger': '#FF4466',
            '--success': '#00E676',
            '--warning': '#FFD740',
            '--text-header': '#FFFFFF',
            '--text-bright': '#D0F0FF',
            '--text-dim': '#8899AA',
            '--text-muted': '#667788',
            '--surface-dark': '#1A1A3E',
            '--surface-hover': '#2A2A55',
            '--surface-overlay': '#15152D',
            '--console-info': '#00D4FF',
            '--console-warning': '#FFD740',
            '--console-error': '#FF4466',
            '--console-success': '#00E676',
        },
        'Hot Pink': {
            '--primary': '#FF1493',
            '--bg-dark': '#0A0008',
            '--bg-panel': '#1E0A1A8f',
            '--accent': '#9B59B6',
            '--danger': '#FF3030',
            '--success': '#00E5A0',
            '--warning': '#FFD700',
            '--text-header': '#FFFFFF',
            '--text-bright': '#FFD0E8',
            '--text-dim': '#AA7799',
            '--text-muted': '#886688',
            '--surface-dark': '#2A1025',
            '--surface-hover': '#3A1A35',
            '--surface-overlay': '#1E0A18',
            '--console-info': '#FF69B4',
            '--console-warning': '#FFD700',
            '--console-error': '#FF3030',
            '--console-success': '#00E5A0',
        },
        'Midnight Purple': {
            '--primary': '#B388FF',
            '--bg-dark': '#08061A',
            '--bg-panel': '#1A1040ef',
            '--accent': '#82B1FF',
            '--danger': '#FF5252',
            '--success': '#69F0AE',
            '--warning': '#FFD740',
            '--text-header': '#FFFFFF',
            '--text-bright': '#E0D0FF',
            '--text-dim': '#9988BB',
            '--text-muted': '#776699',
            '--surface-dark': '#1E1440',
            '--surface-hover': '#2E2060',
            '--surface-overlay': '#160E30',
            '--console-info': '#82B1FF',
            '--console-warning': '#FFD740',
            '--console-error': '#FF5252',
            '--console-success': '#69F0AE',
        },
        'Colourblind friendly': {
            '--primary': '#ffb300',
            '--bg-dark': '#121212',
            '--bg-panel': '#1e1e1e',
            '--accent': '#ffb300',
            '--danger': '#ff3b3b',
            '--success': '#66d020',
            '--warning': '#ffdd00',
            '--text-header': '#ffffff',
            '--text-bright': '#ffffff',
            '--text-dim': '#bfbfbf',
            '--text-muted': '#999999',
            '--surface-dark': '#2a2a2a',
            '--surface-hover': '#383838',
            '--surface-overlay': '#1e1e1e',
            '--console-info': '#ffb300',
            '--console-warning': '#ffdd00',
            '--console-error': '#ff3b3b',
            '--console-success': '#66d020',
        },
        'Tynningö': {
            '--primary': '#6a9955',
            '--bg-dark': '#1e1e1e',
            '--bg-panel': '#252526',
            '--accent': '#6a9955',
            '--danger': '#ff3860',
            '--success': '#6a9955',
            '--warning': '#ce9178',
            '--text-header': '#d4d4d4',
            '--text-bright': '#d4d4d4',
            '--text-dim': '#a9a9a9',
            '--text-muted': '#808080',
            '--surface-dark': '#2d2d2d',
            '--surface-hover': '#3c3c3c',
            '--surface-overlay': '#252526',
            '--console-info': '#6a9955',
            '--console-warning': '#ce9178',
            '--console-error': '#ff3860',
            '--console-success': '#6a9955',
        },
        'Dammstakärret': {
            '--primary': '#7a9a7a',
            '--bg-dark': '#121a12',
            '--bg-panel': '#1b2a1b',
            '--accent': '#7a9a7a',
            '--danger': '#ff6b6b',
            '--success': '#7a9a7a',
            '--warning': '#c9a67a',
            '--text-header': '#d9e6d9',
            '--text-bright': '#d9e6d9',
            '--text-dim': '#a3b3a3',
            '--text-muted': '#7a8a7a',
            '--surface-dark': '#243224',
            '--surface-hover': '#2e3a2e',
            '--surface-overlay': '#1b2a1b',
            '--console-info': '#7a9a7a',
            '--console-warning': '#c9a67a',
            '--console-error': '#ff6b6b',
            '--console-success': '#7a9a7a',
        },
        'Ramsö Sjöwind': {
            '--primary': '#68c1e8',
            '--bg-dark': '#1a2a38',
            '--bg-panel': '#253545',
            '--accent': '#68c1e8',
            '--danger': '#ff6b6b',
            '--success': '#68c1e8',
            '--warning': '#f0ad4e',
            '--text-header': '#e0eaf0',
            '--text-bright': '#e0eaf0',
            '--text-dim': '#b0c0d0',
            '--text-muted': '#8aa0b0',
            '--surface-dark': '#2f4055',
            '--surface-hover': '#3a4c66',
            '--surface-overlay': '#253545',
            '--console-info': '#68c1e8',
            '--console-warning': '#f0ad4e',
            '--console-error': '#ff6b6b',
            '--console-success': '#68c1e8',
        },
        'Rindö Solnedgang': {
            '--primary': '#ff9e7a',
            '--bg-dark': '#272133',
            '--bg-panel': '#332940',
            '--accent': '#ff9e7a',
            '--danger': '#ff6b6b',
            '--success': '#66d020',
            '--warning': '#ffcc66',
            '--text-header': '#f5e6ff',
            '--text-bright': '#f5e6ff',
            '--text-dim': '#d1b6e1',
            '--text-muted': '#b89ac7',
            '--surface-dark': '#3e304d',
            '--surface-hover': '#4b3a5d',
            '--surface-overlay': '#332940',
            '--console-info': '#ff9e7a',
            '--console-warning': '#ffcc66',
            '--console-error': '#ff6b6b',
            '--console-success': '#66d020',
        },
        'Mint Choklad': {
            '--primary': '#7fe0c3',
            '--bg-dark': '#1e2721',
            '--bg-panel': '#26322a',
            '--accent': '#7fe0c3',
            '--danger': '#ff6b6b',
            '--success': '#7fe0c3',
            '--warning': '#d9b382',
            '--text-header': '#e0f0e8',
            '--text-bright': '#e0f0e8',
            '--text-dim': '#b0c5b8',
            '--text-muted': '#8fa9a3',
            '--surface-dark': '#2e3d33',
            '--surface-hover': '#38493e',
            '--surface-overlay': '#26322a',
            '--console-info': '#7fe0c3',
            '--console-warning': '#d9b382',
            '--console-error': '#ff6b6b',
            '--console-success': '#7fe0c3',
        },
        'Lavendel Fält': {
            '--primary': '#b28dff',
            '--bg-dark': '#2b2440',
            '--bg-panel': '#352e4e',
            '--accent': '#b28dff',
            '--danger': '#ff6b6b',
            '--success': '#66d020',
            '--warning': '#ffad9c',
            '--text-header': '#ece8ff',
            '--text-bright': '#ece8ff',
            '--text-dim': '#c7c0e3',
            '--text-muted': '#a99cc9',
            '--surface-dark': '#3f385c',
            '--surface-hover': '#4a426a',
            '--surface-overlay': '#352e4e',
            '--console-info': '#b28dff',
            '--console-warning': '#ffad9c',
            '--console-error': '#ff6b6b',
            '--console-success': '#66d020',
        },
        'Midsommar': {
            '--primary': '#ffd700',
            '--bg-dark': '#1a1a2e',
            '--bg-panel': '#232342',
            '--accent': '#ffd700',
            '--danger': '#ff6b6b',
            '--success': '#66d020',
            '--warning': '#ff6b9d',
            '--text-header': '#fff4d6',
            '--text-bright': '#fff4d6',
            '--text-dim': '#e6d5a8',
            '--text-muted': '#ccc088',
            '--surface-dark': '#2d2d56',
            '--surface-hover': '#3a3a6a',
            '--surface-overlay': '#232342',
            '--console-info': '#ffd700',
            '--console-warning': '#ff6b9d',
            '--console-error': '#ff6b6b',
            '--console-success': '#66d020',
        },
        'Kireness': {
            '--primary': '#80DEEA',
            '--bg-dark': '#0A1215',
            '--bg-panel': '#1225308f',
            '--accent': '#4FC3F7',
            '--danger': '#EF5350',
            '--success': '#66BB6A',
            '--warning': '#FFA726',
            '--text-header': '#ECEFF1',
            '--text-bright': '#B0BEC5',
            '--text-dim': '#78909C',
            '--text-muted': '#546E7A',
            '--surface-dark': '#1A2D35',
            '--surface-hover': '#253D48',
            '--surface-overlay': '#152228',
            '--console-info': '#4FC3F7',
            '--console-warning': '#FFA726',
            '--console-error': '#EF5350',
            '--console-success': '#66BB6A',
        },        
        'Kiruna': {
            '--primary': '#3ddbd9',
            '--bg-dark': '#0d1b2a',
            '--bg-panel': '#1b263b',
            '--accent': '#3ddbd9',
            '--danger': '#ee6c4d',
            '--success': '#3ddbd9',
            '--warning': '#ee6c4d',
            '--text-header': '#e0fbfc',
            '--text-bright': '#e0fbfc',
            '--text-dim': '#98c1d9',
            '--text-muted': '#6a9ab8',
            '--surface-dark': '#253347',
            '--surface-hover': '#2f3e53',
            '--surface-overlay': '#1b263b',
            '--console-info': '#3ddbd9',
            '--console-warning': '#ee6c4d',
            '--console-error': '#ee6c4d',
            '--console-success': '#3ddbd9',
        },
        'Lingon': {
            '--primary': '#ff3864',
            '--bg-dark': '#2d1b1e',
            '--bg-panel': '#3d252a',
            '--accent': '#ff3864',
            '--danger': '#ff3864',
            '--success': '#7fe0c3',
            '--warning': '#ffa07a',
            '--text-header': '#ffe8ed',
            '--text-bright': '#ffe8ed',
            '--text-dim': '#deb8c4',
            '--text-muted': '#c49aaa',
            '--surface-dark': '#4d2f36',
            '--surface-hover': '#5d3942',
            '--surface-overlay': '#3d252a',
            '--console-info': '#ff3864',
            '--console-warning': '#ffa07a',
            '--console-error': '#ff3864',
            '--console-success': '#7fe0c3',
        },
        'Saffran': {
            '--primary': '#f4a261',
            '--bg-dark': '#2a1f15',
            '--bg-panel': '#3a2a1e',
            '--accent': '#f4a261',
            '--danger': '#e76f51',
            '--success': '#66d020',
            '--warning': '#e76f51',
            '--text-header': '#fff5e6',
            '--text-bright': '#fff5e6',
            '--text-dim': '#e6d4b8',
            '--text-muted': '#ccb8a0',
            '--surface-dark': '#4a3527',
            '--surface-hover': '#5a4030',
            '--surface-overlay': '#3a2a1e',
            '--console-info': '#f4a261',
            '--console-warning': '#e76f51',
            '--console-error': '#e76f51',
            '--console-success': '#66d020',
        },
        'Göteborg': {
            '--primary': '#ff10f0',
            '--bg-dark': '#0f0320',
            '--bg-panel': '#1a0835',
            '--accent': '#ff10f0',
            '--danger': '#ffff00',
            '--success': '#39ff14',
            '--warning': '#ffff00',
            '--text-header': '#f0e6ff',
            '--text-bright': '#f0e6ff',
            '--text-dim': '#c8b6e2',
            '--text-muted': '#a899c7',
            '--surface-dark': '#250d4a',
            '--surface-hover': '#30125f',
            '--surface-overlay': '#1a0835',
            '--console-info': '#00f0ff',
            '--console-warning': '#ffff00',
            '--console-error': '#ff10f0',
            '--console-success': '#39ff14',
        },
        'Blåbär': {
            '--primary': '#6b88ff',
            '--bg-dark': '#1a1f3a',
            '--bg-panel': '#242c4a',
            '--accent': '#6b88ff',
            '--danger': '#ff6b6b',
            '--success': '#66d020',
            '--warning': '#c77dff',
            '--text-header': '#e8eeff',
            '--text-bright': '#e8eeff',
            '--text-dim': '#b8c8e8',
            '--text-muted': '#9ab0d8',
            '--surface-dark': '#2e395a',
            '--surface-hover': '#38466a',
            '--surface-overlay': '#242c4a',
            '--console-info': '#6b88ff',
            '--console-warning': '#c77dff',
            '--console-error': '#ff6b6b',
            '--console-success': '#66d020',
        },
        'Rabarber': {
            '--primary': '#ff6b9d',
            '--bg-dark': '#1f1a1d',
            '--bg-panel': '#2d242a',
            '--accent': '#ff6b9d',
            '--danger': '#ff6b9d',
            '--success': '#7fe0c3',
            '--warning': '#ffa07a',
            '--text-header': '#ffe8f5',
            '--text-bright': '#ffe8f5',
            '--text-dim': '#e6b8d8',
            '--text-muted': '#cc9acc',
            '--surface-dark': '#3b2e37',
            '--surface-hover': '#493844',
            '--surface-overlay': '#2d242a',
            '--console-info': '#ff6b9d',
            '--console-warning': '#ffa07a',
            '--console-error': '#ff6b9d',
            '--console-success': '#7fe0c3',
        },
    };

    /*Get the current saved theme from localStorage, or null if none.*/
    function getSavedTheme() {
        try {
            const raw = localStorage.getItem(STORAGE_KEY);
            return raw ? JSON.parse(raw) : null;
        } catch {
            return null;
        }
    }

    /*Save a theme object to localStorage.*/
    function saveTheme(themeObj) {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(themeObj));
    }

    /* Remove the saved theme (revert to CSS defaults).*/
    function clearTheme() {
        localStorage.removeItem(STORAGE_KEY);
        // Remove all overrides from :root
        const root = document.documentElement;
        THEME_VARS.forEach(v => root.style.removeProperty(v.variable));
    }

    /* Apply a theme object to the document (sets CSS custom properties on :root).*/
    function applyTheme(themeObj) {
        if (!themeObj) return;
        const root = document.documentElement;
        // Reset known overrides first so older presets/custom themes naturally
        // inherit defaults for variables introduced in newer SSUI versions.
        THEME_VARS.forEach(v => root.style.removeProperty(v.variable));
        Object.entries(themeObj).forEach(([varName, value]) => {
            if (value) root.style.setProperty(varName, value);
        });
    }

    /* Read the current computed values of all theme variables.*/
    function getCurrentTheme() {
        const computed = getComputedStyle(document.documentElement);
        const theme = {};
        THEME_VARS.forEach(v => {
            // Try inline override first, then computed
            const inline = document.documentElement.style.getPropertyValue(v.variable).trim();
            const comp = computed.getPropertyValue(v.variable).trim();
            theme[v.variable] = inline || comp;
        });
        return theme;
    }

    /* Load and apply saved theme on page load.*/
    function init() {
        const saved = getSavedTheme();
        if (saved) applyTheme(saved);
    }

    /* Export the current theme as a JSON string (for sharing).*/

    function exportTheme() {
        const saved = getSavedTheme();
        return JSON.stringify(saved || getCurrentTheme(), null, 2);
    }

    /* Import a theme from a JSON string.*/
    function importTheme(jsonStr) {
        try {
            const theme = JSON.parse(jsonStr);
            if (typeof theme !== 'object' || theme === null) throw new Error('Invalid theme');
            applyTheme(theme);
            saveTheme(theme);
            return true;
        } catch (e) {
            console.error('Theme import failed:', e);
            return false;
        }
    }

    // Auto-init on load
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

    return {
        THEME_VARS,
        PRESETS,
        getSavedTheme,
        saveTheme,
        clearTheme,
        applyTheme,
        getCurrentTheme,
        exportTheme,
        importTheme,
        init,
    };
})();
