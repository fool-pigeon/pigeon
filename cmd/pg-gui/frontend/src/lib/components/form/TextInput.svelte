<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import type { Snippet } from 'svelte';

	// Icons

	type Prop = { val: string; placeholder?: string; iconSnippet?: Snippet };
	let { val = $bindable(), placeholder, iconSnippet }: Prop = $props();
</script>

<label class="input">
	{#if iconSnippet}
		{@render iconSnippet()}
	{/if}
	<input class="grow" type="text" {placeholder} bind:value={val} />
</label>
