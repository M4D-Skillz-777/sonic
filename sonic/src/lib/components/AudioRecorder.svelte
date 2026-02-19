<script lang="ts">
  import { createEventDispatcher, onDestroy } from 'svelte';

  export let maxDuration = 10;
  export let isProcessing = false;

  const dispatch = createEventDispatcher();

  let isRecording = false;
  let audioBlob: Blob | null = null;
  let audioUrl: string | null = null;
  let recordingTime = 0;
  let audioLevel = 0;
  
  let audioContext: AudioContext | null = null;
  let analyser: AnalyserNode | null = null;
  let animationId: number | null = null;
  let timerInterval: number | null = null;
  let mediaStream: MediaStream | null = null;
  
  let recordedChunks: Float32Array[] = [];
  let sampleRate = 44100;

  async function startRecording() {
    try {
      mediaStream = await navigator.mediaDevices.getUserMedia({ 
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true,
          sampleRate: 48000,
          channelCount: 1
        } 
      });

      audioContext = new AudioContext({ sampleRate: 48000 });
      sampleRate = audioContext.sampleRate;
      
      const source = audioContext.createMediaStreamSource(mediaStream);
      
      // Create noise gate
      const noiseGate = audioContext.createBiquadFilter();
      noiseGate.type = 'highpass';
      noiseGate.frequency.value = 80;
      
      // Create low-pass filter to remove high frequency noise
      const lowPass = audioContext.createBiquadFilter();
      lowPass.type = 'lowpass';
      lowPass.frequency.value = 15000;
      
      // Create compressor for dynamic range
      const compressor = audioContext.createDynamicsCompressor();
      compressor.threshold.value = -50;
      compressor.knee.value = 40;
      compressor.ratio.value = 12;
      compressor.attack.value = 0;
      compressor.release.value = 0.25;
      
      // Create analyser for level meter
      analyser = audioContext.createAnalyser();
      analyser.fftSize = 256;
      
      // Connect nodes: source -> highpass -> lowpass -> compressor -> analyser
      source.connect(noiseGate);
      noiseGate.connect(lowPass);
      lowPass.connect(compressor);
      compressor.connect(analyser);
      
      // Create script processor for recording
      const bufferSize = 4096;
      const scriptProcessor = audioContext.createScriptProcessor(bufferSize, 1, 1);
      
      recordedChunks = [];
      
      scriptProcessor.onaudioprocess = (e) => {
        if (!isRecording) return;
        const inputData = e.inputBuffer.getChannelData(0);
        recordedChunks.push(new Float32Array(inputData));
      };
      
      analyser.connect(scriptProcessor);
      scriptProcessor.connect(audioContext.destination);
      
      isRecording = true;
      recordingTime = 0;
      
      timerInterval = window.setInterval(() => {
        recordingTime++;
        if (recordingTime >= maxDuration) {
          stopRecording();
        }
      }, 1000);
      
      drawAudioLevel();
      
      // Store for cleanup
      (window as any).__scriptProcessor = scriptProcessor;
      (window as any).__source = source;

    } catch (error) {
      console.error('Failed to start recording:', error);
      dispatch('error', { message: 'Microphone access denied' });
    }
  }

  function stopRecording() {
    if (!isRecording) return;
    
    isRecording = false;
    
    if (timerInterval) {
      clearInterval(timerInterval);
      timerInterval = null;
    }
    
    if (animationId) {
      cancelAnimationFrame(animationId);
      animationId = null;
    }
    
    // Process recorded audio
    processRecordedAudio();
  }

  async function processRecordedAudio() {
    if (recordedChunks.length === 0) {
      dispatch('error', { message: 'No audio recorded' });
      return;
    }
    
    // Combine all chunks
    const totalLength = recordedChunks.reduce((acc, chunk) => acc + chunk.length, 0);
    const combinedBuffer = new Float32Array(totalLength);
    let offset = 0;
    for (const chunk of recordedChunks) {
      combinedBuffer.set(chunk, offset);
      offset += chunk.length;
    }
    
    // Trim silence from start and end
    const trimmedBuffer = trimSilence(combinedBuffer, 0.01);
    
    // Normalize audio
    const normalizedBuffer = normalizeAudio(trimmedBuffer);
    
    // Convert to WAV
    const wavBlob = float32ToWav(normalizedBuffer, sampleRate);
    
    audioBlob = wavBlob;
    audioUrl = URL.createObjectURL(wavBlob);
    
    // Cleanup
    if (mediaStream) {
      mediaStream.getTracks().forEach(track => track.stop());
      mediaStream = null;
    }
    
    if (audioContext) {
      audioContext.close();
      audioContext = null;
    }
    
    dispatch('recorded', { blob: audioBlob, url: audioUrl });
  }

  function trimSilence(buffer: Float32Array, threshold: number): Float32Array {
    let start = 0;
    let end = buffer.length;
    
    // Find start (skip silence)
    for (let i = 0; i < buffer.length; i++) {
      if (Math.abs(buffer[i]) > threshold) {
        start = Math.max(0, i - 1000); // Keep 1000 samples before
        break;
      }
    }
    
    // Find end (skip silence)
    for (let i = buffer.length - 1; i >= 0; i--) {
      if (Math.abs(buffer[i]) > threshold) {
        end = Math.min(buffer.length, i + 1000); // Keep 1000 samples after
        break;
      }
    }
    
    return buffer.slice(start, end);
  }

  function normalizeAudio(buffer: Float32Array): Float32Array {
    // Find max amplitude
    let maxAmplitude = 0;
    for (let i = 0; i < buffer.length; i++) {
      const absValue = Math.abs(buffer[i]);
      if (absValue > maxAmplitude) {
        maxAmplitude = absValue;
      }
    }
    
    // Normalize to 90% of max
    if (maxAmplitude > 0) {
      const gain = 0.9 / maxAmplitude;
      for (let i = 0; i < buffer.length; i++) {
        buffer[i] *= gain;
      }
    }
    
    return buffer;
  }

  function float32ToWav(samples: Float32Array, sampleRate: number): Blob {
    const numChannels = 1;
    const bitDepth = 16;
    const bytesPerSample = bitDepth / 8;
    const blockAlign = numChannels * bytesPerSample;
    
    const wavBuffer = new ArrayBuffer(44 + samples.length * bytesPerSample);
    const view = new DataView(wavBuffer);
    
    // WAV header
    writeString(view, 0, 'RIFF');
    view.setUint32(4, 36 + samples.length * bytesPerSample, true);
    writeString(view, 8, 'WAVE');
    writeString(view, 12, 'fmt ');
    view.setUint32(16, 16, true);
    view.setUint16(20, 1, true);
    view.setUint16(22, numChannels, true);
    view.setUint32(24, sampleRate, true);
    view.setUint32(28, sampleRate * blockAlign, true);
    view.setUint16(32, blockAlign, true);
    view.setUint16(34, bitDepth, true);
    writeString(view, 36, 'data');
    view.setUint32(40, samples.length * bytesPerSample, true);
    
    // Convert float32 to int16
    let offset = 44;
    for (let i = 0; i < samples.length; i++) {
      const sample = Math.max(-1, Math.min(1, samples[i]));
      const int16 = sample < 0 ? sample * 0x8000 : sample * 0x7FFF;
      view.setInt16(offset, int16, true);
      offset += 2;
    }
    
    return new Blob([wavBuffer], { type: 'audio/wav' });
  }

  function writeString(view: DataView, offset: number, string: string) {
    for (let i = 0; i < string.length; i++) {
      view.setUint8(offset + i, string.charCodeAt(i));
    }
  }

  function drawAudioLevel() {
    if (!analyser) return;
    
    const dataArray = new Uint8Array(analyser.frequencyBinCount);
    
    function update() {
      if (!isRecording || !analyser) return;
      
      analyser.getByteFrequencyData(dataArray);
      
      let sum = 0;
      for (let i = 0; i < dataArray.length; i++) {
        sum += dataArray[i];
      }
      audioLevel = sum / dataArray.length / 255;

      animationId = requestAnimationFrame(update);
    }
    
    update();
  }

  function clearRecording() {
    if (audioUrl) {
      URL.revokeObjectURL(audioUrl);
    }
    audioBlob = null;
    audioUrl = null;
    recordingTime = 0;
    recordedChunks = [];
    dispatch('clear');
  }

  function formatTime(seconds: number): string {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  }

  onDestroy(() => {
    if (timerInterval) clearInterval(timerInterval);
    if (animationId) cancelAnimationFrame(animationId);
    if (audioContext) audioContext.close();
    if (audioUrl) URL.revokeObjectURL(audioUrl);
    if (mediaStream) {
      mediaStream.getTracks().forEach(track => track.stop());
    }
  });
