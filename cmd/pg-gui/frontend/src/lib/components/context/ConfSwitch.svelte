<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	// Icons
	import IconFileSliders from '@lucide/svelte/icons/file-sliders';

	import { msgAlert } from '@comp/message/AlertNotify.svelte';
	import { ReqCtxService, type FileMeta } from '@services';
	import { untrack } from 'svelte';
	import equal from 'fast-deep-equal';

	type Prop = { current?: FileMeta; path?: string };
	let { current = $bindable(), path }: Prop = $props();

	const cache = new Map<string, FileMeta>();
	let loaded = $state(false);
	let options = $state<FileMeta[]>([]);
	$effect(() => {
		if (path && current) cache.set(path, current);
	});
	$effect(() => {
		if (path) {
			loaded = false;
			untrack(() => {
				ReqCtxService.GetRootConfFiles(path)
					.then((opts) => {
						options = opts ?? [];
						const meta = cache.get(path);
						current = options.find((v) => equal(v, meta));
					})
					.catch(msgAlert.error)
					.finally(() => (loaded = true));
			});
		}
	});
</script>

<label class="flex items-center">
	<IconFileSliders class="shrink-0 size-3 text-base-content" />
	<span class="text-sm">Conf:</span>
	<select
		bind:value={current}
		class="pl-1 select select-sm justify-center outline-none shadow-none"
	>
		{#if !current}
			<option disabled selected value={undefined}>no conf</option>
		{/if}
		{#if loaded}
			{#each options as option}
				<option value={option}>{option.name}</option>
			{/each}
		{/if}
	</select>
</label>
