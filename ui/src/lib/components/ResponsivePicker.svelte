<script lang="ts" generics="Value extends string">
	import { onMount } from 'svelte';
	import { Dialog, Popover, Portal } from '@skeletonlabs/skeleton-svelte';
	import ResponsiveDialogContent from '#lib/components/ResponsiveDialogContent.svelte';
	import SettingMaturityBadge from '#lib/components/SettingMaturityBadge.svelte';
	import type { SettingMaturity } from '#lib/components/settingMaturity.ts';
	import { claimPickerOpen, releasePickerOpen } from '#lib/components/pickerOpen.ts';
	let {
		label,
		value = $bindable(),
		options,
		disabled = false,
		maturity,
		onPreview,
	}: {
		label: string;
		value: Value;
		options: readonly {
			value: Value;
			label: string;
			shortLabel?: string;
			/** Theme token swatches under the option. */
			themeSwatches?: boolean;
		}[];
		disabled?: boolean;
		maturity?: SettingMaturity;
		/** Hover preview; null restores the committed value. */
		onPreview?: (value: Value | null) => void;
	} = $props();
	let selected = $derived(options.find((option) => option.value === value));
	let desktopOpen = $state(false);
	let mobileOpen = $state(false);
	let rootEl: HTMLDivElement | undefined = $state();
	/** Skip restore-on-close after a pick so the old value doesn't flash. */
	let choseValue = false;

	function endPreview() {
		onPreview?.(null);
	}

	function finishClose() {
		if (!choseValue) endPreview();
		choseValue = false;
	}

	const closeSelf = () => {
		desktopOpen = false;
		finishClose();
	};

	function setDesktopOpen(open: boolean) {
		desktopOpen = open;
		if (open) {
			choseValue = false;
			claimPickerOpen(closeSelf);
			return;
		}
		releasePickerOpen(closeSelf);
		finishClose();
	}

	function setMobileOpen(open: boolean) {
		mobileOpen = open;
		if (open) {
			choseValue = false;
			return;
		}
		finishClose();
	}

	function choose(optionValue: Value, close: () => void) {
		choseValue = true;
		// Commit (ThemePicker) before clear so preview restore doesn't flash.
		value = optionValue;
		endPreview();
		close();
	}

	function onListLeave(event: MouseEvent | FocusEvent) {
		const next = event.relatedTarget as Node | null;
		if (next && event.currentTarget instanceof Node && event.currentTarget.contains(next)) {
			return;
		}
		endPreview();
	}

	function onRowClick(event: MouseEvent) {
		const target = event.target as HTMLElement | null;
		if (target?.closest('[data-setting-help]')) return;
		if (target?.closest('[data-setting-control]')) return;
		if (disabled) return;
		if (typeof matchMedia === 'function' && matchMedia('(min-width: 640px)').matches) {
			setDesktopOpen(true);
			return;
		}
		rootEl
			?.querySelector<HTMLButtonElement>('[data-setting-control="mobile"] button')
			?.click();
	}

	onMount(() => {
		const closeOnViewportChange = () => setDesktopOpen(false);
		window.addEventListener('resize', closeOnViewportChange);
		window.addEventListener('orientationchange', closeOnViewportChange);
		return () => {
			window.removeEventListener('resize', closeOnViewportChange);
			window.removeEventListener('orientationchange', closeOnViewportChange);
			releasePickerOpen(closeSelf);
		};
	});
</script>

