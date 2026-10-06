<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts" module>
	import type { Snippet } from 'svelte';
	class ContextMenuState {
		x = $state(0);
		y = $state(0);
		show = $state(false);
		shadow = $state(false);
		content = $state<Snippet>();

		el = $state<HTMLDivElement>();

		open(e: MouseEvent, content?: Snippet, shadow = false) {
			e.preventDefault();
			this.x = e.clientX;
			this.y = e.clientY;
			this.content = content;
			this.shadow = shadow;
			this.show = true;
		}
		close() {
			this.show = false;
			this.shadow = false;
			this.content = undefined;
		}
	}
	const cState = new ContextMenuState();
	export const contextmenu = {
		open: cState.open.bind(cState),
		close: cState.close.bind(cState),
		setShadow: ((s = false) => {
			cState.shadow = s;
		}).bind(cState),
	};
</script>

<script lang="ts">
	$effect(() => {
		if (!cState.show || !cState.el) {
			return;
		}
		const rect = cState.el.getBoundingClientRect();
		if (rect.right > window.innerWidth) {
			cState.x = window.innerWidth - rect.width;
		}
		if (rect.bottom > window.innerHeight) {
			cState.y = window.innerHeight - rect.height;
		}
	});
</script>

<svelte:window
	onclick={(e) => {
		if (cState.show && !cState.el?.contains(e.target as Node)) {
			cState.close();
		}
	}}
	onkeydown={(e) => e.key === 'Escape' && cState.close()}
/>
{#if cState.show}
	<div
		bind:this={cState.el}
		class="z-50 fixed {cState.show ? 'shadow-lg' : ''}"
		style:left="{cState.x}px"
		style:top="{cState.y}px"
	>
		{@render cState.content?.()}
	</div>
{/if}
