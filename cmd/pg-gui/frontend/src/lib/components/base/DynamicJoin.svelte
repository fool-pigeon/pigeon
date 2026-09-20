<script lang="ts" module>
	type Unit = '%' | 'px';
	type Size = {
		value: number;
		unit: Unit;
	};
	export type Panel = {
		snippet: Snippet;
		// size.value<0 means auto resize
		size: Size;
		minSize?: Size;
	};
	/**
	 * Create a panel for join
	 * @param snippet the content in panel
	 * @param size initalize the size of panel.if size.value<0,it will be resize
	 * @param minSize set the minimum size for panel
	 */
	export function NewPanel(
		snippet: Snippet,
		size: Size = { unit: '%', value: -1 },
		minSize?: Size,
	): Panel {
		return { snippet, size, minSize };
	}
</script>

<script lang="ts">
	import { Backpack } from '@lucide/svelte';

	import { untrack, type Snippet } from 'svelte';

	type Props = {
		panels: Panel[];
		direction?: 'horizontal' | 'vertical';
	};
	let { panels = $bindable<Panel[]>(), direction = $bindable('horizontal') }: Props = $props();

	// to the percent of length
	function getPct(length: number, pixelLength: number): number {
		return (length / pixelLength) * 100;
	}
	// to the pixel of length
	function getPx(percent: number, pixelLength: number): number {
		return (percent / 100) * pixelLength;
	}
	// init panels' size
	$effect(() => {
		const len = panels.length;
		// no need to resize when has no panel
		if (len === 0) {
			return;
		}
		// if has resized,return
		if (!panels.some((p) => p.size.value < 0)) {
			return;
		}
		untrack(() => {
			const r = container.getBoundingClientRect();
			const pixelLength = direction === 'horizontal' ? r.width : r.height;
			if (pixelLength <= 0) {
				return;
			}
			let totalPercentSize = 100;
			let needCount = 0;
			for (const p of panels) {
				if (totalPercentSize < 0) {
					needCount = panels.length;
					totalPercentSize = 100;
					break;
				}
				if (p.size.value < 0) {
					needCount++;
					continue;
				}
				totalPercentSize -= p.size.unit === '%' ? p.size.value : getPct(p.size.value, pixelLength);
			}
			const needSize = totalPercentSize / needCount;
			if (needCount > 0) {
				panels = panels.map((p) => {
					if (p.size.value < 0) {
						p.size = { value: needSize, unit: '%' };
					}
					return p;
				});
			}
		});
	});
	$effect(() => {
		if (panels.some((p) => p.size.value < 0)) {
			return;
		}
		const ro = new ResizeObserver(([e]) => {
			const pxLen = direction === 'horizontal' ? e.contentRect.width : e.contentRect.height;
			const needCount = panels.filter((p) => p.size.unit === '%').length;
			if (needCount <= 0) {
				return;
			}
			const deltaPercentSize = panels
				.map((p) => (p.size.unit === '%' ? p.size.value : getPct(p.size.value, pxLen)))
				.reduce((pre, cur) => pre - cur, 100);
			for (const p of panels) {
				if (p.size.unit === '%') {
					p.size.value += deltaPercentSize / needCount;
				}
			}
		});
		ro.observe(container);
		return () => ro.disconnect();
	});

	let container: HTMLDivElement;

	/**
	 * Move divider's place and resize adjacent panel
	 * @param i the index of divider
	 * @param e the PointerEvent
	 */
	function move(i: number, e: PointerEvent) {
		const r = container.getBoundingClientRect();
		// the offset of cursor pointer in container
		const dx = e.clientX - r.left;
		const dy = e.clientY - r.top;

		const pxLen = direction === 'horizontal' ? r.width : r.height;
		const pct = (size: Size) => (size.unit === '%' ? size.value : getPct(size.value, pxLen));
		const px = (size: Size) => (size.unit === 'px' ? size.value : getPx(size.value, pxLen));

		// front panel size and back panel size
		const frontSize = panels[i].size;
		const backSize = panels[i + 1].size;
		// mapping to first panel size
		const current = getPct(direction === 'horizontal' ? dx : dy, pxLen);
		const before = panels.slice(0, i).reduce((a, b) => a + pct(b.size), 0);
		const total = pct(frontSize) + pct(backSize);

		let frontPct = Math.max(0, Math.min(100, current)) - before;

		// limit panel's minimum size
		const minBefore = panels[i].minSize;
		const minAfter = panels[i + 1].minSize;
		if (minBefore && minBefore.value > 0) {
			frontPct = Math.max(pct(minBefore), frontPct);
		}
		if (minAfter && minAfter.value > 0) {
			frontPct = Math.min(total - pct(minAfter), frontPct);
		}
		frontPct = Math.max(0, Math.min(total, frontPct));
		const backPct = total - frontPct;
		// resize adjacent panels
		panels[i].size.value = frontSize.unit === '%' ? frontPct : getPx(frontPct, pxLen);
		panels[i + 1].size.value = backSize.unit === '%' ? backPct : getPx(backPct, pxLen);
	}
</script>

<div
	bind:this={container}
	class="flex w-full h-full
	{direction === 'horizontal' ? 'flex-row' : 'flex-col'}"
>
	{#each panels as panel, i}
		<div
			class="flex min-w-0 min-h-0 overflow-auto"
			style:flex="{panel.size.value}{panel.size.unit} 0 0"
		>
			{@render panel.snippet()}
		</div>
		<!-- add divider between adjacent panels -->
		{#if i < panels.length - 1}
			<div
				role="separator"
				aria-orientation={direction === 'horizontal' ? 'vertical' : 'horizontal'}
				// hidden the width
				class="shrink-0 group relative
				{direction === 'horizontal' ? 'w-1.5 -mx-0.75 cursor-col-resize' : 'h-1 -my-0.5 cursor-row-resize'}
				"
				onpointerdown={(e) => e.currentTarget.setPointerCapture(e.pointerId)}
				onpointermove={(e) => e.buttons === 1 && move(i, e)}
			>
				<div
					// Draw the cue line,hidden when not hover
					class="absolute bg-base-300 transition-colors group-hover:bg-accent
					{direction === 'horizontal'
						? 'inset-y-0 left-1/2 w-px -translate-x-1/2 group-hover:w-full'
						: 'inset-x-0 top-1/2 h-px -translate-y-1/2 group-hover:h-full'}"
				></div>
			</div>
		{/if}
	{/each}
</div>
