<script lang="ts">
  export let progress = 0;
  export let size = 120;
  export let strokeWidth = 8;
  export let showValue = true;

  $: radius = (size - strokeWidth) / 2;
  $: circumference = 2 * Math.PI * radius;
  $: offset = circumference - (progress / 100) * circumference;
  $: gradientId = `gradient-${Math.random().toString(36).substr(2, 9)}`;
</script>

<div class="circular-progress" style="width: {size}px; height: {size}px;">
  <svg viewBox="0 0 {size} {size}">
    <defs>
      <linearGradient id={gradientId} x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" stop-color="#06b6d4" />
        <stop offset="50%" stop-color="#8b5cf6" />
        <stop offset="100%" stop-color="#ec4899" />
      </linearGradient>
    </defs>
    
    <circle
      cx={size / 2}
      cy={size / 2}
      r={radius}
      fill="none"
      stroke="rgba(255, 255, 255, 0.05)"
      stroke-width={strokeWidth}
    />
    
    <circle
      cx={size / 2}
      cy={size / 2}
      r={radius}
      fill="none"
      stroke="url(#{gradientId})"
      stroke-width={strokeWidth}
      stroke-linecap="round"
      stroke-dasharray={circumference}
      stroke-dashoffset={offset}
      transform="rotate(-90 {size / 2} {size / 2})"
      class="progress-ring"
    />
  </svg>
  
  {#if showValue}
    <div class="progress-value">
      <span class="value">{Math.round(progress)}</span>
      <span class="percent">%</span>
    </div>
  {/if}
</div>

<style>
  .circular-progress {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  svg {
    width: 100%;
    height: 100%;
  }

  .progress-ring {
    transition: stroke-dashoffset 0.5s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .progress-value {
    position: absolute;
    display: flex;
    align-items: baseline;
    justify-content: center;
    gap: 2px;
  }

  .value {
    font-size: 1.5rem;
    font-weight: 700;
    background: linear-gradient(135deg, var(--accent-cyan), var(--accent-purple));
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .percent {
    font-size: 0.875rem;
    color: var(--text-secondary);
    font-weight: 500;
  }
</style>
