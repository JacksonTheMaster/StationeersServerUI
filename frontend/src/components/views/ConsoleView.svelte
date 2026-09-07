<script>
  import { onMount, onDestroy } from 'svelte';
  import { apiSSE } from '../../services/api'; // Import the apiSSE function
  
  // State variables
  let consoleElement = $state();
  let detectionConsole = $state();
  let outputEventSource = null;
  let detectionEventSource = null;
  let bootComplete = false;
  let activeTab = $state('console-tab');
  let autoScroll = $state(true);
  
  // Fun boot messages
  const funMessages = [
  "Teaching AI to say 'please' and 'thank you'...",
  "Calibrating interdimensional server relays...",
  "Spinning up player fun generators...",
  "Scanning for gremlins in the server room...",
  "Brewing energy drinks for the GPU...",
  "Charging up epic loot spawners...",
  "Teaching AI to high-five players...",
  "Polishing textures to a glossy finish...",
  "Herding electrons into neat little rows...",
  "Rebooting server dreams for the 42nd time...",
  "Locating your character after that teleport glitch...",
  "Converting lag into pure chaos, efficiently...",
  "Convincing physics engine that walls are solid...",
  "Calculating odds of your inventory vanishing...",
  "Optimizing server ticks (fingers crossed)...",
  "Syncing game world with cosmic clock...",
  "Patching rips in the fabric of virtual reality...",
  "Convincing server that 'uptime' is not a myth...",
  "Mediating truce between CPU and lag spikes...",
  "Compressing map data until the server squeaks...",
  "Measuring distance to the next game-breaking bug...",
  "Explaining 'netcode' to confused server pigeons...",
  "Calculating chance of a crash (don’t ask)...",
  "Corraling rogue packets back into the pipeline...",
  "Sacrificing bandwidth to the frame-rate deities...",
  "Raising sails!"
];
  
  // Console messages storage
  let consoleMessages = $state([]);
  let detectionEvents = $state([]);
  let hasNewDetections = $state(false);
  
  // Switch tabs
  function switchTab(tabId) {
    activeTab = tabId;
    if (tabId === 'detection-tab') {
      hasNewDetections = false;
    }
  }
  
  onMount(async () => {
    // Start boot sequence
    await typeText("Console initializing...");
    
    // Show two random funny messages
    const messageIndex1 = Math.floor(Math.random() * funMessages.length);
    addConsoleMessage(funMessages[messageIndex1], '#0af', 'italic');
    
    let messageIndex2;
    do {
      messageIndex2 = Math.floor(Math.random() * funMessages.length);
    } while (messageIndex2 === messageIndex1);
    addConsoleMessage(funMessages[messageIndex2], '#0af', 'italic');
    
    // Connect to SSE streams
    connectConsoleStream();
    connectDetectionEvents();
    
    // Add a small chance for a "bug" message
    const bugChance = Math.random();
    if (bugChance < 0.5) {
      addConsoleMessage("ERROR: Console was hacked by AI. Initiating counter-measure...", 'red');
      
      setTimeout(() => {
        addConsoleMessage("Sequence complete. AI defeated. Continuing initialization...", 'green');
        completeBootSequence();
      }, 1000);
    } else {
      completeBootSequence();
    }
  });
  
  onDestroy(() => {
    // Clean up SSE connections
    if (outputEventSource) {
      outputEventSource.close();
      outputEventSource = null;
    }
    
    if (detectionEventSource) {
      detectionEventSource.close();
      detectionEventSource = null;
    }
  });
  
  // Helper functions
  async function typeText(text, delay = 30) {
    let index = 0;
    
    return new Promise(resolve => {
      const interval = setInterval(() => {
        if (index < text.length) {
          consoleMessages = [...consoleMessages, { 
            text: text.substring(0, ++index), 
            color: 'white', 
            style: 'normal',
            isTyping: true 
          }];
          consoleMessages = consoleMessages.slice(-1);
        } else {
          clearInterval(interval);
          consoleMessages = consoleMessages.filter(msg => !msg.isTyping);
          resolve();
        }
      }, delay);
    });
  }
  
  function addConsoleMessage(text, color = 'white', style = 'normal') {
    consoleMessages = [...consoleMessages, { text, color, style }];
    if (autoScroll) {
      setTimeout(() => scrollConsole(), 50); // Slightly increased delay for better reliability
    }
  }
  
  function scrollConsole() {
    if (consoleElement && autoScroll) {
      consoleElement.scrollTop = consoleElement.scrollHeight;
    }
  }
  
  function scrollDetectionConsole() {
    if (detectionConsole && autoScroll) {
      detectionConsole.scrollTop = detectionConsole.scrollHeight;
    }
  }
  
  function completeBootSequence() {
    setTimeout(() => {
      addConsoleMessage("Console ready.🎮 Happy gaming! 🎮", '#0f0');
      bootComplete = true;
      if (autoScroll) {
        setTimeout(scrollConsole, 50);
      }
    }, 500);
  }
  
  function connectConsoleStream() {
  // Close any existing connection
  if (outputEventSource) {
    outputEventSource.close();
    outputEventSource = null;
  }
  
  // Use apiSSE instead of direct EventSource
    outputEventSource = apiSSE('/api/v3/streams/console',
    (data) => {
      // Handle the message
      addConsoleMessage(data);
    },
    (error) => {
      console.error("Console stream error:", error);
      outputEventSource = null;
      addConsoleMessage("Warning: Console stream unavailable. Retrying...", '#ff0');
      
      setTimeout(() => {
        if (!outputEventSource && document.visibilityState !== 'hidden') {
          connectConsoleStream();
        }
      }, 2000);
    }
  );
}
  
