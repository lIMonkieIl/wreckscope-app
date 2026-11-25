export type UpdateStatus = {
	current: string;
	latest: string;
	available: boolean;
	mode: 'portable' | 'installed' | string;
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
	return import('../../../../bindings/wreckscope/internal/app/app.js');
}

export async function getAppVersion(): Promise<string> {
	const app = await loadApp();
	return (await app.AppVersion()) ?? '';
}

export async function checkForUpdates(): Promise<UpdateStatus> {
	const app = await loadApp();
	return asStatus((await app.CheckForUpdates()) as Record<string, unknown>);
}

export async function applyUpdate(): Promise<UpdateStatus> {
	const app = await loadApp();
	return asStatus((await app.ApplyUpdate()) as Record<string, unknown>);
}

export async function lastUpdateStatus(): Promise<UpdateStatus> {
	const app = await loadApp();
	return asStatus((await app.LastUpdateStatus()) as Record<string, unknown>);
}

export function onUpdateAvailable(handler: (status: UpdateStatus) => void) {
	let cancelled = false;
	let unsubscribe: (() => void) | undefined;
	(async () => {
		try {
			const { Events } = await import('@wailsio/runtime');
			if (cancelled) return;
			unsubscribe = Events.On('update-available', (event) => {
				handler(asStatus(event.data as Record<string, unknown>));
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
