import type {
	AppearanceSettings,
	DetectionSettings,
	PhoneAccess,
	TelemetrySnapshot,
} from './telemetry.ts';

type TelemetryListener = (snapshot: TelemetrySnapshot) => void;
type SettingsListener = (settings: DetectionSettings) => void;

const settingsListeners = new Set<SettingsListener>();

function publishSettings(settings: DetectionSettings | undefined) {
	if (!settings) return;
	for (const listener of settingsListeners) listener(settings);
}

function hasWails(): boolean {
	return !!(window as unknown as { _wails?: unknown })._wails;
}

async function loadBindings() {
	const app = await import('../../../../bindings/wreckscope/internal/app/app.js');
	const { Events } = await import('@wailsio/runtime');
	return { app, Events };
}

async function postJSON(route: 'mode' | 'settings' | 'appearance' | 'medic', payload: unknown) {
	const response = await fetch('/api/' + route, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(payload),
	});
	if (!response.ok) throw new Error((await response.text()).trim() || response.statusText);
	if (response.status === 204) return undefined;
	return response.json();
}

export function subscribeTelemetry(
	onSnapshot: TelemetryListener,
	onDisconnect: () => void,
	onReconnect?: () => void,
) {
	let cancelled = false;
	let unsubscribe: (() => void) | undefined;
	let stream: EventSource | undefined;

	(async () => {
		try {
			if (hasWails()) {
				const { app, Events } = await loadBindings();
				unsubscribe = Events.On('telemetry', (event) => {
					if (cancelled) return;
					const snapshot = event.data as TelemetrySnapshot;
					publishSettings(snapshot.settings);
					onSnapshot(snapshot);
				});
				const initial = (await app.Snapshot()) as TelemetrySnapshot;
				if (!cancelled) {
					publishSettings(initial.settings);
					onSnapshot(initial);
				}
				return;
			}

			// Phone SSE auto-reconnects; flag drops so UI can hard-refresh instead of stale resume.
			let sawDisconnect = false;
			stream = new EventSource('/api/stream');
			stream.onopen = () => {
				if (cancelled || !sawDisconnect) return;
				sawDisconnect = false;
				onReconnect?.();
			};
			stream.onmessage = (event) => {
				try {
					if (cancelled) return;
					const snapshot = JSON.parse(event.data) as TelemetrySnapshot;
					publishSettings(snapshot.settings);
					onSnapshot(snapshot);
				} catch (error) {
					console.warn('Telemetry snapshot', error);
				}
			};
			stream.onerror = () => {
				if (cancelled) return;
				sawDisconnect = true;
				onDisconnect();
			};
		} catch {
			if (!cancelled) onDisconnect();
		}
	})();

	return () => {
		cancelled = true;
		unsubscribe?.();
		stream?.close();
	};
}

export async function setMode(mode: TelemetrySnapshot['mode']) {
	if (hasWails()) {
		const { app } = await loadBindings();
		await app.SetMode(mode);
		return;
	}
	await postJSON('mode', { mode });
}

/** Assign or clear a medic (PC + LAN; no cap). */
export async function toggleMedic(slot: number): Promise<number[]> {
	if (hasWails()) {
		const { app } = await loadBindings();
		return (await app.ToggleMedic(slot)) as number[];
	}
	const result = (await postJSON('medic', { slot })) as { medics?: number[] };
	return result?.medics ?? [];
}

export async function saveDetectionSettings(settings: DetectionSettings) {
	if (hasWails()) {
		const { app } = await loadBindings();
		const saved = (await app.SaveSettings(settings)) as DetectionSettings;
		publishSettings(saved);
		return saved;
	}
	const saved = (await postJSON('settings', settings)) as DetectionSettings;
	publishSettings(saved);
	return saved;
}

export async function getSettings(): Promise<DetectionSettings> {
	if (hasWails()) {
		const { app } = await loadBindings();
		return (await app.GetSettings()) as DetectionSettings;
	}
	const response = await fetch('/api/settings', { cache: 'no-store' });
	if (!response.ok) throw new Error((await response.text()).trim() || response.statusText);
	return (await response.json()) as DetectionSettings;
}

export async function saveAppearance(appearance: AppearanceSettings) {
	if (hasWails()) {
		const { app } = await loadBindings();
		const saved = (await app.SaveAppearance(appearance.theme, appearance.mode)) as DetectionSettings;
		publishSettings(saved);
		return;
	}
	const saved = (await postJSON('appearance', appearance)) as DetectionSettings;
	publishSettings(saved);
}

/** Shared settings. Desktop: Wails events + GetSettings. Phone: GET + SSE. */
export function subscribeSettings(onSettings: SettingsListener) {
	let cancelled = false;
	let unsubscribe: (() => void) | undefined;
	settingsListeners.add(onSettings);

	(async () => {
		try {
			if (hasWails()) {
				const { app, Events } = await loadBindings();
				unsubscribe = Events.On('settings', (event) => {
					if (!cancelled) onSettings(event.data as DetectionSettings);
				});
				const initial = (await app.GetSettings()) as DetectionSettings;
				if (!cancelled) onSettings(initial);
				return;
			}

			const initial = await getSettings();
			if (!cancelled) onSettings(initial);
		} catch (error) {
			console.warn('Settings subscribe', error);
		}
	})();

	return () => {
		cancelled = true;
		settingsListeners.delete(onSettings);
		unsubscribe?.();
	};
}

export async function getPhoneAccess(): Promise<PhoneAccess> {
	if (hasWails()) {
		const { app } = await loadBindings();
		return (await app.PhoneAccess()) as PhoneAccess;
	}
	const response = await fetch('/api/access', { cache: 'no-store' });
	if (!response.ok) throw new Error('Could not load the phone link.');
	return (await response.json()) as PhoneAccess;
}