{#snippet swatches()}
	<span class="flex shrink-0 -space-x-1" aria-hidden="true">
		<i class="size-2.5 rounded-full bg-primary-500"></i>
		<i class="size-2.5 rounded-full bg-secondary-500"></i>
		<i class="size-2.5 rounded-full bg-tertiary-500"></i>
		<i class="size-2.5 rounded-full bg-surface-500"></i>
	</span>
{/snippet}

{#snippet trigger()}
	<span class="flex min-w-0 flex-1 items-center gap-2">
		<span class="truncate">{selected?.shortLabel ?? selected?.label}</span>
		{#if selected?.themeSwatches}
			<span data-theme={selected.value} class="ms-auto">{@render swatches()}</span>
		{/if}
	</span>
	<svg
		class="size-4 shrink-0"
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		stroke-width="2"
		aria-hidden="true"><path d="m6 9 6 6 6-6" /></svg
	>
{/snippet}

{#snippet choices(close: () => void)}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="grid gap-1" onmouseleave={onListLeave} onfocusout={onListLeave}>
		{#each options as option (option.value)}
			<button
				type="button"
				class="flex min-h-11 cursor-pointer items-center justify-between gap-2 rounded-base px-3 text-left text-sm hover:bg-surface-300-700 aria-pressed:bg-surface-300-700"
				aria-pressed={value === option.value}
				data-theme={option.themeSwatches ? option.value : undefined}
				onmouseenter={() => {
					if (option.themeSwatches) onPreview?.(option.value);
				}}
				onfocus={() => {
					if (option.themeSwatches) onPreview?.(option.value);
				}}
				onclick={() => choose(option.value, close)}
			>
				<span class="truncate">{option.label}</span>
				{#if option.shortLabel && !option.themeSwatches}
					<strong>{option.shortLabel}</strong>
				{:else if option.themeSwatches}
					{@render swatches()}
				{/if}
			</button>
		{/each}
	</div>
{/snippet}

<!-- svelte-ignore a11y_no_static_element_interactions a11y_click_events_have_key_events -->
<div
	bind:this={rootEl}
	class="-mx-2 min-w-0 cursor-pointer rounded-base px-2 py-1 transition-colors hover:bg-surface-300-700"
	onclick={onRowClick}
>
	<span class="mb-1 flex items-center gap-1.5 text-sm leading-none font-semibold">
		<span>{label}</span>
		<SettingMaturityBadge {maturity} />
	</span>
	<div class="hidden sm:block" data-setting-control>
		<Popover
			open={desktopOpen}
			onOpenChange={({ open }) => setDesktopOpen(open)}
			positioning={{ placement: 'bottom-start', sameWidth: true }}
		>
			<Popover.Trigger
				type="button"
				class="input flex min-h-10 w-full cursor-pointer items-center justify-between gap-2 text-left hover:bg-surface-200-800"
				aria-label={label}
				{disabled}>{@render trigger()}</Popover.Trigger
			>
			<Portal>
				<Popover.Positioner class="z-[110]!">
					<Popover.Content
						class="card max-h-[min(24rem,70dvh)] overflow-y-auto overscroll-contain border border-surface-300-700 bg-surface-100-900 p-1 shadow-xl"
					>
						<Popover.Title class="sr-only">{label}</Popover.Title>
						<Popover.Context
							>{#snippet children(popover)}{@render choices(() =>
									popover().setOpen(false),
								)}{/snippet}</Popover.Context
						>
					</Popover.Content>
				</Popover.Positioner>
			</Portal>
		</Popover>
	</div>
	<div class="sm:hidden" data-setting-control="mobile">
		<Dialog
			open={mobileOpen}
			onOpenChange={({ open }) => setMobileOpen(open)}
			preventScroll
			closeOnInteractOutside
		>
			<Dialog.Trigger
				type="button"
				class="input flex min-h-10 w-full cursor-pointer items-center justify-between gap-2 text-left hover:bg-surface-200-800"
				aria-label={label}
				{disabled}>{@render trigger()}</Dialog.Trigger
			>
			<ResponsiveDialogContent title={label}>
				<Dialog.Context
					>{#snippet children(dialog)}{@render choices(() =>
							dialog().setOpen(false),
						)}{/snippet}</Dialog.Context
				>
			</ResponsiveDialogContent>
		</Dialog>
	</div>
</div>
