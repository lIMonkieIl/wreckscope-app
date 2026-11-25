import { trackOutline } from './track-outline';
import { displaySpeed } from './speed';
import type { RecoveryCar, SpeedUnit, TelemetrySnapshot } from './telemetry';

type Point = { x: number; y?: number; z: number };
type ScreenPoint = { x: number; y: number };
type Bounds = { minX: number; maxX: number; minZ: number; maxZ: number };
type TrackMarker = Point & { heading: number };
type Projection = {
	angle: number;
	cx: number;
	cz: number;
	scale: number;
	fitScale: number;
	width: number;
	height: number;
};
type Gesture = {
	point: ScreenPoint;
	distance: number;
	zoom: number;
	anchor: Point;
	projection: Projection;
};

type Vertex = [number, number, number];

// Roof height, roof front/rear, and axle spacing, estimated from the bundled previews.
const carProfiles: Record<string, [number, number, number, number]> = {
	car01: [0.83, -0.42, 0.72, 1.36],
	car02: [0.79, -0.32, 0.79, 1.39],
	car03: [0.9, -0.48, 0.69, 1.34],
	car04: [0.97, -0.65, 1.17, 1.28],
	car05: [0.83, -0.45, 1.02, 1.31],
	car05_half: [0.54, -0.46, 0.46, 1.25],
	car06: [1.13, -0.98, 1.19, 1.25],
	car07: [0.99, -0.52, 1.47, 1.42],
	car08: [1.0, -0.54, 1.05, 1.38],
	car09: [1.15, -0.83, 0.81, 1.16],
	car10: [1.19, -1.13, 1.15, 1.1],
	car11: [0.83, -0.38, 0.91, 1.43],
	car12: [0.72, -0.47, 0.63, 1.32],
	car13: [1.03, -0.87, 1.22, 1.29],
	car14: [0.75, -0.44, 0.58, 1.41],
	car15: [1.04, -0.78, 1.18, 1.29],
	car16: [0.89, -0.48, 1.26, 1.35],
	car18: [1.14, -0.95, 1.32, 1.34],
	car19: [1.1, -0.85, 1.17, 1.26],
	car21: [0.86, -0.37, 0.82, 1.48],
	car25: [0.8, -0.58, 0.88, 1.4],
	car26: [0.87, -0.51, 0.99, 1.36],
	motorhome: [1.65, -1.15, 1.94, 1.33],
	school_bus: [1.51, -1.61, 1.94, 1.48],
};
function makeWireframe([height, front, rear, axle]: [number, number, number, number]) {
	function rectangle(width: number, y: number, start: number, end: number): Vertex[] {
		return [
			[-width, y, start],
			[width, y, start],
			[width, y, end],
			[-width, y, end],
		];
	}
	const chassis = rectangle(0.95, -0.2, -2.1, 2.1);
	const beltline = rectangle(0.95, 0.35, -2.1, 2.1);
	const roof = rectangle(height > 1.4 ? 0.9 : 0.68, height, front, rear);
	const windows = rectangle(0.85, 0.35, Math.max(-2.1, front - 0.55), Math.min(2.1, rear + 0.4));
	const loops = [chassis, beltline, roof, windows];
	const segments = chassis.map((corner, index) => [corner, beltline[index]]);
	segments.push(...windows.map((corner, index) => [corner, roof[index]]));
	for (const side of [-1, 1]) {
		for (const wheel of [-axle, axle]) {
			const tire: Vertex[] = Array.from({ length: 8 }, (_, index) => {
				const angle = (index * Math.PI) / 4;
				return [side * 1.02, -0.25 + Math.sin(angle) * 0.37, wheel + Math.cos(angle) * 0.37];
			});
			loops.push(tire);
			segments.push([tire[0], tire[4]]);
		}
	}
	return { loops, segments };
}
const wireframes = new Map(
	Object.entries(carProfiles).map(([id, profile]) => [id, makeWireframe(profile)]),
);
const fallbackWireframe = makeWireframe([0.95, -0.55, 0.85, 1.32]);

