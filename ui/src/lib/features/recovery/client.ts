import type {
	DetectionSettings,
	PhoneAccess,
	TelemetrySnapshot,
} from './telemetry.ts';
import { hasDesktopBridge, whenDesktopRuntimeReady } from './desktopBridge.ts';

type TelemetryListener = (snapshot: TelemetrySnapshot) => void;

/** Desktop WebView with native IPC ready (not LAN phone). */
async function useDesktopBridge(): Promise<boolean> {
	if (!hasDesktopBridge()) return false;
	return whenDesktopRuntimeReady();
}

async function loadBindings() {
	const app = await import('../../../../bindings/wreckscope/internal/app/app.js');
	const { Events } = await import('@wailsio/runtime');
	return { app, Events };
}

async function postJSON(route: 'mode' | 'settings' | 'medic', payload: unknown) {
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
			if (await useDesktopBridge()) {
				if (cancelled) return;
				const { app, Events } = await loadBindings();
				unsubscribe = Events.On('telemetry', (event) => {
					if (cancelled) return;
					onSnapshot(event.data as TelemetrySnapshot);
				});
				const initial = (await app.Snapshot()) as TelemetrySnapshot;
				if (!cancelled) onSnapshot(initial);
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
					onSnapshot(JSON.parse(event.data) as TelemetrySnapshot);
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
	if (await useDesktopBridge()) {
		const { app } = await loadBindings();
		await app.SetMode(mode);
		return;
	}
	await postJSON('mode', { mode });
}

/** Assign or clear a medic (PC + LAN; no cap). */
export async function toggleMedic(slot: number): Promise<number[]> {
	if (await useDesktopBridge()) {
		const { app } = await loadBindings();
		return (await app.ToggleMedic(slot)) as number[];
	}
	const result = (await postJSON('medic', { slot })) as { medics?: number[] };
	return result?.medics ?? [];
}

export async function saveDetectionSettings(settings: DetectionSettings) {
	if (await useDesktopBridge()) {
		const { app } = await loadBindings();
		return (await app.SaveSettings(settings)) as DetectionSettings;
	}
	return (await postJSON('settings', settings)) as DetectionSettings;
}

export async function getPhoneAccess(): Promise<PhoneAccess> {
	if (await useDesktopBridge()) {
		const { app } = await loadBindings();
		return (await app.PhoneAccess()) as PhoneAccess;
	}
	const response = await fetch('/api/access', { cache: 'no-store' });
	if (!response.ok) throw new Error('Could not load the phone link.');
	return (await response.json()) as PhoneAccess;
}
