<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts" module>
	export { FieldHeader };
</script>

<script lang="ts">
	import type { Snippet } from 'svelte';
	type Prop = {
		content: Snippet;
		noBorder?: boolean;
		noPadding?: boolean;
		enableFieldItemsCenter?: boolean;
	};
	let {
		content,
		noBorder = false,
		noPadding = false,
		enableFieldItemsCenter = false,
	}: Prop = $props();
</script>

{#snippet FieldHeader(title: string, description: string)}
	<span class="flex flex-col">
		<span class="text-base-content/70 text-xs">{title}</span>
		<span class="text-base-content/40 text-[0.6875rem] italic">
			{description}
		</span>
	</span>
{/snippet}

<div
	class="w-full flex flex-col
	*:max-w-full {noPadding ? '' : '*:p-1'} {noBorder ? '' : '*:border-b *:border-base-content/20'}
	*:flex *:justify-between
	{enableFieldItemsCenter ? '*:items-center' : ''} 
	@container *:@max-xs:flex-col"
>
	{@render content()}
</div>
