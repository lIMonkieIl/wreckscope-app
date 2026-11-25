<script lang="ts">
	/** Wreck / Scope halves follow primary / secondary. */
	let { class: className = '' }: { class?: string } = $props();
</script>

<strong class="brand-wordmark {className}" aria-label="WreckScope"
	><span class="brand-wreck">Wreck</span><span class="brand-scope">Scope</span></strong
>

<style>
	.brand-wordmark {
		display: block;
		margin: 0;
		font-size: inherit;
		font-weight: 800;
		letter-spacing: inherit;
		text-transform: uppercase;
		line-height: 1.1;
	}
	/* Dark uses 500s; 300s wash both halves toward the same cyan. */
	.brand-wreck {
		color: light-dark(var(--color-primary-700), var(--color-primary-500));
	}
	.brand-scope {
		color: light-dark(var(--color-secondary-800), var(--color-secondary-500));
	}
</style>
