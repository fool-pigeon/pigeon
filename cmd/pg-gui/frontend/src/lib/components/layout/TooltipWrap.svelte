<script lang="ts">
	import type { Snippet } from 'svelte';

	let { tool = $bindable<Snippet>(), tip = $bindable<Snippet>(), noHoverHidden = false } = $props();
</script>

<div class="group relative">
	{@render tool()}
	<div
		class="absolute top-full left-0 hidden group-hover:block {noHoverHidden ? '' : 'hover:hidden'}"
	>
		{@render tip()}
	</div>
</div>
