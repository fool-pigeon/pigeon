<script lang="ts">
	import type { Snippet } from 'svelte';

	let { content = $bindable<Snippet>() } = $props();
</script>

<div
	class="pb-1 flex overflow-auto
			scrollbar-gutter-auto scrollbar-thin"
	onwheel={(e) => {
		e.preventDefault();
		e.currentTarget.scrollLeft += e.deltaY;
	}}
>
	{@render content()}
</div>
