<script lang="ts">
	import { defaultKeymap, indentWithTab } from '@codemirror/commands';
	import { keymap } from '@codemirror/view';
	import { basicSetup, EditorView } from 'codemirror';
	import { onDestroy, onMount } from 'svelte';
	import type { OpenedFS } from './Editor.svelte';
	import { msgAlert } from '../message/AlertNotify.svelte';
	import { Events } from '@wailsio/runtime';
	import { MenuEvent } from '@bindings/gui';
	import { TextFileService } from '@services';
	import { fileExplorer } from '../navigation/FileExplorer.svelte';
	import CMWrapper from '../form/CMWrapper.svelte';
	import { DataEvent } from '@lib/enum/events';
	type Prop = {
		cur: OpenedFS;
	};
	let { cur = $bindable<OpenedFS>() }: Prop = $props();
	// original file content
	let original = $state('');
	// temp file content
	let temp = $state('');
	onMount(() => {
		const unregister = Events.On(MenuEvent.Save, () => {
			const curFS = cur.fs;
			TextFileService.Write(curFS.id, temp)
				.then((m) => {
					const fsItem = fileExplorer.find(curFS.id);
					if (fsItem) {
						fsItem.additional = m;
					}
					cur.fs.additional = m;
					original = temp;
					Events.Emit(DataEvent.EditorSaved);
					msgAlert.success(`Saved ${curFS.name} Successfully`);
				})
				.catch(msgAlert.error);
		});
		return () => {
			unregister();
			if (!cur.saved) {
				TextFileService.Stage(cur.fs.id, temp).then((m) => (cur.fs.additional = m));
			}
		};
	});
	$effect(() => {
		cur.saved = original === temp;
	});
</script>

{#await Promise.allSettled([TextFileService.Read(cur.fs.id)
		.then((v) => (temp = v))
		.catch(msgAlert.error), TextFileService.ReadOriginal(cur.fs.id)
		.then((v) => (original = v))
		.catch(msgAlert.error)]) then}
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
{/await}
