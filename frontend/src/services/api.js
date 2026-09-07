// api.js

import { writable, get } from 'svelte/store';

// Store for backend configuration
export const backendConfig = writable({
  active: 'default', // Currently active backend
  backends: {
    default: {
      url: '/', // Default backend URL is the current host
      csrf: null
    }
  }
});

// Store for authentication state
export const authState = writable({
  isAuthenticated: false,
  isAuthenticating: false,
  authError: null
});

// Track initialization state
let isInitialized = false;

// Helper to get the current backend configuration
export function getCurrentBackend() {
  const config = get(backendConfig);
  return config.backends[config.active] || config.backends.default;
}

// Helper to get the current backend URL
export function getCurrentBackendUrl() {
  const backend = getCurrentBackend();
  return backend.url === '/' ? '' : backend.url;
}

// Helper to get the current authentication token
export function getCurrentCSRF() {
  return getCurrentBackend().csrf;
}

// Add or update a backend
export function setBackend(id, url) {
  backendConfig.update(config => {
    const csrf = config.backends[id]?.csrf || null;
    config.backends[id] = { url, csrf };
    return config;
  });
}

// Set the active backend and verify authentication
export async function setActiveBackend(id) {
  let success = false;
  let prevBackendId = get(backendConfig).active;
  
  // Only update if different to prevent unnecessary reloads
  if (prevBackendId !== id) {
    backendConfig.update(config => {
      if (config.backends[id]) {
        config.active = id;
      }
      return config;
    });
    
    // After changing backend, check authentication status
    try {
      await syncAuthState();
      success = true;
    } catch (error) {
      console.error('Error syncing auth state after backend change:', error);
    }
  } else {
    // If selecting the same backend, consider it successful
    success = true;
  }
  
  return success;
}

// Update the token for a backend and persist it
export function updateSession(id, csrf) {
  backendConfig.update(config => {
    if (config.backends[id]) {
      config.backends[id].csrf = csrf;
    }
    return config;
  });
  
  // Update the auth state
  if (id === get(backendConfig).active) {
    authState.update(state => ({
      ...state,
      isAuthenticated: !!csrf,
      authError: null
    }));
  }
}

// Clear authentication for the current backend
export async function clearAuthentication() {
  const currentBackendId = get(backendConfig).active;
  try {
    await apiFetch('/api/v3/auth/logout', { method: 'POST' });
  } finally {
    updateSession(currentBackendId, null);
  }
}

/**
 * Fetch wrapper that automatically adds the backend URL and handles authentication
 * @param {string} endpoint - The API endpoint (e.g., "/api/v3/server/status")
 * @param {Object} options - Fetch options
 * @returns {Promise} - The fetch promise
 */
export async function apiFetch(endpoint, options = {}) {
  // Get the current backend configuration
  const backendUrl = getCurrentBackendUrl();
  const csrf = getCurrentCSRF();
  
  // Ensure endpoint starts with "/" if it's not an empty string
  const normalizedEndpoint = endpoint.startsWith('/') || endpoint === '' ? endpoint : `/${endpoint}`;
  
  // Construct the full URL
  const url = `${backendUrl}${normalizedEndpoint}`;
  
  // Set up headers if not provided
  options.headers = options.headers || {};
  
  // Always include credentials for CORS requests
  options.credentials = 'include';
  
  const method = String(options.method || 'GET').toUpperCase();
  if (csrf && !['GET', 'HEAD', 'OPTIONS'].includes(method)) {
    options.headers['X-SSUI-CSRF'] = csrf;
  }
  
  const response = await fetch(url, options);
  const readJSON = response.json.bind(response);
  const readText = response.text.bind(response);
  response.json = async () => {
    const envelope = await readJSON();
    if (envelope && Object.prototype.hasOwnProperty.call(envelope, 'data')) return envelope.data;
    if (envelope?.error) return { error: envelope.error.message, errorCode: envelope.error.code };
    return envelope;
  };
  response.text = async () => {
    const text = await readText();
    try {
      const envelope = JSON.parse(text);
      if (envelope?.error?.message) return envelope.error.message;
      if (envelope?.data?.message) return envelope.data.message;
    } catch (_) {
      // Regular text response.
    }
    return text;
  };
  return response;
}

/**
 * Fetch wrapper with timeout that automatically adds the backend URL and handles authentication
 * @param {string} endpoint - The API endpoint (e.g., "/api/v3/server/status")
 * @param {Object} options - Fetch options
 * @param {number} timeoutMs - Timeout in milliseconds
 * @returns {Promise} - The fetch promise
 */
export async function apiFetchTimeout(endpoint, options = {}, timeoutMs) {
  // Create a new options object to avoid modifying the original
  const timeoutOptions = { ...options };
  
  // Set up AbortController for timeout
  const controller = new AbortController();
  timeoutOptions.signal = controller.signal;
  
  // Set timeout
  const timeoutId = setTimeout(() => controller.abort(), timeoutMs);
  
  try {
    // Use apiFetch for the actual request
    const response = await apiFetch(endpoint, timeoutOptions);
    return response;
  } finally {
    clearTimeout(timeoutId);
  }
}

