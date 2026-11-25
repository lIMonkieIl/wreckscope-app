/** Theme hover preview on document; does not persist. */

let hoverTheme: string | null = null;
let committedTheme = '';

export function setCommittedDocumentTheme(theme: string) {
	committedTheme = theme;
	if (hoverTheme == null) {
		document.documentElement.dataset.theme = theme;
	}
}

/** Preview a theme id, or null to restore the committed one. */
export function setThemeHoverPreview(theme: string | null) {
	hoverTheme = theme;
	document.documentElement.dataset.theme = hoverTheme ?? committedTheme;
}

export function clearThemeHoverPreview() {
	setThemeHoverPreview(null);
}
