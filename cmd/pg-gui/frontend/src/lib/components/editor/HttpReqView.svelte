<script lang="ts">
	// Icons
	import IconSave from '@lucide/svelte/icons/save';

	import type { ReqData } from '@internal/request/httpx';
	import type { OpenedFS } from './Editor.svelte';
	import { onMount, untrack } from 'svelte';
	import { HttpReqFileService } from '@bindings/service';
	import { msgAlert } from '../message/AlertNotify.svelte';
	import { Events } from '@wailsio/runtime';
	import { MenuEvent } from '@bindings/gui';
	import { fileExplorer } from '../navigation/FileExplorer.svelte';
	import equal from 'fast-deep-equal';
	import { DataEvent } from '@lib/enum/events';
	import DynamicJoin, { NewPanel } from '../base/DynamicJoin.svelte';
	import ScrollFlex from '../base/ScrollFlex.svelte';
	import CMInput from '../form/CMInput.svelte';
	import type { KArr } from '@type/form';

	type Prop = {
		cur: OpenedFS;
		isFlexRow: boolean;
	};
	let { cur = $bindable<OpenedFS>(), isFlexRow = $bindable() }: Prop = $props();
	let original = $state<ReqData | null>(null);
	let temp = $state<ReqData | null>(null);

	let rawURL = $state('');
	const Method = {
		GET: 'GET',
		POST: 'POST',
		PUT: 'PUT',
		DELETE: 'DELETE',
		HEAD: 'HEAD',
		PATCH: 'PATCH',
		OPTIONS: 'OPTIONS',
		CONNECT: 'CONNECT',
		TRACE: 'TRACE',
	};
	const EditPart = {
		Path: 'Path',
		Query: 'Query',
		Body: 'Body',
		Header: 'Header',
		Tests: 'Tests',
	};
	let currentPart = $state(EditPart.Path);
	/**
	 * render RawURL from temp
	 */
	function getRawURL(): string {
		if (!temp) {
			return '';
		}
		let u = temp.url;
		if (temp.query && Object.keys(temp.query).length > 0) {
			u += '?';
			u += Object.entries(temp.query)
				.flatMap(([k, arr]) => (!arr ? [] : arr.map((v) => `${k}=${v}`)))
				.join('&');
		}
		return u;
	}
	/**
	 * parse params from u into temp
	 * @param u
	 */
	function parseRawURL(u: string) {
		if (!temp) {
			return;
		}
		const splitQuery = u.indexOf('?');
		temp.url = splitQuery < 0 ? u : u.slice(0, splitQuery);
		if (splitQuery < 0) {
			return;
		}
		const q: KArr = {};
		for (const kv of u
			.slice(splitQuery + 1)
			.split('&')
			.filter((v) => v.length > 0)) {
			const i = kv.indexOf('=');
			const k = i < 0 ? kv : kv.slice(0, i);
			const v = i < 0 ? '' : kv.slice(i + 1);
			// init when undefined
			q[k] ??= [];
			q[k].push(v);
		}
		temp.query = q;
	}
	function equalsRawURL(u1: string, u2: string): boolean {
		const i1 = u1.indexOf('?');
		const i2 = u2.indexOf('?');
		const path1 = i1 === -1 ? u1 : u1.slice(0, i1);
		const path2 = i2 === -1 ? u2 : u2.slice(0, i2);
		const query1 = i1 === -1 ? '' : u1.slice(i1 + 1);
		const query2 = i2 === -1 ? '' : u2.slice(i2 + 1);
		return path1 === path2 && query1 === query2;
	}
	onMount(() => {
		const unregister = Events.On(MenuEvent.Save, () => {
			HttpReqFileService.Write(cur.fs.id, temp)
				.then((m) => {
					const fs = fileExplorer.find(cur.fs.id);
					if (fs) {
						fs.additional = m;
					}
					cur.fs.additional = m;
					original = temp;

					Events.Emit(DataEvent.EditorSaved);
					msgAlert.success(`Saved ${cur.fs.name} Successfully`);
				})
				.catch(msgAlert.error);
		});
		return () => {
			unregister();
			if (!cur.saved) {
				HttpReqFileService.Stage(cur.fs.id, temp).then((m) => (cur.fs.additional = m));
			}
		};
	});
	$effect(() => {
		cur.saved = equal(original, temp);
	});
	// temp -> rawURL (listen temp)
	$effect(() => {
		if (!temp) return;
		const u = untrack(() => getRawURL());
		if (untrack(() => equalsRawURL(u, rawURL))) return;
		untrack(() => (rawURL = u));
	});
	// rawURL -> temp (listen rawURL)
	$effect(() => {
		if (rawURL && untrack(() => !equalsRawURL(rawURL, getRawURL()))) {
			untrack(() => parseRawURL(rawURL));
		}
	});
	let methodClassStyle = $derived(
		temp === null
			? ''
			: temp.method === Method.GET
				? 'text-req-get hover:bg-req-get/50'
				: temp.method === Method.POST
					? 'text-req-post hover:bg-req-post/50'
					: temp.method === Method.PUT
						? 'text-req-put hover:bg-req-put/50'
						: temp.method === Method.DELETE
							? 'text-req-delete hover:bg-req-delete/50'
							: temp.method === Method.HEAD
								? 'text-req-head hover:bg-req-head/50'
								: temp.method === Method.PATCH
									? 'text-req-patch hover:bg-req-patch/50'
									: temp.method === Method.OPTIONS
										? 'text-req-options hover:bg-req-options/50'
										: temp.method === Method.CONNECT
											? 'text-req-connect hover:bg-req-connect/50'
											: temp.method === Method.TRACE
												? 'text-req-trace hover:bg-req-trace/50'
												: '',
	);
