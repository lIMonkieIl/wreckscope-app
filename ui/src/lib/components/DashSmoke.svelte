<script lang="ts">
	import BrandWordmark from '#lib/components/BrandWordmark.svelte';

	let {
		message = '',
		heavy = false,
		contained = false,
		onAction,
	}: {
		message?: string;
		heavy?: boolean;
		/** Absolute in a positioned parent (not full-viewport fixed). */
		contained?: boolean;
		onAction?: () => void;
	} = $props();
</script>

<div
	class="dash-smoke"
	class:heavy
	class:contained
	role={message ? 'status' : 'presentation'}
	aria-live={message ? 'polite' : undefined}
>
	{#if message}
		{#if onAction}
			<button type="button" class="msg action" onclick={onAction}>
				<BrandWordmark />
				<span>{message}</span>
			</button>
		{:else}
			<p class="msg">
				<BrandWordmark />
				<span>{message}</span>
			</p>
		{/if}
	{/if}
</div>

<style>
	.dash-smoke {
		position: fixed;
		inset: 0;
		z-index: 70;
		display: grid;
		place-items: center;
		padding: 1rem;
		background: color-mix(in oklab, var(--color-surface-950, #0b0b0b) 68%, transparent);
		pointer-events: auto;
	}
	.dash-smoke.contained {
		position: absolute;
		z-index: 20;
	}
	.dash-smoke.heavy {
		background: color-mix(in oklab, var(--color-surface-950, #0b0b0b) 78%, transparent);
	}
	.msg {
		display: grid;
		gap: 0.35rem;
		margin: 0;
		padding: 0.85rem 1.1rem;
		border-radius: var(--radius-container, 0.75rem);
		background: light-dark(var(--color-surface-50, #fff), var(--color-surface-900, #161616));
		border: 1px solid light-dark(var(--color-surface-300, #ccc), var(--color-surface-700, #333));
		text-align: center;
		color: inherit;
		font: inherit;
	}
	.msg.action {
		cursor: pointer;
	}
	.msg.action:hover {
		border-color: light-dark(var(--color-primary-500, #888), var(--color-primary-400, #aaa));
	}
	.msg strong {
		font-size: 0.95rem;
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}
	.msg span {
		font-size: 0.85rem;
		opacity: 0.85;
	}
</style>