/**
 * Helper for JSON API calls
 * @param {string} endpoint - The API endpoint
 * @param {Object} options - Fetch options
 * @returns {Promise<Object>} - Parsed JSON response
 */
export async function apiJson(endpoint, options = {}) {
  // Default to JSON content type if not specified
  options.headers = options.headers || {};
  if (!options.headers['Content-Type'] && options.method && options.method !== 'GET') {
    options.headers['Content-Type'] = 'application/json';
  }
  
  const response = await apiFetch(endpoint, options);
  
  // Check for auth errors
  if (response.status === 401) {
    // Update auth state
    authState.update(state => ({
      ...state,
      isAuthenticated: false,
      authError: 'Unauthorized'
    }));
    throw new Error('Authentication required');
  }
  
  if (!response.ok) {
    throw new Error(`API error: ${response.status} ${response.statusText}`);
  }
  
  return response.json();
}

/**
 * Helper for text API calls
 * @param {string} endpoint - The API endpoint
 * @param {Object} options - Fetch options
 * @returns {Promise<string>} - Text response
 */
export async function apiText(endpoint, options = {}) {
  const response = await apiFetch(endpoint, options);
  
  // Check for auth errors
  if (response.status === 401) {
    // Update auth state
    authState.update(state => ({
      ...state,
      isAuthenticated: false,
      authError: 'Unauthorized'
    }));
    throw new Error('Authentication required');
  }
  
  if (!response.ok) {
    throw new Error(`API error: ${response.status} ${response.statusText}`);
  }
  
  return response.text();
}

/**
 * Helper for Server-Sent Events (SSE)
 * @param {string} endpoint - The SSE endpoint
 * @param {function} onMessage - Callback for each message
 * @param {function} onError - Error callback
 * @returns {Object} - Control object with a close() method
 */
export function apiSSE(endpoint, onMessage, onError = console.error) {
  // Get the current backend URL
  const backendUrl = getCurrentBackendUrl();
  
  // Ensure endpoint starts with "/" if it's not an empty string
  const normalizedEndpoint = endpoint.startsWith('/') || endpoint === '' ? endpoint : `/${endpoint}`;
  
  // Construct the full URL 
  const baseUrl = backendUrl || window.location.origin;
  const url = new URL(`${baseUrl}${normalizedEndpoint}`);
  
  let eventSource = null;
  let isActive = true;
  let currentBackendId = get(backendConfig).active;
  
  try {
    // Create EventSource for SSE with withCredentials to send cookies
    const eventSourceOptions = { withCredentials: true };
    eventSource = new EventSource(url.toString(), eventSourceOptions);
    
    // Set up event handlers
    eventSource.onmessage = event => {
      try {
        // Try to parse as JSON first
        const data = JSON.parse(event.data);
        onMessage(data);
      } catch (e) {
        // If not JSON, pass the raw string
        onMessage(event.data);
      }
    };
    
    eventSource.onerror = error => {
      // Check if the error might be an authentication issue
      if (eventSource.readyState === EventSource.CLOSED) {
        // Update auth state if we suspect auth issues
        syncAuthState().catch(console.error);
      }
      onError(error);
      
      // Auto-reconnect after a delay if still active
      if (isActive && !eventSource) {
        setTimeout(() => {
          if (isActive && document.visibilityState !== 'hidden') {
            // Try to reconnect with a fresh EventSource
            try {
              eventSource = new EventSource(url.toString(), eventSourceOptions);
            } catch (reconnectError) {
              onError(reconnectError);
            }
          }
        }, 2000);
      }
    };
    
    // Subscribe to backend config changes to handle backend changes without page reload
    const unsubscribe = backendConfig.subscribe(config => {
      if (isActive && config.active !== currentBackendId) {
        console.log('The Backend changed, I am reconnecting a SSE connection');
        currentBackendId = config.active;
        
        // Close existing connection
        if (eventSource) {
          eventSource.close();
          eventSource = null;
        }
        
        // Create a new connection with updated backend info
        setTimeout(() => {
          if (isActive) {
            try {
              // Get fresh URL and token from new backend
              const newBackendUrl = getCurrentBackendUrl();
              const newUrl = new URL(`${newBackendUrl || window.location.origin}${normalizedEndpoint}`);
              
              // Create new EventSource
              eventSource = new EventSource(newUrl.toString(), eventSourceOptions);
              
              // Set up event handlers again
              eventSource.onmessage = eventSource.onmessage;
              eventSource.onerror = eventSource.onerror;
            } catch (reconnectError) {
              onError(reconnectError);
            }
          }
        }, 100);
      }
    });
    
    // Return control object with enhanced close method
    return {
      close: () => {
        if (eventSource) {
          eventSource.close();
          eventSource = null;
        }
        isActive = false;
        unsubscribe();
      }
    };
  } catch (error) {
    onError(error);
    isActive = false;
    // Return a dummy control object
    return {
      close: () => {}
    };
  }
}

