<script>
    import { apiFetch } from '../../services/api';
    let isLoading = $state(false);
    let lastAction = $state(null);
    let responseMessage = $state('');
    let isError = $state(false);
    
    async function sendCommand(action) {
      isLoading = true;
      lastAction = action;
      isError = false;
      
      try {
        const response = await apiFetch(`/api/v3/server/${action}`, {
          method: 'POST'
        });
        
        const text = await response.text();
        
        if (!response.ok) {
          isError = true;
        }
        
        responseMessage = text;
      } catch (error) {
        isError = true;
        responseMessage = 'Failed to connect to server';
        console.error('Error:', error);
      } finally {
        isLoading = false;
      }
    }
  </script>
  
  <div class="card server-control">
    <div class="button-group">
      <button 
        onclick={() => sendCommand('start')} 
        disabled={isLoading}
        class:active={lastAction === 'start' && !isLoading}
        class="start-button"
      >
        {isLoading && lastAction === 'start' ? 'Starting...' : 'Start Server'}
      </button>
      
      <button 
        onclick={() => sendCommand('stop')} 
        disabled={isLoading}
        class:active={lastAction === 'stop' && !isLoading}
        class="stop-button"
      >
        {isLoading && lastAction === 'stop' ? 'Stopping...' : 'Stop Server'}
      </button>
    </div>
    
    {#if responseMessage}
      <div class={`response-message ${isError ? 'error' : 'success'}`}>
        <div class="response-icon">
          {#if isError}
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path fill="currentColor" d="M12,2C6.48,2,2,6.48,2,12s4.48,10,10,10s10-4.48,10-10S17.52,2,12,2z M13,17h-2v-2h2V17z M13,13h-2V7h2V13z"/>
            </svg>
          {:else}
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path fill="currentColor" d="M12 2C6.5 2 2 6.5 2 12S6.5 22 12 22 22 17.5 22 12 17.5 2 12 2M10 17L5 12L6.41 10.59L10 14.17L17.59 6.58L19 8L10 17Z"/>
            </svg>
          {/if}
        </div>
        <div class="response-text">{responseMessage}</div>
      </div>
    {/if}
  </div>
  
  <style>
    .server-control {
      display: flex;
      flex-direction: column;
    }
    
    .button-group {
      display: flex;
      gap: 0.75rem;
      margin-bottom: 1rem;
    }
    
    .button-group button {
      flex: 1;
      padding: 0.75rem;
      border: none;
      border-radius: 4px;
      font-weight: 500;
      cursor: pointer;
      transition: all 0.2s ease;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    
    .button-group button:disabled {
      opacity: 0.7;
      cursor: not-allowed;
    }
    
    .start-button {
      background-color: rgba(76, 175, 80, 0.2);
      color: #4CAF50;
    }
    
    .start-button:hover:not(:disabled) {
      background-color: rgba(76, 175, 80, 0.3);
    }
    
    .start-button.active {
      background-color: #4CAF50;
      color: white;
    }
    
    .stop-button {
      background-color: rgba(244, 67, 54, 0.2);
      color: #F44336;
    }
    
    .stop-button:hover:not(:disabled) {
      background-color: rgba(244, 67, 54, 0.3);
    }
    
    .stop-button.active {
      background-color: #F44336;
      color: white;
    }
    
    .response-message {
      padding: 0.75rem;
      border-radius: 4px;
      display: flex;
      align-items: flex-start;
      gap: 0.75rem;
      animation: fadeIn 0.3s ease;
    }
    
    .response-message.success {
      background-color: rgba(76, 175, 80, 0.1);
      border-left: 3px solid #4CAF50;
    }
    
    .response-message.error {
      background-color: rgba(244, 67, 54, 0.1);
      border-left: 3px solid #F44336;
    }
    
    .response-icon {
      margin-top: 0.1rem;
    }
    
    .response-text {
      flex: 1;
      font-size: 0.9rem;
    }
    
    @keyframes fadeIn {
      from { opacity: 0; transform: translateY(-5px); }
      to { opacity: 1; transform: translateY(0); }
    }
  </style>