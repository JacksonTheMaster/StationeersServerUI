<script>
  import { apiFetch } from '../../services/api';
  
  let isLoading = $state(false);
  let responseMessage = $state('');
  let isError = $state(false);

  async function reloadAll() {
    isLoading = true;
    isError = false;
    responseMessage = '';

    try {
      const response = await apiFetch('/api/v3/backend/reload', {
        method: 'POST'
      });

      let data;
      try {
        data = await response.json();
      } catch (e) {
        throw new Error('Invalid response format');
      }

      if (!response.ok) {
        isError = true;
        responseMessage = data.error || 'Backend reload failed';
      } else {
        responseMessage = data.message || 'Backend reloaded';
      }
    } catch (error) {
      isError = true;
      responseMessage = error.message || 'Failed to connect to server';
      console.error('Error:', error);
    } finally {
      isLoading = false;
    }
  }
</script>

<div class="card reload-control">
  <button 
    onclick={reloadAll}
    disabled={isLoading}
    style:background-color="var(--bg-hover)"
    class:active={!isLoading}
  >
    {isLoading ? '🕐 Reloading...' : '🔃 Reload Backend'}
  </button>

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
  .reload-control {
    display: flex;
    flex-direction: column;
  }

  button {
    padding: 0.75rem;
    border: none;
    border-radius: 4px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
  }

  button:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

  button.active {
    filter: brightness(1.1);
  }

  .response-message {
    padding: 0.75rem;
    border-radius: 4px;
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    animation: fadeIn 0.3s ease;
    margin-top: 1rem;
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
