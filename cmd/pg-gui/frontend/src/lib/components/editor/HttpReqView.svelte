<script lang="ts">
	import type { ReqData } from '@internal/request/httpx';
	import type { OpenedFS } from './Editor.svelte';
	import { onDestroy, onMount } from 'svelte';
	import { HttpReqFileService } from '@bindings/service';
	import { msgAlert } from '../message/AlertNotify.svelte';
	import { Events } from '@wailsio/runtime';
	import { MenuEvent } from '@bindings/gui';
	import { fileExplorer } from '../navigation/FileExplorer.svelte';
	import equal from 'fast-deep-equal';

	type Prop = {
		cur: OpenedFS;
	};
	let { cur = $bindable<OpenedFS>() }: Prop = $props();
	let original = $state<ReqData | null>(null);
	let temp = $state<ReqData | null>(null);
	onMount(async () => {
		Events.On(MenuEvent.Save, () => {
			HttpReqFileService.Write(cur.fs.id, temp)
				.then((m) => {
					const fs = fileExplorer.find(cur.fs.id);
					if (fs) {
						fs.additional = m;
					}
					cur.fs.additional = m;
					original = temp;
					msgAlert.success(`Saved ${cur.fs.name} Successfully`);
				})
				.catch(msgAlert.error);
		});
	});
	onDestroy(() => {
		Events.Off(MenuEvent.Save);
		if (!cur.saved) {
			HttpReqFileService.Stage(cur.fs.id, temp).then((m) => (cur.fs.additional = m));
		}
	});
	$effect(() => {
		cur.saved = equal(original, temp);
	});
</script>

<div class="w-full h-full">
	{#await Promise.allSettled([HttpReqFileService.Read(cur.fs.id)
			.then((v) => (temp = v))
			.catch(msgAlert.error), HttpReqFileService.ReadOriginal(cur.fs.id)
			.then((v) => (original = v))
			.catch(msgAlert.error)])}
		<div>loading...</div>
	{:then}
		{#if original === null || temp === null}
			<div class="w-full h-full flex justify-center items-center text-error">the value is null</div>
		{:else}
			<!-- TODO finish form input of HttpReq -->
			<!-- for test view -->
			<div>{JSON.stringify(temp, null, 2)}</div>
		{/if}
	{/await}
</div>
