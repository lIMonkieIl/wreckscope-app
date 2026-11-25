export type TelemetrySetup = {
	ready: boolean;
	valid: boolean;
	needsWizard: boolean;
	setupCompleted: boolean;
	canAuto: boolean;
	hasBackup: boolean;
	path: string;
	error: string;
};

import { hasDesktopBridge, whenDesktopRuntimeReady } from './desktopBridge.ts';

export { hasDesktopBridge };

const OPEN_SETTINGS_EVENT = 'wreckscope-open-settings';

async function loadBindings() {
	if (!(await whenDesktopRuntimeReady())) {
		throw new Error('desktop bridge unavailable');
	}
	return import('../../../../bindings/wreckscope/internal/app/app.js');
}

function asSetup(value: Record<string, unknown>): TelemetrySetup {
	return {
		ready: !!value.ready,
		valid: !!value.valid,
		needsWizard: !!value.needsWizard,
		setupCompleted: !!value.setupCompleted,
		canAuto: value.canAuto !== false,
		hasBackup: !!value.hasBackup,
		path: typeof value.path === 'string' ? value.path : '',
		error: typeof value.error === 'string' ? value.error : '',
	};
}

export async function getSetupStatus(): Promise<TelemetrySetup> {
	const app = await loadBindings();
	return asSetup((await app.SetupStatus()) as Record<string, unknown>);
}

export async function completeSetup(): Promise<TelemetrySetup> {
	const app = await loadBindings();
	return asSetup((await app.CompleteSetup()) as Record<string, unknown>);
}

export async function probeTelemetry(): Promise<TelemetrySetup> {
	const app = await loadBindings();
	return asSetup((await app.ProbeTelemetry()) as Record<string, unknown>);
}

export async function probeTelemetryAt(path: string): Promise<TelemetrySetup> {
	const app = await loadBindings();
	return asSetup((await app.ProbeTelemetryAt(path)) as Record<string, unknown>);
}

export async function findTelemetryConfig(): Promise<TelemetrySetup> {
	const app = await loadBindings();
	return asSetup((await app.FindTelemetryConfig()) as Record<string, unknown>);
}

export async function enableTelemetryAuto(): Promise<TelemetrySetup> {
	const app = await loadBindings();
	return asSetup((await app.EnableTelemetry()) as Record<string, unknown>);
}

export async function browseTelemetryConfig(): Promise<string> {
	const app = await loadBindings();
	return (await app.BrowseTelemetryConfig()) ?? '';
}

export async function enableTelemetryAt(path: string): Promise<TelemetrySetup> {
	const app = await loadBindings();
	return asSetup((await app.EnableTelemetryAt(path)) as Record<string, unknown>);
}

export async function restoreTelemetryBackup(path = ''): Promise<TelemetrySetup> {
	const app = await loadBindings();
	return asSetup((await app.RestoreTelemetryBackup(path)) as Record<string, unknown>);
}

export async function resetTelemetrySetup(): Promise<TelemetrySetup> {
	const app = await loadBindings();
	return asSetup((await app.ResetTelemetrySetup()) as Record<string, unknown>);
}

export function requestOpenSettings() {
	window.dispatchEvent(new Event(OPEN_SETTINGS_EVENT));
}

export function onOpenSettingsRequest(handler: () => void) {
	window.addEventListener(OPEN_SETTINGS_EVENT, handler);
	return () => window.removeEventListener(OPEN_SETTINGS_EVENT, handler);
}
