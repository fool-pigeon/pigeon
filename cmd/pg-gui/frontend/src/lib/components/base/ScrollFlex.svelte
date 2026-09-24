<script lang="ts">
	import type { Snippet } from 'svelte';
	type Direction = 'horizontal' | 'vertical';
	let { content = $bindable<Snippet>(), enableFlexCol = $bindable(false) } = $props();
</script>

<div
	class="flex {enableFlexCol ? 'flex-col' : 'flex-row'} overflow-auto custom-scrollbar"
	onwheel={(e) => {
		e.preventDefault();
		e.currentTarget.scrollLeft += e.deltaY;
	}}
>
	{@render content()}
</div>
