<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import type { Snippet } from 'svelte';
	type Prop = { content: Snippet; noBorder?: boolean; noPadding?: boolean };
	let { content, noBorder = false, noPadding = false }: Prop = $props();
</script>

<div
	class="w-full flex flex-col
	*:max-w-full {noPadding ? '' : '*:p-1'} {noBorder ? '' : '*:border-b *:border-base-content/20'}
	*:flex *:justify-between @container *:@max-xs:flex-col"
>
	{@render content()}
</div>
