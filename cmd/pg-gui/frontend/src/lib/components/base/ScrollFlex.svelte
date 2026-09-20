<script lang="ts">
	import type { Snippet } from 'svelte';
	type Direction = 'horizontal' | 'vertical';
	let { content = $bindable<Snippet>(), direction = $bindable<Direction>('horizontal') } = $props();
	let isHorizontal = $derived(direction === 'horizontal');
</script>

<div
	class="flex {isHorizontal ? 'flex-row' : 'flex-col'} overflow-auto custom-scrollbar"
	onwheel={(e) => {
		e.preventDefault();
		e.currentTarget.scrollLeft += e.deltaY;
	}}
>
	{@render content()}
</div>
