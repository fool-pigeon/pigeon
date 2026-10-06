<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

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
