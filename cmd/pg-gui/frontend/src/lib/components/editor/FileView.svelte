<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import { defaultKeymap, indentWithTab } from '@codemirror/commands';
	import { keymap } from '@codemirror/view';
	import { basicSetup, EditorView } from 'codemirror';
	import { onDestroy, onMount } from 'svelte';
	import type { OpenedFS } from './Editor.svelte';
	import { msgAlert } from '../message/AlertNotify.svelte';
	import { Events } from '@wailsio/runtime';
	import { MenuEvent } from '@cmd-gui/gui';
	import { TextFileService } from '@services';
	import { fileExplorer } from '../navigation/FileExplorer.svelte';
	import CMWrapper from '../form/CMWrapper.svelte';
	import { DataEvent } from '@lib/enum/events';
	import type { PageLoadStatus } from '@type/page';
	type Prop = {
		openedFS: OpenedFS;
	};
	let { openedFS }: Prop = $props();
	// original file content
	let original = $state('');
	// temp file content
	let temp = $state('');
	let loadStatus = $state<PageLoadStatus>({ err: null, loaded: false });
	onMount(() => {
		const cur = { ...openedFS };
		const fs = cur.fs;
		Promise.allSettled([
			TextFileService.Read(fs.id).then((v) => (temp = v)),
			TextFileService.ReadOriginal(fs.id).then((v) => (original = v)),
		])
			.catch((err) => (loadStatus.err = err))
			.finally(() => (loadStatus.loaded = true));

		const unregister = Events.On(MenuEvent.Save, () => {
			TextFileService.Write(fs.id, temp)
				.then((m) => {
					fs.additional = m;
					// original should be set to the copy of temp
					original = $state.snapshot(temp);
					Events.Emit(DataEvent.EditorSaved);
					msgAlert.success(`Saved ${fs.name} Successfully`);
				})
				.catch(msgAlert.error);
		});
		return async () => {
			unregister();
			// TODO: do not stage when editor's ofs not contains this
			await TextFileService.Stage(fs.id, temp).then((m) => (fs.additional = m));
		};
	});
	$effect(() => {
		openedFS.saved = original === temp;
	});
</script>

{#if !loadStatus.loaded}
	<div>loading...</div>
{:else if loadStatus.err}
	<div class="w-full h-full flex justify-center items-center">
		<div class="text-2xl text-error">Load data error: {loadStatus.err}</div>
	</div>
{:else}
	<CMWrapper
		bind:value={temp}
		viewInitFunc={() => {
			return new EditorView({
				extensions: [
					basicSetup,
					keymap.of([indentWithTab, ...defaultKeymap]),
					EditorView.baseTheme({
						'&': { height: '100%' },
					}),
				],
			});
		}}
	/>
{/if}
