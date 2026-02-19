<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { MatchResult as MatchResultType, SpotifyTrack } from '$lib/api';
  import CircularProgress from './CircularProgress.svelte';

  export let result: MatchResultType;
  export let mode: 'identify' | 'register';

  const dispatch = createEventDispatcher();

  $: confidencePercent = Math.round((result.confidence || 0) * 100);
  $: isMatch = result.confidence && result.confidence > 0;
  $: spotify = result.spotify;

  function handleRegisterNew() {
    dispatch('register');
  }
</script>

{#if mode === 'register'}
  <div class="result-container animate-scaleIn">
    <div class="success-state">
      <div class="success-icon animate-glow">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <path d="M22 11.08V12a10 10 0 11-5.93-9.14"/>
          <polyline points="22 4 12 14.01 9 11.01"/>
        </svg>
      </div>
      <h3 class="title">Registered Successfully</h3>
      <p class="song-name gradient-text">{result.song_name}</p>
      <div class="meta">
        <span class="badge">{result.fft_impl}</span>
        <span class="hash-count">{result.hashes || 0} hashes</span>
      </div>
    </div>
  </div>

{:else if isMatch}
  <div class="result-container animate-scaleIn">
    <div class="match-state">
      <div class="confidence-section">
        <CircularProgress progress={confidencePercent} size={100} />
      </div>
      
      <div class="match-info">
        <span class="match-label">Match Found</span>
        <h3 class="song-name gradient-text">{result.song_name}</h3>
        {#if result.fft_impl}
          <span class="badge">{result.fft_impl}</span>
        {/if}
        {#if result.source}
          <span class="source-badge">{result.source}</span>
        {/if}
      </div>

      {#if spotify}
        <div class="spotify-card">
          {#if spotify.album_art}
            <img src={spotify.album_art} alt={spotify.album} class="album-art" />
          {/if}
          <div class="spotify-info">
            <p class="artist">{spotify.artist}</p>
            <p class="album">{spotify.album}</p>
            {#if spotify.spotify_url}
              <a href={spotify.spotify_url} target="_blank" rel="noopener" class="spotify-link">
                <svg viewBox="0 0 24 24" fill="currentColor" width="16" height="16">
                  <path d="M12 0C5.4 0 0 5.4 0 12s5.4 12 12 12 12-5.4 12-12S18.66 0 12 0zm5.521 17.34c-.24.359-.66.48-1.021.24-2.82-1.74-6.36-2.101-10.561-1.141-.418.122-.779-.179-.899-.539-.12-.421.18-.78.54-.9 4.56-1.021 8.52-.6 11.64 1.32.42.18.479.659.301 1.02zm1.44-3.3c-.301.42-.841.6-1.262.3-3.239-1.98-8.159-2.58-11.939-1.38-.479.12-1.02-.12-1.14-.6-.12-.48.12-1.021.6-1.141C9.6 9.9 15 10.561 18.72 12.84c.361.181.54.78.241 1.2zm.12-3.36C15.24 8.4 8.82 8.16 5.16 9.301c-.6.179-1.2-.181-1.38-.721-.18-.601.18-1.2.72-1.381 4.26-1.26 11.28-1.02 15.721 1.621.539.3.719 1.02.419 1.56-.299.421-1.02.599-1.559.3z"/>
                </svg>
                Open in Spotify
              </a>
            {/if}
          </div>
        </div>
      {/if}
    </div>
  </div>

{:else}
  <div class="result-container animate-scaleIn">
    <div class="no-match-state">
      <div class="no-match-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"/>
          <path d="m21 21-4.35-4.35"/>
          <path d="M8 11h6"/>
        </svg>
      </div>
      <h3 class="title">No Match Found</h3>
      <p class="subtitle">This song isn't in our database yet</p>
      
      <div class="actions">
        <button class="register-btn spring-bounce" on:click={handleRegisterNew}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 5v14M5 12h14"/>
          </svg>
          Register as New Song
        </button>
      </div>
      
      <p class="hint">Try uploading a longer sample or check audio quality</p>
    </div>
  </div>
{/if}

<style>
  .result-container {
    width: 100%;
    display: flex;
    justify-content: center;
  }

  .success-state,
  .match-state,
  .no-match-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
    padding: 2rem;
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: 1.5rem;
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    min-width: 300px;
  }

  .success-icon,
  .no-match-icon {
    width: 72px;
    height: 72px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
  }

  .success-icon {
    background: linear-gradient(135deg, var(--accent-cyan), var(--accent-purple));
    color: var(--bg-primary);
    box-shadow: 0 0 40px var(--accent-glow);
  }

  .success-icon svg,
  .no-match-icon svg {
    width: 36px;
    height: 36px;
  }

  .no-match-icon {
    background: rgba(255, 255, 255, 0.05);
    color: var(--text-secondary);
  }

  .title {
    font-size: 1.25rem;
    font-weight: 600;
    margin: 0;
    letter-spacing: -0.02em;
  }

  .subtitle {
    color: var(--text-secondary);
    font-size: 0.875rem;
    margin: 0;
  }

  .song-name {
    font-size: 1.125rem;
    font-weight: 600;
    margin: 0;
    text-align: center;
  }

  .gradient-text {
    background: linear-gradient(135deg, var(--accent-cyan), var(--accent-purple));
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .meta {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .badge {
    font-size: 0.75rem;
    padding: 0.375rem 0.75rem;
    background: rgba(6, 182, 212, 0.1);
    border: 1px solid rgba(6, 182, 212, 0.2);
    border-radius: 2rem;
    color: var(--accent-cyan);
    font-weight: 500;
  }

  .hash-count {
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .match-state {
    gap: 1.5rem;
  }

  .confidence-section {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .match-info {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    text-align: center;
  }

  .match-label {
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--success);
    font-weight: 600;
  }

  .source-badge {
    font-size: 0.625rem;
    padding: 0.25rem 0.5rem;
    background: rgba(139, 92, 246, 0.1);
    border-radius: 0.25rem;
    color: var(--accent-purple);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .spotify-card {
    display: flex;
    gap: 1rem;
    padding: 1rem;
    background: rgba(0, 0, 0, 0.3);
    border-radius: 0.75rem;
    width: 100%;
  }

  .album-art {
    width: 64px;
    height: 64px;
    border-radius: 0.5rem;
    object-fit: cover;
  }

  .spotify-info {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    flex: 1;
    min-width: 0;
  }

  .artist {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text-primary);
    margin: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .album {
    font-size: 0.75rem;
    color: var(--text-secondary);
    margin: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .spotify-link {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    font-size: 0.75rem;
    color: #1DB954;
    margin-top: 0.25rem;
    text-decoration: none;
    transition: opacity 0.2s;
  }

  .spotify-link:hover {
    opacity: 0.8;
    text-decoration: underline;
  }

  .no-match-state .actions {
    margin-top: 0.5rem;
    width: 100%;
  }

  .register-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    width: 100%;
    padding: 0.875rem 1.5rem;
    background: linear-gradient(135deg, var(--accent-cyan), var(--accent-purple));
    color: var(--bg-primary);
    font-size: 0.9rem;
    font-weight: 600;
    border-radius: 0.75rem;
    transition: all 0.2s;
  }

  .register-btn svg {
    width: 18px;
    height: 18px;
  }

  .register-btn:hover {
    box-shadow: 0 10px 40px var(--accent-glow);
  }

  .hint {
    font-size: 0.75rem;
    color: var(--text-muted);
    margin: 0;
    text-align: center;
  }
</style>
