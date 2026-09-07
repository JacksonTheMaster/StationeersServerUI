// /static/console-manager.js

const STREAM_MESSAGE_LIMIT = 500;
const streamPanels = new Map();

function registerStreamPanel(tabId, consoleId, expectedConnections = 1) {
    const existing = streamPanels.get(tabId);
    const state = existing || {
        tabId,
        consoleId,
        paused: false,
        queue: [],
        connections: new Map(),
    };
    state.consoleId = consoleId;
    state.expectedConnections = expectedConnections;
    streamPanels.set(tabId, state);
    updateStreamToolbar();
    return state;
}

function getActiveStreamPanel() {
    const activeTab = document.querySelector('.tab-content.active');
    return activeTab ? streamPanels.get(activeTab.id) : null;
}

function trimStreamMessages(state) {
    const consoleElement = document.getElementById(state.consoleId);
    if (!consoleElement) return;
    const messages = Array.from(consoleElement.children)
        .filter(child => !child.classList.contains('sscm-command-container'));
    messages.slice(0, Math.max(0, messages.length - STREAM_MESSAGE_LIMIT)).forEach(message => message.remove());
}

function insertStreamMessage(state, message) {
    const consoleElement = document.getElementById(state.consoleId);
    if (!consoleElement) return;
    const commandInput = consoleElement.querySelector('.sscm-command-container');
    consoleElement.insertBefore(message, commandInput);
    trimStreamMessages(state);
    consoleElement.scrollTop = consoleElement.scrollHeight;
}

function appendStreamMessage(tabId, message) {
    const state = streamPanels.get(tabId);
    if (!state) return;
    if (state.paused) {
        state.queue.push(message);
        if (state.queue.length > STREAM_MESSAGE_LIMIT) state.queue.shift();
    } else {
        insertStreamMessage(state, message);
    }
    updateStreamToolbar();
}

function setStreamConnection(tabId, connectionId, status) {
    const state = streamPanels.get(tabId);
    if (!state) return;
    state.connections.set(connectionId, status);
    updateStreamToolbar();
}

function getStreamConnectionState(state) {
    const statuses = Array.from(state.connections.values());
    if (statuses.length === 0) return 'connecting';
    if (statuses.some(status => status === 'reconnecting')) return 'reconnecting';
    if (statuses.length >= state.expectedConnections && statuses.every(status => status === 'connected')) return 'connected';
    return 'connecting';
}

function updateStreamToolbar() {
    const state = getActiveStreamPanel();
    const text = document.getElementById('stream-status-text');
    const lamp = document.getElementById('stream-status-lamp');
    const pauseButton = document.getElementById('stream-pause-button');
    if (!state || !text || !lamp || !pauseButton) return;

    const connectionState = getStreamConnectionState(state);
    const displayState = state.paused ? 'paused' : connectionState;
    text.textContent = text.dataset[displayState];
    lamp.className = `stream-status-lamp ${displayState}`;
    pauseButton.textContent = state.paused ? pauseButton.dataset.resume : pauseButton.dataset.pause;
    pauseButton.classList.toggle('is-active', state.paused);
}

function toggleActiveStreamPause() {
    const state = getActiveStreamPanel();
    if (!state) return;
    state.paused = !state.paused;
    if (!state.paused && state.queue.length) {
        const queuedMessages = state.queue.splice(0);
        queuedMessages.forEach(message => insertStreamMessage(state, message));
    }
    updateStreamToolbar();
}

function clearActiveStream() {
    const state = getActiveStreamPanel();
    if (!state) return;
    const consoleElement = document.getElementById(state.consoleId);
    if (!consoleElement) return;
    Array.from(consoleElement.children)
        .filter(child => !child.classList.contains('sscm-command-container'))
        .forEach(child => child.remove());
    state.queue = [];
    updateStreamToolbar();
}

// Detection events streaming
function fetchDetectionEvents() {
    registerStreamPanel('detection-tab', 'detection-console');
    
    const connect = () => {
        detectionEventSource = new EventSource('/api/v3/streams/events');
        
        detectionEventSource.onmessage = event => {
            const message = document.createElement('div');
            message.className = `detection-event ${getEventClassName(event.data)}`;
            
            const timestamp = document.createElement('span');
            timestamp.className = 'event-timestamp';
            timestamp.textContent = `${new Date().toLocaleTimeString()}: `;
            
            const content = document.createElement('span');
            content.textContent = event.data;
            
            message.append(timestamp, content);
            appendStreamMessage('detection-tab', message);
            
            const detectionTab = document.getElementById('detection-tab');
            if (!detectionTab.classList.contains('active')) {
                const tabButton = document.querySelector('.tab-button[onclick*="detection-tab"]');
                tabButton.classList.add('notification');
                setTimeout(() => tabButton.classList.remove('notification'), 3000);
            }
        };
        
        detectionEventSource.onopen = () => {
            setStreamConnection('detection-tab', '/api/v3/streams/events', 'connected');
            console.log("Detection events stream connected");
        };
        
        detectionEventSource.onerror = () => {
            console.error("Detection events stream disconnected");
            detectionEventSource.close();
            detectionEventSource = null;
            setStreamConnection('detection-tab', '/api/v3/streams/events', 'reconnecting');
            if (window.location.pathname === '/') {
                setTimeout(connect, 2000);
            }
        };
    };
    connect();
}

