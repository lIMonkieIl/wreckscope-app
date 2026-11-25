<script lang="ts">
	import { onMount } from 'svelte';
	import { Dialog, Popover, Portal } from '@skeletonlabs/skeleton-svelte';
	import ResponsiveDialogContent from '#lib/components/ResponsiveDialogContent.svelte';

	let { label, text }: { label: string; text: string } = $props();
	let desktopOpen = $state(false);

	onMount(() => {
		const closeOnViewportChange = () => {
			desktopOpen = false;
		};
		window.addEventListener('resize', closeOnViewportChange);
		window.addEventListener('orientationchange', closeOnViewportChange);
		return () => {
			window.removeEventListener('resize', closeOnViewportChange);
			window.removeEventListener('orientationchange', closeOnViewportChange);
		};
	});
</script>

{#snippet trigger()}
	<svg
		class="size-4"
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		stroke-width="1.8"
		aria-hidden="true"
		><circle cx="12" cy="12" r="9" /><path d="M12 11v5m0-8h.01" stroke-linecap="round" /></svg
	>
{/snippet}

<span class="inline-flex shrink-0 items-center self-center">
	<span class="hidden sm:inline-flex">
		<Popover
			open={desktopOpen}
			onOpenChange={({ open }) => (desktopOpen = open)}
			positioning={{ placement: 'bottom-start' }}
		>
			<Popover.Trigger
				type="button"
				class="grid size-6 shrink-0 cursor-pointer place-items-center rounded-base text-primary-700-300 hover:bg-surface-300-700"
				aria-label={'Help: ' + label}
			>
				{@render trigger()}
			</Popover.Trigger>
			<Portal>
				<!-- Above Settings dialog (z-[100]). -->
				<Popover.Positioner class="z-[110]!">
					<Popover.Content
						class="card w-[min(17rem,calc(100vw-2rem))] border border-surface-300-700 bg-surface-100-900 p-3 text-sm shadow-xl"
					>
						<Popover.Title class="sr-only">{label}</Popover.Title>
						{text}
					</Popover.Content>
				</Popover.Positioner>
			</Portal>
		</Popover>
	</span>
	<span class="sm:hidden">
		<Dialog preventScroll closeOnInteractOutside>
			<Dialog.Trigger
				type="button"
				class="grid size-6 shrink-0 cursor-pointer place-items-center rounded-base text-primary-700-300 hover:bg-surface-300-700"
				aria-label={'Help: ' + label}
			>
				{@render trigger()}
			</Dialog.Trigger>
			<ResponsiveDialogContent title={label} width="sm:max-w-sm">
				<p class="text-sm text-surface-700-300">{text}</p>
			</ResponsiveDialogContent>
		</Dialog>
	</span>
</span>
