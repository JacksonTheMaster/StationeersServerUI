// sscm.js

// Full command list with descriptions and parameters
const availableCommands = [
    { name: "addgas", params: "[Oxygen,Nitrogen,CarbonDioxide,Volatiles,Pollutant,Water,NitrousOxide]", desc: "Adds GasType to target thing" },
    { name: "atmos", params: "[pipe,world,direction,room,global,thing,cleanup,count,liquid]", desc: "Enables atmosphere debugging" },
    { name: "ban", params: "[<clientId>,refresh]", desc: "Bans a client from the server" },
    { name: "camera", params: "[shake]", desc: "Various camera debug functions" },
    { name: "celestial", params: "[eccentricity,semimajoraxisau,semimajoraxiskm,inclination,periapsis,period,ascendingnode,rotation]", desc: "Allows editing of celestial bodies" },
    { name: "cleanupplayers", params: "[dead,disconnected,all]", desc: "Cleans up player bodies" },
    { name: "clear", params: "", desc: "Clears all console text" },
    { name: "debugthreads", params: "[GameTick,Terrain]", desc: "Show worker thread run times" },
    { name: "deletelooseitems", params: "", desc: "Removes all loose items in world" },
    { name: "deleteoutofbounds", params: "", desc: "Removes out-of-bounds objects" },
    { name: "difficulty", params: "[<?difficulty>]", desc: "Prints or sets difficulty" },
    { name: "discord", params: "", desc: "Interaction with Discord SDK" },
    { name: "dlc", params: "[shared]", desc: "Various DLC debug functions" },
    { name: "emote", params: "[emoteName]", desc: "Triggers player emote" },
    { name: "entity", params: "[state <playerName OR referenceId>]", desc: "Entity debug functions" },
    { name: "exportworld", params: "", desc: "Exports world to WorldSettings file" },
    { name: "help", params: "[commands,list,<key>,tofile]", desc: "Displays command help" },
    { name: "helperhints", params: "[Dismiss,Complete,Trigger]", desc: "Tests world objectives" },
    { name: "keybindings", params: "[reset]", desc: "Displays or resets keybindings" },
    { name: "kick", params: "[<clientId>]", desc: "Kicks a client from server" },
    { name: "legacycpu", params: "[enable,disable]", desc: "Enables Legacy CPU mode" },
    { name: "liquid", params: "[show,renderer,solver,WorldVolume]", desc: "Debugs liquid solver" },
    { name: "listnetworkdevices", params: "[id]", desc: "Lists network devices" },
    { name: "localization", params: "[None,WordCount,Generate,Refresh,CheckKeys,CheckFonts]", desc: "Displays localization info" },
    { name: "log", params: "[<logname>,clear]", desc: "Dumps logs to file" },
    { name: "logtoclipboard", params: "", desc: "Copies console to clipboard" },
    { name: "masterserver", params: "[refresh]", desc: "Interacts with Master Server" },
    { name: "minables", params: "[range,generate]", desc: "Toggles minable debug" },
    { name: "netconfig", params: "[list,print,<PropertyName> <Value>]", desc: "Changes NetConfig.xml" },
    { name: "network", params: "", desc: "Shows network status" },
    { name: "networkdebug", params: "", desc: "Displays network debug window" },
    { name: "orbit", params: "[debug,view,celestials,simulate,set,timescale,makeoffset]", desc: "Controls orbital simulation" },
    { name: "plant", params: "[grow <parent thing id>]", desc: "Plant debug functions" },
    { name: "prefabs", params: "[Thumbnails]", desc: "Validates source prefabs" },
    { name: "printgasinfo", params: "", desc: "Prints gas coefficients" },
    { name: "profiler", params: "[enable,disable]", desc: "Toggles profiler" },
    { name: "regeneraterooms", params: "", desc: "Regenerates world rooms" },
    { name: "rocket", params: "[refresh,print,abandon,debug,chart]", desc: "Rocket debug functions" },
    { name: "save", params: "[<filename>,delete <filename>,list]", desc: "Saves game" },
    { name: "say", params: "", desc: "Sends message to players" },
    { name: "setbatteries", params: "[Empty,Critical,Very Low,Low,Medium,High,Full]", desc: "Sets battery levels" },
    { name: "settings", params: "[list,print,<PropertyName> <Value>]", desc: "Changes settings.xml" },
    { name: "settingspath", params: "[<full-directory-path>]", desc: "Sets settings path" },
    { name: "spacemap", params: "[regenerate,fill,chart,testpaths]", desc: "Space map debug functions" },
    { name: "spacemapnode", params: "[<id>]", desc: "Space map node debug" },
    { name: "status", params: "", desc: "Shows server state" },
    { name: "steam", params: "[Refresh,Store,Achieve,Clear,ClearAll,Invalid]", desc: "Tests Steamworks" },
    { name: "storm", params: "[start,stop,debug]", desc: "Controls weather events" },
    { name: "structure", params: "[completeall]", desc: "Structure debug functions" },
    { name: "structurenetwork", params: "[chute,rocket]", desc: "Debugs structure networks" },
    { name: "systeminfo", params: "", desc: "Prints system info" },
    { name: "test", params: "", desc: "Tests colors" },
    { name: "testbytearray", params: "", desc: "Tests network read/write" },
    { name: "testoctree", params: "[[number of iterations]]", desc: "Benchmarks read density" },
    { name: "thing", params: "[find <id>,delete <id>,spawn <prefabName> [amount],info <id>]", desc: "Manages things" },
    { name: "trader", params: "[regenerate,land,depart,contacts,buys,sells,evaluate,checksum]", desc: "Trader debug commands" },
    { name: "unstuck", params: "", desc: "Attempts to unstick player" },
    { name: "upnp", params: "", desc: "Shows UPnP state" },
    { name: "vegetation", params: "[set,debug]", desc: "Sets vegetation quantity" },
    { name: "version", params: "", desc: "Shows game version" },
    { name: "world", params: "", desc: "Prints world settings" },
    { name: "worldsetting", params: "", desc: "Authors WorldSetting info" }
];

