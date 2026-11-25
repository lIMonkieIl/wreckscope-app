<script lang="ts">
	import { onMount } from 'svelte';
	import RecoveryDashboard from '#lib/features/recovery/RecoveryDashboard.svelte';
	import BootSplash from '#lib/features/recovery/BootSplash.svelte';
	import { hasDesktopBridge } from '#lib/features/recovery/desktopBridge.ts';

	let showSplash = $state(true);
	let splashFading = $state(false);
	let desktopShell = $state(hasDesktopBridge());

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
		desktopShell = hasDesktopBridge();
		// Hand off from the HTML placeholder to the themed Svelte splash, then fade.
		// Never waits on update checks — those stay in the background.
		removeHtmlSplash();
		const timer = window.setTimeout(hideSplash, 320);
		return () => window.clearTimeout(timer);
	});
</script>

<div class="app-shell" class:app-shell-desktop={desktopShell}>
	{#if showSplash}
		<BootSplash fading={splashFading} />
	{/if}
	<div class="app-shell-body">
		<RecoveryDashboard />
	</div>
</div>

<style>
	.app-shell {
		display: flex;
		flex-direction: column;
		min-height: 100dvh;
		min-width: 0;
	}
	.app-shell-desktop {
		height: 100dvh;
		overflow: hidden;
	}
	.app-shell-body {
		flex: 1;
		min-height: 0;
		min-width: 0;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}
	.app-shell-body :global(.race-shell) {
		flex: 1;
		min-height: 0;
		display: flex;
		flex-direction: column;
	}
	.app-shell-body :global(.race-content) {
		flex: 1;
		min-height: 0;
		overflow: auto;
	}
</style>