function connectDetectionEvents() {
  // Close any existing connection
  if (detectionEventSource) {
    detectionEventSource.close();
    detectionEventSource = null;
  }
  
  // Use apiSSE instead of direct EventSource
    detectionEventSource = apiSSE('/api/v3/streams/events',
    (data) => {
      const eventClass = getEventClassName(data);
      const timestamp = new Date().toLocaleTimeString();
      
      detectionEvents = [...detectionEvents, {
        timestamp,
        message: data,
        className: eventClass
      }];
      
      // Limit to max messages
      if (detectionEvents.length > 500) {
        detectionEvents = detectionEvents.slice(-500);
      }
      
      if (autoScroll) {
        setTimeout(scrollDetectionConsole, 50);
      }
      
      // Add notification if detection tab is not active
      if (activeTab !== 'detection-tab') {
        hasNewDetections = true;
      }
    },
    (error) => {
      console.error("Detection events stream error:", error);
      detectionEventSource = null;
      
      setTimeout(() => {
        if (!detectionEventSource && document.visibilityState !== 'hidden') {
          connectDetectionEvents();
        }
      }, 2000);
    }
  );
}
  
  function getEventClassName(message) {
    if (message.includes('ERROR') || message.includes('FATAL')) {
      return 'error';
    } else if (message.includes('WARNING')) {
      return 'warning';
    } else if (message.includes('SUCCESS')) {
      return 'success';
    } else {
      return 'info';
    }
  }
  
  function toggleAutoScroll() {
    autoScroll = !autoScroll;
    if (autoScroll) {
      scrollConsole();
      scrollDetectionConsole();
    }
  }
</script>

