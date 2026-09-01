function reloadConfigTab(tab = 'slp') {
    const url = new URL(window.location.href);
    url.searchParams.set('tab', tab);
    window.location.assign(url.toString());
}

function showNotification(message, type = 'info') {
    const notification = document.getElementById('notification');
    notification.textContent = message;
    notification.className = 'notification ' + type;
    notification.style.display = 'block';
    
    // Auto-hide after 5 seconds for success, keep longer for errors
    const duration = type === 'success' ? 5000 : 7000;
    setTimeout(() => {
        notification.style.display = 'none';
    }, duration);
}

function setButtonLoading(buttonId, isLoading) {
    const button = document.getElementById(buttonId);
    if (!button) return;
    
    if (isLoading) {
        button.disabled = true;
        button.dataset.originalText = button.textContent;
        button.textContent = button.classList.contains('mod-update-button') ? '⏳' : '⏳ Please wait...';
        button.classList.add('loading');
    } else {
        button.disabled = false;
        button.textContent = button.dataset.originalText || button.textContent;
        button.classList.remove('loading');
    }
}

function installSLP() {
    setButtonLoading('installSLPBtn', true);
    showPopup('info', 'Installing Stationeers Launch Pad...');
    
    fetch('/api/v2/slp/install')
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                showPopup('success', 'Stationeers Launch Pad installed successfully! The page will refresh automatically.');
                setButtonLoading('installSLPBtn', false);
                setTimeout(() => reloadConfigTab('slp'), 1200);
            } else {
                showPopup('error', 'Failed to install SLP:\n\n' + (data.error || 'Unknown error'));
                setButtonLoading('installSLPBtn', false);
            }
        })
        .catch(error => {
            showPopup('error', 'Failed to install SLP:\n\n' + (error.message || 'Network error'));
            setButtonLoading('installSLPBtn', false);
        });
}

function uninstallSLP() {
    if (!confirm('Are you sure you want to uninstall SLP? This will DELETE all mods too (you can always reinstall later).')) {
        return;
    }
    setButtonLoading('uninstallSLPBtn', true);
    showPopup('info', 'Uninstalling Stationeers Launch Pad...');

    fetch('/api/v2/slp/uninstall')
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                showPopup('success', 'Stationeers Launch Pad uninstalled successfully! The page will refresh automatically.');
                setButtonLoading('uninstallSLPBtn', false);
                setTimeout(() => reloadConfigTab('slp'), 1200);
            } else {
                showPopup('error', 'Failed to uninstall SLP:\n\n' + (data.error || 'Unknown error'));
                setButtonLoading('uninstallSLPBtn', false);
            }
        })
        .catch(error => {
            showPopup('error', 'Failed to uninstall SLP:\n\n' + (error.message || 'Network error'));
            setButtonLoading('uninstallSLPBtn', false);
        });
}

function reinstallSLP() {
    if (!confirm('Are you sure you want to reinstall SLP? This will re-download and reinstall the SLP plugin. Your mods and modconfig will be preserved.')) {
        return;
    }
    setButtonLoading('reinstallSLPBtn', true);
    showPopup('info', 'Reinstalling Stationeers Launch Pad...\n\nThis will re-download the latest version while keeping your mods intact.');

    fetch('/api/v2/slp/reinstall')
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                showPopup('success', 'Stationeers Launch Pad reinstalled successfully!' + (data.version ? ' (Version: ' + data.version + ')' : '') + ' The page will refresh automatically.');
                setButtonLoading('reinstallSLPBtn', false);
                setTimeout(() => reloadConfigTab('slp'), 1200);
            } else {
                showPopup('error', 'Failed to reinstall SLP:\n\n' + (data.error || 'Unknown error'));
                setButtonLoading('reinstallSLPBtn', false);
            }
        })
        .catch(error => {
            showPopup('error', 'Failed to reinstall SLP:\n\n' + (error.message || 'Network error'));
            setButtonLoading('reinstallSLPBtn', false);
        });
}

function updateSingleMod(workshopHandle, index) {
    const btnId = 'update-mod-btn-' + index;
    setButtonLoading(btnId, true);
    showPopup('info', 'Updating workshop mod ' + workshopHandle + '...\n\nPlease wait.');

    fetch('/api/v2/steamcmd/updatemod', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ workshopHandle: workshopHandle })
    })
        .then(response => response.json())
        .then(data => {
            setButtonLoading(btnId, false);
            if (data.success) {
                showPopup('success', 'Workshop mod updated successfully!\n\nReloading mod list...');
                loadInstalledMods(true);
            } else {
                showPopup('error', 'Failed to update mod:\n\n' + (data.error || 'Unknown error'));
            }
        })
        .catch(error => {
            showPopup('error', 'Failed to update mod:\n\n' + (error.message || 'Network error'));
            setButtonLoading(btnId, false);
        });
}

