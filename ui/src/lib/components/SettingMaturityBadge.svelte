<script lang="ts">
	import {
		settingMaturityHint,
		settingMaturityLabel,
		type SettingMaturity,
	} from './settingMaturity.ts';

	let { maturity }: { maturity?: SettingMaturity } = $props();

	const tone: Record<SettingMaturity, string> = {
		new: 'border-primary-500/40 text-primary-700-300',
		beta: 'border-secondary-500/40 text-secondary-700-300',
		testing: 'border-warning-500/40 text-warning-700-300',
		dev: 'border-tertiary-500/40 text-tertiary-700-300',
	};
</script>

{#if maturity}
	<span
		class="inline-flex shrink-0 items-center rounded-sm border px-1 py-px text-[0.65rem] font-semibold tracking-wide uppercase {tone[
			maturity
		]}"
		title={settingMaturityHint[maturity]}
		>{settingMaturityLabel[maturity]}</span
	>
{/if}
