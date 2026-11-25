<script lang="ts">
	import { onMount } from 'svelte';
	import {
		getPhoneAccess,
		saveDetectionSettings,
		setMode,
		subscribeTelemetry,
		toggleMedic,
	} from './client.ts';
	import {
		browseTelemetryConfig,
		completeSetup,
		enableTelemetryAt,
		findTelemetryConfig,
		hasDesktopBridge,
		onOpenSettingsRequest,
		probeTelemetry,
		probeTelemetryAt,
		restoreTelemetryBackup,
	} from './telemetrySetup.ts';
	import { Dialog, Popover, Portal, QrCode } from '@skeletonlabs/skeleton-svelte';
	import ThemePicker from '#lib/features/appearance/ThemePicker.svelte';
	import { setCommittedDocumentTheme } from '#lib/features/appearance/themeHover.ts';
	import Incident from './Incident.svelte';
	import carPreviewIDs from './carPreviewIDs.json';
	import trackPreviews from './trackPreviewManifest.json';
	import { displaySpeed } from './speed';
	import DashSmoke from '#lib/components/DashSmoke.svelte';
	import BrandMark from '#lib/components/BrandMark.svelte';
	import BrandWordmark from '#lib/components/BrandWordmark.svelte';
	import ResponsiveDialogContent from '#lib/components/ResponsiveDialogContent.svelte';
	import NumberSetting from '#lib/features/settings/NumberSetting.svelte';
	import ResponsivePicker from '#lib/components/ResponsivePicker.svelte';
	import SettingSwitch from '#lib/features/settings/SettingSwitch.svelte';
	import SettingMaturityBadge from '#lib/components/SettingMaturityBadge.svelte';
	import ControlIcon from '#lib/components/ControlIcon.svelte';
	import { mountRescueMap } from './rescue-map';

	import type {
		DashboardView,
		DetectionSettings,
		DisplayMode,
		RecoveryCar,
		Standing,
		TelemetrySnapshot,
	} from './telemetry';

	type DockDrag = { id: number; start: number; size: number; span: number };
	const defaults: DetectionSettings = {
		cautionSeconds: 1,
		alertSeconds: 3.5,
		speedThreshold: 3,
		hideBots: false,
		speedUnit: 'km/h',
		muteAlertSound: false,
		drawTrackOutline: false,
		theme: 'wreckscope',
		mode: 'system',
		view: 'recovery',
	};

	function normalizeView(value?: string): DashboardView {
		return value === 'race' ? 'race' : 'recovery';
	}

	function withNormalizedSettings(settings: DetectionSettings): DetectionSettings {
		return { ...settings, view: normalizeView(settings.view) };
	}

	function settingsFormEqual(a: DetectionSettings, b: DetectionSettings) {
		return (
			a.cautionSeconds === b.cautionSeconds &&
			a.alertSeconds === b.alertSeconds &&
			a.speedThreshold === b.speedThreshold &&
			a.hideBots === b.hideBots &&
			a.speedUnit === b.speedUnit &&
			a.muteAlertSound === b.muteAlertSound &&
			Boolean(a.drawTrackOutline) === Boolean(b.drawTrackOutline) &&
			a.theme === b.theme &&
			a.mode === b.mode &&
			normalizeView(a.view) === normalizeView(b.view)
		);
	}

	let snapshot = $state<TelemetrySnapshot>({
		mode: 'live',
		connection: 'waiting',
		cars: [],
		settings: defaults,
	});
	let driverUpMap = $state(true);
	let elevatedMap = $state(false);
	let dismissed = $state(new Set<string>());
	let mapRoot: HTMLDivElement;
	let map: ReturnType<typeof mountRescueMap> | undefined;
	let compactLayout = $state(false),
		leaderboardOpen = $state(false),
		raceDetailsOpen = $state(false);
	let landscape = $state(false),
		portraitDock = $state(29),
		landscapeDock = $state(30);
	let dockDrag: DockDrag | null = null;
	let dockSize = $derived(landscape ? landscapeDock : portraitDock);
	function resizeDock(value: number) {
		const size = Math.round(Math.max(18, Math.min(55, value)));
		if (landscape) landscapeDock = size;
		else portraitDock = size;
	}
	function startDockDrag(event: PointerEvent & { currentTarget: HTMLDivElement }) {
		if (event.button !== 0) return;
		const handle = event.currentTarget;
		const bounds = handle.parentElement?.getBoundingClientRect();
		if (!bounds) return;
		dockDrag = {
			id: event.pointerId,
			start: landscape ? event.clientX : event.clientY,
			size: dockSize,
			span: landscape ? bounds.width : bounds.height,
		};
		handle.setPointerCapture(event.pointerId);
	}
	function dragDock(event: PointerEvent) {
		if (dockDrag?.id !== event.pointerId) return;
		const position = landscape ? event.clientX : event.clientY;
		resizeDock(dockDrag.size + ((dockDrag.start - position) * 100) / dockDrag.span);
	}
	function resizeDockWithKey(event: KeyboardEvent) {
		const more = landscape ? 'ArrowLeft' : 'ArrowUp',
			less = landscape ? 'ArrowRight' : 'ArrowDown';
		if (![more, less, 'Home', 'End'].includes(event.key)) return;
		event.preventDefault();
		resizeDock(
			event.key === 'Home'
				? 18
				: event.key === 'End'
					? 55
					: dockSize + (event.key === more ? 3 : -3),
		);
	}
	let settingsOpen = $state(false),
		phoneOpen = $state(false);
	let streamReady = $state(false);
	let speedUnit = $derived(snapshot.settings.speedUnit ?? 'km/h');
	// Phone: veil while desktop setup unfinished or stream down.
	let phoneVeil = $derived(
		!hasDesktopBridge() &&
			(snapshot.connection === 'offline' ||
				(streamReady && !snapshot.settings.setupCompleted)),
	);
	let veilMessage = $derived(
		snapshot.connection === 'offline' ? 'Disconnected' : 'Awaiting setup',
	);
	const previews = new Set(carPreviewIDs);
	function carPreview(id?: string) {
		const model = id?.split(':')[0];
		return model && previews.has(model) ? '/car-previews/' + model + '_default.png' : '';
	}
	let notice = $state(''),
		qrURL = $state(''),
		qrStatus = $state('');
	let configPath = $state('');
	let configNotice = $state('');
	let configHasBackup = $state(false);
	let configBusy = $state(false);
	let configReady = $state(false);
	let configValid = $state(false);
	let configReadyPath = $state('');
	let configProbed = $state(false);
	let configLooksReady = $derived(
		configReady && configValid && configPath.trim() !== '' && configPath.trim() === configReadyPath.trim(),
	);
	let configShowConfigure = $derived(configValid && !configLooksReady);
	let configPathSelected = $derived(configPath.trim() !== '');
	let configPathTone = $derived(
		!configPathSelected ? 'none' : configValid ? 'ok' : 'bad',
	);
	/** Show Find when empty or the pick failed validation. */
	let configShowFind = $derived(!configPathSelected || !configValid);
	let settingsPersistTimer: ReturnType<typeof setTimeout> | undefined;
	let settingsSaveGen = 0;
	/** Optimistic settings until the stream confirms. */
	let awaitedSettings: DetectionSettings | null = null;
	/** Optimistic live/demo until the stream confirms. */
	let awaitedFeedMode: 'live' | 'demo' | null = null;
	let liveSettings = $derived(withNormalizedSettings(snapshot.settings));
	/** Desktop first-run UX while setupCompleted is false (independent of stream). */
	let firstRunSetup = $derived(hasDesktopBridge() && !liveSettings.setupCompleted);
	/** Desktop first-run: keep Settings open until setupCompleted. */
	let setupIncomplete = $derived(firstRunSetup && streamReady);
	/** Desktop: bad/missing Config → open Settings and highlight. */
	let configNeedsAttention = $derived(
		hasDesktopBridge() && streamReady && configProbed && !configLooksReady,
	);
	/** First-run or broken Config: emphasize the Config section. */
	let configHighlighted = $derived(firstRunSetup || configNeedsAttention);
	/** Soften other Settings sections until telemetry Config is ready on first run. */
	let settingsSecondary = $derived(firstRunSetup && !configLooksReady);
	let settingsDialogTitle = $derived(firstRunSetup ? 'Set up WreckScope' : 'Settings');
	/** Dim dash until Config is ready; header stays usable. */
	let configSmoke = $derived(configNeedsAttention);
	let dashSmoked = $derived(phoneVeil || configSmoke);
	let configSectionEl: HTMLElement | undefined = $state();
	let raceView = $derived(normalizeView(liveSettings.view) === 'race');
	let systemDark = $state(
		typeof matchMedia === 'function' ? matchMedia('(prefers-color-scheme: dark)').matches : true,
	);
	$effect(() => {
		const theme = liveSettings.theme || defaults.theme;
		const mode = (liveSettings.mode || defaults.mode) as DisplayMode;
		const dark = mode === 'dark' || (mode === 'system' && systemDark);
		setCommittedDocumentTheme(theme);
		document.documentElement.dataset.mode = dark ? 'dark' : 'light';
		document.documentElement.classList.toggle('dark', dark);
	});
	let lastAlerts = new Set<string | number>();
	let audio: AudioContext | undefined;
	const stateLabels = {
		moving: 'Moving',
		slow: 'Stopped',
		caution: 'Possible stuck',
		alert: 'Needs help',
		out: 'Out of race',
	};
	let visibleCars = $derived(snapshot.cars.filter((c) => c.health !== 0 && c.status === 2));
	let incidents = $derived(
		visibleCars
			.filter(
				(c) =>
					c.slot !== snapshot.localSlot &&
					(!c.incidentId || !dismissed.has(c.incidentId)) &&
					['alert', 'caution'].includes(c.state),
			)
			.sort(
				(a, b) =>
					Number(b.state === 'alert') - Number(a.state === 'alert') || b.stopped - a.stopped,
			),
	);
	let alerts = $derived(incidents.filter((c) => c.state === 'alert').length);
	let race = $derived(snapshot.race);
	let carsBySlot = $derived(new Map(snapshot.cars.map((car) => [car.slot, car])));
	let standings = $derived<Standing[]>(snapshot.standings ?? snapshot.cars);
	let followedSlot = $state<number | null>(null);
	let medicSlots = $derived(new Set(snapshot.medics ?? []));
	function toggleFollow(slot: number) {
		followedSlot = followedSlot === slot ? null : slot;
		map?.follow(followedSlot);
	}
	async function markMedic(slot: number) {
		try {
			const medics = await toggleMedic(slot);
			snapshot = { ...snapshot, medics };
			map?.update(snapshot, dismissed, speedUnit);
		} catch (error) {
			console.warn('Medic toggle', error);
		}
	}
	let currentLap = $derived(standings.reduce((lap, car) => Math.max(lap, car.lap), 0));
	let fastestDriver = $derived(
		standings.reduce<Standing | undefined>(
			(best, car) =>
				car.best != null && (best?.best == null || car.best < best.best) ? car : best,
			undefined,
		),
	);
	const driverStatuses = [
		'—',
		'Lobby',
		'Racing',
		'Finished',
		'Eliminated',
		'Disqualified',
		'Retired',
		'Timed out',
		'Wrecked',
	];
	const trackStatuses = ['Normal', 'Out of sectors', 'Wrong sector', 'Wrong direction'];
	let trackImage = $derived.by(() => {
		const file = race?.trackId + '.png';
		return Object.hasOwn(trackPreviews, file) ? '/track-previews/' + file : '';
	});
	const modes = ['Banger Race', 'Deathmatch', 'Last Man Standing'];
	const damageModes = ['Wrecker', 'Normal', 'Realistic'];
	const marshalFlags = [
		[1, 'Green'],
		[2, 'Last lap'],
		[4, 'Finish'],
		[8, 'Disqualified'],
		[16, 'Meatball'],
		[32, 'Warning'],
		[64, 'Blue'],
		[128, 'White'],
		[16384, 'Countdown'],
		[32768, 'Countdown'],
	] as const;
	let flags = $derived(
		race
			? [
					...new Set(
						marshalFlags.filter(([bit]) => (race?.flags ?? 0) & bit).map(([, label]) => label),
					),
				].join(' · ') || (race.flags === 0 ? 'Clear' : `0x${race.flags.toString(16)}`)
			: '—',
	);
	let raceState = $derived(
		snapshot.mode === 'demo'
			? 'Demo'
			: snapshot.connection !== 'live'
				? 'Waiting'
				: !race
					? 'Waiting for race data'
					: race.flags & 4
						? 'Finished'
						: race.flags & 49152
							? 'Countdown'
							: snapshot.sessionRacing
								? 'Racing'
								: race.clock <= 0
									? 'Lobby'
									: 'Race clock stopped',
	);
	function clock(ms?: number | null) {
		if (ms == null || ms < 0) return '—';
		const minutes = Math.floor(ms / 60000);
		return (
			String(minutes).padStart(2, '0') + ':' + ((ms % 60000) / 1000).toFixed(3).padStart(6, '0')
		);
	}
	function gap(ms?: number) {
		return ms == null ? '—' : '+' + (ms / 1000).toFixed(3) + 's';
	}
	let raceRows = $derived([
		['Track name', race?.track || '—'],
		['Track ID', race?.trackId || '—'],
		['Track distance', race?.distance != null ? race.distance.toFixed(1) + ' m' : '—'],
		['Laps', (race?.laps ?? 0) > 0 ? race?.laps + ' laps' : '—'],
		['Mode', modes[race?.mode ?? -1] ?? '—'],
		['Damage', damageModes[race?.damage ?? -1] ?? '—'],
		['Timer', race ? (Math.max(0, race.clock) / 1000).toFixed(3) + 's' : '—'],
		[
			'Best lap',
			fastestDriver
				? clock(fastestDriver.best) + (fastestDriver.name ? ' (' + fastestDriver.name + ')' : '')
				: '—',
		],
		['Grid', race?.grid ? `${race.remaining} / ${race.grid} cars` : '—'],
		['Race status', raceState],
		['Flags', flags],
	]);
	let mapStatus = $derived(
		snapshot.connection === 'live'
			? ['Game connected', 'Waiting for live car positions. Enter a race to see cars.']
			: snapshot.connection === 'stale'
				? ['Telemetry stopped', 'Check that the game is still running.']
				: snapshot.connection === 'offline'
					? ['Dashboard disconnected', 'Waiting to reconnect to WreckScope.']
					: snapshot.mode === 'demo'
						? ['Starting demo', 'Waiting for simulated cars.']
						: ['Waiting for Wreckfest 2', 'Start the game and enter a race.'],
	);

	type IncidentRecord = {
		id: string;
		driver: string;
		started: string;
		state: 'caution' | 'alert';
		score?: number;
		factors?: RecoveryCar['scoreFactors'];
		reason: string;
		outcome: 'Active' | 'Cleared' | 'Session ended' | 'Unrecoverable' | 'False alarm';
	};
	let incidentHistory = $state<IncidentRecord[]>([]);

	function recordIncidents(next: TelemetrySnapshot, sessionChanged: boolean) {
		const active = new Set<string>();
		for (const car of next.cars) {
			const id = car.incidentId;
			if (!id || car.slot === next.localSlot || (car.state !== 'caution' && car.state !== 'alert'))
				continue;
			active.add(id);
			const existing = incidentHistory.find((entry) => entry.id === id);
			if (
				existing?.state === 'alert' ||
				existing?.state === car.state ||
				existing?.outcome === 'False alarm'
			)
				continue;
			const entry: IncidentRecord = {
				id,
				driver: car.name || 'Car ' + (car.slot + 1),
				started: car.incidentSince ?? car.updated,
				state: car.state,
				score: car.score,
				factors: car.scoreFactors,
				reason: car.reason,
				outcome: existing?.outcome ?? 'Active',
			};
			if (existing) Object.assign(existing, entry);
			else incidentHistory = [entry, ...incidentHistory].slice(0, 100);
		}
		for (const entry of incidentHistory) {
			if (entry.outcome === 'Active' && !active.has(entry.id)) {
				entry.outcome = sessionChanged ? 'Session ended' : 'Cleared';
			}
		}
	}

	function markFalseAlarm(entry: IncidentRecord) {
		if (snapshot.cars.some((car) => car.incidentId === entry.id)) dismiss(entry.id);
		entry.outcome = 'False alarm';
	}

	function receive(next: TelemetrySnapshot) {
		streamReady = true;
		const previousMode = snapshot.mode;
		const previousEpoch = snapshot.raceEpoch;
		let feedMode = next.mode;
		if (awaitedFeedMode) {
			if (next.mode === awaitedFeedMode) awaitedFeedMode = null;
			else feedMode = awaitedFeedMode;
		}
		if (previousMode !== feedMode || snapshot.raceEpoch !== next.raceEpoch) dismissed = new Set();
		else {
			const active = new Set(next.cars.map((c) => c.incidentId));
			if ([...dismissed].some((id) => !active.has(id)))
				dismissed = new Set([...dismissed].filter((id) => active.has(id)));
		}
		recordIncidents(next, previousMode !== feedMode || next.raceEpoch !== previousEpoch);
		const incoming = withNormalizedSettings(next.settings);
		// Hold optimistic settings until the stream confirms.
		let settings = incoming;
		if (awaitedSettings) {
			if (settingsFormEqual(incoming, awaitedSettings)) awaitedSettings = null;
			else settings = awaitedSettings;
		}
		snapshot = { ...next, mode: feedMode, settings };

		if (previousMode !== feedMode || next.raceEpoch !== previousEpoch) {
			followedSlot = null;
		}
		map?.update({ ...next, mode: feedMode, settings }, dismissed, speedUnit);
		const ids = new Set(
			incidents.filter((c) => c.state === 'alert').map((c) => c.incidentId || c.slot),
		);
		if (
			!next.settings.muteAlertSound &&
			audio?.state === 'running' &&
			[...ids].some((id) => !lastAlerts.has(id))
		) {
			const osc = audio.createOscillator(),
				gain = audio.createGain();
			osc.frequency.value = 720;
			gain.gain.value = 0.08;
			osc.connect(gain).connect(audio.destination);
			osc.start();
			osc.stop(audio.currentTime + 0.22);
		}
		lastAlerts = ids;
	}
	onMount(() => {
		const compact = matchMedia('(max-width: 65rem)');
		const syncLayout = () => {
			compactLayout = compact.matches;
		};
		syncLayout();
		const orientation = matchMedia('(orientation: landscape)');
		const syncOrientation = () => {
			landscape = orientation.matches;
			dockDrag = null;
		};
		syncOrientation();
		orientation.addEventListener('change', syncOrientation);
		compact.addEventListener('change', syncLayout);
		const colorScheme = matchMedia('(prefers-color-scheme: dark)');
		const syncColorScheme = () => {
			systemDark = colorScheme.matches;
		};
		syncColorScheme();
		colorScheme.addEventListener('change', syncColorScheme);
		map = mountRescueMap(mapRoot, () => (followedSlot = null));
		const unsubscribe = subscribeTelemetry(
			receive,
			() => {
				snapshot = { ...snapshot, connection: 'offline', cars: [] };
				map?.update(snapshot, dismissed, speedUnit);
			},
			() => {
				// Phone SSE: full reload on reconnect so we don't resume stale UI.
				if (hasDesktopBridge()) return;
				location.reload();
			},
		);
		function unlock() {
			try {
				const Audio =
					window.AudioContext ||
					(window as Window & { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
				if (!Audio) return;
				audio ??= new Audio();
				if (audio.state === 'suspended') audio.resume().catch(() => {});
			} catch {}
		}
		document.addEventListener('click', unlock);
		const stopOpenSettings = onOpenSettingsRequest(() => {
			void editSettings(true);
		});
		return () => {
			compact.removeEventListener('change', syncLayout);
			orientation.removeEventListener('change', syncOrientation);
			colorScheme.removeEventListener('change', syncColorScheme);
			stopOpenSettings();
			unsubscribe();
			map?.destroy();
			document.removeEventListener('click', unlock);
			audio?.close().catch(() => {});
			if (settingsPersistTimer) clearTimeout(settingsPersistTimer);
		};
	});
	function dismiss(id?: string) {
		if (!id) return;
		dismissed = new Set([...dismissed, id]);
		const entry = incidentHistory.find((item) => item.id === id);
		if (entry) entry.outcome = 'Unrecoverable';
		map?.update(snapshot, dismissed, speedUnit);
	}
	function undo() {
		const id = [...dismissed].at(-1);
		const entry = incidentHistory.find((item) => item.id === id);
		if (entry) entry.outcome = 'Active';
		dismissed = new Set([...dismissed].slice(0, -1));
		map?.update(snapshot, dismissed, speedUnit);
	}
	let demoPending = $state(false);
	async function toggleDemo(enabled: boolean) {
		const mode = enabled ? 'demo' : 'live';
		awaitedFeedMode = mode;
		snapshot = { ...snapshot, mode };
		demoPending = true;
		notice = '';
		try {
			await setMode(mode);
			window.setTimeout(() => {
				if (awaitedFeedMode === mode) awaitedFeedMode = null;
			}, 2500);
		} catch (error) {
			awaitedFeedMode = null;
			notice = error instanceof Error ? error.message : String(error);
		} finally {
			demoPending = false;
		}
	}
	async function markSetupComplete() {
		if (!hasDesktopBridge() || snapshot.settings.setupCompleted) return;
		try {
			await completeSetup();
			snapshot = {
				...snapshot,
				settings: withNormalizedSettings({
					...snapshot.settings,
					setupCompleted: true,
				}),
			};
		} catch (error) {
			notice = error instanceof Error ? error.message : String(error);
		}
	}

	/** Path hint text only; ok/bad shows on the input border. */
	function configInstructionNotice(message: string): string {
		const text = message.trim();
		if (!text) return '';
		const keep = [
			'no telemetry config found',
			'multiple game profiles found',
			'Could not find Wreckfest 2 telemetry config',
			'Found a file that is not a valid telemetry config',
		];
		const lower = text.toLowerCase();
		return keep.some((hint) => lower.includes(hint.toLowerCase())) ? text : '';
	}

	function applyConfigStatus(
		status: {
			path: string;
			hasBackup: boolean;
			ready: boolean;
			valid: boolean;
			error: string;
		},
		fallbackPath = '',
	) {
		configPath = status.path || fallbackPath;
		configHasBackup = !!status.hasBackup;
		configValid = !!status.valid;
		configReady = !!status.ready;
		configReadyPath = status.ready && status.valid ? configPath.trim() : '';
		configNotice = configInstructionNotice(status.error || '');
	}

	async function loadConfigSection() {
		if (!hasDesktopBridge()) return;
		try {
			const status = await probeTelemetry();
			applyConfigStatus(status, snapshot.settings.telemetryConfigPath || '');
			// Do not mark setupCompleted here — first-run UX stays until the user closes Settings.
		} catch (error) {
			configValid = false;
			configReady = false;
			configReadyPath = '';
			configNotice = configInstructionNotice(
				error instanceof Error ? error.message : String(error),
			);
		} finally {
			configProbed = true;
		}
	}

	async function editSettings(open: boolean) {
		if (!open) {
			// User finished / dismissed first-run (or normal Settings).
			if (hasDesktopBridge() && !snapshot.settings.setupCompleted) {
				await markSetupComplete();
			}
			settingsOpen = false;
			return;
		}
		settingsOpen = true;
		notice = '';
		configNotice = '';
		configPath = snapshot.settings.telemetryConfigPath || configPath;
		void loadConfigSection();
	}

	$effect(() => {
		if (!hasDesktopBridge() || !streamReady) return;
		const needsSettings = setupIncomplete || configNeedsAttention;
		if (needsSettings && !settingsOpen) {
			settingsOpen = true;
			queueMicrotask(() => {
				configSectionEl?.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
			});
		}
		void loadConfigSection();
	});
	let appearanceChanged = $derived(
		(liveSettings.theme || defaults.theme) !== defaults.theme ||
			(liveSettings.mode || defaults.mode) !== defaults.mode,
	);
	let dashboardChanged = $derived(
		liveSettings.speedUnit !== defaults.speedUnit ||
			Boolean(liveSettings.drawTrackOutline) !== Boolean(defaults.drawTrackOutline) ||
			normalizeView(liveSettings.view) !== normalizeView(defaults.view),
	);
	let detectionChanged = $derived(
		liveSettings.cautionSeconds !== defaults.cautionSeconds ||
			liveSettings.alertSeconds !== defaults.alertSeconds ||
			liveSettings.speedThreshold !== defaults.speedThreshold ||
			liveSettings.hideBots !== defaults.hideBots ||
			liveSettings.muteAlertSound !== defaults.muteAlertSound,
	);

	/** Optimistic patch + SaveSettings; stream confirms. */
	function patchSettings(patch: Partial<DetectionSettings>) {
		const next = withNormalizedSettings({ ...snapshot.settings, ...patch });
		if (next.alertSeconds < next.cautionSeconds + 0.5) {
			snapshot = { ...snapshot, settings: next };
			return;
		}
		const gen = ++settingsSaveGen;
		awaitedSettings = next;
		snapshot = { ...snapshot, settings: next };
		notice = '';
		if (settingsPersistTimer) clearTimeout(settingsPersistTimer);
		settingsPersistTimer = setTimeout(() => {
			settingsPersistTimer = undefined;
			const payload = withNormalizedSettings(snapshot.settings);
			void saveDetectionSettings(payload)
				.then((saved) => {
					if (gen !== settingsSaveGen) return;
					const confirmed = withNormalizedSettings(saved ?? payload);
					awaitedSettings = confirmed;
					snapshot = { ...snapshot, settings: confirmed };
					window.setTimeout(() => {
						if (gen === settingsSaveGen && awaitedSettings && settingsFormEqual(awaitedSettings, confirmed)) {
							awaitedSettings = null;
						}
					}, 2500);
				})
				.catch((error) => {
					if (gen !== settingsSaveGen) return;
					awaitedSettings = null;
					notice = error instanceof Error ? error.message : String(error);
				});
		}, 80);
	}

	function resetAppearanceSettings() {
		patchSettings({
			theme: defaults.theme,
			mode: defaults.mode,
		});
	}
	function resetDashboardSettings() {
		patchSettings({
			speedUnit: defaults.speedUnit,
			drawTrackOutline: false,
			view: defaults.view,
		});
	}
	function resetDetectionSettings() {
		patchSettings({
			cautionSeconds: defaults.cautionSeconds,
			alertSeconds: defaults.alertSeconds,
			speedThreshold: defaults.speedThreshold,
			hideBots: defaults.hideBots,
			muteAlertSound: defaults.muteAlertSound,
		});
	}
	async function browseConfigFile() {
		if (!hasDesktopBridge() || configBusy) return;
		configBusy = true;
		configNotice = '';
		try {
			const picked = await browseTelemetryConfig();
			if (!picked) return;
			const status = await probeTelemetryAt(picked);
			applyConfigStatus(status, picked);
		} catch (error) {
			configValid = false;
			configReady = false;
			configReadyPath = '';
			configNotice = configInstructionNotice(
				error instanceof Error ? error.message : String(error),
			);
		} finally {
			configBusy = false;
		}
	}

	async function findConfigFile() {
		if (!hasDesktopBridge() || configBusy) return;
		configBusy = true;
		configNotice = '';
		try {
			const status = await findTelemetryConfig();
			applyConfigStatus(status);
			if (!status.path) {
				configNotice = configInstructionNotice(
					status.error || 'Could not find Wreckfest 2 telemetry config.json',
				);
			} else if (!status.valid) {
				configNotice =
					configInstructionNotice(status.error) ||
					'Found a file that is not a valid telemetry config.';
			}
		} catch (error) {
			configValid = false;
			configReady = false;
			configReadyPath = '';
			configNotice = configInstructionNotice(
				error instanceof Error ? error.message : String(error),
			);
		} finally {
			configBusy = false;
		}
	}

	async function configureTelemetryFile() {
		if (!hasDesktopBridge() || !configValid) return;
		const path = configPath.trim();
		if (!path) return;
		configBusy = true;
		configNotice = '';
		try {
			const status = await enableTelemetryAt(path);
			applyConfigStatus(status, path);
			// Keep first-run dialog until the user closes it.
		} catch (error) {
			configValid = false;
			configReady = false;
			configReadyPath = '';
			configNotice = configInstructionNotice(
				error instanceof Error ? error.message : String(error),
			);
		} finally {
			configBusy = false;
		}
	}

	async function restoreTelemetryFromBackup() {
		if (!hasDesktopBridge()) return;
		configBusy = true;
		configNotice = '';
		try {
			const status = await restoreTelemetryBackup(configPath.trim());
			applyConfigStatus(status, configPath);
		} catch (error) {
			configValid = false;
			configReady = false;
			configReadyPath = '';
			configNotice = configInstructionNotice(
				error instanceof Error ? error.message : String(error),
			);
		} finally {
			configBusy = false;
		}
	}
	async function showPhone(open: boolean) {
		phoneOpen = open;
		if (!open) return;
		qrURL = '';
		qrStatus = 'Loading…';
		try {
			const access = await getPhoneAccess();
			const address = access.urls?.[0];
			if (address) {
				qrURL = new URL(address).origin;
				return;
			}
			qrStatus = access.lan ? 'No primary private IPv4 address found.' : 'Phone access is off.';
		} catch {
			qrStatus = 'Could not load the phone link.';
		}
	}
</script>

<svelte:head><title>WreckScope</title></svelte:head>
<div class="race-shell" class:is-smoked={dashSmoked} bind:this={mapRoot}>
	<header class="race-header">
		<div class="brandline">
			<BrandMark class="brand-bull" size={44} />
			<div class="brand-copy">
				<BrandWordmark />
			</div>
		</div>
		<div class="header-controls">
			<Dialog
				preventScroll
				closeOnInteractOutside
				open={phoneOpen}
				onOpenChange={({ open }) => showPhone(open)}
			>
				<Dialog.Trigger
					class="btn-icon preset-ghost"
					aria-label="Device view"
					title="Device view"><ControlIcon name="device" /></Dialog.Trigger
				>
				<ResponsiveDialogContent title="Device view" width="sm:max-w-sm">
					{#if qrURL}
						<div class="grid justify-items-center gap-3">
							<QrCode value={qrURL} encoding={{ ecc: 'H' }} class="grid w-full place-items-center">
								<QrCode.Frame class="device-qr" aria-label="Phone dashboard QR code">
									<QrCode.Pattern fill="#090909" />
								</QrCode.Frame>
								<QrCode.Overlay
									class="grid size-8 place-items-center rounded-full border-2 border-white bg-surface-950"
								>
									<BrandMark size={22} />
								</QrCode.Overlay>
							</QrCode>
						</div>
					{:else}<p role="status">{qrStatus}</p>{/if}
				</ResponsiveDialogContent>
			</Dialog>
			<Dialog
				preventScroll
				closeOnInteractOutside
				open={settingsOpen}
				onOpenChange={({ open }) => editSettings(open)}
			>
				<Dialog.Trigger class="btn-icon preset-ghost" aria-label="Settings" title="Settings"
					><ControlIcon name="settings" /></Dialog.Trigger
				>
				<ResponsiveDialogContent
					title={settingsDialogTitle}
					width="sm:max-w-3xl lg:max-w-5xl xl:max-w-6xl"
				>
					<div class="@container min-w-0 w-full">
					{#if firstRunSetup}
						<p class="mb-5 text-sm text-surface-700-300">
							Connect Wreckfest 2 telemetry so the live feed can start. Find or choose the game’s
							<code class="font-mono text-xs">config.json</code>, then Configure if needed.
						</p>
					{/if}
					<div class="grid min-w-0 gap-6 @[48rem]:grid-cols-2 @[48rem]:gap-8">
						<div class="min-w-0 space-y-6">
							{#if hasDesktopBridge()}
								<section
									bind:this={configSectionEl}
									class="config-section min-w-0 space-y-3"
									class:opacity-60={configBusy}
									class:needs-attention={configHighlighted}
								>
									<header
										class="flex items-center gap-2 border-b border-surface-300-700/60 pb-1.5 text-sm font-semibold"
									>
										{firstRunSetup ? 'Connect telemetry' : 'Config'}
									</header>
									<div class="min-w-0" inert={configBusy || undefined}>
										<span class="mb-1 flex items-center gap-1.5 text-sm font-semibold">
											{firstRunSetup
												? 'Wreckfest 2 telemetry config.json'
												: 'Telemetry config file'}
											{#if !firstRunSetup}
												<SettingMaturityBadge maturity="new" />
											{/if}
											{#if configPathSelected}
												<span
													class="inline-flex size-5 items-center justify-center {configLooksReady
														? 'text-success-700-300'
														: 'text-error-700-300'}"
													title={configLooksReady
														? 'Configured correctly'
														: 'Not configured correctly'}
													aria-label={configLooksReady
														? 'Configured correctly'
														: 'Not configured correctly'}
												>
													<ControlIcon name={configLooksReady ? 'check' : 'close'} />
												</span>
											{/if}
										</span>
										<div class="flex min-w-0 flex-wrap items-center gap-2">
											<input
												class="input config-path-input min-h-10 min-w-0 flex-1 cursor-pointer font-mono text-xs"
												class:is-ok={configPathTone === 'ok'}
												class:is-bad={configPathTone === 'bad' ||
													(configNeedsAttention && !configPathSelected)}
												type="text"
												readonly
												spellcheck="false"
												value={configPath}
												placeholder="Path to game config.json"
												disabled={configBusy}
												aria-invalid={configPathTone === 'bad' ||
												(configNeedsAttention && !configPathSelected)
													? 'true'
													: undefined}
												aria-label="Telemetry config file, click to choose"
												title="Click to choose config file"
												onclick={() => void browseConfigFile()}
												onkeydown={(event) => {
													if (event.key === 'Enter' || event.key === ' ') {
														event.preventDefault();
														void browseConfigFile();
													}
												}}
											/>
											{#if configShowConfigure}
												<button
													type="button"
													class="btn preset-filled-primary-500 config-action-attention min-h-10 shrink-0"
													disabled={configBusy}
													onclick={configureTelemetryFile}
													>{configBusy ? 'Working…' : 'Configure'}</button
												>
											{/if}
											{#if configShowFind}
												<button
													type="button"
													class="btn-icon preset-filled-primary-500 config-action-attention shrink-0"
													disabled={configBusy}
													aria-label="Find Wreckfest 2 telemetry config"
													title="Find Wreckfest 2 telemetry config"
													onclick={() => void findConfigFile()}
												>
													<ControlIcon name="find" />
												</button>
											{/if}
											{#if configHasBackup}
												<button
													type="button"
													class="btn-icon preset-tonal-error shrink-0"
													disabled={configBusy}
													aria-label="Restore config_backup.wreckscope.json"
													title="Restore config_backup.wreckscope.json"
													onclick={restoreTelemetryFromBackup}
												>
													<ControlIcon name="reset" />
												</button>
											{/if}
										</div>
									</div>
									{#if configNotice}
										<p role="status" class="text-sm text-surface-700-300">{configNotice}</p>
									{/if}
								</section>
							{/if}

							<section
								class="min-w-0 space-y-3"
								class:settings-secondary={settingsSecondary}
							>
								<header
									class="flex items-center gap-2 border-b border-surface-300-700/60 pb-1.5 text-sm font-semibold"
								>
									Appearance
									{#if appearanceChanged}
										<button
											type="button"
											class="btn-icon preset-ghost size-7"
											aria-label="Restore default appearance"
											title="Restore defaults"
											onclick={resetAppearanceSettings}
										>
											<ControlIcon name="reset" />
										</button>
									{/if}
								</header>
								<ThemePicker
									theme={liveSettings.theme || defaults.theme}
									mode={(liveSettings.mode || defaults.mode) as DisplayMode}
									onTheme={(value) => patchSettings({ theme: value })}
									onMode={(value) => patchSettings({ mode: value })}
								/>
							</section>

							<section
								class="min-w-0 space-y-3"
								class:settings-secondary={settingsSecondary}
							>
								<header
									class="flex items-center gap-2 border-b border-surface-300-700/60 pb-1.5 text-sm font-semibold"
								>
									Dashboard
									{#if dashboardChanged}
										<button
											type="button"
											class="btn-icon preset-ghost size-7"
											aria-label="Restore default dashboard settings"
											title="Restore defaults"
											onclick={resetDashboardSettings}
										>
											<ControlIcon name="reset" />
										</button>
									{/if}
								</header>
								<div class="grid gap-3 sm:grid-cols-2">
									<ResponsivePicker
										label="View"
										bind:value={
											() => normalizeView(liveSettings.view),
											(v: DashboardView) => patchSettings({ view: v })
										}
										options={[
											{ value: 'race', label: 'Race view' },
											{ value: 'recovery', label: 'Recovery control' },
										]}
									/>
									<ResponsivePicker
										label="Display speed in"
										bind:value={
											() => liveSettings.speedUnit,
											(v: DetectionSettings['speedUnit']) => patchSettings({ speedUnit: v })
										}
										options={[
											{ value: 'km/h', label: 'Kilometres per hour', shortLabel: 'km/h' },
											{ value: 'mph', label: 'Miles per hour', shortLabel: 'mph' },
										]}
									/>
								</div>
								<div class="space-y-1">
									<SettingSwitch
										label="Draw track outline"
										maturity="testing"
										description="Learned track outline on the map."
										checked={!!liveSettings.drawTrackOutline}
										onChange={(v) => patchSettings({ drawTrackOutline: v })}
									/>
									{#if import.meta.env.DEV}
										<SettingSwitch
											label="Demo mode"
											maturity="dev"
											description="Simulated cars. Live feed pauses while demo is on."
											checked={snapshot.mode === 'demo'}
											onChange={toggleDemo}
											disabled={demoPending}
										/>
									{/if}
								</div>
							</section>
						</div>

						<div class="min-w-0 space-y-6">
							<section
								class="@container min-w-0 space-y-3"
								class:settings-secondary={settingsSecondary}
							>
								<header
									class="flex items-center gap-2 border-b border-surface-300-700/60 pb-1.5 text-sm font-semibold"
								>
									Alerts &amp; detection
									{#if detectionChanged}
										<button
											type="button"
											class="btn-icon preset-ghost size-7"
											aria-label="Restore default detection settings"
											title="Restore defaults"
											onclick={resetDetectionSettings}
										>
											<ControlIcon name="reset" />
										</button>
									{/if}
								</header>
								<div class="grid grid-cols-1 gap-3 @[36rem]:grid-cols-3">
									<NumberSetting
										label="Caution after"
										unit="sec"
										min={0.5}
										max={60}
										bind:value={
											() => liveSettings.cautionSeconds,
											(v: number) => patchSettings({ cautionSeconds: v })
										}
										help="Seconds before a possible stuck caution."
									/>
									<NumberSetting
										label="Recovery alert after"
										unit="sec"
										min={(liveSettings.cautionSeconds ?? 0.5) + 0.5}
										max={180}
										bind:value={
											() => liveSettings.alertSeconds,
											(v: number) => patchSettings({ alertSeconds: v })
										}
										invalid={liveSettings.alertSeconds < liveSettings.cautionSeconds + 0.5}
										help="At least 0.5s above caution."
									/>
									<NumberSetting
										label="Slow below"
										unit="km/h"
										min={0.5}
										max={20}
										bind:value={
											() => liveSettings.speedThreshold,
											(v: number) => patchSettings({ speedThreshold: v })
										}
										help="Below this counts as slow/stopped."
									/>
								</div>
								{#if liveSettings.alertSeconds < liveSettings.cautionSeconds + 0.5}
									<p role="alert" class="text-xs text-error-700-300">
										Recovery delay must be at least 0.5s above caution.
									</p>
								{/if}
								<div class="space-y-1">
									<SettingSwitch
										label="Mute alert sound"
										description="Mute on every connected dashboard."
										checked={!!liveSettings.muteAlertSound}
										onChange={(v) => patchSettings({ muteAlertSound: v })}
									/>
									<SettingSwitch
										label="Hide AI drivers"
										description="Hide bots from map, standings, and alerts."
										checked={!!liveSettings.hideBots}
										onChange={(v) => patchSettings({ hideBots: v })}
									/>
								</div>
							</section>
						</div>
					</div>
					</div>
					{#if notice}<p role="alert" class="mt-4 text-sm text-error-700-300">{notice}</p>{/if}
				</ResponsiveDialogContent>
			</Dialog>
		</div>
	</header>
	<main class="race-content" class:is-smoked={configSmoke} aria-hidden={configSmoke && !settingsOpen}>
		{#if notice && !settingsOpen}<p role="alert">{notice}</p>{/if}

		<div class="race-grid">
			<section class="race-panel card track-panel">
				<div class="race-panel-heading">
					<div><h2 class="panel-title">Track map</h2></div>
					<div class="map-buttons flex items-center gap-2">
						<span class="badge preset-tonal-primary"
							>{snapshot.mode === 'demo'
								? 'Demo'
								: snapshot.connection === 'live'
									? 'Live'
									: snapshot.connection === 'stale'
										? 'Feed lost'
										: snapshot.connection === 'offline'
											? 'Disconnected'
											: 'Waiting'}</span
						>
						<button
							id="mapFull"
							type="button"
							class="btn-icon preset-tonal-surface"
							aria-label="Fullscreen"
							title="Fullscreen"><ControlIcon name="expand" /></button
						>
					</div>
				</div>
				<div
					id="mapSurface"
					class="map-surface"
					style:--portrait-dock={portraitDock + '%'}
					style:--landscape-dock={landscapeDock + '%'}
				>
					<div class="fullscreen-toolbar">
						<button
							id="undoUnrecoverable"
							class="btn preset-tonal-surface"
							type="button"
							hidden={dismissed.size === 0}
							onclick={undo}>Undo dismissal</button
						>
						<div class="fullscreen-zoom" role="group" aria-label="Map controls">
							<button
								type="button"
								class="btn preset-tonal-surface"
								aria-pressed={driverUpMap}
								title="Switch map orientation"
								onclick={() => {
									driverUpMap = !driverUpMap;
									map?.setDriverUp(driverUpMap);
								}}>{driverUpMap ? 'Driver up' : 'North up'}</button
							>
							<button
								type="button"
								class="btn preset-tonal-surface"
								aria-pressed={elevatedMap}
								aria-label="3D elevation view"
								title="Toggle elevation view"
								onclick={() => {
									elevatedMap = !elevatedMap;
									map?.setElevated(elevatedMap);
								}}>{elevatedMap ? '3D' : '2D'}</button
							>
							<button
								id="zoomOut"
								class="btn-icon preset-tonal-surface"
								type="button"
								aria-label="Zoom out"
								title="Zoom out"><ControlIcon name="minus" /></button
							>
							<button
								id="zoomFit"
								class="btn-icon preset-tonal-surface"
								type="button"
								aria-label="Fit track"
								title="Fit track"><ControlIcon name="fit" /></button
							>
							<button
								id="zoomIn"
								class="btn-icon preset-tonal-surface"
								type="button"
								aria-label="Zoom in"
								title="Zoom in"><ControlIcon name="plus" /></button
							>
						</div>
						<button
							id="mapTrueFullscreen"
							class="btn preset-tonal-surface"
							type="button"
							aria-pressed="false">True fullscreen</button
						>
						<button
							id="mapExit"
							class="btn-icon preset-tonal-surface"
							type="button"
							aria-label="Exit fullscreen"
							title="Exit fullscreen"><ControlIcon name="close" /></button
						>
					</div>
					<svg id="map" viewBox="0 0 900 560" role="img" aria-label="Live vehicle map"
						><defs
							><pattern id="grid" width="32" height="32" patternUnits="userSpaceOnUse"
								><path d="M32 0H0V32" fill="none" stroke="#151d1e" stroke-width=".5" /></pattern
							></defs
						><rect width="100%" height="100%" fill="url(#grid)" /><g id="pathLayer"></g><g
							id="trackFeatureLayer"
						></g><g id="carLayer"></g></svg
					>
					{#if !raceView}
						<!-- svelte-ignore a11y_no_noninteractive_tabindex, a11y_no_noninteractive_element_interactions (A movable ARIA window splitter must be focusable.) -->
						<div
							class="dock-resizer"
							role="separator"
							tabindex="0"
							aria-label="Resize recovery alerts"
							title="Drag to resize recovery alerts"
							aria-controls="fullscreenAlerts"
							aria-orientation={landscape ? 'vertical' : 'horizontal'}
							aria-valuemin="18"
							aria-valuemax="55"
							aria-valuenow={dockSize}
							aria-valuetext={dockSize + '% alert space'}
							onpointerdown={startDockDrag}
							onpointermove={dragDock}
							onlostpointercapture={() => (dockDrag = null)}
							onkeydown={resizeDockWithKey}
						>
							<span aria-hidden="true"></span>
						</div>
						<aside
							id="fullscreenAlerts"
							class="recovery-toasts"
							aria-label="Active recovery alerts"
						>
							{#each incidents as car (car.incidentId || car.slot)}
								<Incident
									{car}
									{speedUnit}
									select={() => map?.focusIncident(car.slot)}
									dismiss={() => dismiss(car.incidentId)}
								/>
							{/each}
						</aside>
					{/if}
					{#if !visibleCars.length}<div class="empty" role="status">
							<h3>{mapStatus[0]}</h3>
							<p>{mapStatus[1]}</p>
						</div>{/if}
				</div>
			</section>

			<aside class="race-panel card p-4">
				{#if !raceView}<h2 class="panel-title">
						Recovery queue
						<span class="flex flex-wrap gap-1">
							{#if alerts}<span class="badge preset-tonal-error" title="Need recovery"
									>{alerts} urgent</span
								>{/if}
							{#if incidents.length > alerts}<span
									class="badge preset-tonal-warning"
									title="Possible issues">{incidents.length - alerts} to check</span
								>{/if}
						</span>
					</h2>
					<div class="queue rounded-container bg-surface-100-900 p-2">
						{#each incidents as car (car.incidentId || car.slot)}
							<Incident
								{car}
								{speedUnit}
								select={() => map?.focusIncident(car.slot)}
								dismiss={() => dismiss(car.incidentId)}
							/>
						{:else}<p class="text-sm">All clear</p>{/each}
					</div>
				{/if}
				{#if !raceView}
					<details class="my-3 rounded-container border border-surface-300-700 p-2">
						<summary class="cursor-pointer text-sm font-semibold"
							>Incident history ({incidentHistory.length})</summary
						>
						<div class="mt-2 max-h-72 space-y-2 overflow-auto">
							{#each incidentHistory as entry (entry.id)}
								<article class="rounded-base bg-surface-100-900 p-2 text-sm">
									<strong>{entry.driver}</strong> · {entry.state === 'alert'
										? 'Needs help'
										: 'Check car'} · {entry.outcome}
									<p>{new Date(entry.started).toLocaleTimeString()} · {entry.score ?? '—'} / 100</p>
									{#if entry.factors?.length}
										<ul class="my-1 space-y-0.5">
											{#each entry.factors as factor (factor.label)}
												<li>+{factor.points} · {factor.label}</li>
											{/each}
										</ul>
									{:else}<p>{entry.reason || 'Score breakdown unavailable'}</p>{/if}
									{#if entry.outcome !== 'False alarm'}
										<button
											type="button"
											class="btn preset-tonal-surface mt-1"
											onclick={() => markFalseAlarm(entry)}>False alarm</button
										>
									{/if}
								</article>
							{:else}<p class="text-sm">No incidents recorded.</p>{/each}
						</div>
					</details>
				{/if}
				<h2 class="panel-title">Mini-leaderboard</h2>
				<p class="lap-summary">
					Current lap: <span
						><strong>{currentLap}{race?.laps ? ' / ' + race.laps : ''}</strong>{#if race?.track}<b
								>{race.track}</b
							>{/if}</span
					>
				</p>
				<div class="table-wrap mini-standings">
					<table class="table">
						<thead><tr><th>Pos</th><th>Driver</th><th>HP</th><th>Time</th><th>Gap</th></tr></thead>
						<tbody>
							{#each standings as car (car.slot)}
								<tr
									class="mini-driver-row"
									class:following-driver={followedSlot === car.slot}
									class:medic-driver={medicSlots.has(car.slot)}
								>
									<td>{car.place || '—'}</td>
									<td
										><Popover positioning={{ placement: 'bottom-start' }}>
											<Popover.Trigger
												class="mini-driver-action"
												aria-label={'Actions for ' + (car.name || 'Car ' + (car.slot + 1))}
											>
												<strong>{car.name || 'Car ' + (car.slot + 1)}</strong
												>{#if medicSlots.has(car.slot)}
													<small class="stat-medic">Medic</small>{/if}
												{#if followedSlot === car.slot}
													<small class="stat-good">Following</small>{/if}
											</Popover.Trigger>
											<Portal>
												<Popover.Positioner class="z-40">
													<Popover.Content
														class="card grid border border-surface-300-700 bg-surface-50-950 p-1 shadow-xl"
													>
														<Popover.Title class="sr-only">Driver actions</Popover.Title>
														<Popover.Context>
															{#snippet children(menu)}
																<button
																	type="button"
																	class="btn justify-start"
																	disabled={followedSlot !== car.slot && !carsBySlot.has(car.slot)}
																	onclick={() => {
																		toggleFollow(car.slot);
																		menu().setOpen(false);
																	}}>{followedSlot === car.slot ? 'Unfollow' : 'Follow'}</button
																>
																<button
																	type="button"
																	class="btn justify-start"
																	onclick={() => {
																		void markMedic(car.slot);
																		menu().setOpen(false);
																	}}
																	>{medicSlots.has(car.slot)
																		? 'Unmark as medic'
																		: 'Mark as medic'}</button
																>
															{/snippet}
														</Popover.Context>
													</Popover.Content>
												</Popover.Positioner>
											</Portal>
										</Popover>{#if car.status === 3}
											<span class="stat-finished" title="Finished">(F)</span>
										{:else if car.status >= 4}
											<span class="stat-danger" title={driverStatuses[car.status]}
												>{({ 5: '(DSQ)', 6: '(DNF r)', 8: '(DNF w)' } as Record<number, string>)[
													car.status
												] ?? '(DNF)'}</span
											>
										{/if}</td
									>
									<td
										><span
											class:stat-good={car.health >= 70}
											class:stat-low={car.health >= 30 && car.health < 70}
											class:stat-danger={car.health < 30}>{car.health}%</span
										></td
									>
									<td
										>{clock(
											car.status === 3 ? (car.best ?? car.last ?? car.current) : car.current,
										)}</td
									>
									<td
										>{#if car.place === 1}<strong class="stat-good">Lead</strong>{:else}{gap(
												car.gap,
											)}{/if}</td
									>
								</tr>
							{:else}<tr><td colspan="5">Waiting for standings</td></tr>{/each}
						</tbody>
					</table>
				</div>
			</aside>
			<section class="race-panel card p-4 leaderboard-panel">
				<h2 class="panel-title">
					{#if compactLayout}
						<button
							type="button"
							class="panel-toggle"
							aria-expanded={leaderboardOpen}
							aria-controls="raceLeaderboard"
							onclick={() => (leaderboardOpen = !leaderboardOpen)}
						>
							<span>Ext. leaderboard</span><span class="badge preset-tonal-surface"
								>{standings.length} drivers</span
							><span aria-hidden="true">{leaderboardOpen ? '▴' : '▾'}</span>
						</button>
					{:else}
						Ext. leaderboard <span class="badge preset-tonal-surface"
							>{standings.length} drivers</span
						>
					{/if}
				</h2>
				<div id="raceLeaderboard" class="table-wrap" hidden={compactLayout && !leaderboardOpen}>
					<table class="table extended-standings">
						<thead
							><tr
								><th>Pos</th><th>Player</th><th>Car</th><th>Status</th><th>Track status</th><th
									>Lap</th
								><th>Progress</th><th>HP</th><th>Frags / wrecks</th><th>Gap</th><th>Last lap</th><th
									>Best lap</th
								><th>Best sectors 1/2/3</th><th>Speed</th></tr
							></thead
						>
						<tbody>
							{#each standings as car (car.slot)}
								{@const live = carsBySlot.get(car.slot)}
								<tr>
									<td>{car.place ? 'P' + car.place : '—'}</td>
									<td
										><strong>{car.name || 'Car ' + (car.slot + 1)}</strong><small
											class="driver-slot">Slot: {car.slot}</small
										></td
									>
									<td
										><span class="flex items-center gap-2"
											>{#if carPreview(car.carId)}<img
													src={carPreview(car.carId)}
													class="h-8 w-14 shrink-0 object-contain"
													alt=""
													loading="lazy"
												/>{/if}{car.carName || car.carId || '—'}</span
										></td
									>
									<td
										><span
											class="timing-badge"
											class:stat-good={car.status === 2}
											class:stat-finished={car.status === 3}
											class:stat-danger={car.status >= 4}
											>{#if car.status >= 4 && car.status !== 5}DNF<br />({car.status === 7
													? 'Timeout'
													: driverStatuses[car.status]}){:else}{driverStatuses[car.status] ??
													'—'}{/if}</span
										></td
									>
									<td
										>{car.trackStatus == null
											? '—'
											: (trackStatuses[car.trackStatus] ?? 'Unknown')}</td
									>
									<td
										><strong>{car.lap}</strong>{#if race?.laps}
											<span class="timing-muted">/ {race.laps}</span>{/if}</td
									>
									<td
										>{#if car.progress != null}<span class="timing-progress"
												><progress
													class="lap-progress"
													max="1"
													value={car.progress}
													aria-label="Lap progress"
												></progress><strong>{Math.round(car.progress * 100)}%</strong></span
											>{:else}—{/if}</td
									>
									<td
										><span
											class="health-badge"
											class:stat-good={car.health >= 70}
											class:stat-low={car.health >= 30 && car.health < 70}
											class:stat-danger={car.health < 30}>{car.health}%</span
										></td
									>
									<td
										>{#if car.frags != null && car.wrecks != null}<strong
												class="stat-good sector-time">F: {car.frags}</strong
											><strong class="stat-danger sector-time">W: {car.wrecks}</strong
											>{:else}—{/if}</td
									>
									<td
										>{#if car.place === 1}<strong class="stat-good">Leader</strong>{:else}{gap(
												car.gap,
											)}{/if}</td
									>
									<td>{clock(car.last)}</td>
									<td
										><strong
											class="best-time"
											class:fastest-time={car.best != null && car.best === fastestDriver?.best}
											>{clock(car.best)}</strong
										></td
									>
									<td
										>{#if car.bestSectors}{#each car.bestSectors as sector, index (index)}<span
													class="sector-time"
													><span class="timing-muted">S{index + 1}:</span> {clock(sector)}</span
												>{/each}{:else}—{/if}</td
									>
									<td
										>{#if live}<strong>{displaySpeed(live.speed, speedUnit)}</strong>
											<small class="timing-muted">{speedUnit}</small>{:else}—{/if}</td
									>
								</tr>
							{:else}<tr><td colspan="14">No participants yet</td></tr>{/each}
						</tbody>
					</table>
				</div>
			</section>
			<aside class="race-panel card p-4 session-panel">
				<h2 class="panel-title">
					{#if compactLayout}
						<button
							type="button"
							class="panel-toggle"
							aria-expanded={raceDetailsOpen}
							aria-controls="raceDetails"
							onclick={() => (raceDetailsOpen = !raceDetailsOpen)}
						>
							<span>Race details</span><span class="badge preset-tonal-surface">{raceState}</span
							><span aria-hidden="true">{raceDetailsOpen ? '▴' : '▾'}</span>
						</button>
					{:else}
						Race details
					{/if}
				</h2>
				<div id="raceDetails" hidden={compactLayout && !raceDetailsOpen}>
					{#if trackImage}<img
							src={trackImage}
							alt={race?.track || 'Track preview'}
							class="mt-3 max-h-40 w-full rounded-container object-contain"
							loading="lazy"
						/>{#if snapshot.mode === 'demo'}<p class="mt-1 text-xs">
								Example track preview · map uses simulated positions
							</p>{/if}{/if}
					<dl class="session-rows">
						{#each raceRows as [label, value]}<dt>{label}</dt>
							<dd
								class:stat-good={label === 'Race status' && value === 'Racing'}
								class:race-status-value={label === 'Race status' && value !== 'Racing'}
								class:timing-muted={label === 'Track ID'}
								class:track-name-value={label === 'Track name'}
							>
								{value}
							</dd>{/each}
					</dl>
				</div>
			</aside>
		</div>
	</main>
</div>
{#if phoneVeil}
	<DashSmoke message={veilMessage} heavy />
{/if}

<style>
	/* Phone: full-shell dim (DashSmoke covers header). */
	:global(.race-shell.is-smoked:not(:has(.race-content.is-smoked))) {
		filter: brightness(0.58) saturate(0.72);
		pointer-events: none;
		user-select: none;
	}
	/* Desktop: dim content only; Settings gear stays clickable. Dialogs z-[100] above header. */
	.race-content {
		position: relative;
	}
	.race-content.is-smoked :global(> :not(.dash-smoke)) {
		filter: brightness(0.58) saturate(0.72);
		pointer-events: none;
		user-select: none;
	}
	/* Path input ok/bad borders (theme tokens). */
	:global(input.config-path-input.is-ok) {
		border-color: var(--color-success-500) !important;
		box-shadow: 0 0 0 1px var(--color-success-500);
	}
	:global(input.config-path-input.is-bad) {
		border-color: var(--color-error-500) !important;
		box-shadow: 0 0 0 1px var(--color-error-500);
	}
	.config-section.needs-attention {
		border-radius: var(--radius-container, 0.75rem);
		box-shadow: 0 0 0 2px var(--color-error-500);
		padding: 0.75rem;
	}
	:global(.config-action-attention) {
		box-shadow: 0 0 0 2px var(--color-primary-500);
	}
	/* First-run: Config stays primary; other panels stay usable but quieter. */
	.settings-secondary {
		opacity: 0.55;
	}
</style>
