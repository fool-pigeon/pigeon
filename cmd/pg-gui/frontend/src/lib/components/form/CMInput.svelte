<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import { EditorView } from 'codemirror';
	import CMWrapper from './CMWrapper.svelte';
	import { keymap, placeholder } from '@codemirror/view';
	import { defaultKeymap } from '@codemirror/commands';
	import { daisyColorTheme } from './cmConfig';

	type Prop = {
		value: string;
		placeholder?: string;
		noLineWrap?: boolean;
	};
	let { value = $bindable(), placeholder: p, noLineWrap = false }: Prop = $props();
	// TODO add cm_extension for highlight different env params
</script>

<CMWrapper
	bind:value
	viewInitFunc={() => {
		return new EditorView({
			extensions: [
				keymap.of(defaultKeymap),
				placeholder(p ?? ''),
				noLineWrap ? [] : EditorView.lineWrapping,
				// prevent drag selection
				EditorView.domEventHandlers({
					dragstart: (event) => {
						event.preventDefault();
						return true;
					},
				}),
				EditorView.theme({
					'&': { width: '100%', height: '100%' },
				}),
				daisyColorTheme,
			],
		});
	}}
/>
