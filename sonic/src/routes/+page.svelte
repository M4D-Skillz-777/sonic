<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import '../app.css';
  import DropZone from '$lib/components/DropZone.svelte';
  import AudioRecorder from '$lib/components/AudioRecorder.svelte';
  import WaveformDisplay from '$lib/components/WaveformDisplay.svelte';
  import AudioVisualizer from '$lib/components/AudioVisualizer.svelte';
  import MatchResult from '$lib/components/MatchResult.svelte';
  import SongList from '$lib/components/SongList.svelte';
  import CircularProgress from '$lib/components/CircularProgress.svelte';
  import Confetti from '$lib/components/Confetti.svelte';
  import { registerSong, recognizeSong, listSongs, deleteSong, type MatchResult as MatchResultType } from '$lib/api';

  type Mode = 'identify' | 'register';
  type InputMode = 'file' | 'record';

  let mode: Mode = 'identify';
  let inputMode: InputMode = 'file';
  let file: File | null = null;
  let songName = '';
  let useCustomFFT = false;
  let isProcessing = false;
  let progress = 0;
  let matchResult: MatchResultType | null = null;
  let error = '';
  let songs: string[] = [];
  let audioUrl: string | null = null;
  let audioElement: HTMLAudioElement | null = null;
  let showConfetti = false;
  let modeTransition = false;

  $: audioUrl = file ? URL.createObjectURL(file) : null;

  async function handleFileSelect(selectedFile: File) {
    file = selectedFile;
    inputMode = 'file';
    matchResult = null;
    error = '';
    progress = 0;
  }

  function handleRecorded(e: CustomEvent<{ blob: Blob; url: string }>) {
    const recordedFile = new File([e.detail.blob], 'recording.webm', { type: 'audio/webm' });
    file = recordedFile;
    audioUrl = e.detail.url;
    inputMode = 'record';
    matchResult = null;
    error = '';
    progress = 0;
  }

  function handleRecordingError(e: CustomEvent<{ message: string }>) {
    error = e.detail.message;
  }

  function clearRecording() {
    file = null;
    audioUrl = null;
  }

  async function handleSubmit() {
    if (!file) return;

    isProcessing = true;
    error = '';
    matchResult = null;
    progress = 0;

    try {
      const progressInterval = setInterval(() => {
        if (progress < 90) {
          progress += Math.random() * 15;
        }
      }, 200);

      if (mode === 'register') {
        if (!songName.trim()) {
          throw new Error('Please enter a song name');
        }
        const result = await registerSong(file, songName, useCustomFFT);
        progress = 100;
        matchResult = {
          song_name: result.song_name,
          confidence: 1,
          fft_impl: result.fft_impl,
          hashes: result.hashes
        };
        showConfetti = true;
        setTimeout(() => showConfetti = false, 100);
        await loadSongs();
      } else {
        const result = await recognizeSong(file, useCustomFFT);
        progress = 100;
        matchResult = result;
        if (result.confidence > 0.5) {
          showConfetti = true;
          setTimeout(() => showConfetti = false, 100);
        }
      }

      clearInterval(progressInterval);
    } catch (e: any) {
      error = e.message || 'An error occurred';
    } finally {
      isProcessing = false;
    }
  }

  async function loadSongs() {
    try {
      songs = await listSongs();
    } catch {
      songs = [];
    }
  }

  async function handleDeleteSong(name: string) {
    try {
      await deleteSong(name);
      songs = songs.filter(s => s !== name);
    } catch (e: any) {
      error = e.message;
    }
  }

  function clearFile() {
    file = null;
    audioUrl = null;
    matchResult = null;
    error = '';
    songName = '';
    progress = 0;
  }

  function switchToRegister() {
    modeTransition = true;
    setTimeout(() => {
      mode = 'register';
      modeTransition = false;
    }, 150);
  }

  function handleRegisterFromNoMatch() {
    switchToRegister();
  }

  function setMode(newMode: Mode) {
    modeTransition = true;
    setTimeout(() => {
      mode = newMode;
      clearFile();
      modeTransition = false;
    }, 150);
  }

  onMount(() => {
    loadSongs();
  });