<div class="console-view">
  <div class="console-controls">
    <div class="console-tabs">
      <button 
        class="tab-button {activeTab === 'console-tab' ? 'active' : ''}" 
        onclick={() => switchTab('console-tab')}
      >
        Console
      </button>
      <button 
        class="tab-button {activeTab === 'detection-tab' ? 'active' : ''} {hasNewDetections ? 'notification' : ''}" 
        onclick={() => switchTab('detection-tab')}
      >
        Detections
      </button>
    </div>
    <button 
      class="autoscroll-button" 
      onclick={toggleAutoScroll}
      title={autoScroll ? 'Disable Auto-scroll' : 'Enable Auto-scroll'}
    >
      {autoScroll ? '⏬ Auto-scroll: ON' : '⏫ Auto-scroll: OFF'}
    </button>
  </div>

  <div class="tab-content">
    {#if activeTab === 'console-tab'}
      <div class="console-tab active">
        <div class="console" bind:this={consoleElement}>
          {#each consoleMessages as message}
            <div style="color: {message.color}; font-style: {message.style}">
              {message.text}
            </div>
          {/each}
        </div>
      </div>
    {:else}
      <div class="detection-tab active">
        <div class="detection-console" bind:this={detectionConsole}>
          {#each detectionEvents as event}
            <div class="detection-event {event.className}">
              <span class="event-timestamp">{event.timestamp}: </span>
              <span>{event.message}</span>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  .console-view {
    display: flex;
    flex-direction: column;
    height: 100%;
    max-height: 90vh;
    background-color: var(--bg-tertiary);
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
    overflow: hidden;
  }

  .console-controls {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem;
    background-color: var(--bg-secondary);
    border-bottom: 1px solid var(--border-color);
    flex-shrink: 0;
  }

  .console-tabs {
    display: flex;
    gap: 0.5rem;
  }

  .tab-button {
    padding: 0.75rem 1.5rem;
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 6px 6px 0 0;
    color: var(--text-primary);
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
    position: relative;
    overflow: hidden;
  }

  .tab-button::before {
    content: '';
    position: absolute;
    top: 0;
    left: -100%;
    width: 100%;
    height: 100%;
    background: linear-gradient(
      90deg,
      transparent,
      rgba(255, 255, 255, 0.2),
      transparent
    );
    transition: 0.5s;
  }

  .tab-button:hover::before {
    left: 100%;
  }

  .tab-button:hover {
    background: var(--bg-hover);
    transform: translateY(-2px);
  }

  .tab-button.active {
    background: var(--accent-primary);
    color: white;
    border-bottom: none;
    transform: translateY(0);
  }

  .tab-button.notification {
    animation: pulse 1s infinite ease-in-out;
    position: relative;
  }

  .tab-button.notification::after {
    content: '';
    position: absolute;
    top: 5px;
    right: 5px;
    width: 8px;
    height: 8px;
    background: #ff5555;
    border-radius: 50%;
  }

  .autoscroll-button {
    padding: 0.5rem 1rem;
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    color: var(--text-primary);
    font-family: 'Courier New', monospace;
    font-size: 0.8rem;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .autoscroll-button:hover {
    background: var(--bg-hover);
    transform: translateY(-1px);
  }

  .tab-content {
    flex: 1;
    overflow: hidden;
    position: relative;
  }

  .console-tab, .detection-tab {
    position: absolute;
    width: 100%;
    height: 100%;
    display: none;
  }

  .console-tab.active, .detection-tab.active {
    display: block;
  }

  .console, .detection-console {
    height: 100%;
    overflow-y: auto;
    padding: 1rem;
    font-family: 'Courier New', monospace;
    font-size: 0.9rem;
    line-height: 1.4;
    color: var(--text-primary);
    background-color: rgba(0, 0, 0, 0.8);
    box-sizing: border-box;
  }

  .detection-event {
    margin-bottom: 0.25rem;
    word-break: break-word;
  }

  .detection-event.error {
    color: #ff5555;
  }

  .detection-event.warning {
    color: #ffaa33;
  }

  .detection-event.success {
    color: #33cc33;
  }

  .detection-event.info {
    color: #55aaff;
  }

  .event-timestamp {
    color: #888888;
    font-weight: bold;
  }

  @keyframes pulse {
    0% { opacity: 1; }
    50% { opacity: 0.5; }
    100% { opacity: 1; }
  }
</style>
