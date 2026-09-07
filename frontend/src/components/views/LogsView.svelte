<script>
  import { onMount, onDestroy } from 'svelte';
  import { apiSSE } from '../../services/api';

  // Main state
  let logs = $state([]);
  let filteredLogs = $state([]);
  let logSources = $state({
    all: false,
    debug: false,
    info: true,
    warning: false,
    error: false,
    game: false
  });
  let timeRange = $state('Recent');
  let activeSources = $state({});
  let isReconnecting = $state(false);

  // DOM refs
  let logsViewer;

  // Connection state
  let connections = $state({
    all: null,
    debug: null,
    info: null,
    warning: null,
    error: null,
    game: null
  });
  let connecting = $state({
    all: false,
    debug: false,
    info: false,
    warning: false,
    error: false,
    game: false
  });

  const systemNames = {
    all: 'Full Backend',
    debug: 'Debug',
    info: 'Info',
    warning: 'Warn',
    error: 'Error',
    game: 'Game'
  };

  // Time ranges in milliseconds for easy comparison
  const timeRanges = {
    'Recent': 10 * 1000, //this is a bit of a hack, but it works and looks nice
    'Last 1 Minute': 60 * 1000,
    'Last 10 Minutes': 10 * 60 * 1000,
    'Last 20 Minutes': 20 * 60 * 1000,
    'Last 30 Minutes': 30 * 60 * 1000,
    'Last 1 Hour': 60 * 60 * 1000,
    'Last 12 Hours': 12 * 60 * 60 * 1000,
    'Last 24 Hours': 24 * 60 * 60 * 1000,
    'All Time': Infinity
  };

  // Watch for changes in log sources and update connections
  $effect(() => {
    const newActiveSources = {};
    for (const [key, value] of Object.entries(logSources)) {
      if (value) {
        newActiveSources[key] = true;
      }
    }
    activeSources = newActiveSources;

    // Use microtask to ensure this runs after the state update
    queueMicrotask(() => {
      updateConnections();
    });
  });

  // Watch for changes in logs or time range and update filtered logs
  $effect(() => {
    filteredLogs = applyTimeFilter(logs);
  });

  // Setup on mount
  onMount(() => {
    updateConnections();
    
    // Set up visibility change handler for better connection management
    document.addEventListener('visibilitychange', handleVisibilityChange);
    
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  });

  // Cleanup on destroy
  onDestroy(() => {
    closeAllConnections();
  });

  function handleVisibilityChange() {
    if (document.visibilityState === 'visible') {
      updateConnections();
    }
  }

  function updateConnections() {
    const sources = ['all', 'debug', 'info', 'warning', 'error', 'game'];
    for (const source of sources) {
      if (activeSources[source] && !connections[source] && !connecting[source]) {
        connectToLogSource(source);
      } else if (!activeSources[source] && connections[source]) {
        disconnectLogSource(source);
      }
    }
  }

  function connectToLogSource(source) {
    if (connecting[source]) {
      return;
    }
    
    connecting[source] = true;
    console.log(`Connecting to ${source} log source`);

    const endpoint = source === 'all' ? '/api/v3/streams/logs/backend' :
                     source === 'warning' ? '/api/v3/streams/logs/warn' :
                     source === 'game' ? '/api/v3/streams/console' : `/api/v3/streams/logs/${source}`;
    const systemName = systemNames[source];
    const level = source === 'all' ? 'ALL' : source.toUpperCase();
    const connectionMessage = `${systemName} Log Stream Connected`;

    connections[source] = apiSSE(endpoint,
      (data) => {
        const now = new Date();
        const timestamp = now.toLocaleTimeString();
        
        // Store the exact timestamp for accurate filtering
        const exactTimestamp = now.getTime();
        
        if (data === connectionMessage) {
          const hasConnectionMessage = logs.some(log => 
            log.message === connectionMessage && 
            log.level === level
          );
          
          if (!hasConnectionMessage) {
            addLog({
              timestamp,
              exactTimestamp,
              level,
              server: 'All',
              message: connectionMessage
            });
          }
        } else {
          addLog({
            timestamp,
            exactTimestamp,
            level,
            server: 'All',
            message: data
          });
        }
      },
      (error) => {
        console.error(`${systemName} logs stream error:`, error);
        connections[source] = null;
        connecting[source] = false;

        // Retry connection if still active and page is visible
        if (activeSources[source] && document.visibilityState !== 'hidden') {
          setTimeout(() => {
            if (activeSources[source] && !connections[source] && !connecting[source]) {
              connectToLogSource(source);
            }
          }, 2000);
        }
      }
    );

    connecting[source] = false;
  }

  function disconnectLogSource(source) {
    const systemName = systemNames[source];
    
    if (connections[source]) {
      connections[source].close();
      connections[source] = null;

      if (!isReconnecting) {
        const now = new Date();
        const timestamp = now.toLocaleTimeString();
        const level = source === 'all' ? 'ALL' : source.toUpperCase();
        
        addLog({
          timestamp,
          exactTimestamp: now.getTime(),
          level,
          server: 'All',
          message: `${systemName} Log Stream Disconnected`
        });
      }
    }
  }

  function addLog(log) {
    // Add log to the main array
    logs = [log, ...logs];
    
    // Keep array size manageable (limit to 2000 entries)
    if (logs.length > 2000) {
      logs = logs.slice(-2000);
    }
  }

  function applyTimeFilter(allLogs) {
    if (timeRange === 'All Time') {
      return allLogs;
    }

    const now = Date.now();
    const cutoffTime = now - timeRanges[timeRange];
    
    return allLogs.filter(log => {
      // Use the exact timestamp for precise filtering
      return log.exactTimestamp >= cutoffTime;
    });
  }

  function handleCheckboxChange(level) {
    logSources[level] = !logSources[level];
  }

  function clearLogs() {
    logs = [];
    filteredLogs = [];
  }

  function reconnectLogs() {
    isReconnecting = true;
    closeAllConnections();
    clearLogs();
    
    // Short delay to ensure connections are properly closed
    setTimeout(() => {
      updateConnections();
      isReconnecting = false;
    }, 100);
  }

  function closeAllConnections() {
    for (const source of Object.keys(connections)) {
      if (connections[source]) {
        disconnectLogSource(source);
      }
    }
  }

