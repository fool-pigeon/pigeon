<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

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
