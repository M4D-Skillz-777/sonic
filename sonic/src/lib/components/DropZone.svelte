<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let file: File | null = null;

  const dispatch = createEventDispatcher();
  let isDragging = false;

  function handleDragOver(e: DragEvent) {
    e.preventDefault();
    isDragging = true;
  }

  function handleDragLeave() {
    isDragging = false;
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    isDragging = false;

    const files = e.dataTransfer?.files;
    if (files && files.length > 0) {
      const droppedFile = files[0];
      if (droppedFile.type === 'audio/mpeg' || droppedFile.name.endsWith('.mp3')) {
        dispatch('select', droppedFile);
      }
    }
  }

  function handleFileInput(e: Event) {
    const target = e.target as HTMLInputElement;
    const files = target.files;
    if (files && files.length > 0) {
      dispatch('select', files[0]);
    }
  }

  function clearFile() {
    dispatch('clear');
  }

  function formatFileSize(bytes: number): string {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  }
</script>

{#if file}
  <div class="file-selected animate-slideUp">
    <div class="file-card">
      <div class="file-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M9 18V5l12-2v13"/>
          <circle cx="6" cy="18" r="3"/>
          <circle cx="18" cy="16" r="3"/>
        </svg>
      </div>
      <div class="file-details">
        <span class="file-name">{file.name}</span>
        <span class="file-meta">
          <span class="file-size">{formatFileSize(file.size)}</span>
          <span class="file-type">MP3</span>
        </span>
      </div>
      <button class="clear-btn spring-bounce" on:click={clearFile} title="Remove file" aria-label="Remove file">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M18 6L6 18M6 6l12 12"/>
        </svg>
      </button>
    </div>
  </div>
{:else}
  <div 
    class="dropzone"
    class:dragging={isDragging}
    role="button"
    tabindex="0"
    on:dragover={handleDragOver}
    on:dragleave={handleDragLeave}
    on:drop={handleDrop}
    on:keypress={(e) => e.key === 'Enter' && document.getElementById('file-input')?.click()}
  >
    <input 
      type="file" 
      accept=".mp3,audio/mpeg" 
      id="file-input"
      on:change={handleFileInput}
    />
    <label for="file-input" class="dropzone-content">
      <div class="upload-icon-wrapper">
        <svg class="upload-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/>
          <polyline points="17 8 12 3 7 8"/>
          <line x1="12" y1="3" x2="12" y2="15"/>
        </svg>
        <div class="pulse-ring"></div>
      </div>
      <div class="dropzone-text">
        <span class="dropzone-title">Drop your MP3 here</span>
        <span class="dropzone-subtitle">or click to browse</span>
      </div>
      <div class="dropzone-hint">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <path d="M12 16v-4M12 8h.01"/>
        </svg>
        <span>Best results with 5-10 second samples</span>
      </div>
    </label>
  </div>
{/if}

<style>
  .dropzone {
    width: 100%;
    min-height: 220px;
    border: 2px dashed var(--glass-border);
    border-radius: 1.25rem;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--glass-bg);
    transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
    position: relative;
    overflow: hidden;
  }

  .dropzone::before {
    content: '';
    position: absolute;
    inset: 0;
    background: radial-gradient(circle at center, rgba(6, 182, 212, 0.1) 0%, transparent 70%);
    opacity: 0;
    transition: opacity 0.3s;
  }

  .dropzone:hover::before,
  .dropzone.dragging::before {
    opacity: 1;
  }

  .dropzone:hover,
  .dropzone.dragging {
    border-color: var(--accent-cyan);
    transform: scale(1.01);
    box-shadow: 0 0 40px rgba(6, 182, 212, 0.15);
  }

  .dropzone.dragging {
    transform: scale(1.02);
    border-style: solid;
  }

  .dropzone input {
    position: absolute;
    width: 100%;
    height: 100%;
    opacity: 0;
    cursor: pointer;
  }

  .dropzone-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
    padding: 2rem;
    pointer-events: none;
  }

  .upload-icon-wrapper {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .upload-icon {
    width: 56px;
    height: 56px;
    color: var(--accent-cyan);
    position: relative;
    z-index: 1;
    transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  }

  .dropzone:hover .upload-icon {
    transform: translateY(-5px);
  }

  .pulse-ring {
    position: absolute;
    width: 80px;
    height: 80px;
    border: 2px solid var(--accent-cyan);
    border-radius: 50%;
    opacity: 0;
    animation: pulse-ring 2s ease-out infinite;
  }

  @keyframes pulse-ring {
    0% {
      transform: scale(0.8);
      opacity: 0.5;
    }
    100% {
      transform: scale(1.3);
      opacity: 0;
    }
  }

  .dropzone-text {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.375rem;
  }

  .dropzone-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--text-primary);
    letter-spacing: -0.01em;
  }

  .dropzone-subtitle {
    font-size: 0.875rem;
    color: var(--text-secondary);
  }

  .dropzone-hint {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    background: rgba(6, 182, 212, 0.1);
    border-radius: 2rem;
    font-size: 0.75rem;
    color: var(--accent-cyan);
  }

  .dropzone-hint svg {
    width: 14px;
    height: 14px;
  }

  .file-selected {
    width: 100%;
  }

  .file-card {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1.25rem;
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: 1rem;
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
  }

  .file-icon {
    width: 48px;
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: linear-gradient(135deg, rgba(6, 182, 212, 0.2), rgba(139, 92, 246, 0.2));
    border-radius: 0.75rem;
    color: var(--accent-cyan);
    flex-shrink: 0;
  }

  .file-icon svg {
    width: 24px;
    height: 24px;
  }

  .file-details {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    min-width: 0;
  }

  .file-name {
    font-weight: 500;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 0.9375rem;
  }

  .file-meta {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .file-size {
    font-size: 0.8125rem;
    color: var(--text-secondary);
  }

  .file-type {
    font-size: 0.6875rem;
    padding: 0.125rem 0.5rem;
    background: rgba(6, 182, 212, 0.1);
    border-radius: 1rem;
    color: var(--accent-cyan);
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .clear-btn {
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(239, 68, 68, 0.1);
    border-radius: 0.625rem;
    color: var(--error);
    transition: all 0.2s;
    flex-shrink: 0;
  }

  .clear-btn:hover {
    background: rgba(239, 68, 68, 0.2);
    transform: scale(1.1);
  }

  .clear-btn svg {
    width: 18px;
    height: 18px;
  }
</style>
