<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import type { Snippet } from 'svelte';

	let { content, actived }: { content: Snippet; actived: boolean } = $props();
</script>

<div
	class="[&>button]:btn [&>button]:outline-none [&>button]:border-none
	[&>button]:p-0.5 [&>button]:pl-2 [&>button]:pr-2
	[&>button]:h-auto
	{actived
		? '[&>button]:btn-primary'
		: '[&>button]:btn-ghost [&>button]:text-base-content/50 [&>button]:hover:text-base-content'}"
>
	{@render content()}
</div>
