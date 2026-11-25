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
	}: {
		phase?: UpdateHeaderPhase;
		versionLabel?: string;
		latest?: string;
		errorTitle?: string;
		onCheck?: () => void;
		onDownload?: () => void;
	} = $props();

	let checking = $derived(phase === 'checking');
	let downloading = $derived(phase === 'downloading');
	let downloadVersion = $derived(`v${(latest || 'update').replace(/^v/i, '')}`);
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
			class="btn preset-filled-success-500 btn-sm gap-0 overflow-hidden p-0"
			aria-label={`Download ${downloadVersion}`}
			title={`Download ${downloadVersion}`}
			onclick={() => onDownload?.()}
		>
			<span
				class="inline-flex items-center justify-center border-r border-white/25 px-1.5 py-0.5"
			>
				<ControlIcon name="download" />
			</span>
			<span class="px-1.5 font-mono text-[0.7rem] leading-none">{downloadVersion}</span>
		</button>
	{/if}
	{#if downloading}
		<span
			class="inline-flex size-7 items-center justify-center"
			role="status"
			aria-label="Downloading update"
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
			disabled={checking}
			aria-label={checking ? 'Checking for updates' : 'Check for updates'}
			aria-busy={checking || undefined}
			title={checking ? 'Checking for updates' : 'Check for updates'}
			onclick={() => onCheck?.()}
		>
			<span class="inline-flex" class:animate-spin={checking}>
				<ControlIcon name="update" />
			</span>
		</button>
	{/if}
	<p class="truncate text-xs {versionClass}" role={phase === 'error' ? 'alert' : 'status'} title={versionTitle}>
		{versionLabel}
	</p>
</div>
