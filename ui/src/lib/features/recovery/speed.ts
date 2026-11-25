import type { SpeedUnit } from './telemetry';
export const displaySpeed = (kmh: number, unit: SpeedUnit) =>
	Math.round(unit === 'mph' ? kmh / 1.609344 : kmh);
