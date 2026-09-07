import { writable } from 'svelte/store';
import { apiFetch } from './api.js';

/**
 * User information store
 */
export const userInfo = writable({
  username: null,
  accessLevel: null,
  isLoading: true,
  isAuthenticated: false,
  lastFetched: null,
  error: null
});

/**
 * Fetch current user information from the API
 */
export async function fetchUserInfo() {
  // Set loading state
  userInfo.update(state => ({
    ...state,
    isLoading: true,
    error: null
  }));

  try {
    const response = await apiFetch('/api/v3/auth/session');
    
    // Parse the JSON from the response
    const data = await response.json();
    console.log('You are logged in as:', data);
    
    if (data && data.user?.username) {
      // Update store with successful data
      userInfo.set({
        username: data.user.username,
        accessLevel: data.permissions?.includes('security.manage') ? 'owner' : 'user',
        isLoading: false,
        isAuthenticated: true,
        lastFetched: new Date(),
        error: null
      });
      return data;
    } else {
      throw new Error('Invalid response format: missing username');
    }
  } catch (error) {
    console.error('Failed to fetch user info:', error);
    
    // Update store with error state
    userInfo.set({
      username: null,
      accessLevel: null,
      isLoading: false,
      isAuthenticated: false,
      lastFetched: null,
      error: error.message || 'Failed to fetch user information'
    });
    
    throw error;
  }
}

/**
 * Get formatted user initials for avatar display
 */
export function getUserInitials(username) {
  if (!username) return 'USR';
  
  // Split username and take first letter of each word, max 3 characters
  const words = username.split(/[\s_-]+/);
  if (words.length === 1) {
    return username.substring(0, 3).toUpperCase();
  }
  return words.slice(0, 2).map(word => word.charAt(0).toUpperCase()).join('');
}

/**
 * Format access level for display
 */
export function formatAccessLevel(accessLevel) {
  if (!accessLevel) return 'Unknown';
  return accessLevel.charAt(0).toUpperCase() + accessLevel.slice(1);
}

/**
 * Check if user info needs to be refreshed (older than 5 minutes)
 */
export function shouldRefreshUserInfo(lastFetched) {
  if (!lastFetched) return true;
  const fiveMinutesAgo = new Date(Date.now() - 5 * 60 * 1000);
  return lastFetched < fiveMinutesAgo;
}

/**
 * Initialize user info on app start
 */
export function initUserInfo() {
  return fetchUserInfo().catch(error => {
    console.warn('Initial user info fetch failed:', error);
  });
}

/**
 * Clear user info (for logout)
 */
export function clearUserInfo() {
  userInfo.set({
    username: null,
    accessLevel: null,
    isLoading: false,
    isAuthenticated: false,
    lastFetched: null,
    error: null
  });
}
