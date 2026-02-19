<script lang="ts">
  import { onMount, onDestroy } from 'svelte';

  export let trigger = false;

  interface Particle {
    id: number;
    x: number;
    color: string;
    delay: number;
    duration: number;
    size: number;
  }

  let particles: Particle[] = [];
  let interval: number | null = null;

  const colors = ['#06b6d4', '#8b5cf6', '#ec4899', '#10b981', '#f59e0b', '#ef4444'];

  function createParticles() {
    const newParticles: Particle[] = [];
    for (let i = 0; i < 50; i++) {
      newParticles.push({
        id: i,
        x: Math.random() * 100,
        color: colors[Math.floor(Math.random() * colors.length)],
        delay: Math.random() * 0.5,
        duration: 2 + Math.random() * 2,
        size: 6 + Math.random() * 8
      });
    }
    particles = newParticles;
  }

  $: if (trigger) {
    createParticles();
    setTimeout(() => {
      particles = [];
    }, 4000);
  }
</script>

<div class="confetti-container">
  {#each particles as particle (particle.id)}
    <div 
      class="confetti"
      style="
        left: {particle.x}%;
        background: {particle.color};
        width: {particle.size}px;
        height: {particle.size}px;
        animation-delay: {particle.delay}s;
        animation-duration: {particle.duration}s;
      "
    />
  {/each}
</div>

<style>
  .confetti-container {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
    z-index: 1000;
    overflow: hidden;
  }

  .confetti {
    position: absolute;
    top: -20px;
    border-radius: 2px;
    animation: confetti-fall linear forwards;
  }

  @keyframes confetti-fall {
    0% {
      transform: translateY(0) rotate(0deg);
      opacity: 1;
    }
    100% {
      transform: translateY(100vh) rotate(720deg);
      opacity: 0;
    }
  }
</style>