function updateWorkshopMods() {
    setButtonLoading('updateWorkshopModsBtn', true);
    showPopup('info', 'Updating workshop mods...\n\nThis may take some time depending on the number of mods. Please wait.');
    
    fetch('/api/v2/steamcmd/updatemods')
        .then(response => response.json())
        .then(data => {
            setButtonLoading('updateWorkshopModsBtn', false);
            if (data.success) {
                showPopup('success', 'Workshop mods updated successfully.');
                loadInstalledMods(true);
            } else {
                showPopup('error', 'Failed to update workshop mods:\n\n' + (data.error || 'Unknown error'));
            }
        })
        .catch(error => {
            showPopup('error', 'Failed to update workshop mods:\n\n' + (error.message || 'Network error'));
            setButtonLoading('updateWorkshopModsBtn', false);
        });
}

function installWorkshopMods() {
    const input = document.getElementById('workshopModInput');
    if (!input || !input.value.trim()) {
        showPopup('error', 'Paste at least one Steam Workshop URL or ID.');
        return;
    }

    const workshopHandles = input.value
        .split(/[\s,;]+/)
        .map(value => value.trim())
        .filter(Boolean);

    setButtonLoading('installWorkshopModsBtn', true);
    showPopup('info', 'Downloading and installing ' + workshopHandles.length + ' workshop item(s)...\n\nThis can take a while.');

    fetch('/api/v2/steamcmd/updatemod', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ workshopHandles })
    })
        .then(response => response.json())
        .then(data => {
            setButtonLoading('installWorkshopModsBtn', false);
            if (data.success) {
                input.value = '';
                showPopup('success', 'Workshop mods installed successfully.');
                loadInstalledMods(true);
            } else {
                showPopup('error', 'Failed to install workshop mods:\n\n' + (data.error || 'Unknown error'));
            }
        })
        .catch(error => {
            setButtonLoading('installWorkshopModsBtn', false);
            showPopup('error', 'Failed to install workshop mods:\n\n' + (error.message || 'Network error'));
        });
}

let selectedModFile = null;

function handleModPackageSelection(files) {
    if (!files || files.length === 0) {
        return;
    }

    const file = files[0];
    
    // Validate file is a zip
    if (!file.name.endsWith('.zip')) {
        showPopup('error', 'Invalid file format\n\nPlease select a .zip file');
        document.getElementById('modPackageUpload').value = '';
        selectedModFile = null;
        updateFileDisplay();
        return;
    }

    // Validate filename starts with modpkg_
    if (!file.name.startsWith('modpkg_')) {
        showPopup('error', 'Invalid mod package name: Mod package filename must start with "modpkg_" Example: modpkg_2026-01-09_12-33-01-670.zip');
        document.getElementById('modPackageUpload').value = '';
        selectedModFile = null;
        updateFileDisplay();
        return;
    }

    // Check file size (limit to 500MB)
    const maxSize = 500 * 1024 * 1024;
    if (file.size > maxSize) {
        showPopup('error', 'File too large: Maximum file size is 500MB');
        document.getElementById('modPackageUpload').value = '';
        selectedModFile = null;
        updateFileDisplay();
        return;
    }

    // Store the file for later upload
    selectedModFile = file;
    updateFileDisplay();
    showNotification('✓ File selected: ' + file.name + ' (' + (file.size / 1024 / 1024).toFixed(2) + 'MB)', 'success');
}

function updateFileDisplay() {
    const uploadZone = document.getElementById('modPackageUploadZone');
    const uploadBtn = document.getElementById('uploadModPackageBtn');
    
    if (!uploadZone || !uploadBtn) return;
    
    if (selectedModFile) {
        uploadZone.innerHTML = '<span class="upload-icon">✓</span>' +
                              '<div class="upload-text">File Selected</div>' +
                              '<div class="upload-subtext">' + selectedModFile.name + '</div>' +
                              '<div class="upload-subtext" style="margin-top: 5px; color: var(--accent);">' + (selectedModFile.size / 1024 / 1024).toFixed(2) + ' MB</div>';
        uploadBtn.disabled = false;
        uploadBtn.style.opacity = '1';
    } else {
        uploadZone.innerHTML = '<span class="upload-icon">📦</span>' +
                              '<div class="upload-text">Drag & Drop Mod Package Here</div>' +
                              '<div class="upload-subtext">or click to select a .zip file</div>';
        uploadBtn.disabled = true;
        uploadBtn.style.opacity = '0.5';
    }
}

