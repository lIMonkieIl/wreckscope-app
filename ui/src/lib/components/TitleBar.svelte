<script lang="ts">
	import type { Snippet } from 'svelte';
	import { onMount } from 'svelte';
	import BrandMark from '#lib/components/BrandMark.svelte';
	import BrandWordmark from '#lib/components/BrandWordmark.svelte';
	import { hasDesktopBridge, whenDesktopRuntimeReady } from '#lib/features/recovery/desktopBridge.ts';

	let { actions }: { actions?: Snippet } = $props();

	let maximised = $state(false);
	let windowReady = $state(false);
	let show = $state(hasDesktopBridge());

	async function loadWindow() {
		const { Window } = await import('@wailsio/runtime');
		return Window;
	}

	async function refreshMaximised() {
		try {
			const Window = await loadWindow();
			maximised = await Window.IsMaximised();
		} catch {
			/* phone / bridge not ready */
		}
	}

	async function minimise() {
		try {
			(await loadWindow()).Minimise();
		} catch {
			/* ignore */
		}
	}

	async function toggleMaximise() {
		try {
			await (await loadWindow()).ToggleMaximise();
			await refreshMaximised();
		} catch {
			/* ignore */
		}
	}

	async function closeWindow() {
		try {
			(await loadWindow()).Close();
		} catch {
			/* ignore */
		}
	}

	onMount(() => {
		show = hasDesktopBridge();
		if (!show) return;
		let cancelled = false;
		void (async () => {
			if (!(await whenDesktopRuntimeReady()) || cancelled) return;
			windowReady = true;
			await refreshMaximised();
		})();
		return () => {
			cancelled = true;
		};
	});
</script>

{#if show}
	<header class="title-bar" aria-label="Window">
		<div
			class="title-bar-drag"
			role="presentation"
			ondblclick={() => {
				if (windowReady) void toggleMaximise();
			}}
		>
			<BrandMark size={22} class="title-bar-mark" />
			<BrandWordmark class="title-bar-word" />
		</div>
		{#if actions}
			<div class="title-bar-actions">
				{@render actions()}
			</div>
			<div class="title-bar-divider" aria-hidden="true"></div>
		{/if}
		<div class="title-bar-window">
			<button
				type="button"
				class="title-bar-btn"
				aria-label="Minimise"
				title="Minimise"
				disabled={!windowReady}
				onclick={() => void minimise()}
			>
				<svg viewBox="0 0 12 12" aria-hidden="true"
					><path d="M2 6h8" fill="none" stroke="currentColor" stroke-width="1.2" /></svg
				>
			</button>
			<button
				type="button"
				class="title-bar-btn"
				aria-label={maximised ? 'Restore' : 'Maximise'}
				title={maximised ? 'Restore' : 'Maximise'}
				disabled={!windowReady}
				onclick={() => void toggleMaximise()}
			>
				{#if maximised}
					<svg viewBox="0 0 12 12" aria-hidden="true">
						<path
							d="M3.5 4.5h4v4h-4zM4.5 3.5h4v4"
							fill="none"
							stroke="currentColor"
							stroke-width="1.2"
						/>
					</svg>
				{:else}
					<svg viewBox="0 0 12 12" aria-hidden="true">
						<rect
							x="2.5"
							y="2.5"
							width="7"
							height="7"
							fill="none"
							stroke="currentColor"
							stroke-width="1.2"
						/>
					</svg>
				{/if}
			</button>
			<button
				type="button"
				class="title-bar-btn title-bar-close"
				aria-label="Close"
				title="Close"
				disabled={!windowReady}
				onclick={() => void closeWindow()}
			>
				<svg viewBox="0 0 12 12" aria-hidden="true"
					><path
						d="M3 3l6 6M9 3 3 9"
						fill="none"
						stroke="currentColor"
						stroke-width="1.2"
					/></svg
				>
			</button>
		</div>
	</header>
{/if}

<style>
	.title-bar {
		display: flex;
		align-items: stretch;
		height: 2.5rem;
		flex: none;
		background: light-dark(var(--color-surface-50), var(--color-surface-900));
		border-bottom: 1px solid light-dark(var(--color-surface-300), var(--color-surface-700));
		color: light-dark(var(--color-surface-800), var(--color-surface-200));
		user-select: none;
	}
	.title-bar-drag {
		--wails-draggable: drag;
		app-region: drag;
		display: flex;
		align-items: center;
		gap: 0.55rem;
		min-width: 0;
		flex: 1;
		padding: 0 0.75rem;
	}
	.title-bar-drag :global(.title-bar-mark) {
		flex: none;
	}
	.title-bar-drag :global(.title-bar-word) {
		font-size: 0.78rem;
		letter-spacing: 0.04em;
		line-height: 1;
	}
	.title-bar-actions {
		--wails-draggable: no-drag;
		app-region: no-drag;
		display: flex;
		align-items: center;
		gap: 0.15rem;
		padding-inline: 0.25rem;
	}
	.title-bar-actions :global(.btn-icon) {
		width: 2rem;
		height: 2rem;
	}
	.title-bar-divider {
		width: 1px;
		align-self: center;
		height: 1.15rem;
		margin-inline: 0.15rem 0;
		background: light-dark(var(--color-surface-300), var(--color-surface-700));
	}
	.title-bar-window {
		--wails-draggable: no-drag;
		app-region: no-drag;
		display: flex;
		align-items: stretch;
		flex: none;
	}
	.title-bar-btn {
		display: grid;
		place-items: center;
		width: 2.75rem;
		margin: 0;
		padding: 0;
		border: 0;
		background: transparent;
		color: inherit;
		cursor: pointer;
	}
	.title-bar-btn svg {
		width: 0.7rem;
		height: 0.7rem;
	}
	.title-bar-btn:hover {
		background: light-dark(
			color-mix(in oklab, var(--color-surface-950) 8%, transparent),
			color-mix(in oklab, var(--color-surface-50) 12%, transparent)
		);
	}
	.title-bar-close:hover {
		background: var(--color-error-500);
		color: var(--color-surface-50);
	}
</style>
