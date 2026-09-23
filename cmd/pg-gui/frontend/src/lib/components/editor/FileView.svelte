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
	type Prop = {
		openedFS: OpenedFS;
	};
	let { openedFS = $bindable<OpenedFS>() }: Prop = $props();
	// original file content
	let original = $state('');
	// temp file content
	let temp = $state('');
	let el: HTMLDivElement;
	let view: EditorView;
	onMount(async () => {
		await Promise.allSettled([
			TextFileService.Read(openedFS.fs.id)
				.then((v) => (temp = v))
				.catch(msgAlert.error),
			TextFileService.ReadOriginal(openedFS.fs.id)
				.then((v) => (original = v))
				.catch(msgAlert.error),
			Events.On(MenuEvent.Save, () => {
				const fs = openedFS.fs;
				TextFileService.Write(fs.id, temp)
					.then((m) => {
						const fsItem = fileExplorer.find(openedFS.fs.id);
						if (fsItem) {
							fs.additional = m;
						}
						openedFS.fs.additional = m;
						original = temp;
						msgAlert.success(`Saved ${fs.name} Successfully`);
					})
					.catch(msgAlert.error);
			}),
		]);
		view = new EditorView({
			doc: temp,
			parent: el,
			extensions: [
				basicSetup,
				keymap.of([indentWithTab, ...defaultKeymap]),
				EditorView.updateListener.of((u) => {
					if (u.docChanged) {
						temp = u.state.doc.toString();
					}
				}),
				EditorView.baseTheme({
					'&': { height: '100%' },
				}),
			],
		});
	});
	onDestroy(() => {
		view.destroy();
		Events.Off(MenuEvent.Save);
		if (!openedFS.saved) {
			TextFileService.Stage(openedFS.fs.id, temp).then((m) => (openedFS.fs.additional = m));
		}
	});
	$effect(() => {
		openedFS.saved = original === temp;
	});
</script>

<div class="w-full h-full" bind:this={el}></div>
