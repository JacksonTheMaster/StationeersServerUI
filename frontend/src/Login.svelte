<script>

  import { onMount } from 'svelte';
  import { authState, backendConfig, login, getCurrentBackendUrl, setActiveBackend, setBackend } from './services/api';
  import { get } from 'svelte/store';
  
  // Form data
  let username = $state('');
  let password = $state('');
  let errorMessage = $state('');
  let isSubmitting = $state(false);
  let backends = $state([]);
  let activeBackend = $state('');
  let backendStatuses = $state({});
  let showBackendSelector = $state(false);
  let testingBackend = $state(false);
  
  // New backend form data
  let showNewBackendForm = $state(false);
  let newBackendId = $state('');
  let newBackendUrl = $state('');
  
  // Subscribe to auth state
  let unsubscribe;
  let unsubscribeBackend;
  
  onMount(() => {
    unsubscribe = authState.subscribe(state => {
  if (state.authError) {
    errorMessage = state.authError;
    // Show the form for any auth error unless the backend requires login
    const activeStatus = backendStatuses[activeBackend]?.status;
    const authRequired = backendStatuses[activeBackend]?.authRequired;
    showNewBackendForm = !(activeStatus === 'online' && authRequired);
  } else {
    errorMessage = '';
    showNewBackendForm = false;
  }
  isSubmitting = state.isAuthenticating;
});

unsubscribeBackend = backendConfig.subscribe(config => {
  backends = Object.keys(config.backends);
  activeBackend = config.active;

  // Initialize backend statuses
  backends.forEach(id => {
    if (!(id in backendStatuses)) {
      backendStatuses[id] = { status: 'unknown', lastChecked: null, error: null };
    }
  });

  // Show new backend form for specific backend status errors
  const activeStatus = backendStatuses[activeBackend]?.status;
  const authRequired = backendStatuses[activeBackend]?.authRequired;
  if (['offline', 'unreachable', 'cert-error'].includes(activeStatus)) {
    showNewBackendForm = true;
    errorMessage = backendStatuses[activeBackend]?.error || 'Cannot connect to the server.';
  } else if (activeStatus === 'online' && authRequired) {
    showNewBackendForm = false; // Explicitly hide form if login is required
  }
});

  // Check active backend status
  checkBackendStatus(activeBackend);

  return () => {
    if (unsubscribe) unsubscribe();
    if (unsubscribeBackend) unsubscribeBackend();
  };
});
  
