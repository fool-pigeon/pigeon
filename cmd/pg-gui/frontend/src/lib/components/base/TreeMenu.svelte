<script lang="ts" module>
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	type ItemHandlers = Pick<
		HTMLAttributes<HTMLElement>,
		'ontoggle' | 'oncontextmenu' | 'onclick' | 'onauxclick'
	>;
	type TreeItemType = 'leaf' | 'group';
	/**
	 * @type type: 'leaf'|'group'
	 * @type on:bind oncontextmenu or ontoggle for 'group'.bind onclick,onauxclick or oncontextmenu for 'leaf'
	 */
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
	/**
	 * Create a group node in menu,it can be expanded like folder
	 * @param id unique string for id
	 * @param name the label you can see
	 * @param option the optional options
	 * @param option.on bind oncontextmenu or ontoggle
	 * @param option.snippet the optional render to replace original
	 */
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
	/**
	 * Create a single node in menu,like file in FS
	 * @param id
	 * @param name
	 * @param option the optional options
	 * @param option.on bind onclick or onauxclick or onmousecontextmenu
	 * @param option.snippet the optional render to replace original
	 */
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
	const indentScale = 0.25;
	type Props = {
		items: TreeItem[];
		// the size is for tailwindcss which real is {indentScale*indentSize} rem
		indentSize: number;
		indentGuide: IndentGuideType;
	};
	let {
		items = $bindable(),
		indentSize = $bindable(5),
		indentGuide = $bindable('always'),
	}: Props = $props();
	// the class is style for tailwindcss
	let indentGuidesClass = $derived(
		indentGuide === 'none'
			? ''
			: indentGuide === 'hover'
				? 'before:absolute before:inset-y-0 before:w-px before:bg-transparent hover:before:bg-base-content/20'
				: 'before:absolute before:inset-y-0 before:w-px before:bg-base-content/20',
	);
</script>

<ul class="w-full">
	{@render tree(items, 0)}
</ul>
<!-- offset is the width of indent -->
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
