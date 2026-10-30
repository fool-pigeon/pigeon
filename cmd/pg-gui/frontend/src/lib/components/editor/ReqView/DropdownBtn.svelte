<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts" generics="T">
	import IconChevronDown from '@lucide/svelte/icons/chevron-down';
	import type { Snippet } from 'svelte';
	import type { MouseEventHandler } from 'svelte/elements';

	type Prop = {
		// val: T;
		popoverID: string;
		onclick: MouseEventHandler<HTMLButtonElement>;
		summary: Snippet;
		details: Snippet<[btn: Snippet<[k: T]>, divider: Snippet<[s: string]>]>;
		btn: Snippet<[k: T]>;
	};
	let { popoverID, onclick, summary, details, btn }: Prop = $props();
</script>

<button popovertarget={popoverID} style="anchor-name:--{popoverID}" {onclick}>
	<div class="flex items-center">
		{@render summary()}
		<IconChevronDown class="size-4" />
	</div>
</button>
<ul
	class="dropdown menu max-h-[60vh] bg-base-100 overflow-auto"
	onwheel={(e) => e.stopPropagation()}
	popover
	id={popoverID}
	style="position-anchor:--{popoverID}"
>
	{#snippet divider(s: string)}
		<div class="divider divider-start m-0 p-0">{s}</div>
	{/snippet}
	{@render details(btn, divider)}
</ul>