// Test backend connection
async function checkBackendStatus(id) {
  testingBackend = true;
  backendStatuses[id] = { status: 'checking', lastChecked: new Date(), error: null };
  backendStatuses = { ...backendStatuses };
  
  try {
    const config = get(backendConfig);
    const backendUrl = config.backends[id].url === '/' ? '' : config.backends[id].url;
    
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000);
    
    const response = await fetch(`${backendUrl}/api/v3/server/status`, {
      method: 'GET',
      signal: controller.signal
    });
    
    clearTimeout(timeoutId);
    
    // Handle specific status codes
    if (response.ok) {
      backendStatuses[id] = {
        status: 'online',
        lastChecked: new Date(),
        error: null
      };
    } else if (response.status === 401) {
      // Treat 401 Unauthorized as "ok" to show login box
      backendStatuses[id] = {
        status: 'online',
        lastChecked: new Date(),
        authRequired: true,
        error: null
      };
    } else {
      backendStatuses[id] = {
        status: 'offline',
        lastChecked: new Date(),
        error: `Server responded with ${response.status} ${response.statusText}`
      };
    }
  } catch (error) {
    let status = 'error';
    let errorMsg = 'Failed to connect to the server';
    let certificateHint = false;
    const isHttps = backendUrl.startsWith('https://');

    if (error.name === 'AbortError') {
      errorMsg = isRunningInElectron() 
        ? "You are using the Desktop App, so you must define your Backend."
        : "Connection timed out. The server may be slow, unreachable, or dead.";
    } else if (isHttps) {
      status = 'cert-error';
      certificateHint = true;
      errorMsg = 'There was an error connecting to the remote Backend. If using a self-signed certificate, please visit the server URL to accept the certificate.';
    } else {
      errorMsg = 'Server not found. The server may be down or the URL may be incorrect.';
    }

    backendStatuses[id] = {
      status,
      lastChecked: new Date(),
      error: errorMsg,
      certificateHint,
      backendUrl // Store the backend URL for the link
    };
  } finally {
    testingBackend = false;
    backendStatuses = { ...backendStatuses };
  }
}
  
  // Handle form submission
  async function handleSubmit(event) {
    if (event && event.preventDefault) event.preventDefault();
    errorMessage = '';
    isSubmitting = true;
    
    if (!username || !password) {
      errorMessage = 'Username and password are required';
      isSubmitting = false;
      return;
    }
    
    try {
      const success = await login(username, password);
      
      if (!success) {
        // Error message will be set via the authState subscription
      }
    } catch (error) {
      errorMessage = error.message || 'Login failed';
    } finally {
      isSubmitting = false;
    }
  }
  
  // Change active backend
  async function changeBackend() {
    await setActiveBackend(activeBackend);
    checkBackendStatus(activeBackend);
    showBackendSelector = false;
    // Reset error and new backend form when changing backend
    errorMessage = '';
    showNewBackendForm = false;
  }
  
  // Toggle backend selector
  function toggleBackendSelector() {
    showBackendSelector = !showBackendSelector;
  }
  
  // Add new backend from login page
  async function addNewBackend() {
    if (!newBackendId || !newBackendUrl) {
      return;
    }
    
    // Add the new backend
    setBackend(newBackendId, newBackendUrl);
    
    // Set it as active
    activeBackend = newBackendId;
    await setActiveBackend(newBackendId);
    
    // Check its status
    checkBackendStatus(newBackendId);
    
    // Reset form and hide it
    showNewBackendForm = false;
    newBackendId = '';
    newBackendUrl = '';
    errorMessage = '';
  }

  function isRunningInElectron() {
  const userAgentCheck = typeof navigator === 'object' && 
    typeof navigator.userAgent === 'string' && 
    navigator.userAgent.indexOf('Electron') >= 0;
  return userAgentCheck;
  }


  // Get current backend URL for display
  let currentBackend = $derived($backendConfig.backends[$backendConfig.active]);
  let backendUrl = $derived(currentBackend?.url || '/');
  let displayUrl = $derived(backendUrl === '/' ? window.location.origin : backendUrl);
  let activeStatus = $derived(backendStatuses[activeBackend]?.status || 'unknown');
  let activeError = $derived(backendStatuses[activeBackend]?.error);
</script>

