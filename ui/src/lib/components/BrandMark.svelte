<script lang="ts">
	/** Inline SVG mark; fills follow theme tokens. */
	let {
		size = 44,
		class: className = '',
	}: {
		size?: number;
		class?: string;
	} = $props();
</script>

<svg
	class="brand-mark {className}"
	width={size}
	height={size}
	viewBox="0 0 1248 1248"
	role="img"
	aria-hidden="true"
	focusable="false"
>
	<title>WreckScope</title>
	<!-- Scope → secondary; W → primary. -->
	<g class="brand-mark-scope" fill="var(--color-secondary-500)">
		<path
			d="M 499,221 C 397,254 311,320 258,411 L 319,427 C 362,359 425,309 499,280 Z M 499,1021 C 350,970 243,846 215,692 L 274,692 C 300,809 383,909 499,956 Z"
		/>
		<path
			d="M 499,221 C 397,254 311,320 258,411 L 319,427 C 362,359 425,309 499,280 Z M 499,1021 C 350,970 243,846 215,692 L 274,692 C 300,809 383,909 499,956 Z"
			transform="translate(1248 0) scale(-1 1)"
		/>
		<path d="M606 146H642V377H606Z M606 880H642V1103H606Z" />
	</g>
	<g class="brand-mark-monogram" fill="var(--color-primary-500)">
		<path
			d="M202 423L362 493L474 692L624 494L774 692L886 493L1046 423L779 855L624 631L469 855Z M520 830L624 685L728 830L624 759Z"
		/>
		<path
			d="M160 606H287V635H160Z M961 606H1088V635H961Z M542 235H564V309H542Z M684 235H706V309H684Z M542 947H564V1020H542Z M684 947H706V1020H684Z"
		/>
	</g>
</svg>

<style>
	.brand-mark {
		display: block;
		flex-shrink: 0;
	}
</style>