function uploadModPackage() {
    if (!selectedModFile) {
        showPopup('error', 'No file selected');
        return;
    }

    setButtonLoading('uploadModPackageBtn', true);
    showPopup('info', 'Uploading mod package...\n\n' + selectedModFile.name + ' (' + (selectedModFile.size / 1024 / 1024).toFixed(2) + 'MB)');
    updateUploadProgress(0);

    const reader = new FileReader();
    reader.onload = function(e) {
        const zipData = e.target.result;
        
        fetch('/api/v2/slp/upload', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/zip'
            },
            body: zipData
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                showPopup('success', 'Mod package uploaded successfully!\n\n' + (data.message || 'The mods have been extracted and are ready to use.'));
                selectedModFile = null;
                document.getElementById('modPackageUpload').value = '';
                updateFileDisplay();
                updateUploadProgress(100);
                loadInstalledMods(true);
                setTimeout(() => updateUploadProgress(0), 2000);
                setButtonLoading('uploadModPackageBtn', false);
            } else {
                showPopup('error', 'Failed to upload mod package:\n\n' + (data.error || 'Unknown error'));
                setButtonLoading('uploadModPackageBtn', false);
            }
        })
        .catch(error => {
            showPopup('error', 'Upload failed:\n\n' + (error.message || 'Network error'));
            setButtonLoading('uploadModPackageBtn', false);
        });
    };
    
    reader.onerror = function() {
        showPopup('error', 'Failed to read file');
        setButtonLoading('uploadModPackageBtn', false);
    };
    
    reader.readAsArrayBuffer(selectedModFile);
}

function updateUploadProgress(percent) {
    const progressBar = document.getElementById('modPackageUploadProgress');
    if (progressBar) {
        progressBar.style.width = percent + '%';
        if (percent > 0) {
            progressBar.classList.add('active');
        } else {
            progressBar.classList.remove('active');
        }
    }
}

// Drag and drop support
function initializeDragAndDrop() {
    const uploadZone = document.getElementById('modPackageUploadZone');
    if (!uploadZone) return;

    // Click to open file selector
    uploadZone.addEventListener('click', function() {
        document.getElementById('modPackageUpload').click();
    });

    // Drag and drop events
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        uploadZone.addEventListener(eventName, preventDefaults, false);
    });

    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    ['dragenter', 'dragover'].forEach(eventName => {
        uploadZone.addEventListener(eventName, highlight, false);
    });

    ['dragleave', 'drop'].forEach(eventName => {
        uploadZone.addEventListener(eventName, unhighlight, false);
    });

    function highlight(e) {
        uploadZone.classList.add('highlight');
    }

    function unhighlight(e) {
        uploadZone.classList.remove('highlight');
    }

    uploadZone.addEventListener('drop', handleDrop, false);

    function handleDrop(e) {
        const dt = e.dataTransfer;
        const files = dt.files;
        handleModPackageSelection(files);
    }
    
    // Initialize file display on load
    updateFileDisplay();
}

// Initialize on page load
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeDragAndDrop);
} else {
    initializeDragAndDrop();
}

// Mods List Management
let modsData = [];

function loadInstalledMods(reloadOnFailure = false) {
    const container = document.getElementById('mods-list-container');
    const loader = document.getElementById('mods-loader');
    const modsList = document.getElementById('mods-list');
    
    if (!container) return;
    
    // Show loader
    if (loader) loader.style.display = 'block';
    if (modsList) modsList.innerHTML = '';
    
    fetch('/api/v2/slp/mods')
        .then(response => response.json())
        .then(data => {
            if (loader) loader.style.display = 'none';
            
            if (data.success && data.mods && data.mods.length > 0) {
                modsData = data.mods;
                renderModsList(data.mods);
            } else {
                if (modsList) modsList.innerHTML = '<div class="mods-empty">No mods installed yet. Upload a mod package to get started!</div>';
            }
        })
        .catch(error => {
            if (loader) loader.style.display = 'none';
            console.error('Failed to load mods:', error);
            if (modsList) modsList.innerHTML = '<div class="mods-empty">Failed to load mods. Check console for details.</div>';
            if (reloadOnFailure) {
                reloadConfigTab('slp');
            }
        });
}

function renderModsList(mods) {
    const modsList = document.getElementById('mods-list');
    if (!modsList) return;
    
    modsList.innerHTML = '';
    
    if (!mods || mods.length === 0) {
        modsList.innerHTML = '<div class="mods-empty">No mods installed yet. Upload a mod package to get started!</div>';
        return;
    }
    
    mods.forEach((mod, index) => {
        const modCard = createModCard(mod, index);
        modsList.appendChild(modCard);
    });
}

