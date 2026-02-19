<script lang="ts">
  import { onMount, onDestroy } from 'svelte';

  export let audioUrl: string | null = null;

  let canvas: HTMLCanvasElement;
  let audioContext: AudioContext | null = null;
  let analyser: AnalyserNode | null = null;
  let source: AudioBufferSourceNode | null = null;
  let animationId: number | null = null;
  let audioBuffer: AudioBuffer | null = null;

  async function loadAudio() {
    if (!audioUrl || !canvas) return;

    try {
      if (!audioContext) {
        audioContext = new AudioContext();
      }

      const response = await fetch(audioUrl);
      const arrayBuffer = await response.arrayBuffer();
      audioBuffer = await audioContext.decodeAudioData(arrayBuffer);
      
      drawWaveform();
    } catch (error) {
      console.error('Failed to load audio:', error);
    }
  }

  function drawWaveform() {
    if (!audioBuffer || !canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const data = audioBuffer.getChannelData(0);
    const width = canvas.width;
    const height = canvas.height;
    const step = Math.ceil(data.length / width);
    const amp = height / 2;

    ctx.fillStyle = 'transparent';
    ctx.fillRect(0, 0, width, height);

    const gradient = ctx.createLinearGradient(0, 0, width, 0);
    gradient.addColorStop(0, '#06b6d4');
    gradient.addColorStop(1, '#8b5cf6');

    ctx.fillStyle = gradient;
    ctx.beginPath();
    ctx.moveTo(0, amp);

    for (let i = 0; i < width; i++) {
      let min = 1.0;
      let max = -1.0;
      
      for (let j = 0; j < step; j++) {
        const datum = data[(i * step) + j];
        if (datum < min) min = datum;
        if (datum > max) max = datum;
      }

      ctx.lineTo(i, (1 + min) * amp);
    }

    for (let i = width - 1; i >= 0; i--) {
      let min = 1.0;
      let max = -1.0;
      
      for (let j = 0; j < step; j++) {
        const datum = data[(i * step) + j];
        if (datum < min) min = datum;
        if (datum > max) max = datum;
      }

      ctx.lineTo(i, (1 + max) * amp);
    }

    ctx.closePath();
    ctx.fill();
  }

  function resizeCanvas() {
    if (!canvas) return;
    canvas.width = canvas.offsetWidth * window.devicePixelRatio;
    canvas.height = canvas.offsetHeight * window.devicePixelRatio;
    if (audioBuffer) {
      drawWaveform();
    }
  }

  onMount(() => {
    resizeCanvas();
    window.addEventListener('resize', resizeCanvas);
  });

  onDestroy(() => {
    window.removeEventListener('resize', resizeCanvas);
    if (animationId) {
      cancelAnimationFrame(animationId);
    }
    if (audioContext) {
      audioContext.close();
    }
  });

  $: if (audioUrl) {
    loadAudio();
  }
</script>

<div class="waveform-container">
  <canvas bind:this={canvas} class="waveform-canvas"></canvas>
</div>

<style>
  .waveform-container {
    width: 100%;
    height: 80px;
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: 0.75rem;
    overflow: hidden;
  }

  .waveform-canvas {
    width: 100%;
    height: 100%;
    display: block;
  }
</style>
