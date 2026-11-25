/** One desktop Settings picker open at a time. */
type Closer = () => void;

let active: Closer | null = null;

export function claimPickerOpen(close: Closer) {
	if (active && active !== close) active();
	active = close;
}

export function releasePickerOpen(close: Closer) {
	if (active === close) active = null;
}
