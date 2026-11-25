<script lang="ts">
	import ControlIcon from '#lib/components/ControlIcon.svelte';

	export type UpdateHeaderPhase =
		| 'idle'
		| 'checking'
		| 'up-to-date'
		| 'available'
		| 'downloading'
		| 'error';

	let {
		phase = 'idle',
		versionLabel = 'ver: …',
		latest = '',
		errorTitle = '',
		onCheck,
		onDownload,
		interactive = true,
	}: {
		phase?: UpdateHeaderPhase;
		versionLabel?: string;
		latest?: string;
		errorTitle?: string;
		onCheck?: () => void;
		onDownload?: () => void;
		interactive?: boolean;
	} = $props();

	let busy = $derived(phase === 'checking' || phase === 'downloading');
	let versionClass = $derived(
		phase === 'error'
			? 'text-error-500'
			: phase === 'up-to-date'
				? 'text-success-700-300'
				: 'text-surface-700-300',
	);
	let versionTitle = $derived(
		phase === 'error' && errorTitle
			? errorTitle
			: phase === 'up-to-date'
				? 'Up to date'
				: versionLabel,
	);
</script>

<div class="flex min-w-0 flex-wrap items-center justify-end gap-x-1.5 gap-y-1">
	{#if phase === 'available'}
		<button
			type="button"
			class="btn preset-filled-primary-500 btn-sm"
			disabled={!interactive}
			onclick={() => onDownload?.()}
		>
			Download {latest || 'update'}
		</button>
	{/if}
	{#if busy}
		<span
			class="inline-flex size-7 items-center justify-center"
			role="status"
			aria-label={phase === 'downloading' ? 'Downloading update' : 'Checking for updates'}
		>
			<svg class="size-4 animate-spin text-surface-700-300" viewBox="0 0 24 24" fill="none" aria-hidden="true">
				<circle class="opacity-25" cx="12" cy="12" r="9" stroke="currentColor" stroke-width="2.5" />
				<path
					class="opacity-90"
					d="M21 12a9 9 0 0 0-9-9"
					stroke="currentColor"
					stroke-width="2.5"
					stroke-linecap="round"
				/>
			</svg>
		</span>
	{:else}
		<button
			type="button"
			class="btn-icon preset-ghost size-7"
			disabled={!interactive}
			aria-label="Check for updates"
			title="Check for updates"
			onclick={() => onCheck?.()}
		>
			<ControlIcon name="update" />
		</button>
	{/if}
	<p class="truncate text-xs {versionClass}" role={phase === 'error' ? 'alert' : 'status'} title={versionTitle}>
		{versionLabel}
	</p>
</div>
