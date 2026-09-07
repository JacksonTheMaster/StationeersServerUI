<script>
  import { onMount } from 'svelte';
  import { apiFetch } from '../../services/api';

  // Main state
  let backups = $state([]);
  let isLoading = $state(false);
  let isCreating = $state(false);
  let isRestoring = $state(false);
  let backupStatus = $state({ isRunning: false });
  let selectedBackup = $state(null);
  let showRestoreModal = $state(false);
  let skipPreBackup = $state(false);
  let createMode = $state('tar');
  let error = $state(null);
  let success = $state(null);
  let hasInitialLoad = $state(false);
  let refreshInterval;

  onMount(() => {
    loadBackupStatus();
    loadBackups();
    refreshInterval = setInterval(() => {
        loadBackupStatus();
        loadBackups();
    }, 3000);

    return () => {
      if (refreshInterval) {
        clearInterval(refreshInterval);
      }
    };
  });

  async function loadBackupStatus() {
    try {
      const response = await apiFetch('/api/v3/server/status');
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      
      const data = await response.json();
      backupStatus = { isRunning: Boolean(data?.isRunning) };
    } catch (err) {
      console.error('Failed to load backup status:', err);
      // Silently fail for status updates to avoid UI flicker
      backupStatus = { isRunning: false };
    }
  }

  async function loadBackups() {
  // Don't show loading state for refresh operations after initial load
  const isRefresh = hasInitialLoad && backups.length > 0;
  if (!isRefresh) {
    isLoading = true;
  }
  
  // Only clear error on manual refresh, not auto-refresh
  if (!isRefresh) {
    error = null;
  }
  
  try {
    const response = await apiFetch('/api/v3/backups');
    
    // Always parse the JSON response to get detailed error messages
    const data = await response.json();
    
    if (!response.ok) {
      // Use the detailed error message from the API response
      const errorMessage = (data && data.message) ? 
        data.message : 
        `HTTP ${response.status}: ${response.statusText}`;
      throw new Error(errorMessage);
    }
    
    if (Array.isArray(data)) {
      const newBackups = data.map(backup => ({
        name: backup.name,
        date: new Date(backup.saveTime).toLocaleDateString(),
        time: new Date(backup.saveTime).toLocaleTimeString(),
        displayName: backup.summary?.worldName || backup.name,
        size: 'Unknown size'
      }));
      
      if (JSON.stringify(newBackups) !== JSON.stringify(backups)) {
        backups = newBackups;
      }
      
      if (error && !isRefresh) {
        error = null;
      }
    } else {
      if (!isRefresh) {
        error = (data && data.message) ? data.message : 'Failed to load backups';
      }
    }
  } catch (err) {
    console.error('Failed to load backups:', err);
    if (!isRefresh) {
      error = 'Failed to load backups: ' + err.message;
    }
  } finally {
    if (!isRefresh) {
      isLoading = false;
    }
    hasInitialLoad = true;
    }
  }

  async function createBackup() {
    error = 'SSUI archives completed game autosaves automatically.';
  }

  async function restoreBackup() {
    if (!selectedBackup) return;

    isRestoring = true;
    error = null;
    success = null;

    try {
      const selection = new URLSearchParams({ name: selectedBackup.name });
      const response = await apiFetch(`/api/v3/backups/restore?${selection}`, {
        method: 'POST',
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      if (data?.message) {
        success = data.message;
        showRestoreModal = false;
        selectedBackup = null;
        await loadBackups(); // Refresh the list
      } else {
        error = (data && data.message) ? data.message : 'Failed to restore backup';
      }
    } catch (err) {
      error = 'Failed to restore backup: ' + err.message;
    } finally {
      isRestoring = false;
    }
  }

  function openRestoreModal(backup) {
    selectedBackup = backup;
    showRestoreModal = true;
    skipPreBackup = false;
  }

  function closeRestoreModal() {
    showRestoreModal = false;
    selectedBackup = null;
  }

  function clearMessages() {
    error = null;
    success = null;
  }

  function manualRefresh() {
    hasInitialLoad = false;
    loadBackups();
  }

  function formatFileSize(bytes) {
    if (bytes === 0 || bytes === 'Unknown') return 'Unknown';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }
</script>

<div class="backup-container">
  <div class="backup-header">
    <div class="header-info">
      <h2>Backup Manager</h2>
      <div class="status-indicator">
        <span class="status-dot {backupStatus.isRunning ? 'running' : 'idle'}"></span>
        <span class="status-text">
          {backupStatus.isRunning ? 'Backup or compression in progress...' : 'Ready'}
        </span>
      </div>
    </div>
    
    <div class="header-actions">
      <button 
        class="refresh-button" 
        onclick={manualRefresh}
        disabled={isLoading}
      >
        <span class="button-icon">↻</span>
        {isLoading ? 'Loading...' : 'Refresh'}
      </button>
    </div>
  </div>

  <!-- Messages -->
  {#if error}
    <div class="message error">
      <span>{error}</span>
      <button class="message-close" onclick={clearMessages}>×</button>
    </div>
  {/if}

  {#if success}
    <div class="message success">
      <span>{success}</span>
      <button class="message-close" onclick={clearMessages}>×</button>
    </div>
  {/if}

  <!-- Create Backup Section -->
  <div class="backup-section">
    <h3>Create New Backup</h3>
    <div class="create-backup-form">
      <div class="form-group">
        <label for="backup-mode">Backup Mode:</label>
        <select id="backup-mode" bind:value={createMode}>
          <option value="tar">TAR Archive</option>
          <option value="copy">Full Snapshot</option>
        </select>
      </div>
      
      <button 
        class="create-button"
        onclick={createBackup}
        disabled={isCreating || backupStatus.isRunning}
      >
        <span class="button-icon">+</span>
        {isCreating ? 'Creating...' : 'Create Backup'}
      </button>
    </div>
  </div>

  <!-- Backup List Section -->
  <div class="backup-section">
    <h3>Available Backups ({backups.length})</h3>
    
    {#if isLoading && !hasInitialLoad}
      <div class="loading-message">Loading backups...</div>
    {:else if error && backups.length === 0}
      <div class="error-state">
        <span class="error-icon">⚠️</span>
        <p>Unable to load backups</p>
        <p class="error-subtitle">Please check your connection and Backup path Settings</p>
        <button class="retry-button" onclick={manualRefresh} disabled={isLoading}>
          {isLoading ? 'Retrying...' : 'Retry'}
        </button>
      </div>
    {:else if backups.length === 0}
      <div class="no-backups">
        <span class="no-backups-icon">📦</span>
        <p>No backups found</p>
        <p class="no-backups-subtitle">Create your first backup to get started</p>
      </div>
    {:else}
      <div class="backups-list">
        {#each backups as backup}
          <div class="backup-item">
            <div class="backup-info">
              <div class="backup-name">{backup.displayName}</div>
              <div class="backup-details">
                <span class="backup-filename">{backup.name}</span>
                <span class="backup-size">{backup.size}</span>
              </div>
            </div>
            
            <div class="backup-actions">
              <button 
                class="restore-button"
                onclick={() => openRestoreModal(backup)}
                disabled={isRestoring || backupStatus.isRunning}
              >
                <span class="button-icon">↶</span>
                Restore
              </button>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<!-- Restore Modal -->
{#if showRestoreModal}
  <div class="modal-overlay" onclick={closeRestoreModal}>
    <div class="modal" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h3>Restore Backup</h3>
        <button class="modal-close" onclick={closeRestoreModal}>×</button>
      </div>
      
      <div class="modal-content">
        <p>Are you sure you want to restore this backup?</p>
        <div class="backup-preview">
          <strong>{selectedBackup?.displayName}</strong>
          <br>
          <small>{selectedBackup?.name}</small>
        </div>
        
        <div class="restore-options">
          <label class="checkbox-label">
            <input 
              type="checkbox" 
              bind:checked={skipPreBackup}
            />
            Skip pre-restore backup
          </label>
          <small class="option-help">
            If unchecked, a backup will be created before restoring
          </small>
        </div>
        
        <div class="warning">
          <span class="warning-icon">⚠️</span>
          <span>This action will overwrite current data. This cannot be undone.</span>
        </div>
      </div>
      
      <div class="modal-actions">
        <button class="cancel-button" onclick={closeRestoreModal}>
          Cancel
        </button>
        <button 
          class="confirm-restore-button"
          onclick={restoreBackup}
          disabled={isRestoring}
        >
          <span class="button-icon">↶</span>
          {isRestoring ? 'Restoring...' : 'Restore Backup'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .backup-container {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    height: 100%;
  }

  .backup-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    background-color: var(--bg-secondary);
    padding: 1rem;
    border-radius: 4px;
  }

  .header-info h2 {
    margin: 0 0 0.5rem 0;
    color: var(--text-primary);
  }

  .status-indicator {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background-color: var(--text-secondary);
  }

  .status-dot.running {
    background-color: #4caf50;
    animation: pulse 2s infinite;
  }

  .status-dot.idle {
    background-color: #2196f3;
  }

  @keyframes pulse {
    0% { opacity: 1; }
    50% { opacity: 0.5; }
    100% { opacity: 1; }
  }

  .status-text {
    color: var(--text-secondary);
    font-size: 0.9rem;
  }

  .header-actions {
    display: flex;
    gap: 0.5rem;
  }

  .refresh-button {
    background-color: var(--bg-tertiary);
    color: var(--text-primary);
    border: 1px solid var(--border-color);
    padding: 0.5rem 1rem;
    border-radius: 4px;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .refresh-button:hover:not(:disabled) {
    background-color: var(--bg-hover);
  }

  .refresh-button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .button-icon {
    font-size: 1rem;
  }

  .message {
    padding: 1rem;
    border-radius: 4px;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .message.error {
    background-color: rgba(244, 67, 54, 0.1);
    border: 1px solid rgba(244, 67, 54, 0.3);
    color: #e57373;
  }

  .message.success {
    background-color: rgba(76, 175, 80, 0.1);
    border: 1px solid rgba(76, 175, 80, 0.3);
    color: #81c784;
  }

  .message-close {
    background: none;
    border: none;
    color: inherit;
    font-size: 1.2rem;
    cursor: pointer;
    padding: 0;
    width: 20px;
    height: 20px;
  }

  .backup-section {
    background-color: var(--bg-secondary);
    padding: 1.5rem;
    border-radius: 4px;
  }

  .backup-section h3 {
    margin: 0 0 1rem 0;
    color: var(--text-primary);
  }

  .create-backup-form {
    display: flex;
    gap: 1rem;
    align-items: end;
    flex-wrap: wrap;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .form-group label {
    color: var(--text-secondary);
    font-size: 0.9rem;
  }

  .form-group select {
    background-color: var(--bg-tertiary);
    color: var(--text-primary);
    border: 1px solid var(--border-color);
    padding: 0.5rem;
    border-radius: 4px;
    min-width: 150px;
  }

  .create-button {
    background-color: var(--accent-primary);
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-weight: 500;
  }

  .create-button:hover:not(:disabled) {
    background-color: var(--accent-secondary);
  }

  .create-button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .loading-message {
    text-align: center;
    color: var(--text-secondary);
    padding: 2rem;
  }

  .error-state {
    text-align: center;
    color: var(--text-secondary);
    padding: 3rem 1rem;
  }

  .error-icon {
    font-size: 3rem;
    display: block;
    margin-bottom: 1rem;
  }

  .error-state p {
    margin: 0.5rem 0;
    color: var(--text-primary);
  }

  .error-subtitle {
    font-size: 0.9rem;
    opacity: 0.7;
  }

  .retry-button {
    background-color: var(--accent-primary);
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    cursor: pointer;
    margin-top: 1rem;
  }

  .retry-button:hover:not(:disabled) {
    background-color: var(--accent-secondary);
  }

  .retry-button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .no-backups {
    text-align: center;
    color: var(--text-secondary);
    padding: 3rem 1rem;
  }

  .no-backups-icon {
    font-size: 3rem;
    display: block;
    margin-bottom: 1rem;
  }

  .no-backups p {
    margin: 0.5rem 0;
  }

  .no-backups-subtitle {
    font-size: 0.9rem;
    opacity: 0.7;
  }

  .backups-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .backup-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    background-color: var(--bg-tertiary);
    border-radius: 4px;
    border: 1px solid var(--border-color);
  }

  .backup-item:hover {
    background-color: var(--bg-hover);
  }

  .backup-info {
    flex: 1;
  }

  .backup-name {
    font-weight: 500;
    color: var(--text-primary);
    margin-bottom: 0.25rem;
  }

  .backup-details {
    display: flex;
    gap: 1rem;
    font-size: 0.8rem;
    color: var(--text-secondary);
  }

  .backup-actions {
    display: flex;
    gap: 0.5rem;
  }

  .restore-button {
    background-color: var(--bg-tertiary);
    color: var(--text-primary);
    border: 1px solid var(--border-color);
    padding: 0.4rem 0.8rem;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.9rem;
    display: flex;
    align-items: center;
    gap: 0.25rem;
  }

  .restore-button:hover:not(:disabled) {
    background-color: var(--accent-primary);
    color: white;
  }

  .restore-button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  /* Modal styles */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal {
    background-color: var(--bg-secondary);
    border-radius: 8px;
    min-width: 400px;
    max-width: 90vw;
    max-height: 90vh;
    overflow: hidden;
    border: 1px solid var(--border-color);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    border-bottom: 1px solid var(--border-color);
  }

  .modal-header h3 {
    margin: 0;
    color: var(--text-primary);
  }

  .modal-close {
    background: none;
    border: none;
    color: var(--text-secondary);
    font-size: 1.5rem;
    cursor: pointer;
    padding: 0;
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .modal-close:hover {
    color: var(--text-primary);
  }

  .modal-content {
    padding: 1.5rem;
  }

  .modal-content p {
    margin: 0 0 1rem 0;
    color: var(--text-primary);
  }

  .backup-preview {
    background-color: var(--bg-tertiary);
    padding: 1rem;
    border-radius: 4px;
    margin: 1rem 0;
    color: var(--text-primary);
  }

  .restore-options {
    margin: 1.5rem 0;
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    color: var(--text-primary);
    cursor: pointer;
  }

  .checkbox-label input[type="checkbox"] {
    accent-color: var(--accent-primary);
  }

  .option-help {
    display: block;
    margin-top: 0.5rem;
    color: var(--text-secondary);
    font-size: 0.8rem;
  }

  .warning {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 1rem;
    background-color: rgba(255, 152, 0, 0.1);
    border: 1px solid rgba(255, 152, 0, 0.3);
    border-radius: 4px;
    color: #ffb74d;
    font-size: 0.9rem;
  }

  .warning-icon {
    font-size: 1.1rem;
  }

  .modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    padding: 1rem;
    border-top: 1px solid var(--border-color);
  }

  .cancel-button {
    background-color: var(--bg-tertiary);
    color: var(--text-primary);
    border: 1px solid var(--border-color);
    padding: 0.5rem 1rem;
    border-radius: 4px;
    cursor: pointer;
  }

  .cancel-button:hover {
    background-color: var(--bg-hover);
  }

  .confirm-restore-button {
    background-color: #f44336;
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-weight: 500;
  }

  .confirm-restore-button:hover:not(:disabled) {
    background-color: #d32f2f;
  }

  .confirm-restore-button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>