// Console initialization with SSE stream setup
function handleConsole() {
    const consoleElement = document.getElementById('console');
    consoleElement.innerHTML = '';
    const streamState = registerStreamPanel('console-tab', 'console');
    streamState.queue = [];
    const bootTitle = "Interface initializing...";
    const bootCompleteMessage = "Interface ready.🎮 Happy gaming! 🎮";
    const bugChance = Math.random();
    const bugMessage = "ERROR: Nuclear parts in airflow detected! Initiating repair sequence...";
    
    const funMessages = [
        "Calibrating quantum flux capacitors...",
        "Initializing player happiness modules...",
        "Checking for monsters under the server...",
        "Brewing coffee for the CPU...",
        "Charging laser sharks...",
        "Teaching AI to say 'please' and 'thank you'...",
        "Polishing pixels to a mirror shine...",
        "Convincing electrons to flow in the right direction...",
        "Rebooting atmospheric systems for the 17th time...",
        "Attempting to locate your body after that last airlock malfunction...",
        "Converting oxygen to errors at alarming efficiency...",
        "Persuading physics engine to acknowledge gravity exists...",
        "Calculating ways your base will catastrophically depressurize...",
        "Optimizing unity garbage collection (good luck with that)...",
        "Aligning planetary rotation with server tick rate...",
        "Patching holes in space-time continuum and your habitat...",
        "Convincing solar panels that 'sun' is not just a theoretical concept...",
        "Negotiating peace treaty between logic circuits and the laws of thermodynamics...",
        "Compressing atmosphere until your CPU begs for mercy...",
        "Measuring distance between you and nearest fatal bug...",
        "Attempting to explain 'pipe networks' to confused server hamsters...",
        "Calculating probability of survival (spoiler: it's low)...",
        "Wrangling rogue Unity instances back into containment...",
        "Sacrificing RAM to the gods of stable framerates...",
        "Convincing electrons to flow in the right direction... nope, the power grid's borked.",
        "Patching hull breaches with duct tape and prayers...",
        "Recalculating O2 levels... wait, why is it all CO2 now?",
        "Spinning up the fabricator... hope it doesn't eat the server this time.",
        "Debugging Unity physics... object launched into orbit, send help.",
        "Warming up the furnace... or just setting the base on fire, 50/50 shot.",
        "Rerouting pipes... because who needs logical fluid dynamics anyway?",
        "Loading terrain... oh look, it's floating 3 meters above the ground again.",
        "Processing ore... into a fine paste of lag and despair.",
        "Stabilizing frame rate... lol, just kidding, welcome to 12 FPS city.",
        "Checking for updates... new bug introduced, feature still broken!",
        "Assembling solar tracker... now it's tracking the admin instead.",
        "Balancing gas mixtures... kaboom imminent, run you fool!",
        "Spoiler: object reference not set to an instance of an object, lol",
        "Fun fact: SSUI was originally a simple powershell script",
        "Fun fact: This dashboard was supposed to look like the retro computer from Stationeers. We tried, ok?",
        "Convincing server that 'out of memory' is just a state of mind.",
        "Moo, Moo! I'm a cow!",
        "Welcome home, Sir!"
    ];

    const cssVar = (name) => getComputedStyle(document.documentElement).getPropertyValue(name).trim();
    const addMessage = (text, color, style = 'normal') => {
        const div = document.createElement('div');
        div.textContent = text;
        div.style.color = color;
        div.style.fontStyle = style;
        appendStreamMessage('console-tab', div);
    };

    // Dynamically create SSCM command input
    const createCommandInput = async () => {
        try {
            // Make API call to check if SSCM is enabled
            const response = await fetch('/api/v3/sscm/status', {
                method: 'GET',
                headers: {
                    'Accept': 'application/json'
                }
            });
    
            // If status is not 200, exit the function
            if (response.status !== 200) {
                console.log('SSCM is not enabled, status:', response.status);
                return;
            }
    
            // Proceed to create command input UI if status is 200
            console.log("Creating command input...");
            const commandContainer = document.createElement('div');
            commandContainer.className = 'sscm-command-container';
            
            const prompt = document.createElement('span');
            prompt.className = 'prompt';
            prompt.textContent = '>';
    
            const input = document.createElement('input');
            input.id = 'sscm-command-input';
            input.type = 'text';
            input.placeholder = 'Enter command..';
            input.setAttribute('autocomplete', 'off');
            
            const suggestions = document.createElement('div');
            suggestions.id = 'sscm-autocomplete-suggestions';
            suggestions.className = 'sscm-suggestions';
            
            commandContainer.append(prompt, input, suggestions);
            const commandSlot = document.getElementById('sscm-command-slot') || consoleElement;
            commandSlot.appendChild(commandContainer);
            consoleElement.scrollTop = consoleElement.scrollHeight;
        } catch (error) {
            console.error('Error checking SSCM enabled status:', error);
            return; // Exit on error
        }
    };

    // Start with initializing message
    const bootLine = document.createElement('div');
    appendStreamMessage('console-tab', bootLine);
    typeTextWithCallback(bootLine, bootTitle, 30, () => {
        // Show two funny messages while connecting
        const messageIndex1 = Math.floor(Math.random() * funMessages.length);
        addMessage(funMessages[messageIndex1], cssVar('--console-info'), 'italic');

        let messageIndex2;
        do {
            messageIndex2 = Math.floor(Math.random() * funMessages.length);
        } while (messageIndex2 === messageIndex1);
        addMessage(funMessages[messageIndex2], cssVar('--console-info'), 'italic');

        // Set up the persistent console stream
        outputEventSource = new EventSource('/api/v3/streams/console');
        setStreamConnection('console-tab', '/api/v3/streams/console', 'connecting');
        
        // Persistent message handler
        outputEventSource.onmessage = event => {
            const message = document.createElement('div');
            message.textContent = event.data;
            appendStreamMessage('console-tab', message);
        };

        outputEventSource.onopen = () => {
            setStreamConnection('console-tab', '/api/v3/streams/console', 'connected');
            console.log("Console stream connected");
            finishInitialization();
        };

        outputEventSource.onerror = () => {
            console.error("Console stream disconnected");
            outputEventSource.close();
            outputEventSource = null;
            setStreamConnection('console-tab', '/api/v3/streams/console', 'reconnecting');
            addMessage("Warning: Console stream unavailable. Retrying...", cssVar('--console-warning'));
            if (window.location.pathname === '/') {
                setTimeout(() => {
                    if (!outputEventSource) {
                        // Re-run setup to reconnect
                        consoleElement.innerHTML = ''; // Clear console for fresh start
                        handleConsole();
                    }
                }, 5000);
            }
        };
    });

    function finishInitialization() {
        if (bugChance < 0.05) {
            addMessage(bugMessage, cssVar('--console-error'));
            setTimeout(() => {
                addMessage("Repair complete. Continuing initialization...", cssVar('--console-success'));
                completeBoot();
            }, 1000);
        } else {
            completeBoot();
        }
    }

    function completeBoot() {
        setTimeout(() => {
            createCommandInput(); // Add input after boot
            addMessage(bootCompleteMessage, cssVar('--console-success'));
            //addMessage("StationeersServerUI is becoming SteamServerUI!", '#ff4500');
            //addMessage("Please mind the New Terrain System warning below", '#ff4500');
            consoleElement.scrollTop = consoleElement.scrollHeight;
        }, 500);
    }
}

