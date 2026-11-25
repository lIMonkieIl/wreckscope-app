const WHATS_NEW_KEY = 'wreckscope.whatsNew.seenVersion';

export function seenWhatsNewVersion(): string {
	try {
		return localStorage.getItem(WHATS_NEW_KEY) || '';
	} catch {
		return '';
	}
}

export function markWhatsNewSeen(version: string) {
	const cleaned = version.trim();
	if (!cleaned) return;
	try {
		localStorage.setItem(WHATS_NEW_KEY, cleaned);
	} catch {
		/* private mode */
	}
}