/**
 * Login to the current backend
 * @param {string} username - Username
 * @param {string} password - Password
 * @returns {Promise<boolean>} - Success status
 */
export async function login(username, password) {
  // Set authenticating state
  authState.update(state => ({
    ...state,
    isAuthenticating: true,
    authError: null
  }));
  
  try {
    const response = await apiFetch('/api/v3/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      credentials: 'include', // Ensure cookies are stored
      body: JSON.stringify({ username, password })
    });
    
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Authentication failed');
    }
    
    const data = await response.json();
    
    updateSession(get(backendConfig).active, data.csrf);
    
    // Update auth state
    authState.update(state => ({
      ...state,
      isAuthenticated: true,
      isAuthenticating: false,
      authError: null
    }));
    
    return true;
  } catch (error) {
    // Update auth state with error
    authState.update(state => ({
      ...state,
      isAuthenticated: false,
      isAuthenticating: false,
      authError: error.message
    }));
    
    return false;
  }
}

export async function syncAuthState() {
  const currentBackendId = get(backendConfig).active;
  const backend = getCurrentBackend();
  
  // Update auth state to checking
  authState.update(state => ({
    ...state,
    isAuthenticating: true
  }));
  
  try {
    // Make a simple request with 500ms timeout to verify authentication
    const response = await apiFetchTimeout('/api/v3/auth/session', {
      method: 'GET',
      headers: {
        'Accept': 'application/json'
      },
      credentials: 'include' // Ensure cookies are sent
    }, 500);
    
    if (response.status === 401) {
      // Authentication required but we're not authenticated
      authState.update(state => ({
        ...state,
        isAuthenticated: false,
        isAuthenticating: false,
        authError: 'Authentication required'
      }));
      return false;
    } else if (response.status === 404) {
      // Endpoint not found - this server might not require authentication
      // or is using a different auth endpoint
      authState.update(state => ({
        ...state,
        isAuthenticated: false,
        isAuthenticating: false,
        authError: 'endpoint not found'
      }));
      return false;
    } else if (!response.ok) {
      // Some other error
      authState.update(state => ({
        ...state,
        isAuthenticated: false,
        isAuthenticating: false,
        authError: `API error: ${response.status} ${response.statusText}`
      }));
      return false;
    }
    
    // Successfully authenticated
    authState.update(state => ({
      ...state,
      isAuthenticated: true,
      isAuthenticating: false,
      authError: null
    }));
    return true;
  } catch (error) {
    // Handle timeout or other errors
    const errorMessage = error.name === 'AbortError' ? 'Connection timed out. The server may be slow or unreachable.' : error.message || 'Connection error';
    console.warn('Auth check failed:', error);
    
    // Sleep for 60ms, then retry once, else continue failing
    await new Promise(resolve => setTimeout(resolve, 60));
    try {
      const retryResponse = await apiFetchTimeout('/api/v3/auth/session', {
      method: 'GET',
      headers: {
        'Accept': 'application/json'
      },
      credentials: 'include'
      }, 500);

      if (retryResponse.ok) {
      authState.update(state => ({
        ...state,
        isAuthenticated: true,
        isAuthenticating: false,
        authError: null
      }));
      return true;
      }
    } catch (retryError) {
      console.warn('Retry auth check failed:', retryError);
    }

    authState.update(state => ({
      ...state,
      isAuthenticated: false,
      isAuthenticating: false,
      authError: errorMessage
    }));
    return false;
  }
}

// Initial setup function to load saved backend configurations
export function initializeApiService() {
  if (isInitialized) return;
  
  try {
    // Try to load saved config from localStorage
    const savedConfig = localStorage.getItem('ssui-backend-config');
    if (savedConfig) {
      const parsed = JSON.parse(savedConfig);
      
      // Validate and normalize the loaded configuration
      const validatedConfig = {
        active: parsed.active && parsed.backends?.[parsed.active] ? parsed.active : 'default',
        backends: {
          default: {
            url: '/',
            csrf: parsed.backends?.default?.csrf || null
          }
        }
      };

      // Merge all backends from storage
      if (parsed.backends) {
        for (const [id, backend] of Object.entries(parsed.backends)) {
          if (id !== 'default') {
            validatedConfig.backends[id] = {
              url: backend.url,
              csrf: backend.csrf || null
            };
          }
        }
      }

      // Apply the validated config
      backendConfig.set(validatedConfig);
    }

    // Subscribe to changes and save to localStorage
    const unsubscribe = backendConfig.subscribe(value => {
      try {
        localStorage.setItem('ssui-backend-config', JSON.stringify(value));
      } catch (error) {
        console.error('Error saving backend config:', error);
      }
    });

    isInitialized = true;
    
    // Check authentication status
    syncAuthState().catch(console.error);
    
    // Return cleanup function (though this is a service, so it won't typically be cleaned up)
    return unsubscribe;
  } catch (error) {
    console.error('Error initializing API service:', error);
    isInitialized = true; // Prevent repeated failed initializations
  }
}