</script>

<div class="p-2 w-full h-full">
	{#await Promise.allSettled( [HttpReqFileService.Read(cur.fs.id).then((v) => (temp = v)), HttpReqFileService.ReadOriginal(cur.fs.id).then((v) => (original = v))], )}
		<div>loading...</div>
	{:then}
		{#if original === null || temp === null}
			<div class="w-full h-full flex justify-center items-center text-error">the value is null</div>
		{:else}
			<!-- TODO finish form input of HttpReq -->
			<!-- for test view -->
			<div>{JSON.stringify(temp, null, 2)}</div>
			<!-- ! outline-1 for test -->
			<div class="w-full h-full flex flex-col">
				<!-- URL bar -->
				<div class="flex gap-1">
					<div class="w-full flex gap-1 items-center outline-1">
						<!-- HTTP Method Selector -->
						<select
							class="appearance-none focus:outline-none text-center [text-align-last:center]
							{methodClassStyle}"
							bind:value={temp.method}
						>
							{#each Object.values(Method) as m (m)}
								<option value={m}>{m}</option>
							{/each}
						</select>
						<CMInput bind:value={rawURL} placeholder={'input URL'} />
						<!-- Save file btn -->
						<button
							title="Save"
							class="mr-1 h-full text-base-content/80 hover:text-primary"
							onclick={() => Events.Emit(MenuEvent.Save)}
						>
							<IconSave class="size-5" />
						</button>
					</div>
					<!-- TODO Send Http request -->
					<button class="outline">Send</button>
				</div>
				<div class="mt-1 w-full grow">
					<DynamicJoin bind:isFlexRow panels={[NewPanel(reqEditor), NewPanel(respViewer)]} />
				</div>
			</div>
		{/if}
	{:catch err}
		<div class="w-full h-full flex justify-center items-center">
			<div class="text-2xl text-error">Load data error: {err}</div>
		</div>
	{/await}
</div>

{#snippet reqEditor()}
	<div class="w-full h-full flex flex-col gap-1">
		<!-- edit part select -->
		<div role="tablist" class="w-full text-base-content/50 border-b border-primary">
			<ScrollFlex>
				{#snippet content()}
					{#each Object.values(EditPart) as p (p)}
						<button
							class="p-1 cursor-pointer hover:text-base-content {p === currentPart
								? 'text-base-content bg-primary'
								: ''}"
							onclick={() => (currentPart = p)}>{p}</button
						>
					{/each}
				{/snippet}
			</ScrollFlex>
		</div>
		<!-- content -->
		{#if currentPart === EditPart.Path}
			Path
		{:else if currentPart === EditPart.Query}
			Query
		{:else if currentPart === EditPart.Body}
			Body
		{:else if currentPart === EditPart.Header}
			Header
		{:else if currentPart === EditPart.Tests}
			Tests
		{/if}
	</div>
{/snippet}

{#snippet respViewer()}
	<!-- TODO finish response viewer -->
	<div class="w-full h-full flex justify-center items-center">
		<div class="text-2xl">Response View</div>
	</div>
{/snippet}
