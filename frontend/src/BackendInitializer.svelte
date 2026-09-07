<script>
  import { onMount } from 'svelte';
  import { initializeApiService, syncAuthState, apiFetch, authState } from './services/api';
  import InitializingView from './components/resuables/InitializingView.svelte';
  
  /**
   * @typedef {Object} Props
   * @property {any} [onStatusChange] - Default to a no-op function if no handler is provided
   * @property {import('svelte').Snippet} [children]
   */

  /** @type {Props} */
  let { onStatusChange = (statusObj) => {}, children } = $props();
  let isInitialized = $state(false);
  let serverStatus = $state('checking'); // 'checking', 'online', 'offline', 'error', 'cert-error', 'unreachable'
  let errorMessage = $state(null);
  
  // Add AbortSignal polyfill for older browsers
  if (!AbortSignal.timeout) {
    AbortSignal.timeout = function timeout(ms) {
      const controller = new AbortController();
      setTimeout(() => controller.abort(), ms);
      return controller.signal;
    };
  }
  
  // Check if the server is actually available
  async function checkServerAvailability() {
    try {
      // For the server status check, we don't want to use the standard apiFetch
      // that might include auth tokens, as we're testing basic connectivity
      const currentUrl = new URL(window.location.href);
      const baseUrl = `${currentUrl.protocol}//${currentUrl.host}`;
      
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 5000);
      
      const response = await fetch(`${baseUrl}/api/v3/server/status`, {
        method: 'GET',
        signal: controller.signal
      });
      
      clearTimeout(timeoutId);
      
      if (response.status === 404) {
        // API exists but endpoint not found - still considered available
        serverStatus = 'offline';
        errorMessage = null;
        authState.update(state => ({
        ...state,
        authError: 'endpoint not found'
      }));
        return true;
      } else if (response.ok) {
        serverStatus = 'online';
        return true;
      } else if (response.status === 401 || response.status === 403) {
        // Authentication error means server is online but needs auth
        serverStatus = 'online';
        return true;
      } else {
        serverStatus = 'error';
        errorMessage = `Server error: ${response.status} ${response.statusText}`;
        return false;
      }
    } catch (error) {
      if (error.name === 'AbortError') {
        serverStatus = 'offline';
        errorMessage = 'Connection timed out. The server may be slow or unreachable.';
      } else if (error.message.includes('certificate') || error.message.includes('SSL') || error.message.includes('ERR_CERT')) {
        serverStatus = 'cert-error';
        errorMessage = 'Certificate error. The server may be using an invalid or self-signed certificate. Try accepting the certificate in your browser.';
      } else if (error.message.includes('Failed to fetch') || error.message.includes('NetworkError')) {
        serverStatus = 'unreachable';
        errorMessage = 'Server not found. The server may be down or the URL may be incorrect.';
      } else {
        serverStatus = 'offline';
        errorMessage = error.message || 'Cannot connect to server';
      }
      return false;
    }
  }
  
  onMount(async () => {
    // Initialize the API service
    initializeApiService();
    
    // First check if the server is available at all
    const isAvailable = await checkServerAvailability();
    
    // Only check auth if the server is available
    if (isAvailable) {
      try {
        await syncAuthState();
      } catch (error) {
        console.warn('Auth check failed:', error);
        // Auth check failure doesn't mean the server is offline,
        // it just means we're not authenticated or there's an auth issue
      }
    }
    
    // Notify parent component of status change
    onStatusChange({
      status: serverStatus,
      error: errorMessage
    });
    
    isInitialized = true;
  });
</script>

{#if isInitialized}
  {@render children?.()}
{:else}
  <InitializingView {serverStatus} {errorMessage} />
{/if}