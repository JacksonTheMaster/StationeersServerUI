<!-- UserSettings.svelte -->
<script>
    import { apiFetch } from '../../services/api';
        
    // State management
    let statusMessage = $state('');
    let isError = $state(false);
    let statusTimeout;
    
    // Form state
    let username = $state('');
    let password = $state('');
    let confirmPassword = $state('');
    let accessLevel = $state('superadmin');
    
    // Clear all form fields
    function clearForm() {
      username = '';
      password = '';
      confirmPassword = '';
      accessLevel = 'superadmin';
    }
    
    // Add or change user
    async function addOrChangeUser() {
      // Validation
      if (!username.trim()) {
        showStatus('Username is required', true);
        return;
      }
      
      if (!password) {
        showStatus('Password is required', true);
        return;
      }
      
      if (password !== confirmPassword) {
        showStatus('Passwords do not match', true);
        return;
      }
      
      if (password.length < 10) {
        showStatus('Password must be at least 10 characters long', true);
        return;
      }
      
      try {
        const response = await apiFetch('/api/v3/auth/users', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ 
            username: username.trim(), 
            password: password,
            groupIds: ['system-owner']
          })
        });
        
        const result = await response.json();
        
        if (!response.ok) {
          showStatus(`Failed to add/change user: ${result.error || 'Unknown error'}`, true);
          return;
        }
        
        showStatus(`User "${username}" updated successfully`, false);
        // Clear form on success
        clearForm();
      } catch (e) {
        showStatus(`Error updating user: ${e.message}`, true);
      }
    }
    
    // Show status message
    function showStatus(message, error) {
      statusMessage = message;
      isError = error;
      
      // Clear any existing timeout
      if (statusTimeout) clearTimeout(statusTimeout);
      
      // Auto-hide after 30 seconds
      statusTimeout = setTimeout(() => {
        statusMessage = '';
      }, 30000);
    }
    
    // Handle form submission
    function handleSubmit(event) {
      event.preventDefault();
      addOrChangeUser();
    }
  </script>
  
  <div class="settings-container">
    <h2>Admin: User Settings</h2>
    
    <p class="settings-intro">
      DO NOT USE THIS; REPLACES CURRENT USER IN STATIONEERSERVERUI!
    </p>
    
    <div class="settings-group">
      <h3>Add or Change User</h3>
      <form onsubmit={handleSubmit} class="user-form">
        <div class="form-row">
          <div class="form-field">
            <label for="username">Username</label>
            <input 
              id="username"
              type="text" 
              bind:value={username}
              placeholder="Enter username"
              class="text-input"
              required
            />
          </div>
        </div>
        
        <div class="form-row">
          <div class="form-field">
            <label for="password">Password</label>
            <input 
              id="password"
              type="password" 
              bind:value={password}
              placeholder="Enter password"
              class="text-input"
              required
            />
          </div>
        </div>
        
        <div class="form-row">
          <div class="form-field">
            <label for="confirm-password">Confirm Password</label>
            <input 
              id="confirm-password"
              type="password" 
              bind:value={confirmPassword}
              placeholder="Confirm password"
              class="text-input"
              required
            />
          </div>
        </div>
        
        <div class="form-row">
          <div class="form-field">
            <label for="access-level">Access Level</label>
            <select 
              id="access-level"
              bind:value={accessLevel}
              class="select-input"
              required
            >
              <option value="superadmin">Superadmin</option>
              <option value="user">User</option>
            </select>
          </div>
        </div>
        
        <div class="form-actions">
          <button type="submit" class="primary-button">
            Save User
          </button>
          <button type="button" class="secondary-button" onclick={clearForm}>
            Clear
          </button>
        </div>
      </form>
    </div>
    
    {#if statusMessage}
      <div class="status-message" class:error={isError}>
        <span class="status-icon">{isError ? '⚠️' : '✓'}</span>
        <span>{statusMessage}</span>
        <button class="close-status" onclick={() => statusMessage = ''}>×</button>
      </div>
    {/if}
  </div>
  
  <style>
    h2 {
      margin-top: 0;
      margin-bottom: 1.5rem;
      font-size: 1.5rem;
      font-weight: 500;
      color: var(--text-primary);
    }
    
    .settings-container {
      background-color: var(--bg-tertiary);
      border-radius: 8px;
      padding: 1.5rem;
      box-shadow: var(--shadow-light);
    }
    
    .settings-intro {
      margin-bottom: 2rem;
      color: var(--text-secondary);
      line-height: 1.5;
    }
    
    .settings-group {
      margin-bottom: 2rem;
      animation: fadeIn 0.3s ease;
    }
    
    .settings-group h3 {
      font-size: 1.1rem;
      font-weight: 500;
      margin-bottom: 1.5rem;
      color: var(--text-accent);
      border-bottom: 1px solid var(--border-color);
      padding-bottom: 0.5rem;
    }
    
    .user-form {
      background-color: var(--bg-secondary);
      border: 1px solid var(--border-color);
      border-radius: 8px;
      padding: 1.5rem;
    }
    
    .form-row {
      margin-bottom: 1.25rem;
    }
    
    .form-field label {
      display: block;
      margin-bottom: 0.5rem;
      font-weight: 500;
      color: var(--text-primary);
      font-size: 0.9rem;
    }
    
    .text-input,
    .select-input {
      width: 100%;
      padding: 0.75rem;
      border: 1px solid var(--border-color);
      border-radius: 6px;
      background-color: var(--bg-tertiary);
      color: var(--text-primary);
      font-size: 0.9rem;
      transition: all var(--transition-speed) ease;
    }
    
    .text-input:focus,
    .select-input:focus {
      border-color: var(--accent-primary);
      outline: none;
      box-shadow: 0 0 0 3px rgba(106, 153, 85, 0.1);
    }
    
    .select-input {
      cursor: pointer;
    }
    
    .form-actions {
      display: flex;
      gap: 0.75rem;
      margin-top: 1.5rem;
    }
    
    .primary-button {
      padding: 0.75rem 1.5rem;
      background-color: var(--accent-primary);
      color: white;
      border: none;
      border-radius: 6px;
      cursor: pointer;
      font-weight: 500;
      font-size: 0.9rem;
      transition: all var(--transition-speed) ease;
    }
    
    .primary-button:hover {
      background-color: var(--accent-hover, #5a7a4a);
      transform: translateY(-1px);
    }
    
    .primary-button:active {
      transform: translateY(0);
    }
    
    .secondary-button {
      padding: 0.75rem 1.5rem;
      background-color: transparent;
      color: var(--text-primary);
      border: 1px solid var(--border-color);
      border-radius: 6px;
      cursor: pointer;
      font-weight: 500;
      font-size: 0.9rem;
      transition: all var(--transition-speed) ease;
    }
    
    .secondary-button:hover {
      background-color: var(--bg-hover);
      border-color: var(--accent-primary);
    }
    
    .status-message {
      margin-top: 1.5rem;
      padding: 1rem;
      display: flex;
      align-items: center;
      border-radius: 6px;
      background-color: rgba(106, 153, 85, 0.1);
      color: var(--accent-primary);
      animation: slideIn 0.3s ease;
      position: relative;
    }
    
    .status-message.error {
      background-color: rgba(206, 145, 120, 0.1);
      color: var(--text-warning);
    }
    
    .status-icon {
      margin-right: 0.75rem;
      font-size: 1.2rem;
    }
    
    .close-status {
      position: absolute;
      right: 1rem;
      background: none;
      border: none;
      cursor: pointer;
      font-size: 1.2rem;
      color: currentColor;
      opacity: 0.6;
    }
    
    .close-status:hover {
      opacity: 1;
    }
    
    @keyframes fadeIn {
      from { opacity: 0; }
      to { opacity: 1; }
    }
    
    @keyframes slideIn {
      from { 
        transform: translateY(-10px);
        opacity: 0;
      }
      to { 
        transform: translateY(0);
        opacity: 1;
      }
    }
    
    @media (max-width: 768px) {
      .settings-container {
        padding: 1rem;
      }
      
      .user-form {
        padding: 1rem;
      }
      
      .form-actions {
        flex-direction: column;
      }
      
      .primary-button,
      .secondary-button {
        width: 100%;
      }
    }
  </style>
