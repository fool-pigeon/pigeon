<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import type { Snippet } from 'svelte';

	let { content }: { content: Snippet } = $props();
</script>

<div
	class="w-full flex flex-col
	*:max-w-full *:p-1 *:border-b *:border-base-content/20
	*:flex *:justify-between @container *:@max-md:flex-col"
>
	{@render content()}
</div>