</script>

<div class="logs-container">
  <div class="logs-filter">
    <div class="filter-group">
      <label>Log Level</label>
      <div class="checkbox-group">
        <label><input type="checkbox" checked={logSources.all} onchange={() => handleCheckboxChange('all')} /> All</label>
        <label><input type="checkbox" checked={logSources.info} onchange={() => handleCheckboxChange('info')} /> Info</label>
        <label><input type="checkbox" checked={logSources.warning} onchange={() => handleCheckboxChange('warning')} /> Warning</label>
        <label><input type="checkbox" checked={logSources.error} onchange={() => handleCheckboxChange('error')} /> Error</label>
        <label><input type="checkbox" checked={logSources.debug} onchange={() => handleCheckboxChange('debug')} /> Debug</label>
        <label><input type="checkbox" checked={logSources.game} onchange={() => handleCheckboxChange('game')} /> Game</label>
      </div>
    </div>

    <div class="filter-group">
      <label>Time Range</label>
      <select bind:value={timeRange}>
        <option>Recent</option>
        <option>Last 1 Minute</option>
        <option>Last 10 Minutes</option>
        <option>Last 20 Minutes</option>
        <option>Last 30 Minutes</option>
        <option>Last 1 Hour</option>
        <option>Last 12 Hours</option>
        <option>Last 24 Hours</option>
        <option>All Time</option>
      </select>
    </div>

    <div class="control-buttons">
      <button class="reconnect-button" onclick={reconnectLogs}>Reconnect</button>
      <button class="clear-button" onclick={clearLogs}>Clear</button>
    </div>
  </div>

  <div class="logs-viewer" bind:this={logsViewer}>
    {#if filteredLogs.length === 0}
      <div class="no-logs">No logs to display. Select a log level or change the Time Range to view logs...</div>
    {:else}
      {#each filteredLogs as log}
        <div class="log-line">
          <span class="timestamp">{log.timestamp}</span>
          <span class="level {log.level.toLowerCase()}">{log.level}</span>
          <span class="message">{log.message}</span>
        </div>
      {/each}
    {/if}
  </div>
</div>

<style>
  .disconnect {
    color: #ff5555; /* Optional: Style disconnection messages differently */
    font-style: italic;
  }
  
  .logs-container {
    display: flex;
    flex-direction: column;
    height: 100%;
    gap: 1rem;
  }
  
  .logs-filter {
    display: flex;
    gap: 1rem;
    flex-wrap: wrap;
    background-color: var(--bg-secondary);
    padding: 1rem;
    border-radius: 4px;
  }
  
  .filter-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  
  .filter-group label {
    font-size: 0.9rem;
    color: var(--text-secondary);
  }
  
  .filter-group select {
    background-color: var(--bg-tertiary);
    color: var(--text-primary);
    border: 1px solid var(--border-color);
    padding: 0.5rem;
    border-radius: 4px;
    min-width: 150px;
  }
  
  .checkbox-group {
    display: flex;
    gap: 1rem;
  }
  
  .checkbox-group label {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    cursor: pointer;
  }
  
  .checkbox-group input[type="checkbox"] {
    accent-color: var(--accent-primary);
  }
  
  .control-buttons {
    margin-left: auto;
    display: flex;
    gap: 0.5rem;
    align-self: flex-end;
  }
  
  .reconnect-button, .clear-button {
    background-color: var(--accent-primary);
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    font-weight: 500;
    border-radius: 4px;
  }
  
  .clear-button {
    background-color: var(--bg-tertiary);
  }
  
  .reconnect-button:hover, .clear-button:hover {
    background-color: var(--accent-secondary);
  }
  
  .logs-viewer {
    flex: 1;
    background-color: var(--bg-secondary);
    border-radius: 4px;
    padding: 1rem;
    overflow-y: auto;
    font-family: 'Consolas', 'Courier New', monospace;
    font-size: 0.9rem;
  }
  
  .no-logs {
    color: var(--text-secondary);
    text-align: center;
    padding: 2rem;
  }
  
  .log-line {
    padding: 0.25rem 0;
    border-bottom: 1px solid var(--bg-tertiary);
    white-space: nowrap;
  }
  
  .log-line:hover {
    background-color: var(--bg-hover);
  }
  
  .timestamp {
    color: var(--text-secondary);
    margin-right: 1rem;
    display: inline-block;
    width: 80px;
  }
  
  .level {
    display: inline-block;
    width: 60px;
    text-align: center;
    padding: 0.1rem 0.5rem;
    border-radius: 3px;
    margin-right: 1rem;
    font-size: 0.8rem;
    font-weight: 600;
  }
  
  .level.info {
    background-color: rgba(3, 169, 244, 0.2);
    color: #4fc3f7;
  }
  
  .level.warning {
    background-color: rgba(255, 152, 0, 0.2);
    color: #ffb74d;
  }
  
  .level.error {
    background-color: rgba(244, 67, 54, 0.2);
    color: #e57373;
  }
  
  .level.debug {
    background-color: rgba(158, 158, 158, 0.2);
    color: #bdbdbd;
  }
  
  .level.all {
    background-color: rgba(76, 175, 80, 0.2);
    color: #81c784;
  }
  
  .level.game {
    background-color: rgba(156, 39, 176, 0.2);
    color: #ba68c8;
  }
  
  .message {
    color: var(--text-primary);
  }
</style>
