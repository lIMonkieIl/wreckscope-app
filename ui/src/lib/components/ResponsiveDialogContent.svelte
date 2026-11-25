<script lang="ts">
	import { Dialog, Portal } from '@skeletonlabs/skeleton-svelte';
	import type { Snippet } from 'svelte';
	import ControlIcon from './ControlIcon.svelte';

	let {
		title,
		width = 'sm:max-w-md',
		children,
	}: { title: string; width?: string; children: Snippet } = $props();

	let offset = $state(0);
	let drag: { id: number; y: number } | undefined;
	let suppressClick = false;

	function beginDrag(event: PointerEvent & { currentTarget: HTMLButtonElement }) {
		if (event.button !== 0) return;
		suppressClick = false;
		drag = { id: event.pointerId, y: event.clientY };
		event.currentTarget.setPointerCapture(event.pointerId);
	}
	function moveDrag(event: PointerEvent) {
		if (drag?.id !== event.pointerId) return;
		offset = Math.max(0, event.clientY - drag.y);
		if (Math.abs(event.clientY - drag.y) > 8) suppressClick = true;
	}
	function resetDrag() {
		drag = undefined;
		offset = 0;
	}
	function finishDrag(event: PointerEvent, close: () => void) {
		if (drag?.id !== event.pointerId) return;
		const dismiss = event.clientY - drag.y >= 64;
		resetDrag();
		if (dismiss) {
			suppressClick = true;
			close();
		}
	}
</script>

<Portal>
	<Dialog.Backdrop class="fixed inset-0 z-[100] bg-black/60" />
	<Dialog.Positioner
		class="fixed inset-0 z-[100] grid items-end justify-items-center sm:place-items-center sm:p-4"
	>
		<Dialog.Content
			style={`--sheet-offset: ${offset}px`}
			class="card responsive-dialog flex max-h-[85dvh] w-full max-w-[100%] flex-col overflow-hidden bg-surface-50-950 shadow-xl sm:max-h-[90dvh] {width}"
		>
			<Dialog.Context>
				{#snippet children(dialog)}
					<button
						type="button"
						class="sheet-grab sm:hidden"
						aria-label={'Dismiss ' + title.toLowerCase()}
						onpointerdown={beginDrag}
						onpointermove={moveDrag}
						onpointerup={(event) => finishDrag(event, () => dialog().setOpen(false))}
						onpointercancel={resetDrag}
						onlostpointercapture={resetDrag}
						onclick={() => {
							if (!suppressClick) dialog().setOpen(false);
							suppressClick = false;
						}}><span aria-hidden="true"></span></button
					>
				{/snippet}
			</Dialog.Context>
			<header
				class="sticky top-0 z-10 flex shrink-0 items-center justify-between gap-3 border-b border-surface-300-700 bg-surface-50-950 px-5 pt-[max(0.25rem,env(safe-area-inset-top))] pb-3 max-sm:-mt-2 sm:pt-4"
			>
				<Dialog.Title class="h4">{title}</Dialog.Title>
				<Dialog.CloseTrigger
					class="btn-icon preset-ghost hidden sm:inline-flex"
					aria-label={'Close ' + title.toLowerCase()}
					><ControlIcon name="close" /></Dialog.CloseTrigger
				>
			</header>
			<div class="min-h-0 flex-1 overflow-x-hidden overflow-y-auto overscroll-contain px-5 py-4">
				{@render children()}
			</div>
		</Dialog.Content>
	</Dialog.Positioner>
</Portal>
