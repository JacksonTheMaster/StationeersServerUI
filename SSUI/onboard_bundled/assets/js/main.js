// /static/main.js
document.addEventListener('DOMContentLoaded', () => {

    typeh1(document.querySelector('h1'), 30);
    if (window.location.pathname == '/') {
        setupTabs();
        if (window.SSUIAccess.can('server.view')) {
            fetchDetectionEvents();
        } else {
            document.getElementById('detection-console').textContent = "You don't have permission to view server events.";
        }
        if (window.SSUIAccess.can('console.read')) {
            setupLogStreams({
                consoleId: 'backendlog-console',
                streamUrls: [
                    '/api/v3/streams/logs/info',
                    '/api/v3/streams/logs/warn',
                    '/api/v3/streams/logs/error',
                ],
                maxMessages: 500,
                messageClass: 'log-console-element'
            });
            handleConsole();
        } else {
            document.getElementById('console').textContent = "You don't have permission to view the server console.";
            document.getElementById('backendlog-console').textContent = "You don't have permission to view backend logs.";
        }
        fetchBackups();
        fetchPlayers();
        pollRecurringTasks();
        console.warn("If you see errors for sscm.js or sscm.css, you may want to enable SSCM.");
    }
    // Language flag selection
    const languageFlags = document.querySelectorAll('#language-flags img');
    languageFlags.forEach(flag => {
        flag.addEventListener('click', async () => {
            const lang = flag.dataset.lang;
            try {
                const response = await fetch('/api/v3/settings', {
                    method: 'PATCH',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ language: lang })
                });
                const data = await response.json();
                if (response.ok) {
                    window.location.reload();
                } else {
                    console.error(data.error);
                }
            } catch (error) {
                console.error('Language setting error:', error);
            }
            
        });
    });
});

// Global references to EventSource objects
let outputEventSource = null;
let detectionEventSource = null;

function closeEventSources() {
    [outputEventSource, detectionEventSource].forEach(source => {
        if (source) {
            source.close();
            console.log(`${source === outputEventSource ? 'Output' : 'Detection events'} stream closed`);
        }
    });
    outputEventSource = detectionEventSource = null;
}

function typeh1(element, speed) {
    // Check if typing is already in progress
    if (element.dataset.isTyping === 'true') {
        // Optionally, clear the previous timeout (requires storing it)
        clearTimeout(element.dataset.timeoutId);
    }

    const fullText = element.textContent;
    element.textContent = '';
    element.dataset.isTyping = 'true'; // Mark as typing
    let i = 0;
    
    const typeChar = () => {
        if (i < fullText.length) {
            element.textContent += fullText.charAt(i++);
            const timeoutId = setTimeout(typeChar, speed);
            element.dataset.timeoutId = timeoutId; // Store timeout ID
        } else {
            element.dataset.isTyping = 'false'; // Done typing
            delete element.dataset.timeoutId;
        }
    };
    typeChar();
}

function navigateTo(url) {
    closeEventSources();
    window.location.href = url;
}
