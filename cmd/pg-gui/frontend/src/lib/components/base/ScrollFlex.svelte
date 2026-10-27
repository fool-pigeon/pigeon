<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import type { Snippet } from 'svelte';
	let { content = $bindable<Snippet>(), enableFlexCol = $bindable(false) } = $props();
	let el: HTMLDivElement;
</script>

<div
	bind:this={el}
	class="flex {enableFlexCol ? 'flex-col' : 'flex-row'} overflow-auto custom-scrollbar"
	onwheel={(e) => {
		if (e.currentTarget !== el) return;
		e.preventDefault();
		const delta = e.deltaY !== 0 ? e.deltaY : e.deltaX;
		if (enableFlexCol) {
			e.currentTarget.scrollTop += delta;
		} else {
			e.currentTarget.scrollLeft += delta;
		}
	}}
>
	{@render content()}
</div>