function setupLogStreams({ consoleId, streamUrls, maxMessages, messageClass }) {
    const consoleElement = document.getElementById(consoleId);
    if (!consoleElement) {
        console.error(`Console element with ID '${consoleId}' not found.`);
        return;
    }

    // Clear the console initially
    consoleElement.innerHTML = '';
    registerStreamPanel('log-tab', consoleId, streamUrls.length);

    const connectStream = (streamUrl) => {
        const eventSource = new EventSource(streamUrl);

        eventSource.onmessage = event => {
            const message = document.createElement('div');
            let finalClass = messageClass;

            // Check event.data for specific log levels and modify class
            if (event.data.includes('/INFO')) {
                finalClass += '-info';
            } else if (event.data.includes('/WARN')) {
                finalClass += '-warn';
            } else if (event.data.includes('/ERROR')) {
                finalClass += '-error';
            }

            message.classList.add("log-console-element", finalClass);

            const content = document.createElement('span');
            content.textContent = event.data;

            message.append(content);
            appendStreamMessage('log-tab', message);
        };

        eventSource.onopen = () => {
            setStreamConnection('log-tab', streamUrl, 'connected');
            console.log(`Stream ${streamUrl} connected for console ${consoleId}`);
        };

        eventSource.onerror = () => {
            console.error(`Stream ${streamUrl} disconnected for console ${consoleId}`);
            eventSource.close();
            setStreamConnection('log-tab', streamUrl, 'reconnecting');
            if (window.location.pathname === '/') {
                setTimeout(() => connectStream(streamUrl), 5000); // Reconnect after 5 seconds
            }
        };
    };

    // Connect to all provided stream URLs
    streamUrls.forEach(url => connectStream(url));
}
