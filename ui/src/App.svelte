<script lang="ts">
	import { onMount } from 'svelte';
	import RecoveryDashboard from '#lib/features/recovery/RecoveryDashboard.svelte';
	import BootSplash from '#lib/features/recovery/BootSplash.svelte';
	import { UPDATE_UX_DEMO } from '#lib/features/recovery/updateUxDemo.ts';

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
		const holdMs = UPDATE_UX_DEMO ? 1800 : 320;
		const timer = window.setTimeout(hideSplash, holdMs);
		return () => window.clearTimeout(timer);
	});
</script>

{#if showSplash}
	<BootSplash fading={splashFading} />
{/if}
{#if UPDATE_UX_DEMO && showSplash}
	<p class="update-ux-demo-hint" role="note">
		UPDATE_UX_DEMO on — splash holds ~1.8s; banner forced after. Flip false in updateUxDemo.ts
	</p>
{/if}
<RecoveryDashboard />

<style>
	.update-ux-demo-hint {
		position: fixed;
		top: max(0.5rem, env(safe-area-inset-top));
		left: 50%;
		z-index: 110;
		transform: translateX(-50%);
		margin: 0;
		padding: 0.35rem 0.65rem;
		border-radius: 999px;
		background: color-mix(in oklab, #000 55%, transparent);
		color: #fff;
		font-size: 0.65rem;
		letter-spacing: 0.02em;
		pointer-events: none;
		white-space: nowrap;
	}
</style>
