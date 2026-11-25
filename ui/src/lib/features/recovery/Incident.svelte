<script lang="ts">
	import { displaySpeed } from './speed';
	import type { RecoveryCar, SpeedUnit } from './telemetry';
	let {
		car,
		select,
		dismiss,
		speedUnit,
	}: { car: RecoveryCar; select: () => void; dismiss: () => void; speedUnit: SpeedUnit } = $props();
	let checking = $derived(car.reason === 'Checking sustained recovery');
	let urgent = $derived(car.state === 'alert');
	let duration = $derived(
		Math.max(
			0,
			Math.floor((Date.parse(car.updated) - Date.parse(car.incidentSince ?? car.updated)) / 1000),
		) || 0,
	);
	let title = $derived(
		checking
			? 'Moving again · confirming'
			: /overturn|rollover|upside/i.test(car.reason)
				? 'On its roof'
				: /damage|deceleration/i.test(car.reason)
					? 'Crashed and stopped'
					: /orientation|sideways|tilt/i.test(car.reason)
						? 'Possibly stuck on its side'
						: 'Not moving',
	);
</script>

<div class="incident-entry card" class:urgent>
	<button
		type="button"
		class="incident-brief"
		aria-label={'Select ' + car.name + ' on the map'}
		onclick={select}
	>
		<span class="incident-heading"
			><strong>{car.name}</strong><span class="brief-status"
				>{checking ? 'Checking recovery' : urgent ? 'Needs help' : 'Check car'}</span
			></span
		>
		<span>{title} · {duration}s active</span>
		<span class="incident-meta"
			>{car.place ? 'P' + car.place : 'Car #' + (car.slot + 1)}{car.lap ? ' · Lap ' + car.lap : ''} ·
			{displaySpeed(car.speed, speedUnit)}
			{speedUnit} · {car.health}% health</span
		>
	</button>
	{#if car.incidentId}<button
			type="button"
			class="brief-dismiss"
			aria-label={'Mark ' + car.name + ' unrecoverable'}
			onclick={dismiss}>Unrecoverable</button
		>{/if}
</div>

<style>
	.incident-entry {
		display: grid;
		gap: var(--spacing);
		padding: calc(var(--spacing) * 3);
		--severity: var(--color-warning-500);
		--glass-surface: light-dark(var(--color-surface-50), var(--color-surface-950));
		border: var(--default-border-width) solid color-mix(in srgb, var(--severity) 45%, transparent);
		border-inline-start: 3px solid var(--severity);
		background: color-mix(
			in srgb,
			var(--severity) 12%,
			color-mix(in srgb, var(--glass-surface) 94%, transparent)
		);
		backdrop-filter: blur(10px);
		box-shadow: 0 4px 16px #0002;
	}
	.incident-entry.urgent {
		--severity: var(--color-error-500);
	}
	.incident-heading {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: calc(var(--spacing) * 2);
	}
	.incident-heading {
		flex-wrap: wrap;
		overflow-wrap: anywhere;
	}
	.incident-meta {
		color: light-dark(var(--color-surface-700), var(--color-surface-300));
	}
	.incident-entry {
		grid-template-columns: minmax(0, 1fr) auto;
		align-items: stretch;
		border-inline-start-width: var(--default-border-width);
		border-radius: min(var(--radius-container), 0.45rem);
		gap: 0;
		padding: 0;
		overflow: hidden;
	}
	.incident-brief {
		display: grid;
		gap: 2px;
		min-width: 0;
		padding: calc(var(--spacing) * 2) calc(var(--spacing) * 2.5);
		text-align: start;
		font-size: 0.75rem;
		line-height: 1.35;
		overflow-wrap: anywhere;
		cursor: pointer;
	}
	.brief-status {
		color: light-dark(var(--color-warning-800), var(--color-warning-300));
		font-size: 0.65rem;
		font-weight: 750;
		text-transform: uppercase;
	}
	.urgent .brief-status {
		color: light-dark(var(--color-error-800), var(--color-error-300));
	}
	.brief-dismiss {
		min-height: 44px;
		padding-inline: calc(var(--spacing) * 2);
		color: light-dark(var(--color-surface-950), var(--color-surface-50));
		border-inline-start: var(--default-border-width) solid
			color-mix(in srgb, var(--severity) 45%, transparent);
		background: color-mix(in srgb, var(--severity) 14%, transparent);
		font-size: 0.7rem;
		font-weight: 700;
		cursor: pointer;
	}
	.brief-dismiss:hover {
		background: color-mix(in srgb, var(--severity) 24%, transparent);
	}
	:is(.incident-brief, .brief-dismiss):focus-visible {
		outline: 2px solid var(--severity);
		outline-offset: -2px;
	}
	@media (max-width: 40rem) {
		.brief-dismiss {
			padding-inline: calc(var(--spacing) * 1.5);
		}
	}
</style>
