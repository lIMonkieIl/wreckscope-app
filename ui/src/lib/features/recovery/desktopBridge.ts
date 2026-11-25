/**
 * Desktop (Wails WebView) vs phone (LAN browser / SSE) detection for Wails v3.
 *
 * v2-style `!!window._wails` is unreliable: on Windows, Wails injects `_wails`
 * (environment + invoke) on NavigationCompleted — after the first Svelte paint.
 * An empty `{}` from a late `@wailsio/runtime` import is also not a real bridge.
 */

type WailsBridge = {
	environment?: { OS?: string; Arch?: string; Debug?: boolean };
	invoke?: unknown;
};

type HostWindow = Window & {
	_wails?: WailsBridge;
	chrome?: { webview?: unknown };
	webkit?: { messageHandlers?: { external?: unknown } };
};

const READY_EVENT = 'wails:runtime-config-ready';

function host(): HostWindow | undefined {
	if (typeof window === 'undefined') return undefined;
	return window as HostWindow;
}

/** True when this page is inside a native desktop WebView (not a phone browser). */
export function hasDesktopBridge(): boolean {
	const w = host();
	if (!w) return false;
	if (w._wails?.environment?.OS) return true;
	// WebView2 exists before Wails finishes injecting `_wails`.
	if (w.chrome?.webview) return true;
	// WKWebView / WebKitGTK message handler used by Wails invoke.
	if (w.webkit?.messageHandlers?.external) return true;
	return false;
}

/** True when native runtime config has been injected (safe to Call / Events). */
export function isDesktopRuntimeReady(): boolean {
	return !!host()?._wails?.environment?.OS;
}

/**
 * Resolves when desktop runtime config is ready, or immediately with false on phone.
 * Short timeout so a stuck inject does not hang the UI forever.
 */
export function whenDesktopRuntimeReady(timeoutMs = 5000): Promise<boolean> {
	if (!hasDesktopBridge()) return Promise.resolve(false);
	if (isDesktopRuntimeReady()) return Promise.resolve(true);

	return new Promise((resolve) => {
		let settled = false;
		const finish = (ok: boolean) => {
			if (settled) return;
			settled = true;
			cleanup();
			resolve(ok);
		};
		const onReady = () => finish(isDesktopRuntimeReady());
		const cleanup = () => {
			window.removeEventListener(READY_EVENT, onReady);
			window.clearInterval(poll);
			window.clearTimeout(timer);
		};
		window.addEventListener(READY_EVENT, onReady);
		const poll = window.setInterval(() => {
			if (isDesktopRuntimeReady()) finish(true);
		}, 50);
		const timer = window.setTimeout(() => finish(isDesktopRuntimeReady()), timeoutMs);
	});
}

/** Subscribe to runtime-ready (and an immediate tick if already ready). */
export function onDesktopBridgeChange(handler: (desktop: boolean) => void): () => void {
	handler(hasDesktopBridge());
	if (typeof window === 'undefined') return () => {};

	const notify = () => handler(hasDesktopBridge());
	window.addEventListener(READY_EVENT, notify);
	const poll = window.setInterval(notify, 200);
	const timer = window.setTimeout(() => window.clearInterval(poll), 5000);
	return () => {
		window.removeEventListener(READY_EVENT, notify);
		window.clearInterval(poll);
		window.clearTimeout(timer);
	};
}