// SVG map renderer + gestures.
export function mountRescueMap(root: HTMLElement, onUnfollow: () => void) {
	function element<T extends Element>(id: string): T {
		const result = root.querySelector<T>(`#${id}`);
		if (!result) throw new Error(`Rescue map is missing ${id}`);
		return result;
	}
	const NS = 'http://www.w3.org/2000/svg';
	const zoomIn = element<HTMLButtonElement>('zoomIn'),
		zoomOut = element<HTMLButtonElement>('zoomOut'),
		zoomFit = element<HTMLButtonElement>('zoomFit');
	const mapFull = element<HTMLButtonElement>('mapFull'),
		mapExit = element<HTMLButtonElement>('mapExit');
	const listeners = new AbortController(),
		{ signal } = listeners;
	// Medic marker teal (local You stays sky blue).
	const medicFill = '#4fd1c5';
	const medicElevated = '#1a8f84';
	const medicLabel = '#8fe8df';
	let latest: TelemetrySnapshot | null = null,
		unrecoverable: ReadonlySet<string> = new Set(),
		medicSlots: ReadonlySet<number> = new Set(),
		speedUnit: SpeedUnit = 'km/h';
	function node<K extends keyof SVGElementTagNameMap>(
		tag: K,
		attrs: Record<string, string | number> = {},
		text?: string,
	) {
		const n = document.createElementNS(NS, tag);
		for (const [k, v] of Object.entries(attrs)) n.setAttribute(k, String(v));
		if (text !== undefined) n.textContent = text;
		return n;
	}
	let elevated = false;
	let driverUp = true;
	let cameraAngle = 0;
	let driverCentered = true;
	let headingAt = 0;
	let cameraReady = false;
	let cameraFrame: number | null = null;
	function rotate(x: number, z: number, angle: number): [number, number] {
		return [x * Math.cos(angle) - z * Math.sin(angle), x * Math.sin(angle) + z * Math.cos(angle)];
	}
	function depth(point: Point) {
		return elevated ? -point.z * 0.65 - (point.y ?? 0) * 0.76 : -point.z;
	}

	function wireframe(car: RecoveryCar, x: number, y: number, bonnetOnly = false) {
		const heading = ((carHeadings.get(car.slot) ?? 0) * Math.PI) / 180;
		const [qx, qy, qz, qw] = car.orientation ?? [
			0,
			Math.sin(heading / 2),
			0,
			Math.cos(heading / 2),
		];
		// Scale mesh to a typical captured car (packet units).
		const [width, height, length] = car.extents ?? [90, 60, 230];
		const project = ([vx, vy, vz]: Vertex) => {
			const a = (vx * width) / 90;
			const b = (vy * height) / 60;
			const c = (-vz * length) / 230;
			const tx = 2 * (qy * c - qz * b),
				ty = 2 * (qz * a - qx * c),
				tz = 2 * (qx * b - qy * a);
			const rx = a + qw * tx + qy * tz - qz * ty;
			const ry = b + qw * ty + qz * tx - qx * tz;
			const rz = c + qw * tz + qx * ty - qy * tx;
			const [sx, sy] = rotate(rx, depth({ x: rx, y: ry, z: rz }), cameraAngle);
			return [x + sx * 6, y + sy * 6].map((v) => v.toFixed(1)).join(',');
		};
		const { loops, segments } = wireframes.get(car.carId?.split(':')[0] ?? '') ?? fallbackWireframe;
		if (bonnetOnly)
			return 'M' + project([0, 0.4, -1.6]) + 'm-2.5,0a2.5,2.5 0 1,0 5,0a2.5,2.5 0 1,0 -5,0Z';
		return [
			...loops.map((loop) => 'M' + loop.map(project).join('L') + 'Z'),
			...segments.map((segment) => 'M' + segment.map(project).join('L')),
		].join('');
	}
	const trackSamples = new Map<number, Point[]>();
	const footprintCellSize = 3;
	const trackFootprint = new Map<string, Point>();
	let trackBounds: Bounds | null = null,
		selectedSlot: number | null = null,
		followedSlot: number | null = null,
		lastTrackRender = 0;
	const carHeadings = new Map<number, number>(),
		lapState = new Map<number, { lap: number }>(),
		lapTrace = new Map<number, Point[]>();
	let finishPoint: TrackMarker | null = null,
		sectorPoints: TrackMarker[] = [];
	let zoom = 1,
		viewCenter: Point | null = null,
		mapProjection: Projection | null = null;
	const mapSurface = element<HTMLElement>('mapSurface'),
		mapSVG = element<SVGSVGElement>('map');
	function mapCars(cars: RecoveryCar[]) {
		return cars.filter((c) => c.health !== 0 && c.status === 2);
	}
	function isLocalDriver(c: RecoveryCar) {
		return latest?.localSlot === c.slot;
	}
	function presentedCars(cars: RecoveryCar[]): RecoveryCar[] {
		return mapCars(cars).map((c) =>
			c.incidentId && unrecoverable.has(c.incidentId)
				? { ...c, state: 'out' }
				: isLocalDriver(c)
					? { ...c, state: 'moving' }
					: c,
		);
	}
	function resetMapHistory() {
		trackSamples.clear();
		trackFootprint.clear();
		carHeadings.clear();
		lapState.clear();
		lapTrace.clear();
		finishPoint = null;
		sectorPoints = [];
		trackBounds = null;
		mapProjection = null;
		viewCenter = null;
		cameraReady = false;
		driverCentered = true;
		selectedSlot = null;
		followedSlot = null;
		updateZoomControls(1);
		lastTrackRender = 0;
	}
	function trackHeading(trace: Point[], index: number) {
		const a = trace[Math.max(0, index - 1)],
			b = trace[Math.min(trace.length - 1, index + 1)];
		return Math.atan2(b.z - a.z, b.x - a.x);
	}
	function updateTrackLearning(c: RecoveryCar) {
		if (!Number.isFinite(c.x) || !Number.isFinite(c.z) || !Number.isFinite(c.lap) || c.lap < 1)
			return;
		const old = lapState.get(c.slot);
		const pos = { x: c.x, y: c.y, z: c.z };
		if (!finishPoint && old && c.lap === old.lap + 1) {
			const trace = lapTrace.get(c.slot) || [];
			if (trace.length > 35) {
				const distances = [0];
				for (let i = 1; i < trace.length; i++)
					distances.push(
						distances[i - 1] + Math.hypot(trace[i].x - trace[i - 1].x, trace[i].z - trace[i - 1].z),
					);
				const length = distances.at(-1) ?? 0;
				if (
					length > 100 &&
					Math.hypot(trace[0].x - pos.x, trace[0].z - pos.z) < Math.min(60, length * 0.05)
				) {
					finishPoint = { ...pos, heading: trackHeading(trace, trace.length - 1) };
					sectorPoints = [1 / 3, 2 / 3].map((frac) => {
						const i = distances.findIndex((d) => d >= length * frac);
						return { ...trace[i], heading: trackHeading(trace, i) };
					});
				}
			}
		}
		if (!old || c.lap !== old.lap) lapTrace.set(c.slot, [pos]);
		else {
			const tr = lapTrace.get(c.slot) || [];
			const prev = tr.at(-1);
			if (!prev || Math.hypot(prev.x - pos.x, prev.z - pos.z) > 3) {
				tr.push(pos);
				if (tr.length > 2500) tr.shift();
				lapTrace.set(c.slot, tr);
			}
		}
		lapState.set(c.slot, { lap: c.lap });
	}

	function mapDraw(cars: RecoveryCar[]) {
		const path = element<SVGGElement>('pathLayer'),
			markers = element<SVGGElement>('carLayer');
		markers.replaceChildren();
		if (!cars.length) {
			path.replaceChildren();
			element<SVGGElement>('trackFeatureLayer').replaceChildren();
			mapProjection = null;
			lastTrackRender = 0;
			return;
		}
		const now = performance.now();
		const trackStatuses = new Map(latest?.standings?.map((car) => [car.slot, car.trackStatus]));
		for (const c of cars) {
			if (!Number.isFinite(c.x) || !Number.isFinite(c.z)) continue;
			let h = trackSamples.get(c.slot) || [];
			const last = h[h.length - 1];
			if (last) {
				const dx = c.x - last.x,
					dz = c.z - last.z;
				if (Math.hypot(dx, dz) > 0.7) carHeadings.set(c.slot, (Math.atan2(dx, dz) * 180) / Math.PI);
			}
			if (!last || Math.hypot(c.x - last.x, c.z - last.z) > 1.8) {
				h.push({ x: c.x, y: c.y, z: c.z });
				if (h.length > 220) h.shift();
				trackSamples.set(c.slot, h);
			}
			if (
				trackStatuses.get(c.slot) === 0 &&
				Date.now() - Date.parse(c.updated) <= 2000 &&
				c.speed > 3 &&
				trackFootprint.size < 12000
			) {
				const x = Math.floor(c.x / footprintCellSize);
				const z = Math.floor(c.z / footprintCellSize);
				const key = x + ',' + z;
				if (!trackFootprint.has(key)) trackFootprint.set(key, { x, y: c.y, z });
			}
			if (c.orientation) {
				const [qx, qy, qz, qw] = c.orientation;
				const forwardX = 2 * (qx * qz + qw * qy);
				const forwardZ = 1 - 2 * (qx * qx + qy * qy);
				const forwardY = 2 * (qy * qz - qw * qx);
				if (Math.hypot(forwardX, forwardZ) > 0.1)
					carHeadings.set(
						c.slot,
						(Math.atan2(forwardX, -depth({ x: forwardX, y: forwardY, z: forwardZ })) * 180) /
							Math.PI,
					);
			}
			updateTrackLearning(c);
		}
		const all = [...trackSamples.values()].flat();
		if (!all.length) return;
		const hasAlert = cars.some((c) => c.state === 'alert');
		const xs = all.map((c) => c.x),
			zs = all.map(depth);
		const observed = {
			minX: Math.min(...xs),
			maxX: Math.max(...xs),
			minZ: Math.min(...zs),
			maxZ: Math.max(...zs),
		};
		if (!trackBounds) trackBounds = observed;
		else {
			trackBounds.minX = Math.min(trackBounds.minX, observed.minX);
			trackBounds.maxX = Math.max(trackBounds.maxX, observed.maxX);
			trackBounds.minZ = Math.min(trackBounds.minZ, observed.minZ);
			trackBounds.maxZ = Math.max(trackBounds.maxZ, observed.maxZ);
		}
		const { minX, maxX, minZ, maxZ } = trackBounds;
		const width = mapSVG.clientWidth || 900,
			height = mapSVG.clientHeight || 560;
		mapSVG.setAttribute('viewBox', `0 0 ${width} ${height}`);
		const followed = cars.find((car) => car.slot === followedSlot);
		const local = cars.find(isLocalDriver);
		const tracked = followed ?? (driverUp && driverCentered ? local : undefined);
		const cx = tracked?.x ?? viewCenter?.x ?? (minX + maxX) / 2,
			cz = tracked ? depth(tracked) : (viewCenter?.z ?? (minZ + maxZ) / 2);
		const targetAngle =
			driverUp && local ? (-(carHeadings.get(local.slot) ?? 0) * Math.PI) / 180 : 0;
		if (!cameraReady || !driverUp) {
			cameraAngle = targetAngle;
			cameraReady = true;
		} else {
			const delta = Math.atan2(
				Math.sin(targetAngle - cameraAngle),
				Math.cos(targetAngle - cameraAngle),
			);
			const elapsed = Math.min(0.1, Math.max(0, (now - headingAt) / 1000));
			cameraAngle += delta * (1 - Math.exp(-elapsed / 0.12));
			if (Math.abs(delta) > 0.001 && cameraFrame === null) {
				cameraFrame = requestAnimationFrame(() => {
					cameraFrame = null;
					redrawMap();
				});
			}
		}
		headingAt = now;
		const cos = Math.abs(Math.cos(cameraAngle)),
			sin = Math.abs(Math.sin(cameraAngle));
		// Rotation-invariant fit; stops zoom pumping in corners.
		const diagonal = Math.hypot(maxX - minX, maxZ - minZ);
		const boundsWidth = driverUp ? diagonal : (maxX - minX) * cos + (maxZ - minZ) * sin;
		const boundsHeight = driverUp ? diagonal : (maxX - minX) * sin + (maxZ - minZ) * cos;
		const edgeRoom = Math.min(88, Math.max(48, Math.min(width, height) * 0.13));
		const fitScale = Math.min(
			Math.max(1, width - 2 * edgeRoom) / Math.max(boundsWidth, 80),
			Math.max(1, height - 2 * edgeRoom) / Math.max(boundsHeight, 80),
		);
		const scale = fitScale * zoom;
		mapProjection = { cx, cz, scale, fitScale, width, height, angle: cameraAngle };
		const point = (c: Point): [number, number] => {
			const [sx, sy] = rotate(c.x - cx, depth(c) - cz, cameraAngle);
			return [width / 2 + sx * scale, height / 2 + sy * scale];
		};
		path.setAttribute(
			'transform',
			`translate(${width / 2} ${height / 2}) rotate(${(cameraAngle * 180) / Math.PI}) scale(${scale}) translate(${-cx} ${-cz})`,
		);
		path.setAttribute('opacity', hasAlert ? '.16' : elevated ? '.65' : '.34');
		if (now - lastTrackRender > 180 || lastTrackRender === 0) {
			path.replaceChildren();
			if (trackFootprint.size && latest?.settings.drawTrackOutline) {
				path.append(
					node('path', {
						d: trackOutline(trackFootprint, footprintCellSize, depth),
						fill: 'none',
						stroke: 'var(--map-label)',
						'stroke-width': 1.5,
						'stroke-linecap': 'round',
						'vector-effect': 'non-scaling-stroke',
					}),
				);
			}
			for (const [slot, h] of trackSamples) {
				if (h.length < 2) continue;
				const d = h
					.map((v, i) => `${i ? 'L' : 'M'}${v.x.toFixed(1)},${depth(v).toFixed(1)}`)
					.join(' ');
				path.append(
					node('path', {
						d,
						fill: 'none',
						stroke: ['#cc494f', '#3c8f85', '#b99b47', '#556eaa', '#a66396'][slot % 5],
						'stroke-width': 1.25,
						'vector-effect': 'non-scaling-stroke',
					}),
				);
			}
			lastTrackRender = now;
		}
		const features = element<SVGGElement>('trackFeatureLayer');
		features.replaceChildren();
		const compass = node('g', {
			transform: 'translate(' + (width - 35) + ' 42) rotate(' + (cameraAngle * 180) / Math.PI + ')',
			'pointer-events': 'none',
		});
		compass.append(
			node('title', {}, 'North-up map convention: game +Z points up, +X points right'),
		);
		compass.append(
			node('path', {
				d: 'M0 9 V-10 M-4 -4 L0 -10 L4 -4',
				fill: 'none',
				stroke: 'var(--map-label)',
				'stroke-width': 2,
			}),
		);
		compass.append(
			node(
				'text',
				{
					x: 0,
					y: -18,
					'text-anchor': 'middle',
					fill: 'var(--map-label)',
					'font-size': 12,
					'font-weight': 'bold',
				},
				'N',
			),
		);
		features.append(compass);
		function feature(p: TrackMarker | null, text: string, color: string) {
			if (!p) return;
			const [fx, fy] = point(p);
			const [dx, dy] = rotate(-Math.sin(p.heading) * 13, -Math.cos(p.heading) * 13, cameraAngle);
			features.append(
				node('line', {
					x1: fx - dx,
					y1: fy - dy,
					x2: fx + dx,
					y2: fy + dy,
					stroke: color,
					'stroke-width': 3,
				}),
			);
			features.append(
				node(
					'text',
					{ x: fx + 13, y: fy - 9, fill: color, 'font-size': 12, 'font-weight': 'bold' },
					text,
				),
			);
		}
		feature(finishPoint, 'START / FINISH', '#ff6268');
		sectorPoints.forEach((p, i) => feature(p, `SECTOR ${i + 1} (EST.)`, '#9ca6ad'));
		for (const c of cars) {
			if (!Number.isFinite(c.x) || !Number.isFinite(c.z)) continue;
			const [x, y] = point(c),
				local = isLocalDriver(c),
				medic = !local && medicSlots.has(c.slot),
				special = local || medic,
				color = local
					? '#59b7ff'
					: medic
						? medicFill
						: c.state === 'alert'
							? '#f85e63'
							: c.state === 'caution'
								? '#ffc66e'
								: c.state === 'out'
									? '#606769'
									: '#45d4ac';
			if (c.state === 'alert' || special) {
				markers.append(
					node('circle', {
						cx: x,
						cy: y,
						r: special ? 28 : 21,
						fill: color,
						opacity: special ? 0.28 : 0.2,
					}),
				);
			}
			if (c.slot === followedSlot) {
				markers.append(
					node('circle', { cx: x, cy: y, r: 25, fill: 'none', class: 'followed-driver-ring' }),
				);
			}
			const marker = node('path', {
				d: elevated
					? wireframe(c, x, y)
					: special
						? 'M-8 -13 Q-8 -16 -5 -16 H5 Q8 -16 8 -13 V12 Q8 15 5 15 H-5 Q-8 15 -8 12 Z M-9 -7 H9 M-9 8 H9 M-11 -9 V-3 M11 -9 V-3 M-11 7 V12 M11 7 V12'
						: 'M0 -11 L-7 8 L0 4 L7 8 Z',
				transform: elevated
					? ''
					: `translate(${x} ${y}) rotate(${(carHeadings.get(c.slot) || 0) + (cameraAngle * 180) / Math.PI})`,
				fill: elevated ? 'none' : color,
				stroke: elevated ? color : '#fff',
				'stroke-width':
					special || c.slot === selectedSlot || c.slot === followedSlot ? 2.3 : 1,
				'stroke-linejoin': 'round',
			});
			const opacity =
				hasAlert &&
				!special &&
				c.state !== 'alert' &&
				c.slot !== selectedSlot &&
				c.slot !== followedSlot
					? 0.28
					: 1;
			marker.setAttribute('opacity', String(opacity));
			markers.append(marker);
			if (elevated || special) {
				const dot = node('path', {
					d: elevated
						? wireframe(c, x, y, true)
						: 'M-2.5 -11a2.5,2.5 0 1,0 5,0a2.5,2.5 0 1,0 -5,0Z',
					transform: elevated
						? ''
						: 'translate(' +
							x +
							' ' +
							y +
							') rotate(' +
							((carHeadings.get(c.slot) || 0) + (cameraAngle * 180) / Math.PI) +
							')',
					fill: '#fff',
					stroke: '#163a50',
					'stroke-width': 1,
					'pointer-events': 'none',
					opacity,
				});
				markers.append(dot);
			}
			if (special) {
				const cross = node('g', {
					transform: elevated
						? 'translate(' + (x - 24) + ' ' + (y - 24) + ')'
						: 'translate(' +
							x +
							' ' +
							y +
							') rotate(' +
							((carHeadings.get(c.slot) || 0) + (cameraAngle * 180) / Math.PI) +
							')',
					'pointer-events': 'none',
				});
				cross.append(
					node('path', {
						d: 'M-4 0 H4 M0 -4 V4',
						fill: 'none',
						stroke: '#fff',
						'stroke-width': 2.5,
					}),
				);
				if (elevated)
					cross.prepend(
						node('circle', {
							r: 8,
							fill: local ? '#1467a5' : medicElevated,
							stroke: '#fff',
						}),
					);
				markers.append(cross);
			}
			const nearRight = x > width - 110;
			const label = node(
				'text',
				{
					x: x + (nearRight ? (special ? -20 : -10) : special ? 20 : 10),
					y: y + 4,
					'text-anchor': nearRight ? 'end' : 'start',
					fill: local ? '#83ceff' : medic ? medicLabel : 'var(--map-label)',
					'font-size': special ? 12 : 11,
					'font-weight': special ? 'bold' : 'normal',
					'paint-order': 'stroke',
					stroke: 'var(--map-canvas)',
					'stroke-width': 3,
					'font-family': 'Segoe UI,Arial',
					opacity,
				},
				`${c.place || '-'} ${c.name}${local ? ' (You)' : medic ? ' (Medic)' : ''}`,
			);
			markers.append(label);
			marker.append(
				node(
					'title',
					{},
					`${c.name} | ${displaySpeed(c.speed, speedUnit)} ${speedUnit} | ${c.health}% health | ${c.state}${medic ? ' | Medic' : ''}`,
				),
			);
			marker.addEventListener('click', () => {
				selectedSlot = c.slot;
				redrawMap();
			});
		}
	}

	function redrawMap() {
		mapDraw(presentedCars(latest?.cars || []));
	}
	function updateZoomControls(value: number) {
		zoom = value;
		zoomFit.disabled =
			zoom === 1 && !viewCenter && followedSlot === null && !(driverUp && driverCentered);
	}
	function pointerPoint(event: MouseEvent) {
		const bounds = mapSVG.getBoundingClientRect();
		return { x: event.clientX - bounds.left, y: event.clientY - bounds.top };
	}
	function worldPoint(point: ScreenPoint, projection: Projection) {
		const [x, z] = rotate(
			(point.x - projection.width / 2) / projection.scale,
			(point.y - projection.height / 2) / projection.scale,
			-projection.angle,
		);
		return { x: projection.cx + x, z: projection.cz + z };
	}
	function placeAnchor(anchor: Point, point: ScreenPoint, value: number, projection: Projection) {
		const scale = projection.fitScale * value;
		if (!Number.isFinite(scale) || scale <= 0) return;
		const [dx, dz] = rotate(
			(point.x - projection.width / 2) / scale,
			(point.y - projection.height / 2) / scale,
			-projection.angle,
		);
		const x = anchor.x - dx;
		const z = anchor.z - dz;
		if (!Number.isFinite(x) || !Number.isFinite(z)) return;
		viewCenter = { x, z };
		updateZoomControls(value);
	}
	function zoomAt(value: number, point: ScreenPoint) {
		const projection = mapProjection;
		if (!projection) return;
		const anchor = worldPoint(point, projection);
		if (driverUp && driverCentered && followedSlot === null) {
			updateZoomControls(value);
			redrawMap();
			return;
		}
		placeAnchor(anchor, point, value, projection);
		redrawMap();
	}
	function zoomBy(factor: number) {
		if (mapProjection)
			zoomAt(zoom * factor, { x: mapProjection.width / 2, y: mapProjection.height / 2 });
	}
	zoomIn.onclick = () => zoomBy(1.35);
	zoomOut.onclick = () => zoomBy(1 / 1.35);
	function fitTrack() {
		driverCentered = false;
		viewCenter = null;
		updateZoomControls(1);
		redrawMap();
	}
	zoomFit.onclick = () => {
		if (followedSlot !== null) {
			followedSlot = null;
			onUnfollow();
		}
		fitTrack();
	};
	updateZoomControls(1);

	const pointers = new Map<number, ScreenPoint>();
	let gesture: Gesture | null = null,
		gestureDragged = false,
		mapRedrawPending = false;
	function queueMapRedraw() {
		if (mapRedrawPending) return;
		mapRedrawPending = true;
		requestAnimationFrame(() => {
			mapRedrawPending = false;
			redrawMap();
		});
	}
	function gestureGeometry() {
		const [a, b] = pointers.values();
		return b
			? {
					point: { x: (a.x + b.x) / 2, y: (a.y + b.y) / 2 },
					distance: Math.hypot(a.x - b.x, a.y - b.y),
				}
			: { point: a, distance: 0 };
	}
	function beginGesture() {
		if (!pointers.size || !mapProjection) {
			gesture = null;
			return;
		}
		const { point, distance } = gestureGeometry();
		gesture = {
			point,
			distance,
			zoom,
			anchor: worldPoint(point, mapProjection),
			projection: mapProjection,
		};
	}
	function expandedMap() {
		return (
			document.fullscreenElement === mapSurface || mapSurface.classList.contains('map-expanded')
		);
	}
	mapSVG.addEventListener('pointerdown', (event) => {
		if (
			(!expandedMap() && event.pointerType !== 'mouse') ||
			!mapProjection ||
			event.button !== 0 ||
			pointers.size >= 2
		)
			return;
		if (!pointers.size) gestureDragged = false;
		pointers.set(event.pointerId, pointerPoint(event));
		beginGesture();
	});
	window.addEventListener(
		'pointermove',
		(event) => {
			if (!pointers.has(event.pointerId) || !gesture) return;
			pointers.set(event.pointerId, pointerPoint(event));
			const { point, distance } = gestureGeometry();
			if (
				!gestureDragged &&
				pointers.size === 1 &&
				Math.hypot(point.x - gesture.point.x, point.y - gesture.point.y) < 5
			)
				return;
			gestureDragged = true;
			driverCentered = false;
			if (followedSlot !== null) {
				followedSlot = null;
				onUnfollow();
			}
			const factor = gesture.distance > 0 ? distance / gesture.distance : 1;
			placeAnchor(gesture.anchor, point, gesture.zoom * factor, gesture.projection);
			queueMapRedraw();
		},
		{ signal },
	);
	function endGesture(event: PointerEvent) {
		if (!pointers.delete(event.pointerId)) return;
		if (pointers.size && mapProjection) {
			redrawMap();
			beginGesture();
		} else gesture = null;
	}
	window.addEventListener('pointerup', endGesture, { signal });
	window.addEventListener('pointercancel', endGesture, { signal });
	mapSVG.addEventListener(
		'click',
		(event) => {
			if (gestureDragged) {
				event.stopPropagation();
				gestureDragged = false;
			}
		},
		true,
	);
	mapSVG.addEventListener(
		'wheel',
		(event) => {
			if (!expandedMap() || !mapProjection) return;
			event.preventDefault();
			zoomAt(zoom * (event.deltaY < 0 ? 1.2 : 1 / 1.2), pointerPoint(event));
		},
		{ passive: false },
	);
	const fullscreenToggle = element<HTMLButtonElement>('mapTrueFullscreen');
	mapFull.onclick = () => {
		mapSurface.classList.add('map-expanded');
		document.body.classList.add('map-open');
		mapExit.focus();
		redrawMap();
	};
	fullscreenToggle.disabled = !document.fullscreenEnabled;
	fullscreenToggle.onclick = async () => {
		try {
			if (document.fullscreenElement === mapSurface) await document.exitFullscreen();
			else await mapSurface.requestFullscreen();
		} catch {
			fullscreenToggle.disabled = true;
		}
	};
	async function exitMap() {
		if (document.fullscreenElement === mapSurface) await document.exitFullscreen();
		mapSurface.classList.remove('map-expanded');
		document.body.classList.remove('map-open');
		mapFull.focus();
		redrawMap();
	}
	mapExit.onclick = exitMap;
	document.addEventListener(
		'fullscreenchange',
		() => {
			fullscreenToggle.setAttribute(
				'aria-pressed',
				String(document.fullscreenElement === mapSurface),
			);
			redrawMap();
		},
		{ signal },
	);
	document.addEventListener(
		'keydown',
		(event) => {
			if (
				event.key === 'Escape' &&
				mapSurface.classList.contains('map-expanded') &&
				!document.querySelector('[data-scope=dialog][data-state=open][data-part=content]')
			)
				exitMap();
		},
		{ signal },
	);
	const observer = new ResizeObserver(redrawMap);
	observer.observe(mapSVG);

	return {
		update(snapshot: TelemetrySnapshot, dismissed: ReadonlySet<string>, unit: SpeedUnit = 'km/h') {
			speedUnit = unit;
			if (latest && (latest.mode !== snapshot.mode || latest.raceEpoch !== snapshot.raceEpoch))
				resetMapHistory();
			if (latest?.settings.drawTrackOutline !== snapshot.settings.drawTrackOutline)
				lastTrackRender = 0;
			latest = snapshot;
			unrecoverable = dismissed;
			medicSlots = new Set(snapshot.medics ?? []);
			redrawMap();
		},
		follow(slot: number | null) {
			followedSlot = slot;
			updateZoomControls(zoom);
			if (slot === null) fitTrack();
			else redrawMap();
		},
		setDriverUp(value: boolean) {
			driverUp = value;
			driverCentered = value;
			cameraReady = false;
			viewCenter = null;
			updateZoomControls(zoom);
			redrawMap();
		},
		setElevated(value: boolean) {
			if (elevated === value) return;
			elevated = value;
			trackBounds = null;
			lastTrackRender = 0;
			viewCenter = null;
			driverCentered = driverUp;
			cameraReady = false;
			updateZoomControls(1);
			redrawMap();
		},
		focusIncident(slot: number) {
			const cars = mapCars(latest?.cars ?? []);
			const target = cars.find((car) => car.slot === slot);
			if (!target || !mapProjection) return;
			const local = cars.find(isLocalDriver) ?? target;
			const { width, height, fitScale } = mapProjection;
			const [spanX, spanZ] = rotate(target.x - local.x, depth(target) - depth(local), cameraAngle);
			followedSlot = null;
			onUnfollow();
			selectedSlot = slot;
			driverCentered = false;
			viewCenter = { x: (target.x + local.x) / 2, z: (depth(target) + depth(local)) / 2 };
			updateZoomControls(
				Math.min(
					(width * 0.65) / Math.max(Math.abs(spanX), 1) / fitScale,
					(height * 0.55) / Math.max(Math.abs(spanZ), 1) / fitScale,
				),
			);
			redrawMap();
		},
		destroy() {
			if (cameraFrame !== null) cancelAnimationFrame(cameraFrame);
			listeners.abort();
			observer.disconnect();
			document.body.classList.remove('map-open');
		},
	};
}