<div class="login-page">
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h2>Backend Connection Required</h2>
        
        <div class="server-selector">
          <div class="server-info" onclick={toggleBackendSelector}>
            <span class="server-label">Server:</span>
            <span class="server-url">{displayUrl}</span>
            
            <span class="status-indicator {activeStatus}">
              {#if activeStatus === 'online'}
                🟢
              {:else if activeStatus === 'offline'}
                🔴
              {:else if activeStatus === 'error' || activeStatus === 'unreachable'}
                ⚠️
              {:else if activeStatus === 'cert-error'}
                🔒
              {:else}
                ⚪
              {/if}
            </span>
            
            <button class="change-server-btn" title="Change server">
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="6 9 12 15 18 9"></polyline>
              </svg>
            </button>
          </div>
          
          {#if showBackendSelector}
            <div class="backend-dropdown">
              <div class="dropdown-header">
                <h3>Select Backend</h3>
                <button class="close-btn" onclick={() => showBackendSelector = false}>
                  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                  </svg>
                </button>
              </div>
              
              <div class="backend-list">
                {#each backends as backendId}
                  <div class="backend-item">
                    <label class="backend-option">
                      <input 
                        type="radio" 
                        name="backend" 
                        value={backendId} 
                        bind:group={activeBackend}
                      />
                      <div class="backend-details">
                        <span class="backend-name">{backendId}</span>
                        <span class="backend-url">{$backendConfig.backends[backendId].url}</span>
                      </div>
                      
                      <span class="status-indicator {backendStatuses[backendId]?.status || 'unknown'}">
                        {#if backendStatuses[backendId]?.status === 'online'}
                          🟢
                        {:else if backendStatuses[backendId]?.status === 'offline'}
                          🔴
                        {:else if backendStatuses[backendId]?.status === 'error' || backendStatuses[backendId]?.status === 'unreachable'}
                          ⚠️
                        {:else if backendStatuses[backendId]?.status === 'cert-error'}
                          🔒
                        {:else}
                          ⚪
                        {/if}
                      </span>
                    </label>
                  </div>
                {/each}
              </div>
              
              <div class="dropdown-actions">
                <button 
                  class="test-btn" 
                  onclick={() => checkBackendStatus(activeBackend)}
                  disabled={testingBackend}
                >
                  {testingBackend ? 'Testing...' : 'Test Connection'}
                </button>
                
                <button 
                  class="switch-btn"
                  onclick={changeBackend}
                  disabled={activeBackend === $backendConfig.active}
                >
                  Switch Server
                </button>
              </div>
            </div>
          {/if}
        </div>
      </div>
      {#if errorMessage || activeError}
      <div class="error-message">
        {#if errorMessage === 'endpoint not found'}
          {#if isRunningInElectron()}
            You are using the Desktop App, so you must define your Backend.
          {:else}
            The selected server doesn't have the expected login endpoint.
          {/if}
        {:else if activeStatus === 'cert-error'}
          {activeError}
          <a href={backendStatuses[activeBackend].backendUrl} target="_blank" rel="noopener">
            Check Backend
          </a>
        {:else if activeError}
          {activeError}
        {:else}
          {errorMessage}
        {/if}
      </div>
      {/if}
      
      {#if showNewBackendForm}
        <div class="new-backend-form">
          <h3>Add New Server</h3>
          <p>The current server doesn't have the expected login endpoint or is unreachable. Add a new server with the correct authentication endpoint.</p>
          
          <div class="form-group">
            <label for="new-backend-id">Server Name</label>
            <input 
              type="text" 
              id="new-backend-id" 
              bind:value={newBackendId} 
              placeholder="e.g., production"
              disabled={isSubmitting}
            />
          </div>
          
          <div class="form-group">
            <label for="new-backend-url">Server URL</label>
            <input 
              type="text" 
              id="new-backend-url" 
              bind:value={newBackendUrl} 
              placeholder="e.g., https://api.example.com"
              disabled={isSubmitting}
            />
          </div>
          
          <div class="backend-form-actions">
            <button 
              class="cancel-btn" 
              onclick={() => showNewBackendForm = false}
              disabled={isSubmitting}
            >
              Cancel
            </button>
            <button 
              class="add-backend-btn" 
              onclick={addNewBackend}
              disabled={isSubmitting || !newBackendId || !newBackendUrl}
            >
              Add Server
            </button>
          </div>
        </div>
      {:else}
        <form onsubmit={handleSubmit}>
          <div class="form-group">
            <label for="username">Username</label>
            <div class="input-wrapper">
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="input-icon">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
                <circle cx="12" cy="7" r="4"></circle>
              </svg>
              <input 
                type="text" 
                id="username" 
                bind:value={username} 
                disabled={isSubmitting} 
                placeholder="Enter username"
                autocomplete="username"
              />
            </div>
          </div>
          
          <div class="form-group">
            <label for="password">Password</label>
            <div class="input-wrapper">
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="input-icon">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
              </svg>
              <input 
                type="password" 
                id="password" 
                bind:value={password} 
                disabled={isSubmitting} 
                placeholder="Enter password"
                autocomplete="current-password"
              />
            </div>
          </div>
          
          <button type="submit" class="login-button" disabled={isSubmitting || activeStatus === 'unreachable' || activeStatus === 'cert-error'}>
            {#if isSubmitting}
              <svg class="loading-spinner" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 12a9 9 0 1 1-6.219-8.56"></path>
              </svg>
              Logging in...
            {:else}
              Login
            {/if}
          </button>
        </form>
      {/if}
    </div>
  </div>
</div>


<style>
  .login-page {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 100vh;
    background-color: var(--bg-primary);
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
  }
  
  .login-container {
    display: flex;
    justify-content: center;
    align-items: center;
    width: 100%;
    max-width: 460px;
    padding: 1rem;
  }
  
  .login-card {
    background: var(--bg-secondary);
    padding: 2.5rem;
    box-shadow: var(--shadow-medium);
    border-radius: 12px;
    width: 100%;
    transition: transform var(--transition-speed);
  }
  
  .login-card:hover {
    transform: translateY(-5px);
  }
  
  .login-header {
    margin-bottom: 1.5rem;
  }
  
  h2 {
    text-align: center;
    margin-bottom: 1.5rem;
    color: var(--text-primary);
    font-weight: 600;
    font-size: 1.75rem;
  }
  
  h3 {
    color: var(--text-primary);
    font-weight: 500;
    font-size: 1.2rem;
    margin-top: 0;
    margin-bottom: 1rem;
  }
  
  .server-selector {
    position: relative;
    margin-bottom: 1rem;
  }
  
  .server-info {
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--bg-tertiary);
    border-radius: 8px;
    padding: 0.75rem 1rem;
    cursor: pointer;
    transition: background-color var(--transition-speed);
  }
  
  .server-info:hover {
    background-color: var(--bg-hover);
  }
  
  .server-label {
    font-weight: 500;
    color: var(--text-secondary);
    margin-right: 0.5rem;
  }
  
  .server-url {
    flex: 1;
    color: var(--text-primary);
    font-size: 0.9rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .status-indicator {
    margin: 0 0.5rem;
  }
  
  .change-server-btn {
    background: none;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    padding: 0.25rem;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: color var(--transition-speed);
  }
  
  .change-server-btn:hover {
    color: var(--accent-primary);
  }
  
  .backend-dropdown {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    background: var(--bg-secondary);
    border-radius: 8px;
    box-shadow: var(--shadow-medium);
    margin-top: 0.5rem;
    z-index: 10;
    overflow: hidden;
    animation: slideDown 0.2s ease-out;
  }
  
  @keyframes slideDown {
    from {
      opacity: 0;
      transform: translateY(-10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  
  .dropdown-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    border-bottom: 1px solid var(--border-color);
  }
  
  .dropdown-header h3 {
    margin: 0;
    font-size: 1rem;
    font-weight: 500;
  }
  
  .close-btn {
    background: none;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    padding: 0.25rem;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: color var(--transition-speed);
  }
  
  .close-btn:hover {
    color: var(--accent-primary);
  }
  
  .backend-list {
    max-height: 250px;
    overflow-y: auto;
  }
  
  .backend-item {
    padding: 0.25rem 0.5rem;
  }
  
  .backend-option {
    display: flex;
    align-items: center;
    padding: 0.75rem 1rem;
    border-radius: 6px;
    cursor: pointer;
    transition: background-color var(--transition-speed);
  }
  
  .backend-option:hover {
    background-color: var(--bg-hover);
  }
  
  .backend-option input {
    margin-right: 0.75rem;
    accent-color: var(--accent-primary);
  }
  
  .backend-details {
    display: flex;
    flex-direction: column;
    flex: 1;
    overflow: hidden;
  }
  
  .backend-name {
    font-weight: 500;
    color: var(--text-primary);
  }
  
  .backend-url {
    font-size: 0.85rem;
    color: var(--text-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .backend-error {
    font-size: 0.8rem;
    color: var(--text-warning);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .dropdown-actions {
    display: flex;
    justify-content: space-between;
    padding: 1rem;
    border-top: 1px solid var(--border-color);
  }
  
  .test-btn {
    padding: 0.5rem 1rem;
    background-color: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    font-size: 0.9rem;
    font-weight: 500;
    color: var(--text-primary);
    cursor: pointer;
    transition: all var(--transition-speed);
  }
  
  .test-btn:hover {
    background-color: var(--bg-hover);
    border-color: var(--accent-secondary);
  }
  
  .switch-btn {
    padding: 0.5rem 1rem;
    background-color: var(--accent-primary);
    border: none;
    border-radius: 6px;
    font-size: 0.9rem;
    font-weight: 500;
    color: var(--text-primary);
    cursor: pointer;
    transition: background-color var(--transition-speed);
  }
  
  .switch-btn:hover {
    background-color: var(--accent-tertiary);
  }
  
  .switch-btn:disabled {
    background-color: var(--bg-tertiary);
    cursor: not-allowed;
    opacity: 0.6;
  }
  
  .form-group {
    margin-bottom: 1.5rem;
  }
  
  label {
    display: block;
    margin-bottom: 0.5rem;
    color: var(--text-primary);
    font-weight: 500;
    font-size: 0.95rem;
  }
  
  .input-wrapper {
    position: relative;
    display: flex;
    align-items: center;
  }
  
  .input-icon {
    position: absolute;
    left: 1rem;
    color: var(--text-secondary);
  }
  
  input[type="text"],
  input[type="password"] {
    width: 100%;
    padding: 0.85rem 1rem 0.85rem 2.75rem;
    border: 1px solid var(--border-color);
    border-radius: 8px;
    font-size: 1rem;
    background-color: var(--bg-tertiary);
    color: var(--text-primary);
    transition: all var(--transition-speed);
  }
  
  .new-backend-form input[type="text"] {
    padding: 0.85rem 1rem;
  }
  
  input[type="text"]:focus,
  input[type="password"]:focus {
    outline: none;
    border-color: var(--accent-primary);
    box-shadow: 0 0 0 3px rgba(106, 153, 85, 0.2);
  }
  
  .checkbox {
    display: flex;
    align-items: center;
  }
  
  .login-button {
    width: 100%;
    padding: 0.85rem;
    background-color: var(--accent-primary);
    color: var(--text-primary);
    border: none;
    border-radius: 8px;
    font-size: 1rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color var(--transition-speed);
    display: flex;
    justify-content: center;
    align-items: center;
  }
  
  .login-button:hover {
    background-color: var(--accent-tertiary);
  }
  
  .login-button:disabled {
    background-color: var(--bg-tertiary);
    opacity: 0.6;
    cursor: not-allowed;
  }
  
  .loading-spinner {
    animation: spin 1s linear infinite;
    margin-right: 0.5rem;
  }
  
  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
  
  .error-message {
    padding: 0.85rem;
    background-color: rgba(206, 145, 120, 0.1);
    color: var(--text-warning);
    border-radius: 8px;
    margin-bottom: 1.5rem;
    font-size: 0.95rem;
    display: flex;
    align-items: center;
    animation: fadeIn 0.3s ease-out;
  }
  
  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  
  /* Status colors */
  .status-indicator.online {
    color: var(--accent-primary);
  }
  
  .status-indicator.offline {
    color: var(--text-warning);
  }
  
  .status-indicator.error, .status-indicator.unreachable {
    color: #ff9800;
  }
  
  .status-indicator.unknown {
    color: var(--text-secondary);
  }
  
  .new-backend-form {
    background-color: var(--bg-tertiary);
    border-radius: 8px;
    padding: 1.5rem;
    margin-bottom: 1.5rem;
    border: 1px solid var(--border-color);
    animation: fadeIn 0.3s ease-out;
  }
  
  .new-backend-form p {
    color: var(--text-secondary);
    font-size: 0.95rem;
    margin-bottom: 1.25rem;
  }
  
  .backend-form-actions {
    display: flex;
    justify-content: space-between;
    gap: 1rem;
  }
  
  .cancel-btn {
    flex: 1;
    padding: 0.75rem;
    background-color: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    font-size: 0.95rem;
    font-weight: 500;
    color: var(--text-secondary);
    cursor: pointer;
    transition: all var(--transition-speed);
  }
  
  .cancel-btn:hover {
    background-color: var(--bg-hover);
    border-color: var(--accent-secondary);
  }
  
  .add-backend-btn {
    flex: 2;
    padding: 0.75rem;
    background-color: var(--accent-primary);
    color: var(--text-primary);
    border: none;
    border-radius: 6px;
    font-size: 0.95rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color var(--transition-speed);
  }
  
  .add-backend-btn:hover {
    background-color: var(--accent-tertiary);
  }
  
  .add-backend-btn:disabled {
    background-color: var(--bg-tertiary);
    opacity: 0.6;
    cursor: not-allowed;
  }
  
  @media (max-width: 480px) {
    .login-container {
      padding: 0.5rem;
    }
    
    .login-card {
      padding: 1.5rem;
    }
    
    h2 {
      font-size: 1.5rem;
    }
    
    .dropdown-actions {
      flex-direction: column;
      gap: 0.5rem;
    }
    
    .test-btn, .switch-btn {
      width: 100%;
    }
    
    .backend-form-actions {
      flex-direction: column;
    }
    
    .cancel-btn, .add-backend-btn {
      width: 100%;
    }
  }
</style>