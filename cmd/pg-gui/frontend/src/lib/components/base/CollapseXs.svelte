<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title = $bindable(),
		content = $bindable(),
		checked,
	}: { title: string; content: Snippet; checked?: boolean } = $props();
</script>

<div class="collapse collapse-arrow p-1 border border-base-content/10">
	<input class="checkbox checkbox-xs" type="checkbox" {checked} />
	<div class="collapse-title p-0 font-bold">{title}</div>
	<div class="collapse-content p-0">
		{@render content()}
	</div>
</div>
