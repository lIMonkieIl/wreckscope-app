<script lang="ts">
	import ControlIcon from '#lib/components/ControlIcon.svelte';

	let {
		latest = '',
		busy = false,
		onDownload,
		onLater,
	}: {
		latest?: string;
		busy?: boolean;
		onDownload?: () => void;
		onLater?: () => void;
	} = $props();

	let versionLabel = $derived(`v${(latest || 'update').replace(/^v/i, '')}`);
</script>

<div class="update-banner" role="status" aria-live="polite">
	<p class="update-banner-copy">
		Update <span class="font-mono">{versionLabel}</span> available
	</p>
	<div class="update-banner-actions">
		<button
			type="button"
			class="btn preset-filled-success-500 btn-sm gap-0 overflow-hidden p-0"
			disabled={busy}
			aria-label={`Download ${versionLabel}`}
			title={`Download ${versionLabel}`}
			onclick={() => onDownload?.()}
		>
			<span class="inline-flex items-center justify-center border-r border-white/25 px-1.5 py-0.5">
				{#if busy}
					<svg
						class="size-4 animate-spin"
						viewBox="0 0 24 24"
						fill="none"
						aria-hidden="true"
					>
						<circle class="opacity-25" cx="12" cy="12" r="9" stroke="currentColor" stroke-width="2.5" />
						<path
							class="opacity-90"
							d="M21 12a9 9 0 0 0-9-9"
							stroke="currentColor"
							stroke-width="2.5"
							stroke-linecap="round"
						/>
					</svg>
				{:else}
					<ControlIcon name="download" />
				{/if}
			</span>
			<span class="px-1.5 font-mono text-[0.7rem] leading-none">{versionLabel}</span>
		</button>
		<button type="button" class="btn preset-ghost btn-sm" disabled={busy} onclick={() => onLater?.()}>
			Later
		</button>
	</div>
</div>

<style>
	.update-banner {
		position: fixed;
		right: max(0.75rem, env(safe-area-inset-right));
		bottom: max(0.75rem, env(safe-area-inset-bottom));
		z-index: 60;
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.65rem 0.85rem;
		max-width: min(24rem, calc(100vw - 1.5rem));
		padding: 0.7rem 0.85rem;
		border-radius: var(--radius-container, 0.75rem);
		border: 1px solid light-dark(var(--color-surface-300, #ccc), var(--color-surface-700, #333));
		background: light-dark(var(--color-surface-50, #fff), var(--color-surface-900, #161616));
		box-shadow: 0 10px 30px color-mix(in oklab, #000 22%, transparent);
	}
	.update-banner-copy {
		margin: 0;
		font-size: 0.85rem;
		color: light-dark(var(--color-surface-800, #222), var(--color-surface-200, #ddd));
	}
	.update-banner-actions {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.35rem;
		margin-left: auto;
	}
</style>
