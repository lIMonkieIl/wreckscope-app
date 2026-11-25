<script lang="ts">
	import { Dialog } from '@skeletonlabs/skeleton-svelte';
	import ResponsiveDialogContent from '#lib/components/ResponsiveDialogContent.svelte';
	import ChangelogNotes from './ChangelogNotes.svelte';

	let {
		open = $bindable(false),
		version = '',
		notes = '',
		onDismiss,
	}: {
		open?: boolean;
		version?: string;
		notes?: string;
		onDismiss?: () => void;
	} = $props();

	let versionLabel = $derived(`v${(version || '').replace(/^v/i, '')}`);

	function dismiss() {
		open = false;
		onDismiss?.();
	}
</script>

<Dialog
	preventScroll
	closeOnInteractOutside
	{open}
	onOpenChange={({ open: next }) => {
		open = next;
		if (!next) onDismiss?.();
	}}
>
	<ResponsiveDialogContent
		title={`What’s new in ${versionLabel || 'this version'}`}
		width="sm:max-w-md"
		elevated
	>
		<ChangelogNotes {notes} empty="No release notes were published for this version." />
		<div class="mt-5 flex justify-end">
			<button type="button" class="btn preset-filled-primary-500" onclick={dismiss}>
				Got it
			</button>
		</div>
	</ResponsiveDialogContent>
</Dialog>
