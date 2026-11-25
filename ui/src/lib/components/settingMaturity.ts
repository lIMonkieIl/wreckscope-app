/** Settings maturity tags. Omit when stable. */
export type SettingMaturity = 'new' | 'beta' | 'testing' | 'dev';

export const settingMaturityLabel: Record<SettingMaturity, string> = {
	new: 'New',
	beta: 'Beta',
	testing: 'Testing',
	dev: 'Dev',
};

export const settingMaturityHint: Record<SettingMaturity, string> = {
	new: 'Recently added',
	beta: 'Usable, still settling',
	testing: 'Still working on it',
	dev: 'Development builds only',
};
