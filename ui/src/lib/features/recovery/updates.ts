import { hasDesktopBridge, whenDesktopRuntimeReady } from './desktopBridge.ts';

export type UpdateStatus = {
	current: string;
	latest: string;
	available: boolean;
	mode: 'installed' | string;
	message: string;
};

function emptyStatus(): UpdateStatus {
	return { current: '', latest: '', available: false, mode: '', message: '' };
}

function asStatus(value: Record<string, unknown> | null | undefined): UpdateStatus {
	if (!value) return emptyStatus();
	return {
		current: typeof value.current === 'string' ? value.current : '',
		latest: typeof value.latest === 'string' ? value.latest : '',
		available: !!value.available,
		mode: typeof value.mode === 'string' ? value.mode : '',
		message: typeof value.message === 'string' ? value.message : '',
	};
}

async function loadApp() {
	if (!(await whenDesktopRuntimeReady())) {
		throw new Error('desktop bridge unavailable');
	}
	return import('../../../../bindings/wreckscope/internal/app/app.js');
}

/** Baked app version — desktop bindings, or LAN `/api/version` on phone. */
export async function getAppVersion(): Promise<string> {
	if (hasDesktopBridge()) {
		try {
			const app = await loadApp();
			const v = (await app.AppVersion()) ?? '';
			if (v) return v;
		} catch {
			/* fall through to HTTP */
		}
	}
	try {
		const response = await fetch('/api/version', { cache: 'no-store' });
		if (!response.ok) return '';
		const body = (await response.json()) as { version?: string };
		return typeof body.version === 'string' ? body.version : '';
	} catch {
		return '';
	}
}

export async function applyUpdate(): Promise<UpdateStatus> {
	const app = await loadApp();
	return asStatus((await app.ApplyUpdate()) as unknown as Record<string, unknown>);
}

export async function lastUpdateStatus(): Promise<UpdateStatus> {
	const app = await loadApp();
	return asStatus((await app.LastUpdateStatus()) as unknown as Record<string, unknown>);
}

export async function checkForUpdates(): Promise<UpdateStatus> {
	const app = await loadApp();
	return asStatus((await app.CheckForUpdates()) as unknown as Record<string, unknown>);
}

/** Background quiet-check results (up-to-date or available). */
export function onUpdateStatus(handler: (status: UpdateStatus) => void) {
	let cancelled = false;
	let unsubscribe: (() => void) | undefined;
	(async () => {
		try {
			if (!(await whenDesktopRuntimeReady()) || cancelled) return;
			const { Events } = await import('@wailsio/runtime');
			if (cancelled) return;
			const onEvent = (event: { data?: unknown }) => {
				handler(asStatus(event.data as Record<string, unknown>));
			};
			const stopStatus = Events.On('update-status', onEvent);
			const stopAvailable = Events.On('update-available', onEvent);
			unsubscribe = () => {
				stopStatus?.();
				stopAvailable?.();
			};
		} catch {
			/* phone / no bridge */
		}
	})();
	return () => {
		cancelled = true;
		unsubscribe?.();
	};
}

/** True while a manual or quiet update check is in flight. */
export function onUpdateChecking(handler: (checking: boolean) => void) {
	let cancelled = false;
	let unsubscribe: (() => void) | undefined;
	(async () => {
		try {
			if (!(await whenDesktopRuntimeReady()) || cancelled) return;
			const { Events } = await import('@wailsio/runtime');
			if (cancelled) return;
			unsubscribe = Events.On('update-checking', (event: { data?: unknown }) => {
				const raw = Array.isArray(event.data) ? event.data[0] : event.data;
				handler(raw === true || raw === 'true');
			});
		} catch {
			/* phone / no bridge */
		}
	})();
	return () => {
		cancelled = true;
		unsubscribe?.();
	};
}