async function checkSSCMEnabled() {
    try {
        const response = await fetch('/api/v2/SSCM/enabled', {
            method: 'GET',
            headers: { 'Content-Type': 'application/json' }
        });
        const input = document.getElementById('sscm-command-input');
        if (response.status === 200) {
            input.disabled = false;
            input.placeholder = "Enter command...";
        } else {
            input.onclick = () => {
                window.location.href = "/setup?step=sscm";
            };
            input.placeholder = "SSCM is not enabled, commands unavailable. Click here to configure.";
        }
    } catch (error) {
        console.error('Error checking SSCM status:', error);
        const input = document.getElementById('sscm-command-input');
        input.disabled = true;
        input.placeholder = "Commands unavailable";
    }
}

// Send command to SSCM run endpoint (unchanged)
async function sendSSCMCommand(command) {
    try {
        // Check server status before sending command
        const statusResponse = await fetch('/api/v2/server/status');
        const statusData = await statusResponse.json();
        
        if (!statusData.isRunning) {
            appendToConsole('[SSCM] Error: Gameserver is not running, start the server and try again.');
            return;
        }

        const response = await fetch('/api/v2/SSCM/run', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ command })
        });
        const result = await response.json();
        appendToConsole(result.status === 'success'
            ? `[SSCM] ${result.message}: ${command}`
            : `[SSCM] Error: ${result.message || 'Command failed'}`);
    } catch (error) {
        console.error('Error sending SSCM command:', error);
        appendToConsole(`[SSCM] Error: Failed to send command "${command}"`);
    }
}

// Append message to console
function appendToConsole(message) {
    const consoleDiv = document.getElementById('console');
    const messageElement = document.createElement('p');
    messageElement.textContent = message;
    consoleDiv.appendChild(messageElement);
    consoleDiv.scrollTop = consoleDiv.scrollHeight;
}

