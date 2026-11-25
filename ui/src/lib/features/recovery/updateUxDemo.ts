/** Temporary: force splash + update banner for visual review. Flip to false when done. */
export const UPDATE_UX_DEMO = true;

export const UPDATE_UX_DEMO_LATEST = '0.2.0';

const DISMISS_KEY = 'wreckscope.updateBanner.dismissedVersion';

export function dismissedUpdateVersion(): string {
	try {
		return localStorage.getItem(DISMISS_KEY) || '';
	} catch {
		return '';
	}
}

export function dismissUpdateVersion(version: string) {
	const cleaned = version.trim();
	if (!cleaned) return;
	try {
		localStorage.setItem(DISMISS_KEY, cleaned);
	} catch {
		/* private mode */
	}
}
