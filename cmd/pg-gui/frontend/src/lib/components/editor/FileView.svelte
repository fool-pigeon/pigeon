<script lang="ts">
	import { defaultKeymap, indentWithTab } from '@codemirror/commands';
	import { keymap } from '@codemirror/view';
	import { basicSetup, EditorView } from 'codemirror';
	import { onDestroy, onMount } from 'svelte';

	let { value = $bindable('') } = $props();
	let el: HTMLDivElement;
	let view: EditorView;
	function newView(): EditorView {
		return new EditorView({
			doc: value,
			parent: el,
			extensions: [
				basicSetup,
				keymap.of([indentWithTab, ...defaultKeymap]),
				EditorView.updateListener.of((u) => {
					if (u.docChanged) {
						value = u.state.doc.toString();
					}
				}),
				EditorView.baseTheme({
					'&': { height: '100%' },
				}),
			],
		});
	}
	onMount(() => {
		view = newView();
	});
	onDestroy(() => {
		view.destroy();
	});
	$effect(() => {
		if (!view || value === view.state.doc.toString()) {
			return;
		}
		view.destroy();
		view = newView();
	});
</script>

<div class="w-full h-full" bind:this={el}></div>
