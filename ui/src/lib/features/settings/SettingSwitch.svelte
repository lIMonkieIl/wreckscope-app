<script lang="ts">
	import FieldHelp from '#lib/components/FieldHelp.svelte';
	import SettingMaturityBadge from '#lib/components/SettingMaturityBadge.svelte';
	import type { SettingMaturity } from '#lib/components/settingMaturity.ts';

	let {
		label,
		description,
		checked = $bindable(),
		disabled = false,
		maturity,
		onChange,
	}: {
		label: string;
		description: string;
		checked: boolean;
		disabled?: boolean;
		maturity?: SettingMaturity;
		onChange?: (value: boolean) => void;
	} = $props();
	const id = $props.id();

	function activate() {
		if (disabled) return;
		const next = !checked;
		if (!onChange) {
			checked = next;
			return;
		}
		onChange(next);
	}

	function onRowClick(event: MouseEvent) {
		const target = event.target as HTMLElement | null;
		if (target?.closest('[data-setting-help]')) return;
		activate();
	}
</script>

<!-- svelte-ignore a11y_no_static_element_interactions a11y_click_events_have_key_events -->
<div
	class="-mx-2 flex min-h-10 items-center gap-2 rounded-base px-2 transition-colors hover:bg-surface-300-700 has-[:disabled]:opacity-60 {disabled
		? 'cursor-not-allowed'
		: 'cursor-pointer'}"
	onclick={onRowClick}
>
	<span class="flex min-w-0 flex-1 items-center gap-1.5">
		<span class="truncate text-sm leading-none font-semibold">{label}</span>
		<span
			data-setting-help
			class="relative z-10 inline-flex"
			onclick={(event) => event.stopPropagation()}
			onpointerdown={(event) => event.stopPropagation()}
		>
			<FieldHelp {label} text={description} />
		</span>
		<SettingMaturityBadge {maturity} />
	</span>
	<span class="relative shrink-0" data-setting-control>
		<input
			{id}
			class="peer sr-only"
			type="checkbox"
			role="switch"
			aria-label={label}
			aria-describedby={id + '-help'}
			{checked}
			onchange={(event) => {
				const next = event.currentTarget.checked;
				if (!onChange) {
					checked = next;
					return;
				}
				event.currentTarget.checked = checked;
				onChange(next);
			}}
			{disabled}
		/>
		<span
			aria-hidden="true"
			class="block h-6 w-11 rounded-full bg-surface-400-600 peer-checked:bg-primary-600 peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-focus-visible:outline-primary-500 peer-disabled:opacity-50"
		></span>
		<span
			aria-hidden="true"
			class="absolute top-1 left-1 size-4 rounded-full bg-white transition-transform peer-checked:translate-x-5 motion-reduce:transition-none"
		></span>
	</span>
	<span id={id + '-help'} class="sr-only">{description}</span>
</div>
