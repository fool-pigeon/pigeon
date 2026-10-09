<script lang="ts">
	// Icons
	import IconSave from '@lucide/svelte/icons/save';

	import type { Body, ReqData } from '@internal/request/httpx';
	import type { OpenedFS } from './Editor.svelte';
	import { onMount, untrack } from 'svelte';
	import { HttpReqFileService, ReqCtxService, type HttpRes } from '@pg/service';
	import { msgAlert } from '../message/AlertNotify.svelte';
	import { Events } from '@wailsio/runtime';
	import { MenuEvent } from '@pg/gui';
	import { fileExplorer } from '../navigation/FileExplorer.svelte';
	import equal from 'fast-deep-equal';
	import { DataEvent } from '@lib/enum/events';
	import DynamicJoin, { NewPanel } from '../base/DynamicJoin.svelte';
	import ScrollFlex from '../base/ScrollFlex.svelte';
	import CMInput from '../form/CMInput.svelte';
	import type { KArr } from '@type/form';
	import CMFormKV from '@comp/form/CMFormKV.svelte';
	import CMFormKArr from '@comp/form/CMFormKArr.svelte';
	import CMFromHttpURL from '@comp/form/CMFromHttpURL.svelte';
	import CMTextArea from '@comp/form/CMTextArea.svelte';
	import CMFormHttpMultipart from '@comp/form/CMFormHttpMultipart.svelte';
	import FileSelector from '@comp/form/FileSelector.svelte';
	import type { PageLoadStatus } from '@type/page';

	type Prop = {
		openedFS: OpenedFS;
		isFlexRow: boolean;
	};
	let { openedFS, isFlexRow = $bindable() }: Prop = $props();
	let original = $state<ReqData | null>(null);
	let temp = $state<ReqData | null>(null);

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
		Desc: 'Description',
		PreReq: 'PreRequest',
		// TODO: for more,auth,retry,other setting
	} as const;
	type EditPart = (typeof EditPart)[keyof typeof EditPart];
	let reqEditPart = $state<EditPart>(EditPart.Path);
	const BodyPart = {
		None: 'none',
		Raw: 'raw',
		Multipart: 'multipart/form-data',
		UrlEncoded: 'x-www-form-urlencoded',
		Binary: 'binary',
	} as const;
	type BodyPart = (typeof BodyPart)[keyof typeof BodyPart];
	let bodyPart = $state<BodyPart>(BodyPart.None);
	const ResViewPart = {
		ResBody: 'Response',
		ResHeader: 'Header',
		Tests: 'Tests',
		History: 'History',
	} as const;
	type ResViewPart = (typeof ResViewPart)[keyof typeof ResViewPart];
	let viewPart = $state<ResViewPart>(ResViewPart.ResBody);
	function purifyTemp() {
		if (!temp) {
			return;
		}
		temp.body = purifyBody($state.snapshot(temp.body), bodyPart);
	}
	/**
	 * return the body that only has one part of b
	 * @param b
	 * @param bodyPart
	 */
	function purifyBody(b: Body, bodyPart: string) {
		const res: Body = {
			raw: '',
			multipart: null,
			form: null,
			binary: '',
			contentType: b.contentType,
		};
		switch (bodyPart) {
			case BodyPart.Raw:
				res.raw = b.raw;
				break;
			case BodyPart.Multipart:
				res.multipart = b.multipart;
				break;
			case BodyPart.UrlEncoded:
				res.form = b.form;
				break;
			case BodyPart.Binary:
				res.binary = b.binary;
				break;
		}
		return res;
	}
	let loadStatus = $state<PageLoadStatus>({ loaded: false, err: null });
	onMount(() => {
		const cur = { ...openedFS };
		const fs = cur.fs;
		const unregister = Events.On(MenuEvent.Save, () => {
			purifyTemp();
			HttpReqFileService.Write(fs.id, temp)
				.then((m) => {
					fs.additional = m;
					// original should be set to copy of temp
					original = $state.snapshot(temp);

					Events.Emit(DataEvent.EditorSaved);
					msgAlert.success(`Saved ${fs.name} Successfully`);
				})
				.catch(msgAlert.error);
		});
		Promise.allSettled([
			HttpReqFileService.Read(fs.id).then((v) => {
				temp = v;
				if (temp) {
					const b = temp.body;
					if (b.raw !== '') {
						bodyPart = BodyPart.Raw;
					} else if (b.multipart && b.multipart.length > 0) {
						bodyPart = BodyPart.Multipart;
					} else if (b.form) {
						bodyPart = BodyPart.UrlEncoded;
					} else if (b.binary !== '') {
						bodyPart = BodyPart.Binary;
					}
				}
			}),
			HttpReqFileService.ReadOriginal(fs.id).then((v) => {
				original = v;
			}),
		])
			.catch((err) => {
				loadStatus.err = err;
			})
			.finally(() => {
				loadStatus.loaded = true;
			});
		return async () => {
			unregister();
			purifyTemp();
			await HttpReqFileService.Stage(fs.id, temp).then((m) => (fs.additional = m));
		};
	});
	$effect(() => {
		openedFS.saved = equal(original, temp);
	});
	const METHOD_STYLE_CLASS = {
		[Method.GET]: 'text-req-get hover:bg-req-get/50',
		[Method.POST]: 'text-req-post hover:bg-req-post/50',
		[Method.PUT]: 'text-req-put hover:bg-req-put/50',
		[Method.DELETE]: 'text-req-delete hover:bg-req-delete/50',
		[Method.HEAD]: 'text-req-head hover:bg-req-head/50',
		[Method.PATCH]: 'text-req-patch hover:bg-req-patch/50',
		[Method.OPTIONS]: 'text-req-options hover:bg-req-options/50',
		[Method.CONNECT]: 'text-req-connect hover:bg-req-connect/50',
		[Method.TRACE]: 'text-req-trace hover:bg-req-trace/50',
	};

	type ResViewState = {
		loaded: boolean;
		current?: HttpRes;
		history: HttpRes[];
	};
	let resState = $state<ResViewState>({ loaded: false, history: [] });