function createModCard(mod, index) {
    const card = document.createElement('div');
    card.className = 'mod-card';
    
    const images = mod.Images || {};
    const imageArray = Object.entries(images);
    
    let imageHtml = '';
    if (imageArray.length > 0) {
        const firstImageData = imageArray[0][1];
        imageHtml = `
            <div class="mod-image-container" data-mod-index="${index}">
                <img id="mod-image-${index}" src="data:image/png;base64,${firstImageData}" alt="Mod image">
            </div>
        `;
    } else {
        imageHtml = `
            <div class="mod-image-container no-image">
                📷
            </div>
        `;
    }
    
    const parsedDescription = mod.Description ? parseSteamMarkup(mod.Description) : '';
    
    let descriptionHtml = '';
    if (parsedDescription) {
        descriptionHtml = `
            <div class="mod-description" id="desc-${index}">
                ${parsedDescription}
            </div>
            <button class="mod-expand-button" id="expand-btn-${index}" onclick="toggleModDescription(${index})">▼</button>
        `;
    }
    
    card.innerHTML = `
        ${imageHtml}
        <div class="mod-title">${escapeHtml(mod.Name || 'Unknown Mod')}</div>
        ${mod.Author ? `<div class="mod-author">By ${escapeHtml(mod.Author)}</div>` : ''}
        ${mod.Version ? `<div class="mod-version">v${escapeHtml(mod.Version)}</div>` : ''}
        ${descriptionHtml}
    `;

    if (mod.WorkshopHandle) {
        const actions = document.createElement('div');
        actions.className = 'mod-card-actions';

        const workshopLink = document.createElement('a');
        workshopLink.className = 'mod-workshop-link';
        workshopLink.href = 'https://steamcommunity.com/sharedfiles/filedetails/?id=' + encodeURIComponent(mod.WorkshopHandle);
        workshopLink.target = '_blank';
        workshopLink.rel = 'noopener noreferrer';
        workshopLink.textContent = '↗ Workshop';

        const updateButton = document.createElement('button');
        updateButton.className = 'mod-update-button';
        updateButton.id = `update-mod-btn-${index}`;
        updateButton.textContent = '🔄 Update';
        updateButton.addEventListener('click', () => updateSingleMod(mod.WorkshopHandle, index));

        actions.appendChild(workshopLink);
        actions.appendChild(updateButton);
        card.appendChild(actions);
    }
    
    return card;
}

function parseSteamMarkup(text) {
    if (!text) return '';
    
    // Escape HTML first
    let html = text
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;');
    
    // Replace Steam markup tags
    // Headers
    html = html.replace(/\[h1\](.*?)\[\/h1\]/g, '<h1>$1</h1>');
    html = html.replace(/\[h2\](.*?)\[\/h2\]/g, '<h2>$1</h2>');
    html = html.replace(/\[h3\](.*?)\[\/h3\]/g, '<h3>$1</h3>');
    html = html.replace(/\[h4\](.*?)\[\/h4\]/g, '<h4>$1</h4>');
    
    // Bold, Italic, Underline
    html = html.replace(/\[b\](.*?)\[\/b\]/g, '<b>$1</b>');
    html = html.replace(/\[i\](.*?)\[\/i\]/g, '<i>$1</i>');
    html = html.replace(/\[u\](.*?)\[\/u\]/g, '<u>$1</u>');
    
    // Horizontal rule
    html = html.replace(/\[hr\]/g, '<hr>');
    
    // Links
    html = html.replace(/\[url=(.*?)\](.*?)\[\/url\]/g, '<a href="$1" target="_blank">$2</a>');
    
    // Lists - handle [list] ... [/list] with [*] items
    html = html.replace(/\[list\]([\s\S]*?)\[\/list\]/g, function(match, content) {
        const items = content.split(/\[\*\]/).filter(item => item.trim());
        const listItems = items.map(item => '<li>' + item.trim() + '</li>').join('');
        return '<ul>' + listItems + '</ul>';
    });
    
    // Remove image tags (we don't need external images)
    html = html.replace(/\[img\].*?\[\/img\]/g, '');
    
    // Handle line breaks - replace multiple spaces/newlines with actual line breaks
    html = html.replace(/\n\s*\n/g, '<br><br>');
    html = html.replace(/\n/g, '<br>');
    
    return html;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function toggleModDescription(index) {
    const descElement = document.getElementById(`desc-${index}`);
    const btnElement = document.getElementById(`expand-btn-${index}`);
    
    if (descElement && btnElement) {
        descElement.classList.toggle('expanded');
        btnElement.classList.toggle('expanded');
    }
}

// Initialize mods list when mods section is visible
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function() {
        const modsContainer = document.getElementById('mods-list-container');
        if (modsContainer) {
            loadInstalledMods();
        }
    });
} else {
    const modsContainer = document.getElementById('mods-list-container');
    if (modsContainer) {
        loadInstalledMods();
    }
}
