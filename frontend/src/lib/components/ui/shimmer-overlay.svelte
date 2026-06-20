<script lang="ts">
  import { cn, type WithElementRef, type WithoutChildren } from '$lib/utils.js'
  import type { HTMLAttributes } from 'svelte/elements'

  let {
    ref = $bindable(null),
    class: className,
    ...restProps
  }: WithoutChildren<WithElementRef<HTMLAttributes<HTMLDivElement>>> = $props()
</script>

<div
  bind:this={ref}
  data-slot="shimmer-overlay"
  class={cn(
    'pointer-events-none absolute inset-0 z-10 rounded-lg bg-gradient-to-br from-muted/20 via-background/10 to-muted/20 opacity-80 transition-opacity duration-500',
    className,
  )}
  {...restProps}
>
  <div class="shimmer size-full"></div>
</div>

<style>
  .shimmer {
    position: relative;
    overflow: hidden;
  }
  .shimmer::after {
    content: '';
    position: absolute;
    inset: -50%;
    background: linear-gradient(
      135deg,
      transparent 30%,
      rgba(255, 255, 255, 0.06) 50%,
      transparent 70%
    );
    background-size: 200% 200%;
    animation: shimmer-overlay 4s ease-in-out infinite;
  }
  @keyframes shimmer-overlay {
    0%, 100% { background-position: 0% 0%; }
    50% { background-position: 100% 100%; }
  }
</style>