</script>

<div class="p-2 w-full h-full">
	<!-- FIXME:Out-of-date assignments during concurrent reads; use ‘cancelled’ -->
	{#if !loadStatus.loaded}
		<div>loading...</div>
	{:else if loadStatus.err}
		<div class="w-full h-full flex justify-center items-center">
			<div class="text-2xl text-error">Load data error: {loadStatus.err}</div>
		</div>
	{:else if original === null || temp === null}
		<div class="w-full h-full flex justify-center items-center text-error">the value is null</div>
	{:else}
		<!-- TODO finish form input of HttpReq -->
		<!-- ! outline-1 for test -->
		<div class="w-full h-full flex flex-col">
			<!-- for test view -->
			<div class="w-full h-1/3 flex overflow-auto">
				<div class="w-1/2 text-info">{JSON.stringify(temp, null, 2)}</div>
				<div class="w-1/2 text-success">{JSON.stringify(original, null, 2)}</div>
			</div>
			<!-- URL bar -->
			<div class="flex gap-1">
				<div class="w-full flex gap-1 items-center outline-1">
					<!-- HTTP Method Selector -->
					<select
						class="appearance-none focus:outline-none text-center [text-align-last:center]
							{METHOD_STYLE_CLASS[temp?.method ?? ''] ?? ''}"
						bind:value={temp.method}
					>
						{#each Object.values(Method) as m (m)}
							<option value={m}>{m}</option>
						{/each}
					</select>
					<CMFromHttpURL bind:uri={temp.url} bind:query={temp.query} />
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
				<button
					class="outline"
					onclick={() => {
						ReqCtxService.SendHttp($state.snapshot(temp))
							.then((v) => {
								resState.loaded = true;
								resState.current = v;
								resState.history.push(v);
							})
							.catch(msgAlert.error);
					}}>Send</button
				>
			</div>
			<div class="mt-1 w-full grow">
				<DynamicJoin bind:isFlexRow panels={[NewPanel(reqEditor), NewPanel(respViewer)]} />
			</div>
		</div>
	{/if}
</div>

{#snippet reqEditor()}
	<div class="w-full h-full flex flex-col gap-1">
		<!-- edit part select -->
		<div role="tablist" class="w-full border-b border-primary/50">
			<ScrollFlex>
				{#snippet content()}
					{#each Object.values(EditPart) as p (p)}
						<button
							class="p-1 cursor-pointer text-base-content/50
							{p === reqEditPart ? 'text-primary-content bg-primary' : 'hover:text-base-content'}"
							onclick={() => (reqEditPart = p)}>{p}</button
						>
					{/each}
				{/snippet}
			</ScrollFlex>
		</div>
		<!-- content -->
		<!-- TODO :binding form -->
		{#if !temp}
			<div class="text-error">temp value is null:{temp}</div>
		{:else if reqEditPart === EditPart.Path}
			<CMFormKV bind:value={temp.path} kReadOnly />
		{:else if reqEditPart === EditPart.Query}
			<CMFormKArr bind:value={temp.query} />
		{:else if reqEditPart === EditPart.Body}
			<select
				class="w-48 select select-sm select-primary focus:outline-none text-center [text-align-last:center]"
				bind:value={bodyPart}
			>
				{#each Object.values(BodyPart) as b (b)}
					<option value={b}>{b}</option>
				{/each}
			</select>
			{#if bodyPart === BodyPart.Raw}
				<!-- TODO: different raw type such as json,xml,text -->
				<CMTextArea bind:value={temp.body.raw} />
			{:else if bodyPart === BodyPart.Multipart}
				<CMFormHttpMultipart bind:multipart={temp.body.multipart} />
			{:else if bodyPart === BodyPart.UrlEncoded}
				<CMFormKArr bind:value={temp.body.form} />
			{:else if bodyPart === BodyPart.Binary}
				<FileSelector bind:path={temp.body.binary} />
			{/if}
		{:else if reqEditPart === EditPart.Header}
			<CMFormKArr bind:value={temp.header} />
		{:else if reqEditPart === EditPart.Tests}
			<!-- TODO: pretty Tests form bindings -->
			<CMFormKV bind:value={temp.tests} kHeader="name" vHeader="script" />
		{:else if reqEditPart === EditPart.Desc}
			<CMTextArea bind:value={temp.desc} />
		{:else if reqEditPart === EditPart.PreReq}
			<CMTextArea bind:value={temp.preReq} />
		{/if}
	</div>
{/snippet}

{#snippet respViewer()}
	<!-- TODO finish response viewer -->
	<div class="w-full h-full flex flex-col">
		{#if !resState.loaded || !resState.history || resState.history.length <= 0}
			<p>tip here</p>
		{:else}
			<div class="w-full flex border-b border-primary/50">
				<!-- view part select -->
				<div role="tablist">
					<ScrollFlex>
						{#snippet content()}
							{#each Object.values(ResViewPart) as p (p)}
								<button
									class="p-1 cursor-pointer text-base-content/50
									{p === viewPart ? 'text-primary-content bg-primary' : 'hover:text-base-content'}"
									onclick={() => (viewPart = p)}>{p}</button
								>
							{/each}
						{/snippet}
					</ScrollFlex>
				</div>
				{#if resState.current}
					{#if viewPart === 'Response'}
						<!-- Change response body preview type -->
						change view type
					{/if}
					<!-- current response view status -->
					<div>{resState.current.status}</div>
				{/if}
			</div>
			{#if viewPart === 'Response'}
				<iframe
					title="response-body"
					class="w-full h-full"
					sandbox="allow-same-origin allow-modals allow-forms"
					// srcdoc={resViewState}
				>
				</iframe>
			{:else if viewPart === 'Header'}
				Header
			{:else if viewPart === 'Tests'}
				Tests
			{:else if viewPart === 'History'}
				History
			{/if}
		{/if}
	</div>
{/snippet}
