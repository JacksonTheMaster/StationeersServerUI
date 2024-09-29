document.addEventListener('DOMContentLoaded', () => {
    typeText(document.querySelector('h1'), 60);  // Type out the h1 on load
    setDefaultConsoleMessage();
});

function typeText(element, speed) {
    const fullText = element.innerHTML;
    element.innerHTML = '';  // Clear the text initially
    let i = 0;

    function typeChar() {
        if (i < fullText.length) {
            element.innerHTML += fullText.charAt(i);
            i++;
            setTimeout(typeChar, speed);  // Adjust speed for typing
        }
    }

    typeChar();
}

function typeTextWithCallback(element, text, speed, callback) {
    element.innerHTML = '';  // Clear the text initially
    let i = 0;

    function typeChar() {
        if (i < text.length) {
            element.innerHTML += text.charAt(i);
            i++;
            setTimeout(typeChar, speed);
        } else if (callback) {
            callback();
        }
    }

    typeChar();
}

function startServer() {
    fetch('/start')
        .then(response => response.text())
        .then(data => typeTextWithCallback(document.getElementById('status'), data, 20));
}

function stopServer() {
    fetch('/stop')
        .then(response => response.text())
        .then(data => typeTextWithCallback(document.getElementById('status'), data, 20));
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
        .then(data => typeTextWithCallback(document.getElementById('status'), data, 20));
}


function setDefaultConsoleMessage() {
    const consoleElement = document.getElementById('console');
    consoleElement.innerHTML = ''; // Clear the console initially

    // Define the boot title and sequence
    const bootTitle = "System booting...";
    const bootProgressStages = [
        "[                       ] 0%",
        "[####                   ] 20%",
        "[#####                  ] 30%",
        "[########               ] 40%",
        "[#############          ] 60%",
        "[##################     ] 80%",
        "[#######################] 100%"
    ];
    
    const bootCompleteMessage = "System boot complete.\nINFO: Press 'Start' to launch the game server.";

    // Random chance to trigger a funny bug (e.g., 20% chance)
    const bugChance = Math.random();
    const bugMessage = "ERROR: Nuclear parts in airflow detected! Initiating repair sequence...";

    // First, type the boot title
    typeTextWithCallback(consoleElement, bootTitle, 50, () => {
        // After the title is typed, start the progress bar update sequence
        setTimeout(() => {
            let index = 0;
            const progressElement = document.createElement('div');
            consoleElement.appendChild(progressElement); // Create an element for progress updates

            // Simulate the progress updates
            const bootInterval = setInterval(() => {
                if (index < bootProgressStages.length) {
                    progressElement.textContent = bootProgressStages[index]; // Update the progress in the same line
                    consoleElement.scrollTop = consoleElement.scrollHeight;  // Auto-scroll to bottom
                    index++;
                } else {
                    clearInterval(bootInterval);  // Stop when progress is done
                    
                    if (bugChance < 0.05) {
                        const bugElement = document.createElement('div');
                        bugElement.style.color = 'red'; // Make it look like an error
                        bugElement.textContent = bugMessage;
                        consoleElement.appendChild(bugElement);
                        consoleElement.scrollTop = consoleElement.scrollHeight;

                        // Delay for 2 seconds before "repairing"
                        setTimeout(() => {
                            const repairMessage = "Repair complete. Continuing boot sequence...";
                            const repairElement = document.createElement('div');
                            repairElement.style.color = 'green';
                            repairElement.textContent = repairMessage;
                            consoleElement.appendChild(repairElement);
                            consoleElement.scrollTop = consoleElement.scrollHeight;

                            // Finally, display the completion message after another short delay
                            setTimeout(() => {
                                const completionElement = document.createElement('div');
                                completionElement.innerHTML = bootCompleteMessage.replace(/\n/g, '<br>'); // Add the completion message
                                consoleElement.appendChild(completionElement);
                                consoleElement.scrollTop = consoleElement.scrollHeight;
                            }, 1000); // Delay for 1 second after repair message
                        }, 2000); // Repair after 2 seconds
                    } else {
                        // Display the normal completion message
                        setTimeout(() => {
                            const completionElement = document.createElement('div');
                            completionElement.innerHTML = bootCompleteMessage.replace(/\n/g, '<br>'); // Add the completion message
                            consoleElement.appendChild(completionElement);
                            consoleElement.scrollTop = consoleElement.scrollHeight;
                        }, 200);  // Delay for .5 second after progress reaches 100%
                    }
                }
            }, 150);  // Each progress update every 150ms
        }, 100);  // Initial delay of 2s to simulate a pause
    });
}


fetchOutput();
fetchBackups();