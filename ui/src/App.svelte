<script lang="ts">
	import { onMount } from 'svelte';
	import RecoveryDashboard from '#lib/features/recovery/RecoveryDashboard.svelte';
	import BootSplash from '#lib/features/recovery/BootSplash.svelte';

	let showSplash = $state(true);
	let splashFading = $state(false);

	function removeHtmlSplash() {
		document.getElementById('boot-splash')?.remove();
	}

	function hideSplash() {
		splashFading = true;
		window.setTimeout(() => {
			showSplash = false;
			removeHtmlSplash();
		}, 280);
	}

	onMount(() => {
		// Hand off from the HTML placeholder to the themed Svelte splash, then fade.
		// Never waits on update checks — those stay in the background.
		removeHtmlSplash();
		const timer = window.setTimeout(hideSplash, 320);
		return () => window.clearTimeout(timer);
	});
</script>

{#if showSplash}
	<BootSplash fading={splashFading} />
{/if}
<RecoveryDashboard />
