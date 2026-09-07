// Current known update info (populated from polling)
let currentUpdateVersion = null;

// Poll for update status every 60 seconds
function pollUpdateStatus() {
    if (!window.SSUIAccess.can('server.view')) return;
    fetch('/api/v3/update')
        .then(response => response.json())
        .then(data => {
            if (data.updateAvailable === "true" && data.version) {
                // Update available!
                currentUpdateVersion = data.version;

                // Show update button with bounce animation
                const updateBtn = document.getElementById('update-button');
                updateBtn.style.display = 'block';
                updateBtn.classList.add('bounce');

                // Update modal text when opened
                document.getElementById('modal-version-text').textContent = data.version;
            } else {
                // No update - hide button and reset bounce
                document.getElementById('update-button').style.display = 'none';
                document.getElementById('update-button').classList.remove('bounce');
                currentUpdateVersion = null;
            }
        })
        .catch(err => {
            console.warn('Failed to check for updates:', err);
        });
}

// Open update modal
function openUpdateModal() {
    if (currentUpdateVersion) {
        document.getElementById('modal-version-text').textContent = currentUpdateVersion;
    }
    document.getElementById('update-modal').classList.add('show');

    // Reset state in case it was left in "running"
    document.getElementById('update-status-running').classList.remove('running');
    document.getElementById('update-now-btn').style.display = '';
    document.getElementById('update-later-btn').style.display = '';
}

// Close update modal
function closeUpdateModal() {
    document.getElementById('update-modal').classList.remove('show');
}

// Start the actual update
function startUpdate() {
    if (!window.SSUIAccess.require('update.install', "You don't have permission to update SSUI.")) return;
    // Hide buttons
    document.getElementById('update-now-btn').style.display = 'none';
    document.getElementById('update-later-btn').style.display = 'none';

    // Show running status
    document.getElementById('update-status-running').classList.add('running');

    // Send request to trigger update
    fetch('/api/v3/update', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ allowUpdate: true })
    })
        .then(response => response.json())
        .then(data => {
            console.log('Update triggered:', data);
            // Even if success/fail, we assume update is running
        })
        .catch(err => {
            console.error('Failed to trigger update:', err);
            // Optional: show failure message
            document.getElementById('update-status-failed').classList.add('running');
        });

    // Auto-refresh after 40 seconds (gives time for update to download/apply)
    setTimeout(() => {
        location.reload();
    }, 40000);
}

// Close modal when clicking outside
document.getElementById('update-modal').addEventListener('click', function (e) {
    if (e.target === this) {
        closeUpdateModal();
    }
});

// Start polling when page loads
document.addEventListener('DOMContentLoaded', () => {
    pollUpdateStatus();           // Immediate check
    setInterval(pollUpdateStatus, 60000); // Every minute after
});
