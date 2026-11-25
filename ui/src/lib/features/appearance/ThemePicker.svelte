<script lang="ts">
	import ResponsivePicker from '#lib/components/ResponsivePicker.svelte';
	import type { DisplayMode } from '#lib/features/recovery/telemetry.ts';
	import {
		clearThemeHoverPreview,
		setCommittedDocumentTheme,
		setThemeHoverPreview,
	} from './themeHover.ts';

	const themes = [
		['wreckscope', 'WreckScope'],
		['catppuccin', 'Catppuccin'],
		['cerberus', 'Cerberus'],
		['concord', 'Concord'],
		['crimson', 'Crimson'],
		['dracula', 'Dracula'],
		['fennec', 'Fennec'],
		['hamlindigo', 'Hamlindigo'],
		['legacy', 'Legacy'],
		['mint', 'Mint'],
		['modern', 'Modern'],
		['mona', 'Mona'],
		['nosh', 'Nosh'],
		['nouveau', 'Nouveau'],
		['pine', 'Pine'],
		['reign', 'Reign'],
		['rocket', 'Rocket'],
		['rose', 'Rose'],
		['rosepine', 'Rosepine'],
		['sahara', 'Sahara'],
		['seafoam', 'Seafoam'],
		['terminus', 'Terminus'],
		['vintage', 'Vintage'],
		['vox', 'Vox'],
		['wintry', 'Wintry'],
		['JiffyBagVenetian', 'Jiffy Bag'],
	] as const;

	const modeOptions = [
		{ value: 'system' as const, label: 'System' },
		{ value: 'light' as const, label: 'Light' },
		{ value: 'dark' as const, label: 'Dark' },
	];

	const themeOptions = themes.map(([value, label]) => ({
		value,
		label,
		themeSwatches: true as const,
	}));

	let {
		theme,
		mode,
		onTheme,
		onMode,
	}: {
		theme: string;
		mode: DisplayMode;
		onTheme: (theme: string) => void;
		onMode: (mode: DisplayMode) => void;
	} = $props();

	/** Local commit so close-restore doesn't snap to a stale prop. */
	let localTheme = $state(theme);
	$effect(() => {
		localTheme = theme;
	});

	function previewTheme(value: string | null) {
		setThemeHoverPreview(value);
	}

	function commitTheme(value: string) {
		localTheme = value;
		// Commit before clear so restore doesn't flash the old theme.
		setCommittedDocumentTheme(value);
		clearThemeHoverPreview();
		onTheme(value);
	}
</script>

<div class="grid gap-3 sm:grid-cols-2">
	<ResponsivePicker
		label="Mode"
		bind:value={() => mode, (value: DisplayMode) => onMode(value)}
		options={modeOptions}
	/>
	<ResponsivePicker
		label="Theme"
		bind:value={() => localTheme, commitTheme}
		options={themeOptions}
		onPreview={previewTheme}
	/>
</div>
