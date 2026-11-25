<script lang="ts">
	let { notes = '', empty = 'No release notes for this version.' }: { notes?: string; empty?: string } =
		$props();

	let text = $derived(notes.trim());
</script>

{#if text}
	<div class="changelog-notes" tabindex="0">{text}</div>
{:else}
	<p class="text-sm text-surface-600-400">{empty}</p>
{/if}

<style>
	.changelog-notes {
		margin: 0;
		white-space: pre-wrap;
		overflow-wrap: anywhere;
		font-size: 0.875rem;
		line-height: 1.45;
		color: light-dark(var(--color-surface-800), var(--color-surface-200));
	}
</style>
