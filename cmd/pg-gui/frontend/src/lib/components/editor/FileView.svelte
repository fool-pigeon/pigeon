<script lang="ts">
	import { defaultKeymap, indentWithTab } from '@codemirror/commands';
	import { keymap } from '@codemirror/view';
	import { basicSetup, EditorView } from 'codemirror';
	import { onDestroy, onMount } from 'svelte';
	import type { OpenedFS } from './Editor.svelte';
	import { ReadOriginalText, ReadText, StageText, WriteText } from '@bindings/service/fileservice';
	import { msgAlert } from '../message/AlertNotify.svelte';
	import { Events } from '@wailsio/runtime';
	import { MenuEvent } from '@bindings/gui';

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
			ReadText(openedFS.fs.id)
				.then((v) => (temp = v))
				.catch(msgAlert.error),
			ReadOriginalText(openedFS.fs.id)
				.then((v) => (original = v))
				.catch(msgAlert.error),
			Events.On(MenuEvent.Save, () => {
				const fs = openedFS.fs;
				WriteText(fs.id, temp)
					.then(() => {
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
			StageText(openedFS.fs.id, temp);
		}
	});
	$effect(() => {
		openedFS.saved = original === temp;
	});
</script>

<div class="w-full h-full" bind:this={el}></div>
