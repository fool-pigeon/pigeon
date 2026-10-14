<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts" generics="T">
	import { createVirtualizer, measureElement } from '@tanstack/svelte-virtual';
	import type { Snippet } from 'svelte';
	import { get } from 'svelte/store';
	type Key = number | string | bigint;
	type Prop = {
		list: T[];
		itemRender: Snippet<[item: T, i: number]>;
		getKey?: (item: T) => Key;
		estimateSize?: number;
		overscan?: number;
	};
	let { list, itemRender, getKey, estimateSize, overscan }: Prop = $props();

	const getItemKey = (() => getKey)();
	let virtualListEl: HTMLDivElement;
	let virtualItemEls: HTMLDivElement[] = $state<HTMLDivElement[]>([]);
	const virtualizer = createVirtualizer({
		count: (() => list.length)(),
		getScrollElement: () => virtualListEl,
		estimateSize: () => estimateSize ?? 45,
		getItemKey: getItemKey ? (i) => getItemKey(list[i]) : undefined,
		overscan: (() => overscan ?? 5)(),
	});
	let items = $derived($virtualizer.getVirtualItems());
	$effect(() => {
		get(virtualizer).setOptions({
			count: list.length,
			getScrollElement: () => virtualListEl,
		});
	});
	$effect(() => {
		if (virtualItemEls.length) {
			virtualItemEls.forEach((el) => {
				$virtualizer.measureElement(el);
			});
		}
	});
</script>

<div class="w-full h-full min-h-0 overflow-auto contain-strict" bind:this={virtualListEl}>
	<div class="relative w-full" style:height="{$virtualizer.getTotalSize()}px">
		{#each items as row, idx (row.key)}
			<div
				bind:this={virtualItemEls[idx]}
				data-index={row.index}
				class="absolute top-0 left-0 w-full"
				style="transform: translateY({row.start}px);"
			>
				{@render itemRender(list[row.index], row.index)}
			</div>
		{/each}
	</div>
</div>
