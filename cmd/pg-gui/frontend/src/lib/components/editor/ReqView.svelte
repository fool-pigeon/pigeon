<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts" module>
	// Icons

	import IconLoaderCircle from '@lucide/svelte/icons/loader-circle';
	import IconSend from '@lucide/svelte/icons/send';

	import type { OpenedFS } from './Editor.svelte';
	import { onMount } from 'svelte';
	import { ReqCtxService, ReqFileService } from '@cmd-gui/service';
	import { msgAlert } from '../message/AlertNotify.svelte';
	import { Events } from '@wailsio/runtime';
	import { MenuEvent } from '@cmd-gui/gui';
	import equal from 'fast-deep-equal';
	import { DataEvent } from '@lib/enum/events';
	import DynamicJoin, { NewPanel } from '../base/DynamicJoin.svelte';
	import CMFromHttpURL from '@comp/form/CMFormHttpURL.svelte';
	import type { PageLoadStatus } from '@type/page';
	import HttpResViewer from './HttpResViewer.svelte';
	import ReqEditor from './ReqEditor/ReqEditor.svelte';

	export type HttpResState = {
		loaded: boolean;
		isSending: boolean;
		current?: TaskResHttp;
		history: TaskResHttp[];
		err?: unknown;
	};
</script>

<script lang="ts">
	import HttpReqURLBar from './HttpReqURLBar.svelte';
	import type { Request } from '@pg/protocol';

	type Prop = {
		openedFS: OpenedFS;
		isFlexRow: boolean;
	};
	let { openedFS, isFlexRow = $bindable() }: Prop = $props();
	// original http request content
	let original = $state<Request | null>(null);
	// current http request content
	let temp = $state<Request | null>(null);
	// load status for this page
	let loadStatus = $state<PageLoadStatus>({ loaded: false, err: null });
	// default state of request's run results for result viewer
	let resState = $state<HttpResState>({ loaded: false, isSending: false, history: [] });
	// Current Selected body part
	function send() {
		if (resState.isSending) {
			return;
		}
		resState.isSending = true;
		resState.err = undefined;
		const req = $state.snapshot(temp);
		ReqCtxService.SendHttp(req)
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
			ReqFileService.Write(fs.id, temp)
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
			ReqFileService.Read(fs.id).then((v) => {
				temp = v;
			}),
			ReqFileService.ReadOriginal(fs.id).then((v) => {
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
			// TODO: do not stage when editor's ofs not contains this
			await ReqFileService.Stage(fs.id, temp);
		};
	});
	$effect(() => {
		openedFS.saved = equal(original, temp);
	});
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
		<div class="w-full h-full flex flex-col">
			<!-- URL bar -->
			<HttpReqURLBar bind:req={temp} isSending={resState.isSending} />
			<div class="mt-1 w-full grow min-h-0">
				<DynamicJoin bind:isFlexRow panels={[NewPanel(reqEditor), NewPanel(respViewer)]} />
			</div>
		</div>
	{/if}
</div>

{#snippet reqEditor()}
	<ReqEditor bind:temp />
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
		<div class="w-full h-full flex flex-col gap-1 justify-center items-center">
			<IconSend class="size-32 text-base-content/60" strokeWidth={3} nonScalingStroke />
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
