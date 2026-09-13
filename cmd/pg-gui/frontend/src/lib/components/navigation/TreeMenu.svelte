<script lang="ts" module>
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	type ItemHandlers = Pick<
		HTMLAttributes<HTMLElement>,
		'ontoggle' | 'oncontextmenu' | 'onclick' | 'onauxclick'
	>;
	type TreeItemType = 'leaf' | 'group';
	export type TreeItem = {
		id: string;
		name: string;
		type: TreeItemType;
		isOpen: boolean;
		next?: TreeItem[];
		snippet?: Snippet<[TreeItem]>;
		on?: ItemHandlers;
		class?: string;
		style?: string;
	};
	// Create a group node in menu,it can be expanded like folder
	export function NewTreeGroup(
		id: string,
		name: string,
		option?: {
			isOpen?: boolean;
			next?: TreeItem[];
			snippet?: Snippet<[TreeItem]>;
			on?: ItemHandlers;
			class?: string;
			style?: string;
		},
	): TreeItem {
		return { id, name, type: 'group', isOpen: option?.isOpen ?? false, ...option };
	}
	// Create a single node in menu,like file in FS
	export function NewTreeLeaf(
		id: string,
		name: string,
		option?: {
			snippet?: Snippet<[TreeItem]>;
			on?: ItemHandlers;
			class?: string;
			style?: string;
		},
	): TreeItem {
		return { id, name, type: 'leaf', isOpen: false, ...option };
	}
	export type IndentGuideType = 'none' | 'hover' | 'always';
</script>

<script lang="ts">
	let {
		items = $bindable<TreeItem[]>([] as TreeItem[]),
		indentSize = $bindable<number>(5),
		indentGuide = $bindable<IndentGuideType>('always'),
	} = $props();
	let indentGuidesClass = $derived(
		indentGuide === 'none'
			? ''
			: indentGuide === 'hover'
				? 'before:absolute before:inset-y-0 before:w-px before:bg-transparent hover:before:bg-base-content/20'
				: 'before:absolute before:inset-y-0 before:w-px before:bg-base-content/20',
	);
	const indentScale = 0.25;
</script>

<ul class="w-full">
	{@render tree(items, 0)}
</ul>

{#snippet tree(items: TreeItem[], offset: number)}
	{#each items as item (item.id)}
		<li>
			{#if item.type === 'group'}
				<details bind:open={item.isOpen} ontoggle={item.on?.ontoggle}>
					<summary
						class="flex items-center {item.class}"
						style={item.style}
						style:padding-left="{offset}rem"
						oncontextmenu={item.on?.oncontextmenu}
					>
						<span
							class="inline-flex transition-transform"
							class:rotate-90={item.isOpen}
							style="width: {indentSize * indentScale}rem;height: {indentSize * indentScale}rem;"
						>
							<ChevronRight
								class="text-base-content"
								strokeWidth={1}
								nonScalingStroke
								style="width: {indentSize * indentScale}rem;height: {indentSize * indentScale}rem;"
							/>
						</span>
						{#if item.snippet}
							{@render item.snippet(item)}
						{:else}
							{item.name}
						{/if}
					</summary>
					{#if item.next}
						<ul
							class="relative {indentGuidesClass}"
							style:--guide="{offset + indentSize * indentScale * 0.5}rem"
						>
							{@render tree(item.next, offset + indentSize * indentScale)}
						</ul>
					{/if}
				</details>
			{:else}
				<button
					class="w-full text-left {item.class}"
					style={item.style}
					style:padding-left="{offset}rem"
					onclick={item.on?.onclick}
					onauxclick={item.on?.onauxclick}
					oncontextmenu={item.on?.oncontextmenu}
				>
					<div>
						{#if item.snippet}
							{@render item.snippet(item)}
						{:else}
							{item.name}
						{/if}
					</div>
				</button>
			{/if}
		</li>
	{/each}
{/snippet}

<style>
	.relative::before {
		left: var(--guide, 0);
	}
</style>
