<script lang="ts">
  import { onMount, onDestroy } from 'svelte';

  export let audioUrl: string | null = null;
  export let audioElement: HTMLAudioElement | null = null;

  let canvas: HTMLCanvasElement;
  let audioContext: AudioContext | null = null;
  let analyser: AnalyserNode | null = null;
  let dataArray: Uint8Array | null = null;
  let animationId: number | null = null;
  let isPlaying = false;
  let audio: HTMLAudioElement | null = null;
  let sourceNode: MediaElementAudioSourceNode | null = null;

  async function setupAudio() {
    if (!audioUrl) return;

    try {
      if (!audioContext) {
        audioContext = new AudioContext();
      }

      if (!audio) {
        audio = new Audio(audioUrl);
        audio.crossOrigin = 'anonymous';
        audioElement = audio;
      } else {
        audio.src = audioUrl;
      }

      if (!sourceNode) {
        sourceNode = audioContext.createMediaElementSource(audio);
      }

      analyser = audioContext.createAnalyser();
      analyser.fftSize = 256;
      
      const bufferLength = analyser.frequencyBinCount;
      dataArray = new Uint8Array(bufferLength) as Uint8Array<ArrayBuffer>;

      sourceNode.connect(analyser);
      analyser.connect(audioContext.destination);

      audio.addEventListener('play', () => {
        isPlaying = true;
        draw();
      });

      audio.addEventListener('pause', () => {
        isPlaying = false;
        if (animationId) {
          cancelAnimationFrame(animationId);
        }
      });

      audio.addEventListener('ended', () => {
        isPlaying = false;
      });

    } catch (error) {
      console.error('Failed to setup audio:', error);
    }
  }

  function draw() {
    if (!analyser || !dataArray || !canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    animationId = requestAnimationFrame(draw);

    // @ts-ignore - TypeScript has issues with Uint8Array type inference
    analyser.getByteFrequencyData(dataArray);

    const width = canvas.width;
    const height = canvas.height;

    ctx.fillStyle = 'rgba(0, 0, 0, 0.2)';
    ctx.fillRect(0, 0, width, height);

    const barWidth = (width / dataArray.length) * 2.5;
    let x = 0;

    for (let i = 0; i < dataArray.length; i++) {
      const barHeight = (dataArray[i] / 255) * height;

      const gradient = ctx.createLinearGradient(0, height, 0, height - barHeight);
      gradient.addColorStop(0, '#06b6d4');
      gradient.addColorStop(1, '#8b5cf6');

      ctx.fillStyle = gradient;
      ctx.fillRect(x, height - barHeight, barWidth - 1, barHeight);

      x += barWidth;
    }
  }

  function resizeCanvas() {
    if (!canvas) return;
    canvas.width = canvas.offsetWidth * window.devicePixelRatio;
    canvas.height = canvas.offsetHeight * window.devicePixelRatio;
  }

  function togglePlay() {
    if (!audio) return;

    if (audioContext && audioContext.state === 'suspended') {
      audioContext.resume();
    }

    if (isPlaying) {
      audio.pause();
    } else {
      audio.play();
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
    if (audio) {
      audio.pause();
      audio = null;
    }
    if (audioContext) {
      audioContext.close();
    }
  });

  $: if (audioUrl) {
    setupAudio();
  }
</script>

<div class="visualizer-container">
  <canvas bind:this={canvas} class="visualizer-canvas"></canvas>
  {#if audio}
    <div class="controls">
      <button class="play-btn" on:click={togglePlay}>
        {#if isPlaying}
          <svg viewBox="0 0 24 24" fill="currentColor">
            <rect x="6" y="4" width="4" height="16"/>
            <rect x="14" y="4" width="4" height="16"/>
          </svg>
        {:else}
          <svg viewBox="0 0 24 24" fill="currentColor">
            <polygon points="5,3 19,12 5,21"/>
          </svg>
        {/if}
      </button>
    </div>
  {/if}
</div>

<style>
  .visualizer-container {
    width: 100%;
    height: 120px;
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: 0.75rem;
    overflow: hidden;
    position: relative;
  }

  .visualizer-canvas {
    width: 100%;
    height: 100%;
    display: block;
  }

  .controls {
    position: absolute;
    bottom: 0.75rem;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .play-btn {
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(6, 182, 212, 0.2);
    border-radius: 50%;
    color: var(--accent-cyan);
    transition: all 0.2s;
    backdrop-filter: blur(10px);
  }

  .play-btn:hover {
    background: rgba(6, 182, 212, 0.3);
    transform: scale(1.1);
  }

  .play-btn svg {
    width: 18px;
    height: 18px;
  }
</style>
