<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts" module>
	// Icons
	import IconSave from '@lucide/svelte/icons/save';
	import IconLoaderCircle from '@lucide/svelte/icons/loader-circle';
	import IconSend from '@lucide/svelte/icons/send';

	import { Method, type Body, type ReqData } from '@internal/request/httpx';
	import type { OpenedFS } from './Editor.svelte';
	import { onMount } from 'svelte';
	import { HttpReqFileService, ReqCtxService, type TaskResHttp } from '@bindings/service';
	import { msgAlert } from '../message/AlertNotify.svelte';
	import { Events } from '@wailsio/runtime';
	import { MenuEvent } from '@bindings/gui';
	import equal from 'fast-deep-equal';
	import { DataEvent } from '@lib/enum/events';
	import DynamicJoin, { NewPanel } from '../base/DynamicJoin.svelte';
	import CMFromHttpURL from '@comp/form/CMFromHttpURL.svelte';
	import type { PageLoadStatus } from '@type/page';
	import HttpResViewer from './HttpResViewer.svelte';
	import HttpReqEditor from './HttpReqEditor.svelte';

	export type HttpResState = {
		loaded: boolean;
		isSending: boolean;
		current?: TaskResHttp;
		history: TaskResHttp[];
		err?: unknown;
	};

	export const EditPart = {
		Path: 'Path',
		Query: 'Query',
		Body: 'Body',
		Header: 'Header',
		Tests: 'Tests',
		Desc: 'Description',
		PreReq: 'PreRequest',
		// TODO: for more,auth,retry,other setting
	} as const;
	export type EditPart = (typeof EditPart)[keyof typeof EditPart];

	export const BodyPart = {
		None: 'none',
		Raw: 'raw',
		Multipart: 'multipart/form-data',
		UrlEncoded: 'x-www-form-urlencoded',
		Binary: 'binary',
	} as const;
	export type BodyPart = (typeof BodyPart)[keyof typeof BodyPart];
</script>

<script lang="ts">
	type Prop = {
		openedFS: OpenedFS;
		isFlexRow: boolean;
	};
	let { openedFS, isFlexRow = $bindable() }: Prop = $props();
	// original http request content
	let original = $state<ReqData | null>(null);
	// current http request content
	let temp = $state<ReqData | null>(null);
	// load status for this page
	let loadStatus = $state<PageLoadStatus>({ loaded: false, err: null });
	// default state of request's run results for result viewer
	let resState = $state<HttpResState>({ loaded: false, isSending: false, history: [] });
	// Current Selected body part
	let bodyPart = $state<BodyPart>(BodyPart.None);

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
	function send() {
		if (resState.isSending) {
			return;
		}
		resState.isSending = true;
		resState.err = undefined;
		ReqCtxService.SendHttp($state.snapshot(temp))
			.then((v) => {
				resState.current = v;
				resState.history.push(v);
			})
			.catch((err) => {
				resState.err = err;
				msgAlert.error(err);
			})
			.finally(() => {
				resState.isSending = false;
				resState.loaded = true;
			});
	}

	onMount(() => {
		const cur = { ...openedFS };
		const fs = cur.fs;
		const unregisterSave = Events.On(MenuEvent.Save, () => {
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
		const unregisterSend = Events.On(MenuEvent.RunRequest, () => send());
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
			unregisterSave();
			unregisterSend();
			purifyTemp();
			await HttpReqFileService.Stage(fs.id, temp);
		};
	});
	$effect(() => {
		openedFS.saved = equal(original, temp);
	});

	const METHOD_STYLE_CLASS = {
		[Method.MethodGet]: 'text-req-get hover:bg-req-get/50',
		[Method.MethodPost]: 'text-req-post hover:bg-req-post/50',
		[Method.MethodPut]: 'text-req-put hover:bg-req-put/50',
		[Method.MethodDelete]: 'text-req-delete hover:bg-req-delete/50',
		[Method.MethodHead]: 'text-req-head hover:bg-req-head/50',
		[Method.MethodPatch]: 'text-req-patch hover:bg-req-patch/50',
		[Method.MethodOptions]: 'text-req-options hover:bg-req-options/50',
		[Method.MethodConnect]: 'text-req-connect hover:bg-req-connect/50',
		[Method.MethodTrace]: 'text-req-trace hover:bg-req-trace/50',
	};
</script>

<div class="p-2 w-full h-full min-h-0">
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
			<!-- <div class="w-full h-1/3 flex overflow-auto">
				<div class="w-1/2 text-info">{JSON.stringify(temp, null, 2)}</div>
				<div class="w-1/2 text-success">{JSON.stringify(original, null, 2)}</div>
			</div> -->
			<!-- URL bar -->
			<div class="flex gap-1">
				<div class="p-1 w-full flex gap-1 items-center border border-base-content/50">
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
					<CMFromHttpURL bind:uri={temp.url} bind:query={temp.query} bind:path={temp.path} />
					<!-- Save file btn -->
					<button
						title="Save"
						class="pl-1 pr-1 h-full opacity-50 text-base-content hover:opacity-100 hover:bg-base-content/10"
						onclick={() => Events.Emit(MenuEvent.Save)}
					>
						<IconSave class="size-5" />
					</button>
				</div>
				{#if resState.isSending}
					<button
						class="pt-0 pb-0 btn btn-error"
						onclick={() => Events.Emit(MenuEvent.CancelRequest)}>Cancel</button
					>
				{:else}
					<button
						class="pt-0 pb-0 btn btn-primary"
						onclick={() => Events.Emit(MenuEvent.RunRequest)}>Send</button
					>
				{/if}
			</div>
			<div class="mt-1 w-full grow min-h-0">
				<DynamicJoin bind:isFlexRow panels={[NewPanel(reqEditor), NewPanel(respViewer)]} />
			</div>
		</div>
	{/if}
</div>

{#snippet reqEditor()}
	<HttpReqEditor bind:temp bind:bodyPart />
{/snippet}

{#snippet respViewer()}
	{#if resState.isSending}
		<div class="w-full h-full flex flex-col gap-2 justify-center items-center">
			<IconLoaderCircle class="size-20 animate-spin" strokeWidth={2} nonScalingStroke />
			<button class="btn btn-soft btn-error" onclick={() => Events.Emit(MenuEvent.CancelRequest)}>
				Cancel
			</button>
		</div>
	{:else if !resState.loaded}
		<div class="w-full h-full flex flex-col justify-center items-center">
			<IconSend class="size-48 text-base-content/60" strokeWidth={3} nonScalingStroke />
			<div class="inline-grid grid-cols-2 gap-x-2">
				<span class="text-right"> Send Request </span>
				<span class="text-left">
					<kbd class="kbd">⌘/Ctrl</kbd>
					+
					<kbd class="kbd">Enter</kbd>
				</span>
			</div>
		</div>
	{:else if resState.err}
		<div class="text-error">{resState.err}</div>
	{:else if !resState.current}
		<p>current response result is nil</p>
	{:else}
		<HttpResViewer current={resState.current} history={resState.history} />
	{/if}
{/snippet}
