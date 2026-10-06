<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import type { Snippet } from 'svelte';
	let { content = $bindable<Snippet>(), enableFlexCol = $bindable(false) } = $props();
</script>

<div
	class="flex {enableFlexCol ? 'flex-col' : 'flex-row'} overflow-auto custom-scrollbar"
	onwheel={(e) => {
		if (enableFlexCol) return;
		e.preventDefault();
		e.currentTarget.scrollLeft += e.deltaY;
	}}
>
	{@render content()}
</div>