</script>

<div class="recorder">
  {#if !isRecording && !audioBlob}
    <button 
      class="record-btn spring-bounce" 
      on:click={startRecording}
      disabled={isProcessing}
      title="Record from microphone"
    >
      <div class="mic-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
          <path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
          <line x1="12" y1="19" x2="12" y2="23"/>
          <line x1="8" y1="23" x2="16" y2="23"/>
        </svg>
      </div>
      <div class="btn-text">
        <span class="btn-title">Record</span>
        <span class="btn-hint">Up to {maxDuration}s</span>
      </div>
    </button>
  {:else if isRecording}
    <div class="recording-active">
      <div class="recording-indicator">
        <div class="pulse-dot"></div>
        <span class="recording-text">Recording</span>
      </div>
      
      <div class="level-meter">
        <div class="level-bar" style="width: {audioLevel * 100}%"></div>
      </div>
      
      <div class="timer">
        <span class="time">{formatTime(recordingTime)}</span>
        <span class="max-time">/ {formatTime(maxDuration)}</span>
      </div>
      
      <button class="stop-btn" on:click={stopRecording}>
        <svg viewBox="0 0 24 24" fill="currentColor">
          <rect x="6" y="6" width="12" height="12" rx="2"/>
        </svg>
        Stop
      </button>
    </div>
  {:else if audioBlob}
    <div class="recording-preview">
      <div class="preview-header">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
          <path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
        </svg>
        <span>Recording ready</span>
        <span class="duration">{formatTime(recordingTime)}</span>
      </div>
      
      {#if audioUrl}
        <audio src={audioUrl} controls class="audio-player"></audio>
      {/if}
      
      <div class="preview-actions">
        <button class="action-btn discard" on:click={clearRecording}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6"/>
          </svg>
          Discard
        </button>
      </div>
    </div>
  {/if}
</div>

<style>
  .recorder {
    width: 100%;
  }

  .record-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1rem;
    width: 100%;
    padding: 1.25rem;
    background: linear-gradient(135deg, rgba(6, 182, 212, 0.1), rgba(139, 92, 246, 0.1));
    border: 2px solid rgba(6, 182, 212, 0.3);
    border-radius: 1rem;
    color: var(--text-primary);
    font-size: 0.9375rem;
    font-weight: 500;
    transition: all 0.3s;
  }

  .record-btn:hover:not(:disabled) {
    border-color: var(--accent-cyan);
    background: linear-gradient(135deg, rgba(6, 182, 212, 0.2), rgba(139, 92, 246, 0.2));
    transform: translateY(-2px);
    box-shadow: 0 10px 30px rgba(6, 182, 212, 0.2);
  }

  .record-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .mic-icon {
    width: 32px;
    height: 32px;
    color: var(--accent-cyan);
  }

  .mic-icon svg {
    width: 100%;
    height: 100%;
  }

  .btn-text {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.125rem;
  }

  .btn-title {
    font-weight: 600;
    font-size: 1rem;
  }

  .btn-hint {
    font-size: 0.75rem;
    color: var(--text-secondary);
    font-weight: 400;
  }

  .recording-active {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
    padding: 1.5rem;
    background: linear-gradient(135deg, rgba(239, 68, 68, 0.1), rgba(239, 68, 68, 0.05));
    border: 1px solid rgba(239, 68, 68, 0.3);
    border-radius: 1rem;
  }

  .recording-indicator {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .pulse-dot {
    width: 12px;
    height: 12px;
    background: var(--error);
    border-radius: 50%;
    animation: pulse-dot 1s ease-in-out infinite;
  }

  @keyframes pulse-dot {
    0%, 100% { opacity: 1; transform: scale(1); }
    50% { opacity: 0.5; transform: scale(0.8); }
  }

  .recording-text {
    font-weight: 600;
    color: var(--error);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-size: 0.875rem;
  }

  .level-meter {
    width: 100%;
    height: 8px;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 4px;
    overflow: hidden;
  }

  .level-bar {
    height: 100%;
    background: linear-gradient(90deg, var(--accent-cyan), var(--accent-purple));
    border-radius: 4px;
    transition: width 0.05s;
    min-width: 2%;
  }

  .timer {
    display: flex;
    align-items: baseline;
    gap: 0.25rem;
  }

  .time {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--text-primary);
    font-variant-numeric: tabular-nums;
  }

  .max-time {
    font-size: 0.875rem;
    color: var(--text-muted);
  }

  .stop-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1.5rem;
    background: var(--error);
    color: white;
    border-radius: 0.5rem;
    font-weight: 500;
    font-size: 0.875rem;
    transition: all 0.2s;
  }

  .stop-btn svg {
    width: 16px;
    height: 16px;
  }

  .stop-btn:hover {
    background: #dc2626;
    transform: scale(1.02);
  }

  .recording-preview {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding: 1rem;
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: 1rem;
  }

  .preview-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    color: var(--text-secondary);
  }

  .preview-header svg {
    width: 18px;
    height: 18px;
    color: var(--accent-cyan);
  }

  .duration {
    margin-left: auto;
    font-weight: 500;
    color: var(--text-primary);
    font-variant-numeric: tabular-nums;
  }

  .audio-player {
    width: 100%;
    height: 40px;
    border-radius: 0.5rem;
  }

  .preview-actions {
    display: flex;
    gap: 0.5rem;
  }

  .action-btn {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.75rem;
    border-radius: 0.5rem;
    font-size: 0.8125rem;
    font-weight: 500;
    transition: all 0.2s;
  }

  .action-btn svg {
    width: 16px;
    height: 16px;
  }

  .action-btn.discard {
    background: rgba(239, 68, 68, 0.1);
    color: var(--error);
  }

  .action-btn.discard:hover {
    background: rgba(239, 68, 68, 0.2);
  }
</style>
