<script lang="ts">
	let {
		name,
	}: {
		name:
			| 'theme'
			| 'device'
			| 'settings'
			| 'expand'
			| 'fit'
			| 'plus'
			| 'minus'
			| 'close'
			| 'check'
			| 'reset'
			| 'find'
			| 'update';
	} = $props();
</script>

<svg
	class="size-4 shrink-0"
	viewBox="0 0 24 24"
	fill="none"
	stroke="currentColor"
	stroke-width="1.8"
	stroke-linecap="round"
	stroke-linejoin="round"
	aria-hidden="true"
>
	{#if name === 'theme'}
		<path
			d="M12 3a9 9 0 1 0 0 18h1a2 2 0 0 0 1.4-3.4 1 1 0 0 1 .7-1.6H17a4 4 0 0 0 4-4 9 9 0 0 0-9-9Z"
		/>
		<path d="M7 11h.01M9 7h.01M14 7h.01M17 10h.01" stroke-width="3" />
	{:else if name === 'device'}
		<rect x="7" y="2" width="10" height="20" rx="2" /><path d="M11 18h2" />
	{:else if name === 'settings'}
		<path
			d="M12 15.5a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7Z"
		/><path
			d="M19.4 13a7.8 7.8 0 0 0 .1-2l2-1.2-2-3.4-2.3.6a7.6 7.6 0 0 0-1.7-1L15 4h-4l-.5 2a7.6 7.6 0 0 0-1.7 1L6.5 6.4l-2 3.4 2 1.2a7.8 7.8 0 0 0 0 2l-2 1.2 2 3.4 2.3-.6a7.6 7.6 0 0 0 1.7 1L11 20h4l.5-2a7.6 7.6 0 0 0 1.7-1l2.3.6 2-3.4-2-1.2Z"
		/>
	{:else if name === 'expand'}
		<path d="M8 4H4v4m12-4h4v4M4 16v4h4m12-4v4h-4" />
	{:else if name === 'fit'}
		<rect x="3" y="5" width="18" height="14" rx="2" /><path d="M9 12h6m-3-3v6" />
	{:else if name === 'plus'}
		<path d="M5 12h14m-7-7v14" />
	{:else if name === 'minus'}
		<path d="M5 12h14" />
	{:else if name === 'reset'}
		<path d="M3 12a9 9 0 1 1 2.5 6.2M3 5v7h7" />
	{:else if name === 'close'}
		<path d="m6 6 12 12M18 6 6 18" />
	{:else if name === 'check'}
		<path d="m5 12 5 5L20 7" />
	{:else if name === 'find'}
		<circle cx="11" cy="11" r="7" /><path d="m20 20-3.5-3.5" />
	{:else if name === 'update'}
		<path d="M21 12a9 9 0 0 0-15.5-6.3M3 4v5h5" /><path d="M3 12a9 9 0 0 0 15.5 6.3M21 20v-5h-5" />
	{/if}
</svg>
