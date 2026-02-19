<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let songs: string[] = [];

  const dispatch = createEventDispatcher();

  function handleDelete(song: string) {
    dispatch('delete', song);
  }
</script>

{#if songs.length > 0}
  <div class="song-list-container animate-slideUp">
    <div class="list-header">
      <div class="list-title">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M9 18V5l12-2v13"/>
          <circle cx="6" cy="18" r="3"/>
          <circle cx="18" cy="16" r="3"/>
        </svg>
        <span>Library</span>
      </div>
      <span class="song-count">{songs.length} song{songs.length !== 1 ? 's' : ''}</span>
    </div>
    <div class="song-grid">
      {#each songs as song, i (song)}
        <div class="song-card animate-slideUp" style="animation-delay: {i * 50}ms">
          <div class="song-info">
            <div class="song-avatar">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M9 18V5l12-2v13"/>
                <circle cx="6" cy="18" r="3"/>
                <circle cx="18" cy="16" r="3"/>
              </svg>
            </div>
            <span class="song-name">{song}</span>
          </div>
          <button class="delete-btn spring-bounce" on:click={() => handleDelete(song)} title="Delete song" aria-label="Delete {song}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
              <line x1="10" y1="11" x2="10" y2="17"/>
              <line x1="14" y1="11" x2="14" y2="17"/>
            </svg>
          </button>
        </div>
      {/each}
    </div>
  </div>
{:else}
  <div class="empty-state animate-slideUp">
    <div class="empty-icon">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M9 18V5l12-2v13"/>
        <circle cx="6" cy="18" r="3"/>
        <circle cx="18" cy="16" r="3"/>
        <path d="M12 12v.01"/>
      </svg>
    </div>
    <div class="empty-text">
      <p class="empty-title">No songs yet</p>
      <p class="empty-subtitle">Register your first song to get started</p>
    </div>
  </div>
{/if}

<style>
  .song-list-container {
    width: 100%;
    margin-top: 1rem;
  }

  .list-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1rem;
    padding: 0 0.25rem;
  }

  .list-title {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.9375rem;
    font-weight: 600;
    color: var(--text-primary);
    letter-spacing: -0.01em;
  }

  .list-title svg {
    width: 20px;
    height: 20px;
    color: var(--accent-cyan);
  }

  .song-count {
    font-size: 0.75rem;
    color: var(--text-muted);
    background: var(--glass-bg);
    padding: 0.25rem 0.625rem;
    border-radius: 1rem;
    border: 1px solid var(--glass-border);
  }

  .song-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 0.75rem;
  }

  .song-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.875rem 1rem;
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: 0.875rem;
    transition: all 0.2s;
  }

  .song-card:hover {
    border-color: rgba(6, 182, 212, 0.3);
    background: rgba(6, 182, 212, 0.05);
    transform: translateY(-2px);
  }

  .song-info {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    min-width: 0;
  }

  .song-avatar {
    width: 36px;
    height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: linear-gradient(135deg, rgba(6, 182, 212, 0.2), rgba(139, 92, 246, 0.2));
    border-radius: 0.5rem;
    color: var(--accent-cyan);
    flex-shrink: 0;
  }

  .song-avatar svg {
    width: 18px;
    height: 18px;
  }

  .song-name {
    color: var(--text-primary);
    font-size: 0.875rem;
    font-weight: 500;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .delete-btn {
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border-radius: 0.5rem;
    color: var(--text-muted);
    transition: all 0.2s;
    flex-shrink: 0;
    opacity: 0;
  }

  .song-card:hover .delete-btn {
    opacity: 1;
  }

  .delete-btn:hover {
    background: rgba(239, 68, 68, 0.15);
    color: var(--error);
  }

  .delete-btn svg {
    width: 16px;
    height: 16px;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
    padding: 3rem 2rem;
    text-align: center;
    background: var(--glass-bg);
    border: 1px dashed var(--glass-border);
    border-radius: 1rem;
  }

  .empty-icon {
    width: 64px;
    height: 64px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(6, 182, 212, 0.1);
    border-radius: 1rem;
    color: var(--accent-cyan);
  }

  .empty-icon svg {
    width: 32px;
    height: 32px;
  }

  .empty-text {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .empty-title {
    font-size: 1rem;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
  }

  .empty-subtitle {
    font-size: 0.875rem;
    color: var(--text-secondary);
    margin: 0;
  }

  @media (max-width: 640px) {
    .song-grid {
      grid-template-columns: 1fr;
    }

    .delete-btn {
      opacity: 1;
    }
  }
</style>
