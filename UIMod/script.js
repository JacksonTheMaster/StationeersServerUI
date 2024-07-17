function startServer() {
    fetch('/start')
        .then(response => response.text())
        .then(data => document.getElementById('status').innerText = data);
}

function stopServer() {
    fetch('/stop')
        .then(response => response.text())
        .then(data => document.getElementById('status').innerText = data);
}

function fetchOutput() {
    const eventSource = new EventSource('/output');
    eventSource.onmessage = function(event) {
        const consoleElement = document.getElementById('console');
        const message = document.createElement('div');
        message.textContent = event.data;
        consoleElement.appendChild(message);
        consoleElement.scrollTop = consoleElement.scrollHeight;
    };
}

function fetchBackups() {
    fetch('/backups')
        .then(response => response.text())
        .then(data => {
            const backupList = document.getElementById('backupList');
            backupList.innerHTML = ''; // Clear existing items
            if (data.trim() === "No valid backup files found.") {
                backupList.innerHTML = data;
            } else {
                const backups = data.split('\n');
                backups.forEach(backup => {
                    if (backup.trim()) {
                        const listItem = document.createElement('li');
                        listItem.classList.add('backup-item');
                        listItem.innerHTML = backup + ' <button onclick="restoreBackup(' + extractIndex(backup) + ')">Restore</button>';
                        backupList.appendChild(listItem);
                    }
                });
            }
        });
}

function extractIndex(backupText) {
    const match = backupText.match(/Index: (\d+)/);
    return match ? match[1] : null;
}

function restoreBackup(index) {
    fetch(`/restore?index=${index}`)
        .then(response => response.text())
        .then(data => document.getElementById('status').innerText = data);
}

fetchOutput();
fetchBackups();
