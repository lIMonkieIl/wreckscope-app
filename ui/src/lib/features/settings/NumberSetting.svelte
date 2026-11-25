<script lang="ts">
	import FieldHelp from '#lib/components/FieldHelp.svelte';
	import SettingMaturityBadge from '#lib/components/SettingMaturityBadge.svelte';
	import type { SettingMaturity } from '#lib/components/settingMaturity.ts';
	let {
		label,
		help,
		unit,
		min,
		max,
		value = $bindable(),
		invalid = false,
		maturity,
	}: {
		label: string;
		help: string;
		unit: string;
		min: number;
		max: number;
		value: number;
		invalid?: boolean;
		maturity?: SettingMaturity;
	} = $props();
	const id = $props.id();
	let inputEl: HTMLInputElement | undefined = $state();

	function onRowClick(event: MouseEvent) {
		const target = event.target as HTMLElement | null;
		if (target?.closest('[data-setting-help]')) return;
		if (target?.closest('[data-setting-control]')) return;
		inputEl?.focus();
		inputEl?.select();
	}
</script>

<!-- svelte-ignore a11y_no_static_element_interactions a11y_click_events_have_key_events -->
<div
	class="-mx-2 min-w-0 cursor-pointer rounded-base px-2 py-1 transition-colors hover:bg-surface-300-700"
	onclick={onRowClick}
>
	<div class="mb-1 flex min-h-6 items-center gap-1.5">
		<span class="text-sm leading-none font-semibold">{label}</span>
		<span
			data-setting-help
			class="relative z-10 inline-flex"
			onclick={(event) => event.stopPropagation()}
			onpointerdown={(event) => event.stopPropagation()}
		>
			<FieldHelp {label} text={help} />
		</span>
		<SettingMaturityBadge {maturity} />
	</div>
	<div class="relative min-w-0" data-setting-control>
		<input
			bind:this={inputEl}
			{id}
			class="input unit-input w-full min-w-0 pr-12 tabular-nums"
			type="number"
			{min}
			{max}
			step="any"
			required
			aria-label={label}
			aria-invalid={invalid || undefined}
			aria-describedby={id + '-help'}
			bind:value
		/>
		<span
			aria-hidden="true"
			class="pointer-events-none absolute inset-y-0 right-0 flex w-11 items-center justify-center text-xs text-surface-600-400"
			>{unit}</span
		>
	</div>
	<p id={id + '-help'} class="sr-only">{help}</p>
</div>

<style>
	.unit-input {
		appearance: textfield;
	}
	.unit-input::-webkit-inner-spin-button,
	.unit-input::-webkit-outer-spin-button {
		appearance: none;
		margin: 0;
	}
</style>
