<script lang="ts">
	import BrandMark from '#lib/components/BrandMark.svelte';
	import BrandWordmark from '#lib/components/BrandWordmark.svelte';

	let { fading = false }: { fading?: boolean } = $props();
</script>

<div class="boot-splash" class:fading role="status" aria-live="polite" aria-label="Loading WreckScope">
	<div class="boot-splash-brand">
		<BrandMark size={72} />
		<BrandWordmark class="boot-splash-word" />
	</div>
</div>

<style>
	.boot-splash {
		position: fixed;
		inset: 0;
		z-index: 100;
		display: grid;
		place-items: center;
		background: light-dark(var(--color-surface-100, #f4f6f8), var(--color-surface-950, #111820));
		transition: opacity 280ms ease;
	}
	.boot-splash.fading {
		opacity: 0;
		pointer-events: none;
	}
	.boot-splash-brand {
		display: grid;
		justify-items: center;
		gap: 0.85rem;
	}
	.boot-splash-brand :global(.boot-splash-word) {
		font-size: 1.35rem;
		letter-spacing: 0.04em;
	}
</style>
