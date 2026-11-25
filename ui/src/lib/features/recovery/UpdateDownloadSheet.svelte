<script lang="ts">
	import { Dialog } from '@skeletonlabs/skeleton-svelte';
	import ResponsiveDialogContent from '#lib/components/ResponsiveDialogContent.svelte';
	import ControlIcon from '#lib/components/ControlIcon.svelte';
	import ChangelogNotes from './ChangelogNotes.svelte';

	let {
		open = $bindable(false),
		latest = '',
		notes = '',
		busy = false,
		onConfirm,
		onCancel,
	}: {
		open?: boolean;
		latest?: string;
		notes?: string;
		busy?: boolean;
		onConfirm?: () => void;
		onCancel?: () => void;
	} = $props();

	let versionLabel = $derived(`v${(latest || 'update').replace(/^v/i, '')}`);

	function cancel() {
		open = false;
		onCancel?.();
	}
</script>

<Dialog
	preventScroll
	closeOnInteractOutside={!busy}
	{open}
	onOpenChange={({ open: next }) => {
		open = next;
		if (!next) onCancel?.();
	}}
>
	<ResponsiveDialogContent title={`Update ${versionLabel}`} width="sm:max-w-lg" elevated>
		<p class="mb-3 text-sm text-surface-700-300">
			Review what’s new, then confirm to download the installer.
		</p>
		<ChangelogNotes {notes} empty="No release notes yet for this version." />
		<div class="mt-5 flex flex-wrap items-center justify-end gap-2">
			<button type="button" class="btn preset-ghost" disabled={busy} onclick={cancel}>
				Cancel
			</button>
			<button
				type="button"
				class="btn preset-filled-success-500 gap-0 overflow-hidden p-0"
				disabled={busy}
				aria-label={`Confirm download ${versionLabel}`}
				onclick={() => onConfirm?.()}
			>
				<span class="inline-flex items-center justify-center border-r border-white/25 px-2.5 py-2">
					{#if busy}
						<svg class="size-4 animate-spin" viewBox="0 0 24 24" fill="none" aria-hidden="true">
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
				<span class="px-3 font-mono text-sm leading-none">{versionLabel}</span>
			</button>
		</div>
	</ResponsiveDialogContent>
</Dialog>
