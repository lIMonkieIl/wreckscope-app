type Sample = { x: number; y?: number; z: number };
type Edge = { from: Sample; to: Sample; used: boolean };

export function trackOutline(
	cells: ReadonlyMap<string, Sample>,
	cellSize: number,
	projectDepth: (point: Sample) => number,
): string {
	const edges: Edge[] = [];
	const outgoing = new Map<string, Edge[]>();
	const key = (point: Sample) => point.x + ',' + point.z;
	function add(from: Sample, to: Sample) {
		const edge = { from, to, used: false };
		edges.push(edge);
		const list = outgoing.get(key(from)) ?? [];
		list.push(edge);
		outgoing.set(key(from), list);
	}
	for (const { x, y, z } of cells.values()) {
		const a = { x, y, z },
			b = { x: x + 1, y, z };
		const c = { x: x + 1, y, z: z + 1 },
			d = { x, y, z: z + 1 };
		if (!cells.has(x + ',' + (z - 1))) add(a, b);
		if (!cells.has(x + 1 + ',' + z)) add(b, c);
		if (!cells.has(x + ',' + (z + 1))) add(c, d);
		if (!cells.has(x - 1 + ',' + z)) add(d, a);
	}
	const paths: string[] = [];
	for (const first of edges) {
		if (first.used) continue;
		const loop: Sample[] = [];
		let edge: Edge | undefined = first;
		while (edge && !edge.used) {
			edge.used = true;
			loop.push(edge.from);
			const dx: number = edge.to.x - edge.from.x,
				dz: number = edge.to.z - edge.from.z;
			const choices: Edge[] = (outgoing.get(key(edge.to)) ?? []).filter(
				(candidate) => !candidate.used,
			);
			// Prefer right turn at diagonal joins so loops stay separate.
			edge =
				choices.find(
					(candidate) =>
						candidate.to.x - candidate.from.x === -dz && candidate.to.z - candidate.from.z === dx,
				) ?? choices[0];
		}
		if (loop.length < 4) continue;
		const points = loop.map((point) => ({
			x: point.x * cellSize,
			z: projectDepth({ x: point.x * cellSize, y: point.y, z: point.z * cellSize }),
		}));
		const middle = (a: { x: number; z: number }, b: { x: number; z: number }) =>
			((a.x + b.x) / 2).toFixed(2) + ',' + ((a.z + b.z) / 2).toFixed(2);
		let path = 'M' + middle(points.at(-1)!, points[0]);
		for (let i = 0; i < points.length; i++) {
			const point = points[i];
			path +=
				'Q' +
				point.x.toFixed(2) +
				',' +
				point.z.toFixed(2) +
				' ' +
				middle(point, points[(i + 1) % points.length]);
		}
		paths.push(path + 'Z');
	}
	return paths.join('');
}