// Enhanced autocomplete functionality
function setupAutocomplete() {
    const input = document.getElementById('sscm-command-input');
    const suggestionsDiv = document.getElementById('sscm-autocomplete-suggestions');
    let selectedIndex = -1;

    function updateSuggestions(query) {
        suggestionsDiv.innerHTML = '';
        selectedIndex = -1;
        if (query.length === 0) return;

        const matches = availableCommands
            .filter(cmd => cmd.name.toLowerCase().startsWith(query.toLowerCase()))
            .sort((a, b) => a.name.length - b.name.length) // Prioritize shorter matches
            .slice(0, 5); // Limit to 5 suggestions

        if (matches.length === 0) return;

        matches.forEach((cmd, index) => {
            const suggestion = document.createElement('div');
            suggestion.classList.add('sscm-suggestion-item');
            suggestion.innerHTML = `
                <span class="sscm-suggestion-name">${cmd.name}</span>
                <span class="sscm-suggestion-params">${cmd.params || 'No params'}</span>
                <span class="sscm-suggestion-desc">${cmd.desc}</span>
            `;
            suggestion.addEventListener('click', () => {
                input.value = cmd.name;
                suggestionsDiv.innerHTML = '';
                input.focus();
            });
            suggestion.addEventListener('mouseover', () => {
                selectedIndex = index;
                updateHighlight();
            });
            suggestionsDiv.appendChild(suggestion);
        });
    }

    function updateHighlight() {
        const items = suggestionsDiv.querySelectorAll('.sscm-suggestion-item');
        items.forEach((item, index) => {
            item.classList.toggle('highlighted', index === selectedIndex);
        });
    }

    function selectBestMatch() {
        const query = input.value.toLowerCase();
        const matches = availableCommands.filter(cmd => 
            cmd.name.toLowerCase().startsWith(query)
        ).sort((a, b) => a.name.length - b.name.length); // Shortest match first
        return matches[0]?.name || input.value;
    }

    input.addEventListener('input', () => {
        updateSuggestions(input.value);
    });

    input.addEventListener('keydown', async (e) => {
        const items = suggestionsDiv.querySelectorAll('.sscm-suggestion-item');
        if (e.key === 'ArrowDown') {
            e.preventDefault();
            selectedIndex = Math.min(selectedIndex + 1, items.length - 1);
            updateHighlight();
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            selectedIndex = Math.max(selectedIndex - 1, -1);
            updateHighlight();
        } else if (e.key === 'Tab' || e.key === 'Enter') {
            if (items.length > 0) {
                e.preventDefault();
                input.value = selectedIndex >= 0 
                    ? suggestionsDiv.children[selectedIndex].querySelector('.sscm-suggestion-name').textContent
                    : selectBestMatch();
                suggestionsDiv.innerHTML = '';
                if (e.key === 'Enter' && !input.disabled && input.value.trim()) {
                    await sendSSCMCommand(input.value.trim());
                    input.value = '';
                }
            } else if (e.key === 'Enter' && !input.disabled && input.value.trim()) {
                await sendSSCMCommand(input.value.trim());
                input.value = '';
            }
        } else if (e.key === 'Escape') {
            suggestionsDiv.innerHTML = '';
            selectedIndex = -1;
        }
    });

    // Clear suggestions when clicking outside
    document.addEventListener('click', (e) => {
        if (!input.contains(e.target) && !suggestionsDiv.contains(e.target)) {
            suggestionsDiv.innerHTML = '';
            selectedIndex = -1;
        }
    });
}

// Wait for SSCM input to be created before initializing
function waitForSSCMInput(callback) {
    const checkInput = () => {
        const input = document.getElementById('sscm-command-input');
        const suggestionsDiv = document.getElementById('sscm-autocomplete-suggestions');
        if (input && suggestionsDiv) {
            callback();
        } else {
            setTimeout(checkInput, 100); // Poll every 100ms
        }
    };
    checkInput();
}

// Initialize SSCM functionality
document.addEventListener('DOMContentLoaded', () => {
    waitForSSCMInput(() => {
        checkSSCMEnabled();
        setupAutocomplete();
        setInterval(checkSSCMEnabled, 30000);
    });
});