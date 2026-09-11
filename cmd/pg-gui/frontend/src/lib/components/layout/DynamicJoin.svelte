<script lang="ts" module>
	export type Panel = {
		snippet: Snippet;
		// 0-100, which means the 0-100%,negative means need auto resize
		size: number;
		// 0-100, which means the 0-100%
		minSize?: number;
	};
	export function NewPanel(snippet: Snippet, size: number = -1, minSize?: number): Panel {
		return { snippet, size, minSize };
	}
</script>

<script lang="ts">
	import { untrack, type Snippet } from 'svelte';

	type Props = {
		panels: Panel[];
		direction?: 'horizontal' | 'vertical';
	};
	let { panels = $bindable<Panel[]>(), direction = $bindable('horizontal') }: Props = $props();
	// init panels' size
	$effect(() => {
		const len = panels.length;
		// no need to resize when has no panel
		if (len === 0) {
			return;
		}
		untrack(() => {
			// if has resized,return
			if (!panels.some((p) => p.size < 0)) {
				return;
			}

			let totalSize = 100;
			let needCount = 0;
			panels.forEach((p) => {
				if (p.size >= 0) {
					totalSize -= p.size;
				} else {
					needCount++;
				}
			});
			if (totalSize < 0 || (totalSize === 0 && needCount > 0)) {
				needCount = panels.length;
				totalSize = 100;
			}
			if (needCount > 0) {
				const size = totalSize / needCount;
				panels = panels.map((value) => {
					if (value.size < 0) {
						value.size = size;
					}
					return value;
				});
			}
		});
	});

	let container: HTMLDivElement;

	function move(i: number, e: PointerEvent) {
		const r = container.getBoundingClientRect();
		// the offset of cursor pointer in container
		const dx = e.clientX - r.left;
		const dy = e.clientY - r.top;

		// mapping to first panel size
		const percent = (direction === 'horizontal' ? dx / r.width : dy / r.height) * 100;
		const before = panels.slice(0, i).reduce((a, b) => a + b.size, 0);
		const total = panels[i].size + panels[i + 1].size;
		let firstSize = Math.max(0, Math.min(100, percent)) - before;

		// limit panel's minimum size
		const minBefore = panels[i].minSize;
		const minAfter = panels[i + 1].minSize;
		if (minBefore && minBefore > 0) {
			firstSize = Math.max(minBefore, firstSize);
		}
		if (minAfter && minAfter > 0) {
			firstSize = Math.min(total - minAfter, firstSize);
		}

		// assign to panels where divider before and after
		panels[i].size = Math.max(0, Math.min(total, firstSize));
		panels[i + 1].size = total - panels[i].size;
	}
</script>

<div
	bind:this={container}
	class="flex w-full h-full
	{direction === 'horizontal' ? 'flex-row' : 'flex-col'}"
>
	{#each panels as panel, i}
		<div class="flex min-w-0 min-h-0 overflow-auto" style:flex="{panel.size}% 0 0">
			{@render panel.snippet()}
		</div>
		{#if i < panels.length - 1}
			<div
				role="separator"
				aria-orientation={direction === 'horizontal' ? 'vertical' : 'horizontal'}
				class="shrink-0 group relative
				{direction === 'horizontal' ? 'w-1.5 -mx-0.75 cursor-col-resize' : 'h-1 -my-0.5 cursor-row-resize'}
				"
				onpointerdown={(e) => e.currentTarget.setPointerCapture(e.pointerId)}
				onpointermove={(e) => e.buttons === 1 && move(i, e)}
			>
				<div
					class="absolute bg-base-300 transition-colors group-hover:bg-accent
				{direction === 'horizontal'
						? 'inset-y-0 left-1/2 w-px -translate-x-1/2 group-hover:w-full'
						: 'inset-x-0 top-1/2 h-px -translate-y-1/2 group-hover:h-full'}
				"
				></div>
			</div>
		{/if}
	{/each}
</div>