</script>

<svelte:head>
  <title>Sonic - Audio Fingerprinting</title>
  <meta name="description" content="Identify songs using audio fingerprinting with AI" />
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&display=swap" rel="stylesheet">
</svelte:head>

<Confetti trigger={showConfetti} />

<main class="app">
  <header class="header animate-slideDown">
    <div class="logo animate-float">
      <div class="logo-icon-wrapper">
        <svg class="logo-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M9 18V5L21 3V16" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          <circle cx="6" cy="18" r="3" stroke="currentColor" stroke-width="2"/>
          <circle cx="18" cy="16" r="3" stroke="currentColor" stroke-width="2"/>
        </svg>
        <div class="logo-glow"></div>
      </div>
      <span class="logo-text gradient-text">Sonic</span>
    </div>
    <p class="tagline">Audio fingerprinting</p>
  </header>

  <div class="mode-toggle animate-slideUp delay-100">
    <button 
      class="mode-btn" 
      class:active={mode === 'identify'}
      on:click={() => setMode('identify')}
    >
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="8"/>
        <path d="m21 21-4.35-4.35"/>
      </svg>
      <span>Identify</span>
    </button>
    <button 
      class="mode-btn" 
      class:active={mode === 'register'}
      on:click={() => setMode('register')}
    >
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M12 5v14M5 12h14"/>
      </svg>
      <span>Register</span>
    </button>
    <div class="slider" class:identify={mode === 'identify'} class:register={mode === 'register'}></div>
  </div>

  <div class="fft-toggle animate-slideUp delay-200">
    <label class="toggle-label">
      <input type="checkbox" bind:checked={useCustomFFT} />
      <span class="toggle-slider"></span>
      <span class="toggle-text">
        {useCustomFFT ? 'Custom Cooley-Tukey FFT' : 'Optimized FFT (gofft)'}
      </span>
    </label>
  </div>

  <div class="content" class:transitioning={modeTransition}>
    <div class="input-toggle animate-slideUp">
      <button 
        class="input-btn" 
        class:active={inputMode === 'file'}
        on:click={() => { inputMode = 'file'; clearFile(); }}
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/>
          <polyline points="17 8 12 3 7 8"/>
          <line x1="12" y1="3" x2="12" y2="15"/>
        </svg>
        Upload
      </button>
      <button 
        class="input-btn" 
        class:active={inputMode === 'record'}
        on:click={() => { inputMode = 'record'; clearFile(); }}
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
          <path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
          <line x1="12" y1="19" x2="12" y2="23"/>
          <line x1="8" y1="23" x2="16" y2="23"/>
        </svg>
        Record
      </button>
    </div>

    {#if inputMode === 'file'}
      <DropZone 
        {file} 
        on:select={(e) => handleFileSelect(e.detail)}
        on:clear={clearFile}
      />
    {:else}
      <AudioRecorder 
        maxDuration={10}
        {isProcessing}
        on:recorded={handleRecorded}
        on:error={handleRecordingError}
        on:clear={clearRecording}
      />
    {/if}

    {#if file}
      <div class="audio-section animate-slideUp delay-100">
        <WaveformDisplay {audioUrl} />
        <AudioVisualizer {audioUrl} bind:audioElement />
      </div>

      {#if mode === 'register'}
        <div class="song-name-input animate-slideUp delay-200">
          <input 
            type="text" 
            placeholder="Enter song name..." 
            bind:value={songName}
            disabled={isProcessing}
          />
        </div>
      {/if}

      <div class="actions animate-slideUp delay-300">
        {#if isProcessing}
          <div class="processing-indicator">
            <CircularProgress progress={progress} size={80} />
            <span class="processing-text">Processing audio...</span>
          </div>
        {:else}
          <button 
            class="submit-btn spring-bounce" 
            on:click={handleSubmit}
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              {#if mode === 'identify'}
                <circle cx="11" cy="11" r="8"/>
                <path d="m21 21-4.35-4.35"/>
              {:else}
                <path d="M12 5v14M5 12h14"/>
              {/if}
            </svg>
            {mode === 'identify' ? 'Identify Song' : 'Register Song'}
          </button>
        {/if}
      </div>

      {#if error}
        <div class="error-message animate-slideUp">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <path d="m15 9-6 6M9 9l6 6"/>
          </svg>
          {error}
        </div>
      {/if}

      {#if matchResult}
        <MatchResult result={matchResult} mode={mode} on:register={handleRegisterFromNoMatch} />
      {/if}
    {/if}

    <SongList {songs} on:delete={(e) => handleDeleteSong(e.detail)} />
  </div>

  <footer class="footer animate-fadeIn delay-500">
    <div class="footer-content">
      <p>Powered by FFT audio fingerprinting</p>
      <div class="footer-links">
        <span class="tech-badge">Go</span>
        <span class="tech-badge">Redis</span>
        <span class="tech-badge">Svelte</span>
      </div>
    </div>
  </footer>
</main>

<style>
  .app {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 2rem 1.5rem;
    max-width: 800px;
    margin: 0 auto;
  }

  .header {
    text-align: center;
    margin-bottom: 2rem;
  }

  .logo {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1rem;
    margin-bottom: 0.75rem;
  }

  .logo-icon-wrapper {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .logo-icon {
    width: 48px;
    height: 48px;
    color: var(--accent-cyan);
    position: relative;
    z-index: 1;
  }

  .logo-glow {
    position: absolute;
    width: 80px;
    height: 80px;
    background: radial-gradient(circle, var(--accent-glow) 0%, transparent 70%);
    filter: blur(20px);
    animation: pulse 3s ease-in-out infinite;
  }

  .logo-text {
    font-size: 2.75rem;
    font-weight: 800;
    letter-spacing: -0.03em;
  }

  .tagline {
    color: var(--text-secondary);
    font-size: 0.9rem;
    font-weight: 400;
    letter-spacing: 0.02em;
  }

  .mode-toggle {
    display: flex;
    position: relative;
    gap: 0;
    margin-bottom: 1rem;
    background: var(--glass-bg);
    padding: 0.25rem;
    border-radius: 1rem;
    border: 1px solid var(--glass-border);
    overflow: hidden;
  }

  .mode-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1.75rem;
    border-radius: 0.75rem;
    background: transparent;
    color: var(--text-secondary);
    font-size: 0.875rem;
    font-weight: 500;
    transition: color 0.2s;
    position: relative;
    z-index: 2;
  }

  .mode-btn svg {
    width: 18px;
    height: 18px;
  }

  .mode-btn.active {
    color: var(--bg-primary);
  }

  .slider {
    position: absolute;
    top: 0.25rem;
    left: 0.25rem;
    width: calc(50% - 0.25rem);
    height: calc(100% - 0.5rem);
    background: linear-gradient(135deg, var(--accent-cyan), var(--accent-purple));
    border-radius: 0.75rem;
    transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
    z-index: 1;
  }

  .slider.register {
    transform: translateX(100%);
  }

  .fft-toggle {
    margin-bottom: 1.5rem;
  }

  .toggle-label {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    cursor: pointer;
  }

  .toggle-label input {
    display: none;
  }

  .toggle-slider {
    width: 48px;
    height: 26px;
    background: var(--bg-tertiary);
    border-radius: 13px;
    position: relative;
    transition: background 0.2s;
    border: 1px solid var(--glass-border);
  }

  .toggle-slider::after {
    content: '';
    position: absolute;
    top: 3px;
    left: 3px;
    width: 18px;
    height: 18px;
    background: var(--text-secondary);
    border-radius: 50%;
    transition: all 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
  }

  .toggle-label input:checked + .toggle-slider {
    background: linear-gradient(135deg, var(--accent-cyan), var(--accent-purple));
    border-color: transparent;
  }

  .toggle-label input:checked + .toggle-slider::after {
    left: 25px;
    background: var(--bg-primary);
  }

  .toggle-text {
    color: var(--text-secondary);
    font-size: 0.8125rem;
    font-weight: 500;
  }

  .content {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    transition: opacity 0.15s ease;
  }

  .content.transitioning {
    opacity: 0.5;
  }

  .input-toggle {
    display: flex;
    gap: 0.5rem;
    background: var(--glass-bg);
    padding: 0.25rem;
    border-radius: 0.75rem;
    border: 1px solid var(--glass-border);
  }

  .input-btn {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.625rem 1rem;
    background: transparent;
    border-radius: 0.5rem;
    color: var(--text-secondary);
    font-size: 0.8125rem;
    font-weight: 500;
    transition: all 0.2s;
  }

  .input-btn svg {
    width: 16px;
    height: 16px;
  }

  .input-btn.active {
    background: rgba(6, 182, 212, 0.15);
    color: var(--accent-cyan);
  }

  .audio-section {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .song-name-input {
    width: 100%;
  }

  .song-name-input input {
    width: 100%;
    padding: 1rem 1.25rem;
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: 0.875rem;
    color: var(--text-primary);
    font-size: 0.9375rem;
    font-weight: 500;
    transition: all 0.2s;
  }

  .song-name-input input:focus {
    outline: none;
    border-color: var(--accent-cyan);
    box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
  }

  .song-name-input input::placeholder {
    color: var(--text-muted);
    font-weight: 400;
  }

  .actions {
    display: flex;
    justify-content: center;
  }

  .processing-indicator {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
  }

  .processing-text {
    font-size: 0.875rem;
    color: var(--text-secondary);
  }

  .submit-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.625rem;
    padding: 1rem 2.5rem;
    background: linear-gradient(135deg, var(--accent-cyan), var(--accent-purple));
    color: var(--bg-primary);
    font-size: 0.9375rem;
    font-weight: 600;
    border-radius: 0.875rem;
    transition: all 0.2s;
    min-width: 200px;
    letter-spacing: -0.01em;
  }

  .submit-btn svg {
    width: 18px;
    height: 18px;
  }

  .submit-btn:hover {
    transform: translateY(-2px);
    box-shadow: 0 15px 50px var(--accent-glow);
  }

  .error-message {
    display: flex;
    align-items: center;
    gap: 0.625rem;
    padding: 1rem 1.25rem;
    background: rgba(239, 68, 68, 0.1);
    border: 1px solid rgba(239, 68, 68, 0.2);
    border-radius: 0.875rem;
    color: var(--error);
    font-size: 0.875rem;
  }

  .error-message svg {
    width: 18px;
    height: 18px;
    flex-shrink: 0;
  }

  .footer {
    margin-top: auto;
    padding-top: 3rem;
    text-align: center;
  }

  .footer-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.75rem;
  }

  .footer p {
    color: var(--text-muted);
    font-size: 0.8125rem;
    margin: 0;
  }

  .footer-links {
    display: flex;
    gap: 0.5rem;
  }

  .tech-badge {
    font-size: 0.6875rem;
    padding: 0.25rem 0.625rem;
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: 1rem;
    color: var(--text-muted);
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  @media (max-width: 640px) {
    .app {
      padding: 1.5rem 1rem;
    }

    .logo-text {
      font-size: 2rem;
    }

    .mode-btn {
      padding: 0.625rem 1.25rem;
      font-size: 0.8125rem;
    }

    .submit-btn {
      width: 100%;
    }
  }
</style>
